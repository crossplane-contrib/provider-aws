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

package dbinstance

import (
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	svcsdk "github.com/aws/aws-sdk-go/service/rds"

	"github.com/crossplane-contrib/provider-aws/apis/rds/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
)

// GenerateRestoreDBInstanceFromS3Input from RDSInstanceSpec
func GenerateRestoreDBInstanceFromS3Input(name, password string, p *v1alpha1.DBInstanceParameters) *svcsdk.RestoreDBInstanceFromS3Input {
	// Partially duplicates GenerateCreateDBInstanceInput - make sure any relevant changes are applied there too.
	c := &svcsdk.RestoreDBInstanceFromS3Input{
		DBInstanceIdentifier:               aws.String(name),
		AllocatedStorage:                   p.AllocatedStorage,
		AutoMinorVersionUpgrade:            p.AutoMinorVersionUpgrade,
		AvailabilityZone:                   p.AvailabilityZone,
		BackupRetentionPeriod:              p.BackupRetentionPeriod,
		CopyTagsToSnapshot:                 p.CopyTagsToSnapshot,
		DBInstanceClass:                    p.DBInstanceClass,
		DBName:                             p.DBName,
		DBParameterGroupName:               p.DBParameterGroupName,
		DBSecurityGroups:                   pointer.SliceValueToPtr(p.DBSecurityGroups),
		DBSubnetGroupName:                  p.DBSubnetGroupName,
		DeletionProtection:                 p.DeletionProtection,
		EnableCloudwatchLogsExports:        p.EnableCloudwatchLogsExports,
		EnableIAMDatabaseAuthentication:    p.EnableIAMDatabaseAuthentication,
		EnablePerformanceInsights:          p.EnablePerformanceInsights,
		Engine:                             p.Engine,
		EngineVersion:                      p.EngineVersion,
		Iops:                               p.IOPS,
		KmsKeyId:                           p.KMSKeyID,
		LicenseModel:                       p.LicenseModel,
		MasterUserPassword:                 pointer.ToOrNilIfZeroValue(password),
		MasterUsername:                     p.MasterUsername,
		MonitoringInterval:                 p.MonitoringInterval,
		MonitoringRoleArn:                  p.MonitoringRoleARN,
		MultiAZ:                            p.MultiAZ,
		OptionGroupName:                    p.OptionGroupName,
		PerformanceInsightsKMSKeyId:        p.PerformanceInsightsKMSKeyID,
		PerformanceInsightsRetentionPeriod: p.PerformanceInsightsRetentionPeriod,
		Port:                               p.Port,
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
		VpcSecurityGroupIds:                pointer.SliceValueToPtr(p.VPCSecurityGroupIDs),
	}
	if len(p.ProcessorFeatures) != 0 {
		c.ProcessorFeatures = make([]*svcsdk.ProcessorFeature, len(p.ProcessorFeatures))
		for i, val := range p.ProcessorFeatures {
			c.ProcessorFeatures[i] = &svcsdk.ProcessorFeature{
				Name:  val.Name,
				Value: val.Value,
			}
		}
	}
	if len(p.Tags) != 0 {
		c.Tags = make([]*svcsdk.Tag, len(p.Tags))
		for i, val := range p.Tags {
			c.Tags[i] = &svcsdk.Tag{
				Key:   val.Key,
				Value: val.Value,
			}
		}
	}
	return c
}

