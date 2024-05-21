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

	clientset "github.com/crossplane-contrib/provider-aws/pkg/clients/ec2"
)

// this ensures that the mock implements the client interface
var _ clientset.SecurityGroupClient = (*MockSecurityGroupClient)(nil)

// MockSecurityGroupClient is a type that implements all the methods for SecurityGroupClient interface
type MockSecurityGroupClient struct {
	MockCreate           func(ctx context.Context, input *ec2.CreateSecurityGroupInput, opts []func(*ec2.Options)) (*ec2.CreateSecurityGroupOutput, error)
	MockDelete           func(ctx context.Context, input *ec2.DeleteSecurityGroupInput, opts []func(*ec2.Options)) (*ec2.DeleteSecurityGroupOutput, error)
	MockDescribe         func(ctx context.Context, input *ec2.DescribeSecurityGroupsInput, opts []func(*ec2.Options)) (*ec2.DescribeSecurityGroupsOutput, error)
	MockDescribeRules    func(ctx context.Context, input *ec2.DescribeSecurityGroupRulesInput, opts []func(*ec2.Options)) (*ec2.DescribeSecurityGroupRulesOutput, error)
	MockAuthorizeIngress func(ctx context.Context, input *ec2.AuthorizeSecurityGroupIngressInput, opts []func(*ec2.Options)) (*ec2.AuthorizeSecurityGroupIngressOutput, error)
	MockAuthorizeEgress  func(ctx context.Context, input *ec2.AuthorizeSecurityGroupEgressInput, opts []func(*ec2.Options)) (*ec2.AuthorizeSecurityGroupEgressOutput, error)
	MockRevokeIngress    func(ctx context.Context, input *ec2.RevokeSecurityGroupIngressInput, opts []func(*ec2.Options)) (*ec2.RevokeSecurityGroupIngressOutput, error)
	MockRevokeEgress     func(ctx context.Context, input *ec2.RevokeSecurityGroupEgressInput, opts []func(*ec2.Options)) (*ec2.RevokeSecurityGroupEgressOutput, error)
	MockCreateTags       func(ctx context.Context, input *ec2.CreateTagsInput, opts []func(*ec2.Options)) (*ec2.CreateTagsOutput, error)
	MockDeleteTags       func(ctx context.Context, input *ec2.DeleteTagsInput, opts []func(*ec2.Options)) (*ec2.DeleteTagsOutput, error)
}

// CreateSecurityGroup mocks CreateSecurityGroup method
func (m *MockSecurityGroupClient) CreateSecurityGroup(ctx context.Context, input *ec2.CreateSecurityGroupInput, opts ...func(*ec2.Options)) (*ec2.CreateSecurityGroupOutput, error) {
	return m.MockCreate(ctx, input, opts)
}

// DeleteSecurityGroup mocks DeleteSecurityGroup method
func (m *MockSecurityGroupClient) DeleteSecurityGroup(ctx context.Context, input *ec2.DeleteSecurityGroupInput, opts ...func(*ec2.Options)) (*ec2.DeleteSecurityGroupOutput, error) {
	return m.MockDelete(ctx, input, opts)
}

// DescribeSecurityGroups mocks DescribeSecurityGroups method
func (m *MockSecurityGroupClient) DescribeSecurityGroups(ctx context.Context, input *ec2.DescribeSecurityGroupsInput, opts ...func(*ec2.Options)) (*ec2.DescribeSecurityGroupsOutput, error) {
	return m.MockDescribe(ctx, input, opts)
}

// DescribeSecurityGroups mocks DescribeSecurityGroups method
func (m *MockSecurityGroupClient) DescribeSecurityGroupRules(ctx context.Context, input *ec2.DescribeSecurityGroupRulesInput, opts ...func(*ec2.Options)) (*ec2.DescribeSecurityGroupRulesOutput, error) {
	return m.MockDescribeRules(ctx, input, opts)
}

// AuthorizeSecurityGroupIngress mocks AuthorizeSecurityGroupIngress method
func (m *MockSecurityGroupClient) AuthorizeSecurityGroupIngress(ctx context.Context, input *ec2.AuthorizeSecurityGroupIngressInput, opts ...func(*ec2.Options)) (*ec2.AuthorizeSecurityGroupIngressOutput, error) {
	return m.MockAuthorizeIngress(ctx, input, opts)
}

// AuthorizeSecurityGroupEgress mocks AuthorizeSecurityGroupEgress method
func (m *MockSecurityGroupClient) AuthorizeSecurityGroupEgress(ctx context.Context, input *ec2.AuthorizeSecurityGroupEgressInput, opts ...func(*ec2.Options)) (*ec2.AuthorizeSecurityGroupEgressOutput, error) {
	return m.MockAuthorizeEgress(ctx, input, opts)
}

// RevokeSecurityGroupEgress mocks RevokeSecurityGroupEgress method
func (m *MockSecurityGroupClient) RevokeSecurityGroupEgress(ctx context.Context, input *ec2.RevokeSecurityGroupEgressInput, opts ...func(*ec2.Options)) (*ec2.RevokeSecurityGroupEgressOutput, error) {
	return m.MockRevokeEgress(ctx, input, opts)
}

// RevokeSecurityGroupIngress mocks RevokeSecurityGroupIngress method
func (m *MockSecurityGroupClient) RevokeSecurityGroupIngress(ctx context.Context, input *ec2.RevokeSecurityGroupIngressInput, opts ...func(*ec2.Options)) (*ec2.RevokeSecurityGroupIngressOutput, error) {
	return m.MockRevokeIngress(ctx, input, opts)
}

// CreateTags mocks CreateTagsInput method
func (m *MockSecurityGroupClient) CreateTags(ctx context.Context, input *ec2.CreateTagsInput, opts ...func(*ec2.Options)) (*ec2.CreateTagsOutput, error) {
	return m.MockCreateTags(ctx, input, opts)
}

// DeleteTags mocks DeleteTagsInput method
func (m *MockSecurityGroupClient) DeleteTags(ctx context.Context, input *ec2.DeleteTagsInput, opts ...func(*ec2.Options)) (*ec2.DeleteTagsOutput, error) {
	return m.MockDeleteTags(ctx, input, opts)
}
