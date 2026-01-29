package dbinstance

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	svcsdk "github.com/aws/aws-sdk-go/service/rds"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/rds/v1alpha1"
	rds "github.com/crossplane-contrib/provider-aws/pkg/clients/rds"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/rds/fake"
)

type objectFnCustom func(obj client.Object, diff bool) error

// NewMockGetFn returns a MockGetFn that returns the supplied error.
func newMockGetFnCustomWithDiff(diff bool, err error, ofn []objectFnCustom) test.MockGetFn {
	return func(_ context.Context, _ client.ObjectKey, obj client.Object) error {
		for _, fn := range ofn {
			if err := fn(obj, diff); err != nil {
				return err
			}
		}
		return err
	}
}
func mockGettingSecretData(obj client.Object, diff bool) error {
	secret, ok := obj.(*corev1.Secret)
	if !ok {
		return errors.New("the mock function only supports secret objects")
	}
	secret.Data = map[string][]byte{
		rds.PasswordCacheKey:    []byte("cachedPassword"),
		rds.RestoreFlagCacheKay: []byte(""),
	}
	if diff {
		secret.Data[xpv1.ResourceCredentialsSecretPasswordKey] = []byte("differentPassword")
	} else {
		secret.Data[xpv1.ResourceCredentialsSecretPasswordKey] = []byte("cachedPassword")

	}
	return nil
}

