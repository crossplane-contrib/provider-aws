/*
Copyright 2020 The Crossplane Authors.

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

// PresignedNotebookInstanceURLParameters defines the desired state of PresignedNotebookInstanceURL
type PresignedNotebookInstanceURLParameters struct {
	// Region is which region the PresignedNotebookInstanceURL will be created.
	// +kubebuilder:validation:Required
	Region string `json:"region"`

	// The name of the notebook instance.
	// +kubebuilder:validation:Required
	NotebookInstanceName *string `json:"notebookInstanceName"`

	// The duration of the session, in seconds. The default is 12 hours.
	SessionExpirationDurationInSeconds *int64 `json:"sessionExpirationDurationInSeconds,omitempty"`
}

// PresignedNotebookInstanceURLSpec defines the desired state of PresignedNotebookInstanceURL
type PresignedNotebookInstanceURLSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       PresignedNotebookInstanceURLParameters `json:"forProvider"`
}

// PresignedNotebookInstanceURLObservation defines the observed state of PresignedNotebookInstanceURL
type PresignedNotebookInstanceURLObservation struct {
	// A JSON object that contains the URL string.
	AuthorizedURL *string `json:"authorizedURL,omitempty"`
}

// PresignedNotebookInstanceURLStatus defines the observed state of PresignedNotebookInstanceURL.
type PresignedNotebookInstanceURLStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          PresignedNotebookInstanceURLObservation `json:"atProvider"`
}

// +kubebuilder:object:root=true

// PresignedNotebookInstanceURL is the Schema for the PresignedNotebookInstanceURLS API
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,aws}
type PresignedNotebookInstanceURL struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              PresignedNotebookInstanceURLSpec   `json:"spec,omitempty"`
	Status            PresignedNotebookInstanceURLStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// PresignedNotebookInstanceURLList contains a list of PresignedNotebookInstanceURLS
type PresignedNotebookInstanceURLList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PresignedNotebookInstanceURL `json:"items"`
}

// Repository type metadata.
var (
	PresignedNotebookInstanceURLKind             = "PresignedNotebookInstanceURL"
	PresignedNotebookInstanceURLGroupKind        = schema.GroupKind{Group: Group, Kind: PresignedNotebookInstanceURLKind}.String()
	PresignedNotebookInstanceURLKindAPIVersion   = PresignedNotebookInstanceURLKind + "." + GroupVersion.String()
	PresignedNotebookInstanceURLGroupVersionKind = GroupVersion.WithKind(PresignedNotebookInstanceURLKind)
)

func init() {
	SchemeBuilder.Register(&PresignedNotebookInstanceURL{}, &PresignedNotebookInstanceURLList{})
}
