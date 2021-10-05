/*
Copyright 2020 The Crossplane Authors.

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

	"github.com/aws/aws-sdk-go-v2/service/iam"

	clientset "github.com/crossplane/provider-aws/pkg/clients/iam"
)

// this ensures that the mock implements the client interface
var _ clientset.AccessClient = (*MockAccessClient)(nil)

// MockAccessClient is a type that implements all the methods for AccessClient interface
type MockAccessClient struct {
	MockCreateAccessKey func(ctx context.Context, input *iam.CreateAccessKeyInput, opts []func(*iam.Options)) (*iam.CreateAccessKeyOutput, error)
	MockDeleteAccessKey func(ctx context.Context, input *iam.DeleteAccessKeyInput, opts []func(*iam.Options)) (*iam.DeleteAccessKeyOutput, error)
	MockListAccessKeys  func(ctx context.Context, input *iam.ListAccessKeysInput, opts []func(*iam.Options)) (*iam.ListAccessKeysOutput, error)
	MockUpdateAccessKey func(ctx context.Context, input *iam.UpdateAccessKeyInput, opts []func(*iam.Options)) (*iam.UpdateAccessKeyOutput, error)
}

// UpdateAccessKey mocks UpdateAccessKey method
func (m MockAccessClient) UpdateAccessKey(ctx context.Context, input *iam.UpdateAccessKeyInput, opts ...func(*iam.Options)) (*iam.UpdateAccessKeyOutput, error) {
	return m.MockUpdateAccessKey(ctx, input, opts)
}

// ListAccessKeys mocks ListAccessKeys method
func (m MockAccessClient) ListAccessKeys(ctx context.Context, input *iam.ListAccessKeysInput, opts ...func(*iam.Options)) (*iam.ListAccessKeysOutput, error) {
	return m.MockListAccessKeys(ctx, input, opts)
}

// CreateAccessKey mocks CreateAccessKey method
func (m MockAccessClient) CreateAccessKey(ctx context.Context, input *iam.CreateAccessKeyInput, opts ...func(*iam.Options)) (*iam.CreateAccessKeyOutput, error) {
	return m.MockCreateAccessKey(ctx, input, opts)
}

// DeleteAccessKey mocks DeleteAccessKey method
func (m MockAccessClient) DeleteAccessKey(ctx context.Context, input *iam.DeleteAccessKeyInput, opts ...func(*iam.Options)) (*iam.DeleteAccessKeyOutput, error) {
	return m.MockDeleteAccessKey(ctx, input, opts)
}
