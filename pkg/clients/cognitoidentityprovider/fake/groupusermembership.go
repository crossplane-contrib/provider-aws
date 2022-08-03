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

	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"

	clientset "github.com/crossplane-contrib/provider-aws/pkg/clients/cognitoidentityprovider"
)

// this ensures that the mock implements the client interface
var _ clientset.GroupUserMembershipClient = (*MockGroupUserMembershipClient)(nil)

// MockGroupUserMembershipClient is a type that implements all the methods for GroupUserMembership Client interface
type MockGroupUserMembershipClient struct {
	MockAdminAddUserToGroup      func(ctx context.Context, input *cognitoidentityprovider.AdminAddUserToGroupInput, opts []func(*cognitoidentityprovider.Options)) (*cognitoidentityprovider.AdminAddUserToGroupOutput, error)
	MockAdminRemoveUserFromGroup func(ctx context.Context, input *cognitoidentityprovider.AdminRemoveUserFromGroupInput, opts []func(*cognitoidentityprovider.Options)) (*cognitoidentityprovider.AdminRemoveUserFromGroupOutput, error)
	MockAdminListGroupsForUser   func(ctx context.Context, input *cognitoidentityprovider.AdminListGroupsForUserInput, opts []func(*cognitoidentityprovider.Options)) (*cognitoidentityprovider.AdminListGroupsForUserOutput, error)
}

// AdminAddUserToGroup mocks AdminAddUserToGroup method
func (m *MockGroupUserMembershipClient) AdminAddUserToGroup(ctx context.Context, input *cognitoidentityprovider.AdminAddUserToGroupInput, opts ...func(*cognitoidentityprovider.Options)) (*cognitoidentityprovider.AdminAddUserToGroupOutput, error) {
	return m.MockAdminAddUserToGroup(ctx, input, opts)
}

// AdminRemoveUserFromGroup mocks AdminRemoveUserFromGroup method
func (m *MockGroupUserMembershipClient) AdminRemoveUserFromGroup(ctx context.Context, input *cognitoidentityprovider.AdminRemoveUserFromGroupInput, opts ...func(*cognitoidentityprovider.Options)) (*cognitoidentityprovider.AdminRemoveUserFromGroupOutput, error) {
	return m.MockAdminRemoveUserFromGroup(ctx, input, opts)
}

// AdminListGroupsForUser mocks AdminListGroupsForUser method
func (m *MockGroupUserMembershipClient) AdminListGroupsForUser(ctx context.Context, input *cognitoidentityprovider.AdminListGroupsForUserInput, opts ...func(*cognitoidentityprovider.Options)) (*cognitoidentityprovider.AdminListGroupsForUserOutput, error) {
	return m.MockAdminListGroupsForUser(ctx, input, opts)
}
