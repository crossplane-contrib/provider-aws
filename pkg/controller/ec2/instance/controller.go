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
	"sort"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsec2 "github.com/aws/aws-sdk-go-v2/service/ec2"
	types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/ec2/manualv1alpha1"
	"github.com/crossplane-contrib/provider-aws/apis/v1alpha1"
	awsclient "github.com/crossplane-contrib/provider-aws/pkg/clients"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/ec2"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
)

const (
	errUnexpectedObject = "The managed resource is not an Instance resource"
	errKubeUpdateFailed = "cannot update Instance custom resource"

	errMultipleItems            = "retrieved multiple Instances for the given instanceId"
	errDescribe                 = "failed to describe Instance with id"
	errCreate                   = "failed to create the Instance resource"
	errUpdate                   = "failed to update Instance resource"
	errModifyInstanceAttributes = "failed to modify the Instance resource attributes"
	errCreateTags               = "failed to create tags for the Instance resource"
	errDelete                   = "failed to delete the Instance resource"
)

// SetupInstance adds a controller that reconciles Instances.
func SetupInstance(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(svcapitypes.InstanceGroupKind)

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), v1alpha1.StoreConfigGroupVersionKind))
	}

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		For(&svcapitypes.Instance{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(svcapitypes.InstanceGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), newClientFn: ec2.NewInstanceClient}),
			managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
			managed.WithConnectionPublishers(),
			managed.WithInitializers(managed.NewDefaultProviderConfig(mgr.GetClient()), &tagger{kube: mgr.GetClient()}),
			managed.WithPollInterval(o.PollInterval),
			managed.WithLogger(o.Logger.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
			managed.WithConnectionPublishers(cps...)))
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
	cfg, err := awsclient.GetConfig(ctx, c.kube, mg, awsclient.StringValue(cr.Spec.ForProvider.Region))
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

	response, err := e.client.DescribeInstances(ctx,
		&awsec2.DescribeInstancesInput{
			InstanceIds: []string{meta.GetExternalName(cr)},
		})

	// deleted instances that have not yet been cleaned up from the cluster return a
	// 200 OK with a nil response.Reservations slice
	if err == nil && len(response.Reservations) == 0 {
		return managed.ExternalObservation{}, nil
	}

	if err != nil {
		return managed.ExternalObservation{},
			awsclient.Wrap(resource.Ignore(ec2.IsInstanceNotFoundErr, err), errDescribe)
	}

	// in a successful response, there should be one and only one object
	if len(response.Reservations[0].Instances) != 1 {
		return managed.ExternalObservation{}, errors.New(errMultipleItems)
	}

	observed := response.Reservations[0].Instances[0]

	// update the CRD spec for any new values from provider
	current := cr.Spec.ForProvider.DeepCopy()

	o := awsec2.DescribeInstanceAttributeOutput{}

	for _, input := range []types.InstanceAttributeName{
		types.InstanceAttributeNameDisableApiTermination,
		types.InstanceAttributeNameEbsOptimized,
		types.InstanceAttributeNameInstanceInitiatedShutdownBehavior,
		types.InstanceAttributeNameInstanceType,
		types.InstanceAttributeNameKernel,
		types.InstanceAttributeNameRamdisk,
		types.InstanceAttributeNameUserData,
	} {
		r, err := e.client.DescribeInstanceAttribute(ctx, &awsec2.DescribeInstanceAttributeInput{
			InstanceId: aws.String(meta.GetExternalName(cr)),
			Attribute:  input,
		})

		if err != nil {
			return managed.ExternalObservation{}, awsclient.Wrap(err, errDescribe)
		}

		if r.DisableApiTermination != nil {
			o.DisableApiTermination = r.DisableApiTermination
		}

		if r.EbsOptimized != nil {
			o.EbsOptimized = r.EbsOptimized
		}

		if r.InstanceInitiatedShutdownBehavior != nil {
			o.InstanceInitiatedShutdownBehavior = r.InstanceInitiatedShutdownBehavior
		}

		if r.InstanceType != nil {
			o.InstanceType = r.InstanceType
		}

		if r.KernelId != nil {
			o.KernelId = r.KernelId
		}

		if r.RamdiskId != nil {
			o.RamdiskId = r.RamdiskId
		}

		if r.UserData != nil {
			o.UserData = r.UserData
		}
	}

	ec2.LateInitializeInstance(&cr.Spec.ForProvider, &observed, &o)

	if !cmp.Equal(current, &cr.Spec.ForProvider) {
		if err := e.kube.Update(ctx, cr); err != nil {
			return managed.ExternalObservation{}, errors.Wrap(err, errKubeUpdateFailed)
		}
	}

	observation := ec2.GenerateInstanceObservation(observed)
	condition := ec2.GenerateInstanceCondition(observation)

	switch condition {
	case ec2.Creating:
		cr.SetConditions(xpv1.Creating())
	case ec2.Available:
		cr.SetConditions(xpv1.Available())
	case ec2.Deleting:
		cr.SetConditions(xpv1.Deleting())
	case ec2.Deleted:
		// Terminated instances remain visible on API calls for a time before
		// being automatically deleted. Rather than having the delete command
		// hang for that entire time, return an empty ExternalObservation in
		// this case.
		//
		// ref: (https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/terminating-instances.html)
		return managed.ExternalObservation{}, nil
	}

	cr.Status.AtProvider = observation

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

	result, err := e.client.RunInstances(ctx,
		ec2.GenerateEC2RunInstancesInput(mgd.GetName(), &cr.Spec.ForProvider),
	)
	if err != nil {
		return managed.ExternalCreation{}, awsclient.Wrap(err, errCreate)
	}

	instance := result.Instances[0]

	if _, err := e.client.CreateTags(ctx, &awsec2.CreateTagsInput{
		Resources: []string{awsclient.StringValue(instance.InstanceId)},
		Tags:      svcapitypes.GenerateEC2Tags(cr.Spec.ForProvider.Tags),
	}); err != nil {
		return managed.ExternalCreation{ExternalNameAssigned: false}, awsclient.Wrap(err, errCreateTags)
	}

	meta.SetExternalName(cr, awsclient.StringValue(instance.InstanceId))

	return managed.ExternalCreation{ExternalNameAssigned: true}, nil
}

