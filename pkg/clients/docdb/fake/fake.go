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
	"github.com/aws/aws-sdk-go/service/docdb"
	"github.com/aws/aws-sdk-go/service/docdb/docdbiface"
)

// MockDocDBClient for testing
type MockDocDBClient struct {
	docdbiface.DocDBAPI

	MockListTagsForResource            func(*docdb.ListTagsForResourceInput) (*docdb.ListTagsForResourceOutput, error)
	MockAddTagsToResource              func(*docdb.AddTagsToResourceInput) (*docdb.AddTagsToResourceOutput, error)
	MockRemoveTagsFromResource         func(*docdb.RemoveTagsFromResourceInput) (*docdb.RemoveTagsFromResourceOutput, error)
	MockDescribeDBInstancesWithContext func(context.Context, *docdb.DescribeDBInstancesInput, []request.Option) (*docdb.DescribeDBInstancesOutput, error)
	MockCreateDBInstanceWithContext    func(context.Context, *docdb.CreateDBInstanceInput, []request.Option) (*docdb.CreateDBInstanceOutput, error)
	MockDeleteDBInstanceWithContext    func(context.Context, *docdb.DeleteDBInstanceInput, []request.Option) (*docdb.DeleteDBInstanceOutput, error)
	MockModifyDBInstanceWithContext    func(context.Context, *docdb.ModifyDBInstanceInput, []request.Option) (*docdb.ModifyDBInstanceOutput, error)

	MockDescribeDBSubnetGroupsWithContext func(context.Context, *docdb.DescribeDBSubnetGroupsInput, []request.Option) (*docdb.DescribeDBSubnetGroupsOutput, error)
	MockCreateDBSubnetGroupWithContext    func(context.Context, *docdb.CreateDBSubnetGroupInput, []request.Option) (*docdb.CreateDBSubnetGroupOutput, error)
	MockModifyDBSubnetGroupWithContext    func(context.Context, *docdb.ModifyDBSubnetGroupInput, []request.Option) (*docdb.ModifyDBSubnetGroupOutput, error)
	MockDeleteDBSubnetGroupWithContext    func(context.Context, *docdb.DeleteDBSubnetGroupInput, []request.Option) (*docdb.DeleteDBSubnetGroupOutput, error)

	MockDescribeDBClusterParameters                 func(*docdb.DescribeDBClusterParametersInput) (*docdb.DescribeDBClusterParametersOutput, error)
	MockDescribeDBClusterParametersWithContext      func(context.Context, *docdb.DescribeDBClusterParametersInput, []request.Option) (*docdb.DescribeDBClusterParametersOutput, error)
	MockDescribeDBClusterParameterGroupsWithContext func(context.Context, *docdb.DescribeDBClusterParameterGroupsInput, []request.Option) (*docdb.DescribeDBClusterParameterGroupsOutput, error)
	MockCreateDBClusterParameterGroupWithContext    func(context.Context, *docdb.CreateDBClusterParameterGroupInput, []request.Option) (*docdb.CreateDBClusterParameterGroupOutput, error)
	MockModifyDBClusterParameterGroupWithContext    func(context.Context, *docdb.ModifyDBClusterParameterGroupInput, []request.Option) (*docdb.ModifyDBClusterParameterGroupOutput, error)
	MockDeleteDBClusterParameterGroupWithContext    func(context.Context, *docdb.DeleteDBClusterParameterGroupInput, []request.Option) (*docdb.DeleteDBClusterParameterGroupOutput, error)

	MockDescribeDBClustersWithContext func(context.Context, *docdb.DescribeDBClustersInput, []request.Option) (*docdb.DescribeDBClustersOutput, error)
	MockCreateDBClusterWithContext    func(context.Context, *docdb.CreateDBClusterInput, []request.Option) (*docdb.CreateDBClusterOutput, error)
	MockModifyDBClusterWithContext    func(context.Context, *docdb.ModifyDBClusterInput, []request.Option) (*docdb.ModifyDBClusterOutput, error)
	MockDeleteDBClusterWithContext    func(context.Context, *docdb.DeleteDBClusterInput, []request.Option) (*docdb.DeleteDBClusterOutput, error)

	MockRestoreDBClusterFromSnapshotWithContext  func(context.Context, *docdb.RestoreDBClusterFromSnapshotInput, []request.Option) (*docdb.RestoreDBClusterFromSnapshotOutput, error)
	MockRestoreDBClusterToPointInTimeWithContext func(context.Context, *docdb.RestoreDBClusterToPointInTimeInput, []request.Option) (*docdb.RestoreDBClusterToPointInTimeOutput, error)

	Called MockDocDBClientCall
}

