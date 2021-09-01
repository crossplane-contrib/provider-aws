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
	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

// MockResourceTagClient mocks ResourceTagClient
type MockResourceTagClient struct {
	MockCreateTagsRequest   func(*ec2.CreateTagsInput) ec2.CreateTagsRequest
	MockDescribeTagsRequest func(*ec2.DescribeTagsInput) ec2.DescribeTagsRequest
	MockDeleteTagsRequest   func(*ec2.DeleteTagsInput) ec2.DeleteTagsRequest
}

// CreateTagsRequest calls MockCreateTagsRequest
func (m *MockResourceTagClient) CreateTagsRequest(i *ec2.CreateTagsInput) ec2.CreateTagsRequest {
	return m.MockCreateTagsRequest(i)
}

// DescribeTagsRequest calls MockDescribeTagsRequest
func (m *MockResourceTagClient) DescribeTagsRequest(i *ec2.DescribeTagsInput) ec2.DescribeTagsRequest {
	return m.MockDescribeTagsRequest(i)
}

// DeleteTagsRequest calls MockDeleteTagsRequest
func (m *MockResourceTagClient) DeleteTagsRequest(i *ec2.DeleteTagsInput) ec2.DeleteTagsRequest {
	return m.MockDeleteTagsRequest(i)
}
