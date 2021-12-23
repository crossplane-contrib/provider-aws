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
var _ clientset.UserClient = (*MockUserClient)(nil)

// MockUserClient is a type that implements all the methods for RoleClient interface
type MockUserClient struct {
	MockGetUser    func(ctx context.Context, input *iam.GetUserInput, opts []func(*iam.Options)) (*iam.GetUserOutput, error)
	MockCreateUser func(ctx context.Context, input *iam.CreateUserInput, opts []func(*iam.Options)) (*iam.CreateUserOutput, error)
	MockDeleteUser func(ctx context.Context, input *iam.DeleteUserInput, opts []func(*iam.Options)) (*iam.DeleteUserOutput, error)
	MockUpdateUser func(ctx context.Context, input *iam.UpdateUserInput, opts []func(*iam.Options)) (*iam.UpdateUserOutput, error)
}

// GetUser mocks GetUser method
func (m *MockUserClient) GetUser(ctx context.Context, input *iam.GetUserInput, opts ...func(*iam.Options)) (*iam.GetUserOutput, error) {
	return m.MockGetUser(ctx, input, opts)
}

// CreateUser mocks CreateUser method
func (m *MockUserClient) CreateUser(ctx context.Context, input *iam.CreateUserInput, opts ...func(*iam.Options)) (*iam.CreateUserOutput, error) {
	return m.MockCreateUser(ctx, input, opts)
}

// DeleteUser mocks DeleteUser method
func (m *MockUserClient) DeleteUser(ctx context.Context, input *iam.DeleteUserInput, opts ...func(*iam.Options)) (*iam.DeleteUserOutput, error) {
	return m.MockDeleteUser(ctx, input, opts)
}

// UpdateUser mocks UpdateUser method
func (m *MockUserClient) UpdateUser(ctx context.Context, input *iam.UpdateUserInput, opts ...func(*iam.Options)) (*iam.UpdateUserOutput, error) {
	return m.MockUpdateUser(ctx, input, opts)
}
