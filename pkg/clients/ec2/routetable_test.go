package ec2

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/google/go-cmp/cmp"

	"github.com/crossplane/provider-aws/apis/network/v1beta1"
)

var (
	rtVPC      = "some vpc"
	otherRtVPC = "some other vpc"
	rtID       = "some RT Id"
	rtSubnetID = "some subnet"
	rtOwner    = "some owner"
)

func specAssociations() []v1beta1.Association {
	return []v1beta1.Association{
		{
			SubnetID: rtSubnetID,
		},
	}
}

func rtAssociations() []ec2.RouteTableAssociation {
	return []ec2.RouteTableAssociation{
		{
			SubnetId: aws.String(rtSubnetID),
		},
	}
}

func TestIsRTUpToDate(t *testing.T) {
	type args struct {
		rt ec2.RouteTable
		p  v1beta1.RouteTableParameters
	}

	cases := map[string]struct {
		args args
		want bool
	}{
		"SameFields": {
			args: args{
				rt: ec2.RouteTable{
					VpcId: aws.String(rtVPC),
				},
				p: v1beta1.RouteTableParameters{
					VPCID: rtVPC,
				},
			},
			want: true,
		},
		"DifferentFields": {
			args: args{
				rt: ec2.RouteTable{
					VpcId: aws.String(rtVPC),
				},
				p: v1beta1.RouteTableParameters{
					VPCID: otherRtVPC,
				},
			},
			want: false,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got, _ := IsRtUpToDate(tc.args.p, tc.args.rt)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGenerateRTObservation(t *testing.T) {
	cases := map[string]struct {
		in  ec2.RouteTable
		out v1beta1.RouteTableObservation
	}{
		"AllFilled": {
			in: ec2.RouteTable{
				OwnerId:      aws.String(rtOwner),
				RouteTableId: aws.String(rtID),
			},
			out: v1beta1.RouteTableObservation{
				OwnerID:      rtOwner,
				RouteTableID: rtID,
			},
		},
		"NoOwnerID": {
			in: ec2.RouteTable{
				RouteTableId: aws.String(rtID),
			},
			out: v1beta1.RouteTableObservation{
				RouteTableID: rtID,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			r := GenerateRTObservation(tc.in)
			if diff := cmp.Diff(r, tc.out); diff != "" {
				t.Errorf("GenerateNetworkObservation(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestCreateRTPatch(t *testing.T) {
	type args struct {
		rt ec2.RouteTable
		p  *v1beta1.RouteTableParameters
	}

	type want struct {
		patch *v1beta1.RouteTableParameters
	}

	cases := map[string]struct {
		args
		want
	}{
		"SameFields": {
			args: args{
				rt: ec2.RouteTable{
					Associations: rtAssociations(),
					VpcId:        aws.String(vpcID),
				},
				p: &v1beta1.RouteTableParameters{
					Associations: specAssociations(),
					VPCID:        rtVPC,
				},
			},
			want: want{
				patch: &v1beta1.RouteTableParameters{},
			},
		},
		"DifferentFields": {
			args: args{
				rt: ec2.RouteTable{
					Associations: rtAssociations(),
					VpcId:        aws.String(rtVPC),
				},
				p: &v1beta1.RouteTableParameters{
					Associations: specAssociations(),
					VPCID:        otherRtVPC,
				},
			},
			want: want{
				patch: &v1beta1.RouteTableParameters{
					VPCID: otherRtVPC,
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			result, _ := CreateRTPatch(tc.args.rt, *tc.args.p)
			if diff := cmp.Diff(tc.want.patch, result); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}
