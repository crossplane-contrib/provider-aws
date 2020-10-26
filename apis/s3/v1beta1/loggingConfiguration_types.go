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

import runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"

// LoggingConfiguration describes where logs are stored and the prefix that Amazon S3 assigns to
// all log object keys for a bucket. For more information, see PUT Bucket logging
// (https://docs.aws.amazon.com/AmazonS3/latest/API/RESTBucketPUTlogging.html)
type LoggingConfiguration struct {
	// TargetBucket where logs will be stored, it can be the same bucket.
	// At least one of targetBucket, targetBucketRef or targetBucketSelector is
	// required.
	// +optional
	TargetBucket *string `json:"targetBucket,omitempty"`

	// TargetBucketRef references an S3Bucket to retrieve its name
	// +optional
	TargetBucketRef *runtimev1alpha1.Reference `json:"targetBucketRef,omitempty"`

	// TargetBucketSelector selects a reference to an S3Bucket to retrieve its name
	// +optional
	TargetBucketSelector *runtimev1alpha1.Selector `json:"targetBucketSelector,omitempty"`

	// A prefix for all log object keys.
	TargetPrefix string `json:"targetPrefix"`

	// Container for granting information.
	TargetGrants []TargetGrant `json:"targetGrants,omitempty"`
}

// TargetGrant is the container for granting information.
type TargetGrant struct {
	// Container for the person being granted permissions.
	Grantee TargetGrantee `json:"targetGrantee"`

	// Logging permissions assigned to the Grantee for the bucket.
	// Valid values are "FULL_CONTROL", "READ", "WRITE"
	// +kubebuilder:validation:Enum=FULL_CONTROL;READ;WRITE
	Permission string `json:"bucketLogsPermission"`
}

// TargetGrantee is the container for the person being granted permissions.
type TargetGrantee struct {
	// Screen name of the grantee.
	DisplayName *string `json:"displayName,omitempty"`

	// Email address of the grantee.
	// For a list of all the Amazon S3 supported Regions and endpoints, see Regions
	// and Endpoints (https://docs.aws.amazon.com/general/latest/gr/rande.html#s3_region)
	// in the AWS General Reference.
	EmailAddress *string `json:"emailAddress,omitempty"`

	// The canonical user ID of the grantee.
	ID *string `json:"ID,omitempty"`

	// Type of grantee
	// Type is a required field
	// +kubebuilder:validation:Enum=CanonicalUser;AmazonCustomerByEmail;Group
	Type string `json:"type"`

	// URI of the grantee group.
	URI *string `json:"URI,omitempty"`
}
