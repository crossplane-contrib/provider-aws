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

package globaltable

import (
	"context"

	svcsdk "github.com/aws/aws-sdk-go/service/dynamodb"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	svcapitypes "github.com/crossplane/provider-aws/apis/dynamodb/v1alpha1"
	aws "github.com/crossplane/provider-aws/pkg/clients"
)

// SetupGlobalTable adds a controller that reconciles GlobalTable.
func SetupGlobalTable(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(svcapitypes.GlobalTableGroupKind)
	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&svcapitypes.GlobalTable{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(svcapitypes.GlobalTableGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient()}),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

func (*external) preObserve(context.Context, *svcapitypes.GlobalTable) error {
	return nil
}
func (*external) postObserve(_ context.Context, cr *svcapitypes.GlobalTable, resp *svcsdk.DescribeGlobalTableOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}
	switch aws.StringValue(resp.GlobalTableDescription.GlobalTableStatus) {
	case string(svcapitypes.GlobalTableStatus_SDK_ACTIVE):
		cr.SetConditions(v1alpha1.Available())
	case string(svcapitypes.GlobalTableStatus_SDK_CREATING):
		cr.SetConditions(v1alpha1.Creating())
	case string(svcapitypes.GlobalTableStatus_SDK_DELETING):
		cr.SetConditions(v1alpha1.Deleting())
	}
	return obs, nil
}

func (*external) preCreate(context.Context, *svcapitypes.GlobalTable) error {
	return nil
}

func (*external) postCreate(_ context.Context, _ *svcapitypes.GlobalTable, _ *svcsdk.CreateGlobalTableOutput, cre managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
	return cre, err
}

func (*external) preUpdate(context.Context, *svcapitypes.GlobalTable) error {
	return nil
}

func (*external) postUpdate(_ context.Context, _ *svcapitypes.GlobalTable, upd managed.ExternalUpdate, err error) (managed.ExternalUpdate, error) {
	return upd, err
}
func lateInitialize(*svcapitypes.GlobalTableParameters, *svcsdk.DescribeGlobalTableOutput) error {
	return nil
}

func preGenerateDescribeGlobalTableInput(_ *svcapitypes.GlobalTable, obj *svcsdk.DescribeGlobalTableInput) *svcsdk.DescribeGlobalTableInput {
	return obj
}

func postGenerateDescribeGlobalTableInput(cr *svcapitypes.GlobalTable, obj *svcsdk.DescribeGlobalTableInput) *svcsdk.DescribeGlobalTableInput {
	obj.GlobalTableName = aws.String(meta.GetExternalName(cr))
	return obj
}

func preGenerateCreateGlobalTableInput(_ *svcapitypes.GlobalTable, obj *svcsdk.CreateGlobalTableInput) *svcsdk.CreateGlobalTableInput {
	return obj
}

func postGenerateCreateGlobalTableInput(cr *svcapitypes.GlobalTable, obj *svcsdk.CreateGlobalTableInput) *svcsdk.CreateGlobalTableInput {
	obj.GlobalTableName = aws.String(meta.GetExternalName(cr))
	return obj
}
