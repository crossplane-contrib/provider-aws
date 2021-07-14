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

// GetOperationID returns the last operation id.
func (in *HTTPNamespace) GetOperationID() *string {
	return in.Status.AtProvider.OperationID
}

// SetOperationID sets the last operation id.
func (in *HTTPNamespace) SetOperationID(id *string) {
	in.Status.AtProvider.OperationID = id
}

// GetDescription returns the description.
func (in *HTTPNamespace) GetDescription() *string {
	return in.Spec.ForProvider.Description
}

// SetDescription sets the description.
func (in *HTTPNamespace) SetDescription(d *string) {
	in.Spec.ForProvider.Description = d
}

// GetOperationID returns the last operation id.
func (in *PrivateDNSNamespace) GetOperationID() *string {
	return in.Status.AtProvider.OperationID
}

// SetOperationID sets the last operation id.
func (in *PrivateDNSNamespace) SetOperationID(id *string) {
	in.Status.AtProvider.OperationID = id
}

// GetDescription returns the description.
func (in *PrivateDNSNamespace) GetDescription() *string {
	return in.Spec.ForProvider.Description
}

// SetDescription sets the description.
func (in *PrivateDNSNamespace) SetDescription(d *string) {
	in.Spec.ForProvider.Description = d
}

// GetOperationID returns the last operation id.
func (in *PublicDNSNamespace) GetOperationID() *string {
	return in.Status.AtProvider.OperationID
}

// SetOperationID sets the last operation id.
func (in *PublicDNSNamespace) SetOperationID(id *string) {
	in.Status.AtProvider.OperationID = id
}

// GetDescription returns the description.
func (in *PublicDNSNamespace) GetDescription() *string {
	return in.Spec.ForProvider.Description
}

// SetDescription sets the description.
func (in *PublicDNSNamespace) SetDescription(d *string) {
	in.Spec.ForProvider.Description = d
}
