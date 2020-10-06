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

package s3

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/awserr"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// BucketPolicyClient is the external client used for S3BucketPolicy Custom Resource
type BucketPolicyClient interface {
	GetBucketPolicyRequest(input *s3.GetBucketPolicyInput) s3.GetBucketPolicyRequest
	PutBucketPolicyRequest(input *s3.PutBucketPolicyInput) s3.PutBucketPolicyRequest
	DeleteBucketPolicyRequest(input *s3.DeleteBucketPolicyInput) s3.DeleteBucketPolicyRequest
}

// NewBucketPolicyClient returns a new client given an aws config
func NewBucketPolicyClient(cfg aws.Config) BucketPolicyClient {
	return s3.New(cfg)
}

// IsErrorPolicyNotFound returns true if the error code indicates that the item was not found
func IsErrorPolicyNotFound(err error) bool {
	if s3Err, ok := err.(awserr.Error); ok && s3Err.Code() == "NoSuchBucketPolicy" {
		return true
	}
	return false
}

// IsErrorBucketNotFound returns true if the error code indicates that the bucket was not found
func IsErrorBucketNotFound(err error) bool {
	if s3Err, ok := err.(awserr.Error); ok && s3Err.Code() == s3.ErrCodeNoSuchBucket {
		return true
	}
	return false
}
