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
	rdstypes "github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane/provider-aws/apis/database/v1beta1"
	awsclients "github.com/crossplane/provider-aws/pkg/clients"
)

const (
	errGetPasswordSecretFailed = "cannot get password secret"
)

// Client defines RDS RDSClient operations
type Client interface {
	CreateDBInstance(context.Context, *rds.CreateDBInstanceInput, ...func(*rds.Options)) (*rds.CreateDBInstanceOutput, error)
	DescribeDBInstances(context.Context, *rds.DescribeDBInstancesInput, ...func(*rds.Options)) (*rds.DescribeDBInstancesOutput, error)
	ModifyDBInstance(context.Context, *rds.ModifyDBInstanceInput, ...func(*rds.Options)) (*rds.ModifyDBInstanceOutput, error)
	DeleteDBInstance(context.Context, *rds.DeleteDBInstanceInput, ...func(*rds.Options)) (*rds.DeleteDBInstanceOutput, error)
	AddTagsToResource(context.Context, *rds.AddTagsToResourceInput, ...func(*rds.Options)) (*rds.AddTagsToResourceOutput, error)
}

// NewClient creates new RDS RDSClient with provided AWS Configurations/Credentials
func NewClient(cfg *aws.Config) Client {
	return rds.NewFromConfig(*cfg)
}

// IsErrorAlreadyExists returns true if the supplied error indicates an instance
// already exists.
func IsErrorAlreadyExists(err error) bool {
	var aef *rdstypes.DBInstanceAlreadyExistsFault
	return errors.As(err, &aef)
}

// IsErrorNotFound helper function to test for ErrCodeDBInstanceNotFoundFault error
func IsErrorNotFound(err error) bool {
	var nff *rdstypes.DBInstanceNotFoundFault
	return errors.As(err, &nff)
}

