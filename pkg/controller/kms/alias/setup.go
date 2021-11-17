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
	"time"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	svcsdk "github.com/aws/aws-sdk-go/service/kms"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/ratelimiter"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	svcapitypes "github.com/crossplane/provider-aws/apis/kms/v1alpha1"
	awsclients "github.com/crossplane/provider-aws/pkg/clients"
)

// SetupAlias adds a controller that reconciles Alias.
func SetupAlias(mgr ctrl.Manager, l logging.Logger, limiter workqueue.RateLimiter, poll time.Duration) error {
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
	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(controller.Options{
			RateLimiter: ratelimiter.NewController(limiter),
		}).
		For(&svcapitypes.Alias{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(svcapitypes.AliasGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), opts: opts}),
			managed.WithPollInterval(poll),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

func filterList(cr *svcapitypes.Alias, list *svcsdk.ListAliasesOutput) *svcsdk.ListAliasesOutput {
	for i := range list.Aliases {
		if awsclients.StringValue(list.Aliases[i].AliasName) == "alias/"+meta.GetExternalName(cr) {
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
	obj.AliasName = awsclients.String("alias/" + meta.GetExternalName(cr))
	obj.TargetKeyId = cr.Spec.ForProvider.TargetKeyID
	return nil
}

func preUpdate(_ context.Context, cr *svcapitypes.Alias, obj *svcsdk.UpdateAliasInput) error {
	obj.AliasName = awsclients.String("alias/" + meta.GetExternalName(cr))
	obj.TargetKeyId = cr.Spec.ForProvider.TargetKeyID
	return nil
}

func preDelete(_ context.Context, cr *svcapitypes.Alias, obj *svcsdk.DeleteAliasInput) (bool, error) {
	obj.AliasName = awsclients.String("alias/" + meta.GetExternalName(cr))
	return false, nil
}

func postObserve(_ context.Context, cr *svcapitypes.Alias, _ *svcsdk.ListAliasesOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}
	cr.SetConditions(xpv1.Available())
	return obs, nil
}
