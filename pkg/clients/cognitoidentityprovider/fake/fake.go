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

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider/cognitoidentityprovideriface"
)

// MockCognitoIdentityProviderClient for testing
type MockCognitoIdentityProviderClient struct {
	cognitoidentityprovideriface.CognitoIdentityProviderAPI

	MockGetUserPoolMfaConfig            func(*cognitoidentityprovider.GetUserPoolMfaConfigInput) (*cognitoidentityprovider.GetUserPoolMfaConfigOutput, error)
	MockSetUserPoolMfaConfigWithContext func(context.Context, *cognitoidentityprovider.SetUserPoolMfaConfigInput, []request.Option) (*cognitoidentityprovider.SetUserPoolMfaConfigOutput, error)
	MockAddCustomAttributes             func(*cognitoidentityprovider.AddCustomAttributesInput) (*cognitoidentityprovider.AddCustomAttributesOutput, error)

	Called MockCognitoIdentityProviderClientCall
}

// CallGetUserPoolMfaConfig to log call
type CallGetUserPoolMfaConfig struct {
	Ctx  aws.Context
	I    *cognitoidentityprovider.GetUserPoolMfaConfigInput
	Opts []request.Option
}

// CallAddCustomAttributes to log call
type CallAddCustomAttributes struct {
	I *cognitoidentityprovider.AddCustomAttributesInput
}

// GetUserPoolMfaConfig calls MockGetUserPoolMfaConfig
func (m *MockCognitoIdentityProviderClient) GetUserPoolMfaConfig(i *cognitoidentityprovider.GetUserPoolMfaConfigInput) (*cognitoidentityprovider.GetUserPoolMfaConfigOutput, error) {
	m.Called.GetUserPoolMfaConfig = append(m.Called.GetUserPoolMfaConfig, &CallGetUserPoolMfaConfig{I: i})

	return m.MockGetUserPoolMfaConfig(i)
}

// CallSetUserPoolMfaConfigWithContext to log call
type CallSetUserPoolMfaConfigWithContext struct {
	Ctx  aws.Context
	I    *cognitoidentityprovider.SetUserPoolMfaConfigInput
	Opts []request.Option
}

// SetUserPoolMfaConfigWithContext calls MockSetUserPoolMfaConfigWithContext
func (m *MockCognitoIdentityProviderClient) SetUserPoolMfaConfigWithContext(ctx context.Context, i *cognitoidentityprovider.SetUserPoolMfaConfigInput, opts ...request.Option) (*cognitoidentityprovider.SetUserPoolMfaConfigOutput, error) {
	m.Called.SetUserPoolMfaConfigWithContext = append(m.Called.SetUserPoolMfaConfigWithContext, &CallSetUserPoolMfaConfigWithContext{Ctx: ctx, I: i, Opts: opts})

	return m.MockSetUserPoolMfaConfigWithContext(ctx, i, opts)
}

func (m *MockCognitoIdentityProviderClient) AddCustomAttributes(in *cognitoidentityprovider.AddCustomAttributesInput) (*cognitoidentityprovider.AddCustomAttributesOutput, error) {
	m.Called.MockAddCustomAttributes = append(m.Called.MockAddCustomAttributes, &CallAddCustomAttributes{I: in})
	return m.MockAddCustomAttributes(in)
}

// MockCognitoIdentityProviderClientCall to log calls
type MockCognitoIdentityProviderClientCall struct {
	GetUserPoolMfaConfig            []*CallGetUserPoolMfaConfig
	SetUserPoolMfaConfigWithContext []*CallSetUserPoolMfaConfigWithContext
	MockAddCustomAttributes         []*CallAddCustomAttributes
}
