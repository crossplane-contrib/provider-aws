package cluster

import (
	"context"

	svcsdk "github.com/aws/aws-sdk-go/service/dax"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/google/go-cmp/cmp"
	ctrl "sigs.k8s.io/controller-runtime"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/dax/v1alpha1"
	awsclients "github.com/crossplane-contrib/provider-aws/pkg/clients"
)

// SetupCluster adds a controller that reconciles Cluster.
func SetupCluster(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(svcapitypes.ClusterGroupKind)
	opts := []option{
		func(e *external) {
			e.preObserve = preObserve
			e.postObserve = postObserve
			e.preCreate = preCreate
			e.preUpdate = preUpdate
			e.preDelete = preDelete
			e.lateInitialize = lateInitialize
			e.isUpToDate = isUpToDate
		},
	}
	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		For(&svcapitypes.Cluster{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(svcapitypes.ClusterGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), opts: opts}),
			managed.WithInitializers(),
			managed.WithPollInterval(o.PollInterval),
			managed.WithLogger(o.Logger.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

func preObserve(_ context.Context, cr *svcapitypes.Cluster, obj *svcsdk.DescribeClustersInput) error {
	obj.ClusterNames = append(obj.ClusterNames, awsclients.String(meta.GetExternalName(cr)))
	return nil
}

func postObserve(_ context.Context, cr *svcapitypes.Cluster, resp *svcsdk.DescribeClustersOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}
	switch awsclients.StringValue(resp.Clusters[0].Status) {
	case "available", "modifying":
		cr.SetConditions(xpv1.Available())
	case "creating":
		cr.SetConditions(xpv1.Creating())
	case "deleting", "stopped", "stopping":
		cr.SetConditions(xpv1.Unavailable())
	}
	return obs, nil
}

func preCreate(_ context.Context, cr *svcapitypes.Cluster, obj *svcsdk.CreateClusterInput) error {
	meta.SetExternalName(cr, cr.Name)
	obj.ClusterName = awsclients.String(meta.GetExternalName(cr))
	obj.IamRoleArn = cr.Spec.ForProvider.IAMRoleARN
	obj.ParameterGroupName = cr.Spec.ForProvider.ParameterGroupName
	obj.SubnetGroupName = cr.Spec.ForProvider.SubnetGroupName
	obj.SecurityGroupIds = append(obj.SecurityGroupIds, cr.Spec.ForProvider.SecurityGroupIDs...)
	return nil
}

func preUpdate(_ context.Context, cr *svcapitypes.Cluster, obj *svcsdk.UpdateClusterInput) error {
	obj.ClusterName = awsclients.String(meta.GetExternalName(cr))
	if cr.Spec.ForProvider.ParameterGroupName != nil {
		obj.ParameterGroupName = awsclients.String(*cr.Spec.ForProvider.ParameterGroupName)
	}
	if cr.Spec.ForProvider.SecurityGroupIDs != nil {
		for _, s := range cr.Spec.ForProvider.SecurityGroupIDs {
			obj.SecurityGroupIds = append(obj.SecurityGroupIds, awsclients.String(*s))
		}
	}
	return nil
}

func preDelete(_ context.Context, cr *svcapitypes.Cluster, obj *svcsdk.DeleteClusterInput) (bool, error) {
	obj.ClusterName = awsclients.String(meta.GetExternalName(cr))
	return false, nil
}

func lateInitialize(in *svcapitypes.ClusterParameters, out *svcsdk.DescribeClustersOutput) error {
	c := out.Clusters[0]
	if in.AvailabilityZones == nil {
		in.AvailabilityZones = make([]*string, len(c.Nodes))
		for i, group := range c.Nodes {
			in.AvailabilityZones[i] = group.AvailabilityZone
		}
	}
	in.ClusterEndpointEncryptionType = awsclients.LateInitializeStringPtr(in.ClusterEndpointEncryptionType, c.ClusterEndpointEncryptionType)
	in.PreferredMaintenanceWindow = awsclients.LateInitializeStringPtr(in.PreferredMaintenanceWindow, c.PreferredMaintenanceWindow)
	if in.SecurityGroupIDs == nil {
		in.SecurityGroupIDs = make([]*string, len(c.SecurityGroups))
		for i, group := range c.SecurityGroups {
			in.SecurityGroupIDs[i] = group.SecurityGroupIdentifier
		}
	}
	return nil
}

func isUpToDate(cr *svcapitypes.Cluster, output *svcsdk.DescribeClustersOutput) (bool, error) {
	in := cr.Spec.ForProvider
	out := output.Clusters[0]

	notUpToDate := isNotUpToDate(in, out)

	if notUpToDate {
		return false, nil
	}

	parameterGroupNotEqualNotNil := isUpToDateParameterGroup(in, out)

	if parameterGroupNotEqualNotNil {
		return false, nil
	}

	notificationTopicArnNotEqualNotNil := isUpToDateNotificationTopicArn(in, out)

	if notificationTopicArnNotEqualNotNil {
		return false, nil
	}

	outSecurityGroupIds := getOutSecurityIds(out)
	if !cmp.Equal(in.SecurityGroupIDs, outSecurityGroupIds) {
		return false, nil
	}

	outAvailabilityZones := getOutAvailabilityZones(out)
	if !cmp.Equal(in.AvailabilityZones, outAvailabilityZones) {
		return false, nil
	}

	return true, nil
}

func isNotUpToDate(in svcapitypes.ClusterParameters, out *svcsdk.Cluster) (unequal bool) {
	if !cmp.Equal(in.Description, out.Description) {
		return true
	}

	if !cmp.Equal(in.IAMRoleARN, out.IamRoleArn) {
		return true
	}

	if !cmp.Equal(in.NodeType, out.NodeType) {
		return true
	}

	if !cmp.Equal(in.ClusterEndpointEncryptionType, out.ClusterEndpointEncryptionType) {
		return true
	}

	if !cmp.Equal(in.SubnetGroupName, out.SubnetGroup) {
		return true
	}

	if !cmp.Equal(in.PreferredMaintenanceWindow, out.PreferredMaintenanceWindow) {
		return true
	}
	return false
}

func isUpToDateNotificationTopicArn(in svcapitypes.ClusterParameters, out *svcsdk.Cluster) (equalNotNil bool) {
	if out.NotificationConfiguration == nil {
		return false
	}
	if !cmp.Equal(in.NotificationTopicARN, out.NotificationConfiguration.TopicArn) {
		return true
	}
	return false
}

func isUpToDateParameterGroup(in svcapitypes.ClusterParameters, out *svcsdk.Cluster) (equalNotNil bool) {
	if out.ParameterGroup == nil {
		return false
	}
	if !cmp.Equal(in.ParameterGroupName, out.ParameterGroup.ParameterGroupName) {
		return true
	}
	return false
}

func getOutSecurityIds(out *svcsdk.Cluster) []*string {
	outSecurityGroupIds := make([]*string, len(out.SecurityGroups))
	if len(out.SecurityGroups) > 0 {
		for i, outSecurityGroupID := range out.SecurityGroups {
			outSecurityGroupIds[i] = outSecurityGroupID.SecurityGroupIdentifier
		}
	}
	return outSecurityGroupIds
}

func getOutAvailabilityZones(out *svcsdk.Cluster) []*string {
	outAvailabilityZones := make([]*string, len(out.Nodes))
	if len(out.Nodes) > 0 {
		for i, node := range out.Nodes {
			outAvailabilityZones[i] = node.AvailabilityZone
		}
	}
	return outAvailabilityZones
}
