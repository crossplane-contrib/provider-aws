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

package model

import (
	"context"

	svcsdk "github.com/aws/aws-sdk-go/service/apigatewayv2"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	svcapitypes "github.com/crossplane/provider-aws/apis/apigatewayv2/v1alpha1"
)

// SetupModel adds a controller that reconciles Model.
func SetupModel(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(svcapitypes.ModelGroupKind)
	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&svcapitypes.Model{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(svcapitypes.ModelGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient()}),
			managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
			managed.WithConnectionPublishers(),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

func (*external) preObserve(context.Context, *svcapitypes.Model) error {
	return nil
}
func (*external) postObserve(_ context.Context, _ *svcapitypes.Model, _ *svcsdk.GetModelsOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	return obs, err
}

func (*external) filterList(_ *svcapitypes.Model, list *svcsdk.GetModelsOutput) *svcsdk.GetModelsOutput {
	return list
}

func (*external) preCreate(context.Context, *svcapitypes.Model) error {
	return nil
}

func (*external) postCreate(_ context.Context, _ *svcapitypes.Model, _ *svcsdk.CreateModelOutput, cre managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
	return cre, err
}

func (*external) preUpdate(context.Context, *svcapitypes.Model) error {
	return nil
}

func (*external) postUpdate(_ context.Context, _ *svcapitypes.Model, upd managed.ExternalUpdate, err error) (managed.ExternalUpdate, error) {
	return upd, err
}
func lateInitialize(*svcapitypes.ModelParameters, *svcsdk.GetModelsOutput) error {
	return nil
}

func preGenerateGetModelsInput(_ *svcapitypes.Model, obj *svcsdk.GetModelsInput) *svcsdk.GetModelsInput {
	return obj
}

func postGenerateGetModelsInput(_ *svcapitypes.Model, obj *svcsdk.GetModelsInput) *svcsdk.GetModelsInput {
	return obj
}

func preGenerateCreateModelInput(_ *svcapitypes.Model, obj *svcsdk.CreateModelInput) *svcsdk.CreateModelInput {
	return obj
}

func postGenerateCreateModelInput(_ *svcapitypes.Model, obj *svcsdk.CreateModelInput) *svcsdk.CreateModelInput {
	return obj
}

func preGenerateDeleteModelInput(_ *svcapitypes.Model, obj *svcsdk.DeleteModelInput) *svcsdk.DeleteModelInput {
	return obj
}

func postGenerateDeleteModelInput(_ *svcapitypes.Model, obj *svcsdk.DeleteModelInput) *svcsdk.DeleteModelInput {
	return obj
}
