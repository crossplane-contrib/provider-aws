package dbcluster

import (
	"context"
	_ "embed"
	"errors"
	"testing"

	svcsdk "github.com/aws/aws-sdk-go/service/rds"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/rds/v1alpha1"
	rds "github.com/crossplane-contrib/provider-aws/pkg/clients/rds"
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

func TestIsVPCSecurityGroupIDsUpToDate(t *testing.T) {
	type args struct {
		cr  *svcapitypes.DBCluster
		out *svcsdk.DescribeDBClustersOutput
	}

	type want struct {
		isUpToDate bool
	}

	cases := map[string]struct {
		args
		want
	}{
		"NotAsMany": {
			args: args{
				cr: &svcapitypes.DBCluster{
					Spec: svcapitypes.DBClusterSpec{
						ForProvider: svcapitypes.DBClusterParameters{
							CustomDBClusterParameters: svcapitypes.CustomDBClusterParameters{
								VPCSecurityGroupIDs: []string{"sg-123", "sg-456"},
							},
						},
					},
				},
				out: &svcsdk.DescribeDBClustersOutput{
					DBClusters: []*svcsdk.DBCluster{
						{
							VpcSecurityGroups: []*svcsdk.VpcSecurityGroupMembership{
								{
									VpcSecurityGroupId: ptr.To("sg-123"),
								},
								{
									VpcSecurityGroupId: ptr.To("sg-456"),
								},
								{
									VpcSecurityGroupId: ptr.To("sg-789"),
								},
							},
						},
					},
				},
			},
			want: want{
				isUpToDate: false,
			},
		},
		"DesiredBeingManged": { // AWS default or managed by DBCluster
			args: args{
				cr: &svcapitypes.DBCluster{
					Spec: svcapitypes.DBClusterSpec{
						ForProvider: svcapitypes.DBClusterParameters{
							CustomDBClusterParameters: svcapitypes.CustomDBClusterParameters{
								VPCSecurityGroupIDs: nil,
							},
						},
					},
				},
				out: &svcsdk.DescribeDBClustersOutput{
					DBClusters: []*svcsdk.DBCluster{
						{
							VpcSecurityGroups: []*svcsdk.VpcSecurityGroupMembership{
								{
									VpcSecurityGroupId: ptr.To("sg-456"),
								},
								{
									VpcSecurityGroupId: ptr.To("sg-123"),
								},
							},
						},
					},
				},
			},
			want: want{
				isUpToDate: true,
			},
		},
		"ActualEmpty": {
			args: args{
				cr: &svcapitypes.DBCluster{
					Spec: svcapitypes.DBClusterSpec{
						ForProvider: svcapitypes.DBClusterParameters{
							CustomDBClusterParameters: svcapitypes.CustomDBClusterParameters{
								VPCSecurityGroupIDs: []string{"sg-123", "sg-456"},
							},
						},
					},
				},
				out: &svcsdk.DescribeDBClustersOutput{
					DBClusters: []*svcsdk.DBCluster{
						{
							VpcSecurityGroups: nil,
						},
					},
				},
			},
			want: want{
				isUpToDate: false,
			},
		},
		"Unsorted": {
			args: args{
				cr: &svcapitypes.DBCluster{
					Spec: svcapitypes.DBClusterSpec{
						ForProvider: svcapitypes.DBClusterParameters{
							CustomDBClusterParameters: svcapitypes.CustomDBClusterParameters{
								VPCSecurityGroupIDs: []string{"sg-123", "sg-456"},
							},
						},
					},
				},
				out: &svcsdk.DescribeDBClustersOutput{
					DBClusters: []*svcsdk.DBCluster{
						{
							VpcSecurityGroups: []*svcsdk.VpcSecurityGroupMembership{
								{
									VpcSecurityGroupId: ptr.To("sg-456"),
								},
								{
									VpcSecurityGroupId: ptr.To("sg-123"),
								},
							},
						},
					},
				},
			},
			want: want{
				isUpToDate: true,
			},
		},
		"Identical": {
			args: args{
				cr: &svcapitypes.DBCluster{
					Spec: svcapitypes.DBClusterSpec{
						ForProvider: svcapitypes.DBClusterParameters{
							CustomDBClusterParameters: svcapitypes.CustomDBClusterParameters{
								VPCSecurityGroupIDs: []string{"sg-123", "sg-456"},
							},
						},
					},
				},
				out: &svcsdk.DescribeDBClustersOutput{
					DBClusters: []*svcsdk.DBCluster{
						{
							VpcSecurityGroups: []*svcsdk.VpcSecurityGroupMembership{
								{
									VpcSecurityGroupId: ptr.To("sg-123"),
								},
								{
									VpcSecurityGroupId: ptr.To("sg-456"),
								},
							},
						},
					},
				},
			},
			want: want{
				isUpToDate: true,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			isUpToDate := areVPCSecurityGroupIDsUpToDate(tc.args.cr, tc.args.out)

			if diff := cmp.Diff(tc.want.isUpToDate, isUpToDate); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestIsEngineVersionUpToDate(t *testing.T) {
	type args struct {
		cr  *svcapitypes.DBCluster
		out *svcsdk.DescribeDBClustersOutput
	}

	type want struct {
		isUpToDate bool
	}

	cases := map[string]struct {
		args
		want
	}{
		"Default": {
			args: args{
				cr: &svcapitypes.DBCluster{
					Spec: svcapitypes.DBClusterSpec{
						ForProvider: svcapitypes.DBClusterParameters{
							CustomDBClusterParameters: svcapitypes.CustomDBClusterParameters{
								EngineVersion: nil,
							},
						},
					},
				},
				out: &svcsdk.DescribeDBClustersOutput{
					DBClusters: []*svcsdk.DBCluster{
						{
							EngineVersion: ptr.To("12.3"), // some AWS "default" value
						},
					},
				},
			},
			want: want{
				isUpToDate: true,
			},
		},
		"Identical": {
			args: args{
				cr: &svcapitypes.DBCluster{
					Spec: svcapitypes.DBClusterSpec{
						ForProvider: svcapitypes.DBClusterParameters{
							CustomDBClusterParameters: svcapitypes.CustomDBClusterParameters{
								EngineVersion: ptr.To("12.3"),
							},
						},
					},
				},
				out: &svcsdk.DescribeDBClustersOutput{
					DBClusters: []*svcsdk.DBCluster{
						{
							EngineVersion: ptr.To("12.3"), // some AWS "default" value
						},
					},
				},
			},
			want: want{
				isUpToDate: true,
			},
		},
		"IdenticalMajorVersionOnly": {
			args: args{
				cr: &svcapitypes.DBCluster{
					Spec: svcapitypes.DBClusterSpec{
						ForProvider: svcapitypes.DBClusterParameters{
							CustomDBClusterParameters: svcapitypes.CustomDBClusterParameters{
								EngineVersion: ptr.To("12"),
							},
						},
					},
				},
				out: &svcsdk.DescribeDBClustersOutput{
					DBClusters: []*svcsdk.DBCluster{
						{
							EngineVersion: ptr.To("12.3"), // some AWS "default" value
						},
					},
				},
			},
			want: want{
				isUpToDate: true,
			},
		},
		"IdenticalAtDowngrade": {
			args: args{
				cr: &svcapitypes.DBCluster{
					Spec: svcapitypes.DBClusterSpec{
						ForProvider: svcapitypes.DBClusterParameters{
							CustomDBClusterParameters: svcapitypes.CustomDBClusterParameters{
								EngineVersion: ptr.To("12.1"),
							},
						},
					},
				},
				out: &svcsdk.DescribeDBClustersOutput{
					DBClusters: []*svcsdk.DBCluster{
						{
							EngineVersion: ptr.To("12.3"), // some AWS "default" value
						},
					},
				},
			},
			want: want{
				isUpToDate: true,
			},
		},
		"Different": {
			args: args{
				cr: &svcapitypes.DBCluster{
					Spec: svcapitypes.DBClusterSpec{
						ForProvider: svcapitypes.DBClusterParameters{
							CustomDBClusterParameters: svcapitypes.CustomDBClusterParameters{
								EngineVersion: ptr.To("13"),
							},
						},
					},
				},
				out: &svcsdk.DescribeDBClustersOutput{
					DBClusters: []*svcsdk.DBCluster{
						{
							EngineVersion: ptr.To("12.3"),
						},
					},
				},
			},
			want: want{
				isUpToDate: false,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			isUpToDate := isEngineVersionUpToDate(tc.args.cr, tc.args.out)

			if diff := cmp.Diff(tc.want.isUpToDate, isUpToDate); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestIsDBClusterParameterGroupNameUpToDate(t *testing.T) {
	type args struct {
		cr  *svcapitypes.DBCluster
		out *svcsdk.DescribeDBClustersOutput
	}

	type want struct {
		isUpToDate bool
	}

	cases := map[string]struct {
		args
		want
	}{
		"Nil": {
			args: args{
				cr: &svcapitypes.DBCluster{
					Spec: svcapitypes.DBClusterSpec{
						ForProvider: svcapitypes.DBClusterParameters{
							DBClusterParameterGroupName: nil,
						},
					},
				},
				out: &svcsdk.DescribeDBClustersOutput{
					DBClusters: []*svcsdk.DBCluster{
						{
							DBClusterParameterGroup: nil,
						},
					},
				},
			},
			want: want{
				isUpToDate: true,
			},
		},
		"Default": {
			args: args{
				cr: &svcapitypes.DBCluster{
					Spec: svcapitypes.DBClusterSpec{
						ForProvider: svcapitypes.DBClusterParameters{
							DBClusterParameterGroupName: nil,
						},
					},
				},
				out: &svcsdk.DescribeDBClustersOutput{
					DBClusters: []*svcsdk.DBCluster{
						{
							DBClusterParameterGroup: ptr.To("default.aurora-postgresql14"), // some AWS "default" value
						},
					},
				},
			},
			want: want{
				isUpToDate: true,
			},
		},
		"Identical": {
			args: args{
				cr: &svcapitypes.DBCluster{
					Spec: svcapitypes.DBClusterSpec{
						ForProvider: svcapitypes.DBClusterParameters{
							DBClusterParameterGroupName: ptr.To("default.aurora-postgresql14"),
						},
					},
				},
				out: &svcsdk.DescribeDBClustersOutput{
					DBClusters: []*svcsdk.DBCluster{
						{
							DBClusterParameterGroup: ptr.To("default.aurora-postgresql14"), // some AWS "default" value
						},
					},
				},
			},
			want: want{
				isUpToDate: true,
			},
		},
		"Different": {
			args: args{
				cr: &svcapitypes.DBCluster{
					Spec: svcapitypes.DBClusterSpec{
						ForProvider: svcapitypes.DBClusterParameters{
							DBClusterParameterGroupName: ptr.To("default.aurora-postgresql15"),
						},
					},
				},
				out: &svcsdk.DescribeDBClustersOutput{
					DBClusters: []*svcsdk.DBCluster{
						{
							DBClusterParameterGroup: ptr.To("default.aurora-postgresql14"),
						},
					},
				},
			},
			want: want{
				isUpToDate: false,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			isUpToDate := isDBClusterParameterGroupNameUpToDate(tc.args.cr, tc.args.out)

			if diff := cmp.Diff(tc.want.isUpToDate, isUpToDate); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestIsPreferredBackupWindowUpToDate(t *testing.T) {
	type args struct {
		cr  *svcapitypes.DBCluster
		out *svcsdk.DescribeDBClustersOutput
	}

	type want struct {
		isUpToDate bool
	}

	cases := map[string]struct {
		args
		want
	}{
		"Nil": {
			args: args{
				cr: &svcapitypes.DBCluster{
					Spec: svcapitypes.DBClusterSpec{
						ForProvider: svcapitypes.DBClusterParameters{
							PreferredBackupWindow: nil,
						},
					},
				},
				out: &svcsdk.DescribeDBClustersOutput{
					DBClusters: []*svcsdk.DBCluster{
						{
							PreferredBackupWindow: nil,
						},
					},
				},
			},
			want: want{
				isUpToDate: true,
			},
		},
		"Default": {
			args: args{
				cr: &svcapitypes.DBCluster{
					Spec: svcapitypes.DBClusterSpec{
						ForProvider: svcapitypes.DBClusterParameters{
							PreferredBackupWindow: nil,
						},
					},
				},
				out: &svcsdk.DescribeDBClustersOutput{
					DBClusters: []*svcsdk.DBCluster{
						{
							PreferredBackupWindow: ptr.To("03:00-04:00"), // some AWS "default" value
						},
					},
				},
			},
			want: want{
				isUpToDate: true,
			},
		},
		"Identical": {
			args: args{
				cr: &svcapitypes.DBCluster{
					Spec: svcapitypes.DBClusterSpec{
						ForProvider: svcapitypes.DBClusterParameters{
							PreferredBackupWindow: ptr.To("03:00-04:00"),
						},
					},
				},
				out: &svcsdk.DescribeDBClustersOutput{
					DBClusters: []*svcsdk.DBCluster{
						{
							PreferredBackupWindow: ptr.To("03:00-04:00"),
						},
					},
				},
			},
			want: want{
				isUpToDate: true,
			},
		},
		"Different": {
			args: args{
				cr: &svcapitypes.DBCluster{
					Spec: svcapitypes.DBClusterSpec{
						ForProvider: svcapitypes.DBClusterParameters{
							PreferredBackupWindow: ptr.To("04:00-05:00"),
						},
					},
				},
				out: &svcsdk.DescribeDBClustersOutput{
					DBClusters: []*svcsdk.DBCluster{
						{
							PreferredBackupWindow: ptr.To("03:00-04:00"),
						},
					},
				},
			},
			want: want{
				isUpToDate: false,
			},
		},
		"DifferentButAWSBackupManaged": {
			args: args{
				cr: &svcapitypes.DBCluster{
					Spec: svcapitypes.DBClusterSpec{
						ForProvider: svcapitypes.DBClusterParameters{
							PreferredBackupWindow: ptr.To("04:00-05:00"),
						},
					},
				},
				out: &svcsdk.DescribeDBClustersOutput{
					DBClusters: []*svcsdk.DBCluster{
						{
							AwsBackupRecoveryPointArn: ptr.To("arn:aws:backup:eu-central-1:123456789012:recovery-point:continuous:cluster-random-string-hash"),
							PreferredBackupWindow:     ptr.To("03:00-04:00"),
						},
					},
				},
			},
			want: want{
				isUpToDate: true, // we consider it up to date, because AWS Backup takes ownership of PreferredBackupWindow if AwsBackupRecoveryPointArn is set
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			isUpToDate := isPreferredBackupWindowUpToDate(tc.args.cr, tc.args.out)

			if diff := cmp.Diff(tc.want.isUpToDate, isUpToDate); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestIsBackupRetentionPeriodUpToDate(t *testing.T) {
	type args struct {
		cr  *svcapitypes.DBCluster
		out *svcsdk.DescribeDBClustersOutput
	}

	type want struct {
		isUpToDate bool
	}

	cases := map[string]struct {
		args
		want
	}{
		"Nil": {
			args: args{
				cr: &svcapitypes.DBCluster{
					Spec: svcapitypes.DBClusterSpec{
						ForProvider: svcapitypes.DBClusterParameters{
							BackupRetentionPeriod: nil,
						},
					},
				},
				out: &svcsdk.DescribeDBClustersOutput{
					DBClusters: []*svcsdk.DBCluster{
						{
							BackupRetentionPeriod: nil,
						},
					},
				},
			},
			want: want{
				isUpToDate: true,
			},
		},
		"Default": {
			args: args{
				cr: &svcapitypes.DBCluster{
					Spec: svcapitypes.DBClusterSpec{
						ForProvider: svcapitypes.DBClusterParameters{
							BackupRetentionPeriod: nil,
						},
					},
				},
				out: &svcsdk.DescribeDBClustersOutput{
					DBClusters: []*svcsdk.DBCluster{
						{
							BackupRetentionPeriod: ptr.To(int64(7)), // some AWS "default" value
						},
					},
				},
			},
			want: want{
				isUpToDate: true,
			},
		},
		"Identical": {
			args: args{
				cr: &svcapitypes.DBCluster{
					Spec: svcapitypes.DBClusterSpec{
						ForProvider: svcapitypes.DBClusterParameters{
							BackupRetentionPeriod: ptr.To(int64(7)),
						},
					},
				},
				out: &svcsdk.DescribeDBClustersOutput{
					DBClusters: []*svcsdk.DBCluster{
						{
							BackupRetentionPeriod: ptr.To(int64(7)),
						},
					},
				},
			},
			want: want{
				isUpToDate: true,
			},
		},
		"Different": {
			args: args{
				cr: &svcapitypes.DBCluster{
					Spec: svcapitypes.DBClusterSpec{
						ForProvider: svcapitypes.DBClusterParameters{
							BackupRetentionPeriod: ptr.To(int64(14)),
						},
					},
				},
				out: &svcsdk.DescribeDBClustersOutput{
					DBClusters: []*svcsdk.DBCluster{
						{
							BackupRetentionPeriod: ptr.To(int64(7)),
						},
					},
				},
			},
			want: want{
				isUpToDate: false,
			},
		},
		"DifferentButAWSBackupManaged": {
			args: args{
				cr: &svcapitypes.DBCluster{
					Spec: svcapitypes.DBClusterSpec{
						ForProvider: svcapitypes.DBClusterParameters{
							BackupRetentionPeriod: ptr.To(int64(14)),
						},
					},
				},
				out: &svcsdk.DescribeDBClustersOutput{
					DBClusters: []*svcsdk.DBCluster{
						{
							AwsBackupRecoveryPointArn: ptr.To("arn:aws:backup:eu-central-1:123456789012:recovery-point:continuous:cluster-random-string-hash"),
							BackupRetentionPeriod:     ptr.To(int64(7)),
						},
					},
				},
			},
			want: want{
				isUpToDate: true, // we consider it up to date, because AWS Backup takes ownership of BackupRetentionPeriod if AwsBackupRecoveryPointArn is set
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			isUpToDate := isBackupRetentionPeriodUpToDate(tc.args.cr, tc.args.out)

			if diff := cmp.Diff(tc.want.isUpToDate, isUpToDate); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestIsUpToDate(t *testing.T) {
	type args struct {
		cr   *svcapitypes.DBCluster
		out  *svcsdk.DescribeDBClustersOutput
		kube client.Client
	}

	type want struct {
		isUpToDate bool
		err        error
	}

	cases := map[string]struct {
		args
		want
	}{
		"DifferentServerlessV2ScalingConfigurationMinimum": {
			args: args{
				kube: &test.MockClient{
					// This mock returns the same password for both secrets(masterUserPasswordSecretRef and cached secret)
					MockGet: newMockGetFnCustomWithDiff(false, nil, []objectFnCustom{mockGettingSecretData}),
				},
				cr: &svcapitypes.DBCluster{
					Spec: svcapitypes.DBClusterSpec{
						ForProvider: svcapitypes.DBClusterParameters{
							DBClusterParameterGroupName: ptr.To("default.aurora-postgresql15"),
							ServerlessV2ScalingConfiguration: &svcapitypes.ServerlessV2ScalingConfiguration{
								MaxCapacity: ptr.To[float64](5.0),
								MinCapacity: ptr.To[float64](1.0),
							},
						},
					},
				},
				out: &svcsdk.DescribeDBClustersOutput{
					DBClusters: []*svcsdk.DBCluster{
						{
							DBClusterParameterGroup: ptr.To("default.aurora-postgresql14"),
							ServerlessV2ScalingConfiguration: &svcsdk.ServerlessV2ScalingConfigurationInfo{
								MaxCapacity: ptr.To[float64](5.0),
								MinCapacity: ptr.To[float64](3.0),
							},
						},
					},
				},
			},
			want: want{
				isUpToDate: false,
			},
		},
		"DifferentServerlessV2ScalingConfigurationMaximum": {
			args: args{
				kube: test.NewMockClient(),
				cr: &svcapitypes.DBCluster{
					Spec: svcapitypes.DBClusterSpec{
						ForProvider: svcapitypes.DBClusterParameters{
							DBClusterParameterGroupName: ptr.To("default.aurora-postgresql15"),
							ServerlessV2ScalingConfiguration: &svcapitypes.ServerlessV2ScalingConfiguration{
								MaxCapacity: ptr.To[float64](5.0),
								MinCapacity: ptr.To[float64](1.0),
							},
						},
					},
				},
				out: &svcsdk.DescribeDBClustersOutput{
					DBClusters: []*svcsdk.DBCluster{
						{
							DBClusterParameterGroup: ptr.To("default.aurora-postgresql14"),
							ServerlessV2ScalingConfiguration: &svcsdk.ServerlessV2ScalingConfigurationInfo{
								MaxCapacity: ptr.To[float64](10.0),
								MinCapacity: ptr.To[float64](1.0),
							},
						},
					},
				},
			},
			want: want{
				isUpToDate: false,
			},
		},
		"SameServerlessV2ScalingConfiguration": {
			args: args{
				kube: &test.MockClient{
					// This mock returns the same password for both secrets(masterUserPasswordSecretRef and cached secret)
					MockGet: newMockGetFnCustomWithDiff(false, nil, []objectFnCustom{mockGettingSecretData}),
				},
				cr: &svcapitypes.DBCluster{
					Spec: svcapitypes.DBClusterSpec{
						ForProvider: svcapitypes.DBClusterParameters{
							DBClusterParameterGroupName: ptr.To("default.aurora-postgresql15"),
							ServerlessV2ScalingConfiguration: &svcapitypes.ServerlessV2ScalingConfiguration{
								MaxCapacity: ptr.To[float64](5.0),
								MinCapacity: ptr.To[float64](1.0),
							},
						},
					},
				},
				out: &svcsdk.DescribeDBClustersOutput{
					DBClusters: []*svcsdk.DBCluster{
						{
							DBClusterParameterGroup: ptr.To("default.aurora-postgresql14"),
							ServerlessV2ScalingConfiguration: &svcsdk.ServerlessV2ScalingConfigurationInfo{
								MaxCapacity: ptr.To[float64](5.0),
								MinCapacity: ptr.To[float64](1.0),
							},
						},
					},
				},
			},
			want: want{
				isUpToDate: false,
			},
		},
		"SameEnableCloudwatchLogsExports": {
			args: args{
				kube: &test.MockClient{
					// This mock returns the same password for both secrets(masterUserPasswordSecretRef and cached secret)
					MockGet: newMockGetFnCustomWithDiff(false, nil, []objectFnCustom{mockGettingSecretData}),
				},
				cr: &svcapitypes.DBCluster{
					Spec: svcapitypes.DBClusterSpec{
						ForProvider: svcapitypes.DBClusterParameters{
							DBClusterParameterGroupName: ptr.To("default.aurora-postgresql15"),
							EnableCloudwatchLogsExports: []*string{
								ptr.To("postgresql"),
							},
						},
					},
				},
				out: &svcsdk.DescribeDBClustersOutput{
					DBClusters: []*svcsdk.DBCluster{
						{
							DBClusterParameterGroup: ptr.To("default.aurora-postgresql15"),
							EnabledCloudwatchLogsExports: []*string{
								ptr.To("postgresql"),
							},
						},
					},
				},
			},
			want: want{
				isUpToDate: true,
			},
		},
		"DifferentEnableCloudwatchLogsExportsAddExports": {
			args: args{
				kube: &test.MockClient{
					// This mock returns the same password for both secrets(masterUserPasswordSecretRef and cached secret)
					MockGet: newMockGetFnCustomWithDiff(false, nil, []objectFnCustom{mockGettingSecretData}),
				},
				cr: &svcapitypes.DBCluster{
					Spec: svcapitypes.DBClusterSpec{
						ForProvider: svcapitypes.DBClusterParameters{
							DBClusterParameterGroupName: ptr.To("default.aurora-postgresql15"),
							EnableCloudwatchLogsExports: []*string{
								ptr.To("postgresql"),
							},
						},
					},
				},
				out: &svcsdk.DescribeDBClustersOutput{
					DBClusters: []*svcsdk.DBCluster{
						{
							DBClusterParameterGroup: ptr.To("default.aurora-postgresql14"),
						},
					},
				},
			},
			want: want{
				isUpToDate: false,
			},
		},
		"DifferentEnableCloudwatchLogsExportsRemoveExports": {
			args: args{
				kube: &test.MockClient{
					// This mock returns the same password for both secrets(masterUserPasswordSecretRef and cached secret)
					MockGet: newMockGetFnCustomWithDiff(false, nil, []objectFnCustom{mockGettingSecretData}),
				},
				cr: &svcapitypes.DBCluster{
					Spec: svcapitypes.DBClusterSpec{
						ForProvider: svcapitypes.DBClusterParameters{
							DBClusterParameterGroupName: ptr.To("default.aurora-postgresql15"),
						},
					},
				},
				out: &svcsdk.DescribeDBClustersOutput{
					DBClusters: []*svcsdk.DBCluster{
						{
							DBClusterParameterGroup: ptr.To("default.aurora-postgresql14"),
							EnabledCloudwatchLogsExports: []*string{
								ptr.To("postgresql"),
							},
						},
					},
				},
			},
			want: want{
				isUpToDate: false,
			},
		},

		"ChangedMasterPassword": {
			args: args{
				cr: &svcapitypes.DBCluster{
					Spec: svcapitypes.DBClusterSpec{
						ForProvider: svcapitypes.DBClusterParameters{
							Engine: ptr.To("aurora-postgresql"),
							CustomDBClusterParameters: svcapitypes.CustomDBClusterParameters{
								EngineVersion: ptr.To("16"),
								MasterUserPasswordSecretRef: &xpv1.SecretKeySelector{
									SecretReference: xpv1.SecretReference{
										Name:      "masterUserPassword",
										Namespace: "default",
									},
									Key: "password",
								},
							},
							Region: "eu-central-2",
						},
					},
				},
				out: &svcsdk.DescribeDBClustersOutput{
					DBClusters: []*svcsdk.DBCluster{
						{
							Engine:        ptr.To("aurora-postgresql"),
							EngineVersion: ptr.To("16"),
						},
					},
				},
				kube: &test.MockClient{
					// This mock returns different passwords from masterUserPasswordSecretRef and cached secret
					MockGet: newMockGetFnCustomWithDiff(true, nil, []objectFnCustom{mockGettingSecretData}),
				},
			},
			want: want{
				isUpToDate: false,
			},
		},
		"UpToDatePendingModifiedValue": {
			args: args{
				cr: &svcapitypes.DBCluster{
					Spec: svcapitypes.DBClusterSpec{
						ForProvider: svcapitypes.DBClusterParameters{
							StorageType: ptr.To("gp2"),
						},
					},
				},
				out: &svcsdk.DescribeDBClustersOutput{
					DBClusters: []*svcsdk.DBCluster{
						{
							StorageType: ptr.To("gp3"),
							PendingModifiedValues: &svcsdk.ClusterPendingModifiedValues{
								StorageType: ptr.To("gp2"),
							},
						},
					},
				},
				kube: test.NewMockClient(),
			},
			want: want{
				isUpToDate: true,
				err:        nil,
			},
		},
		"IsNoTUpToDatePendingModifiedValue": {
			args: args{
				cr: &svcapitypes.DBCluster{
					Spec: svcapitypes.DBClusterSpec{
						ForProvider: svcapitypes.DBClusterParameters{
							StorageType: ptr.To("gp2"),
						},
					},
				},
				out: &svcsdk.DescribeDBClustersOutput{
					DBClusters: []*svcsdk.DBCluster{
						{
							StorageType: ptr.To("io1"),
							PendingModifiedValues: &svcsdk.ClusterPendingModifiedValues{
								StorageType: ptr.To("gp3"),
							},
						},
					},
				},
				kube: test.NewMockClient(),
			},
			want: want{
				isUpToDate: false,
				err:        nil,
			},
		},
		"IsNoTUpToDatePendingModifiedValueApplyImmediately": { // The parameter is already scheduled to be updated, but we want to update it immediately
			args: args{
				cr: &svcapitypes.DBCluster{
					Spec: svcapitypes.DBClusterSpec{
						ForProvider: svcapitypes.DBClusterParameters{
							BackupRetentionPeriod: ptr.To(int64(21)),
							CustomDBClusterParameters: svcapitypes.CustomDBClusterParameters{
								ApplyImmediately: ptr.To(true),
							},
						},
					},
				},
				out: &svcsdk.DescribeDBClustersOutput{
					DBClusters: []*svcsdk.DBCluster{
						{
							BackupRetentionPeriod: ptr.To(int64(7)),
							PendingModifiedValues: &svcsdk.ClusterPendingModifiedValues{
								BackupRetentionPeriod: ptr.To(int64(21)),
							},
						},
					},
				},
				kube: test.NewMockClient(),
			},
			want: want{
				isUpToDate: false,
				err:        nil,
			},
		},
		"UpToDatePendingModifiedValueEngineVersion": {
			args: args{
				cr: &svcapitypes.DBCluster{
					Spec: svcapitypes.DBClusterSpec{
						ForProvider: svcapitypes.DBClusterParameters{
							Engine: ptr.To("aurora-postgresql"),
							CustomDBClusterParameters: svcapitypes.CustomDBClusterParameters{
								EngineVersion: ptr.To("15.2"),
							},
						},
					},
				},
				out: &svcsdk.DescribeDBClustersOutput{
					DBClusters: []*svcsdk.DBCluster{
						{
							Engine:        ptr.To("aurora-postgresql"),
							EngineVersion: ptr.To("15.1"),
							PendingModifiedValues: &svcsdk.ClusterPendingModifiedValues{
								EngineVersion: ptr.To("15.2"),
							},
						},
					},
				},
				kube: test.NewMockClient(),
			},
			want: want{
				isUpToDate: true,
				err:        nil,
			},
		},
		"IsNotUpToDatePendingModifiedValueEngineVersion": {
			args: args{
				cr: &svcapitypes.DBCluster{
					Spec: svcapitypes.DBClusterSpec{
						ForProvider: svcapitypes.DBClusterParameters{
							Engine: ptr.To("aurora-postgresql"),
							CustomDBClusterParameters: svcapitypes.CustomDBClusterParameters{
								EngineVersion: ptr.To("15.3"),
							},
						},
					},
				},
				out: &svcsdk.DescribeDBClustersOutput{
					DBClusters: []*svcsdk.DBCluster{
						{
							Engine:        ptr.To("aurora-postgresql"),
							EngineVersion: ptr.To("15.1"),
							PendingModifiedValues: &svcsdk.ClusterPendingModifiedValues{
								EngineVersion: ptr.To("15.2"),
							},
						},
					},
				},
				kube: test.NewMockClient(),
			},
			want: want{
				isUpToDate: false,
				err:        nil,
			},
		},
		"IsNotUpToDatePendingModifiedValueEngineVersionRevert": {
			args: args{
				cr: &svcapitypes.DBCluster{
					Spec: svcapitypes.DBClusterSpec{
						ForProvider: svcapitypes.DBClusterParameters{
							Engine: ptr.To("aurora-postgresql"),
							CustomDBClusterParameters: svcapitypes.CustomDBClusterParameters{
								EngineVersion: ptr.To("15.1"),
							},
						},
					},
				},
				out: &svcsdk.DescribeDBClustersOutput{
					DBClusters: []*svcsdk.DBCluster{
						{
							Engine:        ptr.To("aurora-postgresql"),
							EngineVersion: ptr.To("15.1"),
							PendingModifiedValues: &svcsdk.ClusterPendingModifiedValues{
								EngineVersion: ptr.To("15.2"),
							},
						},
					},
				},
				kube: test.NewMockClient(),
			},
			want: want{
				isUpToDate: false,
				err:        nil,
			},
		},
		"IsNotUpToDatePendingModifiedValueEngineVersionApplyImmediately": {
			args: args{
				cr: &svcapitypes.DBCluster{
					Spec: svcapitypes.DBClusterSpec{
						ForProvider: svcapitypes.DBClusterParameters{
							Engine: ptr.To("aurora-postgresql"),
							CustomDBClusterParameters: svcapitypes.CustomDBClusterParameters{
								ApplyImmediately: ptr.To(true),
								EngineVersion:    ptr.To("15.3"),
							},
						},
					},
				},
				out: &svcsdk.DescribeDBClustersOutput{
					DBClusters: []*svcsdk.DBCluster{
						{
							Engine:        ptr.To("aurora-postgresql"),
							EngineVersion: ptr.To("15.1"),
							PendingModifiedValues: &svcsdk.ClusterPendingModifiedValues{
								EngineVersion: ptr.To("15.3"),
							},
						},
					},
				},
				kube: test.NewMockClient(),
			},
			want: want{
				isUpToDate: false,
				err:        nil,
			},
		},
		"DifferentIAMDatabaseAuthenticationEnabled": {
			args: args{
				cr: &svcapitypes.DBCluster{
					Spec: svcapitypes.DBClusterSpec{
						ForProvider: svcapitypes.DBClusterParameters{
							EnableIAMDatabaseAuthentication: ptr.To(true),
						},
					},
				},
				out: &svcsdk.DescribeDBClustersOutput{
					DBClusters: []*svcsdk.DBCluster{
						{
							IAMDatabaseAuthenticationEnabled: ptr.To(false),
						},
					},
				},
				kube: test.NewMockClient(),
			},
			want: want{
				isUpToDate: false,
				err:        nil,
			},
		},
		"IAMDatabaseAuthenticationEnabledNotMapped": {
			args: args{
				cr: &svcapitypes.DBCluster{
					Spec: svcapitypes.DBClusterSpec{
						ForProvider: svcapitypes.DBClusterParameters{
							Engine: ptr.To("aurora-postgresql"),
						},
					},
				},
				out: &svcsdk.DescribeDBClustersOutput{
					DBClusters: []*svcsdk.DBCluster{
						{
							Engine:                           ptr.To("aurora-postgresql"),
							IAMDatabaseAuthenticationEnabled: ptr.To(true),
						},
					},
				},
				kube: test.NewMockClient(),
			},
			want: want{
				isUpToDate: true,
				err:        nil,
			},
		},
		"DifferentPubliclyAccessible": {
			args: args{
				cr: &svcapitypes.DBCluster{
					Spec: svcapitypes.DBClusterSpec{
						ForProvider: svcapitypes.DBClusterParameters{
							PubliclyAccessible: ptr.To(true),
						},
					},
				},
				out: &svcsdk.DescribeDBClustersOutput{
					DBClusters: []*svcsdk.DBCluster{
						{
							PubliclyAccessible: ptr.To(false),
						},
					},
				},
				kube: test.NewMockClient(),
			},
			want: want{
				isUpToDate: false,
				err:        nil,
			},
		},
		"SamePubliclyAccessible": {
			args: args{
				cr: &svcapitypes.DBCluster{
					Spec: svcapitypes.DBClusterSpec{
						ForProvider: svcapitypes.DBClusterParameters{
							PubliclyAccessible: ptr.To(true),
						},
					},
				},
				out: &svcsdk.DescribeDBClustersOutput{
					DBClusters: []*svcsdk.DBCluster{
						{
							PubliclyAccessible: ptr.To(true),
						},
					},
				},
				kube: test.NewMockClient(),
			},
			want: want{
				isUpToDate: true,
				err:        nil,
			},
		},
		"DifferentStorageEncrypted": {
			args: args{
				cr: &svcapitypes.DBCluster{
					Spec: svcapitypes.DBClusterSpec{
						ForProvider: svcapitypes.DBClusterParameters{
							StorageEncrypted: ptr.To(true),
						},
					},
				},
				out: &svcsdk.DescribeDBClustersOutput{
					DBClusters: []*svcsdk.DBCluster{
						{
							StorageEncrypted: ptr.To(false),
						},
					},
				},
				kube: test.NewMockClient(),
			},
			want: want{
				isUpToDate: false,
				err:        nil,
			},
		},
		"SameStorageEncrypted": {
			args: args{
				cr: &svcapitypes.DBCluster{
					Spec: svcapitypes.DBClusterSpec{
						ForProvider: svcapitypes.DBClusterParameters{
							StorageEncrypted: ptr.To(true),
						},
					},
				},
				out: &svcsdk.DescribeDBClustersOutput{
					DBClusters: []*svcsdk.DBCluster{
						{
							StorageEncrypted: ptr.To(true),
						},
					},
				},
				kube: test.NewMockClient(),
			},
			want: want{
				isUpToDate: true,
				err:        nil,
			},
		},
		"DifferentAutoMinorVersionUpgrade": {
			args: args{
				cr: &svcapitypes.DBCluster{
					Spec: svcapitypes.DBClusterSpec{
						ForProvider: svcapitypes.DBClusterParameters{
							AutoMinorVersionUpgrade: ptr.To(false),
						},
					},
				},
				out: &svcsdk.DescribeDBClustersOutput{
					DBClusters: []*svcsdk.DBCluster{
						{
							AutoMinorVersionUpgrade: ptr.To(true),
						},
					},
				},
				kube: test.NewMockClient(),
			},
			want: want{
				isUpToDate: false,
				err:        nil,
			},
		},
		"SameAutoMinorVersionUpgrade": {
			args: args{
				cr: &svcapitypes.DBCluster{
					Spec: svcapitypes.DBClusterSpec{
						ForProvider: svcapitypes.DBClusterParameters{
							AutoMinorVersionUpgrade: ptr.To(false),
						},
					},
				},
				out: &svcsdk.DescribeDBClustersOutput{
					DBClusters: []*svcsdk.DBCluster{
						{
							AutoMinorVersionUpgrade: ptr.To(false),
						},
					},
				},
				kube: test.NewMockClient(),
			},
			want: want{
				isUpToDate: true,
				err:        nil,
			},
		},
		"DifferentEnablePerformanceInsights": {
			args: args{
				cr: &svcapitypes.DBCluster{
					Spec: svcapitypes.DBClusterSpec{
						ForProvider: svcapitypes.DBClusterParameters{
							Engine: ptr.To("mysql"),
						},
					},
				},
				out: &svcsdk.DescribeDBClustersOutput{
					DBClusters: []*svcsdk.DBCluster{
						{
							Engine:                     ptr.To("mysql"),
							PerformanceInsightsEnabled: ptr.To(false),
						},
					},
				},
				kube: test.NewMockClient(),
			},
			want: want{
				isUpToDate: true,
				err:        nil,
			},
		},
		"SameEnablePerformanceInsights": {
			args: args{
				cr: &svcapitypes.DBCluster{
					Spec: svcapitypes.DBClusterSpec{
						ForProvider: svcapitypes.DBClusterParameters{
							Engine: ptr.To("mysql"),
						},
					},
				},
				out: &svcsdk.DescribeDBClustersOutput{
					DBClusters: []*svcsdk.DBCluster{
						{
							Engine:                     ptr.To("mysql"),
							PerformanceInsightsEnabled: ptr.To(true),
						},
					},
				},
				kube: test.NewMockClient(),
			},
			want: want{
				isUpToDate: true,
				err:        nil,
			},
		},
		// Port tests (integer parameter)
		"DifferentPort": {
			args: args{
				cr: &svcapitypes.DBCluster{
					Spec: svcapitypes.DBClusterSpec{
						ForProvider: svcapitypes.DBClusterParameters{
							Port: ptr.To(int64(3307)),
						},
					},
				},
				out: &svcsdk.DescribeDBClustersOutput{
					DBClusters: []*svcsdk.DBCluster{
						{
							Port: ptr.To(int64(3306)),
						},
					},
				},
				kube: test.NewMockClient(),
			},
			want: want{
				isUpToDate: false,
				err:        nil,
			},
		},
		"SamePort": {
			args: args{
				cr: &svcapitypes.DBCluster{
					Spec: svcapitypes.DBClusterSpec{
						ForProvider: svcapitypes.DBClusterParameters{
							Port: ptr.To(int64(3306)),
						},
					},
				},
				out: &svcsdk.DescribeDBClustersOutput{
					DBClusters: []*svcsdk.DBCluster{
						{
							Port: ptr.To(int64(3306)),
						},
					},
				},
				kube: test.NewMockClient(),
			},
			want: want{
				isUpToDate: true,
				err:        nil,
			},
		},
		"PortNotSet": {
			args: args{
				cr: &svcapitypes.DBCluster{
					Spec: svcapitypes.DBClusterSpec{
						ForProvider: svcapitypes.DBClusterParameters{
							Port: nil,
						},
					},
				},
				out: &svcsdk.DescribeDBClustersOutput{
					DBClusters: []*svcsdk.DBCluster{
						{
							Port: ptr.To(int64(3306)),
						},
					},
				},
				kube: test.NewMockClient(),
			},
			want: want{
				isUpToDate: true,
				err:        nil,
			},
		},
		// StorageType tests
		"DifferentStorageType": {
			args: args{
				cr: &svcapitypes.DBCluster{
					Spec: svcapitypes.DBClusterSpec{
						ForProvider: svcapitypes.DBClusterParameters{
							StorageType: ptr.To("aurora-iopt1"),
						},
					},
				},
				out: &svcsdk.DescribeDBClustersOutput{
					DBClusters: []*svcsdk.DBCluster{
						{
							StorageType: ptr.To("aurora"),
						},
					},
				},
				kube: test.NewMockClient(),
			},
			want: want{
				isUpToDate: false,
				err:        nil,
			},
		},
		"SameStorageType": {
			args: args{
				cr: &svcapitypes.DBCluster{
					Spec: svcapitypes.DBClusterSpec{
						ForProvider: svcapitypes.DBClusterParameters{
							StorageType: ptr.To("aurora-iopt1"),
						},
					},
				},
				out: &svcsdk.DescribeDBClustersOutput{
					DBClusters: []*svcsdk.DBCluster{
						{
							StorageType: ptr.To("aurora-iopt1"),
						},
					},
				},
				kube: test.NewMockClient(),
			},
			want: want{
				isUpToDate: true,
				err:        nil,
			},
		},
		"StorageTypeAuroraDefault": {
			args: args{
				cr: &svcapitypes.DBCluster{
					Spec: svcapitypes.DBClusterSpec{
						ForProvider: svcapitypes.DBClusterParameters{
							StorageType: ptr.To("aurora"),
						},
					},
				},
				out: &svcsdk.DescribeDBClustersOutput{
					DBClusters: []*svcsdk.DBCluster{
						{
							StorageType: ptr.To(""),
						},
					},
				},
				kube: test.NewMockClient(),
			},
			want: want{
				isUpToDate: true,
				err:        nil,
			},
		},
		"DifferentCopyTagsToSnapshot": {
			args: args{
				cr: &svcapitypes.DBCluster{
					Spec: svcapitypes.DBClusterSpec{
						ForProvider: svcapitypes.DBClusterParameters{
							Engine:             ptr.To("aurora-mysql"),
							CopyTagsToSnapshot: ptr.To(true),
						},
					},
				},
				out: &svcsdk.DescribeDBClustersOutput{
					DBClusters: []*svcsdk.DBCluster{
						{
							Engine:             ptr.To("aurora-mysql"),
							CopyTagsToSnapshot: ptr.To(false),
						},
					},
				},
				kube: test.NewMockClient(),
			},
			want: want{
				isUpToDate: false,
				err:        nil,
			},
		},
		"SameCopyTagsToSnapshot": {
			args: args{
				cr: &svcapitypes.DBCluster{
					Spec: svcapitypes.DBClusterSpec{
						ForProvider: svcapitypes.DBClusterParameters{
							CopyTagsToSnapshot: ptr.To(true),
						},
					},
				},
				out: &svcsdk.DescribeDBClustersOutput{
					DBClusters: []*svcsdk.DBCluster{
						{
							CopyTagsToSnapshot: ptr.To(true),
						},
					},
				},
				kube: test.NewMockClient(),
			},
			want: want{
				isUpToDate: true,
				err:        nil,
			},
		},
		// DeletionProtection tests
		"DifferentDeletionProtection": {
			args: args{
				cr: &svcapitypes.DBCluster{
					Spec: svcapitypes.DBClusterSpec{
						ForProvider: svcapitypes.DBClusterParameters{
							DeletionProtection: ptr.To(true),
						},
					},
				},
				out: &svcsdk.DescribeDBClustersOutput{
					DBClusters: []*svcsdk.DBCluster{
						{
							DeletionProtection: ptr.To(false),
						},
					},
				},
				kube: test.NewMockClient(),
			},
			want: want{
				isUpToDate: false,
				err:        nil,
			},
		},
		"SameDeletionProtection": {
			args: args{
				cr: &svcapitypes.DBCluster{
					Spec: svcapitypes.DBClusterSpec{
						ForProvider: svcapitypes.DBClusterParameters{
							DeletionProtection: ptr.To(true),
						},
					},
				},
				out: &svcsdk.DescribeDBClustersOutput{
					DBClusters: []*svcsdk.DBCluster{
						{
							DeletionProtection: ptr.To(true),
						},
					},
				},
				kube: test.NewMockClient(),
			},
			want: want{
				isUpToDate: true,
				err:        nil,
			},
		},

		"DifferentIOPS": {
			args: args{
				cr: &svcapitypes.DBCluster{
					Spec: svcapitypes.DBClusterSpec{
						ForProvider: svcapitypes.DBClusterParameters{
							Engine: ptr.To("mysql"),
							IOPS:   ptr.To(int64(3000)),
						},
					},
				},
				out: &svcsdk.DescribeDBClustersOutput{
					DBClusters: []*svcsdk.DBCluster{
						{
							Engine: ptr.To("mysql"),
							Iops:   ptr.To(int64(1000)),
						},
					},
				},
				kube: test.NewMockClient(),
			},
			want: want{
				isUpToDate: false,
				err:        nil,
			},
		},
		"SameIOPS": {
			args: args{
				cr: &svcapitypes.DBCluster{
					Spec: svcapitypes.DBClusterSpec{
						ForProvider: svcapitypes.DBClusterParameters{
							IOPS: ptr.To(int64(3000)),
						},
					},
				},
				out: &svcsdk.DescribeDBClustersOutput{
					DBClusters: []*svcsdk.DBCluster{
						{
							Iops: ptr.To(int64(3000)),
						},
					},
				},
				kube: test.NewMockClient(),
			},
			want: want{
				isUpToDate: true,
				err:        nil,
			},
		},
		// AllocatedStorage tests
		"DifferentAllocatedStorage": {
			args: args{
				cr: &svcapitypes.DBCluster{
					Spec: svcapitypes.DBClusterSpec{
						ForProvider: svcapitypes.DBClusterParameters{
							AllocatedStorage: ptr.To(int64(100)),
						},
					},
				},
				out: &svcsdk.DescribeDBClustersOutput{
					DBClusters: []*svcsdk.DBCluster{
						{
							AllocatedStorage: ptr.To(int64(50)),
						},
					},
				},
				kube: test.NewMockClient(),
			},
			want: want{
				isUpToDate: false,
				err:        nil,
			},
		},
		// Engine tests (string parameter)
		"DifferentEngine": {
			args: args{
				cr: &svcapitypes.DBCluster{
					Spec: svcapitypes.DBClusterSpec{
						ForProvider: svcapitypes.DBClusterParameters{
							Engine: ptr.To("aurora-mysql"),
						},
					},
				},
				out: &svcsdk.DescribeDBClustersOutput{
					DBClusters: []*svcsdk.DBCluster{
						{
							Engine: ptr.To("aurora-postgresql"),
						},
					},
				},
				kube: test.NewMockClient(),
			},
			want: want{
				isUpToDate: false,
				err:        nil,
			},
		},
		// KMSKeyID tests
		"DifferentKMSKeyID": {
			args: args{
				cr: &svcapitypes.DBCluster{
					Spec: svcapitypes.DBClusterSpec{
						ForProvider: svcapitypes.DBClusterParameters{
							KMSKeyID: ptr.To("arn:aws:kms:us-east-1:123456789012:key/12345678-1234-1234-1234-123456789012"),
						},
					},
				},
				out: &svcsdk.DescribeDBClustersOutput{
					DBClusters: []*svcsdk.DBCluster{
						{
							KmsKeyId: ptr.To("arn:aws:kms:us-east-1:123456789012:key/87654321-4321-4321-4321-210987654321"),
						},
					},
				},
				kube: test.NewMockClient(),
			},
			want: want{
				isUpToDate: false,
				err:        nil,
			},
		},
		// MasterUsername tests
		"DifferentMasterUsername": {
			args: args{
				cr: &svcapitypes.DBCluster{
					Spec: svcapitypes.DBClusterSpec{
						ForProvider: svcapitypes.DBClusterParameters{
							MasterUsername: ptr.To("admin"),
						},
					},
				},
				out: &svcsdk.DescribeDBClustersOutput{
					DBClusters: []*svcsdk.DBCluster{
						{
							MasterUsername: ptr.To("root"),
						},
					},
				},
				kube: test.NewMockClient(),
			},
			want: want{
				isUpToDate: false,
				err:        nil,
			},
		},
		// EngineMode tests
		"DifferentEngineMode": {
			args: args{
				cr: &svcapitypes.DBCluster{
					Spec: svcapitypes.DBClusterSpec{
						ForProvider: svcapitypes.DBClusterParameters{
							EngineMode: ptr.To("provisioned"),
						},
					},
				},
				out: &svcsdk.DescribeDBClustersOutput{
					DBClusters: []*svcsdk.DBCluster{
						{
							EngineMode: ptr.To("serverless"),
						},
					},
				},
				kube: test.NewMockClient(),
			},
			want: want{
				isUpToDate: false,
				err:        nil,
			},
		},

		// NetworkType tests
		"DifferentNetworkType": {
			args: args{
				cr: &svcapitypes.DBCluster{
					Spec: svcapitypes.DBClusterSpec{
						ForProvider: svcapitypes.DBClusterParameters{
							NetworkType: ptr.To("DUAL"),
						},
					},
				},
				out: &svcsdk.DescribeDBClustersOutput{
					DBClusters: []*svcsdk.DBCluster{
						{
							NetworkType: ptr.To("IPV4"),
						},
					},
				},
				kube: test.NewMockClient(),
			},
			want: want{
				isUpToDate: false,
				err:        nil,
			},
		},

		// CharacterSetName tests
		"DifferentCharacterSetName": {
			args: args{
				cr: &svcapitypes.DBCluster{
					Spec: svcapitypes.DBClusterSpec{
						ForProvider: svcapitypes.DBClusterParameters{
							CharacterSetName: ptr.To("UTF8"),
						},
					},
				},
				out: &svcsdk.DescribeDBClustersOutput{
					DBClusters: []*svcsdk.DBCluster{
						{
							CharacterSetName: ptr.To("LATIN1"),
						},
					},
				},
				kube: test.NewMockClient(),
			},
			want: want{
				isUpToDate: false,
				err:        nil,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			custom := shared{kube: tc.args.kube, cache: &cache{}}
			isUpToDate, diffMsg, err := custom.isUpToDate(context.Background(), tc.args.cr, tc.args.out)
			if diff := cmp.Diff(tc.want.isUpToDate, isUpToDate); diff != "" {
				t.Errorf("r: -want, +got:\n%s\ndiff message: %s", diff, diffMsg)
			}
			if diff := cmp.Diff(tc.want.err, err); diff != "" {
				t.Errorf("error: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGenerateRestoreDBClusterToPointInTimeInput(t *testing.T) {
	type args struct {
		cr *svcapitypes.DBCluster
	}

	type want struct {
		restoreType *string
	}

	cases := map[string]struct {
		args
		want
	}{
		"RestoreTypeSet": {
			args: args{
				cr: &svcapitypes.DBCluster{
					Spec: svcapitypes.DBClusterSpec{
						ForProvider: svcapitypes.DBClusterParameters{
							CustomDBClusterParameters: svcapitypes.CustomDBClusterParameters{
								RestoreFrom: &svcapitypes.RestoreDBClusterBackupConfiguration{
									PointInTime: &svcapitypes.PointInTimeRestoreDBClusterBackupConfiguration{
										RestoreType:               ptr.To("copy-on-write"),
										SourceDBClusterIdentifier: ptr.To("source-cluster"),
									},
								},
							},
						},
					},
				},
			},
			want: want{
				restoreType: ptr.To("copy-on-write"),
			},
		},
		"RestoreTypeNotSet": {
			args: args{
				cr: &svcapitypes.DBCluster{
					Spec: svcapitypes.DBClusterSpec{
						ForProvider: svcapitypes.DBClusterParameters{
							CustomDBClusterParameters: svcapitypes.CustomDBClusterParameters{
								RestoreFrom: &svcapitypes.RestoreDBClusterBackupConfiguration{
									PointInTime: &svcapitypes.PointInTimeRestoreDBClusterBackupConfiguration{
										SourceDBClusterIdentifier: ptr.To("source-cluster"),
									},
								},
							},
						},
					},
				},
			},
			want: want{
				restoreType: nil,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			result := generateRestoreDBClusterToPointInTimeInput(tc.args.cr)
			if diff := cmp.Diff(tc.want.restoreType, result.RestoreType); diff != "" {
				t.Errorf("RestoreType: -want, +got:\n%s", diff)
			}
		})
	}
}
