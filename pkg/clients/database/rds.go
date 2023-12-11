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

package rds

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	rdstypes "github.com/aws/aws-sdk-go-v2/service/rds/types"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane-contrib/provider-aws/apis/database/v1beta1"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/rds/utils"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/jsonpatch"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
)

const (
	errGetPasswordSecretFailed = "cannot get password secret"
)

// Client defines RDS RDSClient operations
type Client interface {
	CreateDBInstance(context.Context, *rds.CreateDBInstanceInput, ...func(*rds.Options)) (*rds.CreateDBInstanceOutput, error)
	RestoreDBInstanceFromS3(context.Context, *rds.RestoreDBInstanceFromS3Input, ...func(*rds.Options)) (*rds.RestoreDBInstanceFromS3Output, error)
	RestoreDBInstanceFromDBSnapshot(context.Context, *rds.RestoreDBInstanceFromDBSnapshotInput, ...func(*rds.Options)) (*rds.RestoreDBInstanceFromDBSnapshotOutput, error)
	RestoreDBInstanceToPointInTime(context.Context, *rds.RestoreDBInstanceToPointInTimeInput, ...func(*rds.Options)) (*rds.RestoreDBInstanceToPointInTimeOutput, error)
	DescribeDBInstances(context.Context, *rds.DescribeDBInstancesInput, ...func(*rds.Options)) (*rds.DescribeDBInstancesOutput, error)
	ModifyDBInstance(context.Context, *rds.ModifyDBInstanceInput, ...func(*rds.Options)) (*rds.ModifyDBInstanceOutput, error)
	DeleteDBInstance(context.Context, *rds.DeleteDBInstanceInput, ...func(*rds.Options)) (*rds.DeleteDBInstanceOutput, error)
	AddTagsToResource(context.Context, *rds.AddTagsToResourceInput, ...func(*rds.Options)) (*rds.AddTagsToResourceOutput, error)
	RemoveTagsFromResource(context.Context, *rds.RemoveTagsFromResourceInput, ...func(*rds.Options)) (*rds.RemoveTagsFromResourceOutput, error)
}

// NewClient creates new RDS RDSClient with provided AWS Configurations/Credentials
func NewClient(cfg *aws.Config) Client {
	return rds.NewFromConfig(*cfg)
}

// IsErrorNotFound helper function to test for ErrCodeDBInstanceNotFoundFault error
func IsErrorNotFound(err error) bool {
	var nff *rdstypes.DBInstanceNotFoundFault
	return errors.As(err, &nff)
}

// GenerateCreateRDSInstanceInput from RDSInstanceSpec
func GenerateCreateRDSInstanceInput(name, password string, p *v1beta1.RDSInstanceParameters) *rds.CreateDBInstanceInput {
	// Partially duplicates GenerateRestoreDBInstanceFromS3Input and GenerateRestoreDBInstanceFromSnapshotInput:
	// Make sure any relevant changes are applied there too.
	c := &rds.CreateDBInstanceInput{
		DBInstanceIdentifier:               aws.String(name),
		AllocatedStorage:                   pointer.ToIntAsInt32Ptr(p.AllocatedStorage),
		AutoMinorVersionUpgrade:            p.AutoMinorVersionUpgrade,
		AvailabilityZone:                   p.AvailabilityZone,
		BackupRetentionPeriod:              pointer.ToIntAsInt32Ptr(p.BackupRetentionPeriod),
		CACertificateIdentifier:            p.CACertificateIdentifier,
		CharacterSetName:                   p.CharacterSetName,
		CopyTagsToSnapshot:                 p.CopyTagsToSnapshot,
		DBClusterIdentifier:                p.DBClusterIdentifier,
		DBInstanceClass:                    aws.String(p.DBInstanceClass),
		DBName:                             p.DBName,
		DBParameterGroupName:               p.DBParameterGroupName,
		DBSecurityGroups:                   p.DBSecurityGroups,
		DBSubnetGroupName:                  p.DBSubnetGroupName,
		DeletionProtection:                 p.DeletionProtection,
		Domain:                             p.Domain,
		DomainIAMRoleName:                  p.DomainIAMRoleName,
		EnableCloudwatchLogsExports:        p.EnableCloudwatchLogsExports,
		EnableIAMDatabaseAuthentication:    p.EnableIAMDatabaseAuthentication,
		EnablePerformanceInsights:          p.EnablePerformanceInsights,
		Engine:                             aws.String(p.Engine),
		EngineVersion:                      p.EngineVersion,
		Iops:                               pointer.ToIntAsInt32Ptr(p.IOPS),
		KmsKeyId:                           p.KMSKeyID,
		LicenseModel:                       p.LicenseModel,
		MasterUserPassword:                 pointer.ToOrNilIfZeroValue(password),
		MasterUsername:                     p.MasterUsername,
		MaxAllocatedStorage:                pointer.ToIntAsInt32Ptr(p.MaxAllocatedStorage),
		MonitoringInterval:                 pointer.ToIntAsInt32Ptr(p.MonitoringInterval),
		MonitoringRoleArn:                  p.MonitoringRoleARN,
		MultiAZ:                            p.MultiAZ,
		OptionGroupName:                    p.OptionGroupName,
		PerformanceInsightsKMSKeyId:        p.PerformanceInsightsKMSKeyID,
		PerformanceInsightsRetentionPeriod: pointer.ToIntAsInt32Ptr(p.PerformanceInsightsRetentionPeriod),
		Port:                               pointer.ToIntAsInt32Ptr(p.Port),
		PreferredBackupWindow:              p.PreferredBackupWindow,
		PreferredMaintenanceWindow:         p.PreferredMaintenanceWindow,
		PromotionTier:                      pointer.ToIntAsInt32Ptr(p.PromotionTier),
		PubliclyAccessible:                 p.PubliclyAccessible,
		StorageEncrypted:                   p.StorageEncrypted,
		Timezone:                           p.Timezone,
		StorageType:                        p.StorageType,
		VpcSecurityGroupIds:                p.VPCSecurityGroupIDs,
	}
	if len(p.ProcessorFeatures) != 0 {
		c.ProcessorFeatures = make([]rdstypes.ProcessorFeature, len(p.ProcessorFeatures))
		for i, val := range p.ProcessorFeatures {
			c.ProcessorFeatures[i] = rdstypes.ProcessorFeature{
				Name:  aws.String(val.Name),
				Value: aws.String(val.Value),
			}
		}
	}
	if len(p.Tags) != 0 {
		c.Tags = make([]rdstypes.Tag, len(p.Tags))
		for i, val := range p.Tags {
			c.Tags[i] = rdstypes.Tag{
				Key:   aws.String(val.Key),
				Value: aws.String(val.Value),
			}
		}
	}
	return c
}

