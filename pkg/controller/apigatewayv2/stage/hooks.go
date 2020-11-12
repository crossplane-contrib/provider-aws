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

package stage

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

// SetupStage adds a controller that reconciles Stage.
func SetupStage(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(svcapitypes.StageGroupKind)
	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&svcapitypes.Stage{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(svcapitypes.StageGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient()}),
			managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
			managed.WithConnectionPublishers(),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

func (*external) preObserve(context.Context, *svcapitypes.Stage) error {
	return nil
}
func (*external) postObserve(_ context.Context, cr *svcapitypes.Stage, _ *svcsdk.GetStagesOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	cr.SetConditions(v1alpha1.Available())
	return obs, err
}

func (*external) filterList(cr *svcapitypes.Stage, list *svcsdk.GetStagesOutput) *svcsdk.GetStagesOutput {
	res := &svcsdk.GetStagesOutput{}
	for _, stage := range list.Items {
		if meta.GetExternalName(cr) == aws.StringValue(stage.StageName) {
			res.Items = append(res.Items, stage)
		}
	}
	return res
}

func (*external) preCreate(context.Context, *svcapitypes.Stage) error {
	return nil
}

func (*external) postCreate(_ context.Context, _ *svcapitypes.Stage, _ *svcsdk.CreateStageOutput, cre managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
	return cre, err
}

func (*external) preUpdate(context.Context, *svcapitypes.Stage) error {
	return nil
}

func (*external) postUpdate(_ context.Context, _ *svcapitypes.Stage, upd managed.ExternalUpdate, err error) (managed.ExternalUpdate, error) {
	return upd, err
}
func lateInitialize(*svcapitypes.StageParameters, *svcsdk.GetStagesOutput) error {
	return nil
}

func preGenerateGetStagesInput(_ *svcapitypes.Stage, obj *svcsdk.GetStagesInput) *svcsdk.GetStagesInput {
	return obj
}

func postGenerateGetStagesInput(cr *svcapitypes.Stage, obj *svcsdk.GetStagesInput) *svcsdk.GetStagesInput {
	obj.ApiId = cr.Spec.ForProvider.APIID
	return obj
}

func preGenerateCreateStageInput(_ *svcapitypes.Stage, obj *svcsdk.CreateStageInput) *svcsdk.CreateStageInput {
	return obj
}

func postGenerateCreateStageInput(cr *svcapitypes.Stage, obj *svcsdk.CreateStageInput) *svcsdk.CreateStageInput {
	obj.ApiId = cr.Spec.ForProvider.APIID
	obj.StageName = aws.String(meta.GetExternalName(cr))
	return obj
}

func preGenerateDeleteStageInput(cr *svcapitypes.Stage, obj *svcsdk.DeleteStageInput) *svcsdk.DeleteStageInput {
	obj.StageName = aws.String(meta.GetExternalName(cr))
	obj.ApiId = cr.Spec.ForProvider.CustomStageParameters.APIID
	return obj
}

func postGenerateDeleteStageInput(_ *svcapitypes.Stage, obj *svcsdk.DeleteStageInput) *svcsdk.DeleteStageInput {
	return obj
}
