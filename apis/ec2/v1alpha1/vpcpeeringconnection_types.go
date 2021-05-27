package v1alpha1

// +kubebuilder:storageversion

// CustomVPCPeeringConnectionParameters are custom parameters for VPCPeeringConnection
type CustomVPCPeeringConnectionParameters struct {

	// Automatically accepts the peering connection. If this is not set, the peering connection
	// will be created, but will be in pending-acceptance state. This will only lead to an active
	// connection if both VPCs are in the same tenant.
	AcceptRequest bool `json:"acceptRequest,omitempty"`
}
