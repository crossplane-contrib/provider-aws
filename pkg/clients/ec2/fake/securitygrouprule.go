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
var _ clientset.SecurityGroupRuleClient = (*MockSecurityGroupRuleClient)(nil)

// MockSecurityGroupRuleClient is a type that implements all the methods for SecurityGroupClient interface
type MockSecurityGroupRuleClient struct {
	MockDescribe func(ctx context.Context, params *ec2.DescribeSecurityGroupRulesInput, optFns []func(*ec2.Options)) (*ec2.DescribeSecurityGroupRulesOutput, error)
	MockModify   func(ctx context.Context, params *ec2.ModifySecurityGroupRulesInput, optFns []func(*ec2.Options)) (*ec2.ModifySecurityGroupRulesOutput, error)

	MockAuthorizeIngress func(ctx context.Context, input *ec2.AuthorizeSecurityGroupIngressInput, opts []func(*ec2.Options)) (*ec2.AuthorizeSecurityGroupIngressOutput, error)
	MockAuthorizeEgress  func(ctx context.Context, input *ec2.AuthorizeSecurityGroupEgressInput, opts []func(*ec2.Options)) (*ec2.AuthorizeSecurityGroupEgressOutput, error)
	MockRevokeIngress    func(ctx context.Context, input *ec2.RevokeSecurityGroupIngressInput, opts []func(*ec2.Options)) (*ec2.RevokeSecurityGroupIngressOutput, error)
	MockRevokeEgress     func(ctx context.Context, input *ec2.RevokeSecurityGroupEgressInput, opts []func(*ec2.Options)) (*ec2.RevokeSecurityGroupEgressOutput, error)
}

// DeleteSecurityGroup mocks DeleteSecurityGroup method
func (m *MockSecurityGroupRuleClient) DescribeSecurityGroupRules(ctx context.Context, input *ec2.DescribeSecurityGroupRulesInput, opts ...func(*ec2.Options)) (*ec2.DescribeSecurityGroupRulesOutput, error) {
	return m.MockDescribe(ctx, input, opts)
}

// DescribeSecurityGroups mocks DescribeSecurityGroups method
func (m *MockSecurityGroupRuleClient) ModifySecurityGroupRules(ctx context.Context, input *ec2.ModifySecurityGroupRulesInput, opts ...func(*ec2.Options)) (*ec2.ModifySecurityGroupRulesOutput, error) {
	return m.MockModify(ctx, input, opts)
}

// AuthorizeSecurityGroupIngress mocks AuthorizeSecurityGroupIngress method
func (m *MockSecurityGroupRuleClient) AuthorizeSecurityGroupIngress(ctx context.Context, input *ec2.AuthorizeSecurityGroupIngressInput, opts ...func(*ec2.Options)) (*ec2.AuthorizeSecurityGroupIngressOutput, error) {
	return m.MockAuthorizeIngress(ctx, input, opts)
}

// AuthorizeSecurityGroupEgress mocks AuthorizeSecurityGroupEgress method
func (m *MockSecurityGroupRuleClient) AuthorizeSecurityGroupEgress(ctx context.Context, input *ec2.AuthorizeSecurityGroupEgressInput, opts ...func(*ec2.Options)) (*ec2.AuthorizeSecurityGroupEgressOutput, error) {
	return m.MockAuthorizeEgress(ctx, input, opts)
}

// RevokeSecurityGroupEgress mocks RevokeSecurityGroupEgress method
func (m *MockSecurityGroupRuleClient) RevokeSecurityGroupEgress(ctx context.Context, input *ec2.RevokeSecurityGroupEgressInput, opts ...func(*ec2.Options)) (*ec2.RevokeSecurityGroupEgressOutput, error) {
	return m.MockRevokeEgress(ctx, input, opts)
}

// RevokeSecurityGroupIngress mocks RevokeSecurityGroupIngress method
func (m *MockSecurityGroupRuleClient) RevokeSecurityGroupIngress(ctx context.Context, input *ec2.RevokeSecurityGroupIngressInput, opts ...func(*ec2.Options)) (*ec2.RevokeSecurityGroupIngressOutput, error) {
	return m.MockRevokeIngress(ctx, input, opts)
}
