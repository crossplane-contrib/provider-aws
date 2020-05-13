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
	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CacheSubnetGroupParameters define the desired state of an AWS ElasticCache Subnet Group.
type CacheSubnetGroupParameters struct {
	// A description for the cache subnet group.
	Description string `json:"description"`

	// A list of  Subnet IDs for the cache subnet group.
	// +optional
	SubnetIDs []string `json:"subnetIds,omitempty"`

	// SubnetIDRefs references to a Subnet to and retrieves its SubnetID
	// +optional
	SubnetIDRefs []runtimev1alpha1.Reference `json:"subnetIdRefs,omitempty"`

	// SubnetIDSelector selects a set of references that each retrieve the subnetID from the referenced Subnet
	// +optional
	SubnetIDSelector *runtimev1alpha1.Selector `json:"subnetIdSelector,omitempty"`
}

// A CacheSubnetGroupSpec defines the desired state of a CacheSubnetGroup.
type CacheSubnetGroupSpec struct {
	runtimev1alpha1.ResourceSpec `json:",inline"`
	ForProvider                  CacheSubnetGroupParameters `json:"forProvider"`
}

// CacheSubnetGroupExternalStatus keeps the state for the external resource
type CacheSubnetGroupExternalStatus struct {
	// The Amazon Virtual Private Cloud identifier (VPC ID) of the cache subnet
	// group.
	VPCID string `json:"vpcId"`
}

// A CacheSubnetGroupStatus represents the observed state of a Subnet Group.
type CacheSubnetGroupStatus struct {
	runtimev1alpha1.ResourceStatus `json:",inline"`
	AtProvider                     CacheSubnetGroupExternalStatus `json:"atProvider"`
}

// +kubebuilder:object:root=true

// A CacheSubnetGroup is a managed resource that represents an AWS Subnet Group for ElasticCache.
// +kubebuilder:printcolumn:name="VPCID",type="string",JSONPath=".status.atProvider.vpcId"
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster
type CacheSubnetGroup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CacheSubnetGroupSpec   `json:"spec"`
	Status CacheSubnetGroupStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// CacheSubnetGroupList contains a list of CacheSubnetGroup
type CacheSubnetGroupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CacheSubnetGroup `json:"items"`
}
