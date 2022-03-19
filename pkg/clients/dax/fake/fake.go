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
	MockCreateCluster            func(*dax.CreateClusterInput) (*dax.CreateClusterOutput, error)
	MockCreateClusterWithContext func(context.Context, *dax.CreateClusterInput, []request.Option) (*dax.CreateClusterOutput, error)
	MockCreateClusterRequest     func(*dax.CreateClusterInput) (*request.Request, *dax.CreateClusterOutput)

	MockCreateParameterGroup            func(*dax.CreateParameterGroupInput) (*dax.CreateParameterGroupOutput, error)
	MockCreateParameterGroupWithContext func(context.Context, *dax.CreateParameterGroupInput, []request.Option) (*dax.CreateParameterGroupOutput, error)
	MockCreateParameterGroupRequest     func(*dax.CreateParameterGroupInput) (*request.Request, *dax.CreateParameterGroupOutput)

	MockCreateSubnetGroup            func(*dax.CreateSubnetGroupInput) (*dax.CreateSubnetGroupOutput, error)
	MockCreateSubnetGroupWithContext func(context.Context, *dax.CreateSubnetGroupInput, []request.Option) (*dax.CreateSubnetGroupOutput, error)
	MockCreateSubnetGroupRequest     func(*dax.CreateSubnetGroupInput) (*request.Request, *dax.CreateSubnetGroupOutput)

	MockDecreaseReplicationFactor            func(*dax.DecreaseReplicationFactorInput) (*dax.DecreaseReplicationFactorOutput, error)
	MockDecreaseReplicationFactorWithContext func(context.Context, *dax.DecreaseReplicationFactorInput, ...request.Option) (*dax.DecreaseReplicationFactorOutput, error)
	MockDecreaseReplicationFactorRequest     func(*dax.DecreaseReplicationFactorInput) (*request.Request, *dax.DecreaseReplicationFactorOutput)

	MockDeleteCluster            func(*dax.DeleteClusterInput) (*dax.DeleteClusterOutput, error)
	MockDeleteClusterWithContext func(context.Context, *dax.DeleteClusterInput, []request.Option) (*dax.DeleteClusterOutput, error)
	MockDeleteClusterRequest     func(*dax.DeleteClusterInput) (*request.Request, *dax.DeleteClusterOutput)

	MockDeleteParameterGroup            func(*dax.DeleteParameterGroupInput) (*dax.DeleteParameterGroupOutput, error)
	MockDeleteParameterGroupWithContext func(context.Context, *dax.DeleteParameterGroupInput, []request.Option) (*dax.DeleteParameterGroupOutput, error)
	MockDeleteParameterGroupRequest     func(*dax.DeleteParameterGroupInput) (*request.Request, *dax.DeleteParameterGroupOutput)

	MockDeleteSubnetGroup            func(*dax.DeleteSubnetGroupInput) (*dax.DeleteSubnetGroupOutput, error)
	MockDeleteSubnetGroupWithContext func(context.Context, *dax.DeleteSubnetGroupInput, []request.Option) (*dax.DeleteSubnetGroupOutput, error)
	MockDeleteSubnetGroupRequest     func(*dax.DeleteSubnetGroupInput) (*request.Request, *dax.DeleteSubnetGroupOutput)

	MockDescribeClusters            func(*dax.DescribeClustersInput) (*dax.DescribeClustersOutput, error)
	MockDescribeClustersWithContext func(context.Context, *dax.DescribeClustersInput, []request.Option) (*dax.DescribeClustersOutput, error)
	MockDescribeClustersRequest     func(*dax.DescribeClustersInput) (*request.Request, *dax.DescribeClustersOutput)

	MockDescribeDefaultParameters            func(*dax.DescribeDefaultParametersInput) (*dax.DescribeDefaultParametersOutput, error)
	MockDescribeDefaultParametersWithContext func(context.Context, *dax.DescribeDefaultParametersInput, ...request.Option) (*dax.DescribeDefaultParametersOutput, error)
	MockDescribeDefaultParametersRequest     func(*dax.DescribeDefaultParametersInput) (*request.Request, *dax.DescribeDefaultParametersOutput)

	MockDescribeEvents            func(*dax.DescribeEventsInput) (*dax.DescribeEventsOutput, error)
	MockDescribeEventsWithContext func(context.Context, *dax.DescribeEventsInput, ...request.Option) (*dax.DescribeEventsOutput, error)
	MockDescribeEventsRequest     func(*dax.DescribeEventsInput) (*request.Request, *dax.DescribeEventsOutput)

	MockDescribeParameterGroups            func(*dax.DescribeParameterGroupsInput) (*dax.DescribeParameterGroupsOutput, error)
	MockDescribeParameterGroupsWithContext func(context.Context, *dax.DescribeParameterGroupsInput, []request.Option) (*dax.DescribeParameterGroupsOutput, error)
	MockDescribeParameterGroupsRequest     func(*dax.DescribeParameterGroupsInput) (*request.Request, *dax.DescribeParameterGroupsOutput)

	MockDescribeParameters            func(*dax.DescribeParametersInput) (*dax.DescribeParametersOutput, error)
	MockDescribeParametersWithContext func(context.Context, *dax.DescribeParametersInput, ...request.Option) (*dax.DescribeParametersOutput, error)
	MockDescribeParametersRequest     func(*dax.DescribeParametersInput) (*request.Request, *dax.DescribeParametersOutput)

	MockDescribeSubnetGroups            func(*dax.DescribeSubnetGroupsInput) (*dax.DescribeSubnetGroupsOutput, error)
	MockDescribeSubnetGroupsWithContext func(context.Context, *dax.DescribeSubnetGroupsInput, []request.Option) (*dax.DescribeSubnetGroupsOutput, error)
	MockDescribeSubnetGroupsRequest     func(*dax.DescribeSubnetGroupsInput) (*request.Request, *dax.DescribeSubnetGroupsOutput)

	MockIncreaseReplicationFactor            func(*dax.IncreaseReplicationFactorInput) (*dax.IncreaseReplicationFactorOutput, error)
	MockIncreaseReplicationFactorWithContext func(context.Context, *dax.IncreaseReplicationFactorInput, ...request.Option) (*dax.IncreaseReplicationFactorOutput, error)
	MockIncreaseReplicationFactorRequest     func(*dax.IncreaseReplicationFactorInput) (*request.Request, *dax.IncreaseReplicationFactorOutput)

	MockListTags            func(*dax.ListTagsInput) (*dax.ListTagsOutput, error)
	MockListTagsWithContext func(context.Context, *dax.ListTagsInput, ...request.Option) (*dax.ListTagsOutput, error)
	MockListTagsRequest     func(*dax.ListTagsInput) (*request.Request, *dax.ListTagsOutput)

	MockRebootNode            func(*dax.RebootNodeInput) (*dax.RebootNodeOutput, error)
	MockRebootNodeWithContext func(context.Context, *dax.RebootNodeInput, ...request.Option) (*dax.RebootNodeOutput, error)
	MockRebootNodeRequest     func(*dax.RebootNodeInput) (*request.Request, *dax.RebootNodeOutput)

	MockTagResource            func(*dax.TagResourceInput) (*dax.TagResourceOutput, error)
	MockTagResourceWithContext func(context.Context, *dax.TagResourceInput, ...request.Option) (*dax.TagResourceOutput, error)
	MockTagResourceRequest     func(*dax.TagResourceInput) (*request.Request, *dax.TagResourceOutput)

	MockUntagResource            func(*dax.UntagResourceInput) (*dax.UntagResourceOutput, error)
	MockUntagResourceWithContext func(context.Context, *dax.UntagResourceInput, ...request.Option) (*dax.UntagResourceOutput, error)
	MockUntagResourceRequest     func(*dax.UntagResourceInput) (*request.Request, *dax.UntagResourceOutput)

	MockUpdateCluster            func(*dax.UpdateClusterInput) (*dax.UpdateClusterOutput, error)
	MockUpdateClusterWithContext func(context.Context, *dax.UpdateClusterInput, []request.Option) (*dax.UpdateClusterOutput, error)
	MockUpdateClusterRequest     func(*dax.UpdateClusterInput) (*request.Request, *dax.UpdateClusterOutput)

	MockUpdateParameterGroup            func(*dax.UpdateParameterGroupInput) (*dax.UpdateParameterGroupOutput, error)
	MockUpdateParameterGroupWithContext func(context.Context, *dax.UpdateParameterGroupInput, []request.Option) (*dax.UpdateParameterGroupOutput, error)
	MockUpdateParameterGroupRequest     func(*dax.UpdateParameterGroupInput) (*request.Request, *dax.UpdateParameterGroupOutput)

	MockUpdateSubnetGroup            func(*dax.UpdateSubnetGroupInput) (*dax.UpdateSubnetGroupOutput, error)
	MockUpdateSubnetGroupWithContext func(context.Context, *dax.UpdateSubnetGroupInput, []request.Option) (*dax.UpdateSubnetGroupOutput, error)
	MockUpdateSubnetGroupRequest     func(*dax.UpdateSubnetGroupInput) (*request.Request, *dax.UpdateSubnetGroupOutput)

	Called MockDaxClientCall
}

