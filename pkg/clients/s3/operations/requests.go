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

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// mockery -case snake -all -output fake -outpkg fake

// CreateBucketRequest is a API request type for the CreateBucket API operation.
type CreateBucketRequest interface {
	Send(context.Context) (*s3.CreateBucketResponse, error)
}

// GetBucketVersioningRequest is a API request type for the GetBucketVersioning API operation.
type GetBucketVersioningRequest interface {
	Send(context.Context) (*s3.GetBucketVersioningResponse, error)
}

// PutBucketACLRequest is a API request type for the PutBucketAcl API operation.
type PutBucketACLRequest interface {
	Send(context.Context) (*s3.PutBucketAclResponse, error)
}

// PutBucketVersioningRequest is a API request type for the PutBucketVersioning API operation.
type PutBucketVersioningRequest interface {
	Send(context.Context) (*s3.PutBucketVersioningResponse, error)
}

// PutBucketTaggingRequest is a API request type for the PutBucketTagging API operation.
type PutBucketTaggingRequest interface {
	Send(context.Context) (*s3.PutBucketTaggingResponse, error)
}

// DeleteBucketRequest is a API request type for the DeleteBucket API operation.
type DeleteBucketRequest interface {
	Send(context.Context) (*s3.DeleteBucketResponse, error)
}
