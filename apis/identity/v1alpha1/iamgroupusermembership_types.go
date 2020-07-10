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

// IAMGroupUserMembershipParameters define the desired state of an AWS IAMGroupUserMembership.
type IAMGroupUserMembershipParameters struct {

	// GroupName is the Amazon IAM Group Name (IAMGroup) of the IAM group you want to
	// add User to.
	// +optional
	GroupName *string `json:"groupName,omitempty"`

	// GroupNameRef references to an IAMGroup to retrieve its groupName
	// +optional
	GroupNameRef *runtimev1alpha1.Reference `json:"groupNameRef,omitempty"`

	// GroupNameSelector selects a reference to an IAMGroup to retrieve its groupName
	// +optional
	GroupNameSelector *runtimev1alpha1.Selector `json:"groupNameSelector,omitempty"`

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

// An IAMGroupUserMembershipSpec defines the desired state of an
// IAMGroupUserMembership.
type IAMGroupUserMembershipSpec struct {
	runtimev1alpha1.ResourceSpec `json:",inline"`
	ForProvider                  IAMGroupUserMembershipParameters `json:"forProvider"`
}

// IAMGroupUserMembershipObservation keeps the state for the external resource
type IAMGroupUserMembershipObservation struct {
	// AttachedGroupARN is the arn for the attached group. If nil, the group
	// is not yet attached
	AttachedGroupARN string `json:"attachedGroupArn"`
}

// An IAMGroupUserMembershipStatus represents the observed state of an
// IAMGroupUserMembership.
type IAMGroupUserMembershipStatus struct {
	runtimev1alpha1.ResourceStatus `json:",inline"`
	AtProvider                     IAMGroupUserMembershipObservation `json:"atProvider"`
}

// +kubebuilder:object:root=true

// An IAMGroupUserMembership is a managed resource that represents an AWS IAM
// User group membership.
// +kubebuilder:printcolumn:name="USERNAME",type="string",JSONPath=".spec.forProvider.userName"
// +kubebuilder:printcolumn:name="GROUPNAME",type="string",JSONPath=".spec.forProvider.groupName"
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster
type IAMGroupUserMembership struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   IAMGroupUserMembershipSpec   `json:"spec"`
	Status IAMGroupUserMembershipStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// IAMGroupUserMembershipList contains a list of IAMGroupUserMemberships
type IAMGroupUserMembershipList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []IAMGroupUserMembership `json:"items"`
}
