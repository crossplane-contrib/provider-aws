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

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// BucketLifecycleConfiguration specifies the lifecycle configuration for objects in an Amazon S3 bucket.
// For more information, see Object Lifecycle Management (https://docs.aws.amazon.com/AmazonS3/latest/dev/object-lifecycle-mgmt.html)
// in the Amazon Simple Storage Service Developer Guide.
type BucketLifecycleConfiguration struct {
	// A lifecycle rule for individual objects in an Amazon S3 bucket.
	//
	// Rules is a required field
	Rules []LifecycleRule `json:"rules"`
}

// LifecycleRule for individual objects in an Amazon S3 bucket.
type LifecycleRule struct {
	// Specifies the days since the initiation of an incomplete multipart upload
	// that Amazon S3 will wait before permanently removing all parts of the upload.
	// For more information, see Aborting Incomplete Multipart Uploads Using a Bucket
	// Lifecycle Policy (https://docs.aws.amazon.com/AmazonS3/latest/dev/mpuoverview.html#mpu-abort-incomplete-mpu-lifecycle-config)
	// in the Amazon Simple Storage Service Developer Guide.
	AbortIncompleteMultipartUpload *AbortIncompleteMultipartUpload `json:"abortIncompleteMultipartUpload,omitempty"`

	// Specifies the expiration for the lifecycle of the object in the form of date,
	// days and, whether the object has a delete marker.
	Expiration *LifecycleExpiration `json:"expiration,omitempty"`

	// The Filter is used to identify objects that a Lifecycle Rule applies to.
	// A Filter must have exactly one of Prefix, Tag, or And specified.
	Filter *LifecycleRuleFilter `json:"filter,omitempty"`

	// Unique identifier for the rule. The value cannot be longer than 255 characters.
	ID *string `json:"id,omitempty"`

	// Specifies when noncurrent object versions expire. Upon expiration, Amazon
	// S3 permanently deletes the noncurrent object versions. You set this lifecycle
	// configuration action on a bucket that has versioning enabled (or suspended)
	// to request that Amazon S3 delete noncurrent object versions at a specific
	// period in the object's lifetime.
	NoncurrentVersionExpiration *NoncurrentVersionExpiration `json:"noncurrentVersionExpiration,omitempty"`

	// Specifies the transition rule for the lifecycle rule that describes when
	// noncurrent objects transition to a specific storage class. If your bucket
	// is versioning-enabled (or versioning is suspended), you can set this action
	// to request that Amazon S3 transition noncurrent object versions to a specific
	// storage class at a set period in the object's lifetime.
	NoncurrentVersionTransitions []NoncurrentVersionTransition `json:"noncurrentVersionTransitions,omitempty"`

	// If 'Enabled', the rule is currently being applied. If 'Disabled', the rule
	// is not currently being applied.
	//
	// Status is a required field, valid values are Enabled or Disabled
	// +kubebuilder:validation:Enum=Enabled;Disabled
	Status string `json:"status"`

	// Specifies when an Amazon S3 object transitions to a specified storage class.
	Transitions []Transition `json:"transitions,omitempty"`
}

// AbortIncompleteMultipartUpload specifies the days since the initiation of an incomplete multipart upload
// that Amazon S3 will wait before permanently removing all parts of the upload.
// For more information, see Aborting Incomplete Multipart Uploads Using a Bucket
// Lifecycle Policy (https://docs.aws.amazon.com/AmazonS3/latest/dev/mpuoverview.html#mpu-abort-incomplete-mpu-lifecycle-config)
// in the Amazon Simple Storage Service Developer Guide.
type AbortIncompleteMultipartUpload struct {
	// Specifies the number of days after which Amazon S3 aborts an incomplete multipart
	// upload.
	DaysAfterInitiation int64 `json:"daysAfterInitiation"`
}

// LifecycleExpiration contains for the expiration for the lifecycle of the object.
type LifecycleExpiration struct {
	// Indicates at what date the object is to be moved or deleted.
	Date *metav1.Time `json:"date,omitempty"`

	// Indicates the lifetime, in days, of the objects that are subject to the rule.
	// The value must be a non-zero positive integer.
	Days *int64 `json:"days,omitempty"`

	// Indicates whether Amazon S3 will remove a delete marker with no noncurrent
	// versions. If set to true, the delete marker will be expired; if set to false
	// the policy takes no action. This cannot be specified with Days or Date in
	// a Lifecycle Expiration Policy.
	ExpiredObjectDeleteMarker *bool `json:"expiredObjectDeleteMarker,omitempty"`
}

