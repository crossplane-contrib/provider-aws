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
)

// MockResourceRecordSetClient is a type that implements all the methods for Resource Record Client interface
type MockResourceRecordSetClient struct {
	MockChangeResourceRecordSets func(ctx context.Context, input *route53.ChangeResourceRecordSetsInput, opts []func(*route53.Options)) (*route53.ChangeResourceRecordSetsOutput, error)
	MockListResourceRecordSets   func(ctx context.Context, input *route53.ListResourceRecordSetsInput, opts []func(*route53.Options)) (*route53.ListResourceRecordSetsOutput, error)
}

// ChangeResourceRecordSets mocks ChangeResourceRecordSets method
func (m *MockResourceRecordSetClient) ChangeResourceRecordSets(ctx context.Context, input *route53.ChangeResourceRecordSetsInput, opts ...func(*route53.Options)) (*route53.ChangeResourceRecordSetsOutput, error) {
	return m.MockChangeResourceRecordSets(ctx, input, opts)
}

// ListResourceRecordSets mocks ListResourceRecordSets method
func (m *MockResourceRecordSetClient) ListResourceRecordSets(ctx context.Context, input *route53.ListResourceRecordSetsInput, opts ...func(*route53.Options)) (*route53.ListResourceRecordSetsOutput, error) {
	return m.MockListResourceRecordSets(ctx, input, opts)
}
