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

package computeenvironment

import (
	"context"
	"fmt"

	svcsdk "github.com/aws/aws-sdk-go/service/batch"
	"github.com/aws/aws-sdk-go/service/batch/batchiface"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	ctrl "sigs.k8s.io/controller-runtime"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/batch/v1alpha1"
	awsclients "github.com/crossplane-contrib/provider-aws/pkg/clients"
	svcutils "github.com/crossplane-contrib/provider-aws/pkg/controller/batch"
)

// SetupComputeEnvironment adds a controller that reconciles a ComputeEnvironment.
func SetupComputeEnvironment(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(svcapitypes.ComputeEnvironmentGroupKind)
	opts := []option{
		func(e *external) {
			h := &hooks{client: e.client}
			e.preObserve = preObserve
			e.postObserve = postObserve
			e.lateInitialize = lateInitialize
			e.isUpToDate = isUpToDate
			e.preUpdate = preUpdate
			e.postUpdate = h.postUpdate
			e.preCreate = preCreate
			e.preDelete = h.preDelete
		},
	}

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&svcapitypes.ComputeEnvironment{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(svcapitypes.ComputeEnvironmentGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), opts: opts}),
			managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
			managed.WithInitializers(
				managed.NewDefaultProviderConfig(mgr.GetClient()),
				managed.NewNameAsExternalName(mgr.GetClient())),
			managed.WithPollInterval(o.PollInterval),
			managed.WithLogger(o.Logger.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

type hooks struct {
	client batchiface.BatchAPI
}

func preObserve(_ context.Context, cr *svcapitypes.ComputeEnvironment, obj *svcsdk.DescribeComputeEnvironmentsInput) error {
	obj.ComputeEnvironments = []*string{awsclients.String(meta.GetExternalName(cr))} // we only want to observe our CE
	return nil
}

func postObserve(_ context.Context, cr *svcapitypes.ComputeEnvironment, resp *svcsdk.DescribeComputeEnvironmentsOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}

	cr.Status.AtProvider.ECSClusterARN = resp.ComputeEnvironments[0].EcsClusterArn

	switch awsclients.StringValue(resp.ComputeEnvironments[0].Status) {
	case svcsdk.CEStatusCreating:
		cr.SetConditions(xpv1.Creating())
	case svcsdk.CEStatusDeleting:
		cr.SetConditions(xpv1.Deleting())
	case svcsdk.CEStatusValid:
		cr.SetConditions(xpv1.Available())
	case svcsdk.CEStatusInvalid:
		cr.SetConditions(xpv1.Unavailable().WithMessage(awsclients.StringValue(resp.ComputeEnvironments[0].StatusReason)))
	case svcsdk.CEStatusUpdating:
		cr.SetConditions(xpv1.Unavailable().WithMessage(svcsdk.CEStatusUpdating + " " + awsclients.StringValue(resp.ComputeEnvironments[0].StatusReason)))
		// Prevent Update() call during update status - which will fail.
		obs.ResourceUpToDate = true
	}

	fmt.Printf("cr.Status.ConditionedStatus: %v\n", cr.Status.ConditionedStatus)
	return obs, nil
}

