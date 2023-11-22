/*
Copyright 2023 The Crossplane Authors.

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

package grant

import (
	"context"

	svcsdk "github.com/aws/aws-sdk-go/service/kms"
	svcsdkapi "github.com/aws/aws-sdk-go/service/kms/kmsiface"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/kms/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/apis/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
	custommanaged "github.com/crossplane-contrib/provider-aws/pkg/utils/reconciler/managed"
)

const (
	errRevokeGrant = "cannot revoke grant"
)

// SetupGrant adds a controller that reconciles Grant.
func SetupGrant(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(svcapitypes.GrantGroupKind)
	opts := []option{
		func(e *external) {
			h := &hooks{client: e.client}
			e.preObserve = preObserve
			e.postObserve = postObserve
			e.preCreate = preCreate
			e.postCreate = postCreate
			e.delete = h.delete
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
		managed.WithInitializers(),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		managed.WithConnectionPublishers(cps...),
	}

	if o.Features.Enabled(features.EnableAlphaManagementPolicies) {
		reconcilerOpts = append(reconcilerOpts, managed.WithManagementPolicies())
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(svcapitypes.GrantGroupVersionKind),
		reconcilerOpts...)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&svcapitypes.Grant{}).
		Complete(r)
}

type hooks struct {
	client svcsdkapi.KMSAPI
}

func preObserve(_ context.Context, cr *svcapitypes.Grant, obj *svcsdk.ListGrantsInput) error {
	obj.GrantId = pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))
	obj.KeyId = cr.Spec.ForProvider.KeyID
	return nil
}

func postObserve(_ context.Context, cr *svcapitypes.Grant, obj *svcsdk.ListGrantsResponse, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return obs, err
	}
	cr.SetConditions(xpv1.Available())
	return obs, nil
}

func preCreate(ctx context.Context, cr *svcapitypes.Grant, obj *svcsdk.CreateGrantInput) error {
	obj.KeyId = cr.Spec.ForProvider.KeyID
	return nil
}

func postCreate(_ context.Context, cr *svcapitypes.Grant, obj *svcsdk.CreateGrantOutput, creation managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
	if err != nil {
		return creation, err
	}
	meta.SetExternalName(cr, pointer.StringValue(obj.GrantId))
	return managed.ExternalCreation{}, nil
}

// NOTE: KMS Grants do not support updates.

func (h *hooks) delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*svcapitypes.Grant)
	if !ok {
		return errors.New(errUnexpectedObject)
	}
	cr.SetConditions(xpv1.Deleting())

	_, err := h.client.RevokeGrantWithContext(ctx, &svcsdk.RevokeGrantInput{
		GrantId: pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr)),
		KeyId:   cr.Spec.ForProvider.KeyID,
	})
	return errors.Wrap(err, errRevokeGrant)
}
