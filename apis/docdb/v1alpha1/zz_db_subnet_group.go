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

// DBSubnetGroupParameters defines the desired state of DBSubnetGroup
type DBSubnetGroupParameters struct {
	// Region is which region the DBSubnetGroup will be created.
	// +kubebuilder:validation:Required
	Region string `json:"region"`
	// The description for the subnet group.
	// +kubebuilder:validation:Required
	DBSubnetGroupDescription *string `json:"dbSubnetGroupDescription"`
	// The tags to be assigned to the subnet group.
	Tags                          []*Tag `json:"tags,omitempty"`
	CustomDBSubnetGroupParameters `json:",inline"`
}

// DBSubnetGroupSpec defines the desired state of DBSubnetGroup
type DBSubnetGroupSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       DBSubnetGroupParameters `json:"forProvider"`
}

// DBSubnetGroupObservation defines the observed state of DBSubnetGroup
type DBSubnetGroupObservation struct {
	// The Amazon Resource Name (ARN) for the DB subnet group.
	DBSubnetGroupARN *string `json:"dbSubnetGroupARN,omitempty"`
	// The name of the subnet group.
	DBSubnetGroupName *string `json:"dbSubnetGroupName,omitempty"`
	// Provides the status of the subnet group.
	SubnetGroupStatus *string `json:"subnetGroupStatus,omitempty"`
	// Detailed information about one or more subnets within a subnet group.
	Subnets []*Subnet `json:"subnets,omitempty"`
	// Provides the virtual private cloud (VPC) ID of the subnet group.
	VPCID *string `json:"vpcID,omitempty"`

	CustomDBSubnetGroupObservation `json:",inline"`
}

// DBSubnetGroupStatus defines the observed state of DBSubnetGroup.
type DBSubnetGroupStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          DBSubnetGroupObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// DBSubnetGroup is the Schema for the DBSubnetGroups API
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:storageversion
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,aws}
type DBSubnetGroup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              DBSubnetGroupSpec   `json:"spec"`
	Status            DBSubnetGroupStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// DBSubnetGroupList contains a list of DBSubnetGroups
type DBSubnetGroupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DBSubnetGroup `json:"items"`
}

// Repository type metadata.
var (
	DBSubnetGroupKind             = "DBSubnetGroup"
	DBSubnetGroupGroupKind        = schema.GroupKind{Group: CRDGroup, Kind: DBSubnetGroupKind}.String()
	DBSubnetGroupKindAPIVersion   = DBSubnetGroupKind + "." + GroupVersion.String()
	DBSubnetGroupGroupVersionKind = GroupVersion.WithKind(DBSubnetGroupKind)
)

func init() {
	SchemeBuilder.Register(&DBSubnetGroup{}, &DBSubnetGroupList{})
}
