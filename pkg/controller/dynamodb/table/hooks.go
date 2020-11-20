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

package table

import (
	"context"

	svcsdk "github.com/aws/aws-sdk-go/service/dynamodb"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	svcapitypes "github.com/crossplane/provider-aws/apis/dynamodb/v1alpha1"
)

// SetupTable adds a controller that reconciles Table.
func SetupTable(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(svcapitypes.TableGroupKind)
	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&svcapitypes.Table{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(svcapitypes.TableGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient()}),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

func (*external) preObserve(context.Context, *svcapitypes.Table) error {
	return nil
}
func (*external) postObserve(_ context.Context, _ *svcapitypes.Table, _ *svcsdk.DescribeTableOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	return obs, err
}

func (*external) preCreate(context.Context, *svcapitypes.Table) error {
	return nil
}

func (*external) postCreate(_ context.Context, _ *svcapitypes.Table, _ *svcsdk.CreateTableOutput, cre managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
	return cre, err
}

func (*external) preUpdate(context.Context, *svcapitypes.Table) error {
	return nil
}

func (*external) postUpdate(_ context.Context, _ *svcapitypes.Table, upd managed.ExternalUpdate, err error) (managed.ExternalUpdate, error) {
	return upd, err
}
func lateInitialize(*svcapitypes.TableParameters, *svcsdk.DescribeTableOutput) error {
	return nil
}

func preGenerateDescribeTableInput(_ *svcapitypes.Table, obj *svcsdk.DescribeTableInput) *svcsdk.DescribeTableInput {
	return obj
}

func postGenerateDescribeTableInput(_ *svcapitypes.Table, obj *svcsdk.DescribeTableInput) *svcsdk.DescribeTableInput {
	return obj
}

func preGenerateCreateTableInput(_ *svcapitypes.Table, obj *svcsdk.CreateTableInput) *svcsdk.CreateTableInput {
	return obj
}

func postGenerateCreateTableInput(_ *svcapitypes.Table, obj *svcsdk.CreateTableInput) *svcsdk.CreateTableInput {
	return obj
}
func preGenerateDeleteTableInput(_ *svcapitypes.Table, obj *svcsdk.DeleteTableInput) *svcsdk.DeleteTableInput {
	return obj
}

func postGenerateDeleteTableInput(_ *svcapitypes.Table, obj *svcsdk.DeleteTableInput) *svcsdk.DeleteTableInput {
	return obj
}
