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
	"github.com/aws/aws-sdk-go-v2/service/iam"

	clientset "github.com/crossplane/provider-aws/pkg/clients/iam"
)

// this ensures that the mock implements the client interface
var _ clientset.UserClient = (*MockUserClient)(nil)

// MockUserClient is a type that implements all the methods for RoleClient interface
type MockUserClient struct {
	MockGetUser             func(*iam.GetUserInput) iam.GetUserRequest
	MockCreateUser          func(*iam.CreateUserInput) iam.CreateUserRequest
	MockDeleteUser          func(*iam.DeleteUserInput) iam.DeleteUserRequest
	MockUpdateUser          func(*iam.UpdateUserInput) iam.UpdateUserRequest
	MockRemoveUserFromGroup func(*iam.RemoveUserFromGroupInput) iam.RemoveUserFromGroupRequest
	MockAddUserToGroup      func(*iam.AddUserToGroupInput) iam.AddUserToGroupRequest
	MockListGroupsForUser   func(*iam.ListGroupsForUserInput) iam.ListGroupsForUserRequest
}

// GetUserRequest mocks GetUserRequest method
func (m *MockUserClient) GetUserRequest(input *iam.GetUserInput) iam.GetUserRequest {
	return m.MockGetUser(input)
}

// CreateUserRequest mocks CreateUserRequest method
func (m *MockUserClient) CreateUserRequest(input *iam.CreateUserInput) iam.CreateUserRequest {
	return m.MockCreateUser(input)
}

// DeleteUserRequest mocks DeleteUserRequest method
func (m *MockUserClient) DeleteUserRequest(input *iam.DeleteUserInput) iam.DeleteUserRequest {
	return m.MockDeleteUser(input)
}

// UpdateUserRequest mocks UpdateUserRequest method
func (m *MockUserClient) UpdateUserRequest(input *iam.UpdateUserInput) iam.UpdateUserRequest {
	return m.MockUpdateUser(input)
}

// RemoveUserFromGroupRequest mocks RemoveUserFromGroupRequest method
func (m *MockUserClient) RemoveUserFromGroupRequest(input *iam.RemoveUserFromGroupInput) iam.RemoveUserFromGroupRequest {
	return m.MockRemoveUserFromGroup(input)
}

// AddUserToGroupRequest mocks AddUserToGroupRequest method
func (m *MockUserClient) AddUserToGroupRequest(input *iam.AddUserToGroupInput) iam.AddUserToGroupRequest {
	return m.MockAddUserToGroup(input)
}

// ListGroupsForUserRequest mocks ListGroupsForUserRequest method
func (m *MockUserClient) ListGroupsForUserRequest(input *iam.ListGroupsForUserInput) iam.ListGroupsForUserRequest {
	return m.MockListGroupsForUser(input)
}
