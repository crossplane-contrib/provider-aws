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
	"strconv"
	"testing"

	"github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"

	"github.com/aws/aws-sdk-go-v2/service/elasticache"
	"github.com/google/go-cmp/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	cachev1alpha1 "github.com/crossplane/provider-aws/apis/cache/v1alpha1"
	"github.com/crossplane/provider-aws/apis/cache/v1beta1"
	aws "github.com/crossplane/provider-aws/pkg/clients"
)

const (
	namespace = "coolNamespace"
	name      = "coolGroup"
	uid       = types.UID("definitely-a-uuid")
)

var (
	cacheNodeType            = "n1.super.cool"
	atRestEncryptionEnabled  = true
	authEnabled              = true
	authToken                = "coolToken"
	autoFailoverEnabled      = true
	cacheParameterGroupName  = "coolParamGroup"
	cacheSubnetGroupName     = "coolSubnet"
	engine                   = "redis"
	engineVersion            = "5.0.0"
	notificationTopicARN     = "arn:aws:sns:cooltopic"
	notificationTopicStatus  = "active"
	numCacheClusters         = 2
	numNodeGroups            = 2
	host                     = "coolhost"
	port                     = 6379
	primaryClusterID         = "the-coolest-one"
	maintenanceWindow        = "tomorrow"
	replicasPerNodeGroup     = 2
	snapshotName             = "coolSnapshot"
	snapshotRetentionLimit   = 1
	snapshottingClusterID    = "snapshot-cluster"
	snapshotWindow           = "thedayaftertomorrow"
	tagKey                   = "key-1"
	tagValue                 = "value-1"
	transitEncryptionEnabled = true

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

	meta             = metav1.ObjectMeta{Namespace: namespace, Name: name, UID: uid}
	replicationGroup = &v1beta1.ReplicationGroup{
		ObjectMeta: meta,
		Spec: v1beta1.ReplicationGroupSpec{
			ForProvider: v1beta1.ReplicationGroupParameters{
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
			},
		},
	}
)

var (
	subnetGroupDesc = "some description"
	subnetID1       = "subnetId1"
	subnetID2       = "subnetId2"
)