// CallListTagsForResource to log call
type CallListTagsForResource struct {
	I *docdb.ListTagsForResourceInput
}

// ListTagsForResource calls MockListTagsForResource
func (m *MockDocDBClient) ListTagsForResource(i *docdb.ListTagsForResourceInput) (*docdb.ListTagsForResourceOutput, error) {
	m.Called.ListTagsForResource = append(m.Called.ListTagsForResource, &CallListTagsForResource{I: i})

	return m.MockListTagsForResource(i)
}

// CallDescribeDBInstancesWithContext to log call
type CallDescribeDBInstancesWithContext struct {
	Ctx  aws.Context
	I    *docdb.DescribeDBInstancesInput
	Opts []request.Option
}

// DescribeDBInstancesWithContext calls MockDescribeDBInstancesWithContext
func (m *MockDocDBClient) DescribeDBInstancesWithContext(ctx aws.Context, i *docdb.DescribeDBInstancesInput, opts ...request.Option) (*docdb.DescribeDBInstancesOutput, error) {
	m.Called.DescribeDBInstancesWithContext = append(m.Called.DescribeDBInstancesWithContext, &CallDescribeDBInstancesWithContext{Ctx: ctx, I: i, Opts: opts})

	return m.MockDescribeDBInstancesWithContext(ctx, i, opts)
}

// CallCreateDBInstanceWithContext to log call
type CallCreateDBInstanceWithContext struct {
	Ctx  aws.Context
	I    *docdb.CreateDBInstanceInput
	Opts []request.Option
}

// CreateDBInstanceWithContext calls MockCreateDBInstanceWithContext
func (m *MockDocDBClient) CreateDBInstanceWithContext(ctx aws.Context, i *docdb.CreateDBInstanceInput, opts ...request.Option) (*docdb.CreateDBInstanceOutput, error) {
	m.Called.CreateDBInstanceWithContext = append(m.Called.CreateDBInstanceWithContext, &CallCreateDBInstanceWithContext{Ctx: ctx, I: i, Opts: opts})

	return m.MockCreateDBInstanceWithContext(ctx, i, opts)
}

// CallDeleteDBInstanceWithContext to log call
type CallDeleteDBInstanceWithContext struct {
	Ctx  aws.Context
	I    *docdb.DeleteDBInstanceInput
	Opts []request.Option
}

// DeleteDBInstanceWithContext calls MockDeleteDBInstanceWithContext
func (m *MockDocDBClient) DeleteDBInstanceWithContext(ctx aws.Context, i *docdb.DeleteDBInstanceInput, opts ...request.Option) (*docdb.DeleteDBInstanceOutput, error) {
	m.Called.DeleteDBInstanceWithContext = append(m.Called.DeleteDBInstanceWithContext, &CallDeleteDBInstanceWithContext{Ctx: ctx, I: i, Opts: opts})

	return m.MockDeleteDBInstanceWithContext(ctx, i, opts)
}

// CallModifyDBInstanceWithContext to log call
type CallModifyDBInstanceWithContext struct {
	Ctx  aws.Context
	I    *docdb.ModifyDBInstanceInput
	Opts []request.Option
}

// ModifyDBInstanceWithContext calls MockModifyDBInstanceWithContext
func (m *MockDocDBClient) ModifyDBInstanceWithContext(ctx aws.Context, i *docdb.ModifyDBInstanceInput, opts ...request.Option) (*docdb.ModifyDBInstanceOutput, error) {
	m.Called.ModifyDBInstanceWithContext = append(m.Called.ModifyDBInstanceWithContext, &CallModifyDBInstanceWithContext{Ctx: ctx, I: i, Opts: opts})

	return m.MockModifyDBInstanceWithContext(ctx, i, opts)
}

// CallAddTagsToResource to log call
type CallAddTagsToResource struct {
	I *docdb.AddTagsToResourceInput
}

