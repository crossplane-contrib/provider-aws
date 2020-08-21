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

package v1beta1

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	awsclients "github.com/crossplane/provider-aws/pkg/clients"
)

// BucketParameters are parameters for configuring the calls made to AWS Bucket API.
type BucketParameters struct {

	// The canned ACL to apply to the bucket.
	// +kubebuilder:validation:Enum=private;public-read;public-read-write;authenticated-read
	// +optional
	ACL *string `json:"acl,omitempty"`

	// Specifies the Region where the bucket will be created. If you don't specify
	// a Region, the bucket is created in the US East (N. Virginia) Region (us-east-1).
	// +optional
	LocationConstraint *string `json:"locationConstraint,omitempty"`

	// Allows grantee the read, write, read ACP, and write ACP permissions on the
	// bucket.
	// +optional
	GrantFullControl *string `json:"grantFullControl,omitempty"`

	// Allows grantee to list the objects in the bucket.
	// +optional
	GrantRead *string `json:"grantRead,omitempty"`

	// Allows grantee to read the bucket ACL.
	// +optional
	GrantReadACP *string `json:"grantReadAcp,omitempty"`

	// Allows grantee to create, overwrite, and delete any object in the bucket.
	// +optional
	GrantWrite *string `json:"grantWrite,omitempty"`

	// Allows grantee to write the ACL for the applicable bucket.
	// +optional
	GrantWriteACP *string `json:"grantWriteAcp,omitempty"`

	// Specifies whether you want S3 Object Lock to be enabled for the new bucket.
	// +optional
	ObjectLockEnabledForBucket *bool `json:"objectLockEnabledForBucket,omitempty"`

	// Specifies default encryption for a bucket using server-side encryption with
	// Amazon S3-managed keys (SSE-S3) or customer master keys stored in AWS KMS
	// (SSE-KMS). For information about the Amazon S3 default encryption feature,
	// see Amazon S3 Default Bucket Encryption (https://docs.aws.amazon.com/AmazonS3/latest/dev/bucket-encryption.html)
	// in the Amazon Simple Storage Service Developer Guide.
	// +optional
	ServerSideEncryptionConfiguration *ServerSideEncryptionConfiguration `json:"serverSideEncryptionConfiguration,omitempty"`

	// VersioningConfiguration describes the versioning state of an Amazon S3 bucket.
	// +optional
	VersioningConfiguration *VersioningConfiguration `json:"versioningConfiguration,omitempty"`

	// AccelerateConfiguration configures the transfer acceleration state for an
	// Amazon S3 bucket. For more information, see Amazon S3 Transfer Acceleration
	// (https://docs.aws.amazon.com/AmazonS3/latest/dev/transfer-acceleration.html)
	// in the Amazon Simple Storage Service Developer Guide.
	// +optional
	AccelerateConfiguration *AccelerateConfiguration `json:"accelerateConfiguration,omitempty"`

	// Describes the cross-origin access configuration for objects in an Amazon
	// S3 bucket. For more information, see Enabling Cross-Origin Resource Sharing
	// (https://docs.aws.amazon.com/AmazonS3/latest/dev/cors.html) in the Amazon
	// Simple Storage Service Developer Guide.
	// +optional
	CORSConfiguration *CORSConfiguration `json:"corsConfiguration,omitempty"`

	// Specifies website configuration parameters for an Amazon S3 bucket.
	WebsiteConfiguration *WebsiteConfiguration `json:"websiteConfiguration,omitempty"`

	// Specifies logging parameters for an Amazon S3 bucket.
	LoggingConfiguration *LoggingConfiguration `json:"loggingConfiguration,omitempty"`

	// Specifies payer parameters for an Amazon S3 bucket.
	// For more information, see Request Pays buckets
	// (https://docs.aws.amazon.com/AmazonS3/latest/dev/RequesterPaysBuckets.html) in the Amazon
	// Simple Storage Service Developer Guide.
	PayerConfiguration *PaymentConfiguration `json:"paymentConfiguration,omitempty"`

	// Sets the tags for a bucket.
	// Use tags to organize your AWS bill to reflect your own cost structure.
	// For more information, see Billing and usage reporting for S3 buckets.
	// (https://docs.aws.amazon.com/AmazonS3/latest/dev/BucketBilling.html) in the Amazon
	// Simple Storage Service Developer Guide.
	BucketTagging *Tagging `json:"tagging,omitempty"`

	// Creates a replication configuration or replaces an existing one.
	// For more information, see Replication (https://docs.aws.amazon.com/AmazonS3/latest/dev/replication.html)
	// in the Amazon S3 Developer Guide.
	// ReplicationConfiguration *ReplicationConfiguration `json:"replicationConfiguration,omitempty"`
}

// BucketObservation is observation of Bucket properties.
type BucketObservation struct {
}

// BucketSpec represents the desired state of the Bucket.
type BucketSpec struct {
	runtimev1alpha1.ResourceSpec `json:",inline"`
	ForProvider                  BucketParameters `json:"forProvider"`
}

// BucketStatus represents the observed state of the Bucket.
type BucketStatus struct {
	runtimev1alpha1.ResourceStatus `json:",inline"`
	AtProvider                     BucketObservation `json:"atProvider"`
}

// +kubebuilder:object:root=true

// An Bucket is a managed resource that represents an AWS S3 Bucket.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,aws}
type Bucket struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BucketSpec   `json:"spec"`
	Status BucketStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// BucketList contains a list of Buckets
type BucketList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Bucket `json:"items"`
}

// GenerateCreateBucketInput creates the input for CreateBucket S3 Client request
func GenerateCreateBucketInput(name string, s BucketParameters) *awss3.CreateBucketInput {
	cbi := &awss3.CreateBucketInput{
		ACL:                        awss3.BucketCannedACL(aws.StringValue(s.ACL)),
		Bucket:                     aws.String(name),
		GrantFullControl:           s.GrantFullControl,
		GrantRead:                  s.GrantRead,
		GrantReadACP:               s.GrantReadACP,
		GrantWrite:                 s.GrantWrite,
		GrantWriteACP:              s.GrantWriteACP,
		ObjectLockEnabledForBucket: s.ObjectLockEnabledForBucket,
	}
	if awsclients.StringValue(s.LocationConstraint) != "" {
		cbi.CreateBucketConfiguration = &awss3.CreateBucketConfiguration{LocationConstraint: awss3.BucketLocationConstraint(awsclients.StringValue(s.LocationConstraint))}
	}
	return cbi
}
