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

// VpcCidrBlockState represents the state of a CIDR Block
type VpcCidrBlockState struct {

	// The state of the CIDR block.
	State string `json:"state"`

	// A message about the status of the CIDR block, if applicable.
	StatusMessage *string `json:"statusmessage"`
}

// VpcCidrBlockAssociation represents the association of IPv4 CIDR blocks with the VPC.
type VpcCidrBlockAssociation struct {

	// The association ID for the IPv4 CIDR block.
	AssociationID *string `json:"associationId,omitempty"`

	// The IPv4 CIDR block.
	CIDRBlock *string `json:"cidrBlock,omitempty"`

	// Information about the state of the CIDR block.
	CIDRBlockState *VpcCidrBlockState `json:"cidrBlockState,omitempty"`
}

// VpcIpv6CidrBlockAssociation represents the association of IPv6 CIDR blocks with the VPC.
type VpcIpv6CidrBlockAssociation struct {

	// The association ID for the IPv6 CIDR block.
	AssociationID *string `json:"associationId,omitempty"`

	// The IPv6 CIDR block.
	IPv6CIDRBlock *string `json:"ipv6CidrBlock,omitempty"`

	// Information about the state of the CIDR block.
	IPv6CIDRBlockState *VpcCidrBlockState `json:"ipv6CidrBlockState,omitempty"`

	// The ID of the IPv6 address pool from which the IPv6 CIDR block is allocated.
	Ipv6Pool *string `json:"ipv6Pool,omitempty"`

	// The name of the location from which we advertise the IPV6 CIDR block.
	NetworkBorderGroup *string `json:"networkBorderGroup,omitempty"`
}

// VPCParameters define the desired state of an AWS Virtual Private Cloud.
type VPCParameters struct {
	// CIDRBlock is the IPv4 network range for the VPC, in CIDR notation. For
	// example, 10.0.0.0/16.
	// +kubebuilder:validation:Required
	// +immutable
	CIDRBlock string `json:"cidrBlock"`

	// A boolean flag to enable/disable DNS support in the VPC
	// +optional
	EnableDNSSupport *bool `json:"enableDnsSupport,omitempty"`

	// Tags are used as identification helpers between AWS resources.
	// +optional
	Tags []Tag `json:"tags,omitempty"`

	// Indicates whether the instances launched in the VPC get DNS hostnames.
	// +optional
	EnableDNSHostNames *bool `json:"enableDnsHostNames,omitempty"`

	// The allowed tenancy of instances launched into the VPC.
	// +optional
	InstanceTenancy *string `json:"instanceTenancy,omitempty"`
}

// A VPCSpec defines the desired state of a VPC.
type VPCSpec struct {
	runtimev1alpha1.ResourceSpec `json:",inline"`
	ForProvider                  VPCParameters `json:"forProvider"`
}

// VPCObservation keeps the state for the external resource
type VPCObservation struct {
	// Information about the IPv4 CIDR blocks associated with the VPC.
	CidrBlockAssociationSet []VpcCidrBlockAssociation `json:"cidrBlockAssociationSet,omitempty"`

	// The ID of the set of DHCP options you've associated with the VPC.
	DHCPOptionsID *string `json:"dhcpOptionsId,omitempty"`

	// Information about the IPv6 CIDR blocks associated with the VPC.
	IPv6CIDRBlockAssociationSet []VpcIpv6CidrBlockAssociation `json:"ipv6CidrBlockAssociationSet,omitempty"`

	// Indicates whether the VPC is the default VPC.
	IsDefault bool `json:"isDefault,omitempty"`

	// The ID of the AWS account that owns the VPC.
	OwnerID string `json:"ownerID,omitempty"`

	// VPCState is the current state of the VPC.
	VPCState string `json:"vpcState,omitempty"`

	// Tags represents to current ec2 tags.
	Tags []Tag `json:"tags,omitempty"`

	// The ID of the VPC
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
