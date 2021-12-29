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

package v1alpha1

import (
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
)

// CustomLaunchTemplateParameters includes the custom fields of LaunchTemplate.
type CustomLaunchTemplateParameters struct{}

// CustomLaunchTemplateVersionParameters includes the custom fields of LaunchTemplateVersion.
type CustomLaunchTemplateVersionParameters struct{
		// The ID of the Launch Template. You must specify this parameter in the request.
	// +crossplane:generate:reference:type=github.com/crossplane/provider-aws/apis/ec2/v1alpha1.LaunchTemplate
	LaunchTemplateID *string `json:"launchTemplateID,omitempty"`
	// LaunchTemplateIDRef is a reference to an API used to set
	// the LaunchTemplateID.
	// +optional
	LaunchTemplateIDRef *xpv1.Reference `json:"launchTemplateIDRef,omitempty"`
	// LaunchTemplateIDSelector selects references to API used
	// to set the LaunchTemplateID.
	// +optional
	LaunchTemplateIDSelector *xpv1.Selector `json:"launchTemplateIDSelector,omitempty"`
		// The Name of the Launch Template. You must specify this parameter in the request.
	// +crossplane:generate:reference:type=github.com/crossplane/provider-aws/apis/ec2/v1alpha1.LaunchTemplate
	LaunchTemplateName *string `json:"launchTemplateName,omitempty"`
	// LaunchTemplateNameRef is a reference to an API used to set
	// the LaunchTemplateName.
	// +optional
	LaunchTemplateNameRef *xpv1.Reference `json:"launchTemplateNameRef,omitempty"`
	// LaunchTemplateNameSelector selects references to API used
	// to set the LaunchTemplateName.
	// +optional
	LaunchTemplateNameSelector *xpv1.Selector `json:"launchTemplateNameSelector,omitempty"`
}

// CustomVolumeParameters contains the additional fields for VolumeParameters.
type CustomVolumeParameters struct {
	// The identifier of the AWS Key Management Service (AWS KMS) customer master
	// key (CMK) to use for Amazon EBS encryption. If this parameter is not specified,
	// your AWS managed CMK for EBS is used. If KmsKeyId is specified, the encrypted
	// state must be true.
	//
	// You can specify the CMK using any of the following:
	//
	//    * Key ID. For example, 1234abcd-12ab-34cd-56ef-1234567890ab.
	//
	//    * Key alias. For example, alias/ExampleAlias.
	//
	//    * Key ARN. For example, arn:aws:kms:us-east-1:012345678910:key/1234abcd-12ab-34cd-56ef-1234567890ab.
	//
	//    * Alias ARN. For example, arn:aws:kms:us-east-1:012345678910:alias/ExampleAlias.
	//
	// AWS authenticates the CMK asynchronously. Therefore, if you specify an ID,
	// alias, or ARN that is not valid, the action can appear to complete, but eventually
	// fails.
	// +optional
	// +crossplane:generate:reference:type=github.com/crossplane/provider-aws/apis/kms/v1alpha1.Key
	// +crossplane:generate:reference:refFieldName=KMSKeyIDRef
	// +crossplane:generate:reference:selectorFieldName=KMSKeyIDSelector
	KMSKeyID *string `json:"kmsKeyId,omitempty"`

	// KMSKeyIDRef is a reference to a KMS Key used to set KMSKeyID.
	// +optional
	KMSKeyIDRef *xpv1.Reference `json:"kmsKeyIdRef,omitempty"`

	// KMSKeyIDSelector selects a reference to a KMS Key used to set KMSKeyID.
	// +optional
	KMSKeyIDSelector *xpv1.Selector `json:"kmsKeyIdSelector,omitempty"`
}

// CustomVPCPeeringConnectionParameters are custom parameters for VPCPeeringConnection
type CustomVPCPeeringConnectionParameters struct {
	// The ID of the requester VPC. You must specify this parameter in the request.
	// +crossplane:generate:reference:type=github.com/crossplane/provider-aws/apis/ec2/v1beta1.VPC
	VPCID *string `json:"vpcID,omitempty"`
	// VPCIDRef is a reference to an API used to set
	// the VPCID.
	// +optional
	VPCIDRef *xpv1.Reference `json:"vpcIDRef,omitempty"`
	// VPCIDSelector selects references to API used
	// to set the VPCID.
	// +optional
	VPCIDSelector *xpv1.Selector `json:"vpcIDSelector,omitempty"`
	// The ID of the VPC with which you are creating the VPC peering connection.
	// You must specify this parameter in the request.
	PeerVPCID *string `json:"peerVPCID,omitempty"`
	// PeerVPCIDRef is a reference to an API used to set
	// the PeerVPCID.
	// +optional
	PeerVPCIDRef *xpv1.Reference `json:"peerVPCIDRef,omitempty"`
	// PeerVPCIDSelector selects references to API used
	// to set the PeerVPCID.
	// +optional
	PeerVPCIDSelector *xpv1.Selector `json:"peerVPCIDSelector,omitempty"`
	// Automatically accepts the peering connection. If this is not set, the peering connection
	// will be created, but will be in pending-acceptance state. This will only lead to an active
	// connection if both VPCs are in the same tenant.
	AcceptRequest bool `json:"acceptRequest,omitempty"`
}

