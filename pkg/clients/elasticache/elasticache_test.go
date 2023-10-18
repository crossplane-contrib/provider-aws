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

package elasticache

import (
	"sort"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/elasticache"
	elasticachetypes "github.com/aws/aws-sdk-go-v2/service/elasticache/types"
	"github.com/aws/smithy-go/document"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/pkg/errors"
	"k8s.io/utils/ptr"

	cachev1alpha1 "github.com/crossplane-contrib/provider-aws/apis/cache/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/apis/cache/v1beta1"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
)

const (
	namespace = "coolNamespace"
	name      = "coolGroup"
)

var (
	cacheNodeType             = "n1.super.cool"
	atRestEncryptionEnabled   = true
	authEnabled               = true
	authToken                 = "coolToken"
	autoFailoverEnabled       = true
	cacheParameterGroupName   = "coolParamGroup"
	cacheSubnetGroupName      = "coolSubnet"
	engine                    = "redis"
	engineVersion             = "5.0.0"
	multiAZ                   = true
	notificationTopicARN      = "arn:aws:sns:cooltopic"
	notificationTopicStatus   = "active"
	numCacheClusters          = 1
	numNodeGroups             = 2
	host                      = "coolhost"
	port                      = 6379
	primaryClusterID          = "the-coolest-one"
	maintenanceWindow         = "tomorrow"
	replicasPerNodeGroup      = 2
	snapshotName              = "coolSnapshot"
	snapshotRetentionLimit    = 1
	newSnapshotRetentionLimit = 2
	snapshottingClusterID     = "snapshot-cluster"
	snapshotWindow            = "thedayaftertomorrow"
	tagKey                    = "key-1"
	tagValue                  = "value-1"
	transitEncryptionEnabled  = true

	nodeGroupPrimaryAZ    = "us-cool-1a"
	nodeGroupReplicaCount = 2
	nodeGroupSlots        = "coolslots"

	cacheClusterID = name + "-0001"
)

var (
	cacheSecurityGroupNames  = []string{"coolGroup", "coolerGroup"}
	preferredCacheClusterAZs = []string{"us-cool-1a", "us-cool-1b"}
	securityGroupIDs         = []string{"coolID", "coolerID"}
	snapshotARNs             = []string{"arn:aws:s3:snappy"}

	description = "Crossplane managed " + v1beta1.ReplicationGroupKindAPIVersion + " " + namespace + "/" + name

	nodeGroupAZs = []string{"us-cool-1a", "us-cool-1b"}
)

var (
	subnetGroupDesc = "some description"
	subnetID1       = "subnetId1"
	subnetID2       = "subnetId2"
)

func replicationGroupParams(m ...func(parameters *v1beta1.ReplicationGroupParameters)) *v1beta1.ReplicationGroupParameters {
	o := &v1beta1.ReplicationGroupParameters{
		ApplyModificationsImmediately: true,
		AtRestEncryptionEnabled:       &atRestEncryptionEnabled,
		AuthEnabled:                   &authEnabled,
		AutomaticFailoverEnabled:      &autoFailoverEnabled,
		CacheNodeType:                 cacheNodeType,
		CacheParameterGroupName:       &cacheParameterGroupName,
		CacheSecurityGroupNames:       cacheSecurityGroupNames,
		CacheSubnetGroupName:          &cacheSubnetGroupName,
		Engine:                        engine,
		EngineVersion:                 &engineVersion,
		MultiAZEnabled:                &multiAZ,
		NodeGroupConfiguration: []v1beta1.NodeGroupConfigurationSpec{
			{
				PrimaryAvailabilityZone:  &nodeGroupPrimaryAZ,
				ReplicaAvailabilityZones: nodeGroupAZs,
				ReplicaCount:             &nodeGroupReplicaCount,
				Slots:                    &nodeGroupSlots,
			},
		},
		NotificationTopicARN:        &notificationTopicARN,
		NotificationTopicStatus:     &notificationTopicStatus,
		NumCacheClusters:            &numCacheClusters,
		NumNodeGroups:               &numNodeGroups,
		PrimaryClusterID:            &primaryClusterID,
		Port:                        &port,
		PreferredCacheClusterAZs:    preferredCacheClusterAZs,
		PreferredMaintenanceWindow:  &maintenanceWindow,
		ReplicasPerNodeGroup:        &replicasPerNodeGroup,
		ReplicationGroupDescription: description,
		SecurityGroupIDs:            securityGroupIDs,
		SnapshotARNs:                snapshotARNs,
		SnapshotName:                &snapshotName,
		SnapshotRetentionLimit:      &snapshotRetentionLimit,
		SnapshottingClusterID:       &snapshottingClusterID,
		SnapshotWindow:              &snapshotWindow,
		Tags: []v1beta1.Tag{
			{
				Key:   tagKey,
				Value: tagValue,
			},
		},
		TransitEncryptionEnabled: &transitEncryptionEnabled,
	}

	for _, f := range m {
		f(o)
	}

	return o
}

