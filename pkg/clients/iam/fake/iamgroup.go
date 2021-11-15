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
var _ clientset.GroupClient = (*MockGroupClient)(nil)

// MockGroupClient is a type that implements all the methods for RoleClient interface
type MockGroupClient struct {
	MockGetGroup    func(ctx context.Context, input *iam.GetGroupInput, opts []func(*iam.Options)) (*iam.GetGroupOutput, error)
	MockCreateGroup func(ctx context.Context, input *iam.CreateGroupInput, opts []func(*iam.Options)) (*iam.CreateGroupOutput, error)
	MockDeleteGroup func(ctx context.Context, input *iam.DeleteGroupInput, opts []func(*iam.Options)) (*iam.DeleteGroupOutput, error)
	MockUpdateGroup func(ctx context.Context, input *iam.UpdateGroupInput, opts []func(*iam.Options)) (*iam.UpdateGroupOutput, error)
}

// GetGroup mocks GetGroup method
func (m *MockGroupClient) GetGroup(ctx context.Context, input *iam.GetGroupInput, opts ...func(*iam.Options)) (*iam.GetGroupOutput, error) {
	return m.MockGetGroup(ctx, input, opts)
}

// CreateGroup mocks CreateGroup method
func (m *MockGroupClient) CreateGroup(ctx context.Context, input *iam.CreateGroupInput, opts ...func(*iam.Options)) (*iam.CreateGroupOutput, error) {
	return m.MockCreateGroup(ctx, input, opts)
}

// DeleteGroup mocks DeleteGroup method
func (m *MockGroupClient) DeleteGroup(ctx context.Context, input *iam.DeleteGroupInput, opts ...func(*iam.Options)) (*iam.DeleteGroupOutput, error) {
	return m.MockDeleteGroup(ctx, input, opts)
}

// UpdateGroup mocks UpdateGroup method
func (m *MockGroupClient) UpdateGroup(ctx context.Context, input *iam.UpdateGroupInput, opts ...func(*iam.Options)) (*iam.UpdateGroupOutput, error) {
	return m.MockUpdateGroup(ctx, input, opts)
}
