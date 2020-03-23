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
	"github.com/aws/aws-sdk-go-v2/service/rds"
)

// MockRDSClient for testing.
type MockRDSClient struct {
	MockCreate   func(*rds.CreateDBInstanceInput) rds.CreateDBInstanceRequest
	MockDescribe func(*rds.DescribeDBInstancesInput) rds.DescribeDBInstancesRequest
	MockModify   func(*rds.ModifyDBInstanceInput) rds.ModifyDBInstanceRequest
	MockDelete   func(*rds.DeleteDBInstanceInput) rds.DeleteDBInstanceRequest
	MockAddTags  func(*rds.AddTagsToResourceInput) rds.AddTagsToResourceRequest
}

// DescribeDBInstancesRequest finds RDS Instance by name
func (m *MockRDSClient) DescribeDBInstancesRequest(i *rds.DescribeDBInstancesInput) rds.DescribeDBInstancesRequest {
	return m.MockDescribe(i)
}

// CreateDBInstanceRequest creates RDS Instance with provided Specification
func (m *MockRDSClient) CreateDBInstanceRequest(i *rds.CreateDBInstanceInput) rds.CreateDBInstanceRequest {
	return m.MockCreate(i)
}

// ModifyDBInstanceRequest modifies RDS Instance with provided Specification
func (m *MockRDSClient) ModifyDBInstanceRequest(i *rds.ModifyDBInstanceInput) rds.ModifyDBInstanceRequest {
	return m.MockModify(i)
}

// DeleteDBInstanceRequest deletes RDS Instance
func (m *MockRDSClient) DeleteDBInstanceRequest(i *rds.DeleteDBInstanceInput) rds.DeleteDBInstanceRequest {
	return m.MockDelete(i)
}

// AddTagsToResourceRequest adds tags to RDS Instance.
func (m *MockRDSClient) AddTagsToResourceRequest(i *rds.AddTagsToResourceInput) rds.AddTagsToResourceRequest {
	return m.MockAddTags(i)
}
