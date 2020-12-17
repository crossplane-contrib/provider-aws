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

package userprofile

import (
	"context"

	svcsdk "github.com/aws/aws-sdk-go/service/sagemaker"
	ctrl "sigs.k8s.io/controller-runtime"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	svcapitypes "github.com/crossplane/provider-aws/apis/sagemaker/v1alpha1"
)

// SetupUserProfile adds a controller that reconciles UserProfile.
func SetupUserProfile(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(svcapitypes.UserProfileGroupKind)
	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&svcapitypes.UserProfile{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(svcapitypes.UserProfileGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient()}),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

func (*external) preObserve(context.Context, *svcapitypes.UserProfile) error {
	return nil
}
func (*external) postObserve(_ context.Context, cr *svcapitypes.UserProfile, _ *svcsdk.DescribeUserProfileOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}
	cr.SetConditions(xpv1.Available())
	return obs, err
}

func (*external) preCreate(context.Context, *svcapitypes.UserProfile) error {
	return nil
}

func (*external) postCreate(_ context.Context, _ *svcapitypes.UserProfile, _ *svcsdk.CreateUserProfileOutput, cre managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
	return cre, err
}

func (*external) preUpdate(context.Context, *svcapitypes.UserProfile) error {
	return nil
}

func (*external) postUpdate(_ context.Context, _ *svcapitypes.UserProfile, upd managed.ExternalUpdate, err error) (managed.ExternalUpdate, error) {
	return upd, err
}
func lateInitialize(*svcapitypes.UserProfileParameters, *svcsdk.DescribeUserProfileOutput) error {
	return nil
}

func isUpToDate(*svcapitypes.UserProfile, *svcsdk.DescribeUserProfileOutput) bool {
	return true
}

func preGenerateDescribeUserProfileInput(_ *svcapitypes.UserProfile, obj *svcsdk.DescribeUserProfileInput) *svcsdk.DescribeUserProfileInput {
	return obj
}

func postGenerateDescribeUserProfileInput(_ *svcapitypes.UserProfile, obj *svcsdk.DescribeUserProfileInput) *svcsdk.DescribeUserProfileInput {
	return obj
}

func preGenerateCreateUserProfileInput(_ *svcapitypes.UserProfile, obj *svcsdk.CreateUserProfileInput) *svcsdk.CreateUserProfileInput {
	return obj
}

func postGenerateCreateUserProfileInput(_ *svcapitypes.UserProfile, obj *svcsdk.CreateUserProfileInput) *svcsdk.CreateUserProfileInput {
	return obj
}
func preGenerateDeleteUserProfileInput(_ *svcapitypes.UserProfile, obj *svcsdk.DeleteUserProfileInput) *svcsdk.DeleteUserProfileInput {
	return obj
}

func postGenerateDeleteUserProfileInput(_ *svcapitypes.UserProfile, obj *svcsdk.DeleteUserProfileInput) *svcsdk.DeleteUserProfileInput {
	return obj
}
