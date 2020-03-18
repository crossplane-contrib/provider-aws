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
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// MockS3Client for testing.
type MockS3Client struct {
	MockHeadBucket      func(*s3.HeadBucketInput) s3.HeadBucketRequest
	MockCreateBucket    func(*s3.CreateBucketInput) s3.CreateBucketRequest
	MockGetBucketPolicy func(*s3.GetBucketPolicyInput) s3.GetBucketPolicyRequest
	MockPutBucketPolicy func(*s3.PutBucketPolicyInput) s3.PutBucketPolicyRequest
	MockDeleteBucket    func(*s3.DeleteBucketInput) s3.DeleteBucketRequest
}

// HeadBucketRequest checks if bucket exists.
func (m *MockS3Client) HeadBucketRequest(i *s3.HeadBucketInput) s3.HeadBucketRequest {
	return m.MockHeadBucket(i)
}

// CreateBucketRequest creates a bucket.
func (m *MockS3Client) CreateBucketRequest(i *s3.CreateBucketInput) s3.CreateBucketRequest {
	return m.MockCreateBucket(i)
}

// GetBucketPolicyRequest gets the policy document of bucket.
func (m *MockS3Client) GetBucketPolicyRequest(i *s3.GetBucketPolicyInput) s3.GetBucketPolicyRequest {
	return m.MockGetBucketPolicy(i)
}

// PutBucketPolicyRequest creates the policy document of bucket.
func (m *MockS3Client) PutBucketPolicyRequest(i *s3.PutBucketPolicyInput) s3.PutBucketPolicyRequest {
	return m.MockPutBucketPolicy(i)
}

// DeleteBucketRequest deletes a bucket.
func (m *MockS3Client) DeleteBucketRequest(i *s3.DeleteBucketInput) s3.DeleteBucketRequest {
	return m.MockDeleteBucket(i)
}
