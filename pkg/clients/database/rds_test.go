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
	"strconv"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/rds"
	rdstypes "github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/aws/smithy-go/document"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane-contrib/provider-aws/apis/database/v1beta1"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
)

var (
	allFieldsName                      = "allfieldsName"
	allocatedStorage             int32 = 20
	address                            = "address"
	arn                                = "my:arn"
	az                                 = "az"
	characterSetName                   = "utf8"
	clusterName                        = "testCluster"
	dbName                             = "example-name"
	dbSecurityGroups                   = []string{"test"}
	description                        = "testDescription"
	domain                             = "domain"
	enableCloudwatchExports            = []string{"test"}
	enabledCloudwatchExports           = []string{"test"}
	enabledCloudwatchExportsNone       = []string{}
	engine                             = "5.6.41"
	falseFlag                          = false
	iamRole                            = "iamRole"
	instanceClass                      = "class"
	kmsID                              = "kms"
	monitoring                         = 3
	monitoring32                       = int32(monitoring)
	multiAZ                            = true
	name                               = "testName"
	port                               = 123
	port32                             = int32(port)
	resourceID                         = "resource"
	retention                          = 2
	retention32                        = int32(retention)
	status                             = "testStatus"
	storage                            = 1
	storage32                          = int32(storage)
	storageType                        = "storageType"
	tier                               = 4
	tier32                             = int32(tier)
	trueFlag                           = true
	truncEngine                        = "5.6"
	username                           = "username"
	value                              = "testValue"
	vpc                                = "vpc"
	vpcIds                             = []string{"test"}
	window                             = "window"
	zone                               = "zone"

	secretNamespace      = "crossplane-system"
	connectionSecretName = "my-little-secret"
	connectionSecretKey  = "credentials"
	connectionCredData   = "confidential!"
	outputSecretName     = "my-saved-secret"

	errBoom = errors.New("boom")
)

