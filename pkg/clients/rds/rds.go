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
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"

	"github.com/crossplane/provider-aws/apis/database/v1beta1"
	awsclients "github.com/crossplane/provider-aws/pkg/clients"
)

// Client defines RDS RDSClient operations
type Client interface {
	CreateDBInstanceRequest(*rds.CreateDBInstanceInput) rds.CreateDBInstanceRequest
	DescribeDBInstancesRequest(*rds.DescribeDBInstancesInput) rds.DescribeDBInstancesRequest
	ModifyDBInstanceRequest(*rds.ModifyDBInstanceInput) rds.ModifyDBInstanceRequest
	DeleteDBInstanceRequest(*rds.DeleteDBInstanceInput) rds.DeleteDBInstanceRequest
	AddTagsToResourceRequest(*rds.AddTagsToResourceInput) rds.AddTagsToResourceRequest
}

// NewClient creates new RDS RDSClient with provided AWS Configurations/Credentials
func NewClient(ctx context.Context, credentials []byte, region string, auth awsclients.AuthMethod) (Client, error) {
	cfg, err := auth(ctx, credentials, awsclients.DefaultSection, region)
	if cfg == nil {
		return nil, err
	}
	return rds.New(*cfg), err
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
func GenerateCreateDBInstanceInput(name, password string, p *v1beta1.RDSInstanceParameters) *rds.CreateDBInstanceInput {
	c := &rds.CreateDBInstanceInput{
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
		MonitoringRoleArn:                  p.MonitoringRoleARN,
		MultiAZ:                            p.MultiAZ,
		OptionGroupName:                    p.OptionGroupName,
		PerformanceInsightsKMSKeyId:        p.PerformanceInsightsKMSKeyID,
		PerformanceInsightsRetentionPeriod: awsclients.Int64Address(p.PerformanceInsightsRetentionPeriod),
		Port:                               awsclients.Int64Address(p.Port),
		PreferredBackupWindow:              p.PreferredBackupWindow,
		PreferredMaintenanceWindow:         p.PreferredMaintenanceWindow,
		PromotionTier:                      awsclients.Int64Address(p.PromotionTier),
		PubliclyAccessible:                 p.PubliclyAccessible,
		StorageEncrypted:                   p.StorageEncrypted,
		Timezone:                           p.Timezone,
		StorageType:                        p.StorageType,
		VpcSecurityGroupIds:                p.VPCSecurityGroupIDs,
	}
	if len(p.ProcessorFeatures) != 0 {
		c.ProcessorFeatures = make([]rds.ProcessorFeature, len(p.ProcessorFeatures))
		for i, val := range p.ProcessorFeatures {
			c.ProcessorFeatures[i] = rds.ProcessorFeature{
				Name:  aws.String(val.Name),
				Value: aws.String(val.Value),
			}
		}
	}
	if len(p.Tags) != 0 {
		c.Tags = make([]rds.Tag, len(p.Tags))
		for i, val := range p.Tags {
			c.Tags[i] = rds.Tag{
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
func CreatePatch(in *rds.DBInstance, target *v1beta1.RDSInstanceParameters) (*v1beta1.RDSInstanceParameters, error) {
	currentParams := &v1beta1.RDSInstanceParameters{}
	LateInitialize(currentParams, in)

	jsonPatch, err := awsclients.CreateJSONPatch(currentParams, target)
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
func GenerateModifyDBInstanceInput(name string, p *v1beta1.RDSInstanceParameters) *rds.ModifyDBInstanceInput {
	// NOTE(muvaf): MasterUserPassword is not used here. So, password is set once
	// and kept that way.
	// NOTE(muvaf): Change of DBInstanceIdentifier is supported by AWS but
	// Crossplane assumes identification info never changes, so, we don't support
	// it.
	m := &rds.ModifyDBInstanceInput{
		DBInstanceIdentifier:               aws.String(name),
		AllocatedStorage:                   awsclients.Int64Address(p.AllocatedStorage),
		AllowMajorVersionUpgrade:           p.AllowMajorVersionUpgrade,
		ApplyImmediately:                   p.ApplyModificationsImmediately,
		AutoMinorVersionUpgrade:            p.AutoMinorVersionUpgrade,
		BackupRetentionPeriod:              awsclients.Int64Address(p.BackupRetentionPeriod),
		CACertificateIdentifier:            p.CACertificateIdentifier,
		CopyTagsToSnapshot:                 p.CopyTagsToSnapshot,
		DBInstanceClass:                    awsclients.String(p.DBInstanceClass),
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
		MonitoringRoleArn:                  p.MonitoringRoleARN,
		MultiAZ:                            p.MultiAZ,
		OptionGroupName:                    p.OptionGroupName,
		PerformanceInsightsKMSKeyId:        p.PerformanceInsightsKMSKeyID,
		PerformanceInsightsRetentionPeriod: awsclients.Int64Address(p.PerformanceInsightsRetentionPeriod),
		PreferredBackupWindow:              p.PreferredBackupWindow,
		PreferredMaintenanceWindow:         p.PreferredMaintenanceWindow,
		PromotionTier:                      awsclients.Int64Address(p.PromotionTier),
		PubliclyAccessible:                 p.PubliclyAccessible,
		StorageType:                        p.StorageType,
		UseDefaultProcessorFeatures:        p.UseDefaultProcessorFeatures,
		VpcSecurityGroupIds:                p.VPCSecurityGroupIDs,
	}
	if len(p.ProcessorFeatures) != 0 {
		m.ProcessorFeatures = make([]rds.ProcessorFeature, len(p.ProcessorFeatures))
		for i, val := range p.ProcessorFeatures {
			m.ProcessorFeatures[i] = rds.ProcessorFeature{
				Name:  aws.String(val.Name),
				Value: aws.String(val.Value),
			}
		}
	}
	if p.CloudwatchLogsExportConfiguration != nil {
		m.CloudwatchLogsExportConfiguration = &rds.CloudwatchLogsExportConfiguration{
			DisableLogTypes: p.CloudwatchLogsExportConfiguration.DisableLogTypes,
			EnableLogTypes:  p.CloudwatchLogsExportConfiguration.EnableLogTypes,
		}
	}
	return m
}

// GenerateObservation is used to produce v1alpha3.RDSInstanceObservation from
// rds.DBInstance.
func GenerateObservation(db rds.DBInstance) v1beta1.RDSInstanceObservation { // nolint:gocyclo
	o := v1beta1.RDSInstanceObservation{
		DBInstanceStatus:                      aws.StringValue(db.DBInstanceStatus),
		DBInstanceArn:                         aws.StringValue(db.DBInstanceArn),
		DBInstancePort:                        int(aws.Int64Value(db.DbInstancePort)),
		DBResourceID:                          aws.StringValue(db.DbiResourceId),
		EnhancedMonitoringResourceArn:         aws.StringValue(db.EnhancedMonitoringResourceArn),
		PerformanceInsightsEnabled:            aws.BoolValue(db.PerformanceInsightsEnabled),
		ReadReplicaDBClusterIdentifiers:       db.ReadReplicaDBClusterIdentifiers,
		ReadReplicaDBInstanceIdentifiers:      db.ReadReplicaDBInstanceIdentifiers,
		ReadReplicaSourceDBInstanceIdentifier: aws.StringValue(db.ReadReplicaSourceDBInstanceIdentifier),
		SecondaryAvailabilityZone:             aws.StringValue(db.SecondaryAvailabilityZone),
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
				DBParameterGroupName: aws.StringValue(val.DBParameterGroupName),
				ParameterApplyStatus: aws.StringValue(val.ParameterApplyStatus),
			}
		}
	}
	if len(db.DBSecurityGroups) != 0 {
		o.DBSecurityGroups = make([]v1beta1.DBSecurityGroupMembership, len(db.DBSecurityGroups))
		for i, val := range db.DBSecurityGroups {
			o.DBSecurityGroups[i] = v1beta1.DBSecurityGroupMembership{
				DBSecurityGroupName: aws.StringValue(val.DBSecurityGroupName),
				Status:              aws.StringValue(val.Status),
			}
		}
	}
	if db.DBSubnetGroup != nil {
		o.DBSubnetGroup = v1beta1.DBSubnetGroupInRDS{
			DBSubnetGroupARN:         aws.StringValue(db.DBSubnetGroup.DBSubnetGroupArn),
			DBSubnetGroupDescription: aws.StringValue(db.DBSubnetGroup.DBSubnetGroupDescription),
			DBSubnetGroupName:        aws.StringValue(db.DBSubnetGroup.DBSubnetGroupName),
			SubnetGroupStatus:        aws.StringValue(db.DBSubnetGroup.SubnetGroupStatus),
			VPCID:                    aws.StringValue(db.DBSubnetGroup.VpcId),
		}
		if len(db.DBSubnetGroup.Subnets) != 0 {
			o.DBSubnetGroup.Subnets = make([]v1beta1.SubnetInRDS, len(db.DBSubnetGroup.Subnets))
			for i, val := range db.DBSubnetGroup.Subnets {
				o.DBSubnetGroup.Subnets[i] = v1beta1.SubnetInRDS{
					SubnetIdentifier: aws.StringValue(val.SubnetIdentifier),
					SubnetStatus:     aws.StringValue(val.SubnetStatus),
				}
				if val.SubnetAvailabilityZone != nil {
					o.DBSubnetGroup.Subnets[i].SubnetAvailabilityZone = v1beta1.AvailabilityZone{
						Name: aws.StringValue(val.SubnetAvailabilityZone.Name),
					}
				}
			}
		}
	}
	if len(db.DomainMemberships) != 0 {
		o.DomainMemberships = make([]v1beta1.DomainMembership, len(db.DomainMemberships))
		for i, val := range db.DomainMemberships {
			o.DomainMemberships[i] = v1beta1.DomainMembership{
				Domain:      aws.StringValue(val.Domain),
				FQDN:        aws.StringValue(val.FQDN),
				IAMRoleName: aws.StringValue(val.IAMRoleName),
				Status:      aws.StringValue(val.Status),
			}
		}
	}
	if db.Endpoint != nil {
		o.Endpoint = v1beta1.Endpoint{
			Address:      aws.StringValue(db.Endpoint.Address),
			HostedZoneID: aws.StringValue(db.Endpoint.HostedZoneId),
			Port:         int(aws.Int64Value(db.Endpoint.Port)),
		}
	}
	if len(db.OptionGroupMemberships) != 0 {
		o.OptionGroupMemberships = make([]v1beta1.OptionGroupMembership, len(db.OptionGroupMemberships))
		for i, val := range db.OptionGroupMemberships {
			o.OptionGroupMemberships[i] = v1beta1.OptionGroupMembership{
				OptionGroupName: aws.StringValue(val.OptionGroupName),
				Status:          aws.StringValue(val.Status),
			}
		}
	}
	if db.PendingModifiedValues != nil {
		o.PendingModifiedValues = v1beta1.PendingModifiedValues{
			AllocatedStorage:        int(aws.Int64Value(db.PendingModifiedValues.AllocatedStorage)),
			BackupRetentionPeriod:   int(aws.Int64Value(db.PendingModifiedValues.BackupRetentionPeriod)),
			CACertificateIdentifier: aws.StringValue(db.PendingModifiedValues.CACertificateIdentifier),
			DBInstanceClass:         aws.StringValue(db.PendingModifiedValues.DBInstanceClass),
			DBSubnetGroupName:       aws.StringValue(db.PendingModifiedValues.DBSubnetGroupName),
			IOPS:                    int(aws.Int64Value(db.PendingModifiedValues.Iops)),
			LicenseModel:            aws.StringValue(db.PendingModifiedValues.LicenseModel),
			MultiAZ:                 aws.BoolValue(db.PendingModifiedValues.MultiAZ),
			Port:                    int(aws.Int64Value(db.PendingModifiedValues.Port)),
			StorageType:             aws.StringValue(db.PendingModifiedValues.StorageType),
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
					Name:  aws.StringValue(val.Name),
					Value: aws.StringValue(val.Value),
				}
			}
		}
	}
	if len(db.StatusInfos) != 0 {
		o.StatusInfos = make([]v1beta1.DBInstanceStatusInfo, len(db.StatusInfos))
		for i, val := range db.StatusInfos {
			o.StatusInfos[i] = v1beta1.DBInstanceStatusInfo{
				Message:    aws.StringValue(val.Message),
				Status:     aws.StringValue(val.Status),
				StatusType: aws.StringValue(val.StatusType),
				Normal:     aws.BoolValue(val.Normal),
			}
		}
	}
	if len(db.VpcSecurityGroups) != 0 {
		o.VPCSecurityGroups = make([]v1beta1.VPCSecurityGroupMembership, len(db.VpcSecurityGroups))
		for i, val := range db.VpcSecurityGroups {
			o.VPCSecurityGroups[i] = v1beta1.VPCSecurityGroupMembership{
				Status:             aws.StringValue(val.Status),
				VPCSecurityGroupID: aws.StringValue(val.VpcSecurityGroupId),
			}
		}
	}
	return o
}

// LateInitialize fills the empty fields in *v1alpha3.RDSInstanceParameters with
// the values seen in rds.DBInstance.
func LateInitialize(in *v1beta1.RDSInstanceParameters, db *rds.DBInstance) { // nolint:gocyclo
	if db == nil {
		return
	}
	in.DBInstanceClass = awsclients.LateInitializeString(in.DBInstanceClass, db.DBInstanceClass)
	in.Engine = awsclients.LateInitializeString(in.Engine, db.Engine)

	in.AllocatedStorage = awsclients.LateInitializeIntPtr(in.AllocatedStorage, db.AllocatedStorage)
	in.AutoMinorVersionUpgrade = awsclients.LateInitializeBoolPtr(in.AutoMinorVersionUpgrade, db.AutoMinorVersionUpgrade)
	in.AvailabilityZone = awsclients.LateInitializeStringPtr(in.AvailabilityZone, db.AvailabilityZone)
	in.BackupRetentionPeriod = awsclients.LateInitializeIntPtr(in.BackupRetentionPeriod, db.BackupRetentionPeriod)
	in.CACertificateIdentifier = awsclients.LateInitializeStringPtr(in.CACertificateIdentifier, db.CACertificateIdentifier)
	in.CharacterSetName = awsclients.LateInitializeStringPtr(in.CharacterSetName, db.CharacterSetName)
	in.CopyTagsToSnapshot = awsclients.LateInitializeBoolPtr(in.CopyTagsToSnapshot, db.CopyTagsToSnapshot)
	in.DBClusterIdentifier = awsclients.LateInitializeStringPtr(in.DBClusterIdentifier, db.DBClusterIdentifier)
	in.DBName = awsclients.LateInitializeStringPtr(in.DBName, db.DBName)
	in.DeletionProtection = awsclients.LateInitializeBoolPtr(in.DeletionProtection, db.DeletionProtection)
	in.EnableIAMDatabaseAuthentication = awsclients.LateInitializeBoolPtr(in.EnableIAMDatabaseAuthentication, db.IAMDatabaseAuthenticationEnabled)
	in.EnablePerformanceInsights = awsclients.LateInitializeBoolPtr(in.EnablePerformanceInsights, db.PerformanceInsightsEnabled)
	in.IOPS = awsclients.LateInitializeIntPtr(in.IOPS, db.Iops)
	in.KMSKeyID = awsclients.LateInitializeStringPtr(in.KMSKeyID, db.KmsKeyId)
	in.LicenseModel = awsclients.LateInitializeStringPtr(in.LicenseModel, db.LicenseModel)
	in.MasterUsername = awsclients.LateInitializeStringPtr(in.MasterUsername, db.MasterUsername)
	in.MonitoringInterval = awsclients.LateInitializeIntPtr(in.MonitoringInterval, db.MonitoringInterval)
	in.MonitoringRoleARN = awsclients.LateInitializeStringPtr(in.MonitoringRoleARN, db.MonitoringRoleArn)
	in.MultiAZ = awsclients.LateInitializeBoolPtr(in.MultiAZ, db.MultiAZ)
	in.PerformanceInsightsKMSKeyID = awsclients.LateInitializeStringPtr(in.PerformanceInsightsKMSKeyID, db.PerformanceInsightsKMSKeyId)
	in.PerformanceInsightsRetentionPeriod = awsclients.LateInitializeIntPtr(in.PerformanceInsightsRetentionPeriod, db.PerformanceInsightsRetentionPeriod)
	in.Port = awsclients.LateInitializeIntPtr(in.Port, db.DbInstancePort)
	in.PreferredBackupWindow = awsclients.LateInitializeStringPtr(in.PreferredBackupWindow, db.PreferredBackupWindow)
	in.PreferredMaintenanceWindow = awsclients.LateInitializeStringPtr(in.PreferredMaintenanceWindow, db.PreferredMaintenanceWindow)
	in.PromotionTier = awsclients.LateInitializeIntPtr(in.PromotionTier, db.PromotionTier)
	in.PubliclyAccessible = awsclients.LateInitializeBoolPtr(in.PubliclyAccessible, db.PubliclyAccessible)
	in.StorageEncrypted = awsclients.LateInitializeBoolPtr(in.StorageEncrypted, db.StorageEncrypted)
	in.StorageType = awsclients.LateInitializeStringPtr(in.StorageType, db.StorageType)
	in.Timezone = awsclients.LateInitializeStringPtr(in.Timezone, db.Timezone)

	if len(in.DBSecurityGroups) == 0 && len(db.DBSecurityGroups) != 0 {
		in.DBSecurityGroups = make([]string, len(db.DBSecurityGroups))
		for i, val := range db.DBSecurityGroups {
			in.DBSecurityGroups[i] = aws.StringValue(val.DBSecurityGroupName)
		}
	}
	if aws.StringValue(in.DBSubnetGroupName) == "" && db.DBSubnetGroup != nil {
		in.DBSubnetGroupName = db.DBSubnetGroup.DBSubnetGroupName
	}
	if len(in.EnableCloudwatchLogsExports) == 0 && len(db.EnabledCloudwatchLogsExports) != 0 {
		in.EnableCloudwatchLogsExports = db.EnabledCloudwatchLogsExports
	}
	if len(in.ProcessorFeatures) == 0 && len(db.ProcessorFeatures) != 0 {
		in.ProcessorFeatures = make([]v1beta1.ProcessorFeature, len(db.ProcessorFeatures))
		for i, val := range db.ProcessorFeatures {
			in.ProcessorFeatures[i] = v1beta1.ProcessorFeature{
				Name:  aws.StringValue(val.Name),
				Value: aws.StringValue(val.Value),
			}
		}
	}
	if len(in.VPCSecurityGroupIDs) == 0 && len(db.VpcSecurityGroups) != 0 {
		in.VPCSecurityGroupIDs = make([]string, len(db.VpcSecurityGroups))
		for i, val := range db.VpcSecurityGroups {
			in.VPCSecurityGroupIDs[i] = aws.StringValue(val.VpcSecurityGroupId)
		}
	}
	in.EngineVersion = awsclients.LateInitializeStringPtr(in.EngineVersion, db.EngineVersion)
	// When version 5.6 is chosen, AWS creates 5.6.41 and that's totally valid.
	// But we detect as if we need to update it all the time. Here, we assign
	// the actual full version to our spec to avoid unnecessary update signals.
	if strings.HasPrefix(aws.StringValue(db.EngineVersion), aws.StringValue(in.EngineVersion)) {
		in.EngineVersion = db.EngineVersion
	}
}

// IsUpToDate checks whether there is a change in any of the modifiable fields.
func IsUpToDate(p v1beta1.RDSInstanceParameters, db rds.DBInstance) (bool, error) {
	// TODO(muvaf): ApplyImmediately and other configurations that exist in
	//  <Modify/Create/Delete>DBInstanceInput objects but not in DBInstance
	//  object are not late-inited. So, this func always returns true when
	//  those configurations are changed by the user.
	patch, err := CreatePatch(&db, &p)
	if err != nil {
		return false, err
	}
	return cmp.Equal(&v1beta1.RDSInstanceParameters{}, patch, cmpopts.EquateEmpty(),
		cmpopts.IgnoreTypes(&v1alpha1.Reference{}, &v1alpha1.Selector{}),
		cmpopts.IgnoreFields(v1beta1.RDSInstanceParameters{}, "Tags"),
		cmpopts.IgnoreFields(v1beta1.RDSInstanceParameters{}, "SkipFinalSnapshotBeforeDeletion")), nil
}

// GetConnectionDetails extracts managed.ConnectionDetails out of v1alpha3.RDSInstance.
func GetConnectionDetails(in v1beta1.RDSInstance) managed.ConnectionDetails {
	if in.Status.AtProvider.Endpoint.Address == "" {
		return nil
	}
	return managed.ConnectionDetails{
		v1alpha1.ResourceCredentialsSecretEndpointKey: []byte(in.Status.AtProvider.Endpoint.Address),
		v1alpha1.ResourceCredentialsSecretPortKey:     []byte(strconv.Itoa(in.Status.AtProvider.Endpoint.Port)),
	}
}
