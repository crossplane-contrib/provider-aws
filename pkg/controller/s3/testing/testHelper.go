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

package testing

import (
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/meta"

	"github.com/crossplane/provider-aws/apis/s3/v1beta1"
)

var (
	// an arbitrary managed resource
	acl = "private"
	// Region is the test region of the bucket
	Region           = "us-east-1"
	grantFullControl = "fullGrant"
	grantRead        = "readGrant"
	grantReadACP     = "readACPGrant"
	grantWrite       = "writeGrant"
	grantWriteACP    = "writeACPGrant"
	objectLock       = true
	// BucketName is the name of the s3 bucket in testing
	BucketName = "test.bucket.name"
)

// BucketModifier is a function which modifies the Bucket for testing
type BucketModifier func(bucket *v1beta1.Bucket)

// WithArn sets the ARN for an S3 Bucket
func WithArn(arn string) BucketModifier {
	return func(bucket *v1beta1.Bucket) {
		bucket.Status.AtProvider.ARN = arn
	}
}

// WithConditions sets the Conditions for an S3 Bucket
func WithConditions(c ...xpv1.Condition) BucketModifier { //nolint
	return func(r *v1beta1.Bucket) { r.Status.ConditionedStatus.Conditions = c }
}

// WithAccelerationConfig sets the AccelerateConfiguration for an S3 Bucket
func WithAccelerationConfig(s *v1beta1.AccelerateConfiguration) BucketModifier { //nolint
	return func(r *v1beta1.Bucket) { r.Spec.ForProvider.AccelerateConfiguration = s }
}

// WithSSEConfig sets the ServerSideEncryptionConfiguration for an S3 Bucket
func WithSSEConfig(s *v1beta1.ServerSideEncryptionConfiguration) BucketModifier { //nolint
	return func(r *v1beta1.Bucket) { r.Spec.ForProvider.ServerSideEncryptionConfiguration = s }
}

// WithVersioningConfig sets the VersioningConfiguration for an S3 Bucket
func WithVersioningConfig(s *v1beta1.VersioningConfiguration) BucketModifier { //nolint
	return func(r *v1beta1.Bucket) { r.Spec.ForProvider.VersioningConfiguration = s }
}

// WithCORSConfig sets the CORSConfiguration for an S3 Bucket
func WithCORSConfig(s *v1beta1.CORSConfiguration) BucketModifier { //nolint
	return func(r *v1beta1.Bucket) { r.Spec.ForProvider.CORSConfiguration = s }
}

// WithWebConfig sets the WebsiteConfiguration for an S3 Bucket
func WithWebConfig(s *v1beta1.WebsiteConfiguration) BucketModifier { //nolint
	return func(r *v1beta1.Bucket) { r.Spec.ForProvider.WebsiteConfiguration = s }
}

// WithLoggingConfig sets the LoggingConfiguration for an S3 Bucket
func WithLoggingConfig(s *v1beta1.LoggingConfiguration) BucketModifier { //nolint
	return func(r *v1beta1.Bucket) { r.Spec.ForProvider.LoggingConfiguration = s }
}

// WithPayerConfig sets the PaymentConfiguration for an S3 Bucket
func WithPayerConfig(s *v1beta1.PaymentConfiguration) BucketModifier { //nolint
	return func(r *v1beta1.Bucket) { r.Spec.ForProvider.PayerConfiguration = s }
}

// WithTaggingConfig sets the Tagging for an S3 Bucket
func WithTaggingConfig(s *v1beta1.Tagging) BucketModifier { //nolint
	return func(r *v1beta1.Bucket) { r.Spec.ForProvider.BucketTagging = s }
}

// WithReplConfig sets the ReplicationConfiguration for an S3 Bucket
func WithReplConfig(s *v1beta1.ReplicationConfiguration) BucketModifier { //nolint
	return func(r *v1beta1.Bucket) { r.Spec.ForProvider.ReplicationConfiguration = s }
}

// WithLifecycleConfig sets the BucketLifecycleConfiguration for an S3 Bucket
func WithLifecycleConfig(s *v1beta1.BucketLifecycleConfiguration) BucketModifier { //nolint
	return func(r *v1beta1.Bucket) { r.Spec.ForProvider.LifecycleConfiguration = s }
}

// WithNotificationConfig sets the NotificationConfiguration for an S3 Bucket
func WithNotificationConfig(s *v1beta1.NotificationConfiguration) BucketModifier { //nolint
	return func(r *v1beta1.Bucket) { r.Spec.ForProvider.NotificationConfiguration = s }
}

// Bucket creates a v1beta1 Bucket for use in testing
func Bucket(m ...BucketModifier) *v1beta1.Bucket {
	cr := &v1beta1.Bucket{
		Spec: v1beta1.BucketSpec{
			ForProvider: v1beta1.BucketParameters{
				ACL:                        &acl,
				LocationConstraint:         Region,
				GrantFullControl:           &grantFullControl,
				GrantRead:                  &grantRead,
				GrantReadACP:               &grantReadACP,
				GrantWrite:                 &grantWrite,
				GrantWriteACP:              &grantWriteACP,
				ObjectLockEnabledForBucket: &objectLock,
			},
		},
	}
	for _, f := range m {
		f(cr)
	}
	meta.SetExternalName(cr, BucketName)
	return cr
}
