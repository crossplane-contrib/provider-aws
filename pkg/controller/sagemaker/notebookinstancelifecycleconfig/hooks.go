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

package notebookinstancelifecycleconfig

import (
	"context"

	svcsdk "github.com/aws/aws-sdk-go/service/sagemaker"
	ctrl "sigs.k8s.io/controller-runtime"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	svcapitypes "github.com/crossplane/provider-aws/apis/sagemaker/v1alpha1"
)

// SetupNotebookInstanceLifecycleConfig adds a controller that reconciles NotebookInstanceLifecycleConfig.
func SetupNotebookInstanceLifecycleConfig(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(svcapitypes.NotebookInstanceLifecycleConfigGroupKind)
	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&svcapitypes.NotebookInstanceLifecycleConfig{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(svcapitypes.NotebookInstanceLifecycleConfigGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient()}),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

func (*external) preObserve(context.Context, *svcapitypes.NotebookInstanceLifecycleConfig) error {
	return nil
}
func (*external) postObserve(_ context.Context, cr *svcapitypes.NotebookInstanceLifecycleConfig, _ *svcsdk.DescribeNotebookInstanceLifecycleConfigOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}
	cr.SetConditions(xpv1.Available())
	return obs, err
}

func (*external) preCreate(context.Context, *svcapitypes.NotebookInstanceLifecycleConfig) error {
	return nil
}

func (*external) postCreate(_ context.Context, _ *svcapitypes.NotebookInstanceLifecycleConfig, _ *svcsdk.CreateNotebookInstanceLifecycleConfigOutput, cre managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
	return cre, err
}

func (*external) preUpdate(context.Context, *svcapitypes.NotebookInstanceLifecycleConfig) error {
	return nil
}

func (*external) postUpdate(_ context.Context, _ *svcapitypes.NotebookInstanceLifecycleConfig, upd managed.ExternalUpdate, err error) (managed.ExternalUpdate, error) {
	return upd, err
}
func lateInitialize(*svcapitypes.NotebookInstanceLifecycleConfigParameters, *svcsdk.DescribeNotebookInstanceLifecycleConfigOutput) error {
	return nil
}

func isUpToDate(*svcapitypes.NotebookInstanceLifecycleConfig, *svcsdk.DescribeNotebookInstanceLifecycleConfigOutput) bool {
	return true
}

func preGenerateDescribeNotebookInstanceLifecycleConfigInput(_ *svcapitypes.NotebookInstanceLifecycleConfig, obj *svcsdk.DescribeNotebookInstanceLifecycleConfigInput) *svcsdk.DescribeNotebookInstanceLifecycleConfigInput {
	return obj
}

func postGenerateDescribeNotebookInstanceLifecycleConfigInput(_ *svcapitypes.NotebookInstanceLifecycleConfig, obj *svcsdk.DescribeNotebookInstanceLifecycleConfigInput) *svcsdk.DescribeNotebookInstanceLifecycleConfigInput {
	return obj
}

func preGenerateCreateNotebookInstanceLifecycleConfigInput(_ *svcapitypes.NotebookInstanceLifecycleConfig, obj *svcsdk.CreateNotebookInstanceLifecycleConfigInput) *svcsdk.CreateNotebookInstanceLifecycleConfigInput {
	return obj
}

func postGenerateCreateNotebookInstanceLifecycleConfigInput(_ *svcapitypes.NotebookInstanceLifecycleConfig, obj *svcsdk.CreateNotebookInstanceLifecycleConfigInput) *svcsdk.CreateNotebookInstanceLifecycleConfigInput {
	return obj
}
func preGenerateDeleteNotebookInstanceLifecycleConfigInput(_ *svcapitypes.NotebookInstanceLifecycleConfig, obj *svcsdk.DeleteNotebookInstanceLifecycleConfigInput) *svcsdk.DeleteNotebookInstanceLifecycleConfigInput {
	return obj
}

func postGenerateDeleteNotebookInstanceLifecycleConfigInput(_ *svcapitypes.NotebookInstanceLifecycleConfig, obj *svcsdk.DeleteNotebookInstanceLifecycleConfigInput) *svcsdk.DeleteNotebookInstanceLifecycleConfigInput {
	return obj
}
