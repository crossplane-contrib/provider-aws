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
package rdsinstance

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsrds "github.com/aws/aws-sdk-go-v2/service/rds"
	awsrdstypes "github.com/aws/aws-sdk-go-v2/service/rds/types"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane-contrib/provider-aws/apis/database/v1beta1"
	rds "github.com/crossplane-contrib/provider-aws/pkg/clients/database"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/database/fake"
	errorutils "github.com/crossplane-contrib/provider-aws/pkg/utils/errors"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
)

const (
	secretKey = "credentials"
	credData  = "confidential!"
)

var (
	masterUsername                  = "root"
	engineVersion                   = "5.6"
	s3SourceType                    = "S3"
	snapshotSourceType              = "Snapshot"
	pointInTimeSourceType           = "PointInTime"
	s3BucketName                    = "database-backup"
	backupWindow                    = "21:00-23:00"
	snapshotIdentifier              = "my-snapshot"
	pointInTimeDBInstanceIdentifier = "my-instance"
	awsBackupRecoveryPointARN       = "arn:aws:backup:us-east-1:123456789012:recovery-point:1EB3B5E7-9EB-A80B-108B488B0D45"
	s3Backup                        = v1beta1.RestoreBackupConfiguration{
		Source: &s3SourceType,
		S3: &v1beta1.S3RestoreBackupConfiguration{
			BucketName: &s3BucketName,
		},
	}
	snapshotBackup = v1beta1.RestoreBackupConfiguration{
		Source: &snapshotSourceType,
		Snapshot: &v1beta1.SnapshotRestoreBackupConfiguration{
			SnapshotIdentifier: &snapshotIdentifier,
		},
	}
	pointInTimeRestoreFrom, _ = time.Parse(time.RFC3339, "2011-09-07T23:45:00Z")
	pointInTimeBackup         = v1beta1.RestoreBackupConfiguration{
		Source: &pointInTimeSourceType,
		PointInTime: &v1beta1.PointInTimeRestoreBackupConfiguration{
			RestoreTime:                FromTimePtr(&pointInTimeRestoreFrom),
			SourceDBInstanceIdentifier: &pointInTimeDBInstanceIdentifier,
		},
	}
	pointInTimeLatestRestorableBackup = v1beta1.RestoreBackupConfiguration{
		Source: &pointInTimeSourceType,
		PointInTime: &v1beta1.PointInTimeRestoreBackupConfiguration{
			UseLatestRestorableTime:    true,
			SourceDBInstanceIdentifier: &pointInTimeDBInstanceIdentifier,
		},
	}

	replaceMe = "replace-me!"
	errBoom   = errors.New("boom")
)

// FromTimePtr is a helper for converting a *time.Time to a *metav1.Time
func FromTimePtr(t *time.Time) *metav1.Time {
	if t != nil {
		m := metav1.NewTime(*t)
		return &m
	}
	return nil
}

type args struct {
	rds   rds.Client
	kube  client.Client
	cache Cache
	cr    *v1beta1.RDSInstance
}

type rdsModifier func(*v1beta1.RDSInstance)

func withMasterUsername(s *string) rdsModifier {
	return func(r *v1beta1.RDSInstance) { r.Spec.ForProvider.MasterUsername = s }
}

func withBackupConfiguration(backup *v1beta1.RestoreBackupConfiguration) rdsModifier {
	return func(r *v1beta1.RDSInstance) { r.Spec.ForProvider.RestoreFrom = backup }
}

func withConditions(c ...xpv1.Condition) rdsModifier {
	return func(r *v1beta1.RDSInstance) { r.Status.ConditionedStatus.Conditions = c }
}

func withEngineVersion(s *string) rdsModifier {
	return func(r *v1beta1.RDSInstance) { r.Spec.ForProvider.EngineVersion = s }
}

func withTags(tagMaps ...map[string]string) rdsModifier {
	var tagList []v1beta1.Tag
	for _, tagMap := range tagMaps {
		for k, v := range tagMap {
			tagList = append(tagList, v1beta1.Tag{Key: k, Value: v})
		}
	}
	return func(r *v1beta1.RDSInstance) { r.Spec.ForProvider.Tags = tagList }
}

func withDBInstanceStatus(s string) rdsModifier {
	return func(r *v1beta1.RDSInstance) { r.Status.AtProvider.DBInstanceStatus = s }
}

func withAllocatedStorage(i int) rdsModifier {
	return func(r *v1beta1.RDSInstance) { r.Spec.ForProvider.AllocatedStorage = &i }
}

func withMaxAllocatedStorage(i int) rdsModifier {
	return func(r *v1beta1.RDSInstance) { r.Spec.ForProvider.MaxAllocatedStorage = &i }
}

func withStatusAllocatedStorage(i int) rdsModifier {
	return func(r *v1beta1.RDSInstance) { r.Status.AtProvider.AllocatedStorage = i }
}

