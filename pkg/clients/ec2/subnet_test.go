package ec2

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/google/go-cmp/cmp"

	"github.com/crossplane/provider-aws/apis/network/v1beta1"
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
		subnet ec2.Subnet
		p      v1beta1.SubnetParameters
	}

	cases := map[string]struct {
		args args
		want bool
	}{
		"SameFields": {
			args: args{
				subnet: ec2.Subnet{
					CidrBlock:                   aws.String(cidr),
					VpcId:                       aws.String(vpc),
					AssignIpv6AddressOnCreation: aws.Bool(true),
					MapPublicIpOnLaunch:         aws.Bool(true),
				},
				p: v1beta1.SubnetParameters{
					CIDRBlock:                   cidr,
					VPCID:                       vpc,
					AssignIpv6AddressOnCreation: aws.Bool(true),
					MapPublicIPOnLaunch:         aws.Bool(true),
				},
			},
			want: true,
		},
		"DifferentFields": {
			args: args{
				subnet: ec2.Subnet{
					CidrBlock:                   aws.String(cidr),
					VpcId:                       aws.String(vpc),
					AssignIpv6AddressOnCreation: aws.Bool(true),
					MapPublicIpOnLaunch:         aws.Bool(false),
				},
				p: v1beta1.SubnetParameters{
					CIDRBlock:                   cidr,
					VPCID:                       vpc,
					AssignIpv6AddressOnCreation: aws.Bool(true),
					MapPublicIPOnLaunch:         aws.Bool(true),
				},
			},
			want: false,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got, _ := IsSubnetUpToDate(tc.args.p, tc.args.subnet)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGenerateSubnetObservation(t *testing.T) {
	cases := map[string]struct {
		in  ec2.Subnet
		out v1beta1.SubnetObservation
	}{
		"AllFilled": {
			in: ec2.Subnet{
				AvailableIpAddressCount: aws.Int64(int64(availableIPCount)),
				DefaultForAz:            aws.Bool(true),
				SubnetId:                aws.String(subnetID),
				State:                   ec2.SubnetStateAvailable,
			},
			out: v1beta1.SubnetObservation{
				AvailableIPAddressCount: int64(availableIPCount),
				DefaultForAz:            true,
				SubnetID:                subnetID,
				SubnetState:             state,
			},
		},
		"NoIpCount": {
			in: ec2.Subnet{
				DefaultForAz: aws.Bool(true),
				SubnetId:     aws.String(subnetID),
				State:        ec2.SubnetStateAvailable,
			},
			out: v1beta1.SubnetObservation{
				DefaultForAz: true,
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

func TestCreateSubnetPatch(t *testing.T) {
	type args struct {
		subnet *ec2.Subnet
		p      *v1beta1.SubnetParameters
	}

	type want struct {
		patch *v1beta1.SubnetParameters
	}

	cases := map[string]struct {
		args
		want
	}{
		"SameFields": {
			args: args{
				subnet: &ec2.Subnet{
					CidrBlock:                   aws.String(cidr),
					VpcId:                       aws.String(vpc),
					AssignIpv6AddressOnCreation: aws.Bool(true),
					MapPublicIpOnLaunch:         aws.Bool(true),
				},
				p: &v1beta1.SubnetParameters{
					CIDRBlock:                   cidr,
					VPCID:                       vpc,
					AssignIpv6AddressOnCreation: aws.Bool(true),
					MapPublicIPOnLaunch:         aws.Bool(true),
				},
			},
			want: want{
				patch: &v1beta1.SubnetParameters{},
			},
		},
		"DifferentFields": {
			args: args{
				subnet: &ec2.Subnet{
					CidrBlock:                   aws.String(cidr),
					VpcId:                       aws.String(vpc),
					AssignIpv6AddressOnCreation: aws.Bool(true),
					MapPublicIpOnLaunch:         aws.Bool(false),
				},
				p: &v1beta1.SubnetParameters{
					CIDRBlock:                   cidr,
					VPCID:                       vpc,
					AssignIpv6AddressOnCreation: aws.Bool(true),
					MapPublicIPOnLaunch:         aws.Bool(true),
				},
			},
			want: want{
				patch: &v1beta1.SubnetParameters{
					MapPublicIPOnLaunch: aws.Bool(true),
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			result, _ := CreateSubnetPatch(tc.args.subnet, tc.args.p)
			if diff := cmp.Diff(tc.want.patch, result); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}
