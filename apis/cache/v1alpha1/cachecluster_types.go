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

package v1alpha1

import (
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CacheCluster states.
const (
	StatusCreating            = "creating"
	StatusAvailable           = "available"
	StatusModifying           = "modifying"
	StatusDeleted             = "deleted"
	StatusDeleting            = "deleting"
	StatusCreateFailed        = "create-failed"
	StatusIncompatibleNetwork = "incompatible-network"
	StatusSnapshotting        = "snapshotting"
	StatusRebooting           = "rebooting cluster nodes"
	StatusRestoreFail         = "restore-failed"
)

// A Tag is used to tag the ElastiCache resources in AWS.
type Tag struct {
	// Key for the tag.
	Key string `json:"key"`

	// Value of the tag.
	// +optional
	Value *string `json:"value,omitempty"`
}

// CacheNode represents a node in the cluster
type CacheNode struct {
	// The cache node identifier.
	CacheNodeID string `json:"cacheNodeId,omitempty"`

	// The current state of this cache node, one of the following values:  available, creating,
	// deleted, deleting, incompatible-network, modifying, rebooting cluster nodes, restore-failed, or snapshotting.
	CacheNodeStatus string `json:"cacheNodeStatus,omitempty"`

	// The Availability Zone where this node was created and now resides.
	CustomerAvailabilityZone string `json:"customerAvailabilityZone,omitempty"`

	// The hostname for connecting to this cache node.
	Endpoint *Endpoint `json:"endpoint,omitempty"`

	// The status of the parameter group applied to this cache node.
	ParameterGroupStatus string `json:"parameterGroupStatus,omitempty"`

	// The ID of the primary node to which this read replica node is synchronized.
	SourceCacheNodeID *string `json:"sourceCacheNodeId,omitempty"`
}

// CacheParameterGroupStatus represent status of CacheParameterGroup
type CacheParameterGroupStatus struct {

	// A list of the cache node IDs which need to be rebooted for parameter changes
	// to be applied.
	CacheNodeIDsToReboot []string `json:"cacheNodeIdsToReboot,omitempty"`

	// The name of the cache parameter group.
	CacheParameterGroupName string `json:"cacheParameterGroupName,omitempty"`

	// The status of parameter updates.
	ParameterApplyStatus string `json:"parameterApplyStatus,omitempty"`
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

// NotificationConfiguration represents configuration of a SNS topic
// used to publish Cluster events
type NotificationConfiguration struct {
	// The Amazon Resource Name (ARN) that identifies the topic.
	TopicARN string `json:"topicArn,omitempty"`

	// The current state of the topic.
	TopicStatus *string `json:"topicStatus,omitempty"`
}

// PendingModifiedValues lists values that are applied to cluster in future
type PendingModifiedValues struct {
	// The auth token status
	AuthTokenStatus string `json:"authTokenStatus,omitempty"`

	// A list of cache node IDs that are being removed (or will be removed) from
	// the cluster.
	CacheNodeIDsToRemove []string `json:"cacheNodeIdsToRemove,omitempty"`

	// The cache node type that this cluster or replication group is scaled to.
	CacheNodeType string `json:"cacheNodeType,omitempty"`

	// The new cache engine version that the cluster runs.
	EngineVersion *string `json:"engineVersion,omitempty"`

	// The new number of cache nodes for the cluster.
	NumCacheNodes *int64 `json:"numCacheNodes,omitempty"`
}

// CacheClusterObservation contains the observation of the status of
// the given Cache Cluster.
type CacheClusterObservation struct {
	// A flag that enables encryption at-rest when set to true.
	// Default: false
	AtRestEncryptionEnabled bool `json:"atRestEncryptionEnabled,omitempty"`

	// A flag that enables using an AuthToken (password) when issuing Redis commands.
	// Default: false
	AuthTokenEnabled bool `json:"authTokenEnabled,omitempty"`

	// The current state of this cluster.
	CacheClusterStatus string `json:"cacheClusterStatus,omitempty"`

	// A list of cache nodes that are members of the cluster.
	CacheNodes []CacheNode `json:"cacheNodes,omitempty"`

	// Status of the cache parameter group.
	CacheParameterGroup CacheParameterGroupStatus `json:"cacheParameterGroup,omitempty"`

	// The URL of the web page where you can download the latest ElastiCache client
	// library.
	ClientDownloadLandingPage string `json:"clientDownloadLandingPage,omitempty"`

	// Represents a Memcached cluster endpoint which, if Automatic Discovery is
	// enabled on the cluster, can be used by an application to connect to any node
	// in the cluster. The configuration endpoint will always have .cfg in it.
	ConfigurationEndpoint Endpoint `json:"configurationEndpoint,omitempty"`

	// Describes a notification topic and its status. Notification topics are used
	// for publishing ElastiCache events to subscribers using Amazon Simple Notification
	// Service (SNS).
	NotificationConfiguration NotificationConfiguration `json:"notificationConfiguration,omitempty"`

	// A group of settings that are applied to the cluster in the future, or that
	// are currently being applied.
	PendingModifiedValues PendingModifiedValues `json:"pendingModifiedValues,omitempty"`

	// A flag that enables in-transit encryption when set to true.
	TransitEncryptionEnabled bool `json:"transitEncryptionEnabled,omitempty"`
}

// CacheClusterParameters define the desired state of an AWS ElastiCache
// Cache Cluster. Most fields map directly to an AWS ReplicationGroup:
// https://docs.aws.amazon.com/AmazonElastiCache/latest/APIReference/API_CreateReplicationGroup.html#API_CreateReplicationGroup_RequestParameters
type CacheClusterParameters struct {
	// Region is the region you'd like your CacheSubnetGroup to be created in.
	Region string `json:"region"`

	// If true, this parameter causes the modifications in this request and any
	// pending modifications to be applied, asynchronously and as soon as possible,
	// regardless of the PreferredMaintenanceWindow setting for the cluster.
	// If false, changes to the cluster are applied on the next maintenance reboot,
	// or the next failure reboot, whichever occurs first.
	// +optional
	ApplyImmediately *bool `json:"applyImmediately,omitempty"`

	// Specifies whether the nodes in this Memcached cluster are created in a single
	// Availability Zone or created across multiple Availability Zones in the cluster's
	// region.
	// This parameter is only supported for Memcached clusters.
	// +optional
	AZMode *string `json:"azMode,omitempty"`

	// The password used to access a password protected server.
	// +optional
	AuthToken *string `json:"authToken,omitempty"`

	// Specifies the strategy to use to update the AUTH token. This parameter must
	// be specified with the auth-token parameter. Possible values:
	// +optional
	AuthTokenUpdateStrategy *string `json:"authTokenUpdateStrategy,omitempty"`

	// A list of cache node IDs to be removed.
	// +optional
	CacheNodeIDsToRemove []string `json:"cacheNodeIdsToRemove,omitempty"`

	// The compute and memory capacity of the nodes in the node group (shard).
	CacheNodeType string `json:"cacheNodeType"`

	// The name of the parameter group to associate with this cluster. If this argument
	// is omitted, the default parameter group for the specified engine is used.
	// +optional
	CacheParameterGroupName *string `json:"cacheParameterGroupName,omitempty"`

	// A list of security group names to associate with this cluster.
	// +optional
	CacheSecurityGroupNames []string `json:"cacheSecurityGroupNames,omitempty"`

	// The name of the subnet group to be used for the cluster.
	// +optional
	// +crossplane:generate:reference:type=CacheSubnetGroup
	CacheSubnetGroupName *string `json:"cacheSubnetGroupName,omitempty"`

	// A referencer to retrieve the name of a CacheSubnetGroup
	// +optional
	CacheSubnetGroupNameRef *xpv1.Reference `json:"cacheSubnetGroupNameRef,omitempty"`

	// A selector to select a referencer to retrieve the name of a CacheSubnetGroup
	// +optional
	// +immutable
	CacheSubnetGroupNameSelector *xpv1.Selector `json:"cacheSubnetGroupNameSelector,omitempty"`

	// The name of the cache engine to be used for this cluster.
	// +optional
	// +immutable
	Engine *string `json:"engine,omitempty"`

	// The version number of the cache engine to be used for this cluster.
	// +optional
	EngineVersion *string `json:"engineVersion,omitempty"`

	// The Amazon Resource Name (ARN) of the Amazon Simple Notification Service
	// (SNS) topic to which notifications are sent.
	// +optional
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-aws/apis/sns/v1beta1.Topic
	// +crossplane:generate:reference:extractor=github.com/crossplane-contrib/provider-aws/apis/sns/v1beta1.SNSTopicARN()
	NotificationTopicARN *string `json:"notificationTopicArn,omitempty"`

	// NotificationTopicARNRef references an SNS Topic to retrieve its NotificationTopicARN
	// +optional
	NotificationTopicARNRef *xpv1.Reference `json:"notificationTopicArnRef,omitempty"`

	// NotificationTopicARNSelector selects a reference to an SNS Topic to retrieve its NotificationTopicARN
	// +optional
	NotificationTopicARNSelector *xpv1.Selector `json:"notificationTopicArnSelector,omitempty"`

	// The initial number of cache nodes that the cluster has.
	NumCacheNodes int32 `json:"numCacheNodes"`

	// The port number on which each of the cache nodes accepts connections.
	// +optional
	// +immutable
	Port *int32 `json:"port,omitempty"`

	// The EC2 Availability Zone in which the cluster is created.
	// Default: System chosen Availability Zone.
	// +optional
	PreferredAvailabilityZone *string `json:"preferredAvailabilityZone,omitempty"`

	// A list of the Availability Zones in which cache nodes are created.
	// +optional
	PreferredAvailabilityZones []string `json:"preferredAvailabilityZones,omitempty"`

	// Specifies the weekly time range during which maintenance on the cluster is
	// performed.
	// +optional
	PreferredMaintenanceWindow *string `json:"preferredMaintenanceWindow,omitempty"`

	// The ID of the replication group to which this cluster should belong.
	// +optional
	// +immutable
	ReplicationGroupID *string `json:"replicationGroupId,omitempty"`

	// One or more VPC security groups associated with the cluster.
	// +optional
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-aws/apis/ec2/v1beta1.SecurityGroup
	// +crossplane:generate:reference:refFieldName=SecurityGroupIDRefs
	// +crossplane:generate:reference:selectorFieldName=SecurityGroupIDSelector
	SecurityGroupIDs []string `json:"securityGroupIds,omitempty"`

	// A referencer to retrieve the ID of a Security group
	// +optional
	SecurityGroupIDRefs []xpv1.Reference `json:"securityGroupIDRefs,omitempty"`

	// A selector to select a referencer to retrieve the ID of a Security Group
	// +optional
	SecurityGroupIDSelector *xpv1.Selector `json:"securityGroupIDSelector,omitempty"`

	// A single-element string list containing an Amazon Resource Name (ARN) that
	// uniquely identifies a Redis RDB snapshot file stored in Amazon S3.
	// +optional
	// +immutable
	SnapshotARNs []string `json:"snapshotArns,omitempty"`

	// The name of a Redis snapshot from which to restore data into the new node
	// group (shard).
	// +optional
	// +immutable
	SnapshotName *string `json:"snapshotName,omitempty"`

	// The number of days for which ElastiCache retains automatic snapshots before
	// deleting them.
	// +optional
	SnapshotRetentionLimit *int32 `json:"snapshotRetentionLimit,omitempty"`

	// The daily time range (in UTC) during which ElastiCache begins taking a daily
	// snapshot of your node group (shard).
	// +optional
	SnapshotWindow *string `json:"snapshotWindow,omitempty"`

	// A list of cost allocation tags to be added to this resource.
	// +optional
	// +immutable
	Tags []Tag `json:"tags,omitempty"`
}

// A CacheClusterSpec defines the desired state of a CacheCluster.
type CacheClusterSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       CacheClusterParameters `json:"forProvider"`
}

// A CacheClusterStatus defines the observed state of a CacheCluster.
type CacheClusterStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          CacheClusterObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A CacheCluster is a managed resource that represents an AWS ElastiCache
// Cache Cluster.
// +kubebuilder:printcolumn:name="STATUS",type="string",JSONPath=".status.atProvider.cacheClusterStatus"
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,aws}
type CacheCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CacheClusterSpec   `json:"spec"`
	Status CacheClusterStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// CacheClusterList contains a list of ReplicationGroup
type CacheClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CacheCluster `json:"items"`
}
