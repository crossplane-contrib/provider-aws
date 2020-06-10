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
	"github.com/aws/aws-sdk-go-v2/service/acmpca"

	clientset "github.com/crossplane/provider-aws/pkg/clients/certificatemanager/certificateauthority"
)

// this ensures that the mock implements the client interface
var _ clientset.Client = (*MockCertificateAuthorityClient)(nil)

// MockCertificateAuthorityClient is a type that implements all the methods for Certificate Authority Client interface
type MockCertificateAuthorityClient struct {
	MockCreateCertificateAuthorityRequest   func(*acmpca.CreateCertificateAuthorityInput) acmpca.CreateCertificateAuthorityRequest
	MockCreatePermissionRequest             func(*acmpca.CreatePermissionInput) acmpca.CreatePermissionRequest
	MockDeleteCertificateAuthorityRequest   func(*acmpca.DeleteCertificateAuthorityInput) acmpca.DeleteCertificateAuthorityRequest
	MockDeletePermissionRequest             func(*acmpca.DeletePermissionInput) acmpca.DeletePermissionRequest
	MockUpdateCertificateAuthorityRequest   func(*acmpca.UpdateCertificateAuthorityInput) acmpca.UpdateCertificateAuthorityRequest
	MockDescribeCertificateAuthorityRequest func(*acmpca.DescribeCertificateAuthorityInput) acmpca.DescribeCertificateAuthorityRequest
	MockListTagsRequest                     func(*acmpca.ListTagsInput) acmpca.ListTagsRequest
	MockUntagCertificateAuthorityRequest    func(*acmpca.UntagCertificateAuthorityInput) acmpca.UntagCertificateAuthorityRequest
	MockTagCertificateAuthorityRequest      func(*acmpca.TagCertificateAuthorityInput) acmpca.TagCertificateAuthorityRequest
}

// CreateCertificateAuthorityRequest mocks CreateCertificateAuthorityRequest method
func (m *MockCertificateAuthorityClient) CreateCertificateAuthorityRequest(input *acmpca.CreateCertificateAuthorityInput) acmpca.CreateCertificateAuthorityRequest {
	return m.MockCreateCertificateAuthorityRequest(input)
}

// CreatePermissionRequest mocks CreatePermissionRequest method
func (m *MockCertificateAuthorityClient) CreatePermissionRequest(input *acmpca.CreatePermissionInput) acmpca.CreatePermissionRequest {
	return m.MockCreatePermissionRequest(input)
}

// DeleteCertificateAuthorityRequest mocks DeleteCertificateAuthorityRequest method
func (m *MockCertificateAuthorityClient) DeleteCertificateAuthorityRequest(input *acmpca.DeleteCertificateAuthorityInput) acmpca.DeleteCertificateAuthorityRequest {
	return m.MockDeleteCertificateAuthorityRequest(input)
}

// TagCertificateAuthorityRequest mocks TagCertificateAuthorityRequest method
func (m *MockCertificateAuthorityClient) TagCertificateAuthorityRequest(input *acmpca.TagCertificateAuthorityInput) acmpca.TagCertificateAuthorityRequest {
	return m.MockTagCertificateAuthorityRequest(input)
}

// UntagCertificateAuthorityRequest mocks UntagCertificateAuthorityRequest method
func (m *MockCertificateAuthorityClient) UntagCertificateAuthorityRequest(input *acmpca.UntagCertificateAuthorityInput) acmpca.UntagCertificateAuthorityRequest {
	return m.MockUntagCertificateAuthorityRequest(input)
}

// ListTagsRequest mocks ListTagsRequest method
func (m *MockCertificateAuthorityClient) ListTagsRequest(input *acmpca.ListTagsInput) acmpca.ListTagsRequest {
	return m.MockListTagsRequest(input)
}

// DescribeCertificateAuthorityRequest mocks DescribeCertificateAuthorityRequest method
func (m *MockCertificateAuthorityClient) DescribeCertificateAuthorityRequest(input *acmpca.DescribeCertificateAuthorityInput) acmpca.DescribeCertificateAuthorityRequest {
	return m.MockDescribeCertificateAuthorityRequest(input)
}

// UpdateCertificateAuthorityRequest mocks UpdateCertificateAuthorityRequest method
func (m *MockCertificateAuthorityClient) UpdateCertificateAuthorityRequest(input *acmpca.UpdateCertificateAuthorityInput) acmpca.UpdateCertificateAuthorityRequest {
	return m.MockUpdateCertificateAuthorityRequest(input)
}

// DeletePermissionRequest mocks DeletePermissionRequest method
func (m *MockCertificateAuthorityClient) DeletePermissionRequest(input *acmpca.DeletePermissionInput) acmpca.DeletePermissionRequest {
	return m.MockDeletePermissionRequest(input)
}
