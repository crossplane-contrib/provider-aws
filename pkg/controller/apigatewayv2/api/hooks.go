/*
Copyright 2019 The Crossplane Authors.

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

	"github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"

	svcsdk "github.com/aws/aws-sdk-go/service/apigatewayv2"
	ctrl "sigs.k8s.io/controller-runtime"

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
			managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
			managed.WithConnectionPublishers(),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

func (*external) preObserve(context.Context, *svcapitypes.API) error {
	return nil
}
func (*external) postObserve(_ context.Context, cr *svcapitypes.API, _ *svcsdk.GetApisOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	cr.SetConditions(v1alpha1.Available())
	return obs, err
}

func (*external) filterList(cr *svcapitypes.API, list *svcsdk.GetApisOutput) *svcsdk.GetApisOutput {
	res := &svcsdk.GetApisOutput{}
	for _, api := range list.Items {
		if meta.GetExternalName(cr) == aws.StringValue(api.Name) {
			res.Items = append(res.Items, api)
		}
	}
	return res
}

func (*external) preCreate(context.Context, *svcapitypes.API) error {
	return nil
}

func (*external) postCreate(_ context.Context, _ *svcapitypes.API, _ *svcsdk.CreateApiOutput, cre managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
	return cre, err
}

func (*external) preUpdate(context.Context, *svcapitypes.API) error {
	return nil
}

func (*external) postUpdate(_ context.Context, _ *svcapitypes.API, upd managed.ExternalUpdate, err error) (managed.ExternalUpdate, error) {
	return upd, err
}
func lateInitialize(*svcapitypes.APIParameters, *svcsdk.GetApisOutput) error {
	return nil
}

func preGenerateGetApisInput(_ *svcapitypes.API, obj *svcsdk.GetApisInput) *svcsdk.GetApisInput {
	return obj
}

func postGenerateGetApisInput(_ *svcapitypes.API, obj *svcsdk.GetApisInput) *svcsdk.GetApisInput {
	return obj
}

func preGenerateCreateApiInput(cr *svcapitypes.API, obj *svcsdk.CreateApiInput) *svcsdk.CreateApiInput { //nolint:golint
	obj.Name = aws.String(meta.GetExternalName(cr))
	return obj
}

func postGenerateCreateApiInput(_ *svcapitypes.API, obj *svcsdk.CreateApiInput) *svcsdk.CreateApiInput { //nolint:golint
	return obj
}

func preGenerateDeleteApiInput(_ *svcapitypes.API, obj *svcsdk.DeleteApiInput) *svcsdk.DeleteApiInput { //nolint:golint
	return obj
}

func postGenerateDeleteApiInput(_ *svcapitypes.API, obj *svcsdk.DeleteApiInput) *svcsdk.DeleteApiInput { //nolint:golint
	return obj
}
