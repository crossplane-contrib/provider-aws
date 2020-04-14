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
	"github.com/pkg/errors"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
)

// UserNameReferencerForUserPolicyAttachment is an attribute referencer that retrieves Name from a referenced User
type UserNameReferencerForUserPolicyAttachment struct {
	UserNameReferencer `json:",inline"`
}

// Assign assigns the retrieved name to the managed resource
func (v *UserNameReferencerForUserPolicyAttachment) Assign(res resource.CanReference, value string) error {
	p, ok := res.(*UserPolicyAttachment)
	if !ok {
		return errors.New(errResourceIsNotUserPolicyAttachment)
	}

	p.Spec.ForProvider.UserName = value
	return nil
}

// Error strings
const (
	errResourceIsNotUserPolicyAttachment = "the managed resource is not an UserPolicyAttachment"
)

// UserPolicyAttachmentParameters define the desired state of an AWS IAM
// User policy attachment.
type UserPolicyAttachmentParameters struct {

	// PolicyARN is the Amazon Resource Name (ARN) of the IAM policy you want to
	// attach.
	// +immutable
	PolicyARN string `json:"policyArn"`

	// UserName presents the name of the IAM user.
	// +optional
	UserName string `json:"userName,omitempty"`

	// UserNameRef references to an User to retrieve its Name
	// +optional
	UserNameRef *UserNameReferencerForUserPolicyAttachment `json:"userNameRef,omitempty"`
}

// An UserPolicyAttachmentSpec defines the desired state of an
// UserPolicyAttachment.
type UserPolicyAttachmentSpec struct {
	runtimev1alpha1.ResourceSpec `json:",inline"`
	ForProvider                  UserPolicyAttachmentParameters `json:"forProvider"`
}

// UserPolicyAttachmentObservation keeps the state for the external resource
type UserPolicyAttachmentObservation struct {
	// AttachedPolicyARN is the arn for the attached policy. If nil, the policy
	// is not yet attached
	AttachedPolicyARN string `json:"attachedPolicyArn"`
}

// An UserPolicyAttachmentStatus represents the observed state of an
// UserPolicyAttachment.
type UserPolicyAttachmentStatus struct {
	runtimev1alpha1.ResourceStatus `json:",inline"`
	AtProvider                     UserPolicyAttachmentObservation `json:"atProvider"`
}

// +kubebuilder:object:root=true

// An UserPolicyAttachment is a managed resource that represents an AWS IAM
// User policy attachment.
// +kubebuilder:printcolumn:name="USERNAME",type="string",JSONPath=".spec.forProvider.userName"
// +kubebuilder:printcolumn:name="POLICYARN",type="string",JSONPath=".spec.forProvider.policyArn"
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster
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
