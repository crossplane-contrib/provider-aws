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
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsec2 "github.com/aws/aws-sdk-go-v2/service/ec2"
	awsec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
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

	"github.com/crossplane/provider-aws/apis/ec2/v1beta1"
	awsclient "github.com/crossplane/provider-aws/pkg/clients"
	"github.com/crossplane/provider-aws/pkg/clients/ec2"
)

const (
	errUnexpectedObject = "The managed resource is not an Subnet resource"

	errDescribe      = "failed to describe Subnet"
	errMultipleItems = "retrieved multiple Subnets"
	errCreate        = "failed to create the Subnet resource"
	errDelete        = "failed to delete the Subnet resource"
	errUpdate        = "failed to update the Subnet resource"
	errCreateTags    = "failed to create tags for the Subnet resource"
)

// SetupSubnet adds a controller that reconciles Subnets.
func SetupSubnet(mgr ctrl.Manager, l logging.Logger, rl workqueue.RateLimiter, poll time.Duration) error {
	name := managed.ControllerName(v1beta1.SubnetGroupKind)
	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(controller.Options{
			RateLimiter: ratelimiter.NewController(rl),
		}).
		For(&v1beta1.Subnet{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1beta1.SubnetGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), newClientFn: ec2.NewSubnetClient}),
			managed.WithCreationGracePeriod(3*time.Minute),
			managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
			managed.WithInitializers(),
			managed.WithConnectionPublishers(),
			managed.WithPollInterval(poll),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

type connector struct {
	kube        client.Client
	newClientFn func(config aws.Config) ec2.SubnetClient
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1beta1.Subnet)
	if !ok {
		return nil, errors.New(errUnexpectedObject)
	}
	cfg, err := awsclient.GetConfig(ctx, c.kube, mg, aws.ToString(cr.Spec.ForProvider.Region))
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

	response, err := e.client.DescribeSubnets(ctx, &awsec2.DescribeSubnetsInput{
		SubnetIds: []string{meta.GetExternalName(cr)},
	})

	if err != nil {
		return managed.ExternalObservation{}, awsclient.Wrap(resource.Ignore(ec2.IsSubnetNotFoundErr, err), errDescribe)
	}

	// in a successful response, there should be one and only one object
	if len(response.Subnets) != 1 {
		return managed.ExternalObservation{}, errors.New(errMultipleItems)
	}

	observed := response.Subnets[0]

	// update CRD spec for any new values from provider
	current := cr.Spec.ForProvider.DeepCopy()
	ec2.LateInitializeSubnet(&cr.Spec.ForProvider, &observed)

	switch observed.State {
	case awsec2types.SubnetStateAvailable:
		cr.SetConditions(xpv1.Available())
	case awsec2types.SubnetStatePending:
		cr.SetConditions(xpv1.Creating())
	}

	cr.Status.AtProvider = ec2.GenerateSubnetObservation(observed)

	return managed.ExternalObservation{
		ResourceExists:          true,
		ResourceUpToDate:        ec2.IsSubnetUpToDate(cr.Spec.ForProvider, observed),
		ResourceLateInitialized: !cmp.Equal(current, &cr.Spec.ForProvider),
	}, nil
}

func (e *external) Create(ctx context.Context, mgd resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mgd.(*v1beta1.Subnet)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errUnexpectedObject)
	}

	result, err := e.client.CreateSubnet(ctx, &awsec2.CreateSubnetInput{
		AvailabilityZone:   cr.Spec.ForProvider.AvailabilityZone,
		AvailabilityZoneId: cr.Spec.ForProvider.AvailabilityZoneID,
		CidrBlock:          aws.String(cr.Spec.ForProvider.CIDRBlock),
		Ipv6CidrBlock:      cr.Spec.ForProvider.IPv6CIDRBlock,
		VpcId:              cr.Spec.ForProvider.VPCID,
	})

	if err != nil {
		return managed.ExternalCreation{}, awsclient.Wrap(err, errCreate)
	}

	meta.SetExternalName(cr, aws.ToString(result.Subnet.SubnetId))

	return managed.ExternalCreation{}, nil
}

func (e *external) Update(ctx context.Context, mgd resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mgd.(*v1beta1.Subnet)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errUnexpectedObject)
	}

	response, err := e.client.DescribeSubnets(ctx, &awsec2.DescribeSubnetsInput{
		SubnetIds: []string{meta.GetExternalName(cr)},
	})

	if err != nil {
		return managed.ExternalUpdate{}, awsclient.Wrap(resource.Ignore(ec2.IsSubnetNotFoundErr, err), errDescribe)
	}

	if response.Subnets == nil {
		return managed.ExternalUpdate{}, errors.New(errUpdate)
	}

	subnet := response.Subnets[0]

	if !v1beta1.CompareTags(cr.Spec.ForProvider.Tags, subnet.Tags) {
		if _, err := e.client.CreateTags(ctx, &awsec2.CreateTagsInput{
			Resources: []string{meta.GetExternalName(cr)},
			Tags:      v1beta1.GenerateEC2Tags(cr.Spec.ForProvider.Tags),
		}); err != nil {
			return managed.ExternalUpdate{}, awsclient.Wrap(err, errCreateTags)
		}
	}

	if aws.ToBool(subnet.MapPublicIpOnLaunch) != aws.ToBool(cr.Spec.ForProvider.MapPublicIPOnLaunch) {
		_, err = e.client.ModifySubnetAttribute(ctx, &awsec2.ModifySubnetAttributeInput{
			MapPublicIpOnLaunch: &awsec2types.AttributeBooleanValue{
				Value: cr.Spec.ForProvider.MapPublicIPOnLaunch,
			},
			SubnetId: aws.String(meta.GetExternalName(cr)),
		})
		if err != nil {
			return managed.ExternalUpdate{}, awsclient.Wrap(err, errUpdate)
		}
	}

	if aws.ToBool(subnet.AssignIpv6AddressOnCreation) != aws.ToBool(cr.Spec.ForProvider.AssignIPv6AddressOnCreation) {
		_, err = e.client.ModifySubnetAttribute(ctx, &awsec2.ModifySubnetAttributeInput{
			AssignIpv6AddressOnCreation: &awsec2types.AttributeBooleanValue{
				Value: cr.Spec.ForProvider.AssignIPv6AddressOnCreation,
			},
			SubnetId: aws.String(meta.GetExternalName(cr)),
		})
	}

	return managed.ExternalUpdate{}, awsclient.Wrap(err, errUpdate)
}

func (e *external) Delete(ctx context.Context, mgd resource.Managed) error {
	cr, ok := mgd.(*v1beta1.Subnet)
	if !ok {
		return errors.New(errUnexpectedObject)
	}

	cr.Status.SetConditions(xpv1.Deleting())

	_, err := e.client.DeleteSubnet(ctx, &awsec2.DeleteSubnetInput{
		SubnetId: aws.String(meta.GetExternalName(cr)),
	})

	return awsclient.Wrap(resource.Ignore(ec2.IsSubnetNotFoundErr, err), errDelete)
}
