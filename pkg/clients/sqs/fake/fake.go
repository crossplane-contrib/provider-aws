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

	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

// MockSQSClient for testing.
type MockSQSClient struct {
	MockCreateQueue        func(ctx context.Context, input *sqs.CreateQueueInput, opts []func(*sqs.Options)) (*sqs.CreateQueueOutput, error)
	MockDeleteQueue        func(ctx context.Context, input *sqs.DeleteQueueInput, opts []func(*sqs.Options)) (*sqs.DeleteQueueOutput, error)
	MockTagQueue           func(ctx context.Context, input *sqs.TagQueueInput, opts []func(*sqs.Options)) (*sqs.TagQueueOutput, error)
	MockUntagQueue         func(ctx context.Context, input *sqs.UntagQueueInput, opts []func(*sqs.Options)) (*sqs.UntagQueueOutput, error)
	MockListQueueTags      func(ctx context.Context, input *sqs.ListQueueTagsInput, opts []func(*sqs.Options)) (*sqs.ListQueueTagsOutput, error)
	MockGetQueueAttributes func(ctx context.Context, input *sqs.GetQueueAttributesInput, opts []func(*sqs.Options)) (*sqs.GetQueueAttributesOutput, error)
	MockSetQueueAttributes func(ctx context.Context, input *sqs.SetQueueAttributesInput, opts []func(*sqs.Options)) (*sqs.SetQueueAttributesOutput, error)
	MockGetQueueURL        func(ctx context.Context, input *sqs.GetQueueUrlInput, opts []func(*sqs.Options)) (*sqs.GetQueueUrlOutput, error)
}

// CreateQueue mocks CreateQueue
func (m *MockSQSClient) CreateQueue(ctx context.Context, i *sqs.CreateQueueInput, opts ...func(*sqs.Options)) (*sqs.CreateQueueOutput, error) {
	return m.MockCreateQueue(ctx, i, opts)
}

// DeleteQueue mocks DeleteQueue
func (m *MockSQSClient) DeleteQueue(ctx context.Context, i *sqs.DeleteQueueInput, opts ...func(*sqs.Options)) (*sqs.DeleteQueueOutput, error) {
	return m.MockDeleteQueue(ctx, i, opts)
}

// TagQueue mocks TagQueue
func (m *MockSQSClient) TagQueue(ctx context.Context, i *sqs.TagQueueInput, opts ...func(*sqs.Options)) (*sqs.TagQueueOutput, error) {
	return m.MockTagQueue(ctx, i, opts)
}

// UntagQueue mocks UntagQueue
func (m *MockSQSClient) UntagQueue(ctx context.Context, i *sqs.UntagQueueInput, opts ...func(*sqs.Options)) (*sqs.UntagQueueOutput, error) {
	return m.MockUntagQueue(ctx, i, opts)
}

// ListQueueTags mocks ListQueueTags
func (m *MockSQSClient) ListQueueTags(ctx context.Context, i *sqs.ListQueueTagsInput, opts ...func(*sqs.Options)) (*sqs.ListQueueTagsOutput, error) {
	return m.MockListQueueTags(ctx, i, opts)
}

// GetQueueAttributes mocks GetQueueAttributes
func (m *MockSQSClient) GetQueueAttributes(ctx context.Context, i *sqs.GetQueueAttributesInput, opts ...func(*sqs.Options)) (*sqs.GetQueueAttributesOutput, error) {
	return m.MockGetQueueAttributes(ctx, i, opts)
}

// SetQueueAttributes mocks SetQueueAttributes
func (m *MockSQSClient) SetQueueAttributes(ctx context.Context, i *sqs.SetQueueAttributesInput, opts ...func(*sqs.Options)) (*sqs.SetQueueAttributesOutput, error) {
	return m.MockSetQueueAttributes(ctx, i, opts)
}

// GetQueueUrl mocks GetQueueUrl
func (m *MockSQSClient) GetQueueUrl(ctx context.Context, i *sqs.GetQueueUrlInput, opts ...func(*sqs.Options)) (*sqs.GetQueueUrlOutput, error) { //nolint:golint
	return m.MockGetQueueURL(ctx, i, opts)
}