// GenerateRestoreDBInstanceFromSnapshotInput from RDSInstanceSpec
func GenerateRestoreDBInstanceFromSnapshotInput(name string, p *v1alpha1.DBInstanceParameters) *svcsdk.RestoreDBInstanceFromDBSnapshotInput {
	// Partially duplicates GenerateCreateDBInstanceInput - make sure any relevant changes are applied there too.
	c := &rds.RestoreDBInstanceFromDBSnapshotInput{
		DBInstanceIdentifier:            aws.String(name),
		AutoMinorVersionUpgrade:         p.AutoMinorVersionUpgrade,
		AvailabilityZone:                p.AvailabilityZone,
		CopyTagsToSnapshot:              p.CopyTagsToSnapshot,
		DBInstanceClass:                 p.DBInstanceClass,
		DBParameterGroupName:            p.DBParameterGroupName,
		DBSnapshotIdentifier:            p.RestoreFrom.Snapshot.SnapshotIdentifier,
		DBSubnetGroupName:               p.DBSubnetGroupName,
		DeletionProtection:              p.DeletionProtection,
		Domain:                          p.Domain,
		DomainIAMRoleName:               p.DomainIAMRoleName,
		EnableCloudwatchLogsExports:     p.EnableCloudwatchLogsExports,
		EnableIAMDatabaseAuthentication: p.EnableIAMDatabaseAuthentication,
		Engine:                          p.Engine,
		Iops:                            p.IOPS,
		LicenseModel:                    p.LicenseModel,
		MultiAZ:                         p.MultiAZ,
		OptionGroupName:                 p.OptionGroupName,
		Port:                            p.Port,
		PubliclyAccessible:              p.PubliclyAccessible,
		StorageType:                     p.StorageType,
		VpcSecurityGroupIds:             pointer.SliceValueToPtr(p.VPCSecurityGroupIDs),
	}
	if len(p.ProcessorFeatures) != 0 {
		c.ProcessorFeatures = make([]*svcsdk.ProcessorFeature, len(p.ProcessorFeatures))
		for i, val := range p.ProcessorFeatures {
			c.ProcessorFeatures[i] = &svcsdk.ProcessorFeature{
				Name:  val.Name,
				Value: val.Value,
			}
		}
	}
	if len(p.Tags) != 0 {
		c.Tags = make([]*svcsdk.Tag, len(p.Tags))
		for i, val := range p.Tags {
			c.Tags[i] = &svcsdk.Tag{
				Key:   val.Key,
				Value: val.Value,
			}
		}
	}
	return c
}

// GenerateRestoreDBInstanceToPointInTimeInput from RDSInstanceSpec
func GenerateRestoreDBInstanceToPointInTimeInput(name string, p *v1alpha1.DBInstanceParameters) *svcsdk.RestoreDBInstanceToPointInTimeInput {
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
		DBInstanceClass:                 p.DBInstanceClass,
		DBName:                          p.DBName,
		DBParameterGroupName:            p.DBParameterGroupName,
		DBSubnetGroupName:               p.DBSubnetGroupName,
		DeletionProtection:              p.DeletionProtection,
		Domain:                          p.Domain,
		DomainIAMRoleName:               p.DomainIAMRoleName,
		EnableCloudwatchLogsExports:     p.EnableCloudwatchLogsExports,
		EnableIAMDatabaseAuthentication: p.EnableIAMDatabaseAuthentication,
		Engine:                          p.Engine,
		Iops:                            p.IOPS,
		LicenseModel:                    p.LicenseModel,
		MultiAZ:                         p.MultiAZ,
		OptionGroupName:                 p.OptionGroupName,
		Port:                            p.Port,
		PubliclyAccessible:              p.PubliclyAccessible,
		StorageType:                     p.StorageType,
		VpcSecurityGroupIds:             pointer.SliceValueToPtr(p.VPCSecurityGroupIDs),

		TargetDBInstanceIdentifier:          aws.String(name),
		RestoreTime:                         restoreTime,
		UseLatestRestorableTime:             aws.Bool(p.RestoreFrom.PointInTime.UseLatestRestorableTime),
		SourceDBInstanceAutomatedBackupsArn: p.RestoreFrom.PointInTime.SourceDBInstanceAutomatedBackupsArn,
		SourceDBInstanceIdentifier:          p.RestoreFrom.PointInTime.SourceDBInstanceIdentifier,
		SourceDbiResourceId:                 p.RestoreFrom.PointInTime.SourceDbiResourceID,
	}
	if len(p.ProcessorFeatures) != 0 {
		c.ProcessorFeatures = make([]*svcsdk.ProcessorFeature, len(p.ProcessorFeatures))
		for i, val := range p.ProcessorFeatures {
			c.ProcessorFeatures[i] = &svcsdk.ProcessorFeature{
				Name:  val.Name,
				Value: val.Value,
			}
		}
	}
	if len(p.Tags) != 0 {
		c.Tags = make([]*svcsdk.Tag, len(p.Tags))
		for i, val := range p.Tags {
			c.Tags[i] = &svcsdk.Tag{
				Key:   val.Key,
				Value: val.Value,
			}
		}
	}
	return c
}

