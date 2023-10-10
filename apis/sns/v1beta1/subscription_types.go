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

// SubscriptionParameters define the desired state of a AWS SNS Topic
type SubscriptionParameters struct {
	// Region is the region you'd like your Subscription to be in.
	Region string `json:"region"`

	// TopicArn is the Arn of the SNS Topic
	// +immutable
	TopicARN string `json:"topicArn,omitempty"`

	// TopicArnRef references a SNS Topic and retrieves its TopicArn
	// +optional
	TopicARNRef *xpv1.Reference `json:"topicArnRef,omitempty"`

	// TopicArnSelector selects a reference to a SNS Topic and retrieves
	// its TopicArn
	// +optional
	TopicARNSelector *xpv1.Selector `json:"topicArnSelector,omitempty"`

	// The subscription's protocol.
	// +immutable
	Protocol string `json:"protocol"`

	// The subscription's endpoint
	// +immutable
	Endpoint string `json:"endpoint"`

	//  DeliveryPolicy defines how Amazon SNS retries failed
	//  deliveries to HTTP/S endpoints.
	// +optional
	DeliveryPolicy *string `json:"deliveryPolicy,omitempty"`

	//  The simple JSON object that lets your subscriber receive
	//  only a subset of messages, rather than receiving every message published
	//  to the topic.
	// +optional
	FilterPolicy *string `json:"filterPolicy,omitempty"`

	//  FilterPolicyScope can be MessageAttributes or MessageBody
	// +optional
	FilterPolicyScope *string `json:"filterPolicyScope,omitempty"`

	//  When set to true, enables raw message delivery
	//  to Amazon SQS or HTTP/S endpoints. This eliminates the need for the endpoints
	//  to process JSON formatting, which is otherwise created for Amazon SNS
	//  metadata.
	// +optional
	RawMessageDelivery *string `json:"rawMessageDelivery,omitempty"`

	//  When specified, sends undeliverable messages to the
	//  specified Amazon SQS dead-letter queue. Messages that can't be delivered
	//  due to client errors (for example, when the subscribed endpoint is unreachable)
	//  or server errors (for example, when the service that powers the subscribed
	//  endpoint becomes unavailable) are held in the dead-letter queue for further
	//  analysis or reprocessing.
	// +optional
	RedrivePolicy *string `json:"redrivePolicy,omitempty"`
}

// SubscriptionSpec defined the desired state of a AWS SNS Topic
type SubscriptionSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       SubscriptionParameters `json:"forProvider"`
}

// ConfirmationStatus represents Status of SNS Subscription Confirmation
type ConfirmationStatus string

const (
	// ConfirmationPending represents Pending Confirmation Request for SNS Subscription
	ConfirmationPending ConfirmationStatus = "ConfirmationPending"
	// ConfirmationSuccessful represents confirmed SNS Subscription
	ConfirmationSuccessful ConfirmationStatus = "Confirmed"
)

// SubscriptionObservation represents the observed state of a AWS SNS Topic
type SubscriptionObservation struct {

	// The subscription's owner.
	// +optional
	Owner *string `json:"owner,omitempty"`

	// Status represents Confirmation Status of SNS Subscription
	// +optional
	Status *ConfirmationStatus `json:"status,omitempty"`

	// ConfirmationWasAuthenticated – true if the subscription confirmation
	// request was authenticated.
	// +optional
	ConfirmationWasAuthenticated *bool `json:"confirmationWasAuthenticated,omitempty"`
}

// SubscriptionStatus is the status of AWS SNS Topic
type SubscriptionStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          SubscriptionObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// Subscription defines a managed resource that represents state of
// a AWS SNS Subscription
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="ENDPOINT",type="string",JSONPath=".spec.forProvider.endpoint"
// +kubebuilder:printcolumn:name="PROTOCOL",type="string",JSONPath=".spec.forProvider.protocol"
// +kubebuilder:printcolumn:name="STATUS",type="string",JSONPath=".status.atProvider.status"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,aws}
type Subscription struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SubscriptionSpec   `json:"spec"`
	Status SubscriptionStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// SubscriptionList contains a list of Topic
type SubscriptionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Subscription `json:"items"`
}
