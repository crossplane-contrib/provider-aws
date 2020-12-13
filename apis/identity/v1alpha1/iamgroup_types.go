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

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
)

// IAMGroupParameters define the desired state of an AWS IAM Group.
type IAMGroupParameters struct {
	// The path for the group name.
	// +optional
	Path *string `json:"path,omitempty"`
}

// An IAMGroupSpec defines the desired state of an IAM Group.
type IAMGroupSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       IAMGroupParameters `json:"forProvider,omitempty"`
}

// IAMGroupObservation keeps the state for the external resource
type IAMGroupObservation struct {
	// The Amazon Resource Name (ARN) that identifies the group.
	ARN string `json:"arn,omitempty"`

	// The stable and unique string identifying the group.
	GroupID string `json:"groupId,omitempty"`
}

// An IAMGroupStatus represents the observed state of an IAM Group.
type IAMGroupStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          IAMGroupObservation `json:"atProvider"`
}

// +kubebuilder:object:root=true

// An IAMGroup is a managed resource that represents an AWS IAM IAMGroup.
// An IAMUser is a managed resource that represents an AWS IAM IAMUser.
// +kubebuilder:printcolumn:name="ARN",type="string",JSONPath=".status.atProvider.arn"
// +kubebuilder:printcolumn:name="ID",type="string",JSONPath=".status.atProvider.groupId"
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,aws}
type IAMGroup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec IAMGroupSpec `json:"spec"`

	Status IAMGroupStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// IAMGroupList contains a list of IAM Groups
type IAMGroupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []IAMGroup `json:"items"`
}
