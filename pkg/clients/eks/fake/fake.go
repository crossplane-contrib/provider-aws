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

// https://github.com/golang/mock
// Automatically generate mock interfaces as part of "go generate"
// Ex. go:generate mockgen  --build_flags=--mod=mod -package eksiface -destination pkg/clients/eks/fake/eksiface/fake.go github.com/aws/aws-sdk-go/service/eks/eksiface EKSAPI

package fake

import (
	"context"

	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

// MockClient is a fake implementation of eks.Client.
type MockClient struct {
	MockCreateCluster        func(ctx context.Context, input *eks.CreateClusterInput, opts []func(*eks.Options)) (*eks.CreateClusterOutput, error)
	MockDescribeCluster      func(ctx context.Context, input *eks.DescribeClusterInput, opts []func(*eks.Options)) (*eks.DescribeClusterOutput, error)
	MockDeregisterCluster    func(ctx context.Context, input *eks.DeregisterClusterInput, opts []func(*eks.Options)) (*eks.DeregisterClusterOutput, error)
	MockRegisterCluster      func(ctx context.Context, input *eks.RegisterClusterInput, opts []func(*eks.Options)) (*eks.RegisterClusterOutput, error)
	MockUpdateClusterConfig  func(ctx context.Context, input *eks.UpdateClusterConfigInput, opts []func(*eks.Options)) (*eks.UpdateClusterConfigOutput, error)
	MockDeleteCluster        func(ctx context.Context, input *eks.DeleteClusterInput, opts []func(*eks.Options)) (*eks.DeleteClusterOutput, error)
	MockTagResource          func(ctx context.Context, input *eks.TagResourceInput, opts []func(*eks.Options)) (*eks.TagResourceOutput, error)
	MockUntagResource        func(ctx context.Context, input *eks.UntagResourceInput, opts []func(*eks.Options)) (*eks.UntagResourceOutput, error)
	MockUpdateClusterVersion func(ctx context.Context, input *eks.UpdateClusterVersionInput, opts []func(*eks.Options)) (*eks.UpdateClusterVersionOutput, error)

	MockDescribeNodegroup      func(ctx context.Context, input *eks.DescribeNodegroupInput, opts []func(*eks.Options)) (*eks.DescribeNodegroupOutput, error)
	MockCreateNodegroup        func(ctx context.Context, input *eks.CreateNodegroupInput, opts []func(*eks.Options)) (*eks.CreateNodegroupOutput, error)
	MockUpdateNodegroupVersion func(ctx context.Context, input *eks.UpdateNodegroupVersionInput, opts []func(*eks.Options)) (*eks.UpdateNodegroupVersionOutput, error)
	MockUpdateNodegroupConfig  func(ctx context.Context, input *eks.UpdateNodegroupConfigInput, opts []func(*eks.Options)) (*eks.UpdateNodegroupConfigOutput, error)
	MockDeleteNodegroup        func(ctx context.Context, input *eks.DeleteNodegroupInput, opts []func(*eks.Options)) (*eks.DeleteNodegroupOutput, error)

	MockDescribeFargateProfile func(ctx context.Context, input *eks.DescribeFargateProfileInput, opts []func(*eks.Options)) (*eks.DescribeFargateProfileOutput, error)
	MockCreateFargateProfile   func(ctx context.Context, input *eks.CreateFargateProfileInput, opts []func(*eks.Options)) (*eks.CreateFargateProfileOutput, error)
	MockDeleteFargateProfile   func(ctx context.Context, input *eks.DeleteFargateProfileInput, opts []func(*eks.Options)) (*eks.DeleteFargateProfileOutput, error)

	MockDescribeIdentityProviderConfig     func(ctx context.Context, input *eks.DescribeIdentityProviderConfigInput, opts []func(*eks.Options)) (*eks.DescribeIdentityProviderConfigOutput, error)
	MockAssociateIdentityProviderConfig    func(ctx context.Context, input *eks.AssociateIdentityProviderConfigInput, opts []func(*eks.Options)) (*eks.AssociateIdentityProviderConfigOutput, error)
	MockDisassociateIdentityProviderConfig func(ctx context.Context, input *eks.DisassociateIdentityProviderConfigInput, opts []func(*eks.Options)) (*eks.DisassociateIdentityProviderConfigOutput, error)
	MockListIdentityProviderConfigs        func(ctx context.Context, input *eks.ListIdentityProviderConfigsInput, opts []func(*eks.Options)) (*eks.ListIdentityProviderConfigsOutput, error)

	MockAssociateEncryptionConfig func(ctx context.Context, input *eks.AssociateEncryptionConfigInput, opts []func(*eks.Options)) (*eks.AssociateEncryptionConfigOutput, error)
}

// MockSTSClient mock sts client
type MockSTSClient struct {
	MockPresignGetCallerIdentity func(ctx context.Context, input *sts.GetCallerIdentityInput, opts []func(*sts.PresignOptions)) (*v4.PresignedHTTPRequest, error)
}

// PresignGetCallerIdentity calls the underlying MockPresignGetCallerIdentity method.
func (c *MockSTSClient) PresignGetCallerIdentity(ctx context.Context, input *sts.GetCallerIdentityInput, opts ...func(*sts.PresignOptions)) (*v4.PresignedHTTPRequest, error) {
	return c.MockPresignGetCallerIdentity(ctx, input, opts)
}

// CreateCluster calls the underlying MockCreateCluster method.
func (c *MockClient) CreateCluster(ctx context.Context, input *eks.CreateClusterInput, opts ...func(*eks.Options)) (*eks.CreateClusterOutput, error) {
	return c.MockCreateCluster(ctx, input, opts)
}

// DescribeCluster calls the underlying MockDescribeCluster
// method.
func (c *MockClient) DescribeCluster(ctx context.Context, input *eks.DescribeClusterInput, opts ...func(*eks.Options)) (*eks.DescribeClusterOutput, error) {
	return c.MockDescribeCluster(ctx, input, opts)
}

// UpdateClusterConfig calls the underlying
// MockUpdateClusterConfig method.
func (c *MockClient) UpdateClusterConfig(ctx context.Context, input *eks.UpdateClusterConfigInput, opts ...func(*eks.Options)) (*eks.UpdateClusterConfigOutput, error) {
	return c.MockUpdateClusterConfig(ctx, input, opts)
}

// DeleteCluster calls the underlying MockDeleteCluster method.
func (c *MockClient) DeleteCluster(ctx context.Context, input *eks.DeleteClusterInput, opts ...func(*eks.Options)) (*eks.DeleteClusterOutput, error) {
	return c.MockDeleteCluster(ctx, input, opts)
}

// DeregisterCluster calls the underlying MockDeregisterCluster method.
func (c *MockClient) DeregisterCluster(ctx context.Context, input *eks.DeregisterClusterInput, opts ...func(*eks.Options)) (*eks.DeregisterClusterOutput, error) {
	return c.MockDeregisterCluster(ctx, input, opts)
}

// RegisterCluster calls the underlying MockRegisterCluster method.
func (c *MockClient) RegisterCluster(ctx context.Context, input *eks.RegisterClusterInput, opts ...func(*eks.Options)) (*eks.RegisterClusterOutput, error) {
	return c.MockRegisterCluster(ctx, input, opts)
}

// TagResource calls the underlying MockTagResource method.
func (c *MockClient) TagResource(ctx context.Context, input *eks.TagResourceInput, opts ...func(*eks.Options)) (*eks.TagResourceOutput, error) {
	return c.MockTagResource(ctx, input, opts)
}

// UntagResource calls the underlying MockUntagResource method.
func (c *MockClient) UntagResource(ctx context.Context, input *eks.UntagResourceInput, opts ...func(*eks.Options)) (*eks.UntagResourceOutput, error) {
	return c.MockUntagResource(ctx, input, opts)
}

// UpdateClusterVersion calls the underlying
// MockUpdateClusterVersion method.
func (c *MockClient) UpdateClusterVersion(ctx context.Context, input *eks.UpdateClusterVersionInput, opts ...func(*eks.Options)) (*eks.UpdateClusterVersionOutput, error) {
	return c.MockUpdateClusterVersion(ctx, input, opts)
}

// DescribeNodegroup calls the underlying MockDescribeNodegroup
// method.
func (c *MockClient) DescribeNodegroup(ctx context.Context, input *eks.DescribeNodegroupInput, opts ...func(*eks.Options)) (*eks.DescribeNodegroupOutput, error) {
	return c.MockDescribeNodegroup(ctx, input, opts)
}

// CreateNodegroup calls the underlying MockCreateNodegroup
// method.
func (c *MockClient) CreateNodegroup(ctx context.Context, input *eks.CreateNodegroupInput, opts ...func(*eks.Options)) (*eks.CreateNodegroupOutput, error) {
	return c.MockCreateNodegroup(ctx, input, opts)
}

// UpdateNodegroupVersion calls the underlying
// MockUpdateNodegroupVersion method.
func (c *MockClient) UpdateNodegroupVersion(ctx context.Context, input *eks.UpdateNodegroupVersionInput, opts ...func(*eks.Options)) (*eks.UpdateNodegroupVersionOutput, error) {
	return c.MockUpdateNodegroupVersion(ctx, input, opts)
}

// UpdateNodegroupConfig calls the underlying
// MockUpdateNodegroupConfig method.
func (c *MockClient) UpdateNodegroupConfig(ctx context.Context, input *eks.UpdateNodegroupConfigInput, opts ...func(*eks.Options)) (*eks.UpdateNodegroupConfigOutput, error) {
	return c.MockUpdateNodegroupConfig(ctx, input, opts)
}

// DeleteNodegroup calls the underlying MockDeleteNodegroup
// method.
func (c *MockClient) DeleteNodegroup(ctx context.Context, input *eks.DeleteNodegroupInput, opts ...func(*eks.Options)) (*eks.DeleteNodegroupOutput, error) {
	return c.MockDeleteNodegroup(ctx, input, opts)
}

// DescribeFargateProfile calls the underlying MockDescribeFargateProfile
// method.
func (c *MockClient) DescribeFargateProfile(ctx context.Context, input *eks.DescribeFargateProfileInput, opts ...func(*eks.Options)) (*eks.DescribeFargateProfileOutput, error) {
	return c.MockDescribeFargateProfile(ctx, input, opts)
}

// CreateFargateProfile calls the underlying MockCreateFargateProfile
// method.
func (c *MockClient) CreateFargateProfile(ctx context.Context, input *eks.CreateFargateProfileInput, opts ...func(*eks.Options)) (*eks.CreateFargateProfileOutput, error) {
	return c.MockCreateFargateProfile(ctx, input, opts)
}

// DeleteFargateProfile calls the underlying MockDeleteFargateProfile
// method.
func (c *MockClient) DeleteFargateProfile(ctx context.Context, input *eks.DeleteFargateProfileInput, opts ...func(*eks.Options)) (*eks.DeleteFargateProfileOutput, error) {
	return c.MockDeleteFargateProfile(ctx, input, opts)
}

// DescribeIdentityProviderConfig calls the underlying MockDescribeIdentityProviderConfig
// method
func (c *MockClient) DescribeIdentityProviderConfig(ctx context.Context, input *eks.DescribeIdentityProviderConfigInput, opts ...func(*eks.Options)) (*eks.DescribeIdentityProviderConfigOutput, error) {
	return c.MockDescribeIdentityProviderConfig(ctx, input, opts)
}

// AssociateIdentityProviderConfig calls the underlying MockAssociateIdentityProviderConfig
// method
func (c *MockClient) AssociateIdentityProviderConfig(ctx context.Context, input *eks.AssociateIdentityProviderConfigInput, opts ...func(*eks.Options)) (*eks.AssociateIdentityProviderConfigOutput, error) {
	return c.MockAssociateIdentityProviderConfig(ctx, input, opts)
}

// DisassociateIdentityProviderConfig calls the underlying MockDisassociateIdentityProviderConfig
// method
func (c *MockClient) DisassociateIdentityProviderConfig(ctx context.Context, input *eks.DisassociateIdentityProviderConfigInput, opts ...func(*eks.Options)) (*eks.DisassociateIdentityProviderConfigOutput, error) {
	return c.MockDisassociateIdentityProviderConfig(ctx, input, opts)
}

// ListIdentityProviderConfigs calls the underlying ListIdentityProviderConfigs
// method
func (c *MockClient) ListIdentityProviderConfigs(ctx context.Context, input *eks.ListIdentityProviderConfigsInput, opts ...func(*eks.Options)) (*eks.ListIdentityProviderConfigsOutput, error) {
	return c.MockListIdentityProviderConfigs(ctx, input, opts)
}

// AssociateEncryptionConfig calls the underlying MockAssociateEncryptionConfig
// method
func (c *MockClient) AssociateEncryptionConfig(ctx context.Context, input *eks.AssociateEncryptionConfigInput, opts ...func(*eks.Options)) (*eks.AssociateEncryptionConfigOutput, error) {
	return c.MockAssociateEncryptionConfig(ctx, input, opts)
}
