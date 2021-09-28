package ec2

import (
	"testing"

	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/google/go-cmp/cmp"

	"github.com/crossplane/provider-aws/apis/ec2/v1beta1"
	aws "github.com/crossplane/provider-aws/pkg/clients"
)

var (
	allocationID            = "allocation"
	associationID           = "association"
	testIPAddress           = "0.0.0.0"
	poolName                = "pool"
	domain                  = "vpc"
	instanceID              = "instance"
	networkBorderGroup      = "border"
	networkInterfaceID      = "network"
	networkInterfaceOwnerID = "owner"
	testKey                 = "key"
	testValue               = "value"

	ec2tag     = ec2types.Tag{Key: &testKey, Value: &testValue}
	v1beta1Tag = v1beta1.Tag{Key: testKey, Value: testValue}
)

func TestGenerateAddressObservation(t *testing.T) {
	cases := map[string]struct {
		in  ec2types.Address
		out v1beta1.AddressObservation
	}{
		"AllFilled": {
			in: ec2types.Address{
				AllocationId:            aws.String(allocationID),
				AssociationId:           aws.String(associationID),
				CustomerOwnedIp:         aws.String(testIPAddress),
				CustomerOwnedIpv4Pool:   aws.String(poolName),
				Domain:                  ec2types.DomainType(domain),
				InstanceId:              aws.String(instanceID),
				NetworkBorderGroup:      aws.String(networkBorderGroup),
				NetworkInterfaceId:      aws.String(networkInterfaceID),
				NetworkInterfaceOwnerId: aws.String(networkInterfaceOwnerID),
				PrivateIpAddress:        aws.String(testIPAddress),
				PublicIp:                aws.String(testIPAddress),
				PublicIpv4Pool:          aws.String(poolName),
				Tags:                    []ec2types.Tag{ec2tag},
			},
			out: v1beta1.AddressObservation{
				AllocationID:            allocationID,
				AssociationID:           associationID,
				CustomerOwnedIP:         testIPAddress,
				CustomerOwnedIPv4Pool:   poolName,
				InstanceID:              instanceID,
				NetworkBorderGroup:      networkBorderGroup,
				NetworkInterfaceID:      networkInterfaceID,
				NetworkInterfaceOwnerID: networkInterfaceOwnerID,
				PrivateIPAddress:        testIPAddress,
				PublicIP:                testIPAddress,
				PublicIPv4Pool:          poolName,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			r := GenerateAddressObservation(tc.in)
			if diff := cmp.Diff(tc.out, r); diff != "" {
				t.Errorf("GenerateAddressObservation(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestIsEIPUpToDate(t *testing.T) {
	type args struct {
		eip ec2types.Address
		e   v1beta1.AddressParameters
	}

	cases := map[string]struct {
		args args
		want bool
	}{
		"SameFields": {
			args: args{
				eip: ec2types.Address{
					Tags: []ec2types.Tag{ec2tag},
				},
				e: v1beta1.AddressParameters{
					Tags: []v1beta1.Tag{v1beta1Tag},
				},
			},
			want: true,
		},
		"DifferentFields": {
			args: args{
				eip: ec2types.Address{
					Tags: []ec2types.Tag{},
				},
				e: v1beta1.AddressParameters{
					Tags: []v1beta1.Tag{v1beta1Tag},
				},
			},
			want: false,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := IsAddressUpToDate(tc.args.e, tc.args.eip)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}
