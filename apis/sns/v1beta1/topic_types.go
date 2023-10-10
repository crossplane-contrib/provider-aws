/*
Copyright 2022 The Crossplane Authors.

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

// Tag represent a user-provided metadata that can be associated with a
// SNS Topic. For more information about tagging,
// see Tagging SNS Topics (https://docs.aws.amazon.com/sns/latest/dg/sns-tags.html)
// in the SNS User Guide.
type Tag struct {

	// The key name that can be used to look up or retrieve the associated value.
	// For example, Department or Cost Center are common choices.
	Key string `json:"key"`

	// The value associated with this tag. For example, tags with a key name of
	// Department could have values such as Human Resources, Accounting, and Support.
	// Tags with a key name of Cost Center might have values that consist of the
	// number associated with the different cost centers in your company. Typically,
	// many resources have tags with the same key name but with different values.
	//
	// AWS always interprets the tag Value as a single string. If you need to store
	// an array, you can store comma-separated values in the string. However, you
	// must interpret the value in your code.
	// +optional
	Value *string `json:"value,omitempty"`
}

// TopicParameters define the desired state of a AWS SNS Topic
type TopicParameters struct {
	// Region is the region you'd like your Topic to be created in.
	Region string `json:"region"`

	// Name refers to the name of the AWS SNS Topic
	// +immutable
	Name string `json:"name"`

	// The display name to use for a topic with SNS subscriptions.
	// +optional
	DisplayName *string `json:"displayName,omitempty"`

	// Setting this enables server side encryption at-rest to your topic.
	// The ID of an AWS-managed customer master key (CMK) for Amazon SNS or a custom CMK
	//
	// For more examples, see KeyId (https://docs.aws.amazon.com/kms/latest/APIReference/API_DescribeKey.html#API_DescribeKey_RequestParameters)
	// in the AWS Key Management Service API Reference.
	// +optional
	KMSMasterKeyID *string `json:"kmsMasterKeyId,omitempty"`

	// The policy that defines who can access your topic. By default,
	// only the topic owner can publish or subscribe to the topic.
	// +optional
	Policy *string `json:"policy,omitempty"`

	// DeliveryRetryPolicy - the JSON serialization of the effective
	// delivery policy, taking system defaults into account
	// +optional
	DeliveryPolicy *string `json:"deliveryPolicy,omitempty"`

	// Whether or not this should be a fifo-topic
	// +immutable
	// +optional
	FifoTopic *bool `json:"fifoTopic,omitempty"`

	// Tags represetnt a list of user-provided metadata that can be associated with a
	// SNS Topic. For more information about tagging,
	// see Tagging SNS Topics (https://docs.aws.amazon.com/sns/latest/dg/sns-tags.html)
	// in the SNS User Guide.
	// +immutable
	// +optional
	Tags []Tag `json:"tags,omitempty"`
}

// TopicSpec defined the desired state of a AWS SNS Topic
type TopicSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       TopicParameters `json:"forProvider"`
}

// TopicObservation represents the observed state of a AWS SNS Topic
type TopicObservation struct {

	// Owner refers to owner of SNS Topic
	// +optional
	Owner *string `json:"owner,omitempty"`
	// ConfirmedSubscriptions - The no of confirmed subscriptions
	// +optional
	ConfirmedSubscriptions *int64 `json:"confirmedSubscriptions,omitempty"`

	// PendingSubscriptions - The no of pending subscriptions
	// +optional
	PendingSubscriptions *int64 `json:"pendingSubscriptions,omitempty"`

	// DeletedSubscriptions - The no of deleted subscriptions
	// +optional
	DeletedSubscriptions *int64 `json:"deletedSubscriptions,omitempty"`

	// ARN is the Amazon Resource Name (ARN) specifying the SNS Topic.
	ARN string `json:"arn"`
}

// TopicStatus is the status of AWS SNS Topic
type TopicStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          TopicObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// Topic defines a managed resource that represents state of a AWS Topic
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="TOPIC-NAME",type="string",JSONPath=".spec.forProvider.name"
// +kubebuilder:printcolumn:name="DISPLAY-NAME",type="string",JSONPath=".spec.forProvider.displayName"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,aws}
type Topic struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TopicSpec   `json:"spec"`
	Status TopicStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// TopicList contains a list of Topic
type TopicList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Topic `json:"items"`
}
