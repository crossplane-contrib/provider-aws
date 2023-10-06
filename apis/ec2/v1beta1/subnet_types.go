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
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// SubnetParameters define the desired state of an AWS VPC Subnet.
type SubnetParameters struct {

	// TODO(muvaf): Region is a required field but in order to keep backward compatibility
	// with old Provider type and not bear the cost of bumping to v1beta2, we're
	// keeping it optional for now. Reconsider before v1beta2 or v1.

	// Region is the region you'd like your Subnet to be created in.
	// +optional
	Region *string `json:"region,omitempty"`

	// CIDRBlock is the IPv4 network range for the Subnet, in CIDR notation. For example, 10.0.0.0/18.
	// +immutable
	CIDRBlock string `json:"cidrBlock"`

	// The Availability Zone for the subnet.
	// Default: AWS selects one for you. If you create more than one subnet in your
	// VPC, we may not necessarily select a different zone for each subnet.
	// +optional
	// +immutable
	AvailabilityZone *string `json:"availabilityZone,omitempty"`

	// The AZ ID or the Local Zone ID of the subnet.
	// +optional
	// +immutable
	AvailabilityZoneID *string `json:"availabilityZoneId,omitempty"`

	// Indicates whether a network interface created in this subnet (including a
	// network interface created by RunInstances) receives an IPv6 address.
	// +optional
	AssignIPv6AddressOnCreation *bool `json:"assignIpv6AddressOnCreation,omitempty"`

	// The IPv6 network range for the subnet, in CIDR notation. The subnet size
	// must use a /64 prefix length.
	// +optional
	// +immutable
	IPv6CIDRBlock *string `json:"ipv6CIDRBlock,omitempty"`

	// Indicates whether instances launched in this subnet receive a public IPv4
	// address.
	// +optional
	MapPublicIPOnLaunch *bool `json:"mapPublicIPOnLaunch,omitempty"`

	// Tags represents to current ec2 tags.
	// +optional
	Tags []Tag `json:"tags,omitempty"`

	// VPCID is the ID of the VPC.
	// +optional
	// +immutable
	// +crossplane:generate:reference:type=VPC
	VPCID *string `json:"vpcId,omitempty"`

	// VPCIDRef reference a VPC to retrieve its vpcId
	// +optional
	// +immutable
	VPCIDRef *xpv1.Reference `json:"vpcIdRef,omitempty"`

	// VPCIDSelector selects reference to a VPC to retrieve its vpcId
	// +optional
	VPCIDSelector *xpv1.Selector `json:"vpcIdSelector,omitempty"`
}

// A SubnetSpec defines the desired state of a Subnet.
type SubnetSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       SubnetParameters `json:"forProvider"`
}

// SubnetObservation keeps the state for the external resource
type SubnetObservation struct {
	// The number of unused private IPv4 addresses in the subnet.
	AvailableIPAddressCount int32 `json:"availableIpAddressCount,omitempty"`

	// Indicates whether this is the default subnet for the Availability Zone.
	DefaultForAZ bool `json:"defaultForAz,omitempty"`

	// SubnetState is the current state of the Subnet.
	// +kubebuilder:validation:Enum=pending;available
	SubnetState string `json:"subnetState,omitempty"`

	// SubnetID is the ID of the Subnet.
	SubnetID string `json:"subnetId,omitempty"`
}

// A SubnetStatus represents the observed state of a Subnet.
type SubnetStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          SubnetObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A Subnet is a managed resource that represents an AWS VPC Subnet.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="ID",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="VPC",type="string",JSONPath=".spec.forProvider.vpcId"
// +kubebuilder:printcolumn:name="CIDR",type="string",JSONPath=".spec.forProvider.cidrBlock"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,aws}
// +kubebuilder:storageversion
type Subnet struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SubnetSpec   `json:"spec"`
	Status SubnetStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// SubnetList contains a list of Subnets
type SubnetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Subnet `json:"items"`
}
