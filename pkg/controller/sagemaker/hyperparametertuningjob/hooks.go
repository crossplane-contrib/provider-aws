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

package hyperparametertuningjob

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

// SetupHyperParameterTuningJob adds a controller that reconciles HyperParameterTuningJob.
func SetupHyperParameterTuningJob(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(svcapitypes.HyperParameterTuningJobGroupKind)
	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&svcapitypes.HyperParameterTuningJob{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(svcapitypes.HyperParameterTuningJobGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient()}),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

func (*external) preObserve(context.Context, *svcapitypes.HyperParameterTuningJob) error {
	return nil
}
func (*external) postObserve(_ context.Context, cr *svcapitypes.HyperParameterTuningJob, _ *svcsdk.DescribeHyperParameterTuningJobOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}
	cr.SetConditions(xpv1.Available())
	return obs, err
}

func (*external) preCreate(context.Context, *svcapitypes.HyperParameterTuningJob) error {
	return nil
}

func (*external) postCreate(_ context.Context, _ *svcapitypes.HyperParameterTuningJob, _ *svcsdk.CreateHyperParameterTuningJobOutput, cre managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
	return cre, err
}

func (*external) preUpdate(context.Context, *svcapitypes.HyperParameterTuningJob) error {
	return nil
}

func (*external) postUpdate(_ context.Context, _ *svcapitypes.HyperParameterTuningJob, upd managed.ExternalUpdate, err error) (managed.ExternalUpdate, error) {
	return upd, err
}
func lateInitialize(*svcapitypes.HyperParameterTuningJobParameters, *svcsdk.DescribeHyperParameterTuningJobOutput) error {
	return nil
}

func isUpToDate(*svcapitypes.HyperParameterTuningJob, *svcsdk.DescribeHyperParameterTuningJobOutput) bool {
	return true
}

func preGenerateDescribeHyperParameterTuningJobInput(_ *svcapitypes.HyperParameterTuningJob, obj *svcsdk.DescribeHyperParameterTuningJobInput) *svcsdk.DescribeHyperParameterTuningJobInput {
	return obj
}

func postGenerateDescribeHyperParameterTuningJobInput(_ *svcapitypes.HyperParameterTuningJob, obj *svcsdk.DescribeHyperParameterTuningJobInput) *svcsdk.DescribeHyperParameterTuningJobInput {
	return obj
}

func preGenerateCreateHyperParameterTuningJobInput(_ *svcapitypes.HyperParameterTuningJob, obj *svcsdk.CreateHyperParameterTuningJobInput) *svcsdk.CreateHyperParameterTuningJobInput {
	return obj
}

func postGenerateCreateHyperParameterTuningJobInput(_ *svcapitypes.HyperParameterTuningJob, obj *svcsdk.CreateHyperParameterTuningJobInput) *svcsdk.CreateHyperParameterTuningJobInput {
	return obj
}
func (e *external) Delete(ctx context.Context, mg cpresource.Managed) error {
	// TODO: implement me!
	return nil
}
