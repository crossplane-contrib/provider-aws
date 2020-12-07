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

package domainname

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

// SetupDomainName adds a controller that reconciles DomainName.
func SetupDomainName(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(svcapitypes.DomainNameGroupKind)
	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&svcapitypes.DomainName{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(svcapitypes.DomainNameGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient()}),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

func (*external) preObserve(context.Context, *svcapitypes.DomainName) error {
	return nil
}
func (*external) postObserve(_ context.Context, cr *svcapitypes.DomainName, _ *svcsdk.GetDomainNameOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}
	cr.SetConditions(v1alpha1.Available())
	return obs, nil
}

func (*external) preCreate(context.Context, *svcapitypes.DomainName) error {
	return nil
}

func (*external) postCreate(_ context.Context, _ *svcapitypes.DomainName, _ *svcsdk.CreateDomainNameOutput, cre managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
	return cre, err
}

func (*external) preUpdate(context.Context, *svcapitypes.DomainName) error {
	return nil
}

func (*external) postUpdate(_ context.Context, _ *svcapitypes.DomainName, upd managed.ExternalUpdate, err error) (managed.ExternalUpdate, error) {
	return upd, err
}
func lateInitialize(*svcapitypes.DomainNameParameters, *svcsdk.GetDomainNameOutput) error {
	return nil
}

func preGenerateGetDomainNameInput(_ *svcapitypes.DomainName, obj *svcsdk.GetDomainNameInput) *svcsdk.GetDomainNameInput {
	return obj
}

func postGenerateGetDomainNameInput(cr *svcapitypes.DomainName, obj *svcsdk.GetDomainNameInput) *svcsdk.GetDomainNameInput {
	obj.DomainName = aws.String(meta.GetExternalName(cr))
	return obj
}

func preGenerateCreateDomainNameInput(_ *svcapitypes.DomainName, obj *svcsdk.CreateDomainNameInput) *svcsdk.CreateDomainNameInput {
	return obj
}

func postGenerateCreateDomainNameInput(cr *svcapitypes.DomainName, obj *svcsdk.CreateDomainNameInput) *svcsdk.CreateDomainNameInput {
	obj.DomainName = aws.String(meta.GetExternalName(cr))
	return obj
}

func preGenerateDeleteDomainNameInput(_ *svcapitypes.DomainName, obj *svcsdk.DeleteDomainNameInput) *svcsdk.DeleteDomainNameInput {
	return obj
}

func postGenerateDeleteDomainNameInput(cr *svcapitypes.DomainName, obj *svcsdk.DeleteDomainNameInput) *svcsdk.DeleteDomainNameInput {
	obj.DomainName = aws.String(meta.GetExternalName(cr))
	return obj
}
