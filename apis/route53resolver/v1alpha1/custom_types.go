package v1alpha1

import xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"

// CustomResolverEndpointParameters are custom parameters for ResolverEndpoint
type CustomResolverEndpointParameters struct {
	// The ID of one or more security groups that you want to use to control access
	// to this VPC. The security group that you specify must include one or more
	// inbound rules (for inbound Resolver endpoints) or outbound rules (for outbound
	// Resolver endpoints). Inbound and outbound rules must allow TCP and UDP access.
	// For inbound access, open port 53. For outbound access, open the port that
	// you're using for DNS queries on your network.
	SecurityGroupIDs []string `json:"securityGroupIds,omitempty"`
	// SecurityGroupIDRefs is a list of references to SecurityGroups used to set
	// the SecurityGroupIDs.
	// +optional
	SecurityGroupIDRefs []xpv1.Reference `json:"securityGroupIdRefs,omitempty"`

	// SecurityGroupIDsSelector selects references to SecurityGroupID used
	// to set the SecurityGroupIDs.
	// +optional
	SecurityGroupIDSelector *xpv1.Selector `json:"securityGroupIdSelector,omitempty"`

	// IPAddresses are the subnets and IP addresses in your VPC that DNS queries originate from
	// (for outbound endpoints) or that you forward DNS queries to (for inbound
	// endpoints). The subnet ID uniquely identifies a VPC.
	IPAddresses []*IPAddressRequest `json:"ipAddresses"`
}

// CustomResolverEndpointObservation includes the custom status fields of ResolverEndpoint.
type CustomResolverEndpointObservation struct{}

// CustomResolverRuleParameters are custom parameters for CustomResolverRule
type CustomResolverRuleParameters struct {
	// ResolverEndpointIDRef is the reference to the ResolverEndpoint used
	// to set the ResolverEndpointID
	// +optional
	ResolverEndpointIDRef *xpv1.Reference `json:"resolverEndpointIdRefs,omitempty"`
	// ResolverEndpointIDSelector selects references to ResolverEndpoint used
	// to set the ResolverEndpointID
	// +optional
	ResolverEndpointIDSelector *xpv1.Selector `json:"resolverEndpointIdSelector,omitempty"`
}

// CustomResolverRuleObservation includes the custom status fields of ResolverRule.
type CustomResolverRuleObservation struct{}

// CustomResolverQueryLogConfigParameters are custom parameters for CustomResolverQueryLogConfig
type CustomResolverQueryLogConfigParameters struct {
}

// IPAddressRequest is used by ResolverEndpoint
type IPAddressRequest struct {
	// IP address that you want to use for DNS queries.
	IP *string `json:"ip,omitempty"`

	// SubnetId is the ID of the subnet that contains the IP address.
	SubnetID *string `json:"subnetId,omitempty"`

	// SubnetIDRefs is a list of references to Subnets used to set
	// the SubnetIDs.
	SubnetIDRef *xpv1.Reference `json:"subnetIdRef,omitempty"`

	// SubnetIDSelector selects references to Subnets used
	// to set the SubnetIDs.
	SubnetIDSelector *xpv1.Selector `json:"subnetIdSelector,omitempty"`
}
