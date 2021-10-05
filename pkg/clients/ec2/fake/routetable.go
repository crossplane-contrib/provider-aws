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
var _ clientset.RouteTableClient = (*MockRouteTableClient)(nil)

// MockRouteTableClient is a type that implements all the methods for RouteTableClient interface
type MockRouteTableClient struct {
	MockCreate       func(ctx context.Context, input *ec2.CreateRouteTableInput, opts []func(*ec2.Options)) (*ec2.CreateRouteTableOutput, error)
	MockDelete       func(ctx context.Context, input *ec2.DeleteRouteTableInput, opts []func(*ec2.Options)) (*ec2.DeleteRouteTableOutput, error)
	MockDescribe     func(ctx context.Context, input *ec2.DescribeRouteTablesInput, opts []func(*ec2.Options)) (*ec2.DescribeRouteTablesOutput, error)
	MockCreateRoute  func(ctx context.Context, input *ec2.CreateRouteInput, opts []func(*ec2.Options)) (*ec2.CreateRouteOutput, error)
	MockDeleteRoute  func(ctx context.Context, input *ec2.DeleteRouteInput, opts []func(*ec2.Options)) (*ec2.DeleteRouteOutput, error)
	MockAssociate    func(ctx context.Context, input *ec2.AssociateRouteTableInput, opts []func(*ec2.Options)) (*ec2.AssociateRouteTableOutput, error)
	MockDisassociate func(ctx context.Context, input *ec2.DisassociateRouteTableInput, opts []func(*ec2.Options)) (*ec2.DisassociateRouteTableOutput, error)
	MockCreateTags   func(ctx context.Context, input *ec2.CreateTagsInput, opts []func(*ec2.Options)) (*ec2.CreateTagsOutput, error)
	MockDeleteTags   func(ctx context.Context, input *ec2.DeleteTagsInput, opts []func(*ec2.Options)) (*ec2.DeleteTagsOutput, error)
}

// CreateRouteTable mocks CreateRouteTable method
func (m *MockRouteTableClient) CreateRouteTable(ctx context.Context, input *ec2.CreateRouteTableInput, opts ...func(*ec2.Options)) (*ec2.CreateRouteTableOutput, error) {
	return m.MockCreate(ctx, input, opts)
}

// DeleteRouteTable mocks DeleteRouteTable method
func (m *MockRouteTableClient) DeleteRouteTable(ctx context.Context, input *ec2.DeleteRouteTableInput, opts ...func(*ec2.Options)) (*ec2.DeleteRouteTableOutput, error) {
	return m.MockDelete(ctx, input, opts)
}

// DescribeRouteTables mocks DescribeRouteTables method
func (m *MockRouteTableClient) DescribeRouteTables(ctx context.Context, input *ec2.DescribeRouteTablesInput, opts ...func(*ec2.Options)) (*ec2.DescribeRouteTablesOutput, error) {
	return m.MockDescribe(ctx, input, opts)
}

// AssociateRouteTable mocks AssociateRouteTable method
func (m *MockRouteTableClient) AssociateRouteTable(ctx context.Context, input *ec2.AssociateRouteTableInput, opts ...func(*ec2.Options)) (*ec2.AssociateRouteTableOutput, error) {
	return m.MockAssociate(ctx, input, opts)
}

// DisassociateRouteTable mocks DisassociateRouteTable method
func (m *MockRouteTableClient) DisassociateRouteTable(ctx context.Context, input *ec2.DisassociateRouteTableInput, opts ...func(*ec2.Options)) (*ec2.DisassociateRouteTableOutput, error) {
	return m.MockDisassociate(ctx, input, opts)
}

// CreateRoute mocks CreateRoute method
func (m *MockRouteTableClient) CreateRoute(ctx context.Context, input *ec2.CreateRouteInput, opts ...func(*ec2.Options)) (*ec2.CreateRouteOutput, error) {
	return m.MockCreateRoute(ctx, input, opts)
}

// DeleteRoute mocks DeleteRoute method
func (m *MockRouteTableClient) DeleteRoute(ctx context.Context, input *ec2.DeleteRouteInput, opts ...func(*ec2.Options)) (*ec2.DeleteRouteOutput, error) {
	return m.MockDeleteRoute(ctx, input, opts)
}

// CreateTags mocks CreateTags method
func (m *MockRouteTableClient) CreateTags(ctx context.Context, input *ec2.CreateTagsInput, opts ...func(*ec2.Options)) (*ec2.CreateTagsOutput, error) {
	return m.MockCreateTags(ctx, input, opts)
}

// DeleteTags mocks DeleteTags method
func (m *MockRouteTableClient) DeleteTags(ctx context.Context, input *ec2.DeleteTagsInput, opts ...func(*ec2.Options)) (*ec2.DeleteTagsOutput, error) {
	return m.MockDeleteTags(ctx, input, opts)
}
