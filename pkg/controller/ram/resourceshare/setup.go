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

package resourceshare

import (
	"context"

	svcsdk "github.com/aws/aws-sdk-go/service/ram"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	ctrl "sigs.k8s.io/controller-runtime"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/ram/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/apis/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
	custommanaged "github.com/crossplane-contrib/provider-aws/pkg/utils/reconciler/managed"
)

// SetupResourceShare adds a controller that reconciles ResourceShare.
func SetupResourceShare(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(svcapitypes.ResourceShareGroupKind)
	opts := []option{
		func(e *external) {
			e.preObserve = preObserve
			e.postObserve = postObserve
			e.preDelete = preDelete
			e.postCreate = postCreate
			e.preCreate = preCreate
		},
	}

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), v1alpha1.StoreConfigGroupVersionKind))
	}

	reconcilerOpts := []managed.ReconcilerOption{
		managed.WithInitializers(),
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
		resource.ManagedKind(svcapitypes.ResourceShareGroupVersionKind),
		reconcilerOpts...)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&svcapitypes.ResourceShare{}).
		Complete(r)
}

func preDelete(_ context.Context, cr *svcapitypes.ResourceShare, obj *svcsdk.DeleteResourceShareInput) (bool, error) {
	// IdempotentParameterMismatchException: com.amazonaws.carsservice.IdempotentParameterMismatchException:
	// The request has the same client token as a previous request, but the requests are not the same.
	// client token cannot exceed 64 characters.
	obj.ClientToken = pointer.ToOrNilIfZeroValue(cr.ResourceVersion)
	obj.ResourceShareArn = pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))
	return false, nil
}

func preObserve(_ context.Context, cr *svcapitypes.ResourceShare, obj *svcsdk.GetResourceSharesInput) error {
	obj.MaxResults = pointer.ToIntAsInt64(100)
	obj.ResourceOwner = pointer.ToOrNilIfZeroValue(svcsdk.ResourceOwnerSelf)
	return nil
}

func postObserve(_ context.Context, cr *svcapitypes.ResourceShare, obj *svcsdk.GetResourceSharesOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}

	for _, resourceShare := range obj.ResourceShares {
		if pointer.StringValue(resourceShare.ResourceShareArn) == meta.GetExternalName(cr) {
			switch pointer.StringValue(resourceShare.Status) {
			case string(svcapitypes.ResourceShareStatus_SDK_ACTIVE):
				cr.SetConditions(xpv1.Available())
			case string(svcapitypes.ResourceShareStatus_SDK_PENDING):
				cr.SetConditions(xpv1.Creating())
			case string(svcapitypes.ResourceShareStatus_SDK_FAILED):
				cr.SetConditions(xpv1.Unavailable())
			case string(svcapitypes.ResourceShareStatus_SDK_DELETING):
				cr.SetConditions(xpv1.Deleting())
			case string(svcapitypes.ResourceShareStatus_SDK_DELETED):
				// ram resourceshare is 1 hour in status deleted
				return managed.ExternalObservation{
					ResourceExists: false,
				}, nil
			}

			break
		}
	}
	return obs, nil
}

func preCreate(_ context.Context, cr *svcapitypes.ResourceShare, obj *svcsdk.CreateResourceShareInput) error {
	obj.ClientToken = pointer.ToOrNilIfZeroValue(string(cr.UID))
	return nil
}

func postCreate(_ context.Context, cr *svcapitypes.ResourceShare, obj *svcsdk.CreateResourceShareOutput, _ managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
	if err != nil {
		return managed.ExternalCreation{}, err
	}

	meta.SetExternalName(cr, pointer.StringValue(obj.ResourceShare.ResourceShareArn))
	return managed.ExternalCreation{}, nil
}
