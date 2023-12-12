/*
Copyright 2023 The Crossplane Authors.

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
	extv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// RolePolicyParameters define the desired state of an AWS IAM Role Inline Policy.
type RolePolicyParameters struct {

	// The JSON policy document that is the content for the policy.
	Document extv1.JSON `json:"document"`

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

// An RolePolicySpec defines the desired state of an RolePolicy.
type RolePolicySpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       RolePolicyParameters `json:"forProvider"`
}

// RolePolicyObservation keeps the state for the external resource
type RolePolicyObservation struct {
}

// An RolePolicyStatus represents the observed state of an RolePolicy.
type RolePolicyStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          RolePolicyObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// An RolePolicy is a managed resource that represents an AWS IAM RolePolicy.
// +kubebuilder:printcolumn:name="ROLENAME",type="string",JSONPath=".spec.forProvider.roleName"
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,aws}
type RolePolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RolePolicySpec   `json:"spec"`
	Status RolePolicyStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// RolePolicyList contains a list of Policies
type RolePolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RolePolicy `json:"items"`
}
