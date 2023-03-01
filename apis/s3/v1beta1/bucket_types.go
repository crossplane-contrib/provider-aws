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
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/crossplane-contrib/provider-aws/apis/s3/common"
)

const (
	// ResourceCredentialsSecretRegionKey is the key for region that the S3 bucket is located
	ResourceCredentialsSecretRegionKey = "region"
)

// BucketParameters are parameters for configuring the calls made to AWS Bucket API.
type BucketParameters struct {
	// The canned ACL to apply to the bucket.
	// +kubebuilder:validation:Enum=private;public-read;public-read-write;authenticated-read;aws-exec-read;bucket-owner-read;bucket-owner-full-control;log-delivery-write
	// +optional
	ACL *string `json:"acl,omitempty"`

	// LocationConstraint specifies the Region where the bucket will be created.
	// It is a required field.
	// Due to AWS API limitations lacking on a proper response, when this field is set to a wrong value,
	// or to non-existent region on bucket creation, it's impossible forwarding a meaningful status message to the user
	// about the problem, producing connection errors instead.
	LocationConstraint string `json:"locationConstraint"`

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

	// The container element for object ownership for a bucket's ownership controls.
	// BucketOwnerPreferred - Objects uploaded to the bucket change ownership to the
	// bucket owner if the objects are uploaded with the bucket-owner-full-control
	// canned ACL. ObjectWriter - The uploading account will own the object if the
	// object is uploaded with the bucket-owner-full-control canned ACL.
	// BucketOwnerEnforced - Access control lists (ACLs) are disabled and no longer
	// affect permissions. The bucket owner automatically owns and has full control
	// over every object in the bucket. The bucket only accepts PUT requests that don't
	// specify an ACL or bucket owner full control ACLs, such as the
	// bucket-owner-full-control canned ACL or an equivalent form of this ACL expressed
	// in the XML format.
	// +kubebuilder:validation:Enum=BucketOwnerPreferred;ObjectWriter;BucketOwnerEnforced
	// +optional
	ObjectOwnership *string `json:"objectOwnership,omitempty"`

	// Specifies default encryption for a bucket using server-side encryption with
	// Amazon S3-managed keys (SSE-S3) or customer master keys stored in AWS KMS
	// (SSE-KMS). For information about the Amazon S3 default encryption feature,
	// see Amazon S3 Default Bucket Encryption (https://docs.aws.amazon.com/AmazonS3/latest/dev/bucket-encryption.html)
	// in the Amazon Simple Storage Service Developer Guide.
	// +optional
	ServerSideEncryptionConfiguration *ServerSideEncryptionConfiguration `json:"serverSideEncryptionConfiguration,omitempty"`

	// VersioningConfiguration describes the versioning state of an Amazon S3 bucket.
	// See the AWS API reference guide for Amazon Simple Storage Service's API operation PutBucketVersioning for usage
	// and error information. See also, https://docs.aws.amazon.com/goto/WebAPI/s3-2006-03-01/PutBucketVersioning
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
	// See the AWS API reference guide for Amazon Simple Storage Service's API operation PutBucketWebsite for usage
	// and error information. See also, https://docs.aws.amazon.com/goto/WebAPI/s3-2006-03-01/PutBucketWebsite
	// +optional
	WebsiteConfiguration *WebsiteConfiguration `json:"websiteConfiguration,omitempty"`

	// Specifies logging parameters for an Amazon S3 bucket. Set the logging parameters for a bucket and
	// to specify permissions for who can view and modify the logging parameters. See the AWS API
	// reference guide for Amazon Simple Storage Service's API operation PutBucketLogging for usage
	// and error information. See also, https://docs.aws.amazon.com/goto/WebAPI/s3-2006-03-01/PutBucketLogging
	// +optional
	LoggingConfiguration *LoggingConfiguration `json:"loggingConfiguration,omitempty"`

	// Specifies payer parameters for an Amazon S3 bucket.
	// For more information, see Request Pays buckets
	// (https://docs.aws.amazon.com/AmazonS3/latest/dev/RequesterPaysBuckets.html) in the Amazon
	// Simple Storage Service Developer Guide.
	// +optional
	PayerConfiguration *PaymentConfiguration `json:"paymentConfiguration,omitempty"`

	// Sets the tags for a bucket.
	// Use tags to organize your AWS bill to reflect your own cost structure.
	// For more information, see Billing and usage reporting for S3 buckets.
	// (https://docs.aws.amazon.com/AmazonS3/latest/dev/BucketBilling.html) in the Amazon
	// Simple Storage Service Developer Guide.
	// +optional
	BucketTagging *Tagging `json:"tagging,omitempty"`

	// Creates a replication configuration or replaces an existing one.
	// For more information, see Replication (https://docs.aws.amazon.com/AmazonS3/latest/dev/replication.html)
	// in the Amazon S3 Developer Guide.
	// +optional
	ReplicationConfiguration *ReplicationConfiguration `json:"replicationConfiguration,omitempty"`

	// Creates a new lifecycle configuration for the bucket or replaces an existing
	// lifecycle configuration. For information about lifecycle configuration, see
	// Managing Access Permissions to Your Amazon S3 Resources
	// (https://docs.aws.amazon.com/AmazonS3/latest/dev/s3-access-control.html).
	// +optional
	LifecycleConfiguration *BucketLifecycleConfiguration `json:"lifecycleConfiguration,omitempty"`

	// Enables notifications of specified events for a bucket.
	// For more information about event notifications, see Configuring Event Notifications
	// (https://docs.aws.amazon.com/AmazonS3/latest/dev/NotificationHowTo.html).
	// +optional
	NotificationConfiguration *NotificationConfiguration `json:"notificationConfiguration,omitempty"`

	// PublicAccessBlockConfiguration that you want to apply to this Amazon
	// S3 bucket.
	PublicAccessBlockConfiguration *PublicAccessBlockConfiguration `json:"publicAccessBlockConfiguration,omitempty"`

	// Policy is a well defined type which can be parsed into an JSON S3 Bucket
	// Policy.
	//
	// By default, to ensure compatibility with previous APIs, a bucket policy
	// is never deleted from a bucket if this field is set to null.
	// To change this behaviour, see `policyUpdatePolicy`.
	//
	// +optional
	Policy *common.BucketPolicyBody `json:"policy,omitempty"`

	// PolicyUpdatePolicy specifies the update behaviour of `policy`.
	PolicyUpdatePolicy *BucketPolicyUpdatePolicy `json:"policyUpdatePolicy,omitempty"`
}

