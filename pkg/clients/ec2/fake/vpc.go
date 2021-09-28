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

	"github.com/aws/aws-sdk-go-v2/service/ec2"

	clientset "github.com/crossplane/provider-aws/pkg/clients/ec2"
)

// this ensures that the mock implements the client interface
var _ clientset.VPCClient = (*MockVPCClient)(nil)

// MockVPCClient is a type that implements all the methods for VPCClient interface
type MockVPCClient struct {
	MockCreate               func(ctx context.Context, input *ec2.CreateVpcInput, opts []func(*ec2.Options)) (*ec2.CreateVpcOutput, error)
	MockDelete               func(ctx context.Context, input *ec2.DeleteVpcInput, opts []func(*ec2.Options)) (*ec2.DeleteVpcOutput, error)
	MockDescribe             func(ctx context.Context, input *ec2.DescribeVpcsInput, opts []func(*ec2.Options)) (*ec2.DescribeVpcsOutput, error)
	MockModifyAttribute      func(ctx context.Context, input *ec2.ModifyVpcAttributeInput, opts []func(*ec2.Options)) (*ec2.ModifyVpcAttributeOutput, error)
	MockModifyTenancy        func(ctx context.Context, input *ec2.ModifyVpcTenancyInput, opts []func(*ec2.Options)) (*ec2.ModifyVpcTenancyOutput, error)
	MockCreateTags           func(ctx context.Context, input *ec2.CreateTagsInput, opts []func(*ec2.Options)) (*ec2.CreateTagsOutput, error)
	MockDescribeVpcAttribute func(ctx context.Context, input *ec2.DescribeVpcAttributeInput, opts []func(*ec2.Options)) (*ec2.DescribeVpcAttributeOutput, error)
}

// CreateVpc mocks CreateVpc method
func (m *MockVPCClient) CreateVpc(ctx context.Context, input *ec2.CreateVpcInput, opts ...func(*ec2.Options)) (*ec2.CreateVpcOutput, error) {
	return m.MockCreate(ctx, input, opts)
}

// DeleteVpc mocks DeleteVpc method
func (m *MockVPCClient) DeleteVpc(ctx context.Context, input *ec2.DeleteVpcInput, opts ...func(*ec2.Options)) (*ec2.DeleteVpcOutput, error) {
	return m.MockDelete(ctx, input, opts)
}

// DescribeVpcs mocks DescribeVpcs method
func (m *MockVPCClient) DescribeVpcs(ctx context.Context, input *ec2.DescribeVpcsInput, opts ...func(*ec2.Options)) (*ec2.DescribeVpcsOutput, error) {
	return m.MockDescribe(ctx, input, opts)
}

// ModifyVpcAttribute mocks ModifyVpcAttribute method
func (m *MockVPCClient) ModifyVpcAttribute(ctx context.Context, input *ec2.ModifyVpcAttributeInput, opts ...func(*ec2.Options)) (*ec2.ModifyVpcAttributeOutput, error) {
	return m.MockModifyAttribute(ctx, input, opts)
}

// ModifyVpcTenancy mocks ModifyVpcTenancy method
func (m *MockVPCClient) ModifyVpcTenancy(ctx context.Context, input *ec2.ModifyVpcTenancyInput, opts ...func(*ec2.Options)) (*ec2.ModifyVpcTenancyOutput, error) {
	return m.MockModifyTenancy(ctx, input, opts)
}

// CreateTags mocks CreateTags method
func (m *MockVPCClient) CreateTags(ctx context.Context, input *ec2.CreateTagsInput, opts ...func(*ec2.Options)) (*ec2.CreateTagsOutput, error) {
	return m.MockCreateTags(ctx, input, opts)
}

// DescribeVpcAttribute mocks DescribeVpcAttribute method
func (m *MockVPCClient) DescribeVpcAttribute(ctx context.Context, input *ec2.DescribeVpcAttributeInput, opts ...func(*ec2.Options)) (*ec2.DescribeVpcAttributeOutput, error) {
	return m.MockDescribeVpcAttribute(ctx, input, opts)
}
