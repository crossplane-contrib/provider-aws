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

// NotificationConfiguration specifies the notification configuration of the bucket.
// If this element is empty, notifications are turned off for the bucket.
type NotificationConfiguration struct {
	// Describes the AWS Lambda functions to invoke and the events for which to
	// invoke them.
	// +optional
	LambdaFunctionConfigurations []LambdaFunctionConfiguration `json:"lambdaFunctionConfigurations,omitempty"`

	// The Amazon Simple Queue Service queues to publish messages to and the events
	// for which to publish messages.
	// +optional
	QueueConfigurations []QueueConfiguration `json:"queueConfigurations,omitempty"`

	// The topic to which notifications are sent and the events for which notifications
	// are generated.
	// +optional
	TopicConfigurations []TopicConfiguration `json:"topicConfigurations,omitempty"`
}

// PublicAccessBlockConfiguration that you want to apply to this Amazon
// S3 bucket. You can enable the configuration options in any combination. For
// more information about when Amazon S3 considers a bucket or object public,
// see The Meaning of "Public" (https://docs.aws.amazon.com/AmazonS3/latest/dev/access-control-block-public-access.html#access-control-block-public-access-policy-status)
// in the Amazon Simple Storage Service Developer Guide.
type PublicAccessBlockConfiguration struct {
	// Specifies whether Amazon S3 should block public access control lists (ACLs)
	// for this bucket and objects in this bucket. Setting this element to TRUE
	// causes the following behavior:
	//
	//    * PUT Bucket acl and PUT Object acl calls fail if the specified ACL is
	//    public.
	//
	//    * PUT Object calls fail if the request includes a public ACL.
	//
	//    * PUT Bucket calls fail if the request includes a public ACL.
	//
	// Enabling this setting doesn't affect existing policies or ACLs.
	BlockPublicAcls *bool `json:"blockPublicAcls,omitempty"`

	// Specifies whether Amazon S3 should block public bucket policies for this
	// bucket. Setting this element to TRUE causes Amazon S3 to reject calls to
	// PUT Bucket policy if the specified bucket policy allows public access.
	//
	// Enabling this setting doesn't affect existing bucket policies.
	BlockPublicPolicy *bool `json:"blockPublicPolicy,omitempty"`

	// Specifies whether Amazon S3 should ignore public ACLs for this bucket and
	// objects in this bucket. Setting this element to TRUE causes Amazon S3 to
	// ignore all public ACLs on this bucket and objects in this bucket.
	//
	// Enabling this setting doesn't affect the persistence of any existing ACLs
	// and doesn't prevent new public ACLs from being set.
	IgnorePublicAcls *bool `json:"ignorePublicAcls,omitempty"`

	// Specifies whether Amazon S3 should restrict public bucket policies for this
	// bucket. Setting this element to TRUE restricts access to this bucket to only
	// AWS services and authorized users within this account if the bucket has a
	// public policy.
	//
	// Enabling this setting doesn't affect previously stored bucket policies, except
	// that public and cross-account access within any public bucket policy, including
	// non-public delegation to specific accounts, is blocked.
	RestrictPublicBuckets *bool `json:"restrictPublicBuckets,omitempty"`
}

// LambdaFunctionConfiguration contains the configuration for AWS Lambda notifications.
type LambdaFunctionConfiguration struct {
	// The Amazon S3 bucket event for which to invoke the AWS Lambda function. For
	// more information, see Supported Event Types (https://docs.aws.amazon.com/AmazonS3/latest/dev/NotificationHowTo.html)
	// in the Amazon Simple Storage Service Developer Guide.
	//
	// Events is a required field
	// A full list of valid events can be found in the Amazon S3 Developer guide
	// https://docs.aws.amazon.com/AmazonS3/latest/dev/NotificationHowTo.html#notification-how-to-event-types-and-destinations
	Events []string `json:"events"`

	// Specifies object key name filtering rules. For information about key name
	// filtering, see Configuring Event Notifications (https://docs.aws.amazon.com/AmazonS3/latest/dev/NotificationHowTo.html)
	// in the Amazon Simple Storage Service Developer Guide.
	Filter *NotificationConfigurationFilter `json:"filter,omitempty"`

	// An optional unique identifier for configurations in a notification configuration.
	// If you don't provide one, Amazon S3 will assign an ID.
	// +optional
	ID *string `json:"ID,omitempty"`

	// The Amazon Resource Name (ARN) of the AWS Lambda function that Amazon S3
	// invokes when the specified event type occurs.
	//
	// LambdaFunctionArn is a required field
	LambdaFunctionArn string `json:"lambdaFunctionArn"`
}

