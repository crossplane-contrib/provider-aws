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

package vpccidrblock

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsec2 "github.com/aws/aws-sdk-go-v2/service/ec2"
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

	"github.com/crossplane/provider-aws/apis/ec2/v1alpha1"
	awsclient "github.com/crossplane/provider-aws/pkg/clients"
	"github.com/crossplane/provider-aws/pkg/clients/ec2"
)

const (
	errUnexpectedObject = "The managed resource is not an VPCCIDRBlock resource"
	errDescribe         = "failed to describe VPC with id"
	errMultipleItems    = "retrieved multiple VPCs for the given vpcId"
	errAssociate        = "failed to associate the VPCCIDRBlock resource"
	errDisassociate     = "failed to disassociate the VPCCIDRBlock resource"
)

// SetupVPCCIDRBlock adds a controller that reconciles VPCCIDRBlocks.
func SetupVPCCIDRBlock(mgr ctrl.Manager, l logging.Logger, rl workqueue.RateLimiter) error {
	name := managed.ControllerName(v1alpha1.VPCCIDRBlockGroupKind)
	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(controller.Options{
			RateLimiter: ratelimiter.NewDefaultManagedRateLimiter(rl),
		}).
		For(&v1alpha1.VPCCIDRBlock{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1alpha1.VPCCIDRBlockGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), newClientFn: ec2.NewVPCCIDRBlockClient}),
			managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
			managed.WithConnectionPublishers(),
			managed.WithInitializers(managed.NewDefaultProviderConfig(mgr.GetClient())),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

type connector struct {
	kube        client.Client
	newClientFn func(config aws.Config) ec2.VPCCIDRBlockClient
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1alpha1.VPCCIDRBlock)
	if !ok {
		return nil, errors.New(errUnexpectedObject)
	}
	cfg, err := awsclient.GetConfig(ctx, c.kube, mg, cr.Spec.ForProvider.Region)
	if err != nil {
		return nil, err
	}
	return &external{client: c.newClientFn(*cfg), kube: c.kube}, nil
}

type external struct {
	kube   client.Client
	client ec2.VPCCIDRBlockClient
}

func (e *external) Observe(ctx context.Context, mgd resource.Managed) (managed.ExternalObservation, error) { // nolint:gocyclo
	cr, ok := mgd.(*v1alpha1.VPCCIDRBlock)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errUnexpectedObject)
	}

	if meta.GetExternalName(cr) == "" {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}

	response, err := e.client.DescribeVpcsRequest(&awsec2.DescribeVpcsInput{
		VpcIds: []string{aws.StringValue(cr.Spec.ForProvider.VPCID)},
	}).Send(ctx)

	if err != nil {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, awsclient.Wrap(resource.Ignore(ec2.IsVPCNotFoundErr, err), errDescribe)
	}

	// in a successful response, there should be one and only one object
	if len(response.Vpcs) != 1 {
		return managed.ExternalObservation{}, errors.New(errMultipleItems)
	}

	observed := response.Vpcs[0]

	currentStatusCode, err := ec2.FindVPCCIDRBlockStatus(meta.GetExternalName(cr), observed)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(resource.Ignore(ec2.IsCIDRNotFound, err), errDescribe)
	}

	switch currentStatusCode { //nolint:exhaustive
	case awsec2.VpcCidrBlockStateCodeAssociated:
		cr.SetConditions(xpv1.Available())
	case awsec2.VpcCidrBlockStateCodeAssociating:
		cr.SetConditions(xpv1.Creating())
	case awsec2.VpcCidrBlockStateCodeDisassociated:
		cr.Status.SetConditions(xpv1.Deleting())
	case awsec2.VpcCidrBlockStateCodeDisassociating:
		cr.Status.SetConditions(xpv1.Deleting())
	default:
		cr.SetConditions(xpv1.Unavailable())
	}

	cr.Status.AtProvider = ec2.GenerateVpcCIDRBlockObservation(meta.GetExternalName(cr), observed)

	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: true,
	}, nil
}

func (e *external) Create(ctx context.Context, mgd resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mgd.(*v1alpha1.VPCCIDRBlock)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errUnexpectedObject)
	}

	result, err := e.client.AssociateVpcCidrBlockRequest(&awsec2.AssociateVpcCidrBlockInput{
		AmazonProvidedIpv6CidrBlock:     cr.Spec.ForProvider.AmazonProvidedIPv6CIDRBlock,
		CidrBlock:                       cr.Spec.ForProvider.CIDRBlock,
		Ipv6CidrBlock:                   cr.Spec.ForProvider.IPv6CIDRBlock,
		Ipv6CidrBlockNetworkBorderGroup: cr.Spec.ForProvider.IPv6CIDRBlockNetworkBorderGroup,
		Ipv6Pool:                        cr.Spec.ForProvider.IPv6Pool,
		VpcId:                           cr.Spec.ForProvider.VPCID,
	}).Send(ctx)
	if err != nil {
		return managed.ExternalCreation{}, awsclient.Wrap(err, errAssociate)
	}

	if result != nil {
		if result.CidrBlockAssociation != nil {
			meta.SetExternalName(cr, aws.StringValue(result.CidrBlockAssociation.AssociationId))
		}
		if result.Ipv6CidrBlockAssociation != nil {
			meta.SetExternalName(cr, aws.StringValue(result.Ipv6CidrBlockAssociation.AssociationId))
		}
	}

	return managed.ExternalCreation{ExternalNameAssigned: true}, nil
}

func (e *external) Update(_ context.Context, _ resource.Managed) (managed.ExternalUpdate, error) {
	return managed.ExternalUpdate{}, nil
}

func (e *external) Delete(ctx context.Context, mgd resource.Managed) error {
	cr, ok := mgd.(*v1alpha1.VPCCIDRBlock)
	if !ok {
		return errors.New(errUnexpectedObject)
	}
	if ec2.IsVpcCidrDeleting(cr.Status.AtProvider) {
		return nil
	}

	_, err := e.client.DisassociateVpcCidrBlockRequest(&awsec2.DisassociateVpcCidrBlockInput{
		AssociationId: aws.String(meta.GetExternalName(cr)),
	}).Send(ctx)

	return awsclient.Wrap(resource.Ignore(ec2.IsCIDRNotFound, err), errDisassociate)
}
