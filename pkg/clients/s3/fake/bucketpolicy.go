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

	"github.com/aws/aws-sdk-go-v2/service/s3"

	clientset "github.com/crossplane-contrib/provider-aws/pkg/clients/s3"
)

// this ensures that the mock implements the client interface
var _ clientset.BucketPolicyClient = (*MockBucketPolicyClient)(nil)

// MockBucketPolicyClient is a type that implements all the methods for RolePolicyAttachmentClient interface
type MockBucketPolicyClient struct {
	MockGetBucketPolicy    func(ctx context.Context, input *s3.GetBucketPolicyInput, opts []func(*s3.Options)) (*s3.GetBucketPolicyOutput, error)
	MockPutBucketPolicy    func(ctx context.Context, input *s3.PutBucketPolicyInput, opts []func(*s3.Options)) (*s3.PutBucketPolicyOutput, error)
	MockDeleteBucketPolicy func(ctx context.Context, input *s3.DeleteBucketPolicyInput, opts []func(*s3.Options)) (*s3.DeleteBucketPolicyOutput, error)
}

// GetBucketPolicy mocks GetBucketPolicy method
func (m MockBucketPolicyClient) GetBucketPolicy(ctx context.Context, input *s3.GetBucketPolicyInput, opts ...func(*s3.Options)) (*s3.GetBucketPolicyOutput, error) {
	return m.MockGetBucketPolicy(ctx, input, opts)
}

// PutBucketPolicy mocks PutBucketPolicy method
func (m MockBucketPolicyClient) PutBucketPolicy(ctx context.Context, input *s3.PutBucketPolicyInput, opts ...func(*s3.Options)) (*s3.PutBucketPolicyOutput, error) {
	return m.MockPutBucketPolicy(ctx, input, opts)
}

// DeleteBucketPolicy mocks DeleteBucketPolicy method
func (m MockBucketPolicyClient) DeleteBucketPolicy(ctx context.Context, input *s3.DeleteBucketPolicyInput, opts ...func(*s3.Options)) (*s3.DeleteBucketPolicyOutput, error) {
	return m.MockDeleteBucketPolicy(ctx, input, opts)
}
