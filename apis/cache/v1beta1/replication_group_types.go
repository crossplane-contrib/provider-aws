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

package v1beta1

import (
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
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
	CacheEngineRedis     = "redis"
	CacheEngineMemcached = "memcached"
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

// Endpoint represents the information required for client programs to connect
// to a cache node.
// Please also see https://docs.aws.amazon.com/goto/WebAPI/elasticache-2015-02-02/Endpoint
type Endpoint struct {
	// Address is the DNS hostname of the cache node.
	Address string `json:"address,omitempty"`

	// Port number that the cache engine is listening on.
	Port int `json:"port,omitempty"`
}

// NodeGroup represents a collection of cache nodes in a replication group.
// One node in the node group is the read/write primary node. All the other
// nodes are read-only Replica nodes.
// Please also see https://docs.aws.amazon.com/goto/WebAPI/elasticache-2015-02-02/NodeGroup
type NodeGroup struct {
	// NodeGroupID is the identifier for the node group (shard). A Redis
	// (cluster mode disabled) replication group contains only 1 node group;
	// therefore, the node group ID is 0001. A Redis (cluster mode enabled)
	// replication group contains 1 to 15 node groups numbered 0001 to 0015.
	NodeGroupID string `json:"nodeGroupID,omitempty"`

	// NodeGroupMembers is a list containing information about individual nodes
	// within the node group (shard).
	NodeGroupMembers []NodeGroupMember `json:"nodeGroupMembers,omitempty"`

	// PrimaryEndpoint is the endpoint of the primary node in this
	// node group (shard).
	PrimaryEndpoint Endpoint `json:"primaryEndpoint,omitempty"`

	// ReaderEndpoint is the endpoint of the replica nodes in this node group (shard).
	ReaderEndpoint Endpoint `json:"readerEndpoint,omitempty"`

	// Slots is the keyspace for this node group (shard).
	Slots string `json:"slots,omitempty"`

	// Status of this replication group - creating, available, etc.
	Status string `json:"status,omitempty"`
}

// NodeGroupMember represents a single node within a node group (shard).
// Please also see https://docs.aws.amazon.com/goto/WebAPI/elasticache-2015-02-02/NodeGroupMember
type NodeGroupMember struct {
	// CacheClusterID is the ID of the cluster to which the node belongs.
	CacheClusterID string `json:"cacheClusterId,omitempty"`

	// CacheNodeID is the ID of the node within its cluster. A node ID is a
	// numeric identifier (0001, 0002, etc.).
	CacheNodeID string `json:"cacheNodeId,omitempty"`

	// CurrentRole is the role that is currently assigned to the node - primary
	// or replica. This member is only applicable for Redis (cluster mode
	// disabled) replication groups.
	CurrentRole string `json:"currentRole,omitempty"`

	// PreferredAvailabilityZone is the name of the Availability Zone in
	// which the node is located.
	PreferredAvailabilityZone string `json:"preferredAvailabilityZone,omitempty"`

	// ReadEndpoint is the information required for client programs to connect to a
	// node for read operations. The read endpoint is only applicable on Redis
	// (cluster mode disabled) clusters.
	ReadEndpoint Endpoint `json:"readEndpoint,omitempty"`
}

// ReplicationGroupPendingModifiedValues are the settings to be applied to the
// Redis replication group, either immediately or during the next maintenance
// window. Please also see
// https://docs.aws.amazon.com/goto/WebAPI/elasticache-2015-02-02/ReplicationGroupPendingModifiedValues
type ReplicationGroupPendingModifiedValues struct {
	// AutomaticFailoverStatus indicates the status of Multi-AZ with automatic
	// failover for this Redis replication group.
	AutomaticFailoverStatus string `json:"automaticFailoverStatus,omitempty"`

	// PrimaryClusterID that is applied immediately or during the next
	// maintenance window.
	PrimaryClusterID string `json:"primaryClusterId,omitempty"`

	// Resharding is the status of an online resharding operation.
	Resharding ReshardingStatus `json:"resharding,omitempty"`
}

// ReshardingStatus is the status of an online resharding operation.
// Please also see https://docs.aws.amazon.com/goto/WebAPI/elasticache-2015-02-02/ReshardingStatus
type ReshardingStatus struct {
	// Represents the progress of an online resharding operation.
	SlotMigration SlotMigration `json:"slotMigration"`
}

// SlotMigration represents the progress of an online resharding operation.
// Please also see https://docs.aws.amazon.com/goto/WebAPI/elasticache-2015-02-02/SlotMigration
type SlotMigration struct {
	// NOTE(muvaf): Type of ProgressPercentage is float64 in AWS SDK but
	// float is not supported by controller-runtime.
	// See https://github.com/kubernetes-sigs/controller-tools/issues/245

	// ProgressPercentage is the percentage of the slot migration
	// that is complete.
	ProgressPercentage int `json:"progressPercentage"`
}

// ReplicationGroupObservation contains the observation of the status of
// the given ReplicationGroup.
type ReplicationGroupObservation struct {
	// AutomaticFailover indicates the status of Multi-AZ with automatic failover
	// for this Redis replication group.
	AutomaticFailover string `json:"automaticFailoverStatus,omitempty"`

	// ClusterEnabled is a flag indicating whether or not this replication group
	// is cluster enabled; i.e., whether its data can be partitioned across
	// multiple shards (API/CLI: node groups).
	ClusterEnabled bool `json:"clusterEnabled,omitempty"`

	// ConfigurationEndpoint for this replication group. Use the configuration
	// endpoint to connect to this replication group.
	ConfigurationEndpoint Endpoint `json:"configurationEndpoint,omitempty"`

	// MemberClusters is the list of names of all the cache clusters that are
	// part of this replication group.
	MemberClusters []string `json:"memberClusters,omitempty"`

	// NodeGroups is a list of node groups in this replication group.
	// For Redis (cluster mode disabled) replication groups, this is a
	// single-element list. For Redis (cluster mode enabled) replication groups,
	// the list contains an entry for each node group (shard).
	NodeGroups []NodeGroup `json:"nodeGroups,omitempty"`

	// PendingModifiedValues is a group of settings to be applied to the
	// replication group, either immediately or during the next maintenance window.
	PendingModifiedValues ReplicationGroupPendingModifiedValues `json:"pendingModifiedValues,omitempty"`

	// Status is the current state of this replication group - creating,
	// available, modifying, deleting, create-failed, snapshotting.
	Status string `json:"status,omitempty"`
}

// A Tag is used to tag the ElastiCache resources in AWS.
type Tag struct {
	// Key for the tag.
	Key string `json:"key"`

	// Value of the tag.
	Value string `json:"value"`
}

// A NodeGroupConfigurationSpec specifies the desired state of a node group.
type NodeGroupConfigurationSpec struct {
	// PrimaryAvailabilityZone specifies the Availability Zone where the primary
	// node of this node group (shard) is launched.
	// +optional
	PrimaryAvailabilityZone *string `json:"primaryAvailabilityZone,omitempty"`

	// ReplicaAvailabilityZones specifies a list of Availability Zones to be
	// used for the read replicas. The number of Availability Zones in this list
	// must match the value of ReplicaCount or ReplicasPerNodeGroup if not
	// specified.
	// +optional
	ReplicaAvailabilityZones []string `json:"replicaAvailabilityZones,omitempty"`

	// ReplicaCount specifies the number of read replica nodes in this node
	// group (shard).
	// +optional
	ReplicaCount *int `json:"replicaCount,omitempty"`

	// Slots specifies the keyspace for a particular node group. Keyspaces range
	// from 0 to 16,383. The string is in the format startkey-endkey.
	//
	// Example: "0-3999"
	// +optional
	Slots *string `json:"slots,omitempty"`
}

// ReplicationGroupParameters define the desired state of an AWS ElastiCache
// Replication Group. Most fields map directly to an AWS ReplicationGroup:
// https://docs.aws.amazon.com/AmazonElastiCache/latest/APIReference/API_CreateReplicationGroup.html#API_CreateReplicationGroup_RequestParameters
type ReplicationGroupParameters struct {
	// TODO(muvaf): Region is a required field but in order to keep backward compatibility
	// with old Provider type and not bear the cost of bumping to v1beta2, we're
	// keeping it optional for now. Reconsider before v1beta2 or v1.

	// Region is the region you'd like your ReplicationGroup to be created in.
	// +optional
	Region *string `json:"region,omitempty"`

	// If true, this parameter causes the modifications in this request and any
	// pending modifications to be applied, asynchronously and as soon as possible,
	// regardless of the PreferredMaintenanceWindow setting for the replication
	// group.
	//
	// If false, changes to the nodes in the replication group are applied on the
	// next maintenance reboot, or the next failure reboot, whichever occurs first.
	ApplyModificationsImmediately bool `json:"applyModificationsImmediately"`

	// AtRestEncryptionEnabled enables encryption at rest when set to true.
	//
	// You cannot modify the value of AtRestEncryptionEnabled after the replication
	// group is created. To enable encryption at rest on a replication group you
	// must set AtRestEncryptionEnabled to true when you create the replication
	// group.
	//
	// Only available when creating a replication group in an Amazon VPC
	// using redis version 3.2.6 or 4.x.
	//
	// +immutable
	// +optional
	AtRestEncryptionEnabled *bool `json:"atRestEncryptionEnabled,omitempty"`

	// AuthEnabled enables mandatory authentication when connecting to the
	// managed replication group. AuthEnabled requires TransitEncryptionEnabled
	// to be true.
	//
	// While ReplicationGroupSpec mirrors the fields of the upstream replication
	// group object as closely as possible, we expose a boolean here rather than
	// requiring the operator pass in a string authentication token. Crossplane
	// will generate a token automatically and expose it via a Secret.
	// +immutable
	// +optional
	AuthEnabled *bool `json:"authEnabled,omitempty"`

	// AutomaticFailoverEnabled specifies whether a read-only replica is
	// automatically promoted to read/write primary if the existing primary
	// fails. Must be set to true if Multi-AZ is enabled for this replication group.
	// If false, Multi-AZ cannot be enabled for this replication group.
	//
	// AutomaticFailoverEnabled must be enabled for Redis (cluster mode enabled)
	// replication groups.
	//
	// Amazon ElastiCache for Redis does not support Multi-AZ with automatic
	// failover on:
	// * Redis versions earlier than 2.8.6.
	// * Redis (cluster mode disabled): T1 and T2 cache node types.
	// * Redis (cluster mode enabled): T1 node types.
	// +optional
	AutomaticFailoverEnabled *bool `json:"automaticFailoverEnabled,omitempty"`

	// CacheNodeType specifies the compute and memory capacity of the nodes in
	// the node group (shard).
	// For a complete listing of node types and specifications, see:
	// * Amazon ElastiCache Product Features and Details (http://aws.amazon.com/elasticache/details)
	// * Cache Node Type-Specific Parameters for Memcached (http://docs.aws.amazon.com/AmazonElastiCache/latest/mem-ug/ParameterGroups.Memcached.html#ParameterGroups.Memcached.NodeSpecific)
	// * Cache Node Type-Specific Parameters for Redis (http://docs.aws.amazon.com/AmazonElastiCache/latest/red-ug/ParameterGroups.Redis.html#ParameterGroups.Redis.NodeSpecific)
	CacheNodeType string `json:"cacheNodeType"`

	// CacheParameterGroupName specifies the name of the parameter group to
	// associate with this replication group. If this argument is omitted, the
	// default cache parameter group for the specified engine is used.
	//
	// If you are running Redis version 3.2.4 or later, only one node group (shard),
	// and want to use a default parameter group, we recommend that you specify
	// the parameter group by name.
	// * To create a Redis (cluster mode disabled) replication group,
	// use CacheParameterGroupName=default.redis3.2.
	// * To create a Redis (cluster mode enabled) replication group,
	// use CacheParameterGroupName=default.redis3.2.cluster.on.
	// +optional
	CacheParameterGroupName *string `json:"cacheParameterGroupName,omitempty"`

	// CacheSecurityGroupNames specifies a list of cache security group names to
	// associate with this replication group. Only for EC2-Classic mode.
	// +optional
	CacheSecurityGroupNames []string `json:"cacheSecurityGroupNames,omitempty"`

	// CacheSecurityGroupNameRefs are references to SecurityGroups used to set
	// the CacheSecurityGroupNames.
	// +immutable
	// +optional
	CacheSecurityGroupNameRefs []xpv1.Reference `json:"cacheSecurityGroupNameRefs,omitempty"`

	// CacheSecurityGroupNameSelector selects references to SecurityGroups.
	// +immutable
	// +optional
	CacheSecurityGroupNameSelector *xpv1.Selector `json:"cacheSecurityGroupNameSelector,omitempty"`

	// CacheSubnetGroupName specifies the name of the cache subnet group to be
	// used for the replication group. If you're going to launch your cluster in
	// an Amazon VPC, you need to create a subnet group before you start
	// creating a cluster. For more information, see Subnets and Subnet Groups
	// (http://docs.aws.amazon.com/AmazonElastiCache/latest/red-ug/SubnetGroups.html).
	// +immutable
	// +optional
	CacheSubnetGroupName *string `json:"cacheSubnetGroupName,omitempty"`

	// CacheSubnetGroupNameRef is a reference to a Subnet Group used to set
	// the CacheSubnetGroupName.
	// +immutable
	// +optional
	CacheSubnetGroupNameRef *xpv1.Reference `json:"cacheSubnetGroupNameRef,omitempty"`

	// DeprecatedCacheSubnetGroupNameRef is a reference to a Subnet Group
	// used to set the CacheSubnetGroupName.
	//
	// Deprecated: Use CacheSubnetGroupNameRef. This field exists because we
	// introduced it with the JSON tag cacheSubnetGroupNameRefs (plural)
	// when it should have been cacheSubnetGroupNameRef (singular). This is
	// a bug that we need to avoid a breaking change to this v1beta1 API.
	// +immutable
	// +optional
	DeprecatedCacheSubnetGroupNameRef *xpv1.Reference `json:"cacheSubnetGroupNameRefs,omitempty"`

	// CacheSubnetGroupNameSelector selects a reference to a CacheSubnetGroup.
	// +immutable
	// +optional
	CacheSubnetGroupNameSelector *xpv1.Selector `json:"cacheSubnetGroupNameSelector,omitempty"`

	// Engine is the name of the cache engine (memcached or redis) to be used
	// for the clusters in this replication group.
	// +immutable
	Engine string `json:"engine"`

	// EngineVersion specifies the version number of the cache engine to be
	// used for the clusters in this replication group. To view the supported
	// cache engine versions, use the DescribeCacheEngineVersions operation.
	//
	// Important: You can upgrade to a newer engine version (see Selecting a Cache
	// Engine and Version (http://docs.aws.amazon.com/AmazonElastiCache/latest/red-ug/SelectEngine.html#VersionManagement))
	// in the ElastiCache User Guide, but you cannot downgrade to an earlier engine
	// version. If you want to use an earlier engine version, you must delete the
	// existing cluster or replication group and create it anew with the earlier
	// engine version.
	// +optional
	EngineVersion *string `json:"engineVersion,omitempty"`

	// MultiAZEnabled specifies if Multi-AZ is enabled to enhance fault tolerance
	// You must have nodes across two or more Availability Zones in order to enable
	// this feature.
	// If this feature is set, automaticFailoverEnabled must be set to true.
	// +optional
	MultiAZEnabled *bool `json:"multiAZEnabled,omitempty"`

	// NodeGroupConfigurationSpec specifies a list of node group (shard)
	// configuration options.
	//
	// If you're creating a Redis (cluster mode disabled) or a Redis (cluster mode
	// enabled) replication group, you can use this parameter to individually configure
	// each node group (shard), or you can omit this parameter. However, when seeding
	// a Redis (cluster mode enabled) cluster from a S3 rdb file, you must configure
	// each node group (shard) using this parameter because you must specify the
	// slots for each node group.
	// +immutable
	// +optional
	NodeGroupConfiguration []NodeGroupConfigurationSpec `json:"nodeGroupConfiguration,omitempty"`

	// NotificationTopicARN specifies the Amazon Resource Name (ARN) of the
	// Amazon Simple Notification Service (SNS) topic to which notifications are
	// sent. The Amazon SNS topic owner must be the same as the cluster owner.
	// +optional
	NotificationTopicARN *string `json:"notificationTopicArn,omitempty"`

	// NotificationTopicARNRef references an SNS Topic to retrieve its NotificationTopicARN
	// +optional
	NotificationTopicARNRef *xpv1.Reference `json:"notificationTopicArnRef,omitempty"`

	// NotificationTopicARNSelector selects a reference to an SNS Topic to retrieve its NotificationTopicARN
	// +optional
	NotificationTopicARNSelector *xpv1.Selector `json:"notificationTopicArnSelector,omitempty"`

	// NotificationTopicStatus is the status of the Amazon SNS notification
	// topic for the replication group. Notifications are sent only if the status
	// is active.
	//
	// Valid values: active | inactive
	// +optional
	NotificationTopicStatus *string `json:"notificationTopicStatus,omitempty"`

	// NumCacheClusters specifies the number of clusters this replication group
	// initially has. This parameter is not used if there is more than one node
	// group (shard). You should use ReplicasPerNodeGroup instead.
	//
	// If AutomaticFailoverEnabled is true, the value of this parameter must be
	// at least 2. If AutomaticFailoverEnabled is false you can omit this
	// parameter (it will default to 1), or you can explicitly set it to a value
	// between 2 and 6.
	//
	// The maximum permitted value for NumCacheClusters is 6 (1 primary plus 5 replicas).
	// +kubebuilder:validation:Maximum=6
	// +optional
	NumCacheClusters *int `json:"numCacheClusters,omitempty"`

	// NumNodeGroups specifies the number of node groups (shards) for this Redis
	// (cluster mode enabled) replication group. For Redis (cluster mode
	// disabled) either omit this parameter or set it to 1.
	//
	// Default: 1
	// +immutable
	// +optional
	NumNodeGroups *int `json:"numNodeGroups,omitempty"`

	// Port number on which each member of the replication group accepts
	// connections.
	// +immutable
	// +optional
	Port *int `json:"port,omitempty"`

	// PreferredCacheClusterAZs specifies a list of EC2 Availability Zones in
	// which the replication group's clusters are created. The order of the
	// Availability Zones in the list is the order in which clusters are
	// allocated. The primary cluster is created in the first AZ in the list.
	//
	// This parameter is not used if there is more than one node group (shard).
	// You should use NodeGroupConfigurationSpec instead.
	//
	// If you are creating your replication group in an Amazon VPC (recommended),
	// you can only locate clusters in Availability Zones associated with the subnets
	// in the selected subnet group.
	//
	// The number of Availability Zones listed must equal the value of NumCacheClusters.
	//
	// Default: system chosen Availability Zones.
	// +immutable
	// +optional
	PreferredCacheClusterAZs []string `json:"preferredCacheClusterAzs,omitempty"`

	// PreferredMaintenanceWindow specifies the weekly time range during which
	// maintenance on the cluster is performed. It is specified as a range in
	// the format ddd:hh24:mi-ddd:hh24:mi (24H Clock UTC). The minimum
	// maintenance window is a 60 minute period.
	//
	// Example: sun:23:00-mon:01:30
	// +optional
	PreferredMaintenanceWindow *string `json:"preferredMaintenanceWindow,omitempty"`

	// PrimaryClusterId is the identifier of the cluster that serves as the
	// primary for this replication group. This cluster must already exist
	// and have a status of available.
	//
	// This parameter is not required if NumCacheClusters, NumNodeGroups or
	// ReplicasPerNodeGroup is specified.
	// +optional
	PrimaryClusterID *string `json:"primaryClusterId,omitempty"`

	// ReplicasPerNodeGroup specifies the number of replica nodes in each node
	// group (shard). Valid values are 0 to 5.
	// +immutable
	// +optional
	ReplicasPerNodeGroup *int `json:"replicasPerNodeGroup,omitempty"`

	// ReplicationGroupDescription is the description for the replication group.
	ReplicationGroupDescription string `json:"replicationGroupDescription"`

	// SecurityGroupIDs specifies one or more Amazon VPC security groups
	// associated with this replication group. Use this parameter only when you
	// are creating a replication group in an Amazon VPC.
	// +optional
	SecurityGroupIDs []string `json:"securityGroupIds,omitempty"`

	// SecurityGroupIDRefs are references to SecurityGroups used to set
	// the SecurityGroupIDs.
	// +immutable
	// +optional
	SecurityGroupIDRefs []xpv1.Reference `json:"securityGroupIdRefs,omitempty"`

	// SecurityGroupIDSelector selects references to SecurityGroups used to set
	// the SecurityGroupIDs.
	// +immutable
	// +optional
	SecurityGroupIDSelector *xpv1.Selector `json:"securityGroupIdSelector,omitempty"`

	// SnapshotARNs specifies a list of Amazon Resource Names (ARN) that
	// uniquely identify the Redis RDB snapshot files stored in Amazon S3. The
	// snapshot files are used to populate the new replication group. The Amazon
	// S3 object name in the ARN cannot contain any commas. The new replication
	// group will have the number of node groups (console: shards) specified by
	// the parameter NumNodeGroups or the number of node groups configured by
	// NodeGroupConfigurationSpec regardless of the number of ARNs specified here.
	// +immutable
	// +optional
	SnapshotARNs []string `json:"snapshotArns,omitempty"`

	// SnapshotName specifies the name of a snapshot from which to restore data
	// into the new replication group. The snapshot status changes to restoring
	// while the new replication group is being created.
	// +immutable
	// +optional
	SnapshotName *string `json:"snapshotName,omitempty"`

	// SnapshotRetentionLimit specifies the number of days for which ElastiCache
	// retains automatic snapshots before deleting them. For example, if you set
	// SnapshotRetentionLimit to 5, a snapshot that was taken today is retained
	// for 5 days before being deleted.
	// Default: 0 (i.e., automatic backups are disabled for this cluster).
	// +optional
	SnapshotRetentionLimit *int `json:"snapshotRetentionLimit,omitempty"`

	// SnapshotWindow specifies the daily time range (in UTC) during which
	// ElastiCache begins taking a daily snapshot of your node group (shard).
	//
	// Example: 05:00-09:00
	//
	// If you do not specify this parameter, ElastiCache automatically chooses an
	// appropriate time range.
	// +optional
	SnapshotWindow *string `json:"snapshotWindow,omitempty"`

	// SnapshottingClusterID is used as the daily snapshot source for the replication
	// group. This parameter cannot be set for Redis (cluster mode enabled) replication
	// groups.
	// +optional
	SnapshottingClusterID *string `json:"snapshottingClusterID,omitempty"`

	// A list of cost allocation tags to be added to this resource. A tag is a key-value
	// pair.
	// +optional
	Tags []Tag `json:"tags,omitempty"`

	// TransitEncryptionEnabled enables in-transit encryption when set to true.
	//
	// You cannot modify the value of TransitEncryptionEnabled after the cluster
	// is created. To enable in-transit encryption on a cluster you must
	// TransitEncryptionEnabled to true when you create a cluster.
	//
	// This parameter is valid only if the Engine parameter is redis, the EngineVersion
	// parameter is 3.2.6 or 4.x, and the cluster is being created in an Amazon
	// VPC.
	//
	// If you enable in-transit encryption, you must also specify a value for CacheSubnetGroup.
	//
	// Required: Only available when creating a replication group in an Amazon VPC
	// using redis version 3.2.6 or 4.x.
	//
	// Default: false
	//
	// For HIPAA compliance, you must specify TransitEncryptionEnabled as true,
	// an AuthToken, and a CacheSubnetGroup.
	// +immutable
	// +optional
	TransitEncryptionEnabled *bool `json:"transitEncryptionEnabled,omitempty"`
}

// A ReplicationGroupSpec defines the desired state of a ReplicationGroup.
type ReplicationGroupSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       ReplicationGroupParameters `json:"forProvider"`
}

// A ReplicationGroupStatus defines the observed state of a ReplicationGroup.
type ReplicationGroupStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          ReplicationGroupObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A ReplicationGroup is a managed resource that represents an AWS ElastiCache
// Replication Group.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="STATE",type="string",JSONPath=".status.atProvider.status"
// +kubebuilder:printcolumn:name="VERSION",type="string",JSONPath=".spec.forProvider.engineVersion"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,aws}
type ReplicationGroup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ReplicationGroupSpec   `json:"spec"`
	Status ReplicationGroupStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ReplicationGroupList contains a list of ReplicationGroup
type ReplicationGroupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ReplicationGroup `json:"items"`
}