// CustomTransitGatewayParameters are custom parameters for TransitGateway
type CustomTransitGatewayParameters struct {
	// Metadata tagging key value pairs
	// +optional
	Tags []Tag `json:"tags,omitempty"`
}

// CustomTransitGatewayVPCAttachmentParameters are custom parameters for TransitGatewayVPCAttachment
type CustomTransitGatewayVPCAttachmentParameters struct {
	// The ID of the VPC.
	// +optional
	// +crossplane:generate:reference:type=github.com/crossplane/provider-aws/apis/ec2/v1beta1.VPC
	VPCID *string `json:"vpcId,omitempty"`

	// VPCIDRef is a reference to an API used to set
	// the VPCID.
	// +optional
	VPCIDRef *xpv1.Reference `json:"vpcIdRef,omitempty"`

	// VPCIDSelector selects references to API used
	// to set the VPCID.
	// +optional
	VPCIDSelector *xpv1.Selector `json:"vpcIdSelector,omitempty"`

	// The IDs of one or more subnets. You can specify only one subnet per Availability
	// Zone. You must specify at least one subnet, but we recommend that you specify
	// two subnets for better availability. The transit gateway uses one IP address
	// from each specified subnet.
	// +optional
	// +crossplane:generate:reference:type=github.com/crossplane/provider-aws/apis/ec2/v1beta1.Subnet
	// +crossplane:generate:reference:refFieldName=SubnetIDRefs
	// +crossplane:generate:reference:selectorFieldName=SubnetIDSelector
	SubnetIDs []*string `json:"subnetIds,omitempty"`

	// SubnetIDRefs is a list of references to SubnetIDs used to set
	// the SubnetIDs.
	// +optional
	SubnetIDRefs []xpv1.Reference `json:"subnetIdRefs,omitempty"`

	// SubnetIDSelector selects references to SubnetIDs used
	// to set the SubnetIDs.
	// +optional
	SubnetIDSelector *xpv1.Selector `json:"subnetIdSelector,omitempty"`

	// The ID of the transit gateway.
	// +optional
	// +crossplane:generate:reference:type=TransitGateway
	TransitGatewayID *string `json:"transitGatewayId,omitempty"`

	// TransitGatewayIDRef is a reference to an API used to set
	// the TransitGatewayID.
	// +optional
	TransitGatewayIDRef *xpv1.Reference `json:"transitGatewayIdRef,omitempty"`

	// TransitGatewayIDSelector selects references to API used
	// to set the TransitGatewayID.
	// +optional
	TransitGatewayIDSelector *xpv1.Selector `json:"transitGatewayIdSelector,omitempty"`

	// Metadata tagging key value pairs
	// +optional
	Tags []Tag `json:"tags,omitempty"`
}

// CustomRouteParameters are custom parameters for Route
type CustomRouteParameters struct {
	// The ID of a transit gateway.
	// +optional
	// +crossplane:generate:reference:type=TransitGateway
	TransitGatewayID *string `json:"transitGatewayId,omitempty"`

	// TransitGatewayIDRef is a reference to an API used to set
	// the TransitGatewayID.
	// +optional
	TransitGatewayIDRef *xpv1.Reference `json:"transitGatewayIdRef,omitempty"`

	// TransitGatewayIDSelector selects references to API used
	// to set the TransitGatewayID.
	// +optional
	TransitGatewayIDSelector *xpv1.Selector `json:"transitGatewayIdSelector,omitempty"`

	// [IPv4 traffic only] The ID of a NAT gateway.
	// +optional
	// +crossplane:generate:reference:type=github.com/crossplane/provider-aws/apis/ec2/v1beta1.NATGateway
	NATGatewayID *string `json:"natGatewayId,omitempty"`

	// NATGatewayIDRef is a reference to an API used to set
	// the NATGatewayID.
	// +optional
	NATGatewayIDRef *xpv1.Reference `json:"natGatewayIdRef,omitempty"`

	// NATGatewayIDSelector selects references to API used
	// to set the NATGatewayID.
	// +optional
	NATGatewayIDSelector *xpv1.Selector `json:"natGatewayIdSelector,omitempty"`

	// The ID of a VPC peering connection.
	// +crossplane:generate:reference:type=VPCPeeringConnection
	VPCPeeringConnectionID *string `json:"vpcPeeringConnectionId,omitempty"`

	// VPCPeeringConnectionIDRef is a reference to an API used to set
	// the VPCPeeringConnectionID.
	// +optional
	VPCPeeringConnectionIDRef *xpv1.Reference `json:"vpcPeeringConnectionIdRef,omitempty"`

	// VPCPeeringConnectionIDSelector selects references to API used
	// to set the VPCPeeringConnectionID.
	// +optional
	VPCPeeringConnectionIDSelector *xpv1.Selector `json:"vpcPeeringConnectionIdSelector,omitempty"`

	// The ID of the route table for the route.
	// provider-aws currently provides both a standalone Route resource
	// and a RouteTable resource with routes defined in-line.
	// At this time you cannot use a Route Table with in-line routes
	// in conjunction with any Route resources.
	// Doing so will cause a conflict of rule settings and will overwrite rules.
	// +optional
	RouteTableID *string `json:"routeTableId,omitempty"`

	// The ID of a NAT instance in your VPC. The operation fails if you specify
	// an instance ID unless exactly one network interface is attached.
	// +optional
	// +crossplane:generate:reference:type=github.com/crossplane/provider-aws/apis/ec2/manualv1alpha1.Instance
	InstanceID *string `json:"instanceId,omitempty"`

	// InstanceIDRef is a reference to an API used to set
	// the InstanceID.
	// +optional
	InstanceIDRef *xpv1.Reference `json:"instanceIdRef,omitempty"`

	// InstanceIDSelector selects references to API used
	// to set the InstanceID.
	// +optional
	InstanceIDSelector *xpv1.Selector `json:"instanceIdSelector,omitempty"`

	// The ID of an internet gateway attached to your VPC.
	// +optional
	// +crossplane:generate:reference:type=github.com/crossplane/provider-aws/apis/ec2/v1beta1.InternetGateway
	GatewayID *string `json:"gatewayId,omitempty"`

	// GatewayIDRef is a reference to an API used to set
	// the GatewayID.
	// +optional
	GatewayIDRef *xpv1.Reference `json:"gatewayIdRef,omitempty"`

	// GatewayIDSelector selects references to API used
	// to set the GatewayID.
	// +optional
	GatewayIDSelector *xpv1.Selector `json:"gatewayIdSelector,omitempty"`
}

