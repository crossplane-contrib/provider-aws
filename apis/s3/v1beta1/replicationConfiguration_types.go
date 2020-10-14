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
	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
)

// ReplicationConfiguration contains replication rules. You can add up to 1,000 rules. The maximum
// size of a replication configuration is 2 MB.
type ReplicationConfiguration struct {
	// The Amazon Resource Name (ARN) of the AWS Identity and Access Management
	// (IAM) role that Amazon S3 assumes when replicating objects. For more information,
	// see How to Set Up Replication (https://docs.aws.amazon.com/AmazonS3/latest/dev/replication-how-setup.html)
	// in the Amazon Simple Storage Service Developer Guide.
	//
	// At least one of role, roleRef or roleSelector fields is required.
	// +optional
	Role *string `json:"role,omitempty"`

	// RoleRef references an IAMRole to retrieve its Name
	// +optional
	RoleRef *runtimev1alpha1.Reference `json:"roleRef,omitempty"`

	// RoleSelector selects a reference to an IAMRole to retrieve its Name
	// +optional
	RoleSelector *runtimev1alpha1.Selector `json:"roleSelector,omitempty"`

	// A container for one or more replication rules. A replication configuration
	// must have at least one rule and can contain a maximum of 1,000 rules.
	//
	// Rules is a required field
	Rules []ReplicationRule `json:"rules"`
}

// ReplicationRule specifies which Amazon S3 objects to replicate and where to store the replicas.
type ReplicationRule struct {
	// Specifies whether Amazon S3 replicates the delete markers. If you specify
	// a Filter, you must specify this element. However, in the latest version of
	// replication configuration (when Filter is specified), Amazon S3 doesn't replicate
	// delete markers. Therefore, the DeleteMarkerReplication element can contain
	// only <Status>Disabled</Status>. For an example configuration, see Basic Rule
	// Configuration (https://docs.aws.amazon.com/AmazonS3/latest/dev/replication-add-config.html#replication-config-min-rule-config).
	//
	// If you don't specify the Filter element, Amazon S3 assumes that the replication
	// configuration is the earlier version, V1. In the earlier version, Amazon
	// S3 handled replication of delete markers differently. For more information,
	// see Backward Compatibility (https://docs.aws.amazon.com/AmazonS3/latest/dev/replication-add-config.html#replication-backward-compat-considerations).
	DeleteMarkerReplication *DeleteMarkerReplication `json:"deleteMarkerReplication,omitempty"`

	// A container for information about the replication destination and its configurations
	// including enabling the S3 Replication Time Control (S3 RTC).
	//
	// Destination is a required field
	Destination Destination `json:"destination"`

	// Optional configuration to replicate existing source bucket objects. For more
	// information, see Replicating Existing Objects (https://docs.aws.amazon.com/AmazonS3/latest/dev/replication-what-is-isnot-replicated.html#existing-object-replication)
	// in the Amazon S3 Developer Guide.
	ExistingObjectReplication *ExistingObjectReplication `json:"existingObjectReplication,omitempty"`

	// A filter that identifies the subset of objects to which the replication rule
	// applies. A Filter must specify exactly one Prefix, Tag, or an And child element.
	Filter *ReplicationRuleFilter `json:"filter,omitempty"`

	// A unique identifier for the rule. The maximum value is 255 characters.
	ID *string `json:"id,omitempty"`

	// The priority associated with the rule. If you specify multiple rules in a
	// replication configuration, Amazon S3 prioritizes the rules to prevent conflicts
	// when filtering. If two or more rules identify the same object based on a
	// specified filter, the rule with higher priority takes precedence. For example:
	//
	//    * Same object quality prefix-based filter criteria if prefixes you specified
	//    in multiple rules overlap
	//
	//    * Same object qualify tag-based filter criteria specified in multiple
	//    rules
	//
	// For more information, see Replication (https://docs.aws.amazon.com/AmazonS3/latest/dev/replication.html)
	// in the Amazon Simple Storage Service Developer Guide.
	Priority *int64 `json:"priority,omitempty"`

	// A container that describes additional filters for identifying the source
	// objects that you want to replicate. You can choose to enable or disable the
	// replication of these objects. Currently, Amazon S3 supports only the filter
	// that you can specify for objects created with server-side encryption using
	// a customer master key (CMK) stored in AWS Key Management Service (SSE-KMS).
	SourceSelectionCriteria *SourceSelectionCriteria `json:"sourceSelectionCriteria,omitempty"`

	// Specifies whether the rule is enabled.
	//
	// Status is a required field
	// Valid values are "Enabled" or "Disabled"
	// +kubebuilder:validation:Enum=Enabled;Disabled
	Status string `json:"status"`
}

