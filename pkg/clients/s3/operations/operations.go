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

package operations

import "github.com/aws/aws-sdk-go-v2/service/s3"

// Operations defines common methods for generating bucket requests
// mockery -case snake -name Operations -output fake -outpkg fake
type Operations interface {
	CreateBucketRequest(*s3.CreateBucketInput) CreateBucketRequest
	GetBucketVersioningRequest(*s3.GetBucketVersioningInput) GetBucketVersioningRequest
	PutBucketACLRequest(*s3.PutBucketAclInput) PutBucketACLRequest
	PutBucketVersioningRequest(*s3.PutBucketVersioningInput) PutBucketVersioningRequest
	DeleteBucketRequest(*s3.DeleteBucketInput) DeleteBucketRequest
	PutBucketTaggingRequest(*s3.PutBucketTaggingInput) PutBucketTaggingRequest
}