// nolint:gocyclo
func preUpdate(_ context.Context, cr *svcapitypes.ComputeEnvironment, obj *svcsdk.UpdateComputeEnvironmentInput) error {
	obj.ComputeEnvironment = awsclients.String(meta.GetExternalName(cr))
	obj.ServiceRole = cr.Spec.ForProvider.ServiceRoleARN
	obj.State = cr.Spec.ForProvider.DesiredState

	if obj.ComputeResources != nil {
		obj.ComputeResources.Subnets = cr.Spec.ForProvider.SubnetIDs
		obj.ComputeResources.SecurityGroupIds = cr.Spec.ForProvider.SecurityGroupIDs

		// MANAGED EC2 or SPOT CEs: ComputeResources-update-call does not accept SecurityGroupIds and Subnets
		// when Allocation Strategy is nil or BEST_FIT
		if awsclients.StringValue(cr.Spec.ForProvider.ComputeResources.Type) == string(svcapitypes.CRType_EC2) ||
			awsclients.StringValue(cr.Spec.ForProvider.ComputeResources.Type) == string(svcapitypes.CRType_SPOT) {
			obj.ComputeResources.SecurityGroupIds = nil
			obj.ComputeResources.Subnets = nil
		}

		// fields that can be updated for CE only with Allocation
		// Strategy BEST_FIT_PROGRESSIVE and SPOT_CAPACITY_OPTIMIZED
		if awsclients.StringValue(cr.Spec.ForProvider.ComputeResources.AllocationStrategy) == string(svcapitypes.CRUpdateAllocationStrategy_BEST_FIT_PROGRESSIVE) ||
			awsclients.StringValue(cr.Spec.ForProvider.ComputeResources.AllocationStrategy) == string(svcapitypes.CRUpdateAllocationStrategy_SPOT_CAPACITY_OPTIMIZED) {

			obj.ComputeResources.AllocationStrategy = cr.Spec.ForProvider.ComputeResources.AllocationStrategy
			obj.ComputeResources.BidPercentage = cr.Spec.ForProvider.ComputeResources.BidPercentage
			if cr.Spec.ForProvider.ComputeResources.EC2Configuration != nil {
				updateConfig := []*svcsdk.Ec2Configuration{}
				for _, iter := range cr.Spec.ForProvider.ComputeResources.EC2Configuration {
					ceConfig := &svcsdk.Ec2Configuration{}
					if iter.ImageIDOverride != nil {
						ceConfig.ImageIdOverride = iter.ImageIDOverride
					}
					if iter.ImageType != nil {
						ceConfig.ImageType = iter.ImageType
					}
					updateConfig = append(updateConfig, ceConfig)
				}
				obj.ComputeResources.Ec2Configuration = updateConfig
			}
			obj.ComputeResources.Ec2KeyPair = cr.Spec.ForProvider.ComputeResources.EC2KeyPair
			obj.ComputeResources.InstanceTypes = cr.Spec.ForProvider.ComputeResources.InstanceTypes
			if cr.Spec.ForProvider.ComputeResources.LaunchTemplate != nil {
				updateLaunchTemplate := &svcsdk.LaunchTemplateSpecification{
					LaunchTemplateId:   cr.Spec.ForProvider.ComputeResources.LaunchTemplate.LaunchTemplateID,
					LaunchTemplateName: cr.Spec.ForProvider.ComputeResources.LaunchTemplate.LaunchTemplateName,
					Version:            cr.Spec.ForProvider.ComputeResources.LaunchTemplate.Version,
				}
				obj.ComputeResources.LaunchTemplate = updateLaunchTemplate
			}
			obj.ComputeResources.PlacementGroup = cr.Spec.ForProvider.ComputeResources.PlacementGroup
			obj.ComputeResources.Subnets = cr.Spec.ForProvider.SubnetIDs
			obj.ComputeResources.SecurityGroupIds = cr.Spec.ForProvider.SecurityGroupIDs
			obj.ComputeResources.Tags = cr.Spec.ForProvider.ComputeResources.Tags
			obj.ComputeResources.Type = cr.Spec.ForProvider.ComputeResources.Type
			obj.ComputeResources.UpdateToLatestImageVersion = cr.Spec.ForProvider.UpdateToLatestImageVersion
			if cr.Spec.ForProvider.UpdatePolicy != nil {
				updatePolicy := &svcsdk.UpdatePolicy{
					JobExecutionTimeoutMinutes: cr.Spec.ForProvider.UpdatePolicy.JobExecutionTimeoutMinutes,
					TerminateJobsOnUpdate:      cr.Spec.ForProvider.UpdatePolicy.TerminateJobsOnUpdate,
				}

				obj.UpdatePolicy = updatePolicy
			}
		}

	}

	return nil
}

