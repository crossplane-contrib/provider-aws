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
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/elasticache"
	"github.com/aws/aws-sdk-go-v2/service/elasticache/elasticacheiface"
	"github.com/crossplaneio/stack-aws/apis/cache/v1alpha2"
	aws "github.com/crossplaneio/stack-aws/pkg/clients"
	"github.com/pkg/errors"
)

// A Client handles CRUD operations for ElastiCache resources. This interface is
// compatible with the upstream AWS redis client.
type Client elasticacheiface.ElastiCacheAPI

// NewClient returns a new ElastiCache client. Credentials must be passed as
// JSON encoded data.
func NewClient(credentials []byte, region string) (Client, error) {
	cfg, err := aws.LoadConfig(credentials, aws.DefaultSection, region)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create new AWS configuration")
	}
	return elasticache.New(*cfg), nil
}

// NewReplicationGroupDescription returns a description suitable for use with
// the AWS API.
func NewReplicationGroupDescription(g *v1alpha2.ReplicationGroup) string {
	return fmt.Sprintf("Crossplane managed %s %s/%s", v1alpha2.ReplicationGroupKindAPIVersion, g.GetNamespace(), g.GetName())
}

// TODO(negz): Determine whether we have to handle converting zero values to
// nil for the below types.

// NewCreateReplicationGroupInput returns ElastiCache replication group creation
// input suitable for use with the AWS API.
func NewCreateReplicationGroupInput(g v1alpha2.ReplicationGroupParameters, id, authToken string) *elasticache.CreateReplicationGroupInput {
	c := &elasticache.CreateReplicationGroupInput{
		ReplicationGroupId:          &id,
		ReplicationGroupDescription: &g.ReplicationGroupDescription,

		// The AWS API docs state these fields are not required, but they are.
		// The APi returns an error if they're omitted.
		Engine:        aws.StringAddress(v1alpha2.CacheEngineRedis),
		CacheNodeType: &g.CacheNodeType,

		AtRestEncryptionEnabled:    g.AtRestEncryptionEnabled,
		AuthToken:                  &authToken,
		AutomaticFailoverEnabled:   g.AutomaticFailoverEnabled,
		CacheParameterGroupName:    g.CacheParameterGroupName,
		CacheSecurityGroupNames:    g.CacheSecurityGroupNames,
		CacheSubnetGroupName:       g.CacheSubnetGroupName,
		EngineVersion:              g.EngineVersion,
		NotificationTopicArn:       g.NotificationTopicARN,
		NumCacheClusters:           aws.Int64Address(g.NumCacheClusters),
		NumNodeGroups:              aws.Int64Address(g.NumNodeGroups),
		Port:                       aws.Int64Address(g.Port),
		PreferredCacheClusterAZs:   g.PreferredCacheClusterAZs,
		PreferredMaintenanceWindow: g.PreferredMaintenanceWindow,
		PrimaryClusterId:           g.PrimaryClusterID,
		ReplicasPerNodeGroup:       aws.Int64Address(g.ReplicasPerNodeGroup),
		SecurityGroupIds:           g.SecurityGroupIDs,
		SnapshotArns:               g.SnapshotARNs,
		SnapshotName:               g.SnapshotName,
		SnapshotRetentionLimit:     aws.Int64Address(g.SnapshotRetentionLimit),
		SnapshotWindow:             g.SnapshotWindow,
		TransitEncryptionEnabled:   g.TransitEncryptionEnabled,
	}
	if len(g.Tags) != 0 {
		c.Tags = make([]elasticache.Tag, len(g.Tags))
		for i, tag := range g.Tags {
			c.Tags[i] = elasticache.Tag{
				Key:   &tag.Key,
				Value: &tag.Value,
			}
		}
	}
	if len(g.NodeGroupConfiguration) != 0 {
		c.NodeGroupConfiguration = make([]elasticache.NodeGroupConfiguration, len(c.NodeGroupConfiguration))
		for i, cfg := range g.NodeGroupConfiguration {
			c.NodeGroupConfiguration[i] = elasticache.NodeGroupConfiguration{
				PrimaryAvailabilityZone:  cfg.PrimaryAvailabilityZone,
				ReplicaAvailabilityZones: cfg.ReplicaAvailabilityZones,
				ReplicaCount:             aws.Int64Address(cfg.ReplicaCount),
				Slots:                    cfg.Slots,
			}
		}
	}
	return c
}

