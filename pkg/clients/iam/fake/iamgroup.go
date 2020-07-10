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
var _ clientset.GroupClient = (*MockGroupClient)(nil)

// MockGroupClient is a type that implements all the methods for RoleClient interface
type MockGroupClient struct {
	MockGetGroup    func(*iam.GetGroupInput) iam.GetGroupRequest
	MockCreateGroup func(*iam.CreateGroupInput) iam.CreateGroupRequest
	MockDeleteGroup func(*iam.DeleteGroupInput) iam.DeleteGroupRequest
	MockUpdateGroup func(*iam.UpdateGroupInput) iam.UpdateGroupRequest
}

// GetGroupRequest mocks GetGroupRequest method
func (m *MockGroupClient) GetGroupRequest(input *iam.GetGroupInput) iam.GetGroupRequest {
	return m.MockGetGroup(input)
}

// CreateGroupRequest mocks CreateGroupRequest method
func (m *MockGroupClient) CreateGroupRequest(input *iam.CreateGroupInput) iam.CreateGroupRequest {
	return m.MockCreateGroup(input)
}

// DeleteGroupRequest mocks DeleteGroupRequest method
func (m *MockGroupClient) DeleteGroupRequest(input *iam.DeleteGroupInput) iam.DeleteGroupRequest {
	return m.MockDeleteGroup(input)
}

// UpdateGroupRequest mocks UpdateGroupRequest method
func (m *MockGroupClient) UpdateGroupRequest(input *iam.UpdateGroupInput) iam.UpdateGroupRequest {
	return m.MockUpdateGroup(input)
}
