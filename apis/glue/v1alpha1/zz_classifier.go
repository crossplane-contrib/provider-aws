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

// ClassifierParameters defines the desired state of Classifier
type ClassifierParameters struct {
	// Region is which region the Classifier will be created.
	// +kubebuilder:validation:Required
	Region                     string `json:"region"`
	CustomClassifierParameters `json:",inline"`
}

// ClassifierSpec defines the desired state of Classifier
type ClassifierSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       ClassifierParameters `json:"forProvider"`
}

// ClassifierObservation defines the observed state of Classifier
type ClassifierObservation struct {
}

// ClassifierStatus defines the observed state of Classifier.
type ClassifierStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          ClassifierObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// Classifier is the Schema for the Classifiers API
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,aws}
type Classifier struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              ClassifierSpec   `json:"spec"`
	Status            ClassifierStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ClassifierList contains a list of Classifiers
type ClassifierList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Classifier `json:"items"`
}

// Repository type metadata.
var (
	ClassifierKind             = "Classifier"
	ClassifierGroupKind        = schema.GroupKind{Group: Group, Kind: ClassifierKind}.String()
	ClassifierKindAPIVersion   = ClassifierKind + "." + GroupVersion.String()
	ClassifierGroupVersionKind = GroupVersion.WithKind(ClassifierKind)
)

func init() {
	SchemeBuilder.Register(&Classifier{}, &ClassifierList{})
}