// QueueConfiguration specifies the configuration for publishing messages to an Amazon Simple Queue
// Service (Amazon SQS) queue when Amazon S3 detects specified events.
type QueueConfiguration struct {
	// A collection of bucket events for which to send notifications
	//
	// Events is a required field
	// A full list of valid events can be found in the Amazon S3 Developer guide
	// https://docs.aws.amazon.com/AmazonS3/latest/dev/NotificationHowTo.html#notification-how-to-event-types-and-destinations
	Events []string `json:"events"`

	// Specifies object key name filtering rules. For information about key name
	// filtering, see Configuring Event Notifications (https://docs.aws.amazon.com/AmazonS3/latest/dev/NotificationHowTo.html)
	// in the Amazon Simple Storage Service Developer Guide.
	Filter *NotificationConfigurationFilter `json:"filter,omitempty"`

	// An optional unique identifier for configurations in a notification configuration.
	// If you don't provide one, Amazon S3 will assign an ID.
	// +optional
	ID *string `json:"ID,omitempty"`

	// The Amazon Resource Name (ARN) of the Amazon SQS queue to which Amazon S3
	// publishes a message when it detects events of the specified type.
	//
	// QueueArn is a required field
	QueueArn string `json:"queueArn"`
}

// TopicConfiguration specifies the configuration for publication of messages
// to an Amazon Simple Notification Service (Amazon SNS) topic when Amazon S3
// detects specified events.
type TopicConfiguration struct {
	// The Amazon S3 bucket event about which to send notifications. For more information,
	// see Supported Event Types (https://docs.aws.amazon.com/AmazonS3/latest/dev/NotificationHowTo.html)
	// in the Amazon Simple Storage Service Developer Guide.
	//
	// Events is a required field
	// A full list of valid events can be found in the Amazon S3 Developer guide
	// https://docs.aws.amazon.com/AmazonS3/latest/dev/NotificationHowTo.html#notification-how-to-event-types-and-destinations
	Events []string `json:"events"`

	// Specifies object key name filtering rules. For information about key name
	// filtering, see Configuring Event Notifications (https://docs.aws.amazon.com/AmazonS3/latest/dev/NotificationHowTo.html)
	// in the Amazon Simple Storage Service Developer Guide.
	Filter *NotificationConfigurationFilter `json:"filter,omitempty"`

	// An optional unique identifier for configurations in a notification configuration.
	// If you don't provide one, Amazon S3 will assign an ID.
	// +optional
	ID *string `json:"ID,omitempty"`

	// The Amazon Resource Name (ARN) of the Amazon SNS topic to which Amazon S3
	// publishes a message when it detects events of the specified type.
	// At least one of topicArn, topicArnRef or topicSelector is required.
	// +optional
	TopicArn *string `json:"topicArn,omitempty"`

	// TopicArnRef references an SNS Topic to retrieve its Arn
	// +optional
	TopicArnRef *xpv1.Reference `json:"topicRef,omitempty"`

	// TopicArnSelector selects a reference to an SNS Topic to retrieve its Arn
	// +optional
	TopicArnSelector *xpv1.Selector `json:"topicSelector,omitempty"`
}

// NotificationConfigurationFilter specifies object key name filtering rules. For information about key name
// filtering, see Configuring Event Notifications (https://docs.aws.amazon.com/AmazonS3/latest/dev/NotificationHowTo.html)
// in the Amazon Simple Storage Service Developer Guide.
type NotificationConfigurationFilter struct {
	// A container for object key name prefix and suffix filtering rules.
	Key *S3KeyFilter `json:"key,omitempty"`
}

// S3KeyFilter contains the object key name prefix and suffix filtering rules.
type S3KeyFilter struct {
	// A list of containers for the key-value pair that defines the criteria for
	// the filter rule.
	FilterRules []FilterRule `json:"filterRules"`
}

// FilterRule specifies the Amazon S3 object key name to filter on and whether to filter
// on the suffix or prefix of the key name.
type FilterRule struct {
	// The object key name prefix or suffix identifying one or more objects to which
	// the filtering rule applies. The maximum length is 1,024 characters. Overlapping
	// prefixes and suffixes are not supported. For more information, see Configuring
	// Event Notifications (https://docs.aws.amazon.com/AmazonS3/latest/dev/NotificationHowTo.html)
	// in the Amazon Simple Storage Service Developer Guide.
	// Valid values are "prefix" or "suffix"
	// +kubebuilder:validation:Enum=prefix;suffix
	Name string `json:"name"`

	// The value that the filter searches for in object key names.
	Value *string `json:"value,omitempty"`
}
