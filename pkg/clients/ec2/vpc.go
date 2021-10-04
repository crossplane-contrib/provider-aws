package ec2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/smithy-go"

	"github.com/crossplane/provider-aws/apis/ec2/v1beta1"
	awsclients "github.com/crossplane/provider-aws/pkg/clients"
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
	ModifyVpcTenancy(ctx context.Context, input *ec2.ModifyVpcTenancyInput, opts ...func(*ec2.Options)) (*ec2.ModifyVpcTenancyOutput, error)
}

// NewVPCClient returns a new client using AWS credentials as JSON encoded data.
func NewVPCClient(cfg aws.Config) VPCClient {
	return ec2.NewFromConfig(cfg)
}

// IsVPCNotFoundErr returns true if the error is because the item doesn't exist
func IsVPCNotFoundErr(err error) bool {
	if awsErr, ok := err.(smithy.APIError); ok {
		if awsErr.ErrorCode() == VPCIDNotFound {
			return true
		}
	}

	return false
}

// IsVpcUpToDate returns true if there is no update-able difference between desired
// and observed state of the resource.
func IsVpcUpToDate(spec v1beta1.VPCParameters, vpc ec2types.Vpc, attributes ec2.DescribeVpcAttributeOutput) bool {
	if aws.ToString(spec.InstanceTenancy) != string(vpc.InstanceTenancy) {
		return false
	}

	if aws.ToBool(spec.EnableDNSHostNames) != attributes.EnableDnsHostnames.Value ||
		aws.ToBool(spec.EnableDNSSupport) != attributes.EnableDnsSupport.Value {
		return false
	}

	return v1beta1.CompareTags(spec.Tags, vpc.Tags)
}

// GenerateVpcObservation is used to produce v1beta1.VPCObservation from
// ec2types.Vpc.
func GenerateVpcObservation(vpc ec2types.Vpc) v1beta1.VPCObservation {
	o := v1beta1.VPCObservation{
		IsDefault:     vpc.IsDefault,
		DHCPOptionsID: aws.ToString(vpc.DhcpOptionsId),
		OwnerID:       aws.ToString(vpc.OwnerId),
		VPCState:      string(vpc.State),
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
func LateInitializeVPC(in *v1beta1.VPCParameters, v *ec2types.Vpc, attributes *ec2.DescribeVpcAttributeOutput) { // nolint:gocyclo
	if v == nil {
		return
	}

	in.CIDRBlock = awsclients.LateInitializeString(in.CIDRBlock, v.CidrBlock)
	in.InstanceTenancy = awsclients.LateInitializeStringPtr(in.InstanceTenancy, aws.String(string(v.InstanceTenancy)))
	if attributes.EnableDnsHostnames != nil {
		in.EnableDNSHostNames = awsclients.LateInitializeBoolPtr(in.EnableDNSHostNames, aws.Bool(attributes.EnableDnsHostnames.Value))
	}
	if attributes.EnableDnsHostnames != nil {
		in.EnableDNSSupport = awsclients.LateInitializeBoolPtr(in.EnableDNSSupport, aws.Bool(attributes.EnableDnsSupport.Value))
	}
}