// GenerateRestoreRDSInstanceFromS3Input from RDSInstanceSpec
func GenerateRestoreRDSInstanceFromS3Input(name, password string, p *v1beta1.RDSInstanceParameters) *rds.RestoreDBInstanceFromS3Input {
	// Partially duplicates GenerateCreateDBInstanceInput - make sure any relevant changes are applied there too.
	c := &rds.RestoreDBInstanceFromS3Input{
		DBInstanceIdentifier:               aws.String(name),
		AllocatedStorage:                   pointer.ToIntAsInt32Ptr(p.AllocatedStorage),
		AutoMinorVersionUpgrade:            p.AutoMinorVersionUpgrade,
		AvailabilityZone:                   p.AvailabilityZone,
		BackupRetentionPeriod:              pointer.ToIntAsInt32Ptr(p.BackupRetentionPeriod),
		CopyTagsToSnapshot:                 p.CopyTagsToSnapshot,
		DBInstanceClass:                    aws.String(p.DBInstanceClass),
		DBName:                             p.DBName,
		DBParameterGroupName:               p.DBParameterGroupName,
		DBSecurityGroups:                   p.DBSecurityGroups,
		DBSubnetGroupName:                  p.DBSubnetGroupName,
		DeletionProtection:                 p.DeletionProtection,
		EnableCloudwatchLogsExports:        p.EnableCloudwatchLogsExports,
		EnableIAMDatabaseAuthentication:    p.EnableIAMDatabaseAuthentication,
		EnablePerformanceInsights:          p.EnablePerformanceInsights,
		Engine:                             aws.String(p.Engine),
		EngineVersion:                      p.EngineVersion,
		Iops:                               pointer.ToIntAsInt32Ptr(p.IOPS),
		KmsKeyId:                           p.KMSKeyID,
		LicenseModel:                       p.LicenseModel,
		MasterUserPassword:                 pointer.ToOrNilIfZeroValue(password),
		MasterUsername:                     p.MasterUsername,
		MonitoringInterval:                 pointer.ToIntAsInt32Ptr(p.MonitoringInterval),
		MonitoringRoleArn:                  p.MonitoringRoleARN,
		MultiAZ:                            p.MultiAZ,
		OptionGroupName:                    p.OptionGroupName,
		PerformanceInsightsKMSKeyId:        p.PerformanceInsightsKMSKeyID,
		PerformanceInsightsRetentionPeriod: pointer.ToIntAsInt32Ptr(p.PerformanceInsightsRetentionPeriod),
		Port:                               pointer.ToIntAsInt32Ptr(p.Port),
		PreferredBackupWindow:              p.PreferredBackupWindow,
		PreferredMaintenanceWindow:         p.PreferredMaintenanceWindow,
		PubliclyAccessible:                 p.PubliclyAccessible,
		S3BucketName:                       p.RestoreFrom.S3.BucketName,
		S3IngestionRoleArn:                 p.RestoreFrom.S3.IngestionRoleARN,
		S3Prefix:                           p.RestoreFrom.S3.Prefix,
		SourceEngine:                       p.RestoreFrom.S3.SourceEngine,
		SourceEngineVersion:                p.RestoreFrom.S3.SourceEngineVersion,
		StorageEncrypted:                   p.StorageEncrypted,
		StorageType:                        p.StorageType,
		VpcSecurityGroupIds:                p.VPCSecurityGroupIDs,
	}
	if len(p.ProcessorFeatures) != 0 {
		c.ProcessorFeatures = make([]rdstypes.ProcessorFeature, len(p.ProcessorFeatures))
		for i, val := range p.ProcessorFeatures {
			c.ProcessorFeatures[i] = rdstypes.ProcessorFeature{
				Name:  aws.String(val.Name),
				Value: aws.String(val.Value),
			}
		}
	}
	if len(p.Tags) != 0 {
		c.Tags = make([]rdstypes.Tag, len(p.Tags))
		for i, val := range p.Tags {
			c.Tags[i] = rdstypes.Tag{
				Key:   aws.String(val.Key),
				Value: aws.String(val.Value),
			}
		}
	}
	return c
}

// GenerateRestoreRDSInstanceFromSnapshotInput from RDSInstanceSpec
func GenerateRestoreRDSInstanceFromSnapshotInput(name string, p *v1beta1.RDSInstanceParameters) *rds.RestoreDBInstanceFromDBSnapshotInput {
	// Partially duplicates GenerateCreateDBInstanceInput - make sure any relevant changes are applied there too.
	c := &rds.RestoreDBInstanceFromDBSnapshotInput{
		DBInstanceIdentifier:            aws.String(name),
		AutoMinorVersionUpgrade:         p.AutoMinorVersionUpgrade,
		AvailabilityZone:                p.AvailabilityZone,
		CopyTagsToSnapshot:              p.CopyTagsToSnapshot,
		DBInstanceClass:                 aws.String(p.DBInstanceClass),
		DBName:                          p.DBName,
		DBParameterGroupName:            p.DBParameterGroupName,
		DBSnapshotIdentifier:            p.RestoreFrom.Snapshot.SnapshotIdentifier,
		DBSubnetGroupName:               p.DBSubnetGroupName,
		DeletionProtection:              p.DeletionProtection,
		Domain:                          p.Domain,
		DomainIAMRoleName:               p.DomainIAMRoleName,
		EnableCloudwatchLogsExports:     p.EnableCloudwatchLogsExports,
		EnableIAMDatabaseAuthentication: p.EnableIAMDatabaseAuthentication,
		Engine:                          aws.String(p.Engine),
		Iops:                            pointer.ToIntAsInt32Ptr(p.IOPS),
		LicenseModel:                    p.LicenseModel,
		MultiAZ:                         p.MultiAZ,
		OptionGroupName:                 p.OptionGroupName,
		Port:                            pointer.ToIntAsInt32Ptr(p.Port),
		PubliclyAccessible:              p.PubliclyAccessible,
		StorageType:                     p.StorageType,
		VpcSecurityGroupIds:             p.VPCSecurityGroupIDs,
	}
	if len(p.ProcessorFeatures) != 0 {
		c.ProcessorFeatures = make([]rdstypes.ProcessorFeature, len(p.ProcessorFeatures))
		for i, val := range p.ProcessorFeatures {
			c.ProcessorFeatures[i] = rdstypes.ProcessorFeature{
				Name:  aws.String(val.Name),
				Value: aws.String(val.Value),
			}
		}
	}
	if len(p.Tags) != 0 {
		c.Tags = make([]rdstypes.Tag, len(p.Tags))
		for i, val := range p.Tags {
			c.Tags[i] = rdstypes.Tag{
				Key:   aws.String(val.Key),
				Value: aws.String(val.Value),
			}
		}
	}
	return c
}

