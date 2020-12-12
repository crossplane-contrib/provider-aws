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

package integration

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

// SetupIntegration adds a controller that reconciles Integration.
func SetupIntegration(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(svcapitypes.IntegrationGroupKind)
	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&svcapitypes.Integration{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(svcapitypes.IntegrationGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient()}),
			managed.WithInitializers(managed.NewDefaultProviderConfig(mgr.GetClient())),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

func (*external) preObserve(context.Context, *svcapitypes.Integration) error {
	return nil
}
func (*external) postObserve(_ context.Context, cr *svcapitypes.Integration, _ *svcsdk.GetIntegrationOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}
	cr.SetConditions(xpv1.Available())
	return obs, nil
}

func (*external) preCreate(context.Context, *svcapitypes.Integration) error {
	return nil
}

func (e *external) postCreate(_ context.Context, cr *svcapitypes.Integration, resp *svcsdk.CreateIntegrationOutput, cre managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
	if err != nil {
		return managed.ExternalCreation{}, err
	}
	meta.SetExternalName(cr, aws.StringValue(resp.IntegrationId))
	cre.ExternalNameAssigned = true
	return cre, nil
}

func (*external) preUpdate(context.Context, *svcapitypes.Integration) error {
	return nil
}

func (*external) postUpdate(_ context.Context, _ *svcapitypes.Integration, upd managed.ExternalUpdate, err error) (managed.ExternalUpdate, error) {
	return upd, err
}
func lateInitialize(*svcapitypes.IntegrationParameters, *svcsdk.GetIntegrationOutput) error {
	return nil
}

func preGenerateGetIntegrationInput(_ *svcapitypes.Integration, obj *svcsdk.GetIntegrationInput) *svcsdk.GetIntegrationInput {
	return obj
}

func postGenerateGetIntegrationInput(cr *svcapitypes.Integration, obj *svcsdk.GetIntegrationInput) *svcsdk.GetIntegrationInput {
	obj.ApiId = cr.Spec.ForProvider.APIID
	obj.IntegrationId = aws.String(meta.GetExternalName(cr))
	return obj
}

func preGenerateCreateIntegrationInput(_ *svcapitypes.Integration, obj *svcsdk.CreateIntegrationInput) *svcsdk.CreateIntegrationInput {
	return obj
}

func postGenerateCreateIntegrationInput(cr *svcapitypes.Integration, obj *svcsdk.CreateIntegrationInput) *svcsdk.CreateIntegrationInput {
	obj.ApiId = cr.Spec.ForProvider.APIID
	return obj
}

func preGenerateDeleteIntegrationInput(_ *svcapitypes.Integration, obj *svcsdk.DeleteIntegrationInput) *svcsdk.DeleteIntegrationInput {
	return obj
}

func postGenerateDeleteIntegrationInput(cr *svcapitypes.Integration, obj *svcsdk.DeleteIntegrationInput) *svcsdk.DeleteIntegrationInput {
	obj.ApiId = cr.Spec.ForProvider.APIID
	obj.IntegrationId = aws.String(meta.GetExternalName(cr))
	return obj
}
