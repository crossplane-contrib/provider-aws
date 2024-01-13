package ec2

import (
	"testing"

	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/google/go-cmp/cmp"

	"github.com/crossplane-contrib/provider-aws/apis/ec2/v1beta1"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
)

var (
	boolFalse         = false
	vpcOwner          = "some owner"
	vpcStateAvailable = "available"
)

func TestGenerateVPCObservation(t *testing.T) {
	cases := map[string]struct {
		in  ec2types.Vpc
		out v1beta1.VPCObservation
	}{
		"AllFilled": {
			in: ec2types.Vpc{
				IsDefault: &boolFalse,
				OwnerId:   pointer.ToOrNilIfZeroValue(vpcOwner),
				VpcId:     pointer.ToOrNilIfZeroValue(vpcID),
				State:     ec2types.VpcStateAvailable,
			},
			out: v1beta1.VPCObservation{
				IsDefault: boolFalse,
				OwnerID:   vpcOwner,
				VPCState:  vpcStateAvailable,
				VPCID:     vpcID,
			},
		},
		"NoOwner": {
			in: ec2types.Vpc{
				IsDefault: &boolFalse,
				VpcId:     pointer.ToOrNilIfZeroValue(vpcID),
				State:     ec2types.VpcStateAvailable,
			},
			out: v1beta1.VPCObservation{
				IsDefault: boolFalse,
				VPCState:  vpcStateAvailable,
				VPCID:     vpcID,
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