// AddTagsToResource calls MockAddTagsToResource
func (m *MockDocDBClient) AddTagsToResource(i *docdb.AddTagsToResourceInput) (*docdb.AddTagsToResourceOutput, error) {
	m.Called.AddTagsToResource = append(m.Called.AddTagsToResource, &CallAddTagsToResource{I: i})

	return m.MockAddTagsToResource(i)
}

// CallRemoveTagsFromResource to log call
type CallRemoveTagsFromResource struct {
	I *docdb.RemoveTagsFromResourceInput
}

// RemoveTagsFromResource calls MockRemoveTagsFromResource
func (m *MockDocDBClient) RemoveTagsFromResource(i *docdb.RemoveTagsFromResourceInput) (*docdb.RemoveTagsFromResourceOutput, error) {
	m.Called.RemoveTagsFromResource = append(m.Called.RemoveTagsFromResource, &CallRemoveTagsFromResource{I: i})

	return m.MockRemoveTagsFromResource(i)
}

// CallDescribeDBSubnetGroupsWithContext to log call
type CallDescribeDBSubnetGroupsWithContext struct {
	Ctx  aws.Context
	I    *docdb.DescribeDBSubnetGroupsInput
	Opts []request.Option
}

// DescribeDBSubnetGroupsWithContext calls MockDescribeDBSubnetGroupsWithContext
func (m *MockDocDBClient) DescribeDBSubnetGroupsWithContext(ctx context.Context, i *docdb.DescribeDBSubnetGroupsInput, opts ...request.Option) (*docdb.DescribeDBSubnetGroupsOutput, error) {
	m.Called.DescribeDBSubnetGroupsWithContext = append(m.Called.DescribeDBSubnetGroupsWithContext, &CallDescribeDBSubnetGroupsWithContext{Ctx: ctx, I: i, Opts: opts})

	return m.MockDescribeDBSubnetGroupsWithContext(ctx, i, opts)
}

// CallCreateDBSubnetGroupWithContext to log call
type CallCreateDBSubnetGroupWithContext struct {
	Ctx  aws.Context
	I    *docdb.CreateDBSubnetGroupInput
	Opts []request.Option
}

// CreateDBSubnetGroupWithContext calls MockCreateDBSubnetGroupWithContext
func (m *MockDocDBClient) CreateDBSubnetGroupWithContext(ctx context.Context, i *docdb.CreateDBSubnetGroupInput, opts ...request.Option) (*docdb.CreateDBSubnetGroupOutput, error) {
	m.Called.CreateDBSubnetGroupWithContext = append(m.Called.CreateDBSubnetGroupWithContext, &CallCreateDBSubnetGroupWithContext{Ctx: ctx, I: i, Opts: opts})

	return m.MockCreateDBSubnetGroupWithContext(ctx, i, opts)
}

// CallModifyDBSubnetGroupWithContext to log call
type CallModifyDBSubnetGroupWithContext struct {
	Ctx  aws.Context
	I    *docdb.ModifyDBSubnetGroupInput
	Opts []request.Option
}

// ModifyDBSubnetGroupWithContext calls MockModifyDBSubnetGroupWithContext
func (m *MockDocDBClient) ModifyDBSubnetGroupWithContext(ctx context.Context, i *docdb.ModifyDBSubnetGroupInput, opts ...request.Option) (*docdb.ModifyDBSubnetGroupOutput, error) {
	m.Called.ModifyDBSubnetGroupWithContext = append(m.Called.ModifyDBSubnetGroupWithContext, &CallModifyDBSubnetGroupWithContext{Ctx: ctx, I: i, Opts: opts})

	return m.MockModifyDBSubnetGroupWithContext(ctx, i, opts)
}

// CallDeleteDBSubnetGroupWithContext to log call
type CallDeleteDBSubnetGroupWithContext struct {
	Ctx  aws.Context
	I    *docdb.DeleteDBSubnetGroupInput
	Opts []request.Option
}

// DeleteDBSubnetGroupWithContext calls MockDeleteDBSubnetGroupWithContext
func (m *MockDocDBClient) DeleteDBSubnetGroupWithContext(ctx context.Context, i *docdb.DeleteDBSubnetGroupInput, opts ...request.Option) (*docdb.DeleteDBSubnetGroupOutput, error) {
	m.Called.DeleteDBSubnetGroupWithContext = append(m.Called.DeleteDBSubnetGroupWithContext, &CallDeleteDBSubnetGroupWithContext{Ctx: ctx, I: i, Opts: opts})

	return m.MockDeleteDBSubnetGroupWithContext(ctx, i, opts)
}

