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

package v1alpha3

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
)

// Route describes a route in a route table.
type Route struct {
	// The IPv4 CIDR address block used for the destination match. Routing
	// decisions are based on the most specific match.
	DestinationCIDRBlock string `json:"destinationCidrBlock"`

	// The ID of an internet gateway or virtual private gateway attached to your
	// VPC.
	GatewayID string `json:"gatewayId,omitempty"`

	// A referencer to retrieve the ID of a gateway
	GatewayIDRef *runtimev1alpha1.Reference `json:"gatewayIdRef,omitempty"`

	// A selector to select a referencer to retrieve the ID of a gateway
	GatewayIDSelector *runtimev1alpha1.Selector `json:"gatewayIdSelector,omitempty"`
}

// RouteState describes a route state in the route table.
type RouteState struct {
	// The state of the route. The blackhole state indicates that the route's
	// target isn't available (for example, the specified gateway isn't attached
	// to the VPC, or the specified NAT instance has been terminated).
	RouteState string `json:"routeState,omitempty"`

	Route `json:",inline"`
}

// Association describes an association between a route table and a subnet.
type Association struct {
	// The ID of the subnet. A subnet ID is not returned for an implicit
	// association.
	SubnetID string `json:"subnetId,omitempty"`

	// A referencer to retrieve the ID of a subnet
	SubnetIDRef *runtimev1alpha1.Reference `json:"subnetIdRef,omitempty"`

	// A selector to select a referencer to retrieve the ID of a subnet
	SubnetIDSelector *runtimev1alpha1.Selector `json:"subnetIdSelector,omitempty"`
}

// AssociationState describes an association state in the route table.
type AssociationState struct {
	// Indicates whether this is the main route table.
	Main bool `json:"main"`

	// The ID of the association between a route table and a subnet.
	AssociationID string `json:"associationId"`

	Association `json:",inline"`
}

// RouteTableParameters define the desired state of an AWS VPC Route Table.
type RouteTableParameters struct {
	// VPCID is the ID of the VPC.
	VPCID string `json:"vpcId,omitempty"`

	// VPCIDRef references a VPC to retrieve its vpcId
	VPCIDRef *runtimev1alpha1.Reference `json:"vpcIdRef,omitempty"`

	// VPCIDSelector selects a reference to a VPC to retrieve its vpcId
	VPCIDSelector *runtimev1alpha1.Selector `json:"vpcIdSelector,omitempty"`

	// the routes in the route table
	Routes []Route `json:"routes,omitempty"`

	// The associations between the route table and one or more subnets.
	Associations []Association `json:"associations,omitempty"`
}

// A RouteTableSpec defines the desired state of a RouteTable.
type RouteTableSpec struct {
	runtimev1alpha1.ResourceSpec `json:",inline"`
	RouteTableParameters         `json:",inline"`
}

// RouteTableExternalStatus keeps the state for the external resource
type RouteTableExternalStatus struct {
	// The actual routes created for the route table.
	Routes []RouteState `json:"routes,omitempty"`

	// The actual associations created for the route table.
	Associations []AssociationState `json:"associations,omitempty"`
}

// A RouteTableStatus represents the observed state of a RouteTable.
type RouteTableStatus struct {
	runtimev1alpha1.ResourceStatus `json:",inline"`
	RouteTableExternalStatus       `json:",inline"`
}

// +kubebuilder:object:root=true

// A RouteTable is a managed resource that represents an AWS VPC Route Table.
// +kubebuilder:printcolumn:name="TABLEID",type="string",JSONPath=".status.routeTableId"
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster
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

// UpdateExternalStatus updates the external status object, given the observation
func (t *RouteTable) UpdateExternalStatus(observation ec2.RouteTable) {
	st := RouteTableExternalStatus{
		Routes:       []RouteState{},
		Associations: []AssociationState{},
	}

	st.Routes = make([]RouteState, len(observation.Routes))
	for i, rt := range observation.Routes {
		st.Routes[i] = RouteState{
			RouteState: string(rt.State),
			Route: Route{
				DestinationCIDRBlock: aws.StringValue(rt.DestinationCidrBlock),
				GatewayID:            aws.StringValue(rt.GatewayId),
			},
		}
	}

	st.Associations = make([]AssociationState, len(observation.Associations))
	for i, asc := range observation.Associations {
		st.Associations[i] = AssociationState{
			Main:          aws.BoolValue(asc.Main),
			AssociationID: aws.StringValue(asc.RouteTableAssociationId),
			Association: Association{
				SubnetID: aws.StringValue(asc.SubnetId),
			},
		}
	}

	t.Status.RouteTableExternalStatus = st
}
