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
	"github.com/aws/aws-sdk-go-v2/service/ec2"

	clientset "github.com/crossplane/provider-aws/pkg/clients/ec2"
)

// this ensures that the mock implements the client interface
var _ clientset.RouteTableClient = (*MockRouteTableClient)(nil)

// MockRouteTableClient is a type that implements all the methods for RouteTableClient interface
type MockRouteTableClient struct {
	MockCreate       func(*ec2.CreateRouteTableInput) ec2.CreateRouteTableRequest
	MockDelete       func(*ec2.DeleteRouteTableInput) ec2.DeleteRouteTableRequest
	MockDescribe     func(*ec2.DescribeRouteTablesInput) ec2.DescribeRouteTablesRequest
	MockCreateRoute  func(*ec2.CreateRouteInput) ec2.CreateRouteRequest
	MockDeleteRoute  func(*ec2.DeleteRouteInput) ec2.DeleteRouteRequest
	MockAssociate    func(*ec2.AssociateRouteTableInput) ec2.AssociateRouteTableRequest
	MockDisassociate func(*ec2.DisassociateRouteTableInput) ec2.DisassociateRouteTableRequest
	MockCreateTags   func(*ec2.CreateTagsInput) ec2.CreateTagsRequest
	MockDeleteTags   func(*ec2.DeleteTagsInput) ec2.DeleteTagsRequest
}

// CreateRouteTableRequest mocks CreateRouteTableRequest method
func (m *MockRouteTableClient) CreateRouteTableRequest(input *ec2.CreateRouteTableInput) ec2.CreateRouteTableRequest {
	return m.MockCreate(input)
}

// DeleteRouteTableRequest mocks DeleteRouteTableRequest method
func (m *MockRouteTableClient) DeleteRouteTableRequest(input *ec2.DeleteRouteTableInput) ec2.DeleteRouteTableRequest {
	return m.MockDelete(input)
}

// DescribeRouteTablesRequest mocks DescribeRouteTablesRequest method
func (m *MockRouteTableClient) DescribeRouteTablesRequest(input *ec2.DescribeRouteTablesInput) ec2.DescribeRouteTablesRequest {
	return m.MockDescribe(input)
}

// AssociateRouteTableRequest mocks AssociateRouteTableRequest method
func (m *MockRouteTableClient) AssociateRouteTableRequest(input *ec2.AssociateRouteTableInput) ec2.AssociateRouteTableRequest {
	return m.MockAssociate(input)
}

// DisassociateRouteTableRequest mocks DisassociateRouteTableRequest method
func (m *MockRouteTableClient) DisassociateRouteTableRequest(input *ec2.DisassociateRouteTableInput) ec2.DisassociateRouteTableRequest {
	return m.MockDisassociate(input)
}

// CreateRouteRequest mocks CreateRouteRequest method
func (m *MockRouteTableClient) CreateRouteRequest(input *ec2.CreateRouteInput) ec2.CreateRouteRequest {
	return m.MockCreateRoute(input)
}

// DeleteRouteRequest mocks DeleteRouteRequest method
func (m *MockRouteTableClient) DeleteRouteRequest(input *ec2.DeleteRouteInput) ec2.DeleteRouteRequest {
	return m.MockDeleteRoute(input)
}

// CreateTagsRequest mocks CreateTagsInput method
func (m *MockRouteTableClient) CreateTagsRequest(input *ec2.CreateTagsInput) ec2.CreateTagsRequest {
	return m.MockCreateTags(input)
}

// DeleteTagsRequest mocks DeleteTagsInput method
func (m *MockRouteTableClient) DeleteTagsRequest(input *ec2.DeleteTagsInput) ec2.DeleteTagsRequest {
	return m.MockDeleteTags(input)
}
