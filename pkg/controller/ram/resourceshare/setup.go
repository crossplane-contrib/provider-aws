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

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	ctrl "sigs.k8s.io/controller-runtime"

	svcsdk "github.com/aws/aws-sdk-go/service/ram"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	svcapitypes "github.com/crossplane/provider-aws/apis/ram/v1alpha1"
	"github.com/crossplane/provider-aws/apis/v1alpha1"
	awsclients "github.com/crossplane/provider-aws/pkg/clients"
	"github.com/crossplane/provider-aws/pkg/features"
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

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		For(&svcapitypes.ResourceShare{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(svcapitypes.ResourceShareGroupVersionKind),
			managed.WithInitializers(managed.NewDefaultProviderConfig(mgr.GetClient())),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), opts: opts}),
			managed.WithPollInterval(o.PollInterval),
			managed.WithLogger(o.Logger.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
			managed.WithConnectionPublishers(cps...)))
}

func preDelete(_ context.Context, cr *svcapitypes.ResourceShare, obj *svcsdk.DeleteResourceShareInput) (bool, error) {
	// IdempotentParameterMismatchException: com.amazonaws.carsservice.IdempotentParameterMismatchException:
	// The request has the same client token as a previous request, but the requests are not the same.
	// client token cannot exceed 64 characters.
	obj.ClientToken = awsclients.String(cr.ResourceVersion)
	obj.ResourceShareArn = awsclients.String(meta.GetExternalName(cr))
	return false, nil
}

func preObserve(_ context.Context, cr *svcapitypes.ResourceShare, obj *svcsdk.GetResourceSharesInput) error {
	obj.MaxResults = awsclients.Int64(100)
	obj.ResourceOwner = awsclients.String(svcsdk.ResourceOwnerSelf)
	return nil
}

func postObserve(_ context.Context, cr *svcapitypes.ResourceShare, obj *svcsdk.GetResourceSharesOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}

	for _, resourceShare := range obj.ResourceShares {
		if awsclients.StringValue(resourceShare.ResourceShareArn) == meta.GetExternalName(cr) {
			switch awsclients.StringValue(resourceShare.Status) {
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
	obj.ClientToken = awsclients.String(string(cr.UID))
	return nil
}

func postCreate(_ context.Context, cr *svcapitypes.ResourceShare, obj *svcsdk.CreateResourceShareOutput, _ managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
	if err != nil {
		return managed.ExternalCreation{}, err
	}

	meta.SetExternalName(cr, awsclients.StringValue(obj.ResourceShare.ResourceShareArn))
	return managed.ExternalCreation{}, nil
}
