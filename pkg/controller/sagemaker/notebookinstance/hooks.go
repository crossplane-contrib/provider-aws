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

package notebookinstance

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

// SetupNotebookInstance adds a controller that reconciles NotebookInstance.
func SetupNotebookInstance(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(svcapitypes.NotebookInstanceGroupKind)
	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&svcapitypes.NotebookInstance{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(svcapitypes.NotebookInstanceGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient()}),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

func (*external) preObserve(context.Context, *svcapitypes.NotebookInstance) error {
	return nil
}
func (*external) postObserve(_ context.Context, cr *svcapitypes.NotebookInstance, _ *svcsdk.DescribeNotebookInstanceOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}
	cr.SetConditions(xpv1.Available())
	return obs, err
}

func (*external) preCreate(context.Context, *svcapitypes.NotebookInstance) error {
	return nil
}

func (*external) postCreate(_ context.Context, _ *svcapitypes.NotebookInstance, _ *svcsdk.CreateNotebookInstanceOutput, cre managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
	return cre, err
}

func (*external) preUpdate(context.Context, *svcapitypes.NotebookInstance) error {
	return nil
}

func (*external) postUpdate(_ context.Context, _ *svcapitypes.NotebookInstance, upd managed.ExternalUpdate, err error) (managed.ExternalUpdate, error) {
	return upd, err
}
func lateInitialize(*svcapitypes.NotebookInstanceParameters, *svcsdk.DescribeNotebookInstanceOutput) error {
	return nil
}

func isUpToDate(*svcapitypes.NotebookInstance, *svcsdk.DescribeNotebookInstanceOutput) bool {
	return true
}

func preGenerateDescribeNotebookInstanceInput(_ *svcapitypes.NotebookInstance, obj *svcsdk.DescribeNotebookInstanceInput) *svcsdk.DescribeNotebookInstanceInput {
	return obj
}

func postGenerateDescribeNotebookInstanceInput(_ *svcapitypes.NotebookInstance, obj *svcsdk.DescribeNotebookInstanceInput) *svcsdk.DescribeNotebookInstanceInput {
	return obj
}

func preGenerateCreateNotebookInstanceInput(_ *svcapitypes.NotebookInstance, obj *svcsdk.CreateNotebookInstanceInput) *svcsdk.CreateNotebookInstanceInput {
	return obj
}

func postGenerateCreateNotebookInstanceInput(_ *svcapitypes.NotebookInstance, obj *svcsdk.CreateNotebookInstanceInput) *svcsdk.CreateNotebookInstanceInput {
	return obj
}
func preGenerateDeleteNotebookInstanceInput(_ *svcapitypes.NotebookInstance, obj *svcsdk.DeleteNotebookInstanceInput) *svcsdk.DeleteNotebookInstanceInput {
	return obj
}

func postGenerateDeleteNotebookInstanceInput(_ *svcapitypes.NotebookInstance, obj *svcsdk.DeleteNotebookInstanceInput) *svcsdk.DeleteNotebookInstanceInput {
	return obj
}