// GenerateCreateDBInstanceReadReplicaInput returns a create input.
func GenerateCreateDBInstanceReadReplicaInput(cr *v1alpha1.DBInstance) *svcsdk.CreateDBInstanceReadReplicaInput { //nolint:gocyclo
	res := &svcsdk.CreateDBInstanceReadReplicaInput{}

	if cr.Spec.ForProvider.AllocatedStorage != nil {
		res.SetAllocatedStorage(*cr.Spec.ForProvider.AllocatedStorage)
	}
	if cr.Spec.ForProvider.AutoMinorVersionUpgrade != nil {
		res.SetAutoMinorVersionUpgrade(*cr.Spec.ForProvider.AutoMinorVersionUpgrade)
	}
	if cr.Spec.ForProvider.AvailabilityZone != nil {
		res.SetAvailabilityZone(*cr.Spec.ForProvider.AvailabilityZone)
	}
	if cr.Spec.ForProvider.CopyTagsToSnapshot != nil {
		res.SetCopyTagsToSnapshot(*cr.Spec.ForProvider.CopyTagsToSnapshot)
	}
	if cr.Spec.ForProvider.CustomIAMInstanceProfile != nil {
		res.SetCustomIamInstanceProfile(*cr.Spec.ForProvider.CustomIAMInstanceProfile)
	}
	if cr.Spec.ForProvider.DBInstanceClass != nil {
		res.SetDBInstanceClass(*cr.Spec.ForProvider.DBInstanceClass)
	}
	if cr.Spec.ForProvider.DBParameterGroupName != nil {
		res.SetDBParameterGroupName(*cr.Spec.ForProvider.DBParameterGroupName)
	}
	if cr.Spec.ForProvider.DBSubnetGroupName != nil {
		res.SetDBSubnetGroupName(*cr.Spec.ForProvider.DBSubnetGroupName)
	}
	if cr.Spec.ForProvider.DedicatedLogVolume != nil {
		res.SetDedicatedLogVolume(*cr.Spec.ForProvider.DedicatedLogVolume)
	}
	if cr.Spec.ForProvider.DeletionProtection != nil {
		res.SetDeletionProtection(*cr.Spec.ForProvider.DeletionProtection)
	}
	if cr.Spec.ForProvider.Domain != nil {
		res.SetDomain(*cr.Spec.ForProvider.Domain)
	}
	if cr.Spec.ForProvider.DomainAuthSecretARN != nil {
		res.SetDomainAuthSecretArn(*cr.Spec.ForProvider.DomainAuthSecretARN)
	}
	if cr.Spec.ForProvider.DomainDNSIPs != nil {
		res.SetDomainDnsIps(cr.Spec.ForProvider.DomainDNSIPs)
	}
	if cr.Spec.ForProvider.DomainFqdn != nil {
		res.SetDomainFqdn(*cr.Spec.ForProvider.DomainFqdn)
	}
	if cr.Spec.ForProvider.DomainIAMRoleName != nil {
		res.SetDomainIAMRoleName(*cr.Spec.ForProvider.DomainIAMRoleName)
	}
	if cr.Spec.ForProvider.DomainOu != nil {
		res.SetDomainOu(*cr.Spec.ForProvider.DomainOu)
	}
	if cr.Spec.ForProvider.EnableCloudwatchLogsExports != nil {
		res.SetEnableCloudwatchLogsExports(cr.Spec.ForProvider.EnableCloudwatchLogsExports)
	}
	if cr.Spec.ForProvider.EnableCustomerOwnedIP != nil {
		res.SetEnableCustomerOwnedIp(*cr.Spec.ForProvider.EnableCustomerOwnedIP)
	}
	if cr.Spec.ForProvider.EnableIAMDatabaseAuthentication != nil {
		res.SetEnableIAMDatabaseAuthentication(*cr.Spec.ForProvider.EnableIAMDatabaseAuthentication)
	}
	if cr.Spec.ForProvider.EnablePerformanceInsights != nil {
		res.SetEnablePerformanceInsights(*cr.Spec.ForProvider.EnablePerformanceInsights)
	}
	if cr.Spec.ForProvider.IOPS != nil {
		res.SetIops(*cr.Spec.ForProvider.IOPS)
	}
	if cr.Spec.ForProvider.KMSKeyID != nil {
		res.SetKmsKeyId(*cr.Spec.ForProvider.KMSKeyID)

	}
	if cr.Spec.ForProvider.MaxAllocatedStorage != nil {
		res.SetMaxAllocatedStorage(*cr.Spec.ForProvider.MaxAllocatedStorage)
	}
	if cr.Spec.ForProvider.MonitoringInterval != nil {
		res.SetMonitoringInterval(*cr.Spec.ForProvider.MonitoringInterval)
	}
	if cr.Spec.ForProvider.MonitoringRoleARN != nil {
		res.SetMonitoringRoleArn(*cr.Spec.ForProvider.MonitoringRoleARN)
	}
	if cr.Spec.ForProvider.MultiAZ != nil {
		res.SetMultiAZ(*cr.Spec.ForProvider.MultiAZ)
	}
	if cr.Spec.ForProvider.NetworkType != nil {
		res.SetNetworkType(*cr.Spec.ForProvider.NetworkType)
	}
	if cr.Spec.ForProvider.OptionGroupName != nil {
		res.SetOptionGroupName(*cr.Spec.ForProvider.OptionGroupName)
	}
	if cr.Spec.ForProvider.PerformanceInsightsKMSKeyID != nil {
		res.SetPerformanceInsightsKMSKeyId(*cr.Spec.ForProvider.PerformanceInsightsKMSKeyID)
	}
	if cr.Spec.ForProvider.PerformanceInsightsRetentionPeriod != nil {
		res.SetPerformanceInsightsRetentionPeriod(*cr.Spec.ForProvider.PerformanceInsightsRetentionPeriod)
	}
	if cr.Spec.ForProvider.Port != nil {
		res.SetPort(*cr.Spec.ForProvider.Port)
	}
	if cr.Spec.ForProvider.ProcessorFeatures != nil {
		var processorFeatures []*svcsdk.ProcessorFeature
		for _, pf := range cr.Spec.ForProvider.ProcessorFeatures {
			pfeature := &svcsdk.ProcessorFeature{}
			if pf.Name != nil {
				pfeature.SetName(*pf.Name)
			}
			if pf.Value != nil {
				pfeature.SetValue(*pf.Value)
			}
			processorFeatures = append(processorFeatures, pfeature)
		}
		res.SetProcessorFeatures(processorFeatures)
	}
	if cr.Spec.ForProvider.PubliclyAccessible != nil {
		res.SetPubliclyAccessible(*cr.Spec.ForProvider.PubliclyAccessible)
	}
	if cr.Spec.ForProvider.SourceDBClusterID != nil {
		res.SetSourceDBClusterIdentifier(*cr.Spec.ForProvider.SourceDBClusterID)
	}
	if cr.Spec.ForProvider.SourceDBInstanceID != nil {
		res.SetSourceDBInstanceIdentifier(*cr.Spec.ForProvider.SourceDBInstanceID)
	}
	if cr.Spec.ForProvider.StorageThroughput != nil {
		res.SetStorageThroughput(*cr.Spec.ForProvider.StorageThroughput)
	}
	if cr.Spec.ForProvider.StorageType != nil {
		res.SetStorageType(*cr.Spec.ForProvider.StorageType)
	}
	if cr.Spec.ForProvider.Tags != nil {
		var tags []*svcsdk.Tag
		for _, t := range cr.Spec.ForProvider.Tags {
			tag := &svcsdk.Tag{}
			if t.Key != nil {
				tag.SetKey(*t.Key)
			}
			if t.Value != nil {
				tag.SetValue(*t.Value)
			}
			tags = append(tags, tag)
		}
		res.SetTags(tags)
	}
	if cr.Spec.ForProvider.VPCSecurityGroupIDs != nil {
		var vpcSecurityGroupIDs []*string
		for _, v := range cr.Spec.ForProvider.VPCSecurityGroupIDs {
			vpcSecurityGroupIDs = append(vpcSecurityGroupIDs, &v)
		}
		res.SetVpcSecurityGroupIds(vpcSecurityGroupIDs)
	}
	return res
}
