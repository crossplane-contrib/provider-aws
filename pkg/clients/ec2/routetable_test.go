package ec2

import (
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/google/go-cmp/cmp"

	"github.com/crossplane-contrib/provider-aws/apis/ec2/v1beta1"
)

var (
	rtVPC           = "some vpc"
	otherRtVPC      = "some other vpc"
	rtID            = "some RT Id"
	rtSubnetID      = "some subnet"
	rtOwner         = "some owner"
	rtTagName       = "tag1"
	rtTagValue      = "value1"
	otherRtTagName  = "tag2"
	otherRtTagValue = "value2"
)

func specAssociations() []v1beta1.Association {
	return []v1beta1.Association{
		{
			SubnetID: aws.String(rtSubnetID),
		},
	}
}

func rtAssociations() []ec2types.RouteTableAssociation {
	return []ec2types.RouteTableAssociation{
		{
			SubnetId: aws.String(rtSubnetID),
		},
	}
}

func TestIsRTUpToDate(t *testing.T) {
	type args struct {
		rt ec2types.RouteTable
		p  v1beta1.RouteTableParameters
	}

	cases := map[string]struct {
		args args
		want bool
	}{
		"SameFields": {
			args: args{
				rt: ec2types.RouteTable{
					VpcId: aws.String(rtVPC),
				},
				p: v1beta1.RouteTableParameters{
					VPCID: aws.String(rtVPC),
				},
			},
			want: true,
		},
		"DifferentFields": {
			args: args{
				rt: ec2types.RouteTable{
					VpcId: aws.String(rtVPC),
				},
				p: v1beta1.RouteTableParameters{
					VPCID: aws.String(otherRtVPC),
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
		in  ec2types.RouteTable
		out v1beta1.RouteTableObservation
	}{
		"AllFilled": {
			in: ec2types.RouteTable{
				OwnerId:      aws.String(rtOwner),
				RouteTableId: aws.String(rtID),
			},
			out: v1beta1.RouteTableObservation{
				OwnerID:      rtOwner,
				RouteTableID: rtID,
			},
		},
		"NoOwnerID": {
			in: ec2types.RouteTable{
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
		rt ec2types.RouteTable
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
				rt: ec2types.RouteTable{
					Associations: rtAssociations(),
					VpcId:        aws.String(vpcID),
				},
				p: &v1beta1.RouteTableParameters{
					Associations: specAssociations(),
					VPCID:        aws.String(rtVPC),
				},
			},
			want: want{
				patch: &v1beta1.RouteTableParameters{},
			},
		},
		"DifferentFields": {
			args: args{
				rt: ec2types.RouteTable{
					Associations: rtAssociations(),
					VpcId:        aws.String(rtVPC),
				},
				p: &v1beta1.RouteTableParameters{
					Associations: specAssociations(),
					VPCID:        aws.String(otherRtVPC),
				},
			},
			want: want{
				patch: &v1beta1.RouteTableParameters{
					VPCID: aws.String(otherRtVPC),
				},
			},
		},
		"DifferentTagOrder": {
			args: args{
				rt: ec2types.RouteTable{
					Tags:         []ec2types.Tag{{Key: &rtTagName, Value: &rtTagValue}, {Key: &otherRtTagName, Value: &otherRtTagValue}},
					Associations: rtAssociations(),
					VpcId:        aws.String(rtVPC),
				},
				p: &v1beta1.RouteTableParameters{
					Tags:         []v1beta1.Tag{{Key: otherRtTagName, Value: otherRtTagValue}, {Key: rtTagName, Value: rtTagValue}},
					Associations: specAssociations(),
					VPCID:        aws.String(vpcID),
				},
			},
			want: want{
				patch: &v1beta1.RouteTableParameters{},
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

func TestSortRoutes(t *testing.T) {
	type args struct {
		route    []v1beta1.RouteBeta
		ec2Route []ec2types.Route
	}

	type want struct {
		route    []v1beta1.RouteBeta
		ec2Route []ec2types.Route
	}

	var (
		v4RouteMin = aws.String("v4_0")
		v4RouteMax = aws.String("v4_1")
		v6RouteMin = aws.String("v6_0")
		v6RouteMax = aws.String("v6_1")
	)

	cases := map[string]struct {
		args
		want
	}{
		"v4": {
			args: args{
				route: []v1beta1.RouteBeta{
					{
						DestinationCIDRBlock: v4RouteMax,
					},
					{
						DestinationCIDRBlock: v4RouteMin,
					},
				},
				ec2Route: []ec2types.Route{
					{
						DestinationCidrBlock: v4RouteMax,
					},
					{
						DestinationCidrBlock: v4RouteMin,
					},
				},
			},
			want: want{
				route: []v1beta1.RouteBeta{
					{
						DestinationCIDRBlock: v4RouteMin,
					},
					{
						DestinationCIDRBlock: v4RouteMax,
					},
				},
				ec2Route: []ec2types.Route{
					{
						DestinationCidrBlock: v4RouteMin,
					},
					{
						DestinationCidrBlock: v4RouteMax,
					},
				},
			},
		},
		"v6": {
			args: args{
				route: []v1beta1.RouteBeta{
					{
						DestinationIPV6CIDRBlock: v6RouteMax,
					},
					{
						DestinationIPV6CIDRBlock: v6RouteMin,
					},
				},
				ec2Route: []ec2types.Route{
					{
						DestinationCidrBlock: v6RouteMax,
					},
					{
						DestinationCidrBlock: v6RouteMin,
					},
				},
			},
			want: want{
				route: []v1beta1.RouteBeta{
					{
						DestinationIPV6CIDRBlock: v6RouteMin,
					},
					{
						DestinationIPV6CIDRBlock: v6RouteMax,
					},
				},
				ec2Route: []ec2types.Route{
					{
						DestinationCidrBlock: v6RouteMin,
					},
					{
						DestinationCidrBlock: v6RouteMax,
					},
				},
			},
		},
		"both": {
			args: args{
				route: []v1beta1.RouteBeta{
					{
						DestinationCIDRBlock: v4RouteMax,
					},
					{
						DestinationIPV6CIDRBlock: v6RouteMax,
					},
					{
						DestinationCIDRBlock: v4RouteMin,
					},
					{
						DestinationIPV6CIDRBlock: v6RouteMin,
					},
				},
				ec2Route: []ec2types.Route{
					{
						DestinationCidrBlock: v4RouteMax,
					},
					{
						DestinationIpv6CidrBlock: v6RouteMax,
					},
					{
						DestinationCidrBlock: v4RouteMin,
					},
					{
						DestinationIpv6CidrBlock: v6RouteMin,
					},
				},
			},
			want: want{
				route: []v1beta1.RouteBeta{
					{
						DestinationCIDRBlock: v4RouteMin,
					},
					{
						DestinationCIDRBlock: v4RouteMax,
					},
					{
						DestinationIPV6CIDRBlock: v6RouteMin,
					},
					{
						DestinationIPV6CIDRBlock: v6RouteMax,
					},
				},
				ec2Route: []ec2types.Route{
					{
						DestinationCidrBlock: v4RouteMin,
					},
					{
						DestinationCidrBlock: v4RouteMax,
					},
					{
						DestinationIpv6CidrBlock: v6RouteMin,
					},
					{
						DestinationIpv6CidrBlock: v6RouteMax,
					},
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			SortRoutes(tc.args.route, tc.args.ec2Route)
			if !reflect.DeepEqual(tc.args.route, tc.want.route) {
				t.Errorf("SortRoutes() route got = %v, want %v", tc.args.route, tc.want.route)
			}
			if !reflect.DeepEqual(tc.args.ec2Route, tc.want.ec2Route) {
				t.Errorf("SortRoutes() ec2 route got = %v, want %v", tc.args.ec2Route, tc.want.ec2Route)
			}
		})
	}
}

func TestValidateRoutes(t *testing.T) {
	type args struct {
		route []v1beta1.RouteBeta
	}
	tests := map[string]struct {
		args args
		err  string
	}{
		"valid": {
			args: args{
				route: []v1beta1.RouteBeta{{DestinationCIDRBlock: aws.String("0.0.0.0/0")}},
			},
		},
		"empty cidrs": {
			args: args{
				route: []v1beta1.RouteBeta{{}},
			},
			err: "invalid routes: route[0]: both v4 and v6 cidrs are empty",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			err := ValidateRoutes(tt.args.route)
			if err != nil && err.Error() != tt.err {
				t.Errorf("ValidateRoutes() error = %s, wantErr %s", err, err)
			}
			if err == nil && tt.err != "" {
				t.Errorf("ValidateRoutes() error = %s, wantErr %s", err, err)
			}
		})
	}
}
