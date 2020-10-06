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
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/crossplane/provider-aws/pkg/clients/iam"
	"github.com/crossplane/provider-aws/pkg/clients/iam/fake"
	clientset "github.com/crossplane/provider-aws/pkg/clients/s3"
)

// this ensures that the mock implements the client interface
var _ clientset.BucketPolicyClient = (*MockBucketPolicyClient)(nil)

// MockBucketPolicyClient is a type that implements all the methods for RolePolicyAttachmentClient interface
type MockBucketPolicyClient struct {
	MockGetBucketPolicyRequest    func(*s3.GetBucketPolicyInput) s3.GetBucketPolicyRequest
	MockPutBucketPolicyRequest    func(*s3.PutBucketPolicyInput) s3.PutBucketPolicyRequest
	MockDeleteBucketPolicyRequest func(*s3.DeleteBucketPolicyInput) s3.DeleteBucketPolicyRequest
}

// GetBucketPolicyRequest mocks GetBucketPolicyRequest method
func (m *MockBucketPolicyClient) GetBucketPolicyRequest(input *s3.GetBucketPolicyInput) s3.GetBucketPolicyRequest {
	return m.MockGetBucketPolicyRequest(input)
}

// PutBucketPolicyRequest mocks PutBucketPolicyRequest method
func (m *MockBucketPolicyClient) PutBucketPolicyRequest(input *s3.PutBucketPolicyInput) s3.PutBucketPolicyRequest {
	return m.MockPutBucketPolicyRequest(input)
}

// DeleteBucketPolicyRequest mocks DeleteBucketPolicyRequest method
func (m *MockBucketPolicyClient) DeleteBucketPolicyRequest(input *s3.DeleteBucketPolicyInput) s3.DeleteBucketPolicyRequest {
	return m.MockDeleteBucketPolicyRequest(input)
}

// NewMockBucketPolicyClient returns a new client given an aws config
func NewMockBucketPolicyClient(conf *aws.Config) (clientset.BucketPolicyClient, iam.Client, error) {
	s3client := MockBucketPolicyClient{}
	iamclient := fake.Client{}
	return &s3client, &iamclient, nil
}
