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

package api

import (
	"context"

	svcsdk "github.com/aws/aws-sdk-go/service/apigatewayv2"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	svcapitypes "github.com/crossplane/provider-aws/apis/apigatewayv2/v1alpha1"
	aws "github.com/crossplane/provider-aws/pkg/clients"
)

// SetupAPI adds a controller that reconciles API.
func SetupAPI(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(svcapitypes.APIGroupKind)
	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&svcapitypes.API{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(svcapitypes.APIGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient()}),
			managed.WithInitializers(managed.NewDefaultProviderConfig(mgr.GetClient())),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

func (*external) preObserve(context.Context, *svcapitypes.API) error {
	return nil
}
func (*external) postObserve(_ context.Context, cr *svcapitypes.API, _ *svcsdk.GetApiOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	cr.SetConditions(v1alpha1.Available())
	return obs, err
}

func (*external) preCreate(context.Context, *svcapitypes.API) error {
	return nil
}

func (*external) postCreate(_ context.Context, cr *svcapitypes.API, resp *svcsdk.CreateApiOutput, cre managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
	if err != nil {
		return managed.ExternalCreation{}, err
	}
	meta.SetExternalName(cr, aws.StringValue(resp.ApiId))
	cre.ExternalNameAssigned = true
	return cre, nil
}

func (*external) preUpdate(context.Context, *svcapitypes.API) error {
	return nil
}

func (*external) postUpdate(_ context.Context, _ *svcapitypes.API, upd managed.ExternalUpdate, err error) (managed.ExternalUpdate, error) {
	return upd, err
}
func lateInitialize(*svcapitypes.APIParameters, *svcsdk.GetApiOutput) error {
	return nil
}

func preGenerateGetApiInput(_ *svcapitypes.API, obj *svcsdk.GetApiInput) *svcsdk.GetApiInput { //nolint:golint
	return obj
}

func postGenerateGetApiInput(cr *svcapitypes.API, obj *svcsdk.GetApiInput) *svcsdk.GetApiInput { //nolint:golint
	obj.ApiId = aws.String(meta.GetExternalName(cr))
	return obj
}

func preGenerateCreateApiInput(_ *svcapitypes.API, obj *svcsdk.CreateApiInput) *svcsdk.CreateApiInput { //nolint:golint

	return obj
}

func postGenerateCreateApiInput(cr *svcapitypes.API, obj *svcsdk.CreateApiInput) *svcsdk.CreateApiInput { //nolint:golint
	return obj
}

func preGenerateDeleteApiInput(_ *svcapitypes.API, obj *svcsdk.DeleteApiInput) *svcsdk.DeleteApiInput { //nolint:golint
	return obj
}

func postGenerateDeleteApiInput(cr *svcapitypes.API, obj *svcsdk.DeleteApiInput) *svcsdk.DeleteApiInput { //nolint:golint
	obj.ApiId = aws.String(meta.GetExternalName(cr))
	return obj
}
