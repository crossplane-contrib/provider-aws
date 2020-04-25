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
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

// MockDynamoClient for testing.
type MockDynamoClient struct {
	MockDescribe func(input *dynamodb.DescribeTableInput) dynamodb.DescribeTableRequest
	MockCreate   func(input *dynamodb.CreateTableInput) dynamodb.CreateTableRequest
	MockDelete   func(input *dynamodb.DeleteTableInput) dynamodb.DeleteTableRequest
	MockUpdate   func(input *dynamodb.UpdateTableInput) dynamodb.UpdateTableRequest
}

// DescribeTableRequest finds DynamoDB Table by name
func (m *MockDynamoClient) DescribeTableRequest(i *dynamodb.DescribeTableInput) dynamodb.DescribeTableRequest {
	return m.MockDescribe(i)
}

// CreateTableRequest creates DynamoDB Table with provided Specification
func (m *MockDynamoClient) CreateTableRequest(i *dynamodb.CreateTableInput) dynamodb.CreateTableRequest {
	return m.MockCreate(i)
}

// DeleteTableRequest modifies DynamoDB Table with provided Specification
func (m *MockDynamoClient) DeleteTableRequest(i *dynamodb.DeleteTableInput) dynamodb.DeleteTableRequest {
	return m.MockDelete(i)
}

// UpdateTableRequest deletes DynamoDB Table
func (m *MockDynamoClient) UpdateTableRequest(i *dynamodb.UpdateTableInput) dynamodb.UpdateTableRequest {
	return m.MockUpdate(i)
}
