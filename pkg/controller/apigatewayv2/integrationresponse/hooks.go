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

package integrationresponse

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

// SetupIntegrationResponse adds a controller that reconciles IntegrationResponse.
func SetupIntegrationResponse(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(svcapitypes.IntegrationResponseGroupKind)
	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&svcapitypes.IntegrationResponse{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(svcapitypes.IntegrationResponseGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient()}),
			managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
			managed.WithConnectionPublishers(),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

func (*external) preObserve(context.Context, *svcapitypes.IntegrationResponse) error {
	return nil
}
func (*external) postObserve(_ context.Context, _ *svcapitypes.IntegrationResponse, _ *svcsdk.GetIntegrationResponsesOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	return obs, err
}

func (*external) filterList(_ *svcapitypes.IntegrationResponse, list *svcsdk.GetIntegrationResponsesOutput) *svcsdk.GetIntegrationResponsesOutput {
	return list
}

func (*external) preCreate(context.Context, *svcapitypes.IntegrationResponse) error {
	return nil
}

func (*external) postCreate(_ context.Context, _ *svcapitypes.IntegrationResponse, _ *svcsdk.CreateIntegrationResponseOutput, cre managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
	return cre, err
}

func (*external) preUpdate(context.Context, *svcapitypes.IntegrationResponse) error {
	return nil
}

func (*external) postUpdate(_ context.Context, _ *svcapitypes.IntegrationResponse, upd managed.ExternalUpdate, err error) (managed.ExternalUpdate, error) {
	return upd, err
}
func lateInitialize(*svcapitypes.IntegrationResponseParameters, *svcsdk.GetIntegrationResponsesOutput) error {
	return nil
}

func preGenerateGetIntegrationResponsesInput(_ *svcapitypes.IntegrationResponse, obj *svcsdk.GetIntegrationResponsesInput) *svcsdk.GetIntegrationResponsesInput {
	return obj
}

func postGenerateGetIntegrationResponsesInput(_ *svcapitypes.IntegrationResponse, obj *svcsdk.GetIntegrationResponsesInput) *svcsdk.GetIntegrationResponsesInput {
	return obj
}

func preGenerateCreateIntegrationResponseInput(_ *svcapitypes.IntegrationResponse, obj *svcsdk.CreateIntegrationResponseInput) *svcsdk.CreateIntegrationResponseInput {
	return obj
}

func postGenerateCreateIntegrationResponseInput(_ *svcapitypes.IntegrationResponse, obj *svcsdk.CreateIntegrationResponseInput) *svcsdk.CreateIntegrationResponseInput {
	return obj
}

func preGenerateDeleteIntegrationResponseInput(_ *svcapitypes.IntegrationResponse, obj *svcsdk.DeleteIntegrationResponseInput) *svcsdk.DeleteIntegrationResponseInput {
	return obj
}

func postGenerateDeleteIntegrationResponseInput(_ *svcapitypes.IntegrationResponse, obj *svcsdk.DeleteIntegrationResponseInput) *svcsdk.DeleteIntegrationResponseInput {
	return obj
}
