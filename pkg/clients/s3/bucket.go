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
	"context"
	"fmt"
	"sort"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/awserr"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/crossplane/crossplane-runtime/pkg/meta"

	"github.com/crossplane/provider-aws/apis/s3/v1beta1"
)

var (
	// BucketNotFoundErrCode is the error code sent by AWS when a bucket does not exist
	BucketNotFoundErrCode = "NotFound"
	// CORSNotFoundErrCode is the error code sent by AWS when the CORS configuration does not exist
	CORSNotFoundErrCode = "NoSuchCORSConfiguration"
	// PublicAccessBlockNotFoundErrCode is NotFound error for PublicAccessBlock
	PublicAccessBlockNotFoundErrCode = "NoSuchPublicAccessBlockConfiguration"
	// ReplicationNotFoundErrCode is the error code sent by AWS when the replication config does not exist
	ReplicationNotFoundErrCode = "ReplicationConfigurationNotFoundError"
	// LifecycleNotFoundErrCode is the error code sent by AWS when the lifecycle config does not exist
	LifecycleNotFoundErrCode = "NoSuchLifecycleConfiguration"
	// SSENotFoundErrCode is the error code sent by AWS when the SSE config does not exist
	SSENotFoundErrCode = "ServerSideEncryptionConfigurationNotFoundError"
	// TaggingNotFoundErrCode is the error code sent by AWS when the tagging does not exist
	TaggingNotFoundErrCode = "NoSuchTagSet"
	// WebsiteNotFoundErrCode is the error code sent by AWS when the website config does not exist
	WebsiteNotFoundErrCode = "NoSuchWebsiteConfiguration"
	// MethodNotAllowed is the error code sent by AWS when the request method for an object is not allowed
	MethodNotAllowed = "MethodNotAllowed"
	// UnsupportedArgument is the error code sent by AWS when the request fields contain an argument that is not supported
	UnsupportedArgument = "UnsupportedArgument"
)

// BucketClient is the interface for Client for making S3 Bucket requests.
type BucketClient interface {
	HeadBucketRequest(input *s3.HeadBucketInput) s3.HeadBucketRequest
	CreateBucketRequest(input *s3.CreateBucketInput) s3.CreateBucketRequest
	DeleteBucketRequest(input *s3.DeleteBucketInput) s3.DeleteBucketRequest

	PutBucketEncryptionRequest(input *s3.PutBucketEncryptionInput) s3.PutBucketEncryptionRequest
	GetBucketEncryptionRequest(input *s3.GetBucketEncryptionInput) s3.GetBucketEncryptionRequest
	DeleteBucketEncryptionRequest(input *s3.DeleteBucketEncryptionInput) s3.DeleteBucketEncryptionRequest

	PutBucketVersioningRequest(input *s3.PutBucketVersioningInput) s3.PutBucketVersioningRequest
	GetBucketVersioningRequest(input *s3.GetBucketVersioningInput) s3.GetBucketVersioningRequest

	PutBucketAccelerateConfigurationRequest(input *s3.PutBucketAccelerateConfigurationInput) s3.PutBucketAccelerateConfigurationRequest
	GetBucketAccelerateConfigurationRequest(input *s3.GetBucketAccelerateConfigurationInput) s3.GetBucketAccelerateConfigurationRequest

	PutBucketCorsRequest(input *s3.PutBucketCorsInput) s3.PutBucketCorsRequest
	GetBucketCorsRequest(input *s3.GetBucketCorsInput) s3.GetBucketCorsRequest
	DeleteBucketCorsRequest(input *s3.DeleteBucketCorsInput) s3.DeleteBucketCorsRequest

	PutBucketWebsiteRequest(input *s3.PutBucketWebsiteInput) s3.PutBucketWebsiteRequest
	GetBucketWebsiteRequest(input *s3.GetBucketWebsiteInput) s3.GetBucketWebsiteRequest
	DeleteBucketWebsiteRequest(input *s3.DeleteBucketWebsiteInput) s3.DeleteBucketWebsiteRequest

	PutBucketLoggingRequest(input *s3.PutBucketLoggingInput) s3.PutBucketLoggingRequest
	GetBucketLoggingRequest(input *s3.GetBucketLoggingInput) s3.GetBucketLoggingRequest

	PutBucketReplicationRequest(input *s3.PutBucketReplicationInput) s3.PutBucketReplicationRequest
	GetBucketReplicationRequest(input *s3.GetBucketReplicationInput) s3.GetBucketReplicationRequest
	DeleteBucketReplicationRequest(input *s3.DeleteBucketReplicationInput) s3.DeleteBucketReplicationRequest

	PutBucketRequestPaymentRequest(input *s3.PutBucketRequestPaymentInput) s3.PutBucketRequestPaymentRequest
	GetBucketRequestPaymentRequest(input *s3.GetBucketRequestPaymentInput) s3.GetBucketRequestPaymentRequest

	PutBucketTaggingRequest(input *s3.PutBucketTaggingInput) s3.PutBucketTaggingRequest
	GetBucketTaggingRequest(input *s3.GetBucketTaggingInput) s3.GetBucketTaggingRequest
	DeleteBucketTaggingRequest(input *s3.DeleteBucketTaggingInput) s3.DeleteBucketTaggingRequest

	PutBucketAnalyticsConfigurationRequest(input *s3.PutBucketAnalyticsConfigurationInput) s3.PutBucketAnalyticsConfigurationRequest
	GetBucketAnalyticsConfigurationRequest(input *s3.GetBucketAnalyticsConfigurationInput) s3.GetBucketAnalyticsConfigurationRequest

	PutBucketLifecycleConfigurationRequest(input *s3.PutBucketLifecycleConfigurationInput) s3.PutBucketLifecycleConfigurationRequest
	GetBucketLifecycleConfigurationRequest(input *s3.GetBucketLifecycleConfigurationInput) s3.GetBucketLifecycleConfigurationRequest
	DeleteBucketLifecycleRequest(input *s3.DeleteBucketLifecycleInput) s3.DeleteBucketLifecycleRequest

	PutBucketNotificationConfigurationRequest(input *s3.PutBucketNotificationConfigurationInput) s3.PutBucketNotificationConfigurationRequest
	GetBucketNotificationConfigurationRequest(input *s3.GetBucketNotificationConfigurationInput) s3.GetBucketNotificationConfigurationRequest

	GetBucketAclRequest(*s3.GetBucketAclInput) s3.GetBucketAclRequest
	PutBucketAclRequest(*s3.PutBucketAclInput) s3.PutBucketAclRequest

	GetPublicAccessBlockRequest(input *s3.GetPublicAccessBlockInput) s3.GetPublicAccessBlockRequest
	PutPublicAccessBlockRequest(input *s3.PutPublicAccessBlockInput) s3.PutPublicAccessBlockRequest
	DeletePublicAccessBlockRequest(input *s3.DeletePublicAccessBlockInput) s3.DeletePublicAccessBlockRequest
}