func TestNewCreateReplicationGroupInput(t *testing.T) {
	cases := []struct {
		name      string
		params    *v1beta1.ReplicationGroupParameters
		authToken *string
		want      *elasticache.CreateReplicationGroupInput
	}{
		{
			name:      "AllPossibleFields",
			params:    replicationGroupParams(),
			authToken: &authToken,
			want: &elasticache.CreateReplicationGroupInput{
				ReplicationGroupId:          ptr.To(name),
				ReplicationGroupDescription: ptr.To(description),
				Engine:                      ptr.To(v1beta1.CacheEngineRedis),
				CacheNodeType:               ptr.To(cacheNodeType),
				AtRestEncryptionEnabled:     pointer.ToOrNilIfZeroValue(atRestEncryptionEnabled),
				AuthToken:                   pointer.ToOrNilIfZeroValue(authToken),
				AutomaticFailoverEnabled:    pointer.ToOrNilIfZeroValue(autoFailoverEnabled),
				CacheParameterGroupName:     pointer.ToOrNilIfZeroValue(cacheParameterGroupName),
				CacheSecurityGroupNames:     cacheSecurityGroupNames,
				CacheSubnetGroupName:        pointer.ToOrNilIfZeroValue(cacheSubnetGroupName),
				EngineVersion:               pointer.ToOrNilIfZeroValue(engineVersion),
				MultiAZEnabled:              pointer.ToOrNilIfZeroValue(multiAZ),
				NodeGroupConfiguration: []elasticachetypes.NodeGroupConfiguration{
					{
						PrimaryAvailabilityZone:  pointer.ToOrNilIfZeroValue(nodeGroupPrimaryAZ),
						ReplicaAvailabilityZones: nodeGroupAZs,
						ReplicaCount:             pointer.ToIntAsInt32Ptr(&nodeGroupReplicaCount),
						Slots:                    pointer.ToOrNilIfZeroValue(nodeGroupSlots),
					},
				},
				NotificationTopicArn:       pointer.ToOrNilIfZeroValue(notificationTopicARN),
				NumCacheClusters:           pointer.ToIntAsInt32Ptr(&numCacheClusters),
				NumNodeGroups:              pointer.ToIntAsInt32Ptr(&numNodeGroups),
				Port:                       pointer.ToIntAsInt32Ptr(&port),
				PreferredCacheClusterAZs:   preferredCacheClusterAZs,
				PreferredMaintenanceWindow: pointer.ToOrNilIfZeroValue(maintenanceWindow),
				PrimaryClusterId:           pointer.ToOrNilIfZeroValue(primaryClusterID),
				ReplicasPerNodeGroup:       pointer.ToIntAsInt32Ptr(&replicasPerNodeGroup),
				SecurityGroupIds:           securityGroupIDs,
				SnapshotArns:               snapshotARNs,
				SnapshotName:               pointer.ToOrNilIfZeroValue(snapshotName),
				SnapshotRetentionLimit:     pointer.ToIntAsInt32Ptr(&snapshotRetentionLimit),
				SnapshotWindow:             pointer.ToOrNilIfZeroValue(snapshotWindow),
				Tags: []elasticachetypes.Tag{
					{
						Key:   &tagKey,
						Value: &tagValue,
					},
				},
				TransitEncryptionEnabled: pointer.ToOrNilIfZeroValue(transitEncryptionEnabled),
			},
		},
		{
			name: "UnsetFieldsAreNilNotZeroType",
			params: &v1beta1.ReplicationGroupParameters{
				CacheNodeType:               cacheNodeType,
				ReplicationGroupDescription: description,
				Engine:                      engine,
			},
			want: &elasticache.CreateReplicationGroupInput{
				ReplicationGroupId:          ptr.To(name),
				ReplicationGroupDescription: ptr.To(description),
				Engine:                      ptr.To(engine),
				CacheNodeType:               ptr.To(cacheNodeType),
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := NewCreateReplicationGroupInput(*tc.params, name, tc.authToken)

			if diff := cmp.Diff(tc.want, got, cmpopts.IgnoreTypes(document.NoSerde{})); diff != "" {
				t.Errorf("NewCreateReplicationGroupInput(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestNewModifyReplicationGroupInput(t *testing.T) {
	cases := []struct {
		name   string
		params *v1beta1.ReplicationGroupParameters
		want   *elasticache.ModifyReplicationGroupInput
	}{
		{
			name:   "AllPossibleFields",
			params: replicationGroupParams(),
			want: &elasticache.ModifyReplicationGroupInput{
				ReplicationGroupId:          ptr.To(name),
				ApplyImmediately:            true,
				AutomaticFailoverEnabled:    pointer.ToOrNilIfZeroValue(autoFailoverEnabled),
				CacheNodeType:               pointer.ToOrNilIfZeroValue(cacheNodeType),
				CacheParameterGroupName:     pointer.ToOrNilIfZeroValue(cacheParameterGroupName),
				CacheSecurityGroupNames:     cacheSecurityGroupNames,
				EngineVersion:               pointer.ToOrNilIfZeroValue(engineVersion),
				MultiAZEnabled:              pointer.ToOrNilIfZeroValue(true),
				NotificationTopicArn:        pointer.ToOrNilIfZeroValue(notificationTopicARN),
				NotificationTopicStatus:     pointer.ToOrNilIfZeroValue(notificationTopicStatus),
				PreferredMaintenanceWindow:  pointer.ToOrNilIfZeroValue(maintenanceWindow),
				PrimaryClusterId:            pointer.ToOrNilIfZeroValue(primaryClusterID),
				ReplicationGroupDescription: pointer.ToOrNilIfZeroValue(description),
				SecurityGroupIds:            securityGroupIDs,
				SnapshotRetentionLimit:      pointer.ToIntAsInt32Ptr(&snapshotRetentionLimit),
				SnapshotWindow:              pointer.ToOrNilIfZeroValue(snapshotWindow),
				SnapshottingClusterId:       pointer.ToOrNilIfZeroValue(snapshottingClusterID),
			},
		},
		{
			name: "UnsetFieldsAreNilNotZeroType",
			params: &v1beta1.ReplicationGroupParameters{
				CacheNodeType:               cacheNodeType,
				ReplicationGroupDescription: description,
			},
			want: &elasticache.ModifyReplicationGroupInput{
				ApplyImmediately:            false,
				ReplicationGroupId:          ptr.To(name),
				ReplicationGroupDescription: ptr.To(description),
				CacheNodeType:               ptr.To(cacheNodeType),
			},
		},
		{
			name: "SuperfluousFields",
			params: &v1beta1.ReplicationGroupParameters{
				AtRestEncryptionEnabled:     &atRestEncryptionEnabled,
				CacheNodeType:               cacheNodeType,
				ReplicationGroupDescription: description,
			},
			want: &elasticache.ModifyReplicationGroupInput{
				ApplyImmediately:            false,
				ReplicationGroupId:          ptr.To(name),
				ReplicationGroupDescription: ptr.To(description),
				CacheNodeType:               ptr.To(cacheNodeType),
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := NewModifyReplicationGroupInput(*tc.params, name)

			if diff := cmp.Diff(tc.want, got, cmpopts.IgnoreTypes(document.NoSerde{})); diff != "" {
				t.Errorf("NewModifyReplicationGroupInput(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestNewModifyReplicationGroupShardConfigurationInput(t *testing.T) {
	cases := []struct {
		name     string
		params   *v1beta1.ReplicationGroupParameters
		observed elasticachetypes.ReplicationGroup
		want     *elasticache.ModifyReplicationGroupShardConfigurationInput
	}{
		{
			name:   "ScaleUp",
			params: replicationGroupParams(),
			observed: elasticachetypes.ReplicationGroup{
				NodeGroups: []elasticachetypes.NodeGroup{
					{
						NodeGroupId: pointer.ToOrNilIfZeroValue("ng-01"),
					},
				},
			},
			want: &elasticache.ModifyReplicationGroupShardConfigurationInput{
				ApplyImmediately:   true,
				NodeGroupCount:     2,
				ReplicationGroupId: ptr.To(name),
			},
		},
		{
			name:   "ScaleDown",
			params: replicationGroupParams(),
			observed: elasticachetypes.ReplicationGroup{
				NodeGroups: []elasticachetypes.NodeGroup{
					{NodeGroupId: pointer.ToOrNilIfZeroValue("ng-01")},
					{NodeGroupId: pointer.ToOrNilIfZeroValue("ng-02")},
					{NodeGroupId: pointer.ToOrNilIfZeroValue("ng-03")},
				},
			},
			want: &elasticache.ModifyReplicationGroupShardConfigurationInput{
				ApplyImmediately:   true,
				NodeGroupCount:     2,
				NodeGroupsToRemove: []string{"ng-01"},
				ReplicationGroupId: pointer.ToOrNilIfZeroValue(name),
			},
		},
		{
			name: "ApplyImmediatelyFromRG",
			params: &v1beta1.ReplicationGroupParameters{
				ApplyModificationsImmediately: false,
				NumNodeGroups:                 &numNodeGroups,
			},
			observed: elasticachetypes.ReplicationGroup{
				NodeGroups: []elasticachetypes.NodeGroup{
					{NodeGroupId: pointer.ToOrNilIfZeroValue("ng-01")},
					{NodeGroupId: pointer.ToOrNilIfZeroValue("ng-02")},
					{NodeGroupId: pointer.ToOrNilIfZeroValue("ng-03")},
				},
			},
			want: &elasticache.ModifyReplicationGroupShardConfigurationInput{
				ApplyImmediately:   false,
				NodeGroupCount:     2,
				NodeGroupsToRemove: []string{"ng-01"},
				ReplicationGroupId: pointer.ToOrNilIfZeroValue(name),
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := NewModifyReplicationGroupShardConfigurationInput(*tc.params, name, tc.observed)

			if diff := cmp.Diff(tc.want, got, cmpopts.IgnoreTypes(document.NoSerde{})); diff != "" {
				t.Errorf("NewModifyReplicationGroupShardConfigurationInput(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestNewDeleteReplicationGroupInput(t *testing.T) {
	cases := []struct {
		name string
		want *elasticache.DeleteReplicationGroupInput
	}{
		{
			name: "Successful",
			want: &elasticache.DeleteReplicationGroupInput{ReplicationGroupId: ptr.To(name)},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := NewDeleteReplicationGroupInput(name)

			if diff := cmp.Diff(tc.want, got, cmpopts.IgnoreTypes(document.NoSerde{})); diff != "" {
				t.Errorf("NewDeleteReplicationGroupInput(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestNewDescribeReplicationGroupsInput(t *testing.T) {
	cases := []struct {
		name string
		want *elasticache.DescribeReplicationGroupsInput
	}{
		{
			name: "Successful",
			want: &elasticache.DescribeReplicationGroupsInput{ReplicationGroupId: ptr.To(name)},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := NewDescribeReplicationGroupsInput(name)
			if diff := cmp.Diff(tc.want, got, cmpopts.IgnoreTypes(document.NoSerde{})); diff != "" {
				t.Errorf("NewDescribeReplicationGroupsInput(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestNewDescribeCacheClustersInput(t *testing.T) {
	cases := []struct {
		name    string
		cluster string
		want    *elasticache.DescribeCacheClustersInput
	}{
		{
			name:    "Successful",
			cluster: cacheClusterID,
			want:    &elasticache.DescribeCacheClustersInput{CacheClusterId: ptr.To(cacheClusterID)},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := NewDescribeCacheClustersInput(tc.cluster)
			if diff := cmp.Diff(tc.want, got, cmpopts.IgnoreTypes(document.NoSerde{})); diff != "" {
				t.Errorf("NewDescribeCacheClustersInput(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestLateInitialize(t *testing.T) {
	cases := []struct {
		name   string
		params *v1beta1.ReplicationGroupParameters
		rg     elasticachetypes.ReplicationGroup
		cc     elasticachetypes.CacheCluster
		want   *v1beta1.ReplicationGroupParameters
	}{
		{
			name: "NoChange",
			params: &v1beta1.ReplicationGroupParameters{
				AtRestEncryptionEnabled:    &atRestEncryptionEnabled,
				AuthEnabled:                &authEnabled,
				AutomaticFailoverEnabled:   &autoFailoverEnabled,
				SnapshotRetentionLimit:     &snapshotRetentionLimit,
				SnapshotWindow:             &snapshotWindow,
				SnapshottingClusterID:      &snapshottingClusterID,
				TransitEncryptionEnabled:   &transitEncryptionEnabled,
				EngineVersion:              &engineVersion,
				CacheParameterGroupName:    &cacheParameterGroupName,
				NotificationTopicARN:       &notificationTopicARN,
				NotificationTopicStatus:    &notificationTopicStatus,
				PreferredMaintenanceWindow: &maintenanceWindow,
				SecurityGroupIDs:           []string{securityGroupIDs[0]},
				CacheSecurityGroupNames:    []string{cacheSecurityGroupNames[0]},
			},
			rg: elasticachetypes.ReplicationGroup{
				AtRestEncryptionEnabled:  &atRestEncryptionEnabled,
				AuthTokenEnabled:         &authEnabled,
				AutomaticFailover:        elasticachetypes.AutomaticFailoverStatusEnabled,
				SnapshotRetentionLimit:   pointer.ToIntAsInt32Ptr(&snapshotRetentionLimit),
				SnapshotWindow:           pointer.ToOrNilIfZeroValue(snapshotWindow),
				SnapshottingClusterId:    pointer.ToOrNilIfZeroValue(snapshottingClusterID),
				TransitEncryptionEnabled: &transitEncryptionEnabled,
			},
			cc: elasticachetypes.CacheCluster{
				EngineVersion:       pointer.ToOrNilIfZeroValue(engineVersion),
				CacheParameterGroup: &elasticachetypes.CacheParameterGroupStatus{CacheParameterGroupName: pointer.ToOrNilIfZeroValue(cacheParameterGroupName)},
				NotificationConfiguration: &elasticachetypes.NotificationConfiguration{
					TopicArn:    pointer.ToOrNilIfZeroValue(notificationTopicARN),
					TopicStatus: pointer.ToOrNilIfZeroValue(notificationTopicStatus),
				},
				PreferredMaintenanceWindow: pointer.ToOrNilIfZeroValue(maintenanceWindow),
				SecurityGroups: []elasticachetypes.SecurityGroupMembership{
					{
						SecurityGroupId: pointer.ToOrNilIfZeroValue(securityGroupIDs[0]),
					},
				},
				CacheSecurityGroups: []elasticachetypes.CacheSecurityGroupMembership{
					{
						CacheSecurityGroupName: pointer.ToOrNilIfZeroValue(cacheSecurityGroupNames[0]),
					},
				},
			},
			want: &v1beta1.ReplicationGroupParameters{
				AtRestEncryptionEnabled:    &atRestEncryptionEnabled,
				AuthEnabled:                &authEnabled,
				AutomaticFailoverEnabled:   &autoFailoverEnabled,
				SnapshotRetentionLimit:     &snapshotRetentionLimit,
				SnapshotWindow:             &snapshotWindow,
				SnapshottingClusterID:      &snapshottingClusterID,
				TransitEncryptionEnabled:   &transitEncryptionEnabled,
				EngineVersion:              &engineVersion,
				CacheParameterGroupName:    &cacheParameterGroupName,
				NotificationTopicARN:       &notificationTopicARN,
				NotificationTopicStatus:    &notificationTopicStatus,
				PreferredMaintenanceWindow: &maintenanceWindow,
				SecurityGroupIDs:           []string{securityGroupIDs[0]},
				CacheSecurityGroupNames:    []string{cacheSecurityGroupNames[0]},
			},
		},
		{
			name:   "AllChanged",
			params: &v1beta1.ReplicationGroupParameters{},
			rg: elasticachetypes.ReplicationGroup{
				AtRestEncryptionEnabled:  &atRestEncryptionEnabled,
				AuthTokenEnabled:         &authEnabled,
				AutomaticFailover:        elasticachetypes.AutomaticFailoverStatusEnabled,
				SnapshotRetentionLimit:   pointer.ToIntAsInt32Ptr(&snapshotRetentionLimit),
				SnapshotWindow:           pointer.ToOrNilIfZeroValue(snapshotWindow),
				SnapshottingClusterId:    pointer.ToOrNilIfZeroValue(snapshottingClusterID),
				TransitEncryptionEnabled: &transitEncryptionEnabled,
			},
			cc: elasticachetypes.CacheCluster{
				EngineVersion:       pointer.ToOrNilIfZeroValue(engineVersion),
				CacheParameterGroup: &elasticachetypes.CacheParameterGroupStatus{CacheParameterGroupName: pointer.ToOrNilIfZeroValue(cacheParameterGroupName)},
				NotificationConfiguration: &elasticachetypes.NotificationConfiguration{
					TopicArn:    pointer.ToOrNilIfZeroValue(notificationTopicARN),
					TopicStatus: pointer.ToOrNilIfZeroValue(notificationTopicStatus),
				},
				PreferredMaintenanceWindow: pointer.ToOrNilIfZeroValue(maintenanceWindow),
				SecurityGroups: []elasticachetypes.SecurityGroupMembership{
					{
						SecurityGroupId: pointer.ToOrNilIfZeroValue(securityGroupIDs[0]),
					},
				},
				CacheSecurityGroups: []elasticachetypes.CacheSecurityGroupMembership{
					{
						CacheSecurityGroupName: pointer.ToOrNilIfZeroValue(cacheSecurityGroupNames[0]),
					},
				},
			},
			want: &v1beta1.ReplicationGroupParameters{
				AtRestEncryptionEnabled:    &atRestEncryptionEnabled,
				AuthEnabled:                &authEnabled,
				AutomaticFailoverEnabled:   &autoFailoverEnabled,
				SnapshotRetentionLimit:     &snapshotRetentionLimit,
				SnapshotWindow:             &snapshotWindow,
				SnapshottingClusterID:      &snapshottingClusterID,
				TransitEncryptionEnabled:   &transitEncryptionEnabled,
				EngineVersion:              &engineVersion,
				CacheParameterGroupName:    &cacheParameterGroupName,
				NotificationTopicARN:       &notificationTopicARN,
				NotificationTopicStatus:    &notificationTopicStatus,
				PreferredMaintenanceWindow: &maintenanceWindow,
				SecurityGroupIDs:           []string{securityGroupIDs[0]},
				CacheSecurityGroupNames:    []string{cacheSecurityGroupNames[0]},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			LateInitialize(tc.params, tc.rg, tc.cc)
			if diff := cmp.Diff(tc.want, tc.params); diff != "" {
				t.Errorf("LateInitialize(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGenerateObservation(t *testing.T) {
	automaticFailover := elasticachetypes.AutomaticFailoverStatusEnabled
	clusterEnabled := true
	configurationEndpoint := &elasticachetypes.Endpoint{
		Address: pointer.ToOrNilIfZeroValue("istanbul"),
		Port:    34,
	}
	memberClusters := []string{"member-1", "member-2"}
	status := "creating"
	nodeGroups := []elasticachetypes.NodeGroup{
		{
			NodeGroupId: pointer.ToOrNilIfZeroValue("my-id"),
			Slots:       pointer.ToOrNilIfZeroValue("special-slots"),
			Status:      pointer.ToOrNilIfZeroValue("creating"),
			PrimaryEndpoint: &elasticachetypes.Endpoint{
				Address: pointer.ToOrNilIfZeroValue("random-12"),
				Port:    124,
			},
			NodeGroupMembers: []elasticachetypes.NodeGroupMember{
				{
					CacheClusterId:            pointer.ToOrNilIfZeroValue("my-cache-cluster"),
					CacheNodeId:               pointer.ToOrNilIfZeroValue("cluster-0001"),
					CurrentRole:               pointer.ToOrNilIfZeroValue("secret-role"),
					PreferredAvailabilityZone: pointer.ToOrNilIfZeroValue("us-east-1"),
					ReadEndpoint: &elasticachetypes.Endpoint{
						Address: pointer.ToOrNilIfZeroValue("random-1"),
						Port:    23,
					},
				},
			},
		},
	}
	percentage := float64(54)
	rgpmdv := elasticachetypes.ReplicationGroupPendingModifiedValues{
		AutomaticFailoverStatus: elasticachetypes.PendingAutomaticFailoverStatusEnabled,
		PrimaryClusterId:        pointer.ToOrNilIfZeroValue("my-coolest-cluster"),
		Resharding: &elasticachetypes.ReshardingStatus{
			SlotMigration: &elasticachetypes.SlotMigration{
				ProgressPercentage: percentage,
			},
		},
	}
	cases := []struct {
		name string
		rg   elasticachetypes.ReplicationGroup
		want v1beta1.ReplicationGroupObservation
	}{
		{
			name: "AllFields",
			rg: elasticachetypes.ReplicationGroup{
				AutomaticFailover:     automaticFailover,
				ClusterEnabled:        &clusterEnabled,
				ConfigurationEndpoint: configurationEndpoint,
				MemberClusters:        memberClusters,
				Status:                &status,
				NodeGroups:            nodeGroups,
				PendingModifiedValues: &rgpmdv,
			},
			want: v1beta1.ReplicationGroupObservation{
				AutomaticFailover: string(automaticFailover),
				ClusterEnabled:    clusterEnabled,
				ConfigurationEndpoint: v1beta1.Endpoint{
					Address: *configurationEndpoint.Address,
					Port:    int(configurationEndpoint.Port),
				},
				MemberClusters: memberClusters,
				NodeGroups: []v1beta1.NodeGroup{
					generateNodeGroup(nodeGroups[0]),
				},
				PendingModifiedValues: generateReplicationGroupPendingModifiedValues(rgpmdv),
				Status:                status,
			},
		},
		{
			name: "cluster-mode-no-endpoint",
			rg: elasticachetypes.ReplicationGroup{
				AutomaticFailover:     automaticFailover,
				ClusterEnabled:        &clusterEnabled,
				ConfigurationEndpoint: nil,
				MemberClusters:        memberClusters,
				Status:                &status,
				NodeGroups:            nodeGroups,
				PendingModifiedValues: &rgpmdv,
			},
			want: v1beta1.ReplicationGroupObservation{
				AutomaticFailover:     string(automaticFailover),
				ClusterEnabled:        clusterEnabled,
				ConfigurationEndpoint: v1beta1.Endpoint{},
				MemberClusters:        memberClusters,
				NodeGroups: []v1beta1.NodeGroup{
					generateNodeGroup(nodeGroups[0]),
				},
				PendingModifiedValues: generateReplicationGroupPendingModifiedValues(rgpmdv),
				Status:                status,
			},
		},
		{
			name: "non-cluster-mode",
			rg: elasticachetypes.ReplicationGroup{
				AutomaticFailover:     automaticFailover,
				ClusterEnabled:        pointer.ToOrNilIfZeroValue(false),
				ConfigurationEndpoint: configurationEndpoint,
				MemberClusters:        memberClusters,
				Status:                &status,
				NodeGroups:            nodeGroups,
				PendingModifiedValues: &rgpmdv,
			},
			want: v1beta1.ReplicationGroupObservation{
				AutomaticFailover: string(automaticFailover),
				ClusterEnabled:    false,
				ConfigurationEndpoint: v1beta1.Endpoint{
					Address: *nodeGroups[0].PrimaryEndpoint.Address,
					Port:    int(nodeGroups[0].PrimaryEndpoint.Port),
				},
				MemberClusters: memberClusters,
				NodeGroups: []v1beta1.NodeGroup{
					generateNodeGroup(nodeGroups[0]),
				},
				PendingModifiedValues: generateReplicationGroupPendingModifiedValues(rgpmdv),
				Status:                status,
			},
		},
		{
			name: "non-cluster-mode-no-nodes",
			rg: elasticachetypes.ReplicationGroup{
				AutomaticFailover:     automaticFailover,
				ClusterEnabled:        pointer.ToOrNilIfZeroValue(false),
				ConfigurationEndpoint: configurationEndpoint,
				MemberClusters:        memberClusters,
				Status:                &status,
				PendingModifiedValues: &rgpmdv,
			},
			want: v1beta1.ReplicationGroupObservation{
				AutomaticFailover:     string(automaticFailover),
				ClusterEnabled:        false,
				ConfigurationEndpoint: v1beta1.Endpoint{},
				MemberClusters:        memberClusters,
				NodeGroups:            nil,
				PendingModifiedValues: generateReplicationGroupPendingModifiedValues(rgpmdv),
				Status:                status,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			o := GenerateObservation(tc.rg)
			if diff := cmp.Diff(tc.want, o); diff != "" {
				t.Errorf("LateInitialize(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestReplicationGroupNeedsUpdate(t *testing.T) {
	cases := []struct {
		name   string
		kube   *v1beta1.ReplicationGroupParameters
		rg     elasticachetypes.ReplicationGroup
		ccList []elasticachetypes.CacheCluster
		want   bool
	}{
		{
			name: "NeedsFailoverEnabled",
			kube: replicationGroupParams(),
			rg:   elasticachetypes.ReplicationGroup{AutomaticFailover: elasticachetypes.AutomaticFailoverStatusDisabled},
			want: true,
		},
		{
			name: "NeedsNewCacheNodeType",
			kube: replicationGroupParams(),
			rg: elasticachetypes.ReplicationGroup{
				AutomaticFailover: elasticachetypes.AutomaticFailoverStatusEnabling,
				CacheNodeType:     pointer.ToOrNilIfZeroValue("n1.insufficiently.cool"),
			},
			want: true,
		},
		{
			name: "NeedsNewSnapshotRetentionLimit",
			kube: replicationGroupParams(),
			rg: elasticachetypes.ReplicationGroup{
				AutomaticFailover:      elasticachetypes.AutomaticFailoverStatusEnabling,
				CacheNodeType:          pointer.ToOrNilIfZeroValue(cacheNodeType),
				SnapshotRetentionLimit: pointer.ToIntAsInt32Ptr(&newSnapshotRetentionLimit),
			},
			want: true,
		},
		{
			name: "NeedsNewSnapshotWindow",
			kube: replicationGroupParams(),
			rg: elasticachetypes.ReplicationGroup{
				AutomaticFailover:      elasticachetypes.AutomaticFailoverStatusEnabling,
				CacheNodeType:          pointer.ToOrNilIfZeroValue(cacheNodeType),
				SnapshotRetentionLimit: pointer.ToIntAsInt32Ptr(&snapshotRetentionLimit),
				SnapshotWindow:         pointer.ToOrNilIfZeroValue("yesterday"),
			},
			want: true,
		},
		{
			name: "CacheClusterNeedsUpdate",
			kube: replicationGroupParams(),
			rg: elasticachetypes.ReplicationGroup{
				AutomaticFailover:      elasticachetypes.AutomaticFailoverStatusEnabling,
				CacheNodeType:          pointer.ToOrNilIfZeroValue(cacheNodeType),
				SnapshotRetentionLimit: pointer.ToIntAsInt32Ptr(&snapshotRetentionLimit),
				SnapshotWindow:         pointer.ToOrNilIfZeroValue(snapshotWindow),
			},
			ccList: []elasticachetypes.CacheCluster{
				{
					EngineVersion: pointer.ToOrNilIfZeroValue("4.0.0"),
				},
			},
			want: true,
		},
		{
			name: "NeedsNoUpdate",
			kube: replicationGroupParams(),
			rg: elasticachetypes.ReplicationGroup{
				AutomaticFailover:      elasticachetypes.AutomaticFailoverStatusEnabling,
				CacheNodeType:          pointer.ToOrNilIfZeroValue(cacheNodeType),
				MultiAZ:                elasticachetypes.MultiAZStatusEnabled,
				SnapshotRetentionLimit: pointer.ToIntAsInt32Ptr(&snapshotRetentionLimit),
				SnapshotWindow:         pointer.ToOrNilIfZeroValue(snapshotWindow),
			},
			ccList: []elasticachetypes.CacheCluster{
				{
					EngineVersion:              pointer.ToOrNilIfZeroValue(engineVersion),
					CacheParameterGroup:        &elasticachetypes.CacheParameterGroupStatus{CacheParameterGroupName: pointer.ToOrNilIfZeroValue(cacheParameterGroupName)},
					NotificationConfiguration:  &elasticachetypes.NotificationConfiguration{TopicArn: pointer.ToOrNilIfZeroValue(notificationTopicARN), TopicStatus: pointer.ToOrNilIfZeroValue(notificationTopicStatus)},
					PreferredMaintenanceWindow: pointer.ToOrNilIfZeroValue(maintenanceWindow),
					SecurityGroups: func() []elasticachetypes.SecurityGroupMembership {
						ids := make([]elasticachetypes.SecurityGroupMembership, len(securityGroupIDs))
						for i, id := range securityGroupIDs {
							ids[i] = elasticachetypes.SecurityGroupMembership{SecurityGroupId: pointer.ToOrNilIfZeroValue(id)}
						}
						return ids
					}(),
					CacheSecurityGroups: func() []elasticachetypes.CacheSecurityGroupMembership {
						names := make([]elasticachetypes.CacheSecurityGroupMembership, len(cacheSecurityGroupNames))
						for i, n := range cacheSecurityGroupNames {
							names[i] = elasticachetypes.CacheSecurityGroupMembership{CacheSecurityGroupName: pointer.ToOrNilIfZeroValue(n)}
						}
						return names
					}(),
				},
			},
			want: false,
		},
		{
			name: "NeedsMultiAZUpdate",
			kube: replicationGroupParams(),
			rg: elasticachetypes.ReplicationGroup{
				AutomaticFailover:      elasticachetypes.AutomaticFailoverStatusEnabling,
				CacheNodeType:          pointer.ToOrNilIfZeroValue(cacheNodeType),
				MultiAZ:                elasticachetypes.MultiAZStatusDisabled, // trigger Update
				SnapshotRetentionLimit: pointer.ToIntAsInt32Ptr(&snapshotRetentionLimit),
				SnapshotWindow:         pointer.ToOrNilIfZeroValue(snapshotWindow),
			},
			ccList: []elasticachetypes.CacheCluster{
				{
					EngineVersion:              pointer.ToOrNilIfZeroValue(engineVersion),
					CacheParameterGroup:        &elasticachetypes.CacheParameterGroupStatus{CacheParameterGroupName: pointer.ToOrNilIfZeroValue(cacheParameterGroupName)},
					NotificationConfiguration:  &elasticachetypes.NotificationConfiguration{TopicArn: pointer.ToOrNilIfZeroValue(notificationTopicARN), TopicStatus: pointer.ToOrNilIfZeroValue(notificationTopicStatus)},
					PreferredMaintenanceWindow: pointer.ToOrNilIfZeroValue(maintenanceWindow),
					SecurityGroups: func() []elasticachetypes.SecurityGroupMembership {
						ids := make([]elasticachetypes.SecurityGroupMembership, len(securityGroupIDs))
						for i, id := range securityGroupIDs {
							ids[i] = elasticachetypes.SecurityGroupMembership{SecurityGroupId: pointer.ToOrNilIfZeroValue(id)}
						}
						return ids
					}(),
					CacheSecurityGroups: func() []elasticachetypes.CacheSecurityGroupMembership {
						names := make([]elasticachetypes.CacheSecurityGroupMembership, len(cacheSecurityGroupNames))
						for i, n := range cacheSecurityGroupNames {
							names[i] = elasticachetypes.CacheSecurityGroupMembership{CacheSecurityGroupName: pointer.ToOrNilIfZeroValue(n)}
						}
						return names
					}(),
				},
			},
			want: true,
		},
		{
			name: "NeedsUpdateNumCacheClusters",
			kube: replicationGroupParams(),
			rg: elasticachetypes.ReplicationGroup{
				AutomaticFailover:      elasticachetypes.AutomaticFailoverStatusEnabling,
				CacheNodeType:          pointer.ToOrNilIfZeroValue(cacheNodeType),
				MultiAZ:                elasticachetypes.MultiAZStatusEnabled,
				SnapshotRetentionLimit: pointer.ToIntAsInt32Ptr(&snapshotRetentionLimit),
				SnapshotWindow:         pointer.ToOrNilIfZeroValue(snapshotWindow),
			},
			ccList: []elasticachetypes.CacheCluster{},
			want:   true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := ReplicationGroupNeedsUpdate(*tc.kube, tc.rg, tc.ccList) != ""
			if got != tc.want {
				t.Errorf("ReplicationGroupNeedsUpdate(...): want %t, got %t", tc.want, got)
			}
		})
	}
}

func TestReplicationGroupShardConfigurationNeedsUpdate(t *testing.T) {
	cases := []struct {
		name   string
		kube   *v1beta1.ReplicationGroupParameters
		rg     elasticachetypes.ReplicationGroup
		ccList []elasticachetypes.CacheCluster
		want   bool
	}{
		{
			name: "NodeMismatch",
			kube: replicationGroupParams(), // 2
			rg: elasticachetypes.ReplicationGroup{
				NodeGroups: make([]elasticachetypes.NodeGroup, 3),
			},
			want: true,
		},
		{
			name: "UpToDate",
			kube: replicationGroupParams(),
			rg: elasticachetypes.ReplicationGroup{
				NodeGroups: make([]elasticachetypes.NodeGroup, numNodeGroups),
			},
			want: false,
		},
		{
			name: "NilNumNodes",
			kube: &v1beta1.ReplicationGroupParameters{
				NumNodeGroups: nil,
			},
			rg: elasticachetypes.ReplicationGroup{
				NodeGroups: make([]elasticachetypes.NodeGroup, numNodeGroups),
			},
			want: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := ReplicationGroupShardConfigurationNeedsUpdate(*tc.kube, tc.rg)
			if got != tc.want {
				t.Errorf("ReplicationGroupShardConfigurationNeedsUpdate(...): want %t, got %t", tc.want, got)
			}
		})
	}
}

func TestCacheClusterNeedsUpdate(t *testing.T) {
	cases := []struct {
		name string
		kube *v1beta1.ReplicationGroupParameters
		cc   elasticachetypes.CacheCluster
		want bool
	}{
		{
			name: "NeedsNewEngineVersion",
			kube: replicationGroupParams(),
			cc: elasticachetypes.CacheCluster{
				EngineVersion: pointer.ToOrNilIfZeroValue("4.0.0"),
			},
			want: true,
		},
		{
			name: "NeedsNewCacheParameterGroup",
			kube: replicationGroupParams(),
			cc: elasticachetypes.CacheCluster{
				EngineVersion:       pointer.ToOrNilIfZeroValue(engineVersion),
				CacheParameterGroup: &elasticachetypes.CacheParameterGroupStatus{CacheParameterGroupName: pointer.ToOrNilIfZeroValue("okaygroupiguess")},
			},
			want: true,
		},
		{
			name: "NeedsNewNotificationTopicARN",
			kube: replicationGroupParams(),
			cc: elasticachetypes.CacheCluster{
				EngineVersion:             pointer.ToOrNilIfZeroValue(engineVersion),
				CacheParameterGroup:       &elasticachetypes.CacheParameterGroupStatus{CacheParameterGroupName: pointer.ToOrNilIfZeroValue(cacheParameterGroupName)},
				NotificationConfiguration: &elasticachetypes.NotificationConfiguration{TopicArn: pointer.ToOrNilIfZeroValue("aws:arn:sqs:nope"), TopicStatus: pointer.ToOrNilIfZeroValue(notificationTopicStatus)},
			},
			want: true,
		},
		{
			name: "NeedsNewMaintenanceWindow",
			kube: replicationGroupParams(),
			cc: elasticachetypes.CacheCluster{
				EngineVersion:              pointer.ToOrNilIfZeroValue(engineVersion),
				CacheParameterGroup:        &elasticachetypes.CacheParameterGroupStatus{CacheParameterGroupName: pointer.ToOrNilIfZeroValue(cacheParameterGroupName)},
				NotificationConfiguration:  &elasticachetypes.NotificationConfiguration{TopicArn: pointer.ToOrNilIfZeroValue(notificationTopicARN), TopicStatus: pointer.ToOrNilIfZeroValue(notificationTopicStatus)},
				PreferredMaintenanceWindow: pointer.ToOrNilIfZeroValue("never!"),
			},
			want: true,
		},
		{
			name: "NeedsNewSecurityGroupIDs",
			kube: replicationGroupParams(),
			cc: elasticachetypes.CacheCluster{
				EngineVersion:              pointer.ToOrNilIfZeroValue(engineVersion),
				CacheParameterGroup:        &elasticachetypes.CacheParameterGroupStatus{CacheParameterGroupName: pointer.ToOrNilIfZeroValue(cacheParameterGroupName)},
				NotificationConfiguration:  &elasticachetypes.NotificationConfiguration{TopicArn: pointer.ToOrNilIfZeroValue(notificationTopicARN), TopicStatus: pointer.ToOrNilIfZeroValue(notificationTopicStatus)},
				PreferredMaintenanceWindow: pointer.ToOrNilIfZeroValue(maintenanceWindow),
				SecurityGroups: []elasticachetypes.SecurityGroupMembership{
					{SecurityGroupId: pointer.ToOrNilIfZeroValue("notaverysecuregroupid")},
					{SecurityGroupId: pointer.ToOrNilIfZeroValue("evenlesssecuregroupid")},
				},
			},
			want: true,
		},
		{
			name: "NeedsSecurityGroupIDs",
			kube: replicationGroupParams(),
			cc: elasticachetypes.CacheCluster{
				EngineVersion:              pointer.ToOrNilIfZeroValue(engineVersion),
				CacheParameterGroup:        &elasticachetypes.CacheParameterGroupStatus{CacheParameterGroupName: pointer.ToOrNilIfZeroValue(cacheParameterGroupName)},
				NotificationConfiguration:  &elasticachetypes.NotificationConfiguration{TopicArn: pointer.ToOrNilIfZeroValue(notificationTopicARN), TopicStatus: pointer.ToOrNilIfZeroValue(notificationTopicStatus)},
				PreferredMaintenanceWindow: pointer.ToOrNilIfZeroValue(maintenanceWindow),
			},
			want: true,
		},
		{
			name: "NeedsNewSecurityGroupNames",
			kube: replicationGroupParams(),
			cc: elasticachetypes.CacheCluster{
				EngineVersion:              pointer.ToOrNilIfZeroValue(engineVersion),
				CacheParameterGroup:        &elasticachetypes.CacheParameterGroupStatus{CacheParameterGroupName: pointer.ToOrNilIfZeroValue(cacheParameterGroupName)},
				NotificationConfiguration:  &elasticachetypes.NotificationConfiguration{TopicArn: pointer.ToOrNilIfZeroValue(notificationTopicARN), TopicStatus: pointer.ToOrNilIfZeroValue(notificationTopicStatus)},
				PreferredMaintenanceWindow: pointer.ToOrNilIfZeroValue(maintenanceWindow),
				SecurityGroups: func() []elasticachetypes.SecurityGroupMembership {
					ids := make([]elasticachetypes.SecurityGroupMembership, len(securityGroupIDs))
					for i, id := range securityGroupIDs {
						ids[i] = elasticachetypes.SecurityGroupMembership{SecurityGroupId: pointer.ToOrNilIfZeroValue(id)}
					}
					return ids
				}(),
				CacheSecurityGroups: []elasticachetypes.CacheSecurityGroupMembership{
					{CacheSecurityGroupName: pointer.ToOrNilIfZeroValue("notaverysecuregroup")},
					{CacheSecurityGroupName: pointer.ToOrNilIfZeroValue("evenlesssecuregroup")},
				},
			},
			want: true,
		},
		{
			name: "NeedsSecurityGroupNames",
			kube: replicationGroupParams(),
			cc: elasticachetypes.CacheCluster{
				EngineVersion:              pointer.ToOrNilIfZeroValue(engineVersion),
				CacheParameterGroup:        &elasticachetypes.CacheParameterGroupStatus{CacheParameterGroupName: pointer.ToOrNilIfZeroValue(cacheParameterGroupName)},
				NotificationConfiguration:  &elasticachetypes.NotificationConfiguration{TopicArn: pointer.ToOrNilIfZeroValue(notificationTopicARN), TopicStatus: pointer.ToOrNilIfZeroValue(notificationTopicStatus)},
				PreferredMaintenanceWindow: pointer.ToOrNilIfZeroValue(maintenanceWindow),
				SecurityGroups: func() []elasticachetypes.SecurityGroupMembership {
					ids := make([]elasticachetypes.SecurityGroupMembership, len(securityGroupIDs))
					for i, id := range securityGroupIDs {
						ids[i] = elasticachetypes.SecurityGroupMembership{SecurityGroupId: pointer.ToOrNilIfZeroValue(id)}
					}
					return ids
				}(),
			},
			want: true,
		},
		{
			name: "NeedsNoUpdate",
			kube: replicationGroupParams(),
			cc: elasticachetypes.CacheCluster{
				EngineVersion:              pointer.ToOrNilIfZeroValue(engineVersion),
				CacheParameterGroup:        &elasticachetypes.CacheParameterGroupStatus{CacheParameterGroupName: pointer.ToOrNilIfZeroValue(cacheParameterGroupName)},
				NotificationConfiguration:  &elasticachetypes.NotificationConfiguration{TopicArn: pointer.ToOrNilIfZeroValue(notificationTopicARN), TopicStatus: pointer.ToOrNilIfZeroValue(notificationTopicStatus)},
				PreferredMaintenanceWindow: pointer.ToOrNilIfZeroValue(maintenanceWindow),
				SecurityGroups: func() []elasticachetypes.SecurityGroupMembership {
					ids := make([]elasticachetypes.SecurityGroupMembership, len(securityGroupIDs))
					for i, id := range securityGroupIDs {
						ids[i] = elasticachetypes.SecurityGroupMembership{SecurityGroupId: pointer.ToOrNilIfZeroValue(id)}
					}
					return ids
				}(),
				CacheSecurityGroups: func() []elasticachetypes.CacheSecurityGroupMembership {
					names := make([]elasticachetypes.CacheSecurityGroupMembership, len(cacheSecurityGroupNames))
					for i, n := range cacheSecurityGroupNames {
						names[i] = elasticachetypes.CacheSecurityGroupMembership{CacheSecurityGroupName: pointer.ToOrNilIfZeroValue(n)}
					}
					return names
				}(),
			},
			want: false,
		},
		{
			name: "NeedsNoUpdateNormalizedMaintenanceWindow",
			kube: replicationGroupParams(func(p *v1beta1.ReplicationGroupParameters) {
				p.PreferredMaintenanceWindow = pointer.ToOrNilIfZeroValue("Mon:00:00-Fri:23:59")

			}),
			cc: elasticachetypes.CacheCluster{
				EngineVersion:              pointer.ToOrNilIfZeroValue(engineVersion),
				CacheParameterGroup:        &elasticachetypes.CacheParameterGroupStatus{CacheParameterGroupName: pointer.ToOrNilIfZeroValue(cacheParameterGroupName)},
				NotificationConfiguration:  &elasticachetypes.NotificationConfiguration{TopicArn: pointer.ToOrNilIfZeroValue(notificationTopicARN), TopicStatus: pointer.ToOrNilIfZeroValue(notificationTopicStatus)},
				PreferredMaintenanceWindow: pointer.ToOrNilIfZeroValue("mon:00:00-fri:23:59"),
				SecurityGroups: func() []elasticachetypes.SecurityGroupMembership {
					ids := make([]elasticachetypes.SecurityGroupMembership, len(securityGroupIDs))
					for i, id := range securityGroupIDs {
						ids[i] = elasticachetypes.SecurityGroupMembership{SecurityGroupId: pointer.ToOrNilIfZeroValue(id)}
					}
					return ids
				}(),
				CacheSecurityGroups: func() []elasticachetypes.CacheSecurityGroupMembership {
					names := make([]elasticachetypes.CacheSecurityGroupMembership, len(cacheSecurityGroupNames))
					for i, n := range cacheSecurityGroupNames {
						names[i] = elasticachetypes.CacheSecurityGroupMembership{CacheSecurityGroupName: pointer.ToOrNilIfZeroValue(n)}
					}
					return names
				}(),
			},
			want: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := cacheClusterNeedsUpdate(*tc.kube, tc.cc) != ""
			if got != tc.want {
				t.Errorf("cacheClusterNeedsUpdate(...): want %t, got %t", tc.want, got)
			}
		})
	}
}

func TestConnectionEndpoint(t *testing.T) {
	cases := []struct {
		name string
		rg   elasticachetypes.ReplicationGroup
		want managed.ConnectionDetails
	}{
		{
			name: "ClusterModeEnabled",
			rg: elasticachetypes.ReplicationGroup{
				ClusterEnabled: pointer.ToOrNilIfZeroValue(true),
				ConfigurationEndpoint: &elasticachetypes.Endpoint{
					Address: pointer.ToOrNilIfZeroValue(host),
					Port:    int32(port),
				},
			},
			want: managed.ConnectionDetails{
				xpv1.ResourceCredentialsSecretEndpointKey: []byte(host),
				xpv1.ResourceCredentialsSecretPortKey:     []byte(strconv.Itoa(port)),
			},
		},
		{
			name: "ClusterModeEnabledMissingConfigurationEndpoint",
			rg: elasticachetypes.ReplicationGroup{
				ClusterEnabled: pointer.ToOrNilIfZeroValue(true),
			},
			want: nil,
		},
		{
			name: "ClusterModeDisabled",
			rg: elasticachetypes.ReplicationGroup{
				NodeGroups: []elasticachetypes.NodeGroup{{
					PrimaryEndpoint: &elasticachetypes.Endpoint{
						Address: pointer.ToOrNilIfZeroValue(host),
						Port:    int32(port),
					}},
				},
			},
			want: managed.ConnectionDetails{
				xpv1.ResourceCredentialsSecretEndpointKey: []byte(host),
				xpv1.ResourceCredentialsSecretPortKey:     []byte(strconv.Itoa(port)),
			},
		},
		{
			name: "ClusterModeDisabledMissingPrimaryEndpoint",
			rg:   elasticachetypes.ReplicationGroup{NodeGroups: []elasticachetypes.NodeGroup{{}}},
			want: nil,
		},
		{
			name: "ClusterModeDisabledMissingNodeGroups",
			rg:   elasticachetypes.ReplicationGroup{},
			want: nil,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := ConnectionEndpoint(tc.rg)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("ConnectionEndpoint(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestIsSubnetGroupUpToDate(t *testing.T) {
	type args struct {
		subnetGroup elasticachetypes.CacheSubnetGroup
		p           cachev1alpha1.CacheSubnetGroupParameters
	}

	cases := map[string]struct {
		args args
		want bool
	}{
		"SameFields": {
			args: args{
				subnetGroup: elasticachetypes.CacheSubnetGroup{
					CacheSubnetGroupDescription: pointer.ToOrNilIfZeroValue(subnetGroupDesc),
					Subnets: []elasticachetypes.Subnet{
						{
							SubnetIdentifier: pointer.ToOrNilIfZeroValue(subnetID1),
						},
						{
							SubnetIdentifier: pointer.ToOrNilIfZeroValue(subnetID2),
						},
					},
				},
				p: cachev1alpha1.CacheSubnetGroupParameters{
					Description: subnetGroupDesc,
					SubnetIDs:   []string{subnetID1, subnetID2},
				},
			},
			want: true,
		},
		"DifferentFields": {
			args: args{
				subnetGroup: elasticachetypes.CacheSubnetGroup{
					CacheSubnetGroupDescription: pointer.ToOrNilIfZeroValue(subnetGroupDesc),
					Subnets: []elasticachetypes.Subnet{
						{
							SubnetIdentifier: pointer.ToOrNilIfZeroValue(subnetID1),
						},
						{
							SubnetIdentifier: pointer.ToOrNilIfZeroValue(subnetID2),
						},
					},
				},
				p: cachev1alpha1.CacheSubnetGroupParameters{
					Description: subnetGroupDesc,
					SubnetIDs:   []string{subnetID1},
				},
			},
			want: false,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := IsSubnetGroupUpToDate(tc.args.p, tc.args.subnetGroup)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestVersionMatches(t *testing.T) {
	cases := []struct {
		name        string
		kubeVersion *string
		awsVersion  *string
		want        bool
	}{
		{
			name:        "Same value",
			kubeVersion: pointer.ToOrNilIfZeroValue("5.0.8"),
			awsVersion:  pointer.ToOrNilIfZeroValue("5.0.8"),
			want:        true,
		},
		{
			name:        "Same pattern", // currently this will never happen, but if it does it should match..
			kubeVersion: pointer.ToOrNilIfZeroValue("6.x"),
			awsVersion:  pointer.ToOrNilIfZeroValue("6.x"),
			want:        true,
		},
		{
			name:        "Same with nil",
			kubeVersion: nil,
			awsVersion:  nil,
			want:        true,
		},
		{
			name:        "nil in kubernetes",
			kubeVersion: nil,
			awsVersion:  pointer.ToOrNilIfZeroValue("5.0.8"),
			want:        false,
		},
		{
			name:        "nil from aws",
			kubeVersion: pointer.ToOrNilIfZeroValue("5.0.8"),
			awsVersion:  nil,
			want:        false,
		},
		{
			name:        "mismatch",
			kubeVersion: pointer.ToOrNilIfZeroValue("5.0.8"),
			awsVersion:  pointer.ToOrNilIfZeroValue("5.0.9"),
			want:        false,
		},
		{
			name:        "pattern match",
			kubeVersion: pointer.ToOrNilIfZeroValue("6.x"),
			awsVersion:  pointer.ToOrNilIfZeroValue("6.0.5"),
			want:        true,
		},
		{
			name:        "minor match",
			kubeVersion: pointer.ToOrNilIfZeroValue("6.2"),
			awsVersion:  pointer.ToOrNilIfZeroValue("6.2.6"),
			want:        true,
		},
		{
			name:        "zero major mismatch",
			kubeVersion: pointer.ToOrNilIfZeroValue("0.2"),
			awsVersion:  pointer.ToOrNilIfZeroValue("6.2.6"),
			want:        false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := versionMatches(tc.kubeVersion, tc.awsVersion)
			if got != tc.want {
				t.Errorf("versionMatches(%+v) - got %v", tc, got)
			}
		})
	}
}

func TestParseVersion(t *testing.T) {
	cases := []struct {
		name    string
		version *string
		parsed  *PartialSemanticVersion
		wantErr error
	}{
		{
			name:    "nil",
			version: nil,
			parsed:  nil,
			wantErr: errors.New("empty string"),
		},
		{
			name:    "",
			version: nil,
			parsed:  nil,
			wantErr: errors.New("empty string"),
		},
		{
			name:    "bad version",
			version: pointer.ToOrNilIfZeroValue("badversion"),
			parsed:  nil,
			wantErr: errors.New("major version must be a number"),
		},
		{
			name:    "major only",
			version: pointer.ToOrNilIfZeroValue("6"),
			parsed:  &PartialSemanticVersion{Major: ptr.To[int64](6)},
			wantErr: nil,
		},
		{
			name:    "major.minor",
			version: pointer.ToOrNilIfZeroValue("6.2"),
			parsed:  &PartialSemanticVersion{Major: ptr.To[int64](6), Minor: ptr.To[int64](2)},
			wantErr: nil,
		},
		{
			name:    "major.x",
			version: pointer.ToOrNilIfZeroValue("6.x"),
			parsed:  &PartialSemanticVersion{Major: ptr.To[int64](6)},
			wantErr: nil,
		},
		{
			name:    "major.",
			version: pointer.ToOrNilIfZeroValue("6."),
			parsed:  &PartialSemanticVersion{Major: ptr.To[int64](6)},
			wantErr: nil,
		},
		{
			name:    "majorLarge.",
			version: pointer.ToOrNilIfZeroValue("999."),
			parsed:  &PartialSemanticVersion{Major: ptr.To[int64](999)},
			wantErr: nil,
		},
		{
			name:    "major.minor.patch",
			version: pointer.ToOrNilIfZeroValue("5.0.9"),
			parsed:  &PartialSemanticVersion{Major: ptr.To[int64](5), Minor: ptr.To[int64](0), Patch: ptr.To[int64](9)},
			wantErr: nil,
		},
		{
			name:    "major.minor.x",
			version: pointer.ToOrNilIfZeroValue("5.0.x"),
			parsed:  &PartialSemanticVersion{Major: ptr.To[int64](5), Minor: ptr.To[int64](0)},
			wantErr: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, gotErr := ParseVersion(tc.version)
			if diff := cmp.Diff(tc.parsed, got); diff != "" {
				t.Errorf("ParseVersion(...): -want, +got:\n%s", diff)
			}
			if (tc.wantErr == nil) != (gotErr == nil) {
				t.Errorf("ParseVersion Error (%+v) - got %v", tc.wantErr, gotErr)
			}
			if tc.wantErr != nil {
				if tc.wantErr.Error() != gotErr.Error() {
					t.Errorf("ParseVersion ErrorString (%s) - got %s", tc.wantErr.Error(), gotErr.Error())
				}
			}

		})
	}
}

func TestMultiAZEnabled(t *testing.T) {
	f := false
	tr := true
	cases := []struct {
		name string
		maz  elasticachetypes.MultiAZStatus
		want *bool
	}{
		{
			name: "empty status",
			maz:  elasticachetypes.MultiAZStatus(""),
			want: nil,
		},
		{
			name: "enabled",
			maz:  elasticachetypes.MultiAZStatusEnabled,
			want: &tr,
		},
		{
			name: "disabled",
			maz:  elasticachetypes.MultiAZStatusDisabled,
			want: &f,
		},
		{
			name: "unknown status",
			maz:  elasticachetypes.MultiAZStatus("unknown"),
			want: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := multiAZEnabled(tc.maz)
			if tc.want == nil && got != nil {
				t.Errorf("MultiAZEnabled(%+v) - got %v", tc.want, got)
			}
			if pointer.BoolValue(got) != pointer.BoolValue(tc.want) {
				t.Errorf("MultiAZEnabled(%+v) - got %v", tc.want, got)
			}
		})
	}
}

func TestDiffTags(t *testing.T) {
	type args struct {
		local  []v1beta1.Tag
		remote []elasticachetypes.Tag
	}
	type want struct {
		add    map[string]string
		remove []string
	}
	cases := map[string]struct {
		args args
		want want
	}{
		"AllNew": {
			args: args{
				local: []v1beta1.Tag{
					{Key: "key", Value: "val"},
				},
			},
			want: want{
				add:    map[string]string{"key": "val"},
				remove: []string{},
			},
		},
		"SomeNew": {
			args: args{
				local: []v1beta1.Tag{
					{Key: "key", Value: "val"},
					{Key: "key1", Value: "val1"},
					{Key: "key2", Value: "val2"},
				},
				remote: []elasticachetypes.Tag{
					{Key: pointer.ToOrNilIfZeroValue("key"), Value: pointer.ToOrNilIfZeroValue("val")},
				},
			},
			want: want{
				add: map[string]string{
					"key1": "val1",
					"key2": "val2",
				},
				remove: []string{},
			},
		},
		"Update": {
			args: args{
				local: []v1beta1.Tag{
					{Key: "key", Value: "different"},
					{Key: "key1", Value: "val1"},
					{Key: "key2", Value: "val2"},
				},
				remote: []elasticachetypes.Tag{
					{Key: pointer.ToOrNilIfZeroValue("key"), Value: pointer.ToOrNilIfZeroValue("val")},
					{Key: pointer.ToOrNilIfZeroValue("key1"), Value: pointer.ToOrNilIfZeroValue("val1")},
					{Key: pointer.ToOrNilIfZeroValue("key2"), Value: pointer.ToOrNilIfZeroValue("val2")},
				},
			},
			want: want{
				add:    map[string]string{"key": "different"},
				remove: []string{"key"},
			},
		},
		"RemoveAll": {
			args: args{
				remote: []elasticachetypes.Tag{
					{Key: pointer.ToOrNilIfZeroValue("key"), Value: pointer.ToOrNilIfZeroValue("val")},
					{Key: pointer.ToOrNilIfZeroValue("key1"), Value: pointer.ToOrNilIfZeroValue("val1")},
					{Key: pointer.ToOrNilIfZeroValue("key2"), Value: pointer.ToOrNilIfZeroValue("val2")},
				},
			},
			want: want{
				add:    map[string]string{},
				remove: []string{"key", "key1", "key2"},
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			tagCmp := cmpopts.SortSlices(func(i, j elasticachetypes.Tag) bool {
				return pointer.StringValue(i.Key) < pointer.StringValue(j.Key)
			})
			add, remove := DiffTags(tc.args.local, tc.args.remote)
			if diff := cmp.Diff(tc.want.add, add, tagCmp, cmpopts.IgnoreTypes(document.NoSerde{})); diff != "" {
				t.Errorf("add: -want, +got:\n%s", diff)
			}
			sort.Strings(tc.want.remove)
			sort.Strings(remove)
			if diff := cmp.Diff(tc.want.remove, remove, tagCmp); diff != "" {
				t.Errorf("remove: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestReplicationGroupTagsNeedsUpdate(t *testing.T) {
	type args struct {
		local  []v1beta1.Tag
		remote []elasticachetypes.Tag
	}
	type want struct {
		res bool
	}
	cases := map[string]struct {
		args args
		want want
	}{
		"AllNewRemoteNil": {
			args: args{
				local: []v1beta1.Tag{
					{Key: "key", Value: "val"},
				},
			},
			want: want{
				res: true,
			},
		},
		"AllNew": {
			args: args{
				local: []v1beta1.Tag{
					{Key: "key", Value: "val"},
				},
				remote: []elasticachetypes.Tag{},
			},
			want: want{
				res: true,
			},
		},
		"SomeNew": {
			args: args{
				local: []v1beta1.Tag{
					{Key: "key", Value: "val"},
					{Key: "key1", Value: "val1"},
					{Key: "key2", Value: "val2"},
				},
				remote: []elasticachetypes.Tag{
					{Key: pointer.ToOrNilIfZeroValue("key"), Value: pointer.ToOrNilIfZeroValue("val")},
				},
			},
			want: want{
				res: true,
			},
		},
		"Update": {
			args: args{
				local: []v1beta1.Tag{
					{Key: "key", Value: "different"},
					{Key: "key1", Value: "val1"},
					{Key: "key2", Value: "val2"},
				},
				remote: []elasticachetypes.Tag{
					{Key: pointer.ToOrNilIfZeroValue("key"), Value: pointer.ToOrNilIfZeroValue("val")},
					{Key: pointer.ToOrNilIfZeroValue("key1"), Value: pointer.ToOrNilIfZeroValue("val1")},
					{Key: pointer.ToOrNilIfZeroValue("key2"), Value: pointer.ToOrNilIfZeroValue("val2")},
				},
			},
			want: want{
				res: true,
			},
		},
		"Equal": {
			args: args{
				local: []v1beta1.Tag{
					{Key: "key1", Value: "val1"},
					{Key: "key2", Value: "val2"},
				},
				remote: []elasticachetypes.Tag{
					{Key: pointer.ToOrNilIfZeroValue("key1"), Value: pointer.ToOrNilIfZeroValue("val1")},
					{Key: pointer.ToOrNilIfZeroValue("key2"), Value: pointer.ToOrNilIfZeroValue("val2")},
				},
			},
			want: want{
				res: false,
			},
		},
		"EqualEmpty": {
			args: args{
				local:  []v1beta1.Tag{},
				remote: []elasticachetypes.Tag{},
			},
			want: want{
				res: false,
			},
		},
		"RemoveAll": {
			args: args{
				remote: []elasticachetypes.Tag{
					{Key: pointer.ToOrNilIfZeroValue("key"), Value: pointer.ToOrNilIfZeroValue("val")},
					{Key: pointer.ToOrNilIfZeroValue("key1"), Value: pointer.ToOrNilIfZeroValue("val1")},
					{Key: pointer.ToOrNilIfZeroValue("key2"), Value: pointer.ToOrNilIfZeroValue("val2")},
				},
			},
			want: want{
				res: true,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			res := ReplicationGroupTagsNeedsUpdate(tc.args.local, tc.args.remote)
			if diff := cmp.Diff(tc.want.res, res); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestReplicationGroupNumCacheClustersNeedsUpdate(t *testing.T) {
	var numCacheClusters5 = 5
	type args struct {
		kube   v1beta1.ReplicationGroupParameters
		ccList []elasticachetypes.CacheCluster
	}
	type want struct {
		res bool
	}
	cases := map[string]struct {
		args args
		want want
	}{
		"Equal": {
			args: args{
				kube: v1beta1.ReplicationGroupParameters{
					NumCacheClusters: &numCacheClusters,
				},
				ccList: []elasticachetypes.CacheCluster{
					{EngineVersion: pointer.ToOrNilIfZeroValue(engineVersion)},
				},
			},
			want: want{res: false},
		},
		"NotEqual": {
			args: args{
				kube: v1beta1.ReplicationGroupParameters{
					NumCacheClusters: &numCacheClusters5,
				},
				ccList: []elasticachetypes.CacheCluster{
					{EngineVersion: pointer.ToOrNilIfZeroValue(engineVersion)},
				},
			},
			want: want{res: true},
		},
		"NilRequest": {
			args: args{
				kube:   v1beta1.ReplicationGroupParameters{},
				ccList: []elasticachetypes.CacheCluster{},
			},
			want: want{res: false},
		},
		"NilRequestCC": {
			args: args{
				kube: v1beta1.ReplicationGroupParameters{},
				ccList: []elasticachetypes.CacheCluster{
					{EngineVersion: pointer.ToOrNilIfZeroValue(engineVersion)},
				},
			},
			want: want{res: false},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			res := ReplicationGroupNumCacheClustersNeedsUpdate(tc.args.kube, tc.args.ccList)
			if diff := cmp.Diff(tc.want.res, res); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}

}
