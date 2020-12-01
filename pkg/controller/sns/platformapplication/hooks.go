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
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	svcapitypes "github.com/crossplane/provider-aws/apis/sns/v1alpha1"
)

// SetupPlatformApplication adds a controller that reconciles PlatformApplication.
func SetupPlatformApplication(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(svcapitypes.PlatformApplicationGroupKind)
	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&svcapitypes.PlatformApplication{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(svcapitypes.PlatformApplicationGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient()}),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

func (*external) preObserve(context.Context, *svcapitypes.PlatformApplication) error {
	return nil
}
func (*external) postObserve(_ context.Context, _ *svcapitypes.PlatformApplication, _ *svcsdk.GetPlatformApplicationAttributesOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	return obs, err
}

func (*external) preCreate(context.Context, *svcapitypes.PlatformApplication) error {
	return nil
}

func (*external) postCreate(_ context.Context, _ *svcapitypes.PlatformApplication, _ *svcsdk.CreatePlatformApplicationOutput, cre managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
	return cre, err
}

func (*external) preUpdate(context.Context, *svcapitypes.PlatformApplication) error {
	return nil
}

func (*external) postUpdate(_ context.Context, _ *svcapitypes.PlatformApplication, upd managed.ExternalUpdate, err error) (managed.ExternalUpdate, error) {
	return upd, err
}
func lateInitialize(*svcapitypes.PlatformApplicationParameters, *svcsdk.GetPlatformApplicationAttributesOutput) error {
	return nil
}

func isUpToDate(*svcapitypes.PlatformApplication, *svcsdk.GetPlatformApplicationAttributesOutput) bool {
	return true
}

func preGenerateGetPlatformApplicationAttributesInput(_ *svcapitypes.PlatformApplication, obj *svcsdk.GetPlatformApplicationAttributesInput) *svcsdk.GetPlatformApplicationAttributesInput {
	return obj
}

func postGenerateGetPlatformApplicationAttributesInput(_ *svcapitypes.PlatformApplication, obj *svcsdk.GetPlatformApplicationAttributesInput) *svcsdk.GetPlatformApplicationAttributesInput {
	return obj
}

func preGenerateCreatePlatformApplicationInput(_ *svcapitypes.PlatformApplication, obj *svcsdk.CreatePlatformApplicationInput) *svcsdk.CreatePlatformApplicationInput {
	return obj
}

func postGenerateCreatePlatformApplicationInput(_ *svcapitypes.PlatformApplication, obj *svcsdk.CreatePlatformApplicationInput) *svcsdk.CreatePlatformApplicationInput {
	return obj
}
func preGenerateDeletePlatformApplicationInput(_ *svcapitypes.PlatformApplication, obj *svcsdk.DeletePlatformApplicationInput) *svcsdk.DeletePlatformApplicationInput {
	return obj
}

func postGenerateDeletePlatformApplicationInput(_ *svcapitypes.PlatformApplication, obj *svcsdk.DeletePlatformApplicationInput) *svcsdk.DeletePlatformApplicationInput {
	return obj
}
