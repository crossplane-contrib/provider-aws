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

package deployment

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

// SetupDeployment adds a controller that reconciles Deployment.
func SetupDeployment(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(svcapitypes.DeploymentGroupKind)
	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&svcapitypes.Deployment{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(svcapitypes.DeploymentGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient()}),
			managed.WithInitializers(managed.NewDefaultProviderConfig(mgr.GetClient())),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

func (*external) preObserve(context.Context, *svcapitypes.Deployment) error {
	return nil
}
func (*external) postObserve(_ context.Context, cr *svcapitypes.Deployment, _ *svcsdk.GetDeploymentOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}
	cr.SetConditions(v1alpha1.Available())
	return obs, nil
}

func (*external) preCreate(context.Context, *svcapitypes.Deployment) error {
	return nil
}

func (e *external) postCreate(_ context.Context, cr *svcapitypes.Deployment, resp *svcsdk.CreateDeploymentOutput, cre managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
	if err != nil {
		return managed.ExternalCreation{}, err
	}
	meta.SetExternalName(cr, aws.StringValue(resp.DeploymentId))
	cre.ExternalNameAssigned = true
	return cre, nil
}

func (*external) preUpdate(context.Context, *svcapitypes.Deployment) error {
	return nil
}

func (*external) postUpdate(_ context.Context, _ *svcapitypes.Deployment, upd managed.ExternalUpdate, err error) (managed.ExternalUpdate, error) {
	return upd, err
}
func lateInitialize(*svcapitypes.DeploymentParameters, *svcsdk.GetDeploymentOutput) error {
	return nil
}

func preGenerateGetDeploymentInput(_ *svcapitypes.Deployment, obj *svcsdk.GetDeploymentInput) *svcsdk.GetDeploymentInput {
	return obj
}

func postGenerateGetDeploymentInput(cr *svcapitypes.Deployment, obj *svcsdk.GetDeploymentInput) *svcsdk.GetDeploymentInput {
	obj.ApiId = cr.Spec.ForProvider.APIID
	obj.DeploymentId = aws.String(meta.GetExternalName(cr))
	return obj
}

func preGenerateCreateDeploymentInput(_ *svcapitypes.Deployment, obj *svcsdk.CreateDeploymentInput) *svcsdk.CreateDeploymentInput {
	return obj
}

func postGenerateCreateDeploymentInput(cr *svcapitypes.Deployment, obj *svcsdk.CreateDeploymentInput) *svcsdk.CreateDeploymentInput {
	obj.ApiId = cr.Spec.ForProvider.APIID
	return obj
}

func preGenerateDeleteDeploymentInput(_ *svcapitypes.Deployment, obj *svcsdk.DeleteDeploymentInput) *svcsdk.DeleteDeploymentInput {
	return obj
}

func postGenerateDeleteDeploymentInput(cr *svcapitypes.Deployment, obj *svcsdk.DeleteDeploymentInput) *svcsdk.DeleteDeploymentInput {
	obj.ApiId = cr.Spec.ForProvider.APIID
	obj.DeploymentId = aws.String(meta.GetExternalName(cr))
	return obj
}
