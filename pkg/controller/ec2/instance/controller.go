/*
Copyright 2021 The Crossplane Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package instance

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsec2 "github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/ratelimiter"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	svcapitypes "github.com/crossplane/provider-aws/apis/ec2/manualv1alpha1"
	awsclient "github.com/crossplane/provider-aws/pkg/clients"
	"github.com/crossplane/provider-aws/pkg/clients/ec2"
)

// TODO update errors
const (
	errUnexpectedObject = "The managed resource is not an VPC resource"
	errKubeUpdateFailed = "cannot update VPC custom resource"

	errDescribe            = "failed to describe VPC with id"
	errMultipleItems       = "retrieved multiple VPCs for the given vpcId"
	errCreate              = "failed to create the VPC resource"
	errUpdate              = "failed to update VPC resource"
	errModifyVPCAttributes = "failed to modify the VPC resource attributes"
	errCreateTags          = "failed to create tags for the VPC resource"
	errDelete              = "failed to delete the VPC resource"
)

// SetupInstance adds a controller that reconciles Instances.
func SetupInstance(mgr ctrl.Manager, l logging.Logger, rl workqueue.RateLimiter, poll time.Duration) error {
	name := managed.ControllerName(svcapitypes.InstanceGroupKind)
	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(controller.Options{
			RateLimiter: ratelimiter.NewDefaultManagedRateLimiter(rl),
		}).
		For(&svcapitypes.Instance{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(svcapitypes.InstanceGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), newClientFn: ec2.NewInstanceClient}),
			managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
			managed.WithConnectionPublishers(),
			managed.WithInitializers(managed.NewDefaultProviderConfig(mgr.GetClient())),
			managed.WithPollInterval(poll),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

type connector struct {
	kube        client.Client
	newClientFn func(config aws.Config) ec2.InstanceClient
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*svcapitypes.Instance)
	if !ok {
		return nil, errors.New(errUnexpectedObject)
	}
	cfg, err := awsclient.GetConfig(ctx, c.kube, mg, aws.StringValue(cr.Spec.ForProvider.Region))
	if err != nil {
		return nil, err
	}
	return &external{client: c.newClientFn(*cfg), kube: c.kube}, nil
}

type external struct {
	kube   client.Client
	client ec2.InstanceClient
}

func (e *external) Observe(ctx context.Context, mgd resource.Managed) (managed.ExternalObservation, error) { // nolint:gocyclo
	cr, ok := mgd.(*svcapitypes.Instance)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errUnexpectedObject)
	}

	if meta.GetExternalName(cr) == "" {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}

	response, err := e.client.DescribeInstancesRequest(&awsec2.DescribeInstancesInput{
		InstanceIds: []string{meta.GetExternalName(cr)},
	}).Send(ctx)

	// deleted instances that have not yet been cleaned up from the cluster return a
	// 200 OK with a nil response.Reservations slice
	if err == nil && len(response.Reservations) == 0 {
		return managed.ExternalObservation{}, nil
	}

	if err != nil {
		return managed.ExternalObservation{}, awsclient.Wrap(resource.Ignore(ec2.IsInstanceNotFoundErr, err), errDescribe)
	}

	// in a successful response, there should be one and only one object
	if len(response.Reservations[0].Instances) != 1 {
		return managed.ExternalObservation{}, errors.New(errMultipleItems)
	}

	observed := response.Reservations[0].Instances[0]

	// (tnthornton) Terminated instances remain visible on API calls for a time before
	// being automatically deleted. Rather than having the delete command hang for that
	// entire time, return an empty ExternalObservation in this case.
	//
	// ref: (https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/terminating-instances.html)
	if observed.State.Name == awsec2.InstanceStateNameTerminated {
		return managed.ExternalObservation{}, nil
	}

	// update the CRD spec for any new values from provider
	current := cr.Spec.ForProvider.DeepCopy()

	o := awsec2.DescribeInstanceAttributeOutput{}

	// for _, input := range []awsec2.InstanceAttributeName{
	// 	awsec2.VpcAttributeNameEnableDnsSupport,
	// 	awsec2.VpcAttributeNameEnableDnsHostnames,
	// } {
	// 	_, err := e.client.DescribeInstanceAttributeRequest(&awsec2.DescribeInstanceAttributeInput{
	// 		InstanceId: aws.String(meta.GetExternalName(cr)),
	// 		Attribute:  input,
	// 	}).Send(context.Background())

	// 	if err != nil {
	// 		return managed.ExternalObservation{}, awsclient.Wrap(err, errDescribe)
	// 	}

	// 	// if r.EnableDnsHostnames != nil {
	// 	// 	o.EnableDnsHostnames = r.EnableDnsHostnames
	// 	// }

	// 	// if r.EnableDnsSupport != nil {
	// 	// 	o.EnableDnsSupport = r.EnableDnsSupport
	// 	// }
	// }

	ec2.LateInitializeInstance(&cr.Spec.ForProvider, &observed, &o)

	switch observed.State.Name {
	case awsec2.InstanceStateNameRunning:
		cr.SetConditions(xpv1.Available())
	case awsec2.InstanceStateNamePending:
		cr.SetConditions(xpv1.Creating())
	case awsec2.InstanceStateNameShuttingDown:
		cr.SetConditions(xpv1.Deleting())
	}

	cr.Status.AtProvider = ec2.GenerateInstanceObservation(observed)

	return managed.ExternalObservation{
		ResourceExists:          true,
		ResourceUpToDate:        ec2.IsInstanceUpToDate(cr.Spec.ForProvider, observed, o),
		ResourceLateInitialized: !cmp.Equal(current, &cr.Spec.ForProvider),
	}, nil
}

func (e *external) Create(ctx context.Context, mgd resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mgd.(*svcapitypes.Instance)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errUnexpectedObject)
	}

	result, err := e.client.RunInstancesRequest(&awsec2.RunInstancesInput{
		// ClientToken:           cr.Spec.ForProvider.ClientToken,
		// DisableApiTermination: cr.Spec.ForProvider.DisableAPITermination,
		// DryRun:                cr.Spec.ForProvider.DryRun,
		// EbsOptimized:          cr.Spec.ForProvider.EBSOptimized,
		ImageId: cr.Spec.ForProvider.ImageID,
		// InstanceType:          awsec2.InstanceType(*cr.Spec.ForProvider.InstanceType), //optional
		// Ipv6AddressCount: cr.Spec.ForProvider.Ipv6AddressCount,
		// KernelId: cr.Spec.ForProvider.KernelID,
		// KeyName:  cr.Spec.ForProvider.KeyName,
		MaxCount: cr.Spec.ForProvider.MaxCount,
		MinCount: cr.Spec.ForProvider.MinCount,
		// Monitoring:       &awsec2.RunInstancesMonitoringEnabled{Enabled: aws.Bool(*cr.Spec.ForProvider.Monitoring)}, // default is returning an error
		// PrivateIpAddress: cr.Spec.ForProvider.PrivateIPAddress,
		// RamdiskId:        cr.Spec.ForProvider.RAMDiskID,
		// SecurityGroupIds: cr.Spec.ForProvider.SecurityGroupIDs,
		// TODO fill in refs

		// SecurityGroups: cr.Spec.ForProvider.SecurityGroups,
		// SubnetId:       cr.Spec.ForProvider.SubnetID,

		UserData: cr.Spec.ForProvider.UserData,
		// special type of tag for specifying the instance name. probably want to allow it to be overridden by the spec, but
		// maybe set it from the metadata.Name if not set in spec?
		TagSpecifications: []awsec2.TagSpecification{
			{
				ResourceType: awsec2.ResourceType("instance"),
				Tags:         []awsec2.Tag{{Key: aws.String("Name"), Value: aws.String(mgd.GetName())}},
			},
		},
	}).Send(ctx)
	if err != nil {
		return managed.ExternalCreation{}, awsclient.Wrap(err, errCreate)
	}

	meta.SetExternalName(cr, aws.StringValue(result.Instances[0].InstanceId))

	return managed.ExternalCreation{ExternalNameAssigned: true}, nil
}

func (e *external) Update(ctx context.Context, mgd resource.Managed) (managed.ExternalUpdate, error) {
	_, ok := mgd.(*svcapitypes.Instance)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errUnexpectedObject)
	}

	// if cr.Spec.ForProvider.EnableDNSSupport != nil {
	// 	modifyInput := &awsec2.ModifyInstanceAttributeInput{
	// 		InstanceId:       aws.String(meta.GetExternalName(cr)),
	// 		EnableDnsSupport: &awsec2.AttributeBooleanValue{Value: cr.Spec.ForProvider.EnableDNSSupport},
	// 	}
	// 	if _, err := e.client.ModifyInstanceAttributeRequest(modifyInput).Send(ctx); err != nil {
	// 		return managed.ExternalUpdate{}, awsclient.Wrap(err, errModifyVPCAttributes)
	// 	}
	// }

	// if cr.Spec.ForProvider.EnableDNSHostNames != nil {
	// 	modifyInput := &awsec2.ModifyVpcAttributeInput{
	// 		VpcId:              aws.String(meta.GetExternalName(cr)),
	// 		EnableDnsHostnames: &awsec2.AttributeBooleanValue{Value: cr.Spec.ForProvider.EnableDNSHostNames},
	// 	}
	// 	if _, err := e.client.ModifyVpcAttributeRequest(modifyInput).Send(ctx); err != nil {
	// 		return managed.ExternalUpdate{}, awsclient.Wrap(err, errModifyVPCAttributes)
	// 	}
	// }

	// // NOTE(muvaf): VPCs can only be tagged after the creation and this request
	// // is idempotent.
	// if _, err := e.client.CreateTagsRequest(&awsec2.CreateTagsInput{
	// 	Resources: []string{meta.GetExternalName(cr)},
	// 	Tags:      v1beta1.GenerateEC2Tags(cr.Spec.ForProvider.Tags),
	// }).Send(ctx); err != nil {
	// 	return managed.ExternalUpdate{}, awsclient.Wrap(err, errCreateTags)
	// }

	// _, err := e.client.ModifyVpcTenancyRequest(&awsec2.ModifyVpcTenancyInput{
	// 	InstanceTenancy: awsec2.VpcTenancy(aws.StringValue(cr.Spec.ForProvider.InstanceTenancy)),
	// 	VpcId:           aws.String(meta.GetExternalName(cr)),
	// }).Send(ctx)

	return managed.ExternalUpdate{}, awsclient.Wrap(errors.New("fix this"), errUpdate)
}

func (e *external) Delete(ctx context.Context, mgd resource.Managed) error {
	cr, ok := mgd.(*svcapitypes.Instance)
	if !ok {
		return errors.New(errUnexpectedObject)
	}

	cr.Status.SetConditions(xpv1.Deleting())

	_, err := e.client.TerminateInstancesRequest(&awsec2.TerminateInstancesInput{
		InstanceIds: []string{meta.GetExternalName(cr)}, // TODO handle a list of instances
	}).Send(ctx)

	return awsclient.Wrap(resource.Ignore(ec2.IsVPCNotFoundErr, err), errDelete)
}
