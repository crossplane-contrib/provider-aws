package ec2

import (
	"testing"
	"time"

	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/google/go-cmp/cmp"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/crossplane-contrib/provider-aws/apis/ec2/v1beta1"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
)

var (
	natAllocationID       = "some allocation id"
	natNetworkInterfaceID = "some network interface id"
	natPrivateIP          = "some private ip"
	natPublicIP           = "some public ip"
	natGatewayID          = "some gateway id"
	natSubnetID           = "some subnet id"
	natVpcID              = "some vpc"
	natFailureCode        = "some failure code"
	natFailureMessage     = "some failure message"
)

func natTags() []ec2types.Tag {
	return []ec2types.Tag{
		{
			Key:   pointer.String("key1"),
			Value: pointer.String("value1"),
		},
		{
			Key:   pointer.String("key2"),
			Value: pointer.String("value2"),
		},
	}
}

func natAddresses() []ec2types.NatGatewayAddress {
	return []ec2types.NatGatewayAddress{
		{
			AllocationId:       pointer.String(natAllocationID),
			NetworkInterfaceId: pointer.String(natNetworkInterfaceID),
			PrivateIp:          pointer.String(natPrivateIP),
			PublicIp:           pointer.String(natPublicIP),
		},
	}
}

func specAddresses() []v1beta1.NATGatewayAddress {
	return []v1beta1.NATGatewayAddress{
		{
			AllocationID:       natAllocationID,
			NetworkInterfaceID: natNetworkInterfaceID,
			PrivateIP:          natPrivateIP,
			PublicIP:           natPublicIP,
		},
	}
}

func TestGenerateNATGatewayObservation(t *testing.T) {
	time := time.Now()
	cases := map[string]struct {
		in  ec2types.NatGateway
		out v1beta1.NATGatewayObservation
	}{
		"AllFilled": {
			in: ec2types.NatGateway{
				CreateTime:          &time,
				NatGatewayAddresses: natAddresses(),
				NatGatewayId:        pointer.String(natGatewayID),
				State:               v1beta1.NatGatewayStatusAvailable,
				SubnetId:            pointer.String(natSubnetID),
				Tags:                natTags(),
				VpcId:               pointer.String(natVpcID),
			},
			out: v1beta1.NATGatewayObservation{
				CreateTime:          &v1.Time{Time: time},
				NatGatewayAddresses: specAddresses(),
				NatGatewayID:        natGatewayID,
				State:               v1beta1.NatGatewayStatusAvailable,
				VpcID:               natVpcID,
			},
		},
		"DeleteTime": {
			in: ec2types.NatGateway{
				CreateTime:          &time,
				DeleteTime:          &time,
				NatGatewayAddresses: natAddresses(),
				NatGatewayId:        pointer.String(natGatewayID),
				State:               v1beta1.NatGatewayStatusPending,
				SubnetId:            pointer.String(natSubnetID),
				Tags:                natTags(),
				VpcId:               pointer.String(natVpcID),
			},
			out: v1beta1.NATGatewayObservation{
				CreateTime:          &v1.Time{Time: time},
				DeleteTime:          &v1.Time{Time: time},
				NatGatewayAddresses: specAddresses(),
				NatGatewayID:        natGatewayID,
				State:               v1beta1.NatGatewayStatusPending,
				VpcID:               natVpcID,
			},
		},
		"stateFailed": {
			in: ec2types.NatGateway{
				CreateTime:          &time,
				DeleteTime:          &time,
				FailureCode:         pointer.String(natFailureCode),
				FailureMessage:      pointer.String(natFailureMessage),
				NatGatewayAddresses: natAddresses(),
				NatGatewayId:        pointer.String(natGatewayID),
				State:               v1beta1.NatGatewayStatusFailed,
				SubnetId:            pointer.String(natSubnetID),
				Tags:                natTags(),
				VpcId:               pointer.String(natVpcID),
			},
			out: v1beta1.NATGatewayObservation{
				CreateTime:          &v1.Time{Time: time},
				DeleteTime:          &v1.Time{Time: time},
				FailureCode:         natFailureCode,
				FailureMessage:      natFailureMessage,
				NatGatewayAddresses: specAddresses(),
				NatGatewayID:        natGatewayID,
				State:               v1beta1.NatGatewayStatusFailed,
				VpcID:               natVpcID,
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			r := GenerateNATGatewayObservation(tc.in)
			if diff := cmp.Diff(r, tc.out); diff != "" {
				t.Errorf("GenerateNATGatewayObservation(...): -want, +got:\n%s", diff)
			}
		})
	}
}
