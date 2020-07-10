package elasticache

import (
	"testing"

	awscache "github.com/aws/aws-sdk-go-v2/service/elasticache"
	"github.com/google/go-cmp/cmp"

	"github.com/crossplane/provider-aws/apis/cache/v1alpha1"
	aws "github.com/crossplane/provider-aws/pkg/clients"
)

var (
	clusterID          = "someID"
	nodeType           = "t2.small"
	subnetGroup        = "someSubnetGroup"
	redisEngine        = "redis"
	az                 = "us-east-1a"
	friday             = "friday"
	replicationGroupID = "some-replication-group"
	timeWindow         = "05:00-09:00"
	boolTrue           = true
)

func clusterParams(m ...func(*v1alpha1.CacheClusterParameters)) *v1alpha1.CacheClusterParameters {
	o := &v1alpha1.CacheClusterParameters{
		CacheNodeType:              nodeType,
		CacheSubnetGroupName:       aws.String(subnetGroup),
		Engine:                     aws.String(redisEngine),
		NumCacheNodes:              2,
		PreferredAvailabilityZone:  aws.String(az),
		PreferredMaintenanceWindow: aws.String(friday),
		ReplicationGroupID:         aws.String(replicationGroupID),
		SnapshotRetentionLimit:     aws.Int64(5),
		SnapshotWindow:             aws.String(timeWindow),
	}

	for _, f := range m {
		f(o)
	}

	return o
}

func cluster(m ...func(*awscache.CacheCluster)) *awscache.CacheCluster {
	o := &awscache.CacheCluster{
		AtRestEncryptionEnabled:    &boolTrue,
		AuthTokenEnabled:           &boolTrue,
		CacheClusterStatus:         aws.String(v1alpha1.StatusAvailable),
		CacheClusterId:             aws.String(clusterID),
		CacheNodeType:              aws.String(nodeType),
		CacheSubnetGroupName:       aws.String(subnetGroup),
		Engine:                     aws.String(redisEngine),
		NumCacheNodes:              aws.Int64(2),
		PreferredMaintenanceWindow: aws.String(friday),
		PreferredAvailabilityZone:  aws.String(az),
		ReplicationGroupId:         aws.String(replicationGroupID),
		SnapshotWindow:             aws.String(timeWindow),
		SnapshotRetentionLimit:     aws.Int64(5),
	}

	for _, f := range m {
		f(o)
	}

	return o
}

func TestLateInitializeCluster(t *testing.T) {
	type args struct {
		spec *v1alpha1.CacheClusterParameters
		in   awscache.CacheCluster
	}
	cases := map[string]struct {
		args args
		want *v1alpha1.CacheClusterParameters
	}{
		"AllFilledNoDiff": {
			args: args{
				spec: clusterParams(),
				in:   *cluster(),
			},
			want: clusterParams(),
		},
		"AllFilledExternalDiff": {
			args: args{
				spec: clusterParams(),
				in: *cluster(func(r *awscache.CacheCluster) {
					r.ReplicationGroupId = nil
				}),
			},
			want: clusterParams(),
		},
		"PartialFilled": {
			args: args{
				spec: clusterParams(func(p *v1alpha1.CacheClusterParameters) {
					p.ReplicationGroupID = nil
				}),
				in: *cluster(),
			},
			want: clusterParams(func(p *v1alpha1.CacheClusterParameters) {
				p.ReplicationGroupID = aws.String(replicationGroupID)
			}),
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			LateInitializeCluster(tc.args.spec, tc.args.in)
			if diff := cmp.Diff(tc.args.spec, tc.want); diff != "" {
				t.Errorf("LateInitializeSpec(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGenerateCreateCacheClusterInput(t *testing.T) {
	cases := map[string]struct {
		in  v1alpha1.CacheClusterParameters
		out awscache.CreateCacheClusterInput
	}{
		"FilledInput": {
			in: *clusterParams(),
			out: awscache.CreateCacheClusterInput{
				CacheClusterId:             &clusterID,
				CacheNodeType:              aws.String(nodeType),
				CacheSubnetGroupName:       aws.String(subnetGroup),
				Engine:                     aws.String(redisEngine),
				NumCacheNodes:              aws.Int64(2),
				PreferredAvailabilityZone:  aws.String(az),
				PreferredMaintenanceWindow: aws.String(friday),
				ReplicationGroupId:         aws.String(replicationGroupID),
				SnapshotRetentionLimit:     aws.Int64(5),
				SnapshotWindow:             aws.String(timeWindow),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			r := GenerateCreateCacheClusterInput(tc.in, clusterID)
			if diff := cmp.Diff(r, &tc.out); diff != "" {
				t.Errorf("GenerateNetworkObservation(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGenerateModifyCacheClusterInput(t *testing.T) {
	cases := map[string]struct {
		in  v1alpha1.CacheClusterParameters
		out awscache.ModifyCacheClusterInput
	}{
		"FilledInput": {
			in: *clusterParams(),
			out: awscache.ModifyCacheClusterInput{
				CacheClusterId:             &clusterID,
				CacheNodeType:              aws.String(nodeType),
				NumCacheNodes:              aws.Int64(2),
				PreferredMaintenanceWindow: aws.String(friday),
				SnapshotRetentionLimit:     aws.Int64(5),
				SnapshotWindow:             aws.String(timeWindow),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			r := GenerateModifyCacheClusterInput(tc.in, clusterID)
			if diff := cmp.Diff(r, &tc.out); diff != "" {
				t.Errorf("GenerateNetworkObservation(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestIsClusterUpToDate(t *testing.T) {
	type args struct {
		c awscache.CacheCluster
		p v1alpha1.CacheClusterParameters
	}

	cases := map[string]struct {
		args args
		want bool
	}{
		"SameFields": {
			args: args{
				c: *cluster(),
				p: *clusterParams(),
			},
			want: true,
		},
		"DifferentFields": {
			args: args{
				c: *cluster(),
				p: *clusterParams(func(c *v1alpha1.CacheClusterParameters) {
					c.CacheNodeType = "t2.large"
				}),
			},
			want: false,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got, _ := IsClusterUpToDate(clusterID, &tc.args.p, &tc.args.c)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGenerateClusterObservation(t *testing.T) {
	cases := map[string]struct {
		in  awscache.CacheCluster
		out v1alpha1.CacheClusterObservation
	}{
		"AllFilled": {
			in: *cluster(),
			out: v1alpha1.CacheClusterObservation{
				AtRestEncryptionEnabled: boolTrue,
				AuthTokenEnabled:        boolTrue,
				CacheClusterStatus:      v1alpha1.StatusAvailable,
			},
		},
		"CacheNodes": {
			in: *cluster(func(c *awscache.CacheCluster) {
				c.CacheNodes = []awscache.CacheNode{
					{
						CacheNodeStatus: aws.String(v1alpha1.StatusAvailable),
					},
				}
			}),
			out: v1alpha1.CacheClusterObservation{
				AtRestEncryptionEnabled: boolTrue,
				AuthTokenEnabled:        boolTrue,
				CacheClusterStatus:      v1alpha1.StatusAvailable,
				CacheNodes: []v1alpha1.CacheNode{{
					CacheNodeStatus: v1alpha1.StatusAvailable,
				}},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			r := GenerateClusterObservation(tc.in)
			if diff := cmp.Diff(r, tc.out); diff != "" {
				t.Errorf("GenerateNetworkObservation(...): -want, +got:\n%s", diff)
			}
		})
	}
}
