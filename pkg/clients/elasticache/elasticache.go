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
	"context"
	"reflect"
	"strconv"

	"github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/elasticache"
	"github.com/aws/aws-sdk-go-v2/service/elasticache/elasticacheiface"

	cachev1alpha1 "github.com/crossplane/provider-aws/apis/cache/v1alpha1"
	"github.com/crossplane/provider-aws/apis/cache/v1beta1"
	clients "github.com/crossplane/provider-aws/pkg/clients"
)

// A Client handles CRUD operations for ElastiCache resources. This interface is
// compatible with the upstream AWS redis client.
type Client elasticacheiface.ClientAPI

// NewClient returns a new ElastiCache client. Credentials must be passed as
// JSON encoded data.
func NewClient(ctx context.Context, credentials []byte, region string, auth clients.AuthMethod) (Client, error) {
	cfg, err := auth(ctx, credentials, clients.DefaultSection, region)
	if cfg == nil {
		return nil, err
	}
	return elasticache.New(*cfg), err
}

// TODO(negz): Determine whether we have to handle converting zero values to
// nil for the below types.

// NewCreateReplicationGroupInput returns ElastiCache replication group creation
// input suitable for use with the AWS API.
func NewCreateReplicationGroupInput(g v1beta1.ReplicationGroupParameters, id string, authToken *string) *elasticache.CreateReplicationGroupInput {
	c := &elasticache.CreateReplicationGroupInput{
		ReplicationGroupId:          &id,
		ReplicationGroupDescription: &g.ReplicationGroupDescription,
		Engine:                      &g.Engine,
		CacheNodeType:               &g.CacheNodeType,

		AtRestEncryptionEnabled:    g.AtRestEncryptionEnabled,
		AuthToken:                  authToken,
		AutomaticFailoverEnabled:   g.AutomaticFailoverEnabled,
		CacheParameterGroupName:    g.CacheParameterGroupName,
		CacheSecurityGroupNames:    g.CacheSecurityGroupNames,
		CacheSubnetGroupName:       g.CacheSubnetGroupName,
		EngineVersion:              g.EngineVersion,
		NotificationTopicArn:       g.NotificationTopicARN,
		NumCacheClusters:           clients.Int64Address(g.NumCacheClusters),
		NumNodeGroups:              clients.Int64Address(g.NumNodeGroups),
		Port:                       clients.Int64Address(g.Port),
		PreferredCacheClusterAZs:   g.PreferredCacheClusterAZs,
		PreferredMaintenanceWindow: g.PreferredMaintenanceWindow,
		PrimaryClusterId:           g.PrimaryClusterID,
		ReplicasPerNodeGroup:       clients.Int64Address(g.ReplicasPerNodeGroup),
		SecurityGroupIds:           g.SecurityGroupIDs,
		SnapshotArns:               g.SnapshotARNs,
		SnapshotName:               g.SnapshotName,
		SnapshotRetentionLimit:     clients.Int64Address(g.SnapshotRetentionLimit),
		SnapshotWindow:             g.SnapshotWindow,
		TransitEncryptionEnabled:   g.TransitEncryptionEnabled,
	}
	if len(g.Tags) != 0 {
		c.Tags = make([]elasticache.Tag, len(g.Tags))
		for i, tag := range g.Tags {
			c.Tags[i] = elasticache.Tag{
				Key:   clients.String(tag.Key),
				Value: clients.String(tag.Value),
			}
		}
	}
	if len(g.NodeGroupConfiguration) != 0 {
		c.NodeGroupConfiguration = make([]elasticache.NodeGroupConfiguration, len(g.NodeGroupConfiguration))
		for i, cfg := range g.NodeGroupConfiguration {
			c.NodeGroupConfiguration[i] = elasticache.NodeGroupConfiguration{
				PrimaryAvailabilityZone:  cfg.PrimaryAvailabilityZone,
				ReplicaAvailabilityZones: cfg.ReplicaAvailabilityZones,
				ReplicaCount:             clients.Int64Address(cfg.ReplicaCount),
				Slots:                    cfg.Slots,
			}
		}
	}
	return c
}

