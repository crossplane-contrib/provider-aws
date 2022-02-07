package fake

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/sns"
)

// MockSubscriptionClient is a type that implements all the methods for SubscriptionClient interface
type MockSubscriptionClient struct {
	MockSubscribe                 func(ctx context.Context, input *sns.SubscribeInput, opts []func(*sns.Options)) (*sns.SubscribeOutput, error)
	MockUnsubscribe               func(ctx context.Context, input *sns.UnsubscribeInput, opts []func(*sns.Options)) (*sns.UnsubscribeOutput, error)
	MockGetSubscriptionAttributes func(ctx context.Context, input *sns.GetSubscriptionAttributesInput, opts []func(*sns.Options)) (*sns.GetSubscriptionAttributesOutput, error)
	MockSetSubscriptionAttributes func(ctx context.Context, input *sns.SetSubscriptionAttributesInput, opts []func(*sns.Options)) (*sns.SetSubscriptionAttributesOutput, error)
}

// Subscribe mocks Subscribe method
func (m *MockSubscriptionClient) Subscribe(ctx context.Context, input *sns.SubscribeInput, opts ...func(*sns.Options)) (*sns.SubscribeOutput, error) {
	return m.MockSubscribe(ctx, input, opts)
}

// Unsubscribe mocks Unsubscribe method
func (m *MockSubscriptionClient) Unsubscribe(ctx context.Context, input *sns.UnsubscribeInput, opts ...func(*sns.Options)) (*sns.UnsubscribeOutput, error) {
	return m.MockUnsubscribe(ctx, input, opts)
}

// GetSubscriptionAttributes mocks GetSubscriptionAttributes method
func (m *MockSubscriptionClient) GetSubscriptionAttributes(ctx context.Context, input *sns.GetSubscriptionAttributesInput, opts ...func(*sns.Options)) (*sns.GetSubscriptionAttributesOutput, error) {
	return m.MockGetSubscriptionAttributes(ctx, input, opts)
}

// SetSubscriptionAttributes mocks SetSubscriptionAttributes method
func (m *MockSubscriptionClient) SetSubscriptionAttributes(ctx context.Context, input *sns.SetSubscriptionAttributesInput, opts ...func(*sns.Options)) (*sns.SetSubscriptionAttributesOutput, error) {
	return m.MockSetSubscriptionAttributes(ctx, input, opts)
}
