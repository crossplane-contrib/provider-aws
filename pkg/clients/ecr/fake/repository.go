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

	"github.com/aws/aws-sdk-go-v2/service/ecr"

	clientset "github.com/crossplane/provider-aws/pkg/clients/ecr"
)

// this ensures that the mock implements the client interface
var _ clientset.RepositoryClient = (*MockRepositoryClient)(nil)

// MockRepositoryClient is a type that implements all the methods for ECRClient interface
type MockRepositoryClient struct {
	MockCreate                func(ctx context.Context, input *ecr.CreateRepositoryInput, opts []func(*ecr.Options)) (*ecr.CreateRepositoryOutput, error)
	MockDelete                func(ctx context.Context, input *ecr.DeleteRepositoryInput, opts []func(*ecr.Options)) (*ecr.DeleteRepositoryOutput, error)
	MockDescribe              func(ctx context.Context, input *ecr.DescribeRepositoriesInput, opts []func(*ecr.Options)) (*ecr.DescribeRepositoriesOutput, error)
	MockListTags              func(ctx context.Context, input *ecr.ListTagsForResourceInput, opts []func(*ecr.Options)) (*ecr.ListTagsForResourceOutput, error)
	MockTag                   func(ctx context.Context, input *ecr.TagResourceInput, opts []func(*ecr.Options)) (*ecr.TagResourceOutput, error)
	MockUntag                 func(ctx context.Context, input *ecr.UntagResourceInput, opts []func(*ecr.Options)) (*ecr.UntagResourceOutput, error)
	MockPutImageScan          func(ctx context.Context, input *ecr.PutImageScanningConfigurationInput, opts []func(*ecr.Options)) (*ecr.PutImageScanningConfigurationOutput, error)
	MockPutImageTagMutability func(ctx context.Context, input *ecr.PutImageTagMutabilityInput, opts []func(*ecr.Options)) (*ecr.PutImageTagMutabilityOutput, error)
}

// CreateRepository mocks CreateRepository method
func (m *MockRepositoryClient) CreateRepository(ctx context.Context, input *ecr.CreateRepositoryInput, opts ...func(*ecr.Options)) (*ecr.CreateRepositoryOutput, error) {
	return m.MockCreate(ctx, input, opts)
}

// DeleteRepository mocks DeleteRepository method
func (m *MockRepositoryClient) DeleteRepository(ctx context.Context, input *ecr.DeleteRepositoryInput, opts ...func(*ecr.Options)) (*ecr.DeleteRepositoryOutput, error) {
	return m.MockDelete(ctx, input, opts)
}

// DescribeRepositories mocks DescribeRepositories method
func (m *MockRepositoryClient) DescribeRepositories(ctx context.Context, input *ecr.DescribeRepositoriesInput, opts ...func(*ecr.Options)) (*ecr.DescribeRepositoriesOutput, error) {
	return m.MockDescribe(ctx, input, opts)
}

// ListTagsForResource mocks ListTagsForResource method
func (m *MockRepositoryClient) ListTagsForResource(ctx context.Context, input *ecr.ListTagsForResourceInput, opts ...func(*ecr.Options)) (*ecr.ListTagsForResourceOutput, error) {
	return m.MockListTags(ctx, input, opts)
}

// TagResource mocks TagResource method
func (m *MockRepositoryClient) TagResource(ctx context.Context, input *ecr.TagResourceInput, opts ...func(*ecr.Options)) (*ecr.TagResourceOutput, error) {
	return m.MockTag(ctx, input, opts)
}

// UntagResource mocks UntagResource method
func (m *MockRepositoryClient) UntagResource(ctx context.Context, input *ecr.UntagResourceInput, opts ...func(*ecr.Options)) (*ecr.UntagResourceOutput, error) {
	return m.MockUntag(ctx, input, opts)
}

// PutImageTagMutability mocks PutImageTagMutability method
func (m *MockRepositoryClient) PutImageTagMutability(ctx context.Context, input *ecr.PutImageTagMutabilityInput, opts ...func(*ecr.Options)) (*ecr.PutImageTagMutabilityOutput, error) {
	return m.MockPutImageTagMutability(ctx, input, opts)
}

// PutImageScanningConfiguration mocks PutImageScanningConfiguration method
func (m *MockRepositoryClient) PutImageScanningConfiguration(ctx context.Context, input *ecr.PutImageScanningConfigurationInput, opts ...func(*ecr.Options)) (*ecr.PutImageScanningConfigurationOutput, error) {
	return m.MockPutImageScan(ctx, input, opts)
}