// CallDescribeDBClusterParameters to log call
type CallDescribeDBClusterParameters struct {
	I *docdb.DescribeDBClusterParametersInput
}

// DescribeDBClusterParameters calls MockDescribeDBClusterParameters
func (m *MockDocDBClient) DescribeDBClusterParameters(i *docdb.DescribeDBClusterParametersInput) (*docdb.DescribeDBClusterParametersOutput, error) {
	m.Called.DescribeDBClusterParameters = append(m.Called.DescribeDBClusterParameters, &CallDescribeDBClusterParameters{I: i})

	return m.MockDescribeDBClusterParameters(i)
}

// CallDescribeDBClusterParametersWithContext to log call
type CallDescribeDBClusterParametersWithContext struct {
	Ctx  aws.Context
	I    *docdb.DescribeDBClusterParametersInput
	Opts []request.Option
}

// DescribeDBClusterParametersWithContext calls MockDescribeDBClusterParametersWithContext
func (m *MockDocDBClient) DescribeDBClusterParametersWithContext(ctx context.Context, i *docdb.DescribeDBClusterParametersInput, opts ...request.Option) (*docdb.DescribeDBClusterParametersOutput, error) {
	m.Called.DescribeDBClusterParametersWithContext = append(m.Called.DescribeDBClusterParametersWithContext, &CallDescribeDBClusterParametersWithContext{I: i, Ctx: ctx, Opts: opts})

	return m.MockDescribeDBClusterParametersWithContext(ctx, i, opts)
}

// CallDescribeDBClusterParameterGroupsWithContext to log call
type CallDescribeDBClusterParameterGroupsWithContext struct {
	Ctx  aws.Context
	I    *docdb.DescribeDBClusterParameterGroupsInput
	Opts []request.Option
}

// DescribeDBClusterParameterGroupsWithContext calls MockDescribeDBClusterParameterGroupsWithContext
func (m *MockDocDBClient) DescribeDBClusterParameterGroupsWithContext(ctx context.Context, i *docdb.DescribeDBClusterParameterGroupsInput, opts ...request.Option) (*docdb.DescribeDBClusterParameterGroupsOutput, error) {
	m.Called.DescribeDBClusterParameterGroupsWithContext = append(m.Called.DescribeDBClusterParameterGroupsWithContext, &CallDescribeDBClusterParameterGroupsWithContext{Ctx: ctx, I: i, Opts: opts})

	return m.MockDescribeDBClusterParameterGroupsWithContext(ctx, i, opts)
}

// CallCreateDBClusterParameterGroupWithContext to log call
type CallCreateDBClusterParameterGroupWithContext struct {
	Ctx  aws.Context
	I    *docdb.CreateDBClusterParameterGroupInput
	Opts []request.Option
}

// CreateDBClusterParameterGroupWithContext calls MockCreateDBClusterParameterGroupWithContext
func (m *MockDocDBClient) CreateDBClusterParameterGroupWithContext(ctx context.Context, i *docdb.CreateDBClusterParameterGroupInput, opts ...request.Option) (*docdb.CreateDBClusterParameterGroupOutput, error) {
	m.Called.CreateDBClusterParameterGroupWithContext = append(m.Called.CreateDBClusterParameterGroupWithContext, &CallCreateDBClusterParameterGroupWithContext{Ctx: ctx, I: i, Opts: opts})

	return m.MockCreateDBClusterParameterGroupWithContext(ctx, i, opts)
}

// CallModifyDBClusterParameterGroupWithContext to log call
type CallModifyDBClusterParameterGroupWithContext struct {
	Ctx  aws.Context
	I    *docdb.ModifyDBClusterParameterGroupInput
	Opts []request.Option
}

// ModifyDBClusterParameterGroupWithContext calls MockModifyDBClusterParameterGroupWithContext
func (m *MockDocDBClient) ModifyDBClusterParameterGroupWithContext(ctx context.Context, i *docdb.ModifyDBClusterParameterGroupInput, opts ...request.Option) (*docdb.ModifyDBClusterParameterGroupOutput, error) {
	m.Called.ModifyDBClusterParameterGroupWithContext = append(m.Called.ModifyDBClusterParameterGroupWithContext, &CallModifyDBClusterParameterGroupWithContext{Ctx: ctx, I: i, Opts: opts})

	return m.MockModifyDBClusterParameterGroupWithContext(ctx, i, opts)
}

