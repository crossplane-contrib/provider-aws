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

package v1beta1

import (
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Enum values for Queue attribute names
const (
	AttributeAll                                   string = "All"
	AttributePolicy                                string = "Policy"
	AttributeVisibilityTimeout                     string = "VisibilityTimeout"
	AttributeMaximumMessageSize                    string = "MaximumMessageSize"
	AttributeMessageRetentionPeriod                string = "MessageRetentionPeriod"
	AttributeApproximateNumberOfMessages           string = "ApproximateNumberOfMessages"
	AttributeApproximateNumberOfMessagesNotVisible string = "ApproximateNumberOfMessagesNotVisible"
	AttributeCreatedTimestamp                      string = "CreatedTimestamp"
	AttributeLastModifiedTimestamp                 string = "LastModifiedTimestamp"
	AttributeQueueArn                              string = "QueueArn"
	AttributeApproximateNumberOfMessagesDelayed    string = "ApproximateNumberOfMessagesDelayed"
	AttributeDelaySeconds                          string = "DelaySeconds"
	AttributeReceiveMessageWaitTimeSeconds         string = "ReceiveMessageWaitTimeSeconds"
	AttributeRedrivePolicy                         string = "RedrivePolicy"
	AttributeFifoQueue                             string = "FifoQueue"
	AttributeContentBasedDeduplication             string = "ContentBasedDeduplication"
	AttributeKmsMasterKeyID                        string = "KmsMasterKeyId"
	AttributeKmsDataKeyReusePeriodSeconds          string = "KmsDataKeyReusePeriodSeconds"
	AttributeSqsManagedSseEnabled                  string = "SqsManagedSseEnabled"
)

// RedrivePolicy includes the parameters for the dead-letter queue functionality of the source queue.
type RedrivePolicy struct {
	// The Amazon Resource Name (ARN) of the dead-letter queue to which Amazon
	// SQS moves messages after the value of maxReceiveCount is exceeded.
	// +crossplane:generate:reference:type=Queue
	// +crossplane:generate:reference:extractor=QueueARN()
	DeadLetterTargetARN *string `json:"deadLetterTargetArn,omitempty"`

	// DeadLetterTargetARNRef reference a Queue to retrieve its ARN.
	// +optional
	DeadLetterTargetARNRef *xpv1.Reference `json:"deadLetterTargetArnRef,omitempty"`

	// DeadLetterTargetARNSelector selects reference to a Queue to retrieve its ARN
	// +optional
	DeadLetterTargetARNSelector *xpv1.Selector `json:"deadLetterTargetArnSelector,omitempty"`

	// The number of times a message is delivered to the source queue before
	// being moved to the dead-letter queue.
	MaxReceiveCount int64 `json:"maxReceiveCount"`
}

// QueueParameters define the desired state of an AWS Queue
type QueueParameters struct {
	// Region is the region you'd like your Queue to be created in.
	Region string `json:"region"`

	// DelaySeconds - The length of time, in seconds, for which the delivery
	// of all messages in the queue is delayed. Valid values: An integer from
	// 0 to 900 (15 minutes). Default: 0.
	// +optional
	DelaySeconds *int64 `json:"delaySeconds,omitempty"`

	// MaximumMessageSize is the limit of how many bytes a message can contain
	// before Amazon SQS rejects it. Valid values: An integer from 1,024 bytes
	// (1 KiB) up to 262,144 bytes (256 KiB). Default: 262,144 (256 KiB).
	// +optional
	MaximumMessageSize *int64 `json:"maximumMessageSize,omitempty"`

	// MessageRetentionPeriod - The length of time, in seconds, for which Amazon
	// SQS retains a message. Valid values: An integer representing seconds,
	// from 60 (1 minute) to 1,209,600 (14 days). Default: 345,600 (4 days).
	// +optional
	MessageRetentionPeriod *int64 `json:"messageRetentionPeriod,omitempty"`

	// The queue's policy. A valid AWS policy. For more information
	// about policy structure, see Overview of AWS IAM Policies (https://docs.aws.amazon.com/IAM/latest/UserGuide/PoliciesOverview.html)
	// in the Amazon IAM User Guide.
	// +optional
	Policy *string `json:"policy,omitempty"`

	// ReceiveMessageWaitTimeSeconds - The length of time, in seconds, for
	// which a ReceiveMessage action waits for a message to arrive. Valid values:
	// an integer from 0 to 20 (seconds). Default: 0.
	// +optional
	ReceiveMessageWaitTimeSeconds *int64 `json:"receiveMessageWaitTimeSeconds,omitempty"`

	// RedrivePolicy includes the parameters for the dead-letter
	// queue functionality of the source queue. For more information about the
	// redrive policy and dead-letter queues, see Using Amazon SQS Dead-Letter
	// Queues (https://docs.aws.amazon.com/AWSSimpleQueueService/latest/SQSDeveloperGuide/sqs-dead-letter-queues.html)
	// in the Amazon Simple Queue Service Developer Guide
	// +optional
	RedrivePolicy *RedrivePolicy `json:"redrivePolicy,omitempty"`

	// VisibilityTimeout - The visibility timeout for the queue, in seconds.
	// Valid values: an integer from 0 to 43,200 (12 hours). Default: 30. For
	// more information about the visibility timeout, see Visibility Timeout
	// (https://docs.aws.amazon.com/AWSSimpleQueueService/latest/SQSDeveloperGuide/sqs-visibility-timeout.html)
	// in the Amazon Simple Queue Service Developer Guide.
	// +optional
	VisibilityTimeout *int64 `json:"visibilityTimeout,omitempty"`

	// KMSMasterKeyID - The ID of an AWS-managed customer master key (CMK)
	// for Amazon SQS or a custom CMK. For more information, see Key Terms (https://docs.aws.amazon.com/AWSSimpleQueueService/latest/SQSDeveloperGuide/sqs-server-side-encryption.html#sqs-sse-key-terms).
	// While the alias of the AWS-managed CMK for Amazon SQS is always alias/aws/sqs,
	// the alias of a custom CMK can, for example, be alias/MyAlias . For more
	// examples, see KeyId (https://docs.aws.amazon.com/kms/latest/APIReference/API_DescribeKey.html#API_DescribeKey_RequestParameters)
	// in the AWS Key Management Service API Reference.
	// Applies only to server-side-encryption (https://docs.aws.amazon.com/AWSSimpleQueueService/latest/SQSDeveloperGuide/sqs-server-side-encryption.html):
	// +optional
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-aws/apis/kms/v1alpha1.Key
	KMSMasterKeyID *string `json:"kmsMasterKeyId,omitempty"`

	KMSMasterKeyIDRef *xpv1.Reference `json:"kmsMasterKeyIdRef,omitempty"`

	KMSMasterKeyIDSelector *xpv1.Selector `json:"kmsMasterKeyIdSelector,omitempty"`

	// KMSDataKeyReusePeriodSeconds - The length of time, in seconds, for which
	// Amazon SQS can reuse a data key (https://docs.aws.amazon.com/kms/latest/developerguide/concepts.html#data-keys)
	// to encrypt or decrypt messages before calling AWS KMS again. An integer
	// representing seconds, between 60 seconds (1 minute) and 86,400 seconds
	// (24 hours). Default: 300 (5 minutes). A shorter time period provides better
	// security but results in more calls to KMS which might incur charges after
	// Free Tier. For more information, see How Does the Data Key Reuse Period
	// Work? (https://docs.aws.amazon.com/AWSSimpleQueueService/latest/SQSDeveloperGuide/sqs-server-side-encryption.html#sqs-how-does-the-data-key-reuse-period-work).
	// Applies only to server-side-encryption (https://docs.aws.amazon.com/AWSSimpleQueueService/latest/SQSDeveloperGuide/sqs-server-side-encryption.html):
	// +optional
	KMSDataKeyReusePeriodSeconds *int64 `json:"kmsDataKeyReusePeriodSeconds,omitempty"`

	// FIFOQueue - Designates a queue as FIFO. Valid values: true, false. If
	//	you don't specify the FifoQueue attribute, Amazon SQS creates a standard
	//	queue. You can provide this attribute only during queue creation. You
	//	can't change it for an existing queue. When you set this attribute, you
	//	must also provide the MessageGroupId for your messages explicitly. For
	//	more information, see FIFO Queue Logic (https://docs.aws.amazon.com/AWSSimpleQueueService/latest/SQSDeveloperGuide/FIFO-queues.html#FIFO-queues-understanding-logic)
	//	in the Amazon Simple Queue Service Developer Guide.
	// +immutable
	// +optional
	FIFOQueue *bool `json:"fifoQueue,omitempty"`

	// ContentBasedDeduplication - Enables content-based deduplication. Valid
	// values: true, false. For more information, see Exactly-Once Processing
	// (https://docs.aws.amazon.com/AWSSimpleQueueService/latest/SQSDeveloperGuide/FIFO-queues.html#FIFO-queues-exactly-once-processing)
	// in the Amazon Simple Queue Service Developer Guide. Every message must
	// have a unique MessageDeduplicationId, You may provide a MessageDeduplicationId
	// explicitly. If you aren't able to provide a MessageDeduplicationId and
	// you enable ContentBasedDeduplication for your queue, Amazon SQS uses a
	// SHA-256 hash to generate the MessageDeduplicationId using the body of
	// the message (but not the attributes of the message). If you don't provide
	// a MessageDeduplicationId and the queue doesn't have ContentBasedDeduplication
	// set, the action fails with an error. If the queue has ContentBasedDeduplication
	// set, your MessageDeduplicationId overrides the generated one. When ContentBasedDeduplication
	// is in effect, messages with identical content sent within the deduplication
	// interval are treated as duplicates and only one copy of the message is
	// delivered. If you send one message with ContentBasedDeduplication enabled
	// and then another message with a MessageDeduplicationId that is the same
	// as the one generated for the first MessageDeduplicationId, the two messages
	// are treated as duplicates and only one copy of the message is delivered.
	// +optional
	ContentBasedDeduplication *bool `json:"contentBasedDeduplication,omitempty"`

	// Boolean to enable server-side encryption (SSE) of
	// message content with SQS-owned encryption keys. See Encryption at rest.
	SqsManagedSseEnabled *bool `json:"sseEnabled,omitempty"`

	// Tags add cost allocation tags to the specified Amazon SQS queue.
	// +optional
	Tags map[string]string `json:"tags,omitempty"`
}

// QueueSpec defines the desired state of a Queue.
type QueueSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       QueueParameters `json:"forProvider"`
}

