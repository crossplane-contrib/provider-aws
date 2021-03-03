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

package platformapplication

import (
	"context"

	svcsdk "github.com/aws/aws-sdk-go/service/sns"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/ratelimiter"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	svcapitypes "github.com/crossplane/provider-aws/apis/sns/v1alpha1"
	aws "github.com/crossplane/provider-aws/pkg/clients"
)

// SetupPlatformApplication adds a controller that reconciles PlatformApplication.
func SetupPlatformApplication(mgr ctrl.Manager, l logging.Logger, rl workqueue.RateLimiter) error {
	opts := []option{
		func(e *external) {
			e.preObserve = preObserve
			e.postObserve = postObserve
			e.postCreate = postCreate
			e.preDelete = preDelete
		},
	}
	name := managed.ControllerName(svcapitypes.PlatformApplicationGroupKind)
	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(controller.Options{
			RateLimiter: ratelimiter.NewDefaultManagedRateLimiter(rl),
		}).
		For(&svcapitypes.PlatformApplication{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(svcapitypes.PlatformApplicationGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), opts: opts}),
			managed.WithInitializers(managed.NewDefaultProviderConfig(mgr.GetClient())),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

func preObserve(_ context.Context, cr *svcapitypes.PlatformApplication, obj *svcsdk.GetPlatformApplicationAttributesInput) error {
	obj.PlatformApplicationArn = aws.String(meta.GetExternalName(cr))
	return nil
}

func postObserve(_ context.Context, cr *svcapitypes.PlatformApplication, _ *svcsdk.GetPlatformApplicationAttributesOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}
	cr.SetConditions(xpv1.Available())
	return obs, nil
}

func postCreate(_ context.Context, cr *svcapitypes.PlatformApplication, resp *svcsdk.CreatePlatformApplicationOutput, cre managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
	if err != nil {
		return managed.ExternalCreation{}, err
	}
	meta.SetExternalName(cr, aws.StringValue(resp.PlatformApplicationArn))
	cre.ExternalNameAssigned = true
	return cre, nil
}

func preDelete(_ context.Context, cr *svcapitypes.PlatformApplication, obj *svcsdk.DeletePlatformApplicationInput) (bool, error) {
	obj.PlatformApplicationArn = aws.String(meta.GetExternalName(cr))
	return false, nil
}
