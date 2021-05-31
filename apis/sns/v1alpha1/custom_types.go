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

import runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"

// CustomPlatformApplicationParameters are the additional parameters for
// PlatformApplicationParameters.
type CustomPlatformApplicationParameters struct {
	// FailureFeedbackRoleARNRef is a reference to an IAM Role used to set
	// the FailureFeedbackRoleARN.
	// +optional
	FailureFeedbackRoleARNRef *runtimev1alpha1.Reference `json:"failureFeedbackRoleARNRef,omitempty"`

	// FailureFeedbackRoleArnSelector selects references to IAM Role used
	// to set the FailureFeedbackRoleArn.
	// +optional
	FailureFeedbackRoleARNSelector *runtimev1alpha1.Selector `json:"failureFeedbackRoleARNSelector,omitempty"`

	// SuccessFeedbackRoleARNRef is a reference to an IAM Role used to set
	// the SuccessFeedbackRoleARN.
	// +optional
	SuccessFeedbackRoleARNRef *runtimev1alpha1.Reference `json:"successFeedbackRoleARNRef,omitempty"`

	// SuccessFeedbackRoleARNSelector selects references to IAM Role used
	// to set the SuccessFeedbackRoleARN.
	// +optional
	SuccessFeedbackRoleARNSelector *runtimev1alpha1.Selector `json:"successFeedbackRoleARNSelector,omitempty"`
}

// CustomPlatformEndpointParameters are the additional parameters for
// PlatformEndpointParameters.
type CustomPlatformEndpointParameters struct{}

// CustomTopicParameters are the additional parameters for TopicParameters.
type CustomTopicParameters struct{}
