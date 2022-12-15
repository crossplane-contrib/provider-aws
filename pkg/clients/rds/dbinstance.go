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
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	svcsdk "github.com/aws/aws-sdk-go/service/rds"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane-contrib/provider-aws/apis/rds/v1alpha1"
	awsclients "github.com/crossplane-contrib/provider-aws/pkg/clients"
)

const (
	errGetPasswordSecretFailed = "cannot get password secret"
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
		DBSecurityGroups:                   awsclients.StringSliceToPtr(p.DBSecurityGroups),
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
		MasterUserPassword:                 awsclients.String(password),
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
		VpcSecurityGroupIds:                awsclients.StringSliceToPtr(p.VPCSecurityGroupIDs),
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
		VpcSecurityGroupIds:             awsclients.StringSliceToPtr(p.VPCSecurityGroupIDs),
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
		VpcSecurityGroupIds:             awsclients.StringSliceToPtr(p.VPCSecurityGroupIDs),

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

// // GetConnectionDetails extracts managed.ConnectionDetails out of v1beta1.RDSInstance.
// func GetConnectionDetails(in v1beta1.RDSInstance) managed.ConnectionDetails {
// 	if in.Status.AtProvider.Endpoint.Address == "" {
// 		return nil
// 	}
// 	return managed.ConnectionDetails{
// 		xpv1.ResourceCredentialsSecretEndpointKey: []byte(in.Status.AtProvider.Endpoint.Address),
// 		xpv1.ResourceCredentialsSecretPortKey:     []byte(strconv.Itoa(in.Status.AtProvider.Endpoint.Port)),
// 	}
// }
