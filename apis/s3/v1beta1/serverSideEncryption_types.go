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

import xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"

// ServerSideEncryptionConfiguration specifies the default server-side-encryption configuration.
type ServerSideEncryptionConfiguration struct {
	// Container for information about a particular server-side encryption configuration
	// rule.
	Rules []ServerSideEncryptionRule `json:"rules"`
}

// ServerSideEncryptionRule Specifies the default server-side encryption configuration.
type ServerSideEncryptionRule struct {
	// Specifies the default server-side encryption to apply to new objects in the
	// bucket. If a PUT Object request doesn't specify any server-side encryption,
	// this default encryption will be applied.
	ApplyServerSideEncryptionByDefault ServerSideEncryptionByDefault `json:"applyServerSideEncryptionByDefault"`

	// Specifies whether Amazon S3 should use an S3 Bucket Key with server-side
	// encryption using KMS (SSE-KMS) for new objects in the bucket. Existing objects
	// are not affected. Setting the BucketKeyEnabled element to true causes Amazon S3
	// to use an S3 Bucket Key. By default, S3 Bucket Key is not enabled. For more
	// information, see Amazon S3 Bucket Keys
	// (https://docs.aws.amazon.com/AmazonS3/latest/dev/bucket-key.html) in the Amazon
	// S3 User Guide.
	BucketKeyEnabled bool `json:"bucketKeyEnabled,omitempty"`
}

// ServerSideEncryptionByDefault describes the default server-side encryption to
// apply to new objects in the bucket. If a PUT Object request doesn't specify
// any server-side encryption, this default encryption will be applied.
type ServerSideEncryptionByDefault struct {
	// AWS Key Management Service (KMS) customer master key ID to use for the default
	// encryption. This parameter is allowed if and only if SSEAlgorithm is set
	// to aws:kms.
	//
	// You can specify the key ID or the Amazon Resource Name (ARN) of the CMK.
	// However, if you are using encryption with cross-account operations, you must
	// use a fully qualified CMK ARN. For more information, see Using encryption
	// for cross-account operations (https://docs.aws.amazon.com/AmazonS3/latest/dev/bucket-encryption.html#bucket-encryption-update-bucket-policy).
	//
	// For example:
	//
	//    * Key ID: 1234abcd-12ab-34cd-56ef-1234567890ab
	//
	//    * Key ARN: arn:aws:kms:us-east-2:111122223333:key/1234abcd-12ab-34cd-56ef-1234567890ab
	//
	// Amazon S3 only supports symmetric CMKs and not asymmetric CMKs. For more
	// information, see Using Symmetric and Asymmetric Keys (https://docs.aws.amazon.com/kms/latest/developerguide/symmetric-asymmetric.html)
	// in the AWS Key Management Service Developer Guide.
	// +optional
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-aws/apis/kms/v1alpha1.Key
	KMSMasterKeyID *string `json:"kmsMasterKeyId,omitempty"`

	// KMSMasterKeyIDRef references an KMSKey to retrieve its ID
	// +optional
	KMSMasterKeyIDRef *xpv1.Reference `json:"kmsMasterKeyIdRef,omitempty"`

	// KMSMasterKeyIDSelector selects a reference to an KMSKey to retrieve its ID
	// +optional
	KMSMasterKeyIDSelector *xpv1.Selector `json:"kmsMasterKeyIdSelector,omitempty"`

	// NOTE(muvaf): aws:kms is not accepted by kubebuilder enum.

	// Server-side encryption algorithm to use for the default encryption.
	// Options are AES256 or aws:kms
	SSEAlgorithm string `json:"sseAlgorithm"`
}
