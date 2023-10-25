package ec2

import (
	"testing"

	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/google/go-cmp/cmp"

	"github.com/crossplane-contrib/provider-aws/apis/ec2/v1beta1"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
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
				AllocationId:            pointer.ToOrNilIfZeroValue(allocationID),
				AssociationId:           pointer.ToOrNilIfZeroValue(associationID),
				CustomerOwnedIp:         pointer.ToOrNilIfZeroValue(testIPAddress),
				CustomerOwnedIpv4Pool:   pointer.ToOrNilIfZeroValue(poolName),
				Domain:                  ec2types.DomainType(domain),
				InstanceId:              pointer.ToOrNilIfZeroValue(instanceID),
				NetworkBorderGroup:      pointer.ToOrNilIfZeroValue(networkBorderGroup),
				NetworkInterfaceId:      pointer.ToOrNilIfZeroValue(networkInterfaceID),
				NetworkInterfaceOwnerId: pointer.ToOrNilIfZeroValue(networkInterfaceOwnerID),
				PrivateIpAddress:        pointer.ToOrNilIfZeroValue(testIPAddress),
				PublicIp:                pointer.ToOrNilIfZeroValue(testIPAddress),
				PublicIpv4Pool:          pointer.ToOrNilIfZeroValue(poolName),
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
