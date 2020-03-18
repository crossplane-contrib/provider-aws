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
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/awserr"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/go-cmp/cmp"

	"github.com/crossplane/provider-aws/apis/storage/v1beta1"
	awsclients "github.com/crossplane/provider-aws/pkg/clients"
)

const (
	// https://docs.aws.amazon.com/general/latest/gr/rande.html#s3_region
	regionWithNoConstraint = "us-east-1"
)

// Client implements S3 Client
type Client interface {
	HeadBucketRequest(*s3.HeadBucketInput) s3.HeadBucketRequest
	CreateBucketRequest(*s3.CreateBucketInput) s3.CreateBucketRequest
	GetBucketPolicyRequest(*s3.GetBucketPolicyInput) s3.GetBucketPolicyRequest
	PutBucketPolicyRequest(*s3.PutBucketPolicyInput) s3.PutBucketPolicyRequest
	DeleteBucketRequest(*s3.DeleteBucketInput) s3.DeleteBucketRequest
}

// NewClient creates new S3 Client with provided AWS Configurations/Credentials
func NewClient(ctx context.Context, credentials []byte, region string, auth awsclients.AuthMethod) (Client, error) {
	cfg, err := auth(ctx, credentials, awsclients.DefaultSection, region)
	if cfg == nil {
		return nil, err
	}
	return s3.New(*cfg), err
}

// CreatePatch creates a *v1beta1.RDSInstanceParameters that has only the changed
// values between the target *v1beta1.RDSInstanceParameters and the current
// *rds.DBInstance
func CreatePatch(p v1beta1.S3BucketParameters, policy *string) (*v1beta1.S3BucketParameters, error) {
	patch := &v1beta1.S3BucketParameters{}

	if p.Policy != nil && *p.Policy != *policy {
		patch.Policy = p.Policy
	}

	return patch, nil
}

// IsUpToDate checks if the current spec is in sync with the observed bucket properties.
func IsUpToDate(p v1beta1.S3BucketParameters, policy string) (bool, error) {
	patch, err := CreatePatch(p, &policy)
	if err != nil {
		return false, err
	}
	return cmp.Equal(&v1beta1.S3BucketParameters{}, patch), nil
}

// IsErrorNotFound helper function to test for ErrCodeNoSuchEntityException error
func IsErrorNotFound(err error) bool {
	if bucketErr, ok := err.(awserr.Error); ok && bucketErr.Code() == s3.ErrCodeNoSuchBucket {
		return true
	}

	var awsErr awserr.RequestFailure
	if errors.As(err, &awsErr) {
		return awsErr.StatusCode() == 404 || awsErr.StatusCode() == 301
	}

	return false
}

// GenerateCreateBucketInput returns a CreateBucketInput from the supplied S3Bucket.
func GenerateCreateBucketInput(name string, p *v1beta1.S3BucketParameters) *s3.CreateBucketInput {
	bucketInput := &s3.CreateBucketInput{
		Bucket: aws.String(name),
	}

	if p.Region != regionWithNoConstraint {
		bucketInput.CreateBucketConfiguration = &s3.CreateBucketConfiguration{LocationConstraint: s3.BucketLocationConstraint(p.Region)}
	}

	if p.CannedACL != nil {
		bucketInput.ACL = *p.CannedACL
	}
	return bucketInput
}

// GenerateObservation is used to produce v1beta1.S3BucketObservation from
// s3.Bucket.
func GenerateObservation(policy string) v1beta1.S3BucketObservation { // nolint:gocyclo
	return v1beta1.S3BucketObservation{
		Policy: policy,
	}
}
