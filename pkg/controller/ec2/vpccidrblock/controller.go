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
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsec2 "github.com/aws/aws-sdk-go-v2/service/ec2"
	awsec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	awsgo "github.com/aws/aws-sdk-go/aws"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
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
	errUnexpectedObject = "The managed resource is not an VPCCIDRBlock resource"
	errDescribe         = "failed to describe VPC with id"
	errMultipleItems    = "retrieved multiple VPCs for the given vpcId"
	errAssociate        = "failed to associate the VPCCIDRBlock resource"
	errDisassociate     = "failed to disassociate the VPCCIDRBlock resource"
)

// SetupVPCCIDRBlock adds a controller that reconciles VPCCIDRBlocks.
func SetupVPCCIDRBlock(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1beta1.VPCCIDRBlockGroupKind)

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), v1alpha1.StoreConfigGroupVersionKind))
	}

	reconcilerOpts := []managed.ReconcilerOption{
		managed.WithCriticalAnnotationUpdater(custommanaged.NewRetryingCriticalAnnotationUpdater(mgr.GetClient())),
		managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), newClientFn: ec2.NewVPCCIDRBlockClient}),
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
		resource.ManagedKind(v1beta1.VPCCIDRBlockGroupVersionKind),
		reconcilerOpts...)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&v1beta1.VPCCIDRBlock{}).
		Complete(r)
}

type connector struct {
	kube        client.Client
	newClientFn func(config aws.Config) ec2.VPCCIDRBlockClient
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1beta1.VPCCIDRBlock)
	if !ok {
		return nil, errors.New(errUnexpectedObject)
	}
	cfg, err := connectaws.GetConfig(ctx, c.kube, mg, cr.Spec.ForProvider.Region)
	if err != nil {
		return nil, err
	}
	return &external{client: c.newClientFn(*cfg), kube: c.kube}, nil
}

type external struct {
	kube   client.Client
	client ec2.VPCCIDRBlockClient
}

func (e *external) Observe(ctx context.Context, mgd resource.Managed) (managed.ExternalObservation, error) { //nolint:gocyclo
	cr, ok := mgd.(*v1beta1.VPCCIDRBlock)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errUnexpectedObject)
	}

	if meta.GetExternalName(cr) == "" {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}

	response, err := e.client.DescribeVpcs(ctx, &awsec2.DescribeVpcsInput{
		VpcIds: []string{aws.ToString(cr.Spec.ForProvider.VPCID)},
	})

	if err != nil {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, errorutils.Wrap(resource.Ignore(ec2.IsVPCNotFoundErr, err), errDescribe)
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
	case awsec2types.VpcCidrBlockStateCodeAssociated:
		cr.SetConditions(xpv1.Available())
	case awsec2types.VpcCidrBlockStateCodeAssociating:
		cr.SetConditions(xpv1.Creating())
	case awsec2types.VpcCidrBlockStateCodeDisassociated:
		if meta.WasDeleted(mgd) {
			return managed.ExternalObservation{
				ResourceExists:   false,
				ResourceUpToDate: false,
			}, nil
		}
		cr.Status.SetConditions(xpv1.Deleting())
	case awsec2types.VpcCidrBlockStateCodeDisassociating:
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
	cr, ok := mgd.(*v1beta1.VPCCIDRBlock)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errUnexpectedObject)
	}

	result, err := e.client.AssociateVpcCidrBlock(ctx, &awsec2.AssociateVpcCidrBlockInput{
		AmazonProvidedIpv6CidrBlock:     cr.Spec.ForProvider.AmazonProvidedIPv6CIDRBlock,
		CidrBlock:                       cr.Spec.ForProvider.CIDRBlock,
		Ipv6CidrBlock:                   cr.Spec.ForProvider.IPv6CIDRBlock,
		Ipv6CidrBlockNetworkBorderGroup: cr.Spec.ForProvider.IPv6CIDRBlockNetworkBorderGroup,
		Ipv6Pool:                        cr.Spec.ForProvider.IPv6Pool,
		VpcId:                           cr.Spec.ForProvider.VPCID,
	})
	if err != nil {
		return managed.ExternalCreation{}, errorutils.Wrap(err, errAssociate)
	}

	if result != nil {
		if result.CidrBlockAssociation != nil {
			meta.SetExternalName(cr, awsgo.StringValue(result.CidrBlockAssociation.AssociationId))
		}
		if result.Ipv6CidrBlockAssociation != nil {
			meta.SetExternalName(cr, awsgo.StringValue(result.Ipv6CidrBlockAssociation.AssociationId))
		}
	}

	return managed.ExternalCreation{}, nil
}

func (e *external) Update(_ context.Context, _ resource.Managed) (managed.ExternalUpdate, error) {
	return managed.ExternalUpdate{}, nil
}

func (e *external) Delete(ctx context.Context, mgd resource.Managed) error {
	cr, ok := mgd.(*v1beta1.VPCCIDRBlock)
	if !ok {
		return errors.New(errUnexpectedObject)
	}
	if ec2.IsVpcCidrDeleting(cr.Status.AtProvider) {
		return nil
	}

	_, err := e.client.DisassociateVpcCidrBlock(ctx, &awsec2.DisassociateVpcCidrBlockInput{
		AssociationId: aws.String(meta.GetExternalName(cr)),
	})

	return errorutils.Wrap(resource.Ignore(ec2.IsCIDRNotFound, err), errDisassociate)
}
