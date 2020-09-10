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

package subnet

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsec2 "github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane/provider-aws/apis/ec2/v1beta1"
	awscommon "github.com/crossplane/provider-aws/pkg/clients"
	"github.com/crossplane/provider-aws/pkg/clients/ec2"
)

const (
	errUnexpectedObject = "The managed resource is not an Subnet resource"
	errKubeUpdateFailed = "cannot update Subnet custom resource"

	errDescribe      = "failed to describe Subnet"
	errMultipleItems = "retrieved multiple Subnets"
	errCreate        = "failed to create the Subnet resource"
	errDelete        = "failed to delete the Subnet resource"
	errUpdate        = "failed to update the Subnet resource"
	errSpecUpdate    = "cannot update spec of the Subnet custom resource"
	errStatusUpdate  = "cannot update status of the Subnet custom resource"
	errCreateTags    = "failed to create tags for the Subnet resource"
)

// SetupSubnet adds a controller that reconciles Subnets.
func SetupSubnet(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(v1beta1.SubnetGroupKind)
	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1beta1.Subnet{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1beta1.SubnetGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), newClientFn: ec2.NewSubnetClient}),
			managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
			managed.WithInitializers(),
			managed.WithConnectionPublishers(),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

type connector struct {
	kube        client.Client
	newClientFn func(config aws.Config) ec2.SubnetClient
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cfg, err := awscommon.GetConfig(ctx, c.kube, mg, "")
	if err != nil {
		return nil, err
	}
	return &external{client: c.newClientFn(*cfg), kube: c.kube}, nil
}

type external struct {
	kube   client.Client
	client ec2.SubnetClient
}

func (e *external) Observe(ctx context.Context, mgd resource.Managed) (managed.ExternalObservation, error) { // nolint:gocyclo
	cr, ok := mgd.(*v1beta1.Subnet)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errUnexpectedObject)
	}

	if meta.GetExternalName(cr) == "" {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}

	response, err := e.client.DescribeSubnetsRequest(&awsec2.DescribeSubnetsInput{
		SubnetIds: []string{meta.GetExternalName(cr)},
	}).Send(ctx)

	if err != nil {
		// when the subnet is deleted externally and we sent a create request with the cr data in k8s
		// either of AvailabilityZone or AvailabilityZoneID needs to be set
		if cr.Spec.ForProvider.AvailabilityZone != nil && cr.Spec.ForProvider.AvailabilityZoneID != nil {
			cr.Spec.ForProvider.AvailabilityZoneID = nil
			if err := e.kube.Update(ctx, cr); err != nil {
				return managed.ExternalObservation{}, errors.Wrap(err, errKubeUpdateFailed)
			}
		}
		return managed.ExternalObservation{}, errors.Wrap(resource.Ignore(ec2.IsSubnetNotFoundErr, err), errDescribe)
	}

	// in a successful response, there should be one and only one object
	if len(response.Subnets) != 1 {
		return managed.ExternalObservation{}, errors.New(errMultipleItems)
	}

	observed := response.Subnets[0]

	// update CRD spec for any new values from provider
	current := cr.Spec.ForProvider.DeepCopy()
	ec2.LateInitializeSubnet(&cr.Spec.ForProvider, &observed)
	if !cmp.Equal(current, &cr.Spec.ForProvider) {
		if err := e.kube.Update(ctx, cr); err != nil {
			return managed.ExternalObservation{}, errors.Wrap(err, errKubeUpdateFailed)
		}
	}

	switch observed.State {
	case awsec2.SubnetStateAvailable:
		cr.SetConditions(runtimev1alpha1.Available())
	case awsec2.SubnetStatePending:
		cr.SetConditions(runtimev1alpha1.Creating())
	}

	cr.Status.AtProvider = ec2.GenerateSubnetObservation(observed)

	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: ec2.IsSubnetUpToDate(cr.Spec.ForProvider, observed),
	}, nil
}

