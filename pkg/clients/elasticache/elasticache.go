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
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/elasticache"
	elasticachetypes "github.com/aws/aws-sdk-go-v2/service/elasticache/types"
	"github.com/aws/smithy-go/document"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/pkg/errors"

	cachev1alpha1 "github.com/crossplane-contrib/provider-aws/apis/cache/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/apis/cache/v1beta1"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/elasticache/convert"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
	tagutils "github.com/crossplane-contrib/provider-aws/pkg/utils/tags"
)

const (
	errCheckUpToDate = "unable to determine if external resource is up to date"
	errVersionInput  = "unable to parse version number"
)

// A Client handles CRUD operations for ElastiCache resources.
type Client interface {
	DescribeReplicationGroups(context.Context, *elasticache.DescribeReplicationGroupsInput, ...func(*elasticache.Options)) (*elasticache.DescribeReplicationGroupsOutput, error)
	CreateReplicationGroup(context.Context, *elasticache.CreateReplicationGroupInput, ...func(*elasticache.Options)) (*elasticache.CreateReplicationGroupOutput, error)
	ModifyReplicationGroup(context.Context, *elasticache.ModifyReplicationGroupInput, ...func(*elasticache.Options)) (*elasticache.ModifyReplicationGroupOutput, error)
	DeleteReplicationGroup(context.Context, *elasticache.DeleteReplicationGroupInput, ...func(*elasticache.Options)) (*elasticache.DeleteReplicationGroupOutput, error)

	DescribeCacheSubnetGroups(context.Context, *elasticache.DescribeCacheSubnetGroupsInput, ...func(*elasticache.Options)) (*elasticache.DescribeCacheSubnetGroupsOutput, error)
	CreateCacheSubnetGroup(context.Context, *elasticache.CreateCacheSubnetGroupInput, ...func(*elasticache.Options)) (*elasticache.CreateCacheSubnetGroupOutput, error)
	ModifyCacheSubnetGroup(context.Context, *elasticache.ModifyCacheSubnetGroupInput, ...func(*elasticache.Options)) (*elasticache.ModifyCacheSubnetGroupOutput, error)
	DeleteCacheSubnetGroup(context.Context, *elasticache.DeleteCacheSubnetGroupInput, ...func(*elasticache.Options)) (*elasticache.DeleteCacheSubnetGroupOutput, error)

	DescribeCacheClusters(context.Context, *elasticache.DescribeCacheClustersInput, ...func(*elasticache.Options)) (*elasticache.DescribeCacheClustersOutput, error)
	CreateCacheCluster(context.Context, *elasticache.CreateCacheClusterInput, ...func(*elasticache.Options)) (*elasticache.CreateCacheClusterOutput, error)
	DeleteCacheCluster(context.Context, *elasticache.DeleteCacheClusterInput, ...func(*elasticache.Options)) (*elasticache.DeleteCacheClusterOutput, error)
	ModifyCacheCluster(context.Context, *elasticache.ModifyCacheClusterInput, ...func(*elasticache.Options)) (*elasticache.ModifyCacheClusterOutput, error)

	DecreaseReplicaCount(ctx context.Context, params *elasticache.DecreaseReplicaCountInput, optFns ...func(*elasticache.Options)) (*elasticache.DecreaseReplicaCountOutput, error)
	IncreaseReplicaCount(ctx context.Context, params *elasticache.IncreaseReplicaCountInput, optFns ...func(*elasticache.Options)) (*elasticache.IncreaseReplicaCountOutput, error)

	ModifyReplicationGroupShardConfiguration(context.Context, *elasticache.ModifyReplicationGroupShardConfigurationInput, ...func(*elasticache.Options)) (*elasticache.ModifyReplicationGroupShardConfigurationOutput, error)

	ListTagsForResource(context.Context, *elasticache.ListTagsForResourceInput, ...func(*elasticache.Options)) (*elasticache.ListTagsForResourceOutput, error)
	AddTagsToResource(context.Context, *elasticache.AddTagsToResourceInput, ...func(*elasticache.Options)) (*elasticache.AddTagsToResourceOutput, error)
	RemoveTagsFromResource(context.Context, *elasticache.RemoveTagsFromResourceInput, ...func(*elasticache.Options)) (*elasticache.RemoveTagsFromResourceOutput, error)
}

// NewClient returns a new ElastiCache client. Credentials must be passed as
// JSON encoded data.
func NewClient(cfg aws.Config) Client {
	return elasticache.NewFromConfig(cfg)
}

