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
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	ctrl "sigs.k8s.io/controller-runtime"

	svcapitypes "github.com/crossplane/provider-aws/apis/cloudfront/v1alpha1"
	awsclients "github.com/crossplane/provider-aws/pkg/clients"
)

// SetupCachePolicy adds a controller that reconciles CachePolicy.
func SetupCachePolicy(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(svcapitypes.CachePolicyGroupKind)
	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		For(&svcapitypes.CachePolicy{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(svcapitypes.CachePolicyGroupVersionKind),
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
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

func postCreate(_ context.Context, cp *svcapitypes.CachePolicy, cpo *svcsdk.CreateCachePolicyOutput,
	ec managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
	if err != nil {
		return managed.ExternalCreation{}, err
	}

	meta.SetExternalName(cp, awsclients.StringValue(cpo.CachePolicy.Id))
	return ec, nil
}

func preObserve(_ context.Context, cp *svcapitypes.CachePolicy, gpi *svcsdk.GetCachePolicyInput) error {
	gpi.Id = awsclients.String(meta.GetExternalName(cp))
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
	upi.Id = awsclients.String(meta.GetExternalName(cp))
	upi.SetIfMatch(awsclients.StringValue(cp.Status.AtProvider.ETag))
	return nil
}

func preDelete(_ context.Context, cp *svcapitypes.CachePolicy, dpi *svcsdk.DeleteCachePolicyInput) (bool, error) {
	dpi.Id = awsclients.String(meta.GetExternalName(cp))
	dpi.SetIfMatch(awsclients.StringValue(cp.Status.AtProvider.ETag))
	return false, nil
}

var mappingOptions = []LateInitOption{Replacer("ID", "Id")}

func lateInitialize(in *svcapitypes.CachePolicyParameters, gpo *svcsdk.GetCachePolicyOutput) error {
	_, err := LateInitializeFromResponse("",
		in.CachePolicyConfig, gpo.CachePolicy.CachePolicyConfig, mappingOptions...)
	return err
}

func isUpToDate(cp *svcapitypes.CachePolicy, gpo *svcsdk.GetCachePolicyOutput) (bool, error) {
	return IsUpToDate(gpo.CachePolicy.CachePolicyConfig, cp.Spec.ForProvider.CachePolicyConfig,
		mappingOptions...)
}