func TestCreatePatch(t *testing.T) {
	type args struct {
		db *rdstypes.DBInstance
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
				db: &rdstypes.DBInstance{
					AllocatedStorage: allocatedStorage,
					CharacterSetName: &characterSetName,
					DBName:           &dbName,
				},
				p: &v1beta1.RDSInstanceParameters{
					AllocatedStorage: ptr.To(20),
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
				db: &rdstypes.DBInstance{
					AllocatedStorage: allocatedStorage,
					CharacterSetName: &characterSetName,
					DBName:           &dbName,
					AvailabilityZone: ptr.To("az1"),
				},
				p: &v1beta1.RDSInstanceParameters{
					AllocatedStorage: ptr.To(30),
					CharacterSetName: &characterSetName,
					DBName:           &dbName,
					AvailabilityZone: ptr.To("az2"),
				},
			},
			want: want{
				patch: &v1beta1.RDSInstanceParameters{
					AllocatedStorage: ptr.To(30),
					AvailabilityZone: ptr.To("az2"),
				},
			},
		},
		"IgnoreDifferentAvailabilityZoneForMultiAZ": {
			args: args{
				db: &rdstypes.DBInstance{
					AvailabilityZone: ptr.To("az1"),
					MultiAZ:          true,
				},
				p: &v1beta1.RDSInstanceParameters{
					AvailabilityZone: ptr.To("az2"),
					MultiAZ:          ptr.To(true),
				},
			},
			want: want{
				patch: &v1beta1.RDSInstanceParameters{},
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
		db   rdstypes.DBInstance
		r    v1beta1.RDSInstance
		kube client.Client
	}

	cases := map[string]struct {
		args args
		want bool
	}{
		"SameFields": {
			args: args{
				db: rdstypes.DBInstance{
					AllocatedStorage: allocatedStorage,
					CharacterSetName: &characterSetName,
					DBName:           &dbName,
				},
				r: v1beta1.RDSInstance{
					Spec: v1beta1.RDSInstanceSpec{
						ForProvider: v1beta1.RDSInstanceParameters{
							AllocatedStorage: ptr.To(20),
							CharacterSetName: &characterSetName,
							DBName:           &dbName,
						},
					},
				},
			},
			want: true,
		},
		"IgnoreDeletionOptions": {
			args: args{
				db: rdstypes.DBInstance{
					AllocatedStorage: allocatedStorage,
					CharacterSetName: &characterSetName,
					DBName:           &dbName,
				},
				r: v1beta1.RDSInstance{
					Spec: v1beta1.RDSInstanceSpec{
						ForProvider: v1beta1.RDSInstanceParameters{
							AllocatedStorage:                ptr.To(20),
							CharacterSetName:                &characterSetName,
							DBName:                          &dbName,
							DeleteAutomatedBackups:          pointer.ToOrNilIfZeroValue(true),
							SkipFinalSnapshotBeforeDeletion: pointer.ToOrNilIfZeroValue(true),
							FinalDBSnapshotIdentifier:       pointer.ToOrNilIfZeroValue("final"),
						},
					},
				},
			},
			want: true,
		},
		"DifferentFields": {
			args: args{
				db: rdstypes.DBInstance{
					AllocatedStorage: allocatedStorage,
					CharacterSetName: &characterSetName,
					DBName:           &dbName,
				},
				r: v1beta1.RDSInstance{
					Spec: v1beta1.RDSInstanceSpec{
						ForProvider: v1beta1.RDSInstanceParameters{
							AllocatedStorage: ptr.To(30),
							CharacterSetName: &characterSetName,
							DBName:           &dbName,
						},
					},
				},
			},
			want: false,
		},
		"IgnoresRefs": {
			args: args{
				db: rdstypes.DBInstance{
					DBName:        &dbName,
					DBSubnetGroup: &rdstypes.DBSubnetGroup{DBSubnetGroupName: &dbSubnetGroupName},
				},
				r: v1beta1.RDSInstance{
					Spec: v1beta1.RDSInstanceSpec{
						ForProvider: v1beta1.RDSInstanceParameters{
							DBName:               &dbName,
							DBSubnetGroupName:    &dbSubnetGroupName,
							DBSubnetGroupNameRef: &xpv1.Reference{Name: "coolgroup"},
						},
					},
				},
			},
			want: true,
		},
		"IgnoresDBName": {
			args: args{
				db: rdstypes.DBInstance{
					DBName: nil,
				},
				r: v1beta1.RDSInstance{
					Spec: v1beta1.RDSInstanceSpec{
						ForProvider: v1beta1.RDSInstanceParameters{
							DBName: &dbName,
						},
					},
				},
			},
			want: true,
		},
		"SamePassword": {
			args: args{
				db: rdstypes.DBInstance{
					DBName: &dbName,
				},
				r: v1beta1.RDSInstance{
					Spec: v1beta1.RDSInstanceSpec{
						ForProvider: v1beta1.RDSInstanceParameters{
							DBName: &dbName,
							MasterPasswordSecretRef: &xpv1.SecretKeySelector{
								SecretReference: xpv1.SecretReference{
									Name:      connectionSecretName,
									Namespace: secretNamespace,
								},
								Key: connectionSecretKey,
							},
						},
						ResourceSpec: xpv1.ResourceSpec{
							WriteConnectionSecretToReference: &xpv1.SecretReference{
								Name:      outputSecretName,
								Namespace: secretNamespace,
							},
						},
					},
				},
				kube: &test.MockClient{
					MockGet: func(_ context.Context, key client.ObjectKey, obj client.Object) error {
						switch key.Name {
						case connectionSecretName:
							secret := corev1.Secret{
								Data: map[string][]byte{},
							}
							secret.Data[connectionSecretKey] = []byte(connectionCredData)
							secret.DeepCopyInto(obj.(*corev1.Secret))
							return nil
						case outputSecretName:
							secret := corev1.Secret{
								Data: map[string][]byte{},
							}
							secret.Data[xpv1.ResourceCredentialsSecretPasswordKey] = []byte(connectionCredData)
							secret.DeepCopyInto(obj.(*corev1.Secret))
							return nil
						default:
							return nil
						}
					},
				},
			},
			want: true,
		},
		"DifferentPassword": {
			args: args{
				db: rdstypes.DBInstance{
					DBName: &dbName,
				},
				r: v1beta1.RDSInstance{
					Spec: v1beta1.RDSInstanceSpec{
						ForProvider: v1beta1.RDSInstanceParameters{
							DBName: &dbName,
							MasterPasswordSecretRef: &xpv1.SecretKeySelector{
								SecretReference: xpv1.SecretReference{
									Name:      connectionSecretName,
									Namespace: secretNamespace,
								},
								Key: connectionSecretKey,
							},
						},
						ResourceSpec: xpv1.ResourceSpec{
							WriteConnectionSecretToReference: &xpv1.SecretReference{
								Name:      outputSecretName,
								Namespace: secretNamespace,
							},
						},
					},
				},
				kube: &test.MockClient{
					MockGet: func(_ context.Context, key client.ObjectKey, obj client.Object) error {
						switch key.Name {
						case connectionSecretName:
							secret := corev1.Secret{
								Data: map[string][]byte{},
							}
							secret.Data[connectionSecretKey] = []byte(connectionCredData)
							secret.DeepCopyInto(obj.(*corev1.Secret))
							return nil
						case outputSecretName:
							secret := corev1.Secret{
								Data: map[string][]byte{},
							}
							secret.Data[xpv1.ResourceCredentialsSecretPasswordKey] = []byte("not" + connectionCredData)
							secret.DeepCopyInto(obj.(*corev1.Secret))
							return nil
						default:
							return nil
						}
					},
				},
			},
			want: false,
		},
		"EngineVersionUpgrade": {
			args: args{
				db: rdstypes.DBInstance{
					EngineVersion: ptr.To("12.3"),
				},
				r: v1beta1.RDSInstance{
					Spec: v1beta1.RDSInstanceSpec{
						ForProvider: v1beta1.RDSInstanceParameters{
							EngineVersion: ptr.To("12.7"),
						},
					},
				},
			},
			want: false,
		},
		"EngineVersionUpgradeMajorVersion": {
			args: args{
				db: rdstypes.DBInstance{
					EngineVersion: ptr.To("12.3"),
				},
				r: v1beta1.RDSInstance{
					Spec: v1beta1.RDSInstanceSpec{
						ForProvider: v1beta1.RDSInstanceParameters{
							EngineVersion: ptr.To("13.7"),
						},
					},
				},
			},
			want: false,
		},
		"EngineVersionMajorVersionOnly": {
			args: args{
				db: rdstypes.DBInstance{
					EngineVersion: ptr.To("12.3"),
				},
				r: v1beta1.RDSInstance{
					Spec: v1beta1.RDSInstanceSpec{
						ForProvider: v1beta1.RDSInstanceParameters{
							EngineVersion: ptr.To("12"),
						},
					},
				},
			},
			want: true,
		},
		"EngineVersionDowngrade": {
			args: args{
				db: rdstypes.DBInstance{
					EngineVersion: ptr.To("12.3"),
				},
				r: v1beta1.RDSInstance{
					Spec: v1beta1.RDSInstanceSpec{
						ForProvider: v1beta1.RDSInstanceParameters{
							EngineVersion: ptr.To("12.1"),
						},
					},
				},
			},
			want: true,
		},
		"NoUpdateForDifferentAvailabilityZoneWhenMultiAZ": {
			args: args{
				db: rdstypes.DBInstance{
					AvailabilityZone: ptr.To("az1"),
					MultiAZ:          true,
				},
				r: v1beta1.RDSInstance{
					Spec: v1beta1.RDSInstanceSpec{
						ForProvider: v1beta1.RDSInstanceParameters{
							AvailabilityZone: ptr.To("az2"),
							MultiAZ:          ptr.To(true),
						},
					},
				},
			},
			want: true,
		},
		"SameTags": {
			args: args{
				db: rdstypes.DBInstance{
					TagList: []rdstypes.Tag{
						{Key: ptr.To("tag1")},
						{Key: ptr.To("tag2")},
						{Key: ptr.To("tag3")},
					},
				},
				r: v1beta1.RDSInstance{
					Spec: v1beta1.RDSInstanceSpec{
						ForProvider: v1beta1.RDSInstanceParameters{
							Tags: []v1beta1.Tag{
								{Key: "tag1"},
								{Key: "tag2"},
								{Key: "tag3"},
							},
						},
					},
				},
			},
			want: true,
		},
		"SameTagsDifferentOrder": {
			args: args{
				db: rdstypes.DBInstance{
					TagList: []rdstypes.Tag{
						{Key: ptr.To("tag1"), Value: ptr.To("val")},
						{Key: ptr.To("tag2"), Value: ptr.To("val")},
						{Key: ptr.To("tag3"), Value: ptr.To("val")},
					},
				},
				r: v1beta1.RDSInstance{
					Spec: v1beta1.RDSInstanceSpec{
						ForProvider: v1beta1.RDSInstanceParameters{
							Tags: []v1beta1.Tag{
								{Key: "tag3", Value: "val"},
								{Key: "tag2", Value: "val"},
								{Key: "tag1", Value: "val"},
							},
						},
					},
				},
			},
			want: true,
		},
		"DifferentTags": {
			args: args{
				db: rdstypes.DBInstance{
					TagList: []rdstypes.Tag{
						{Key: ptr.To("tag1")},
						{Key: ptr.To("tag2")},
						{Key: ptr.To("tag3")},
					},
				},
				r: v1beta1.RDSInstance{
					Spec: v1beta1.RDSInstanceSpec{
						ForProvider: v1beta1.RDSInstanceParameters{
							Tags: []v1beta1.Tag{
								{Key: "tag1"},
								{Key: "tag5"},
								{Key: "tag6"},
							},
						},
					},
				},
			},
			want: false,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			got, _, _, _, _ := IsUpToDate(ctx, tc.args.kube, &tc.args.r, tc.args.db)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGetPassword(t *testing.T) {
	type args struct {
		in   *xpv1.SecretKeySelector
		out  *xpv1.SecretReference
		kube client.Client
	}
	type want struct {
		Pwd     string
		Changed bool
		Err     error
	}

	cases := map[string]struct {
		args args
		want want
	}{
		"SamePassword": {
			args: args{
				in: &xpv1.SecretKeySelector{
					SecretReference: xpv1.SecretReference{
						Name:      connectionSecretName,
						Namespace: secretNamespace,
					},
					Key: connectionSecretKey,
				},
				out: &xpv1.SecretReference{
					Name:      outputSecretName,
					Namespace: secretNamespace,
				},
				kube: &test.MockClient{
					MockGet: func(_ context.Context, key client.ObjectKey, obj client.Object) error {
						switch key.Name {
						case connectionSecretName:
							secret := corev1.Secret{
								Data: map[string][]byte{},
							}
							secret.Data[connectionSecretKey] = []byte(connectionCredData)
							secret.DeepCopyInto(obj.(*corev1.Secret))
							return nil
						case outputSecretName:
							secret := corev1.Secret{
								Data: map[string][]byte{},
							}
							secret.Data[xpv1.ResourceCredentialsSecretPasswordKey] = []byte(connectionCredData)
							secret.DeepCopyInto(obj.(*corev1.Secret))
							return nil
						default:
							return nil
						}
					},
				},
			},
			want: want{
				Pwd:     connectionCredData,
				Changed: false,
				Err:     nil,
			},
		},
		"DifferentPassword": {
			args: args{
				in: &xpv1.SecretKeySelector{
					SecretReference: xpv1.SecretReference{
						Name:      connectionSecretName,
						Namespace: secretNamespace,
					},
					Key: connectionSecretKey,
				},
				out: &xpv1.SecretReference{
					Name:      outputSecretName,
					Namespace: secretNamespace,
				},
				kube: &test.MockClient{
					MockGet: func(_ context.Context, key client.ObjectKey, obj client.Object) error {
						switch key.Name {
						case connectionSecretName:
							secret := corev1.Secret{
								Data: map[string][]byte{},
							}
							secret.Data[connectionSecretKey] = []byte(connectionCredData)
							secret.DeepCopyInto(obj.(*corev1.Secret))
							return nil
						case outputSecretName:
							secret := corev1.Secret{
								Data: map[string][]byte{},
							}
							secret.Data[xpv1.ResourceCredentialsSecretPasswordKey] = []byte("not" + connectionCredData)
							secret.DeepCopyInto(obj.(*corev1.Secret))
							return nil
						default:
							return nil
						}
					},
				},
			},
			want: want{
				Pwd:     connectionCredData,
				Changed: true,
				Err:     nil,
			},
		},
		"ErrorOnInput": {
			args: args{
				in: &xpv1.SecretKeySelector{
					SecretReference: xpv1.SecretReference{
						Name:      connectionSecretName,
						Namespace: secretNamespace,
					},
					Key: connectionSecretKey,
				},
				kube: &test.MockClient{
					MockGet: func(_ context.Context, key client.ObjectKey, obj client.Object) error {
						return errBoom
					},
				},
			},
			want: want{
				Pwd:     "",
				Changed: false,
				Err:     errors.Wrap(errBoom, errGetPasswordSecretFailed),
			},
		},
		"OutputDoesNotExistYet": {
			args: args{
				in: &xpv1.SecretKeySelector{
					SecretReference: xpv1.SecretReference{
						Name:      connectionSecretName,
						Namespace: secretNamespace,
					},
					Key: connectionSecretKey,
				},
				out: &xpv1.SecretReference{
					Name:      outputSecretName,
					Namespace: secretNamespace,
				},
				kube: &test.MockClient{
					MockGet: func(_ context.Context, key client.ObjectKey, obj client.Object) error {
						switch key.Name {
						case connectionSecretName:
							secret := corev1.Secret{
								Data: map[string][]byte{},
							}
							secret.Data[connectionSecretKey] = []byte(connectionCredData)
							secret.DeepCopyInto(obj.(*corev1.Secret))
							return nil
						case outputSecretName:
							return kerrors.NewNotFound(schema.GroupResource{
								Resource: "Secret",
							}, outputSecretName)
						default:
							return nil
						}
					},
				},
			},
			want: want{
				Pwd:     connectionCredData,
				Changed: true,
				Err:     nil,
			},
		},
		"NoInputPassword": {
			args: args{
				out: &xpv1.SecretReference{
					Name:      outputSecretName,
					Namespace: secretNamespace,
				},
				kube: &test.MockClient{
					MockGet: func(_ context.Context, key client.ObjectKey, obj client.Object) error {
						secret := corev1.Secret{
							Data: map[string][]byte{},
						}
						secret.Data[xpv1.ResourceCredentialsSecretPasswordKey] = []byte("not" + connectionCredData)
						secret.DeepCopyInto(obj.(*corev1.Secret))
						return nil
					},
				},
			},
			want: want{
				Pwd:     "",
				Changed: false,
				Err:     nil,
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			pwd, changed, err := GetPassword(ctx, tc.args.kube, tc.args.in, tc.args.out)
			if diff := cmp.Diff(tc.want, want{
				Pwd:     pwd,
				Changed: changed,
				Err:     err,
			}, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGenerateObservation(t *testing.T) {
	lastRestoreTime, createTime := time.Now(), time.Now()
	rdsAz := rdstypes.AvailabilityZone{Name: &name}
	replicaSourceIdentifier := "replicaSource"
	secondaryAZ := "secondary"
	normal := true
	replicaClusters := []string{"replicaCluster"}
	subnetGroup := rdstypes.DBSubnetGroup{
		DBSubnetGroupArn:         &arn,
		DBSubnetGroupDescription: &description,
		DBSubnetGroupName:        &name,
		SubnetGroupStatus:        &status,
		VpcId:                    &vpc,
	}
	subnetGroup.Subnets = []rdstypes.Subnet{{
		SubnetIdentifier:       &name,
		SubnetStatus:           &status,
		SubnetAvailabilityZone: &rdsAz,
	}}
	endpoint := rdstypes.Endpoint{
		Address:      &address,
		HostedZoneId: &zone,
		Port:         port32,
	}
	pendingModifiedValues := rdstypes.PendingModifiedValues{
		AllocatedStorage:        &storage32,
		BackupRetentionPeriod:   &retention32,
		CACertificateIdentifier: &name,
		DBInstanceClass:         &instanceClass,
		DBSubnetGroupName:       &name,
		Iops:                    &storage32,
		LicenseModel:            &name,
		MultiAZ:                 &multiAZ,
		Port:                    &port32,
		StorageType:             &storageType,
	}
	pendingCloudwatch := rdstypes.PendingCloudwatchLogsExports{
		LogTypesToDisable: nil,
		LogTypesToEnable:  enableCloudwatchExports,
	}
	pendingModifiedValues.PendingCloudwatchLogsExports = &pendingCloudwatch
	pendingModifiedValues.ProcessorFeatures = []rdstypes.ProcessorFeature{{
		Name:  &name,
		Value: &value,
	}}

	cases := map[string]struct {
		rds  rdstypes.DBInstance
		want v1beta1.RDSInstanceObservation
	}{
		"AllFields": {
			rds: rdstypes.DBInstance{
				DBInstanceStatus:                      &status,
				DBInstanceArn:                         &arn,
				InstanceCreateTime:                    &createTime,
				DbInstancePort:                        port32,
				DbiResourceId:                         &resourceID,
				BackupRetentionPeriod:                 retention32,
				EnabledCloudwatchLogsExports:          enabledCloudwatchExports,
				EnhancedMonitoringResourceArn:         &arn,
				PerformanceInsightsEnabled:            &trueFlag,
				ReadReplicaDBClusterIdentifiers:       replicaClusters,
				ReadReplicaDBInstanceIdentifiers:      replicaClusters,
				ReadReplicaSourceDBInstanceIdentifier: &replicaSourceIdentifier,
				SecondaryAvailabilityZone:             &secondaryAZ,
				LatestRestorableTime:                  &lastRestoreTime,
				DBParameterGroups:                     []rdstypes.DBParameterGroupStatus{{DBParameterGroupName: &name}},
				DBSecurityGroups:                      []rdstypes.DBSecurityGroupMembership{{DBSecurityGroupName: &name, Status: &status}},
				DBSubnetGroup:                         &subnetGroup,
				DomainMemberships: []rdstypes.DomainMembership{{
					Domain:      &domain,
					FQDN:        &name,
					IAMRoleName: &name,
					Status:      &status,
				}},
				Endpoint: &endpoint,
				OptionGroupMemberships: []rdstypes.OptionGroupMembership{{
					OptionGroupName: &name,
					Status:          &status,
				}},
				PendingModifiedValues: &pendingModifiedValues,
				StatusInfos: []rdstypes.DBInstanceStatusInfo{{
					Message:    &status,
					Status:     &status,
					StatusType: &status,
					Normal:     normal,
				}},
				VpcSecurityGroups: []rdstypes.VpcSecurityGroupMembership{{
					Status:             &status,
					VpcSecurityGroupId: &name,
				}},
			},
			want: v1beta1.RDSInstanceObservation{
				DBInstanceStatus:  status,
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
						SubnetAvailabilityZone: v1beta1.AvailabilityZone{Name: *rdsAz.Name},
					}},
				},
				DBInstancePort:                port,
				DBResourceID:                  resourceID,
				BackupRetentionPeriod:         retention,
				DomainMemberships:             []v1beta1.DomainMembership{{Domain: domain, FQDN: name, IAMRoleName: name, Status: status}},
				InstanceCreateTime:            &metav1.Time{Time: createTime},
				Endpoint:                      v1beta1.Endpoint{Port: port, HostedZoneID: zone, Address: address},
				EnabledCloudwatchLogsExports:  enabledCloudwatchExports,
				EnhancedMonitoringResourceArn: arn,
				LatestRestorableTime:          &metav1.Time{Time: lastRestoreTime},
				OptionGroupMemberships:        []v1beta1.OptionGroupMembership{{OptionGroupName: name, Status: status}},
				PendingModifiedValues: v1beta1.PendingModifiedValues{
					AllocatedStorage:        storage,
					BackupRetentionPeriod:   retention,
					CACertificateIdentifier: name,
					DBInstanceClass:         instanceClass,
					DBSubnetGroupName:       name,
					IOPS:                    storage,
					LicenseModel:            name,
					MultiAZ:                 multiAZ,
					Port:                    port,
					StorageType:             storageType,
					PendingCloudwatchLogsExports: v1beta1.PendingCloudwatchLogsExports{
						LogTypesToDisable: nil,
						LogTypesToEnable:  enableCloudwatchExports,
					},
					ProcessorFeatures: []v1beta1.ProcessorFeature{{Name: name, Value: value}},
				},
				PerformanceInsightsEnabled:            trueFlag,
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
			rds: rdstypes.DBInstance{
				DBInstanceStatus:                      &status,
				DBInstanceArn:                         &arn,
				InstanceCreateTime:                    &createTime,
				DbInstancePort:                        port32,
				DbiResourceId:                         &resourceID,
				EnhancedMonitoringResourceArn:         &arn,
				PerformanceInsightsEnabled:            &trueFlag,
				ReadReplicaDBClusterIdentifiers:       replicaClusters,
				ReadReplicaDBInstanceIdentifiers:      replicaClusters,
				ReadReplicaSourceDBInstanceIdentifier: &replicaSourceIdentifier,
				SecondaryAvailabilityZone:             &secondaryAZ,
				LatestRestorableTime:                  &lastRestoreTime,
				DomainMemberships: []rdstypes.DomainMembership{{
					Domain:      &domain,
					FQDN:        &name,
					IAMRoleName: &name,
					Status:      &status,
				}},
				Endpoint: &endpoint,
				OptionGroupMemberships: []rdstypes.OptionGroupMembership{{
					OptionGroupName: &name,
					Status:          &status,
				}},
				StatusInfos: []rdstypes.DBInstanceStatusInfo{{
					Message:    &status,
					Status:     &status,
					StatusType: &status,
					Normal:     normal,
				}},
				VpcSecurityGroups: []rdstypes.VpcSecurityGroupMembership{{
					Status:             &status,
					VpcSecurityGroupId: &name,
				}},
			},
			want: v1beta1.RDSInstanceObservation{
				DBInstanceStatus:                      status,
				DBInstanceArn:                         arn,
				DBInstancePort:                        port,
				DBResourceID:                          resourceID,
				DomainMemberships:                     []v1beta1.DomainMembership{{Domain: domain, FQDN: name, IAMRoleName: name, Status: status}},
				InstanceCreateTime:                    &metav1.Time{Time: createTime},
				Endpoint:                              v1beta1.Endpoint{Port: port, HostedZoneID: zone, Address: address},
				EnhancedMonitoringResourceArn:         arn,
				LatestRestorableTime:                  &metav1.Time{Time: lastRestoreTime},
				OptionGroupMemberships:                []v1beta1.OptionGroupMembership{{OptionGroupName: name, Status: status}},
				PerformanceInsightsEnabled:            trueFlag,
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
			rds:  rdstypes.DBInstance{},
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
				xpv1.ResourceCredentialsSecretEndpointKey: []byte(address),
				xpv1.ResourceCredentialsSecretPortKey:     []byte(strconv.Itoa(port)),
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
	existingName := "existing"
	subnetGroup := rdstypes.DBSubnetGroup{
		DBSubnetGroupArn:         &arn,
		DBSubnetGroupDescription: &description,
		DBSubnetGroupName:        &name,
		SubnetGroupStatus:        &status,
		VpcId:                    &vpc,
	}

	cases := map[string]struct {
		rds    rdstypes.DBInstance
		params v1beta1.RDSInstanceParameters
		want   v1beta1.RDSInstanceParameters
	}{
		"AllFields": {
			rds: rdstypes.DBInstance{
				AllocatedStorage:                   storage32,
				DBInstanceClass:                    &instanceClass,
				Engine:                             &engine,
				AutoMinorVersionUpgrade:            trueFlag,
				AvailabilityZone:                   &az,
				BackupRetentionPeriod:              storage32,
				CACertificateIdentifier:            &name,
				CharacterSetName:                   &name,
				CopyTagsToSnapshot:                 trueFlag,
				DBClusterIdentifier:                &clusterName,
				DBName:                             &name,
				DeletionProtection:                 trueFlag,
				IAMDatabaseAuthenticationEnabled:   trueFlag,
				PerformanceInsightsEnabled:         &trueFlag,
				Iops:                               &storage32,
				KmsKeyId:                           &kmsID,
				LicenseModel:                       &name,
				MasterUsername:                     &username,
				MonitoringInterval:                 &monitoring32,
				MonitoringRoleArn:                  &arn,
				MultiAZ:                            multiAZ,
				PerformanceInsightsKMSKeyId:        &kmsID,
				PerformanceInsightsRetentionPeriod: &retention32,
				Endpoint:                           &rdstypes.Endpoint{Port: port32},
				PreferredBackupWindow:              &window,
				PreferredMaintenanceWindow:         &window,
				PromotionTier:                      &tier32,
				PubliclyAccessible:                 trueFlag,
				StorageEncrypted:                   trueFlag,
				StorageType:                        &storageType,
				Timezone:                           &zone,
				DBSecurityGroups:                   []rdstypes.DBSecurityGroupMembership{{DBSecurityGroupName: &name, Status: &status}},
				DBSubnetGroup:                      &subnetGroup,
				EnabledCloudwatchLogsExports:       enabledCloudwatchExports,
				ProcessorFeatures: []rdstypes.ProcessorFeature{{
					Name:  &name,
					Value: &value,
				}},
				VpcSecurityGroups: []rdstypes.VpcSecurityGroupMembership{{
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
				EnableCloudwatchLogsExports:        nil,
				ProcessorFeatures: []v1beta1.ProcessorFeature{{
					Name:  name,
					Value: value,
				}},
				VPCSecurityGroupIDs: []string{name},
				EngineVersion:       &engine,
			},
		},
		"SubnetGroupNameSet": {
			rds: rdstypes.DBInstance{
				DBSubnetGroup: &subnetGroup,
			},
			params: v1beta1.RDSInstanceParameters{},
			want: v1beta1.RDSInstanceParameters{
				DBSubnetGroupName:               subnetGroup.DBSubnetGroupName,
				AutoMinorVersionUpgrade:         &falseFlag,
				BackupRetentionPeriod:           new(int),
				CopyTagsToSnapshot:              &falseFlag,
				DeletionProtection:              &falseFlag,
				EnableIAMDatabaseAuthentication: &falseFlag,
				MultiAZ:                         &falseFlag,
				PubliclyAccessible:              &falseFlag,
				StorageEncrypted:                &falseFlag,
			},
		},
		"SubnetGroupNameNotOverwritten": {
			rds: rdstypes.DBInstance{
				DBSubnetGroup: &subnetGroup,
			},
			params: v1beta1.RDSInstanceParameters{
				DBSubnetGroupName: &existingName,
			},
			want: v1beta1.RDSInstanceParameters{
				DBSubnetGroupName:               &existingName,
				AutoMinorVersionUpgrade:         &falseFlag,
				BackupRetentionPeriod:           new(int),
				CopyTagsToSnapshot:              &falseFlag,
				DeletionProtection:              &falseFlag,
				EnableIAMDatabaseAuthentication: &falseFlag,
				MultiAZ:                         &falseFlag,
				PubliclyAccessible:              &falseFlag,
				StorageEncrypted:                &falseFlag,
			},
		},
		"SecurityGroupNotOverwritten": {
			rds: rdstypes.DBInstance{
				DBSecurityGroups: []rdstypes.DBSecurityGroupMembership{{DBSecurityGroupName: &name, Status: &status}},
			},
			params: v1beta1.RDSInstanceParameters{
				DBSecurityGroups: []string{"newGroup"},
			},
			want: v1beta1.RDSInstanceParameters{
				DBSecurityGroups: []string{"newGroup"},

				AutoMinorVersionUpgrade:         &falseFlag,
				BackupRetentionPeriod:           new(int),
				CopyTagsToSnapshot:              &falseFlag,
				DeletionProtection:              &falseFlag,
				EnableIAMDatabaseAuthentication: &falseFlag,
				MultiAZ:                         &falseFlag,
				PubliclyAccessible:              &falseFlag,
				StorageEncrypted:                &falseFlag,
			},
		},
		"ProcessorFeaturesNotOverwritten": {
			rds: rdstypes.DBInstance{
				ProcessorFeatures: []rdstypes.ProcessorFeature{{
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

				AutoMinorVersionUpgrade:         &falseFlag,
				BackupRetentionPeriod:           new(int),
				CopyTagsToSnapshot:              &falseFlag,
				DeletionProtection:              &falseFlag,
				EnableIAMDatabaseAuthentication: &falseFlag,
				MultiAZ:                         &falseFlag,
				PubliclyAccessible:              &falseFlag,
				StorageEncrypted:                &falseFlag,
			},
		},
		"VPCSecurityGroupIdsNotOverwritten": {
			rds: rdstypes.DBInstance{
				ProcessorFeatures: []rdstypes.ProcessorFeature{{
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

				AutoMinorVersionUpgrade:         &falseFlag,
				BackupRetentionPeriod:           new(int),
				CopyTagsToSnapshot:              &falseFlag,
				DeletionProtection:              &falseFlag,
				EnableIAMDatabaseAuthentication: &falseFlag,
				MultiAZ:                         &falseFlag,
				PubliclyAccessible:              &falseFlag,
				StorageEncrypted:                &falseFlag,
			},
		},
		"EngineVersion": {
			rds: rdstypes.DBInstance{
				EngineVersion: &engine,
			},
			params: v1beta1.RDSInstanceParameters{
				EngineVersion: &truncEngine,
			},
			want: v1beta1.RDSInstanceParameters{
				EngineVersion: &engine,

				AutoMinorVersionUpgrade:         &falseFlag,
				BackupRetentionPeriod:           new(int),
				CopyTagsToSnapshot:              &falseFlag,
				DeletionProtection:              &falseFlag,
				EnableIAMDatabaseAuthentication: &falseFlag,
				MultiAZ:                         &falseFlag,
				PubliclyAccessible:              &falseFlag,
				StorageEncrypted:                &falseFlag,
			},
		},
		"EmptyInstance": {
			rds:    rdstypes.DBInstance{},
			params: v1beta1.RDSInstanceParameters{},
			want: v1beta1.RDSInstanceParameters{
				AutoMinorVersionUpgrade:         &falseFlag,
				BackupRetentionPeriod:           new(int),
				CopyTagsToSnapshot:              &falseFlag,
				DeletionProtection:              &falseFlag,
				EnableIAMDatabaseAuthentication: &falseFlag,
				MultiAZ:                         &falseFlag,
				PubliclyAccessible:              &falseFlag,
				StorageEncrypted:                &falseFlag,
			},
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
	emptyName := "emptyProcessor"

	cases := map[string]struct {
		name   string
		params v1beta1.RDSInstanceParameters
		db     rdstypes.DBInstance
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
				CopyTagsToSnapshot:                 &trueFlag,
				DBClusterIdentifier:                &clusterName,
				DBName:                             &name,
				DBParameterGroupName:               &name,
				DeletionProtection:                 &trueFlag,
				Domain:                             &domain,
				DomainIAMRoleName:                  &iamRole,
				EnableCloudwatchLogsExports:        enableCloudwatchExports,
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
				ProcessorFeatures: []v1beta1.ProcessorFeature{{
					Name:  name,
					Value: value,
				}},
				VPCSecurityGroupIDs:         vpcIds,
				EngineVersion:               &engine,
				UseDefaultProcessorFeatures: &trueFlag,
			},
			db: rdstypes.DBInstance{
				EnabledCloudwatchLogsExports: enabledCloudwatchExportsNone,
			},
			want: rds.ModifyDBInstanceInput{
				DBInstanceIdentifier:     &allFieldsName,
				AllocatedStorage:         &storage32,
				AllowMajorVersionUpgrade: trueFlag,
				ApplyImmediately:         trueFlag,
				AutoMinorVersionUpgrade:  &trueFlag,
				BackupRetentionPeriod:    &retention32,
				CACertificateIdentifier:  &name,
				CloudwatchLogsExportConfiguration: &rdstypes.CloudwatchLogsExportConfiguration{
					DisableLogTypes: []string{},
					EnableLogTypes:  enableCloudwatchExports,
				},
				CopyTagsToSnapshot:                 &trueFlag,
				DBInstanceClass:                    &instanceClass,
				DBParameterGroupName:               &name,
				DBPortNumber:                       &port32,
				DBSecurityGroups:                   dbSecurityGroups,
				DBSubnetGroupName:                  &name,
				DeletionProtection:                 &trueFlag,
				Domain:                             &domain,
				DomainIAMRoleName:                  &iamRole,
				EnableIAMDatabaseAuthentication:    &trueFlag,
				EnablePerformanceInsights:          &trueFlag,
				EngineVersion:                      &engine,
				Iops:                               &storage32,
				LicenseModel:                       &name,
				MonitoringInterval:                 &monitoring32,
				MonitoringRoleArn:                  &arn,
				MultiAZ:                            &multiAZ,
				OptionGroupName:                    &name,
				PerformanceInsightsRetentionPeriod: &retention32,
				PerformanceInsightsKMSKeyId:        &kmsID,
				PreferredBackupWindow:              &window,
				PreferredMaintenanceWindow:         &window,
				PromotionTier:                      &tier32,
				PubliclyAccessible:                 &trueFlag,
				StorageType:                        &storageType,
				UseDefaultProcessorFeatures:        &trueFlag,
				VpcSecurityGroupIds:                vpcIds,
				ProcessorFeatures: []rdstypes.ProcessorFeature{{
					Name:  &name,
					Value: &value,
				}},
			},
		},
		"Empty": {
			name:   emptyName,
			params: v1beta1.RDSInstanceParameters{},
			want: rds.ModifyDBInstanceInput{
				DBInstanceIdentifier: &emptyName,
				CloudwatchLogsExportConfiguration: &rdstypes.CloudwatchLogsExportConfiguration{
					DisableLogTypes: []string{},
					EnableLogTypes:  []string{},
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := GenerateModifyDBInstanceInput(tc.name, &tc.params, &tc.db)
			if diff := cmp.Diff(&tc.want, got, cmpopts.IgnoreTypes(document.NoSerde{})); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGenerateCreateRDSInstanceInput(t *testing.T) {
	cases := map[string]struct {
		name     string
		password string
		params   v1beta1.RDSInstanceParameters
		want     rds.CreateDBInstanceInput
	}{
		"AllFields": {
			name: allFieldsName,
			params: v1beta1.RDSInstanceParameters{
				AllocatedStorage:                   &storage,
				DBInstanceClass:                    instanceClass,
				ApplyModificationsImmediately:      &trueFlag,
				Engine:                             engine,
				EngineVersion:                      &engine,
				AutoMinorVersionUpgrade:            &trueFlag,
				AllowMajorVersionUpgrade:           &trueFlag,
				AvailabilityZone:                   &az,
				BackupRetentionPeriod:              &retention,
				CACertificateIdentifier:            &name,
				CharacterSetName:                   &name,
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
				EnableCloudwatchLogsExports:        enableCloudwatchExports,
				ProcessorFeatures: []v1beta1.ProcessorFeature{{
					Name:  name,
					Value: value,
				}},
				VPCSecurityGroupIDs:         vpcIds,
				UseDefaultProcessorFeatures: &trueFlag,
			},
			want: rds.CreateDBInstanceInput{
				DBInstanceIdentifier:               &allFieldsName,
				AvailabilityZone:                   &az,
				AllocatedStorage:                   &storage32,
				AutoMinorVersionUpgrade:            &trueFlag,
				BackupRetentionPeriod:              &retention32,
				CACertificateIdentifier:            &name,
				CharacterSetName:                   &name,
				CopyTagsToSnapshot:                 &trueFlag,
				DBClusterIdentifier:                &clusterName,
				DBInstanceClass:                    &instanceClass,
				DBName:                             &name,
				DBParameterGroupName:               &name,
				DBSecurityGroups:                   dbSecurityGroups,
				DBSubnetGroupName:                  &name,
				DeletionProtection:                 &trueFlag,
				Domain:                             &domain,
				DomainIAMRoleName:                  &iamRole,
				EnableIAMDatabaseAuthentication:    &trueFlag,
				EnableCloudwatchLogsExports:        enableCloudwatchExports,
				EnablePerformanceInsights:          &trueFlag,
				Engine:                             &engine,
				EngineVersion:                      &engine,
				Iops:                               &storage32,
				KmsKeyId:                           &kmsID,
				LicenseModel:                       &name,
				MasterUsername:                     &username,
				MonitoringInterval:                 &monitoring32,
				MonitoringRoleArn:                  &arn,
				MultiAZ:                            &multiAZ,
				OptionGroupName:                    &name,
				PerformanceInsightsRetentionPeriod: &retention32,
				PerformanceInsightsKMSKeyId:        &kmsID,
				Port:                               &port32,
				PreferredBackupWindow:              &window,
				PreferredMaintenanceWindow:         &window,
				PromotionTier:                      &tier32,
				PubliclyAccessible:                 &trueFlag,
				StorageEncrypted:                   &trueFlag,
				StorageType:                        &storageType,
				Timezone:                           &zone,
				VpcSecurityGroupIds:                vpcIds,
				ProcessorFeatures: []rdstypes.ProcessorFeature{{
					Name:  &name,
					Value: &value,
				}},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := GenerateCreateRDSInstanceInput(tc.name, tc.password, &tc.params)
			if diff := cmp.Diff(&tc.want, got, cmpopts.IgnoreTypes(document.NoSerde{})); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}