// TODO(negz): Determine whether we have to handle converting zero values to
// nil for the below elasticachetypes.

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
		MultiAZEnabled:             g.MultiAZEnabled,
		NotificationTopicArn:       g.NotificationTopicARN,
		NumCacheClusters:           pointer.ToIntAsInt32Ptr(g.NumCacheClusters),
		NumNodeGroups:              pointer.ToIntAsInt32Ptr(g.NumNodeGroups),
		Port:                       pointer.ToIntAsInt32Ptr(g.Port),
		PreferredCacheClusterAZs:   g.PreferredCacheClusterAZs,
		PreferredMaintenanceWindow: g.PreferredMaintenanceWindow,
		PrimaryClusterId:           g.PrimaryClusterID,
		ReplicasPerNodeGroup:       pointer.ToIntAsInt32Ptr(g.ReplicasPerNodeGroup),
		SecurityGroupIds:           g.SecurityGroupIDs,
		SnapshotArns:               g.SnapshotARNs,
		SnapshotName:               g.SnapshotName,
		SnapshotRetentionLimit:     pointer.ToIntAsInt32Ptr(g.SnapshotRetentionLimit),
		SnapshotWindow:             g.SnapshotWindow,
		TransitEncryptionEnabled:   g.TransitEncryptionEnabled,
	}
	if len(g.Tags) != 0 {
		c.Tags = make([]elasticachetypes.Tag, len(g.Tags))
		for i, tag := range g.Tags {
			c.Tags[i] = elasticachetypes.Tag{
				Key:   pointer.ToOrNilIfZeroValue(tag.Key),
				Value: pointer.ToOrNilIfZeroValue(tag.Value),
			}
		}
	}
	if len(g.NodeGroupConfiguration) != 0 {
		c.NodeGroupConfiguration = make([]elasticachetypes.NodeGroupConfiguration, len(g.NodeGroupConfiguration))
		for i, cfg := range g.NodeGroupConfiguration {
			c.NodeGroupConfiguration[i] = elasticachetypes.NodeGroupConfiguration{
				PrimaryAvailabilityZone:  cfg.PrimaryAvailabilityZone,
				ReplicaAvailabilityZones: cfg.ReplicaAvailabilityZones,
				ReplicaCount:             pointer.ToIntAsInt32Ptr(cfg.ReplicaCount),
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
		ReplicationGroupId:          aws.String(id),
		ApplyImmediately:            g.ApplyModificationsImmediately,
		AutomaticFailoverEnabled:    g.AutomaticFailoverEnabled,
		CacheNodeType:               aws.String(g.CacheNodeType),
		CacheParameterGroupName:     g.CacheParameterGroupName,
		CacheSecurityGroupNames:     g.CacheSecurityGroupNames,
		EngineVersion:               g.EngineVersion,
		MultiAZEnabled:              g.MultiAZEnabled,
		NotificationTopicArn:        g.NotificationTopicARN,
		NotificationTopicStatus:     g.NotificationTopicStatus,
		PreferredMaintenanceWindow:  g.PreferredMaintenanceWindow,
		PrimaryClusterId:            g.PrimaryClusterID,
		ReplicationGroupDescription: aws.String(g.ReplicationGroupDescription),
		SecurityGroupIds:            g.SecurityGroupIDs,
		SnapshotRetentionLimit:      pointer.ToIntAsInt32Ptr(g.SnapshotRetentionLimit),
		SnapshotWindow:              g.SnapshotWindow,
		SnapshottingClusterId:       g.SnapshottingClusterID,
	}
}

// NewModifyReplicationGroupShardConfigurationInput returns ElastiCache replication group
// shard configuration modification input suitable for use with the AWS API.
func NewModifyReplicationGroupShardConfigurationInput(g v1beta1.ReplicationGroupParameters, id string, rg elasticachetypes.ReplicationGroup) *elasticache.ModifyReplicationGroupShardConfigurationInput {
	input := &elasticache.ModifyReplicationGroupShardConfigurationInput{
		ApplyImmediately:   g.ApplyModificationsImmediately,
		NodeGroupCount:     int32(*g.NumNodeGroups),
		ReplicationGroupId: aws.String(id),
	}

	// For scale down we must name the nodes. This code picks the oldest rg
	// now, but there might be a better algorithm, such as the one with least
	// data
	remove := len(rg.NodeGroups) - int(input.NodeGroupCount)
	for i := 0; i < remove; i++ {
		input.NodeGroupsToRemove = append(input.NodeGroupsToRemove, aws.ToString(rg.NodeGroups[i].NodeGroupId))
	}

	return input
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

// NewListTagsForResourceInput returns ElastiCache list tags input usable with the
// AWS API
func NewListTagsForResourceInput(arn *string) *elasticache.ListTagsForResourceInput {
	return &elasticache.ListTagsForResourceInput{ResourceName: arn}
}

// NewDecreaseReplicaCountInput returns Elasticache replication group decrease
// the number of replicaGroup cache clusters
func NewDecreaseReplicaCountInput(replicationGroupID string, newReplicaCount *int32) *elasticache.DecreaseReplicaCountInput {
	return &elasticache.DecreaseReplicaCountInput{
		ApplyImmediately:   true, // false is not supported by the API
		ReplicationGroupId: &replicationGroupID,
		NewReplicaCount:    newReplicaCount,
	}

}

// NewIncreaseReplicaCountInput returns Elasticache replication group increase
// the number of replicaGroup cache clusters
func NewIncreaseReplicaCountInput(replicationGroupID string, newReplicaCount *int32) *elasticache.IncreaseReplicaCountInput {
	return &elasticache.IncreaseReplicaCountInput{
		ApplyImmediately:   true, // false is not supported by the API
		ReplicationGroupId: &replicationGroupID,
		NewReplicaCount:    newReplicaCount,
	}

}

// LateInitialize assigns the observed configurations and assigns them to the
// corresponding fields in ReplicationGroupParameters in order to let user
// know the defaults and make the changes as wished on that value.
func LateInitialize(s *v1beta1.ReplicationGroupParameters, rg elasticachetypes.ReplicationGroup, cc elasticachetypes.CacheCluster) {
	if s == nil {
		return
	}
	s.AtRestEncryptionEnabled = pointer.LateInitialize(s.AtRestEncryptionEnabled, rg.AtRestEncryptionEnabled)
	s.AuthEnabled = pointer.LateInitialize(s.AuthEnabled, rg.AuthTokenEnabled)
	s.AutomaticFailoverEnabled = pointer.LateInitialize(s.AutomaticFailoverEnabled, automaticFailoverEnabled(rg.AutomaticFailover))
	s.SnapshotRetentionLimit = pointer.LateInitializeIntFromInt32Ptr(s.SnapshotRetentionLimit, rg.SnapshotRetentionLimit)
	s.SnapshotWindow = pointer.LateInitialize(s.SnapshotWindow, rg.SnapshotWindow)
	s.SnapshottingClusterID = pointer.LateInitialize(s.SnapshottingClusterID, rg.SnapshottingClusterId)
	s.TransitEncryptionEnabled = pointer.LateInitialize(s.TransitEncryptionEnabled, rg.TransitEncryptionEnabled)

	// NOTE(muvaf): ReplicationGroup managed N identical CacheCluster objects.
	// While configuration of those CacheClusters flow through ReplicationGroup API,
	// their statuses are fetched independently. Since we check for drifts against
	// the current state, late-init and up-to-date checks have to be made against
	// CacheClusters as well.
	s.EngineVersion = pointer.LateInitialize(s.EngineVersion, cc.EngineVersion)
	if cc.CacheParameterGroup != nil {
		s.CacheParameterGroupName = pointer.LateInitialize(s.CacheParameterGroupName, cc.CacheParameterGroup.CacheParameterGroupName)
	}
	if cc.NotificationConfiguration != nil {
		s.NotificationTopicARN = pointer.LateInitialize(s.NotificationTopicARN, cc.NotificationConfiguration.TopicArn)
		s.NotificationTopicStatus = pointer.LateInitialize(s.NotificationTopicStatus, cc.NotificationConfiguration.TopicStatus)
	}
	s.PreferredMaintenanceWindow = pointer.LateInitialize(s.PreferredMaintenanceWindow, cc.PreferredMaintenanceWindow)
	if len(s.SecurityGroupIDs) == 0 && len(cc.SecurityGroups) != 0 {
		s.SecurityGroupIDs = make([]string, len(cc.SecurityGroups))
		for i, val := range cc.SecurityGroups {
			s.SecurityGroupIDs[i] = aws.ToString(val.SecurityGroupId)
		}
	}
	if len(s.CacheSecurityGroupNames) == 0 && len(cc.CacheSecurityGroups) != 0 {
		s.CacheSecurityGroupNames = make([]string, len(cc.CacheSecurityGroups))
		for i, val := range cc.CacheSecurityGroups {
			s.CacheSecurityGroupNames[i] = aws.ToString(val.CacheSecurityGroupName)
		}
	}
}

// ReplicationGroupShardConfigurationNeedsUpdate returns true if the supplied ReplicationGroup and
// the configuration shards.
func ReplicationGroupShardConfigurationNeedsUpdate(kube v1beta1.ReplicationGroupParameters, rg elasticachetypes.ReplicationGroup) bool {
	return kube.NumNodeGroups != nil && *kube.NumNodeGroups != len(rg.NodeGroups)
}

// ReplicationGroupNeedsUpdate returns true if the supplied ReplicationGroup and
// the configuration of its member clusters differ from given desired state.
func ReplicationGroupNeedsUpdate(kube v1beta1.ReplicationGroupParameters, rg elasticachetypes.ReplicationGroup, ccList []elasticachetypes.CacheCluster) string {
	switch {
	case !reflect.DeepEqual(kube.AutomaticFailoverEnabled, automaticFailoverEnabled(rg.AutomaticFailover)):
		return "AutomaticFailover"
	case !reflect.DeepEqual(&kube.CacheNodeType, rg.CacheNodeType):
		return "CacheNotType"
	case !reflect.DeepEqual(kube.SnapshotRetentionLimit, pointer.ToInt32FromIntPtr(rg.SnapshotRetentionLimit)):
		return "SnapshotRetentionLimit"
	case !reflect.DeepEqual(kube.SnapshotWindow, rg.SnapshotWindow):
		return "SnapshotWindow"
	case aws.ToBool(kube.MultiAZEnabled) != aws.ToBool(multiAZEnabled(rg.MultiAZ)):
		return "MultiAZ"
	case ReplicationGroupNumCacheClustersNeedsUpdate(kube, ccList):
		return "NumCacheClusters"
	}

	for _, cc := range ccList {
		if reason := cacheClusterNeedsUpdate(kube, cc); reason != "" {
			return reason
		}
	}
	return ""
}

func automaticFailoverEnabled(af elasticachetypes.AutomaticFailoverStatus) *bool {
	if af == "" {
		return nil
	}
	r := af == elasticachetypes.AutomaticFailoverStatusEnabled || af == elasticachetypes.AutomaticFailoverStatusEnabling
	return &r
}

func multiAZEnabled(maz elasticachetypes.MultiAZStatus) *bool {
	switch maz {
	case elasticachetypes.MultiAZStatusEnabled:
		return aws.Bool(true)
	case elasticachetypes.MultiAZStatusDisabled:
		return aws.Bool(false)
	default:
		return nil
	}
}

// PartialSemanticVersion is semantic version that does not fulfill
// the specification. This allows for partial matching.
type PartialSemanticVersion struct {
	Major *int64
	Minor *int64
	Patch *int64
}

// ParseVersion parses the semantic version of an Elasticache Cluster
// See https://docs.aws.amazon.com/memorydb/latest/devguide/engine-versions.html
func ParseVersion(ver *string) (*PartialSemanticVersion, error) {
	if ver == nil || aws.ToString(ver) == "" {
		return nil, errors.New("empty string")
	}

	parts := strings.Split(strings.TrimSpace(aws.ToString(ver)), ".")

	major, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return nil, errors.New("major version must be a number")
	}

	p := &PartialSemanticVersion{Major: aws.Int64(major)}

	if len(parts) > 1 {
		minor, err := strconv.ParseInt(parts[1], 10, 64)
		// if not a digit (i.e. .x, ignore)
		if err != nil {
			return p, nil //nolint:nilerr
		}
		p.Minor = aws.Int64(minor)
	}

	if len(parts) > 2 {
		patch, err := strconv.ParseInt(parts[2], 10, 64)
		if err != nil {
			return p, nil //nolint:nilerr
		}
		p.Patch = aws.Int64(patch)
	}

	return p, nil
}

// For versions before 6.x, the version string can be exact (i.e. 5.0.6)
// For versions 6, 6.2, 6.x, etc., we only need a major version
func versionMatches(kubeVersion *string, awsVersion *string) bool { //nolint: gocyclo

	if pointer.StringValue(kubeVersion) == pointer.StringValue(awsVersion) {
		return true
	}

	if kubeVersion == nil || awsVersion == nil {
		return false
	}

	kv, err := ParseVersion(kubeVersion)
	if err != nil {
		return false
	}

	av, err := ParseVersion(awsVersion)
	if err != nil {
		return false
	}

	if aws.ToInt64(kv.Major) != aws.ToInt64(av.Major) {
		return false
	}

	if kv.Minor != nil {
		if aws.ToInt64(kv.Minor) != aws.ToInt64(av.Minor) {
			return false
		}
	}

	// Setting the patch level is valid for Redis versions < 6
	if kv.Patch != nil && aws.ToInt64(kv.Major) < 6 {
		if aws.ToInt64(kv.Patch) != aws.ToInt64(av.Patch) {
			return false
		}
	}

	return true
}

func cacheClusterNeedsUpdate(kube v1beta1.ReplicationGroupParameters, cc elasticachetypes.CacheCluster) string { //nolint:gocyclo
	// AWS will set and return a default version if we don't specify one.
	if !versionMatches(kube.EngineVersion, cc.EngineVersion) {
		return "EngineVersion"
	}
	if pg, name := cc.CacheParameterGroup, kube.CacheParameterGroupName; pg != nil && !reflect.DeepEqual(name, pg.CacheParameterGroupName) {
		return "CacheParameterGroup"
	}
	if cc.NotificationConfiguration != nil {
		if !reflect.DeepEqual(kube.NotificationTopicARN, cc.NotificationConfiguration.TopicArn) {
			return "NoticationTopicARN"
		}
		if !reflect.DeepEqual(cc.NotificationConfiguration.TopicStatus, kube.NotificationTopicStatus) {
			return "TopicStatus"
		}
	} else if pointer.StringValue(kube.NotificationTopicARN) != "" {
		return "NotificationTopicARN"
	}
	// AWS will normalize preferred maintenance windows to lowercase
	if !strings.EqualFold(pointer.StringValue(kube.PreferredMaintenanceWindow),
		pointer.StringValue(cc.PreferredMaintenanceWindow)) {
		return "PreferredMaintainenceWindow"
	}
	if sgIDsNeedUpdate(kube.SecurityGroupIDs, cc.SecurityGroups) || sgNamesNeedUpdate(kube.CacheSecurityGroupNames, cc.CacheSecurityGroups) {
		return "SecurityGroups"
	}
	return ""
}

func sgIDsNeedUpdate(kube []string, cc []elasticachetypes.SecurityGroupMembership) bool {
	if len(kube) != len(cc) {
		return true
	}
	existingOnes := map[string]bool{}
	for _, sg := range cc {
		existingOnes[pointer.StringValue(sg.SecurityGroupId)] = true
	}
	for _, desired := range kube {
		if !existingOnes[desired] {
			return true
		}
	}
	return false
}

func sgNamesNeedUpdate(kube []string, cc []elasticachetypes.CacheSecurityGroupMembership) bool {
	if len(kube) != len(cc) {
		return true
	}
	existingOnes := map[string]bool{}
	for _, sg := range cc {
		existingOnes[pointer.StringValue(sg.CacheSecurityGroupName)] = true
	}
	for _, desired := range kube {
		if !existingOnes[desired] {
			return true
		}
	}
	return false
}

// ReplicationGroupTagsNeedsUpdate indicates whether tags need updating
func ReplicationGroupTagsNeedsUpdate(kube []v1beta1.Tag, tags []elasticachetypes.Tag) bool {
	if len(kube) != len(tags) {
		return true
	}

	add, remove := DiffTags(kube, tags)

	return len(add) != 0 || len(remove) != 0
}

// DiffTags returns tags that should be added or removed.
func DiffTags(rgtags []v1beta1.Tag, tags []elasticachetypes.Tag) (add map[string]string, remove []string) {

	local := make(map[string]string, len(rgtags))
	for _, t := range rgtags {
		local[t.Key] = t.Value
	}

	remote := make(map[string]string, len(tags))
	for _, t := range tags {
		remote[aws.ToString(t.Key)] = aws.ToString(t.Value)
	}

	return tagutils.DiffTags(local, remote)
}

// ReplicationGroupNumCacheClustersNeedsUpdate determines if the number of Cache Clusters
// in a replication group needs to be updated
func ReplicationGroupNumCacheClustersNeedsUpdate(kube v1beta1.ReplicationGroupParameters, ccList []elasticachetypes.CacheCluster) bool {
	return kube.NumCacheClusters != nil && aws.ToInt(kube.NumCacheClusters) != len(ccList)
}

// GenerateObservation produces a ReplicationGroupObservation object out of
// received elasticache.ReplicationGroup object.
func GenerateObservation(rg elasticachetypes.ReplicationGroup) v1beta1.ReplicationGroupObservation {
	o := v1beta1.ReplicationGroupObservation{
		AutomaticFailover:     string(rg.AutomaticFailover),
		ClusterEnabled:        aws.ToBool(rg.ClusterEnabled),
		ConfigurationEndpoint: newEndpoint(rg),
		MemberClusters:        rg.MemberClusters,
		Status:                pointer.StringValue(rg.Status),
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

func generateNodeGroup(ng elasticachetypes.NodeGroup) v1beta1.NodeGroup {
	r := v1beta1.NodeGroup{
		NodeGroupID: pointer.StringValue(ng.NodeGroupId),
		Slots:       pointer.StringValue(ng.Slots),
		Status:      pointer.StringValue(ng.Status),
	}
	if len(ng.NodeGroupMembers) != 0 {
		r.NodeGroupMembers = make([]v1beta1.NodeGroupMember, len(ng.NodeGroupMembers))
		for i, m := range ng.NodeGroupMembers {
			r.NodeGroupMembers[i] = v1beta1.NodeGroupMember{
				CacheClusterID:            pointer.StringValue(m.CacheClusterId),
				CacheNodeID:               pointer.StringValue(m.CacheNodeId),
				CurrentRole:               pointer.StringValue(m.CurrentRole),
				PreferredAvailabilityZone: pointer.StringValue(m.PreferredAvailabilityZone),
			}
			if m.ReadEndpoint != nil {
				r.NodeGroupMembers[i].ReadEndpoint = v1beta1.Endpoint{
					Address: pointer.StringValue(m.ReadEndpoint.Address),
					Port:    int(m.ReadEndpoint.Port),
				}
			}

		}
	}
	return r
}

func generateReplicationGroupPendingModifiedValues(in elasticachetypes.ReplicationGroupPendingModifiedValues) v1beta1.ReplicationGroupPendingModifiedValues {
	r := v1beta1.ReplicationGroupPendingModifiedValues{
		AutomaticFailoverStatus: string(in.AutomaticFailoverStatus),
		PrimaryClusterID:        pointer.StringValue(in.PrimaryClusterId),
	}
	if in.Resharding != nil && in.Resharding.SlotMigration != nil {
		r.Resharding = v1beta1.ReshardingStatus{
			SlotMigration: v1beta1.SlotMigration{
				ProgressPercentage: int(in.Resharding.SlotMigration.ProgressPercentage),
			},
		}
	}
	return r
}

// newEndpoint returns the endpoint end users should use to connect to this cluster.
// https://docs.aws.amazon.com/AmazonElastiCache/latest/red-ug/Endpoints.html
func newEndpoint(rg elasticachetypes.ReplicationGroup) v1beta1.Endpoint {
	var e *elasticachetypes.Endpoint
	switch {
	case !aws.ToBool(rg.ClusterEnabled) && len(rg.NodeGroups) > 0 && rg.NodeGroups[0].PrimaryEndpoint != nil:
		e = rg.NodeGroups[0].PrimaryEndpoint
	case aws.ToBool(rg.ClusterEnabled) && rg.ConfigurationEndpoint != nil:
		e = rg.ConfigurationEndpoint
	default:
		return v1beta1.Endpoint{}
	}
	return v1beta1.Endpoint{Address: pointer.StringValue(e.Address), Port: int(e.Port)}
}

// ConnectionEndpoint returns the connection endpoint for a Replication Group.
// https://docs.aws.amazon.com/AmazonElastiCache/latest/red-ug/Endpoints.html
func ConnectionEndpoint(rg elasticachetypes.ReplicationGroup) managed.ConnectionDetails {
	// "Cluster enabled" Replication Groups have multiple node groups, and an
	// explicit configuration endpoint that should be used for read and write.
	if aws.ToBool(rg.ClusterEnabled) &&
		rg.ConfigurationEndpoint != nil &&
		rg.ConfigurationEndpoint.Address != nil {
		return managed.ConnectionDetails{
			xpv1.ResourceCredentialsSecretEndpointKey: []byte(aws.ToString(rg.ConfigurationEndpoint.Address)),
			xpv1.ResourceCredentialsSecretPortKey:     []byte(strconv.Itoa(int(rg.ConfigurationEndpoint.Port))),
		}
	}

	// "Cluster disabled" Replication Groups have a single node group, with a
	// primary endpoint that should be used for write. Any node's endpoint can
	// be used for read, but we support only a single endpoint so we return the
	// primary's.
	if len(rg.NodeGroups) > 0 {
		hasData := false
		cd := managed.ConnectionDetails{}

		if rg.NodeGroups[0].PrimaryEndpoint != nil &&
			rg.NodeGroups[0].PrimaryEndpoint.Address != nil {
			hasData = true
			cd[xpv1.ResourceCredentialsSecretEndpointKey] = []byte(aws.ToString(rg.NodeGroups[0].PrimaryEndpoint.Address))
			cd[xpv1.ResourceCredentialsSecretPortKey] = []byte(strconv.Itoa(int(rg.NodeGroups[0].PrimaryEndpoint.Port)))
		}
		if rg.NodeGroups[0].ReaderEndpoint != nil &&
			rg.NodeGroups[0].ReaderEndpoint.Address != nil {
			hasData = true
			cd["readerEndpoint"] = []byte(aws.ToString(rg.NodeGroups[0].ReaderEndpoint.Address))
			cd["readerPort"] = []byte(strconv.Itoa(int(rg.NodeGroups[0].ReaderEndpoint.Port)))
		}
		if hasData {
			return cd
		}
	}

	// If the AWS API docs are to be believed we should never get here.
	return nil
}

// IsNotFound returns true if the supplied error indicates a Replication Group
// was not found.
func IsNotFound(err error) bool {
	var gnf *elasticachetypes.ReplicationGroupNotFoundFault
	return errors.As(err, &gnf)
}

// IsSubnetGroupNotFound returns true if the supplied error indicates a Cache Subnet Group
// was not found.
func IsSubnetGroupNotFound(err error) bool {
	var gnf *elasticachetypes.CacheSubnetGroupNotFoundFault
	return errors.As(err, &gnf)
}

// IsAlreadyExists returns true if the supplied error indicates a Replication Group
// already exists.
func IsAlreadyExists(err error) bool {
	var gae *elasticachetypes.ReplicationGroupAlreadyExistsFault
	return errors.As(err, &gae)
}

// IsSubnetGroupUpToDate checks if CacheSubnetGroupParameters are in sync with provider values
func IsSubnetGroupUpToDate(p cachev1alpha1.CacheSubnetGroupParameters, sg elasticachetypes.CacheSubnetGroup) bool {
	if p.Description != aws.ToString(sg.CacheSubnetGroupDescription) {
		return false
	}

	if len(p.SubnetIDs) != len(sg.Subnets) {
		return false
	}

	exists := make(map[string]bool)
	for _, s := range sg.Subnets {
		exists[*s.SubnetIdentifier] = true
	}
	for _, id := range p.SubnetIDs {
		if !exists[id] {
			return false
		}
	}

	return true
}

// GenerateCreateCacheClusterInput returns Cache Cluster creation input
func GenerateCreateCacheClusterInput(p cachev1alpha1.CacheClusterParameters, id string) (*elasticache.CreateCacheClusterInput, error) {
	c := &elasticache.CreateCacheClusterInput{
		AZMode:                     elasticachetypes.AZMode(aws.ToString(p.AZMode)),
		AuthToken:                  p.AZMode,
		CacheClusterId:             aws.String(id),
		CacheNodeType:              aws.String(p.CacheNodeType),
		CacheParameterGroupName:    p.CacheParameterGroupName,
		CacheSubnetGroupName:       p.CacheSubnetGroupName,
		CacheSecurityGroupNames:    p.CacheSecurityGroupNames,
		Engine:                     p.Engine,
		NotificationTopicArn:       p.NotificationTopicARN,
		NumCacheNodes:              aws.Int32(p.NumCacheNodes),
		Port:                       p.Port,
		PreferredAvailabilityZone:  p.PreferredAvailabilityZone,
		PreferredAvailabilityZones: p.PreferredAvailabilityZones,
		PreferredMaintenanceWindow: p.PreferredMaintenanceWindow,
		ReplicationGroupId:         p.ReplicationGroupID,
		SecurityGroupIds:           p.SecurityGroupIDs,
		SnapshotArns:               p.SnapshotARNs,
		SnapshotName:               p.SnapshotName,
		SnapshotRetentionLimit:     p.SnapshotRetentionLimit,
		SnapshotWindow:             p.SnapshotWindow,
	}

	if p.EngineVersion != nil {
		version, err := getVersion(p.EngineVersion)
		if err != nil {
			return nil, err
		}
		c.EngineVersion = version
	} else {
		c.EngineVersion = p.EngineVersion
	}

	if len(p.Tags) != 0 {
		c.Tags = make([]elasticachetypes.Tag, len(p.Tags))
		for i, tag := range p.Tags {
			c.Tags[i] = elasticachetypes.Tag{
				Key:   pointer.ToOrNilIfZeroValue(tag.Key),
				Value: tag.Value,
			}
		}
	}

	return c, nil
}

// GenerateModifyCacheClusterInput returns ElastiCache Cache Cluster
// modification input suitable for use with the AWS API.
func GenerateModifyCacheClusterInput(p cachev1alpha1.CacheClusterParameters, id string) (*elasticache.ModifyCacheClusterInput, error) {
	c := &elasticache.ModifyCacheClusterInput{
		CacheClusterId:             aws.String(id),
		AZMode:                     elasticachetypes.AZMode(aws.ToString(p.AZMode)),
		ApplyImmediately:           aws.ToBool(p.ApplyImmediately),
		AuthToken:                  p.AuthToken,
		AuthTokenUpdateStrategy:    elasticachetypes.AuthTokenUpdateStrategyType(aws.ToString(p.AuthTokenUpdateStrategy)),
		CacheNodeIdsToRemove:       p.CacheNodeIDsToRemove,
		CacheNodeType:              aws.String(p.CacheNodeType),
		CacheParameterGroupName:    p.CacheParameterGroupName,
		CacheSecurityGroupNames:    p.CacheSecurityGroupNames,
		EngineVersion:              p.EngineVersion,
		NewAvailabilityZones:       p.PreferredAvailabilityZones,
		NotificationTopicArn:       p.NotificationTopicARN,
		NumCacheNodes:              aws.Int32(p.NumCacheNodes),
		PreferredMaintenanceWindow: p.PreferredMaintenanceWindow,
		SecurityGroupIds:           p.SecurityGroupIDs,
		SnapshotRetentionLimit:     p.SnapshotRetentionLimit,
		SnapshotWindow:             p.SnapshotWindow,
	}

	if p.EngineVersion != nil {
		version, err := getVersion(p.EngineVersion)
		if err != nil {
			return nil, err
		}
		c.EngineVersion = version
	} else {
		c.EngineVersion = p.EngineVersion
	}

	return c, nil
}

// GenerateClusterObservation produces a CacheClusterObservation object out of
// received elasticache.CacheCluster object.
func GenerateClusterObservation(c elasticachetypes.CacheCluster) cachev1alpha1.CacheClusterObservation {
	o := cachev1alpha1.CacheClusterObservation{
		AtRestEncryptionEnabled:   aws.ToBool(c.AtRestEncryptionEnabled),
		AuthTokenEnabled:          aws.ToBool(c.AtRestEncryptionEnabled),
		CacheClusterStatus:        aws.ToString(c.CacheClusterStatus),
		ClientDownloadLandingPage: aws.ToString(c.ClientDownloadLandingPage),
	}

	if len(c.CacheNodes) > 0 {
		cacheNodes := make([]cachev1alpha1.CacheNode, len(c.CacheNodes))
		for i, v := range c.CacheNodes {
			cacheNodes[i] = cachev1alpha1.CacheNode{
				CacheNodeID:              aws.ToString(v.CacheNodeId),
				CacheNodeStatus:          aws.ToString(v.CacheNodeStatus),
				CustomerAvailabilityZone: aws.ToString(v.CustomerAvailabilityZone),
				ParameterGroupStatus:     aws.ToString(v.ParameterGroupStatus),
				SourceCacheNodeID:        v.SourceCacheNodeId,
			}
			if v.Endpoint != nil {
				cacheNodes[i].Endpoint = &cachev1alpha1.Endpoint{
					Address: aws.ToString(v.Endpoint.Address),
					Port:    int(v.Endpoint.Port),
				}
			}
		}
		o.CacheNodes = cacheNodes
	}
	return o
}

// IsClusterNotFound returns true if the supplied error indicates a Cache Cluster
// already exists.
func IsClusterNotFound(err error) bool {
	var gnf *elasticachetypes.CacheClusterNotFoundFault
	return errors.As(err, &gnf)
}

// LateInitializeCluster assigns the observed configurations and assigns them to the
// corresponding fields in CacheClusterParameters in order to let user
// know the defaults and make the changes as wished on that value.
func LateInitializeCluster(p *cachev1alpha1.CacheClusterParameters, c elasticachetypes.CacheCluster) {
	p.SnapshotRetentionLimit = pointer.LateInitialize(p.SnapshotRetentionLimit, c.SnapshotRetentionLimit)
	p.SnapshotWindow = pointer.LateInitialize(p.SnapshotWindow, c.SnapshotWindow)
	p.CacheSubnetGroupName = pointer.LateInitialize(p.CacheSubnetGroupName, c.CacheSubnetGroupName)
	p.EngineVersion = pointer.LateInitialize(p.EngineVersion, c.EngineVersion)
	p.PreferredAvailabilityZone = pointer.LateInitialize(p.PreferredAvailabilityZone, c.PreferredAvailabilityZone)
	p.PreferredMaintenanceWindow = pointer.LateInitialize(p.PreferredMaintenanceWindow, c.PreferredMaintenanceWindow)
	p.ReplicationGroupID = pointer.LateInitialize(p.ReplicationGroupID, c.ReplicationGroupId)

	if c.NotificationConfiguration != nil {
		p.NotificationTopicARN = pointer.LateInitialize(p.NotificationTopicARN, c.NotificationConfiguration.TopicArn)
	}
	if c.CacheParameterGroup != nil {
		p.CacheParameterGroupName = pointer.LateInitialize(p.CacheParameterGroupName, c.CacheParameterGroup.CacheParameterGroupName)
	}
}

// GenerateCluster modifies elasticache.CacheCluster with values from cachev1alpha1.CacheClusterParameters
func GenerateCluster(name string, p cachev1alpha1.CacheClusterParameters, c *elasticachetypes.CacheCluster) {
	c.CacheClusterId = aws.String(name)
	c.CacheNodeType = aws.String(p.CacheNodeType)
	c.EngineVersion = p.EngineVersion
	c.NumCacheNodes = aws.Int32(p.NumCacheNodes)
	c.PreferredMaintenanceWindow = p.PreferredMaintenanceWindow
	c.SnapshotRetentionLimit = p.SnapshotRetentionLimit
	c.SnapshotWindow = p.SnapshotWindow

	if len(p.SecurityGroupIDs) > 0 {
		sg := make([]elasticachetypes.SecurityGroupMembership, len(p.SecurityGroupIDs))
		for i, v := range p.SecurityGroupIDs {
			sg[i] = elasticachetypes.SecurityGroupMembership{
				SecurityGroupId: aws.String(v),
				Status:          aws.String("active"),
			}
		}
		c.SecurityGroups = sg
	}

	if c.CacheParameterGroup != nil {
		c.CacheParameterGroup.CacheParameterGroupName = p.CacheParameterGroupName
	}

	if c.NotificationConfiguration != nil {
		c.NotificationConfiguration.TopicArn = p.NotificationTopicARN
	}
}

// IsClusterUpToDate checks whether current state is up-to-date compared to the given
// set of parameters.
func IsClusterUpToDate(name string, in *cachev1alpha1.CacheClusterParameters, observed *elasticachetypes.CacheCluster) (bool, error) {
	desired := (&convert.ConverterImpl{}).DeepCopyAWSCacheCluster(observed)
	GenerateCluster(name, *in, desired)

	if desired.EngineVersion != nil {
		observedVersion := observed.EngineVersion
		desiredVersion := desired.EngineVersion

		observedVersionSplit := strings.Split(aws.ToString(observedVersion), ".")
		desiredVersionSplit := strings.Split(aws.ToString(desiredVersion), ".")
		if observedVersionSplit[0] != desiredVersionSplit[0] {
			return false, nil
		}
		if len(desiredVersionSplit) > 1 {
			if observedVersionSplit[1] != desiredVersionSplit[1] {
				return false, nil
			}
		}
		// to ignore in following equal
		desired.EngineVersion = observed.EngineVersion
	}

	return cmp.Equal(desired, observed, cmpopts.EquateEmpty(), cmpopts.IgnoreTypes(document.NoSerde{})), nil
}

func getVersion(version *string) (*string, error) {
	versionSplit := strings.Split(aws.ToString(version), ".")
	version1, err := strconv.Atoi(versionSplit[0])
	if err != nil {
		return nil, errors.Wrap(err, errVersionInput)
	}
	versionOut := strconv.Itoa(version1)
	if len(versionSplit) > 1 {
		version2, err := strconv.Atoi(versionSplit[1])
		if err != nil {
			return nil, errors.Wrap(err, errVersionInput)
		}
		versionOut += "." + strconv.Itoa(version2)
	}
	return &versionOut, nil
}
