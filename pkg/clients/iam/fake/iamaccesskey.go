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
	"github.com/aws/aws-sdk-go-v2/service/iam"

	clientset "github.com/crossplane/provider-aws/pkg/clients/iam"
)

// this ensures that the mock implements the client interface
var _ clientset.AccessClient = (*MockAccessClient)(nil)

// MockAccessClient is a type that implements all the methods for AccessClient interface
type MockAccessClient struct {
	MockCreateAccessKeyRequest func(*iam.CreateAccessKeyInput) iam.CreateAccessKeyRequest
	MockDeleteAccessKeyRequest func(*iam.DeleteAccessKeyInput) iam.DeleteAccessKeyRequest
	MockListAccessKeysRequest  func(*iam.ListAccessKeysInput) iam.ListAccessKeysRequest
	MockUpdateAccessKeyRequest func(*iam.UpdateAccessKeyInput) iam.UpdateAccessKeyRequest
}

// UpdateAccessKeyRequest mocks UpdateAccessKeyRequest method
func (m MockAccessClient) UpdateAccessKeyRequest(input *iam.UpdateAccessKeyInput) iam.UpdateAccessKeyRequest {
	return m.MockUpdateAccessKeyRequest(input)
}

// ListAccessKeysRequest mocks ListAccessKeysRequest method
func (m MockAccessClient) ListAccessKeysRequest(input *iam.ListAccessKeysInput) iam.ListAccessKeysRequest {
	return m.MockListAccessKeysRequest(input)
}

// CreateAccessKeyRequest mocks CreateAccessKeyRequest method
func (m MockAccessClient) CreateAccessKeyRequest(input *iam.CreateAccessKeyInput) iam.CreateAccessKeyRequest {
	return m.MockCreateAccessKeyRequest(input)
}

// DeleteAccessKeyRequest mocks DeleteAccessKeyRequest method
func (m MockAccessClient) DeleteAccessKeyRequest(input *iam.DeleteAccessKeyInput) iam.DeleteAccessKeyRequest {
	return m.MockDeleteAccessKeyRequest(input)
}
