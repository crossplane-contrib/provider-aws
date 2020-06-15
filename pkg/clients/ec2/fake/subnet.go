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
	"github.com/aws/aws-sdk-go-v2/service/ec2"

	clientset "github.com/crossplane/provider-aws/pkg/clients/ec2"
)

// this ensures that the mock implements the client interface
var _ clientset.SubnetClient = (*MockSubnetClient)(nil)

// MockSubnetClient is a type that implements all the methods for SubnetClient interface
type MockSubnetClient struct {
	MockCreate     func(*ec2.CreateSubnetInput) ec2.CreateSubnetRequest
	MockDelete     func(*ec2.DeleteSubnetInput) ec2.DeleteSubnetRequest
	MockDescribe   func(*ec2.DescribeSubnetsInput) ec2.DescribeSubnetsRequest
	MockModify     func(*ec2.ModifySubnetAttributeInput) ec2.ModifySubnetAttributeRequest
	MockCreateTags func(*ec2.CreateTagsInput) ec2.CreateTagsRequest
}

// CreateSubnetRequest mocks CreateSubnetRequest method
func (m *MockSubnetClient) CreateSubnetRequest(input *ec2.CreateSubnetInput) ec2.CreateSubnetRequest {
	return m.MockCreate(input)
}

// DeleteSubnetRequest mocks DeleteSubnetRequest method
func (m *MockSubnetClient) DeleteSubnetRequest(input *ec2.DeleteSubnetInput) ec2.DeleteSubnetRequest {
	return m.MockDelete(input)
}

// DescribeSubnetsRequest mocks DescribeSubnetsRequest method
func (m *MockSubnetClient) DescribeSubnetsRequest(input *ec2.DescribeSubnetsInput) ec2.DescribeSubnetsRequest {
	return m.MockDescribe(input)
}

// ModifySubnetAttributeRequest mocks ModifySubnetAttributeInput method
func (m *MockSubnetClient) ModifySubnetAttributeRequest(input *ec2.ModifySubnetAttributeInput) ec2.ModifySubnetAttributeRequest {
	return m.MockModify(input)
}

// CreateTagsRequest mocks CreateTagsInput method
func (m *MockSubnetClient) CreateTagsRequest(input *ec2.CreateTagsInput) ec2.CreateTagsRequest {
	return m.MockCreateTags(input)
}
