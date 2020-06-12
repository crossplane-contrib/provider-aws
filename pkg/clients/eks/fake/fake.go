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
	MockUpdateClusterVersionRequest func(*eks.UpdateClusterVersionInput) eks.UpdateClusterVersionRequest
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

// UpdateClusterVersionRequest calls the underlying
// MockUpdateClusterVersionRequest method.
func (c *MockClient) UpdateClusterVersionRequest(i *eks.UpdateClusterVersionInput) eks.UpdateClusterVersionRequest {
	return c.MockUpdateClusterVersionRequest(i)
}