// LifecycleRuleFilter is used to identify objects that a Lifecycle Rule applies to.
// A Filter must have exactly one of Prefix, Tag, or And specified.
type LifecycleRuleFilter struct {
	// This is used in a Lifecycle Rule Filter to apply a logical AND to two or
	// more predicates. The Lifecycle Rule will apply to any object matching all
	// of the predicates configured inside the And operator.
	And *LifecycleRuleAndOperator `json:"and,omitempty"`

	// Prefix identifying one or more objects to which the rule applies.
	Prefix *string `json:"prefix,omitempty"`

	// This tag must exist in the object's tag set in order for the rule to apply.
	Tag *Tag `json:"tag,omitempty"`
}

// LifecycleRuleAndOperator is used in a Lifecycle Rule Filter to apply a logical AND to two or
// more predicates. The Lifecycle Rule will apply to any object matching all
// of the predicates configured inside the And operator.
type LifecycleRuleAndOperator struct {
	// Prefix identifying one or more objects to which the rule applies.
	Prefix *string `json:"prefix,omitempty"`

	// All of these tags must exist in the object's tag set in order for the rule
	// to apply.
	Tags []Tag `json:"tags"`
}

// NoncurrentVersionExpiration specifies when noncurrent object versions expire. Upon expiration, Amazon
// S3 permanently deletes the noncurrent object versions. You set this lifecycle
// configuration action on a bucket that has versioning enabled (or suspended)
// to request that Amazon S3 delete noncurrent object versions at a specific
// period in the object's lifetime.
type NoncurrentVersionExpiration struct {
	// Specifies the number of days an object is noncurrent before Amazon S3 can
	// perform the associated action. For information about the noncurrent days
	// calculations, see How Amazon S3 Calculates When an Object Became Noncurrent
	// (https://docs.aws.amazon.com/AmazonS3/latest/dev/intro-lifecycle-rules.html#non-current-days-calculations)
	// in the Amazon Simple Storage Service Developer Guide.
	NoncurrentDays *int64 `json:"noncurrentDays,omitempty"`
}

// NoncurrentVersionTransition contains the transition rule that describes when noncurrent objects
// transition to the STANDARD_IA, ONEZONE_IA, INTELLIGENT_TIERING, GLACIER,
// or DEEP_ARCHIVE storage class. If your bucket is versioning-enabled (or versioning
// is suspended), you can set this action to request that Amazon S3 transition
// noncurrent object versions to the STANDARD_IA, ONEZONE_IA, INTELLIGENT_TIERING,
// GLACIER, or DEEP_ARCHIVE storage class at a specific period in the object's
// lifetime.
type NoncurrentVersionTransition struct {
	// Specifies the number of days an object is noncurrent before Amazon S3 can
	// perform the associated action. For information about the noncurrent days
	// calculations, see How Amazon S3 Calculates How Long an Object Has Been Noncurrent
	// (https://docs.aws.amazon.com/AmazonS3/latest/dev/intro-lifecycle-rules.html#non-current-days-calculations)
	// in the Amazon Simple Storage Service Developer Guide.
	NoncurrentDays *int64 `json:"noncurrentDays,omitempty"`

	// The class of storage used to store the object.
	// Valid values are: GLACIER, STANDARD_IA, ONEZONE_IA, INTELLIGENT_TIERING, DEEP_ARCHIVE
	// +kubebuilder:validation:Enum=GLACIER;STANDARD_IA;ONEZONE_IA;INTELLIGENT_TIERING;DEEP_ARCHIVE
	StorageClass string `json:"storageClass"`
}

// Transition specifies when an object transitions to a specified storage class. For more
// information about Amazon S3 lifecycle configuration rules, see Transitioning
// Objects Using Amazon S3 Lifecycle (https://docs.aws.amazon.com/AmazonS3/latest/dev/lifecycle-transition-general-considerations.html)
// in the Amazon Simple Storage Service Developer Guide.
type Transition struct {
	// Indicates when objects are transitioned to the specified storage class. The
	// date value must be in ISO 8601 format. The time is always midnight UTC.
	Date *metav1.Time `json:"date,omitempty"`

	// Indicates the number of days after creation when objects are transitioned
	// to the specified storage class. The value must be a positive integer.
	Days *int64 `json:"days,omitempty"`

	// The storage class to which you want the object to transition.
	// Valid values are: GLACIER, STANDARD_IA, ONEZONE_IA, INTELLIGENT_TIERING, DEEP_ARCHIVE
	// +kubebuilder:validation:Enum=GLACIER;STANDARD_IA;ONEZONE_IA;INTELLIGENT_TIERING;DEEP_ARCHIVE
	StorageClass string `json:"storageClass"`
}
