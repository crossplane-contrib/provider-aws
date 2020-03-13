package ec2

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/awserr"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/google/go-cmp/cmp"

	"github.com/crossplane/provider-aws/apis/network/v1alpha3"
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
}

// NewVPCClient returns a new client using AWS credentials as JSON encoded data.
func NewVPCClient(cfg *aws.Config) (VPCClient, error) {
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

// IsUpToDate returns true if there is no update-able difference between desired
// and observed state of the resource.
func IsUpToDate(spec v1alpha3.VPCParameters, o ec2.Vpc) bool {
	actual := v1alpha3.BuildFromEC2Tags(o.Tags)
	return cmp.Equal(spec.Tags, actual)
}