// NewModifyReplicationGroupInput returns ElastiCache replication group
// modification input suitable for use with the AWS API.
func NewModifyReplicationGroupInput(g v1alpha2.ReplicationGroupParameters, id string) *elasticache.ModifyReplicationGroupInput {
	return &elasticache.ModifyReplicationGroupInput{
		ReplicationGroupId:          &id,
		ApplyImmediately:            &g.ApplyImmediately,
		AutomaticFailoverEnabled:    g.AutomaticFailoverEnabled,
		CacheNodeType:               &g.CacheNodeType,
		CacheParameterGroupName:     g.CacheParameterGroupName,
		CacheSecurityGroupNames:     g.CacheSecurityGroupNames,
		EngineVersion:               g.EngineVersion,
		NotificationTopicArn:        g.NotificationTopicARN,
		NotificationTopicStatus:     g.NotificationTopicStatus,
		PreferredMaintenanceWindow:  g.PreferredMaintenanceWindow,
		PrimaryClusterId:            g.PrimaryClusterID,
		ReplicationGroupDescription: &g.ReplicationGroupDescription,
		SecurityGroupIds:            g.SecurityGroupIDs,
		SnapshotRetentionLimit:      aws.Int64Address(g.SnapshotRetentionLimit),
		SnapshotWindow:              g.SnapshotWindow,
		SnapshottingClusterId:       g.SnapshottingClusterID,
	}
}

// NewDeleteReplicationGroupInput returns ElastiCache replication group deletion
// input suitable for use with the AWS API.
func NewDeleteReplicationGroupInput(id string) *elasticache.DeleteReplicationGroupInput {
	return &elasticache.DeleteReplicationGroupInput{ReplicationGroupId: &id}
}

// NewDescribeReplicationGroupsInput returns ElastiCache replication group describe
// input suitable for use with the AWS API.
func NewDescribeReplicationGroupsInput(id string) *elasticache.DescribeReplicationGroupsInput {
	return &elasticache.DescribeReplicationGroupsInput{ReplicationGroupId: &id}
}

// NewDescribeCacheClustersInput returns ElastiCache cache cluster describe
// input suitable for use with the AWS API.
func NewDescribeCacheClustersInput(clusterID string) *elasticache.DescribeCacheClustersInput {
	return &elasticache.DescribeCacheClustersInput{CacheClusterId: &clusterID}
}

func LateInitialize(s *v1alpha2.ReplicationGroupParameters, rg elasticache.ReplicationGroup) {
	// NOTE(muvaf): there are many other parameters that elasticache.ReplicationGroup
	// does not include for some reason.
	s.AtRestEncryptionEnabled = aws.LateInitializeBoolPtr(s.AtRestEncryptionEnabled, rg.AtRestEncryptionEnabled)
	s.AuthEnabled = aws.LateInitializeBoolPtr(s.AuthEnabled, rg.AuthTokenEnabled)
	s.AutomaticFailoverEnabled = aws.LateInitializeBool(s.AuthEnabled, automaticFailoverEnabled(rg.AutomaticFailover))
	s.SnapshotRetentionLimit = aws.LateInitializeIntPtr(s.SnapshotRetentionLimit, rg.SnapshotRetentionLimit)
	s.SnapshotWindow = aws.LateInitializeStringPtr(s.SnapshotWindow, rg.SnapshotWindow)
	s.SnapshottingClusterID = aws.LateInitializeStringPtr(s.SnapshottingClusterID, rg.SnapshottingClusterId)
	s.TransitEncryptionEnabled = aws.LateInitializeBoolPtr(s.TransitEncryptionEnabled, rg.TransitEncryptionEnabled)
}

