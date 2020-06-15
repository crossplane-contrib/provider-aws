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

	clientset "github.com/crossplane/provider-aws/pkg/clients/certificatemanager/certificateauthoritypermission"
)

// this ensures that the mock implements the client interface
var _ clientset.Client = (*MockCertificateAuthorityPermissionClient)(nil)

// MockCertificateAuthorityPermissionClient is a type that implements all the methods for Certificate Authority Permission Client interface
type MockCertificateAuthorityPermissionClient struct {
	MockCreatePermissionRequest func(*acmpca.CreatePermissionInput) acmpca.CreatePermissionRequest
	MockDeletePermissionRequest func(*acmpca.DeletePermissionInput) acmpca.DeletePermissionRequest
	MockListPermissionsRequest  func(*acmpca.ListPermissionsInput) acmpca.ListPermissionsRequest
}

// CreatePermissionRequest mocks CreatePermissionRequest method
func (m *MockCertificateAuthorityPermissionClient) CreatePermissionRequest(input *acmpca.CreatePermissionInput) acmpca.CreatePermissionRequest {
	return m.MockCreatePermissionRequest(input)
}

// DeletePermissionRequest mocks DeletePermissionRequest method
func (m *MockCertificateAuthorityPermissionClient) DeletePermissionRequest(input *acmpca.DeletePermissionInput) acmpca.DeletePermissionRequest {
	return m.MockDeletePermissionRequest(input)
}

// ListPermissionsRequest mocks ListPermissionsRequest method
func (m *MockCertificateAuthorityPermissionClient) ListPermissionsRequest(input *acmpca.ListPermissionsInput) acmpca.ListPermissionsRequest {
	return m.MockListPermissionsRequest(input)
}
