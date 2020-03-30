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
var _ clientset.VPCClient = (*MockVPCClient)(nil)

// MockVPCClient is a type that implements all the methods for VPCClient interface
type MockVPCClient struct {
	MockCreate            func(*ec2.CreateVpcInput) ec2.CreateVpcRequest
	MockDelete            func(*ec2.DeleteVpcInput) ec2.DeleteVpcRequest
	MockDescribe          func(*ec2.DescribeVpcsInput) ec2.DescribeVpcsRequest
	MockModifyAttribute   func(*ec2.ModifyVpcAttributeInput) ec2.ModifyVpcAttributeRequest
	MockModifyTenancy     func(*ec2.ModifyVpcTenancyInput) ec2.ModifyVpcTenancyRequest
	MockCreateTagsRequest func(*ec2.CreateTagsInput) ec2.CreateTagsRequest
}

// CreateVpcRequest mocks CreateVpcRequest method
func (m *MockVPCClient) CreateVpcRequest(input *ec2.CreateVpcInput) ec2.CreateVpcRequest {
	return m.MockCreate(input)
}

// DeleteVpcRequest mocks DeleteVpcRequest method
func (m *MockVPCClient) DeleteVpcRequest(input *ec2.DeleteVpcInput) ec2.DeleteVpcRequest {
	return m.MockDelete(input)
}

// DescribeVpcsRequest mocks DescribeVpcsRequest method
func (m *MockVPCClient) DescribeVpcsRequest(input *ec2.DescribeVpcsInput) ec2.DescribeVpcsRequest {
	return m.MockDescribe(input)
}

// ModifyVpcAttributeRequest mocks ModifyVpcAttributeRequest method
func (m *MockVPCClient) ModifyVpcAttributeRequest(input *ec2.ModifyVpcAttributeInput) ec2.ModifyVpcAttributeRequest {
	return m.MockModifyAttribute(input)
}

// ModifyVpcTenancyRequest mocks ModifyVpcTenancyRequest method
func (m *MockVPCClient) ModifyVpcTenancyRequest(input *ec2.ModifyVpcTenancyInput) ec2.ModifyVpcTenancyRequest {
	return m.MockModifyTenancy(input)
}

// CreateTagsRequest mocks CreateTagsRequest method
func (m *MockVPCClient) CreateTagsRequest(input *ec2.CreateTagsInput) ec2.CreateTagsRequest {
	return m.MockCreateTagsRequest(input)
}
