package ec2

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/google/go-cmp/cmp"

	"github.com/crossplane/provider-aws/apis/ec2/v1alpha1"
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
	ec2tag                  = ec2.Tag{Key: &testKey, Value: &testValue}
	beta1tag                = v1beta1.Tag{Key: testKey, Value: testValue}
)

func TestGenerateElasticIPObservation(t *testing.T) {
	cases := map[string]struct {
		in  ec2.Address
		out v1alpha1.ElasticIPObservation
	}{
		"AllFilled": {
			in: ec2.Address{
				AllocationId:            aws.String(allocationID),
				AssociationId:           aws.String(associationID),
				CustomerOwnedIp:         aws.String(testIPAddress),
				CustomerOwnedIpv4Pool:   aws.String(poolName),
				Domain:                  ec2.DomainType(domain),
				InstanceId:              aws.String(instanceID),
				NetworkBorderGroup:      aws.String(networkBorderGroup),
				NetworkInterfaceId:      aws.String(networkInterfaceID),
				NetworkInterfaceOwnerId: aws.String(networkInterfaceOwnerID),
				PrivateIpAddress:        aws.String(testIPAddress),
				PublicIp:                aws.String(testIPAddress),
				PublicIpv4Pool:          aws.String(poolName),
				Tags:                    []ec2.Tag{ec2tag},
			},
			out: v1alpha1.ElasticIPObservation{
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
			r := GenerateElasticIPObservation(tc.in)
			if diff := cmp.Diff(tc.out, r); diff != "" {
				t.Errorf("GenerateElasticIPObservation(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestIsEIPUpToDate(t *testing.T) {
	type args struct {
		eip ec2.Address
		e   v1alpha1.ElasticIPParameters
	}

	cases := map[string]struct {
		args args
		want bool
	}{
		"SameFields": {
			args: args{
				eip: ec2.Address{
					Tags: []ec2.Tag{ec2tag},
				},
				e: v1alpha1.ElasticIPParameters{
					Tags: []v1beta1.Tag{beta1tag},
				},
			},
			want: true,
		},
		"DifferentFields": {
			args: args{
				eip: ec2.Address{
					Tags: []ec2.Tag{},
				},
				e: v1alpha1.ElasticIPParameters{
					Tags: []v1beta1.Tag{beta1tag},
				},
			},
			want: false,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := IsElasticIPUpToDate(tc.args.e, tc.args.eip)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}
