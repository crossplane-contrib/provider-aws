package autoscalinggroup

import (
	"context"
	"fmt"
	"sort"

	"github.com/aws/aws-sdk-go/aws"
	svcsdk "github.com/aws/aws-sdk-go/service/autoscaling"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/google/go-cmp/cmp"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/autoscaling/v1beta1"
	"github.com/crossplane-contrib/provider-aws/apis/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
	custommanaged "github.com/crossplane-contrib/provider-aws/pkg/utils/reconciler/managed"
)

// SetupAutoScalingGroup adds a controller that reconciles AutoScalingGroup.
func SetupAutoScalingGroup(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(svcapitypes.AutoScalingGroupGroupKind)
	opts := []option{
		func(e *external) {
			e.isUpToDate = isUpToDate
			e.lateInitialize = lateInitialize
			e.preObserve = preObserve
			e.postObserve = postObserve
			e.preCreate = preCreate
			e.preUpdate = preUpdate
			e.preDelete = preDelete
		},
	}
	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), v1alpha1.StoreConfigGroupVersionKind))
	}

	reconcilerOpts := []managed.ReconcilerOption{
		managed.WithCriticalAnnotationUpdater(custommanaged.NewRetryingCriticalAnnotationUpdater(mgr.GetClient())),
		managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), opts: opts}),
		managed.WithPollInterval(o.PollInterval),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(
			name))),
		managed.WithConnectionPublishers(cps...),
	}

	if o.Features.Enabled(features.EnableAlphaManagementPolicies) {
		reconcilerOpts = append(reconcilerOpts, managed.WithManagementPolicies())
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(svcapitypes.AutoScalingGroupGroupVersionKind),
		reconcilerOpts...)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&svcapitypes.AutoScalingGroup{}).
		Complete(r)
}

