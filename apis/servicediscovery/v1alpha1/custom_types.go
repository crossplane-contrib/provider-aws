package v1alpha1

import (
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/meta"

	aws "github.com/crossplane/provider-aws/pkg/clients"
)

// AnnotationKeyOperationID is the key in the annotations map of a
// Cloud Map managed resource for the OperationId returned by API calls
const AnnotationKeyOperationID = CRDGroup + "/operation-id"

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
	if val, ok := in.GetAnnotations()[AnnotationKeyOperationID]; ok {
		return &val
	}
	return nil
}

// SetOperationID sets the last operation id.
func (in *HTTPNamespace) SetOperationID(id *string) {
	meta.AddAnnotations(in, map[string]string{AnnotationKeyOperationID: aws.StringValue(id)})
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
	if val, ok := in.GetAnnotations()[AnnotationKeyOperationID]; ok {
		return &val
	}
	return nil
}

// SetOperationID sets the last operation id.
func (in *PrivateDNSNamespace) SetOperationID(id *string) {
	meta.AddAnnotations(in, map[string]string{AnnotationKeyOperationID: aws.StringValue(id)})
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
	if val, ok := in.GetAnnotations()[AnnotationKeyOperationID]; ok {
		return &val
	}
	return nil
}

// SetOperationID sets the last operation id.
func (in *PublicDNSNamespace) SetOperationID(id *string) {
	meta.AddAnnotations(in, map[string]string{AnnotationKeyOperationID: aws.StringValue(id)})
}

// GetDescription returns the description.
func (in *PublicDNSNamespace) GetDescription() *string {
	return in.Spec.ForProvider.Description
}

// SetDescription sets the description.
func (in *PublicDNSNamespace) SetDescription(d *string) {
	in.Spec.ForProvider.Description = d
}
