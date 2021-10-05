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

	clientset "github.com/crossplane/provider-aws/pkg/clients/dbsubnetgroup"
)

// this ensures that the mock implements the client interface
var _ clientset.Client = (*MockDBSubnetGroupClient)(nil)

// MockDBSubnetGroupClient is a type that implements all the methods for DBSubnetGroupClient interface
type MockDBSubnetGroupClient struct {
	MockCreateDBSubnetGroup    func(context.Context, *rds.CreateDBSubnetGroupInput, []func(*rds.Options)) (*rds.CreateDBSubnetGroupOutput, error)
	MockDeleteDBSubnetGroup    func(context.Context, *rds.DeleteDBSubnetGroupInput, []func(*rds.Options)) (*rds.DeleteDBSubnetGroupOutput, error)
	MockDescribeDBSubnetGroups func(context.Context, *rds.DescribeDBSubnetGroupsInput, []func(*rds.Options)) (*rds.DescribeDBSubnetGroupsOutput, error)
	MockModifyDBSubnetGroup    func(context.Context, *rds.ModifyDBSubnetGroupInput, []func(*rds.Options)) (*rds.ModifyDBSubnetGroupOutput, error)
	MockAddTagsToResource      func(context.Context, *rds.AddTagsToResourceInput, []func(*rds.Options)) (*rds.AddTagsToResourceOutput, error)
	MockListTagsForResource    func(context.Context, *rds.ListTagsForResourceInput, []func(*rds.Options)) (*rds.ListTagsForResourceOutput, error)
}

// CreateDBSubnetGroup mocks CreateDBSubnetGroup method
func (m *MockDBSubnetGroupClient) CreateDBSubnetGroup(ctx context.Context, input *rds.CreateDBSubnetGroupInput, opts ...func(*rds.Options)) (*rds.CreateDBSubnetGroupOutput, error) {
	return m.MockCreateDBSubnetGroup(ctx, input, opts)
}

// DeleteDBSubnetGroup mocks DeleteDBSubnetGroup method
func (m *MockDBSubnetGroupClient) DeleteDBSubnetGroup(ctx context.Context, input *rds.DeleteDBSubnetGroupInput, opts ...func(*rds.Options)) (*rds.DeleteDBSubnetGroupOutput, error) {
	return m.MockDeleteDBSubnetGroup(ctx, input, opts)
}

// DescribeDBSubnetGroups mocks DescribeDBSubnetGroups method
func (m *MockDBSubnetGroupClient) DescribeDBSubnetGroups(ctx context.Context, input *rds.DescribeDBSubnetGroupsInput, opts ...func(*rds.Options)) (*rds.DescribeDBSubnetGroupsOutput, error) {
	return m.MockDescribeDBSubnetGroups(ctx, input, opts)

}

// ModifyDBSubnetGroup mocks ModifyDBSubnetGroup method
func (m *MockDBSubnetGroupClient) ModifyDBSubnetGroup(ctx context.Context, input *rds.ModifyDBSubnetGroupInput, opts ...func(*rds.Options)) (*rds.ModifyDBSubnetGroupOutput, error) {
	return m.MockModifyDBSubnetGroup(ctx, input, opts)

}

// AddTagsToResource mocks AddTagsToResource method
func (m *MockDBSubnetGroupClient) AddTagsToResource(ctx context.Context, input *rds.AddTagsToResourceInput, opts ...func(*rds.Options)) (*rds.AddTagsToResourceOutput, error) {
	return m.MockAddTagsToResource(ctx, input, opts)
}

// ListTagsForResource mocks ListTagsForResource method
func (m *MockDBSubnetGroupClient) ListTagsForResource(ctx context.Context, input *rds.ListTagsForResourceInput, opts ...func(*rds.Options)) (*rds.ListTagsForResourceOutput, error) {
	return m.MockListTagsForResource(ctx, input, opts)
}
