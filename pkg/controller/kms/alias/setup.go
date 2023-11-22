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

package alias

import (
	"context"

	svcsdk "github.com/aws/aws-sdk-go/service/kms"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	ctrl "sigs.k8s.io/controller-runtime"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/kms/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/apis/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
	custommanaged "github.com/crossplane-contrib/provider-aws/pkg/utils/reconciler/managed"
)

// SetupAlias adds a controller that reconciles Alias.
func SetupAlias(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(svcapitypes.AliasGroupKind)
	opts := []option{
		func(e *external) {
			e.preObserve = preObserve
			e.postObserve = postObserve
			e.preCreate = preCreate
			e.preUpdate = preUpdate
			e.preDelete = preDelete
			e.filterList = filterList
		},
	}

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), v1alpha1.StoreConfigGroupVersionKind))
	}

	reconcilerOpts := []managed.ReconcilerOption{
		managed.WithCriticalAnnotationUpdater(custommanaged.NewRetryingCriticalAnnotationUpdater(mgr.GetClient())),
		managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), opts: opts}),
		managed.WithPollInterval(o.PollInterval),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		managed.WithConnectionPublishers(cps...),
	}

	if o.Features.Enabled(features.EnableAlphaManagementPolicies) {
		reconcilerOpts = append(reconcilerOpts, managed.WithManagementPolicies())
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(svcapitypes.AliasGroupVersionKind),
		reconcilerOpts...)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&svcapitypes.Alias{}).
		Complete(r)
}

func filterList(cr *svcapitypes.Alias, list *svcsdk.ListAliasesOutput) *svcsdk.ListAliasesOutput {
	for i := range list.Aliases {
		if pointer.StringValue(list.Aliases[i].AliasName) == "alias/"+meta.GetExternalName(cr) {
			return &svcsdk.ListAliasesOutput{
				Aliases: []*svcsdk.AliasListEntry{
					list.Aliases[i],
				}}
		}
	}
	return &svcsdk.ListAliasesOutput{}
}

func preObserve(_ context.Context, cr *svcapitypes.Alias, obj *svcsdk.ListAliasesInput) error {
	obj.KeyId = cr.Spec.ForProvider.TargetKeyID
	return nil
}

func preCreate(_ context.Context, cr *svcapitypes.Alias, obj *svcsdk.CreateAliasInput) error {
	obj.AliasName = pointer.ToOrNilIfZeroValue("alias/" + meta.GetExternalName(cr))
	obj.TargetKeyId = cr.Spec.ForProvider.TargetKeyID
	return nil
}

func preUpdate(_ context.Context, cr *svcapitypes.Alias, obj *svcsdk.UpdateAliasInput) error {
	obj.AliasName = pointer.ToOrNilIfZeroValue("alias/" + meta.GetExternalName(cr))
	obj.TargetKeyId = cr.Spec.ForProvider.TargetKeyID
	return nil
}

func preDelete(_ context.Context, cr *svcapitypes.Alias, obj *svcsdk.DeleteAliasInput) (bool, error) {
	obj.AliasName = pointer.ToOrNilIfZeroValue("alias/" + meta.GetExternalName(cr))
	return false, nil
}

func postObserve(_ context.Context, cr *svcapitypes.Alias, _ *svcsdk.ListAliasesOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}
	cr.SetConditions(xpv1.Available())
	return obs, nil
}