// NewModifyReplicationGroupInput returns ElastiCache replication group
// modification input suitable for use with the AWS API.
func NewModifyReplicationGroupInput(g v1beta1.ReplicationGroupParameters, id string) *elasticache.ModifyReplicationGroupInput {
	return &elasticache.ModifyReplicationGroupInput{
		ReplicationGroupId:          &id,
		ApplyImmediately:            &g.ApplyModificationsImmediately,
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
		SnapshotRetentionLimit:      clients.Int64Address(g.SnapshotRetentionLimit),
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

// LateInitialize assigns the observed configurations and assigns them to the
// corresponding fields in ReplicationGroupParameters in order to let user
// know the defaults and make the changes as wished on that value.
func LateInitialize(s *v1beta1.ReplicationGroupParameters, rg elasticache.ReplicationGroup) {
	// NOTE(muvaf): there are many other parameters that elasticache.ReplicationGroup
	// does not include for some reason.
	s.AtRestEncryptionEnabled = clients.LateInitializeBoolPtr(s.AtRestEncryptionEnabled, rg.AtRestEncryptionEnabled)
	s.AuthEnabled = clients.LateInitializeBoolPtr(s.AuthEnabled, rg.AuthTokenEnabled)
	s.AutomaticFailoverEnabled = clients.LateInitializeBoolPtr(s.AutomaticFailoverEnabled, automaticFailoverEnabled(rg.AutomaticFailover))
	s.SnapshotRetentionLimit = clients.LateInitializeIntPtr(s.SnapshotRetentionLimit, rg.SnapshotRetentionLimit)
	s.SnapshotWindow = clients.LateInitializeStringPtr(s.SnapshotWindow, rg.SnapshotWindow)
	s.SnapshottingClusterID = clients.LateInitializeStringPtr(s.SnapshottingClusterID, rg.SnapshottingClusterId)
	s.TransitEncryptionEnabled = clients.LateInitializeBoolPtr(s.TransitEncryptionEnabled, rg.TransitEncryptionEnabled)
}

// ReplicationGroupNeedsUpdate returns true if the supplied ReplicationGroup and
// the configuration of its member clusters differ from given desired state.
func ReplicationGroupNeedsUpdate(kube v1beta1.ReplicationGroupParameters, rg elasticache.ReplicationGroup, ccList []elasticache.CacheCluster) bool {
	switch {
	case !reflect.DeepEqual(kube.AutomaticFailoverEnabled, automaticFailoverEnabled(rg.AutomaticFailover)):
		return true
	case !reflect.DeepEqual(&kube.CacheNodeType, rg.CacheNodeType):
		return true
	case !reflect.DeepEqual(kube.SnapshotRetentionLimit, clients.IntAddress(rg.SnapshotRetentionLimit)):
		return true
	case !reflect.DeepEqual(kube.SnapshotWindow, rg.SnapshotWindow):
		return true
	}
	for _, cc := range ccList {
		if cacheClusterNeedsUpdate(kube, cc) {
			return true
		}
	}
	return false
}

func automaticFailoverEnabled(af elasticache.AutomaticFailoverStatus) *bool {
	if af == "" {
		return nil
	}
	r := af == elasticache.AutomaticFailoverStatusEnabled || af == elasticache.AutomaticFailoverStatusEnabling
	return &r
}

func cacheClusterNeedsUpdate(kube v1beta1.ReplicationGroupParameters, cc elasticache.CacheCluster) bool { // nolint:gocyclo
	// AWS will set and return a default version if we don't specify one.
	if !reflect.DeepEqual(kube.EngineVersion, cc.EngineVersion) {
		return true
	}
	if pg, name := cc.CacheParameterGroup, kube.CacheParameterGroupName; pg != nil && !reflect.DeepEqual(name, pg.CacheParameterGroupName) {
		return true
	}
	if cc.NotificationConfiguration != nil {
		if !reflect.DeepEqual(kube.NotificationTopicARN, cc.NotificationConfiguration.TopicArn) {
			return true
		}
		if !reflect.DeepEqual(cc.NotificationConfiguration.TopicStatus, kube.NotificationTopicStatus) {
			return true
		}
	} else if clients.StringValue(kube.NotificationTopicARN) != "" {
		return true
	}
	if !reflect.DeepEqual(kube.PreferredMaintenanceWindow, cc.PreferredMaintenanceWindow) {
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
		existingOnes[clients.StringValue(sg.SecurityGroupId)] = true
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
		existingOnes[clients.StringValue(sg.CacheSecurityGroupName)] = true
	}
	for _, desired := range kube {
		if !existingOnes[desired] {
			return true
		}
	}
	return false
}

// GenerateObservation produces a ReplicationGroupObservation object out of
// received elasticache.ReplicationGroup object.
func GenerateObservation(rg elasticache.ReplicationGroup) v1beta1.ReplicationGroupObservation {
	o := v1beta1.ReplicationGroupObservation{
		AutomaticFailover:     string(rg.AutomaticFailover),
		ClusterEnabled:        aws.BoolValue(rg.ClusterEnabled),
		ConfigurationEndpoint: newEndpoint(rg.ConfigurationEndpoint),
		MemberClusters:        rg.MemberClusters,
		Status:                clients.StringValue(rg.Status),
	}
	if len(rg.NodeGroups) != 0 {
		o.NodeGroups = make([]v1beta1.NodeGroup, len(rg.NodeGroups))
		for i, ng := range rg.NodeGroups {
			o.NodeGroups[i] = generateNodeGroup(ng)
		}
	}
	if rg.PendingModifiedValues != nil {
		o.PendingModifiedValues = generateReplicationGroupPendingModifiedValues(*rg.PendingModifiedValues)
	}
	return o
}

func generateNodeGroup(ng elasticache.NodeGroup) v1beta1.NodeGroup {
	r := v1beta1.NodeGroup{
		NodeGroupID: clients.StringValue(ng.NodeGroupId),
		Slots:       clients.StringValue(ng.Slots),
		Status:      clients.StringValue(ng.Status),
	}
	if len(ng.NodeGroupMembers) != 0 {
		r.NodeGroupMembers = make([]v1beta1.NodeGroupMember, len(ng.NodeGroupMembers))
		for i, m := range ng.NodeGroupMembers {
			r.NodeGroupMembers[i] = v1beta1.NodeGroupMember{
				CacheClusterID:            clients.StringValue(m.CacheClusterId),
				CacheNodeID:               clients.StringValue(m.CacheNodeId),
				CurrentRole:               clients.StringValue(m.CurrentRole),
				PreferredAvailabilityZone: clients.StringValue(m.PreferredAvailabilityZone),
			}
			if m.ReadEndpoint != nil {
				r.NodeGroupMembers[i].ReadEndpoint = v1beta1.Endpoint{
					Address: clients.StringValue(m.ReadEndpoint.Address),
					Port:    int(aws.Int64Value(m.ReadEndpoint.Port)),
				}
			}
		}
	}
	return r
}

func generateReplicationGroupPendingModifiedValues(in elasticache.ReplicationGroupPendingModifiedValues) v1beta1.ReplicationGroupPendingModifiedValues {
	r := v1beta1.ReplicationGroupPendingModifiedValues{
		AutomaticFailoverStatus: string(in.AutomaticFailoverStatus),
		PrimaryClusterID:        clients.StringValue(in.PrimaryClusterId),
	}
	if in.Resharding != nil && in.Resharding.SlotMigration != nil {
		r.Resharding = v1beta1.ReshardingStatus{
			SlotMigration: v1beta1.SlotMigration{
				ProgressPercentage: int(aws.Float64Value(in.Resharding.SlotMigration.ProgressPercentage)),
			},
		}
	}
	return r
}

func newEndpoint(e *elasticache.Endpoint) v1beta1.Endpoint {
	if e == nil {
		return v1beta1.Endpoint{}
	}

	return v1beta1.Endpoint{Address: clients.StringValue(e.Address), Port: int(aws.Int64Value(e.Port))}
}

// ConnectionEndpoint returns the connection endpoint for a Replication Group.
// https://docs.aws.amazon.com/AmazonElastiCache/latest/red-ug/Endpoints.html
func ConnectionEndpoint(rg elasticache.ReplicationGroup) managed.ConnectionDetails {
	// "Cluster enabled" Replication Groups have multiple node groups, and an
	// explicit configuration endpoint that should be used for read and write.
	if aws.BoolValue(rg.ClusterEnabled) &&
		rg.ConfigurationEndpoint != nil &&
		rg.ConfigurationEndpoint.Address != nil {
		return managed.ConnectionDetails{
			v1alpha1.ResourceCredentialsSecretEndpointKey: []byte(aws.StringValue(rg.ConfigurationEndpoint.Address)),
			v1alpha1.ResourceCredentialsSecretPortKey:     []byte(strconv.Itoa(int(aws.Int64Value(rg.ConfigurationEndpoint.Port)))),
		}
	}

	// "Cluster disabled" Replication Groups have a single node group, with a
	// primary endpoint that should be used for write. Any node's endpoint can
	// be used for read, but we support only a single endpoint so we return the
	// primary's.
	if len(rg.NodeGroups) > 0 &&
		rg.NodeGroups[0].PrimaryEndpoint != nil &&
		rg.NodeGroups[0].PrimaryEndpoint.Address != nil {
		return managed.ConnectionDetails{
			v1alpha1.ResourceCredentialsSecretEndpointKey: []byte(aws.StringValue(rg.NodeGroups[0].PrimaryEndpoint.Address)),
			v1alpha1.ResourceCredentialsSecretPortKey:     []byte(strconv.Itoa(int(aws.Int64Value(rg.NodeGroups[0].PrimaryEndpoint.Port)))),
		}
	}

	// If the AWS API docs are to be believed we should never get here.
	return nil
}

// IsNotFound returns true if the supplied error indicates a Replication Group
// was not found.
func IsNotFound(err error) bool {
	return isErrorCodeEqual(elasticache.ErrCodeReplicationGroupNotFoundFault, err)
}

// IsSubnetGroupNotFound returns true if the supplied error indicates a Cache Subnet Group
// was not found.
func IsSubnetGroupNotFound(err error) bool {
	return isErrorCodeEqual(elasticache.ErrCodeCacheSubnetGroupNotFoundFault, err)
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

// IsSubnetGroupUpToDate checks if CacheSubnetGroupParameters are in sync with provider values
func IsSubnetGroupUpToDate(p cachev1alpha1.CacheSubnetGroupParameters, sg elasticache.CacheSubnetGroup) bool {
	subnetEqual := len(p.SubnetIds) == len(sg.Subnets) && (func() bool {
		r := false
		for _, id := range p.SubnetIds {
			for _, subnet := range sg.Subnets {
				if id == *subnet.SubnetIdentifier {
					r = true
					break
				} else {
					r = false
				}
			}
		}
		return r
	})()

	return subnetEqual && p.Description == aws.StringValue(sg.CacheSubnetGroupDescription)
}
