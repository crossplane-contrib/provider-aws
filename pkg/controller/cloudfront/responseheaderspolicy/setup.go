/*
Copyright 2022 The Crossplane Authors.

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

package responseheaderspolicy

import (
	"context"

	svcsdk "github.com/aws/aws-sdk-go/service/cloudfront"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	ctrl "sigs.k8s.io/controller-runtime"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/cloudfront/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/apis/v1alpha1"
	cloudfront "github.com/crossplane-contrib/provider-aws/pkg/controller/cloudfront/utils"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
	custommanaged "github.com/crossplane-contrib/provider-aws/pkg/utils/reconciler/managed"
)

// SetupResponseHeadersPolicy adds a controller that reconciles ResponseHeadersPolicy.
func SetupResponseHeadersPolicy(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(svcapitypes.ResponseHeadersPolicyGroupKind)

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), v1alpha1.StoreConfigGroupVersionKind))
	}

	reconcilerOpts := []managed.ReconcilerOption{
		managed.WithCriticalAnnotationUpdater(custommanaged.NewRetryingCriticalAnnotationUpdater(mgr.GetClient())),
		managed.WithExternalConnecter(&connector{
			kube: mgr.GetClient(),
			opts: []option{
				func(e *external) {
					e.preCreate = preCreate
					e.postCreate = postCreate
					e.lateInitialize = lateInitialize
					e.preObserve = preObserve
					e.postObserve = postObserve
					e.isUpToDate = isUpToDate
					e.preUpdate = preUpdate
					e.postUpdate = postUpdate
					d := &deleter{external: e}
					e.preDelete = d.preDelete
				},
			},
		}),
		managed.WithPollInterval(o.PollInterval),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		managed.WithConnectionPublishers(cps...),
	}

	if o.Features.Enabled(features.EnableAlphaManagementPolicies) {
		reconcilerOpts = append(reconcilerOpts, managed.WithManagementPolicies())
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(svcapitypes.ResponseHeadersPolicyGroupVersionKind),
		reconcilerOpts...)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&svcapitypes.ResponseHeadersPolicy{}).
		Complete(r)
}

func preCreate(_ context.Context, cr *svcapitypes.ResponseHeadersPolicy, crhpi *svcsdk.CreateResponseHeadersPolicyInput) error {
	crhpi.ResponseHeadersPolicyConfig.Name = pointer.ToOrNilIfZeroValue(cr.Name)

	if cr.Spec.ForProvider.ResponseHeadersPolicyConfig.CustomHeadersConfig != nil {
		crhpi.ResponseHeadersPolicyConfig.CustomHeadersConfig.Quantity =
			pointer.ToIntAsInt64(len(cr.Spec.ForProvider.ResponseHeadersPolicyConfig.CustomHeadersConfig.Items))
	}

	if cr.Spec.ForProvider.ResponseHeadersPolicyConfig.CORSConfig != nil && cr.Spec.ForProvider.ResponseHeadersPolicyConfig.CORSConfig.AccessControlAllowHeaders != nil {
		crhpi.ResponseHeadersPolicyConfig.CorsConfig.AccessControlAllowHeaders.Quantity =
			pointer.ToIntAsInt64(len(cr.Spec.ForProvider.ResponseHeadersPolicyConfig.CORSConfig.AccessControlAllowHeaders.Items))
	}
	if cr.Spec.ForProvider.ResponseHeadersPolicyConfig.CORSConfig != nil && cr.Spec.ForProvider.ResponseHeadersPolicyConfig.CORSConfig.AccessControlAllowMethods != nil {
		crhpi.ResponseHeadersPolicyConfig.CorsConfig.AccessControlAllowMethods.Quantity =
			pointer.ToIntAsInt64(len(cr.Spec.ForProvider.ResponseHeadersPolicyConfig.CORSConfig.AccessControlAllowMethods.Items))
	}

	if cr.Spec.ForProvider.ResponseHeadersPolicyConfig.CORSConfig != nil && cr.Spec.ForProvider.ResponseHeadersPolicyConfig.CORSConfig.AccessControlAllowOrigins != nil {
		crhpi.ResponseHeadersPolicyConfig.CorsConfig.AccessControlAllowOrigins.Quantity =
			pointer.ToIntAsInt64(len(cr.Spec.ForProvider.ResponseHeadersPolicyConfig.CORSConfig.AccessControlAllowOrigins.Items))
	}

	if cr.Spec.ForProvider.ResponseHeadersPolicyConfig.CORSConfig != nil && cr.Spec.ForProvider.ResponseHeadersPolicyConfig.CORSConfig.AccessControlExposeHeaders != nil {
		crhpi.ResponseHeadersPolicyConfig.CorsConfig.AccessControlExposeHeaders.Quantity =
			pointer.ToIntAsInt64(len(cr.Spec.ForProvider.ResponseHeadersPolicyConfig.CORSConfig.AccessControlExposeHeaders.Items))
	}

	return nil
}

func postCreate(_ context.Context, cr *svcapitypes.ResponseHeadersPolicy, crhpo *svcsdk.CreateResponseHeadersPolicyOutput,
	ec managed.ExternalCreation, err error) (managed.ExternalCreation, error) {

	if err != nil {
		return managed.ExternalCreation{}, err
	}

	meta.SetExternalName(cr, pointer.StringValue(crhpo.ResponseHeadersPolicy.Id))
	return ec, nil
}

func preObserve(_ context.Context, cr *svcapitypes.ResponseHeadersPolicy, grhpi *svcsdk.GetResponseHeadersPolicyInput) error {
	grhpi.Id = pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))
	return nil
}

func postObserve(_ context.Context, cr *svcapitypes.ResponseHeadersPolicy, grhpo *svcsdk.GetResponseHeadersPolicyOutput,
	eo managed.ExternalObservation, err error) (managed.ExternalObservation, error) {

	if err != nil {
		return managed.ExternalObservation{}, err
	}
	cr.SetConditions(xpv1.Available())
	return eo, nil
}

func preUpdate(_ context.Context, cr *svcapitypes.ResponseHeadersPolicy, urhpi *svcsdk.UpdateResponseHeadersPolicyInput) error {
	urhpi.Id = pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))
	urhpi.SetIfMatch(pointer.StringValue(cr.Status.AtProvider.ETag))

	if cr.Spec.ForProvider.ResponseHeadersPolicyConfig.CustomHeadersConfig != nil {
		urhpi.ResponseHeadersPolicyConfig.CustomHeadersConfig.Quantity =
			pointer.ToIntAsInt64(len(cr.Spec.ForProvider.ResponseHeadersPolicyConfig.CustomHeadersConfig.Items))
	}

	if cr.Spec.ForProvider.ResponseHeadersPolicyConfig.CORSConfig != nil && cr.Spec.ForProvider.ResponseHeadersPolicyConfig.CORSConfig.AccessControlAllowHeaders != nil {
		urhpi.ResponseHeadersPolicyConfig.CorsConfig.AccessControlAllowHeaders.Quantity =
			pointer.ToIntAsInt64(len(cr.Spec.ForProvider.ResponseHeadersPolicyConfig.CORSConfig.AccessControlAllowHeaders.Items))
	}
	if cr.Spec.ForProvider.ResponseHeadersPolicyConfig.CORSConfig != nil && cr.Spec.ForProvider.ResponseHeadersPolicyConfig.CORSConfig.AccessControlAllowMethods != nil {
		urhpi.ResponseHeadersPolicyConfig.CorsConfig.AccessControlAllowMethods.Quantity =
			pointer.ToIntAsInt64(len(cr.Spec.ForProvider.ResponseHeadersPolicyConfig.CORSConfig.AccessControlAllowMethods.Items))
	}

	if cr.Spec.ForProvider.ResponseHeadersPolicyConfig.CORSConfig != nil && cr.Spec.ForProvider.ResponseHeadersPolicyConfig.CORSConfig.AccessControlAllowOrigins != nil {
		urhpi.ResponseHeadersPolicyConfig.CorsConfig.AccessControlAllowOrigins.Quantity =
			pointer.ToIntAsInt64(len(cr.Spec.ForProvider.ResponseHeadersPolicyConfig.CORSConfig.AccessControlAllowOrigins.Items))
	}

	if cr.Spec.ForProvider.ResponseHeadersPolicyConfig.CORSConfig != nil && cr.Spec.ForProvider.ResponseHeadersPolicyConfig.CORSConfig.AccessControlExposeHeaders != nil {
		urhpi.ResponseHeadersPolicyConfig.CorsConfig.AccessControlExposeHeaders.Quantity =
			pointer.ToIntAsInt64(len(cr.Spec.ForProvider.ResponseHeadersPolicyConfig.CORSConfig.AccessControlExposeHeaders.Items))
	}
	return nil
}

func postUpdate(_ context.Context, cr *svcapitypes.ResponseHeadersPolicy, urhpo *svcsdk.UpdateResponseHeadersPolicyOutput,
	upd managed.ExternalUpdate, err error) (managed.ExternalUpdate, error) {
	if err != nil {
		return managed.ExternalUpdate{}, err
	}
	// We need etag of update operation for the next operations.
	cr.Status.AtProvider.ETag = urhpo.ETag
	return upd, nil
}

type deleter struct {
	external *external
}

func (d *deleter) preDelete(_ context.Context, cr *svcapitypes.ResponseHeadersPolicy, drhpi *svcsdk.DeleteResponseHeadersPolicyInput) (bool, error) {
	drhpi.Id = pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))
	drhpi.SetIfMatch(pointer.StringValue(cr.Status.AtProvider.ETag))
	return false, nil
}

var mappingOptions = []cloudfront.LateInitOption{cloudfront.Replacer("ID", "Id")}

func lateInitialize(in *svcapitypes.ResponseHeadersPolicyParameters, grhpo *svcsdk.GetResponseHeadersPolicyOutput) error {
	_, err := cloudfront.LateInitializeFromResponse("",
		in.ResponseHeadersPolicyConfig, grhpo.ResponseHeadersPolicy.ResponseHeadersPolicyConfig, mappingOptions...)
	return err
}

func isUpToDate(_ context.Context, rhp *svcapitypes.ResponseHeadersPolicy, grhpo *svcsdk.GetResponseHeadersPolicyOutput) (bool, string, error) {
	return cloudfront.IsUpToDate(grhpo.ResponseHeadersPolicy, rhp.Spec.ForProvider.ResponseHeadersPolicyConfig,
		mappingOptions...)
}
