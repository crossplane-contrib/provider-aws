// /*
// Copyright 2020 The Crossplane Authors.

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

// 	http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
// */

package fake

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/route53"
)

// MockHostedZoneClient is a type that implements all the methods for Hosted Zone Client interface
type MockHostedZoneClient struct {
	MockCreateHostedZone        func(ctx context.Context, input *route53.CreateHostedZoneInput, opts []func(*route53.Options)) (*route53.CreateHostedZoneOutput, error)
	MockDeleteHostedZone        func(ctx context.Context, input *route53.DeleteHostedZoneInput, opts []func(*route53.Options)) (*route53.DeleteHostedZoneOutput, error)
	MockGetHostedZone           func(ctx context.Context, input *route53.GetHostedZoneInput, opts []func(*route53.Options)) (*route53.GetHostedZoneOutput, error)
	MockUpdateHostedZoneComment func(ctx context.Context, input *route53.UpdateHostedZoneCommentInput, opts []func(*route53.Options)) (*route53.UpdateHostedZoneCommentOutput, error)
	MockListTagsForResource     func(ctx context.Context, params *route53.ListTagsForResourceInput, opts []func(*route53.Options)) (*route53.ListTagsForResourceOutput, error)
	MockChangeTagsForResource   func(ctx context.Context, params *route53.ChangeTagsForResourceInput, optFns []func(*route53.Options)) (*route53.ChangeTagsForResourceOutput, error)
}

// GetHostedZone mocks GetHostedZone method
func (m *MockHostedZoneClient) GetHostedZone(ctx context.Context, input *route53.GetHostedZoneInput, opts ...func(*route53.Options)) (*route53.GetHostedZoneOutput, error) {
	return m.MockGetHostedZone(ctx, input, opts)
}

// CreateHostedZone mocks CreateHostedZone method
func (m *MockHostedZoneClient) CreateHostedZone(ctx context.Context, input *route53.CreateHostedZoneInput, opts ...func(*route53.Options)) (*route53.CreateHostedZoneOutput, error) {
	return m.MockCreateHostedZone(ctx, input, opts)
}

// UpdateHostedZoneComment mocks UpdateHostedZoneComment method
func (m *MockHostedZoneClient) UpdateHostedZoneComment(ctx context.Context, input *route53.UpdateHostedZoneCommentInput, opts ...func(*route53.Options)) (*route53.UpdateHostedZoneCommentOutput, error) {
	return m.MockUpdateHostedZoneComment(ctx, input, opts)
}

// DeleteHostedZone mocks DeleteHostedZone method
func (m *MockHostedZoneClient) DeleteHostedZone(ctx context.Context, input *route53.DeleteHostedZoneInput, opts ...func(*route53.Options)) (*route53.DeleteHostedZoneOutput, error) {
	return m.MockDeleteHostedZone(ctx, input, opts)
}

// ListTagsForResource mocks ListTagsForResource method
func (m *MockHostedZoneClient) ListTagsForResource(ctx context.Context, input *route53.ListTagsForResourceInput, opts ...func(*route53.Options)) (*route53.ListTagsForResourceOutput, error) {
	return m.MockListTagsForResource(ctx, input, opts)
}

// ChangeTagsForResource mocks ChangeTagsForResource method
func (m *MockHostedZoneClient) ChangeTagsForResource(ctx context.Context, input *route53.ChangeTagsForResourceInput, opts ...func(*route53.Options)) (*route53.ChangeTagsForResourceOutput, error) {
	return m.MockChangeTagsForResource(ctx, input, opts)
}
