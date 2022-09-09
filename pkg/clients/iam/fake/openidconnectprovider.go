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

	clientset "github.com/crossplane-contrib/provider-aws/pkg/clients/iam"
)

// this ensures that the mock implements the client interface
var _ clientset.OpenIDConnectProviderClient = (*MockOpenIDConnectProviderClient)(nil)

// MockOpenIDConnectProviderInput holds the input structures for future inspections
type MockOpenIDConnectProviderInput struct {
	CreateOIDCProviderInput            *iam.CreateOpenIDConnectProviderInput
	TagOpenIDConnectProviderInput      *iam.TagOpenIDConnectProviderInput
	UntagOpenIDConnectProviderInput    *iam.UntagOpenIDConnectProviderInput
	ListOpenIDConnectProvidersInput    *iam.ListOpenIDConnectProvidersInput
	ListOpenIDConnectProviderTagsInput *iam.ListOpenIDConnectProviderTagsInput
}

// MockOpenIDConnectProviderClient is a type that implements all the methods for OpenIDConnectProviderClient interface
type MockOpenIDConnectProviderClient struct {
	MockOpenIDConnectProviderInput              MockOpenIDConnectProviderInput
	MockGetOpenIDConnectProvider                func(ctx context.Context, input *iam.GetOpenIDConnectProviderInput, opts []func(*iam.Options)) (*iam.GetOpenIDConnectProviderOutput, error)
	MockCreateOpenIDConnectProvider             func(ctx context.Context, input *iam.CreateOpenIDConnectProviderInput, opts []func(*iam.Options)) (*iam.CreateOpenIDConnectProviderOutput, error)
	MockAddClientIDToOpenIDConnectProvider      func(ctx context.Context, input *iam.AddClientIDToOpenIDConnectProviderInput, opts []func(*iam.Options)) (*iam.AddClientIDToOpenIDConnectProviderOutput, error)
	MockRemoveClientIDFromOpenIDConnectProvider func(ctx context.Context, input *iam.RemoveClientIDFromOpenIDConnectProviderInput, opts []func(*iam.Options)) (*iam.RemoveClientIDFromOpenIDConnectProviderOutput, error)
	MockUpdateOpenIDConnectProviderThumbprint   func(ctx context.Context, input *iam.UpdateOpenIDConnectProviderThumbprintInput, opts []func(*iam.Options)) (*iam.UpdateOpenIDConnectProviderThumbprintOutput, error)
	MockDeleteOpenIDConnectProvider             func(ctx context.Context, input *iam.DeleteOpenIDConnectProviderInput, opts []func(*iam.Options)) (*iam.DeleteOpenIDConnectProviderOutput, error)
	MockTagOpenIDConnectProvider                func(ctx context.Context, input *iam.TagOpenIDConnectProviderInput, opts []func(*iam.Options)) (*iam.TagOpenIDConnectProviderOutput, error)
	MockUntagOpenIDConnectProvider              func(ctx context.Context, input *iam.UntagOpenIDConnectProviderInput, opts []func(*iam.Options)) (*iam.UntagOpenIDConnectProviderOutput, error)
	MockListOpenIDConnectProviders              func(ctx context.Context, input *iam.ListOpenIDConnectProvidersInput, opts []func(*iam.Options)) (*iam.ListOpenIDConnectProvidersOutput, error)
	MockListOpenIDConnectProviderTags           func(ctx context.Context, input *iam.ListOpenIDConnectProviderTagsInput, opts []func(*iam.Options)) (*iam.ListOpenIDConnectProviderTagsOutput, error)
}

// GetOpenIDConnectProvider mocks client call.
func (m *MockOpenIDConnectProviderClient) GetOpenIDConnectProvider(ctx context.Context, input *iam.GetOpenIDConnectProviderInput, opts ...func(*iam.Options)) (*iam.GetOpenIDConnectProviderOutput, error) {
	return m.MockGetOpenIDConnectProvider(ctx, input, opts)
}

// CreateOpenIDConnectProvider mocks client call.
func (m *MockOpenIDConnectProviderClient) CreateOpenIDConnectProvider(ctx context.Context, input *iam.CreateOpenIDConnectProviderInput, opts ...func(*iam.Options)) (*iam.CreateOpenIDConnectProviderOutput, error) {
	m.MockOpenIDConnectProviderInput.CreateOIDCProviderInput = input
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

// TagOpenIDConnectProvider mocks client call
func (m *MockOpenIDConnectProviderClient) TagOpenIDConnectProvider(ctx context.Context, input *iam.TagOpenIDConnectProviderInput, opts ...func(*iam.Options)) (*iam.TagOpenIDConnectProviderOutput, error) {
	m.MockOpenIDConnectProviderInput.TagOpenIDConnectProviderInput = input
	return m.MockTagOpenIDConnectProvider(ctx, input, opts)
}

// UntagOpenIDConnectProvider mocks client call
func (m *MockOpenIDConnectProviderClient) UntagOpenIDConnectProvider(ctx context.Context, input *iam.UntagOpenIDConnectProviderInput, opts ...func(*iam.Options)) (*iam.UntagOpenIDConnectProviderOutput, error) {
	m.MockOpenIDConnectProviderInput.UntagOpenIDConnectProviderInput = input
	return m.MockUntagOpenIDConnectProvider(ctx, input, opts)
}

// ListOpenIDConnectProviders mocks client call
func (m *MockOpenIDConnectProviderClient) ListOpenIDConnectProviders(ctx context.Context, input *iam.ListOpenIDConnectProvidersInput, opts ...func(*iam.Options)) (*iam.ListOpenIDConnectProvidersOutput, error) {
	m.MockOpenIDConnectProviderInput.ListOpenIDConnectProvidersInput = input
	return m.MockListOpenIDConnectProviders(ctx, input, opts)
}

// ListOpenIDConnectProviderTags mocks client call
func (m *MockOpenIDConnectProviderClient) ListOpenIDConnectProviderTags(ctx context.Context, input *iam.ListOpenIDConnectProviderTagsInput, opts ...func(*iam.Options)) (*iam.ListOpenIDConnectProviderTagsOutput, error) {
	m.MockOpenIDConnectProviderInput.ListOpenIDConnectProviderTagsInput = input
	return m.MockListOpenIDConnectProviderTags(ctx, input, opts)
}
