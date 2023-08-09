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
type CustomLaunchTemplateParameters struct {
	// Metadata tagging key value pairs
	// +optional
	Tags []Tag `json:"tags,omitempty"`
}

// CustomVPCEndpointServiceConfigurationParameters contains the additional fields
// for VPCEndpointServiceConfigurationParameter.
type CustomVPCEndpointServiceConfigurationParameters struct {
	// Metadata tagging key value pairs
	// +optional
	Tags []Tag `json:"tags,omitempty"`

	// The Amazon Resource Names (ARNs) of one or more Gateway Load Balancers.
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-aws/apis/elbv2/v1alpha1.LoadBalancer
	// +crossplane:generate:reference:refFieldName=GatewayLoadBalancerARNRefs
	// +crossplane:generate:reference:selectorFieldName=GatewayLoadBalancerARNSelector
	GatewayLoadBalancerARNs []*string `json:"gatewayLoadBalancerARNs,omitempty"`

	// GatewayLoadBalancerARNRefs is a list of references to GatewayLoadBalancerARNs used to set
	// the GatewayLoadBalancerARNs.
	// +optional
	GatewayLoadBalancerARNRefs []xpv1.Reference `json:"gatewayLoadBalancerARNRefs,omitempty"`

	// GatewayLoadBalancerARNSelector selects references to GatewayLoadBalancerARNs used
	// to set the GatewayLoadBalancerARNs.
	// +optional
	GatewayLoadBalancerARNSelector *xpv1.Selector `json:"gatewayLoadBalancerARNSelector,omitempty"`

	// The Amazon Resource Names (ARNs) of one or more Network Load Balancers for
	// your service.
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-aws/apis/elbv2/v1alpha1.LoadBalancer
	// +crossplane:generate:reference:refFieldName=NetworkLoadBalancerARNRefs
	// +crossplane:generate:reference:selectorFieldName=NetworkLoadBalancerARNSelector
	NetworkLoadBalancerARNs []*string `json:"networkLoadBalancerARNs,omitempty"`

	// NetworkLoadBalancerARNRefs is a list of references to NetworkLoadBalancerARNs used to set
	// the NetworkLoadBalancerARNs.
	// +optional
	NetworkLoadBalancerARNRefs []xpv1.Reference `json:"networkLoadBalancerARNRefs,omitempty"`

	// NetworkLoadBalancerARNSelector selects references to NetworkLoadBalancerARNs used
	// to set the NetworkLoadBalancerARNs.
	// +optional
	NetworkLoadBalancerARNSelector *xpv1.Selector `json:"networkLoadBalancerARNSelector,omitempty"`
}

