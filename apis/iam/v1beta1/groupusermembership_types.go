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

// GroupUserMembershipParameters define the desired state of an AWS GroupUserMembership.
type GroupUserMembershipParameters struct {

	// GroupName is the Amazon IAM Group Name (Group) of the IAM group you want to
	// add User to.
	// +immutable
	// +crossplane:generate:reference:type=Group
	GroupName string `json:"groupName,omitempty"`

	// GroupNameRef references to a Group to retrieve its groupName
	// +optional
	// +immutable
	GroupNameRef *xpv1.Reference `json:"groupNameRef,omitempty"`

	// GroupNameSelector selects a reference to a Group to retrieve its groupName
	// +optional
	GroupNameSelector *xpv1.Selector `json:"groupNameSelector,omitempty"`

	// UserName presents the name of the User.
	// +immutable
	// +crossplane:generate:reference:type=User
	UserName string `json:"userName,omitempty"`

	// UserNameRef references to a User to retrieve its userName
	// +optional
	// +immutable
	UserNameRef *xpv1.Reference `json:"userNameRef,omitempty"`

	// UserNameSelector selects a reference to a User to retrieve its userName
	// +optional
	UserNameSelector *xpv1.Selector `json:"userNameSelector,omitempty"`
}

// A GroupUserMembershipSpec defines the desired state of an
// GroupUserMembership.
type GroupUserMembershipSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       GroupUserMembershipParameters `json:"forProvider"`
}

// GroupUserMembershipObservation keeps the state for the external resource
type GroupUserMembershipObservation struct {
	// AttachedGroupARN is the arn for the attached group. If nil, the group
	// is not yet attached
	AttachedGroupARN string `json:"attachedGroupArn"`
}

// A GroupUserMembershipStatus represents the observed state of a GroupUserMembership.
type GroupUserMembershipStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          GroupUserMembershipObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A GroupUserMembership is a managed resource that represents an AWS IAM
// User group membership.
// +kubebuilder:printcolumn:name="USERNAME",type="string",JSONPath=".spec.forProvider.userName"
// +kubebuilder:printcolumn:name="GROUPNAME",type="string",JSONPath=".spec.forProvider.groupName"
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,aws}
type GroupUserMembership struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GroupUserMembershipSpec   `json:"spec"`
	Status GroupUserMembershipStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// GroupUserMembershipList contains a list of GroupUserMemberships
type GroupUserMembershipList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GroupUserMembership `json:"items"`
}
