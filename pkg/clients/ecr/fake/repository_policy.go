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
	"github.com/aws/aws-sdk-go-v2/service/ecr"

	clientset "github.com/crossplane/provider-aws/pkg/clients/ecr"
)

// this ensures that the mock implements the client interface
var _ clientset.RepositoryPolicyClient = (*MockRepositoryPolicyClient)(nil)

// MockRepositoryPolicyClient is a type that implements all the methods for ECRClient interface
type MockRepositoryPolicyClient struct {
	MockSet    func(*ecr.SetRepositoryPolicyInput) ecr.SetRepositoryPolicyRequest
	MockDelete func(*ecr.DeleteRepositoryPolicyInput) ecr.DeleteRepositoryPolicyRequest
	MockGet    func(*ecr.GetRepositoryPolicyInput) ecr.GetRepositoryPolicyRequest
}

// SetRepositoryPolicyRequest mocks SetRepositoryPolicyRequest method
func (m *MockRepositoryPolicyClient) SetRepositoryPolicyRequest(input *ecr.SetRepositoryPolicyInput) ecr.SetRepositoryPolicyRequest {
	return m.MockSet(input)
}

// DeleteRepositoryPolicyRequest mocks DeleteRepositoryPolicyRequest method
func (m *MockRepositoryPolicyClient) DeleteRepositoryPolicyRequest(input *ecr.DeleteRepositoryPolicyInput) ecr.DeleteRepositoryPolicyRequest {
	return m.MockDelete(input)
}

// GetRepositoryPolicyRequest mocks GetRepositoryPolicyRequest method
func (m *MockRepositoryPolicyClient) GetRepositoryPolicyRequest(input *ecr.GetRepositoryPolicyInput) ecr.GetRepositoryPolicyRequest {
	return m.MockGet(input)
}
