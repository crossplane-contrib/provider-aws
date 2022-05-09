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

package fake

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/elasticache"
)

// MockClient is a fake implementation of cloudmemorystore.Client.
type MockClient struct {
	elasticache.Client

	MockDescribeReplicationGroups func(context.Context, *elasticache.DescribeReplicationGroupsInput, []func(*elasticache.Options)) (*elasticache.DescribeReplicationGroupsOutput, error)
	MockCreateReplicationGroup    func(context.Context, *elasticache.CreateReplicationGroupInput, []func(*elasticache.Options)) (*elasticache.CreateReplicationGroupOutput, error)
	MockModifyReplicationGroup    func(context.Context, *elasticache.ModifyReplicationGroupInput, []func(*elasticache.Options)) (*elasticache.ModifyReplicationGroupOutput, error)
	MockDeleteReplicationGroup    func(context.Context, *elasticache.DeleteReplicationGroupInput, []func(*elasticache.Options)) (*elasticache.DeleteReplicationGroupOutput, error)

	MockDescribeCacheSubnetGroups func(context.Context, *elasticache.DescribeCacheSubnetGroupsInput, []func(*elasticache.Options)) (*elasticache.DescribeCacheSubnetGroupsOutput, error)
	MockCreateCacheSubnetGroup    func(context.Context, *elasticache.CreateCacheSubnetGroupInput, []func(*elasticache.Options)) (*elasticache.CreateCacheSubnetGroupOutput, error)
	MockModifyCacheSubnetGroup    func(context.Context, *elasticache.ModifyCacheSubnetGroupInput, []func(*elasticache.Options)) (*elasticache.ModifyCacheSubnetGroupOutput, error)
	MockDeleteCacheSubnetGroup    func(context.Context, *elasticache.DeleteCacheSubnetGroupInput, []func(*elasticache.Options)) (*elasticache.DeleteCacheSubnetGroupOutput, error)

	MockDescribeCacheClusters func(context.Context, *elasticache.DescribeCacheClustersInput, []func(*elasticache.Options)) (*elasticache.DescribeCacheClustersOutput, error)
	MockCreateCacheCluster    func(context.Context, *elasticache.CreateCacheClusterInput, []func(*elasticache.Options)) (*elasticache.CreateCacheClusterOutput, error)
	MockDeleteCacheCluster    func(context.Context, *elasticache.DeleteCacheClusterInput, []func(*elasticache.Options)) (*elasticache.DeleteCacheClusterOutput, error)
	MockModifyCacheCluster    func(context.Context, *elasticache.ModifyCacheClusterInput, []func(*elasticache.Options)) (*elasticache.ModifyCacheClusterOutput, error)
	MockIncreaseReplicaCount  func(context.Context, *elasticache.IncreaseReplicaCountInput, []func(*elasticache.Options)) (*elasticache.IncreaseReplicaCountOutput, error)
	MockDecreaseReplicaCount  func(context.Context, *elasticache.DecreaseReplicaCountInput, []func(*elasticache.Options)) (*elasticache.DecreaseReplicaCountOutput, error)

	MockModifyReplicationGroupShardConfiguration func(context.Context, *elasticache.ModifyReplicationGroupShardConfigurationInput, []func(*elasticache.Options)) (*elasticache.ModifyReplicationGroupShardConfigurationOutput, error)

	MockListTagsForResource    func(context.Context, *elasticache.ListTagsForResourceInput, []func(*elasticache.Options)) (*elasticache.ListTagsForResourceOutput, error)
	MockAddTagsToResource      func(context.Context, *elasticache.AddTagsToResourceInput, []func(*elasticache.Options)) (*elasticache.AddTagsToResourceOutput, error)
	MockRemoveTagsFromResource func(context.Context, *elasticache.RemoveTagsFromResourceInput, []func(*elasticache.Options)) (*elasticache.RemoveTagsFromResourceOutput, error)
}

// DescribeReplicationGroups calls the underlying
// MockDescribeReplicationGroups method.
func (c *MockClient) DescribeReplicationGroups(ctx context.Context, i *elasticache.DescribeReplicationGroupsInput, opts ...func(*elasticache.Options)) (*elasticache.DescribeReplicationGroupsOutput, error) {
	return c.MockDescribeReplicationGroups(ctx, i, opts)
}

// CreateReplicationGroup calls the underlying
// MockCreateReplicationGroup method.
func (c *MockClient) CreateReplicationGroup(ctx context.Context, i *elasticache.CreateReplicationGroupInput, opts ...func(*elasticache.Options)) (*elasticache.CreateReplicationGroupOutput, error) {
	return c.MockCreateReplicationGroup(ctx, i, opts)
}

// ModifyReplicationGroup calls the underlying
// MockModifyReplicationGroup method.
func (c *MockClient) ModifyReplicationGroup(ctx context.Context, i *elasticache.ModifyReplicationGroupInput, opts ...func(*elasticache.Options)) (*elasticache.ModifyReplicationGroupOutput, error) {
	return c.MockModifyReplicationGroup(ctx, i, opts)
}

// DeleteReplicationGroup calls the underlying
// MockDeleteReplicationGroup method.
func (c *MockClient) DeleteReplicationGroup(ctx context.Context, i *elasticache.DeleteReplicationGroupInput, opts ...func(*elasticache.Options)) (*elasticache.DeleteReplicationGroupOutput, error) {
	return c.MockDeleteReplicationGroup(ctx, i, opts)
}

