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

package resourcerecordset

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	route53types "github.com/aws/aws-sdk-go-v2/service/route53/types"
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

	route53v1alpha1 "github.com/crossplane-contrib/provider-aws/apis/route53/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/apis/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/resourcerecordset"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
	connectaws "github.com/crossplane-contrib/provider-aws/pkg/utils/connect/aws"
	errorutils "github.com/crossplane-contrib/provider-aws/pkg/utils/errors"
	custommanaged "github.com/crossplane-contrib/provider-aws/pkg/utils/reconciler/managed"
)

const (
	errUnexpectedObject = "The managed resource is not an ResourceRecordSet resource"
	errKubeUpdate       = "failed to update the ResourceRecordSet custom resource"
	errList             = "failed to list the ResourceRecordSet resource"
	errCreate           = "failed to create the ResourceRecordSet resource"
	errUpdate           = "failed to update the ResourceRecordSet resource"
	errDelete           = "failed to delete the ResourceRecordSet resource"
	errState            = "failed to determine resource state"
)

// SetupResourceRecordSet adds a controller that reconciles ResourceRecordSets.
func SetupResourceRecordSet(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(route53v1alpha1.ResourceRecordSetGroupKind)

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), v1alpha1.StoreConfigGroupVersionKind))
	}

	reconcilerOpts := []managed.ReconcilerOption{
		managed.WithCriticalAnnotationUpdater(custommanaged.NewRetryingCriticalAnnotationUpdater(mgr.GetClient())),
		managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), newClientFn: resourcerecordset.NewClient}),
		managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
		managed.WithConnectionPublishers(),
		managed.WithPollInterval(o.PollInterval),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		managed.WithConnectionPublishers(cps...),
	}

	if o.Features.Enabled(features.EnableAlphaManagementPolicies) {
		reconcilerOpts = append(reconcilerOpts, managed.WithManagementPolicies())
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(route53v1alpha1.ResourceRecordSetGroupVersionKind),
		reconcilerOpts...)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&route53v1alpha1.ResourceRecordSet{}).
		Complete(r)
}

type connector struct {
	kube        client.Client
	newClientFn func(config aws.Config) resourcerecordset.Client
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cfg, err := connectaws.GetConfig(ctx, c.kube, mg, connectaws.GlobalRegion)
	if err != nil {
		return nil, err
	}
	return &external{client: c.newClientFn(*cfg), kube: c.kube}, nil
}

type external struct {
	kube   client.Client
	client resourcerecordset.Client
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*route53v1alpha1.ResourceRecordSet)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errUnexpectedObject)
	}

	rrs, err := resourcerecordset.GetResourceRecordSet(ctx, meta.GetExternalName(cr), cr.Spec.ForProvider, e.client)
	if err != nil {
		// Either there is err and retry. Or Resource does not exist.
		return managed.ExternalObservation{
			ResourceExists: false,
		}, errorutils.Wrap(resource.Ignore(resourcerecordset.IsNotFound, err), errList)
	}

	current := cr.Spec.ForProvider.DeepCopy()
	resourcerecordset.LateInitialize(&cr.Spec.ForProvider, rrs)
	if !cmp.Equal(current, &cr.Spec.ForProvider) {
		if err := e.kube.Update(ctx, cr); err != nil {
			return managed.ExternalObservation{}, errors.Wrap(err, errKubeUpdate)
		}
	}

	cr.Status.SetConditions(xpv1.Available())
	upToDate, err := resourcerecordset.IsUpToDate(cr.Spec.ForProvider, *rrs)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errState)
	}

	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: upToDate,
	}, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*route53v1alpha1.ResourceRecordSet)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errUnexpectedObject)
	}

	cr.Status.SetConditions(xpv1.Creating())

	input := resourcerecordset.GenerateChangeResourceRecordSetsInput(meta.GetExternalName(cr), cr.Spec.ForProvider, route53types.ChangeActionUpsert)
	_, err := e.client.ChangeResourceRecordSets(ctx, input)

	return managed.ExternalCreation{}, errorutils.Wrap(err, errCreate)
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*route53v1alpha1.ResourceRecordSet)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errUnexpectedObject)
	}
	input := resourcerecordset.GenerateChangeResourceRecordSetsInput(meta.GetExternalName(cr), cr.Spec.ForProvider, route53types.ChangeActionUpsert)
	_, err := e.client.ChangeResourceRecordSets(ctx, input)
	return managed.ExternalUpdate{}, errorutils.Wrap(err, errUpdate)
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*route53v1alpha1.ResourceRecordSet)
	if !ok {
		return errors.New(errUnexpectedObject)
	}

	cr.Status.SetConditions(xpv1.Deleting())
	_, err := e.client.ChangeResourceRecordSets(ctx,
		resourcerecordset.GenerateChangeResourceRecordSetsInput(meta.GetExternalName(cr), cr.Spec.ForProvider, route53types.ChangeActionDelete),
	)

	// There is no way to confirm 404 (from response) when deleting a recordset
	// which isn't present using ChangeResourceRecordSetRequest.
	return errorutils.Wrap(resource.Ignore(resourcerecordset.IsNotFound, err), errDelete)
}
