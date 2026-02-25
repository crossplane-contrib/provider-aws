package dbinstance

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	svcsdk "github.com/aws/aws-sdk-go/service/rds"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"sigs.k8s.io/controller-runtime/pkg/client"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/rds/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/rds/fake"
)

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
		"upToDateStandaloneInstance": {
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
				statusAtProvider: &svcapitypes.CustomDBInstanceObservation{
					DatabaseRole: aws.String(databaseRoleStandalone),
				},
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