func (e *hooks) postUpdate(ctx context.Context, cr *svcapitypes.ComputeEnvironment, obj *svcsdk.UpdateComputeEnvironmentOutput, upd managed.ExternalUpdate, err error) (managed.ExternalUpdate, error) {
	if err != nil {
		return managed.ExternalUpdate{}, err
	}

	return upd, svcutils.UpdateTagsForResource(ctx, e.client, cr.Spec.ForProvider.Tags, obj.ComputeEnvironmentArn)
}

func preCreate(_ context.Context, cr *svcapitypes.ComputeEnvironment, obj *svcsdk.CreateComputeEnvironmentInput) error {
	obj.ComputeEnvironmentName = awsclients.String(cr.Name)
	obj.ServiceRole = cr.Spec.ForProvider.ServiceRoleARN

	if obj.ComputeResources != nil {
		obj.ComputeResources.SecurityGroupIds = cr.Spec.ForProvider.SecurityGroupIDs
		obj.ComputeResources.Subnets = cr.Spec.ForProvider.SubnetIDs
		obj.ComputeResources.InstanceRole = cr.Spec.ForProvider.InstanceRole
		obj.ComputeResources.SpotIamFleetRole = cr.Spec.ForProvider.SpotIAMFleetRole
	}

	return nil
}

func (e *hooks) preDelete(ctx context.Context, cr *svcapitypes.ComputeEnvironment, obj *svcsdk.DeleteComputeEnvironmentInput) (bool, error) {
	obj.ComputeEnvironment = awsclients.String(meta.GetExternalName(cr))

	// Skip Deletion if CE is updating or already deleting
	if awsclients.StringValue(cr.Status.AtProvider.Status) == svcsdk.CEStatusUpdating ||
		awsclients.StringValue(cr.Status.AtProvider.Status) == svcsdk.CEStatusDeleting {
		return true, nil
	}

	// CE state needs to be DISABLED to be able to be deleted
	// If the CE is already or finally DISABLED, we are done here and
	// the controller can request the deletion of the CE
	if awsclients.StringValue(cr.Status.AtProvider.State) == svcsdk.CEStateDisabled {
		return false, nil
	}
	// Update the CE to set the state to DISABLED
	_, err := e.client.UpdateComputeEnvironmentWithContext(ctx, &svcsdk.UpdateComputeEnvironmentInput{
		ComputeEnvironment: awsclients.String(meta.GetExternalName(cr)),
		State:              awsclients.String(svcsdk.CEStateDisabled)})
	return true, awsclients.Wrap(err, errUpdate)

}

