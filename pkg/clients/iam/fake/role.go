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

	clientset "github.com/crossplane-contrib/provider-aws/pkg/clients/iam"
)

// this ensures that the mock implements the client interface
var _ clientset.RoleClient = (*MockRoleClient)(nil)

// MockRoleClient is a type that implements all the methods for RoleClient interface
type MockRoleClient struct {
	MockGetRole                       func(ctx context.Context, input *iam.GetRoleInput, opts []func(*iam.Options)) (*iam.GetRoleOutput, error)
	MockCreateRole                    func(ctx context.Context, input *iam.CreateRoleInput, opts []func(*iam.Options)) (*iam.CreateRoleOutput, error)
	MockDeleteRole                    func(ctx context.Context, input *iam.DeleteRoleInput, opts []func(*iam.Options)) (*iam.DeleteRoleOutput, error)
	MockUpdateRole                    func(ctx context.Context, input *iam.UpdateRoleInput, opts []func(*iam.Options)) (*iam.UpdateRoleOutput, error)
	MockPutRolePermissionsBoundary    func(ctx context.Context, input *iam.PutRolePermissionsBoundaryInput, opts []func(*iam.Options)) (*iam.PutRolePermissionsBoundaryOutput, error)
	MockDeleteRolePermissionsBoundary func(ctx context.Context, input *iam.DeleteRolePermissionsBoundaryInput, opts []func(*iam.Options)) (*iam.DeleteRolePermissionsBoundaryOutput, error)
	MockUpdateAssumeRolePolicy        func(ctx context.Context, input *iam.UpdateAssumeRolePolicyInput, opts []func(*iam.Options)) (*iam.UpdateAssumeRolePolicyOutput, error)
	MockTagRole                       func(ctx context.Context, input *iam.TagRoleInput, opts []func(*iam.Options)) (*iam.TagRoleOutput, error)
	MockUntagRole                     func(ctx context.Context, input *iam.UntagRoleInput, opts []func(*iam.Options)) (*iam.UntagRoleOutput, error)
}

// GetRole mocks GetRole method
func (m *MockRoleClient) GetRole(ctx context.Context, input *iam.GetRoleInput, opts ...func(*iam.Options)) (*iam.GetRoleOutput, error) {
	return m.MockGetRole(ctx, input, opts)
}

// CreateRole mocks CreateRole method
func (m *MockRoleClient) CreateRole(ctx context.Context, input *iam.CreateRoleInput, opts ...func(*iam.Options)) (*iam.CreateRoleOutput, error) {
	return m.MockCreateRole(ctx, input, opts)
}

// DeleteRole mocks DeleteRole method
func (m *MockRoleClient) DeleteRole(ctx context.Context, input *iam.DeleteRoleInput, opts ...func(*iam.Options)) (*iam.DeleteRoleOutput, error) {
	return m.MockDeleteRole(ctx, input, opts)
}

// UpdateRole mocks UpdateRole method
func (m *MockRoleClient) UpdateRole(ctx context.Context, input *iam.UpdateRoleInput, opts ...func(*iam.Options)) (*iam.UpdateRoleOutput, error) {
	return m.MockUpdateRole(ctx, input, opts)
}

// PutRolePermissionsBoundary mocks PutRolePermissionsBoundary method
func (m *MockRoleClient) PutRolePermissionsBoundary(ctx context.Context, input *iam.PutRolePermissionsBoundaryInput, opts ...func(*iam.Options)) (*iam.PutRolePermissionsBoundaryOutput, error) {
	return m.MockPutRolePermissionsBoundary(ctx, input, opts)
}

// DeleteRolePermissionsBoundary mocks DeleteRolePermissionsBoundary method
func (m *MockRoleClient) DeleteRolePermissionsBoundary(ctx context.Context, input *iam.DeleteRolePermissionsBoundaryInput, opts ...func(*iam.Options)) (*iam.DeleteRolePermissionsBoundaryOutput, error) {
	return m.MockDeleteRolePermissionsBoundary(ctx, input, opts)
}

// UpdateAssumeRolePolicy mocks UpdateAssumeRolePolicy method
func (m *MockRoleClient) UpdateAssumeRolePolicy(ctx context.Context, input *iam.UpdateAssumeRolePolicyInput, opts ...func(*iam.Options)) (*iam.UpdateAssumeRolePolicyOutput, error) {
	return m.MockUpdateAssumeRolePolicy(ctx, input, opts)
}

// TagRole mocks TagRole method
func (m *MockRoleClient) TagRole(ctx context.Context, input *iam.TagRoleInput, opts ...func(*iam.Options)) (*iam.TagRoleOutput, error) {
	return m.MockTagRole(ctx, input, opts)
}

// UntagRole mocks UntagRole method
func (m *MockRoleClient) UntagRole(ctx context.Context, input *iam.UntagRoleInput, opts ...func(*iam.Options)) (*iam.UntagRoleOutput, error) {
	return m.MockUntagRole(ctx, input, opts)
}
