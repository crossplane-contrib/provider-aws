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

// RolePolicyAttachmentParameters define the desired state of an AWS IAM
// Role policy attachment.
type RolePolicyAttachmentParameters struct {

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

	// RoleName presents the name of the IAM role.
	// +immutable
	// +crossplane:generate:reference:type=Role
	RoleName string `json:"roleName,omitempty"`

	// RoleNameRef references a Role to retrieve its Name
	// +optional
	RoleNameRef *xpv1.Reference `json:"roleNameRef,omitempty"`

	// RoleNameSelector selects a reference to a Role to retrieve its Name
	// +optional
	RoleNameSelector *xpv1.Selector `json:"roleNameSelector,omitempty"`
}

// A RolePolicyAttachmentSpec defines the desired state of an
// RolePolicyAttachment.
type RolePolicyAttachmentSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       RolePolicyAttachmentParameters `json:"forProvider"`
}

// RolePolicyAttachmentExternalStatus keeps the state for the external resource
type RolePolicyAttachmentExternalStatus struct {
	// AttachedPolicyARN is the arn for the attached policy. If nil, the policy
	// is not yet attached
	AttachedPolicyARN string `json:"attachedPolicyArn"`
}

// A RolePolicyAttachmentStatus represents the observed state of an
// RolePolicyAttachment.
type RolePolicyAttachmentStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          RolePolicyAttachmentExternalStatus `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A RolePolicyAttachment is a managed resource that represents an AWS IAM
// Role policy attachment.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="ROLENAME",type="string",JSONPath=".spec.forProvider.roleName"
// +kubebuilder:printcolumn:name="POLICYARN",type="string",JSONPath=".spec.forProvider.policyArn"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,aws}
type RolePolicyAttachment struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RolePolicyAttachmentSpec   `json:"spec"`
	Status RolePolicyAttachmentStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// RolePolicyAttachmentList contains a list of RolePolicyAttachments
type RolePolicyAttachmentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RolePolicyAttachment `json:"items"`
}
