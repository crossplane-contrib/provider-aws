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
	"github.com/aws/aws-sdk-go-v2/service/ec2"

	clientset "github.com/crossplane/provider-aws/pkg/clients/ec2"
)

// this ensures that the mock implements the client interface
var _ clientset.ElasticIPClient = (*MockElasticIPClient)(nil)

// MockElasticIPClient is a type that implements all the methods for ElasticIPClient interface
type MockElasticIPClient struct {
	MockAllocate          func(*ec2.AllocateAddressInput) ec2.AllocateAddressRequest
	MockRelease           func(*ec2.ReleaseAddressInput) ec2.ReleaseAddressRequest
	MockDescribe          func(*ec2.DescribeAddressesInput) ec2.DescribeAddressesRequest
	MockCreateTagsRequest func(*ec2.CreateTagsInput) ec2.CreateTagsRequest
}

// AllocateAddressRequest mocks AllocateAddressRequest method
func (m *MockElasticIPClient) AllocateAddressRequest(input *ec2.AllocateAddressInput) ec2.AllocateAddressRequest {
	return m.MockAllocate(input)
}

// ReleaseAddressRequest mocks ReleaseAddressRequest method
func (m *MockElasticIPClient) ReleaseAddressRequest(input *ec2.ReleaseAddressInput) ec2.ReleaseAddressRequest {
	return m.MockRelease(input)
}

// DescribeAddressesRequest mocks DescribeAddressesRequest method
func (m *MockElasticIPClient) DescribeAddressesRequest(input *ec2.DescribeAddressesInput) ec2.DescribeAddressesRequest {
	return m.MockDescribe(input)
}

// CreateTagsRequest mocks CreateTagsRequest method
func (m *MockElasticIPClient) CreateTagsRequest(input *ec2.CreateTagsInput) ec2.CreateTagsRequest {
	return m.MockCreateTagsRequest(input)
}
