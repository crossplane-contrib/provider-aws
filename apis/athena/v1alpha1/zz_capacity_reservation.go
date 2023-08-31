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

// Code generated by ack-generate. DO NOT EDIT.

package v1alpha1

import (
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// CapacityReservationParameters defines the desired state of CapacityReservation
type CapacityReservationParameters struct {
	// Region is which region the CapacityReservation will be created.
	// +kubebuilder:validation:Required
	Region string `json:"region"`
	// The name of the capacity reservation to create.
	// +kubebuilder:validation:Required
	Name *string `json:"name"`
	// The tags for the capacity reservation.
	Tags []*Tag `json:"tags,omitempty"`
	// The number of requested data processing units.
	// +kubebuilder:validation:Required
	TargetDpus                          *int64 `json:"targetDpus"`
	CustomCapacityReservationParameters `json:",inline"`
}

// CapacityReservationSpec defines the desired state of CapacityReservation
type CapacityReservationSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       CapacityReservationParameters `json:"forProvider"`
}

// CapacityReservationObservation defines the observed state of CapacityReservation
type CapacityReservationObservation struct {
}

// CapacityReservationStatus defines the observed state of CapacityReservation.
type CapacityReservationStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          CapacityReservationObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// CapacityReservation is the Schema for the CapacityReservations API
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:storageversion
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,aws}
type CapacityReservation struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              CapacityReservationSpec   `json:"spec"`
	Status            CapacityReservationStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// CapacityReservationList contains a list of CapacityReservations
type CapacityReservationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CapacityReservation `json:"items"`
}

// Repository type metadata.
var (
	CapacityReservationKind             = "CapacityReservation"
	CapacityReservationGroupKind        = schema.GroupKind{Group: CRDGroup, Kind: CapacityReservationKind}.String()
	CapacityReservationKindAPIVersion   = CapacityReservationKind + "." + GroupVersion.String()
	CapacityReservationGroupVersionKind = GroupVersion.WithKind(CapacityReservationKind)
)

func init() {
	SchemeBuilder.Register(&CapacityReservation{}, &CapacityReservationList{})
}