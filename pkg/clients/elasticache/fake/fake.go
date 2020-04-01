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
	"github.com/aws/aws-sdk-go-v2/service/elasticache"
	"github.com/aws/aws-sdk-go-v2/service/elasticache/elasticacheiface"
)

var _ elasticacheiface.ClientAPI = &MockClient{}

// MockClient is a fake implementation of cloudmemorystore.Client.
type MockClient struct {
	elasticacheiface.ClientAPI

	MockDescribeReplicationGroupsRequest func(*elasticache.DescribeReplicationGroupsInput) elasticache.DescribeReplicationGroupsRequest
	MockCreateReplicationGroupRequest    func(*elasticache.CreateReplicationGroupInput) elasticache.CreateReplicationGroupRequest
	MockModifyReplicationGroupRequest    func(*elasticache.ModifyReplicationGroupInput) elasticache.ModifyReplicationGroupRequest
	MockDeleteReplicationGroupRequest    func(*elasticache.DeleteReplicationGroupInput) elasticache.DeleteReplicationGroupRequest
	MockDescribeCacheClustersRequest     func(*elasticache.DescribeCacheClustersInput) elasticache.DescribeCacheClustersRequest

	MockDescribeCacheSubnetGroupsRequest func(*elasticache.DescribeCacheSubnetGroupsInput) elasticache.DescribeCacheSubnetGroupsRequest
	MockCreateCacheSubnetGroupRequest    func(*elasticache.CreateCacheSubnetGroupInput) elasticache.CreateCacheSubnetGroupRequest
	MockModifyCacheSubnetGroupRequest    func(*elasticache.ModifyCacheSubnetGroupInput) elasticache.ModifyCacheSubnetGroupRequest
	MockDeleteCacheSubnetGroupRequest    func(*elasticache.DeleteCacheSubnetGroupInput) elasticache.DeleteCacheSubnetGroupRequest
}

// DescribeReplicationGroupsRequest calls the underlying
// MockDescribeReplicationGroupsRequest method.
func (c *MockClient) DescribeReplicationGroupsRequest(i *elasticache.DescribeReplicationGroupsInput) elasticache.DescribeReplicationGroupsRequest {
	return c.MockDescribeReplicationGroupsRequest(i)
}

// CreateReplicationGroupRequest calls the underlying
// MockCreateReplicationGroupRequest method.
func (c *MockClient) CreateReplicationGroupRequest(i *elasticache.CreateReplicationGroupInput) elasticache.CreateReplicationGroupRequest {
	return c.MockCreateReplicationGroupRequest(i)
}

// ModifyReplicationGroupRequest calls the underlying
// MockModifyReplicationGroupRequest method.
func (c *MockClient) ModifyReplicationGroupRequest(i *elasticache.ModifyReplicationGroupInput) elasticache.ModifyReplicationGroupRequest {
	return c.MockModifyReplicationGroupRequest(i)
}

// DeleteReplicationGroupRequest calls the underlying
// MockDeleteReplicationGroupRequest method.
func (c *MockClient) DeleteReplicationGroupRequest(i *elasticache.DeleteReplicationGroupInput) elasticache.DeleteReplicationGroupRequest {
	return c.MockDeleteReplicationGroupRequest(i)
}

// DescribeCacheClustersRequest calls the underlying
// MockDescribeCacheClustersRequest method.
func (c *MockClient) DescribeCacheClustersRequest(i *elasticache.DescribeCacheClustersInput) elasticache.DescribeCacheClustersRequest {
	return c.MockDescribeCacheClustersRequest(i)
}

// DescribeCacheSubnetGroupsRequest calls the underlying
// MockDescribeCacheSubnetGroupsRequest method.
func (c *MockClient) DescribeCacheSubnetGroupsRequest(i *elasticache.DescribeCacheSubnetGroupsInput) elasticache.DescribeCacheSubnetGroupsRequest {
	return c.MockDescribeCacheSubnetGroupsRequest(i)
}

// CreateCacheSubnetGroupRequest calls the underlying
// MockCreateCacheSubnetGroupRequest method.
func (c *MockClient) CreateCacheSubnetGroupRequest(i *elasticache.CreateCacheSubnetGroupInput) elasticache.CreateCacheSubnetGroupRequest {
	return c.MockCreateCacheSubnetGroupRequest(i)
}

// ModifyCacheSubnetGroupRequest calls the underlying
// MockCreateCacheSubnetGroupRequest method.
func (c *MockClient) ModifyCacheSubnetGroupRequest(i *elasticache.ModifyCacheSubnetGroupInput) elasticache.ModifyCacheSubnetGroupRequest {
	return c.MockModifyCacheSubnetGroupRequest(i)
}

// DeleteCacheSubnetGroupRequest calls the underlying
// MockDeleteCacheSubnetGroupRequest method.
func (c *MockClient) DeleteCacheSubnetGroupRequest(i *elasticache.DeleteCacheSubnetGroupInput) elasticache.DeleteCacheSubnetGroupRequest {
	return c.MockDeleteCacheSubnetGroupRequest(i)
}
