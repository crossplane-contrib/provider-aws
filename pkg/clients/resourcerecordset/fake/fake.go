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
	"github.com/aws/aws-sdk-go-v2/service/route53"
)

// MockResourceRecordSetClient is a type that implements all the methods for Resource Record Client interface
type MockResourceRecordSetClient struct {
	MockChangeResourceRecordSetsRequest func(*route53.ChangeResourceRecordSetsInput) route53.ChangeResourceRecordSetsRequest
	MockListResourceRecordSetsRequest   func(*route53.ListResourceRecordSetsInput) route53.ListResourceRecordSetsRequest
}

// ChangeResourceRecordSetsRequest mocks ChangeResourceRecordSetsRequest method
func (m *MockResourceRecordSetClient) ChangeResourceRecordSetsRequest(input *route53.ChangeResourceRecordSetsInput) route53.ChangeResourceRecordSetsRequest {
	return m.MockChangeResourceRecordSetsRequest(input)
}

// ListResourceRecordSetsRequest mocks ListResourceRecordSetsRequest method
func (m *MockResourceRecordSetClient) ListResourceRecordSetsRequest(input *route53.ListResourceRecordSetsInput) route53.ListResourceRecordSetsRequest {
	return m.MockListResourceRecordSetsRequest(input)
}