// BucketPolicyUpdatePolicy specifies the update behaviour of a bucket policy.
type BucketPolicyUpdatePolicy struct {
	// DeletionPolicy specifies how and if the bucket policy should be deleted
	// from the bucket.
	//
	// * `Never`: The bucket policy is never deleted, but only overwritten. If a
	// change is detected.
	//
	// * `IfNull`: The bucket policy is deleted if the `policy` field is set to
	// null.
	//
	// The default is `Never`.
	//
	// +kubebuilder:default="Never"
	// +kubebuilder:Enum=Never;IfNull
	DeletionPolicy BucketPolicyDeletionPolicy `json:"deletionPolicy,omitempty"`
}

// BucketPolicyDeletionPolicy determines how a bucket policy should be deleted
// from a bucket.
type BucketPolicyDeletionPolicy string

const (
	// BucketPolicyDeletionPolicyNever specifies that a bucket policy is never
	// deleted but only overwritten.
	BucketPolicyDeletionPolicyNever BucketPolicyDeletionPolicy = "Never"

	// BucketPolicyDeletionPolicyIfNull specifies that a bucket policy is
	// deleted if the corresponding spec field is null.
	BucketPolicyDeletionPolicyIfNull BucketPolicyDeletionPolicy = "IfNull"
)

// BucketSpec represents the desired state of the Bucket.
type BucketSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       BucketParameters `json:"forProvider"`
}

// BucketExternalStatus keeps the state for the external resource
type BucketExternalStatus struct {
	// ARN is the Amazon Resource Name (ARN) specifying the S3 Bucket. For more information
	// about ARNs and how to use them, see S3 Resources (https://docs.aws.amazon.com/AmazonS3/latest/dev/s3-arn-format.html)
	// in the Amazon Simple Storage Service guide.
	ARN string `json:"arn"`
}

// BucketStatus represents the observed state of the Bucket.
type BucketStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          BucketExternalStatus `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// An Bucket is a managed resource that represents an AWS S3 Bucket.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
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
