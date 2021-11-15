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
	"context"

	"github.com/aws/aws-sdk-go-v2/service/iam"

	clientset "github.com/crossplane/provider-aws/pkg/clients/iam"
)

// this ensures that the mock implements the client interface
var _ clientset.OpenIDConnectProviderClient = (*MockOpenIDConnectProviderClient)(nil)

// MockOpenIDConnectProviderClient is a type that implements all the methods for OpenIDConnectProviderClient interface
type MockOpenIDConnectProviderClient struct {
	MockGetOpenIDConnectProvider                func(ctx context.Context, input *iam.GetOpenIDConnectProviderInput, opts []func(*iam.Options)) (*iam.GetOpenIDConnectProviderOutput, error)
	MockCreateOpenIDConnectProvider             func(ctx context.Context, input *iam.CreateOpenIDConnectProviderInput, opts []func(*iam.Options)) (*iam.CreateOpenIDConnectProviderOutput, error)
	MockAddClientIDToOpenIDConnectProvider      func(ctx context.Context, input *iam.AddClientIDToOpenIDConnectProviderInput, opts []func(*iam.Options)) (*iam.AddClientIDToOpenIDConnectProviderOutput, error)
	MockRemoveClientIDFromOpenIDConnectProvider func(ctx context.Context, input *iam.RemoveClientIDFromOpenIDConnectProviderInput, opts []func(*iam.Options)) (*iam.RemoveClientIDFromOpenIDConnectProviderOutput, error)
	MockUpdateOpenIDConnectProviderThumbprint   func(ctx context.Context, input *iam.UpdateOpenIDConnectProviderThumbprintInput, opts []func(*iam.Options)) (*iam.UpdateOpenIDConnectProviderThumbprintOutput, error)
	MockDeleteOpenIDConnectProvider             func(ctx context.Context, input *iam.DeleteOpenIDConnectProviderInput, opts []func(*iam.Options)) (*iam.DeleteOpenIDConnectProviderOutput, error)
}

// GetOpenIDConnectProvider mocks client call.
func (m *MockOpenIDConnectProviderClient) GetOpenIDConnectProvider(ctx context.Context, input *iam.GetOpenIDConnectProviderInput, opts ...func(*iam.Options)) (*iam.GetOpenIDConnectProviderOutput, error) {
	return m.MockGetOpenIDConnectProvider(ctx, input, opts)
}

// CreateOpenIDConnectProvider mocks client call.
func (m *MockOpenIDConnectProviderClient) CreateOpenIDConnectProvider(ctx context.Context, input *iam.CreateOpenIDConnectProviderInput, opts ...func(*iam.Options)) (*iam.CreateOpenIDConnectProviderOutput, error) {
	return m.MockCreateOpenIDConnectProvider(ctx, input, opts)
}

// AddClientIDToOpenIDConnectProvider mocks client call.
func (m *MockOpenIDConnectProviderClient) AddClientIDToOpenIDConnectProvider(ctx context.Context, input *iam.AddClientIDToOpenIDConnectProviderInput, opts ...func(*iam.Options)) (*iam.AddClientIDToOpenIDConnectProviderOutput, error) {
	return m.MockAddClientIDToOpenIDConnectProvider(ctx, input, opts)
}

// RemoveClientIDFromOpenIDConnectProvider mocks client call.
func (m *MockOpenIDConnectProviderClient) RemoveClientIDFromOpenIDConnectProvider(ctx context.Context, input *iam.RemoveClientIDFromOpenIDConnectProviderInput, opts ...func(*iam.Options)) (*iam.RemoveClientIDFromOpenIDConnectProviderOutput, error) {
	return m.MockRemoveClientIDFromOpenIDConnectProvider(ctx, input, opts)
}

// UpdateOpenIDConnectProviderThumbprint mocks client call.
func (m *MockOpenIDConnectProviderClient) UpdateOpenIDConnectProviderThumbprint(ctx context.Context, input *iam.UpdateOpenIDConnectProviderThumbprintInput, opts ...func(*iam.Options)) (*iam.UpdateOpenIDConnectProviderThumbprintOutput, error) {
	return m.MockUpdateOpenIDConnectProviderThumbprint(ctx, input, opts)
}

// DeleteOpenIDConnectProvider mocks client call.
func (m *MockOpenIDConnectProviderClient) DeleteOpenIDConnectProvider(ctx context.Context, input *iam.DeleteOpenIDConnectProviderInput, opts ...func(*iam.Options)) (*iam.DeleteOpenIDConnectProviderOutput, error) {
	return m.MockDeleteOpenIDConnectProvider(ctx, input, opts)
}