// CustomLaunchTemplateVersionParameters includes the custom fields of LaunchTemplateVersion.
type CustomLaunchTemplateVersionParameters struct {
	// The ID of the Launch Template. You must specify this parameter in the request.
	// +crossplane:generate:reference:type=LaunchTemplate
	LaunchTemplateID *string `json:"launchTemplateId,omitempty"`
	// LaunchTemplateIDRef is a reference to an API used to set
	// the LaunchTemplateID.
	// +optional
	LaunchTemplateIDRef *xpv1.Reference `json:"launchTemplateIdRef,omitempty"`
	// LaunchTemplateIDSelector selects references to API used
	// to set the LaunchTemplateID.
	// +optional
	LaunchTemplateIDSelector *xpv1.Selector `json:"launchTemplateIdSelector,omitempty"`
	// The Name of the Launch Template. You must specify this parameter in the request.
	// +crossplane:generate:reference:type=LaunchTemplate
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
	// Metadata tagging key value pairs
	// +optional
	Tags map[string]string `json:"tags,omitempty"`

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
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-aws/apis/kms/v1alpha1.Key
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
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-aws/apis/ec2/v1beta1.VPC
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
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-aws/apis/ec2/v1beta1.VPC
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

	// Metadata tagging key value pairs
	// +optional
	Tags []Tag `json:"tags,omitempty"`

	// RequesterPeeringOptions describes the Requester VPC peering connection options.
	// +optional
	RequesterPeeringOptions *VPCPeeringConnectionOptionsDescription `json:"requesterPeeringOptions,omitempty"`
	// AccepterRequesterPeeringOptions describes the Accepter VPC peering connection options.
	// +optional
	AccepterPeeringOptions *VPCPeeringConnectionOptionsDescription `json:"accepterPeeringOptions,omitempty"`
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
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-aws/apis/ec2/v1beta1.VPC
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
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-aws/apis/ec2/v1beta1.Subnet
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
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-aws/apis/ec2/v1beta1.NATGateway
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
	// At this time you cannot use a RouteTable with in-line routes
	// in conjunction with any Route resources.
	// Set the flag ignoreRoutes=true in RouteTable to use Route resource
	// +optional
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-aws/apis/ec2/v1beta1.RouteTable
	RouteTableID *string `json:"routeTableId,omitempty"`

	// RouteTableIDRef is a reference to an API used to set
	// the RouteTableID.
	// +optional
	RouteTableIDRef *xpv1.Reference `json:"routeTableIdRef,omitempty"`

	// RouteTableIDSelector selects references to API used
	// to set the RouteTableID.
	// +optional
	RouteTableIDSelector *xpv1.Selector `json:"routeTableIdSelector,omitempty"`

	// The ID of a NAT instance in your VPC. The operation fails if you specify
	// an instance ID unless exactly one network interface is attached.
	// +optional
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-aws/apis/ec2/manualv1alpha1.Instance
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
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-aws/apis/ec2/v1beta1.InternetGateway
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
	// Metadata tagging key value pairs
	// +optional
	Tags map[string]string `json:"tags,omitempty"`

	// The ID of the VPC. You must specify this parameter in the request.
	// +optional
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-aws/apis/ec2/v1beta1.VPC
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
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-aws/apis/ec2/v1beta1.SecurityGroup
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
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-aws/apis/ec2/v1beta1.Subnet
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
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-aws/apis/ec2/v1beta1.RouteTable
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

// CustomTransitGatewayRouteParameters are custom parameters for TransitGatewayRouteParameters
type CustomTransitGatewayRouteParameters struct {
	// The ID of the attachment.
	// +optional
	// +crossplane:generate:reference:type=TransitGatewayVPCAttachment
	TransitGatewayAttachmentID *string `json:"transitGatewayAttachmentId,omitempty"`

	// TransitGatewayAttachmentIDRef is a reference to an API used to set
	// the TransitGatewayAttachmentID.
	// +optional
	TransitGatewayAttachmentIDRef *xpv1.Reference `json:"transitGatewayAttachmentIdRef,omitempty"`

	// TransitGatewayAttachmentIDSelector selects references to API used
	// to set the TransitGatewayAttachmentID.
	// +optional
	TransitGatewayAttachmentIDSelector *xpv1.Selector `json:"transitGatewayAttachmentIdSelector,omitempty"`

	// The ID of the transit gateway route table.
	// +optional
	// +crossplane:generate:reference:type=TransitGatewayRouteTable
	TransitGatewayRouteTableID *string `json:"transitGatewayRouteTableId,omitempty"`

	// TransitGatewayRouteTableIDRef is a reference to an API used to set
	// the TransitGatewayRouteTableID.
	// +optional
	TransitGatewayRouteTableIDRef *xpv1.Reference `json:"transitGatewayRouteTableIdRef,omitempty"`

	// TransitGatewayRouteTableIDSelector selects references to API used
	// to set the TransitGatewayRouteTableID.
	// +optional
	TransitGatewayRouteTableIDSelector *xpv1.Selector `json:"transitGatewayRouteTableIdSelector,omitempty"`
}

// CustomTransitGatewayRouteTableParameters are custom parameters for TransitGatewayRouteTableParameters
type CustomTransitGatewayRouteTableParameters struct {
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

// CustomFlowLogParameters are custom parameters for FlowLogParameters
type CustomFlowLogParameters struct {
	// Metadata tagging key value pairs
	// +optional
	Tags []Tag `json:"tags,omitempty"`
	// The ID of the subnet, network interface, or VPC for which you want to create
	// a flow log.
	//
	// Constraints: Maximum of 1000 resources
	// +optional
	ResourceIDs []*string `json:"resourceIDs"`
	// The type of resource for which to create the flow log. For example, if you
	// specified a VPC ID for the ResourceId property, specify VPC for this property.
	// +optional
	ResourceType *string `json:"resourceType"`
	// The ID of the VPC for which you want to create a flow log.
	// +optional
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-aws/apis/ec2/v1beta1.VPC
	VPCID *string `json:"vpcId,omitempty"`
	// VPCIDRef is a reference to an API used to set
	// the VPCID.
	// +optional
	VPCIDRef *xpv1.Reference `json:"vpcIdRef,omitempty"`
	// VPCIDSelector selects references to API used
	// to set the VPCID.
	// +optional
	VPCIDSelector *xpv1.Selector `json:"vpcIdSelector,omitempty"`

	// The ID of the TransitGateway for which you want to create a flow log.
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

	// The ID of the TransitGatewayAttachment for which you want to create a flow log.
	// +optional
	// +crossplane:generate:reference:type=TransitGatewayVPCAttachment
	TransitGatewayAttachmentID *string `json:"transitGatewayAttachmentId,omitempty"`

	// TransitGatewayAttachmentIDRef is a reference to an API used to set
	// the TransitGatewayAttachmentID.
	// +optional
	TransitGatewayAttachmentIDRef *xpv1.Reference `json:"transitGatewayAttachmentIdRef,omitempty"`

	// TransitGatewayAttachmentIDSelector selects references to API used
	// to set the TransitGatewayAttachmentID.
	// +optional
	TransitGatewayAttachmentIDSelector *xpv1.Selector `json:"transitGatewayAttachmentIdSelector,omitempty"`

	// The ID of the one or more Subnets for which you want to create a flow log.
	// +optional
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-aws/apis/ec2/v1beta1.Subnet
	SubnetID *string `json:"subnetId,omitempty"`

	// SubnetIDRefs is a referenc to a SubnetID used to set
	// the SubnetID.
	// +optional
	SubnetIDRef *xpv1.Reference `json:"subnetIdRef,omitempty"`

	// SubnetIDSelector selects reference to a SubnetID used
	// to set the SubnetID.
	// +optional
	SubnetIDSelector *xpv1.Selector `json:"subnetIdSelector,omitempty"`

	// The ID of the NetworkInterface for which you want to create a flow log.
	// +optional
	NetworkInterfaceID *string `json:"networkInterfaceId"`

	// The ARN for the IAM role that permits Amazon EC2
	// to publish flow logs to a CloudWatch Logs log group in your
	// account. \n If you specify LogDestinationType as s3, do not
	// specify DeliverLogsPermissionArn or LogGroupName.
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-aws/apis/iam/v1beta1.Role
	// +crossplane:generate:reference:extractor=github.com/crossplane-contrib/provider-aws/apis/iam/v1beta1.RoleARN()
	DeliverLogsPermissionARN *string `json:"deliverLogsPermissionArn,omitempty"`

	// DeliverLogsPermissionARNRef is a reference to DeliverLogsPermissionARN used to set
	// the DeliverLogsPermissionARN.
	// +optional
	DeliverLogsPermissionARNRef *xpv1.Reference `json:"deliverLogsPermissionArnRef,omitempty"`

	// DeliverLogsPermissionARNelector selects a reference to DeliverLogsPermissionARN used
	// to set the DeliverLogsPermissionARN.
	// +optional
	DeliverLogsPermissionARNSelector *xpv1.Selector `json:"deliverLogsPermissionArnSelector,omitempty"`

	// The Amazon Resource Names (ARNs) of a CloudWatch Log Group.
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-aws/apis/cloudwatchlogs/v1alpha1.LogGroup
	// +crossplane:generate:reference:extractor=github.com/crossplane-contrib/provider-aws/apis/cloudwatchlogs/v1alpha1.LogGroupARN()
	CloudWatchLogDestination *string `json:"cloudWatchLogDestination,omitempty"`

	// CloudWatchLogDestinationRef is a reference to a CloudWatch Log Group used to set
	// the CloudWatchLogDestination.
	// +optional
	CloudWatchLogDestinationRef *xpv1.Reference `json:"cloudWatchLogDestinationRef,omitempty"`

	// CloudWatchLogDestinationSelector selects a reference to a CloudWatch Log Group used
	// to set the CloudWatchLogDestination.
	// +optional
	CloudWatchLogDestinationSelector *xpv1.Selector `json:"cloudWatchLogDestinationSelector,omitempty"`

	// The Amazon Resource Names (ARNs) of a CloudWatch Log Group.
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-aws/apis/s3/v1beta1.Bucket
	// +crossplane:generate:reference:extractor=github.com/crossplane-contrib/provider-aws/apis/s3/v1beta1.BucketARN()
	S3BucketLogDestination *string `json:"s3BucketLogDestination,omitempty"`

	// S3BucketLogDestinationRef is a reference to a S3 Bucket  to set
	// the S3BucketLogDestination.
	// +optional
	S3BucketLogDestinationRef *xpv1.Reference `json:"s3BucketLogDestinationRef,omitempty"`

	// S3BucketLogDestinationSelector selects a reference to a S3 Bucket used
	// to set the S3BucketLogDestination.
	// +optional
	S3BucketLogDestinationSelector *xpv1.Selector `json:"s3BucketLogDestinationSelector,omitempty"`

	// S3BucketSubfolder allows you to use a subfolder in a S3Bucket, if S3BucketLogDestination is used to determin the ARN of a Bucket managed by crossplane
	// S3BucketSubfolder starts and ends without an slash e.g. test/test2
	// You cannot use AWSLogs as a subfolder name. This is a reserved term.
	// +optional
	S3BucketSubfolder *string `json:"s3BucketSubfolder,omitempty"`
}

// CustomCapacityReservationParameters includes the custom fields of Stage.
type CustomCapacityReservationParameters struct {
}