// ParameterGroup Mocks

type CallDescribeParameterGroupsWithContext struct {
	Ctx  aws.Context
	I    *dax.DescribeParameterGroupsInput
	Opts []request.Option
}

// DescribeParameterGroupsWithContext calls MockDescribeParameterGroupsWithContext
func (m *MockDaxClient) DescribeParameterGroupsWithContext(ctx aws.Context, i *dax.DescribeParameterGroupsInput, opts ...request.Option) (*dax.DescribeParameterGroupsOutput, error) {
	m.Called.DescribeParameterGroupsWithContext = append(m.Called.DescribeParameterGroupsWithContext, &CallDescribeParameterGroupsWithContext{Ctx: ctx, I: i, Opts: opts})

	return m.MockDescribeParameterGroupsWithContext(ctx, i, opts)
}

type CallDescribeParameterGroups struct {
	I *dax.DescribeParameterGroupsInput
}

// DescribeParameterGroups calls MockDescribeParameterGroups
func (m *MockDaxClient) DescribeParameterGroups(i *dax.DescribeParameterGroupsInput) (*dax.DescribeParameterGroupsOutput, error) {
	m.Called.DescribeParameterGroups = append(m.Called.DescribeParameterGroups, &CallDescribeParameterGroups{I: i})
	return m.MockDescribeParameterGroups(i)
}

