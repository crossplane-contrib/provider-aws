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
var _ clientset.GroupPolicyAttachmentClient = (*MockGroupPolicyAttachmentClient)(nil)

// MockGroupPolicyAttachmentClient is a type that implements all the methods for GroupPolicyAttachmentClient interface
type MockGroupPolicyAttachmentClient struct {
	MockAttachGroupPolicy         func(ctx context.Context, input *iam.AttachGroupPolicyInput, opts []func(*iam.Options)) (*iam.AttachGroupPolicyOutput, error)
	MockListAttachedGroupPolicies func(ctx context.Context, input *iam.ListAttachedGroupPoliciesInput, opts []func(*iam.Options)) (*iam.ListAttachedGroupPoliciesOutput, error)
	MockDetachGroupPolicy         func(ctx context.Context, input *iam.DetachGroupPolicyInput, opts []func(*iam.Options)) (*iam.DetachGroupPolicyOutput, error)
}

// AttachGroupPolicy mocks AttachGroupPolicy method
func (m *MockGroupPolicyAttachmentClient) AttachGroupPolicy(ctx context.Context, input *iam.AttachGroupPolicyInput, opts ...func(*iam.Options)) (*iam.AttachGroupPolicyOutput, error) {
	return m.MockAttachGroupPolicy(ctx, input, opts)
}

// ListAttachedGroupPolicies mocks ListAttachedGroupPolicies method
func (m *MockGroupPolicyAttachmentClient) ListAttachedGroupPolicies(ctx context.Context, input *iam.ListAttachedGroupPoliciesInput, opts ...func(*iam.Options)) (*iam.ListAttachedGroupPoliciesOutput, error) {
	return m.MockListAttachedGroupPolicies(ctx, input, opts)
}

// DetachGroupPolicy mocks DetachGroupPolicy method
func (m *MockGroupPolicyAttachmentClient) DetachGroupPolicy(ctx context.Context, input *iam.DetachGroupPolicyInput, opts ...func(*iam.Options)) (*iam.DetachGroupPolicyOutput, error) {
	return m.MockDetachGroupPolicy(ctx, input, opts)
}
