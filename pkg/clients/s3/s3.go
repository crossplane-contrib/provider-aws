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
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/awserr"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"k8s.io/apimachinery/pkg/util/rand"

	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	storage "github.com/crossplane/crossplane/apis/storage/v1alpha1"

	"github.com/crossplane/provider-aws/apis/storage/v1alpha3"
	iamc "github.com/crossplane/provider-aws/pkg/clients/iam"
	"github.com/crossplane/provider-aws/pkg/clients/s3/operations"
)

const (
	bucketUser           = "crossplane-bucket-%s"
	bucketObjectARN      = "arn:aws:s3:::%s"
	maxIAMUsernameLength = 64
	// https://docs.aws.amazon.com/general/latest/gr/rande.html#s3_region
	regionWithNoConstraint = "us-east-1"
)

// Service defines S3 Client operations
type Service interface {
	CreateOrUpdateBucket(bucket *v1alpha3.S3Bucket) error
	GetBucketInfo(username string, bucket *v1alpha3.S3Bucket) (*Bucket, error)
	CreateUser(username string, bucket *v1alpha3.S3Bucket) (*iam.AccessKey, string, error)
	UpdateBucketACL(bucket *v1alpha3.S3Bucket) error
	UpdateVersioning(bucket *v1alpha3.S3Bucket) error
	UpdatePolicyDocument(username string, bucket *v1alpha3.S3Bucket) (string, error)
	DeleteBucket(bucket *v1alpha3.S3Bucket) error
}

// Client implements S3 Client
type Client struct {
	s3        operations.Operations
	iamClient iamc.Client
}

// NewClient creates new S3 Client with provided AWS Configurations/Credentials
func NewClient(config *aws.Config) Service {
	ops := operations.NewS3Operations(s3.New(*config))
	return &Client{s3: ops, iamClient: iamc.NewClient(config)}
}

// CreateOrUpdateBucket creates or updates the supplied S3 bucket with provided
// specification, and returns access keys with permissions of localPermission
func (c *Client) CreateOrUpdateBucket(bucket *v1alpha3.S3Bucket) error {
	input := CreateBucketInput(bucket)
	_, err := c.s3.CreateBucketRequest(input).Send(context.TODO())
	if err != nil {
		if isErrorAlreadyExists(err) {
			return c.UpdateBucketACL(bucket)
		}
	}
	return err
}

// Bucket represents crossplane metadata about the bucket
type Bucket struct {
	Versioning        bool
	UserPolicyVersion string
}

// GetBucketInfo returns the status of key bucket settings including user's policy version for permission status
func (c *Client) GetBucketInfo(username string, bucket *v1alpha3.S3Bucket) (*Bucket, error) {
	b := Bucket{}
	bucketVersioning, err := c.s3.GetBucketVersioningRequest(&s3.GetBucketVersioningInput{Bucket: aws.String(meta.GetExternalName(bucket))}).Send(context.TODO())
	if err != nil {
		return nil, err
	}
	b.Versioning = bucketVersioning.Status == s3.BucketVersioningStatusEnabled
	policyVersion, err := c.iamClient.GetPolicyVersion(username)
	if err != nil {
		return nil, err
	}
	b.UserPolicyVersion = policyVersion

	return &b, err
}

// CreateUser - Create as user to access bucket per permissions in BucketSpec returing access key and policy version
func (c *Client) CreateUser(username string, bucket *v1alpha3.S3Bucket) (*iam.AccessKey, string, error) {
	policyDocument, err := newPolicyDocument(bucket)
	if err != nil {
		return nil, "", fmt.Errorf("could not update policy, %s", err.Error())
	}
	accessKeys, err := c.iamClient.CreateUser(username)
	if err != nil {
		return nil, "", fmt.Errorf("could not create user %s", err)
	}

	currentVersion, err := c.iamClient.CreatePolicyAndAttach(username, username, policyDocument)
	if err != nil {
		return nil, "", fmt.Errorf("could not create policy %s", err)
	}

	return accessKeys, currentVersion, nil
}

// UpdateBucketACL - Updated CannedACL on Bucket
func (c *Client) UpdateBucketACL(bucket *v1alpha3.S3Bucket) error {
	if bucket.Spec.CannedACL == nil {
		return nil
	}
	input := &s3.PutBucketAclInput{
		ACL:    *bucket.Spec.CannedACL,
		Bucket: aws.String(meta.GetExternalName(bucket)),
	}
	_, err := c.s3.PutBucketACLRequest(input).Send(context.TODO())
	return err
}

// UpdateVersioning configuration for Bucket
func (c *Client) UpdateVersioning(bucket *v1alpha3.S3Bucket) error {
	versioningStatus := s3.BucketVersioningStatusSuspended
	if bucket.Spec.Versioning {
		versioningStatus = s3.BucketVersioningStatusEnabled
	}
	input := &s3.PutBucketVersioningInput{Bucket: aws.String(meta.GetExternalName(bucket)), VersioningConfiguration: &s3.VersioningConfiguration{Status: versioningStatus}}
	_, err := c.s3.PutBucketVersioningRequest(input).Send(context.TODO())
	if err != nil {
		return err
	}
	return nil
}

