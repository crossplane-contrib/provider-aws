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

// GroupParameters define the desired state of an AWS IAM Group.
type GroupParameters struct {
	// The path for the group name.
	// +optional
	Path *string `json:"path,omitempty"`
}

// A GroupSpec defines the desired state of an IAM Group.
type GroupSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       GroupParameters `json:"forProvider,omitempty"`
}

// GroupObservation keeps the state for the external resource
type GroupObservation struct {
	// The Amazon Resource Name (ARN) that identifies the group.
	ARN string `json:"arn,omitempty"`

	// The stable and unique string identifying the group.
	GroupID string `json:"groupId,omitempty"`
}

// A GroupStatus represents the observed state of an IAM Group.
type GroupStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          GroupObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A Group is a managed resource that represents an AWS IAM Group.
// A User is a managed resource that represents an AWS IAM User.
// +kubebuilder:printcolumn:name="ARN",type="string",JSONPath=".status.atProvider.arn"
// +kubebuilder:printcolumn:name="ID",type="string",JSONPath=".status.atProvider.groupId"
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,aws}
type Group struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec GroupSpec `json:"spec"`

	Status GroupStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// GroupList contains a list of IAM Groups
type GroupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Group `json:"items"`
}