// QueueObservation is the representation of the current state that is observed
type QueueObservation struct {
	// The URL of the created Amazon SQS queue.
	URL string `json:"url,omitempty"`

	// The Amazon resource name (ARN) of the queue.
	ARN string `json:"arn,omitempty"`

	// ApproximateNumberOfMessages - The approximate number of messages
	// available for retrieval from the queue.
	ApproximateNumberOfMessages int64 `json:"approximateNumberOfMessages,omitempty"`

	// ApproximateNumberOfMessagesDelayed - The approximate number
	// of messages in the queue that are delayed and not available for reading
	// immediately. This can happen when the queue is configured as a delay queue
	// or when a message has been sent with a delay parameter.
	ApproximateNumberOfMessagesDelayed int64 `json:"approximateNumberOfMessagesDelayed,omitempty"`

	// ApproximateNumberOfMessagesNotVisible - The approximate number
	// of messages that are in flight. Messages are considered to be in flight
	// if they have been sent to a client but have not yet been deleted or have
	// not yet reached the end of their visibility window.
	ApproximateNumberOfMessagesNotVisible int64 `json:"approximateNumberOfMessagesNotVisible,omitempty"`

	// CreatedTimestamp is the time when the queue was created
	CreatedTimestamp *metav1.Time `json:"createdTimestamp,omitempty"`

	// LastModifiedTimestamp - Returns the time when the queue was last changed.
	LastModifiedTimestamp *metav1.Time `json:"lastModifiedTimestamp,omitempty"`
}

// QueueStatus represents the observed state of a Queue.
type QueueStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          QueueObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A Queue is a managed resource that represents a AWS Simple Queue
// +kubebuilder:printcolumn:name="ARN",type="string",JSONPath=".status.atProvider.arn"
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,aws}
type Queue struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   QueueSpec   `json:"spec"`
	Status QueueStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// QueueList contains a list of Queue
type QueueList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Queue `json:"items"`
}