// NewClient returns a new client using AWS credentials as JSON encoded data.
func NewClient(cfg aws.Config) BucketClient {
	return s3.New(cfg)
}

// IsNotFound helper function to test for NotFound error
func IsNotFound(err error) bool {
	if err == nil {
		return false
	}
	if bucketErr, ok := err.(awserr.Error); ok && bucketErr.Code() == BucketNotFoundErrCode {
		return true
	}
	return false
}

// IsAlreadyExists helper function to test for ErrCodeBucketAlreadyOwnedByYou error
func IsAlreadyExists(err error) bool {
	if err == nil {
		return false
	}
	if bucketErr, ok := err.(awserr.Error); ok && bucketErr.Code() == s3.ErrCodeBucketAlreadyOwnedByYou {
		return true
	}
	return false
}

// GenerateCreateBucketInput creates the input for CreateBucket S3 Client request
func GenerateCreateBucketInput(name string, s v1beta1.BucketParameters) *s3.CreateBucketInput {
	cbi := &s3.CreateBucketInput{
		ACL:                        s3.BucketCannedACL(aws.StringValue(s.ACL)),
		Bucket:                     aws.String(name),
		GrantFullControl:           s.GrantFullControl,
		GrantRead:                  s.GrantRead,
		GrantReadACP:               s.GrantReadACP,
		GrantWrite:                 s.GrantWrite,
		GrantWriteACP:              s.GrantWriteACP,
		ObjectLockEnabledForBucket: s.ObjectLockEnabledForBucket,
	}
	if s.LocationConstraint != "us-east-1" {
		cbi.CreateBucketConfiguration = &s3.CreateBucketConfiguration{LocationConstraint: s3.BucketLocationConstraint(s.LocationConstraint)}
	}
	return cbi
}

// GenerateBucketObservation generates the ARN string for the external status
func GenerateBucketObservation(name string) v1beta1.BucketExternalStatus {
	return v1beta1.BucketExternalStatus{
		ARN: fmt.Sprintf("arn:aws:s3:::%s", name),
	}
}

