/*
Copyright 2020 The Crossplane Authors.
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

package resolverruleassociation

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsr53r "github.com/aws/aws-sdk-go-v2/service/route53resolver"
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

	"github.com/crossplane/provider-aws/apis/route53resolver/v1alpha1"
	awsclient "github.com/crossplane/provider-aws/pkg/clients"
	r53r "github.com/crossplane/provider-aws/pkg/clients/route53resolver"
)

const (
	errUnexpectedObject = "The managed resource is not an Resolver Rule Association resource"

	errDescribe     = "failed to describe Resolver Rule Association with id"
	errCreate       = "failed to create the  Resolver Rule Association resource"
	errDelete       = "failed to delete the Resolver Rule Association resource"
	errStatusUpdate = "cannot update status of Resolver Rule Association custom resource"
)

// SetupResolverRuleAssociation adds a controller that reconciles ResolverRuleAssociation
func SetupResolverRuleAssociation(mgr ctrl.Manager, l logging.Logger, rl workqueue.RateLimiter) error {
	name := managed.ControllerName(v1alpha1.ResolverRuleAssociationKind)
	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(controller.Options{
			RateLimiter: ratelimiter.NewDefaultManagedRateLimiter(rl),
		}).
		For(&v1alpha1.ResolverRuleAssociation{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1alpha1.ResolverRuleAssociationGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient()}),
			managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
			managed.WithConnectionPublishers(),
			managed.WithInitializers(),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

type connector struct {
	kube client.Client
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1alpha1.ResolverRuleAssociation)
	if !ok {
		return nil, errors.New(errUnexpectedObject)
	}
	cfg, err := awsclient.GetConfig(ctx, c.kube, mg, cr.Spec.ForProvider.Region)
	if err != nil {
		return nil, err
	}
	return &external{client: awsr53r.New(*cfg), kube: c.kube}, nil
}

type external struct {
	kube   client.Client
	client r53r.ResolverRuleAssociationClient
}

func (e *external) Observe(ctx context.Context, mgd resource.Managed) (managed.ExternalObservation, error) { // nolint:gocyclo
	cr, ok := mgd.(*v1alpha1.ResolverRuleAssociation)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errUnexpectedObject)
	}

	if meta.GetExternalName(cr) == "" {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}

	response, err := e.client.GetResolverRuleAssociationRequest(&awsr53r.GetResolverRuleAssociationInput{
		ResolverRuleAssociationId: aws.String(meta.GetExternalName(cr))}).Send(ctx)
	if err != nil {
		return managed.ExternalObservation{}, awsclient.Wrap(err, errDescribe)
	}

	cr.SetConditions(xpv1.Available())

	cr.Status.AtProvider = r53r.GenerateRoute53ResolverObservation(*response.ResolverRuleAssociation)

	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: true,
	}, nil
}

func (e *external) Create(ctx context.Context, mgd resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mgd.(*v1alpha1.ResolverRuleAssociation)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errUnexpectedObject)
	}

	result, err := e.client.AssociateResolverRuleRequest(&awsr53r.AssociateResolverRuleInput{
		ResolverRuleId: cr.Spec.ForProvider.ResolverRuleID,
		VPCId:          cr.Spec.ForProvider.VPCID,
	}).Send(ctx)

	if err != nil {
		return managed.ExternalCreation{}, awsclient.Wrap(err, errCreate)
	}
	meta.SetExternalName(cr, aws.StringValue(result.ResolverRuleAssociation.Id))

	return managed.ExternalCreation{ExternalNameAssigned: true}, nil
}

func (e *external) Update(ctx context.Context, mgd resource.Managed) (managed.ExternalUpdate, error) {
	return managed.ExternalUpdate{}, nil
}

func (e *external) Delete(ctx context.Context, mgd resource.Managed) error {
	cr, ok := mgd.(*v1alpha1.ResolverRuleAssociation)
	if !ok {
		return errors.New(errUnexpectedObject)
	}

	_, err := e.client.DisassociateResolverRuleRequest(&awsr53r.DisassociateResolverRuleInput{
		ResolverRuleId: cr.Spec.ForProvider.ResolverRuleID,
		VPCId:          cr.Spec.ForProvider.VPCID,
	}).Send(ctx)

	return awsclient.Wrap(resource.Ignore(r53r.IsResolverRuleAssociationNotFoundErr, err), errDelete)
}
