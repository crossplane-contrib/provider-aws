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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
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
	AttributeDeadLetterQueueARN                    string = "DeadLetterQueueARN"
	AttributeMaxReceiveCount                       string = "MaxReceiveCount"
	AttributeFifoQueue                             string = "FifoQueue"
	AttributeContentBasedDeduplication             string = "ContentBasedDeduplication"
	AttributeKmsMasterKeyID                        string = "KmsMasterKeyId"
	AttributeKmsDataKeyReusePeriodSeconds          string = "KmsDataKeyReusePeriodSeconds"
)

// Tag is a key value pairs attached to a Amazon SQS queue.
type Tag struct {

	// The key name that can be used to look up or retrieve the associated value.
	Key string `json:"key"`

	// The value associated with a key in a tag.
	// +optional
	Value string `json:"value,omitempty"`
}

// RedrivePolicy includes the parameters for the dead-letter queue functionality of the source queue.
type RedrivePolicy struct {
	// The Amazon Resource Name (ARN) of the dead-letter queue to which Amazon
	// SQS moves messages after the value of maxReceiveCount is exceeded.
	// +optional
	DeadLetterQueueARN *string `json:"deadLetterQueueARN,omitempty"`

	// The number of times a message is delivered to the source queue before
	// being moved to the dead-letter queue.
	// +optional
	MaxReceiveCount *int64 `json:"maxReceiveCount,omitempty"`
}

// QueueParameters define the desired state of an AWS Queue
type QueueParameters struct {
	// The length of time, in seconds, for which the delivery
	// of all messages in the queue is delayed.
	// +optional
	DelaySeconds *int64 `json:"delaySeconds,omitempty"`

	// Designates a queue as FIFO.
	// +immutable
	// +optional
	FIFOQueue *bool `json:"fifoQueue,omitempty"`

	// The limit of how many bytes a message can contain before Amazon SQS rejects it.
	// +optional
	MaximumMessageSize *int64 `json:"maximumMessageSize,omitempty"`

	// The length of time, in seconds, for which Amazon SQS retains a message.
	// +optional
	MessageRetentionPeriod *int64 `json:"messageRetentionPeriod,omitempty"`

	// The length of time, in seconds, for which a ReceiveMessage
	// action waits for a message to arrive.
	// +optional
	ReceiveMessageWaitTimeSeconds *int64 `json:"receiveMessageWaitTimeSeconds,omitempty"`

	// RedrivePolicy includes the parameters for the dead-letter queue
	// functionality of the source queue.
	// +optional
	RedrivePolicy *RedrivePolicy `json:"redrivePolicy,omitempty"`

	// The visibility timeout for the queue, in seconds.
	// +optional
	VisibilityTimeout *int64 `json:"visibilityTimeout,omitempty"`

	// The ID of an AWS-managed customer master key (CMK) for Amazon SQS or a custom CMK.
	// +optional
	KMSMasterKeyID *string `json:"kmsMasterKeyId,omitempty"`

	// The length of time, in seconds, for which
	// Amazon SQS can reuse a data key to encrypt or decrypt messages before calling AWS KMS again.
	// +optional
	KMSDataKeyReusePeriodSeconds *int64 `json:"kmsDataKeyReusePeriodSeconds,omitempty"`

	// Tags add cost allocation tags to the specified Amazon SQS queue.
	// +optional
	Tags []Tag `json:"tags,omitempty"`
}

// QueueSpec defines the desired state of a Queue.
type QueueSpec struct {
	runtimev1alpha1.ResourceSpec `json:",inline"`
	ForProvider                  QueueParameters `json:"forProvider"`
}

// QueueObservation is the representation of the current state that is observed
type QueueObservation struct {
	// The URL of the created Amazon SQS queue.
	URL string `json:"url,omitempty"`

	// The Amazon resource name (ARN) of the queue.
	ARN string `json:"arn,omitempty"`
}

// QueueStatus represents the observed state of a Queue.
type QueueStatus struct {
	runtimev1alpha1.ResourceStatus `json:",inline"`
	AtProvider                     QueueObservation `json:"atProvider"`
}

// +kubebuilder:object:root=true

// A Queue is a managed resource that represents a AWS Simple Queue
// +kubebuilder:printcolumn:name="QUEUENAME",type="string",JSONPath=".spec.forProvider.name"
// +kubebuilder:printcolumn:name="ARN",type="string",JSONPath=".status.atProvider.arn"
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster
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
