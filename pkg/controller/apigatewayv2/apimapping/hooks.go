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

package apimapping

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

// SetupAPIMapping adds a controller that reconciles APIMapping.
func SetupAPIMapping(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(svcapitypes.APIMappingGroupKind)
	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&svcapitypes.APIMapping{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(svcapitypes.APIMappingGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient()}),
			managed.WithInitializers(managed.NewDefaultProviderConfig(mgr.GetClient())),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

func (*external) preObserve(context.Context, *svcapitypes.APIMapping) error {
	return nil
}
func (*external) postObserve(_ context.Context, cr *svcapitypes.APIMapping, _ *svcsdk.GetApiMappingsOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}
	cr.SetConditions(v1alpha1.Available())
	return obs, nil
}

func (*external) filterList(cr *svcapitypes.APIMapping, list *svcsdk.GetApiMappingsOutput) *svcsdk.GetApiMappingsOutput {
	res := &svcsdk.GetApiMappingsOutput{}
	for _, am := range list.Items {
		if meta.GetExternalName(cr) == aws.StringValue(am.ApiMappingId) {
			res.Items = append(res.Items, am)
		}
	}
	return res
}

func (*external) preCreate(context.Context, *svcapitypes.APIMapping) error {
	return nil
}

func (e *external) postCreate(ctx context.Context, cr *svcapitypes.APIMapping, resp *svcsdk.CreateApiMappingOutput, cre managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
	if err != nil {
		return managed.ExternalCreation{}, err
	}
	meta.SetExternalName(cr, aws.StringValue(resp.ApiMappingId))
	return cre, e.kube.Update(ctx, cr)
}

func (*external) preUpdate(context.Context, *svcapitypes.APIMapping) error {
	return nil
}

func (*external) postUpdate(_ context.Context, _ *svcapitypes.APIMapping, upd managed.ExternalUpdate, err error) (managed.ExternalUpdate, error) {
	return upd, err
}
func lateInitialize(*svcapitypes.APIMappingParameters, *svcsdk.GetApiMappingsOutput) error {
	return nil
}

func preGenerateGetApiMappingsInput(_ *svcapitypes.APIMapping, obj *svcsdk.GetApiMappingsInput) *svcsdk.GetApiMappingsInput { // nolint:golint
	return obj
}

func postGenerateGetApiMappingsInput(cr *svcapitypes.APIMapping, obj *svcsdk.GetApiMappingsInput) *svcsdk.GetApiMappingsInput { // nolint:golint
	obj.DomainName = cr.Spec.ForProvider.DomainName
	return obj
}

func preGenerateCreateApiMappingInput(_ *svcapitypes.APIMapping, obj *svcsdk.CreateApiMappingInput) *svcsdk.CreateApiMappingInput { // nolint:golint
	return obj
}

func postGenerateCreateApiMappingInput(cr *svcapitypes.APIMapping, obj *svcsdk.CreateApiMappingInput) *svcsdk.CreateApiMappingInput { // nolint:golint
	obj.ApiId = cr.Spec.ForProvider.APIID
	obj.DomainName = cr.Spec.ForProvider.DomainName
	obj.Stage = cr.Spec.ForProvider.Stage
	return obj
}

func preGenerateDeleteApiMappingInput(_ *svcapitypes.APIMapping, obj *svcsdk.DeleteApiMappingInput) *svcsdk.DeleteApiMappingInput { // nolint:golint
	return obj
}

func postGenerateDeleteApiMappingInput(cr *svcapitypes.APIMapping, obj *svcsdk.DeleteApiMappingInput) *svcsdk.DeleteApiMappingInput { // nolint:golint
	obj.ApiMappingId = aws.String(meta.GetExternalName(cr))
	obj.DomainName = cr.Spec.ForProvider.DomainName
	return obj
}
