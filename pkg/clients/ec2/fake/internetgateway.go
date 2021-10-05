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
var _ clientset.InternetGatewayClient = (*MockInternetGatewayClient)(nil)

// MockInternetGatewayClient is a type that implements all the methods for InternetGatewayClient interface
type MockInternetGatewayClient struct {
	MockCreate     func(ctx context.Context, input *ec2.CreateInternetGatewayInput, opts []func(*ec2.Options)) (*ec2.CreateInternetGatewayOutput, error)
	MockDelete     func(ctx context.Context, input *ec2.DeleteInternetGatewayInput, opts []func(*ec2.Options)) (*ec2.DeleteInternetGatewayOutput, error)
	MockDescribe   func(ctx context.Context, input *ec2.DescribeInternetGatewaysInput, opts []func(*ec2.Options)) (*ec2.DescribeInternetGatewaysOutput, error)
	MockAttach     func(ctx context.Context, input *ec2.AttachInternetGatewayInput, opts []func(*ec2.Options)) (*ec2.AttachInternetGatewayOutput, error)
	MockDetach     func(ctx context.Context, input *ec2.DetachInternetGatewayInput, opts []func(*ec2.Options)) (*ec2.DetachInternetGatewayOutput, error)
	MockCreateTags func(ctx context.Context, input *ec2.CreateTagsInput, opts []func(*ec2.Options)) (*ec2.CreateTagsOutput, error)
}

// CreateInternetGateway mocks CreateInternetGateway method
func (m *MockInternetGatewayClient) CreateInternetGateway(ctx context.Context, input *ec2.CreateInternetGatewayInput, opts ...func(*ec2.Options)) (*ec2.CreateInternetGatewayOutput, error) {
	return m.MockCreate(ctx, input, opts)
}

// DeleteInternetGateway mocks DeleteInternetGateway method
func (m *MockInternetGatewayClient) DeleteInternetGateway(ctx context.Context, input *ec2.DeleteInternetGatewayInput, opts ...func(*ec2.Options)) (*ec2.DeleteInternetGatewayOutput, error) {
	return m.MockDelete(ctx, input, opts)
}

// DescribeInternetGateways mocks DescribeInternetGateways method
func (m *MockInternetGatewayClient) DescribeInternetGateways(ctx context.Context, input *ec2.DescribeInternetGatewaysInput, opts ...func(*ec2.Options)) (*ec2.DescribeInternetGatewaysOutput, error) {
	return m.MockDescribe(ctx, input, opts)
}

// AttachInternetGateway mocks AttachInternetGateway method
func (m *MockInternetGatewayClient) AttachInternetGateway(ctx context.Context, input *ec2.AttachInternetGatewayInput, opts ...func(*ec2.Options)) (*ec2.AttachInternetGatewayOutput, error) {
	return m.MockAttach(ctx, input, opts)
}

// DetachInternetGateway mocks DetachInternetGateway
func (m *MockInternetGatewayClient) DetachInternetGateway(ctx context.Context, input *ec2.DetachInternetGatewayInput, opts ...func(*ec2.Options)) (*ec2.DetachInternetGatewayOutput, error) {
	return m.MockDetach(ctx, input, opts)
}

// CreateTags mocks CreateTags method
func (m *MockInternetGatewayClient) CreateTags(ctx context.Context, input *ec2.CreateTagsInput, opts ...func(*ec2.Options)) (*ec2.CreateTagsOutput, error) {
	return m.MockCreateTags(ctx, input, opts)
}