func TestCreate(t *testing.T) {
	type args struct {
		cr           *svcapitypes.DBInstance
		kube         client.Client
		awsRDSClient fake.MockRDSClient
	}

	type want struct {
		statusAtProvider *svcapitypes.CustomDBInstanceObservation
		err              error
	}

	cases := map[string]struct {
		args
		want
	}{
		"CreateReadReplica": {
			args: args{
				cr: &svcapitypes.DBInstance{
					Spec: svcapitypes.DBInstanceSpec{
						ForProvider: svcapitypes.DBInstanceParameters{
							CustomDBInstanceParameters: svcapitypes.CustomDBInstanceParameters{
								SourceDBInstanceID: aws.String("source-db-instance-id"),
							},
						},
					},
				},
				kube: test.NewMockClient(),
				awsRDSClient: fake.MockRDSClient{
					MockCreateDBInstanceReadReplicaWithContext: func(ctx context.Context, input *svcsdk.CreateDBInstanceReadReplicaInput, optFns ...request.Option) (*svcsdk.CreateDBInstanceReadReplicaOutput, error) {
						return &svcsdk.CreateDBInstanceReadReplicaOutput{}, nil
					},
					MockCreateDBInstanceWithContext: func(ctx context.Context, input *svcsdk.CreateDBInstanceInput, optFns ...request.Option) (*svcsdk.CreateDBInstanceOutput, error) {
						return &svcsdk.CreateDBInstanceOutput{}, nil
					},
				},
			},
			want: want{
				statusAtProvider: &svcapitypes.CustomDBInstanceObservation{
					DatabaseRole: aws.String(databaseRoleReadReplica),
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			cr := tc.args.cr
			ce := newCustomExternal(tc.kube, &tc.awsRDSClient)
			_, err := ce.Create(context.TODO(), cr)

			if diff := cmp.Diff(tc.want.err, err, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got error: \n%s", diff)
			}
			if diff := cmp.Diff(tc.want.statusAtProvider.DatabaseRole, cr.Status.AtProvider.DatabaseRole); diff != "" {
				t.Errorf("r: -want, +got: \n%s", diff)
			}
		})
	}
}

func TestIsUpToDate(t *testing.T) {
	type args struct {
		cr   *svcapitypes.DBInstance
		out  *svcsdk.DescribeDBInstancesOutput
		kube client.Client
	}

	type want struct {
		upToDate bool
		err      error
	}

	cases := map[string]struct {
		args
		want
	}{

		"PreferredBackupWindowNotUpToDate": {
			args: args{
				cr: &svcapitypes.DBInstance{
					Spec: svcapitypes.DBInstanceSpec{
						ForProvider: svcapitypes.DBInstanceParameters{
							PreferredBackupWindow: aws.String("01:00-02:00"),
						},
					},
				},
				out: &svcsdk.DescribeDBInstancesOutput{
					DBInstances: []*svcsdk.DBInstance{
						{
							PreferredBackupWindow: aws.String("02:00-03:00"),
						},
					},
				},
				kube: test.NewMockClient(),
			},
			want: want{
				upToDate: false,
				err:      nil,
			},
		},
		"PreferredBackupWindowAndBackupRetentionPeriodIgnoredDueToAWSBackup": {
			args: args{
				cr: &svcapitypes.DBInstance{
					Spec: svcapitypes.DBInstanceSpec{
						ForProvider: svcapitypes.DBInstanceParameters{
							BackupRetentionPeriod: aws.Int64(1),
							PreferredBackupWindow: aws.String("01:00-02:00"),
						},
					},
				},
				out: &svcsdk.DescribeDBInstancesOutput{
					DBInstances: []*svcsdk.DBInstance{
						{
							AwsBackupRecoveryPointArn: aws.String("arn:aws:backup:eu-central-1:123456789012:recovery-point:continuous:db-random-string-hash"),
							BackupRetentionPeriod:     aws.Int64(7),
							PreferredBackupWindow:     aws.String("02:00-03:00"),
						},
					},
				},
				kube: test.NewMockClient(),
			},
			want: want{
				upToDate: true,
				err:      nil,
			},
		},
		"UpToDatePendingModifiedValue": {
			args: args{
				cr: &svcapitypes.DBInstance{
					Spec: svcapitypes.DBInstanceSpec{
						ForProvider: svcapitypes.DBInstanceParameters{
							DBInstanceClass: aws.String("db.t4.small"),
						},
					},
				},
				out: &svcsdk.DescribeDBInstancesOutput{
					DBInstances: []*svcsdk.DBInstance{
						{
							DBInstanceClass: aws.String("db.t4.micro"),
							PendingModifiedValues: &svcsdk.PendingModifiedValues{
								DBInstanceClass: aws.String("db.t4.small"),
							},
						},
					},
				},
				kube: test.NewMockClient(),
			},
			want: want{
				upToDate: true,
				err:      nil,
			},
		},
		"IsNoTUpToDatePendingModifiedValue": {
			args: args{
				cr: &svcapitypes.DBInstance{
					Spec: svcapitypes.DBInstanceSpec{
						ForProvider: svcapitypes.DBInstanceParameters{
							DBInstanceClass: aws.String("db.t4.medium"),
						},
					},
				},
				out: &svcsdk.DescribeDBInstancesOutput{
					DBInstances: []*svcsdk.DBInstance{
						{
							DBInstanceClass: aws.String("db.t4.micro"),
							PendingModifiedValues: &svcsdk.PendingModifiedValues{
								DBInstanceClass: aws.String("db.t4.small"),
							},
						},
					},
				},
				kube: test.NewMockClient(),
			},
			want: want{
				upToDate: false,
				err:      nil,
			},
		},
		"IsNoTUpToDatePendingModifiedValueApplyImmediately": { // Instance class is already scheduled to be updated, but we want to update it immediately
			args: args{
				cr: &svcapitypes.DBInstance{
					Spec: svcapitypes.DBInstanceSpec{
						ForProvider: svcapitypes.DBInstanceParameters{
							DBInstanceClass: aws.String("db.t4.medium"),
							CustomDBInstanceParameters: svcapitypes.CustomDBInstanceParameters{
								ApplyImmediately: aws.Bool(true),
							},
						},
					},
				},
				out: &svcsdk.DescribeDBInstancesOutput{
					DBInstances: []*svcsdk.DBInstance{
						{
							DBInstanceClass: aws.String("db.t4.micro"),
							PendingModifiedValues: &svcsdk.PendingModifiedValues{
								DBInstanceClass: aws.String("db.t4.medium"),
							},
						},
					},
				},
				kube: test.NewMockClient(),
			},
			want: want{
				upToDate: false,
				err:      nil,
			},
		},
		"UpToDatePendingModifiedValueEngineVersion": {
			args: args{
				cr: &svcapitypes.DBInstance{
					Spec: svcapitypes.DBInstanceSpec{
						ForProvider: svcapitypes.DBInstanceParameters{
							Engine: aws.String("mariadb"),
							CustomDBInstanceParameters: svcapitypes.CustomDBInstanceParameters{
								EngineVersion: aws.String("11.8.5"),
							},
						},
					},
				},
				out: &svcsdk.DescribeDBInstancesOutput{
					DBInstances: []*svcsdk.DBInstance{
						{
							Engine:        aws.String("mariadb"),
							EngineVersion: aws.String("11.8.3"),
							PendingModifiedValues: &svcsdk.PendingModifiedValues{
								EngineVersion: aws.String("11.8.5"),
							},
						},
					},
				},
				kube: test.NewMockClient(),
			},
			want: want{
				upToDate: true,
				err:      nil,
			},
		},
		"IsNoTUpToDatePendingModifiedValueEngineVersionRevert": {
			args: args{
				cr: &svcapitypes.DBInstance{
					Spec: svcapitypes.DBInstanceSpec{
						ForProvider: svcapitypes.DBInstanceParameters{
							Engine: aws.String("mariadb"),
							CustomDBInstanceParameters: svcapitypes.CustomDBInstanceParameters{
								EngineVersion: aws.String("11.8.3"),
							},
						},
					},
				},
				out: &svcsdk.DescribeDBInstancesOutput{
					DBInstances: []*svcsdk.DBInstance{
						{
							Engine:        aws.String("mariadb"),
							EngineVersion: aws.String("11.8.3"),
							PendingModifiedValues: &svcsdk.PendingModifiedValues{
								EngineVersion: aws.String("11.8.5"),
							},
						},
					},
				},
				kube: test.NewMockClient(),
			},
			want: want{
				upToDate: false,
				err:      nil,
			},
		},
		"IsNoTUpToDatePendingModifiedValueEngineVersionApplyImmediately": { // Engine version is already scheduled to be updated, but we want to update it immediately
			args: args{
				cr: &svcapitypes.DBInstance{
					Spec: svcapitypes.DBInstanceSpec{
						ForProvider: svcapitypes.DBInstanceParameters{
							Engine: aws.String("mariadb"),
							CustomDBInstanceParameters: svcapitypes.CustomDBInstanceParameters{
								ApplyImmediately: aws.Bool(true),
								EngineVersion:    aws.String("11.8.5"),
							},
						},
					},
				},
				out: &svcsdk.DescribeDBInstancesOutput{
					DBInstances: []*svcsdk.DBInstance{
						{
							Engine:        aws.String("mariadb"),
							EngineVersion: aws.String("11.8.3"),
							PendingModifiedValues: &svcsdk.PendingModifiedValues{
								EngineVersion: aws.String("11.8.5"),
							},
						},
					},
				},
				kube: test.NewMockClient(),
			},
			want: want{
				upToDate: false,
				err:      nil,
			},
		},
		"UpToDateMasterPassword": {
			args: args{
				cr: &svcapitypes.DBInstance{
					Spec: svcapitypes.DBInstanceSpec{
						ForProvider: svcapitypes.DBInstanceParameters{
							Engine: aws.String("mariadb"),
							CustomDBInstanceParameters: svcapitypes.CustomDBInstanceParameters{
								EngineVersion: aws.String("10.5.12"),
								MasterUserPasswordSecretRef: &xpv1.SecretKeySelector{
									SecretReference: xpv1.SecretReference{
										Name:      "masterUserPassword",
										Namespace: "default",
									},
									Key: "password",
								},
							},
						},
					},
				},
				out: &svcsdk.DescribeDBInstancesOutput{
					DBInstances: []*svcsdk.DBInstance{
						{
							Engine:        aws.String("mariadb"),
							EngineVersion: aws.String("10.5.12"),
						},
					},
				},
				kube: &test.MockClient{
					// This mock returns the same password for both secrets(masterUserPasswordSecretRef and cached secret)
					MockGet: newMockGetFnCustomWithDiff(false, nil, []objectFnCustom{mockGettingSecretData}),
				},
			},
			want: want{
				upToDate: true,
				err:      nil,
			},
		},
		"ChangedMasterPassword": {
			args: args{
				cr: &svcapitypes.DBInstance{
					Spec: svcapitypes.DBInstanceSpec{
						ForProvider: svcapitypes.DBInstanceParameters{
							Engine: aws.String("mariadb"),
							CustomDBInstanceParameters: svcapitypes.CustomDBInstanceParameters{
								EngineVersion: aws.String("10.5.12"),
								MasterUserPasswordSecretRef: &xpv1.SecretKeySelector{
									SecretReference: xpv1.SecretReference{
										Name:      "masterUserPassword",
										Namespace: "default",
									},
									Key: "password",
								},
							},
						},
					},
				},
				out: &svcsdk.DescribeDBInstancesOutput{
					DBInstances: []*svcsdk.DBInstance{
						{
							Engine:        aws.String("mariadb"),
							EngineVersion: aws.String("10.5.12"),
						},
					},
				},
				kube: &test.MockClient{
					// This mock returns different passwords from masterUserPasswordSecretRef and cached secret
					MockGet: newMockGetFnCustomWithDiff(true, nil, []objectFnCustom{mockGettingSecretData}),
				},
			},
			want: want{
				upToDate: false,
				err:      nil,
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			cr := tc.args.cr
			ce := newCustomExternal(tc.kube, nil)
			upToDate, diffMsg, err := ce.isUpToDate(context.TODO(), cr, tc.args.out)

			if diff := cmp.Diff(tc.want.err, err); diff != "" {
				t.Errorf("r: -want, +got error: \n%s", diff)
			}
			if diff := cmp.Diff(tc.want.upToDate, upToDate); diff != "" {
				t.Errorf("r: -want, +got: \n%s\ndiff message: %s", diff, diffMsg)
			}
		})
	}
}

func TestPostObserve(t *testing.T) {
	type args struct {
		awsRDSClient fake.MockRDSClient
		cr           *svcapitypes.DBInstance
		out          *svcsdk.DescribeDBInstancesOutput
		kube         client.Client
	}

	type want struct {
		err              error
		statusAtProvider *svcapitypes.CustomDBInstanceObservation
	}

	cases := map[string]struct {
		args
		want
	}{
		"databaseRoleReplica": {
			args: args{
				cr: &svcapitypes.DBInstance{
					Status: svcapitypes.DBInstanceStatus{},
				},
				out: &svcsdk.DescribeDBInstancesOutput{
					DBInstances: []*svcsdk.DBInstance{
						{
							DeletionProtection:                   aws.Bool(true),
							ReadReplicaSourceDBClusterIdentifier: aws.String("source-db-instance-id"),
						},
					},
				},
				awsRDSClient: fake.MockRDSClient{
					MockDescribeDBClustersWithContext: func(ctx context.Context, input *svcsdk.DescribeDBClustersInput, optFns ...request.Option) (*svcsdk.DescribeDBClustersOutput, error) {
						return &svcsdk.DescribeDBClustersOutput{}, nil
					},
				},
				kube: &test.MockClient{
					MockGet: func(ctx context.Context, key client.ObjectKey, obj client.Object) error {
						return errors.New("not found")
					},
				},
			},
			want: want{
				err: nil,
				statusAtProvider: &svcapitypes.CustomDBInstanceObservation{
					DatabaseRole: aws.String(databaseRoleReadReplica),
				},
			},
		},
		"databaseRolePrimary": {
			args: args{
				cr: &svcapitypes.DBInstance{
					Status: svcapitypes.DBInstanceStatus{},
				},
				kube: test.NewMockClient(),
				out: &svcsdk.DescribeDBInstancesOutput{
					DBInstances: []*svcsdk.DBInstance{
						{
							DeletionProtection:               aws.Bool(true),
							ReadReplicaDBInstanceIdentifiers: []*string{aws.String("db-read-replica-id")},
						},
					},
				},
			},
			want: want{
				err: nil,
				statusAtProvider: &svcapitypes.CustomDBInstanceObservation{
					DatabaseRole: aws.String(databaseRolePrimary),
				},
			},
		},
		"databaseRoleClusterWriter": {
			args: args{
				cr: &svcapitypes.DBInstance{
					Status: svcapitypes.DBInstanceStatus{},
				},
				out: &svcsdk.DescribeDBInstancesOutput{
					DBInstances: []*svcsdk.DBInstance{
						{
							DBClusterIdentifier:  aws.String("db-cluster-id"),
							DBInstanceIdentifier: aws.String("db-instance-id"),
							DeletionProtection:   aws.Bool(true),
						},
					},
				},
				awsRDSClient: fake.MockRDSClient{
					MockDescribeDBClustersWithContext: func(ctx context.Context, input *svcsdk.DescribeDBClustersInput, optFns ...request.Option) (*svcsdk.DescribeDBClustersOutput, error) {
						return &svcsdk.DescribeDBClustersOutput{
							DBClusters: []*svcsdk.DBCluster{
								{
									DBClusterIdentifier: aws.String("db-cluster-id"),
									DeletionProtection:  aws.Bool(true),
									DBClusterMembers: []*svcsdk.DBClusterMember{
										{
											DBInstanceIdentifier: aws.String("db-instance-id"),
											IsClusterWriter:      aws.Bool(true),
										},
									},
								},
							},
						}, nil
					},
				},
				kube: test.NewMockClient(),
			},
			want: want{
				err: nil,
				statusAtProvider: &svcapitypes.CustomDBInstanceObservation{
					DatabaseRole: aws.String(databaseRoleClusterWriter),
				},
			},
		},
		"databaseRoleClusterReader": {
			args: args{
				cr: &svcapitypes.DBInstance{
					Status: svcapitypes.DBInstanceStatus{},
				},
				out: &svcsdk.DescribeDBInstancesOutput{
					DBInstances: []*svcsdk.DBInstance{
						{
							DBClusterIdentifier:  aws.String("db-cluster-id"),
							DBInstanceIdentifier: aws.String("db-instance-id"),
							DeletionProtection:   aws.Bool(true),
						},
					},
				},
				awsRDSClient: fake.MockRDSClient{
					MockDescribeDBClustersWithContext: func(ctx context.Context, input *svcsdk.DescribeDBClustersInput, optFns ...request.Option) (*svcsdk.DescribeDBClustersOutput, error) {
						return &svcsdk.DescribeDBClustersOutput{
							DBClusters: []*svcsdk.DBCluster{
								{
									DBClusterIdentifier: aws.String("db-cluster-id"),
									DeletionProtection:  aws.Bool(true),
									DBClusterMembers: []*svcsdk.DBClusterMember{
										{
											DBInstanceIdentifier: aws.String("another-db-instance-id"),
											IsClusterWriter:      aws.Bool(true),
										},
										{
											DBInstanceIdentifier: aws.String("db-instance-id"),
											IsClusterWriter:      aws.Bool(false),
										},
									},
								},
							},
						}, nil
					},
				},
				kube: test.NewMockClient(),
			},
			want: want{
				err: nil,
				statusAtProvider: &svcapitypes.CustomDBInstanceObservation{
					DatabaseRole: aws.String(databaseRoleClusterReader),
				},
			},
		},
		"databaseRoleInstance": {
			args: args{
				cr: &svcapitypes.DBInstance{
					Status: svcapitypes.DBInstanceStatus{},
				},
				out: &svcsdk.DescribeDBInstancesOutput{
					DBInstances: []*svcsdk.DBInstance{
						{
							DeletionProtection: aws.Bool(true),
						},
					},
				},
				kube: test.NewMockClient(),
			},
			want: want{
				err: nil,
				statusAtProvider: &svcapitypes.CustomDBInstanceObservation{
					DatabaseRole: aws.String(databaseRoleStandalone),
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			cr := tc.args.cr
			ce := newCustomExternal(tc.kube, &tc.awsRDSClient)
			_, err := ce.postObserve(context.TODO(), cr, tc.args.out, managed.ExternalObservation{}, nil)

			if diff := cmp.Diff(tc.want.err, err); diff != "" {
				t.Errorf("r: -want, +got error: \n%s", diff)
			}
			if diff := cmp.Diff(tc.want.statusAtProvider.DatabaseRole, cr.Status.AtProvider.DatabaseRole); diff != "" {
				t.Errorf("r: -want, +got: \n%s", diff)
			}
		})
	}
}
