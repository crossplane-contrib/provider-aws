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
var _ clientset.SecurityGroupClient = (*MockSecurityGroupClient)(nil)

// MockSecurityGroupClient is a type that implements all the methods for SecurityGroupClient interface
type MockSecurityGroupClient struct {
	MockCreate          func(*ec2.CreateSecurityGroupInput) ec2.CreateSecurityGroupRequest
	MockDelete          func(*ec2.DeleteSecurityGroupInput) ec2.DeleteSecurityGroupRequest
	MockDescribe        func(*ec2.DescribeSecurityGroupsInput) ec2.DescribeSecurityGroupsRequest
	MockAuthorizeIgress func(*ec2.AuthorizeSecurityGroupIngressInput) ec2.AuthorizeSecurityGroupIngressRequest
	MockAuthorizeEgress func(*ec2.AuthorizeSecurityGroupEgressInput) ec2.AuthorizeSecurityGroupEgressRequest
}

// CreateSecurityGroupRequest mocks CreateSecurityGroupRequest method
func (m *MockSecurityGroupClient) CreateSecurityGroupRequest(input *ec2.CreateSecurityGroupInput) ec2.CreateSecurityGroupRequest {
	return m.MockCreate(input)
}

// DeleteSecurityGroupRequest mocks DeleteSecurityGroupRequest method
func (m *MockSecurityGroupClient) DeleteSecurityGroupRequest(input *ec2.DeleteSecurityGroupInput) ec2.DeleteSecurityGroupRequest {
	return m.MockDelete(input)
}

// DescribeSecurityGroupsRequest mocks DescribeSecurityGroupsRequest method
func (m *MockSecurityGroupClient) DescribeSecurityGroupsRequest(input *ec2.DescribeSecurityGroupsInput) ec2.DescribeSecurityGroupsRequest {
	return m.MockDescribe(input)
}

// AuthorizeSecurityGroupIngressRequest mocks AuthorizeSecurityGroupIngressRequest method
func (m *MockSecurityGroupClient) AuthorizeSecurityGroupIngressRequest(input *ec2.AuthorizeSecurityGroupIngressInput) ec2.AuthorizeSecurityGroupIngressRequest {
	return m.MockAuthorizeIgress(input)
}

// AuthorizeSecurityGroupEgressRequest mocks AuthorizeSecurityGroupEgressRequest method
func (m *MockSecurityGroupClient) AuthorizeSecurityGroupEgressRequest(input *ec2.AuthorizeSecurityGroupEgressInput) ec2.AuthorizeSecurityGroupEgressRequest {
	return m.MockAuthorizeEgress(input)
}
