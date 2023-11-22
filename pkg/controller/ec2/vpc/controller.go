/*
Copyright 2019 The Crossplane Authors.

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

package vpc

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsec2 "github.com/aws/aws-sdk-go-v2/service/ec2"
	awsec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane-contrib/provider-aws/apis/ec2/v1beta1"
	"github.com/crossplane-contrib/provider-aws/apis/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/ec2"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
	connectaws "github.com/crossplane-contrib/provider-aws/pkg/utils/connect/aws"
	errorutils "github.com/crossplane-contrib/provider-aws/pkg/utils/errors"
	custommanaged "github.com/crossplane-contrib/provider-aws/pkg/utils/reconciler/managed"
)

const (
	errUnexpectedObject = "The managed resource is not an VPC resource"
	errKubeUpdateFailed = "cannot update VPC custom resource"

	errDescribe            = "failed to describe VPC with id"
	errMultipleItems       = "retrieved multiple VPCs for the given vpcId"
	errCreate              = "failed to create the VPC resource"
	errUpdate              = "failed to update VPC resource"
	errModifyVPCAttributes = "failed to modify the VPC resource attributes"
	errCreateTags          = "failed to create tags for the VPC resource"
	errDeleteTags          = "failed to delete tags for the VPC resource"
	errDelete              = "failed to delete the VPC resource"
)

// SetupVPC adds a controller that reconciles VPCs.
func SetupVPC(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1beta1.VPCGroupKind)

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), v1alpha1.StoreConfigGroupVersionKind))
	}

	reconcilerOpts := []managed.ReconcilerOption{
		managed.WithCriticalAnnotationUpdater(custommanaged.NewRetryingCriticalAnnotationUpdater(mgr.GetClient())),
		managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), newClientFn: ec2.NewVPCClient}),
		managed.WithCreationGracePeriod(3 * time.Minute),
		managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
		managed.WithConnectionPublishers(),
		managed.WithInitializers(),
		managed.WithPollInterval(o.PollInterval),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		managed.WithConnectionPublishers(cps...),
	}

	if o.Features.Enabled(features.EnableAlphaManagementPolicies) {
		reconcilerOpts = append(reconcilerOpts, managed.WithManagementPolicies())
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1beta1.VPCGroupVersionKind),
		reconcilerOpts...)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&v1beta1.VPC{}).
		Complete(r)
}

type connector struct {
	kube        client.Client
	newClientFn func(config aws.Config) ec2.VPCClient
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1beta1.VPC)
	if !ok {
		return nil, errors.New(errUnexpectedObject)
	}
	cfg, err := connectaws.GetConfig(ctx, c.kube, mg, aws.ToString(cr.Spec.ForProvider.Region))
	if err != nil {
		return nil, err
	}
	return &external{client: c.newClientFn(*cfg), kube: c.kube}, nil
}

type external struct {
	kube   client.Client
	client ec2.VPCClient
}

func (e *external) Observe(ctx context.Context, mgd resource.Managed) (managed.ExternalObservation, error) { //nolint:gocyclo
	cr, ok := mgd.(*v1beta1.VPC)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errUnexpectedObject)
	}

	if meta.GetExternalName(cr) == "" {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}

	response, err := e.client.DescribeVpcs(ctx, &awsec2.DescribeVpcsInput{
		VpcIds: []string{meta.GetExternalName(cr)},
	})
	if err != nil {
		return managed.ExternalObservation{}, errorutils.Wrap(resource.Ignore(ec2.IsVPCNotFoundErr, err), errDescribe)
	}

	// in a successful response, there should be one and only one object
	if len(response.Vpcs) != 1 {
		return managed.ExternalObservation{}, errors.New(errMultipleItems)
	}

	observed := response.Vpcs[0]

	o := awsec2.DescribeVpcAttributeOutput{}

	for _, input := range []awsec2types.VpcAttributeName{
		awsec2types.VpcAttributeNameEnableDnsSupport,
		awsec2types.VpcAttributeNameEnableDnsHostnames,
	} {
		r, err := e.client.DescribeVpcAttribute(ctx, &awsec2.DescribeVpcAttributeInput{
			VpcId:     aws.String(meta.GetExternalName(cr)),
			Attribute: input,
		})

		if err != nil {
			return managed.ExternalObservation{}, errorutils.Wrap(err, errDescribe)
		}

		if r.EnableDnsHostnames != nil {
			o.EnableDnsHostnames = r.EnableDnsHostnames
		}

		if r.EnableDnsSupport != nil {
			o.EnableDnsSupport = r.EnableDnsSupport
		}
	}

	// update the CRD spec for any new values from provider
	current := cr.Spec.ForProvider.DeepCopy()
	ec2.LateInitializeVPC(&cr.Spec.ForProvider, &observed, &o)

	switch observed.State {
	case awsec2types.VpcStateAvailable:
		cr.SetConditions(xpv1.Available())
	case awsec2types.VpcStatePending:
		cr.SetConditions(xpv1.Creating())
	}

	cr.Status.AtProvider = ec2.GenerateVpcObservation(observed)

	ec2.LateInitializeVPC(&cr.Spec.ForProvider, &observed, &o)

	switch observed.State {
	case awsec2types.VpcStateAvailable:
		cr.SetConditions(xpv1.Available())
	case awsec2types.VpcStatePending:
		cr.SetConditions(xpv1.Creating())
	}

	cr.Status.AtProvider = ec2.GenerateVpcObservation(observed)

	return managed.ExternalObservation{
		ResourceExists:          true,
		ResourceUpToDate:        ec2.IsVpcUpToDate(cr.Spec.ForProvider, observed, o),
		ResourceLateInitialized: !cmp.Equal(current, &cr.Spec.ForProvider),
	}, nil
}

func (e *external) Create(ctx context.Context, mgd resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mgd.(*v1beta1.VPC)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errUnexpectedObject)
	}

	result, err := e.client.CreateVpc(ctx, &awsec2.CreateVpcInput{
		CidrBlock:                   aws.String(cr.Spec.ForProvider.CIDRBlock),
		Ipv6CidrBlock:               cr.Spec.ForProvider.Ipv6CIDRBlock,
		AmazonProvidedIpv6CidrBlock: cr.Spec.ForProvider.AmazonProvidedIpv6CIDRBlock,
		Ipv6Pool:                    cr.Spec.ForProvider.Ipv6Pool,
		InstanceTenancy:             awsec2types.Tenancy(aws.ToString(cr.Spec.ForProvider.InstanceTenancy)),
	})
	if err != nil {
		return managed.ExternalCreation{}, errorutils.Wrap(err, errCreate)
	}

	meta.SetExternalName(cr, aws.ToString(result.Vpc.VpcId))

	return managed.ExternalCreation{}, nil
}

func (e *external) Update(ctx context.Context, mgd resource.Managed) (managed.ExternalUpdate, error) { //nolint:gocyclo
	cr, ok := mgd.(*v1beta1.VPC)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errUnexpectedObject)
	}

	response, err := e.client.DescribeVpcs(ctx, &awsec2.DescribeVpcsInput{
		VpcIds: []string{meta.GetExternalName(cr)},
	})

	if err != nil {
		return managed.ExternalUpdate{}, errorutils.Wrap(resource.Ignore(ec2.IsSubnetNotFoundErr, err), errDescribe)
	}

	if response.Vpcs == nil {
		return managed.ExternalUpdate{}, errors.New(errUpdate)
	}

	vpc := response.Vpcs[0]

	if cr.Spec.ForProvider.EnableDNSSupport != nil {
		modifyInput := &awsec2.ModifyVpcAttributeInput{
			VpcId:            aws.String(meta.GetExternalName(cr)),
			EnableDnsSupport: &awsec2types.AttributeBooleanValue{Value: cr.Spec.ForProvider.EnableDNSSupport},
		}
		if _, err := e.client.ModifyVpcAttribute(ctx, modifyInput); err != nil {
			return managed.ExternalUpdate{}, errorutils.Wrap(err, errModifyVPCAttributes)
		}
	}

	if cr.Spec.ForProvider.EnableDNSHostNames != nil {
		modifyInput := &awsec2.ModifyVpcAttributeInput{
			VpcId:              aws.String(meta.GetExternalName(cr)),
			EnableDnsHostnames: &awsec2types.AttributeBooleanValue{Value: cr.Spec.ForProvider.EnableDNSHostNames},
		}
		if _, err := e.client.ModifyVpcAttribute(ctx, modifyInput); err != nil {
			return managed.ExternalUpdate{}, errorutils.Wrap(err, errModifyVPCAttributes)
		}
	}

	add, remove := ec2.DiffEC2Tags(ec2.GenerateEC2TagsV1Beta1(cr.Spec.ForProvider.Tags), vpc.Tags)
	if len(remove) > 0 {
		if _, err := e.client.DeleteTags(ctx, &awsec2.DeleteTagsInput{
			Resources: []string{meta.GetExternalName(cr)},
			Tags:      remove,
		}); err != nil {
			return managed.ExternalUpdate{}, errorutils.Wrap(err, errDeleteTags)
		}
	}

	if len(add) > 0 {
		if _, err := e.client.CreateTags(ctx, &awsec2.CreateTagsInput{
			Resources: []string{meta.GetExternalName(cr)},
			Tags:      add,
		}); err != nil {
			return managed.ExternalUpdate{}, errorutils.Wrap(err, errCreateTags)
		}
	}

	_, err = e.client.ModifyVpcTenancy(ctx, &awsec2.ModifyVpcTenancyInput{
		InstanceTenancy: awsec2types.VpcTenancy(aws.ToString(cr.Spec.ForProvider.InstanceTenancy)),
		VpcId:           aws.String(meta.GetExternalName(cr)),
	})

	return managed.ExternalUpdate{}, errorutils.Wrap(err, errUpdate)
}

func (e *external) Delete(ctx context.Context, mgd resource.Managed) error {
	cr, ok := mgd.(*v1beta1.VPC)
	if !ok {
		return errors.New(errUnexpectedObject)
	}

	cr.Status.SetConditions(xpv1.Deleting())

	_, err := e.client.DeleteVpc(ctx, &awsec2.DeleteVpcInput{
		VpcId: aws.String(meta.GetExternalName(cr)),
	})

	return errorutils.Wrap(resource.Ignore(ec2.IsVPCNotFoundErr, err), errDelete)
}
