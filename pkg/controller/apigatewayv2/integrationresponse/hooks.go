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

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	svcapitypes "github.com/crossplane/provider-aws/apis/apigatewayv2/v1alpha1"
	aws "github.com/crossplane/provider-aws/pkg/clients"
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
			managed.WithInitializers(managed.NewDefaultProviderConfig(mgr.GetClient())),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

func (*external) preObserve(context.Context, *svcapitypes.IntegrationResponse) error {
	return nil
}
func (*external) postObserve(_ context.Context, cr *svcapitypes.IntegrationResponse, _ *svcsdk.GetIntegrationResponseOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}
	cr.SetConditions(xpv1.Available())
	return obs, nil
}

func (*external) preCreate(context.Context, *svcapitypes.IntegrationResponse) error {
	return nil
}

func (e *external) postCreate(_ context.Context, cr *svcapitypes.IntegrationResponse, resp *svcsdk.CreateIntegrationResponseOutput, cre managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
	if err != nil {
		return managed.ExternalCreation{}, err
	}
	meta.SetExternalName(cr, aws.StringValue(resp.IntegrationResponseId))
	cre.ExternalNameAssigned = true
	return cre, nil
}

func (*external) preUpdate(context.Context, *svcapitypes.IntegrationResponse) error {
	return nil
}

func (*external) postUpdate(_ context.Context, _ *svcapitypes.IntegrationResponse, upd managed.ExternalUpdate, err error) (managed.ExternalUpdate, error) {
	return upd, err
}
func lateInitialize(*svcapitypes.IntegrationResponseParameters, *svcsdk.GetIntegrationResponseOutput) error {
	return nil
}

func preGenerateGetIntegrationResponseInput(_ *svcapitypes.IntegrationResponse, obj *svcsdk.GetIntegrationResponseInput) *svcsdk.GetIntegrationResponseInput {
	return obj
}

func postGenerateGetIntegrationResponseInput(cr *svcapitypes.IntegrationResponse, obj *svcsdk.GetIntegrationResponseInput) *svcsdk.GetIntegrationResponseInput {
	obj.ApiId = cr.Spec.ForProvider.APIID
	obj.IntegrationId = cr.Spec.ForProvider.IntegrationID
	obj.IntegrationResponseId = aws.String(meta.GetExternalName(cr))
	return obj
}

func preGenerateCreateIntegrationResponseInput(_ *svcapitypes.IntegrationResponse, obj *svcsdk.CreateIntegrationResponseInput) *svcsdk.CreateIntegrationResponseInput {
	return obj
}

func postGenerateCreateIntegrationResponseInput(cr *svcapitypes.IntegrationResponse, obj *svcsdk.CreateIntegrationResponseInput) *svcsdk.CreateIntegrationResponseInput {
	obj.ApiId = cr.Spec.ForProvider.APIID
	obj.IntegrationId = cr.Spec.ForProvider.IntegrationID
	return obj
}

func preGenerateDeleteIntegrationResponseInput(_ *svcapitypes.IntegrationResponse, obj *svcsdk.DeleteIntegrationResponseInput) *svcsdk.DeleteIntegrationResponseInput {
	return obj
}

func postGenerateDeleteIntegrationResponseInput(cr *svcapitypes.IntegrationResponse, obj *svcsdk.DeleteIntegrationResponseInput) *svcsdk.DeleteIntegrationResponseInput {
	obj.ApiId = cr.Spec.ForProvider.APIID
	obj.IntegrationId = cr.Spec.ForProvider.IntegrationID
	obj.IntegrationResponseId = aws.String(meta.GetExternalName(cr))
	return obj
}
