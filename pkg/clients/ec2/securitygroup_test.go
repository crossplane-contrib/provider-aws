package ec2

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/smithy-go/document"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/crossplane/provider-aws/apis/ec2/v1beta1"
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

func TestLateInitializeSG(t *testing.T) {
	type args struct {
		in *v1beta1.SecurityGroupParameters
		sg *ec2types.SecurityGroup
	}

	cases := map[string]struct {
		args args
		want *v1beta1.SecurityGroupParameters
	}{
		"NilSG": {
			args: args{
				in: &v1beta1.SecurityGroupParameters{
					Tags:        []v1beta1.Tag{{Key: "k", Value: "v"}},
					Description: sgDesc,
					GroupName:   sgName,
					VPCID:       aws.String(sgVpc),
					Ingress:     specIPPermission(80),
				},
			},
			want: &v1beta1.SecurityGroupParameters{
				Tags:        []v1beta1.Tag{{Key: "k", Value: "v"}},
				Description: sgDesc,
				GroupName:   sgName,
				VPCID:       aws.String(sgVpc),
				Ingress:     specIPPermission(80),
			},
		},
		"NoOp": {
			args: args{
				in: &v1beta1.SecurityGroupParameters{
					Tags:        []v1beta1.Tag{{Key: "k", Value: "v"}},
					Description: sgDesc,
					GroupName:   sgName,
					VPCID:       aws.String(sgVpc),
					Ingress:     specIPPermission(80),
				},
				sg: &ec2types.SecurityGroup{
					Tags:          []ec2types.Tag{{Key: aws.String("k"), Value: aws.String("v")}},
					Description:   aws.String(sgDesc),
					GroupName:     aws.String(sgName),
					VpcId:         aws.String(sgVpc),
					IpPermissions: sgIPPermission(80),
				},
			},
			want: &v1beta1.SecurityGroupParameters{
				Tags:        []v1beta1.Tag{{Key: "k", Value: "v"}},
				Description: sgDesc,
				GroupName:   sgName,
				VPCID:       aws.String(sgVpc),
				Ingress:     specIPPermission(80),
			},
		},
		"EmptySG": {
			args: args{
				in: &v1beta1.SecurityGroupParameters{},
				sg: &ec2types.SecurityGroup{
					Tags:                []ec2types.Tag{{Key: aws.String("k"), Value: aws.String("v")}},
					Description:         aws.String(sgDesc),
					GroupName:           aws.String(sgName),
					VpcId:               aws.String(sgVpc),
					IpPermissions:       sgIPPermission(80),
					IpPermissionsEgress: sgIPPermission(80),
				},
			},
			want: &v1beta1.SecurityGroupParameters{
				Tags:        []v1beta1.Tag{{Key: "k", Value: "v"}},
				Description: sgDesc,
				GroupName:   sgName,
				VPCID:       aws.String(sgVpc),
				Ingress:     specIPPermission(80),
				Egress:      specIPPermission(80),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			LateInitializeSG(tc.args.in, tc.args.sg)
			if diff := cmp.Diff(tc.want, tc.args.in); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
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
					Tags:          []ec2types.Tag{{Key: aws.String("k"), Value: aws.String("v")}},
					Description:   aws.String(sgDesc),
					GroupName:     aws.String(sgName),
					VpcId:         aws.String(sgVpc),
					IpPermissions: sgIPPermission(80),
				},
				p: v1beta1.SecurityGroupParameters{
					Tags:        []v1beta1.Tag{{Key: "k", Value: "v"}},
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
		"DifferentTags": {
			args: args{
				sg: ec2types.SecurityGroup{
					Description: aws.String(sgDesc),
					GroupName:   aws.String(sgName),
					VpcId:       aws.String(sgVpc),
				},
				p: v1beta1.SecurityGroupParameters{
					Tags:        []v1beta1.Tag{{Key: "k", Value: "v"}},
					Ingress:     specIPPermission(100),
					Description: sgDesc,
					GroupName:   sgName,
					VPCID:       aws.String(sgVpc),
				},
			},
			want: false,
		},
		"DifferentIngress": {
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
		"DifferentEgress": {
			args: args{
				sg: ec2types.SecurityGroup{
					Description:         aws.String(sgDesc),
					GroupName:           aws.String(sgName),
					VpcId:               aws.String(sgVpc),
					IpPermissions:       sgIPPermission(80),
					IpPermissionsEgress: sgIPPermission(80),
				},
				p: v1beta1.SecurityGroupParameters{
					Description: sgDesc,
					GroupName:   sgName,
					VPCID:       aws.String(sgVpc),
					Ingress:     specIPPermission(80),
					Egress:      specIPPermission(100),
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

func TestBuildFromEC2Permissions(t *testing.T) {
	cases := map[string]struct {
		in   []ec2types.IpPermission
		want []v1beta1.IPPermission
	}{
		"NilPermissions": {
			in:   nil,
			want: nil,
		},
		"FullyPopulated": {
			in: []ec2types.IpPermission{{
				FromPort:   aws.Int32(80),
				IpProtocol: aws.String("tcp"),
				ToPort:     aws.Int32(80),
				IpRanges: []ec2types.IpRange{{
					CidrIp:      aws.String("10.0.0.0/8"),
					Description: aws.String("Only the finest IP addresses."),
				}},
				Ipv6Ranges: []ec2types.Ipv6Range{{
					CidrIpv6:    aws.String("2001:db8:1234:1a00::/64"),
					Description: aws.String("Only the finest IPv6 addresses."),
				}},
				PrefixListIds: []ec2types.PrefixListId{{
					PrefixListId: aws.String("really-good-prefix"),
					Description:  aws.String("Only the finest prefixes."),
				}},
				UserIdGroupPairs: []ec2types.UserIdGroupPair{{
					Description:            aws.String("Only the finest pairs."),
					GroupId:                aws.String("really-good-group-id"),
					GroupName:              aws.String("really-good-group"),
					UserId:                 aws.String("really-good-user-id"),
					VpcId:                  aws.String("really-good-vpc-id"),
					VpcPeeringConnectionId: aws.String("really-good-peering-id"),
				}},
			}},
			want: []v1beta1.IPPermission{{
				FromPort:   aws.Int32(80),
				IPProtocol: "tcp",
				ToPort:     aws.Int32(80),
				IPRanges: []v1beta1.IPRange{{
					CIDRIP:      "10.0.0.0/8",
					Description: aws.String("Only the finest IP addresses."),
				}},
				IPv6Ranges: []v1beta1.IPv6Range{{
					CIDRIPv6:    "2001:db8:1234:1a00::/64",
					Description: aws.String("Only the finest IPv6 addresses."),
				}},
				PrefixListIDs: []v1beta1.PrefixListID{{
					PrefixListID: "really-good-prefix",
					Description:  aws.String("Only the finest prefixes."),
				}},
				UserIDGroupPairs: []v1beta1.UserIDGroupPair{{
					Description:            aws.String("Only the finest pairs."),
					GroupID:                aws.String("really-good-group-id"),
					GroupName:              aws.String("really-good-group"),
					UserID:                 aws.String("really-good-user-id"),
					VPCID:                  aws.String("really-good-vpc-id"),
					VPCPeeringConnectionID: aws.String("really-good-peering-id"),
				}},
			}},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := BuildFromEC2Permissions(tc.in)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("BuildFromEC2Permissions(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGenerateEC2Permissions(t *testing.T) {
	cases := map[string]struct {
		in   []v1beta1.IPPermission
		want []ec2types.IpPermission
	}{
		"NilPermissions": {
			in:   nil,
			want: nil,
		},
		"FullyPopulated": {
			in: []v1beta1.IPPermission{{
				FromPort:   aws.Int32(80),
				IPProtocol: "tcp",
				ToPort:     aws.Int32(80),
				IPRanges: []v1beta1.IPRange{{
					CIDRIP:      "10.0.0.0/8",
					Description: aws.String("Only the finest IP addresses."),
				}},
				IPv6Ranges: []v1beta1.IPv6Range{{
					CIDRIPv6:    "2001:db8:1234:1a00::/64",
					Description: aws.String("Only the finest IPv6 addresses."),
				}},
				PrefixListIDs: []v1beta1.PrefixListID{{
					PrefixListID: "really-good-prefix",
					Description:  aws.String("Only the finest prefixes."),
				}},
				UserIDGroupPairs: []v1beta1.UserIDGroupPair{{
					Description:            aws.String("Only the finest pairs."),
					GroupID:                aws.String("really-good-group-id"),
					GroupName:              aws.String("really-good-group"),
					UserID:                 aws.String("really-good-user-id"),
					VPCID:                  aws.String("really-good-vpc-id"),
					VPCPeeringConnectionID: aws.String("really-good-peering-id"),
				}},
			}},
			want: []ec2types.IpPermission{{
				FromPort:   aws.Int32(80),
				IpProtocol: aws.String("tcp"),
				ToPort:     aws.Int32(80),
				IpRanges: []ec2types.IpRange{{
					CidrIp:      aws.String("10.0.0.0/8"),
					Description: aws.String("Only the finest IP addresses."),
				}},
				Ipv6Ranges: []ec2types.Ipv6Range{{
					CidrIpv6:    aws.String("2001:db8:1234:1a00::/64"),
					Description: aws.String("Only the finest IPv6 addresses."),
				}},
				PrefixListIds: []ec2types.PrefixListId{{
					PrefixListId: aws.String("really-good-prefix"),
					Description:  aws.String("Only the finest prefixes."),
				}},
				UserIdGroupPairs: []ec2types.UserIdGroupPair{{
					Description:            aws.String("Only the finest pairs."),
					GroupId:                aws.String("really-good-group-id"),
					GroupName:              aws.String("really-good-group"),
					UserId:                 aws.String("really-good-user-id"),
					VpcId:                  aws.String("really-good-vpc-id"),
					VpcPeeringConnectionId: aws.String("really-good-peering-id"),
				}},
			}},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := GenerateEC2Permissions(tc.in)
			if diff := cmp.Diff(tc.want, got, cmpopts.IgnoreTypes(document.NoSerde{})); diff != "" {
				t.Errorf("GenerateEC2Permissions(...): -want, +got:\n%s", diff)
			}
		})
	}
}