// GenerateRestoreRDSInstanceToPointInTimeInput from RDSInstanceSpec
func GenerateRestoreRDSInstanceToPointInTimeInput(name string, p *v1beta1.RDSInstanceParameters) *rds.RestoreDBInstanceToPointInTimeInput {
	// Partially duplicates GenerateCreateDBInstanceInput - make sure any relevant changes are applied there too.
	// Need to convert restoreTime from *metav1.Time to *time.Time
	var restoreTime *time.Time
	if p.RestoreFrom.PointInTime.RestoreTime != nil {
		t, _ := time.Parse(time.RFC3339, p.RestoreFrom.PointInTime.RestoreTime.Format(time.RFC3339))
		restoreTime = &t
	}
	c := &rds.RestoreDBInstanceToPointInTimeInput{
		AutoMinorVersionUpgrade:         p.AutoMinorVersionUpgrade,
		AvailabilityZone:                p.AvailabilityZone,
		CopyTagsToSnapshot:              p.CopyTagsToSnapshot,
		DBInstanceClass:                 aws.String(p.DBInstanceClass),
		DBName:                          p.DBName,
		DBParameterGroupName:            p.DBParameterGroupName,
		DBSubnetGroupName:               p.DBSubnetGroupName,
		DeletionProtection:              p.DeletionProtection,
		Domain:                          p.Domain,
		DomainIAMRoleName:               p.DomainIAMRoleName,
		EnableCloudwatchLogsExports:     p.EnableCloudwatchLogsExports,
		EnableIAMDatabaseAuthentication: p.EnableIAMDatabaseAuthentication,
		Engine:                          aws.String(p.Engine),
		Iops:                            pointer.ToIntAsInt32Ptr(p.IOPS),
		LicenseModel:                    p.LicenseModel,
		MultiAZ:                         p.MultiAZ,
		OptionGroupName:                 p.OptionGroupName,
		Port:                            pointer.ToIntAsInt32Ptr(p.Port),
		PubliclyAccessible:              p.PubliclyAccessible,
		StorageType:                     p.StorageType,
		VpcSecurityGroupIds:             p.VPCSecurityGroupIDs,

		TargetDBInstanceIdentifier:          aws.String(name),
		RestoreTime:                         restoreTime,
		UseLatestRestorableTime:             p.RestoreFrom.PointInTime.UseLatestRestorableTime,
		SourceDBInstanceAutomatedBackupsArn: p.RestoreFrom.PointInTime.SourceDBInstanceAutomatedBackupsArn,
		SourceDBInstanceIdentifier:          p.RestoreFrom.PointInTime.SourceDBInstanceIdentifier,
		SourceDbiResourceId:                 p.RestoreFrom.PointInTime.SourceDbiResourceID,
	}
	if len(p.ProcessorFeatures) != 0 {
		c.ProcessorFeatures = make([]rdstypes.ProcessorFeature, len(p.ProcessorFeatures))
		for i, val := range p.ProcessorFeatures {
			c.ProcessorFeatures[i] = rdstypes.ProcessorFeature{
				Name:  aws.String(val.Name),
				Value: aws.String(val.Value),
			}
		}
	}
	if len(p.Tags) != 0 {
		c.Tags = make([]rdstypes.Tag, len(p.Tags))
		for i, val := range p.Tags {
			c.Tags[i] = rdstypes.Tag{
				Key:   aws.String(val.Key),
				Value: aws.String(val.Value),
			}
		}
	}
	return c
}

// CreatePatch creates a *v1beta1.RDSInstanceParameters that has only the changed
// values between the target *v1beta1.RDSInstanceParameters and the current
// *rds.DBInstance
func CreatePatch(in *rdstypes.DBInstance, spec *v1beta1.RDSInstanceParameters) (*v1beta1.RDSInstanceParameters, error) {
	target := spec.DeepCopy()
	currentParams := &v1beta1.RDSInstanceParameters{}
	LateInitialize(currentParams, in)

	// AvailabilityZone parameters is not allowed for MultiAZ deployments.
	// So set this to nil if that is the case to avoid unnecessary diffs.
	if ptr.Deref(target.MultiAZ, false) {
		target.AvailabilityZone = nil
	}
	if ptr.Deref(currentParams.MultiAZ, false) {
		currentParams.AvailabilityZone = nil
	}

	// Don't attempt to scale down storage if autoscaling is enabled,
	// and the current storage is larger than what was once
	// requested. We still want to allow the user to manually scale
	// the storage, so we only remove it if we're certain AWS will
	// reject the value
	if target.MaxAllocatedStorage != nil && aws.ToInt(target.AllocatedStorage) < aws.ToInt(currentParams.AllocatedStorage) {
		// By making the values equal, CreateJSONPatch will exclude
		// the field. It might seem more sensible to change the target
		// object, but it's a pointer to the CR so we do not want to
		// mutate it.
		currentParams.AllocatedStorage = target.AllocatedStorage
	}

	// AWS Backup takes ownership of backupRetentionPeriod and
	// preferredBackupWindow if it is in use, so we need to exclude
	// the field in the diff
	if in.AwsBackupRecoveryPointArn != nil {
		if target.BackupRetentionPeriod != nil {
			currentParams.BackupRetentionPeriod = target.BackupRetentionPeriod
		}

		if target.PreferredBackupWindow != nil {
			currentParams.PreferredBackupWindow = target.PreferredBackupWindow
		}
	}

	jsonPatch, err := jsonpatch.CreateJSONPatch(currentParams, target)
	if err != nil {
		return nil, err
	}
	patch := &v1beta1.RDSInstanceParameters{}
	if err := json.Unmarshal(jsonPatch, patch); err != nil {
		return nil, err
	}
	return patch, nil
}

