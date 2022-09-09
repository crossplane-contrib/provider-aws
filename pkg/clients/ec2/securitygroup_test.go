package ec2

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/google/go-cmp/cmp"

	"github.com/crossplane-contrib/provider-aws/apis/ec2/v1beta1"
)

var (
	sgID       = "some id"
	sgVpc      = "some vpc"
	sgDesc     = "some description"
	sgName     = "some name"
	sgProtocol = "tcp"
	sgCidr     = "192.168.0.0/32"
	sgOwner    = "some owner"
)

func specIPPermission(ports ...int) (ret []v1beta1.IPPermission) {
	for _, port := range ports {
		ret = append(ret, v1beta1.IPPermission{
			FromPort:   aws.Int32(int32(port)),
			ToPort:     aws.Int32(int32(port)),
			IPProtocol: "tcp",
			IPRanges: []v1beta1.IPRange{
				{
					CIDRIP: sgCidr,
				},
			},
		})
	}
	return ret
}

func sgIPPermission(ports ...int) (ret []ec2types.IpPermission) {
	for _, port := range ports {
		ret = append(ret, ec2types.IpPermission{
			FromPort:   aws.Int32(int32(port)),
			ToPort:     aws.Int32(int32(port)),
			IpProtocol: aws.String(sgProtocol),
			IpRanges: []ec2types.IpRange{
				{
					CidrIp: aws.String(sgCidr),
				},
			},
		})
	}
	return ret
}

func TestIsSGUpToDate(t *testing.T) {
	type args struct {
		sg ec2types.SecurityGroup
		p  v1beta1.SecurityGroupParameters
	}

	cases := map[string]struct {
		args args
		want bool
	}{
		"SameFields": {
			args: args{
				sg: ec2types.SecurityGroup{
					Description:   aws.String(sgDesc),
					GroupName:     aws.String(sgName),
					VpcId:         aws.String(sgVpc),
					IpPermissions: sgIPPermission(80),
				},
				p: v1beta1.SecurityGroupParameters{
					Description: sgDesc,
					GroupName:   sgName,
					VPCID:       aws.String(sgVpc),
					Ingress:     specIPPermission(80),
				},
			},
			want: true,
		},
		"SameFieldsUnsorted": {
			args: args{
				sg: ec2types.SecurityGroup{
					Description:   aws.String(sgDesc),
					GroupName:     aws.String(sgName),
					VpcId:         aws.String(sgVpc),
					IpPermissions: sgIPPermission(80, 100, 90),
				},
				p: v1beta1.SecurityGroupParameters{
					Description: sgDesc,
					GroupName:   sgName,
					VPCID:       aws.String(sgVpc),
					Ingress:     specIPPermission(100, 90, 80),
				},
			},
			want: true,
		},
		"DifferentFields": {
			args: args{
				sg: ec2types.SecurityGroup{
					Description:   aws.String(sgDesc),
					GroupName:     aws.String(sgName),
					VpcId:         aws.String(sgVpc),
					IpPermissions: sgIPPermission(80),
				},
				p: v1beta1.SecurityGroupParameters{
					Description: sgDesc,
					GroupName:   sgName,
					VPCID:       aws.String(sgVpc),
					Ingress:     specIPPermission(100),
				},
			},
			want: false,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := IsSGUpToDate(tc.args.p, tc.args.sg)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGenerateSGObservation(t *testing.T) {
	cases := map[string]struct {
		in  ec2types.SecurityGroup
		out v1beta1.SecurityGroupObservation
	}{
		"AllFilled": {
			in: ec2types.SecurityGroup{
				OwnerId: aws.String(sgOwner),
				GroupId: aws.String(sgID),
			},
			out: v1beta1.SecurityGroupObservation{
				OwnerID:         sgOwner,
				SecurityGroupID: sgID,
			},
		},
		"NoIpCount": {
			in: ec2types.SecurityGroup{
				OwnerId: aws.String(sgOwner),
			},
			out: v1beta1.SecurityGroupObservation{
				OwnerID: sgOwner,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			r := GenerateSGObservation(tc.in)
			if diff := cmp.Diff(r, tc.out); diff != "" {
				t.Errorf("GenerateNetworkObservation(...): -want, +got:\n%s", diff)
			}
		})
	}
}
