package v1alpha1

import xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"

// CustomJobRunParameters includes the custom fields of JobRun.
type CustomJobRunParameters struct {
	// The virtual cluster ID for which the job run request is submitted.
	// +kubebuilder:validation:Required
	// +crossplane:generate:reference:type=VirtualCluster
	VirtualClusterID *string `json:"virtualClusterId,omitempty"`
	// VirtualClusterIdRef is a reference to an API used to set
	// the VirtualClusterId.
	// +optional
	VirtualClusterIDRef *xpv1.Reference `json:"virtualClusterIdRef,omitempty"`
	// VirtualClusterIdSelector is a reference to an API used to set
	// the VirtualClusterIdSelector.
	// +optional
	VirtualClusterIDSelector *xpv1.Selector `json:"virtualClusterIdSelector,omitempty"`
}

// CustomJobRunObservation includes the custom status fields of JobRun.
type CustomJobRunObservation struct{}

// CustomVirtualClusterParameters includes the custom fields of VirtualCluster.
type CustomVirtualClusterParameters struct{}

// CustomVirtualClusterObservation includes the custom status fields of VirtualCluster.
type CustomVirtualClusterObservation struct{}