// GenerateModifyDBInstanceInput from RDSInstanceSpec
func GenerateModifyDBInstanceInput(name string, p *v1beta1.RDSInstanceParameters, db *rdstypes.DBInstance) *rds.ModifyDBInstanceInput {
	// NOTE(muvaf): MasterUserPassword is not used here. So, password is set once
	// and kept that way.
	// NOTE(muvaf): Change of DBInstanceIdentifier is supported by AWS but
	// Crossplane assumes identification info never changes, so, we don't support
	// it.
	m := &rds.ModifyDBInstanceInput{
		DBInstanceIdentifier:               aws.String(name),
		AllocatedStorage:                   pointer.ToIntAsInt32Ptr(p.AllocatedStorage),
		AllowMajorVersionUpgrade:           aws.ToBool(p.AllowMajorVersionUpgrade),
		ApplyImmediately:                   aws.ToBool(p.ApplyModificationsImmediately),
		AutoMinorVersionUpgrade:            p.AutoMinorVersionUpgrade,
		BackupRetentionPeriod:              pointer.ToIntAsInt32Ptr(p.BackupRetentionPeriod),
		CACertificateIdentifier:            p.CACertificateIdentifier,
		CopyTagsToSnapshot:                 p.CopyTagsToSnapshot,
		DBInstanceClass:                    pointer.ToOrNilIfZeroValue(p.DBInstanceClass),
		DBParameterGroupName:               p.DBParameterGroupName,
		DBPortNumber:                       pointer.ToIntAsInt32Ptr(p.Port),
		DBSecurityGroups:                   p.DBSecurityGroups,
		DBSubnetGroupName:                  p.DBSubnetGroupName,
		DeletionProtection:                 p.DeletionProtection,
		Domain:                             p.Domain,
		DomainIAMRoleName:                  p.DomainIAMRoleName,
		EnableIAMDatabaseAuthentication:    p.EnableIAMDatabaseAuthentication,
		EnablePerformanceInsights:          p.EnablePerformanceInsights,
		EngineVersion:                      p.EngineVersion,
		Iops:                               pointer.ToIntAsInt32Ptr(p.IOPS),
		LicenseModel:                       p.LicenseModel,
		MaxAllocatedStorage:                pointer.ToIntAsInt32Ptr(p.MaxAllocatedStorage),
		MonitoringInterval:                 pointer.ToIntAsInt32Ptr(p.MonitoringInterval),
		MonitoringRoleArn:                  p.MonitoringRoleARN,
		MultiAZ:                            p.MultiAZ,
		OptionGroupName:                    p.OptionGroupName,
		PerformanceInsightsKMSKeyId:        p.PerformanceInsightsKMSKeyID,
		PerformanceInsightsRetentionPeriod: pointer.ToIntAsInt32Ptr(p.PerformanceInsightsRetentionPeriod),
		PreferredBackupWindow:              p.PreferredBackupWindow,
		PreferredMaintenanceWindow:         p.PreferredMaintenanceWindow,
		PromotionTier:                      pointer.ToIntAsInt32Ptr(p.PromotionTier),
		PubliclyAccessible:                 p.PubliclyAccessible,
		StorageType:                        p.StorageType,
		UseDefaultProcessorFeatures:        p.UseDefaultProcessorFeatures,
		VpcSecurityGroupIds:                p.VPCSecurityGroupIDs,
	}
	if len(p.ProcessorFeatures) != 0 {
		m.ProcessorFeatures = make([]rdstypes.ProcessorFeature, len(p.ProcessorFeatures))
		for i, val := range p.ProcessorFeatures {
			m.ProcessorFeatures[i] = rdstypes.ProcessorFeature{
				Name:  aws.String(val.Name),
				Value: aws.String(val.Value),
			}
		}
	}

	m.CloudwatchLogsExportConfiguration = generateCloudWatchExportConfiguration(
		p.EnableCloudwatchLogsExports,
		db.EnabledCloudwatchLogsExports)

	return m
}

