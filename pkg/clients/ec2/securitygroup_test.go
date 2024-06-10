package ec2

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/google/go-cmp/cmp"
	"k8s.io/utils/ptr"

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
	type args struct {
		sg    ec2types.SecurityGroup
		rules []ec2types.SecurityGroupRule
	}

	cases := map[string]struct {
		in  args
		out v1beta1.SecurityGroupObservation
	}{
		"AllFilled": {
			in: args{
				sg: ec2types.SecurityGroup{
					OwnerId: aws.String(sgOwner),
					GroupId: aws.String(sgID),
				},
				rules: []ec2types.SecurityGroupRule{
					{
						CidrIpv4:    ptr.To("10.0.0.16/32"),
						Description: ptr.To("egress rule"),
						GroupId:     ptr.To("abcd"),
						IpProtocol:  ptr.To("tcp"),
						IsEgress:    ptr.To(true),
					},
					{
						CidrIpv4:    ptr.To("10.0.100.16/16"),
						Description: ptr.To("ingress rule"),
						GroupId:     ptr.To("efgh"),
						IpProtocol:  ptr.To("tcp"),
						IsEgress:    ptr.To(false),
					},
					{
						CidrIpv4:    ptr.To("10.0.0.0/16"),
						Description: ptr.To("ingress rule"),
						FromPort:    aws.Int32(int32(8080)),
						ToPort:      aws.Int32(int32(8443)),
						GroupId:     ptr.To("efgh"),
						IpProtocol:  ptr.To("tcp"),
						IsEgress:    ptr.To(false),
					},
					{
						ReferencedGroupInfo: &ec2types.ReferencedSecurityGroup{
							GroupId: ptr.To("groupId"),
						},
						Description: ptr.To("ingress rule sg"),
						FromPort:    aws.Int32(int32(8080)),
						ToPort:      aws.Int32(int32(8443)),
						GroupId:     ptr.To("efgh"),
						IpProtocol:  ptr.To("tcp"),
						IsEgress:    ptr.To(false),
					},
					{
						PrefixListId: ptr.To("pl-12345676"),
						Description:  ptr.To("ingress rule pl"),
						FromPort:     aws.Int32(int32(8080)),
						ToPort:       aws.Int32(int32(8443)),
						GroupId:      ptr.To("efgh"),
						IpProtocol:   ptr.To("tcp"),
						IsEgress:     ptr.To(false),
					},
				},
			},
			out: v1beta1.SecurityGroupObservation{
				OwnerID:         sgOwner,
				SecurityGroupID: sgID,
				EgressRules: []v1beta1.SecurityGroupRuleObservation{
					{
						CidrIpv4:    ptr.To("10.0.0.16/32"),
						IpProtocol:  ptr.To("tcp"),
						Description: ptr.To("egress rule"),
					},
				},
				IngressRules: []v1beta1.SecurityGroupRuleObservation{
					{
						CidrIpv4:    ptr.To("10.0.100.16/16"),
						IpProtocol:  ptr.To("tcp"),
						Description: ptr.To("ingress rule"),
					},
					{
						CidrIpv4:    ptr.To("10.0.0.0/16"),
						IpProtocol:  ptr.To("tcp"),
						Description: ptr.To("ingress rule"),
						FromPort:    aws.Int32(int32(8080)),
						ToPort:      aws.Int32(int32(8443)),
					},
					{
						IpProtocol:  ptr.To("tcp"),
						Description: ptr.To("ingress rule sg"),
						FromPort:    aws.Int32(int32(8080)),
						ToPort:      aws.Int32(int32(8443)),
						ReferencedGroupInfo: &v1beta1.ReferencedSecurityGroup{
							GroupId: ptr.To("groupId"),
						},
					},
					{
						PrefixListId: ptr.To("pl-12345676"),
						IpProtocol:   ptr.To("tcp"),
						Description:  ptr.To("ingress rule pl"),
						FromPort:     aws.Int32(int32(8080)),
						ToPort:       aws.Int32(int32(8443)),
					},
				},
			},
		},
		"NoIpCount": {
			in: args{
				sg: ec2types.SecurityGroup{
					OwnerId: aws.String(sgOwner),
				},
			},
			out: v1beta1.SecurityGroupObservation{
				OwnerID:      sgOwner,
				IngressRules: []v1beta1.SecurityGroupRuleObservation{},
				EgressRules:  []v1beta1.SecurityGroupRuleObservation{},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			r := GenerateSGObservation(tc.in.sg, tc.in.rules)
			if diff := cmp.Diff(r, tc.out); diff != "" {
				t.Errorf("GenerateNetworkObservation(...): -want, +got:\n%s", diff)
			}
		})
	}
}
