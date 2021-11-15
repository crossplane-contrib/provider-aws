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

	"github.com/aws/aws-sdk-go-v2/service/acm"

	clientset "github.com/crossplane/provider-aws/pkg/clients/acm"
)

// this ensures that the mock implements the client interface
var _ clientset.Client = (*MockCertificateClient)(nil)

// MockCertificateClient is a type that implements all the methods for Certificate Client interface
type MockCertificateClient struct {
	MockDescribeCertificate       func(context.Context, *acm.DescribeCertificateInput, []func(*acm.Options)) (*acm.DescribeCertificateOutput, error)
	MockAddTagsToCertificate      func(context.Context, *acm.AddTagsToCertificateInput, []func(*acm.Options)) (*acm.AddTagsToCertificateOutput, error)
	MockRequestCertificate        func(context.Context, *acm.RequestCertificateInput, []func(*acm.Options)) (*acm.RequestCertificateOutput, error)
	MockDeleteCertificate         func(context.Context, *acm.DeleteCertificateInput, []func(*acm.Options)) (*acm.DeleteCertificateOutput, error)
	MockUpdateCertificateOptions  func(context.Context, *acm.UpdateCertificateOptionsInput, []func(*acm.Options)) (*acm.UpdateCertificateOptionsOutput, error)
	MockListTagsForCertificate    func(context.Context, *acm.ListTagsForCertificateInput, []func(*acm.Options)) (*acm.ListTagsForCertificateOutput, error)
	MockRenewCertificate          func(context.Context, *acm.RenewCertificateInput, []func(*acm.Options)) (*acm.RenewCertificateOutput, error)
	MockRemoveTagsFromCertificate func(context.Context, *acm.RemoveTagsFromCertificateInput, []func(*acm.Options)) (*acm.RemoveTagsFromCertificateOutput, error)
}

// DescribeCertificate mocks DescribeCertificate method
func (m *MockCertificateClient) DescribeCertificate(ctx context.Context, input *acm.DescribeCertificateInput, opts ...func(*acm.Options)) (*acm.DescribeCertificateOutput, error) {
	return m.MockDescribeCertificate(ctx, input, opts)
}

// RequestCertificate mocks RequestCertificate method
func (m *MockCertificateClient) RequestCertificate(ctx context.Context, input *acm.RequestCertificateInput, opts ...func(*acm.Options)) (*acm.RequestCertificateOutput, error) {
	return m.MockRequestCertificate(ctx, input, opts)
}

// DeleteCertificate mocks DeleteCertificate method
func (m *MockCertificateClient) DeleteCertificate(ctx context.Context, input *acm.DeleteCertificateInput, opts ...func(*acm.Options)) (*acm.DeleteCertificateOutput, error) {
	return m.MockDeleteCertificate(ctx, input, opts)
}

// UpdateCertificateOptions mocks UpdateCertificateOptions method
func (m *MockCertificateClient) UpdateCertificateOptions(ctx context.Context, input *acm.UpdateCertificateOptionsInput, opts ...func(*acm.Options)) (*acm.UpdateCertificateOptionsOutput, error) {
	return m.MockUpdateCertificateOptions(ctx, input, opts)
}

// ListTagsForCertificate mocks ListTagsForCertificate method
func (m *MockCertificateClient) ListTagsForCertificate(ctx context.Context, input *acm.ListTagsForCertificateInput, opts ...func(*acm.Options)) (*acm.ListTagsForCertificateOutput, error) {
	return m.MockListTagsForCertificate(ctx, input, opts)
}

// RenewCertificate mocks RenewCertificate method
func (m *MockCertificateClient) RenewCertificate(ctx context.Context, input *acm.RenewCertificateInput, opts ...func(*acm.Options)) (*acm.RenewCertificateOutput, error) {
	return m.MockRenewCertificate(ctx, input, opts)
}

// RemoveTagsFromCertificate mocks RemoveTagsFromCertificate method
func (m *MockCertificateClient) RemoveTagsFromCertificate(ctx context.Context, input *acm.RemoveTagsFromCertificateInput, opts ...func(*acm.Options)) (*acm.RemoveTagsFromCertificateOutput, error) {
	return m.MockRemoveTagsFromCertificate(ctx, input, opts)
}

// AddTagsToCertificate mocks AddTagsToCertificate method
func (m *MockCertificateClient) AddTagsToCertificate(ctx context.Context, input *acm.AddTagsToCertificateInput, opts ...func(*acm.Options)) (*acm.AddTagsToCertificateOutput, error) {
	return m.MockAddTagsToCertificate(ctx, input, opts)
}
