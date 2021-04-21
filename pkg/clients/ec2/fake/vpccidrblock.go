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
var _ clientset.VPCCIDRBlockClient = (*MockVPCCIDRBlockClient)(nil)

// MockVPCCIDRBlockClient is a type that implements all the methods for MockVPCCIDRBlockClient interface
type MockVPCCIDRBlockClient struct {
	MockDescribe     func(*ec2.DescribeVpcsInput) ec2.DescribeVpcsRequest
	MockAssociate    func(*ec2.AssociateVpcCidrBlockInput) ec2.AssociateVpcCidrBlockRequest
	MockDisassociate func(*ec2.DisassociateVpcCidrBlockInput) ec2.DisassociateVpcCidrBlockRequest
}

// DescribeVpcsRequest mocks DescribeVpcsRequest method
func (m *MockVPCCIDRBlockClient) DescribeVpcsRequest(input *ec2.DescribeVpcsInput) ec2.DescribeVpcsRequest {
	return m.MockDescribe(input)
}

// AssociateVpcCidrBlockRequest mocks AssociateVpcCidrBlockRequest method
func (m *MockVPCCIDRBlockClient) AssociateVpcCidrBlockRequest(input *ec2.AssociateVpcCidrBlockInput) ec2.AssociateVpcCidrBlockRequest {
	return m.MockAssociate(input)
}

// DisassociateVpcCidrBlockRequest mocks DisassociateVpcCidrBlockRequest method
func (m *MockVPCCIDRBlockClient) DisassociateVpcCidrBlockRequest(input *ec2.DisassociateVpcCidrBlockInput) ec2.DisassociateVpcCidrBlockRequest {
	return m.MockDisassociate(input)
}
