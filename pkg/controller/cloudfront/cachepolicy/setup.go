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

package cachepolicy

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

// SetupCachePolicy adds a controller that reconciles CachePolicy.
func SetupCachePolicy(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(svcapitypes.CachePolicyGroupKind)

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
					e.preObserve = preObserve
					e.postObserve = postObserve
					e.postCreate = postCreate
					e.lateInitialize = lateInitialize
					e.preUpdate = preUpdate
					e.isUpToDate = isUpToDate
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
		resource.ManagedKind(svcapitypes.CachePolicyGroupVersionKind),
		reconcilerOpts...)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&svcapitypes.CachePolicy{}).
		Complete(r)
}

func postCreate(_ context.Context, cp *svcapitypes.CachePolicy, cpo *svcsdk.CreateCachePolicyOutput,
	ec managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
	if err != nil {
		return managed.ExternalCreation{}, err
	}

	meta.SetExternalName(cp, pointer.StringValue(cpo.CachePolicy.Id))
	return ec, nil
}

func preObserve(_ context.Context, cp *svcapitypes.CachePolicy, gpi *svcsdk.GetCachePolicyInput) error {
	gpi.Id = pointer.ToOrNilIfZeroValue(meta.GetExternalName(cp))
	return nil
}

func postObserve(_ context.Context, cp *svcapitypes.CachePolicy, _ *svcsdk.GetCachePolicyOutput,
	eo managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}
	cp.SetConditions(xpv1.Available())
	return eo, nil
}

func preUpdate(_ context.Context, cp *svcapitypes.CachePolicy, upi *svcsdk.UpdateCachePolicyInput) error {
	upi.Id = pointer.ToOrNilIfZeroValue(meta.GetExternalName(cp))
	upi.SetIfMatch(pointer.StringValue(cp.Status.AtProvider.ETag))
	return nil
}

func preDelete(_ context.Context, cp *svcapitypes.CachePolicy, dpi *svcsdk.DeleteCachePolicyInput) (bool, error) {
	dpi.Id = pointer.ToOrNilIfZeroValue(meta.GetExternalName(cp))
	dpi.SetIfMatch(pointer.StringValue(cp.Status.AtProvider.ETag))
	return false, nil
}

var mappingOptions = []cloudfront.LateInitOption{cloudfront.Replacer("ID", "Id")}

func lateInitialize(in *svcapitypes.CachePolicyParameters, gpo *svcsdk.GetCachePolicyOutput) error {
	_, err := cloudfront.LateInitializeFromResponse("",
		in.CachePolicyConfig, gpo.CachePolicy.CachePolicyConfig, mappingOptions...)
	return err
}

func isUpToDate(_ context.Context, cp *svcapitypes.CachePolicy, gpo *svcsdk.GetCachePolicyOutput) (bool, string, error) {
	return cloudfront.IsUpToDate(gpo.CachePolicy.CachePolicyConfig, cp.Spec.ForProvider.CachePolicyConfig,
		mappingOptions...)
}