func isUpToDate(_ context.Context, obj *svcapitypes.AutoScalingGroup, obs *svcsdk.DescribeAutoScalingGroupsOutput) (bool, string, error) { //nolint:gocyclo
	in := obj.Spec.ForProvider
	asg := obs.AutoScalingGroups[0]

	if !cmp.Equal(in.CapacityRebalance, asg.CapacityRebalance) {
		return false, "spec.forProvider.capacityRebalance", nil
	}
	// DefaultInstanceWarmup can be updated
	if !cmp.Equal(in.DefaultInstanceWarmup, asg.DefaultInstanceWarmup) {
		return false, "spec.forProvider.defaultInstanceWarmup", nil
	}
	// DesiredCapacityType can be updated
	if !cmp.Equal(in.DesiredCapacityType, asg.DesiredCapacityType) {
		return false, "spec.forProvider.desiredCapacityType", nil
	}
	// Context is reserved
	// if !cmp.Equal(*in.Context, *asg.Context) {
	// 	return false, nil
	// }
	// DefaultCooldown can be updated
	if !cmp.Equal(in.DefaultCooldown, asg.DefaultCooldown) {
		return false, "spec.forProvider.defaultCooldown", nil
	}
	// DesiredCapacity can be updated
	if !cmp.Equal(in.DesiredCapacity, asg.DesiredCapacity) {
		return false, "spec.forProvider.desiredCapacity", nil
	}
	// HealthCheckGracePeriod can be updated
	if !cmp.Equal(in.HealthCheckGracePeriod, asg.HealthCheckGracePeriod) {
		return false, "spec.forProvider.healthCheckGracePeriod", nil
	}
	// HealthCheckType can be updated
	if !cmp.Equal(in.HealthCheckType, asg.HealthCheckType) {
		return false, "spec.forProvider.healthCheckType", nil
	}
	// MaxInstanceLifetime can be updated
	if !cmp.Equal(in.MaxInstanceLifetime, asg.MaxInstanceLifetime) {
		return false, "spec.forProvider.maxInstanceLifetime", nil
	}
	// MaxSize can be updated
	if !cmp.Equal(in.MaxSize, asg.MaxSize) {
		return false, "spec.forProvider.maxSize", nil
	}
	// MinSize can be updated
	if !cmp.Equal(in.MinSize, asg.MinSize) {
		return false, "spec.forProvider.minSize", nil
	}
	// NewInstancesProtectedFromScaleIn can be updated
	if !cmp.Equal(in.NewInstancesProtectedFromScaleIn, asg.NewInstancesProtectedFromScaleIn) {
		return false, "spec.forProvider.newInstancesProtectedFromScaleIn", nil
	}
	if !cmp.Equal(in.PlacementGroup, asg.PlacementGroup) {
		return false, "spec.forProvider.placementGroup", nil
	}
	// VPCZoneIdentifier can be updated
	if !cmp.Equal(in.VPCZoneIdentifier, asg.VPCZoneIdentifier) {
		return false, "spec.forProvider.vpcZoneIdentifier", nil
	}
	// LaunchTemplate can be updated
	if in.LaunchTemplate != nil && asg.LaunchTemplate != nil {
		if in.LaunchTemplate.LaunchTemplateID != nil {
			if !cmp.Equal(in.LaunchTemplate.LaunchTemplateID, asg.LaunchTemplate.LaunchTemplateId) {
				return false, "spec.forProvider.launchTemplate.launchTemplateID", nil
			}
		}
		if in.LaunchTemplate.LaunchTemplateName != nil {
			if !cmp.Equal(in.LaunchTemplate.LaunchTemplateName, asg.LaunchTemplate.LaunchTemplateName) {
				return false, "spec.forProvider.launchTemplate.launchTemplateName", nil
			}
		}
		if in.LaunchTemplate.Version != nil {
			if !cmp.Equal(in.LaunchTemplate.Version, asg.LaunchTemplate.Version) {
				return false, "spec.forProvider.launchTemplate.version", nil
			}
		}
	}
	// MixedInstancesPolicy can be updated
	if in.MixedInstancesPolicy != nil && asg.MixedInstancesPolicy != nil {
		if in.MixedInstancesPolicy.InstancesDistribution != nil && asg.MixedInstancesPolicy.InstancesDistribution != nil {
			if !cmp.Equal(in.MixedInstancesPolicy.InstancesDistribution, asg.MixedInstancesPolicy.InstancesDistribution) {
				return false, "spec.forProvider.mixedInstancesPolicy.instancesDistribution", nil
			}
		}
		if in.MixedInstancesPolicy.LaunchTemplate != nil && asg.MixedInstancesPolicy.LaunchTemplate != nil {
			if !cmp.Equal(in.MixedInstancesPolicy.LaunchTemplate.LaunchTemplateSpecification, asg.MixedInstancesPolicy.LaunchTemplate.LaunchTemplateSpecification) {
				return false, "spec.forProvider.mixedInstancesPolicy.launchTemplate.launchTemplateSpecification", nil
			}
			if !cmp.Equal(in.MixedInstancesPolicy.LaunchTemplate.Overrides, asg.MixedInstancesPolicy.LaunchTemplate.Overrides) {
				return false, "spec.forProvider.mixedInstancesPolicy.launchTemplate.overrides", nil
			}
		}
	}
	// AvailabilityZones can be updated
	sort.Slice(in.AvailabilityZones, func(i, j int) bool {
		return *in.AvailabilityZones[i] < *in.AvailabilityZones[j]
	})
	sort.Slice(asg.AvailabilityZones, func(i, j int) bool {
		return *asg.AvailabilityZones[i] < *asg.AvailabilityZones[j]
	})
	if !cmp.Equal(in.AvailabilityZones, asg.AvailabilityZones) {
		return false, "spec.forProvider.availabilityZones", nil
	}
	// LoadBalancerNames can be updated
	sort.Slice(in.LoadBalancerNames, func(i, j int) bool {
		return *in.LoadBalancerNames[i] < *in.LoadBalancerNames[j]
	})
	sort.Slice(asg.LoadBalancerNames, func(i, j int) bool {
		return *asg.LoadBalancerNames[i] < *asg.LoadBalancerNames[j]
	})
	if !cmp.Equal(in.LoadBalancerNames, asg.LoadBalancerNames) {
		return false, "spec.forProvider.loadBalancerNames", nil
	}
	// Tags can be updated
	sort.Slice(in.Tags, func(i, j int) bool {
		return *in.Tags[i].Key < *in.Tags[j].Key
	})
	sort.Slice(asg.Tags, func(i, j int) bool {
		return *asg.Tags[i].Key < *asg.Tags[j].Key
	})
	if len(in.Tags) != len(asg.Tags) {
		return false, "spec.forProvider.tags", nil
	}
	for i := range in.Tags {
		a := in.Tags[i]
		b := asg.Tags[i]
		if !ptr.Equal(a.Key, b.Key) {
			return false, fmt.Sprintf("spec.forProvider.tags[%d].key", i), nil
		}
		if !ptr.Equal(a.Value, b.Value) {
			return false, fmt.Sprintf("spec.forProvider.tags[%d].value", i), nil
		}
		if !ptr.Equal(a.PropagateAtLaunch, b.PropagateAtLaunch) {
			return false, fmt.Sprintf("spec.forProvider.tags[%d].propagateAtLaunch", i), nil
		}
	}

	// TargetGroupARNs
	// sort.Slice(in.TargetGroupARNs, func(i, j int) bool {
	//	return *in.TargetGroupARNs[i] < *in.TargetGroupARNs[j]
	// })
	// sort.Slice(asg.TargetGroupARNs, func(i, j int) bool {
	//	return *asg.TargetGroupARNs[i] < *asg.TargetGroupARNs[j]
	// })
	// if !cmp.Equal(in.TargetGroupARNs, asg.TargetGroupARNs) {
	//	return false, nil
	// }

	// TerminationPolicies can be updated
	sort.Slice(in.TerminationPolicies, func(i, j int) bool {
		return *in.TerminationPolicies[i] < *in.TerminationPolicies[j]
	})
	sort.Slice(asg.TerminationPolicies, func(i, j int) bool {
		return *asg.TerminationPolicies[i] < *asg.TerminationPolicies[j]
	})
	if !cmp.Equal(in.TerminationPolicies, asg.TerminationPolicies) {
		return false, "spec.forProvider.terminationPolicies", nil
	}

	// TrafficSources
	// sort.Slice(in.TrafficSources, func(i, j int) bool {
	//	return *in.TrafficSources[i].Identifier < *in.TrafficSources[j].Identifier
	// })
	// sort.Slice(asg.TrafficSources, func(i, j int) bool {
	//	return *asg.TrafficSources[i].Identifier < *asg.TrafficSources[j].Identifier
	// })
	// if !cmp.Equal(in.TrafficSources, asg.TrafficSources) {
	// 	return false, nil
	// }
	return true, "", nil
}

