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
	// BucketErrCode is the error code sent by AWS when a bucket does not exist
	BucketErrCode = "NotFound"
	// CORSErrCode is the error code sent by AWS when the CORS configuration does not exist
	CORSErrCode = "NoSuchCORSConfiguration"
	// ReplicationErrCode is the error code sent by AWS when the replication config does not exist
	ReplicationErrCode = "ReplicationConfigurationNotFoundError"
	// LifecycleErrCode is the error code sent by AWS when the lifecycle config does not exist
	LifecycleErrCode = "NoSuchLifecycleConfiguration"
	// SSEErrCode is the error code sent by AWS when the SSE config does not exist
	SSEErrCode = "ServerSideEncryptionConfigurationNotFoundError"
	// TaggingErrCode is the error code sent by AWS when the tagging does not exist
	TaggingErrCode = "NoSuchTagSet"
	// WebsiteErrCode is the error code sent by AWS when the website config does not exist
	WebsiteErrCode = "NoSuchWebsiteConfiguration"
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
	if bucketErr, ok := err.(awserr.Error); ok && bucketErr.Code() == BucketErrCode {
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
	if s3Err, ok := err.(awserr.Error); ok && s3Err.Code() == CORSErrCode {
		return true
	}
	return false
}

// ReplicationConfigurationNotFound is parses the aws Error and validates if the replication configuration does not exist
func ReplicationConfigurationNotFound(err error) bool {
	if s3Err, ok := err.(awserr.Error); ok && s3Err.Code() == ReplicationErrCode {
		return true
	}
	return false
}

// LifecycleConfigurationNotFound is parses the aws Error and validates if the lifecycle configuration does not exist
func LifecycleConfigurationNotFound(err error) bool {
	if s3Err, ok := err.(awserr.Error); ok && s3Err.Code() == LifecycleErrCode {
		return true
	}
	return false
}

// SSEConfigurationNotFound is parses the aws Error and validates if the SSE configuration does not exist
func SSEConfigurationNotFound(err error) bool {
	if s3Err, ok := err.(awserr.Error); ok && s3Err.Code() == SSEErrCode {
		return true
	}
	return false
}

// TaggingNotFound is parses the aws Error and validates if the tagging configuration does not exist
func TaggingNotFound(err error) bool {
	if s3Err, ok := err.(awserr.Error); ok && s3Err.Code() == TaggingErrCode {
		return true
	}
	return false
}

// WebsiteConfigurationNotFound is parses the aws Error and validates if the website configuration does not exist
func WebsiteConfigurationNotFound(err error) bool {
	if s3Err, ok := err.(awserr.Error); ok && s3Err.Code() == WebsiteErrCode {
		return true
	}
	return false
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

// CopyTag converts a local v1beta.Tag to an S3 Tag
func CopyTag(tag *v1beta1.Tag) *s3.Tag {
	if tag == nil {
		return nil
	}
	return &s3.Tag{
		Key:   aws.String(tag.Key),
		Value: aws.String(tag.Value),
	}
}

// CopyTags converts a list of local v1beta.Tags to S3 Tags
func CopyTags(tags []v1beta1.Tag) []s3.Tag {
	if tags == nil {
		return nil
	}
	out := make([]s3.Tag, len(tags))
	for i := range tags {
		out[i] = *CopyTag(&tags[i])
	}
	return out
}

// SortS3TagSet stable sorts an external s3 tag list by the key and value.
func SortS3TagSet(tags []s3.Tag) []s3.Tag {
	if len(tags) == 0 {
		return tags
	}
	sort.SliceStable(tags, func(i, j int) bool {
		if aws.StringValue(tags[i].Key) < aws.StringValue(tags[j].Key) {
			return true
		} else if aws.StringValue(tags[i].Key) == aws.StringValue(tags[j].Key) {
			return aws.StringValue(tags[i].Value) < aws.StringValue(tags[j].Value)
		}
		return false
	})
	return tags
}
