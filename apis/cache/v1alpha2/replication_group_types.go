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

package v1alpha2

import (
	runtimev1alpha1 "github.com/crossplaneio/crossplane-runtime/apis/core/v1alpha1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ReplicationGroup states.
const (
	StatusCreating     = "creating"
	StatusAvailable    = "available"
	StatusModifying    = "modifying"
	StatusDeleting     = "deleting"
	StatusCreateFailed = "create-failed"
	StatusSnapshotting = "snapshotting"
)

// Supported cache engines.
const (
	CacheEngineRedis = "redis"
)

// TODO(negz): Lookup supported patch versions in the ElastiCache API?
// AWS requires we specify desired Redis versions down to the patch version,
// but the RedisCluster resource claim supports only minor versions (which are
// the lowest common denominator between supported clouds). We perform this
// lookup in the claim provisioning code, which does not have an AWS client
// plumbed in to perform such a lookup.
// https://docs.aws.amazon.com/AmazonElastiCache/latest/APIReference/API_DescribeCacheEngineVersions.html

// MinorVersion represents a supported minor version of Redis.
type MinorVersion string

// PatchVersion represents a supported patch version of Redis.
type PatchVersion string

// UnsupportedVersion indicates the requested MinorVersion is unsupported.
const UnsupportedVersion PatchVersion = ""

// LatestSupportedPatchVersion returns the latest supported patch version
// for a given minor version.
var LatestSupportedPatchVersion = map[MinorVersion]PatchVersion{
	MinorVersion("5.0"): PatchVersion("5.0.0"),
	MinorVersion("4.0"): PatchVersion("4.0.10"),
	MinorVersion("3.2"): PatchVersion("3.2.10"),
	MinorVersion("2.8"): PatchVersion("2.8.24"),
}

// ReplicationGroupParameters define the desired state of an AWS ElastiCache
// Replication Group. Most fields map directly to an AWS ReplicationGroup:
// https://docs.aws.amazon.com/AmazonElastiCache/latest/APIReference/API_CreateReplicationGroup.html#API_CreateReplicationGroup_RequestParameters
type ReplicationGroupParameters struct {
	// AtRestEncryptionEnabled enables encryption at rest when set to true.
	//
	// You cannot modify the value of AtRestEncryptionEnabled after the replication
	// group is created. To enable encryption at rest on a replication group you
	// must set AtRestEncryptionEnabled to true when you create the replication
	// group.
	// +optional
	AtRestEncryptionEnabled bool `json:"atRestEncryptionEnabled,omitempty"`

	// AuthEnabled enables mandatory authentication when connecting to the
	// managed replication group. AuthEnabled requires TransitEncryptionEnabled
	// to be true.
	//
	// While ReplicationGroupSpec mirrors the fields of the upstream replication
	// group object as closely as possible, we expose a boolean here rather than
	// requiring the operator pass in a string authentication token. Crossplane
	// will generate a token automatically and expose it via a Secret.
	// +optional
	AuthEnabled bool `json:"authEnabled,omitempty"`

	// AutomaticFailoverEnabled specifies whether a read-only replica is
	// automatically promoted to read/write primary if the existing primary
	// fails. If true, Multi-AZ is enabled for this replication group. If false,
	// Multi-AZ is disabled for this replication group.
	//
	// AutomaticFailoverEnabled must be enabled for Redis (cluster mode enabled)
	// replication groups.
	// +optional
	AutomaticFailoverEnabled bool `json:"automaticFailoverEnabled,omitempty"`

	// CacheNodeType specifies the compute and memory capacity of the nodes in
	// the node group (shard).
	CacheNodeType string `json:"cacheNodeType"`

	// CacheParameterGroupName specifies the name of the parameter group to
	// associate with this replication group. If this argument is omitted, the
	// default cache parameter group for the specified engine is used.
	// +optional
	CacheParameterGroupName string `json:"cacheParameterGroupName,omitempty"`

	// CacheSecurityGroupNames specifies a list of cache security group names to
	// associate with this replication group.
	// +optional
	CacheSecurityGroupNames []string `json:"cacheSecurityGroupNames,omitempty"`

	// CacheSubnetGroupName specifies the name of the cache subnet group to be
	// used for the replication group. If you're going to launch your cluster in
	// an Amazon VPC, you need to create a subnet group before you start
	// creating a cluster.
	// +optional
	CacheSubnetGroupName string `json:"cacheSubnetGroupName,omitempty"`

	// EngineVersion specifies the version number of the cache engine to be
	// used for the clusters in this replication group. To view the supported
	// cache engine versions, use the DescribeCacheEngineVersions operation.
	// +optional
	EngineVersion string `json:"engineVersion,omitempty"`

	// NodeGroupConfiguration specifies a list of node group (shard)
	// configuration options.
	// +optional
	NodeGroupConfiguration []NodeGroupConfigurationSpec `json:"nodeGroupConfiguration,omitempty"`

	// NotificationTopicARN specifies the Amazon Resource Name (ARN) of the
	// Amazon Simple Notification Service (SNS) topic to which notifications are
	// sent. The Amazon SNS topic owner must be the same as the cluster owner.
	// +optional
	NotificationTopicARN string `json:"notificationTopicArn,omitempty"`

	// NumCacheClusters specifies the number of clusters this replication group
	// initially has. This parameter is not used if there is more than one node
	// group (shard). You should use ReplicasPerNodeGroup instead.
	//
	// If AutomaticFailoverEnabled is true, the value of this parameter must be
	// at least 2. If AutomaticFailoverEnabled is false you can omit this
	// parameter (it will default to 1), or you can explicitly set it to a value
	// between 2 and 6.
	// +optional
	NumCacheClusters int `json:"numCacheClusters,omitempty"`

	// NumNodeGroups specifies the number of node groups (shards) for this Redis
	// (cluster mode enabled) replication group. For Redis (cluster mode
	// disabled) either omit this parameter or set it to 1.
	// +optional
	NumNodeGroups int `json:"numNodeGroups,omitempty"`

	// Port number on which each member of the replication group accepts
	// connections.
	// +optional
	Port int `json:"port,omitempty"`

	// PreferredCacheClusterAZs specifies a list of EC2 Availability Zones in
	// which the replication group's clusters are created. The order of the
	// Availability Zones in the list is the order in which clusters are
	// allocated. The primary cluster is created in the first AZ in the list.
	//
	// This parameter is not used if there is more than one node group (shard).
	// You should use NodeGroupConfiguration instead.
	//
	// The number of Availability Zones listed must equal the value of
	// NumCacheClusters.
	// +optional
	PreferredCacheClusterAZs []string `json:"preferredCacheClusterAzs,omitempty"`

	// PreferredMaintenanceWindow specifies the weekly time range during which
	// maintenance on the cluster is performed. It is specified as a range in
	// the format ddd:hh24:mi-ddd:hh24:mi (24H Clock UTC). The minimum
	// maintenance window is a 60 minute period.
	//
	// Example: sun:23:00-mon:01:30
	// +optional
	PreferredMaintenanceWindow string `json:"preferredMaintenanceWindow,omitempty"`

	// ReplicasPerNodeGroup specifies the number of replica nodes in each node
	// group (shard). Valid values are 0 to 5.
	// +optional
	ReplicasPerNodeGroup int `json:"replicasPerNodeGroup,omitempty"`

	// SecurityGroupIDs specifies one or more Amazon VPC security groups
	// associated with this replication group. Use this parameter only when you
	// are creating a replication group in an Amazon VPC.
	// +optional
	SecurityGroupIDs []string `json:"securityGroupIds,omitempty"`

	// SnapshotARNs specifies a list of Amazon Resource Names (ARN) that
	// uniquely identify the Redis RDB snapshot files stored in Amazon S3. The
	// snapshot files are used to populate the new replication group. The Amazon
	// S3 object name in the ARN cannot contain any commas. The new replication
	// group will have the number of node groups (console: shards) specified by
	// the parameter NumNodeGroups or the number of node groups configured by
	// NodeGroupConfiguration regardless of the number of ARNs specified here.
	// +optional
	SnapshotARNs []string `json:"snapshotArns,omitempty"`

	// SnapshotName specifies the name of a snapshot from which to restore data
	// into the new replication group. The snapshot status changes to restoring
	// while the new replication group is being created.
	// +optional
	SnapshotName string `json:"snapshotName,omitempty"`

	// SnapshotRetentionLimit specifies the number of days for which ElastiCache
	// retains automatic snapshots before deleting them. For example, if you set
	// SnapshotRetentionLimit to 5, a snapshot that was taken today is retained
	// for 5 days before being deleted.
	// +optional
	SnapshotRetentionLimit int `json:"snapshotRetentionLimit,omitempty"`

	// SnapshotWindow specifies the daily time range (in UTC) during which
	// ElastiCache begins taking a daily snapshot of your node group (shard).
	//
	// Example: 05:00-09:00
	//
	// If you do not specify this parameter, ElastiCache automatically chooses an
	// appropriate time range.
	// +optional
	SnapshotWindow string `json:"snapshotWindow,omitempty"`

	// TransitEncryptionEnabled enables in-transit encryption when set to true.
	//
	// You cannot modify the value of TransitEncryptionEnabled after the cluster
	// is created. To enable in-transit encryption on a cluster you must
	// TransitEncryptionEnabled to true when you create a cluster.
	// +optional
	TransitEncryptionEnabled bool `json:"transitEncryptionEnabled,omitempty"`
}

// A ReplicationGroupSpec defines the desired state of a ReplicationGroup.
type ReplicationGroupSpec struct {
	runtimev1alpha1.ResourceSpec `json:",inline"`
	ReplicationGroupParameters   `json:",inline"`
}

// A NodeGroupConfigurationSpec specifies the desired state of a node group.
type NodeGroupConfigurationSpec struct {
	// PrimaryAvailabilityZone specifies the Availability Zone where the primary
	// node of this node group (shard) is launched.
	// +optional
	PrimaryAvailabilityZone string `json:"primaryAvailabilityZone,omitempty"`

	// ReplicaAvailabilityZones specifies a list of Availability Zones to be
	// used for the read replicas. The number of Availability Zones in this list
	// must match the value of ReplicaCount or ReplicasPerNodeGroup if not
	// specified.
	// +optional
	ReplicaAvailabilityZones []string `json:"replicaAvailabilityZones,omitempty"`

	// ReplicaCount specifies the number of read replica nodes in this node
	// group (shard).
	// +optional
	ReplicaCount int `json:"replicaCount,omitempty"`

	// Slots specifies the keyspace for a particular node group. Keyspaces range
	// from 0 to 16,383. The string is in the format startkey-endkey.
	//
	// Example: "0-3999"
	// +optional
	Slots string `json:"slots,omitempty"`
}

// A ReplicationGroupStatus defines the observed state of a ReplicationGroup.
type ReplicationGroupStatus struct {
	runtimev1alpha1.ResourceStatus `json:",inline"`

	// State of the Replication Group.
	State string `json:"state,omitempty"`

	// ProviderID is the external ID to identify this resource in the cloud
	// provider
	ProviderID string `json:"providerID,omitempty"`

	// Endpoint of the Replication Group used in connection strings.
	Endpoint string `json:"endpoint,omitempty"`

	// Port at which the Replication Group endpoint is listening.
	Port int `json:"port,omitempty"`

	// ClusterEnabled indicates whether cluster mode is enabled, i.e. whether
	// this replication group's data can be partitioned across multiple shards.
	ClusterEnabled bool `json:"clusterEnabled,omitempty"`

	// MemberClusters that are part of this replication group.
	MemberClusters []string `json:"memberClusters,omitempty"`

	// Groupname of the Replication Group.
	GroupName string `json:"groupName,omitempty"`

	// TODO(negz): Support PendingModifiedValues?
	// https://docs.aws.amazon.com/AmazonElastiCache/latest/APIReference/API_ReplicationGroupPendingModifiedValues.html
}

// +kubebuilder:object:root=true

// A ReplicationGroup is a managed resource that represents an AWS ElastiCache
// Replication Group.
// +kubebuilder:printcolumn:name="STATUS",type="string",JSONPath=".status.state"
// +kubebuilder:printcolumn:name="CLASS",type="string",JSONPath=".spec.classRef.name"
// +kubebuilder:printcolumn:name="VERSION",type="string",JSONPath=".spec.engineVersion"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
type ReplicationGroup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ReplicationGroupSpec   `json:"spec,omitempty"`
	Status ReplicationGroupStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ReplicationGroupList contains a list of ReplicationGroup
type ReplicationGroupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ReplicationGroup `json:"items"`
}

// A ReplicationGroupClassSpecTemplate is a template for the spec of a
// dynamically provisioned ReplicationGroup.
type ReplicationGroupClassSpecTemplate struct {
	runtimev1alpha1.NonPortableClassSpecTemplate `json:",inline"`
	ReplicationGroupParameters                   `json:",inline"`
}

// +kubebuilder:object:root=true

// A ReplicationGroupClass is a non-portable resource class. It defines the
// desired spec of resource claims that use it to dynamically provision a
// managed resource.
// +kubebuilder:printcolumn:name="PROVIDER-REF",type="string",JSONPath=".specTemplate.providerRef.name"
// +kubebuilder:printcolumn:name="RECLAIM-POLICY",type="string",JSONPath=".specTemplate.reclaimPolicy"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
type ReplicationGroupClass struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// SpecTemplate is a template for the spec of a dynamically provisioned
	// ReplicationGroup.
	SpecTemplate ReplicationGroupClassSpecTemplate `json:"specTemplate"`
}

// +kubebuilder:object:root=true

// ReplicationGroupClassList contains a list of cloud memorystore resource classes.
type ReplicationGroupClassList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ReplicationGroupClass `json:"items"`
}
