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
	"github.com/aws/aws-sdk-go/service/dax"
	"github.com/aws/aws-sdk-go/service/dax/daxiface"
)

// MockDaxClient for testing
type MockDaxClient struct {
	daxiface.DAXAPI
	MockCreateClusterWithContext func(context.Context, *dax.CreateClusterInput, []request.Option) (*dax.CreateClusterOutput, error)

	MockCreateParameterGroupWithContext func(context.Context, *dax.CreateParameterGroupInput, []request.Option) (*dax.CreateParameterGroupOutput, error)

	MockCreateSubnetGroupWithContext func(context.Context, *dax.CreateSubnetGroupInput, []request.Option) (*dax.CreateSubnetGroupOutput, error)

	MockDeleteClusterWithContext func(context.Context, *dax.DeleteClusterInput, []request.Option) (*dax.DeleteClusterOutput, error)

	MockDeleteParameterGroupWithContext func(context.Context, *dax.DeleteParameterGroupInput, []request.Option) (*dax.DeleteParameterGroupOutput, error)

	MockDeleteSubnetGroupWithContext func(context.Context, *dax.DeleteSubnetGroupInput, []request.Option) (*dax.DeleteSubnetGroupOutput, error)

	MockDescribeClusters            func(*dax.DescribeClustersInput) (*dax.DescribeClustersOutput, error)
	MockDescribeClustersWithContext func(context.Context, *dax.DescribeClustersInput, []request.Option) (*dax.DescribeClustersOutput, error)

	MockDescribeParameterGroups            func(*dax.DescribeParameterGroupsInput) (*dax.DescribeParameterGroupsOutput, error)
	MockDescribeParameterGroupsWithContext func(context.Context, *dax.DescribeParameterGroupsInput, []request.Option) (*dax.DescribeParameterGroupsOutput, error)

	MockDescribeParameters            func(*dax.DescribeParametersInput) (*dax.DescribeParametersOutput, error)
	MockDescribeParametersWithContext func(context.Context, *dax.DescribeParametersInput, []request.Option) (*dax.DescribeParametersOutput, error)

	MockDescribeSubnetGroups            func(*dax.DescribeSubnetGroupsInput) (*dax.DescribeSubnetGroupsOutput, error)
	MockDescribeSubnetGroupsWithContext func(context.Context, *dax.DescribeSubnetGroupsInput, []request.Option) (*dax.DescribeSubnetGroupsOutput, error)

	MockUpdateCluster            func(*dax.UpdateClusterInput) (*dax.UpdateClusterOutput, error)
	MockUpdateClusterWithContext func(context.Context, *dax.UpdateClusterInput, []request.Option) (*dax.UpdateClusterOutput, error)

	MockUpdateParameterGroup            func(*dax.UpdateParameterGroupInput) (*dax.UpdateParameterGroupOutput, error)
	MockUpdateParameterGroupWithContext func(context.Context, *dax.UpdateParameterGroupInput, []request.Option) (*dax.UpdateParameterGroupOutput, error)

	MockUpdateSubnetGroup            func(*dax.UpdateSubnetGroupInput) (*dax.UpdateSubnetGroupOutput, error)
	MockUpdateSubnetGroupWithContext func(context.Context, *dax.UpdateSubnetGroupInput, []request.Option) (*dax.UpdateSubnetGroupOutput, error)

	Called MockDaxClientCall
}

// CallDescribeParameterGroupsWithContext to log calls
type CallDescribeParameterGroupsWithContext struct {
	Ctx  aws.Context
	I    *dax.DescribeParameterGroupsInput
	Opts []request.Option
}

// DescribeParameterGroupsWithContext mocks DescribeParameterGroupsWithContext method
func (m *MockDaxClient) DescribeParameterGroupsWithContext(ctx aws.Context, i *dax.DescribeParameterGroupsInput, opts ...request.Option) (*dax.DescribeParameterGroupsOutput, error) {
	m.Called.DescribeParameterGroupsWithContext = append(m.Called.DescribeParameterGroupsWithContext, &CallDescribeParameterGroupsWithContext{Ctx: ctx, I: i, Opts: opts})

	return m.MockDescribeParameterGroupsWithContext(ctx, i, opts)
}

// CallDescribeParameterGroups to log calls
type CallDescribeParameterGroups struct {
	I *dax.DescribeParameterGroupsInput
}

// DescribeParameterGroups mocks DescribeParameterGroups method
func (m *MockDaxClient) DescribeParameterGroups(i *dax.DescribeParameterGroupsInput) (*dax.DescribeParameterGroupsOutput, error) {
	m.Called.DescribeParameterGroups = append(m.Called.DescribeParameterGroups, &CallDescribeParameterGroups{I: i})
	return m.MockDescribeParameterGroups(i)
}

