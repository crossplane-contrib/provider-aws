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
	"strings"
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
			e.postCreate = postCreate
			e.postObserve = postObserve
			e.preCreate = preCreate
		},
	}
	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(controller.Options{
			RateLimiter: ratelimiter.NewDefaultManagedRateLimiter(limiter),
		}).
		For(&svcapitypes.Alias{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(svcapitypes.AliasGroupVersionKind),
			managed.WithInitializers(managed.NewDefaultProviderConfig(mgr.GetClient())),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), opts: opts}),
			managed.WithPollInterval(poll),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

func preCreate(_ context.Context, cr *svcapitypes.Alias, obj *svcsdk.CreateAliasInput) error {
	obj.TargetKeyId = cr.Spec.ForProvider.TargetKeyID
	return nil
}

func postCreate(_ context.Context, cr *svcapitypes.Alias, obj *svcsdk.CreateAliasOutput, cre managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
	if err != nil {
		return managed.ExternalCreation{}, err
	}
	// CreateAliasOutput is empty
	meta.SetExternalName(cr, *cr.Spec.ForProvider.AliasName)
	return managed.ExternalCreation{ExternalNameAssigned: true}, nil
}

func postObserve(_ context.Context, cr *svcapitypes.Alias, obj *svcsdk.ListAliasesOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}

	// get all alias
	for i := range obj.Aliases {
		if awsclients.StringValue(obj.Aliases[i].AliasName) == awsclients.StringValue(cr.Spec.ForProvider.AliasName) {
			// obj.Aliases[i].TargetKeyId is in ListAliasesOutput the KMSKey.ARN
			if strings.Contains(awsclients.StringValue(cr.Spec.ForProvider.TargetKeyID), awsclients.StringValue(obj.Aliases[i].TargetKeyId)) {
				// alias found and TargetKeyId included
				cr.SetConditions(xpv1.Available())
				return obs, nil
			}
		}
	}

	cr.SetConditions(xpv1.Unavailable())
	return managed.ExternalObservation{}, err
}
