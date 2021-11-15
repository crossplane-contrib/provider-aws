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
var _ clientset.VPCCIDRBlockClient = (*MockVPCCIDRBlockClient)(nil)

// MockVPCCIDRBlockClient is a type that implements all the methods for MockVPCCIDRBlockClient interface
type MockVPCCIDRBlockClient struct {
	MockDescribe     func(ctx context.Context, input *ec2.DescribeVpcsInput, opts []func(*ec2.Options)) (*ec2.DescribeVpcsOutput, error)
	MockAssociate    func(ctx context.Context, input *ec2.AssociateVpcCidrBlockInput, opts []func(*ec2.Options)) (*ec2.AssociateVpcCidrBlockOutput, error)
	MockDisassociate func(ctx context.Context, input *ec2.DisassociateVpcCidrBlockInput, opts []func(*ec2.Options)) (*ec2.DisassociateVpcCidrBlockOutput, error)
}

// DescribeVpcs mocks DescribeVpcs method
func (m *MockVPCCIDRBlockClient) DescribeVpcs(ctx context.Context, input *ec2.DescribeVpcsInput, opts ...func(*ec2.Options)) (*ec2.DescribeVpcsOutput, error) {
	return m.MockDescribe(ctx, input, opts)
}

// AssociateVpcCidrBlock mocks AssociateVpcCidrBlockmethod
func (m *MockVPCCIDRBlockClient) AssociateVpcCidrBlock(ctx context.Context, input *ec2.AssociateVpcCidrBlockInput, opts ...func(*ec2.Options)) (*ec2.AssociateVpcCidrBlockOutput, error) {
	return m.MockAssociate(ctx, input, opts)
}

// DisassociateVpcCidrBlock mocks DisassociateVpcCidrBlock method
func (m *MockVPCCIDRBlockClient) DisassociateVpcCidrBlock(ctx context.Context, input *ec2.DisassociateVpcCidrBlockInput, opts ...func(*ec2.Options)) (*ec2.DisassociateVpcCidrBlockOutput, error) {
	return m.MockDisassociate(ctx, input, opts)
}
