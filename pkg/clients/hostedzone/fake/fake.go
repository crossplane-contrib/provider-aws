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
	"github.com/aws/aws-sdk-go-v2/service/route53"
)

// MockHostedZoneClient is a type that implements all the methods for Hosted Zone Client interface
type MockHostedZoneClient struct {
	MockCreateHostedZoneRequest        func(input *route53.CreateHostedZoneInput) route53.CreateHostedZoneRequest
	MockDeleteHostedZoneRequest        func(input *route53.DeleteHostedZoneInput) route53.DeleteHostedZoneRequest
	MockGetHostedZoneRequest           func(input *route53.GetHostedZoneInput) route53.GetHostedZoneRequest
	MockUpdateHostedZoneCommentRequest func(input *route53.UpdateHostedZoneCommentInput) route53.UpdateHostedZoneCommentRequest
}

// GetHostedZoneRequest mocks GetHostedZoneRequest method
func (m *MockHostedZoneClient) GetHostedZoneRequest(input *route53.GetHostedZoneInput) route53.GetHostedZoneRequest {
	return m.MockGetHostedZoneRequest(input)
}

// CreateHostedZoneRequest mocks CreateHostedZoneRequest method
func (m *MockHostedZoneClient) CreateHostedZoneRequest(input *route53.CreateHostedZoneInput) route53.CreateHostedZoneRequest {
	return m.MockCreateHostedZoneRequest(input)
}

// UpdateHostedZoneCommentRequest mocks UpdateHostedZoneCommentRequest method
func (m *MockHostedZoneClient) UpdateHostedZoneCommentRequest(input *route53.UpdateHostedZoneCommentInput) route53.UpdateHostedZoneCommentRequest {
	return m.MockUpdateHostedZoneCommentRequest(input)
}

// DeleteHostedZoneRequest mocks DeleteHostedZoneRequest method
func (m *MockHostedZoneClient) DeleteHostedZoneRequest(input *route53.DeleteHostedZoneInput) route53.DeleteHostedZoneRequest {
	return m.MockDeleteHostedZoneRequest(input)
}
