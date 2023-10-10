/*
Copyright 2021 The Crossplane Authors.

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

package manualv1alpha1

import (
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ExternalAnnotation defines a virtual external name annotation for importing existing resources.
type ExternalAnnotation struct {
	Groupname  *string `json:"groupname"`
	UserPoolID *string `json:"userPoolId"`
	Username   *string `json:"username"`
}

// GroupUserMembershipParameters define the desired state of an AWS GroupUserMembership.
type GroupUserMembershipParameters struct {
	// Region is which region the Group will be created.
	// +kubebuilder:validation:Required
	Region string `json:"region"`

	// UserPoolID is the Amazon Cognito Group Name (Group) of the Cognito User-Pool group you want to
	// add User to.
	// +immutable
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-aws/apis/cognitoidentityprovider/v1alpha1.UserPool
	UserPoolID string `json:"userPoolId,omitempty"`

	// UserPoolIDRef references to an Group to retrieve its userPoolId
	// +optional
	// +immutable
	UserPoolIDRef *xpv1.Reference `json:"userPoolIdRef,omitempty"`

	// UserPoolIDSelector selects a reference to an Group to retrieve its userPoolId
	// +optional
	UserPoolIDSelector *xpv1.Selector `json:"userPoolIdSelector,omitempty"`

	// Groupname is the Amazon Cognito Group Name (Group) of the Cognito User-Pool group you want to
	// add User to.
	// +immutable
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-aws/apis/cognitoidentityprovider/v1alpha1.Group
	Groupname string `json:"groupname,omitempty"`

	// GroupnameRef references to an Group to retrieve its groupName
	// +optional
	// +immutable
	GroupnameRef *xpv1.Reference `json:"groupnameRef,omitempty"`

	// GroupnameSelector selects a reference to an Group to retrieve its groupName
	// +optional
	GroupnameSelector *xpv1.Selector `json:"groupnameSelector,omitempty"`

	// Username presents the name of the User.
	// +immutable
	Username string `json:"username,omitempty"`
}

// An GroupUserMembershipSpec defines the desired state of an
// GroupUserMembership.
type GroupUserMembershipSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       GroupUserMembershipParameters `json:"forProvider"`
}

// GroupUserMembershipObservation keeps the state for the external resource
type GroupUserMembershipObservation struct {
	// Groupname is the name of the attached group. If nil, the group
	// is not yet attached
	Groupname *string `json:"groupname,omitempty"`
}

// An GroupUserMembershipStatus represents the observed state of an
// GroupUserMembership.
type GroupUserMembershipStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          GroupUserMembershipObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// An GroupUserMembership is a managed resource that represents an AWS Cognito
// User group membership.
// +kubebuilder:printcolumn:name="USERNAME",type="string",JSONPath=".spec.forProvider.username"
// +kubebuilder:printcolumn:name="GROUPNAME",type="string",JSONPath=".spec.forProvider.groupname"
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
