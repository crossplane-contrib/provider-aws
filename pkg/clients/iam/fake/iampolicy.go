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

	"github.com/aws/aws-sdk-go-v2/service/iam"

	clientset "github.com/crossplane/provider-aws/pkg/clients/iam"
)

// this ensures that the mock implements the client interface
var _ clientset.PolicyClient = (*MockPolicyClient)(nil)

// MockPolicyClient is a type that implements all the methods for PolicyClient interface
type MockPolicyClient struct {
	MockGetPolicy           func(ctx context.Context, input *iam.GetPolicyInput, opts []func(*iam.Options)) (*iam.GetPolicyOutput, error)
	MockCreatePolicy        func(ctx context.Context, input *iam.CreatePolicyInput, opts []func(*iam.Options)) (*iam.CreatePolicyOutput, error)
	MockDeletePolicy        func(ctx context.Context, input *iam.DeletePolicyInput, opts []func(*iam.Options)) (*iam.DeletePolicyOutput, error)
	MockGetPolicyVersion    func(ctx context.Context, input *iam.GetPolicyVersionInput, opts []func(*iam.Options)) (*iam.GetPolicyVersionOutput, error)
	MockCreatePolicyVersion func(ctx context.Context, input *iam.CreatePolicyVersionInput, opts []func(*iam.Options)) (*iam.CreatePolicyVersionOutput, error)
	MockListPolicyVersions  func(ctx context.Context, input *iam.ListPolicyVersionsInput, opts []func(*iam.Options)) (*iam.ListPolicyVersionsOutput, error)
	MockDeletePolicyVersion func(ctx context.Context, input *iam.DeletePolicyVersionInput, opts []func(*iam.Options)) (*iam.DeletePolicyVersionOutput, error)
}

// GetPolicy mocks GetPolicy method
func (m *MockPolicyClient) GetPolicy(ctx context.Context, input *iam.GetPolicyInput, opts ...func(*iam.Options)) (*iam.GetPolicyOutput, error) {
	return m.MockGetPolicy(ctx, input, opts)
}

// CreatePolicy mocks CreatePolicy method
func (m *MockPolicyClient) CreatePolicy(ctx context.Context, input *iam.CreatePolicyInput, opts ...func(*iam.Options)) (*iam.CreatePolicyOutput, error) {
	return m.MockCreatePolicy(ctx, input, opts)
}

// DeletePolicy mocks DeletePolicy method
func (m *MockPolicyClient) DeletePolicy(ctx context.Context, input *iam.DeletePolicyInput, opts ...func(*iam.Options)) (*iam.DeletePolicyOutput, error) {
	return m.MockDeletePolicy(ctx, input, opts)
}

// GetPolicyVersion mocks GetPolicyVersion method
func (m *MockPolicyClient) GetPolicyVersion(ctx context.Context, input *iam.GetPolicyVersionInput, opts ...func(*iam.Options)) (*iam.GetPolicyVersionOutput, error) {
	return m.MockGetPolicyVersion(ctx, input, opts)
}

// CreatePolicyVersion mocks CreatePolicyVersion method
func (m *MockPolicyClient) CreatePolicyVersion(ctx context.Context, input *iam.CreatePolicyVersionInput, opts ...func(*iam.Options)) (*iam.CreatePolicyVersionOutput, error) {
	return m.MockCreatePolicyVersion(ctx, input, opts)
}

// ListPolicyVersions mocks ListPolicyVersions method
func (m *MockPolicyClient) ListPolicyVersions(ctx context.Context, input *iam.ListPolicyVersionsInput, opts ...func(*iam.Options)) (*iam.ListPolicyVersionsOutput, error) {
	return m.MockListPolicyVersions(ctx, input, opts)
}

// DeletePolicyVersion mocks DeletePolicyVersion method
func (m *MockPolicyClient) DeletePolicyVersion(ctx context.Context, input *iam.DeletePolicyVersionInput, opts ...func(*iam.Options)) (*iam.DeletePolicyVersionOutput, error) {
	return m.MockDeletePolicyVersion(ctx, input, opts)
}
