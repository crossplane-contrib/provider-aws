package fake

import "github.com/aws/aws-sdk-go-v2/service/sns"

// MockSubscriptionClient is a type that implements all the methods for SubscriptionClient interface
type MockSubscriptionClient struct {
	MockSubscribeRequest                 func(*sns.SubscribeInput) sns.SubscribeRequest
	MockUnsubscribeRequest               func(*sns.UnsubscribeInput) sns.UnsubscribeRequest
	MockGetSubscriptionAttributesRequest func(*sns.GetSubscriptionAttributesInput) sns.GetSubscriptionAttributesRequest
	MockSetSubscriptionAttributesRequest func(*sns.SetSubscriptionAttributesInput) sns.SetSubscriptionAttributesRequest
}

// SubscribeRequest mocks SubscribeRequest method
func (m *MockSubscriptionClient) SubscribeRequest(input *sns.SubscribeInput) sns.SubscribeRequest {
	return m.MockSubscribeRequest(input)
}

// UnsubscribeRequest mocks UnsubscribeRequest method
func (m *MockSubscriptionClient) UnsubscribeRequest(input *sns.UnsubscribeInput) sns.UnsubscribeRequest {
	return m.MockUnsubscribeRequest(input)
}

// GetSubscriptionAttributesRequest mocks GetSubscriptionAttributesRequest method
func (m *MockSubscriptionClient) GetSubscriptionAttributesRequest(input *sns.GetSubscriptionAttributesInput) sns.GetSubscriptionAttributesRequest {
	return m.MockGetSubscriptionAttributesRequest(input)
}

// SetSubscriptionAttributesRequest mocks SetSubscriptionAttributesRequest method
func (m *MockSubscriptionClient) SetSubscriptionAttributesRequest(input *sns.SetSubscriptionAttributesInput) sns.SetSubscriptionAttributesRequest {
	return m.MockSetSubscriptionAttributesRequest(input)
}
