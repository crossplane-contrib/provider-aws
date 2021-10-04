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

	"github.com/aws/aws-sdk-go-v2/service/rds"
)

// MockRDSClient for testing.
type MockRDSClient struct {
	MockCreate   func(context.Context, *rds.CreateDBInstanceInput, []func(*rds.Options)) (*rds.CreateDBInstanceOutput, error)
	MockDescribe func(context.Context, *rds.DescribeDBInstancesInput, []func(*rds.Options)) (*rds.DescribeDBInstancesOutput, error)
	MockModify   func(context.Context, *rds.ModifyDBInstanceInput, []func(*rds.Options)) (*rds.ModifyDBInstanceOutput, error)
	MockDelete   func(context.Context, *rds.DeleteDBInstanceInput, []func(*rds.Options)) (*rds.DeleteDBInstanceOutput, error)
	MockAddTags  func(context.Context, *rds.AddTagsToResourceInput, []func(*rds.Options)) (*rds.AddTagsToResourceOutput, error)
}

// DescribeDBInstances finds RDS Instance by name
func (m *MockRDSClient) DescribeDBInstances(ctx context.Context, i *rds.DescribeDBInstancesInput, opts ...func(*rds.Options)) (*rds.DescribeDBInstancesOutput, error) {
	return m.MockDescribe(ctx, i, opts)
}

// CreateDBInstance creates RDS Instance with provided Specification
func (m *MockRDSClient) CreateDBInstance(ctx context.Context, i *rds.CreateDBInstanceInput, opts ...func(*rds.Options)) (*rds.CreateDBInstanceOutput, error) {
	return m.MockCreate(ctx, i, opts)
}

// ModifyDBInstance modifies RDS Instance with provided Specification
func (m *MockRDSClient) ModifyDBInstance(ctx context.Context, i *rds.ModifyDBInstanceInput, opts ...func(*rds.Options)) (*rds.ModifyDBInstanceOutput, error) {
	return m.MockModify(ctx, i, opts)
}

// DeleteDBInstance deletes RDS Instance
func (m *MockRDSClient) DeleteDBInstance(ctx context.Context, i *rds.DeleteDBInstanceInput, opts ...func(*rds.Options)) (*rds.DeleteDBInstanceOutput, error) {
	return m.MockDelete(ctx, i, opts)
}

// AddTagsToResource adds tags to RDS Instance.
func (m *MockRDSClient) AddTagsToResource(ctx context.Context, i *rds.AddTagsToResourceInput, opts ...func(*rds.Options)) (*rds.AddTagsToResourceOutput, error) {
	return m.MockAddTags(ctx, i, opts)
}
