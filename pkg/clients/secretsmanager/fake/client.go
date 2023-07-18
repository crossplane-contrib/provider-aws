/*
Copyright 2023 The Crossplane Authors.

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
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/aws/aws-sdk-go/service/secretsmanager/secretsmanageriface"
)

// MockSecretsManagerClient mocks the secretsmanageriface.SecretsManagerAPI.
type MockSecretsManagerClient struct {
	secretsmanageriface.SecretsManagerAPI

	MockDescribeSecretWithContext    func(*secretsmanager.DescribeSecretInput) (*secretsmanager.DescribeSecretOutput, error)
	MockGetSecretValueWithContext    func(*secretsmanager.GetSecretValueInput) (*secretsmanager.GetSecretValueOutput, error)
	MockGetResourcePolicyWithContext func(*secretsmanager.GetResourcePolicyInput) (*secretsmanager.GetResourcePolicyOutput, error)
}

// DescribeSecretWithContext calls c.MockDescribeSecretWithContext
func (c *MockSecretsManagerClient) DescribeSecretWithContext(_ aws.Context, in *secretsmanager.DescribeSecretInput, _ ...request.Option) (*secretsmanager.DescribeSecretOutput, error) {
	return c.MockDescribeSecretWithContext(in)
}

// GetSecretValueWithContext calls c.MockGetSecretValueWithContext
func (c *MockSecretsManagerClient) GetSecretValueWithContext(_ aws.Context, in *secretsmanager.GetSecretValueInput, _ ...request.Option) (*secretsmanager.GetSecretValueOutput, error) {
	return c.MockGetSecretValueWithContext(in)
}

// GetResourcePolicyWithContext calls c.MockGetResourcePolicyWithContext
func (c *MockSecretsManagerClient) GetResourcePolicyWithContext(_ aws.Context, in *secretsmanager.GetResourcePolicyInput, _ ...request.Option) (*secretsmanager.GetResourcePolicyOutput, error) {
	return c.MockGetResourcePolicyWithContext(in)
}
