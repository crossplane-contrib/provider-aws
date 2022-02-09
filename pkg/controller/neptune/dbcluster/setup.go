/*
Copyright 2022 The Crossplane Authors.
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

package dbcluster

import (
	"context"
	"time"

	svcsdk "github.com/aws/aws-sdk-go/service/neptune"
	svcsdkapi "github.com/aws/aws-sdk-go/service/neptune/neptuneiface"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/ratelimiter"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	svcapitypes "github.com/crossplane/provider-aws/apis/neptune/v1alpha1"
	aws "github.com/crossplane/provider-aws/pkg/clients"
)

type dbClusterStatus string

const (
	statusAvailable dbClusterStatus = "available"
	statusCreating  dbClusterStatus = "creating"
	statusDeleted   dbClusterStatus = "deleted"
	statusUpdating  dbClusterStatus = "updating"
)

// SetupDBCluster adds a controller that reconciles DB Cluster.
func SetupDBCluster(mgr ctrl.Manager, l logging.Logger, rl workqueue.RateLimiter, poll time.Duration) error {
	name := managed.ControllerName(svcapitypes.DBClusterKind)
	opts := []option{
		func(e *external) {
			e.lateInitialize = lateInitialize
			e.isUpToDate = isUpToDate
			e.preObserve = preObserve
			e.preCreate = preCreate
			e.preDelete = preDelete
			e.postObserve = postObserve
			u := &updateClient{client: e.client}
			e.preUpdate = u.preUpdate
		},
	}
	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(controller.Options{
			RateLimiter: ratelimiter.NewController(rl),
		}).
		For(&svcapitypes.DBCluster{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(svcapitypes.DBClusterGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), opts: opts}),
			managed.WithPollInterval(poll),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

func preObserve(_ context.Context, cr *svcapitypes.DBCluster, obj *svcsdk.DescribeDBClustersInput) error {
	obj.DBClusterIdentifier = aws.String(meta.GetExternalName(cr))
	return nil
}

func preCreate(_ context.Context, cr *svcapitypes.DBCluster, obj *svcsdk.CreateDBClusterInput) error {
	obj.DBClusterIdentifier = aws.String(meta.GetExternalName(cr))
	return nil
}

type updateClient struct {
	client svcsdkapi.NeptuneAPI
}

func (e *updateClient) preUpdate(_ context.Context, cr *svcapitypes.DBCluster, mci *svcsdk.ModifyDBClusterInput) error {
	switch aws.StringValue(cr.Status.AtProvider.Status) {
	case string(statusUpdating), string(statusCreating):
		return nil
	}

	mci.DBClusterIdentifier = aws.String(meta.GetExternalName(cr))
	mci.SetApplyImmediately(*cr.Spec.ForProvider.ApplyImmediately)
	mci.SetDBClusterParameterGroupName(*cr.Spec.ForProvider.DBClusterParameterGroupName)
	mci.SetDeletionProtection(*cr.Spec.ForProvider.DeletionProtection)
	mci.SetEnableIAMDatabaseAuthentication(*cr.Spec.ForProvider.EnableIAMDatabaseAuthentication)
	mci.SetBackupRetentionPeriod(*cr.Spec.ForProvider.BackupRetentionPeriod)
	mci.SetPreferredMaintenanceWindow(*cr.Spec.ForProvider.PreferredMaintenanceWindow)
	mci.SetVpcSecurityGroupIds(cr.Spec.ForProvider.VPCSecurityGroupIDs)

	c, err := e.client.DescribeDBClusters(&svcsdk.DescribeDBClustersInput{DBClusterIdentifier: aws.String(meta.GetExternalName(cr))})
	if err != nil {
		return errors.Wrap(err, "could not describe DB Cluster")
	}

	if len(c.DBClusters) != 0 && len(c.DBClusters[0].DBClusterMembers) != 0 &&
		c.DBClusters[0].DBClusterMembers[0].DBInstanceIdentifier != nil {
		mci.SetEngineVersion(*cr.Spec.ForProvider.EngineVersion)
	}

	if len(c.DBClusters) != 0 && len(c.DBClusters[0].DBClusterMembers) != 0 &&
		c.DBClusters[0].DBClusterMembers[0].DBInstanceIdentifier != nil {
		mci.SetPort(*cr.Spec.ForProvider.Port)
	}

	cloudwatchConfig := svcsdk.CloudwatchLogsExportConfiguration{
		EnableLogTypes: cr.Spec.ForProvider.EnableCloudwatchLogsExports,
	}
	mci.SetCloudwatchLogsExportConfiguration(&cloudwatchConfig)

	return nil
}

func preDelete(_ context.Context, cr *svcapitypes.DBCluster, obj *svcsdk.DeleteDBClusterInput) (bool, error) {
	obj.DBClusterIdentifier = aws.String(meta.GetExternalName(cr))
	obj.SkipFinalSnapshot = cr.Spec.ForProvider.SkipFinalSnapshot

	return false, nil
}

func lateInitialize(in *svcapitypes.DBClusterParameters, out *svcsdk.DescribeDBClustersOutput) error {
	if out == nil || len(out.DBClusters) == 0 {
		return nil
	}

	if in == nil {
		in = &svcapitypes.DBClusterParameters{}
	}

	from := out.DBClusters[0]

	in.AvailabilityZones = aws.LateInitializeStringPtrSlice(in.AvailabilityZones, from.AvailabilityZones)
	in.BackupRetentionPeriod = aws.LateInitializeInt64Ptr(in.BackupRetentionPeriod, from.BackupRetentionPeriod)
	in.CharacterSetName = aws.LateInitializeStringPtr(in.CharacterSetName, from.CharacterSetName)
	in.DatabaseName = aws.LateInitializeStringPtr(in.DatabaseName, from.DatabaseName)
	in.DBClusterParameterGroupName = aws.LateInitializeStringPtr(in.DBClusterParameterGroupName, from.DBClusterParameterGroup)
	in.DBSubnetGroupName = aws.LateInitializeStringPtr(in.DBSubnetGroupName, from.DBSubnetGroup)
	in.DeletionProtection = aws.LateInitializeBoolPtr(in.DeletionProtection, from.DeletionProtection)
	in.EnableCloudwatchLogsExports = aws.LateInitializeStringPtrSlice(in.EnableCloudwatchLogsExports, from.EnabledCloudwatchLogsExports)
	in.Engine = aws.LateInitializeStringPtr(in.Engine, from.Engine)
	in.EngineVersion = aws.LateInitializeStringPtr(in.EngineVersion, from.EngineVersion)
	in.EnableIAMDatabaseAuthentication = aws.LateInitializeBoolPtr(in.EnableIAMDatabaseAuthentication, from.IAMDatabaseAuthenticationEnabled)
	in.KMSKeyID = aws.LateInitializeStringPtr(in.KMSKeyID, from.KmsKeyId)
	in.MasterUsername = aws.LateInitializeStringPtr(in.MasterUsername, from.MasterUsername)
	in.Port = aws.LateInitializeInt64Ptr(in.Port, from.Port)
	in.PreferredBackupWindow = aws.LateInitializeStringPtr(in.PreferredBackupWindow, from.PreferredBackupWindow)
	in.PreferredMaintenanceWindow = aws.LateInitializeStringPtr(in.PreferredMaintenanceWindow, from.PreferredMaintenanceWindow)
	in.ReplicationSourceIdentifier = aws.LateInitializeStringPtr(in.ReplicationSourceIdentifier, from.ReplicationSourceIdentifier)
	in.StorageEncrypted = aws.LateInitializeBoolPtr(in.StorageEncrypted, from.StorageEncrypted)

	if len(in.VPCSecurityGroupIDs) == 0 && len(from.VpcSecurityGroups) != 0 {
		in.VPCSecurityGroupIDs = make([]*string, len(from.VpcSecurityGroups))
		for i, val := range from.VpcSecurityGroups {
			in.VPCSecurityGroupIDs[i] = aws.LateInitializeStringPtr(in.VPCSecurityGroupIDs[i], val.VpcSecurityGroupId)
		}
	}
	return nil
}

// nolint:gocyclo
func isUpToDate(cr *svcapitypes.DBCluster, output *svcsdk.DescribeDBClustersOutput) (bool, error) {
	in := cr.Spec.ForProvider
	out := output.DBClusters[0]

	if aws.Int64Value(in.BackupRetentionPeriod) != aws.Int64Value(out.BackupRetentionPeriod) {
		return false, nil
	}
	if aws.StringValue(in.DBClusterParameterGroupName) != aws.StringValue(out.DBClusterParameterGroup) {
		return false, nil
	}
	if aws.BoolValue(in.DeletionProtection) != aws.BoolValue(out.DeletionProtection) {
		return false, nil
	}
	if !cmp.Equal(in.EnableCloudwatchLogsExports, out.EnabledCloudwatchLogsExports) {
		return false, nil
	}
	if aws.StringValue(in.EngineVersion) != aws.StringValue(out.EngineVersion) {
		return false, nil
	}
	if aws.BoolValue(in.EnableIAMDatabaseAuthentication) != aws.BoolValue(out.IAMDatabaseAuthenticationEnabled) {
		return false, nil
	}
	if aws.Int64Value(in.Port) != aws.Int64Value(out.Port) {
		return false, nil
	}
	if aws.StringValue(in.PreferredBackupWindow) != aws.StringValue(out.PreferredBackupWindow) {
		return false, nil
	}
	if aws.StringValue(in.PreferredMaintenanceWindow) != aws.StringValue(out.PreferredMaintenanceWindow) {
		return false, nil
	}
	if len(in.VPCSecurityGroupIDs) != len(out.VpcSecurityGroups) {
		return true, nil
	}

	vcpArr := make([]*string, len(in.VPCSecurityGroupIDs))
	for i := range out.VpcSecurityGroups {
		vcpArr[i] = out.VpcSecurityGroups[i].VpcSecurityGroupId
	}
	if !cmp.Equal(in.VPCSecurityGroupIDs, vcpArr) {
		return false, nil
	}

	return true, nil
}

func postObserve(_ context.Context, cr *svcapitypes.DBCluster, resp *svcsdk.DescribeDBClustersOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}

	switch aws.StringValue(resp.DBClusters[0].Status) {
	case string(statusAvailable):
		cr.SetConditions(xpv1.Available())
	case string(statusCreating):
		cr.SetConditions(xpv1.Creating())
	case string(statusDeleted):
		cr.SetConditions(xpv1.Unavailable())
	}

	return obs, nil
}
