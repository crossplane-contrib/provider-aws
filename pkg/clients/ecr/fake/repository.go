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
	"github.com/aws/aws-sdk-go-v2/service/ecr"

	clientset "github.com/crossplane/provider-aws/pkg/clients/ecr"
)

// this ensures that the mock implements the client interface
var _ clientset.RepositoryClient = (*MockRepositoryClient)(nil)

// MockRepositoryClient is a type that implements all the methods for ECRClient interface
type MockRepositoryClient struct {
	MockCreate                func(*ecr.CreateRepositoryInput) ecr.CreateRepositoryRequest
	MockDelete                func(*ecr.DeleteRepositoryInput) ecr.DeleteRepositoryRequest
	MockDescribe              func(*ecr.DescribeRepositoriesInput) ecr.DescribeRepositoriesRequest
	MockListTags              func(*ecr.ListTagsForResourceInput) ecr.ListTagsForResourceRequest
	MockTag                   func(*ecr.TagResourceInput) ecr.TagResourceRequest
	MockUntag                 func(*ecr.UntagResourceInput) ecr.UntagResourceRequest
	MockPutImageScan          func(*ecr.PutImageScanningConfigurationInput) ecr.PutImageScanningConfigurationRequest
	MockPutImageTagMutability func(*ecr.PutImageTagMutabilityInput) ecr.PutImageTagMutabilityRequest
}

// CreateRepositoryRequest mocks CreateRepositoryRequest method
func (m *MockRepositoryClient) CreateRepositoryRequest(input *ecr.CreateRepositoryInput) ecr.CreateRepositoryRequest {
	return m.MockCreate(input)
}

// DeleteRepositoryRequest mocks DeleteRepositoryRequest method
func (m *MockRepositoryClient) DeleteRepositoryRequest(input *ecr.DeleteRepositoryInput) ecr.DeleteRepositoryRequest {
	return m.MockDelete(input)
}

// DescribeRepositoriesRequest mocks DescribeRepositoriesRequest method
func (m *MockRepositoryClient) DescribeRepositoriesRequest(input *ecr.DescribeRepositoriesInput) ecr.DescribeRepositoriesRequest {
	return m.MockDescribe(input)
}

// ListTagsForResourceRequest mocks ListTagsForResourceRequest method
func (m *MockRepositoryClient) ListTagsForResourceRequest(input *ecr.ListTagsForResourceInput) ecr.ListTagsForResourceRequest {
	return m.MockListTags(input)
}

// TagResourceRequest mocks TagResourceRequest method
func (m *MockRepositoryClient) TagResourceRequest(input *ecr.TagResourceInput) ecr.TagResourceRequest {
	return m.MockTag(input)
}

// UntagResourceRequest mocks UntagResourceRequest method
func (m *MockRepositoryClient) UntagResourceRequest(input *ecr.UntagResourceInput) ecr.UntagResourceRequest {
	return m.MockUntag(input)
}

// PutImageTagMutabilityRequest mocks PutImageTagMutabilityRequest method
func (m *MockRepositoryClient) PutImageTagMutabilityRequest(input *ecr.PutImageTagMutabilityInput) ecr.PutImageTagMutabilityRequest {
	return m.MockPutImageTagMutability(input)
}

// PutImageScanningConfigurationRequest mocks PutImageScanningConfigurationRequest method
func (m *MockRepositoryClient) PutImageScanningConfigurationRequest(input *ecr.PutImageScanningConfigurationInput) ecr.PutImageScanningConfigurationRequest {
	return m.MockPutImageScan(input)
}
