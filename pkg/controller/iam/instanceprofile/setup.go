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

package instanceprofile

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	svcsdk "github.com/aws/aws-sdk-go/service/iam"
	svcsdkapi "github.com/aws/aws-sdk-go/service/iam/iamiface"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	ctrl "sigs.k8s.io/controller-runtime"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/iam/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/apis/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
	custommanaged "github.com/crossplane-contrib/provider-aws/pkg/utils/reconciler/managed"
)

// SetupInstanceProfile adds a controller that reconciles InstanceProfile.
func SetupInstanceProfile(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(svcapitypes.InstanceProfileGroupKind)
	opts := []option{
		func(e *external) {
			u := &updater{client: e.client}
			e.preObserve = preObserve
			e.postObserve = postObserve
			e.preCreate = preCreate
			e.postCreate = u.postCreate
			e.preDelete = u.preDelete
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
		resource.ManagedKind(svcapitypes.InstanceProfileGroupVersionKind),
		reconcilerOpts...)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		For(&svcapitypes.InstanceProfile{}).
		WithEventFilter(resource.DesiredStateChanged()).
		Complete(r)
}

func preObserve(_ context.Context, cr *svcapitypes.InstanceProfile, obj *svcsdk.GetInstanceProfileInput) error {
	obj.InstanceProfileName = aws.String(meta.GetExternalName(cr))
	return nil
}

func postObserve(_ context.Context, cr *svcapitypes.InstanceProfile, _ *svcsdk.GetInstanceProfileOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}
	cr.SetConditions(xpv1.Available())
	return obs, nil
}

func preCreate(_ context.Context, cr *svcapitypes.InstanceProfile, obj *svcsdk.CreateInstanceProfileInput) error {
	obj.InstanceProfileName = aws.String(meta.GetExternalName(cr))
	return nil
}

type updater struct {
	client svcsdkapi.IAMAPI
}

func (u *updater) postCreate(ctx context.Context, cr *svcapitypes.InstanceProfile, resp *svcsdk.CreateInstanceProfileOutput, cre managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
	if err != nil {
		return managed.ExternalCreation{}, err
	}

	input := &svcsdk.AddRoleToInstanceProfileInput{
		InstanceProfileName: aws.String(meta.GetExternalName(cr)),
		RoleName:            cr.Spec.ForProvider.Role,
	}

	_, err = u.client.AddRoleToInstanceProfileWithContext(ctx, input)
	return cre, err
}

func (u *updater) preDelete(ctx context.Context, cr *svcapitypes.InstanceProfile, obj *svcsdk.DeleteInstanceProfileInput) (bool, error) {
	obj.InstanceProfileName = aws.String(meta.GetExternalName(cr))
	input := &svcsdk.RemoveRoleFromInstanceProfileInput{
		InstanceProfileName: aws.String(meta.GetExternalName(cr)),
		RoleName:            cr.Spec.ForProvider.Role,
	}

	_, err := u.client.RemoveRoleFromInstanceProfileWithContext(ctx, input)
	if IsNotFound(err) {
		return false, nil
	}

	return false, err
}
