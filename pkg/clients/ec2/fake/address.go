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

package fake

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ec2"

	clientset "github.com/crossplane/provider-aws/pkg/clients/ec2"
)

// this ensures that the mock implements the client interface
var _ clientset.AddressClient = (*MockAddressClient)(nil)

// MockAddressClient is a type that implements all the methods for ElasticIPClient interface
type MockAddressClient struct {
	MockAllocate   func(ctx context.Context, input *ec2.AllocateAddressInput, opts []func(*ec2.Options)) (*ec2.AllocateAddressOutput, error)
	MockRelease    func(ctx context.Context, input *ec2.ReleaseAddressInput, opts []func(*ec2.Options)) (*ec2.ReleaseAddressOutput, error)
	MockDescribe   func(ctx context.Context, input *ec2.DescribeAddressesInput, opts []func(*ec2.Options)) (*ec2.DescribeAddressesOutput, error)
	MockCreateTags func(ctx context.Context, input *ec2.CreateTagsInput, opts []func(*ec2.Options)) (*ec2.CreateTagsOutput, error)
}

// AllocateAddress mocks AllocateAddress method
func (m *MockAddressClient) AllocateAddress(ctx context.Context, input *ec2.AllocateAddressInput, opts ...func(*ec2.Options)) (*ec2.AllocateAddressOutput, error) {
	return m.MockAllocate(ctx, input, opts)
}

// ReleaseAddress mocks ReleaseAddress method
func (m *MockAddressClient) ReleaseAddress(ctx context.Context, input *ec2.ReleaseAddressInput, opts ...func(*ec2.Options)) (*ec2.ReleaseAddressOutput, error) {
	return m.MockRelease(ctx, input, opts)
}

// DescribeAddresses mocks DescribeAddresses method
func (m *MockAddressClient) DescribeAddresses(ctx context.Context, input *ec2.DescribeAddressesInput, opts ...func(*ec2.Options)) (*ec2.DescribeAddressesOutput, error) {
	return m.MockDescribe(ctx, input, opts)
}

// CreateTags mocks CreateTags method
func (m *MockAddressClient) CreateTags(ctx context.Context, input *ec2.CreateTagsInput, opts ...func(*ec2.Options)) (*ec2.CreateTagsOutput, error) {
	return m.MockCreateTags(ctx, input, opts)
}
