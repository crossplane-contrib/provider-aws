package ec2

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/smithy-go"

	"github.com/crossplane-contrib/provider-aws/apis/ec2/v1beta1"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
)

const (
	// VPCIDNotFound is the code that is returned by ec2 when the given VPCID is not valid
	VPCIDNotFound = "InvalidVpcID.NotFound"
)

// VPCClient is the external client used for VPC Custom Resource
type VPCClient interface {
	CreateVpc(ctx context.Context, input *ec2.CreateVpcInput, opts ...func(*ec2.Options)) (*ec2.CreateVpcOutput, error)
	DeleteVpc(ctx context.Context, input *ec2.DeleteVpcInput, opts ...func(*ec2.Options)) (*ec2.DeleteVpcOutput, error)
	DescribeVpcs(ctx context.Context, input *ec2.DescribeVpcsInput, opts ...func(*ec2.Options)) (*ec2.DescribeVpcsOutput, error)
	DescribeVpcAttribute(ctx context.Context, input *ec2.DescribeVpcAttributeInput, opts ...func(*ec2.Options)) (*ec2.DescribeVpcAttributeOutput, error)
	ModifyVpcAttribute(ctx context.Context, input *ec2.ModifyVpcAttributeInput, opts ...func(*ec2.Options)) (*ec2.ModifyVpcAttributeOutput, error)
	CreateTags(ctx context.Context, input *ec2.CreateTagsInput, opts ...func(*ec2.Options)) (*ec2.CreateTagsOutput, error)
	DeleteTags(ctx context.Context, input *ec2.DeleteTagsInput, opts ...func(*ec2.Options)) (*ec2.DeleteTagsOutput, error)
	ModifyVpcTenancy(ctx context.Context, input *ec2.ModifyVpcTenancyInput, opts ...func(*ec2.Options)) (*ec2.ModifyVpcTenancyOutput, error)
}

// NewVPCClient returns a new client using AWS credentials as JSON encoded data.
func NewVPCClient(cfg aws.Config) VPCClient {
	return ec2.NewFromConfig(cfg)
}

// IsVPCNotFoundErr returns true if the error is because the item doesn't exist
func IsVPCNotFoundErr(err error) bool {
	var awsErr smithy.APIError
	return errors.As(err, &awsErr) && awsErr.ErrorCode() == VPCIDNotFound
}

// IsVpcUpToDate returns true if there is no update-able difference between desired
// and observed state of the resource.
func IsVpcUpToDate(spec v1beta1.VPCParameters, vpc ec2types.Vpc, attributes ec2.DescribeVpcAttributeOutput) bool {
	if aws.ToString(spec.InstanceTenancy) != string(vpc.InstanceTenancy) {
		return false
	}

	if aws.ToBool(spec.EnableDNSHostNames) != aws.ToBool(attributes.EnableDnsHostnames.Value) ||
		aws.ToBool(spec.EnableDNSSupport) != aws.ToBool(attributes.EnableDnsSupport.Value) {
		return false
	}

	return CompareTagsV1Beta1(spec.Tags, vpc.Tags)
}

// GenerateVpcObservation is used to produce v1beta1.VPCObservation from
// ec2types.Vpc.
func GenerateVpcObservation(vpc ec2types.Vpc) v1beta1.VPCObservation {
	o := v1beta1.VPCObservation{
		IsDefault:     aws.ToBool(vpc.IsDefault),
		DHCPOptionsID: aws.ToString(vpc.DhcpOptionsId),
		OwnerID:       aws.ToString(vpc.OwnerId),
		VPCState:      string(vpc.State),
		VPCID:         aws.ToString(vpc.VpcId),
	}

	if len(vpc.CidrBlockAssociationSet) > 0 {
		o.CIDRBlockAssociationSet = make([]v1beta1.VPCCIDRBlockAssociation, len(vpc.CidrBlockAssociationSet))
		for i, v := range vpc.CidrBlockAssociationSet {
			o.CIDRBlockAssociationSet[i] = v1beta1.VPCCIDRBlockAssociation{
				AssociationID: aws.ToString(v.AssociationId),
				CIDRBlock:     aws.ToString(v.CidrBlock),
			}
			o.CIDRBlockAssociationSet[i].CIDRBlockState = v1beta1.VPCCIDRBlockState{
				State:         string(v.CidrBlockState.State),
				StatusMessage: aws.ToString(v.CidrBlockState.StatusMessage),
			}
		}
	}

	if len(vpc.Ipv6CidrBlockAssociationSet) > 0 {
		o.IPv6CIDRBlockAssociationSet = make([]v1beta1.VPCIPv6CidrBlockAssociation, len(vpc.Ipv6CidrBlockAssociationSet))
		for i, v := range vpc.Ipv6CidrBlockAssociationSet {
			o.IPv6CIDRBlockAssociationSet[i] = v1beta1.VPCIPv6CidrBlockAssociation{
				AssociationID:      aws.ToString(v.AssociationId),
				IPv6CIDRBlock:      aws.ToString(v.Ipv6CidrBlock),
				IPv6Pool:           aws.ToString(v.Ipv6Pool),
				NetworkBorderGroup: aws.ToString(v.NetworkBorderGroup),
			}
			o.IPv6CIDRBlockAssociationSet[i].IPv6CIDRBlockState = v1beta1.VPCCIDRBlockState{
				State:         string(v.Ipv6CidrBlockState.State),
				StatusMessage: aws.ToString(v.Ipv6CidrBlockState.StatusMessage),
			}
		}
	}

	return o
}

// LateInitializeVPC fills the empty fields in *v1beta1.VPCParameters with
// the values seen in ec2.Vpc and ec2.DescribeVpcAttributeOutput.
func LateInitializeVPC(in *v1beta1.VPCParameters, v *ec2types.Vpc, attributes *ec2.DescribeVpcAttributeOutput) {
	if v == nil {
		return
	}

	in.CIDRBlock = pointer.LateInitializeValueFromPtr(in.CIDRBlock, v.CidrBlock)
	in.InstanceTenancy = pointer.LateInitialize(in.InstanceTenancy, pointer.ToOrNilIfZeroValue(string(v.InstanceTenancy)))
	if len(v.Ipv6CidrBlockAssociationSet) != 0 {
		ipv6Association := v.Ipv6CidrBlockAssociationSet[0]
		in.Ipv6CIDRBlock = pointer.LateInitialize(in.Ipv6CIDRBlock, ipv6Association.Ipv6CidrBlock)
	}
	if attributes.EnableDnsHostnames != nil {
		in.EnableDNSHostNames = pointer.LateInitialize(in.EnableDNSHostNames, attributes.EnableDnsHostnames.Value)
	}
	if attributes.EnableDnsHostnames != nil {
		in.EnableDNSSupport = pointer.LateInitialize(in.EnableDNSSupport, attributes.EnableDnsSupport.Value)
	}
}
