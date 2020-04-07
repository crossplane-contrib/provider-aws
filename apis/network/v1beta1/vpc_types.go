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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
)

// VPCParameters define the desired state of an AWS Virtual Private Cloud.
type VPCParameters struct {
	// CIDRBlock is the IPv4 network range for the VPC, in CIDR notation. For
	// example, 10.0.0.0/16.
	// +kubebuilder:validation:Required
	CIDRBlock string `json:"cidrBlock"`

	// A boolean flag to enable/disable DNS support in the VPC
	// +optional
	EnableDNSSupport *bool `json:"enableDnsSupport,omitempty"`

	// Tags are used as identification helpers between AWS resources.
	Tags []Tag `json:"tags,omitempty"`

	// +optional
	EnableDNSHostNames *bool `json:"enableDnsHostNames,omitempty"`

	// The allowed tenancy of instances launched into the VPC.
	// +optinal
	InstanceTenancy *string `json:"instanceTenancy,omitempty"`
}

// A VPCSpec defines the desired state of a VPC.
type VPCSpec struct {
	runtimev1alpha1.ResourceSpec `json:",inline"`
	ForProvider                  VPCParameters `json:"forProvider"`
}

// VPCObservation keeps the state for the external resource
type VPCObservation struct {
	// Indicates whether the VPC is the default VPC.
	IsDefault bool `json:"isDefault,omitempty"`

	// The ID of the AWS account that owns the VPC.
	OwnerID string `json:"ownerID,omitempty"`

	// VPCState is the current state of the VPC.
	// +kubebuilder:validation:Enum=pending;available
	VPCState string `json:"vpcState,omitempty"`

	// Tags represents to current ec2 tags.
	Tags []Tag `json:"tags,omitempty"`

	VPCID string `json:"vpcId,omitempty"`
}

// A VPCStatus represents the observed state of a VPC.
type VPCStatus struct {
	runtimev1alpha1.ResourceStatus `json:",inline"`
	AtProvider                     VPCObservation `json:"atProvider"`
}

// +kubebuilder:object:root=true

// A VPC is a managed resource that represents an AWS Virtual Private Cloud.
// +kubebuilder:printcolumn:name="VPCID",type="string",JSONPath=".status.atProvider.vpcId"
// +kubebuilder:printcolumn:name="CIDRBLOCK",type="string",JSONPath=".spec.forProvider.cidrBlock"
// +kubebuilder:printcolumn:name="STATE",type="string",JSONPath=".status.atProvider.vpcState"
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,aws}
type VPC struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   VPCSpec   `json:"spec"`
	Status VPCStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// VPCList contains a list of VPCs
type VPCList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VPC `json:"items"`
}
