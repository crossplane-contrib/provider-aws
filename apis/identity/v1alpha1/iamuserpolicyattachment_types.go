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

// IAMUserPolicyAttachmentParameters define the desired state of an AWS IAMUserPolicyAttachment.
type IAMUserPolicyAttachmentParameters struct {

	// PolicyARN is the Amazon Resource Name (ARN) of the IAM policy you want to
	// attach.
	// +immutable
	PolicyARN string `json:"policyArn"`

	// UserName presents the name of the IAMUser.
	// +optional
	UserName *string `json:"userName,omitempty"`

	// UserNameRef references to an IAMUser to retrieve its userName
	// +optional
	UserNameRef *runtimev1alpha1.Reference `json:"userNameRef,omitempty"`

	// UserNameSelector selects a reference to an IAMUser to retrieve its userName
	// +optional
	UserNameSelector *runtimev1alpha1.Selector `json:"userNameSelector,omitempty"`
}

// An IAMUserPolicyAttachmentSpec defines the desired state of an
// IAMUserPolicyAttachment.
type IAMUserPolicyAttachmentSpec struct {
	runtimev1alpha1.ResourceSpec `json:",inline"`
	ForProvider                  IAMUserPolicyAttachmentParameters `json:"forProvider"`
}

// IAMUserPolicyAttachmentObservation keeps the state for the external resource
type IAMUserPolicyAttachmentObservation struct {
	// AttachedPolicyARN is the arn for the attached policy. If nil, the policy
	// is not yet attached
	AttachedPolicyARN string `json:"attachedPolicyArn"`
}

// An IAMUserPolicyAttachmentStatus represents the observed state of an
// IAMUserPolicyAttachment.
type IAMUserPolicyAttachmentStatus struct {
	runtimev1alpha1.ResourceStatus `json:",inline"`
	AtProvider                     IAMUserPolicyAttachmentObservation `json:"atProvider"`
}

// +kubebuilder:object:root=true

// An IAMUserPolicyAttachment is a managed resource that represents an AWS IAM
// User policy attachment.
// +kubebuilder:printcolumn:name="USERNAME",type="string",JSONPath=".spec.forProvider.userName"
// +kubebuilder:printcolumn:name="POLICYARN",type="string",JSONPath=".spec.forProvider.policyArn"
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster
type IAMUserPolicyAttachment struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   IAMUserPolicyAttachmentSpec   `json:"spec"`
	Status IAMUserPolicyAttachmentStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// IAMUserPolicyAttachmentList contains a list of IAMUserPolicyAttachments
type IAMUserPolicyAttachmentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []IAMUserPolicyAttachment `json:"items"`
}
