package dbinstance

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	svcsdk "github.com/aws/aws-sdk-go/service/rds"
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
		"CreateStandaloneInstance": {
			args: args{
				cr: &svcapitypes.DBInstance{
					Spec: svcapitypes.DBInstanceSpec{
						ForProvider: svcapitypes.DBInstanceParameters{
							CustomDBInstanceParameters: svcapitypes.CustomDBInstanceParameters{
								AutogeneratePassword: true,
							},
						},
					},
					Status: svcapitypes.DBInstanceStatus{
						AtProvider: svcapitypes.DBInstanceObservation{},
					},
				},
				kube: test.NewMockClient(),
				awsRDSClient: fake.MockRDSClient{
					MockCreateDBInstanceWithContext: func(ctx context.Context, input *svcsdk.CreateDBInstanceInput, optFns ...request.Option) (*svcsdk.CreateDBInstanceOutput, error) {
						return &svcsdk.CreateDBInstanceOutput{DBInstance: &svcsdk.DBInstance{}}, nil
					},
					MockCreateDBInstanceReadReplicaWithContext: func(ctx context.Context, input *svcsdk.CreateDBInstanceReadReplicaInput, optFns ...request.Option) (*svcsdk.CreateDBInstanceReadReplicaOutput, error) {
						return &svcsdk.CreateDBInstanceReadReplicaOutput{}, nil
					},
				},
			},
			want: want{
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
		kube client.Client
		cr   *svcapitypes.DBInstance
		out  *svcsdk.DescribeDBInstancesOutput
	}

	type want struct {
		upToDate         bool
		err              error
		statusAtProvider *svcapitypes.CustomDBInstanceObservation
	}

	cases := map[string]struct {
		args
		want
	}{
		"UpToDateReadReplicaWithReplicatedMasterCredentials": {
			args: args{
				cr: &svcapitypes.DBInstance{
					Spec: svcapitypes.DBInstanceSpec{
						ForProvider: svcapitypes.DBInstanceParameters{
							DeletionProtection: aws.Bool(true),
						},
					},
				},
				out: &svcsdk.DescribeDBInstancesOutput{
					DBInstances: []*svcsdk.DBInstance{
						{
							DeletionProtection:                   aws.Bool(true),
							ReadReplicaSourceDBClusterIdentifier: aws.String("source-db-instance-id"),
						},
					},
				},
				kube: &test.MockClient{
					MockGet: func(ctx context.Context, key client.ObjectKey, obj client.Object) error {
						return errors.New("not found")
					},
				},
			},
			want: want{
				upToDate: true,
				err:      nil,
				statusAtProvider: &svcapitypes.CustomDBInstanceObservation{
					DatabaseRole: aws.String(databaseRoleReadReplica),
				},
			},
		},
		"databaseRolePrimary": {
			args: args{
				cr: &svcapitypes.DBInstance{
					Spec: svcapitypes.DBInstanceSpec{
						ForProvider: svcapitypes.DBInstanceParameters{
							DeletionProtection: aws.Bool(true),
						},
					},
				},
				out: &svcsdk.DescribeDBInstancesOutput{
					DBInstances: []*svcsdk.DBInstance{
						{
							DeletionProtection:               aws.Bool(true),
							ReadReplicaDBInstanceIdentifiers: []*string{aws.String("db-read-replica-id")},
						},
					},
				},
				kube: test.NewMockClient(),
			},
			want: want{
				upToDate: true,
				err:      nil,
				statusAtProvider: &svcapitypes.CustomDBInstanceObservation{
					DatabaseRole: aws.String(databaseRolePrimary),
				},
			},
		},
		"UpToDateStandaloneInstance": {
			args: args{
				cr: &svcapitypes.DBInstance{
					Spec: svcapitypes.DBInstanceSpec{
						ForProvider: svcapitypes.DBInstanceParameters{
							DeletionProtection: aws.Bool(true),
						},
					},
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
				upToDate: true,
				err:      nil,
				statusAtProvider: &svcapitypes.CustomDBInstanceObservation{
					DatabaseRole: aws.String(databaseRoleStandalone),
				},
			},
		},
		"Ignores Tags with TagsIgnore prefix*": {
			args: args{
				cr: &svcapitypes.DBInstance{
					Spec: svcapitypes.DBInstanceSpec{
						ForProvider: svcapitypes.DBInstanceParameters{
							CustomDBInstanceParameters: svcapitypes.CustomDBInstanceParameters{
								TagsIgnore: []svcapitypes.TagIgnoreRule{{Key: "aws:*"}, {Key: "c7n:*"}},
							},
							Tags: []*svcapitypes.Tag{
								{Key: aws.String("env"), Value: aws.String("prod")},
							},
							DeletionProtection: aws.Bool(true),
						},
					},
				},
				out: &svcsdk.DescribeDBInstancesOutput{
					DBInstances: []*svcsdk.DBInstance{
						{
							DeletionProtection: aws.Bool(true),
							TagList: []*svcsdk.Tag{
								{Key: aws.String("aws:createdBy"), Value: aws.String("terraform")},
								{Key: aws.String("c7n:policy"), Value: aws.String("auto")},
								{Key: aws.String("env"), Value: aws.String("prod")},
							},
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
		"Ignores Tags with TagsIgnore exact": {
			args: args{
				cr: &svcapitypes.DBInstance{
					Spec: svcapitypes.DBInstanceSpec{
						ForProvider: svcapitypes.DBInstanceParameters{
							CustomDBInstanceParameters: svcapitypes.CustomDBInstanceParameters{
								TagsIgnore: []svcapitypes.TagIgnoreRule{{Key: "aws:*"}, {Key: "c7n:policy"}},
							},
							Tags: []*svcapitypes.Tag{
								{Key: aws.String("env"), Value: aws.String("prod")},
							},
							DeletionProtection: aws.Bool(true),
						},
					},
				},
				out: &svcsdk.DescribeDBInstancesOutput{
					DBInstances: []*svcsdk.DBInstance{
						{
							DeletionProtection: aws.Bool(true),
							TagList: []*svcsdk.Tag{
								{Key: aws.String("aws:createdBy"), Value: aws.String("terraform")},
								{Key: aws.String("c7n:policy"), Value: aws.String("auto")},
								{Key: aws.String("c7n:other"), Value: aws.String("x")},
								{Key: aws.String("env"), Value: aws.String("prod")},
							},
						},
					},
				},
				kube: test.NewMockClient(),
			},
			want: want{
				upToDate: false, // c7n:other should be removed since not ignored and not in spec
				err:      nil,
				statusAtProvider: &svcapitypes.CustomDBInstanceObservation{
					DatabaseRole: aws.String(databaseRoleStandalone),
				},
			},
		},
		"DoesNotIgnoreAllWithStarOnlyRule": {
			args: args{
				cr: &svcapitypes.DBInstance{
					Spec: svcapitypes.DBInstanceSpec{
						ForProvider: svcapitypes.DBInstanceParameters{
							CustomDBInstanceParameters: svcapitypes.CustomDBInstanceParameters{
								// User attempts to ignore all tags with a single "*" rule; guard should prevent this.
								TagsIgnore: []svcapitypes.TagIgnoreRule{{Key: "*"}},
							},
							// Desired spec has no tags.
							Tags:               []*svcapitypes.Tag{},
							DeletionProtection: aws.Bool(true),
						},
					},
				},
				out: &svcsdk.DescribeDBInstancesOutput{
					DBInstances: []*svcsdk.DBInstance{
						{
							DeletionProtection: aws.Bool(true),
							TagList: []*svcsdk.Tag{
								{Key: aws.String("env"), Value: aws.String("prod")}, // Should not be ignored; will cause diff
							},
						},
					},
				},
				kube: test.NewMockClient(),
			},
			want: want{
				upToDate: false, // env tag should be scheduled for removal
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
			ce := newCustomExternal(tc.kube, nil)
			upToDate, _, err := ce.isUpToDate(context.TODO(), cr, tc.args.out)

			if diff := cmp.Diff(tc.want.err, err); diff != "" {
				t.Errorf("r: -want, +got error: \n%s", diff)
			}
			if diff := cmp.Diff(tc.want.upToDate, upToDate); diff != "" {
				t.Errorf("r: -want, +got: \n%s", diff)
			}
			if diff := cmp.Diff(tc.want.statusAtProvider.DatabaseRole, cr.Status.AtProvider.DatabaseRole); diff != "" {
				t.Errorf("r: -want, +got: \n%s", diff)
			}
		})
	}
}
