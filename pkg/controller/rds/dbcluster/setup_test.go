package dbcluster

import (
	"context"
	_ "embed"
	"testing"

	svcsdk "github.com/aws/aws-sdk-go/service/rds"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/rds/v1alpha1"
)

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
		"Nil": {
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
							EngineVersion: nil,
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
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			custom := custom{kube: tc.args.kube}
			isUpToDate, _, err := custom.isUpToDate(context.Background(), tc.args.cr, tc.args.out)
			if diff := cmp.Diff(tc.want.isUpToDate, isUpToDate); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.err, err); diff != "" {
				t.Errorf("error: -want, +got:\n%s", diff)
			}
		})
	}
}
