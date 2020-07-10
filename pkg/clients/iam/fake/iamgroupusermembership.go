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
var _ clientset.GroupUserMembershipClient = (*MockGroupUserMembershipClient)(nil)

// MockGroupUserMembershipClient is a type that implements all the methods for IAMGroupUserMembershipClient interface
type MockGroupUserMembershipClient struct {
	MockAddUserToGroup      func(*iam.AddUserToGroupInput) iam.AddUserToGroupRequest
	MockRemoveUserFromGroup func(*iam.RemoveUserFromGroupInput) iam.RemoveUserFromGroupRequest
	MockListGroupsForUser   func(*iam.ListGroupsForUserInput) iam.ListGroupsForUserRequest
}

// AddUserToGroupRequest mocks AddUserToGroupRequest method
func (m *MockGroupUserMembershipClient) AddUserToGroupRequest(input *iam.AddUserToGroupInput) iam.AddUserToGroupRequest {
	return m.MockAddUserToGroup(input)
}

// RemoveUserFromGroupRequest mocks RemoveUserFromGroupRequest method
func (m *MockGroupUserMembershipClient) RemoveUserFromGroupRequest(input *iam.RemoveUserFromGroupInput) iam.RemoveUserFromGroupRequest {
	return m.MockRemoveUserFromGroup(input)
}

// ListGroupsForUserRequest mocks ListGroupsForUserRequest method
func (m *MockGroupUserMembershipClient) ListGroupsForUserRequest(input *iam.ListGroupsForUserInput) iam.ListGroupsForUserRequest {
	return m.MockListGroupsForUser(input)
}
