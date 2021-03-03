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

package v1alpha1

import xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"

// CustomPlatformApplicationParameters are the additional parameters for
// PlatformApplicationParameters.
type CustomPlatformApplicationParameters struct {
	// FailureFeedbackRoleARNRef is a reference to an IAM Role used to set
	// the FailureFeedbackRoleARN.
	// +optional
	FailureFeedbackRoleARNRef *xpv1.Reference `json:"failureFeedbackRoleARNRef,omitempty"`

	// FailureFeedbackRoleArnSelector selects references to IAM Role used
	// to set the FailureFeedbackRoleArn.
	// +optional
	FailureFeedbackRoleARNSelector *xpv1.Selector `json:"failureFeedbackRoleARNSelector,omitempty"`

	// SuccessFeedbackRoleARNRef is a reference to an IAM Role used to set
	// the SuccessFeedbackRoleARN.
	// +optional
	SuccessFeedbackRoleARNRef *xpv1.Reference `json:"successFeedbackRoleARNRef,omitempty"`

	// SuccessFeedbackRoleARNSelector selects references to IAM Role used
	// to set the SuccessFeedbackRoleARN.
	// +optional
	SuccessFeedbackRoleARNSelector *xpv1.Selector `json:"successFeedbackRoleARNSelector,omitempty"`

	// EventDeliveryFailureRef is a reference to a a Topic used to set
	// the EventDeliveryFailure.
	// +optional
	EventDeliveryFailureRef *xpv1.Reference `json:"eventDeliveryFailureRef,omitempty"`

	// EventDeliveryFailureSelector selects references to Topic used to set the
	// EventDeliveryFailure.
	// +optional
	EventDeliveryFailureSelector *xpv1.Selector `json:"eventDeliveryFailureSelector,omitempty"`

	// EventEndpointCreatedRef is a reference to a a Topic used to set
	// the EventEndpointCreated.
	// +optional
	EventEndpointCreatedRef *xpv1.Reference `json:"eventEndpointCreatedRef,omitempty"`

	// EventEndpointCreatedSelector selects references to Topic used to set the
	// EventEndpointCreated.
	// +optional
	EventEndpointCreatedSelector *xpv1.Selector `json:"eventEndpointCreatedSelector,omitempty"`

	// EventEndpointDeletedRef is a reference to a a Topic used to set
	// the EventEndpointDeleted.
	// +optional
	EventEndpointDeletedRef *xpv1.Reference `json:"eventEndpointDeletedRef,omitempty"`

	// EventEndpointDeletedSelector selects references to Topic used to set the
	// EventEndpointDeleted.
	// +optional
	EventEndpointDeletedSelector *xpv1.Selector `json:"eventEndpointDeletedSelector,omitempty"`

	// EventEndpointUpdatedRef is a reference to a a Topic used to set
	// the EventEndpointUpdated.
	// +optional
	EventEndpointUpdatedRef *xpv1.Reference `json:"eventEndpointUpdatedRef,omitempty"`

	// EventEndpointUpdatedSelector selects references to Topic used to set the
	// EventEndpointUpdated.
	// +optional
	EventEndpointUpdatedSelector *xpv1.Selector `json:"eventEndpointUpdatedSelector,omitempty"`
}

// CustomPlatformEndpointParameters are the additional parameters for
// PlatformEndpointParameters.
type CustomPlatformEndpointParameters struct {

	// PlatformApplicationARN is ARN of the Platform Application that this Endpoint
	// targets.
	PlatformApplicationARN *string `json:"platformApplicationArn,omitempty"`

	// PlatformApplicationARNRef is a reference to a a PlatformApplication used
	// to set the PlatformApplicationARN.
	// +optional
	PlatformApplicationARNRef *xpv1.Reference `json:"platformApplicationArnRef,omitempty"`

	// PlatformApplicationARNSelector selects references to PlatformApplication
	// used to set the PlatformApplicationARN.
	// +optional
	PlatformApplicationARNSelector *xpv1.Selector `json:"platformApplicationArnSelector,omitempty"`
}

// TopicAttributes refers to AWS SNS Topic Attributes List
// ref: https://docs.aws.amazon.com/cli/latest/reference/sns/get-topic-attributes.html#output
type TopicAttributes string

const (
	// TopicDisplayName is Display Name of SNS Topic
	TopicDisplayName TopicAttributes = "DisplayName"
	// TopicDeliveryPolicy is Delivery Policy of SNS Topic
	TopicDeliveryPolicy TopicAttributes = "DeliveryPolicy"
	// TopicKmsMasterKeyID is KmsMasterKeyId of SNS Topic
	TopicKmsMasterKeyID TopicAttributes = "KmsMasterKeyId"
	// TopicPolicy is Policy of SNS Topic
	TopicPolicy TopicAttributes = "Policy"
	// TopicOwner is Owner of SNS Topic
	TopicOwner TopicAttributes = "Owner"
)

// CustomTopicParameters are the additional parameters for TopicParameters.
type CustomTopicParameters struct{}
