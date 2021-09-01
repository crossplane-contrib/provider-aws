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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
)

// ResourceTagTag defines a tag for a subnet resource.
type ResourceTagTag struct {
	// Key of the Tag.
	// +kubebuilder:validation:Required
	Key string `json:"key"`

	// Value of the Tag
	// +kubebuilder:validation:Required
	Value string `json:"value"`
}

// ResourceTagParameters define the desired state of an AWS Resource ResourceTag.
type ResourceTagParameters struct {
	// Region is the region you'd like your Tag to be created in.
	// +optional
	Region *string `json:"region,omitempty"`

	// IDs of the resources that should be tagged
	// +optional
	ResourceIDs []string `json:"resourceIds,omitempty"`

	// Tags with which the resource should be tagged
	// +kubebuilder:validation:Required
	// +immutable
	Tags []ResourceTagTag `json:"tags"`
}

// A ResourceTagSpec defines the desired state of a ResourceTag.
type ResourceTagSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       ResourceTagParameters `json:"forProvider"`
}

// A ResourceTagStatus represents the observed state of a ResourceTag.
type ResourceTagStatus struct {
	xpv1.ResourceStatus `json:",inline"`
}

// +kubebuilder:object:root=true

// A ResourceTag is a managed resource that represents an AWS Resource Tag.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,aws}
type ResourceTag struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ResourceTagSpec   `json:"spec"`
	Status ResourceTagStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ResourceTagList contains a list of ResourceTags
type ResourceTagList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ResourceTag `json:"items"`
}
