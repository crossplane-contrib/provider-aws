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

package routeresponse

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

// SetupRouteResponse adds a controller that reconciles RouteResponse.
func SetupRouteResponse(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(svcapitypes.RouteResponseGroupKind)
	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&svcapitypes.RouteResponse{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(svcapitypes.RouteResponseGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient()}),
			managed.WithInitializers(managed.NewDefaultProviderConfig(mgr.GetClient())),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

func (*external) preObserve(context.Context, *svcapitypes.RouteResponse) error {
	return nil
}
func (*external) postObserve(_ context.Context, cr *svcapitypes.RouteResponse, _ *svcsdk.GetRouteResponsesOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}
	cr.SetConditions(v1alpha1.Available())
	return obs, nil
}

func (*external) filterList(cr *svcapitypes.RouteResponse, list *svcsdk.GetRouteResponsesOutput) *svcsdk.GetRouteResponsesOutput {
	res := &svcsdk.GetRouteResponsesOutput{}
	for _, rr := range list.Items {
		if meta.GetExternalName(cr) == aws.StringValue(rr.RouteResponseId) {
			res.Items = append(res.Items, rr)
			break
		}
	}
	return res
}

func (*external) preCreate(context.Context, *svcapitypes.RouteResponse) error {
	return nil
}

func (e *external) postCreate(_ context.Context, cr *svcapitypes.RouteResponse, resp *svcsdk.CreateRouteResponseOutput, cre managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
	if err != nil {
		return managed.ExternalCreation{}, err
	}
	meta.SetExternalName(cr, aws.StringValue(resp.RouteResponseId))
	cre.ExternalNameAssigned = true
	return cre, nil
}

func (*external) preUpdate(context.Context, *svcapitypes.RouteResponse) error {
	return nil
}

func (*external) postUpdate(_ context.Context, _ *svcapitypes.RouteResponse, upd managed.ExternalUpdate, err error) (managed.ExternalUpdate, error) {
	return upd, err
}
func lateInitialize(*svcapitypes.RouteResponseParameters, *svcsdk.GetRouteResponsesOutput) error {
	return nil
}

func preGenerateGetRouteResponsesInput(_ *svcapitypes.RouteResponse, obj *svcsdk.GetRouteResponsesInput) *svcsdk.GetRouteResponsesInput {
	return obj
}

func postGenerateGetRouteResponsesInput(cr *svcapitypes.RouteResponse, obj *svcsdk.GetRouteResponsesInput) *svcsdk.GetRouteResponsesInput {
	obj.ApiId = cr.Spec.ForProvider.APIID
	obj.RouteId = cr.Spec.ForProvider.RouteID
	return obj
}

func preGenerateCreateRouteResponseInput(_ *svcapitypes.RouteResponse, obj *svcsdk.CreateRouteResponseInput) *svcsdk.CreateRouteResponseInput {
	return obj
}

func postGenerateCreateRouteResponseInput(cr *svcapitypes.RouteResponse, obj *svcsdk.CreateRouteResponseInput) *svcsdk.CreateRouteResponseInput {
	obj.ApiId = cr.Spec.ForProvider.APIID
	obj.RouteId = cr.Spec.ForProvider.RouteID
	return obj
}

func preGenerateDeleteRouteResponseInput(_ *svcapitypes.RouteResponse, obj *svcsdk.DeleteRouteResponseInput) *svcsdk.DeleteRouteResponseInput {
	return obj
}

func postGenerateDeleteRouteResponseInput(cr *svcapitypes.RouteResponse, obj *svcsdk.DeleteRouteResponseInput) *svcsdk.DeleteRouteResponseInput {
	obj.ApiId = cr.Spec.ForProvider.APIID
	obj.RouteId = cr.Spec.ForProvider.RouteID
	obj.RouteResponseId = aws.String(meta.GetExternalName(cr))
	return obj
}
