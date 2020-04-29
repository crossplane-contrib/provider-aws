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

// IAMUserParameters define the desired state of an AWS IAM User.
type IAMUserParameters struct {
	// The path for the user name.
	// +optional
	Path *string `json:"path,omitempty"`

	// The ARN of the policy that is used to set the permissions boundary for the
	// user.
	// +optional
	PermissionsBoundary *string `json:"permissionsBoundary,omitempty"`

	// A list of tags that you want to attach to the newly created user.
	// +optional
	Tags []Tag `json:"tags,omitempty"`
}

// An IAMUserSpec defines the desired state of an IAM User.
type IAMUserSpec struct {
	runtimev1alpha1.ResourceSpec `json:",inline"`
	ForProvider                  IAMUserParameters `json:"forProvider"`
}

// IAMUserObservation keeps the state for the external resource
type IAMUserObservation struct {
	// The Amazon Resource Name (ARN) that identifies the user.
	ARN string `json:"arn,omitempty"`

	// The stable and unique string identifying the user.
	UserID string `json:"userId,omitempty"`
}

// An IAMUserStatus represents the observed state of an IAM User.
type IAMUserStatus struct {
	runtimev1alpha1.ResourceStatus `json:",inline"`
	AtProvider                     IAMUserObservation `json:"atProvider"`
}

// +kubebuilder:object:root=true

// An IAMUser is a managed resource that represents an AWS IAM IAMUser.
// +kubebuilder:printcolumn:name="ARN",type="string",JSONPath=".status.atProvider.arn"
// +kubebuilder:printcolumn:name="ID",type="string",JSONPath=".status.atProvider.userId"
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,aws}
type IAMUser struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   IAMUserSpec   `json:"spec"`
	Status IAMUserStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// IAMUserList contains a list of IAM Users
type IAMUserList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []IAMUser `json:"items"`
}
