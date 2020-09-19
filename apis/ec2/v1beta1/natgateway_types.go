package v1beta1

import (
	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NatGatewayParameters defined the desired state of an AWS VPC NAT Gateway
type NatGatewayParameters struct {
	// AllocationID is the Elastic IP allocation ID
	// +immutable
	// +optional
	AllocationID *string `json:"allocationId,omitempty"`

	// AllocationIDRef references an EIP and retrieves it's allocation id
	// +immutable
	// +optional
	AllocationIDRef *runtimev1alpha1.Reference `json:"allocationIdRef,omitempty"`

	// AllocationIDSelector references an EIP by selector and retrieves it's allocation id
	// +immutable
	// +optional
	AllocationIDSelector *runtimev1alpha1.Selector `json:"allocationIdSelector,omitempty"`

	// SubnetID is the subnet the NAT gateways needs to be associated to
	// +immutable
	// +optional
	SubnetID *string `json:"subnetId,omitempty"`

	// SubnetIDRef references a subnet and retrives it's subnet id
	// +immutable
	// +optional
	SubnetIDRef *runtimev1alpha1.Reference `json:"subnetIdRef,omitempty"`

	// SubnetIDSelector references a subnet by selector and retrives it's subnet id
	// +immutable
	// +optional
	SubnetIDSelector *runtimev1alpha1.Selector `json:"subnetIdSelector,omitempty"`

	// Tags represents to current ec2 tags.
	// +optional
	Tags []Tag `json:"tags,omitempty"`
}

// NatGatewaySpec defines the desired state of a NAT Gateway
type NatGatewaySpec struct {
	runtimev1alpha1.ResourceSpec `json:",inline"`
	ForProvider                  NatGatewayParameters `json:"forProvider"`
}

// NatGatewayObservation keeps the state for the CR
type NatGatewayObservation struct {
	CreateTime          *metav1.Time        `json:"createTime"`
	NatGatewayAddresses []NatGatewayAddress `json:"natGatewayAddresses"`
	NatGatewayID        string              `json:"natGatewayId"`
	SubnetID            string              `json:"subnetId"`
	Tags                []Tag               `json:"tags,omitempty"`
	VpcID               string              `json:"vpcId"`
}

// NatGatewayAddress describes the details of network
type NatGatewayAddress struct {
	AllocationID       string `json:"allocationId"`
	NetworkInterfaceID string `json:"networkInterfaceId"`
	PrivateIP          string `json:"privateIp"`
	PublicIP           string `json:"publicIp"`
}

// NatGatewayStatus describes the observed state
type NatGatewayStatus struct {
	runtimev1alpha1.ResourceStatus `json:",inline"`
	AtProvider                     NatGatewayObservation `json:"atProvider"`
}

// +kubebuilder:object:root=true

// An NatGateway is a managed resource that represents an AWS VPC NAT
// Gateway.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="ID",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="VPC",type="string",JSONPath=".spec.atProvider.vpcId"
// +kubebuilder:printcolumn:name="SUBNET",type="string",JSONPath=".spec.forProvider.subnetId"
// +kubebuilder:printcolumn:name="ALLOCATION ID",type="string",JSONPath=".spec.forProvider.allocationId"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,aws}
type NatGateway struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NatGatewaySpec   `json:"spec"`
	Status NatGatewayStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// NatGatewayList contains a list of NatGateways
type NatGatewayList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NatGateway `json:"items"`
}
