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

package fake

import (
	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/aws/aws-sdk-go-v2/service/eks/eksiface"
)

var _ eksiface.ClientAPI = &MockClient{}

// MockClient is a fake implementation of cloudmemorystore.Client.
type MockClient struct {
	eksiface.ClientAPI

	MockCreateClusterRequest        func(*eks.CreateClusterInput) eks.CreateClusterRequest
	MockDescribeClusterRequest      func(*eks.DescribeClusterInput) eks.DescribeClusterRequest
	MockUpdateClusterConfigRequest  func(*eks.UpdateClusterConfigInput) eks.UpdateClusterConfigRequest
	MockDeleteClusterRequest        func(*eks.DeleteClusterInput) eks.DeleteClusterRequest
	MockTagResourceRequest          func(*eks.TagResourceInput) eks.TagResourceRequest
	MockUntagResourceRequest        func(*eks.UntagResourceInput) eks.UntagResourceRequest
	MockUpdateClusterVersionRequest func(*eks.UpdateClusterVersionInput) eks.UpdateClusterVersionRequest

	MockDescribeNodegroupRequest      func(*eks.DescribeNodegroupInput) eks.DescribeNodegroupRequest
	MockCreateNodegroupRequest        func(*eks.CreateNodegroupInput) eks.CreateNodegroupRequest
	MockUpdateNodegroupVersionRequest func(*eks.UpdateNodegroupVersionInput) eks.UpdateNodegroupVersionRequest
	MockUpdateNodegroupConfigRequest  func(*eks.UpdateNodegroupConfigInput) eks.UpdateNodegroupConfigRequest
	MockDeleteNodegroupRequest        func(*eks.DeleteNodegroupInput) eks.DeleteNodegroupRequest

	MockDescribeFargateProfileRequest func(*eks.DescribeFargateProfileInput) eks.DescribeFargateProfileRequest
	MockCreateFargateProfileRequest   func(*eks.CreateFargateProfileInput) eks.CreateFargateProfileRequest
	MockDeleteFargateProfileRequest   func(*eks.DeleteFargateProfileInput) eks.DeleteFargateProfileRequest
}

// CreateClusterRequest calls the underlying MockCreateClusterRequest method.
func (c *MockClient) CreateClusterRequest(i *eks.CreateClusterInput) eks.CreateClusterRequest {
	return c.MockCreateClusterRequest(i)
}

// DescribeClusterRequest calls the underlying MockDescribeClusterRequest
// method.
func (c *MockClient) DescribeClusterRequest(i *eks.DescribeClusterInput) eks.DescribeClusterRequest {
	return c.MockDescribeClusterRequest(i)
}

// UpdateClusterConfigRequest calls the underlying
// MockUpdateClusterConfigRequest method.
func (c *MockClient) UpdateClusterConfigRequest(i *eks.UpdateClusterConfigInput) eks.UpdateClusterConfigRequest {
	return c.MockUpdateClusterConfigRequest(i)
}

// DeleteClusterRequest calls the underlying MockDeleteClusterRequest method.
func (c *MockClient) DeleteClusterRequest(i *eks.DeleteClusterInput) eks.DeleteClusterRequest {
	return c.MockDeleteClusterRequest(i)
}

// TagResourceRequest calls the underlying MockTagResourceRequest method.
func (c *MockClient) TagResourceRequest(i *eks.TagResourceInput) eks.TagResourceRequest {
	return c.MockTagResourceRequest(i)
}

// UntagResourceRequest calls the underlying MockUntagResourceRequest method.
func (c *MockClient) UntagResourceRequest(i *eks.UntagResourceInput) eks.UntagResourceRequest {
	return c.MockUntagResourceRequest(i)
}

// UpdateClusterVersionRequest calls the underlying
// MockUpdateClusterVersionRequest method.
func (c *MockClient) UpdateClusterVersionRequest(i *eks.UpdateClusterVersionInput) eks.UpdateClusterVersionRequest {
	return c.MockUpdateClusterVersionRequest(i)
}

// DescribeNodegroupRequest calls the underlying MockDescribeNodegroupRequest
// method.
func (c *MockClient) DescribeNodegroupRequest(i *eks.DescribeNodegroupInput) eks.DescribeNodegroupRequest {
	return c.MockDescribeNodegroupRequest(i)
}

// CreateNodegroupRequest calls the underlying MockCreateNodegroupRequest
// method.
func (c *MockClient) CreateNodegroupRequest(i *eks.CreateNodegroupInput) eks.CreateNodegroupRequest {
	return c.MockCreateNodegroupRequest(i)
}

// UpdateNodegroupVersionRequest calls the underlying
// MockUpdateNodegroupVersionRequest method.
func (c *MockClient) UpdateNodegroupVersionRequest(i *eks.UpdateNodegroupVersionInput) eks.UpdateNodegroupVersionRequest {
	return c.MockUpdateNodegroupVersionRequest(i)
}

// UpdateNodegroupConfigRequest calls the underlying
// MockUpdateNodegroupConfigRequest method.
func (c *MockClient) UpdateNodegroupConfigRequest(i *eks.UpdateNodegroupConfigInput) eks.UpdateNodegroupConfigRequest {
	return c.MockUpdateNodegroupConfigRequest(i)
}

// DeleteNodegroupRequest calls the underlying MockDeleteNodegroupRequest
// method.
func (c *MockClient) DeleteNodegroupRequest(i *eks.DeleteNodegroupInput) eks.DeleteNodegroupRequest {
	return c.MockDeleteNodegroupRequest(i)
}

// DescribeFargateProfileRequest calls the underlying MockDescribeFargateProfileRequest
// method.
func (c *MockClient) DescribeFargateProfileRequest(i *eks.DescribeFargateProfileInput) eks.DescribeFargateProfileRequest {
	return c.MockDescribeFargateProfileRequest(i)
}

// CreateFargateProfileRequest calls the underlying MockCreateFargateProfileRequest
// method.
func (c *MockClient) CreateFargateProfileRequest(i *eks.CreateFargateProfileInput) eks.CreateFargateProfileRequest {
	return c.MockCreateFargateProfileRequest(i)
}

// DeleteFargateProfileRequest calls the underlying MockDeleteFargateProfileRequest
// method.
func (c *MockClient) DeleteFargateProfileRequest(i *eks.DeleteFargateProfileInput) eks.DeleteFargateProfileRequest {
	return c.MockDeleteFargateProfileRequest(i)
}
