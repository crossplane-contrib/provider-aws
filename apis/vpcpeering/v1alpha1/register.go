package v1alpha1

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

// Package type metadata.
const (
	Group   = "vpcpeering.aws.crossplane.io"
	Version = "v1alpha1"
)

var (
	// SchemeGroupVersion is group version used to register these objects
	SchemeGroupVersion = schema.GroupVersion{Group: Group, Version: Version}

	// SchemeBuilder is used to add go types to the GroupVersionKind scheme
	SchemeBuilder = &scheme.Builder{GroupVersion: SchemeGroupVersion}
)

// Peering type metadata.
var (
	VPCPeeringConnectionKind             = "VPCPeeringConnection"
	VPCPeeringConnectionGroupKind        = schema.GroupKind{Group: Group, Kind: VPCPeeringConnectionKind}.String()
	VPCPeeringConnectionKindAPIVersion   = VPCPeeringConnectionKind + "." + SchemeGroupVersion.String()
	VPCPeeringConnectionGroupVersionKind = SchemeGroupVersion.WithKind(VPCPeeringConnectionKind)
)

func init() {
	SchemeBuilder.Register(&VPCPeeringConnection{}, &VPCPeeringConnectionList{})
}
