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
var _ clientset.RolePolicyAttachmentClient = (*MockRolePolicyAttachmentClient)(nil)

// MockRolePolicyAttachmentClient is a type that implements all the methods for RolePolicyAttachmentClient interface
type MockRolePolicyAttachmentClient struct {
	MockAttachRolePolicy         func(ctx context.Context, input *iam.AttachRolePolicyInput, opts []func(*iam.Options)) (*iam.AttachRolePolicyOutput, error)
	MockListAttachedRolePolicies func(ctx context.Context, input *iam.ListAttachedRolePoliciesInput, opts []func(*iam.Options)) (*iam.ListAttachedRolePoliciesOutput, error)
	MockDetachRolePolicy         func(ctx context.Context, input *iam.DetachRolePolicyInput, opts []func(*iam.Options)) (*iam.DetachRolePolicyOutput, error)
}

// AttachRolePolicy mocks AttachRolePolicy method
func (m *MockRolePolicyAttachmentClient) AttachRolePolicy(ctx context.Context, input *iam.AttachRolePolicyInput, opts ...func(*iam.Options)) (*iam.AttachRolePolicyOutput, error) {
	return m.MockAttachRolePolicy(ctx, input, opts)
}

// ListAttachedRolePolicies mocks ListAttachedRolePolicies method
func (m *MockRolePolicyAttachmentClient) ListAttachedRolePolicies(ctx context.Context, input *iam.ListAttachedRolePoliciesInput, opts ...func(*iam.Options)) (*iam.ListAttachedRolePoliciesOutput, error) {
	return m.MockListAttachedRolePolicies(ctx, input, opts)
}

// DetachRolePolicy mocks DetachRolePolicy method
func (m *MockRolePolicyAttachmentClient) DetachRolePolicy(ctx context.Context, input *iam.DetachRolePolicyInput, opts ...func(*iam.Options)) (*iam.DetachRolePolicyOutput, error) {
	return m.MockDetachRolePolicy(ctx, input, opts)
}
