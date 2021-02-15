/*
Copyright 2020 The Crossplane Authors.

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
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager/secretsmanageriface"
)

var _ secretsmanageriface.ClientAPI = &MockClient{}

// MockClient is a fake implementation of secretsmanager.Client.
type MockClient struct {
	secretsmanageriface.ClientAPI

	MockDescribeSecretRequest func(*secretsmanager.DescribeSecretInput) secretsmanager.DescribeSecretRequest
	MockGetSecretValueRequest func(*secretsmanager.GetSecretValueInput) secretsmanager.GetSecretValueRequest
	MockCreateSecretRequest   func(*secretsmanager.CreateSecretInput) secretsmanager.CreateSecretRequest
	MockDeleteSecretRequest   func(*secretsmanager.DeleteSecretInput) secretsmanager.DeleteSecretRequest
	MockUpdateSecretRequest   func(*secretsmanager.UpdateSecretInput) secretsmanager.UpdateSecretRequest
	MockTagResourceRequest    func(*secretsmanager.TagResourceInput) secretsmanager.TagResourceRequest
	MockUntagResourceRequest  func(*secretsmanager.UntagResourceInput) secretsmanager.UntagResourceRequest
}

// DescribeSecretRequest calls the underlying MockDescribeSecretRequest method.
func (c *MockClient) DescribeSecretRequest(i *secretsmanager.DescribeSecretInput) secretsmanager.DescribeSecretRequest {
	return c.MockDescribeSecretRequest(i)
}

// GetSecretValueRequest calls the underlying MockGetSecretValueRequest method.
func (c *MockClient) GetSecretValueRequest(i *secretsmanager.GetSecretValueInput) secretsmanager.GetSecretValueRequest {
	return c.MockGetSecretValueRequest(i)
}

// CreateSecretRequest calls the underlying MockCreateSecretRequest method.
func (c *MockClient) CreateSecretRequest(i *secretsmanager.CreateSecretInput) secretsmanager.CreateSecretRequest {
	return c.MockCreateSecretRequest(i)
}

// DeleteSecretRequest calls the underlying MockDeleteSecretRequest method.
func (c *MockClient) DeleteSecretRequest(i *secretsmanager.DeleteSecretInput) secretsmanager.DeleteSecretRequest {
	return c.MockDeleteSecretRequest(i)
}

// UpdateSecretRequest calls the underlying MockUpdateSecretRequest method.
func (c *MockClient) UpdateSecretRequest(i *secretsmanager.UpdateSecretInput) secretsmanager.UpdateSecretRequest {
	return c.MockUpdateSecretRequest(i)
}

// TagResourceRequest calls the underlying MockTagResourceRequest method.
func (c *MockClient) TagResourceRequest(i *secretsmanager.TagResourceInput) secretsmanager.TagResourceRequest {
	return c.MockTagResourceRequest(i)
}

// UntagResourceRequest calls the underlying UntagResourceRequest method.
func (c *MockClient) UntagResourceRequest(i *secretsmanager.UntagResourceInput) secretsmanager.UntagResourceRequest {
	return c.MockUntagResourceRequest(i)
}