// ReplicationGroupNeedsUpdate returns true if the supplied Kubernetes resource
// differs from the supplied AWS resource.
func ReplicationGroupNeedsUpdate(kube v1alpha2.ReplicationGroupParameters, rg elasticache.ReplicationGroup, ccList []elasticache.CacheCluster) bool {
	switch {
	case aws.BoolValue(kube.AutomaticFailoverEnabled) != automaticFailoverEnabled(rg.AutomaticFailover):
		return true
	case kube.CacheNodeType != aws.StringValue(rg.CacheNodeType):
		return true
	case aws.IntValue(kube.SnapshotRetentionLimit) != aws.Int64Value(rg.SnapshotRetentionLimit):
		return true
	case kube.SnapshotWindow != rg.SnapshotWindow:
		return true
	}
	for _, cc := range ccList {
		if cacheClusterNeedsUpdate(kube, cc) {
			return true
		}
	}
	return false
}

func automaticFailoverEnabled(af elasticache.AutomaticFailoverStatus) bool {
	return af == elasticache.AutomaticFailoverStatusEnabled || af == elasticache.AutomaticFailoverStatusEnabling
}

func cacheClusterNeedsUpdate(kube v1alpha2.ReplicationGroupParameters, cc elasticache.CacheCluster) bool { // nolint:gocyclo
	// AWS will set and return a default version if we don't specify one.
	if kube.EngineVersion != cc.EngineVersion {
		return true
	}
	if pg, name := cc.CacheParameterGroup, kube.CacheParameterGroupName; pg != nil && name != pg.CacheParameterGroupName {
		return true
	}
	if cc.NotificationConfiguration != nil {
		if kube.NotificationTopicARN != cc.NotificationConfiguration.TopicArn {
			return true
		}
		if cc.NotificationConfiguration.TopicStatus != kube.NotificationTopicStatus {
			return true
		}
	} else if aws.StringValue(kube.NotificationTopicARN) != "" {
		return true
	}
	if kube.PreferredMaintenanceWindow != cc.PreferredMaintenanceWindow {
		return true
	}
	return sgIDsNeedUpdate(kube.SecurityGroupIDs, cc.SecurityGroups) || sgNamesNeedUpdate(kube.CacheSecurityGroupNames, cc.CacheSecurityGroups)
}

func sgIDsNeedUpdate(kube []string, cc []elasticache.SecurityGroupMembership) bool {
	if len(kube) != len(cc) {
		return true
	}
	existingOnes := map[string]bool{}
	for _, sg := range cc {
		existingOnes[aws.StringValue(sg.SecurityGroupId)] = true
	}
	for _, desired := range kube {
		if !existingOnes[desired] {
			return true
		}
	}
	return false
}

func sgNamesNeedUpdate(kube []string, cc []elasticache.CacheSecurityGroupMembership) bool {
	if len(kube) != len(cc) {
		return true
	}
	existingOnes := map[string]bool{}
	for _, sg := range cc {
		existingOnes[aws.StringValue(sg.CacheSecurityGroupName)] = true
	}
	for _, desired := range kube {
		if !existingOnes[desired] {
			return true
		}
	}
	return false
}

func GenerateObservation(rg elasticache.ReplicationGroup) v1alpha2.ReplicationGroupObservation {
	o := v1alpha2.ReplicationGroupObservation{
		AutomaticFailover:     string(rg.AutomaticFailover),
		ClusterEnabled:        aws.BoolValue(rg.ClusterEnabled),
		ConfigurationEndpoint: ConnectionEndpoint(rg),
		MemberClusters:        rg.MemberClusters,
		Status:                aws.StringValue(rg.Status),
	}
	if len(rg.NodeGroups) != 0 {
		o.NodeGroups = make([]v1alpha2.NodeGroup, len(rg.NodeGroups))
		for i, ng := range rg.NodeGroups {
			o.NodeGroups[i] = generateNodeGroup(ng)
		}
	}
	if rg.PendingModifiedValues != nil {
		o.PendingModifiedValues = generateReplicationGroupPendingModifiedValues(*rg.PendingModifiedValues)
	}
	return o
}

