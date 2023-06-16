package dbcluster

import (
	"testing"

	svcsdk "github.com/aws/aws-sdk-go/service/rds"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/rds/v1alpha1"

	"github.com/google/go-cmp/cmp"
)

func ptr(str string) *string {
	return &str
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
									VpcSecurityGroupId: ptr("sg-123"),
								},
								{
									VpcSecurityGroupId: ptr("sg-456"),
								},
								{
									VpcSecurityGroupId: ptr("sg-789"),
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
									VpcSecurityGroupId: ptr("sg-456"),
								},
								{
									VpcSecurityGroupId: ptr("sg-123"),
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
									VpcSecurityGroupId: ptr("sg-456"),
								},
								{
									VpcSecurityGroupId: ptr("sg-123"),
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
									VpcSecurityGroupId: ptr("sg-123"),
								},
								{
									VpcSecurityGroupId: ptr("sg-456"),
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
							EngineVersion: ptr("12.3"), // some AWS "default" value
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
								EngineVersion: ptr("12.3"),
							},
						},
					},
				},
				out: &svcsdk.DescribeDBClustersOutput{
					DBClusters: []*svcsdk.DBCluster{
						{
							EngineVersion: ptr("12.3"), // some AWS "default" value
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
								EngineVersion: ptr("12"),
							},
						},
					},
				},
				out: &svcsdk.DescribeDBClustersOutput{
					DBClusters: []*svcsdk.DBCluster{
						{
							EngineVersion: ptr("12.3"), // some AWS "default" value
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
								EngineVersion: ptr("12.1"),
							},
						},
					},
				},
				out: &svcsdk.DescribeDBClustersOutput{
					DBClusters: []*svcsdk.DBCluster{
						{
							EngineVersion: ptr("12.3"), // some AWS "default" value
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
								EngineVersion: ptr("13"),
							},
						},
					},
				},
				out: &svcsdk.DescribeDBClustersOutput{
					DBClusters: []*svcsdk.DBCluster{
						{
							EngineVersion: ptr("12.3"),
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
							DBClusterParameterGroup: ptr("default.aurora-postgresql14"), // some AWS "default" value
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
							DBClusterParameterGroupName: ptr("default.aurora-postgresql14"),
						},
					},
				},
				out: &svcsdk.DescribeDBClustersOutput{
					DBClusters: []*svcsdk.DBCluster{
						{
							DBClusterParameterGroup: ptr("default.aurora-postgresql14"), // some AWS "default" value
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
							DBClusterParameterGroupName: ptr("default.aurora-postgresql15"),
						},
					},
				},
				out: &svcsdk.DescribeDBClustersOutput{
					DBClusters: []*svcsdk.DBCluster{
						{
							DBClusterParameterGroup: ptr("default.aurora-postgresql14"),
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