// CallUpdateParameterGroupsWithContext to log calls
type CallUpdateParameterGroupsWithContext struct {
	Ctx  aws.Context
	I    *dax.UpdateParameterGroupInput
	Opts []request.Option
}

// UpdateParameterGroupWithContext mocks UpdateParameterGroupWithContext method
func (m *MockDaxClient) UpdateParameterGroupWithContext(ctx aws.Context, i *dax.UpdateParameterGroupInput, opts ...request.Option) (*dax.UpdateParameterGroupOutput, error) {
	m.Called.UpdateParameterGroupsWithContext = append(m.Called.UpdateParameterGroupsWithContext, &CallUpdateParameterGroupsWithContext{Ctx: ctx, I: i, Opts: opts})

	return m.MockUpdateParameterGroupWithContext(ctx, i, opts)
}

// CallUpdateParameterGroups to log calls
type CallUpdateParameterGroups struct {
	I *dax.UpdateParameterGroupInput
}

// UpdateParameterGroups mocks UpdateParameterGroups method
func (m *MockDaxClient) UpdateParameterGroups(i *dax.UpdateParameterGroupInput) (*dax.UpdateParameterGroupOutput, error) {
	m.Called.UpdateParameterGroups = append(m.Called.UpdateParameterGroups, &CallUpdateParameterGroups{I: i})
	return m.MockUpdateParameterGroup(i)
}

// CallDescribeParameters to log calls
type CallDescribeParameters struct {
	I *dax.DescribeParametersInput
}

// DescribeParameters mocks DescribeParameters method
func (m *MockDaxClient) DescribeParameters(i *dax.DescribeParametersInput) (*dax.DescribeParametersOutput, error) {
	m.Called.DescribeParameters = append(m.Called.DescribeParameters, &CallDescribeParameters{I: i})

	return m.MockDescribeParameters(i)
}

// CallDescribeParametersWithContext to log calls
type CallDescribeParametersWithContext struct {
	Ctx  context.Context
	I    *dax.DescribeParametersInput
	Opts []request.Option
}

// DescribeParametersWithContext mocks DescribeParametersWithContext method
func (m *MockDaxClient) DescribeParametersWithContext(ctx context.Context, i *dax.DescribeParametersInput, opts ...request.Option) (*dax.DescribeParametersOutput, error) {
	m.Called.DescribeParametersWithContext = append(m.Called.DescribeParametersWithContext, &CallDescribeParametersWithContext{Ctx: ctx, I: i})

	return m.MockDescribeParametersWithContext(ctx, i, opts)
}

// CallCreateParameterGroupWithContext to log calls
type CallCreateParameterGroupWithContext struct {
	Ctx  aws.Context
	I    *dax.CreateParameterGroupInput
	Opts []request.Option
}

// CreateParameterGroupWithContext mocks CreateParameterGroupWithContext method
func (m *MockDaxClient) CreateParameterGroupWithContext(ctx aws.Context, i *dax.CreateParameterGroupInput, opts ...request.Option) (*dax.CreateParameterGroupOutput, error) {
	m.Called.CreateParameterGroupWithContext = append(m.Called.CreateParameterGroupWithContext, &CallCreateParameterGroupWithContext{Ctx: ctx, I: i, Opts: opts})

	return m.MockCreateParameterGroupWithContext(ctx, i, opts)
}

// CallDeleteParameterGroupWithContext to log calls
type CallDeleteParameterGroupWithContext struct {
	Ctx  aws.Context
	I    *dax.DeleteParameterGroupInput
	Opts []request.Option
}

// DeleteParameterGroupWithContext mocks DeleteParameterGroupWithContext method
func (m *MockDaxClient) DeleteParameterGroupWithContext(ctx aws.Context, i *dax.DeleteParameterGroupInput, opts ...request.Option) (*dax.DeleteParameterGroupOutput, error) {
	m.Called.DeleteParameterGroupWithContext = append(m.Called.DeleteParameterGroupWithContext, &CallDeleteParameterGroupWithContext{Ctx: ctx, I: i, Opts: opts})

	return m.MockDeleteParameterGroupWithContext(ctx, i, opts)
}

// CallDescribeSubnetGroupsWithContext to log calls
type CallDescribeSubnetGroupsWithContext struct {
	Ctx  aws.Context
	I    *dax.DescribeSubnetGroupsInput
	Opts []request.Option
}