type CallUpdateParameterGroupsWithContext struct {
	Ctx  aws.Context
	I    *dax.UpdateParameterGroupInput
	Opts []request.Option
}

// UpdateParameterGroupWithContext calls MockUpdateParameterGroupWithContext
func (m *MockDaxClient) UpdateParameterGroupWithContext(ctx aws.Context, i *dax.UpdateParameterGroupInput, opts ...request.Option) (*dax.UpdateParameterGroupOutput, error) {
	m.Called.UpdateParameterGroupsWithContext = append(m.Called.UpdateParameterGroupsWithContext, &CallUpdateParameterGroupsWithContext{Ctx: ctx, I: i, Opts: opts})

	return m.MockUpdateParameterGroupWithContext(ctx, i, opts)
}

type CallUpdateParameterGroups struct {
	I *dax.UpdateParameterGroupInput
}

// UpdateParameterGroups calls MockUpdateParameterGroups
func (m *MockDaxClient) UpdateParameterGroups(i *dax.UpdateParameterGroupInput) (*dax.UpdateParameterGroupOutput, error) {
	m.Called.UpdateParameterGroups = append(m.Called.UpdateParameterGroups, &CallUpdateParameterGroups{I: i})
	return m.MockUpdateParameterGroup(i)
}

type CallDescribeParameters struct {
	I *dax.DescribeParametersInput
}

// DescribeParameters calls MockDescribeParameters
func (m *MockDaxClient) DescribeParameters(i *dax.DescribeParametersInput) (*dax.DescribeParametersOutput, error) {
	m.Called.DescribeParameters = append(m.Called.DescribeParameters, &CallDescribeParameters{I: i})

	return m.MockDescribeParameters(i)
}

type CallCreateParameterGroupWithContext struct {
	Ctx  aws.Context
	I    *dax.CreateParameterGroupInput
	Opts []request.Option
}

func (m *MockDaxClient) CreateParameterGroupWithContext(ctx aws.Context, i *dax.CreateParameterGroupInput, opts ...request.Option) (*dax.CreateParameterGroupOutput, error) {
	m.Called.CreateParameterGroupWithContext = append(m.Called.CreateParameterGroupWithContext, &CallCreateParameterGroupWithContext{Ctx: ctx, I: i, Opts: opts})

	return m.MockCreateParameterGroupWithContext(ctx, i, opts)
}

