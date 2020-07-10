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
var _ clientset.GroupPolicyAttachmentClient = (*MockGroupPolicyAttachmentClient)(nil)

// MockGroupPolicyAttachmentClient is a type that implements all the methods for GroupPolicyAttachmentClient interface
type MockGroupPolicyAttachmentClient struct {
	MockAttachGroupPolicy         func(*iam.AttachGroupPolicyInput) iam.AttachGroupPolicyRequest
	MockListAttachedGroupPolicies func(*iam.ListAttachedGroupPoliciesInput) iam.ListAttachedGroupPoliciesRequest
	MockDetachGroupPolicy         func(*iam.DetachGroupPolicyInput) iam.DetachGroupPolicyRequest
}

// AttachGroupPolicyRequest mocks AttachGroupPolicyRequest method
func (m *MockGroupPolicyAttachmentClient) AttachGroupPolicyRequest(input *iam.AttachGroupPolicyInput) iam.AttachGroupPolicyRequest {
	return m.MockAttachGroupPolicy(input)
}

// ListAttachedGroupPoliciesRequest mocks ListAttachedGroupPoliciesRequest method
func (m *MockGroupPolicyAttachmentClient) ListAttachedGroupPoliciesRequest(input *iam.ListAttachedGroupPoliciesInput) iam.ListAttachedGroupPoliciesRequest {
	return m.MockListAttachedGroupPolicies(input)
}

// DetachGroupPolicyRequest mocks DetachGroupPolicyRequest method
func (m *MockGroupPolicyAttachmentClient) DetachGroupPolicyRequest(input *iam.DetachGroupPolicyInput) iam.DetachGroupPolicyRequest {
	return m.MockDetachGroupPolicy(input)
}