// DeleteMarkerReplication specifies whether Amazon S3 replicates the delete markers.
// If you specify a Filter, you must specify this element. However, in the latest version of
// replication configuration (when Filter is specified), Amazon S3 doesn't replicate
// delete markers. Therefore, the DeleteMarkerReplication element can contain
// only <Status>Disabled</Status>. For an example configuration, see Basic Rule
// Configuration (https://docs.aws.amazon.com/AmazonS3/latest/dev/replication-add-config.html#replication-config-min-rule-config).
//
// If you don't specify the Filter element, Amazon S3 assumes that the replication
// configuration is the earlier version, V1. In the earlier version, Amazon
// S3 handled replication of delete markers differently. For more information,
// see Backward Compatibility (https://docs.aws.amazon.com/AmazonS3/latest/dev/replication-add-config.html#replication-backward-compat-considerations).
type DeleteMarkerReplication struct {
	// Indicates whether to replicate delete markers.
	// In the current implementation, Amazon S3 doesn't replicate the delete markers.
	// The status must be "Disabled".
	// +kubebuilder:validation:Enum=Disabled
	Status string `json:"Status"`
}

// Destination specifies information about where to publish analysis or configuration results
// for an Amazon S3 bucket and S3 Replication Time Control (S3 RTC).
type Destination struct {
	// Specify this only in a cross-account scenario (where source and destination
	// bucket owners are not the same), and you want to change replica ownership
	// to the AWS account that owns the destination bucket. If this is not specified
	// in the replication configuration, the replicas are owned by same AWS account
	// that owns the source object.
	// +optional
	AccessControlTranslation *AccessControlTranslation `json:"accessControlTranslation,omitempty"`

	// Destination bucket owner account ID. In a cross-account scenario, if you
	// direct Amazon S3 to change replica ownership to the AWS account that owns
	// the destination bucket by specifying the AccessControlTranslation property,
	// this is the account ID of the destination bucket owner. For more information,
	// see Replication Additional Configuration: Changing the Replica Owner (https://docs.aws.amazon.com/AmazonS3/latest/dev/replication-change-owner.html)
	// in the Amazon Simple Storage Service Developer Guide.
	Account *string `json:"account,omitempty"`

	// The Amazon Resource Name (ARN) of the bucket where you want Amazon S3 to
	// store the results.
	// At least one of bucket, bucketRef or bucketSelector is required.
	// +optional
	Bucket *string `json:"bucket,omitempty"`

	// BucketRef references a Bucket to retrieve its Name
	// +optional
	BucketRef *runtimev1alpha1.Reference `json:"bucketRef,omitempty"`

	// BucketSelector selects a reference to a Bucket to retrieve its Name
	// +optional
	BucketSelector *runtimev1alpha1.Selector `json:"bucketSelector,omitempty"`

	// A container that provides information about encryption. If SourceSelectionCriteria
	// is specified, you must specify this element.
	// +optional
	EncryptionConfiguration *EncryptionConfiguration `json:"encryptionConfiguration,omitempty"`

	// A container specifying replication metrics-related settings enabling metrics
	// and Amazon S3 events for S3 Replication Time Control (S3 RTC). Must be specified
	// together with a ReplicationTime block.
	Metrics *Metrics `json:"metrics,omitempty"`

	// A container specifying S3 Replication Time Control (S3 RTC), including whether
	// S3 RTC is enabled and the time when all objects and operations on objects
	// must be replicated. Must be specified together with a Metrics block.
	ReplicationTime *ReplicationTime `json:"replicationTime,omitempty"`

	// The storage class to use when replicating objects, such as S3 Standard or
	// reduced redundancy. By default, Amazon S3 uses the storage class of the source
	// object to create the object replica.
	// For valid values, see the StorageClass element of the PUT Bucket replication
	// (https://docs.aws.amazon.com/AmazonS3/latest/API/RESTBucketPUTreplication.html)
	// action in the Amazon Simple Storage Service API Reference.
	// +kubebuilder:validation:Enum=GLACIER;STANDARD_IA;ONEZONE_IA;INTELLIGENT_TIERING;DEEP_ARCHIVE
	// +optional
	StorageClass *string `json:"storageClass"`
}

// AccessControlTranslation contains information about access control for replicas.
type AccessControlTranslation struct {
	// Specifies the replica ownership. For default and valid values, see PUT bucket
	// replication (https://docs.aws.amazon.com/AmazonS3/latest/API/RESTBucketPUTreplication.html)
	// in the Amazon Simple Storage Service API Reference.
	// Owner is a required field
	Owner string `json:"ownerOverride"`
}

// EncryptionConfiguration specifies encryption-related information for
// an Amazon S3 bucket that is a destination for replicated objects.
type EncryptionConfiguration struct {
	// Specifies the ID (Key ARN or Alias ARN) of the customer managed customer
	// master key (CMK) stored in AWS Key Management Service (KMS) for the destination
	// bucket. Amazon S3 uses this key to encrypt replica objects. Amazon S3 only
	// supports symmetric customer managed CMKs. For more information, see Using
	// Symmetric and Asymmetric Keys (https://docs.aws.amazon.com/kms/latest/developerguide/symmetric-asymmetric.html)
	// in the AWS Key Management Service Developer Guide.
	ReplicaKmsKeyID string `json:"replicaKmsKeyId"`
}