type CallDeleteParameterGroupWithContext struct {
	Ctx  aws.Context
	I    *dax.DeleteParameterGroupInput
	Opts []request.Option
}

func (m *MockDaxClient) DeleteParameterGroupWithContext(ctx aws.Context, i *dax.DeleteParameterGroupInput, opts ...request.Option) (*dax.DeleteParameterGroupOutput, error) {
	m.Called.DeleteParameterGroupWithContext = append(m.Called.DeleteParameterGroupWithContext, &CallDeleteParameterGroupWithContext{Ctx: ctx, I: i, Opts: opts})

	return m.MockDeleteParameterGroupWithContext(ctx, i, opts)
}

// SubnetGroup Mocks

type CallDescribeSubnetGroupsWithContext struct {
	Ctx  aws.Context
	I    *dax.DescribeSubnetGroupsInput
	Opts []request.Option
}

// DescribeSubnetGroupsWithContext calls MockDescribeSubnetGroupsWithContext
func (m *MockDaxClient) DescribeSubnetGroupsWithContext(ctx aws.Context, i *dax.DescribeSubnetGroupsInput, opts ...request.Option) (*dax.DescribeSubnetGroupsOutput, error) {
	m.Called.DescribeSubnetGroupsWithContext = append(m.Called.DescribeSubnetGroupsWithContext, &CallDescribeSubnetGroupsWithContext{Ctx: ctx, I: i, Opts: opts})

	return m.MockDescribeSubnetGroupsWithContext(ctx, i, opts)
}

type CallDescribeSubnetGroups struct {
	I *dax.DescribeSubnetGroupsInput
}

// DescribeSubnetGroups calls MockDescribeSubnetGroups
func (m *MockDaxClient) DescribeSubnetGroups(i *dax.DescribeSubnetGroupsInput) (*dax.DescribeSubnetGroupsOutput, error) {
	m.Called.DescribeSubnetGroups = append(m.Called.DescribeSubnetGroups, &CallDescribeSubnetGroups{I: i})
	return m.DescribeSubnetGroups(i)
}

type CallUpdateSubnetGroupsWithContext struct {
	Ctx  aws.Context
	I    *dax.UpdateSubnetGroupInput
	Opts []request.Option
}

// UpdateSubnetGroupWithContext calls MockUpdateSubnetGroupWithContext
func (m *MockDaxClient) UpdateSubnetGroupWithContext(ctx aws.Context, i *dax.UpdateSubnetGroupInput, opts ...request.Option) (*dax.UpdateSubnetGroupOutput, error) {
	m.Called.UpdateSubnetGroupsWithContext = append(m.Called.UpdateSubnetGroupsWithContext, &CallUpdateSubnetGroupsWithContext{Ctx: ctx, I: i, Opts: opts})

	return m.MockUpdateSubnetGroupWithContext(ctx, i, opts)
}

type CallUpdateSubnetGroups struct {
	I *dax.UpdateSubnetGroupInput
}

// UpdateSubnetGroups calls MockUpdateSubnetGroups
func (m *MockDaxClient) UpdateSubnetGroups(i *dax.UpdateSubnetGroupInput) (*dax.UpdateSubnetGroupOutput, error) {
	m.Called.UpdateSubnetGroups = append(m.Called.UpdateSubnetGroups, &CallUpdateSubnetGroups{I: i})
	return m.MockUpdateSubnetGroup(i)
}

type CallCreateSubnetGroupWithContext struct {
	Ctx  aws.Context
	I    *dax.CreateSubnetGroupInput
	Opts []request.Option
}

func (m *MockDaxClient) CreateSubnetGroupWithContext(ctx aws.Context, i *dax.CreateSubnetGroupInput, opts ...request.Option) (*dax.CreateSubnetGroupOutput, error) {
	m.Called.CreateSubnetGroupWithContext = append(m.Called.CreateSubnetGroupWithContext, &CallCreateSubnetGroupWithContext{Ctx: ctx, I: i, Opts: opts})

	return m.MockCreateSubnetGroupWithContext(ctx, i, opts)
}

type CallDeleteSubnetGroupWithContext struct {
	Ctx  aws.Context
	I    *dax.DeleteSubnetGroupInput
	Opts []request.Option
}

