package ec2

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/google/go-cmp/cmp"

	"github.com/crossplane/provider-aws/apis/ec2/v1beta1"
	aws "github.com/crossplane/provider-aws/pkg/clients"
)

var (
	boolFalse         = false
	vpcOwner          = "some owner"
	vpcStateAvailable = "available"
)

func TestGenerateVPCObservation(t *testing.T) {
	cases := map[string]struct {
		in  ec2.Vpc
		out v1beta1.VPCObservation
	}{
		"AllFilled": {
			in: ec2.Vpc{
				IsDefault: aws.Bool(boolFalse),
				OwnerId:   aws.String(vpcOwner),
				VpcId:     aws.String(vpcID),
				State:     ec2.VpcStateAvailable,
			},
			out: v1beta1.VPCObservation{
				IsDefault: boolFalse,
				OwnerID:   vpcOwner,
				VPCState:  vpcStateAvailable,
			},
		},
		"NoOwner": {
			in: ec2.Vpc{
				IsDefault: aws.Bool(boolFalse),
				VpcId:     aws.String(vpcID),
				State:     ec2.VpcStateAvailable,
			},
			out: v1beta1.VPCObservation{
				IsDefault: boolFalse,
				VPCState:  vpcStateAvailable,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			r := GenerateVpcObservation(tc.in)
			if diff := cmp.Diff(tc.out, r); diff != "" {
				t.Errorf("GenerateNetworkObservation(...): -want, +got:\n%s", diff)
			}
		})
	}
}