// GenerateObservation is used to produce v1alpha3.RDSInstanceObservation from
// rds.DBInstance.
func GenerateObservation(db rdstypes.DBInstance) v1beta1.RDSInstanceObservation { //nolint:gocyclo
	o := v1beta1.RDSInstanceObservation{
		AllocatedStorage:                      int(db.AllocatedStorage),
		AWSBackupRecoveryPointARN:             aws.ToString(db.AwsBackupRecoveryPointArn),
		BackupRetentionPeriod:                 int(db.BackupRetentionPeriod),
		DBInstanceStatus:                      aws.ToString(db.DBInstanceStatus),
		DBInstanceArn:                         aws.ToString(db.DBInstanceArn),
		DBInstancePort:                        int(db.DbInstancePort),
		DBResourceID:                          aws.ToString(db.DbiResourceId),
		EnabledCloudwatchLogsExports:          db.EnabledCloudwatchLogsExports,
		EnhancedMonitoringResourceArn:         aws.ToString(db.EnhancedMonitoringResourceArn),
		PerformanceInsightsEnabled:            aws.ToBool(db.PerformanceInsightsEnabled),
		ReadReplicaDBClusterIdentifiers:       db.ReadReplicaDBClusterIdentifiers,
		ReadReplicaDBInstanceIdentifiers:      db.ReadReplicaDBInstanceIdentifiers,
		ReadReplicaSourceDBInstanceIdentifier: aws.ToString(db.ReadReplicaSourceDBInstanceIdentifier),
		SecondaryAvailabilityZone:             aws.ToString(db.SecondaryAvailabilityZone),
	}
	if db.LatestRestorableTime != nil {
		t := metav1.NewTime(*db.LatestRestorableTime)
		o.LatestRestorableTime = &t
	}
	if db.InstanceCreateTime != nil {
		t := metav1.NewTime(*db.InstanceCreateTime)
		o.InstanceCreateTime = &t
	}
	if len(db.DBParameterGroups) != 0 {
		o.DBParameterGroups = make([]v1beta1.DBParameterGroupStatus, len(db.DBParameterGroups))
		for i, val := range db.DBParameterGroups {
			o.DBParameterGroups[i] = v1beta1.DBParameterGroupStatus{
				DBParameterGroupName: aws.ToString(val.DBParameterGroupName),
				ParameterApplyStatus: aws.ToString(val.ParameterApplyStatus),
			}
		}
	}
	if len(db.DBSecurityGroups) != 0 {
		o.DBSecurityGroups = make([]v1beta1.DBSecurityGroupMembership, len(db.DBSecurityGroups))
		for i, val := range db.DBSecurityGroups {
			o.DBSecurityGroups[i] = v1beta1.DBSecurityGroupMembership{
				DBSecurityGroupName: aws.ToString(val.DBSecurityGroupName),
				Status:              aws.ToString(val.Status),
			}
		}
	}
	if db.DBSubnetGroup != nil {
		o.DBSubnetGroup = v1beta1.DBSubnetGroupInRDS{
			DBSubnetGroupARN:         aws.ToString(db.DBSubnetGroup.DBSubnetGroupArn),
			DBSubnetGroupDescription: aws.ToString(db.DBSubnetGroup.DBSubnetGroupDescription),
			DBSubnetGroupName:        aws.ToString(db.DBSubnetGroup.DBSubnetGroupName),
			SubnetGroupStatus:        aws.ToString(db.DBSubnetGroup.SubnetGroupStatus),
			VPCID:                    aws.ToString(db.DBSubnetGroup.VpcId),
		}
		if len(db.DBSubnetGroup.Subnets) != 0 {
			o.DBSubnetGroup.Subnets = make([]v1beta1.SubnetInRDS, len(db.DBSubnetGroup.Subnets))
			for i, val := range db.DBSubnetGroup.Subnets {
				o.DBSubnetGroup.Subnets[i] = v1beta1.SubnetInRDS{
					SubnetIdentifier: aws.ToString(val.SubnetIdentifier),
					SubnetStatus:     aws.ToString(val.SubnetStatus),
				}
				if val.SubnetAvailabilityZone != nil {
					o.DBSubnetGroup.Subnets[i].SubnetAvailabilityZone = v1beta1.AvailabilityZone{
						Name: aws.ToString(val.SubnetAvailabilityZone.Name),
					}
				}
			}
		}
	}
	if len(db.DomainMemberships) != 0 {
		o.DomainMemberships = make([]v1beta1.DomainMembership, len(db.DomainMemberships))
		for i, val := range db.DomainMemberships {
			o.DomainMemberships[i] = v1beta1.DomainMembership{
				Domain:      aws.ToString(val.Domain),
				FQDN:        aws.ToString(val.FQDN),
				IAMRoleName: aws.ToString(val.IAMRoleName),
				Status:      aws.ToString(val.Status),
			}
		}
	}
	if db.Endpoint != nil {
		o.Endpoint = v1beta1.Endpoint{
			Address:      aws.ToString(db.Endpoint.Address),
			HostedZoneID: aws.ToString(db.Endpoint.HostedZoneId),
			Port:         int(db.Endpoint.Port),
		}
	}
	if len(db.OptionGroupMemberships) != 0 {
		o.OptionGroupMemberships = make([]v1beta1.OptionGroupMembership, len(db.OptionGroupMemberships))
		for i, val := range db.OptionGroupMemberships {
			o.OptionGroupMemberships[i] = v1beta1.OptionGroupMembership{
				OptionGroupName: aws.ToString(val.OptionGroupName),
				Status:          aws.ToString(val.Status),
			}
		}
	}
	if db.PendingModifiedValues != nil {
		o.PendingModifiedValues = v1beta1.PendingModifiedValues{
			AllocatedStorage:        int(aws.ToInt32(db.PendingModifiedValues.AllocatedStorage)),
			BackupRetentionPeriod:   int(aws.ToInt32(db.PendingModifiedValues.BackupRetentionPeriod)),
			CACertificateIdentifier: aws.ToString(db.PendingModifiedValues.CACertificateIdentifier),
			DBInstanceClass:         aws.ToString(db.PendingModifiedValues.DBInstanceClass),
			DBSubnetGroupName:       aws.ToString(db.PendingModifiedValues.DBSubnetGroupName),
			IOPS:                    int(aws.ToInt32(db.PendingModifiedValues.Iops)),
			LicenseModel:            aws.ToString(db.PendingModifiedValues.LicenseModel),
			MultiAZ:                 aws.ToBool(db.PendingModifiedValues.MultiAZ),
			Port:                    int(aws.ToInt32(db.PendingModifiedValues.Port)),
			StorageType:             aws.ToString(db.PendingModifiedValues.StorageType),
		}
		if db.PendingModifiedValues.PendingCloudwatchLogsExports != nil {
			o.PendingModifiedValues.PendingCloudwatchLogsExports = v1beta1.PendingCloudwatchLogsExports{
				LogTypesToDisable: db.PendingModifiedValues.PendingCloudwatchLogsExports.LogTypesToDisable,
				LogTypesToEnable:  db.PendingModifiedValues.PendingCloudwatchLogsExports.LogTypesToEnable,
			}
		}
		if len(db.PendingModifiedValues.ProcessorFeatures) != 0 {
			o.PendingModifiedValues.ProcessorFeatures = make([]v1beta1.ProcessorFeature, len(db.PendingModifiedValues.ProcessorFeatures))
			for i, val := range db.PendingModifiedValues.ProcessorFeatures {
				o.PendingModifiedValues.ProcessorFeatures[i] = v1beta1.ProcessorFeature{
					Name:  aws.ToString(val.Name),
					Value: aws.ToString(val.Value),
				}
			}
		}
	}
	if len(db.StatusInfos) != 0 {
		o.StatusInfos = make([]v1beta1.DBInstanceStatusInfo, len(db.StatusInfos))
		for i, val := range db.StatusInfos {
			o.StatusInfos[i] = v1beta1.DBInstanceStatusInfo{
				Message:    aws.ToString(val.Message),
				Status:     aws.ToString(val.Status),
				StatusType: aws.ToString(val.StatusType),
				Normal:     val.Normal,
			}
		}
	}
	if len(db.VpcSecurityGroups) != 0 {
		o.VPCSecurityGroups = make([]v1beta1.VPCSecurityGroupMembership, len(db.VpcSecurityGroups))
		for i, val := range db.VpcSecurityGroups {
			o.VPCSecurityGroups[i] = v1beta1.VPCSecurityGroupMembership{
				Status:             aws.ToString(val.Status),
				VPCSecurityGroupID: aws.ToString(val.VpcSecurityGroupId),
			}
		}
	}
	return o
}

