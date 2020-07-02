package fake

import (
	"github.com/aws/aws-sdk-go-v2/service/sns"
)

// MockTopicClient is a type that implements all the methods for TopicClient interface
type MockTopicClient struct {
	MockCreateTopicRequest        func(*sns.CreateTopicInput) sns.CreateTopicRequest
	MockDeleteTopicRequest        func(*sns.DeleteTopicInput) sns.DeleteTopicRequest
	MockGetTopicAttributesRequest func(*sns.GetTopicAttributesInput) sns.GetTopicAttributesRequest
	MockSetTopicAttributesRequest func(*sns.SetTopicAttributesInput) sns.SetTopicAttributesRequest
}

// CreateTopicRequest mocks CreateTopicRequest method
func (m *MockTopicClient) CreateTopicRequest(input *sns.CreateTopicInput) sns.CreateTopicRequest {
	return m.MockCreateTopicRequest(input)
}

// DeleteTopicRequest mocks DeleteTopicRequest method
func (m *MockTopicClient) DeleteTopicRequest(input *sns.DeleteTopicInput) sns.DeleteTopicRequest {
	return m.MockDeleteTopicRequest(input)
}

// GetTopicAttributesRequest mocks GetTopicAttributesRequest method
func (m *MockTopicClient) GetTopicAttributesRequest(input *sns.GetTopicAttributesInput) sns.GetTopicAttributesRequest {
	return m.MockGetTopicAttributesRequest(input)
}

// SetTopicAttributesRequest mocks SetTopicAttributesRequest method
func (m *MockTopicClient) SetTopicAttributesRequest(input *sns.SetTopicAttributesInput) sns.SetTopicAttributesRequest {
	return m.MockSetTopicAttributesRequest(input)
}