func lateInitialize(in *svcapitypes.AutoScalingGroupParameters, asg *svcsdk.DescribeAutoScalingGroupsOutput) error {
	obs := asg.AutoScalingGroups[0]
	in.AvailabilityZones = pointer.LateInitializeSlice(in.AvailabilityZones, obs.AvailabilityZones)
	in.Context = pointer.LateInitialize(in.Context, obs.Context)
	in.CapacityRebalance = pointer.LateInitialize(in.CapacityRebalance, obs.CapacityRebalance)
	in.DefaultCooldown = pointer.LateInitialize(in.DefaultCooldown, obs.DefaultCooldown)
	in.DefaultInstanceWarmup = pointer.LateInitialize(in.DefaultInstanceWarmup, obs.DefaultInstanceWarmup)
	in.DesiredCapacity = pointer.LateInitialize(in.DesiredCapacity, obs.DesiredCapacity)
	in.DesiredCapacityType = pointer.LateInitialize(in.DesiredCapacityType, obs.DesiredCapacityType)
	in.HealthCheckGracePeriod = pointer.LateInitialize(in.HealthCheckGracePeriod, obs.HealthCheckGracePeriod)
	in.HealthCheckType = pointer.LateInitialize(in.HealthCheckType, obs.HealthCheckType)
	in.LoadBalancerNames = pointer.LateInitializeSlice(in.LoadBalancerNames, obs.LoadBalancerNames)
	in.MaxInstanceLifetime = pointer.LateInitialize(in.MaxInstanceLifetime, obs.MaxInstanceLifetime)
	in.NewInstancesProtectedFromScaleIn = pointer.LateInitialize(in.NewInstancesProtectedFromScaleIn, obs.NewInstancesProtectedFromScaleIn)
	in.PlacementGroup = pointer.LateInitialize(in.PlacementGroup, obs.PlacementGroup)
	in.ServiceLinkedRoleARN = pointer.LateInitialize(in.ServiceLinkedRoleARN, obs.ServiceLinkedRoleARN)
	in.TargetGroupARNs = pointer.LateInitializeSlice(in.TargetGroupARNs, obs.TargetGroupARNs)
	in.TerminationPolicies = pointer.LateInitializeSlice(in.TerminationPolicies, obs.TerminationPolicies)
	in.VPCZoneIdentifier = pointer.LateInitialize(in.VPCZoneIdentifier, obs.VPCZoneIdentifier)

	return nil
}

func preObserve(_ context.Context, cr *svcapitypes.AutoScalingGroup, obj *svcsdk.DescribeAutoScalingGroupsInput) error {
	obj.AutoScalingGroupNames = append(obj.AutoScalingGroupNames, aws.String(meta.GetExternalName(cr)))
	return nil
}

func postObserve(_ context.Context, cr *svcapitypes.AutoScalingGroup, resp *svcsdk.DescribeAutoScalingGroupsOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}
	cr.SetConditions(xpv1.Available())
	// cr.Status.AtProvider.CreatedTime = fromTimePtr(resp.AutoScalingGroups[0].CreatedTime)
	return obs, nil
}

func preCreate(_ context.Context, cr *svcapitypes.AutoScalingGroup, obj *svcsdk.CreateAutoScalingGroupInput) error {
	obj.AutoScalingGroupName = aws.String(meta.GetExternalName(cr))
	return nil
}

func preUpdate(_ context.Context, cr *svcapitypes.AutoScalingGroup, obj *svcsdk.UpdateAutoScalingGroupInput) error {
	obj.AutoScalingGroupName = aws.String(meta.GetExternalName(cr))
	return nil
}

func preDelete(_ context.Context, cr *svcapitypes.AutoScalingGroup, obj *svcsdk.DeleteAutoScalingGroupInput) (bool, error) {
	obj.AutoScalingGroupName = aws.String(meta.GetExternalName(cr))
	f := true
	obj.ForceDelete = &f
	return false, nil
}

// fromTimePtr probably not needed if metav1 import issue in zz_conversions.go is fixed
// see https://github.com/aws-controllers-k8s/community/issues/1372

// fromTimePtr is a helper for converting a *time.Time to a *metav1.Time
// func fromTimePtr(t *time.Time) *metav1.Time {
//	if t != nil {
//		m := metav1.NewTime(*t)
//		return &m
//	}
//	return nil
// }
