/*
Copyright 2020 The Crossplane Authors.

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

package elasticache

import (
	"testing"

	awscache "github.com/aws/aws-sdk-go-v2/service/elasticache"
	awscachetypes "github.com/aws/aws-sdk-go-v2/service/elasticache/types"
	"github.com/aws/smithy-go/document"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/crossplane-contrib/provider-aws/apis/cache/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
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
		CacheSubnetGroupName:       pointer.ToOrNilIfZeroValue(subnetGroup),
		Engine:                     pointer.ToOrNilIfZeroValue(redisEngine),
		NumCacheNodes:              2,
		PreferredAvailabilityZone:  pointer.ToOrNilIfZeroValue(az),
		PreferredMaintenanceWindow: pointer.ToOrNilIfZeroValue(friday),
		ReplicationGroupID:         pointer.ToOrNilIfZeroValue(replicationGroupID),
		SnapshotRetentionLimit:     pointer.ToIntAsInt32(5),
		SnapshotWindow:             pointer.ToOrNilIfZeroValue(timeWindow),
	}

	for _, f := range m {
		f(o)
	}

	return o
}

func cluster(m ...func(*awscachetypes.CacheCluster)) *awscachetypes.CacheCluster {
	o := &awscachetypes.CacheCluster{
		AtRestEncryptionEnabled:    &boolTrue,
		AuthTokenEnabled:           &boolTrue,
		CacheClusterStatus:         pointer.ToOrNilIfZeroValue(v1alpha1.StatusAvailable),
		CacheClusterId:             pointer.ToOrNilIfZeroValue(clusterID),
		CacheNodeType:              pointer.ToOrNilIfZeroValue(nodeType),
		CacheSubnetGroupName:       pointer.ToOrNilIfZeroValue(subnetGroup),
		Engine:                     pointer.ToOrNilIfZeroValue(redisEngine),
		NumCacheNodes:              pointer.ToIntAsInt32(2),
		PreferredMaintenanceWindow: pointer.ToOrNilIfZeroValue(friday),
		PreferredAvailabilityZone:  pointer.ToOrNilIfZeroValue(az),
		ReplicationGroupId:         pointer.ToOrNilIfZeroValue(replicationGroupID),
		SnapshotWindow:             pointer.ToOrNilIfZeroValue(timeWindow),
		SnapshotRetentionLimit:     pointer.ToIntAsInt32(5),
	}

	for _, f := range m {
		f(o)
	}

	return o
}

func TestLateInitializeCluster(t *testing.T) {
	type args struct {
		spec *v1alpha1.CacheClusterParameters
		in   awscachetypes.CacheCluster
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
				in: *cluster(func(r *awscachetypes.CacheCluster) {
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
				p.ReplicationGroupID = pointer.ToOrNilIfZeroValue(replicationGroupID)
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
				CacheNodeType:              pointer.ToOrNilIfZeroValue(nodeType),
				CacheSubnetGroupName:       pointer.ToOrNilIfZeroValue(subnetGroup),
				Engine:                     pointer.ToOrNilIfZeroValue(redisEngine),
				NumCacheNodes:              pointer.ToIntAsInt32(2),
				PreferredAvailabilityZone:  pointer.ToOrNilIfZeroValue(az),
				PreferredMaintenanceWindow: pointer.ToOrNilIfZeroValue(friday),
				ReplicationGroupId:         pointer.ToOrNilIfZeroValue(replicationGroupID),
				SnapshotRetentionLimit:     pointer.ToIntAsInt32(5),
				SnapshotWindow:             pointer.ToOrNilIfZeroValue(timeWindow),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			r, _ := GenerateCreateCacheClusterInput(tc.in, clusterID)
			if diff := cmp.Diff(r, &tc.out, cmpopts.IgnoreTypes(document.NoSerde{})); diff != "" {
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
				CacheNodeType:              pointer.ToOrNilIfZeroValue(nodeType),
				NumCacheNodes:              pointer.ToIntAsInt32(2),
				PreferredMaintenanceWindow: pointer.ToOrNilIfZeroValue(friday),
				SnapshotRetentionLimit:     pointer.ToIntAsInt32(5),
				SnapshotWindow:             pointer.ToOrNilIfZeroValue(timeWindow),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			r, _ := GenerateModifyCacheClusterInput(tc.in, clusterID)
			if diff := cmp.Diff(r, &tc.out, cmpopts.IgnoreTypes(document.NoSerde{})); diff != "" {
				t.Errorf("GenerateNetworkObservation(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestIsClusterUpToDate(t *testing.T) {
	type args struct {
		c awscachetypes.CacheCluster
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
		in  awscachetypes.CacheCluster
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
			in: *cluster(func(c *awscachetypes.CacheCluster) {
				c.CacheNodes = []awscachetypes.CacheNode{
					{
						CacheNodeStatus: pointer.ToOrNilIfZeroValue(v1alpha1.StatusAvailable),
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
