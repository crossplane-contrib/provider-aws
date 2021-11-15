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

	"github.com/aws/aws-sdk-go-v2/service/acmpca"

	clientset "github.com/crossplane/provider-aws/pkg/clients/acmpca"
)

// this ensures that the mock implements the client interface
var _ clientset.CAPermissionClient = (*MockCertificateAuthorityPermissionClient)(nil)

// MockCertificateAuthorityPermissionClient is a type that implements all the methods for Certificate Authority Permission Client interface
type MockCertificateAuthorityPermissionClient struct {
	MockCreatePermission func(context.Context, *acmpca.CreatePermissionInput, []func(*acmpca.Options)) (*acmpca.CreatePermissionOutput, error)
	MockDeletePermission func(context.Context, *acmpca.DeletePermissionInput, []func(*acmpca.Options)) (*acmpca.DeletePermissionOutput, error)
	MockListPermissions  func(context.Context, *acmpca.ListPermissionsInput, []func(*acmpca.Options)) (*acmpca.ListPermissionsOutput, error)
}

// CreatePermission mocks CreatePermission method
func (m *MockCertificateAuthorityPermissionClient) CreatePermission(ctx context.Context, input *acmpca.CreatePermissionInput, opts ...func(*acmpca.Options)) (*acmpca.CreatePermissionOutput, error) {
	return m.MockCreatePermission(ctx, input, opts)
}

// DeletePermission mocks DeletePermission method
func (m *MockCertificateAuthorityPermissionClient) DeletePermission(ctx context.Context, input *acmpca.DeletePermissionInput, opts ...func(*acmpca.Options)) (*acmpca.DeletePermissionOutput, error) {
	return m.MockDeletePermission(ctx, input, opts)
}

// ListPermissions mocks ListPermissions method
func (m *MockCertificateAuthorityPermissionClient) ListPermissions(ctx context.Context, input *acmpca.ListPermissionsInput, opts ...func(*acmpca.Options)) (*acmpca.ListPermissionsOutput, error) {
	return m.MockListPermissions(ctx, input, opts)
}
