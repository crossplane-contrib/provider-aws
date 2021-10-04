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
var _ clientset.SubnetClient = (*MockSubnetClient)(nil)

// MockSubnetClient is a type that implements all the methods for SubnetClient interface
type MockSubnetClient struct {
	MockCreate     func(ctx context.Context, input *ec2.CreateSubnetInput, opts []func(*ec2.Options)) (*ec2.CreateSubnetOutput, error)
	MockDelete     func(ctx context.Context, input *ec2.DeleteSubnetInput, opts []func(*ec2.Options)) (*ec2.DeleteSubnetOutput, error)
	MockDescribe   func(ctx context.Context, input *ec2.DescribeSubnetsInput, opts []func(*ec2.Options)) (*ec2.DescribeSubnetsOutput, error)
	MockModify     func(ctx context.Context, input *ec2.ModifySubnetAttributeInput, opts []func(*ec2.Options)) (*ec2.ModifySubnetAttributeOutput, error)
	MockCreateTags func(ctx context.Context, input *ec2.CreateTagsInput, opts []func(*ec2.Options)) (*ec2.CreateTagsOutput, error)
}

// CreateSubnet mocks CreateSubnet method
func (m *MockSubnetClient) CreateSubnet(ctx context.Context, input *ec2.CreateSubnetInput, opts ...func(*ec2.Options)) (*ec2.CreateSubnetOutput, error) {
	return m.MockCreate(ctx, input, opts)
}

// DeleteSubnet mocks DeleteSubnet method
func (m *MockSubnetClient) DeleteSubnet(ctx context.Context, input *ec2.DeleteSubnetInput, opts ...func(*ec2.Options)) (*ec2.DeleteSubnetOutput, error) {
	return m.MockDelete(ctx, input, opts)
}

// DescribeSubnets mocks DescribeSubnets method
func (m *MockSubnetClient) DescribeSubnets(ctx context.Context, input *ec2.DescribeSubnetsInput, opts ...func(*ec2.Options)) (*ec2.DescribeSubnetsOutput, error) {
	return m.MockDescribe(ctx, input, opts)
}

// ModifySubnetAttribute mocks ModifySubnetAttributeInput method
func (m *MockSubnetClient) ModifySubnetAttribute(ctx context.Context, input *ec2.ModifySubnetAttributeInput, opts ...func(*ec2.Options)) (*ec2.ModifySubnetAttributeOutput, error) {
	return m.MockModify(ctx, input, opts)
}

// CreateTags mocks CreateTags method
func (m *MockSubnetClient) CreateTags(ctx context.Context, input *ec2.CreateTagsInput, opts ...func(*ec2.Options)) (*ec2.CreateTagsOutput, error) {
	return m.MockCreateTags(ctx, input, opts)
}