// DescribeSubnetGroupsWithContext mocks DescribeSubnetGroupsWithContext method
func (m *MockDaxClient) DescribeSubnetGroupsWithContext(ctx aws.Context, i *dax.DescribeSubnetGroupsInput, opts ...request.Option) (*dax.DescribeSubnetGroupsOutput, error) {
	m.Called.DescribeSubnetGroupsWithContext = append(m.Called.DescribeSubnetGroupsWithContext, &CallDescribeSubnetGroupsWithContext{Ctx: ctx, I: i, Opts: opts})

	return m.MockDescribeSubnetGroupsWithContext(ctx, i, opts)
}

// CallDescribeSubnetGroups to log calls
type CallDescribeSubnetGroups struct {
	I *dax.DescribeSubnetGroupsInput
}

// DescribeSubnetGroups mocks DescribeSubnetGroups method
func (m *MockDaxClient) DescribeSubnetGroups(i *dax.DescribeSubnetGroupsInput) (*dax.DescribeSubnetGroupsOutput, error) {
	m.Called.DescribeSubnetGroups = append(m.Called.DescribeSubnetGroups, &CallDescribeSubnetGroups{I: i})
	return m.MockDescribeSubnetGroups(i)
}

// CallUpdateSubnetGroupsWithContext to log calls
type CallUpdateSubnetGroupsWithContext struct {
	Ctx  aws.Context
	I    *dax.UpdateSubnetGroupInput
	Opts []request.Option
}

// UpdateSubnetGroupWithContext mocks UpdateSubnetGroupWithContext method
func (m *MockDaxClient) UpdateSubnetGroupWithContext(ctx aws.Context, i *dax.UpdateSubnetGroupInput, opts ...request.Option) (*dax.UpdateSubnetGroupOutput, error) {
	m.Called.UpdateSubnetGroupsWithContext = append(m.Called.UpdateSubnetGroupsWithContext, &CallUpdateSubnetGroupsWithContext{Ctx: ctx, I: i, Opts: opts})

	return m.MockUpdateSubnetGroupWithContext(ctx, i, opts)
}

// CallUpdateSubnetGroups to log calls
type CallUpdateSubnetGroups struct {
	I *dax.UpdateSubnetGroupInput
}

// UpdateSubnetGroups mocks UpdateSubnetGroups method
func (m *MockDaxClient) UpdateSubnetGroups(i *dax.UpdateSubnetGroupInput) (*dax.UpdateSubnetGroupOutput, error) {
	m.Called.UpdateSubnetGroups = append(m.Called.UpdateSubnetGroups, &CallUpdateSubnetGroups{I: i})
	return m.MockUpdateSubnetGroup(i)
}

// CallCreateSubnetGroupWithContext to log calls
type CallCreateSubnetGroupWithContext struct {
	Ctx  aws.Context
	I    *dax.CreateSubnetGroupInput
	Opts []request.Option
}

// CreateSubnetGroupWithContext mocks CreateSubnetGroupWithContext method
func (m *MockDaxClient) CreateSubnetGroupWithContext(ctx aws.Context, i *dax.CreateSubnetGroupInput, opts ...request.Option) (*dax.CreateSubnetGroupOutput, error) {
	m.Called.CreateSubnetGroupWithContext = append(m.Called.CreateSubnetGroupWithContext, &CallCreateSubnetGroupWithContext{Ctx: ctx, I: i, Opts: opts})

	return m.MockCreateSubnetGroupWithContext(ctx, i, opts)
}

// CallDeleteSubnetGroupWithContext to log calls
type CallDeleteSubnetGroupWithContext struct {
	Ctx  aws.Context
	I    *dax.DeleteSubnetGroupInput
	Opts []request.Option
}

// DeleteSubnetGroupWithContext mocks DeleteSubnetGroupWithContext method
func (m *MockDaxClient) DeleteSubnetGroupWithContext(ctx aws.Context, i *dax.DeleteSubnetGroupInput, opts ...request.Option) (*dax.DeleteSubnetGroupOutput, error) {
	m.Called.DeleteSubnetGroupWithContext = append(m.Called.DeleteSubnetGroupWithContext, &CallDeleteSubnetGroupWithContext{Ctx: ctx, I: i, Opts: opts})

	return m.MockDeleteSubnetGroupWithContext(ctx, i, opts)
}

// CallDescribeClustersWithContext to log calls
type CallDescribeClustersWithContext struct {
	Ctx  aws.Context
	I    *dax.DescribeClustersInput
	Opts []request.Option
}

// DescribeClustersWithContext mocks DescribeClustersWithContext method
func (m *MockDaxClient) DescribeClustersWithContext(ctx aws.Context, i *dax.DescribeClustersInput, opts ...request.Option) (*dax.DescribeClustersOutput, error) {
	m.Called.DescribeClustersWithContext = append(m.Called.DescribeClustersWithContext, &CallDescribeClustersWithContext{Ctx: ctx, I: i, Opts: opts})

	return m.MockDescribeClustersWithContext(ctx, i, opts)
}

