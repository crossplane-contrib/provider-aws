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

package resolverruleassociation

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane-contrib/provider-aws/apis/route53resolver/manualv1alpha1"
	"github.com/crossplane-contrib/provider-aws/apis/v1alpha1"
	resolverruleassociation "github.com/crossplane-contrib/provider-aws/pkg/clients/resolverruleassociation"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
	connectaws "github.com/crossplane-contrib/provider-aws/pkg/utils/connect/aws"
	errorutils "github.com/crossplane-contrib/provider-aws/pkg/utils/errors"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
	custommanaged "github.com/crossplane-contrib/provider-aws/pkg/utils/reconciler/managed"
)

const (
	errUnexpectedObject = "The managed resource is not an AssociatedResolverRule resource"
	errCreate           = "failed to create the AssociatedResolverRule"
	errDelete           = "failed to delete the AssociatedResolverRule"
	errGet              = "failed to get the AssociatedResolverRule"
)

// SetupResolverRuleAssociation adds a controller that reconciles ResolverRuleAssociation
func SetupResolverRuleAssociation(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(manualv1alpha1.ResolverRuleAssociationGroupKind)

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), v1alpha1.StoreConfigGroupVersionKind))
	}

	reconcilerOpts := []managed.ReconcilerOption{
		managed.WithCriticalAnnotationUpdater(custommanaged.NewRetryingCriticalAnnotationUpdater(mgr.GetClient())),
		managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), newRoute53ResolverClientFn: resolverruleassociation.NewRoute53ResolverClient}),
		managed.WithInitializers(),
		managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
		managed.WithPollInterval(o.PollInterval),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		managed.WithConnectionPublishers(cps...),
	}

	if o.Features.Enabled(features.EnableAlphaManagementPolicies) {
		reconcilerOpts = append(reconcilerOpts, managed.WithManagementPolicies())
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(manualv1alpha1.ResolverRuleAssociationGroupVersionKind),
		reconcilerOpts...)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&manualv1alpha1.ResolverRuleAssociation{}).
		Complete(r)
}

type connector struct {
	kube                       client.Client
	newRoute53ResolverClientFn func(config aws.Config) resolverruleassociation.Client
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*manualv1alpha1.ResolverRuleAssociation)
	if !ok {
		return nil, errors.New(errUnexpectedObject)
	}
	cfg, err := connectaws.GetConfig(ctx, c.kube, mg, cr.Spec.ForProvider.Region)
	if err != nil {
		return nil, err
	}
	return &external{client: c.newRoute53ResolverClientFn(*cfg), kube: c.kube}, nil
}

type external struct {
	client resolverruleassociation.Client
	kube   client.Client
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*manualv1alpha1.ResolverRuleAssociation)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errUnexpectedObject)
	}

	if meta.GetExternalName(cr) == "" {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}

	res, err := e.client.GetResolverRuleAssociation(ctx, resolverruleassociation.GenerateGetAssociateResolverRuleAssociationInput(pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))))
	if err != nil {
		return managed.ExternalObservation{}, errorutils.Wrap(resource.Ignore(resolverruleassociation.IsNotFound, err), errGet)
	}

	switch res.ResolverRuleAssociation.Status {
	case manualv1alpha1.ResolverRuleAssociationStatusComplete:
		cr.Status.SetConditions(xpv1.Available())
	case manualv1alpha1.ResolverRuleAssociationStatusCreating:
		cr.Status.SetConditions(xpv1.Creating())
	case manualv1alpha1.ResolverRuleAssociationStatusDeleting:
		cr.Status.SetConditions(xpv1.Deleting())
	case manualv1alpha1.ResolverRuleAssociationStatusFailed, manualv1alpha1.ResolverRuleAssociationStatusOverridden:
		cr.Status.SetConditions(xpv1.Unavailable())
	default:
		cr.Status.SetConditions(xpv1.Unavailable())
	}

	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: true,
	}, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*manualv1alpha1.ResolverRuleAssociation)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errUnexpectedObject)
	}

	rsp, err := e.client.AssociateResolverRule(ctx, resolverruleassociation.GenerateCreateAssociateResolverRuleInput(cr))
	if err != nil {
		return managed.ExternalCreation{}, errorutils.Wrap(err, errCreate)
	}

	meta.SetExternalName(cr, pointer.StringValue(rsp.ResolverRuleAssociation.Id))
	return managed.ExternalCreation{}, nil
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*manualv1alpha1.ResolverRuleAssociation)
	if !ok {
		return errors.New(errUnexpectedObject)
	}

	_, err := e.client.DisassociateResolverRule(ctx, resolverruleassociation.GenerateDeleteAssociateResolverRuleInput(cr))

	return errorutils.Wrap(resource.Ignore(resolverruleassociation.IsNotFound, err), errDelete)
}

func (e *external) Update(_ context.Context, _ resource.Managed) (managed.ExternalUpdate, error) {
	return managed.ExternalUpdate{}, nil
}