// DescribeCacheClusters calls the underlying
// MockDescribeCacheClusters method.
func (c *MockClient) DescribeCacheClusters(ctx context.Context, i *elasticache.DescribeCacheClustersInput, opts ...func(*elasticache.Options)) (*elasticache.DescribeCacheClustersOutput, error) {
	return c.MockDescribeCacheClusters(ctx, i, opts)
}

// ModifyReplicationGroupShardConfiguration calls the underlying
// MockModifyReplicationGroupShardConfiguration method.
func (c *MockClient) ModifyReplicationGroupShardConfiguration(ctx context.Context, i *elasticache.ModifyReplicationGroupShardConfigurationInput, opts ...func(*elasticache.Options)) (*elasticache.ModifyReplicationGroupShardConfigurationOutput, error) {
	return c.MockModifyReplicationGroupShardConfiguration(ctx, i, opts)
}

// DescribeCacheSubnetGroups calls the underlying
// MockDescribeCacheSubnetGroups method.
func (c *MockClient) DescribeCacheSubnetGroups(ctx context.Context, i *elasticache.DescribeCacheSubnetGroupsInput, opts ...func(*elasticache.Options)) (*elasticache.DescribeCacheSubnetGroupsOutput, error) {
	return c.MockDescribeCacheSubnetGroups(ctx, i, opts)
}

// CreateCacheSubnetGroup calls the underlying
// MockCreateCacheSubnetGroup method.
func (c *MockClient) CreateCacheSubnetGroup(ctx context.Context, i *elasticache.CreateCacheSubnetGroupInput, opts ...func(*elasticache.Options)) (*elasticache.CreateCacheSubnetGroupOutput, error) {
	return c.MockCreateCacheSubnetGroup(ctx, i, opts)
}

// ModifyCacheSubnetGroup calls the underlying
// MockCreateCacheSubnetGroup method.
func (c *MockClient) ModifyCacheSubnetGroup(ctx context.Context, i *elasticache.ModifyCacheSubnetGroupInput, opts ...func(*elasticache.Options)) (*elasticache.ModifyCacheSubnetGroupOutput, error) {
	return c.MockModifyCacheSubnetGroup(ctx, i, opts)
}

// DeleteCacheSubnetGroup calls the underlying
// MockDeleteCacheSubnetGroup method.
func (c *MockClient) DeleteCacheSubnetGroup(ctx context.Context, i *elasticache.DeleteCacheSubnetGroupInput, opts ...func(*elasticache.Options)) (*elasticache.DeleteCacheSubnetGroupOutput, error) {
	return c.MockDeleteCacheSubnetGroup(ctx, i, opts)
}

// CreateCacheCluster calls the underlying
// MockCreateCacheCluster method.
func (c *MockClient) CreateCacheCluster(ctx context.Context, i *elasticache.CreateCacheClusterInput, opts ...func(*elasticache.Options)) (*elasticache.CreateCacheClusterOutput, error) {
	return c.MockCreateCacheCluster(ctx, i, opts)
}

// DeleteCacheCluster calls the underlying
// MockDeleteCacheCluster method.
func (c *MockClient) DeleteCacheCluster(ctx context.Context, i *elasticache.DeleteCacheClusterInput, opts ...func(*elasticache.Options)) (*elasticache.DeleteCacheClusterOutput, error) {
	return c.MockDeleteCacheCluster(ctx, i, opts)
}

// ModifyCacheCluster calls the underlying
// MockModifyCacheCluster method.
func (c *MockClient) ModifyCacheCluster(ctx context.Context, i *elasticache.ModifyCacheClusterInput, opts ...func(*elasticache.Options)) (*elasticache.ModifyCacheClusterOutput, error) {
	return c.MockModifyCacheCluster(ctx, i, opts)
}

// ListTagsForResource calls the underlying
// MockListTagsForResource method
func (c *MockClient) ListTagsForResource(ctx context.Context, i *elasticache.ListTagsForResourceInput, opts ...func(*elasticache.Options)) (*elasticache.ListTagsForResourceOutput, error) {
	return c.MockListTagsForResource(ctx, i, opts)
}

// AddTagsToResource calls the underlying
// MockAddTagsToResource method
func (c *MockClient) AddTagsToResource(ctx context.Context, i *elasticache.AddTagsToResourceInput, opts ...func(*elasticache.Options)) (*elasticache.AddTagsToResourceOutput, error) {
	return c.MockAddTagsToResource(ctx, i, opts)
}

// RemoveTagsFromResource calls the underlying
// MockRemoveTagsFromResource method
func (c *MockClient) RemoveTagsFromResource(ctx context.Context, i *elasticache.RemoveTagsFromResourceInput, opts ...func(*elasticache.Options)) (*elasticache.RemoveTagsFromResourceOutput, error) {
	return c.MockRemoveTagsFromResource(ctx, i, opts)
}

// DecreaseReplicaCount calls the underlying
// MockDecreaseReplicaCount method
func (c *MockClient) DecreaseReplicaCount(ctx context.Context, i *elasticache.DecreaseReplicaCountInput, opts ...func(*elasticache.Options)) (*elasticache.DecreaseReplicaCountOutput, error) {
	return c.MockDecreaseReplicaCount(ctx, i, opts)
}

// IncreaseReplicaCount calls the underlying
// MockIncreaseReplicaCount method
func (c *MockClient) IncreaseReplicaCount(ctx context.Context, i *elasticache.IncreaseReplicaCountInput, opts ...func(*elasticache.Options)) (*elasticache.IncreaseReplicaCountOutput, error) {
	return c.MockIncreaseReplicaCount(ctx, i, opts)
}
