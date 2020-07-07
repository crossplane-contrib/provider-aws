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

	"github.com/aws/aws-sdk-go-v2/service/route53"

	"github.com/crossplane/provider-aws/apis/route53/v1alpha1"
	rr "github.com/crossplane/provider-aws/pkg/clients/resourcerecordset"
)

// MockResourceRecordSetClient is a type that implements all the methods for Resource Record Client interface
type MockResourceRecordSetClient struct {
	MockChangeResourceRecordSetsRequest       func(*route53.ChangeResourceRecordSetsInput) route53.ChangeResourceRecordSetsRequest
	MockListResourceRecordSetsRequest         func(*route53.ListResourceRecordSetsInput) route53.ListResourceRecordSetsRequest
	MockGetResourceRecordSet                  func(ctx context.Context, c rr.Client, id, rrName, si *string) (route53.ResourceRecordSet, error)
	MockGenerateChangeResourceRecordSetsInput func(p *v1alpha1.ResourceRecordSetParameters, action route53.ChangeAction) *route53.ChangeResourceRecordSetsInput
}

// ChangeResourceRecordSetsRequest mocks ChangeResourceRecordSetsRequest method
func (m *MockResourceRecordSetClient) ChangeResourceRecordSetsRequest(input *route53.ChangeResourceRecordSetsInput) route53.ChangeResourceRecordSetsRequest {
	return m.MockChangeResourceRecordSetsRequest(input)
}

// ListResourceRecordSetsRequest mocks ListResourceRecordSetsRequest method
func (m *MockResourceRecordSetClient) ListResourceRecordSetsRequest(input *route53.ListResourceRecordSetsInput) route53.ListResourceRecordSetsRequest {
	return m.MockListResourceRecordSetsRequest(input)
}

// GetResourceRecordSet mocks GetResourceRecordSet method
func (m *MockResourceRecordSetClient) GetResourceRecordSet(ctx context.Context, c rr.Client, id, rrName, si *string) (route53.ResourceRecordSet, error) {
	return m.MockGetResourceRecordSet(ctx, c, id, rrName, si)
}

// GenerateChangeResourceRecordSetsInput mocks GenerateChangeResourceRecordSetsInput method
func (m *MockResourceRecordSetClient) GenerateChangeResourceRecordSetsInput(p *v1alpha1.ResourceRecordSetParameters, action route53.ChangeAction) *route53.ChangeResourceRecordSetsInput {
	return m.MockGenerateChangeResourceRecordSetsInput(p, action)
}
