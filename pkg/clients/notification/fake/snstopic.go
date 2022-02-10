package fake

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/sns"
)

// MockTopicClient is a type that implements all the methods for TopicClient interface
type MockTopicClient struct {
	MockCreateTopic        func(ctx context.Context, input *sns.CreateTopicInput, opts []func(*sns.Options)) (*sns.CreateTopicOutput, error)
	MockDeleteTopic        func(ctx context.Context, input *sns.DeleteTopicInput, opts []func(*sns.Options)) (*sns.DeleteTopicOutput, error)
	MockGetTopicAttributes func(ctx context.Context, input *sns.GetTopicAttributesInput, opts []func(*sns.Options)) (*sns.GetTopicAttributesOutput, error)
	MockSetTopicAttributes func(ctx context.Context, input *sns.SetTopicAttributesInput, opts []func(*sns.Options)) (*sns.SetTopicAttributesOutput, error)
}

// CreateTopic mocks CreateTopic method
func (m *MockTopicClient) CreateTopic(ctx context.Context, input *sns.CreateTopicInput, opts ...func(*sns.Options)) (*sns.CreateTopicOutput, error) {
	return m.MockCreateTopic(ctx, input, opts)
}

// DeleteTopic mocks DeleteTopic method
func (m *MockTopicClient) DeleteTopic(ctx context.Context, input *sns.DeleteTopicInput, opts ...func(*sns.Options)) (*sns.DeleteTopicOutput, error) {
	return m.MockDeleteTopic(ctx, input, opts)
}

// GetTopicAttributes mocks GetTopicAttributes method
func (m *MockTopicClient) GetTopicAttributes(ctx context.Context, input *sns.GetTopicAttributesInput, opts ...func(*sns.Options)) (*sns.GetTopicAttributesOutput, error) {
	return m.MockGetTopicAttributes(ctx, input, opts)
}

// SetTopicAttributes mocks SetTopicAttributes method
func (m *MockTopicClient) SetTopicAttributes(ctx context.Context, input *sns.SetTopicAttributesInput, opts ...func(*sns.Options)) (*sns.SetTopicAttributesOutput, error) {
	return m.MockSetTopicAttributes(ctx, input, opts)
}
