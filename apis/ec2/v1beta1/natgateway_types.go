package v1beta1

import (
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Defines the states of NatGateway
const (
	NatGatewayStatusPending   = "pending"
	NatGatewayStatusFailed    = "failed"
	NatGatewayStatusAvailable = "available"
	NatGatewayStatusDeleting  = "deleting"
	NatGatewayStatusDeleted   = "deleted"
)

// NATGatewayParameters defined the desired state of an AWS VPC NAT Gateway
type NATGatewayParameters struct {

	// TODO(muvaf): Region is a required field but in order to keep backward compatibility
	// with old Provider type and not bear the cost of bumping to v1beta2, we're
	// keeping it optional for now. Reconsider before v1beta2 or v1.

	// Region is the region you'd like your NATGateway to be created in.
	// +immutable
	Region string `json:"region"`

	// AllocationID is the Elastic IP allocation ID
	// +immutable
	// +optional
	// +crossplane:generate:reference:type=Address
	AllocationID *string `json:"allocationId,omitempty"`

	// AllocationIDRef references an EIP and retrieves it's allocation id
	// +immutable
	// +optional
	AllocationIDRef *xpv1.Reference `json:"allocationIdRef,omitempty"`

	// AllocationIDSelector references an EIP by selector and retrieves it's allocation id
	// +immutable
	// +optional
	AllocationIDSelector *xpv1.Selector `json:"allocationIdSelector,omitempty"`

	// SubnetID is the subnet the NAT gateways needs to be associated to
	// +immutable
	// +optional
	// +crossplane:generate:reference:type=Subnet
	SubnetID *string `json:"subnetId,omitempty"`

	// SubnetIDRef references a subnet and retrives it's subnet id
	// +immutable
	// +optional
	SubnetIDRef *xpv1.Reference `json:"subnetIdRef,omitempty"`

	// SubnetIDSelector references a subnet by selector and retrives it's subnet id
	// +immutable
	// +optional
	SubnetIDSelector *xpv1.Selector `json:"subnetIdSelector,omitempty"`

	// Indicates whether the NAT gateway supports public or private connectivity. The
	// default is public connectivity.
	// +optional
	// +kubebuilder:validation:Enum=public;private
	ConnectivityType string `json:"connectivityType,omitempty"`

	// Tags represents to current ec2 tags.
	// +optional
	Tags []Tag `json:"tags,omitempty"`
}

// NATGatewaySpec defines the desired state of a NAT Gateway
type NATGatewaySpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       NATGatewayParameters `json:"forProvider"`
}

// NATGatewayObservation keeps the state for the CR
type NATGatewayObservation struct {
	CreateTime          *metav1.Time        `json:"createTime,omitempty"`
	DeleteTime          *metav1.Time        `json:"deleteTime,omitempty"`
	FailureCode         string              `json:"failureCode,omitempty"`
	FailureMessage      string              `json:"failureMessage,omitempty"`
	NatGatewayAddresses []NATGatewayAddress `json:"natGatewayAddresses,omitempty"`
	NatGatewayID        string              `json:"natGatewayId,omitempty"`
	State               string              `json:"state,omitempty"`
	VpcID               string              `json:"vpcId,omitempty"`
}

// NATGatewayAddress describes the details of network
type NATGatewayAddress struct {
	AllocationID       string `json:"allocationId,omitempty"`
	NetworkInterfaceID string `json:"networkInterfaceId,omitempty"`
	PrivateIP          string `json:"privateIp,omitempty"`
	PublicIP           string `json:"publicIp,omitempty"`
}

// NATGatewayStatus describes the observed state
type NATGatewayStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          NATGatewayObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A NATGateway is a managed resource that represents an AWS VPC NAT
// Gateway.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="ID",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="VPC",type="string",JSONPath=".status.atProvider.vpcId"
// +kubebuilder:printcolumn:name="SUBNET",type="string",JSONPath=".spec.forProvider.subnetId"
// +kubebuilder:printcolumn:name="ALLOCATION ID",type="string",JSONPath=".spec.forProvider.allocationId"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,aws}
// +kubebuilder:storageversion
type NATGateway struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NATGatewaySpec   `json:"spec"`
	Status NATGatewayStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// NATGatewayList contains a list of NatGateways
type NATGatewayList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NATGateway `json:"items"`
}
