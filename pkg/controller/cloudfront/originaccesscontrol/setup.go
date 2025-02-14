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

package originaccesscontrol

import (
	"context"
	"slices"

	svcsdk "github.com/aws/aws-sdk-go/service/cloudfront"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/errors"
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

var (
	OriginAccessControlOriginTypes     = []string{"s3", "mediastore", "lambda", "mediapackagev2"}
	OriginAccessControlSigningBehavior = []string{"never", "no-override", "always"}
	OriginAccessControlSigningProtocol = []string{"sigv4"}
)

func SetupOriginAccessControl(mgr ctrl.Manager, o controller.Options) error {
	_ = custommanaged.NewRetryingCriticalAnnotationUpdater
	name := managed.ControllerName(svcapitypes.OriginAccessControlKind)

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), v1alpha1.StoreConfigGroupVersionKind))
	}

	reconcilerOpts := []managed.ReconcilerOption{
		managed.WithCriticalAnnotationUpdater(custommanaged.NewRetryingCriticalAnnotationUpdater(mgr.GetClient())),
		managed.WithTypedExternalConnector(&connector{
			kube: mgr.GetClient(),
			opts: []option{
				func(e *external) {
					e.preCreate = preCreate
					e.postCreate = postCreate
					e.preObserve = preObserve
					e.postObserve = postObserve
					e.isUpToDate = isUpToDate
					e.preUpdate = preUpdate
					e.lateInitialize = lateInitialize
					e.postUpdate = postUpdate
					e.preDelete = preDelete
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
		resource.ManagedKind(svcapitypes.OriginAccessControlGroupVersionKind),
		reconcilerOpts...)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&svcapitypes.OriginAccessControl{}).
		Complete(r)
}

func validateOriginAccessControl(cr *svcapitypes.OriginAccessControl) error {
	if !slices.Contains(OriginAccessControlOriginTypes, pointer.StringValue(cr.Spec.ForProvider.OriginAccessControlConfig.OriginAccessControlOriginType)) {
		return errors.New("originAccessControlOriginType invalid")
	}

	if !slices.Contains(OriginAccessControlSigningBehavior, pointer.StringValue(cr.Spec.ForProvider.OriginAccessControlConfig.SigningBehavior)) {
		return errors.New("signingBehavior invalid")
	}

	if !slices.Contains(OriginAccessControlSigningProtocol, pointer.StringValue(cr.Spec.ForProvider.OriginAccessControlConfig.SigningProtocol)) {
		return errors.New("signingProtocol invalid")
	}

	if len(pointer.StringValue(cr.Spec.ForProvider.OriginAccessControlConfig.Name)) > 64 {
		return errors.New("name is more than 64 characters")
	}

	return nil
}

func preCreate(_ context.Context, cr *svcapitypes.OriginAccessControl, coaci *svcsdk.CreateOriginAccessControlInput) error {
	return validateOriginAccessControl(cr)
}

func postCreate(_ context.Context, cr *svcapitypes.OriginAccessControl, coaco *svcsdk.CreateOriginAccessControlOutput, ec managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
	if err != nil {
		return managed.ExternalCreation{}, err
	}

	meta.SetExternalName(cr, pointer.StringValue(coaco.OriginAccessControl.Id))
	return ec, nil
}

func preObserve(_ context.Context, cr *svcapitypes.OriginAccessControl, goaci *svcsdk.GetOriginAccessControlInput) error {
	goaci.Id = pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))
	return nil
}

func postObserve(_ context.Context, cr *svcapitypes.OriginAccessControl, goaco *svcsdk.GetOriginAccessControlOutput, eo managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}

	cr.SetConditions(xpv1.Unavailable())
	if pointer.StringValue(goaco.OriginAccessControl.Id) != "" {
		cr.SetConditions(xpv1.Available())
	}
	return eo, nil
}

func isUpToDate(_ context.Context, cr *svcapitypes.OriginAccessControl, goaco *svcsdk.GetOriginAccessControlOutput) (bool, string, error) {
	return cloudfront.IsUpToDate(goaco.OriginAccessControl.OriginAccessControlConfig, cr.Spec.ForProvider.OriginAccessControlConfig)
}

func preUpdate(_ context.Context, cr *svcapitypes.OriginAccessControl, uoaci *svcsdk.UpdateOriginAccessControlInput) error {
	uoaci.Id = pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))
	uoaci.SetIfMatch(pointer.StringValue(cr.Status.AtProvider.ETag))
	return validateOriginAccessControl(cr)
}

func postUpdate(_ context.Context, cr *svcapitypes.OriginAccessControl, goaco *svcsdk.UpdateOriginAccessControlOutput, upd managed.ExternalUpdate, err error) (managed.ExternalUpdate, error) {
	if err != nil {
		return managed.ExternalUpdate{}, err
	}
	// We need etag of update operation for the next operations.
	cr.Status.AtProvider.ETag = goaco.ETag
	return upd, nil
}

func lateInitialize(in *svcapitypes.OriginAccessControlParameters, goaco *svcsdk.GetOriginAccessControlOutput) error {
	_, err := cloudfront.LateInitializeFromResponse("",
		in.OriginAccessControlConfig, goaco.OriginAccessControl.OriginAccessControlConfig)
	return err
}

func preDelete(_ context.Context, cp *svcapitypes.OriginAccessControl, doaci *svcsdk.DeleteOriginAccessControlInput) (bool, error) {
	doaci.Id = pointer.ToOrNilIfZeroValue(meta.GetExternalName(cp))
	doaci.SetIfMatch(pointer.StringValue(cp.Status.AtProvider.ETag))
	return false, nil
}
