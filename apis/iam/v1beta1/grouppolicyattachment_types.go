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

// GroupPolicyAttachmentParameters define the desired state of an AWS GroupPolicyAttachment.
type GroupPolicyAttachmentParameters struct {

	// PolicyARN is the Amazon Resource Name (ARN) of the IAM policy you want to
	// attach.
	// +immutable
	// +crossplane:generate:reference:type=Policy
	// +crossplane:generate:reference:extractor=PolicyARN()
	PolicyARN string `json:"policyArn,omitempty"`

	// PolicyARNRef references a Policy to retrieve its Policy ARN.
	// +optional
	PolicyARNRef *xpv1.Reference `json:"policyArnRef,omitempty"`

	// PolicyARNSelector selects a reference to a Policy to retrieve its
	// Policy ARN
	// +optional
	PolicyARNSelector *xpv1.Selector `json:"policyArnSelector,omitempty"`

	// GroupName presents the name of the Group.
	// +immutable
	// +crossplane:generate:reference:type=Group
	GroupName string `json:"groupName,omitempty"`

	// GroupNameRef references to an Group to retrieve its groupName
	// +optional
	GroupNameRef *xpv1.Reference `json:"groupNameRef,omitempty"`

	// GroupNameSelector selects a reference to an Group to retrieve its groupName
	// +optional
	GroupNameSelector *xpv1.Selector `json:"groupNameSelector,omitempty"`
}

// A GroupPolicyAttachmentSpec defines the desired state of a
// GroupPolicyAttachment.
type GroupPolicyAttachmentSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       GroupPolicyAttachmentParameters `json:"forProvider"`
}

// GroupPolicyAttachmentObservation keeps the state for the external resource
type GroupPolicyAttachmentObservation struct {
	// AttachedPolicyARN is the arn for the attached policy. If nil, the policy
	// is not yet attached
	AttachedPolicyARN string `json:"attachedPolicyArn"`
}

// A GroupPolicyAttachmentStatus represents the observed state of a
// GroupPolicyAttachment.
type GroupPolicyAttachmentStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          GroupPolicyAttachmentObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A GroupPolicyAttachment is a managed resource that represents an AWS IAM
// Group policy attachment.
// +kubebuilder:printcolumn:name="GROUPNAME",type="string",JSONPath=".spec.forProvider.groupName"
// +kubebuilder:printcolumn:name="POLICYARN",type="string",JSONPath=".spec.forProvider.policyArn"
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,aws}
type GroupPolicyAttachment struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GroupPolicyAttachmentSpec   `json:"spec"`
	Status GroupPolicyAttachmentStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// GroupPolicyAttachmentList contains a list of GroupPolicyAttachments
type GroupPolicyAttachmentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GroupPolicyAttachment `json:"items"`
}
