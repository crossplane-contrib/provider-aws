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

// DBParameterGroupParameters defines the desired state of DBParameterGroup
type DBParameterGroupParameters struct {
	// Region is which region the DBParameterGroup will be created.
	// +kubebuilder:validation:Required
	Region string `json:"region"`
	// The description for the DB parameter group.
	// +kubebuilder:validation:Required
	Description *string `json:"description"`
	// Tags to assign to the DB parameter group.
	Tags                             []*Tag `json:"tags,omitempty"`
	CustomDBParameterGroupParameters `json:",inline"`
}

// DBParameterGroupSpec defines the desired state of DBParameterGroup
type DBParameterGroupSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       DBParameterGroupParameters `json:"forProvider"`
}

// DBParameterGroupObservation defines the observed state of DBParameterGroup
type DBParameterGroupObservation struct {
	// The Amazon Resource Name (ARN) for the DB parameter group.
	DBParameterGroupARN *string `json:"dbParameterGroupARN,omitempty"`
	// The name of the DB parameter group family that this DB parameter group is
	// compatible with.
	DBParameterGroupFamily *string `json:"dbParameterGroupFamily,omitempty"`
	// The name of the DB parameter group.
	DBParameterGroupName *string `json:"dbParameterGroupName,omitempty"`
}

// DBParameterGroupStatus defines the observed state of DBParameterGroup.
type DBParameterGroupStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          DBParameterGroupObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// DBParameterGroup is the Schema for the DBParameterGroups API
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:subresource:status
// +kubebuilder:storageversion
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,aws}
type DBParameterGroup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              DBParameterGroupSpec   `json:"spec"`
	Status            DBParameterGroupStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// DBParameterGroupList contains a list of DBParameterGroups
type DBParameterGroupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DBParameterGroup `json:"items"`
}

// Repository type metadata.
var (
	DBParameterGroupKind             = "DBParameterGroup"
	DBParameterGroupGroupKind        = schema.GroupKind{Group: CRDGroup, Kind: DBParameterGroupKind}.String()
	DBParameterGroupKindAPIVersion   = DBParameterGroupKind + "." + GroupVersion.String()
	DBParameterGroupGroupVersionKind = GroupVersion.WithKind(DBParameterGroupKind)
)

func init() {
	SchemeBuilder.Register(&DBParameterGroup{}, &DBParameterGroupList{})
}
