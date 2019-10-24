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
	"strings"

	"github.com/crossplaneio/crossplane-runtime/pkg/resource"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"

	"github.com/crossplaneio/stack-aws/apis/database/v1alpha2"
	awsclients "github.com/crossplaneio/stack-aws/pkg/clients"
)

// Client defines RDS RDSClient operations
type Client interface {
	CreateDBInstanceRequest(*rds.CreateDBInstanceInput) rds.CreateDBInstanceRequest
	DescribeDBInstancesRequest(*rds.DescribeDBInstancesInput) rds.DescribeDBInstancesRequest
	ModifyDBInstanceRequest(*rds.ModifyDBInstanceInput) rds.ModifyDBInstanceRequest
	DeleteDBInstanceRequest(*rds.DeleteDBInstanceInput) rds.DeleteDBInstanceRequest
}

// NewClient creates new RDS RDSClient with provided AWS Configurations/Credentials
func NewClient(credentials []byte, region string) (Client, error) {
	cfg, err := awsclients.LoadConfig(credentials, awsclients.DefaultSection, region)
	if err != nil {
		return nil, err
	}
	return rds.New(*cfg), nil
}

// IsErrorAlreadyExists returns true if the supplied error indicates an instance
// already exists.
func IsErrorAlreadyExists(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), rds.ErrCodeDBInstanceAlreadyExistsFault)
}

// IsErrorNotFound helper function to test for ErrCodeDBInstanceNotFoundFault error
func IsErrorNotFound(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), rds.ErrCodeDBInstanceNotFoundFault)
}

// GenerateCreateDBInstanceInput from RDSInstanceSpec
func GenerateCreateDBInstanceInput(name, password string, p *v1alpha2.RDSInstanceParameters) *rds.CreateDBInstanceInput {
	return &rds.CreateDBInstanceInput{
		DBInstanceIdentifier:               aws.String(name),
		AllocatedStorage:                   awsclients.Int64Address(p.AllocatedStorage),
		AutoMinorVersionUpgrade:            p.AutoMinorVersionUpgrade,
		AvailabilityZone:                   p.AvailabilityZone,
		BackupRetentionPeriod:              aws.Int64(0),
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
		Iops:                               awsclients.Int64Address(p.IOPS),
		KmsKeyId:                           p.KMSKeyID,
		LicenseModel:                       p.LicenseModel,
		MasterUserPassword:                 aws.String(password),
		MasterUsername:                     p.MasterUsername,
		MonitoringInterval:                 awsclients.Int64Address(p.MonitoringInterval),
		MonitoringRoleArn:                  p.MonitoringRoleArn,
		MultiAZ:                            p.MultiAZ,
		OptionGroupName:                    p.OptionGroupName,
		PerformanceInsightsKMSKeyId:        p.PerformanceInsightsKMSKeyID,
		PerformanceInsightsRetentionPeriod: awsclients.Int64Address(p.PerformanceInsightsRetentionPeriod),
		Port:                               awsclients.Int64Address(p.Port),
		PreferredBackupWindow:              p.PreferredBackupWindow,
		PreferredMaintenanceWindow:         p.PreferredMaintenanceWindow,
		ProcessorFeatures:                  convertProcessorFeatures(p.ProcessorFeatures),
		PromotionTier:                      awsclients.Int64Address(p.PromotionTier),
		PubliclyAccessible:                 p.PubliclyAccessible,
		StorageEncrypted:                   p.StorageEncrypted,
		Tags:                               convertTags(p.Tags),
		TdeCredentialArn:                   p.TdeCredentialArn,
		TdeCredentialPassword:              p.TdeCredentialPassword,
		Timezone:                           p.Timezone,
		StorageType:                        p.StorageType,
		VpcSecurityGroupIds:                p.VPCSecurityGroupIDs,
	}
}

