package v1beta1

import (
	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/provider-aws/apis/identity/v1beta1"
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
}

// NatGatewaySpec defines the desired state of a NAT Gateway
type NatGatewaySpec struct {
	runtimev1alpha1.ResourceSpec `json:",inline"`
	ForProvider                  NatGatewayParameters `json:"forProvider"`
}

// NatGatewayObservation defines the observed state
type NatGatewayObservation struct {
	CreateTime          *metav1.Time `json:"createTime"`
	NatGatewayAddresses []NatGatewayAddress
	NatGatewayID        *string       `json:"natGatewayId"`
	SubnetID            *string       `json:"subnetId"`
	Tags                []v1beta1.Tag `json:"tags,omitempty"`
	VpcID               *string       `json:"vpcId"`
}

// NatGatewayAddress describes the details of network
type NatGatewayAddress struct {
	AllocationID       *string `json:"allocationId"`
	NetworkInterfaceID *string `json:"networkInterfaceId"`
	PrivateIP          *string `json:"privateIp"`
	PublicIP           *string `json:"publicIp"`
}