func (e *external) Create(ctx context.Context, mgd resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mgd.(*v1beta1.Subnet)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errUnexpectedObject)
	}

	if cr.Spec.ForProvider.AvailabilityZone != nil && cr.Spec.ForProvider.AvailabilityZoneID != nil {
		return managed.ExternalCreation{}, errors.New("Both AvailabilityZone and AvailabilityZoneID cannot be passed")
	}

	cr.Status.SetConditions(runtimev1alpha1.Creating())
	if err := e.kube.Status().Update(ctx, cr); err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errStatusUpdate)
	}

	result, err := e.client.CreateSubnetRequest(&awsec2.CreateSubnetInput{
		AvailabilityZone:   cr.Spec.ForProvider.AvailabilityZone,
		AvailabilityZoneId: cr.Spec.ForProvider.AvailabilityZoneID,
		CidrBlock:          aws.String(cr.Spec.ForProvider.CIDRBlock),
		Ipv6CidrBlock:      cr.Spec.ForProvider.IPv6CIDRBlock,
		VpcId:              cr.Spec.ForProvider.VPCID,
	}).Send(ctx)

	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreate)
	}

	meta.SetExternalName(cr, aws.StringValue(result.Subnet.SubnetId))

	return managed.ExternalCreation{}, errors.Wrap(e.kube.Update(ctx, cr), errSpecUpdate)
}

func (e *external) Update(ctx context.Context, mgd resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mgd.(*v1beta1.Subnet)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errUnexpectedObject)
	}

	response, err := e.client.DescribeSubnetsRequest(&awsec2.DescribeSubnetsInput{
		SubnetIds: []string{meta.GetExternalName(cr)},
	}).Send(ctx)

	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrapf(resource.Ignore(ec2.IsSubnetNotFoundErr, err), errDescribe)
	}

	if response.Subnets == nil {
		return managed.ExternalUpdate{}, errors.New(errUpdate)
	}

	subnet := response.Subnets[0]

	if !v1beta1.CompareTags(cr.Spec.ForProvider.Tags, subnet.Tags) {
		if _, err := e.client.CreateTagsRequest(&awsec2.CreateTagsInput{
			Resources: []string{meta.GetExternalName(cr)},
			Tags:      v1beta1.GenerateEC2Tags(cr.Spec.ForProvider.Tags),
		}).Send(ctx); err != nil {
			return managed.ExternalUpdate{}, errors.Wrap(err, errCreateTags)
		}
	}

	if subnet.MapPublicIpOnLaunch != cr.Spec.ForProvider.MapPublicIPOnLaunch {
		_, err = e.client.ModifySubnetAttributeRequest(&awsec2.ModifySubnetAttributeInput{
			MapPublicIpOnLaunch: &awsec2.AttributeBooleanValue{
				Value: cr.Spec.ForProvider.MapPublicIPOnLaunch,
			},
			SubnetId: aws.String(meta.GetExternalName(cr)),
		}).Send((ctx))
		if err != nil {
			return managed.ExternalUpdate{}, errors.Wrap(err, errUpdate)
		}
	}

	if subnet.AssignIpv6AddressOnCreation != cr.Spec.ForProvider.AssignIPv6AddressOnCreation {
		_, err = e.client.ModifySubnetAttributeRequest(&awsec2.ModifySubnetAttributeInput{
			AssignIpv6AddressOnCreation: &awsec2.AttributeBooleanValue{
				Value: cr.Spec.ForProvider.AssignIPv6AddressOnCreation,
			},
			SubnetId: aws.String(meta.GetExternalName(cr)),
		}).Send((ctx))
	}

	return managed.ExternalUpdate{}, errors.Wrap(err, errUpdate)
}

func (e *external) Delete(ctx context.Context, mgd resource.Managed) error {
	cr, ok := mgd.(*v1beta1.Subnet)
	if !ok {
		return errors.New(errUnexpectedObject)
	}

	cr.Status.SetConditions(runtimev1alpha1.Deleting())

	_, err := e.client.DeleteSubnetRequest(&awsec2.DeleteSubnetInput{
		SubnetId: aws.String(meta.GetExternalName(cr)),
	}).Send(ctx)

	return errors.Wrap(resource.Ignore(ec2.IsSubnetNotFoundErr, err), errDelete)
}
