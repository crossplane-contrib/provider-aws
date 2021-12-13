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
var _ clientset.UserPolicyAttachmentClient = (*MockUserPolicyAttachmentClient)(nil)

// MockUserPolicyAttachmentClient is a type that implements all the methods for UserPolicyAttachmentClient interface
type MockUserPolicyAttachmentClient struct {
	MockAttachUserPolicy         func(ctx context.Context, input *iam.AttachUserPolicyInput, opts []func(*iam.Options)) (*iam.AttachUserPolicyOutput, error)
	MockListAttachedUserPolicies func(ctx context.Context, input *iam.ListAttachedUserPoliciesInput, opts []func(*iam.Options)) (*iam.ListAttachedUserPoliciesOutput, error)
	MockDetachUserPolicy         func(ctx context.Context, input *iam.DetachUserPolicyInput, opts []func(*iam.Options)) (*iam.DetachUserPolicyOutput, error)
}

// AttachUserPolicy mocks AttachUserPolicy method
func (m *MockUserPolicyAttachmentClient) AttachUserPolicy(ctx context.Context, input *iam.AttachUserPolicyInput, opts ...func(*iam.Options)) (*iam.AttachUserPolicyOutput, error) {
	return m.MockAttachUserPolicy(ctx, input, opts)
}

// ListAttachedUserPolicies mocks ListAttachedUserPolicies method
func (m *MockUserPolicyAttachmentClient) ListAttachedUserPolicies(ctx context.Context, input *iam.ListAttachedUserPoliciesInput, opts ...func(*iam.Options)) (*iam.ListAttachedUserPoliciesOutput, error) {
	return m.MockListAttachedUserPolicies(ctx, input, opts)
}

// DetachUserPolicy mocks DetachUserPolicy method
func (m *MockUserPolicyAttachmentClient) DetachUserPolicy(ctx context.Context, input *iam.DetachUserPolicyInput, opts ...func(*iam.Options)) (*iam.DetachUserPolicyOutput, error) {
	return m.MockDetachUserPolicy(ctx, input, opts)
}
