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
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

// MockSQSClient for testing.
type MockSQSClient struct {
	MockCreateQueueRequest        func(input *sqs.CreateQueueInput) sqs.CreateQueueRequest
	MockDeleteQueueRequest        func(input *sqs.DeleteQueueInput) sqs.DeleteQueueRequest
	MockTagQueueRequest           func(input *sqs.TagQueueInput) sqs.TagQueueRequest
	MockUntagQueueRequest         func(input *sqs.UntagQueueInput) sqs.UntagQueueRequest
	MockListQueueTagsRequest      func(input *sqs.ListQueueTagsInput) sqs.ListQueueTagsRequest
	MockGetQueueAttributesRequest func(input *sqs.GetQueueAttributesInput) sqs.GetQueueAttributesRequest
	MockSetQueueAttributesRequest func(input *sqs.SetQueueAttributesInput) sqs.SetQueueAttributesRequest
	MockGetQueueURLRequest        func(input *sqs.GetQueueUrlInput) sqs.GetQueueUrlRequest
}

// CreateQueueRequest mocks CreateQueueRequest
func (m *MockSQSClient) CreateQueueRequest(i *sqs.CreateQueueInput) sqs.CreateQueueRequest {
	return m.MockCreateQueueRequest(i)
}

// DeleteQueueRequest mocks DeleteQueueRequest
func (m *MockSQSClient) DeleteQueueRequest(i *sqs.DeleteQueueInput) sqs.DeleteQueueRequest {
	return m.MockDeleteQueueRequest(i)
}

// TagQueueRequest mocks TagQueueRequest
func (m *MockSQSClient) TagQueueRequest(i *sqs.TagQueueInput) sqs.TagQueueRequest {
	return m.MockTagQueueRequest(i)
}

// UntagQueueRequest mocks UntagQueueRequest
func (m *MockSQSClient) UntagQueueRequest(i *sqs.UntagQueueInput) sqs.UntagQueueRequest {
	return m.MockUntagQueueRequest(i)
}

// ListQueueTagsRequest mocks ListQueueTagsRequest
func (m *MockSQSClient) ListQueueTagsRequest(i *sqs.ListQueueTagsInput) sqs.ListQueueTagsRequest {
	return m.MockListQueueTagsRequest(i)
}

// GetQueueAttributesRequest mocks GetQueueAttributesRequest
func (m *MockSQSClient) GetQueueAttributesRequest(i *sqs.GetQueueAttributesInput) sqs.GetQueueAttributesRequest {
	return m.MockGetQueueAttributesRequest(i)
}

// SetQueueAttributesRequest mocks SetQueueAttributesRequest
func (m *MockSQSClient) SetQueueAttributesRequest(i *sqs.SetQueueAttributesInput) sqs.SetQueueAttributesRequest {
	return m.MockSetQueueAttributesRequest(i)
}

// GetQueueUrlRequest mocks GetQueueUrlRequest
func (m *MockSQSClient) GetQueueUrlRequest(i *sqs.GetQueueUrlInput) sqs.GetQueueUrlRequest { //nolint:golint
	return m.MockGetQueueURLRequest(i)
}
