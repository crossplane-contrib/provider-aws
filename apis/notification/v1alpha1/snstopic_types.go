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

// SNSTopicParameters define the desired state of a AWS SNS Topic
type SNSTopicParameters struct {
	// Region is the region you'd like your SNSTopic to be created in.
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

	// Tags represetnt a list of user-provided metadata that can be associated with a
	// SNS Topic. For more information about tagging,
	// see Tagging SNS Topics (https://docs.aws.amazon.com/sns/latest/dg/sns-tags.html)
	// in the SNS User Guide.
	// +immutable
	// +optional
	Tags []Tag `json:"tags,omitempty"`
}

// SNSTopicSpec defined the desired state of a AWS SNS Topic
type SNSTopicSpec struct {
	runtimev1alpha1.ResourceSpec `json:",inline"`
	ForProvider                  SNSTopicParameters `json:"forProvider"`
}

// SNSTopicObservation represents the observed state of a AWS SNS Topic
type SNSTopicObservation struct {

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

// SNSTopicStatus is the status of AWS SNS Topic
type SNSTopicStatus struct {
	runtimev1alpha1.ResourceStatus `json:",inline"`
	AtProvider                     SNSTopicObservation `json:"atProvider"`
}

// +kubebuilder:object:root=true

// SNSTopic defines a managed resource that represents state of a AWS SNSTopic
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="TOPIC-NAME",type="string",JSONPath=".spec.forProvider.name"
// +kubebuilder:printcolumn:name="DISPLAY-NAME",type="string",JSONPath=".spec.forProvider.displayName"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,aws}
type SNSTopic struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SNSTopicSpec   `json:"spec"`
	Status SNSTopicStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

//SNSTopicList contains a list of SNSTopic
type SNSTopicList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SNSTopic `json:"items"`
}