func (m *MockDaxClient) DeleteSubnetGroupWithContext(ctx aws.Context, i *dax.DeleteSubnetGroupInput, opts ...request.Option) (*dax.DeleteSubnetGroupOutput, error) {
	m.Called.DeleteSubnetGroupWithContext = append(m.Called.DeleteSubnetGroupWithContext, &CallDeleteSubnetGroupWithContext{Ctx: ctx, I: i, Opts: opts})

	return m.MockDeleteSubnetGroupWithContext(ctx, i, opts)
}

// Cluster Mocks

type CallDescribeClustersWithContext struct {
	Ctx  aws.Context
	I    *dax.DescribeClustersInput
	Opts []request.Option
}

// DescribeClustersWithContext calls MockDescribeClustersWithContext
func (m *MockDaxClient) DescribeClustersWithContext(ctx aws.Context, i *dax.DescribeClustersInput, opts ...request.Option) (*dax.DescribeClustersOutput, error) {
	m.Called.DescribeClustersWithContext = append(m.Called.DescribeClustersWithContext, &CallDescribeClustersWithContext{Ctx: ctx, I: i, Opts: opts})

	return m.MockDescribeClustersWithContext(ctx, i, opts)
}

type CallDescribeClusters struct {
	I *dax.DescribeClustersInput
}

// DescribeClusters calls MockDescribeClusters
func (m *MockDaxClient) DescribeClusters(i *dax.DescribeClustersInput) (*dax.DescribeClustersOutput, error) {
	m.Called.DescribeClusters = append(m.Called.DescribeClusters, &CallDescribeClusters{I: i})
	return m.MockDescribeClusters(i)
}

type CallUpdateClusterWithContext struct {
	Ctx  aws.Context
	I    *dax.UpdateClusterInput
	Opts []request.Option
}

// UpdateClusterWithContext calls MockUpdateClusterWithContext
func (m *MockDaxClient) UpdateClusterWithContext(ctx aws.Context, i *dax.UpdateClusterInput, opts ...request.Option) (*dax.UpdateClusterOutput, error) {
	m.Called.UpdateClusterWithContext = append(m.Called.UpdateClusterWithContext, &CallUpdateClusterWithContext{Ctx: ctx, I: i, Opts: opts})

	return m.MockUpdateClusterWithContext(ctx, i, opts)
}

type CallUpdateCluster struct {
	I *dax.UpdateClusterInput
}

// UpdateCluster calls MockUpdateCluster
func (m *MockDaxClient) UpdateCluster(i *dax.UpdateClusterInput) (*dax.UpdateClusterOutput, error) {
	m.Called.UpdateCluster = append(m.Called.UpdateCluster, &CallUpdateCluster{I: i})
	return m.MockUpdateCluster(i)
}

type CallCreateClusterWithContext struct {
	Ctx  aws.Context
	I    *dax.CreateClusterInput
	Opts []request.Option
}

func (m *MockDaxClient) CreateClusterWithContext(ctx aws.Context, i *dax.CreateClusterInput, opts ...request.Option) (*dax.CreateClusterOutput, error) {
	m.Called.CreateClusterWithContext = append(m.Called.CreateClusterWithContext, &CallCreateClusterWithContext{Ctx: ctx, I: i, Opts: opts})

	return m.MockCreateClusterWithContext(ctx, i, opts)
}

type CallDeleteClusterWithContext struct {
	Ctx  aws.Context
	I    *dax.DeleteClusterInput
	Opts []request.Option
}

func (m *MockDaxClient) DeleteClusterWithContext(ctx aws.Context, i *dax.DeleteClusterInput, opts ...request.Option) (*dax.DeleteClusterOutput, error) {
	m.Called.DeleteClusterWithContext = append(m.Called.DeleteClusterWithContext, &CallDeleteClusterWithContext{Ctx: ctx, I: i, Opts: opts})

	return m.MockDeleteClusterWithContext(ctx, i, opts)
}

// MockDaxClientCall to log calls
type MockDaxClientCall struct {
	UpdateParameterGroupsWithContext   []*CallUpdateParameterGroupsWithContext
	UpdateParameterGroups              []*CallUpdateParameterGroups
	DescribeParameterGroupsWithContext []*CallDescribeParameterGroupsWithContext
	DescribeParameterGroups            []*CallDescribeParameterGroups
	DescribeParameters                 []*CallDescribeParameters
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
