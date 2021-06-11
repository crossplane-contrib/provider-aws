package v1alpha1

import xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"

// CustomServiceParameters are custom parameters for Services.
type CustomServiceParameters struct{}

// CustomPrivateDNSNamespaceParameters are custom parameters for PrivateDNSNamespaces.
type CustomPrivateDNSNamespaceParameters struct {

	// VPC of the PrivateDNSNamespace.
	// One if vpc, vpcRef or vpcSelector has to be supplied.
	VPC *string `json:"vpc,omitempty"`

	// A referencer to retrieve the ID of a VPC
	VPCRef *xpv1.Reference `json:"vpcRef,omitempty"`

	// A selector to select a referencer to retrieve the ID of a VPC.
	VPCSelector *xpv1.Selector `json:"vpcSelector,omitempty"`
}

// CustomHTTPNamespaceParameters are custom parameters for HTTPNamespaces.
type CustomHTTPNamespaceParameters struct{}

// CustomPublicDNSNamespaceParameters are custom parameters for PublicDNSNamespaces.
type CustomPublicDNSNamespaceParameters struct{}