// Metrics specifies replication metrics-related settings enabling metrics
// and Amazon S3 events for S3 Replication Time Control (S3 RTC). Must be specified
// together with a ReplicationTime block.
type Metrics struct {
	// A container specifying the time threshold for emitting the s3:Replication:OperationMissedThreshold
	// event.
	// EventThreshold is a required field
	EventThreshold ReplicationTimeValue `json:"eventThreshold"`

	// Specifies whether the replication metrics are enabled.
	//
	// Status is a required field, valid values are "Enabled" and "Disabled"
	// +kubebuilder:validation:Enum=Enabled;Disabled
	Status string `json:"status"`
}

// ReplicationTimeValue specifies the time value for S3 Replication Time Control (S3
// RTC) and replication metrics EventThreshold.
type ReplicationTimeValue struct {
	// Contains an integer specifying time in minutes.
	//
	// Valid values: 15 minutes.
	Minutes int64 `json:"minutes"`
}

// ReplicationTime specifies S3 Replication Time Control (S3 RTC) related information,
// including whether S3 RTC is enabled and the time when all objects and operations
// on objects must be replicated. Must be specified together with a Metrics
// block.
type ReplicationTime struct {
	// Specifies whether the replication time is enabled
	// Status is a required field
	// Valid values are "Enabled" and "Disabled"
	// +kubebuilder:validation:Enum=Enabled;Disabled
	Status string `json:"status"`

	// A container specifying the time by which replication should be complete for
	// all objects and operations on objects.
	// Time is a required field
	Time ReplicationTimeValue `json:"time"`
}

// ExistingObjectReplication optional configuration to replicate existing source bucket objects. For more
// information, see Replicating Existing Objects
// (https://docs.aws.amazon.com/AmazonS3/latest/dev/replication-what-is-isnot-replicated.html#existing-object-replication)
// in the Amazon S3 Developer Guide.
type ExistingObjectReplication struct {
	// Status is a required field
	// Valid values are "Enabled" and "Disabled"
	// +kubebuilder:validation:Enum=Enabled;Disabled
	Status string `json:"status"`
}

// ReplicationRuleFilter identifies the subset of objects to which the replication rule
// applies. A Filter must specify exactly one Prefix, Tag, or an And child element.
type ReplicationRuleFilter struct {
	// A container for specifying rule filters. The filters determine the subset
	// of objects to which the rule applies. This element is required only if you
	// specify more than one filter. For example:
	//
	//    * If you specify both a Prefix and a Tag filter, wrap these filters in
	//    an And tag.
	//
	//    * If you specify a filter based on multiple tags, wrap the Tag elements
	//    in an And tag.
	And *ReplicationRuleAndOperator `json:"and,omitempty"`

	// An object key name prefix that identifies the subset of objects to which
	// the rule applies.
	Prefix *string `json:"prefix,omitempty"`

	// A container for specifying a tag key and value.
	// The rule applies only to objects that have the tag in their tag set.
	Tag *Tag `json:"tag,omitempty"`
}

// ReplicationRuleAndOperator specifies rule filters. The filters determine the subset
// of objects to which the rule applies. This element is required only if you
// specify more than one filter.
//
// For example:
//
//    * If you specify both a Prefix and a Tag filter, wrap these filters in
//    an And tag.
//
//    * If you specify a filter based on multiple tags, wrap the Tag elements
//    in an And tag
type ReplicationRuleAndOperator struct {
	// An object key name prefix that identifies the subset of objects to which
	// the rule applies.
	Prefix *string `json:"prefix,omitempty"`

	// An array of tags containing key and value pairs.
	Tags []Tag `json:"tag,omitempty"`
}

// SourceSelectionCriteria describes additional filters for identifying the source
// objects that you want to replicate. You can choose to enable or disable the
// replication of these objects. Currently, Amazon S3 supports only the filter
// that you can specify for objects created with server-side encryption using
// a customer master key (CMK) stored in AWS Key Management Service (SSE-KMS).
type SourceSelectionCriteria struct {
	// A container for filter information for the selection of Amazon S3 objects
	// encrypted with AWS KMS. If you include SourceSelectionCriteria in the replication
	// configuration, this element is required.
	SseKmsEncryptedObjects SseKmsEncryptedObjects `json:"sseKmsEncryptedObjects"`
}

// SseKmsEncryptedObjects is the container for filter information
// for the selection of S3 objects encrypted with AWS KMS.
type SseKmsEncryptedObjects struct {
	// Specifies whether Amazon S3 replicates objects created with server-side encryption
	// using a customer master key (CMK) stored in AWS Key Management Service.
	//
	// Status is a required field
	// Valid values are "Enabled" or "Disabled"
	// +kubebuilder:validation:Enum=Enabled;Disabled
	Status string `json:"status"`
}