func TestNewCreateReplicationGroupInput(t *testing.T) {
	cases := []struct {
		name      string
		params    v1beta1.ReplicationGroupParameters
		authToken *string
		want      *elasticache.CreateReplicationGroupInput
	}{
		{
			name:      "AllPossibleFields",
			params:    replicationGroup.Spec.ForProvider,
			authToken: &authToken,
			want: &elasticache.CreateReplicationGroupInput{
				ReplicationGroupId:          aws.String(name, aws.FieldRequired),
				ReplicationGroupDescription: aws.String(description, aws.FieldRequired),
				Engine:                      aws.String(v1beta1.CacheEngineRedis, aws.FieldRequired),
				CacheNodeType:               aws.String(cacheNodeType, aws.FieldRequired),
				AtRestEncryptionEnabled:     aws.Bool(atRestEncryptionEnabled),
				AuthToken:                   aws.String(authToken),
				AutomaticFailoverEnabled:    aws.Bool(autoFailoverEnabled),
				CacheParameterGroupName:     aws.String(cacheParameterGroupName),
				CacheSecurityGroupNames:     cacheSecurityGroupNames,
				CacheSubnetGroupName:        aws.String(cacheSubnetGroupName),
				EngineVersion:               aws.String(engineVersion),
				NodeGroupConfiguration: []elasticache.NodeGroupConfiguration{
					{
						PrimaryAvailabilityZone:  aws.String(nodeGroupPrimaryAZ),
						ReplicaAvailabilityZones: nodeGroupAZs,
						ReplicaCount:             aws.Int64(nodeGroupReplicaCount),
						Slots:                    aws.String(nodeGroupSlots),
					},
				},
				NotificationTopicArn:       aws.String(notificationTopicARN),
				NumCacheClusters:           aws.Int64(numCacheClusters),
				NumNodeGroups:              aws.Int64(numNodeGroups),
				Port:                       aws.Int64(port),
				PreferredCacheClusterAZs:   preferredCacheClusterAZs,
				PreferredMaintenanceWindow: aws.String(maintenanceWindow),
				PrimaryClusterId:           aws.String(primaryClusterID),
				ReplicasPerNodeGroup:       aws.Int64(replicasPerNodeGroup),
				SecurityGroupIds:           securityGroupIDs,
				SnapshotArns:               snapshotARNs,
				SnapshotName:               aws.String(snapshotName),
				SnapshotRetentionLimit:     aws.Int64(snapshotRetentionLimit),
				SnapshotWindow:             aws.String(snapshotWindow),
				Tags: []elasticache.Tag{
					{
						Key:   &tagKey,
						Value: &tagValue,
					},
				},
				TransitEncryptionEnabled: aws.Bool(transitEncryptionEnabled),
			},
		},
		{
			name: "UnsetFieldsAreNilNotZeroType",
			params: v1beta1.ReplicationGroupParameters{
				CacheNodeType:               cacheNodeType,
				ReplicationGroupDescription: description,
				Engine:                      engine,
			},
			want: &elasticache.CreateReplicationGroupInput{
				ReplicationGroupId:          aws.String(name, aws.FieldRequired),
				ReplicationGroupDescription: aws.String(description, aws.FieldRequired),
				Engine:                      aws.String(engine, aws.FieldRequired),
				CacheNodeType:               aws.String(cacheNodeType, aws.FieldRequired),
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := NewCreateReplicationGroupInput(tc.params, name, tc.authToken)

			if err := got.Validate(); err != nil {
				t.Errorf("NewCreateReplicationGroupInput(...): invalid input: %v", err)
			}
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("NewCreateReplicationGroupInput(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestNewModifyReplicationGroupInput(t *testing.T) {
	cases := []struct {
		name   string
		params v1beta1.ReplicationGroupParameters
		want   *elasticache.ModifyReplicationGroupInput
	}{
		{
			name:   "AllPossibleFields",
			params: replicationGroup.Spec.ForProvider,
			want: &elasticache.ModifyReplicationGroupInput{
				ReplicationGroupId:          aws.String(name, aws.FieldRequired),
				ApplyImmediately:            aws.Bool(true),
				AutomaticFailoverEnabled:    aws.Bool(autoFailoverEnabled),
				CacheNodeType:               aws.String(cacheNodeType),
				CacheParameterGroupName:     aws.String(cacheParameterGroupName),
				CacheSecurityGroupNames:     cacheSecurityGroupNames,
				EngineVersion:               aws.String(engineVersion),
				NotificationTopicArn:        aws.String(notificationTopicARN),
				NotificationTopicStatus:     aws.String(notificationTopicStatus),
				PreferredMaintenanceWindow:  aws.String(maintenanceWindow),
				PrimaryClusterId:            aws.String(primaryClusterID),
				ReplicationGroupDescription: aws.String(description),
				SecurityGroupIds:            securityGroupIDs,
				SnapshotRetentionLimit:      aws.Int64(snapshotRetentionLimit),
				SnapshotWindow:              aws.String(snapshotWindow),
				SnapshottingClusterId:       aws.String(snapshottingClusterID),
			},
		},
		{
			name: "UnsetFieldsAreNilNotZeroType",
			params: v1beta1.ReplicationGroupParameters{
				CacheNodeType:               cacheNodeType,
				ReplicationGroupDescription: description,
			},
			want: &elasticache.ModifyReplicationGroupInput{
				ApplyImmediately:            aws.Bool(false, aws.FieldRequired),
				ReplicationGroupId:          aws.String(name, aws.FieldRequired),
				ReplicationGroupDescription: aws.String(description, aws.FieldRequired),
				CacheNodeType:               aws.String(cacheNodeType, aws.FieldRequired),
			},
		},
		{
			name: "SuperfluousFields",
			params: v1beta1.ReplicationGroupParameters{
				AtRestEncryptionEnabled:     &atRestEncryptionEnabled,
				CacheNodeType:               cacheNodeType,
				ReplicationGroupDescription: description,
			},
			want: &elasticache.ModifyReplicationGroupInput{
				ApplyImmediately:            aws.Bool(false, aws.FieldRequired),
				ReplicationGroupId:          aws.String(name, aws.FieldRequired),
				ReplicationGroupDescription: aws.String(description, aws.FieldRequired),
				CacheNodeType:               aws.String(cacheNodeType, aws.FieldRequired),
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := NewModifyReplicationGroupInput(tc.params, name)

			if err := got.Validate(); err != nil {
				t.Errorf("NewModifyReplicationGroupInput(...): invalid input: %v", err)
			}
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("NewModifyReplicationGroupInput(...): -want, +got:\n%s", diff)
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
			want: &elasticache.DeleteReplicationGroupInput{ReplicationGroupId: aws.String(name, aws.FieldRequired)},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := NewDeleteReplicationGroupInput(name)

			if err := got.Validate(); err != nil {
				t.Errorf("NewDeleteReplicationGroupInput(...): invalid input: %v", err)
			}
			if diff := cmp.Diff(tc.want, got); diff != "" {
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
			want: &elasticache.DescribeReplicationGroupsInput{ReplicationGroupId: aws.String(name, aws.FieldRequired)},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := NewDescribeReplicationGroupsInput(name)
			if diff := cmp.Diff(tc.want, got); diff != "" {
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
			want:    &elasticache.DescribeCacheClustersInput{CacheClusterId: aws.String(cacheClusterID, aws.FieldRequired)},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := NewDescribeCacheClustersInput(tc.cluster)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("NewDescribeCacheClustersInput(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestLateInitialize(t *testing.T) {
	cases := []struct {
		name   string
		params *v1beta1.ReplicationGroupParameters
		rg     elasticache.ReplicationGroup
		want   *v1beta1.ReplicationGroupParameters
	}{
		{
			name: "NoChange",
			params: &v1beta1.ReplicationGroupParameters{
				AtRestEncryptionEnabled:  &atRestEncryptionEnabled,
				AuthEnabled:              &authEnabled,
				AutomaticFailoverEnabled: &autoFailoverEnabled,
				SnapshotRetentionLimit:   &snapshotRetentionLimit,
				SnapshotWindow:           &snapshotWindow,
				SnapshottingClusterID:    &snapshottingClusterID,
				TransitEncryptionEnabled: &transitEncryptionEnabled,
			},
			rg: elasticache.ReplicationGroup{
				AtRestEncryptionEnabled:  &atRestEncryptionEnabled,
				AuthTokenEnabled:         &authEnabled,
				AutomaticFailover:        elasticache.AutomaticFailoverStatusEnabled,
				SnapshotRetentionLimit:   aws.Int64(snapshotRetentionLimit),
				SnapshotWindow:           aws.String(snapshotWindow),
				SnapshottingClusterId:    aws.String(snapshottingClusterID),
				TransitEncryptionEnabled: &transitEncryptionEnabled,
			},
			want: &v1beta1.ReplicationGroupParameters{
				AtRestEncryptionEnabled:  &atRestEncryptionEnabled,
				AuthEnabled:              &authEnabled,
				AutomaticFailoverEnabled: &autoFailoverEnabled,
				SnapshotRetentionLimit:   &snapshotRetentionLimit,
				SnapshotWindow:           &snapshotWindow,
				SnapshottingClusterID:    &snapshottingClusterID,
				TransitEncryptionEnabled: &transitEncryptionEnabled,
			},
		},
		{
			name:   "AllChanged",
			params: &v1beta1.ReplicationGroupParameters{},
			rg: elasticache.ReplicationGroup{
				AtRestEncryptionEnabled:  &atRestEncryptionEnabled,
				AuthTokenEnabled:         &authEnabled,
				AutomaticFailover:        elasticache.AutomaticFailoverStatusEnabled,
				SnapshotRetentionLimit:   aws.Int64(snapshotRetentionLimit),
				SnapshotWindow:           aws.String(snapshotWindow),
				SnapshottingClusterId:    aws.String(snapshottingClusterID),
				TransitEncryptionEnabled: &transitEncryptionEnabled,
			},
			want: &v1beta1.ReplicationGroupParameters{
				AtRestEncryptionEnabled:  &atRestEncryptionEnabled,
				AuthEnabled:              &authEnabled,
				AutomaticFailoverEnabled: &autoFailoverEnabled,
				SnapshotRetentionLimit:   &snapshotRetentionLimit,
				SnapshotWindow:           &snapshotWindow,
				SnapshottingClusterID:    &snapshottingClusterID,
				TransitEncryptionEnabled: &transitEncryptionEnabled,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			LateInitialize(tc.params, tc.rg)
			if diff := cmp.Diff(tc.want, tc.params); diff != "" {
				t.Errorf("LateInitialize(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGenerateObservation(t *testing.T) {
	automaticFailover := elasticache.AutomaticFailoverStatusEnabled
	clusterEnabled := true
	configurationEndpoint := &elasticache.Endpoint{
		Address: aws.String("istanbul"),
		Port:    aws.Int64(34),
	}
	memberClusters := []string{"member-1", "member-2"}
	status := "creating"
	nodeGroups := []elasticache.NodeGroup{
		{
			NodeGroupId: aws.String("my-id"),
			Slots:       aws.String("special-slots"),
			Status:      aws.String("creating"),
			PrimaryEndpoint: &elasticache.Endpoint{
				Address: aws.String("random-12"),
				Port:    aws.Int64(124),
			},
			NodeGroupMembers: []elasticache.NodeGroupMember{
				{
					CacheClusterId:            aws.String("my-cache-cluster"),
					CacheNodeId:               aws.String("cluster-0001"),
					CurrentRole:               aws.String("secret-role"),
					PreferredAvailabilityZone: aws.String("us-east-1"),
					ReadEndpoint: &elasticache.Endpoint{
						Address: aws.String("random-1"),
						Port:    aws.Int64(123),
					},
				},
			},
		},
	}
	percentage := float64(54)
	rgpmdv := elasticache.ReplicationGroupPendingModifiedValues{
		AutomaticFailoverStatus: elasticache.PendingAutomaticFailoverStatusEnabled,
		PrimaryClusterId:        aws.String("my-coolest-cluster"),
		Resharding: &elasticache.ReshardingStatus{
			SlotMigration: &elasticache.SlotMigration{
				ProgressPercentage: &percentage,
			},
		},
	}
	cases := []struct {
		name string
		rg   elasticache.ReplicationGroup
		want v1beta1.ReplicationGroupObservation
	}{
		{
			name: "AllFields",
			rg: elasticache.ReplicationGroup{
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
					Port:    int(*configurationEndpoint.Port),
				},
				MemberClusters: memberClusters,
				NodeGroups: []v1beta1.NodeGroup{
					generateNodeGroup(nodeGroups[0]),
				},
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
		kube   v1beta1.ReplicationGroupParameters
		rg     elasticache.ReplicationGroup
		ccList []elasticache.CacheCluster
		want   bool
	}{
		{
			name: "NeedsFailoverEnabled",
			kube: replicationGroup.Spec.ForProvider,
			rg:   elasticache.ReplicationGroup{AutomaticFailover: elasticache.AutomaticFailoverStatusDisabled},
			want: true,
		},
		{
			name: "NeedsNewCacheNodeType",
			kube: replicationGroup.Spec.ForProvider,
			rg: elasticache.ReplicationGroup{
				AutomaticFailover: elasticache.AutomaticFailoverStatusEnabling,
				CacheNodeType:     aws.String("n1.insufficiently.cool"),
			},
			want: true,
		},
		{
			name: "NeedsNewSnapshotRetentionLimit",
			kube: replicationGroup.Spec.ForProvider,
			rg: elasticache.ReplicationGroup{
				AutomaticFailover:      elasticache.AutomaticFailoverStatusEnabling,
				CacheNodeType:          aws.String(cacheNodeType),
				SnapshotRetentionLimit: aws.Int64(snapshotRetentionLimit + 1),
			},
			want: true,
		},
		{
			name: "NeedsNewSnapshotWindow",
			kube: replicationGroup.Spec.ForProvider,
			rg: elasticache.ReplicationGroup{
				AutomaticFailover:      elasticache.AutomaticFailoverStatusEnabling,
				CacheNodeType:          aws.String(cacheNodeType),
				SnapshotRetentionLimit: aws.Int64(snapshotRetentionLimit),
				SnapshotWindow:         aws.String("yesterday"),
			},
			want: true,
		},
		{
			name: "CacheClusterNeedsUpdate",
			kube: replicationGroup.Spec.ForProvider,
			rg: elasticache.ReplicationGroup{
				AutomaticFailover:      elasticache.AutomaticFailoverStatusEnabling,
				CacheNodeType:          aws.String(cacheNodeType),
				SnapshotRetentionLimit: aws.Int64(snapshotRetentionLimit),
				SnapshotWindow:         aws.String(snapshotWindow),
			},
			ccList: []elasticache.CacheCluster{
				{
					EngineVersion: aws.String("4.0.0"),
				},
			},
			want: true,
		},
		{
			name: "NeedsNoUpdate",
			kube: replicationGroup.Spec.ForProvider,
			rg: elasticache.ReplicationGroup{
				AutomaticFailover:      elasticache.AutomaticFailoverStatusEnabling,
				CacheNodeType:          aws.String(cacheNodeType),
				SnapshotRetentionLimit: aws.Int64(snapshotRetentionLimit),
				SnapshotWindow:         aws.String(snapshotWindow),
			},
			ccList: []elasticache.CacheCluster{
				{
					EngineVersion:              aws.String(engineVersion),
					CacheParameterGroup:        &elasticache.CacheParameterGroupStatus{CacheParameterGroupName: aws.String(cacheParameterGroupName)},
					NotificationConfiguration:  &elasticache.NotificationConfiguration{TopicArn: aws.String(notificationTopicARN), TopicStatus: aws.String(notificationTopicStatus)},
					PreferredMaintenanceWindow: aws.String(maintenanceWindow),
					SecurityGroups: func() []elasticache.SecurityGroupMembership {
						ids := make([]elasticache.SecurityGroupMembership, len(securityGroupIDs))
						for i, id := range securityGroupIDs {
							ids[i] = elasticache.SecurityGroupMembership{SecurityGroupId: aws.String(id)}
						}
						return ids
					}(),
					CacheSecurityGroups: func() []elasticache.CacheSecurityGroupMembership {
						names := make([]elasticache.CacheSecurityGroupMembership, len(cacheSecurityGroupNames))
						for i, n := range cacheSecurityGroupNames {
							names[i] = elasticache.CacheSecurityGroupMembership{CacheSecurityGroupName: aws.String(n)}
						}
						return names
					}(),
				},
			},
			want: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := ReplicationGroupNeedsUpdate(tc.kube, tc.rg, tc.ccList)
			if got != tc.want {
				t.Errorf("ReplicationGroupNeedsUpdate(...): want %t, got %t", tc.want, got)
			}
		})
	}
}

func TestCacheClusterNeedsUpdate(t *testing.T) {
	cases := []struct {
		name string
		kube v1beta1.ReplicationGroupParameters
		cc   elasticache.CacheCluster
		want bool
	}{
		{
			name: "NeedsNewEngineVersion",
			kube: replicationGroup.Spec.ForProvider,
			cc: elasticache.CacheCluster{
				EngineVersion: aws.String("4.0.0"),
			},
			want: true,
		},
		{
			name: "NeedsNewCacheParameterGroup",
			kube: replicationGroup.Spec.ForProvider,
			cc: elasticache.CacheCluster{
				EngineVersion:       aws.String(engineVersion),
				CacheParameterGroup: &elasticache.CacheParameterGroupStatus{CacheParameterGroupName: aws.String("okaygroupiguess")},
			},
			want: true,
		},
		{
			name: "NeedsNewNotificationTopicARN",
			kube: replicationGroup.Spec.ForProvider,
			cc: elasticache.CacheCluster{
				EngineVersion:             aws.String(engineVersion),
				CacheParameterGroup:       &elasticache.CacheParameterGroupStatus{CacheParameterGroupName: aws.String(cacheParameterGroupName)},
				NotificationConfiguration: &elasticache.NotificationConfiguration{TopicArn: aws.String("aws:arn:sqs:nope"), TopicStatus: aws.String(notificationTopicStatus)},
			},
			want: true,
		},
		{
			name: "NeedsNewMaintenanceWindow",
			kube: replicationGroup.Spec.ForProvider,
			cc: elasticache.CacheCluster{
				EngineVersion:              aws.String(engineVersion),
				CacheParameterGroup:        &elasticache.CacheParameterGroupStatus{CacheParameterGroupName: aws.String(cacheParameterGroupName)},
				NotificationConfiguration:  &elasticache.NotificationConfiguration{TopicArn: aws.String(notificationTopicARN), TopicStatus: aws.String(notificationTopicStatus)},
				PreferredMaintenanceWindow: aws.String("never!"),
			},
			want: true,
		},
		{
			name: "NeedsNewSecurityGroupIDs",
			kube: replicationGroup.Spec.ForProvider,
			cc: elasticache.CacheCluster{
				EngineVersion:              aws.String(engineVersion),
				CacheParameterGroup:        &elasticache.CacheParameterGroupStatus{CacheParameterGroupName: aws.String(cacheParameterGroupName)},
				NotificationConfiguration:  &elasticache.NotificationConfiguration{TopicArn: aws.String(notificationTopicARN), TopicStatus: aws.String(notificationTopicStatus)},
				PreferredMaintenanceWindow: aws.String(maintenanceWindow),
				SecurityGroups: []elasticache.SecurityGroupMembership{
					{SecurityGroupId: aws.String("notaverysecuregroupid")},
					{SecurityGroupId: aws.String("evenlesssecuregroupid")},
				},
			},
			want: true,
		},
		{
			name: "NeedsSecurityGroupIDs",
			kube: replicationGroup.Spec.ForProvider,
			cc: elasticache.CacheCluster{
				EngineVersion:              aws.String(engineVersion),
				CacheParameterGroup:        &elasticache.CacheParameterGroupStatus{CacheParameterGroupName: aws.String(cacheParameterGroupName)},
				NotificationConfiguration:  &elasticache.NotificationConfiguration{TopicArn: aws.String(notificationTopicARN), TopicStatus: aws.String(notificationTopicStatus)},
				PreferredMaintenanceWindow: aws.String(maintenanceWindow),
			},
			want: true,
		},
		{
			name: "NeedsNewSecurityGroupNames",
			kube: replicationGroup.Spec.ForProvider,
			cc: elasticache.CacheCluster{
				EngineVersion:              aws.String(engineVersion),
				CacheParameterGroup:        &elasticache.CacheParameterGroupStatus{CacheParameterGroupName: aws.String(cacheParameterGroupName)},
				NotificationConfiguration:  &elasticache.NotificationConfiguration{TopicArn: aws.String(notificationTopicARN), TopicStatus: aws.String(notificationTopicStatus)},
				PreferredMaintenanceWindow: aws.String(maintenanceWindow),
				SecurityGroups: func() []elasticache.SecurityGroupMembership {
					ids := make([]elasticache.SecurityGroupMembership, len(securityGroupIDs))
					for i, id := range securityGroupIDs {
						ids[i] = elasticache.SecurityGroupMembership{SecurityGroupId: aws.String(id)}
					}
					return ids
				}(),
				CacheSecurityGroups: []elasticache.CacheSecurityGroupMembership{
					{CacheSecurityGroupName: aws.String("notaverysecuregroup")},
					{CacheSecurityGroupName: aws.String("evenlesssecuregroup")},
				},
			},
			want: true,
		},
		{
			name: "NeedsSecurityGroupNames",
			kube: replicationGroup.Spec.ForProvider,
			cc: elasticache.CacheCluster{
				EngineVersion:              aws.String(engineVersion),
				CacheParameterGroup:        &elasticache.CacheParameterGroupStatus{CacheParameterGroupName: aws.String(cacheParameterGroupName)},
				NotificationConfiguration:  &elasticache.NotificationConfiguration{TopicArn: aws.String(notificationTopicARN), TopicStatus: aws.String(notificationTopicStatus)},
				PreferredMaintenanceWindow: aws.String(maintenanceWindow),
				SecurityGroups: func() []elasticache.SecurityGroupMembership {
					ids := make([]elasticache.SecurityGroupMembership, len(securityGroupIDs))
					for i, id := range securityGroupIDs {
						ids[i] = elasticache.SecurityGroupMembership{SecurityGroupId: aws.String(id)}
					}
					return ids
				}(),
			},
			want: true,
		},
		{
			name: "NeedsNoUpdate",
			kube: replicationGroup.Spec.ForProvider,
			cc: elasticache.CacheCluster{
				EngineVersion:              aws.String(engineVersion),
				CacheParameterGroup:        &elasticache.CacheParameterGroupStatus{CacheParameterGroupName: aws.String(cacheParameterGroupName)},
				NotificationConfiguration:  &elasticache.NotificationConfiguration{TopicArn: aws.String(notificationTopicARN), TopicStatus: aws.String(notificationTopicStatus)},
				PreferredMaintenanceWindow: aws.String(maintenanceWindow),
				SecurityGroups: func() []elasticache.SecurityGroupMembership {
					ids := make([]elasticache.SecurityGroupMembership, len(securityGroupIDs))
					for i, id := range securityGroupIDs {
						ids[i] = elasticache.SecurityGroupMembership{SecurityGroupId: aws.String(id)}
					}
					return ids
				}(),
				CacheSecurityGroups: func() []elasticache.CacheSecurityGroupMembership {
					names := make([]elasticache.CacheSecurityGroupMembership, len(cacheSecurityGroupNames))
					for i, n := range cacheSecurityGroupNames {
						names[i] = elasticache.CacheSecurityGroupMembership{CacheSecurityGroupName: aws.String(n)}
					}
					return names
				}(),
			},
			want: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := cacheClusterNeedsUpdate(tc.kube, tc.cc)
			if got != tc.want {
				t.Errorf("cacheClusterNeedsUpdate(...): want %t, got %t", tc.want, got)
			}
		})
	}

}

func TestConnectionEndpoint(t *testing.T) {
	cases := []struct {
		name string
		rg   elasticache.ReplicationGroup
		want managed.ConnectionDetails
	}{
		{
			name: "ClusterModeEnabled",
			rg: elasticache.ReplicationGroup{
				ClusterEnabled: aws.Bool(true),
				ConfigurationEndpoint: &elasticache.Endpoint{
					Address: aws.String(host),
					Port:    aws.Int64(port),
				},
			},
			want: managed.ConnectionDetails{
				v1alpha1.ResourceCredentialsSecretEndpointKey: []byte(host),
				v1alpha1.ResourceCredentialsSecretPortKey:     []byte(strconv.Itoa(port)),
			},
		},
		{
			name: "ClusterModeEnabledMissingConfigurationEndpoint",
			rg: elasticache.ReplicationGroup{
				ClusterEnabled: aws.Bool(true),
			},
			want: nil,
		},
		{
			name: "ClusterModeDisabled",
			rg: elasticache.ReplicationGroup{
				NodeGroups: []elasticache.NodeGroup{{
					PrimaryEndpoint: &elasticache.Endpoint{
						Address: aws.String(host),
						Port:    aws.Int64(port),
					}},
				},
			},
			want: managed.ConnectionDetails{
				v1alpha1.ResourceCredentialsSecretEndpointKey: []byte(host),
				v1alpha1.ResourceCredentialsSecretPortKey:     []byte(strconv.Itoa(port)),
			},
		},
		{
			name: "ClusterModeDisabledMissingPrimaryEndpoint",
			rg:   elasticache.ReplicationGroup{NodeGroups: []elasticache.NodeGroup{{}}},
			want: nil,
		},
		{
			name: "ClusterModeDisabledMissingNodeGroups",
			rg:   elasticache.ReplicationGroup{},
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
		subnetGroup elasticache.CacheSubnetGroup
		p           cachev1alpha1.CacheSubnetGroupParameters
	}

	cases := map[string]struct {
		args args
		want bool
	}{
		"SameFields": {
			args: args{
				subnetGroup: elasticache.CacheSubnetGroup{
					CacheSubnetGroupDescription: aws.String(subnetGroupDesc),
					Subnets: []elasticache.Subnet{
						{
							SubnetIdentifier: aws.String(subnetID1),
						},
						{
							SubnetIdentifier: aws.String(subnetID2),
						},
					},
				},
				p: cachev1alpha1.CacheSubnetGroupParameters{
					Description: subnetGroupDesc,
					SubnetIds:   []string{subnetID1, subnetID2},
				},
			},
			want: true,
		},
		"DifferentFields": {
			args: args{
				subnetGroup: elasticache.CacheSubnetGroup{
					CacheSubnetGroupDescription: aws.String(subnetGroupDesc),
					Subnets: []elasticache.Subnet{
						{
							SubnetIdentifier: aws.String(subnetID1),
						},
						{
							SubnetIdentifier: aws.String(subnetID2),
						},
					},
				},
				p: cachev1alpha1.CacheSubnetGroupParameters{
					Description: subnetGroupDesc,
					SubnetIds:   []string{subnetID1},
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
