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

package labelingjob

import (
	"context"

	svcsdk "github.com/aws/aws-sdk-go/service/sagemaker"
	ctrl "sigs.k8s.io/controller-runtime"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	cpresource "github.com/crossplane/crossplane-runtime/pkg/resource"

	svcapitypes "github.com/crossplane/provider-aws/apis/sagemaker/v1alpha1"
)

// SetupLabelingJob adds a controller that reconciles LabelingJob.
func SetupLabelingJob(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(svcapitypes.LabelingJobGroupKind)
	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&svcapitypes.LabelingJob{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(svcapitypes.LabelingJobGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient()}),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

func (*external) preObserve(context.Context, *svcapitypes.LabelingJob) error {
	return nil
}
func (*external) postObserve(_ context.Context, cr *svcapitypes.LabelingJob, _ *svcsdk.DescribeLabelingJobOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}
	cr.SetConditions(xpv1.Available())
	return obs, err
}

func (*external) preCreate(context.Context, *svcapitypes.LabelingJob) error {
	return nil
}

func (*external) postCreate(_ context.Context, _ *svcapitypes.LabelingJob, _ *svcsdk.CreateLabelingJobOutput, cre managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
	return cre, err
}

func (*external) preUpdate(context.Context, *svcapitypes.LabelingJob) error {
	return nil
}

func (*external) postUpdate(_ context.Context, _ *svcapitypes.LabelingJob, upd managed.ExternalUpdate, err error) (managed.ExternalUpdate, error) {
	return upd, err
}
func lateInitialize(*svcapitypes.LabelingJobParameters, *svcsdk.DescribeLabelingJobOutput) error {
	return nil
}

func isUpToDate(*svcapitypes.LabelingJob, *svcsdk.DescribeLabelingJobOutput) bool {
	return true
}

func preGenerateDescribeLabelingJobInput(_ *svcapitypes.LabelingJob, obj *svcsdk.DescribeLabelingJobInput) *svcsdk.DescribeLabelingJobInput {
	return obj
}

func postGenerateDescribeLabelingJobInput(_ *svcapitypes.LabelingJob, obj *svcsdk.DescribeLabelingJobInput) *svcsdk.DescribeLabelingJobInput {
	return obj
}

func preGenerateCreateLabelingJobInput(_ *svcapitypes.LabelingJob, obj *svcsdk.CreateLabelingJobInput) *svcsdk.CreateLabelingJobInput {
	return obj
}

func postGenerateCreateLabelingJobInput(_ *svcapitypes.LabelingJob, obj *svcsdk.CreateLabelingJobInput) *svcsdk.CreateLabelingJobInput {
	return obj
}
func (e *external) Delete(ctx context.Context, mg cpresource.Managed) error {
	// TODO: implement me!
	return nil
}