// LateInitialize fills the empty fields in *v1beta1.RDSInstanceParameters with
// the values seen in rds.DBInstance.
func LateInitialize(in *v1beta1.RDSInstanceParameters, db *rdstypes.DBInstance) { //nolint:gocyclo
	if db == nil {
		return
	}
	in.DBInstanceClass = pointer.LateInitializeValueFromPtr(in.DBInstanceClass, db.DBInstanceClass)
	in.Engine = pointer.LateInitializeValueFromPtr(in.Engine, db.Engine)

	in.AllocatedStorage = pointer.LateInitializeIntFrom32Ptr(in.AllocatedStorage, &db.AllocatedStorage)
	in.AutoMinorVersionUpgrade = pointer.LateInitialize(in.AutoMinorVersionUpgrade, ptr.To(db.AutoMinorVersionUpgrade))
	in.AvailabilityZone = pointer.LateInitialize(in.AvailabilityZone, db.AvailabilityZone)
	in.BackupRetentionPeriod = pointer.LateInitializeIntFromInt32Ptr(in.BackupRetentionPeriod, &db.BackupRetentionPeriod)
	in.CACertificateIdentifier = pointer.LateInitialize(in.CACertificateIdentifier, db.CACertificateIdentifier)
	in.CharacterSetName = pointer.LateInitialize(in.CharacterSetName, db.CharacterSetName)
	in.CopyTagsToSnapshot = pointer.LateInitialize(in.CopyTagsToSnapshot, ptr.To(db.CopyTagsToSnapshot))
	in.DBClusterIdentifier = pointer.LateInitialize(in.DBClusterIdentifier, db.DBClusterIdentifier)
	in.DBName = pointer.LateInitialize(in.DBName, db.DBName)
	in.DeletionProtection = pointer.LateInitialize(in.DeletionProtection, ptr.To(db.DeletionProtection))
	in.EnableIAMDatabaseAuthentication = pointer.LateInitialize(in.EnableIAMDatabaseAuthentication, ptr.To(db.IAMDatabaseAuthenticationEnabled))
	in.EnablePerformanceInsights = pointer.LateInitialize(in.EnablePerformanceInsights, db.PerformanceInsightsEnabled)
	in.IOPS = pointer.LateInitializeIntFrom32Ptr(in.IOPS, db.Iops)
	in.KMSKeyID = pointer.LateInitialize(in.KMSKeyID, db.KmsKeyId)
	in.LicenseModel = pointer.LateInitialize(in.LicenseModel, db.LicenseModel)
	in.MasterUsername = pointer.LateInitialize(in.MasterUsername, db.MasterUsername)
	in.MaxAllocatedStorage = pointer.LateInitializeIntFrom32Ptr(in.MaxAllocatedStorage, db.MaxAllocatedStorage)
	in.MonitoringInterval = pointer.LateInitializeIntFrom32Ptr(in.MonitoringInterval, db.MonitoringInterval)
	in.MonitoringRoleARN = pointer.LateInitialize(in.MonitoringRoleARN, db.MonitoringRoleArn)
	in.MultiAZ = pointer.LateInitialize(in.MultiAZ, ptr.To(db.MultiAZ))
	in.PerformanceInsightsKMSKeyID = pointer.LateInitialize(in.PerformanceInsightsKMSKeyID, db.PerformanceInsightsKMSKeyId)
	in.PerformanceInsightsRetentionPeriod = pointer.LateInitializeIntFrom32Ptr(in.PerformanceInsightsRetentionPeriod, db.PerformanceInsightsRetentionPeriod)
	in.PreferredBackupWindow = pointer.LateInitialize(in.PreferredBackupWindow, db.PreferredBackupWindow)
	in.PreferredMaintenanceWindow = pointer.LateInitialize(in.PreferredMaintenanceWindow, db.PreferredMaintenanceWindow)
	in.PromotionTier = pointer.LateInitializeIntFrom32Ptr(in.PromotionTier, db.PromotionTier)
	in.PubliclyAccessible = pointer.LateInitialize(in.PubliclyAccessible, ptr.To(db.PubliclyAccessible))
	in.StorageEncrypted = pointer.LateInitialize(in.StorageEncrypted, ptr.To(db.StorageEncrypted))
	in.StorageType = pointer.LateInitialize(in.StorageType, db.StorageType)
	in.Timezone = pointer.LateInitialize(in.Timezone, db.Timezone)

	// NOTE(muvaf): Do not use db.DbInstancePort as that always returns 0 for
	// some reason. See the bug here:
	// https://github.com/aws/aws-sdk-java/issues/924#issuecomment-658089792
	if db.Endpoint != nil {
		in.Port = pointer.LateInitializeIntFrom32Ptr(in.Port, &db.Endpoint.Port)
	}

	if len(in.DBSecurityGroups) == 0 && len(db.DBSecurityGroups) != 0 {
		in.DBSecurityGroups = make([]string, len(db.DBSecurityGroups))
		for i, val := range db.DBSecurityGroups {
			in.DBSecurityGroups[i] = aws.ToString(val.DBSecurityGroupName)
		}
	}
	if aws.ToString(in.DBSubnetGroupName) == "" && db.DBSubnetGroup != nil {
		in.DBSubnetGroupName = db.DBSubnetGroup.DBSubnetGroupName
	}
	if len(in.ProcessorFeatures) == 0 && len(db.ProcessorFeatures) != 0 {
		in.ProcessorFeatures = make([]v1beta1.ProcessorFeature, len(db.ProcessorFeatures))
		for i, val := range db.ProcessorFeatures {
			in.ProcessorFeatures[i] = v1beta1.ProcessorFeature{
				Name:  aws.ToString(val.Name),
				Value: aws.ToString(val.Value),
			}
		}
	}
	if len(in.VPCSecurityGroupIDs) == 0 && len(db.VpcSecurityGroups) != 0 {
		in.VPCSecurityGroupIDs = make([]string, len(db.VpcSecurityGroups))
		for i, val := range db.VpcSecurityGroups {
			in.VPCSecurityGroupIDs[i] = aws.ToString(val.VpcSecurityGroupId)
		}
	}
	in.EngineVersion = pointer.LateInitialize(in.EngineVersion, db.EngineVersion)
	// When version 5.6 is chosen, AWS creates 5.6.41 and that's totally valid.
	// But we detect as if we need to update it all the time. Here, we assign
	// the actual full version to our spec to avoid unnecessary update signals.
	if strings.HasPrefix(aws.ToString(db.EngineVersion), aws.ToString(in.EngineVersion)) {
		in.EngineVersion = db.EngineVersion
	}
	if in.DBParameterGroupName == nil {
		for i := range db.DBParameterGroups {
			if db.DBParameterGroups[i].DBParameterGroupName != nil {
				in.DBParameterGroupName = db.DBParameterGroups[i].DBParameterGroupName
				break
			}
		}
	}
	// TODO: remove deprecated field + code. Mapping to EnableCloudwatchLogsExports while in deprecation.
	//nolint:staticcheck
	if len(in.EnableCloudwatchLogsExports) == 0 && in.CloudwatchLogsExportConfiguration != nil {
		in.EnableCloudwatchLogsExports = in.CloudwatchLogsExportConfiguration.EnableLogTypes
	}

	in.OptionGroupName = lateInitializeOptionGroupName(in.OptionGroupName, db.OptionGroupMemberships)

}

