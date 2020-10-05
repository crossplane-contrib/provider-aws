package ec2

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/google/go-cmp/cmp"

	"github.com/crossplane/provider-aws/apis/ec2/v1beta1"
	aws "github.com/crossplane/provider-aws/pkg/clients"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

func natTags() []ec2.Tag {
	return []ec2.Tag{
		{
			Key:   aws.String("key1"),
			Value: aws.String("value1"),
		},
		{
			Key:   aws.String("key2"),
			Value: aws.String("value2"),
		},
	}
}

func specTags() []v1beta1.Tag {
	return []v1beta1.Tag{
		{
			Key:   "key1",
			Value: "value1",
		},
		{
			Key:   "key2",
			Value: "value2",
		},
	}
}

func natAddresses() []ec2.NatGatewayAddress {
	return []ec2.NatGatewayAddress{
		{
			AllocationId:       aws.String(natAllocationID),
			NetworkInterfaceId: aws.String(natNetworkInterfaceID),
			PrivateIp:          aws.String(natPrivateIP),
			PublicIp:           aws.String(natPublicIP),
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
		in  ec2.NatGateway
		out v1beta1.NATGatewayObservation
	}{
		"AllFilled": {
			in: ec2.NatGateway{
				CreateTime:          &time,
				NatGatewayAddresses: natAddresses(),
				NatGatewayId:        aws.String(natGatewayID),
				State:               v1beta1.NatGatewayStatusAvailable,
				SubnetId:            aws.String(natSubnetID),
				Tags:                natTags(),
				VpcId:               aws.String(natVpcID),
			},
			out: v1beta1.NATGatewayObservation{
				CreateTime:          &v1.Time{Time: time},
				NatGatewayAddresses: specAddresses(),
				NatGatewayID:        natGatewayID,
				State:               v1beta1.NatGatewayStatusAvailable,
				SubnetID:            natSubnetID,
				Tags:                specTags(),
				VpcID:               natVpcID,
			},
		},
		"DeleteTime": {
			in: ec2.NatGateway{
				CreateTime:          &time,
				DeleteTime:          &time,
				NatGatewayAddresses: natAddresses(),
				NatGatewayId:        aws.String(natGatewayID),
				State:               v1beta1.NatGatewayStatusPending,
				SubnetId:            aws.String(natSubnetID),
				Tags:                natTags(),
				VpcId:               aws.String(natVpcID),
			},
			out: v1beta1.NATGatewayObservation{
				CreateTime:          &v1.Time{Time: time},
				DeleteTime:          &v1.Time{Time: time},
				NatGatewayAddresses: specAddresses(),
				NatGatewayID:        natGatewayID,
				State:               v1beta1.NatGatewayStatusPending,
				SubnetID:            natSubnetID,
				Tags:                specTags(),
				VpcID:               natVpcID,
			},
		},
		"stateFailed": {
			in: ec2.NatGateway{
				CreateTime:          &time,
				DeleteTime:          &time,
				FailureCode:         aws.String(natFailureCode),
				FailureMessage:      aws.String(natFailureMessage),
				NatGatewayAddresses: natAddresses(),
				NatGatewayId:        aws.String(natGatewayID),
				State:               v1beta1.NatGatewayStatusFailed,
				SubnetId:            aws.String(natSubnetID),
				Tags:                natTags(),
				VpcId:               aws.String(natVpcID),
			},
			out: v1beta1.NATGatewayObservation{
				CreateTime:          &v1.Time{Time: time},
				DeleteTime:          &v1.Time{Time: time},
				FailureCode:         natFailureCode,
				FailureMessage:      natFailureMessage,
				NatGatewayAddresses: specAddresses(),
				NatGatewayID:        natGatewayID,
				State:               v1beta1.NatGatewayStatusFailed,
				SubnetID:            natSubnetID,
				Tags:                specTags(),
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
