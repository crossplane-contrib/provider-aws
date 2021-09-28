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

	"github.com/aws/aws-sdk-go-v2/service/ecr"

	clientset "github.com/crossplane/provider-aws/pkg/clients/ecr"
)

// this ensures that the mock implements the client interface
var _ clientset.RepositoryPolicyClient = (*MockRepositoryPolicyClient)(nil)

// MockRepositoryPolicyClient is a type that implements all the methods for ECRClient interface
type MockRepositoryPolicyClient struct {
	MockSet    func(ctx context.Context, input *ecr.SetRepositoryPolicyInput, opts []func(*ecr.Options)) (*ecr.SetRepositoryPolicyOutput, error)
	MockDelete func(ctx context.Context, input *ecr.DeleteRepositoryPolicyInput, opts []func(*ecr.Options)) (*ecr.DeleteRepositoryPolicyOutput, error)
	MockGet    func(ctx context.Context, input *ecr.GetRepositoryPolicyInput, opts []func(*ecr.Options)) (*ecr.GetRepositoryPolicyOutput, error)
}

// SetRepositoryPolicy mocks ecr method
func (m *MockRepositoryPolicyClient) SetRepositoryPolicy(ctx context.Context, input *ecr.SetRepositoryPolicyInput, opts ...func(*ecr.Options)) (*ecr.SetRepositoryPolicyOutput, error) {
	return m.MockSet(ctx, input, opts)
}

// DeleteRepositoryPolicy mocks ecr method
func (m *MockRepositoryPolicyClient) DeleteRepositoryPolicy(ctx context.Context, input *ecr.DeleteRepositoryPolicyInput, opts ...func(*ecr.Options)) (*ecr.DeleteRepositoryPolicyOutput, error) {
	return m.MockDelete(ctx, input, opts)
}

// GetRepositoryPolicy mocks ecr method
func (m *MockRepositoryPolicyClient) GetRepositoryPolicy(ctx context.Context, input *ecr.GetRepositoryPolicyInput, opts ...func(*ecr.Options)) (*ecr.GetRepositoryPolicyOutput, error) {
	return m.MockGet(ctx, input, opts)
}
