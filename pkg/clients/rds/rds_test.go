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
	"strconv"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/google/go-cmp/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"

	"github.com/crossplane/provider-aws/apis/database/v1beta1"
	aws "github.com/crossplane/provider-aws/pkg/clients"
)

var (
	dbName           = "example-name"
	characterSetName = "utf8"
)

func TestCreatePatch(t *testing.T) {
	type args struct {
		db *rds.DBInstance
		p  *v1beta1.RDSInstanceParameters
	}

	type want struct {
		patch *v1beta1.RDSInstanceParameters
	}

	cases := map[string]struct {
		args
		want
	}{
		"SameFields": {
			args: args{
				db: &rds.DBInstance{
					AllocatedStorage: aws.Int64(20),
					CharacterSetName: &characterSetName,
					DBName:           &dbName,
				},
				p: &v1beta1.RDSInstanceParameters{
					AllocatedStorage: aws.IntAddress(aws.Int64(20)),
					CharacterSetName: &characterSetName,
					DBName:           &dbName,
				},
			},
			want: want{
				patch: &v1beta1.RDSInstanceParameters{},
			},
		},
		"DifferentFields": {
			args: args{
				db: &rds.DBInstance{
					AllocatedStorage: aws.Int64(20),
					CharacterSetName: &characterSetName,
					DBName:           &dbName,
				},
				p: &v1beta1.RDSInstanceParameters{
					AllocatedStorage: aws.IntAddress(aws.Int64(30)),
					CharacterSetName: &characterSetName,
					DBName:           &dbName,
				},
			},
			want: want{
				patch: &v1beta1.RDSInstanceParameters{
					AllocatedStorage: aws.IntAddress(aws.Int64(30)),
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			result, _ := CreatePatch(tc.args.db, tc.args.p)
			if diff := cmp.Diff(tc.want.patch, result); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestIsUpToDate(t *testing.T) {
	dbSubnetGroupName := "example-subnet"

	type args struct {
		db rds.DBInstance
		p  v1beta1.RDSInstanceParameters
	}

	cases := map[string]struct {
		args args
		want bool
	}{
		"SameFields": {
			args: args{
				db: rds.DBInstance{
					AllocatedStorage: aws.Int64(20),
					CharacterSetName: &characterSetName,
					DBName:           &dbName,
				},
				p: v1beta1.RDSInstanceParameters{
					AllocatedStorage: aws.IntAddress(aws.Int64(20)),
					CharacterSetName: &characterSetName,
					DBName:           &dbName,
				},
			},
			want: true,
		},
		"DifferentFields": {
			args: args{
				db: rds.DBInstance{
					AllocatedStorage: aws.Int64(20),
					CharacterSetName: &characterSetName,
					DBName:           &dbName,
				},
				p: v1beta1.RDSInstanceParameters{
					AllocatedStorage: aws.IntAddress(aws.Int64(30)),
					CharacterSetName: &characterSetName,
					DBName:           &dbName,
				},
			},
			want: false,
		},
		"IgnoresRefs": {
			args: args{
				db: rds.DBInstance{
					DBName:        &dbName,
					DBSubnetGroup: &rds.DBSubnetGroup{DBSubnetGroupName: &dbSubnetGroupName},
				},
				p: v1beta1.RDSInstanceParameters{
					DBName:               &dbName,
					DBSubnetGroupName:    &dbSubnetGroupName,
					DBSubnetGroupNameRef: &v1alpha1.Reference{Name: "coolgroup"},
				},
			},
			want: true,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got, _ := IsUpToDate(tc.args.p, tc.args.db)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGenerateObservation(t *testing.T) {
	lastRestoreTime, createTime := time.Now(), time.Now()
	instanceStatus := "status"
	name := "testName"
	value := "testValue"
	description := "testDescription"
	status := "testStatus"
	domain := "testDomain"
	arn := "my:arn"
	address := "testAddress"
	zone := "zone"
	az := rds.AvailabilityZone{Name: &name}
	vpc := "vpc"
	port := int64(123)
	resourceID := "resource"
	insightsEnabled := true
	replicaSourceIdentifier := "replicaSource"
	secondaryAZ := "secondary"
	storage := int64(1)
	retention := int64(2)
	instanceClass := "class"
	multiAZ := false
	normal := true
	replicaClusters := []string{"replicaCluster"}
	subnetGroup := rds.DBSubnetGroup{
		DBSubnetGroupArn:         &arn,
		DBSubnetGroupDescription: &description,
		DBSubnetGroupName:        &name,
		SubnetGroupStatus:        &status,
		VpcId:                    &vpc,
	}
	subnetGroup.Subnets = []rds.Subnet{{
		SubnetIdentifier:       &name,
		SubnetStatus:           &status,
		SubnetAvailabilityZone: &az,
	}}
	endpoint := rds.Endpoint{
		Address:      &address,
		HostedZoneId: &zone,
		Port:         &port,
	}
	pendingModifiedValues := rds.PendingModifiedValues{
		AllocatedStorage:        &storage,
		BackupRetentionPeriod:   &retention,
		CACertificateIdentifier: &name,
		DBInstanceClass:         &instanceClass,
		DBSubnetGroupName:       &name,
		Iops:                    &storage,
		LicenseModel:            &name,
		MultiAZ:                 &multiAZ,
		Port:                    &port,
		StorageType:             &name,
	}
	pendingCloudwatch := rds.PendingCloudwatchLogsExports{
		LogTypesToDisable: []string{"test"},
		LogTypesToEnable:  []string{"test"},
	}
	pendingModifiedValues.PendingCloudwatchLogsExports = &pendingCloudwatch
	pendingModifiedValues.ProcessorFeatures = []rds.ProcessorFeature{{
		Name:  &name,
		Value: &value,
	}}

	cases := map[string]struct {
		rds  rds.DBInstance
		want v1beta1.RDSInstanceObservation
	}{
		"AllFields": {
			rds: rds.DBInstance{
				DBInstanceStatus:                      &instanceStatus,
				DBInstanceArn:                         &arn,
				InstanceCreateTime:                    &createTime,
				DbInstancePort:                        &port,
				DbiResourceId:                         &resourceID,
				EnhancedMonitoringResourceArn:         &arn,
				PerformanceInsightsEnabled:            &insightsEnabled,
				ReadReplicaDBClusterIdentifiers:       replicaClusters,
				ReadReplicaDBInstanceIdentifiers:      replicaClusters,
				ReadReplicaSourceDBInstanceIdentifier: &replicaSourceIdentifier,
				SecondaryAvailabilityZone:             &secondaryAZ,
				LatestRestorableTime:                  &lastRestoreTime,
				DBParameterGroups:                     []rds.DBParameterGroupStatus{{DBParameterGroupName: &name}},
				DBSecurityGroups:                      []rds.DBSecurityGroupMembership{{DBSecurityGroupName: &name, Status: &status}},
				DBSubnetGroup:                         &subnetGroup,
				DomainMemberships: []rds.DomainMembership{{
					Domain:      &domain,
					FQDN:        &name,
					IAMRoleName: &name,
					Status:      &status,
				}},
				Endpoint: &endpoint,
				OptionGroupMemberships: []rds.OptionGroupMembership{{
					OptionGroupName: &name,
					Status:          &status,
				}},
				PendingModifiedValues: &pendingModifiedValues,
				StatusInfos: []rds.DBInstanceStatusInfo{{
					Message:    &status,
					Status:     &status,
					StatusType: &status,
					Normal:     &normal,
				}},
				VpcSecurityGroups: []rds.VpcSecurityGroupMembership{{
					Status:             &status,
					VpcSecurityGroupId: &name,
				}},
			},
			want: v1beta1.RDSInstanceObservation{
				DBInstanceStatus:  instanceStatus,
				DBInstanceArn:     arn,
				DBParameterGroups: []v1beta1.DBParameterGroupStatus{{DBParameterGroupName: name}},
				DBSecurityGroups:  []v1beta1.DBSecurityGroupMembership{{DBSecurityGroupName: name, Status: status}},
				DBSubnetGroup: v1beta1.DBSubnetGroupInRDS{
					DBSubnetGroupARN:         arn,
					DBSubnetGroupDescription: description,
					DBSubnetGroupName:        name,
					SubnetGroupStatus:        status,
					VPCID:                    vpc,
					Subnets: []v1beta1.SubnetInRDS{{
						SubnetIdentifier:       name,
						SubnetStatus:           status,
						SubnetAvailabilityZone: v1beta1.AvailabilityZone{Name: *az.Name},
					}},
				},
				DBInstancePort:                int(port),
				DBResourceID:                  resourceID,
				DomainMemberships:             []v1beta1.DomainMembership{{Domain: domain, FQDN: name, IAMRoleName: name, Status: status}},
				InstanceCreateTime:            &metav1.Time{Time: createTime},
				Endpoint:                      v1beta1.Endpoint{Port: int(port), HostedZoneID: zone, Address: address},
				EnhancedMonitoringResourceArn: arn,
				LatestRestorableTime:          &metav1.Time{Time: lastRestoreTime},
				OptionGroupMemberships:        []v1beta1.OptionGroupMembership{{OptionGroupName: name, Status: status}},
				PendingModifiedValues: v1beta1.PendingModifiedValues{
					AllocatedStorage:        int(storage),
					BackupRetentionPeriod:   int(retention),
					CACertificateIdentifier: name,
					DBInstanceClass:         instanceClass,
					DBSubnetGroupName:       name,
					IOPS:                    int(storage),
					LicenseModel:            name,
					MultiAZ:                 multiAZ,
					Port:                    int(port),
					StorageType:             name,
					PendingCloudwatchLogsExports: v1beta1.PendingCloudwatchLogsExports{
						LogTypesToDisable: []string{"test"},
						LogTypesToEnable:  []string{"test"},
					},
					ProcessorFeatures: []v1beta1.ProcessorFeature{{Name: name, Value: value}},
				},
				PerformanceInsightsEnabled:            insightsEnabled,
				ReadReplicaDBClusterIdentifiers:       replicaClusters,
				ReadReplicaDBInstanceIdentifiers:      replicaClusters,
				ReadReplicaSourceDBInstanceIdentifier: replicaSourceIdentifier,
				SecondaryAvailabilityZone:             secondaryAZ,
				StatusInfos: []v1beta1.DBInstanceStatusInfo{{
					Message:    status,
					Status:     status,
					StatusType: status,
					Normal:     normal,
				}},
				VPCSecurityGroups: []v1beta1.VPCSecurityGroupMembership{{
					Status:             status,
					VPCSecurityGroupID: name,
				}},
			},
		},
		"SomeFields": {
			rds: rds.DBInstance{
				DBInstanceStatus:                      &instanceStatus,
				DBInstanceArn:                         &arn,
				InstanceCreateTime:                    &createTime,
				DbInstancePort:                        &port,
				DbiResourceId:                         &resourceID,
				EnhancedMonitoringResourceArn:         &arn,
				PerformanceInsightsEnabled:            &insightsEnabled,
				ReadReplicaDBClusterIdentifiers:       replicaClusters,
				ReadReplicaDBInstanceIdentifiers:      replicaClusters,
				ReadReplicaSourceDBInstanceIdentifier: &replicaSourceIdentifier,
				SecondaryAvailabilityZone:             &secondaryAZ,
				LatestRestorableTime:                  &lastRestoreTime,
				DomainMemberships: []rds.DomainMembership{{
					Domain:      &domain,
					FQDN:        &name,
					IAMRoleName: &name,
					Status:      &status,
				}},
				Endpoint: &endpoint,
				OptionGroupMemberships: []rds.OptionGroupMembership{{
					OptionGroupName: &name,
					Status:          &status,
				}},
				StatusInfos: []rds.DBInstanceStatusInfo{{
					Message:    &status,
					Status:     &status,
					StatusType: &status,
					Normal:     &normal,
				}},
				VpcSecurityGroups: []rds.VpcSecurityGroupMembership{{
					Status:             &status,
					VpcSecurityGroupId: &name,
				}},
			},
			want: v1beta1.RDSInstanceObservation{
				DBInstanceStatus:                      instanceStatus,
				DBInstanceArn:                         arn,
				DBInstancePort:                        int(port),
				DBResourceID:                          resourceID,
				DomainMemberships:                     []v1beta1.DomainMembership{{Domain: domain, FQDN: name, IAMRoleName: name, Status: status}},
				InstanceCreateTime:                    &metav1.Time{Time: createTime},
				Endpoint:                              v1beta1.Endpoint{Port: int(port), HostedZoneID: zone, Address: address},
				EnhancedMonitoringResourceArn:         arn,
				LatestRestorableTime:                  &metav1.Time{Time: lastRestoreTime},
				OptionGroupMemberships:                []v1beta1.OptionGroupMembership{{OptionGroupName: name, Status: status}},
				PerformanceInsightsEnabled:            insightsEnabled,
				ReadReplicaDBClusterIdentifiers:       replicaClusters,
				ReadReplicaDBInstanceIdentifiers:      replicaClusters,
				ReadReplicaSourceDBInstanceIdentifier: replicaSourceIdentifier,
				SecondaryAvailabilityZone:             secondaryAZ,
				StatusInfos: []v1beta1.DBInstanceStatusInfo{{
					Message:    status,
					Status:     status,
					StatusType: status,
					Normal:     normal,
				}},
				VPCSecurityGroups: []v1beta1.VPCSecurityGroupMembership{{
					Status:             status,
					VPCSecurityGroupID: name,
				}},
			},
		},
		"EmptyInstance": {
			rds:  rds.DBInstance{},
			want: v1beta1.RDSInstanceObservation{},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := GenerateObservation(tc.rds)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGetConnectionDetails(t *testing.T) {
	address := "testAddress"
	port := 123
	cases := map[string]struct {
		rds  v1beta1.RDSInstance
		want managed.ConnectionDetails
	}{
		"ValidInstance": {
			rds: v1beta1.RDSInstance{
				Status: v1beta1.RDSInstanceStatus{
					AtProvider: v1beta1.RDSInstanceObservation{
						Endpoint: v1beta1.Endpoint{
							Address: address,
							Port:    port,
						},
					},
				},
			},
			want: managed.ConnectionDetails{
				v1alpha1.ResourceCredentialsSecretEndpointKey: []byte(address),
				v1alpha1.ResourceCredentialsSecretPortKey:     []byte(strconv.Itoa(port)),
			},
		},
		"NilInstance": {
			rds:  v1beta1.RDSInstance{},
			want: nil,
		}}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := GetConnectionDetails(tc.rds)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestLateInitialize(t *testing.T) {
	name := "testName"
	clusterName := "testCluster"
	value := "testValue"
	description := "testDescription"
	status := "testStatus"
	arn := "my:arn"
	zone := "zone"
	az := "az"
	vpc := "vpc"
	port := 123
	port64 := int64(port)
	storage := 1
	storage64 := int64(storage)
	retention := 2
	retention64 := int64(retention)
	monitoring := 3
	monitoring64 := int64(monitoring)
	tier := 4
	tier64 := int64(tier)
	storageType := "storageType"
	instanceClass := "class"
	engine := "5.6.41"
	truncEngine := "5.6"
	multiAZ := false
	trueFlag := true
	kmsID := "kms"
	username := "username"
	window := "window"
	existingName := "existing"
	subnetGroup := rds.DBSubnetGroup{
		DBSubnetGroupArn:         &arn,
		DBSubnetGroupDescription: &description,
		DBSubnetGroupName:        &name,
		SubnetGroupStatus:        &status,
		VpcId:                    &vpc,
	}
	cloudwatchExports := []string{"test"}

	cases := map[string]struct {
		rds    rds.DBInstance
		params v1beta1.RDSInstanceParameters
		want   v1beta1.RDSInstanceParameters
	}{
		"AllFields": {
			rds: rds.DBInstance{
				AllocatedStorage:                   &storage64,
				DBInstanceClass:                    &instanceClass,
				Engine:                             &engine,
				AutoMinorVersionUpgrade:            &trueFlag,
				AvailabilityZone:                   &az,
				BackupRetentionPeriod:              &storage64,
				CACertificateIdentifier:            &name,
				CharacterSetName:                   &name,
				CopyTagsToSnapshot:                 &trueFlag,
				DBClusterIdentifier:                &clusterName,
				DBName:                             &name,
				DeletionProtection:                 &trueFlag,
				IAMDatabaseAuthenticationEnabled:   &trueFlag,
				PerformanceInsightsEnabled:         &trueFlag,
				Iops:                               &storage64,
				KmsKeyId:                           &kmsID,
				LicenseModel:                       &name,
				MasterUsername:                     &username,
				MonitoringInterval:                 &monitoring64,
				MonitoringRoleArn:                  &arn,
				MultiAZ:                            &multiAZ,
				PerformanceInsightsKMSKeyId:        &kmsID,
				PerformanceInsightsRetentionPeriod: &retention64,
				DbInstancePort:                     &port64,
				PreferredBackupWindow:              &window,
				PreferredMaintenanceWindow:         &window,
				PromotionTier:                      &tier64,
				PubliclyAccessible:                 &trueFlag,
				StorageEncrypted:                   &trueFlag,
				StorageType:                        &storageType,
				Timezone:                           &zone,
				DBSecurityGroups:                   []rds.DBSecurityGroupMembership{{DBSecurityGroupName: &name, Status: &status}},
				DBSubnetGroup:                      &subnetGroup,
				EnabledCloudwatchLogsExports:       cloudwatchExports,
				ProcessorFeatures: []rds.ProcessorFeature{{
					Name:  &name,
					Value: &value,
				}},
				VpcSecurityGroups: []rds.VpcSecurityGroupMembership{{
					Status:             &status,
					VpcSecurityGroupId: &name,
				}},
				EngineVersion: &engine,
			},
			params: v1beta1.RDSInstanceParameters{},
			want: v1beta1.RDSInstanceParameters{
				AllocatedStorage:                   &storage,
				DBInstanceClass:                    instanceClass,
				Engine:                             engine,
				AutoMinorVersionUpgrade:            &trueFlag,
				AvailabilityZone:                   &az,
				BackupRetentionPeriod:              &storage,
				CACertificateIdentifier:            &name,
				CharacterSetName:                   &name,
				CopyTagsToSnapshot:                 &trueFlag,
				DBClusterIdentifier:                &clusterName,
				DBName:                             &name,
				DeletionProtection:                 &trueFlag,
				EnableIAMDatabaseAuthentication:    &trueFlag,
				EnablePerformanceInsights:          &trueFlag,
				IOPS:                               &storage,
				KMSKeyID:                           &kmsID,
				LicenseModel:                       &name,
				MasterUsername:                     &username,
				MonitoringInterval:                 &monitoring,
				MonitoringRoleARN:                  &arn,
				MultiAZ:                            &multiAZ,
				PerformanceInsightsKMSKeyID:        &kmsID,
				PerformanceInsightsRetentionPeriod: &retention,
				Port:                               &port,
				PreferredBackupWindow:              &window,
				PreferredMaintenanceWindow:         &window,
				PromotionTier:                      &tier,
				PubliclyAccessible:                 &trueFlag,
				StorageEncrypted:                   &trueFlag,
				StorageType:                        &storageType,
				Timezone:                           &zone,
				DBSecurityGroups:                   []string{name},
				DBSubnetGroupName:                  subnetGroup.DBSubnetGroupName,
				EnableCloudwatchLogsExports:        cloudwatchExports,
				ProcessorFeatures: []v1beta1.ProcessorFeature{{
					Name:  name,
					Value: value,
				}},
				VPCSecurityGroupIDs: []string{name},
				EngineVersion:       &engine,
			},
		},
		"SubnetGroupNameSet": {
			rds: rds.DBInstance{
				DBSubnetGroup: &subnetGroup,
			},
			params: v1beta1.RDSInstanceParameters{},
			want: v1beta1.RDSInstanceParameters{
				DBSubnetGroupName: subnetGroup.DBSubnetGroupName,
			},
		},
		"SubnetGroupNameNotOverwritten": {
			rds: rds.DBInstance{
				DBSubnetGroup: &subnetGroup,
			},
			params: v1beta1.RDSInstanceParameters{
				DBSubnetGroupName: &existingName,
			},
			want: v1beta1.RDSInstanceParameters{
				DBSubnetGroupName: &existingName,
			},
		},
		"SecurityGroupNotOverwritten": {
			rds: rds.DBInstance{
				DBSecurityGroups: []rds.DBSecurityGroupMembership{{DBSecurityGroupName: &name, Status: &status}},
			},
			params: v1beta1.RDSInstanceParameters{
				DBSecurityGroups: []string{"newGroup"},
			},
			want: v1beta1.RDSInstanceParameters{
				DBSecurityGroups: []string{"newGroup"},
			},
		},
		"CloudwatchExportsNotOverwritten": {
			rds: rds.DBInstance{
				EnabledCloudwatchLogsExports: cloudwatchExports,
			},
			params: v1beta1.RDSInstanceParameters{
				EnableCloudwatchLogsExports: []string{"newExport"},
			},
			want: v1beta1.RDSInstanceParameters{
				EnableCloudwatchLogsExports: []string{"newExport"},
			},
		},
		"ProcessorFeaturesNotOverwritten": {
			rds: rds.DBInstance{
				ProcessorFeatures: []rds.ProcessorFeature{{
					Name:  &name,
					Value: &value,
				}},
			},
			params: v1beta1.RDSInstanceParameters{
				ProcessorFeatures: []v1beta1.ProcessorFeature{{
					Name:  existingName,
					Value: existingName,
				}},
			},
			want: v1beta1.RDSInstanceParameters{
				ProcessorFeatures: []v1beta1.ProcessorFeature{{
					Name:  existingName,
					Value: existingName,
				}},
			},
		},
		"VPCSecurityGroupIdsNotOverwritten": {
			rds: rds.DBInstance{
				ProcessorFeatures: []rds.ProcessorFeature{{
					Name:  &name,
					Value: &value,
				}},
			},
			params: v1beta1.RDSInstanceParameters{
				ProcessorFeatures: []v1beta1.ProcessorFeature{{
					Name:  existingName,
					Value: existingName,
				}},
			},
			want: v1beta1.RDSInstanceParameters{
				ProcessorFeatures: []v1beta1.ProcessorFeature{{
					Name:  existingName,
					Value: existingName,
				}},
			},
		},
		"EngineVersion": {
			rds: rds.DBInstance{
				EngineVersion: &engine,
			},
			params: v1beta1.RDSInstanceParameters{
				EngineVersion: &truncEngine,
			},
			want: v1beta1.RDSInstanceParameters{
				EngineVersion: &engine,
			},
		},
		"EmptyInstance": {
			rds:    rds.DBInstance{},
			params: v1beta1.RDSInstanceParameters{},
			want:   v1beta1.RDSInstanceParameters{},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			LateInitialize(&tc.params, &tc.rds)
			if diff := cmp.Diff(tc.want, tc.params); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGenerateModifyDBInstanceInput(t *testing.T) {
	name := "testName"
	clusterName := "testCluster"
	value := "testValue"
	arn := "my:arn"
	zone := "zone"
	az := "az"
	port := 123
	port64 := int64(port)
	storage := 1
	storage64 := int64(storage)
	retention := 2
	retention64 := int64(retention)
	monitoring := 3
	monitoring64 := int64(monitoring)
	tier := 4
	tier64 := int64(tier)
	instanceClass := "class"
	engine := "5.6.41"
	multiAZ := false
	trueFlag := true
	kmsID := "kms"
	username := "username"
	window := "window"
	domain := "domain"
	storageType := "storageType"
	dbSecurityGroups := []string{name}
	cloudwatchExports := []string{"test"}
	allFieldsName := "allfieldsName"
	emptyName := "emptyProcessor"
	iamRole := "iamRole"
	vpcIds := []string{name}
	rdsCloudwatchLogsExportConfig := rds.CloudwatchLogsExportConfiguration{
		DisableLogTypes: []string{"test"},
		EnableLogTypes:  []string{"test"},
	}
	cloudwatchLogsExportConfig := v1beta1.CloudwatchLogsExportConfiguration{
		DisableLogTypes: []string{"test"},
		EnableLogTypes:  []string{"test"},
	}
	cases := map[string]struct {
		name   string
		params v1beta1.RDSInstanceParameters
		want   rds.ModifyDBInstanceInput
	}{
		"AllFields": {
			name: allFieldsName,
			params: v1beta1.RDSInstanceParameters{
				AllocatedStorage:                   &storage,
				DBInstanceClass:                    instanceClass,
				ApplyModificationsImmediately:      &trueFlag,
				Engine:                             engine,
				AutoMinorVersionUpgrade:            &trueFlag,
				AllowMajorVersionUpgrade:           &trueFlag,
				AvailabilityZone:                   &az,
				BackupRetentionPeriod:              &retention,
				CACertificateIdentifier:            &name,
				CharacterSetName:                   &name,
				CloudwatchLogsExportConfiguration:  &cloudwatchLogsExportConfig,
				CopyTagsToSnapshot:                 &trueFlag,
				DBClusterIdentifier:                &clusterName,
				DBName:                             &name,
				DBParameterGroupName:               &name,
				DeletionProtection:                 &trueFlag,
				Domain:                             &domain,
				DomainIAMRoleName:                  &iamRole,
				EnableIAMDatabaseAuthentication:    &trueFlag,
				EnablePerformanceInsights:          &trueFlag,
				IOPS:                               &storage,
				KMSKeyID:                           &kmsID,
				LicenseModel:                       &name,
				MasterUsername:                     &username,
				MonitoringInterval:                 &monitoring,
				MonitoringRoleARN:                  &arn,
				MultiAZ:                            &multiAZ,
				OptionGroupName:                    &name,
				PerformanceInsightsKMSKeyID:        &kmsID,
				PerformanceInsightsRetentionPeriod: &retention,
				Port:                               &port,
				PreferredBackupWindow:              &window,
				PreferredMaintenanceWindow:         &window,
				PromotionTier:                      &tier,
				PubliclyAccessible:                 &trueFlag,
				StorageEncrypted:                   &trueFlag,
				StorageType:                        &storageType,
				Timezone:                           &zone,
				DBSecurityGroups:                   dbSecurityGroups,
				DBSubnetGroupName:                  &name,
				EnableCloudwatchLogsExports:        cloudwatchExports,
				ProcessorFeatures: []v1beta1.ProcessorFeature{{
					Name:  name,
					Value: value,
				}},
				VPCSecurityGroupIDs:         vpcIds,
				EngineVersion:               &engine,
				UseDefaultProcessorFeatures: &trueFlag,
			},
			want: rds.ModifyDBInstanceInput{
				DBInstanceIdentifier:               &allFieldsName,
				AllocatedStorage:                   &storage64,
				AllowMajorVersionUpgrade:           &trueFlag,
				ApplyImmediately:                   &trueFlag,
				AutoMinorVersionUpgrade:            &trueFlag,
				BackupRetentionPeriod:              &retention64,
				CACertificateIdentifier:            &name,
				CopyTagsToSnapshot:                 &trueFlag,
				DBInstanceClass:                    &instanceClass,
				DBParameterGroupName:               &name,
				DBPortNumber:                       &port64,
				DBSecurityGroups:                   dbSecurityGroups,
				DBSubnetGroupName:                  &name,
				DeletionProtection:                 &trueFlag,
				Domain:                             &domain,
				DomainIAMRoleName:                  &iamRole,
				EnableIAMDatabaseAuthentication:    &trueFlag,
				EnablePerformanceInsights:          &trueFlag,
				EngineVersion:                      &engine,
				Iops:                               &storage64,
				LicenseModel:                       &name,
				MonitoringInterval:                 &monitoring64,
				MonitoringRoleArn:                  &arn,
				MultiAZ:                            &multiAZ,
				OptionGroupName:                    &name,
				PerformanceInsightsRetentionPeriod: &retention64,
				PerformanceInsightsKMSKeyId:        &kmsID,
				PreferredBackupWindow:              &window,
				PreferredMaintenanceWindow:         &window,
				PromotionTier:                      &tier64,
				PubliclyAccessible:                 &trueFlag,
				StorageType:                        &storageType,
				UseDefaultProcessorFeatures:        &trueFlag,
				VpcSecurityGroupIds:                vpcIds,
				ProcessorFeatures: []rds.ProcessorFeature{{
					Name:  &name,
					Value: &value,
				}},
				CloudwatchLogsExportConfiguration: &rdsCloudwatchLogsExportConfig,
			},
		},
		"Empty": {
			name:   emptyName,
			params: v1beta1.RDSInstanceParameters{},
			want: rds.ModifyDBInstanceInput{
				DBInstanceIdentifier: &emptyName,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := GenerateModifyDBInstanceInput(tc.name, &tc.params)
			if diff := cmp.Diff(&tc.want, got); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}
