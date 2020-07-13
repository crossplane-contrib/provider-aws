package s3

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/awserr"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/crossplane/provider-aws/pkg/clients/iam"
)

// BucketPolicyClient is the external client used for S3BucketPolicy Custom Resource
type BucketPolicyClient interface {
	GetBucketPolicyRequest(input *s3.GetBucketPolicyInput) s3.GetBucketPolicyRequest
	PutBucketPolicyRequest(input *s3.PutBucketPolicyInput) s3.PutBucketPolicyRequest
	DeleteBucketPolicyRequest(input *s3.DeleteBucketPolicyInput) s3.DeleteBucketPolicyRequest
}

// NewBucketPolicyClient returns a new client given an aws config
func NewBucketPolicyClient(conf *aws.Config) (BucketPolicyClient, iam.Client, error) {
	s3client := s3.New(*conf)
	iamclient := iam.NewClient(conf)
	return s3client, iamclient, nil
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