func generateNodeGroup(ng elasticache.NodeGroup) v1alpha2.NodeGroup {
	r := v1alpha2.NodeGroup{
		NodeGroupId: aws.StringValue(ng.NodeGroupId),
		Slots:       aws.StringValue(ng.Slots),
		Status:      aws.StringValue(ng.Status),
	}
	if len(ng.NodeGroupMembers) != 0 {
		r.NodeGroupMembers = make([]v1alpha2.NodeGroupMember, len(ng.NodeGroupMembers))
		for i, m := range ng.NodeGroupMembers {
			r.NodeGroupMembers[i] = v1alpha2.NodeGroupMember{
				CacheClusterId:            aws.StringValue(m.CacheClusterId),
				CacheNodeId:               aws.StringValue(m.CacheNodeId),
				CurrentRole:               aws.StringValue(m.CurrentRole),
				PreferredAvailabilityZone: aws.StringValue(m.PreferredAvailabilityZone),
			}
			if m.ReadEndpoint != nil {
				r.NodeGroupMembers[i].ReadEndpoint = v1alpha2.Endpoint{
					Address: aws.StringValue(m.ReadEndpoint.Address),
					Port:    aws.Int64Value(m.ReadEndpoint.Port),
				}
			}
		}
	}
	return r
}

func generateReplicationGroupPendingModifiedValues(in elasticache.ReplicationGroupPendingModifiedValues) v1alpha2.ReplicationGroupPendingModifiedValues {
	r := v1alpha2.ReplicationGroupPendingModifiedValues{
		AutomaticFailoverStatus: string(in.AutomaticFailoverStatus),
		PrimaryClusterID:        aws.StringValue(in.PrimaryClusterId),
	}
	if in.Resharding != nil && in.Resharding.SlotMigration != nil {
		r.Resharding = v1alpha2.ReshardingStatus{
			SlotMigration: v1alpha2.SlotMigration{
				ProgressPercentage: aws.Float64Value(in.Resharding.SlotMigration.ProgressPercentage),
			},
		}
	}
	return r
}

func newEndpoint(e *elasticache.Endpoint) v1alpha2.Endpoint {
	if e == nil {
		return v1alpha2.Endpoint{}
	}

	return v1alpha2.Endpoint{Address: aws.StringValue(e.Address), Port: aws.Int64Value(e.Port)}
}

// ConnectionEndpoint returns the connection endpoint for a Replication Group.
// https://docs.aws.amazon.com/AmazonElastiCache/latest/red-ug/Endpoints.html
func ConnectionEndpoint(rg elasticache.ReplicationGroup) v1alpha2.Endpoint {
	// "Cluster enabled" Replication Groups have multiple node groups, and an
	// explicit configuration endpoint that should be used for read and write.
	if aws.BoolValue(rg.ClusterEnabled) {
		return newEndpoint(rg.ConfigurationEndpoint)
	}

	// "Cluster disabled" Replication Groups have a single node group, with a
	// primary endpoint that should be used for write. Any node's endpoint can
	// be used for read, but we support only a single endpoint so we return the
	// primary's.
	if len(rg.NodeGroups) > 0 {
		return newEndpoint(rg.NodeGroups[0].PrimaryEndpoint)
	}

	// If the AWS API docs are to be believed we should never get here.
	return v1alpha2.Endpoint{}
}

// IsNotFound returns true if the supplied error indicates a Replication Group
// was not found.
func IsNotFound(err error) bool {
	return isErrorCodeEqual(elasticache.ErrCodeReplicationGroupNotFoundFault, err)
}

// IsAlreadyExists returns true if the supplied error indicates a Replication Group
// already exists.
func IsAlreadyExists(err error) bool {
	return isErrorCodeEqual(elasticache.ErrCodeReplicationGroupAlreadyExistsFault, err)
}

func isErrorCodeEqual(errorCode string, err error) bool {
	ce, ok := err.(interface {
		Code() string
	})
	if !ok {
		return false
	}

	return ce.Code() == errorCode
}
