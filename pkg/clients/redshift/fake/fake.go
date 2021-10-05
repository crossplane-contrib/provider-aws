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
	"context"

	"github.com/aws/aws-sdk-go-v2/service/redshift"
)

// MockRedshiftClient for testing.
type MockRedshiftClient struct {
	MockCreate   func(ctx context.Context, input *redshift.CreateClusterInput, opts []func(*redshift.Options)) (*redshift.CreateClusterOutput, error)
	MockDescribe func(ctx context.Context, input *redshift.DescribeClustersInput, opts []func(*redshift.Options)) (*redshift.DescribeClustersOutput, error)
	MockModify   func(ctx context.Context, input *redshift.ModifyClusterInput, opts []func(*redshift.Options)) (*redshift.ModifyClusterOutput, error)
	MockDelete   func(ctx context.Context, input *redshift.DeleteClusterInput, opts []func(*redshift.Options)) (*redshift.DeleteClusterOutput, error)
}

// DescribeClusters finds Redshift Instance by name
func (m *MockRedshiftClient) DescribeClusters(ctx context.Context, input *redshift.DescribeClustersInput, opts ...func(*redshift.Options)) (*redshift.DescribeClustersOutput, error) {
	return m.MockDescribe(ctx, input, opts)
}

// CreateCluster creates Redshift Instance with provided Specification
func (m *MockRedshiftClient) CreateCluster(ctx context.Context, input *redshift.CreateClusterInput, opts ...func(*redshift.Options)) (*redshift.CreateClusterOutput, error) {
	return m.MockCreate(ctx, input, opts)
}

// ModifyCluster modifies Redshift Instance with provided Specification
func (m *MockRedshiftClient) ModifyCluster(ctx context.Context, input *redshift.ModifyClusterInput, opts ...func(*redshift.Options)) (*redshift.ModifyClusterOutput, error) {
	return m.MockModify(ctx, input, opts)
}

// DeleteCluster deletes Redshift Instance
func (m *MockRedshiftClient) DeleteCluster(ctx context.Context, input *redshift.DeleteClusterInput, opts ...func(*redshift.Options)) (*redshift.DeleteClusterOutput, error) {
	return m.MockDelete(ctx, input, opts)
}
