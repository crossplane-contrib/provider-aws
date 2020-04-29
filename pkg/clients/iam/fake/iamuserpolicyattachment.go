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
var _ clientset.UserPolicyAttachmentClient = (*MockUserPolicyAttachmentClient)(nil)

// MockUserPolicyAttachmentClient is a type that implements all the methods for UserPolicyAttachmentClient interface
type MockUserPolicyAttachmentClient struct {
	MockAttachUserPolicy         func(*iam.AttachUserPolicyInput) iam.AttachUserPolicyRequest
	MockListAttachedUserPolicies func(*iam.ListAttachedUserPoliciesInput) iam.ListAttachedUserPoliciesRequest
	MockDetachUserPolicy         func(*iam.DetachUserPolicyInput) iam.DetachUserPolicyRequest
}

// AttachUserPolicyRequest mocks AttachUserPolicyRequest method
func (m *MockUserPolicyAttachmentClient) AttachUserPolicyRequest(input *iam.AttachUserPolicyInput) iam.AttachUserPolicyRequest {
	return m.MockAttachUserPolicy(input)
}

// ListAttachedUserPoliciesRequest mocks ListAttachedUserPoliciesRequest method
func (m *MockUserPolicyAttachmentClient) ListAttachedUserPoliciesRequest(input *iam.ListAttachedUserPoliciesInput) iam.ListAttachedUserPoliciesRequest {
	return m.MockListAttachedUserPolicies(input)
}

// DetachUserPolicyRequest mocks DetachUserPolicyRequest method
func (m *MockUserPolicyAttachmentClient) DetachUserPolicyRequest(input *iam.DetachUserPolicyInput) iam.DetachUserPolicyRequest {
	return m.MockDetachUserPolicy(input)
}