// GenerateCreateDBInstanceInput from RDSInstanceSpec
func GenerateCreateDBInstanceInput(name, password string, p *v1beta1.RDSInstanceParameters) *rds.CreateDBInstanceInput {
	c := &rds.CreateDBInstanceInput{
		DBInstanceIdentifier:               aws.String(name),
		AllocatedStorage:                   awsclients.Int32Address(p.AllocatedStorage),
		AutoMinorVersionUpgrade:            p.AutoMinorVersionUpgrade,
		AvailabilityZone:                   p.AvailabilityZone,
		BackupRetentionPeriod:              awsclients.Int32Address(p.BackupRetentionPeriod),
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
		Iops:                               awsclients.Int32Address(p.IOPS),
		KmsKeyId:                           p.KMSKeyID,
		LicenseModel:                       p.LicenseModel,
		MasterUserPassword:                 awsclients.String(password),
		MasterUsername:                     p.MasterUsername,
		MonitoringInterval:                 awsclients.Int32Address(p.MonitoringInterval),
		MonitoringRoleArn:                  p.MonitoringRoleARN,
		MultiAZ:                            p.MultiAZ,
		OptionGroupName:                    p.OptionGroupName,
		PerformanceInsightsKMSKeyId:        p.PerformanceInsightsKMSKeyID,
		PerformanceInsightsRetentionPeriod: awsclients.Int32Address(p.PerformanceInsightsRetentionPeriod),
		Port:                               awsclients.Int32Address(p.Port),
		PreferredBackupWindow:              p.PreferredBackupWindow,
		PreferredMaintenanceWindow:         p.PreferredMaintenanceWindow,
		PromotionTier:                      awsclients.Int32Address(p.PromotionTier),
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

// CreatePatch creates a *v1beta1.RDSInstanceParameters that has only the changed
// values between the target *v1beta1.RDSInstanceParameters and the current
// *rds.DBInstance
func CreatePatch(in *rdstypes.DBInstance, target *v1beta1.RDSInstanceParameters) (*v1beta1.RDSInstanceParameters, error) {
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
		AllocatedStorage:                   awsclients.Int32Address(p.AllocatedStorage),
		AllowMajorVersionUpgrade:           aws.ToBool(p.AllowMajorVersionUpgrade),
		ApplyImmediately:                   aws.ToBool(p.ApplyModificationsImmediately),
		AutoMinorVersionUpgrade:            p.AutoMinorVersionUpgrade,
		BackupRetentionPeriod:              awsclients.Int32Address(p.BackupRetentionPeriod),
		CACertificateIdentifier:            p.CACertificateIdentifier,
		CopyTagsToSnapshot:                 p.CopyTagsToSnapshot,
		DBInstanceClass:                    awsclients.String(p.DBInstanceClass),
		DBParameterGroupName:               p.DBParameterGroupName,
		DBPortNumber:                       awsclients.Int32Address(p.Port),
		DBSecurityGroups:                   p.DBSecurityGroups,
		DBSubnetGroupName:                  p.DBSubnetGroupName,
		DeletionProtection:                 p.DeletionProtection,
		Domain:                             p.Domain,
		DomainIAMRoleName:                  p.DomainIAMRoleName,
		EnableIAMDatabaseAuthentication:    p.EnableIAMDatabaseAuthentication,
		EnablePerformanceInsights:          p.EnablePerformanceInsights,
		EngineVersion:                      p.EngineVersion,
		Iops:                               awsclients.Int32Address(p.IOPS),
		LicenseModel:                       p.LicenseModel,
		MonitoringInterval:                 awsclients.Int32Address(p.MonitoringInterval),
		MonitoringRoleArn:                  p.MonitoringRoleARN,
		MultiAZ:                            p.MultiAZ,
		OptionGroupName:                    p.OptionGroupName,
		PerformanceInsightsKMSKeyId:        p.PerformanceInsightsKMSKeyID,
		PerformanceInsightsRetentionPeriod: awsclients.Int32Address(p.PerformanceInsightsRetentionPeriod),
		PreferredBackupWindow:              p.PreferredBackupWindow,
		PreferredMaintenanceWindow:         p.PreferredMaintenanceWindow,
		PromotionTier:                      awsclients.Int32Address(p.PromotionTier),
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
	if p.CloudwatchLogsExportConfiguration != nil {
		m.CloudwatchLogsExportConfiguration = &rdstypes.CloudwatchLogsExportConfiguration{
			DisableLogTypes: p.CloudwatchLogsExportConfiguration.DisableLogTypes,
			EnableLogTypes:  p.CloudwatchLogsExportConfiguration.EnableLogTypes,
		}
	}
	return m
}

// GenerateObservation is used to produce v1alpha3.RDSInstanceObservation from
// rds.DBInstance.
func GenerateObservation(db rdstypes.DBInstance) v1beta1.RDSInstanceObservation { // nolint:gocyclo
	o := v1beta1.RDSInstanceObservation{
		DBInstanceStatus:                      aws.ToString(db.DBInstanceStatus),
		DBInstanceArn:                         aws.ToString(db.DBInstanceArn),
		DBInstancePort:                        int(db.DbInstancePort),
		DBResourceID:                          aws.ToString(db.DbiResourceId),
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
func LateInitialize(in *v1beta1.RDSInstanceParameters, db *rdstypes.DBInstance) { // nolint:gocyclo
	if db == nil {
		return
	}
	in.DBInstanceClass = awsclients.LateInitializeString(in.DBInstanceClass, db.DBInstanceClass)
	in.Engine = awsclients.LateInitializeString(in.Engine, db.Engine)

	in.AllocatedStorage = awsclients.LateInitializeIntFrom32Ptr(in.AllocatedStorage, &db.AllocatedStorage)
	in.AutoMinorVersionUpgrade = awsclients.LateInitializeBoolPtr(in.AutoMinorVersionUpgrade, awsclients.Bool(db.AutoMinorVersionUpgrade))
	in.AvailabilityZone = awsclients.LateInitializeStringPtr(in.AvailabilityZone, db.AvailabilityZone)
	in.BackupRetentionPeriod = awsclients.LateInitializeIntFrom32Ptr(in.BackupRetentionPeriod, &db.BackupRetentionPeriod)
	in.CACertificateIdentifier = awsclients.LateInitializeStringPtr(in.CACertificateIdentifier, db.CACertificateIdentifier)
	in.CharacterSetName = awsclients.LateInitializeStringPtr(in.CharacterSetName, db.CharacterSetName)
	in.CopyTagsToSnapshot = awsclients.LateInitializeBoolPtr(in.CopyTagsToSnapshot, awsclients.Bool(db.CopyTagsToSnapshot))
	in.DBClusterIdentifier = awsclients.LateInitializeStringPtr(in.DBClusterIdentifier, db.DBClusterIdentifier)
	in.DBName = awsclients.LateInitializeStringPtr(in.DBName, db.DBName)
	in.DeletionProtection = awsclients.LateInitializeBoolPtr(in.DeletionProtection, awsclients.Bool(db.DeletionProtection))
	in.EnableIAMDatabaseAuthentication = awsclients.LateInitializeBoolPtr(in.EnableIAMDatabaseAuthentication, awsclients.Bool(db.IAMDatabaseAuthenticationEnabled))
	in.EnablePerformanceInsights = awsclients.LateInitializeBoolPtr(in.EnablePerformanceInsights, db.PerformanceInsightsEnabled)
	in.IOPS = awsclients.LateInitializeIntFrom32Ptr(in.IOPS, db.Iops)
	in.KMSKeyID = awsclients.LateInitializeStringPtr(in.KMSKeyID, db.KmsKeyId)
	in.LicenseModel = awsclients.LateInitializeStringPtr(in.LicenseModel, db.LicenseModel)
	in.MasterUsername = awsclients.LateInitializeStringPtr(in.MasterUsername, db.MasterUsername)
	in.MonitoringInterval = awsclients.LateInitializeIntFrom32Ptr(in.MonitoringInterval, db.MonitoringInterval)
	in.MonitoringRoleARN = awsclients.LateInitializeStringPtr(in.MonitoringRoleARN, db.MonitoringRoleArn)
	in.MultiAZ = awsclients.LateInitializeBoolPtr(in.MultiAZ, awsclients.Bool(db.MultiAZ))
	in.PerformanceInsightsKMSKeyID = awsclients.LateInitializeStringPtr(in.PerformanceInsightsKMSKeyID, db.PerformanceInsightsKMSKeyId)
	in.PerformanceInsightsRetentionPeriod = awsclients.LateInitializeIntFrom32Ptr(in.PerformanceInsightsRetentionPeriod, db.PerformanceInsightsRetentionPeriod)
	in.PreferredBackupWindow = awsclients.LateInitializeStringPtr(in.PreferredBackupWindow, db.PreferredBackupWindow)
	in.PreferredMaintenanceWindow = awsclients.LateInitializeStringPtr(in.PreferredMaintenanceWindow, db.PreferredMaintenanceWindow)
	in.PromotionTier = awsclients.LateInitializeIntFrom32Ptr(in.PromotionTier, db.PromotionTier)
	in.PubliclyAccessible = awsclients.LateInitializeBoolPtr(in.PubliclyAccessible, awsclients.Bool(db.PubliclyAccessible))
	in.StorageEncrypted = awsclients.LateInitializeBoolPtr(in.StorageEncrypted, awsclients.Bool(db.StorageEncrypted))
	in.StorageType = awsclients.LateInitializeStringPtr(in.StorageType, db.StorageType)
	in.Timezone = awsclients.LateInitializeStringPtr(in.Timezone, db.Timezone)

	// NOTE(muvaf): Do not use db.DbInstancePort as that always returns 0 for
	// some reason. See the bug here:
	// https://github.com/aws/aws-sdk-java/issues/924#issuecomment-658089792
	if db.Endpoint != nil {
		in.Port = awsclients.LateInitializeIntFrom32Ptr(in.Port, &db.Endpoint.Port)
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
	if len(in.EnableCloudwatchLogsExports) == 0 && len(db.EnabledCloudwatchLogsExports) != 0 {
		in.EnableCloudwatchLogsExports = db.EnabledCloudwatchLogsExports
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
	in.EngineVersion = awsclients.LateInitializeStringPtr(in.EngineVersion, db.EngineVersion)
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
}

// IsUpToDate checks whether there is a change in any of the modifiable fields.
func IsUpToDate(ctx context.Context, kube client.Client, r *v1beta1.RDSInstance, db rdstypes.DBInstance) (bool, error) {
	_, pwdChanged, err := GetPassword(ctx, kube, r.Spec.ForProvider.MasterPasswordSecretRef, r.Spec.WriteConnectionSecretToReference)
	if err != nil {
		return false, err
	}
	patch, err := CreatePatch(&db, &r.Spec.ForProvider)
	if err != nil {
		return false, err
	}
	return cmp.Equal(&v1beta1.RDSInstanceParameters{}, patch, cmpopts.EquateEmpty(),
		cmpopts.IgnoreTypes(&xpv1.Reference{}, &xpv1.Selector{}, []xpv1.Reference{}),
		cmpopts.IgnoreFields(v1beta1.RDSInstanceParameters{}, "Region"),
		cmpopts.IgnoreFields(v1beta1.RDSInstanceParameters{}, "Tags"),
		cmpopts.IgnoreFields(v1beta1.RDSInstanceParameters{}, "SkipFinalSnapshotBeforeDeletion"),
		cmpopts.IgnoreFields(v1beta1.RDSInstanceParameters{}, "FinalDBSnapshotIdentifier"),
		cmpopts.IgnoreFields(v1beta1.RDSInstanceParameters{}, "ApplyModificationsImmediately"),
		cmpopts.IgnoreFields(v1beta1.RDSInstanceParameters{}, "AllowMajorVersionUpgrade"),
		cmpopts.IgnoreFields(v1beta1.RDSInstanceParameters{}, "MasterPasswordSecretRef"),
	) && !pwdChanged, nil
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
