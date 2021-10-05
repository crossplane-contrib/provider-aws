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
var _ clientset.GroupUserMembershipClient = (*MockGroupUserMembershipClient)(nil)

// MockGroupUserMembershipClient is a type that implements all the methods for IAMGroupUserMembershipClient interface
type MockGroupUserMembershipClient struct {
	MockAddUserToGroup      func(ctx context.Context, input *iam.AddUserToGroupInput, opts []func(*iam.Options)) (*iam.AddUserToGroupOutput, error)
	MockRemoveUserFromGroup func(ctx context.Context, input *iam.RemoveUserFromGroupInput, opts []func(*iam.Options)) (*iam.RemoveUserFromGroupOutput, error)
	MockListGroupsForUser   func(ctx context.Context, input *iam.ListGroupsForUserInput, opts []func(*iam.Options)) (*iam.ListGroupsForUserOutput, error)
}

// AddUserToGroup mocks AddUserToGroup method
func (m *MockGroupUserMembershipClient) AddUserToGroup(ctx context.Context, input *iam.AddUserToGroupInput, opts ...func(*iam.Options)) (*iam.AddUserToGroupOutput, error) {
	return m.MockAddUserToGroup(ctx, input, opts)
}

// RemoveUserFromGroup mocks RemoveUserFromGroup method
func (m *MockGroupUserMembershipClient) RemoveUserFromGroup(ctx context.Context, input *iam.RemoveUserFromGroupInput, opts ...func(*iam.Options)) (*iam.RemoveUserFromGroupOutput, error) {
	return m.MockRemoveUserFromGroup(ctx, input, opts)
}

// ListGroupsForUser mocks ListGroupsForUser method
func (m *MockGroupUserMembershipClient) ListGroupsForUser(ctx context.Context, input *iam.ListGroupsForUserInput, opts ...func(*iam.Options)) (*iam.ListGroupsForUserOutput, error) {
	return m.MockListGroupsForUser(ctx, input, opts)
}
