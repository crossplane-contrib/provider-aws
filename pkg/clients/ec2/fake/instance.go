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
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ec2"

	clientset "github.com/crossplane/provider-aws/pkg/clients/ec2"
)

// this ensures that the mock implements the client interface
var _ clientset.InstanceClient = (*MockInstanceClient)(nil)

// MockInstanceClient is a type that implements all the methods for MockInstanceClient interface
type MockInstanceClient struct {
	MockRunInstances              func(context.Context, *ec2.RunInstancesInput, []func(*ec2.Options)) (*ec2.RunInstancesOutput, error)
	MockTerminateInstances        func(context.Context, *ec2.TerminateInstancesInput, []func(*ec2.Options)) (*ec2.TerminateInstancesOutput, error)
	MockDescribeInstances         func(context.Context, *ec2.DescribeInstancesInput, []func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error)
	MockDescribeInstanceAttribute func(context.Context, *ec2.DescribeInstanceAttributeInput, []func(*ec2.Options)) (*ec2.DescribeInstanceAttributeOutput, error)
	MockModifyInstanceAttribute   func(context.Context, *ec2.ModifyInstanceAttributeInput, []func(*ec2.Options)) (*ec2.ModifyInstanceAttributeOutput, error)
	MockCreateTags                func(context.Context, *ec2.CreateTagsInput, []func(*ec2.Options)) (*ec2.CreateTagsOutput, error)
}

// RunInstances mocks RunInstances method
func (m *MockInstanceClient) RunInstances(ctx context.Context, input *ec2.RunInstancesInput, opts ...func(*ec2.Options)) (*ec2.RunInstancesOutput, error) {
	return m.MockRunInstances(ctx, input, opts)
}

// TerminateInstances mocks TerminateInstances method
func (m *MockInstanceClient) TerminateInstances(ctx context.Context, input *ec2.TerminateInstancesInput, opts ...func(*ec2.Options)) (*ec2.TerminateInstancesOutput, error) {
	return m.MockTerminateInstances(ctx, input, opts)
}

// DescribeInstances mocks DescribeInstances method
func (m *MockInstanceClient) DescribeInstances(ctx context.Context, input *ec2.DescribeInstancesInput, opts ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error) {
	return m.MockDescribeInstances(ctx, input, opts)
}

// DescribeInstanceAttribute mocks DescribeInstanceAttribute method
func (m *MockInstanceClient) DescribeInstanceAttribute(ctx context.Context, input *ec2.DescribeInstanceAttributeInput, opts ...func(*ec2.Options)) (*ec2.DescribeInstanceAttributeOutput, error) {
	return m.MockDescribeInstanceAttribute(ctx, input, opts)
}

// ModifyInstanceAttribute mocks ModifyInstanceAttribute method
func (m *MockInstanceClient) ModifyInstanceAttribute(ctx context.Context, input *ec2.ModifyInstanceAttributeInput, opts ...func(*ec2.Options)) (*ec2.ModifyInstanceAttributeOutput, error) {
	return m.MockModifyInstanceAttribute(ctx, input, opts)
}

// CreateTags mocks CreateTags method
func (m *MockInstanceClient) CreateTags(ctx context.Context, input *ec2.CreateTagsInput, opts ...func(*ec2.Options)) (*ec2.CreateTagsOutput, error) {
	return m.MockCreateTags(ctx, input, opts)
}