// CallDeleteDBClusterParameterGroupWithContext to log call
type CallDeleteDBClusterParameterGroupWithContext struct {
	Ctx  aws.Context
	I    *docdb.DeleteDBClusterParameterGroupInput
	Opts []request.Option
}

// DeleteDBClusterParameterGroupWithContext calls MockDeleteDBClusterParameterGroupWithContext
func (m *MockDocDBClient) DeleteDBClusterParameterGroupWithContext(ctx context.Context, i *docdb.DeleteDBClusterParameterGroupInput, opts ...request.Option) (*docdb.DeleteDBClusterParameterGroupOutput, error) {
	m.Called.DeleteDBClusterParameterGroupWithContext = append(m.Called.DeleteDBClusterParameterGroupWithContext, &CallDeleteDBClusterParameterGroupWithContext{Ctx: ctx, I: i, Opts: opts})

	return m.MockDeleteDBClusterParameterGroupWithContext(ctx, i, opts)
}

// CallDescribeDBClustersWithContext to log call
type CallDescribeDBClustersWithContext struct {
	Ctx  aws.Context
	I    *docdb.DescribeDBClustersInput
	Opts []request.Option
}

// DescribeDBClustersWithContext calls MockDescribeDBClustersWithContext
func (m *MockDocDBClient) DescribeDBClustersWithContext(ctx context.Context, i *docdb.DescribeDBClustersInput, opts ...request.Option) (*docdb.DescribeDBClustersOutput, error) {
	m.Called.DescribeDBClustersWithContext = append(m.Called.DescribeDBClustersWithContext, &CallDescribeDBClustersWithContext{Ctx: ctx, I: i, Opts: opts})

	return m.MockDescribeDBClustersWithContext(ctx, i, opts)
}

// CallCreateDBClusterWithContext to log call
type CallCreateDBClusterWithContext struct {
	Ctx  aws.Context
	I    *docdb.CreateDBClusterInput
	Opts []request.Option
}

// CreateDBClusterWithContext calls MockCreateDBClusterWithContext
func (m *MockDocDBClient) CreateDBClusterWithContext(ctx context.Context, i *docdb.CreateDBClusterInput, opts ...request.Option) (*docdb.CreateDBClusterOutput, error) {
	m.Called.CreateDBClusterWithContext = append(m.Called.CreateDBClusterWithContext, &CallCreateDBClusterWithContext{Ctx: ctx, I: i, Opts: opts})

	return m.MockCreateDBClusterWithContext(ctx, i, opts)
}

// CallModifyDBClusterWithContext to log call
type CallModifyDBClusterWithContext struct {
	Ctx  aws.Context
	I    *docdb.ModifyDBClusterInput
	Opts []request.Option
}

// ModifyDBClusterWithContext calls MockModifyDBClusterWithContext
func (m *MockDocDBClient) ModifyDBClusterWithContext(ctx context.Context, i *docdb.ModifyDBClusterInput, opts ...request.Option) (*docdb.ModifyDBClusterOutput, error) {
	m.Called.ModifyDBClusterWithContext = append(m.Called.ModifyDBClusterWithContext, &CallModifyDBClusterWithContext{Ctx: ctx, I: i, Opts: opts})

	return m.MockModifyDBClusterWithContext(ctx, i, opts)
}

// CallDeleteDBClusterWithContext to log call
type CallDeleteDBClusterWithContext struct {
	Ctx  aws.Context
	I    *docdb.DeleteDBClusterInput
	Opts []request.Option
}

// DeleteDBClusterWithContext calls MockDeleteDBClusterWithContext
func (m *MockDocDBClient) DeleteDBClusterWithContext(ctx context.Context, i *docdb.DeleteDBClusterInput, opts ...request.Option) (*docdb.DeleteDBClusterOutput, error) {
	m.Called.DeleteDBClusterWithContext = append(m.Called.DeleteDBClusterWithContext, &CallDeleteDBClusterWithContext{Ctx: ctx, I: i, Opts: opts})

	return m.MockDeleteDBClusterWithContext(ctx, i, opts)
}