// CustomVPCEndpointParameters are custom parameters for VPCEndpoint
type CustomVPCEndpointParameters struct {
	// The ID of the VPC. You must specify this parameter in the request.
	// +optional
	// +crossplane:generate:reference:type=github.com/crossplane/provider-aws/apis/ec2/v1beta1.VPC
	VPCID *string `json:"vpcId,omitempty"`

	// VPCIDRef is a reference to an API used to set
	// the VPCID.
	// +optional
	VPCIDRef *xpv1.Reference `json:"vpcIdRef,omitempty"`

	// VPCIDSelector selects references to API used
	// to set the VPCID.
	// +optional
	VPCIDSelector *xpv1.Selector `json:"vpcIdSelector,omitempty"`

	// (Interface endpoint) The ID of one or more security groups to associate with
	// the endpoint network interface.
	// +crossplane:generate:reference:type=github.com/crossplane/provider-aws/apis/ec2/v1beta1.SecurityGroup
	// +crossplane:generate:reference:refFieldName=SecurityGroupIDRefs
	// +crossplane:generate:reference:selectorFieldName=SecurityGroupIDSelector
	SecurityGroupIDs []*string `json:"securityGroupIds,omitempty"`

	// SecurityGroupIDRefs is a list of references to SecurityGroups used to set
	// the SecurityGroupIDs.
	// +optional
	SecurityGroupIDRefs []xpv1.Reference `json:"securityGroupIdRefs,omitempty"`

	// SecurityGroupIDsSelector selects references to SecurityGroupID used
	// to set the SecurityGroupIDs.
	// +optional
	SecurityGroupIDSelector *xpv1.Selector `json:"securityGroupIdSelector,omitempty"`

	// (Interface and Gateway Load Balancer endpoints) The ID of one or more subnets
	// in which to create an endpoint network interface. For a Gateway Load Balancer
	// endpoint, you can specify one subnet only.
	// +crossplane:generate:reference:type=github.com/crossplane/provider-aws/apis/ec2/v1beta1.Subnet
	// +crossplane:generate:reference:refFieldName=SubnetIDRefs
	// +crossplane:generate:reference:selectorFieldName=SubnetIDSelector
	SubnetIDs []*string `json:"subnetIds,omitempty"`

	// SubnetIDRefs is a list of references to Subnets used to set
	// the SubnetIDs.
	// +optional
	SubnetIDRefs []xpv1.Reference `json:"subnetIdRefs,omitempty"`

	// SubnetIDsSelector selects references to Subnets used
	// to set the SubnetIDs.
	// +optional
	SubnetIDSelector *xpv1.Selector `json:"subnetIdSelector,omitempty"`

	// (Gateway endpoint) One or more route table IDs.
	// +crossplane:generate:reference:type=github.com/crossplane/provider-aws/apis/ec2/v1beta1.RouteTable
	// +crossplane:generate:reference:refFieldName=RouteTableIDRefs
	// +crossplane:generate:reference:selectorFieldName=RouteTableIDSelector
	RouteTableIDs []*string `json:"routeTableIds,omitempty"`

	// RouteTableIDRefs is a list of references to RouteTables used to set
	// the RouteTableIDs.
	// +optional
	RouteTableIDRefs []xpv1.Reference `json:"routeTableIdRefs,omitempty"`

	// RouteTableIDsSelector selects references to RouteTables used
	// to set the RouteTableIDs.
	// +optional
	RouteTableIDSelector *xpv1.Selector `json:"routeTableIdSelector,omitempty"`
}