func (e *external) Update(ctx context.Context, mgd resource.Managed) (managed.ExternalUpdate, error) { // nolint:gocyclo
	cr, ok := mgd.(*svcapitypes.Instance)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errUnexpectedObject)
	}

	if cr.Spec.ForProvider.DisableAPITermination != nil {
		modifyInput := &awsec2.ModifyInstanceAttributeInput{
			InstanceId: aws.String(meta.GetExternalName(cr)),
			DisableApiTermination: &types.AttributeBooleanValue{
				Value: cr.Spec.ForProvider.DisableAPITermination,
			},
		}
		_, err := e.client.ModifyInstanceAttribute(ctx, modifyInput)

		if err != nil {
			return managed.ExternalUpdate{}, awsclient.Wrap(err, errModifyInstanceAttributes)
		}
	}

	if cr.Spec.ForProvider.InstanceInitiatedShutdownBehavior != "" {
		modifyInput := &awsec2.ModifyInstanceAttributeInput{
			InstanceId: aws.String(meta.GetExternalName(cr)),
			InstanceInitiatedShutdownBehavior: &types.AttributeValue{
				Value: aws.String(cr.Spec.ForProvider.InstanceInitiatedShutdownBehavior),
			},
		}
		_, err := e.client.ModifyInstanceAttribute(ctx, modifyInput)

		if err != nil {
			return managed.ExternalUpdate{}, awsclient.Wrap(err, errModifyInstanceAttributes)
		}
	}

	if cr.Spec.ForProvider.KernelID != nil {
		modifyInput := &awsec2.ModifyInstanceAttributeInput{
			InstanceId: aws.String(meta.GetExternalName(cr)),
			Kernel: &types.AttributeValue{
				Value: cr.Spec.ForProvider.KernelID,
			},
		}
		_, err := e.client.ModifyInstanceAttribute(ctx, modifyInput)

		if err != nil {
			return managed.ExternalUpdate{}, awsclient.Wrap(err, errModifyInstanceAttributes)
		}
	}

	if cr.Spec.ForProvider.RAMDiskID != nil {
		modifyInput := &awsec2.ModifyInstanceAttributeInput{
			InstanceId: aws.String(meta.GetExternalName(cr)),
			Ramdisk: &types.AttributeValue{
				Value: cr.Spec.ForProvider.RAMDiskID,
			},
		}
		_, err := e.client.ModifyInstanceAttribute(ctx, modifyInput)

		if err != nil {
			return managed.ExternalUpdate{}, awsclient.Wrap(err, errModifyInstanceAttributes)
		}
	}

	if cr.Spec.ForProvider.UserData != nil {
		modifyInput := &awsec2.ModifyInstanceAttributeInput{
			InstanceId: aws.String(meta.GetExternalName(cr)),
			UserData: &types.BlobAttributeValue{
				Value: []byte(*cr.Spec.ForProvider.UserData),
			},
		}
		_, err := e.client.ModifyInstanceAttribute(ctx, modifyInput)

		if err != nil {
			return managed.ExternalUpdate{}, awsclient.Wrap(err, errModifyInstanceAttributes)
		}
	}

	_, err := e.client.CreateTags(ctx, &awsec2.CreateTagsInput{
		Resources: []string{meta.GetExternalName(cr)},
		Tags:      svcapitypes.GenerateEC2Tags(cr.Spec.ForProvider.Tags),
	})

	return managed.ExternalUpdate{}, awsclient.Wrap(err, errUpdate)
}

func (e *external) Delete(ctx context.Context, mgd resource.Managed) error {
	cr, ok := mgd.(*svcapitypes.Instance)
	if !ok {
		return errors.New(errUnexpectedObject)
	}

	cr.Status.SetConditions(xpv1.Deleting())

	_, err := e.client.TerminateInstances(ctx, &awsec2.TerminateInstancesInput{
		InstanceIds: []string{meta.GetExternalName(cr)},
	})

	return awsclient.Wrap(resource.Ignore(ec2.IsInstanceNotFoundErr, err), errDelete)
}

type tagger struct {
	kube client.Client
}

func (t *tagger) Initialize(ctx context.Context, mgd resource.Managed) error {
	cr, ok := mgd.(*svcapitypes.Instance)
	if !ok {
		return errors.New(errUnexpectedObject)
	}
	tagMap := map[string]string{}
	for _, t := range cr.Spec.ForProvider.Tags {
		tagMap[t.Key] = t.Value
	}
	for k, v := range resource.GetExternalTags(mgd) {
		tagMap[k] = v
	}
	cr.Spec.ForProvider.Tags = make([]svcapitypes.Tag, len(tagMap))
	i := 0
	for k, v := range tagMap {
		cr.Spec.ForProvider.Tags[i] = svcapitypes.Tag{Key: k, Value: v}
		i++
	}
	sort.Slice(cr.Spec.ForProvider.Tags, func(i, j int) bool {
		return cr.Spec.ForProvider.Tags[i].Key < cr.Spec.ForProvider.Tags[j].Key
	})
	return errors.Wrap(t.kube.Update(ctx, cr), errKubeUpdateFailed)
}
