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
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// VPCCIDRBlockParameters define the desired state of an VPC CIDR Block
type VPCCIDRBlockParameters struct {
	// Region is the region you'd like your VPC CIDR to be created in.
	Region string `json:"region"`

	// Requests an Amazon-provided IPv6 CIDR block with a /56 prefix length for
	// the VPC. You cannot specify the range of IPv6 addresses, or the size of the
	// CIDR block.
	// +immutable
	// +optional
	AmazonProvidedIPv6CIDRBlock *bool `json:"amazonProvidedIpv6CidrBlock,omitempty"`

	// An IPv4 CIDR block to associate with the VPC.
	// +immutable
	// +optional
	CIDRBlock *string `json:"cidrBlock,omitempty"`

	// An IPv6 CIDR block from the IPv6 address pool. You must also specify Ipv6Pool
	// in the request.
	//
	// To let Amazon choose the IPv6 CIDR block for you, omit this parameter.
	// +immutable
	// +optional
	IPv6CIDRBlock *string `json:"ipv6CdirBlock,omitempty"`

	// The name of the location from which we advertise the IPV6 CIDR block. Use
	// this parameter to limit the CiDR block to this location.
	//
	// You must set AmazonProvidedIpv6CIDRBlock to true to use this parameter.
	//
	// You can have one IPv6 CIDR block association per network border group.
	// +immutable
	// +optional
	IPv6CIDRBlockNetworkBorderGroup *string `json:"ipv6CidrBlockNetworkBorderGroup,omitempty"`

	// The ID of an IPv6 address pool from which to allocate the IPv6 CIDR block.
	// +immutable
	// +optional
	IPv6Pool *string `json:"ipv6Pool,omitempty"`

	// VPCID is the ID of the VPC.
	// +optional
	VPCID *string `json:"vpcId,omitempty"`

	// VPCIDRef references a VPC to and retrieves its vpcId
	// +optional
	VPCIDRef *xpv1.Reference `json:"vpcIdRef,omitempty"`

	// VPCIDSelector selects a reference to a VPC to and retrieves its vpcId
	// +optional
	VPCIDSelector *xpv1.Selector `json:"vpcIdSelector,omitempty"`
}

// A VPCCIDRBlockSpec defines the desired state of a VPCCIDRBlock.
type VPCCIDRBlockSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       VPCCIDRBlockParameters `json:"forProvider"`
}

// VPCCIDRBlockObservation keeps the state for the external resource
type VPCCIDRBlockObservation struct {
	// The association ID for the CIDR block.
	AssociationID *string `json:"associationId,omitempty"`

	// The IPv4 CIDR block.
	CIDRBlock *string `json:"cidrBlock,omitempty"`

	// The IPv6 CIDR block.
	IPv6CIDRBlock *string `json:"ipv6CidrBlock,omitempty"`

	// Information about the state of the CIDR block.
	IPv6CIDRBlockState *VPCCIDRBlockState `json:"ipv6CidrrBlockState,omitempty"`

	// The ID of the IPv6 address pool from which the IPv6 CIDR block is allocated.
	IPv6Pool *string `json:"ipv6Pool,omitempty"`

	// The name of the location from which we advertise the IPV6 CIDR block.
	NetworkBorderGroup *string `json:"networkBorderGroup,omitempty"`

	// Information about the state of the CIDR block.
	CIDRBlockState *VPCCIDRBlockState `json:"cidrBlockState,omitempty"`
}

// VPCCIDRBlockState represents the state of a CIDR Block
type VPCCIDRBlockState struct {

	// The state of the CIDR block.
	State *string `json:"state,omitempty"`

	// A message about the status of the CIDR block, if applicable.
	StatusMessage *string `json:"statusMessage,omitempty"`
}

// A VPCCIDRBlockStatus represents the observed state of a ElasticIP.
type VPCCIDRBlockStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          VPCCIDRBlockObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A VPCCIDRBlock is a managed resource that represents an secondary CIDR block for a VPC
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="ID",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="CIDR",type="string",JSONPath=".status.atProvider.cidrBlock"
// +kubebuilder:printcolumn:name="IPv6CIDR",type="string",JSONPath=".status.atProvider.ipv6CIDRBlock"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,aws}
// +kubebuilder:deprecatedversion:warning="Please use v1beta1 version of this resource that has identical schema."
// Deprecated: Please use v1beta1 version of this resource.
type VPCCIDRBlock struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   VPCCIDRBlockSpec   `json:"spec"`
	Status VPCCIDRBlockStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// VPCCIDRBlockList contains a list of VPCCIDRBlocks
// +kubebuilder:deprecatedversion:warning="Please use v1beta1 version of this resource that has identical schema."
// Deprecated: Please use v1beta1 version of this resource.
type VPCCIDRBlockList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VPCCIDRBlock `json:"items"`
}
