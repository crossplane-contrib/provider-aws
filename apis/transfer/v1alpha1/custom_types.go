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

import xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"

// CustomUserParameters includes custom additional fields for UserParameters.
type CustomUserParameters struct {

	// A system-assigned unique identifier for a server instance. This is the specific
	// server that you added your user to.
	// +optional
	ServerID *string `json:"serverID,omitempty"`

	// ServerIDRef is a reference to an server instance.
	// +optional
	ServerIDRef *xpv1.Reference `json:"serverIDRef,omitempty"`

	// ServerIDSelector selects references to an server instance.
	// +optional
	ServerIDSelector *xpv1.Selector `json:"serverIDSelector,omitempty"`

	// The IAM role that controls your users' access to your Amazon S3 bucket. The
	// policies attached to this role will determine the level of access you want
	// to provide your users when transferring files into and out of your Amazon
	// S3 bucket or buckets. The IAM role should also contain a trust relationship
	// that allows the server to access your resources when servicing your users'
	// transfer requests.
	// +optional
	Role *string `json:"role,omitempty"`

	// RoleRef is a reference to a IAM role.
	// +optional
	RoleRef *xpv1.Reference `json:"roleRef,omitempty"`

	// RoleSelector selects references to a IAM role.
	// +optional
	RoleSelector *xpv1.Selector `json:"roleSelector,omitempty"`
}

// CustomServerParameters includes custom additional fields for ServerParameters.
type CustomServerParameters struct {
	// The virtual private cloud (VPC) endpoint settings that are configured for
	// your server. When you host your endpoint within your VPC, you can make it
	// accessible only to resources within your VPC, or you can attach Elastic IPs
	// and make it accessible to clients over the internet. Your VPC's default security
	// groups are automatically assigned to your endpoint.
	CustomEndpointDetails *CustomEndpointDetails `json:"endpointDetails,omitempty"`

	// CertificateRef is a reference to a Certificate.
	// +optional
	CertificateRef *xpv1.Reference `json:"certificateRef,omitempty"`

	// CertificateSelector selects references to a Certificate.
	// +optional
	CertificateSelector *xpv1.Selector `json:"certificateSelector,omitempty"`

	// LoggingRoleRef is a reference to a IAM role.
	// +optional
	LoggingRoleRef *xpv1.Reference `json:"loggingRoleRef,omitempty"`

	// LoggingRoleSelector selects references to a IAM role.
	// +optional
	LoggingRoleSelector *xpv1.Selector `json:"loggingRoleSelector,omitempty"`
}

// CustomEndpointDetails includes custom additional fields for UserParameters.
type CustomEndpointDetails struct {
	// A list of address allocation IDs that are required to attach an Elastic IP
	// address to your server's endpoint.
	//
	// This property can only be set when EndpointType is set to VPC and it is only
	// valid in the UpdateServer API.
	AddressAllocationIDs []*string `json:"addressAllocationIDs,omitempty"`

	// A list of security groups IDs that are available to attach to your server's
	// endpoint.
	//
	// This property can only be set when EndpointType is set to VPC.
	//
	// You can edit the SecurityGroupIds property in the UpdateServer (https://docs.aws.amazon.com/transfer/latest/userguide/API_UpdateServer.html)
	// API only if you are changing the EndpointType from PUBLIC or VPC_ENDPOINT
	// to VPC. To change security groups associated with your server's VPC endpoint
	// after creation, use the Amazon EC2 ModifyVpcEndpoint (https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_ModifyVpcEndpoint.html)
	// API.
	SecurityGroupIDs []*string `json:"securityGroupIDs,omitempty"`

	// SecurityGroupIDRefs is a list of references to SecurityGroups used to set
	// the SecurityGroupIDs.
	// +optional
	SecurityGroupIDRefs []xpv1.Reference `json:"securityGroupIDRefs,omitempty"`

	// SecurityGroupIDsSelector selects references to SecurityGroupID used
	// to set the SecurityGroupIDs.
	// +optional
	SecurityGroupIDSelector *xpv1.Selector `json:"securityGroupIDSelector,omitempty"`

	// A list of subnet IDs that are required to host your server endpoint in your
	// VPC.
	//
	// This property can only be set when EndpointType is set to VPC.
	SubnetIDs []*string `json:"subnetIDs,omitempty"`

	// SubnetIDsRefs is a list of references to Subnets used to set
	// the SubnetIDs.
	// +optional
	SubnetIDRefs []xpv1.Reference `json:"subnetIDRefs,omitempty"`

	// SubnetIDsSelector selects references to Subnets used
	// to set the SubnetIds.
	// +optional
	SubnetIDSelector *xpv1.Selector `json:"subnetIDSelector,omitempty"`

	// The ID of the VPC endpoint.
	//
	// This property can only be set when EndpointType is set to VPC_ENDPOINT.
	//
	// For more information, see https://docs.aws.amazon.com/transfer/latest/userguide/create-server-in-vpc.html#deprecate-vpc-endpoint.
	VPCEndpointID *string `json:"vpcEndpointID,omitempty"`

	// The VPC ID of the VPC in which a server's endpoint will be hosted.
	//
	// This property can only be set when EndpointType is set to VPC.
	VPCID *string `json:"vpcID,omitempty"`

	// VPCIDRef is a reference to a VPCID.
	// +optional
	VPCIDRef *xpv1.Reference `json:"vpcIDRef,omitempty"`

	// VPCIDSelector selects references to a VPCID.
	// +optional
	VPCIDSelector *xpv1.Selector `json:"vpcIDSelector,omitempty"`
}
