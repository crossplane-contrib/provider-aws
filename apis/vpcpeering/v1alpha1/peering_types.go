package v1alpha1

import (
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// VPCPeeringConnectionParameters defines the desired state of VPCPeeringConnection
type VPCPeeringConnectionParameters struct {
	// Region is which region the VPCPeeringConnection will be created.
	// +kubebuilder:validation:Required
	Region string `json:"region"`
	// The AWS account ID of the owner of the accepter VPC.
	//
	// Default: Your AWS account ID
	PeerOwnerID *string `json:"peerOwnerID,omitempty"`
	// The Region code for the accepter VPC, if the accepter VPC is located in a
	// Region other than the Region in which you make the request.
	//
	// Default: The Region in which you make the request.
	PeerRegion *string `json:"peerRegion,omitempty"`

	PeerCIDR *string `json:"peerCidr,omitempty"`
	// The ID of the VPC with which you are creating the VPC peering connection.
	// You must specify this parameter in the request.
	PeerVPCID *string `json:"peerVPCID,omitempty"`
	// The tags to assign to the peering connection.
	Tags []*Tag `json:"tags,omitempty"`
	// The ID of the requester VPC. You must specify this parameter in the request.
	VPCID                                *string `json:"vpcID,omitempty"`
	HostZoneID *string `json:"hostZoneID,omitempty "`
}

type Tag struct {
	Key *string `json:"key,omitempty"`

	Value *string `json:"value,omitempty"`
}

// VPCPeeringConnectionSpec defines the desired state of VPCPeeringConnection
type VPCPeeringConnectionSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       VPCPeeringConnectionParameters `json:"forProvider"`
}

// VPCPeeringConnectionObservation defines the observed state of VPCPeeringConnection
type VPCPeeringConnectionObservation struct {
	// Information about the accepter VPC. CIDR block information is only returned
	// when describing an active VPC peering connection.
	AccepterVPCInfo *VPCPeeringConnectionVPCInfo `json:"accepterVPCInfo,omitempty"`
	// The time that an unaccepted VPC peering connection will expire.
	ExpirationTime *metav1.Time `json:"expirationTime,omitempty"`
	// Information about the requester VPC. CIDR block information is only returned
	// when describing an active VPC peering connection.
	RequesterVPCInfo *VPCPeeringConnectionVPCInfo `json:"requesterVPCInfo,omitempty"`
	// The status of the VPC peering connection.
	Status *VPCPeeringConnectionStateReason `json:"status,omitempty"`
	// Any tags assigned to the resource.
	Tags []*Tag `json:"tags,omitempty"`
	// The ID of the VPC peering connection.
	VPCPeeringConnectionID *string `json:"vpcPeeringConnectionID,omitempty"`

	Phase  *string `json:"phase,omitempty"`
}

type VPCPeeringConnectionStateReason struct {
	Code *string `json:"code,omitempty"`

	Message *string `json:"message,omitempty"`
}

type VPCPeeringConnectionVPCInfo struct {
	CIDRBlock *string `json:"cidrBlock,omitempty"`

	CIDRBlockSet []*CIDRBlock `json:"cidrBlockSet,omitempty"`

	IPv6CIDRBlockSet []*IPv6CIDRBlock `json:"ipv6CIDRBlockSet,omitempty"`

	OwnerID *string `json:"ownerID,omitempty"`
	// Describes the VPC peering connection options.
	PeeringOptions *VPCPeeringConnectionOptionsDescription `json:"peeringOptions,omitempty"`

	Region *string `json:"region,omitempty"`

	VPCID *string `json:"vpcID,omitempty"`
}

type VPCPeeringConnectionOptionsDescription struct {
	AllowDNSResolutionFromRemoteVPC *bool `json:"allowDNSResolutionFromRemoteVPC,omitempty"`

	AllowEgressFromLocalClassicLinkToRemoteVPC *bool `json:"allowEgressFromLocalClassicLinkToRemoteVPC,omitempty"`

	AllowEgressFromLocalVPCToRemoteClassicLink *bool `json:"allowEgressFromLocalVPCToRemoteClassicLink,omitempty"`
}

type CIDRBlock struct {
	CIDRBlock *string `json:"cidrBlock,omitempty"`
}

type IPv6CIDRBlock struct {
	IPv6CIDRBlock *string `json:"ipv6CIDRBlock,omitempty"`
}

// VPCPeeringConnectionStatus defines the observed state of VPCPeeringConnection.
type VPCPeeringConnectionStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          VPCPeeringConnectionObservation `json:"atProvider"`
}

// +kubebuilder:object:root=true

// VPCPeeringConnection is the Schema for the VPCPeeringConnections API
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,aws}
type VPCPeeringConnection struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              VPCPeeringConnectionSpec   `json:"spec,omitempty"`
	Status            VPCPeeringConnectionStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// VPCPeeringConnectionList contains a list of VPCPeeringConnections
type VPCPeeringConnectionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VPCPeeringConnection `json:"items"`
}