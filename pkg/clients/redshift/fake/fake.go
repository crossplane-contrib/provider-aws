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
	"github.com/aws/aws-sdk-go-v2/service/redshift"
)

// MockRedshiftClient for testing.
type MockRedshiftClient struct {
	MockCreate   func(*redshift.CreateClusterInput) redshift.CreateClusterRequest
	MockDescribe func(*redshift.DescribeClustersInput) redshift.DescribeClustersRequest
	MockModify   func(*redshift.ModifyClusterInput) redshift.ModifyClusterRequest
	MockDelete   func(*redshift.DeleteClusterInput) redshift.DeleteClusterRequest
}

// DescribeClustersRequest finds Redshift Instance by name
func (m *MockRedshiftClient) DescribeClustersRequest(i *redshift.DescribeClustersInput) redshift.DescribeClustersRequest {
	return m.MockDescribe(i)
}

// CreateClusterRequest creates Redshift Instance with provided Specification
func (m *MockRedshiftClient) CreateClusterRequest(i *redshift.CreateClusterInput) redshift.CreateClusterRequest {
	return m.MockCreate(i)
}

// ModifyClusterRequest modifies Redshift Instance with provided Specification
func (m *MockRedshiftClient) ModifyClusterRequest(i *redshift.ModifyClusterInput) redshift.ModifyClusterRequest {
	return m.MockModify(i)
}

// DeleteClusterRequest deletes Redshift Instance
func (m *MockRedshiftClient) DeleteClusterRequest(i *redshift.DeleteClusterInput) redshift.DeleteClusterRequest {
	return m.MockDelete(i)
}
