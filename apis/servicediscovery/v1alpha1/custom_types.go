package v1alpha1

import (
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"k8s.io/utils/ptr"
)

// AnnotationKeyOperationID is the key in the annotations map of a
// Cloud Map managed resource for the OperationId returned by API calls
const AnnotationKeyOperationID = CRDGroup + "/operation-id"

// CustomServiceParameters are custom parameters for Services.
type CustomServiceParameters struct {

	// +optional
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-aws/apis/servicediscovery/v1alpha1.PrivateDNSNamespace
	// +crossplane:generate:reference:refFieldName=ServiceNameRef
	// +crossplane:generate:reference:selectorFieldName=ServiceNameSelector
	ServiceName *string `json:"serviceName,omitempty"`

	// ServiceNameRef is a reference to a service used to set
	// the ServiceName.
	// +optional
	ServiceNameRef *xpv1.Reference `json:"serviceNameRef,omitempty"`

	// ServiceNameSelector selects references to service used
	// to set the ServiceName.
	// +optional
	ServiceNameSelector *xpv1.Selector `json:"serviceNameSelector,omitempty"`
}

// CustomServiceObservation includes the custom status fields of Services.
type CustomServiceObservation struct{}

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

// CustomPrivateDNSNamespaceObservation includes the custom status fields of PrivateDNSNamespaces.
type CustomPrivateDNSNamespaceObservation struct{}

// CustomHTTPNamespaceParameters are custom parameters for HTTPNamespaces.
type CustomHTTPNamespaceParameters struct{}

// CustomHTTPNamespaceObservation includes the custom status fields of HTTPNamespaces.
type CustomHTTPNamespaceObservation struct{}

// CustomPublicDNSNamespaceParameters are custom parameters for PublicDNSNamespaces.
type CustomPublicDNSNamespaceParameters struct{}

// CustomPublicDNSNamespaceObservation includes the custom status fields of PublicDNSNamespaces.
type CustomPublicDNSNamespaceObservation struct{}

// GetOperationID returns the last operation id.
func (in *HTTPNamespace) GetOperationID() *string {
	if val, ok := in.GetAnnotations()[AnnotationKeyOperationID]; ok {
		return &val
	}
	return nil
}

// SetOperationID sets the last operation id.
func (in *HTTPNamespace) SetOperationID(id *string) {
	meta.AddAnnotations(in, map[string]string{AnnotationKeyOperationID: ptr.Deref(id, "")})
}

// GetDescription returns the description.
func (in *HTTPNamespace) GetDescription() *string {
	return in.Spec.ForProvider.Description
}

// GetTTL returns the TTL.
func (in *HTTPNamespace) GetTTL() *int64 {
	return nil
}

// GetTags returns the tags.
func (in *HTTPNamespace) GetTags() []*Tag {
	return in.Spec.ForProvider.Tags
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
	meta.AddAnnotations(in, map[string]string{AnnotationKeyOperationID: ptr.Deref(id, "")})
}

// GetDescription returns the description.
func (in *PrivateDNSNamespace) GetDescription() *string {
	return in.Spec.ForProvider.Description
}

// GetTTL returns the TTL.
func (in *PrivateDNSNamespace) GetTTL() *int64 {
	if in.Spec.ForProvider.Properties == nil || in.Spec.ForProvider.Properties.DNSProperties == nil || in.Spec.ForProvider.Properties.DNSProperties.SOA == nil {
		return nil
	}
	return in.Spec.ForProvider.Properties.DNSProperties.SOA.TTL
}

// GetTags returns the tags.
func (in *PrivateDNSNamespace) GetTags() []*Tag {
	return in.Spec.ForProvider.Tags
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
	meta.AddAnnotations(in, map[string]string{AnnotationKeyOperationID: ptr.Deref(id, "")})
}

// GetDescription returns the description.
func (in *PublicDNSNamespace) GetDescription() *string {
	return in.Spec.ForProvider.Description
}

// GetTTL returns the TTL.
func (in *PublicDNSNamespace) GetTTL() *int64 {
	if in.Spec.ForProvider.Properties == nil || in.Spec.ForProvider.Properties.DNSProperties == nil || in.Spec.ForProvider.Properties.DNSProperties.SOA == nil {
		return nil
	}
	return in.Spec.ForProvider.Properties.DNSProperties.SOA.TTL
}

// GetTags returns the tags.
func (in *PublicDNSNamespace) GetTags() []*Tag {
	return in.Spec.ForProvider.Tags
}
