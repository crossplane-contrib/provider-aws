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
var _ clientset.InternetGatewayClient = (*MockInternetGatewayClient)(nil)

// MockInternetGatewayClient is a type that implements all the methods for InternetGatewayClient interface
type MockInternetGatewayClient struct {
	MockCreate     func(*ec2.CreateInternetGatewayInput) ec2.CreateInternetGatewayRequest
	MockDelete     func(*ec2.DeleteInternetGatewayInput) ec2.DeleteInternetGatewayRequest
	MockDescribe   func(*ec2.DescribeInternetGatewaysInput) ec2.DescribeInternetGatewaysRequest
	MockAttach     func(*ec2.AttachInternetGatewayInput) ec2.AttachInternetGatewayRequest
	MockDetach     func(*ec2.DetachInternetGatewayInput) ec2.DetachInternetGatewayRequest
	MockCreateTags func(*ec2.CreateTagsInput) ec2.CreateTagsRequest
}

// CreateInternetGatewayRequest mocks CreateInternetGatewayRequest method
func (m *MockInternetGatewayClient) CreateInternetGatewayRequest(input *ec2.CreateInternetGatewayInput) ec2.CreateInternetGatewayRequest {
	return m.MockCreate(input)
}

// DeleteInternetGatewayRequest mocks DeleteInternetGatewayRequest method
func (m *MockInternetGatewayClient) DeleteInternetGatewayRequest(input *ec2.DeleteInternetGatewayInput) ec2.DeleteInternetGatewayRequest {
	return m.MockDelete(input)
}

// DescribeInternetGatewaysRequest mocks DescribeInternetGatewaysRequest method
func (m *MockInternetGatewayClient) DescribeInternetGatewaysRequest(input *ec2.DescribeInternetGatewaysInput) ec2.DescribeInternetGatewaysRequest {
	return m.MockDescribe(input)
}

// AttachInternetGatewayRequest mocks AttachInternetGatewayRequest method
func (m *MockInternetGatewayClient) AttachInternetGatewayRequest(input *ec2.AttachInternetGatewayInput) ec2.AttachInternetGatewayRequest {
	return m.MockAttach(input)
}

// DetachInternetGatewayRequest mocks DetachInternetGatewayRequest
func (m *MockInternetGatewayClient) DetachInternetGatewayRequest(input *ec2.DetachInternetGatewayInput) ec2.DetachInternetGatewayRequest {
	return m.MockDetach(input)
}

// CreateTagsRequest mocks CreateTagsInput method
func (m *MockInternetGatewayClient) CreateTagsRequest(input *ec2.CreateTagsInput) ec2.CreateTagsRequest {
	return m.MockCreateTags(input)
}
