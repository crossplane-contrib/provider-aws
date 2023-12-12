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
var _ clientset.RolePolicyClient = (*MockRolePolicyClient)(nil)

// MockRolePolicyClient is a type that implements all the methods for RolePolicyClient interface
type MockRolePolicyClient struct {
	MockGetRolePolicy    func(ctx context.Context, input *iam.GetRolePolicyInput, opts []func(*iam.Options)) (*iam.GetRolePolicyOutput, error)
	MockPutRolePolicy    func(ctx context.Context, input *iam.PutRolePolicyInput, opts []func(*iam.Options)) (*iam.PutRolePolicyOutput, error)
	MockDeleteRolePolicy func(ctx context.Context, input *iam.DeleteRolePolicyInput, opts []func(*iam.Options)) (*iam.DeleteRolePolicyOutput, error)
}

// GetRolePolicy mocks GetRolePolicy method
func (m *MockRolePolicyClient) GetRolePolicy(ctx context.Context, input *iam.GetRolePolicyInput, opts ...func(*iam.Options)) (*iam.GetRolePolicyOutput, error) {
	return m.MockGetRolePolicy(ctx, input, opts)
}

// PutRolePolicy mocks PutRolePolicy method
func (m *MockRolePolicyClient) PutRolePolicy(ctx context.Context, input *iam.PutRolePolicyInput, opts ...func(*iam.Options)) (*iam.PutRolePolicyOutput, error) {
	return m.MockPutRolePolicy(ctx, input, opts)
}

// DeleteRolePolicy mocks DeleteRolePolicy method
func (m *MockRolePolicyClient) DeleteRolePolicy(ctx context.Context, input *iam.DeleteRolePolicyInput, opts ...func(*iam.Options)) (*iam.DeleteRolePolicyOutput, error) {
	return m.MockDeleteRolePolicy(ctx, input, opts)
}