func lateInitializeOptionGroupName(inOptionGroupName *string, members []rdstypes.OptionGroupMembership) *string {

	if inOptionGroupName == nil && len(members) != 0 {

		for _, group := range members {
			if group.OptionGroupName != nil && group.Status != nil {

				// find the OptionGroup that is applied or will be applied to the DB
				switch pointer.StringValue(group.Status) {
				case "in-sync", "applying", "pending-apply", "pending-maintenance-apply":
					return group.OptionGroupName
				}
			}
		}
	}
	return inOptionGroupName
}

// IsUpToDate checks whether there is a change in any of the modifiable fields.
func IsUpToDate(ctx context.Context, kube client.Client, r *v1beta1.RDSInstance, db rdstypes.DBInstance) (bool, string, []rdstypes.Tag, []string, error) { //nolint:gocyclo

	addTags := []rdstypes.Tag{}
	removeTags := []string{}

	_, pwdChanged, err := GetPassword(ctx, kube, r.Spec.ForProvider.MasterPasswordSecretRef, r.Spec.WriteConnectionSecretToReference)
	if err != nil {
		return false, "", addTags, removeTags, err
	}
	patch, err := CreatePatch(&db, &r.Spec.ForProvider)
	if err != nil {
		return false, "", addTags, removeTags, err
	}
	diff := cmp.Diff(&v1beta1.RDSInstanceParameters{}, patch, cmpopts.EquateEmpty(),
		cmpopts.IgnoreTypes(&xpv1.Reference{}, &xpv1.Selector{}, []xpv1.Reference{}),
		cmpopts.IgnoreFields(v1beta1.RDSInstanceParameters{}, "Region"),
		cmpopts.IgnoreFields(v1beta1.RDSInstanceParameters{}, "Tags"),
		cmpopts.IgnoreFields(v1beta1.RDSInstanceParameters{}, "DBName"),
		cmpopts.IgnoreFields(v1beta1.RDSInstanceParameters{}, "EngineVersion"),
		cmpopts.IgnoreFields(v1beta1.RDSInstanceParameters{}, "SkipFinalSnapshotBeforeDeletion"),
		cmpopts.IgnoreFields(v1beta1.RDSInstanceParameters{}, "FinalDBSnapshotIdentifier"),
		cmpopts.IgnoreFields(v1beta1.RDSInstanceParameters{}, "DeleteAutomatedBackups"),
		cmpopts.IgnoreFields(v1beta1.RDSInstanceParameters{}, "ApplyModificationsImmediately"),
		cmpopts.IgnoreFields(v1beta1.RDSInstanceParameters{}, "AllowMajorVersionUpgrade"),
		cmpopts.IgnoreFields(v1beta1.RDSInstanceParameters{}, "MasterPasswordSecretRef"),
		cmpopts.IgnoreFields(v1beta1.RDSInstanceParameters{}, "OptionGroupName"),
		cmpopts.IgnoreFields(v1beta1.RDSInstanceParameters{}, "KMSKeyID"),
		cmpopts.IgnoreFields(v1beta1.RDSInstanceParameters{}, "EnableCloudwatchLogsExports"),
		// TODO: remove deprecated field + code. Mapping to EnableCloudwatchLogsExports while in deprecation.
		cmpopts.IgnoreFields(v1beta1.RDSInstanceParameters{}, "CloudwatchLogsExportConfiguration"),
	)

	addTags, removeTags = DiffTags(r.Spec.ForProvider.Tags, db.TagList)
	tagsChanged := len(addTags) != 0 || len(removeTags) != 0

	engineVersionChanged := !isEngineVersionUpToDate(r, db)

	optionGroupChanged := !isOptionGroupUpToDate(r, db)

	cloudwatchLogsExportChanged := false
	// only check CloudwatchLogsExports if there are no pending cloudwatchlogs exports (ignores the apply immediately setting)
	// (to avoid: "api error InvalidParameterCombination: You cannot configure CloudWatch Logs while a previous configuration is in progress.")
	if db.PendingModifiedValues != nil && db.PendingModifiedValues.PendingCloudwatchLogsExports == nil {
		cloudwatchLogsExportChanged = !areSameElements(r.Spec.ForProvider.EnableCloudwatchLogsExports, db.EnabledCloudwatchLogsExports)
	}

	if diff == "" && !pwdChanged && !engineVersionChanged && !optionGroupChanged && !cloudwatchLogsExportChanged && !tagsChanged {
		return true, "", addTags, removeTags, nil
	}

	diff = "Found observed difference in rds\n" + diff

	if tagsChanged {
		diff += fmt.Sprintf("\nadd %d tag(s) and remove %d tag(s)", len(addTags), len(removeTags))
	}

	return false, diff, addTags, removeTags, nil
}