func isUpToDate(cr *svcapitypes.ComputeEnvironment, obj *svcsdk.DescribeComputeEnvironmentsOutput) (bool, error) {

	status := awsclients.StringValue(cr.Status.AtProvider.Status)
	ce := obj.ComputeEnvironments[0]
	spec := cr.Spec.ForProvider

	// Skip when updating, deleting or creating
	if status == svcsdk.CEStatusUpdating || status == svcsdk.CEStatusDeleting || status == svcsdk.CEStatusCreating {
		return true, nil
	}

	currentParams := GenerateComputeEnvironment(obj).Spec.ForProvider

	if awsclients.StringValue(cr.Spec.ForProvider.Type) == string(svcapitypes.CEType_MANAGED) {

		switch {
		case !cmp.Equal(spec.SubnetIDs, ce.ComputeResources.Subnets),
			!cmp.Equal(spec.SecurityGroupIDs, ce.ComputeResources.SecurityGroupIds):
			return false, nil
		}

		// fields that can be updated for CE only with Allocation
		// Strategy BEST_FIT_PROGRESSIVE and SPOT_CAPACITY_OPTIMIZED
		if awsclients.StringValue(ce.ComputeResources.AllocationStrategy) == string(svcapitypes.CRUpdateAllocationStrategy_BEST_FIT_PROGRESSIVE) ||
			awsclients.StringValue(ce.ComputeResources.AllocationStrategy) == string(svcapitypes.CRUpdateAllocationStrategy_SPOT_CAPACITY_OPTIMIZED) {

			// for instance role profile ARN and name is possible,
			// however AWS seems to always give userinput back, so simple check is fine
			switch {
			case !cmp.Equal(spec.ComputeResources, currentParams.ComputeResources, cmpopts.EquateEmpty()),
				awsclients.StringValue(spec.InstanceRole) != awsclients.StringValue(ce.ComputeResources.InstanceRole),
				!areUpdatePolicyEqual(spec.UpdatePolicy, ce.UpdatePolicy):
				return false, nil
			}

		}
	}

	switch {
	case awsclients.StringValue(spec.DesiredState) != awsclients.StringValue(ce.State),
		awsclients.StringValue(spec.ServiceRoleARN) != awsclients.StringValue(ce.ServiceRole),
		!cmp.Equal(spec, currentParams, cmpopts.EquateEmpty(),
			cmpopts.IgnoreTypes(&xpv1.Reference{}, &xpv1.Selector{}, []xpv1.Reference{}),
			cmpopts.IgnoreFields(svcapitypes.ComputeEnvironmentParameters{}, "Region", "Type", "InstanceRole", "SpotIAMFleetRole",
				"UpdatePolicy", "UpdateToLatestImageVersion", "SubnetIDs", "SecurityGroupIDs", "ServiceRoleARN", "DesiredState"),
			cmpopts.IgnoreFields(svcapitypes.ComputeResource{}, "AllocationStrategy", "BidPercentage", "EC2Configuration", "EC2KeyPair",
				"InstanceTypes", "LaunchTemplate", "PlacementGroup", "Tags", "Type")):
		return false, nil
	}

	return true, nil
}

func areUpdatePolicyEqual(spec *svcapitypes.UpdatePolicy, current *svcsdk.UpdatePolicy) bool {

	if spec != nil {
		if current == nil {
			return false
		}
		switch {
		case awsclients.Int64Value(spec.JobExecutionTimeoutMinutes) != awsclients.Int64Value(current.JobExecutionTimeoutMinutes),
			awsclients.BoolValue(spec.TerminateJobsOnUpdate) != awsclients.BoolValue(current.TerminateJobsOnUpdate):
			return false
		}
	}
	return true
}

func lateInitialize(spec *svcapitypes.ComputeEnvironmentParameters, resp *svcsdk.DescribeComputeEnvironmentsOutput) error {

	ce := resp.ComputeEnvironments[0]

	spec.DesiredState = awsclients.LateInitializeStringPtr(spec.DesiredState, ce.State)
	spec.ServiceRoleARN = awsclients.LateInitializeStringPtr(spec.ServiceRoleARN, ce.ServiceRole)

	if ce.ComputeResources != nil {
		spec.ComputeResources.MinvCPUs = awsclients.LateInitializeInt64Ptr(spec.ComputeResources.MinvCPUs, ce.ComputeResources.MinvCpus)
		spec.ComputeResources.MaxvCPUs = awsclients.LateInitializeInt64Ptr(spec.ComputeResources.MaxvCPUs, ce.ComputeResources.MaxvCpus)

		if awsclients.StringValue(ce.ComputeResources.Type) == string(svcsdk.CRTypeEc2) ||
			awsclients.StringValue(ce.ComputeResources.Type) == string(svcsdk.CRTypeSpot) {

			if ce.ComputeResources.Ec2Configuration != nil && spec.ComputeResources.EC2Configuration == nil {

				specConfig := []*svcapitypes.EC2Configuration{}
				for _, iter := range ce.ComputeResources.Ec2Configuration {
					ceConfig := &svcapitypes.EC2Configuration{}
					if iter.ImageIdOverride != nil {
						ceConfig.ImageIDOverride = iter.ImageIdOverride
					}
					if iter.ImageType != nil {
						ceConfig.ImageType = iter.ImageType
					}
					specConfig = append(specConfig, ceConfig)
				}
				spec.ComputeResources.EC2Configuration = specConfig
			}
		}
	}

	return nil
}