// CORSConfigurationNotFound is parses the aws Error and validates if the cors configuration does not exist
func CORSConfigurationNotFound(err error) bool {
	s3Err, ok := err.(awserr.Error)
	return ok && s3Err.Code() == CORSNotFoundErrCode
}

// ReplicationConfigurationNotFound is parses the aws Error and validates if the replication configuration does not exist
func ReplicationConfigurationNotFound(err error) bool {
	s3Err, ok := err.(awserr.Error)
	return ok && s3Err.Code() == ReplicationNotFoundErrCode
}

// PublicAccessBlockConfigurationNotFound is parses the aws Error and validates if the public access block does not exist
func PublicAccessBlockConfigurationNotFound(err error) bool {
	s3Err, ok := err.(awserr.Error)
	return ok && s3Err.Code() == PublicAccessBlockNotFoundErrCode
}

// LifecycleConfigurationNotFound is parses the aws Error and validates if the lifecycle configuration does not exist
func LifecycleConfigurationNotFound(err error) bool {
	s3Err, ok := err.(awserr.Error)
	return ok && s3Err.Code() == LifecycleNotFoundErrCode
}

// SSEConfigurationNotFound is parses the aws Error and validates if the SSE configuration does not exist
func SSEConfigurationNotFound(err error) bool {
	s3Err, ok := err.(awserr.Error)
	return ok && s3Err.Code() == SSENotFoundErrCode
}

// TaggingNotFound is parses the aws Error and validates if the tagging configuration does not exist
func TaggingNotFound(err error) bool {
	s3Err, ok := err.(awserr.Error)
	return ok && s3Err.Code() == TaggingNotFoundErrCode
}

// WebsiteConfigurationNotFound is parses the aws Error and validates if the website configuration does not exist
func WebsiteConfigurationNotFound(err error) bool {
	s3Err, ok := err.(awserr.Error)
	return ok && s3Err.Code() == WebsiteNotFoundErrCode
}

// MethodNotSupported is parses the aws Error and validates if the method is allowed for a request
func MethodNotSupported(err error) bool {
	s3Err, ok := err.(awserr.Error)
	return ok && s3Err.Code() == MethodNotAllowed
}

// ArgumentNotSupported is parses the aws Error and validates if parameters are now allowed for a request
func ArgumentNotSupported(err error) bool {
	s3Err, ok := err.(awserr.Error)
	return ok && s3Err.Code() == UnsupportedArgument
}

// UpdateBucketACL creates the ACLInput, sends the request to put an ACL based on the bucket
func UpdateBucketACL(ctx context.Context, client BucketClient, bucket *v1beta1.Bucket) error {
	config := &s3.PutBucketAclInput{
		ACL:              s3.BucketCannedACL(aws.StringValue(bucket.Spec.ForProvider.ACL)),
		Bucket:           aws.String(meta.GetExternalName(bucket)),
		GrantFullControl: bucket.Spec.ForProvider.GrantFullControl,
		GrantRead:        bucket.Spec.ForProvider.GrantRead,
		GrantReadACP:     bucket.Spec.ForProvider.GrantReadACP,
		GrantWrite:       bucket.Spec.ForProvider.GrantWrite,
		GrantWriteACP:    bucket.Spec.ForProvider.GrantWriteACP,
	}
	_, err := client.PutBucketAclRequest(config).Send(ctx)
	return err
}

// CopyTags converts a list of local v1beta.Tags to S3 Tags
func CopyTags(tags []v1beta1.Tag) []s3.Tag {
	out := make([]s3.Tag, 0)
	for _, one := range tags {
		out = append(out, s3.Tag{Key: aws.String(one.Key), Value: aws.String(one.Value)})
	}
	return out
}

// CopyAWSTags converts a list of external s3.Tags to local Tags
func CopyAWSTags(tags []s3.Tag) []v1beta1.Tag {
	out := make([]v1beta1.Tag, len(tags))
	for i, one := range tags {
		out[i] = v1beta1.Tag{Key: aws.StringValue(one.Key), Value: aws.StringValue(one.Value)}
	}
	return out
}

// SortS3TagSet stable sorts an external s3 tag list by the key and value.
func SortS3TagSet(tags []s3.Tag) []s3.Tag {
	outTags := make([]s3.Tag, len(tags))
	copy(outTags, tags)
	sort.SliceStable(outTags, func(i, j int) bool {
		return aws.StringValue(outTags[i].Key) < aws.StringValue(outTags[j].Key)
	})
	return outTags
}
