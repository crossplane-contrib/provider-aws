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
var _ clientset.Client = (*MockCertificateAuthorityClient)(nil)

// MockCertificateAuthorityClient is a type that implements all the methods for Certificate Authority Client interface
type MockCertificateAuthorityClient struct {
	MockCreateCertificateAuthority   func(context.Context, *acmpca.CreateCertificateAuthorityInput, []func(*acmpca.Options)) (*acmpca.CreateCertificateAuthorityOutput, error)
	MockCreatePermission             func(context.Context, *acmpca.CreatePermissionInput, []func(*acmpca.Options)) (*acmpca.CreatePermissionOutput, error)
	MockDeleteCertificateAuthority   func(context.Context, *acmpca.DeleteCertificateAuthorityInput, []func(*acmpca.Options)) (*acmpca.DeleteCertificateAuthorityOutput, error)
	MockDeletePermission             func(context.Context, *acmpca.DeletePermissionInput, []func(*acmpca.Options)) (*acmpca.DeletePermissionOutput, error)
	MockUpdateCertificateAuthority   func(context.Context, *acmpca.UpdateCertificateAuthorityInput, []func(*acmpca.Options)) (*acmpca.UpdateCertificateAuthorityOutput, error)
	MockDescribeCertificateAuthority func(context.Context, *acmpca.DescribeCertificateAuthorityInput, []func(*acmpca.Options)) (*acmpca.DescribeCertificateAuthorityOutput, error)
	MockListTags                     func(context.Context, *acmpca.ListTagsInput, []func(*acmpca.Options)) (*acmpca.ListTagsOutput, error)
	MockUntagCertificateAuthority    func(context.Context, *acmpca.UntagCertificateAuthorityInput, []func(*acmpca.Options)) (*acmpca.UntagCertificateAuthorityOutput, error)
	MockTagCertificateAuthority      func(context.Context, *acmpca.TagCertificateAuthorityInput, []func(*acmpca.Options)) (*acmpca.TagCertificateAuthorityOutput, error)
}

// CreateCertificateAuthority mocks CreateCertificateAuthority method
func (m *MockCertificateAuthorityClient) CreateCertificateAuthority(ctx context.Context, input *acmpca.CreateCertificateAuthorityInput, opts ...func(*acmpca.Options)) (*acmpca.CreateCertificateAuthorityOutput, error) {
	return m.MockCreateCertificateAuthority(ctx, input, opts)
}

// CreatePermission mocks CreatePermission method
func (m *MockCertificateAuthorityClient) CreatePermission(ctx context.Context, input *acmpca.CreatePermissionInput, opts ...func(*acmpca.Options)) (*acmpca.CreatePermissionOutput, error) {
	return m.MockCreatePermission(ctx, input, opts)
}

// DeleteCertificateAuthority mocks DeleteCertificateAuthority method
func (m *MockCertificateAuthorityClient) DeleteCertificateAuthority(ctx context.Context, input *acmpca.DeleteCertificateAuthorityInput, opts ...func(*acmpca.Options)) (*acmpca.DeleteCertificateAuthorityOutput, error) {
	return m.MockDeleteCertificateAuthority(ctx, input, opts)
}

// TagCertificateAuthority mocks TagCertificateAuthority method
func (m *MockCertificateAuthorityClient) TagCertificateAuthority(ctx context.Context, input *acmpca.TagCertificateAuthorityInput, opts ...func(*acmpca.Options)) (*acmpca.TagCertificateAuthorityOutput, error) {
	return m.MockTagCertificateAuthority(ctx, input, opts)
}

// UntagCertificateAuthority mocks UntagCertificateAuthority method
func (m *MockCertificateAuthorityClient) UntagCertificateAuthority(ctx context.Context, input *acmpca.UntagCertificateAuthorityInput, opts ...func(*acmpca.Options)) (*acmpca.UntagCertificateAuthorityOutput, error) {
	return m.MockUntagCertificateAuthority(ctx, input, opts)
}

// ListTags mocks ListTags method
func (m *MockCertificateAuthorityClient) ListTags(ctx context.Context, input *acmpca.ListTagsInput, opts ...func(*acmpca.Options)) (*acmpca.ListTagsOutput, error) {
	return m.MockListTags(ctx, input, opts)
}

// DescribeCertificateAuthority mocks DescribeCertificateAuthority method
func (m *MockCertificateAuthorityClient) DescribeCertificateAuthority(ctx context.Context, input *acmpca.DescribeCertificateAuthorityInput, opts ...func(*acmpca.Options)) (*acmpca.DescribeCertificateAuthorityOutput, error) {
	return m.MockDescribeCertificateAuthority(ctx, input, opts)
}

// UpdateCertificateAuthority mocks UpdateCertificateAuthority method
func (m *MockCertificateAuthorityClient) UpdateCertificateAuthority(ctx context.Context, input *acmpca.UpdateCertificateAuthorityInput, opts ...func(*acmpca.Options)) (*acmpca.UpdateCertificateAuthorityOutput, error) {
	return m.MockUpdateCertificateAuthority(ctx, input, opts)
}

// DeletePermission mocks DeletePermission method
func (m *MockCertificateAuthorityClient) DeletePermission(ctx context.Context, input *acmpca.DeletePermissionInput, opts ...func(*acmpca.Options)) (*acmpca.DeletePermissionOutput, error) {
	return m.MockDeletePermission(ctx, input, opts)
}
