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

// IAMGroupPolicyAttachmentParameters define the desired state of an AWS IAMGroupPolicyAttachment.
type IAMGroupPolicyAttachmentParameters struct {

	// PolicyARN is the Amazon Resource Name (ARN) of the IAM policy you want to
	// attach.
	// +immutable
	PolicyARN string `json:"policyArn,omitempty"`

	// PolicyARNRef references an IAMPolicy to retrieve its Policy ARN.
	// +optional
	PolicyARNRef *runtimev1alpha1.Reference `json:"policyArnRef,omitempty"`

	// PolicyARNSelector selects a reference to an IAMPolicy to retrieve its
	// Policy ARN
	// +optional
	PolicyARNSelector *runtimev1alpha1.Selector `json:"policyArnSelector,omitempty"`

	// GroupName presents the name of the IAMGroup.
	// +immutable
	GroupName string `json:"groupName,omitempty"`

	// GroupNameRef references to an IAMGroup to retrieve its groupName
	// +optional
	GroupNameRef *runtimev1alpha1.Reference `json:"groupNameRef,omitempty"`

	// GroupNameSelector selects a reference to an IAMGroup to retrieve its groupName
	// +optional
	GroupNameSelector *runtimev1alpha1.Selector `json:"groupNameSelector,omitempty"`
}

// An IAMGroupPolicyAttachmentSpec defines the desired state of an
// IAMGroupPolicyAttachment.
type IAMGroupPolicyAttachmentSpec struct {
	runtimev1alpha1.ResourceSpec `json:",inline"`
	ForProvider                  IAMGroupPolicyAttachmentParameters `json:"forProvider"`
}

// IAMGroupPolicyAttachmentObservation keeps the state for the external resource
type IAMGroupPolicyAttachmentObservation struct {
	// AttachedPolicyARN is the arn for the attached policy. If nil, the policy
	// is not yet attached
	AttachedPolicyARN string `json:"attachedPolicyArn"`
}

// An IAMGroupPolicyAttachmentStatus represents the observed state of an
// IAMGroupPolicyAttachment.
type IAMGroupPolicyAttachmentStatus struct {
	runtimev1alpha1.ResourceStatus `json:",inline"`
	AtProvider                     IAMGroupPolicyAttachmentObservation `json:"atProvider"`
}

// +kubebuilder:object:root=true

// An IAMGroupPolicyAttachment is a managed resource that represents an AWS IAM
// Group policy attachment.
// +kubebuilder:printcolumn:name="GROUPNAME",type="string",JSONPath=".spec.forProvider.groupName"
// +kubebuilder:printcolumn:name="POLICYARN",type="string",JSONPath=".spec.forProvider.policyArn"
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,aws}
type IAMGroupPolicyAttachment struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   IAMGroupPolicyAttachmentSpec   `json:"spec"`
	Status IAMGroupPolicyAttachmentStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// IAMGroupPolicyAttachmentList contains a list of IAMGroupPolicyAttachments
type IAMGroupPolicyAttachmentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []IAMGroupPolicyAttachment `json:"items"`
}