func isEngineVersionUpToDate(cr *v1beta1.RDSInstance, db rdstypes.DBInstance) bool {
	// If EngineVersion is not set, AWS sets a default value,
	// so we do not try to update in this case
	if cr.Spec.ForProvider.EngineVersion != nil {
		if db.EngineVersion == nil {
			return false
		}

		// Upgrade is only necessary if the spec version is higher.
		// Downgrades are not possible in AWS.
		c := utils.CompareEngineVersions(*cr.Spec.ForProvider.EngineVersion, *db.EngineVersion)
		return c <= 0
	}
	return true
}

func isOptionGroupUpToDate(cr *v1beta1.RDSInstance, db rdstypes.DBInstance) bool {
	// If OptionGroupName is not set, AWS sets a default OptionGroup,
	// so we do not try to update in this case
	if cr.Spec.ForProvider.OptionGroupName != nil {
		for _, group := range db.OptionGroupMemberships {
			if group.OptionGroupName != nil && (pointer.StringValue(group.OptionGroupName) == pointer.StringValue(cr.Spec.ForProvider.OptionGroupName)) {

				switch pointer.StringValue(group.Status) {
				case "pending-maintenance-apply":
					// If ApplyModificationsImmediately was turned on after the OptionGroup change was requested,
					// we can make a new Modify request
					if pointer.BoolValue(cr.Spec.ForProvider.ApplyModificationsImmediately) {
						return false
					}
					return true
				case "pending-maintenance-removal":
					return false
				default: // "in-sync", "applying", "pending-apply", "pending-removal", "removing", "failed"
					return true
				}
			}
		}
		return false
	}
	return true
}

// GetPassword fetches the referenced input password for an RDSInstance CRD and determines whether it has changed or not
func GetPassword(ctx context.Context, kube client.Client, in *xpv1.SecretKeySelector, out *xpv1.SecretReference) (newPwd string, changed bool, err error) {
	if in == nil {
		return "", false, nil
	}
	nn := types.NamespacedName{
		Name:      in.Name,
		Namespace: in.Namespace,
	}
	s := &corev1.Secret{}
	if err := kube.Get(ctx, nn, s); err != nil {
		return "", false, errors.Wrap(err, errGetPasswordSecretFailed)
	}
	newPwd = string(s.Data[in.Key])

	if out != nil {
		nn = types.NamespacedName{
			Name:      out.Name,
			Namespace: out.Namespace,
		}
		s = &corev1.Secret{}
		// the output secret may not exist yet, so we can skip returning an
		// error if the error is NotFound
		if err := kube.Get(ctx, nn, s); resource.IgnoreNotFound(err) != nil {
			return "", false, err
		}
		// if newPwd was set to some value, compare value in output secret with
		// newPwd
		changed = newPwd != "" && newPwd != string(s.Data[xpv1.ResourceCredentialsSecretPasswordKey])
	}

	return newPwd, changed, nil
}

// GetConnectionDetails extracts managed.ConnectionDetails out of v1beta1.RDSInstance.
func GetConnectionDetails(in v1beta1.RDSInstance) managed.ConnectionDetails {
	if in.Status.AtProvider.Endpoint.Address == "" {
		return nil
	}
	return managed.ConnectionDetails{
		xpv1.ResourceCredentialsSecretEndpointKey: []byte(in.Status.AtProvider.Endpoint.Address),
		xpv1.ResourceCredentialsSecretPortKey:     []byte(strconv.Itoa(in.Status.AtProvider.Endpoint.Port)),
	}
}

func generateCloudWatchExportConfiguration(spec, current []string) *rdstypes.CloudwatchLogsExportConfiguration {
	toEnable := []string{}
	toDisable := []string{}

	currentMap := make(map[string]struct{}, len(current))
	for _, currentID := range current {
		currentMap[currentID] = struct{}{}
	}

	specMap := make(map[string]struct{}, len(spec))
	for _, specID := range spec {
		specMap[specID] = struct{}{}

		if _, exists := currentMap[specID]; !exists {
			toEnable = append(toEnable, specID)
		}
	}

	for _, currentID := range current {
		if _, exists := specMap[currentID]; !exists {
			toDisable = append(toDisable, currentID)
		}
	}

	return &rdstypes.CloudwatchLogsExportConfiguration{
		EnableLogTypes:  toEnable,
		DisableLogTypes: toDisable,
	}
}

func areSameElements(a1, a2 []string) bool {
	if len(a1) != len(a2) {
		return false
	}

	m2 := make(map[string]struct{}, len(a2))
	for _, s2 := range a2 {
		m2[s2] = struct{}{}
	}

	for _, s1 := range a1 {
		if _, exists := m2[s1]; !exists {
			return false
		}
	}

	return true
}

// DiffTags between spec and current
func DiffTags(spec []v1beta1.Tag, current []rdstypes.Tag) (addTags []rdstypes.Tag, removeTags []string) {
	currentMap := make(map[string]string, len(current))
	for _, t := range current {
		currentMap[pointer.StringValue(t.Key)] = pointer.StringValue(t.Value)
	}

	specMap := make(map[string]string, len(spec))
	for _, t := range spec {
		key := t.Key
		val := t.Value
		specMap[key] = t.Value

		if currentVal, exists := currentMap[key]; exists {
			if currentVal != val {
				addTags = append(addTags, rdstypes.Tag{
					Key:   pointer.ToOrNilIfZeroValue(key),
					Value: pointer.ToOrNilIfZeroValue(val),
				})
			}
		} else {
			addTags = append(addTags, rdstypes.Tag{
				Key:   pointer.ToOrNilIfZeroValue(key),
				Value: pointer.ToOrNilIfZeroValue(val),
			})
		}
	}

	for _, t := range current {
		key := pointer.StringValue(t.Key)
		if _, exists := specMap[key]; !exists {
			removeTags = append(removeTags, key)
		}
	}

	return addTags, removeTags
}