func withPasswordSecretRef(s xpv1.SecretKeySelector) rdsModifier {
	return func(r *v1beta1.RDSInstance) { r.Spec.ForProvider.MasterPasswordSecretRef = &s }
}

func withDeleteAutomatedBackups(b bool) rdsModifier {
	return func(r *v1beta1.RDSInstance) { r.Spec.ForProvider.DeleteAutomatedBackups = &b }
}

func withBackupRetentionPeriod(i int) rdsModifier {
	return func(r *v1beta1.RDSInstance) { r.Spec.ForProvider.BackupRetentionPeriod = &i }
}

func withPreferredBackupWindow(s string) rdsModifier {
	return func(r *v1beta1.RDSInstance) { r.Spec.ForProvider.PreferredBackupWindow = &s }
}

func withStatusBackupRetentionPeriod(i int) rdsModifier {
	return func(r *v1beta1.RDSInstance) { r.Status.AtProvider.BackupRetentionPeriod = i }
}

func withStatusAWSBackupRecoveryPointARN(s string) rdsModifier {
	return func(r *v1beta1.RDSInstance) { r.Status.AtProvider.AWSBackupRecoveryPointARN = s }
}

func instance(m ...rdsModifier) *v1beta1.RDSInstance {
	falseFlag := false
	cr := &v1beta1.RDSInstance{
		Spec: v1beta1.RDSInstanceSpec{
			ForProvider: v1beta1.RDSInstanceParameters{
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
	for _, f := range m {
		f(cr)
	}
	return cr
}

func cache(addTagsMap map[string]string, removeTagsMap map[string]string) Cache {
	c := Cache{}
	for k, v := range addTagsMap {
		c.AddTags = append(c.AddTags, awsrdstypes.Tag{Key: pointer.ToOrNilIfZeroValue(k), Value: pointer.ToOrNilIfZeroValue(v)})
	}
	for k := range removeTagsMap {
		c.RemoveTags = append(c.RemoveTags, k)
	}
	return c
}

var _ managed.ExternalClient = &external{}
var _ managed.ExternalConnecter = &connector{}

func TestObserve(t *testing.T) {
	type want struct {
		cr     *v1beta1.RDSInstance
		result managed.ExternalObservation
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"SuccessfulAvailable": {
			args: args{
				rds: &fake.MockRDSClient{
					MockDescribe: func(ctx context.Context, input *awsrds.DescribeDBInstancesInput, opts []func(*awsrds.Options)) (*awsrds.DescribeDBInstancesOutput, error) {
						return &awsrds.DescribeDBInstancesOutput{
							DBInstances: []awsrdstypes.DBInstance{
								{
									DBInstanceStatus: aws.String(string(v1beta1.RDSInstanceStateAvailable)),
								},
							},
						}, nil
					},
				},
				cr: instance(),
			},
			want: want{
				cr: instance(
					withConditions(xpv1.Available()),
					withDBInstanceStatus(string(v1beta1.RDSInstanceStateAvailable))),
				result: managed.ExternalObservation{
					ResourceExists:    true,
					ResourceUpToDate:  true,
					ConnectionDetails: rds.GetConnectionDetails(v1beta1.RDSInstance{}),
				},
			},
		},
		"AutoscaledStorageIsUpToDate": { // if aws scales storage up, we should still consider it up to date, even if initial storage size was provided
			args: args{
				rds: &fake.MockRDSClient{
					MockDescribe: func(ctx context.Context, input *awsrds.DescribeDBInstancesInput, opts []func(*awsrds.Options)) (*awsrds.DescribeDBInstancesOutput, error) {
						return &awsrds.DescribeDBInstancesOutput{
							DBInstances: []awsrdstypes.DBInstance{
								{
									DBInstanceStatus:    aws.String(string(v1beta1.RDSInstanceStateAvailable)),
									MaxAllocatedStorage: aws.Int32(100),
									AllocatedStorage:    30,
								},
							},
						}, nil
					},
				},
				cr: instance(withMaxAllocatedStorage(100), withAllocatedStorage(20)),
			},
			want: want{
				cr: instance(
					withMaxAllocatedStorage(100),
					withAllocatedStorage(20),
					withStatusAllocatedStorage(30),
					withConditions(xpv1.Available()),
					withDBInstanceStatus(string(v1beta1.RDSInstanceStateAvailable))),
				result: managed.ExternalObservation{
					ResourceExists:    true,
					ResourceUpToDate:  true,
					ConnectionDetails: rds.GetConnectionDetails(v1beta1.RDSInstance{}),
				},
			},
		},
		"AWSBackupManagedBackupRetentionPeriodIsUpToDate": { // Ignore BackupRetentionPeriod if using AWS Backup
			args: args{
				rds: &fake.MockRDSClient{
					MockDescribe: func(ctx context.Context, input *awsrds.DescribeDBInstancesInput, opts []func(*awsrds.Options)) (*awsrds.DescribeDBInstancesOutput, error) {
						return &awsrds.DescribeDBInstancesOutput{
							DBInstances: []awsrdstypes.DBInstance{
								{
									DBInstanceStatus:          aws.String(string(v1beta1.RDSInstanceStateAvailable)),
									BackupRetentionPeriod:     10,
									AwsBackupRecoveryPointArn: aws.String(awsBackupRecoveryPointARN),
								},
							},
						}, nil
					},
				},
				cr: instance(withBackupRetentionPeriod(1), withStatusBackupRetentionPeriod(1), withStatusAWSBackupRecoveryPointARN(awsBackupRecoveryPointARN)),
			},
			want: want{
				cr: instance(
					withBackupRetentionPeriod(1),
					withStatusBackupRetentionPeriod(10),
					withStatusAWSBackupRecoveryPointARN(awsBackupRecoveryPointARN),
					withConditions(xpv1.Available()),
					withDBInstanceStatus(string(v1beta1.RDSInstanceStateAvailable))),
				result: managed.ExternalObservation{
					ResourceExists:    true,
					ResourceUpToDate:  true,
					ConnectionDetails: rds.GetConnectionDetails(v1beta1.RDSInstance{}),
				},
			},
		},
		"RDSManagedBackupRetentionPeriod": {
			args: args{
				rds: &fake.MockRDSClient{
					MockDescribe: func(ctx context.Context, input *awsrds.DescribeDBInstancesInput, opts []func(*awsrds.Options)) (*awsrds.DescribeDBInstancesOutput, error) {
						return &awsrds.DescribeDBInstancesOutput{
							DBInstances: []awsrdstypes.DBInstance{
								{
									DBInstanceStatus:      aws.String(string(v1beta1.RDSInstanceStateAvailable)),
									BackupRetentionPeriod: 10,
								},
							},
						}, nil
					},
				},
				cr: instance(withBackupRetentionPeriod(10), withStatusBackupRetentionPeriod(10)),
			},
			want: want{
				cr: instance(
					withBackupRetentionPeriod(10),
					withStatusBackupRetentionPeriod(10),
					withConditions(xpv1.Available()),
					withDBInstanceStatus(string(v1beta1.RDSInstanceStateAvailable))),
				result: managed.ExternalObservation{
					ResourceExists:    true,
					ResourceUpToDate:  true,
					ConnectionDetails: rds.GetConnectionDetails(v1beta1.RDSInstance{}),
				},
			},
		},
		"DeletingState": {
			args: args{
				rds: &fake.MockRDSClient{
					MockDescribe: func(ctx context.Context, input *awsrds.DescribeDBInstancesInput, opts []func(*awsrds.Options)) (*awsrds.DescribeDBInstancesOutput, error) {
						return &awsrds.DescribeDBInstancesOutput{
							DBInstances: []awsrdstypes.DBInstance{
								{
									DBInstanceStatus: aws.String(string(v1beta1.RDSInstanceStateDeleting)),
								},
							},
						}, nil
					},
				},
				cr: instance(),
			},
			want: want{
				cr: instance(
					withConditions(xpv1.Deleting()),
					withDBInstanceStatus(string(v1beta1.RDSInstanceStateDeleting))),
				result: managed.ExternalObservation{
					ResourceExists:    true,
					ResourceUpToDate:  true,
					ConnectionDetails: rds.GetConnectionDetails(v1beta1.RDSInstance{}),
				},
			},
		},
		"FailedState": {
			args: args{
				rds: &fake.MockRDSClient{
					MockDescribe: func(ctx context.Context, input *awsrds.DescribeDBInstancesInput, opts []func(*awsrds.Options)) (*awsrds.DescribeDBInstancesOutput, error) {
						return &awsrds.DescribeDBInstancesOutput{
							DBInstances: []awsrdstypes.DBInstance{
								{
									DBInstanceStatus: aws.String(string(v1beta1.RDSInstanceStateFailed)),
								},
							},
						}, nil
					},
				},
				cr: instance(),
			},
			want: want{
				cr: instance(
					withConditions(xpv1.Unavailable()),
					withDBInstanceStatus(string(v1beta1.RDSInstanceStateFailed))),
				result: managed.ExternalObservation{
					ResourceExists:    true,
					ResourceUpToDate:  true,
					ConnectionDetails: rds.GetConnectionDetails(v1beta1.RDSInstance{}),
				},
			},
		},
		"FailedDescribeRequest": {
			args: args{
				rds: &fake.MockRDSClient{
					MockDescribe: func(ctx context.Context, input *awsrds.DescribeDBInstancesInput, opts []func(*awsrds.Options)) (*awsrds.DescribeDBInstancesOutput, error) {
						return nil, errBoom
					},
				},
				cr: instance(),
			},
			want: want{
				cr:  instance(),
				err: errorutils.Wrap(errBoom, errDescribeFailed),
			},
		},
		"NotFound": {
			args: args{
				rds: &fake.MockRDSClient{
					MockDescribe: func(ctx context.Context, input *awsrds.DescribeDBInstancesInput, opts []func(*awsrds.Options)) (*awsrds.DescribeDBInstancesOutput, error) {
						return nil, &awsrdstypes.DBInstanceNotFoundFault{}
					},
				},
				cr: instance(),
			},
			want: want{
				cr: instance(),
			},
		},
		"LateInitSuccess": {
			args: args{
				rds: &fake.MockRDSClient{
					MockDescribe: func(ctx context.Context, input *awsrds.DescribeDBInstancesInput, opts []func(*awsrds.Options)) (*awsrds.DescribeDBInstancesOutput, error) {
						return &awsrds.DescribeDBInstancesOutput{
							DBInstances: []awsrdstypes.DBInstance{
								{
									EngineVersion:    aws.String(engineVersion),
									DBInstanceStatus: aws.String(string(v1beta1.RDSInstanceStateCreating)),
								},
							},
						}, nil
					},
				},
				cr: instance(),
			},
			want: want{
				cr: instance(
					withEngineVersion(&engineVersion),
					withDBInstanceStatus(string(v1beta1.RDSInstanceStateCreating)),
					withConditions(xpv1.Creating()),
				),
				result: managed.ExternalObservation{
					ResourceExists:          true,
					ResourceUpToDate:        true,
					ResourceLateInitialized: true,
					ConnectionDetails:       rds.GetConnectionDetails(v1beta1.RDSInstance{}),
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.rds}
			o, err := e.Observe(context.Background(), tc.args.cr)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.cr, tc.args.cr, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.result, o); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestCreate(t *testing.T) {
	type want struct {
		cr     *v1beta1.RDSInstance
		result managed.ExternalCreation
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"SuccessfulCreate": {
			args: args{
				rds: &fake.MockRDSClient{
					MockCreate: func(ctx context.Context, input *awsrds.CreateDBInstanceInput, opts []func(*awsrds.Options)) (*awsrds.CreateDBInstanceOutput, error) {
						return &awsrds.CreateDBInstanceOutput{}, nil
					},
				},
				cr: instance(withMasterUsername(&masterUsername)),
			},
			want: want{
				cr: instance(
					withMasterUsername(&masterUsername),
					withConditions(xpv1.Creating())),
				result: managed.ExternalCreation{
					ConnectionDetails: managed.ConnectionDetails{
						xpv1.ResourceCredentialsSecretUserKey:     []byte(masterUsername),
						xpv1.ResourceCredentialsSecretPasswordKey: []byte(replaceMe),
					},
				},
			},
		},
		"SuccessfulS3Restore": {
			args: args{
				rds: &fake.MockRDSClient{
					MockS3Restore: func(ctx context.Context, input *awsrds.RestoreDBInstanceFromS3Input, opts []func(*awsrds.Options)) (*awsrds.RestoreDBInstanceFromS3Output, error) {
						return &awsrds.RestoreDBInstanceFromS3Output{}, nil
					},
				},
				cr: instance(
					withMasterUsername(&masterUsername),
					withBackupConfiguration(&s3Backup)),
			},
			want: want{
				cr: instance(
					withMasterUsername(&masterUsername),
					withBackupConfiguration(&s3Backup),
					withConditions(xpv1.Creating())),
				result: managed.ExternalCreation{
					ConnectionDetails: managed.ConnectionDetails{
						xpv1.ResourceCredentialsSecretUserKey:     []byte(masterUsername),
						xpv1.ResourceCredentialsSecretPasswordKey: []byte(replaceMe),
					},
				},
			},
		},
		"SuccessfulSnapshotRestore": {
			args: args{
				rds: &fake.MockRDSClient{
					MockSnapshotRestore: func(ctx context.Context, input *awsrds.RestoreDBInstanceFromDBSnapshotInput, opts []func(*awsrds.Options)) (*awsrds.RestoreDBInstanceFromDBSnapshotOutput, error) {
						return &awsrds.RestoreDBInstanceFromDBSnapshotOutput{}, nil
					},
				},
				cr: instance(
					withMasterUsername(&masterUsername),
					withBackupConfiguration(&snapshotBackup)),
			},
			want: want{
				cr: instance(
					withMasterUsername(&masterUsername),
					withBackupConfiguration(&snapshotBackup),
					withConditions(xpv1.Creating())),
				result: managed.ExternalCreation{
					ConnectionDetails: managed.ConnectionDetails{
						xpv1.ResourceCredentialsSecretUserKey:     []byte(masterUsername),
						xpv1.ResourceCredentialsSecretPasswordKey: []byte(replaceMe),
					},
				},
			},
		},
		"SuccessfulPointInTimeLatestRestorable": {
			args: args{
				rds: &fake.MockRDSClient{
					MockPointInTimeRestore: func(ctx context.Context, input *awsrds.RestoreDBInstanceToPointInTimeInput, opts []func(*awsrds.Options)) (*awsrds.RestoreDBInstanceToPointInTimeOutput, error) {
						return &awsrds.RestoreDBInstanceToPointInTimeOutput{}, nil
					},
				},
				cr: instance(
					withMasterUsername(&masterUsername),
					withBackupConfiguration(&pointInTimeLatestRestorableBackup)),
			},
			want: want{
				cr: instance(
					withMasterUsername(&masterUsername),
					withBackupConfiguration(&pointInTimeLatestRestorableBackup),
					withConditions(xpv1.Creating())),
				result: managed.ExternalCreation{
					ConnectionDetails: managed.ConnectionDetails{
						xpv1.ResourceCredentialsSecretUserKey:     []byte(masterUsername),
						xpv1.ResourceCredentialsSecretPasswordKey: []byte(replaceMe),
					},
				},
			},
		},
		"SuccessfulPointInTimeRestore": {
			args: args{
				rds: &fake.MockRDSClient{
					MockPointInTimeRestore: func(ctx context.Context, input *awsrds.RestoreDBInstanceToPointInTimeInput, opts []func(*awsrds.Options)) (*awsrds.RestoreDBInstanceToPointInTimeOutput, error) {
						return &awsrds.RestoreDBInstanceToPointInTimeOutput{}, nil
					},
				},
				cr: instance(
					withMasterUsername(&masterUsername),
					withBackupConfiguration(&pointInTimeBackup)),
			},
			want: want{
				cr: instance(
					withMasterUsername(&masterUsername),
					withBackupConfiguration(&pointInTimeBackup),
					withConditions(xpv1.Creating())),
				result: managed.ExternalCreation{
					ConnectionDetails: managed.ConnectionDetails{
						xpv1.ResourceCredentialsSecretUserKey:     []byte(masterUsername),
						xpv1.ResourceCredentialsSecretPasswordKey: []byte(replaceMe),
					},
				},
			},
		},
		"SuccessfulNoNeedForCreate": {
			args: args{
				cr: instance(withDBInstanceStatus(v1beta1.RDSInstanceStateCreating)),
			},
			want: want{
				cr: instance(
					withDBInstanceStatus(v1beta1.RDSInstanceStateCreating),
					withConditions(xpv1.Creating())),
			},
		},
		"SuccessfulNoUsername": {
			args: args{
				rds: &fake.MockRDSClient{
					MockCreate: func(ctx context.Context, input *awsrds.CreateDBInstanceInput, opts []func(*awsrds.Options)) (*awsrds.CreateDBInstanceOutput, error) {
						return &awsrds.CreateDBInstanceOutput{}, nil
					},
				},
				cr: instance(withMasterUsername(nil)),
			},
			want: want{
				cr: instance(
					withMasterUsername(nil),
					withConditions(xpv1.Creating())),
				result: managed.ExternalCreation{
					ConnectionDetails: managed.ConnectionDetails{
						xpv1.ResourceCredentialsSecretPasswordKey: []byte(replaceMe),
					},
				},
			},
		},
		"SuccessfulWithSecret": {
			args: args{
				rds: &fake.MockRDSClient{
					MockCreate: func(ctx context.Context, input *awsrds.CreateDBInstanceInput, opts []func(*awsrds.Options)) (*awsrds.CreateDBInstanceOutput, error) {
						return &awsrds.CreateDBInstanceOutput{}, nil
					},
				},
				kube: &test.MockClient{
					MockGet: func(_ context.Context, key types.NamespacedName, obj client.Object) error {
						secret := corev1.Secret{
							Data: map[string][]byte{},
						}
						secret.Data[secretKey] = []byte(credData)
						secret.DeepCopyInto(obj.(*corev1.Secret))
						return nil
					},
				},
				cr: instance(withMasterUsername(&masterUsername), withPasswordSecretRef(xpv1.SecretKeySelector{Key: secretKey})),
			},
			want: want{
				cr: instance(
					withMasterUsername(&masterUsername),
					withPasswordSecretRef(xpv1.SecretKeySelector{Key: secretKey}),
					withConditions(xpv1.Creating())),
				result: managed.ExternalCreation{
					ConnectionDetails: managed.ConnectionDetails{
						xpv1.ResourceCredentialsSecretPasswordKey: []byte(credData),
						xpv1.ResourceCredentialsSecretUserKey:     []byte(masterUsername),
					},
				},
			},
		},
		"FailedWhileGettingSecret": {
			args: args{
				kube: &test.MockClient{
					MockGet: test.NewMockGetFn(errBoom),
				},
				cr: instance(withMasterUsername(&masterUsername), withPasswordSecretRef(xpv1.SecretKeySelector{})),
			},
			want: want{
				cr: instance(
					withMasterUsername(&masterUsername),
					withPasswordSecretRef(xpv1.SecretKeySelector{}),
					withConditions(xpv1.Creating())),
				err: errorutils.Wrap(errBoom, errGetPasswordSecretFailed),
			},
		},
		"FailedRequest": {
			args: args{
				rds: &fake.MockRDSClient{
					MockCreate: func(ctx context.Context, input *awsrds.CreateDBInstanceInput, opts []func(*awsrds.Options)) (*awsrds.CreateDBInstanceOutput, error) {
						return nil, errBoom
					},
				},
				cr: instance(),
			},
			want: want{
				cr:  instance(withConditions(xpv1.Creating())),
				err: errorutils.Wrap(errBoom, errCreateFailed),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.rds}
			o, err := e.Create(context.Background(), tc.args.cr)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.cr, tc.args.cr, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if string(tc.want.result.ConnectionDetails[xpv1.ResourceCredentialsSecretPasswordKey]) == replaceMe {
				tc.want.result.ConnectionDetails[xpv1.ResourceCredentialsSecretPasswordKey] =
					o.ConnectionDetails[xpv1.ResourceCredentialsSecretPasswordKey]
			}
			if diff := cmp.Diff(tc.want.result, o); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestUpdate(t *testing.T) {
	type want struct {
		cr     *v1beta1.RDSInstance
		result managed.ExternalUpdate
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"Successful": {
			args: args{
				rds: &fake.MockRDSClient{
					MockModify: func(ctx context.Context, input *awsrds.ModifyDBInstanceInput, opts []func(*awsrds.Options)) (*awsrds.ModifyDBInstanceOutput, error) {
						return &awsrds.ModifyDBInstanceOutput{}, nil
					},
					MockDescribe: func(ctx context.Context, input *awsrds.DescribeDBInstancesInput, opts []func(*awsrds.Options)) (*awsrds.DescribeDBInstancesOutput, error) {
						return &awsrds.DescribeDBInstancesOutput{
							DBInstances: []awsrdstypes.DBInstance{{}},
						}, nil
					},
					MockAddTags: func(ctx context.Context, input *awsrds.AddTagsToResourceInput, opts []func(*awsrds.Options)) (*awsrds.AddTagsToResourceOutput, error) {
						return &awsrds.AddTagsToResourceOutput{}, nil
					},
				},
				cr: instance(withTags(map[string]string{"foo": "bar"})),
			},
			want: want{
				cr: instance(withTags(map[string]string{"foo": "bar"})),
			},
		},
		"AutoscaleExcludeStorage": {
			args: args{
				rds: &fake.MockRDSClient{
					MockModify: func(ctx context.Context, input *awsrds.ModifyDBInstanceInput, opts []func(*awsrds.Options)) (*awsrds.ModifyDBInstanceOutput, error) {
						if input.AllocatedStorage != nil {
							return &awsrds.ModifyDBInstanceOutput{}, errors.New("AllocatedStorage must not be set when on a modify request when AWS has autoscaled the storage")
						}
						return &awsrds.ModifyDBInstanceOutput{}, nil
					},
					MockDescribe: func(ctx context.Context, input *awsrds.DescribeDBInstancesInput, opts []func(*awsrds.Options)) (*awsrds.DescribeDBInstancesOutput, error) {
						return &awsrds.DescribeDBInstancesOutput{
							DBInstances: []awsrdstypes.DBInstance{{
								MaxAllocatedStorage: aws.Int32(100),
								AllocatedStorage:    30,
							}},
						}, nil
					},
					MockAddTags: func(ctx context.Context, input *awsrds.AddTagsToResourceInput, opts []func(*awsrds.Options)) (*awsrds.AddTagsToResourceOutput, error) {
						return &awsrds.AddTagsToResourceOutput{}, nil
					},
				},
				cr: instance(withMaxAllocatedStorage(100), withAllocatedStorage(20)),
			},
			want: want{
				cr: instance(withMaxAllocatedStorage(100), withAllocatedStorage(20)),
			},
		},
		"AWSManagedBackupRetentionTargetIgnore": {
			args: args{
				rds: &fake.MockRDSClient{
					MockModify: func(ctx context.Context, input *awsrds.ModifyDBInstanceInput, opts []func(*awsrds.Options)) (*awsrds.ModifyDBInstanceOutput, error) {
						if input.BackupRetentionPeriod != nil {
							return &awsrds.ModifyDBInstanceOutput{}, errors.New("BackupRetentionPeriod must not be set when AWS Backup is used")
						}
						if input.PreferredBackupWindow != nil {
							return &awsrds.ModifyDBInstanceOutput{}, errors.New("PreferredBackupWindow must not be set when AWS Backup is used")
						}
						return &awsrds.ModifyDBInstanceOutput{}, nil
					},
					MockDescribe: func(ctx context.Context, input *awsrds.DescribeDBInstancesInput, opts []func(*awsrds.Options)) (*awsrds.DescribeDBInstancesOutput, error) {
						return &awsrds.DescribeDBInstancesOutput{
							DBInstances: []awsrdstypes.DBInstance{{
								BackupRetentionPeriod:     7,
								PreferredBackupWindow:     &backupWindow,
								AwsBackupRecoveryPointArn: aws.String(awsBackupRecoveryPointARN),
							}},
						}, nil
					},
					MockAddTags: func(ctx context.Context, input *awsrds.AddTagsToResourceInput, opts []func(*awsrds.Options)) (*awsrds.AddTagsToResourceOutput, error) {
						return &awsrds.AddTagsToResourceOutput{}, nil
					},
				},
				cr: instance(withBackupRetentionPeriod(0), withStatusBackupRetentionPeriod(7), withPreferredBackupWindow("x")),
			},
			want: want{
				cr: instance(withBackupRetentionPeriod(0), withStatusBackupRetentionPeriod(7), withPreferredBackupWindow("x")),
			},
		},
		"AlreadyModifying": {
			args: args{
				cr: instance(withDBInstanceStatus(v1beta1.RDSInstanceStateModifying)),
			},
			want: want{
				cr: instance(withDBInstanceStatus(v1beta1.RDSInstanceStateModifying)),
			},
		},
		"FailedDescribe": {
			args: args{
				rds: &fake.MockRDSClient{
					MockDescribe: func(ctx context.Context, input *awsrds.DescribeDBInstancesInput, opts []func(*awsrds.Options)) (*awsrds.DescribeDBInstancesOutput, error) {
						return nil, errBoom
					},
				},
				cr: instance(),
			},
			want: want{
				cr:  instance(),
				err: errorutils.Wrap(errBoom, errDescribeFailed),
			},
		},
		"FailedModify": {
			args: args{
				rds: &fake.MockRDSClient{
					MockModify: func(ctx context.Context, input *awsrds.ModifyDBInstanceInput, opts []func(*awsrds.Options)) (*awsrds.ModifyDBInstanceOutput, error) {
						return nil, errBoom
					},
					MockDescribe: func(ctx context.Context, input *awsrds.DescribeDBInstancesInput, opts []func(*awsrds.Options)) (*awsrds.DescribeDBInstancesOutput, error) {
						return &awsrds.DescribeDBInstancesOutput{
							DBInstances: []awsrdstypes.DBInstance{{}},
						}, nil
					},
				},
				cr: instance(),
			},
			want: want{
				cr:  instance(),
				err: errorutils.Wrap(errBoom, errModifyFailed),
			},
		},
		"FailedAddTags": {
			args: args{
				rds: &fake.MockRDSClient{
					MockModify: func(ctx context.Context, input *awsrds.ModifyDBInstanceInput, opts []func(*awsrds.Options)) (*awsrds.ModifyDBInstanceOutput, error) {
						return &awsrds.ModifyDBInstanceOutput{}, nil
					},
					MockDescribe: func(ctx context.Context, input *awsrds.DescribeDBInstancesInput, opts []func(*awsrds.Options)) (*awsrds.DescribeDBInstancesOutput, error) {
						return &awsrds.DescribeDBInstancesOutput{
							DBInstances: []awsrdstypes.DBInstance{{}},
						}, nil
					},
					MockAddTags: func(ctx context.Context, input *awsrds.AddTagsToResourceInput, opts []func(*awsrds.Options)) (*awsrds.AddTagsToResourceOutput, error) {
						return nil, errBoom
					},
				},
				cr:    instance(withTags(map[string]string{"foo": "bar"})),
				cache: cache(map[string]string{"foo": "bar"}, map[string]string{}),
			},
			want: want{
				cr:  instance(withTags(map[string]string{"foo": "bar"})),
				err: errorutils.Wrap(errBoom, errAddTagsFailed),
			},
		},
		"FailedRemoveTags": {
			args: args{
				rds: &fake.MockRDSClient{
					MockModify: func(ctx context.Context, input *awsrds.ModifyDBInstanceInput, opts []func(*awsrds.Options)) (*awsrds.ModifyDBInstanceOutput, error) {
						return &awsrds.ModifyDBInstanceOutput{}, nil
					},
					MockDescribe: func(ctx context.Context, input *awsrds.DescribeDBInstancesInput, opts []func(*awsrds.Options)) (*awsrds.DescribeDBInstancesOutput, error) {
						return &awsrds.DescribeDBInstancesOutput{
							DBInstances: []awsrdstypes.DBInstance{{}},
						}, nil
					},
					MockRemoveTags: func(ctx context.Context, input *awsrds.RemoveTagsFromResourceInput, opts []func(*awsrds.Options)) (*awsrds.RemoveTagsFromResourceOutput, error) {
						return nil, errBoom
					},
				},
				cr:    instance(withTags(map[string]string{"foo": "bar"})),
				cache: cache(map[string]string{}, map[string]string{"foo": "bar"}),
			},
			want: want{
				cr:  instance(withTags(map[string]string{"foo": "bar"})),
				err: errorutils.Wrap(errBoom, errRemoveTagsFailed),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.rds, cache: tc.cache}
			u, err := e.Update(context.Background(), tc.args.cr)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.cr, tc.args.cr, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.result, u); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	type want struct {
		cr  *v1beta1.RDSInstance
		err error
	}

	cases := map[string]struct {
		args
		want
	}{
		"Successful": {
			args: args{
				rds: &fake.MockRDSClient{
					MockDelete: func(ctx context.Context, input *awsrds.DeleteDBInstanceInput, opts []func(*awsrds.Options)) (*awsrds.DeleteDBInstanceOutput, error) {
						return &awsrds.DeleteDBInstanceOutput{}, nil
					},
					MockModify: func(ctx context.Context, input *awsrds.ModifyDBInstanceInput, opts []func(*awsrds.Options)) (*awsrds.ModifyDBInstanceOutput, error) {
						return &awsrds.ModifyDBInstanceOutput{}, nil
					},
					MockDescribe: func(ctx context.Context, input *awsrds.DescribeDBInstancesInput, opts []func(*awsrds.Options)) (*awsrds.DescribeDBInstancesOutput, error) {
						return &awsrds.DescribeDBInstancesOutput{
							DBInstances: []awsrdstypes.DBInstance{{}},
						}, nil
					},
				},
				cr: instance(),
			},
			want: want{
				cr: instance(withConditions(xpv1.Deleting())),
			},
		},
		"SuccessfulWithOptions": {
			args: args{
				rds: &fake.MockRDSClient{
					MockDelete: func(ctx context.Context, input *awsrds.DeleteDBInstanceInput, opts []func(*awsrds.Options)) (*awsrds.DeleteDBInstanceOutput, error) {
						if input.DeleteAutomatedBackups == nil || !*input.DeleteAutomatedBackups {
							return nil, errors.New("expected DeletedAutomatedBackups to be set")
						}
						return &awsrds.DeleteDBInstanceOutput{}, nil
					},
					MockModify: func(ctx context.Context, input *awsrds.ModifyDBInstanceInput, opts []func(*awsrds.Options)) (*awsrds.ModifyDBInstanceOutput, error) {
						return &awsrds.ModifyDBInstanceOutput{}, nil
					},
					MockDescribe: func(ctx context.Context, input *awsrds.DescribeDBInstancesInput, opts []func(*awsrds.Options)) (*awsrds.DescribeDBInstancesOutput, error) {
						return &awsrds.DescribeDBInstancesOutput{
							DBInstances: []awsrdstypes.DBInstance{{}},
						}, nil
					},
				},
				cr: instance(withDeleteAutomatedBackups(true)),
			},
			want: want{
				cr: instance(withDeleteAutomatedBackups(true), withConditions(xpv1.Deleting())),
			},
		},
		"AlreadyDeleting": {
			args: args{
				cr: instance(withDBInstanceStatus(v1beta1.RDSInstanceStateDeleting)),
			},
			want: want{
				cr: instance(withDBInstanceStatus(v1beta1.RDSInstanceStateDeleting),
					withConditions(xpv1.Deleting())),
			},
		},
		"AlreadyDeleted": {
			args: args{
				rds: &fake.MockRDSClient{
					MockDescribe: func(ctx context.Context, input *awsrds.DescribeDBInstancesInput, opts []func(*awsrds.Options)) (*awsrds.DescribeDBInstancesOutput, error) {
						return nil, &awsrdstypes.DBInstanceNotFoundFault{}
					},
				},
				cr: instance(),
			},
			want: want{
				cr: instance(withConditions(xpv1.Deleting())),
			},
		},
		"Failed": {
			args: args{
				rds: &fake.MockRDSClient{
					MockDelete: func(ctx context.Context, input *awsrds.DeleteDBInstanceInput, opts []func(*awsrds.Options)) (*awsrds.DeleteDBInstanceOutput, error) {
						return nil, errBoom
					},
					MockModify: func(ctx context.Context, input *awsrds.ModifyDBInstanceInput, opts []func(*awsrds.Options)) (*awsrds.ModifyDBInstanceOutput, error) {
						return &awsrds.ModifyDBInstanceOutput{}, nil
					},
					MockDescribe: func(ctx context.Context, input *awsrds.DescribeDBInstancesInput, opts []func(*awsrds.Options)) (*awsrds.DescribeDBInstancesOutput, error) {
						return &awsrds.DescribeDBInstancesOutput{
							DBInstances: []awsrdstypes.DBInstance{{}},
						}, nil
					},
				},
				cr: instance(),
			},
			want: want{
				cr:  instance(withConditions(xpv1.Deleting())),
				err: errorutils.Wrap(errBoom, errDeleteFailed),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.rds}
			err := e.Delete(context.Background(), tc.args.cr)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.cr, tc.args.cr, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}
