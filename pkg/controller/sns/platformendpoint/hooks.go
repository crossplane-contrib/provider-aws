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

package platformendpoint

import (
	"context"

	svcsdk "github.com/aws/aws-sdk-go/service/sns"
	"github.com/crossplane/crossplane-runtime/pkg/ratelimiter"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	cpresource "github.com/crossplane/crossplane-runtime/pkg/resource"

	svcapitypes "github.com/crossplane/provider-aws/apis/sns/v1alpha1"
)

// SetupPlatformEndpoint adds a controller that reconciles PlatformEndpoint.
func SetupPlatformEndpoint(mgr ctrl.Manager, l logging.Logger, rl workqueue.RateLimiter) error {
	name := managed.ControllerName(svcapitypes.PlatformEndpointGroupKind)
	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(controller.Options{
			RateLimiter: ratelimiter.NewDefaultManagedRateLimiter(rl),
		}).
		For(&svcapitypes.PlatformEndpoint{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(svcapitypes.PlatformEndpointGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient()}),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

func (e *external) Observe(_ context.Context, _ cpresource.Managed) (managed.ExternalObservation, error) {
	return managed.ExternalObservation{}, nil
}

func (*external) preCreate(context.Context, *svcapitypes.PlatformEndpoint) error {
	return nil
}

func (*external) postCreate(_ context.Context, _ *svcapitypes.PlatformEndpoint, _ *svcsdk.CreatePlatformEndpointOutput, cre managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
	return cre, err
}

func (*external) preUpdate(context.Context, *svcapitypes.PlatformEndpoint) error {
	return nil
}

func (*external) postUpdate(_ context.Context, _ *svcapitypes.PlatformEndpoint, upd managed.ExternalUpdate, err error) (managed.ExternalUpdate, error) {
	return upd, err
}

func preGenerateCreatePlatformEndpointInput(_ *svcapitypes.PlatformEndpoint, obj *svcsdk.CreatePlatformEndpointInput) *svcsdk.CreatePlatformEndpointInput {
	return obj
}

func postGenerateCreatePlatformEndpointInput(_ *svcapitypes.PlatformEndpoint, obj *svcsdk.CreatePlatformEndpointInput) *svcsdk.CreatePlatformEndpointInput {
	return obj
}