// GenerateModifyDBInstanceInput from RDSInstanceSpec
func GenerateModifyDBInstanceInput(name string, p *v1alpha2.RDSInstanceParameters) *rds.ModifyDBInstanceInput {
	// NOTE(muvaf): MasterUserPassword is not used here. So, password is set once
	// and kept that way.
	// NOTE(muvaf): Change of DBInstanceIdentifier is supported by AWS but
	// Crossplane assumes identification info never changes, so, we don't support
	// it.
	return &rds.ModifyDBInstanceInput{
		DBInstanceIdentifier:               aws.String(name),
		AllocatedStorage:                   awsclients.Int64Address(p.AllocatedStorage),
		AllowMajorVersionUpgrade:           p.AllowMajorVersionUpgrade,
		ApplyImmediately:                   p.ApplyModificationsImmediately,
		AutoMinorVersionUpgrade:            p.AutoMinorVersionUpgrade,
		BackupRetentionPeriod:              awsclients.Int64Address(p.BackupRetentionPeriod),
		CACertificateIdentifier:            p.CACertificateIdentifier,
		CloudwatchLogsExportConfiguration:  convertCloudwatchLogsExportConfiguration(p.CloudwatchLogsExportConfiguration),
		CopyTagsToSnapshot:                 p.CopyTagsToSnapshot,
		DBInstanceClass:                    aws.String(p.DBInstanceClass),
		DBParameterGroupName:               p.DBParameterGroupName,
		DBPortNumber:                       awsclients.Int64Address(p.Port),
		DBSecurityGroups:                   p.DBSecurityGroups,
		DBSubnetGroupName:                  p.DBSubnetGroupName,
		DeletionProtection:                 p.DeletionProtection,
		Domain:                             p.Domain,
		DomainIAMRoleName:                  p.DomainIAMRoleName,
		EnableIAMDatabaseAuthentication:    p.EnableIAMDatabaseAuthentication,
		EnablePerformanceInsights:          p.EnablePerformanceInsights,
		EngineVersion:                      p.EngineVersion,
		Iops:                               awsclients.Int64Address(p.IOPS),
		LicenseModel:                       p.LicenseModel,
		MonitoringInterval:                 awsclients.Int64Address(p.MonitoringInterval),
		MonitoringRoleArn:                  p.MonitoringRoleArn,
		MultiAZ:                            p.MultiAZ,
		OptionGroupName:                    p.OptionGroupName,
		PerformanceInsightsKMSKeyId:        p.PerformanceInsightsKMSKeyID,
		PerformanceInsightsRetentionPeriod: awsclients.Int64Address(p.PerformanceInsightsRetentionPeriod),
		PreferredBackupWindow:              p.PreferredBackupWindow,
		PreferredMaintenanceWindow:         p.PreferredMaintenanceWindow,
		ProcessorFeatures:                  convertProcessorFeatures(p.ProcessorFeatures),
		PromotionTier:                      awsclients.Int64Address(p.PromotionTier),
		PubliclyAccessible:                 p.PubliclyAccessible,
		StorageType:                        p.StorageType,
		TdeCredentialArn:                   p.TdeCredentialArn,
		TdeCredentialPassword:              p.TdeCredentialPassword,
		UseDefaultProcessorFeatures:        p.UseDefaultProcessorFeatures,
		VpcSecurityGroupIds:                p.VPCSecurityGroupIDs,
	}
}

func GenerateObservation(db rds.DBInstance) v1alpha2.RDSInstanceObservation {
	return v1alpha2.RDSInstanceObservation{}
}

func LateInitialize(in *v1alpha2.RDSInstanceParameters, db rds.DBInstance) {

}

func NeedsUpdate(in v1alpha2.RDSInstanceParameters, db rds.DBInstance) bool {
	return true
}

func GetConnectionDetails(in v1alpha2.RDSInstance) resource.ConnectionDetails {
	return resource.ConnectionDetails{}
}

func convertProcessorFeatures(in []v1alpha2.ProcessorFeature) []rds.ProcessorFeature {
	if len(in) == 0 {
		return nil
	}
	out := make([]rds.ProcessorFeature, len(in))
	for i, f := range in {
		out[i] = rds.ProcessorFeature{
			Name:  aws.String(f.Name),
			Value: aws.String(f.Value),
		}
	}
	return out
}

func convertTags(in []v1alpha2.Tag) []rds.Tag {
	if len(in) == 0 {
		return nil
	}
	out := make([]rds.Tag, len(in))
	for i, f := range in {
		out[i] = rds.Tag{
			Key:   aws.String(f.Key),
			Value: aws.String(f.Value),
		}
	}
	return out
}

func convertCloudwatchLogsExportConfiguration(in *v1alpha2.CloudwatchLogsExportConfiguration) *rds.CloudwatchLogsExportConfiguration {
	if in == nil {
		return nil
	}
	out := &rds.CloudwatchLogsExportConfiguration{}
	if len(in.DisableLogTypes) != 0 {
		out.DisableLogTypes = make([]string, len(in.DisableLogTypes))
		for i, s := range in.DisableLogTypes {
			out.DisableLogTypes[i] = s
		}
	}
	if len(in.EnableLogTypes) != 0 {
		out.EnableLogTypes = make([]string, len(in.EnableLogTypes))
		for i, s := range in.EnableLogTypes {
			out.EnableLogTypes[i] = s
		}
	}
	return out
}