// UpdatePolicyDocument based on localPermissions
func (c *Client) UpdatePolicyDocument(username string, bucket *v1alpha3.S3Bucket) (string, error) {
	policyDocument, err := newPolicyDocument(bucket)
	if err != nil {
		return "", fmt.Errorf("could not generate policy, %s", err.Error())
	}
	currentVersion, err := c.iamClient.UpdatePolicy(username, policyDocument)
	if err != nil {
		return "", fmt.Errorf("could not update policy, %s", err.Error())
	}
	return currentVersion, nil
}

// DeleteBucket deletes s3 bucket, and related IAM
func (c *Client) DeleteBucket(bucket *v1alpha3.S3Bucket) error {
	_, err := c.s3.DeleteBucketRequest(&s3.DeleteBucketInput{Bucket: aws.String(meta.GetExternalName(bucket))}).Send(context.TODO())
	if resource.Ignore(isErrorNotFound, err) != nil {
		return err
	}

	if bucket.Spec.IAMUsername != "" {
		err := c.iamClient.DeletePolicyAndDetach(bucket.Spec.IAMUsername, bucket.Spec.IAMUsername)
		if err != nil {
			return err
		}
		return c.iamClient.DeleteUser(bucket.Spec.IAMUsername)
	}

	return nil
}

// isErrorAlreadyExists helper function to test for ErrCodeBucketAlreadyOwnedByYou error
func isErrorAlreadyExists(err error) bool {
	if err == nil {
		return false
	}
	if bucketErr, ok := err.(awserr.Error); ok && bucketErr.Code() == s3.ErrCodeBucketAlreadyOwnedByYou {
		return true
	}
	return false
}

// isErrorNotFound helper function to test for ErrCodeNoSuchEntityException error
func isErrorNotFound(err error) bool {
	if err == nil {
		return false
	}
	if bucketErr, ok := err.(awserr.Error); ok && bucketErr.Code() == s3.ErrCodeNoSuchBucket {
		return true
	}
	return false
}

// CreateBucketInput returns a CreateBucketInput from the supplied S3Bucket.
func CreateBucketInput(bucket *v1alpha3.S3Bucket) *s3.CreateBucketInput {
	bucketInput := &s3.CreateBucketInput{
		Bucket: aws.String(meta.GetExternalName(bucket)),
	}

	if bucket.Spec.Region != regionWithNoConstraint {
		bucketInput.CreateBucketConfiguration = &s3.CreateBucketConfiguration{LocationConstraint: s3.BucketLocationConstraint(bucket.Spec.Region)}
	}

	if bucket.Spec.CannedACL != nil {
		bucketInput.ACL = *bucket.Spec.CannedACL
	}
	return bucketInput
}

// GenerateBucketUsername generates a username that is within AWS size
// specifications, and adds a random suffix.
func GenerateBucketUsername(bucket *v1alpha3.S3Bucket) string {
	name := fmt.Sprintf(bucketUser, meta.GetExternalName(bucket))
	max := maxIAMUsernameLength - 6
	if len(name) <= max {
		return name
	}
	return fmt.Sprintf("%s-%s", name[:max], rand.String(5))
}

func newPolicyDocument(bucket *v1alpha3.S3Bucket) (string, error) {
	bucketARN := fmt.Sprintf(bucketObjectARN, meta.GetExternalName(bucket))
	read := iamc.StatementEntry{
		Sid:    "crossplaneRead",
		Effect: "Allow",
		Action: []string{
			"s3:Get*",
			"s3:List*",
		},
		Resource: []string{bucketARN, bucketARN + "/*"},
	}

	write := iamc.StatementEntry{
		Sid:    "crossplaneWrite",
		Effect: "Allow",
		Action: []string{
			"s3:DeleteObject",
			"s3:Put*",
		},
		Resource: []string{bucketARN + "/*"},
	}

	policy := iamc.PolicyDocument{
		Version:   "2012-10-17",
		Statement: []iamc.StatementEntry{},
	}

	if bucket.Spec.LocalPermission != nil {
		switch *bucket.Spec.LocalPermission {
		case storage.ReadOnlyPermission:
			policy.Statement = append(policy.Statement, read)
		case storage.WriteOnlyPermission:
			policy.Statement = append(policy.Statement, write)
		case storage.ReadWritePermission:
			policy.Statement = append(policy.Statement, read, write)
		default:
			return "", fmt.Errorf("unknown permission, %s", *bucket.Spec.LocalPermission)
		}
	}

	b, err := json.Marshal(&policy)
	if err != nil {
		return "", fmt.Errorf("error marshaling policy, %s", err.Error())
	}

	return string(b), nil
}
