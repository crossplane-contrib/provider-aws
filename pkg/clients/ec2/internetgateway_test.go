package ec2

import (
	"testing"

	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/google/go-cmp/cmp"

	"github.com/crossplane/provider-aws/apis/ec2/v1beta1"
	aws "github.com/crossplane/provider-aws/pkg/clients"
)

var (
	vpcID   = "some vpc"
	igID    = "some id"
	ownerID = "some owner"
)

func igAttachments() []ec2types.InternetGatewayAttachment {
	return []ec2types.InternetGatewayAttachment{
		{
			VpcId: aws.String(vpcID),
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
					InternetGatewayId: aws.String(igID),
				},
				p: v1beta1.InternetGatewayParameters{
					VPCID: aws.String(vpcID),
				},
			},
			want: true,
		},
		"DifferentFields": {
			args: args{
				ig: ec2types.InternetGateway{
					Attachments:       igAttachments(),
					InternetGatewayId: aws.String(igID),
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
				InternetGatewayId: aws.String(igID),
				OwnerId:           aws.String(ownerID),
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
				InternetGatewayId: aws.String(igID),
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