// CallDescribeClusters to log calls
type CallDescribeClusters struct {
	I *dax.DescribeClustersInput
}

// DescribeClusters mocks DescribeClusters method
func (m *MockDaxClient) DescribeClusters(i *dax.DescribeClustersInput) (*dax.DescribeClustersOutput, error) {
	m.Called.DescribeClusters = append(m.Called.DescribeClusters, &CallDescribeClusters{I: i})
	return m.MockDescribeClusters(i)
}

// CallUpdateClusterWithContext to log calls
type CallUpdateClusterWithContext struct {
	Ctx  aws.Context
	I    *dax.UpdateClusterInput
	Opts []request.Option
}

// UpdateClusterWithContext mocks UpdateClusterWithContext method
func (m *MockDaxClient) UpdateClusterWithContext(ctx aws.Context, i *dax.UpdateClusterInput, opts ...request.Option) (*dax.UpdateClusterOutput, error) {
	m.Called.UpdateClusterWithContext = append(m.Called.UpdateClusterWithContext, &CallUpdateClusterWithContext{Ctx: ctx, I: i, Opts: opts})

	return m.MockUpdateClusterWithContext(ctx, i, opts)
}

// CallUpdateCluster to log calls
type CallUpdateCluster struct {
	I *dax.UpdateClusterInput
}

// UpdateCluster mocks UpdateCluster method
func (m *MockDaxClient) UpdateCluster(i *dax.UpdateClusterInput) (*dax.UpdateClusterOutput, error) {
	m.Called.UpdateCluster = append(m.Called.UpdateCluster, &CallUpdateCluster{I: i})
	return m.MockUpdateCluster(i)
}

// CallCreateClusterWithContext to log calls
type CallCreateClusterWithContext struct {
	Ctx  aws.Context
	I    *dax.CreateClusterInput
	Opts []request.Option
}

// CreateClusterWithContext mocks CreateClusterWithContext method
func (m *MockDaxClient) CreateClusterWithContext(ctx aws.Context, i *dax.CreateClusterInput, opts ...request.Option) (*dax.CreateClusterOutput, error) {
	m.Called.CreateClusterWithContext = append(m.Called.CreateClusterWithContext, &CallCreateClusterWithContext{Ctx: ctx, I: i, Opts: opts})

	return m.MockCreateClusterWithContext(ctx, i, opts)
}

// CallDeleteClusterWithContext to log calls
type CallDeleteClusterWithContext struct {
	Ctx  aws.Context
	I    *dax.DeleteClusterInput
	Opts []request.Option
}

// DeleteClusterWithContext mocks DeleteClusterWithContext method
func (m *MockDaxClient) DeleteClusterWithContext(ctx aws.Context, i *dax.DeleteClusterInput, opts ...request.Option) (*dax.DeleteClusterOutput, error) {
	m.Called.DeleteClusterWithContext = append(m.Called.DeleteClusterWithContext, &CallDeleteClusterWithContext{Ctx: ctx, I: i, Opts: opts})

	return m.MockDeleteClusterWithContext(ctx, i, opts)
}

// MockDaxClientCall is a type that implements all the methods for the Dax Client interface
type MockDaxClientCall struct {
	UpdateParameterGroupsWithContext   []*CallUpdateParameterGroupsWithContext
	UpdateParameterGroups              []*CallUpdateParameterGroups
	DescribeParameterGroupsWithContext []*CallDescribeParameterGroupsWithContext
	DescribeParameterGroups            []*CallDescribeParameterGroups
	DescribeParameters                 []*CallDescribeParameters
	DescribeParametersWithContext      []*CallDescribeParametersWithContext
	CreateParameterGroupWithContext    []*CallCreateParameterGroupWithContext
	DeleteParameterGroupWithContext    []*CallDeleteParameterGroupWithContext

	UpdateSubnetGroupsWithContext   []*CallUpdateSubnetGroupsWithContext
	UpdateSubnetGroups              []*CallUpdateSubnetGroups
	DescribeSubnetGroupsWithContext []*CallDescribeSubnetGroupsWithContext
	DescribeSubnetGroups            []*CallDescribeSubnetGroups
	CreateSubnetGroupWithContext    []*CallCreateSubnetGroupWithContext
	DeleteSubnetGroupWithContext    []*CallDeleteSubnetGroupWithContext

	UpdateClusterWithContext    []*CallUpdateClusterWithContext
	UpdateCluster               []*CallUpdateCluster
	DescribeClustersWithContext []*CallDescribeClustersWithContext
	DescribeClusters            []*CallDescribeClusters
	CreateClusterWithContext    []*CallCreateClusterWithContext
	DeleteClusterWithContext    []*CallDeleteClusterWithContext
}
