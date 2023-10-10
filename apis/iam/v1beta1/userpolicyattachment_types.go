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

// UserPolicyAttachmentParameters define the desired state of an AWS UserPolicyAttachment.
type UserPolicyAttachmentParameters struct {

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

	// UserName presents the name of the User.
	// +immutable
	// +crossplane:generate:reference:type=User
	UserName string `json:"userName,omitempty"`

	// UserNameRef references to an User to retrieve its userName
	// +optional
	UserNameRef *xpv1.Reference `json:"userNameRef,omitempty"`

	// UserNameSelector selects a reference to an User to retrieve its userName
	// +optional
	UserNameSelector *xpv1.Selector `json:"userNameSelector,omitempty"`
}

// A UserPolicyAttachmentSpec defines the desired state of an
// UserPolicyAttachment.
type UserPolicyAttachmentSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       UserPolicyAttachmentParameters `json:"forProvider"`
}

// UserPolicyAttachmentObservation keeps the state for the external resource
type UserPolicyAttachmentObservation struct {
	// AttachedPolicyARN is the arn for the attached policy. If nil, the policy
	// is not yet attached
	AttachedPolicyARN string `json:"attachedPolicyArn"`
}

// A UserPolicyAttachmentStatus represents the observed state of a UserPolicyAttachment.
type UserPolicyAttachmentStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          UserPolicyAttachmentObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A UserPolicyAttachment is a managed resource that represents an AWS IAM
// User policy attachment.
// +kubebuilder:printcolumn:name="USERNAME",type="string",JSONPath=".spec.forProvider.userName"
// +kubebuilder:printcolumn:name="POLICYARN",type="string",JSONPath=".spec.forProvider.policyArn"
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,aws}
type UserPolicyAttachment struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   UserPolicyAttachmentSpec   `json:"spec"`
	Status UserPolicyAttachmentStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// UserPolicyAttachmentList contains a list of UserPolicyAttachments
type UserPolicyAttachmentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []UserPolicyAttachment `json:"items"`
}
