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
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
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
	cpresource "github.com/crossplane/crossplane-runtime/pkg/resource"

	svcapitypes "github.com/crossplane/provider-aws/apis/sns/v1alpha1"
	aws "github.com/crossplane/provider-aws/pkg/clients"
)

// NOTE(muvaf): Support in ACK for this resource is poor due to mismatch in
// method names; in Create its name is PlatformEndpoint, but in other calls it's
// only Endpoint.

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

func (e *external) Observe(ctx context.Context, mg cpresource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*svcapitypes.PlatformEndpoint)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errUnexpectedObject)
	}
	if meta.GetExternalName(cr) == "" {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}
	input := GenerateGetEndpointAttributesInput(cr)
	resp, err := e.client.GetEndpointAttributesWithContext(ctx, input)
	if err != nil {
		return managed.ExternalObservation{ResourceExists: false}, errors.Wrap(cpresource.Ignore(IsNotFound, err), errDescribe)
	}
	currentSpec := cr.Spec.ForProvider.DeepCopy()
	lateInitialize(&cr.Spec.ForProvider, resp)
	cr.SetConditions(runtimev1alpha1.Available())
	return managed.ExternalObservation{
		ResourceExists:          true,
		ResourceUpToDate:        isUpToDate(cr, resp),
		ResourceLateInitialized: !cmp.Equal(&cr.Spec.ForProvider, currentSpec),
	}, nil
}

func (e *external) Delete(ctx context.Context, mg cpresource.Managed) error {
	cr, ok := mg.(*svcapitypes.PlatformEndpoint)
	if !ok {
		return errors.New(errUnexpectedObject)
	}
	cr.SetConditions(xpv1.Deleting())
	input := &svcsdk.DeleteEndpointInput{EndpointArn: aws.String(meta.GetExternalName(cr))}
	_, err := e.client.DeleteEndpointWithContext(ctx, input)
	return errors.Wrap(cpresource.Ignore(IsNotFound, err), errDelete)
}

func lateInitialize(_ *svcapitypes.PlatformEndpointParameters, _ *svcsdk.GetEndpointAttributesOutput) {
}

func isUpToDate(_ *svcapitypes.PlatformEndpoint, _ *svcsdk.GetEndpointAttributesOutput) bool {
	return true
}

// GenerateGetEndpointAttributesInput converts parameters in PlatformEndpoint
// into GetEndpointAttributesInput to be used in the API requests.
func GenerateGetEndpointAttributesInput(cr *svcapitypes.PlatformEndpoint) *svcsdk.GetEndpointAttributesInput {
	return &svcsdk.GetEndpointAttributesInput{EndpointArn: aws.String(meta.GetExternalName(cr))}
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

func postGenerateCreatePlatformEndpointInput(cr *svcapitypes.PlatformEndpoint, obj *svcsdk.CreatePlatformEndpointInput) *svcsdk.CreatePlatformEndpointInput {
	obj.SetPlatformApplicationArn(aws.StringValue(cr.Spec.ForProvider.PlatformApplicationARN))
	return obj
}
