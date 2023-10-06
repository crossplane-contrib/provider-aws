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

package v1beta1

import (
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ToDo (haarchri): changed Route to RouteBeta otherwise we got error "CRD for Route.ec2.aws.crossplane.io has no storage version"

// RouteBeta describes a route in a route table.
// provider-aws currently provides both a standalone Route resource
// and a RouteTable resource with routes defined in-line.
// At this time you cannot use a Route Table with in-line routes
// in conjunction with any Route resources.
// Doing so will cause a conflict of rule settings and will overwrite rules.
type RouteBeta struct {
	// The IPv4 CIDR address block used for the destination match. Routing
	// decisions are based on the most specific match.
	// +optional
	DestinationCIDRBlock *string `json:"destinationCidrBlock,omitempty"`

	// The IPv6 CIDR address block used for the destination match. Routing
	// decisions are based on the most specific match.
	// +optional
	DestinationIPV6CIDRBlock *string `json:"destinationIpv6CidrBlock,omitempty"`

	// [IPv6 traffic only] The ID of an egress-only internet gateway.
	EgressOnlyInternetGatewayID *string `json:"egressOnlyInternetGatewayId,omitempty"`

	// The ID of an internet gateway or virtual private gateway attached to your
	// VPC.
	// +optional
	// +crossplane:generate:reference:type=InternetGateway
	GatewayID *string `json:"gatewayId,omitempty"`

	// A referencer to retrieve the ID of a gateway
	GatewayIDRef *xpv1.Reference `json:"gatewayIdRef,omitempty"`

	// A selector to select a referencer to retrieve the ID of a gateway
	GatewayIDSelector *xpv1.Selector `json:"gatewayIdSelector,omitempty"`

	// The ID of a NAT instance in your VPC. The operation fails if you specify
	// an instance ID unless exactly one network interface is attached.
	InstanceID *string `json:"instanceId,omitempty"`

	// The ID of the local gateway.
	LocalGatewayID *string `json:"localGatewayId,omitempty"`

	// [IPv4 traffic only] The ID of a NAT gateway.
	// +optional
	// +crossplane:generate:reference:type=NATGateway
	NatGatewayID *string `json:"natGatewayId,omitempty"`

	// A referencer to retrieve the ID of a NAT gateway
	NatGatewayIDRef *xpv1.Reference `json:"natGatewayIdRef,omitempty"`

	// A selector to select a referencer to retrieve the ID of a NAT gateway
	NatGatewayIDSelector *xpv1.Selector `json:"natGatewayIdSelector,omitempty"`

	// The ID of a network interface.
	NetworkInterfaceID *string `json:"networkInterfaceId,omitempty"`

	// The ID of a transit gateway.
	TransitGatewayID *string `json:"transitGatewayId,omitempty"`

	// The ID of a VPC peering connection.
	VpcPeeringConnectionID *string `json:"vpcPeeringConnectionId,omitempty"`
}

// ClearRefSelectors nils out ref and selectors
func (r *RouteBeta) ClearRefSelectors() {
	r.GatewayIDRef = nil
	r.GatewayIDSelector = nil
	r.NatGatewayIDSelector = nil
	r.NatGatewayIDRef = nil
}

// RouteState describes a route state in the route table.
type RouteState struct {
	// The state of the route. The blackhole state indicates that the route's
	// target isn't available (for example, the specified gateway isn't attached
	// to the VPC, or the specified NAT instance has been terminated).
	State string `json:"state,omitempty"`

	// The IPv4 CIDR address block used for the destination match. Routing
	// decisions are based on the most specific match.
	DestinationCIDRBlock string `json:"destinationCidrBlock,omitempty"`

	// The IPv6 CIDR address block used for the destination match. Routing
	// decisions are based on the most specific match.
	DestinationIPV6CIDRBlock string `json:"destinationIpv6CidrBlock,omitempty"`

	// The ID of an internet gateway or virtual private gateway attached to your
	// VPC.
	GatewayID string `json:"gatewayId,omitempty"`

	// The ID of a NAT instance in your VPC. The operation fails if you specify
	// an instance ID unless exactly one network interface is attached.
	InstanceID string `json:"instanceId,omitempty"`

	// The ID of the local gateway.
	LocalGatewayID string `json:"localGatewayId,omitempty"`

	// [IPv4 traffic only] The ID of a NAT gateway.
	NatGatewayID string `json:"natGatewayId,omitempty"`

	// The ID of a network interface.
	NetworkInterfaceID string `json:"networkInterfaceId,omitempty"`

	// The ID of a transit gateway.
	TransitGatewayID string `json:"transitGatewayId,omitempty"`

	// The ID of a VPC peering connection.
	VpcPeeringConnectionID string `json:"vpcPeeringConnectionId,omitempty"`
}

// Association describes an association between a route table and a subnet.
type Association struct {
	// The ID of the subnet. A subnet ID is not returned for an implicit
	// association.
	// +optional
	// +crossplane:generate:reference:type=Subnet
	SubnetID *string `json:"subnetId,omitempty"`

	// A referencer to retrieve the ID of a subnet
	// +optional
	SubnetIDRef *xpv1.Reference `json:"subnetIdRef,omitempty"`

	// A selector to select a referencer to retrieve the ID of a subnet
	// +optional
	SubnetIDSelector *xpv1.Selector `json:"subnetIdSelector,omitempty"`
}

// ClearRefSelectors nils out ref and selectors
func (a *Association) ClearRefSelectors() {
	a.SubnetIDRef = nil
	a.SubnetIDSelector = nil
}

// AssociationState describes an association state in the route table.
type AssociationState struct {
	// Indicates whether this is the main route table.
	Main *bool `json:"main"`

	// The ID of the association between a route table and a subnet.
	AssociationID string `json:"associationId,omitempty"`

	// The state of the association.
	State string `json:"state,omitempty"`

	// The ID of the subnet. A subnet ID is not returned for an implicit
	// association.
	SubnetID string `json:"subnetId,omitempty"`
}

// RouteTableParameters define the desired state of an AWS VPC Route Table.
type RouteTableParameters struct {
	// Region is the region you'd like your VPC to be created in.
	Region string `json:"region"`

	// Indicates whether we reconcile inline routes
	// +optional
	IgnoreRoutes *bool `json:"ignoreRoutes,omitempty"`

	// The associations between the route table and one or more subnets.
	Associations []Association `json:"associations"`

	// inline routes in the route table
	// Deprecated: Routes inline exists for historical compatibility
	// and should not be used. Please use separate route resource.
	// +optional
	Routes []RouteBeta `json:"routes,omitempty"`

	// Tags represents to current ec2 tags.
	// +optional
	Tags []Tag `json:"tags,omitempty"`

	// VPCID is the ID of the VPC.
	// +optional
	// +immutable
	// +crossplane:generate:reference:type=VPC
	VPCID *string `json:"vpcId,omitempty"`

	// VPCIDRef references a VPC to retrieve its vpcId
	// +optional
	// +immutable
	VPCIDRef *xpv1.Reference `json:"vpcIdRef,omitempty"`

	// VPCIDSelector selects a reference to a VPC to retrieve its vpcId
	// +optional
	VPCIDSelector *xpv1.Selector `json:"vpcIdSelector,omitempty"`
}

// A RouteTableSpec defines the desired state of a RouteTable.
type RouteTableSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       RouteTableParameters `json:"forProvider"`
}

// RouteTableObservation keeps the state for the external resource
type RouteTableObservation struct {
	// The ID of the AWS account that owns the route table.
	OwnerID string `json:"ownerId,omitempty"`

	// RouteTableID is the ID of the RouteTable.
	RouteTableID string `json:"routeTableId,omitempty"`

	// The actual routes created for the route table.
	Routes []RouteState `json:"routes,omitempty"`

	// The actual associations created for the route table.
	Associations []AssociationState `json:"associations,omitempty"`
}

// A RouteTableStatus represents the observed state of a RouteTable.
type RouteTableStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          RouteTableObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A RouteTable is a managed resource that represents an AWS VPC Route Table.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="ID",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="VPC",type="string",JSONPath=".spec.forProvider.vpcId"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,aws}
// +kubebuilder:storageversion
type RouteTable struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RouteTableSpec   `json:"spec"`
	Status RouteTableStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// RouteTableList contains a list of RouteTables
type RouteTableList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RouteTable `json:"items"`
}
