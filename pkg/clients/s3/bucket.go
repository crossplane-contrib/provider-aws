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

package s3

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/awserr"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/crossplane/provider-aws/apis/s3/v1beta1"
	awsclients "github.com/crossplane/provider-aws/pkg/clients"
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
func NewClient(ctx context.Context, credentials []byte, region string, auth awsclients.AuthMethod) (BucketClient, error) {
	cfg, err := auth(ctx, credentials, awsclients.DefaultSection, region)
	if cfg == nil {
		return nil, err
	}
	return s3.New(*cfg), nil
}

// IsNotFound helper function to test for ErrCodeNoSuchEntityException error
func IsNotFound(err error) bool {
	if err == nil {
		return false
	}
	if bucketErr, ok := err.(awserr.Error); ok && bucketErr.Code() == "NotFound" {
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
	if awsclients.StringValue(s.LocationConstraint) != "us-east-1" {
		cbi.CreateBucketConfiguration = &s3.CreateBucketConfiguration{LocationConstraint: s3.BucketLocationConstraint(awsclients.StringValue(s.LocationConstraint))}
	}
	return cbi
}
