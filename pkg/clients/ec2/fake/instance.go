/*
Copyright 2021 The Crossplane Authors.

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
var _ clientset.InstanceClient = (*MockInstanceClient)(nil)

// MockInstanceClient is a type that implements all the methods for MockInstanceClient interface
type MockInstanceClient struct {
	MockRunInstancesRequest              func(*ec2.RunInstancesInput) ec2.RunInstancesRequest
	MockTerminateInstancesRequest        func(*ec2.TerminateInstancesInput) ec2.TerminateInstancesRequest
	MockDescribeInstancesRequest         func(*ec2.DescribeInstancesInput) ec2.DescribeInstancesRequest
	MockDescribeInstanceAttributeRequest func(*ec2.DescribeInstanceAttributeInput) ec2.DescribeInstanceAttributeRequest
	MockModifyInstanceAttributeRequest   func(*ec2.ModifyInstanceAttributeInput) ec2.ModifyInstanceAttributeRequest
}

// RunInstancesRequest mocks RunInstancesRequest method
func (m *MockInstanceClient) RunInstancesRequest(input *ec2.RunInstancesInput) ec2.RunInstancesRequest {
	return m.MockRunInstancesRequest(input)
}

// TerminateInstancesRequest mocks TerminateInstancesRequest method
func (m *MockInstanceClient) TerminateInstancesRequest(input *ec2.TerminateInstancesInput) ec2.TerminateInstancesRequest {
	return m.MockTerminateInstancesRequest(input)
}

// DescribeInstancesRequest mocks DescribeInstancesRequest method
func (m *MockInstanceClient) DescribeInstancesRequest(input *ec2.DescribeInstancesInput) ec2.DescribeInstancesRequest {
	return m.MockDescribeInstancesRequest(input)
}

// DescribeInstanceAttributeRequest mocks DescribeInstanceAttributeRequest method
func (m *MockInstanceClient) DescribeInstanceAttributeRequest(input *ec2.DescribeInstanceAttributeInput) ec2.DescribeInstanceAttributeRequest {
	return m.MockDescribeInstanceAttributeRequest(input)
}

// ModifyInstanceAttributeRequest mocks ModifyInstanceAttributeRequest method
func (m *MockInstanceClient) ModifyInstanceAttributeRequest(input *ec2.ModifyInstanceAttributeInput) ec2.ModifyInstanceAttributeRequest {
	return m.MockModifyInstanceAttributeRequest(input)
}
