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
var _ clientset.PolicyClient = (*MockPolicyClient)(nil)

// MockPolicyClient is a type that implements all the methods for PolicyClient interface
type MockPolicyClient struct {
	MockGetPolicyRequest           func(*iam.GetPolicyInput) iam.GetPolicyRequest
	MockCreatePolicyRequest        func(*iam.CreatePolicyInput) iam.CreatePolicyRequest
	MockDeletePolicyRequest        func(*iam.DeletePolicyInput) iam.DeletePolicyRequest
	MockGetPolicyVersionRequest    func(*iam.GetPolicyVersionInput) iam.GetPolicyVersionRequest
	MockCreatePolicyVersionRequest func(*iam.CreatePolicyVersionInput) iam.CreatePolicyVersionRequest
	MockListPolicyVersionsRequest  func(*iam.ListPolicyVersionsInput) iam.ListPolicyVersionsRequest
	MockDeletePolicyVersionRequest func(*iam.DeletePolicyVersionInput) iam.DeletePolicyVersionRequest
}

// GetPolicyRequest mocks GetPolicyRequest method
func (m *MockPolicyClient) GetPolicyRequest(input *iam.GetPolicyInput) iam.GetPolicyRequest {
	return m.MockGetPolicyRequest(input)
}

// CreatePolicyRequest mocks CreatePolicyRequest method
func (m *MockPolicyClient) CreatePolicyRequest(input *iam.CreatePolicyInput) iam.CreatePolicyRequest {
	return m.MockCreatePolicyRequest(input)
}

// DeletePolicyRequest mocks DeletePolicyRequest method
func (m *MockPolicyClient) DeletePolicyRequest(input *iam.DeletePolicyInput) iam.DeletePolicyRequest {
	return m.MockDeletePolicyRequest(input)
}

// GetPolicyVersionRequest mocks GetPolicyVersionRequest method
func (m *MockPolicyClient) GetPolicyVersionRequest(input *iam.GetPolicyVersionInput) iam.GetPolicyVersionRequest {
	return m.MockGetPolicyVersionRequest(input)
}

// CreatePolicyVersionRequest mocks CreatePolicyVersionRequest method
func (m *MockPolicyClient) CreatePolicyVersionRequest(input *iam.CreatePolicyVersionInput) iam.CreatePolicyVersionRequest {
	return m.MockCreatePolicyVersionRequest(input)
}

// ListPolicyVersionsRequest mocks ListPolicyVersionsRequest method
func (m *MockPolicyClient) ListPolicyVersionsRequest(input *iam.ListPolicyVersionsInput) iam.ListPolicyVersionsRequest {
	return m.MockListPolicyVersionsRequest(input)
}

// DeletePolicyVersionRequest mocks DeletePolicyVersionRequest method
func (m *MockPolicyClient) DeletePolicyVersionRequest(input *iam.DeletePolicyVersionInput) iam.DeletePolicyVersionRequest {
	return m.MockDeletePolicyVersionRequest(input)
}