// CallRestoreDBClusterFromSnapshotWithContext to log call
type CallRestoreDBClusterFromSnapshotWithContext struct {
	Ctx  aws.Context
	I    *docdb.RestoreDBClusterFromSnapshotInput
	Opts []request.Option
}

// RestoreDBClusterFromSnapshotWithContext calls MockRestoreDBClusterFromSnapshotWithContext
func (m *MockDocDBClient) RestoreDBClusterFromSnapshotWithContext(ctx context.Context, i *docdb.RestoreDBClusterFromSnapshotInput, opts ...request.Option) (*docdb.RestoreDBClusterFromSnapshotOutput, error) {
	m.Called.RestoreDBClusterFromSnapshotWithContext = append(m.Called.RestoreDBClusterFromSnapshotWithContext, &CallRestoreDBClusterFromSnapshotWithContext{Ctx: ctx, I: i, Opts: opts})

	return m.MockRestoreDBClusterFromSnapshotWithContext(ctx, i, opts)
}

// CallRestoreDBClusterToPointInTimeWithContext to log call
type CallRestoreDBClusterToPointInTimeWithContext struct {
	Ctx  aws.Context
	I    *docdb.RestoreDBClusterToPointInTimeInput
	Opts []request.Option
}

// RestoreDBClusterToPointInTimeWithContext calls MockRestoreDBClusterToPointInTimeWithContext
func (m *MockDocDBClient) RestoreDBClusterToPointInTimeWithContext(ctx context.Context, i *docdb.RestoreDBClusterToPointInTimeInput, opts ...request.Option) (*docdb.RestoreDBClusterToPointInTimeOutput, error) {
	m.Called.RestoreDBClusterToPointInTimeWithContext = append(m.Called.RestoreDBClusterToPointInTimeWithContext, &CallRestoreDBClusterToPointInTimeWithContext{Ctx: ctx, I: i, Opts: opts})

	return m.MockRestoreDBClusterToPointInTimeWithContext(ctx, i, opts)
}

// MockDocDBClientCall to log calls
type MockDocDBClientCall struct {
	ListTagsForResource            []*CallListTagsForResource
	AddTagsToResource              []*CallAddTagsToResource
	RemoveTagsFromResource         []*CallRemoveTagsFromResource
	DescribeDBInstancesWithContext []*CallDescribeDBInstancesWithContext
	CreateDBInstanceWithContext    []*CallCreateDBInstanceWithContext
	ModifyDBInstanceWithContext    []*CallModifyDBInstanceWithContext
	DeleteDBInstanceWithContext    []*CallDeleteDBInstanceWithContext

	DescribeDBSubnetGroupsWithContext []*CallDescribeDBSubnetGroupsWithContext
	CreateDBSubnetGroupWithContext    []*CallCreateDBSubnetGroupWithContext
	ModifyDBSubnetGroupWithContext    []*CallModifyDBSubnetGroupWithContext
	DeleteDBSubnetGroupWithContext    []*CallDeleteDBSubnetGroupWithContext

	DescribeDBClusterParameters                 []*CallDescribeDBClusterParameters
	DescribeDBClusterParametersWithContext      []*CallDescribeDBClusterParametersWithContext
	DescribeDBClusterParameterGroupsWithContext []*CallDescribeDBClusterParameterGroupsWithContext
	CreateDBClusterParameterGroupWithContext    []*CallCreateDBClusterParameterGroupWithContext
	ModifyDBClusterParameterGroupWithContext    []*CallModifyDBClusterParameterGroupWithContext
	DeleteDBClusterParameterGroupWithContext    []*CallDeleteDBClusterParameterGroupWithContext

	DescribeDBClustersWithContext []*CallDescribeDBClustersWithContext
	CreateDBClusterWithContext    []*CallCreateDBClusterWithContext
	ModifyDBClusterWithContext    []*CallModifyDBClusterWithContext
	DeleteDBClusterWithContext    []*CallDeleteDBClusterWithContext

	RestoreDBClusterFromSnapshotWithContext  []*CallRestoreDBClusterFromSnapshotWithContext
	RestoreDBClusterToPointInTimeWithContext []*CallRestoreDBClusterToPointInTimeWithContext
}
