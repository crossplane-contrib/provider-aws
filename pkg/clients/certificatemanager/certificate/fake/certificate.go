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
	"github.com/aws/aws-sdk-go-v2/service/acm"

	clientset "github.com/crossplane/provider-aws/pkg/clients/certificatemanager/certificate"
)

// this ensures that the mock implements the client interface
var _ clientset.Client = (*MockCertificateClient)(nil)

// MockCertificateClient is a type that implements all the methods for Certificate Client interface
type MockCertificateClient struct {
	MockDescribeCertificateRequest       func(*acm.DescribeCertificateInput) acm.DescribeCertificateRequest
	MockAddTagsToCertificateRequest      func(*acm.AddTagsToCertificateInput) acm.AddTagsToCertificateRequest
	MockRequestCertificateRequest        func(*acm.RequestCertificateInput) acm.RequestCertificateRequest
	MockDeleteCertificateRequest         func(*acm.DeleteCertificateInput) acm.DeleteCertificateRequest
	MockUpdateCertificateOptionsRequest  func(*acm.UpdateCertificateOptionsInput) acm.UpdateCertificateOptionsRequest
	MockListTagsForCertificateRequest    func(*acm.ListTagsForCertificateInput) acm.ListTagsForCertificateRequest
	MockRenewCertificateRequest          func(*acm.RenewCertificateInput) acm.RenewCertificateRequest
	MockRemoveTagsFromCertificateRequest func(*acm.RemoveTagsFromCertificateInput) acm.RemoveTagsFromCertificateRequest
}

// DescribeCertificateRequest mocks DescribeCertificateRequest method
func (m *MockCertificateClient) DescribeCertificateRequest(input *acm.DescribeCertificateInput) acm.DescribeCertificateRequest {
	return m.MockDescribeCertificateRequest(input)
}

// RequestCertificateRequest mocks RequestCertificateRequest method
func (m *MockCertificateClient) RequestCertificateRequest(input *acm.RequestCertificateInput) acm.RequestCertificateRequest {
	return m.MockRequestCertificateRequest(input)
}

// DeleteCertificateRequest mocks DeleteCertificateRequest method
func (m *MockCertificateClient) DeleteCertificateRequest(input *acm.DeleteCertificateInput) acm.DeleteCertificateRequest {
	return m.MockDeleteCertificateRequest(input)
}

// UpdateCertificateOptionsRequest mocks UpdateCertificateOptionsRequest method
func (m *MockCertificateClient) UpdateCertificateOptionsRequest(input *acm.UpdateCertificateOptionsInput) acm.UpdateCertificateOptionsRequest {
	return m.MockUpdateCertificateOptionsRequest(input)
}

// ListTagsForCertificateRequest mocks ListTagsForCertificateRequest method
func (m *MockCertificateClient) ListTagsForCertificateRequest(input *acm.ListTagsForCertificateInput) acm.ListTagsForCertificateRequest {
	return m.MockListTagsForCertificateRequest(input)
}

// RenewCertificateRequest mocks RenewCertificateRequest method
func (m *MockCertificateClient) RenewCertificateRequest(input *acm.RenewCertificateInput) acm.RenewCertificateRequest {
	return m.MockRenewCertificateRequest(input)
}

// RemoveTagsFromCertificateRequest mocks RemoveTagsFromCertificateRequest method
func (m *MockCertificateClient) RemoveTagsFromCertificateRequest(input *acm.RemoveTagsFromCertificateInput) acm.RemoveTagsFromCertificateRequest {
	return m.MockRemoveTagsFromCertificateRequest(input)
}

// AddTagsToCertificateRequest mocks AddTagsToCertificateRequest method
func (m *MockCertificateClient) AddTagsToCertificateRequest(input *acm.AddTagsToCertificateInput) acm.AddTagsToCertificateRequest {
	return m.MockAddTagsToCertificateRequest(input)
}
