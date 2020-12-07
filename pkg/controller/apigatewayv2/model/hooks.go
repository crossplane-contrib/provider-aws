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

	"github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	svcapitypes "github.com/crossplane/provider-aws/apis/apigatewayv2/v1alpha1"
	aws "github.com/crossplane/provider-aws/pkg/clients"
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
			managed.WithInitializers(managed.NewDefaultProviderConfig(mgr.GetClient())),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

func (*external) preObserve(context.Context, *svcapitypes.Model) error {
	return nil
}
func (*external) postObserve(_ context.Context, cr *svcapitypes.Model, _ *svcsdk.GetModelOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}
	cr.SetConditions(v1alpha1.Available())
	return obs, err
}

func (*external) preCreate(context.Context, *svcapitypes.Model) error {
	return nil
}

func (*external) postCreate(_ context.Context, cr *svcapitypes.Model, resp *svcsdk.CreateModelOutput, cre managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
	if err != nil {
		return managed.ExternalCreation{}, err
	}
	meta.SetExternalName(cr, aws.StringValue(resp.ModelId))
	cre.ExternalNameAssigned = true
	return cre, nil
}

func (*external) preUpdate(context.Context, *svcapitypes.Model) error {
	return nil
}

func (*external) postUpdate(_ context.Context, _ *svcapitypes.Model, upd managed.ExternalUpdate, err error) (managed.ExternalUpdate, error) {
	return upd, err
}
func lateInitialize(*svcapitypes.ModelParameters, *svcsdk.GetModelOutput) error {
	return nil
}

func preGenerateGetModelInput(_ *svcapitypes.Model, obj *svcsdk.GetModelInput) *svcsdk.GetModelInput {
	return obj
}

func postGenerateGetModelInput(cr *svcapitypes.Model, obj *svcsdk.GetModelInput) *svcsdk.GetModelInput {
	obj.ApiId = cr.Spec.ForProvider.APIID
	obj.ModelId = aws.String(meta.GetExternalName(cr))
	return obj
}

func preGenerateCreateModelInput(_ *svcapitypes.Model, obj *svcsdk.CreateModelInput) *svcsdk.CreateModelInput {
	return obj
}

func postGenerateCreateModelInput(cr *svcapitypes.Model, obj *svcsdk.CreateModelInput) *svcsdk.CreateModelInput {
	obj.ApiId = cr.Spec.ForProvider.APIID
	return obj
}

func preGenerateDeleteModelInput(_ *svcapitypes.Model, obj *svcsdk.DeleteModelInput) *svcsdk.DeleteModelInput {
	return obj
}

func postGenerateDeleteModelInput(cr *svcapitypes.Model, obj *svcsdk.DeleteModelInput) *svcsdk.DeleteModelInput {
	obj.ApiId = cr.Spec.ForProvider.APIID
	obj.ModelId = aws.String(meta.GetExternalName(cr))
	return obj
}
