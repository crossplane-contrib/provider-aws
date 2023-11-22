/*
Copyright 2019 The Crossplane Authors.
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

package jobqueue

import (
	"context"
	"errors"

	awsarn "github.com/aws/aws-sdk-go/aws/arn"
	svcsdk "github.com/aws/aws-sdk-go/service/batch"
	"github.com/aws/aws-sdk-go/service/batch/batchiface"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	ctrl "sigs.k8s.io/controller-runtime"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/batch/v1alpha1"
	svcutils "github.com/crossplane-contrib/provider-aws/pkg/controller/batch/utils"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
	errorutils "github.com/crossplane-contrib/provider-aws/pkg/utils/errors"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
	custommanaged "github.com/crossplane-contrib/provider-aws/pkg/utils/reconciler/managed"
)

const (
	errComputeEnvironmentARN = "missing or incorrect ARN for ComputeEnvironment"
)

// SetupJobQueue adds a controller that reconciles a JobQueue.
func SetupJobQueue(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(svcapitypes.JobQueueGroupKind)
	opts := []option{
		func(e *external) {
			h := &hooks{client: e.client}
			e.preObserve = preObserve
			e.postObserve = postObserve
			e.lateInitialize = lateInitialize
			e.isUpToDate = isUpToDate
			e.preUpdate = preUpdate
			e.postUpdate = h.postUpdate
			e.preCreate = h.preCreate
			e.preDelete = h.preDelete
		},
	}

	reconcilerOpts := []managed.ReconcilerOption{
		managed.WithCriticalAnnotationUpdater(custommanaged.NewRetryingCriticalAnnotationUpdater(mgr.GetClient())),
		managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), opts: opts}),
		managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
		managed.WithInitializers(managed.NewNameAsExternalName(mgr.GetClient())),
		managed.WithPollInterval(o.PollInterval),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
	}

	if o.Features.Enabled(features.EnableAlphaManagementPolicies) {
		reconcilerOpts = append(reconcilerOpts, managed.WithManagementPolicies())
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(svcapitypes.JobQueueGroupVersionKind),
		reconcilerOpts...)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&svcapitypes.JobQueue{}).
		Complete(r)
}

type hooks struct {
	client batchiface.BatchAPI
}

func preObserve(_ context.Context, cr *svcapitypes.JobQueue, obj *svcsdk.DescribeJobQueuesInput) error {
	obj.JobQueues = []*string{pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))} // we only want to observe our JQ
	return nil
}

func postObserve(_ context.Context, cr *svcapitypes.JobQueue, resp *svcsdk.DescribeJobQueuesOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}

	switch pointer.StringValue(resp.JobQueues[0].Status) {
	case svcsdk.JQStatusCreating:
		cr.SetConditions(xpv1.Creating())
	case svcsdk.JQStatusDeleting:
		cr.SetConditions(xpv1.Deleting())
	case svcsdk.JQStatusValid:
		cr.SetConditions(xpv1.Available())
	case svcsdk.JQStatusInvalid:
		cr.SetConditions(xpv1.Unavailable().WithMessage(pointer.StringValue(resp.JobQueues[0].StatusReason)))
	case svcsdk.JQStatusUpdating:
		cr.SetConditions(xpv1.Unavailable().WithMessage(svcsdk.JQStatusUpdating + " " + pointer.StringValue(resp.JobQueues[0].StatusReason)))
		// Prevent Update() call during update status - which will fail.
		obs.ResourceUpToDate = true
	}

	return obs, nil
}

func preUpdate(_ context.Context, cr *svcapitypes.JobQueue, obj *svcsdk.UpdateJobQueueInput) error {
	obj.JobQueue = pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))
	obj.State = cr.Spec.ForProvider.DesiredState

	for i := range cr.Spec.ForProvider.ComputeEnvironmentOrder {
		if awsarn.IsARN(cr.Spec.ForProvider.ComputeEnvironmentOrder[i].ComputeEnvironment) {
			obj.ComputeEnvironmentOrder = append(obj.ComputeEnvironmentOrder, &svcsdk.ComputeEnvironmentOrder{
				ComputeEnvironment: pointer.ToOrNilIfZeroValue(cr.Spec.ForProvider.ComputeEnvironmentOrder[i].ComputeEnvironment),
				Order:              &cr.Spec.ForProvider.ComputeEnvironmentOrder[i].Order,
			})
		} else {
			return errors.New(errComputeEnvironmentARN)
		}
	}

	return nil
}

func (e *hooks) postUpdate(ctx context.Context, cr *svcapitypes.JobQueue, obj *svcsdk.UpdateJobQueueOutput, upd managed.ExternalUpdate, err error) (managed.ExternalUpdate, error) {
	if err != nil {
		return managed.ExternalUpdate{}, err
	}

	return upd, svcutils.UpdateTagsForResource(ctx, e.client, cr.Spec.ForProvider.Tags, obj.JobQueueArn)
}

func (e *hooks) preCreate(_ context.Context, cr *svcapitypes.JobQueue, obj *svcsdk.CreateJobQueueInput) error {
	obj.JobQueueName = pointer.ToOrNilIfZeroValue(cr.Name)
	obj.State = cr.Spec.ForProvider.DesiredState

	for i := range cr.Spec.ForProvider.ComputeEnvironmentOrder {
		if awsarn.IsARN(cr.Spec.ForProvider.ComputeEnvironmentOrder[i].ComputeEnvironment) {
			obj.ComputeEnvironmentOrder = append(obj.ComputeEnvironmentOrder, &svcsdk.ComputeEnvironmentOrder{
				ComputeEnvironment: pointer.ToOrNilIfZeroValue(cr.Spec.ForProvider.ComputeEnvironmentOrder[i].ComputeEnvironment),
				Order:              &cr.Spec.ForProvider.ComputeEnvironmentOrder[i].Order,
			})
		} else {
			return errors.New(errComputeEnvironmentARN)
		}
	}
	return nil
}

func (e *hooks) preDelete(ctx context.Context, cr *svcapitypes.JobQueue, obj *svcsdk.DeleteJobQueueInput) (bool, error) {
	obj.JobQueue = pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr))

	// Skip Deletion if JQ is updating or already deleting
	if pointer.StringValue(cr.Status.AtProvider.Status) == svcsdk.JQStatusUpdating ||
		pointer.StringValue(cr.Status.AtProvider.Status) == svcsdk.JQStatusDeleting {
		return true, nil
	}
	// JQ needs to be DISABLED to be able to be deleted
	// If the JQ is already or finally DISABLED, we are done here and the controller can request the deletion of the JQ
	if pointer.StringValue(cr.Status.AtProvider.State) == svcsdk.JQStateDisabled {
		return false, nil
	}
	// Update the JQ to set the state to DISABLED
	_, err := e.client.UpdateJobQueueWithContext(ctx, &svcsdk.UpdateJobQueueInput{
		JobQueue: pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr)),
		State:    pointer.ToOrNilIfZeroValue(svcsdk.JQStateDisabled)})
	return true, errorutils.Wrap(err, errUpdate)
}

func isUpToDate(_ context.Context, cr *svcapitypes.JobQueue, obj *svcsdk.DescribeJobQueuesOutput) (bool, string, error) {
	status := pointer.StringValue(cr.Status.AtProvider.Status)

	// Skip when updating, deleting or creating
	if status == svcsdk.JQStatusUpdating || status == svcsdk.JQStatusDeleting || status == svcsdk.JQStatusCreating {
		return true, "", nil
	}

	if pointer.StringValue(cr.Spec.ForProvider.DesiredState) != pointer.StringValue(obj.JobQueues[0].State) {
		return false, "", nil
	}

	currentParams := GenerateJobQueue(obj).Spec.ForProvider

	// Need to set the CustomComputeEnvironmentOrder by translating from the ComputeEnvironmentOrder to be able to compare with specs
	for i := range obj.JobQueues[0].ComputeEnvironmentOrder {
		if obj.JobQueues[0].ComputeEnvironmentOrder[i].ComputeEnvironment != nil {
			currentParams.ComputeEnvironmentOrder = append(currentParams.ComputeEnvironmentOrder, svcapitypes.CustomComputeEnvironmentOrder{
				ComputeEnvironment: pointer.StringValue(obj.JobQueues[0].ComputeEnvironmentOrder[i].ComputeEnvironment),
				Order:              pointer.Int64Value(obj.JobQueues[0].ComputeEnvironmentOrder[i].Order),
			})
		}
	}

	diff := cmp.Diff(cr.Spec.ForProvider, currentParams, cmpopts.EquateEmpty(),
		cmpopts.IgnoreTypes(&xpv1.Reference{}, &xpv1.Selector{}, []xpv1.Reference{}),
		cmpopts.IgnoreFields(svcapitypes.JobQueueParameters{}, "Region", "DesiredState"))

	return diff == "", diff, nil
}

func lateInitialize(spec *svcapitypes.JobQueueParameters, resp *svcsdk.DescribeJobQueuesOutput) error {
	jq := resp.JobQueues[0]
	spec.DesiredState = pointer.LateInitialize(spec.DesiredState, jq.State)

	return nil
}
