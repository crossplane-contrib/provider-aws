/*
Copyright 2021 The Crossplane Authors.

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
var _ clientset.OpenIDConnectProviderClient = (*MockOpenIDConnectProviderClient)(nil)

// MockOpenIDConnectProviderClient is a type that implements all the methods for OpenIDConnectProviderClient interface
type MockOpenIDConnectProviderClient struct {
	MockGetOpenIDConnectProviderRequest                func(*iam.GetOpenIDConnectProviderInput) iam.GetOpenIDConnectProviderRequest
	MockCreateOpenIDConnectProviderRequest             func(*iam.CreateOpenIDConnectProviderInput) iam.CreateOpenIDConnectProviderRequest
	MockAddClientIDToOpenIDConnectProviderRequest      func(*iam.AddClientIDToOpenIDConnectProviderInput) iam.AddClientIDToOpenIDConnectProviderRequest
	MockRemoveClientIDFromOpenIDConnectProviderRequest func(*iam.RemoveClientIDFromOpenIDConnectProviderInput) iam.RemoveClientIDFromOpenIDConnectProviderRequest
	MockUpdateOpenIDConnectProviderThumbprintRequest   func(*iam.UpdateOpenIDConnectProviderThumbprintInput) iam.UpdateOpenIDConnectProviderThumbprintRequest
	MockDeleteOpenIDConnectProviderRequest             func(*iam.DeleteOpenIDConnectProviderInput) iam.DeleteOpenIDConnectProviderRequest
}

// GetOpenIDConnectProviderRequest mocks client call.
func (m *MockOpenIDConnectProviderClient) GetOpenIDConnectProviderRequest(input *iam.GetOpenIDConnectProviderInput) iam.GetOpenIDConnectProviderRequest {
	return m.MockGetOpenIDConnectProviderRequest(input)
}

// CreateOpenIDConnectProviderRequest mocks client call.
func (m *MockOpenIDConnectProviderClient) CreateOpenIDConnectProviderRequest(input *iam.CreateOpenIDConnectProviderInput) iam.CreateOpenIDConnectProviderRequest {
	return m.MockCreateOpenIDConnectProviderRequest(input)
}

// AddClientIDToOpenIDConnectProviderRequest mocks client call.
func (m *MockOpenIDConnectProviderClient) AddClientIDToOpenIDConnectProviderRequest(input *iam.AddClientIDToOpenIDConnectProviderInput) iam.AddClientIDToOpenIDConnectProviderRequest {
	return m.MockAddClientIDToOpenIDConnectProviderRequest(input)
}

// RemoveClientIDFromOpenIDConnectProviderRequest mocks client call.
func (m *MockOpenIDConnectProviderClient) RemoveClientIDFromOpenIDConnectProviderRequest(input *iam.RemoveClientIDFromOpenIDConnectProviderInput) iam.RemoveClientIDFromOpenIDConnectProviderRequest {
	return m.MockRemoveClientIDFromOpenIDConnectProviderRequest(input)
}

// UpdateOpenIDConnectProviderThumbprintRequest mocks client call.
func (m *MockOpenIDConnectProviderClient) UpdateOpenIDConnectProviderThumbprintRequest(input *iam.UpdateOpenIDConnectProviderThumbprintInput) iam.UpdateOpenIDConnectProviderThumbprintRequest {
	return m.MockUpdateOpenIDConnectProviderThumbprintRequest(input)
}

// DeleteOpenIDConnectProviderRequest mocks client call.
func (m *MockOpenIDConnectProviderClient) DeleteOpenIDConnectProviderRequest(input *iam.DeleteOpenIDConnectProviderInput) iam.DeleteOpenIDConnectProviderRequest {
	return m.MockDeleteOpenIDConnectProviderRequest(input)
}
