package ec2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/awserr"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/google/go-cmp/cmp"

	"github.com/crossplane/provider-aws/apis/network/v1beta1"
	awsclients "github.com/crossplane/provider-aws/pkg/clients"
)

const (
	// VPCIDNotFound is the code that is returned by ec2 when the given VPCID is not valid
	VPCIDNotFound = "InvalidVpcID.NotFound"
)

// VPCClient is the external client used for VPC Custom Resource
type VPCClient interface {
	CreateVpcRequest(*ec2.CreateVpcInput) ec2.CreateVpcRequest
	DeleteVpcRequest(*ec2.DeleteVpcInput) ec2.DeleteVpcRequest
	DescribeVpcsRequest(*ec2.DescribeVpcsInput) ec2.DescribeVpcsRequest
	ModifyVpcAttributeRequest(*ec2.ModifyVpcAttributeInput) ec2.ModifyVpcAttributeRequest
	CreateTagsRequest(*ec2.CreateTagsInput) ec2.CreateTagsRequest
	ModifyVpcTenancyRequest(*ec2.ModifyVpcTenancyInput) ec2.ModifyVpcTenancyRequest
}

// NewVpcClient returns a new client using AWS credentials as JSON encoded data.
func NewVpcClient(ctx context.Context, credentials []byte, region string, auth awsclients.AuthMethod) (VPCClient, error) {
	cfg, err := auth(ctx, credentials, awsclients.DefaultSection, region)
	if cfg == nil {
		return nil, err
	}
	return ec2.New(*cfg), nil
}

// IsVPCNotFoundErr returns true if the error is because the item doesn't exist
func IsVPCNotFoundErr(err error) bool {
	if awsErr, ok := err.(awserr.Error); ok {
		if awsErr.Code() == VPCIDNotFound {
			return true
		}
	}

	return false
}

// IsVpcUpToDate returns true if there is no update-able difference between desired
// and observed state of the resource.
func IsVpcUpToDate(spec v1beta1.VPCParameters, o ec2.Vpc) bool {
	actual := v1beta1.BuildFromEC2Tags(o.Tags)
	return cmp.Equal(spec.Tags, actual) && (aws.StringValue(spec.InstanceTenancy) == string(o.InstanceTenancy))
}

// GenerateVpcObservation is used to produce v1beta1.VPCObservation from
// ec2.Vpc.
func GenerateVpcObservation(vpc ec2.Vpc) v1beta1.VPCObservation {
	o := v1beta1.VPCObservation{
		IsDefault: aws.BoolValue(vpc.IsDefault),
		OwnerID:   aws.StringValue(vpc.OwnerId),
		VPCID:     aws.StringValue(vpc.VpcId),
		Tags:      v1beta1.BuildFromEC2Tags(vpc.Tags),
		VPCState:  string(vpc.State),
	}

	return o
}

// LateInitializeVPC fills the empty fields in *v1beta1.VPCParameters with
// the values seen in ec2.Vpc.
func LateInitializeVPC(in *v1beta1.VPCParameters, v *ec2.Vpc) { // nolint:gocyclo
	if v == nil {
		return
	}

	in.CIDRBlock = awsclients.LateInitializeString(in.CIDRBlock, v.CidrBlock)
	in.InstanceTenancy = awsclients.LateInitializeStringPtr(in.InstanceTenancy, aws.String(string(v.InstanceTenancy)))
}
