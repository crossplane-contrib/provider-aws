package ec2

import (
	"testing"

	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/google/go-cmp/cmp"

	"github.com/crossplane-contrib/provider-aws/apis/ec2/v1beta1"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
)

var (
	vpcID   = "some vpc"
	igID    = "some id"
	ownerID = "some owner"
)

func igAttachments() []ec2types.InternetGatewayAttachment {
	return []ec2types.InternetGatewayAttachment{
		{
			VpcId: pointer.ToOrNilIfZeroValue(vpcID),
			State: ec2types.AttachmentStatusAttached,
		},
	}
}

func specAttachments() []v1beta1.InternetGatewayAttachment {
	return []v1beta1.InternetGatewayAttachment{
		{
			AttachmentStatus: string(ec2types.AttachmentStatusAttached),
			VPCID:            vpcID,
		},
	}
}
func TestIsIGUpToDate(t *testing.T) {
	type args struct {
		ig ec2types.InternetGateway
		p  v1beta1.InternetGatewayParameters
	}

	cases := map[string]struct {
		args args
		want bool
	}{
		"SameFields": {
			args: args{
				ig: ec2types.InternetGateway{
					Attachments:       igAttachments(),
					InternetGatewayId: pointer.ToOrNilIfZeroValue(igID),
				},
				p: v1beta1.InternetGatewayParameters{
					VPCID: pointer.ToOrNilIfZeroValue(vpcID),
				},
			},
			want: true,
		},
		"DifferentFields": {
			args: args{
				ig: ec2types.InternetGateway{
					Attachments:       igAttachments(),
					InternetGatewayId: pointer.ToOrNilIfZeroValue(igID),
				},
				p: v1beta1.InternetGatewayParameters{},
			},
			want: false,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := IsIgUpToDate(tc.args.p, tc.args.ig)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGenerateIGObservation(t *testing.T) {
	cases := map[string]struct {
		in  ec2types.InternetGateway
		out v1beta1.InternetGatewayObservation
	}{
		"AllFilled": {
			in: ec2types.InternetGateway{
				Attachments:       igAttachments(),
				InternetGatewayId: pointer.ToOrNilIfZeroValue(igID),
				OwnerId:           pointer.ToOrNilIfZeroValue(ownerID),
			},
			out: v1beta1.InternetGatewayObservation{
				Attachments:       specAttachments(),
				InternetGatewayID: igID,
				OwnerID:           ownerID,
			},
		},
		"NoOwnerId": {
			in: ec2types.InternetGateway{
				Attachments:       igAttachments(),
				InternetGatewayId: pointer.ToOrNilIfZeroValue(igID),
			},
			out: v1beta1.InternetGatewayObservation{
				Attachments:       specAttachments(),
				InternetGatewayID: igID,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			r := GenerateIGObservation(tc.in)
			if diff := cmp.Diff(r, tc.out); diff != "" {
				t.Errorf("GenerateNetworkObservation(...): -want, +got:\n%s", diff)
			}
		})
	}
}
