package ec2

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/google/go-cmp/cmp"

	"github.com/crossplane/provider-aws/apis/ec2/v1beta1"
)

var (
	cidr             = "192.18.0.0/32"
	vpc              = "some vpc"
	availableIPCount = 10
	subnetID         = "some subnet"
	state            = "available"
)

func TestIsSubnetUpToDate(t *testing.T) {
	type args struct {
		subnet ec2types.Subnet
		p      v1beta1.SubnetParameters
	}

	cases := map[string]struct {
		args args
		want bool
	}{
		"SameFields": {
			args: args{
				subnet: ec2types.Subnet{
					CidrBlock:                   aws.String(cidr),
					VpcId:                       aws.String(vpc),
					AssignIpv6AddressOnCreation: true,
					MapPublicIpOnLaunch:         true,
				},
				p: v1beta1.SubnetParameters{
					CIDRBlock:                   cidr,
					VPCID:                       aws.String(vpc),
					AssignIPv6AddressOnCreation: aws.Bool(true),
					MapPublicIPOnLaunch:         aws.Bool(true),
				},
			},
			want: true,
		},
		"DifferentFields": {
			args: args{
				subnet: ec2types.Subnet{
					CidrBlock:                   aws.String(cidr),
					VpcId:                       aws.String(vpc),
					AssignIpv6AddressOnCreation: true,
					MapPublicIpOnLaunch:         false,
				},
				p: v1beta1.SubnetParameters{
					CIDRBlock:                   cidr,
					VPCID:                       aws.String(vpc),
					AssignIPv6AddressOnCreation: aws.Bool(true),
					MapPublicIPOnLaunch:         aws.Bool(true),
				},
			},
			want: false,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := IsSubnetUpToDate(tc.args.p, tc.args.subnet)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGenerateSubnetObservation(t *testing.T) {
	cases := map[string]struct {
		in  ec2types.Subnet
		out v1beta1.SubnetObservation
	}{
		"AllFilled": {
			in: ec2types.Subnet{
				AvailableIpAddressCount: int32(availableIPCount),
				DefaultForAz:            true,
				SubnetId:                aws.String(subnetID),
				State:                   ec2types.SubnetStateAvailable,
			},
			out: v1beta1.SubnetObservation{
				AvailableIPAddressCount: int32(availableIPCount),
				DefaultForAZ:            true,
				SubnetID:                subnetID,
				SubnetState:             state,
			},
		},
		"NoIpCount": {
			in: ec2types.Subnet{
				DefaultForAz: true,
				SubnetId:     aws.String(subnetID),
				State:        ec2types.SubnetStateAvailable,
			},
			out: v1beta1.SubnetObservation{
				DefaultForAZ: true,
				SubnetID:     subnetID,
				SubnetState:  state,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			r := GenerateSubnetObservation(tc.in)
			if diff := cmp.Diff(r, tc.out); diff != "" {
				t.Errorf("GenerateNetworkObservation(...): -want, +got:\n%s", diff)
			}
		})
	}
}
