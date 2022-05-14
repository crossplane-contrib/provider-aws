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

	"github.com/aws/aws-sdk-go/aws/request"
	svcsdk "github.com/aws/aws-sdk-go/service/apigateway"

	clientset "github.com/crossplane/provider-aws/pkg/clients/apigateway"
)

// ensures its a valid client
var _ clientset.Client = (*MockAPIGatewayClient)(nil)

// MockAPIGatewayClient is a type that implements all the methods for ApiGatewayClient interface
type MockAPIGatewayClient struct {
	MockGetRestAPIByID         func(context.Context, *string) (*svcsdk.RestApi, error)
	MockGetRestAPIRootResource func(context.Context, *string) (*string, error)
	MockGetResource            func(context.Context, *svcsdk.GetResourceInput, ...request.Option) (*svcsdk.Resource, error)
}

// GetRestAPIByID mocks GetRestAPIByID method
func (m *MockAPIGatewayClient) GetRestAPIByID(ctx context.Context, id *string) (*svcsdk.RestApi, error) {
	return m.MockGetRestAPIByID(ctx, id)
}

// GetResource mocks GetResource method
func (m *MockAPIGatewayClient) GetResource(ctx context.Context, in *svcsdk.GetResourceInput, opts ...request.Option) (*svcsdk.Resource, error) {
	return m.MockGetResource(ctx, in, opts...)
}

// GetRestAPIRootResource mocks GetRestAPIRootResource method
func (m *MockAPIGatewayClient) GetRestAPIRootResource(ctx context.Context, id *string) (*string, error) {
	return m.MockGetRestAPIRootResource(ctx, id)
}
