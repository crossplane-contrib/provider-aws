package ec2

import (
	"context"
	"errors"

	awsgo "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/smithy-go"

	"github.com/crossplane-contrib/provider-aws/apis/ec2/v1beta1"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
)

const (
	errCIDRAssociationNotFound = "InvalidVpcCidrBlockAssociationID.NotFound"
)

// VPCCIDRBlockClient is the external client used for VPC CIDR Block Custom Resource
type VPCCIDRBlockClient interface {
	DescribeVpcs(ctx context.Context, input *ec2.DescribeVpcsInput, opts ...func(*ec2.Options)) (*ec2.DescribeVpcsOutput, error)
	AssociateVpcCidrBlock(ctx context.Context, input *ec2.AssociateVpcCidrBlockInput, opts ...func(*ec2.Options)) (*ec2.AssociateVpcCidrBlockOutput, error)
	DisassociateVpcCidrBlock(ctx context.Context, input *ec2.DisassociateVpcCidrBlockInput, opts ...func(*ec2.Options)) (*ec2.DisassociateVpcCidrBlockOutput, error)
}

// NewVPCCIDRBlockClient returns a new client using AWS credentials as JSON encoded data.
func NewVPCCIDRBlockClient(cfg awsgo.Config) VPCCIDRBlockClient {
	return ec2.NewFromConfig(cfg)
}

// CIDRNotFoundError will be raised when there is no Association
type CIDRNotFoundError struct{}

// Error satisfies the Error interface for CIDRNotFoundError.
func (r *CIDRNotFoundError) Error() string {
	return errCIDRAssociationNotFound
}

// IsCIDRNotFound returns true if the error code indicates that the CIDR Block Association was not found
func IsCIDRNotFound(err error) bool {
	var notFoundError *CIDRNotFoundError
	if errors.As(err, &notFoundError) {
		return true
	}
	var awsErr smithy.APIError
	return errors.As(err, &awsErr) && awsErr.ErrorCode() == errCIDRAssociationNotFound
}

// IsVpcCidrDeleting returns true if the CIDR Block is already disassociated or disassociating
func IsVpcCidrDeleting(observation v1beta1.VPCCIDRBlockObservation) bool {
	switch {
	case observation.CIDRBlockState.State == string(ec2types.VpcCidrBlockStateCodeDisassociating) || observation.CIDRBlockState.State == string(ec2types.VpcCidrBlockStateCodeDisassociated):
		return true
	case observation.IPv6CIDRBlockState.State == string(ec2types.VpcCidrBlockStateCodeDisassociating) || observation.IPv6CIDRBlockState.State == string(ec2types.VpcCidrBlockStateCodeDisassociated):
		return true
	default:
		return false
	}
}

// GenerateVpcCIDRBlockObservation is used to produce v1alpha1.VPCObservation from
// ec2.Vpc.
func GenerateVpcCIDRBlockObservation(associationID string, vpc ec2types.Vpc) v1beta1.VPCCIDRBlockObservation {
	o := v1beta1.VPCCIDRBlockObservation{}

	IPv4, IPv6 := FindCIDRAssociation(associationID, vpc)

	if IPv4 != nil {
		o.AssociationID = pointer.StringValue(IPv4.AssociationId)
		o.CIDRBlockState = v1beta1.VPCCIDRBlockState{
			State:         string(IPv4.CidrBlockState.State),
			StatusMessage: pointer.StringValue(IPv4.CidrBlockState.StatusMessage),
		}
		o.CIDRBlock = pointer.StringValue(IPv4.CidrBlock)
		return o
	}

	if IPv6 != nil {
		o.AssociationID = pointer.StringValue(IPv6.AssociationId)
		o.IPv6CIDRBlockState = v1beta1.VPCCIDRBlockState{
			State:         string(IPv6.Ipv6CidrBlockState.State),
			StatusMessage: pointer.StringValue(IPv6.Ipv6CidrBlockState.StatusMessage),
		}
		o.IPv6CIDRBlock = pointer.StringValue(IPv6.Ipv6CidrBlock)
		o.IPv6Pool = pointer.StringValue(IPv6.Ipv6Pool)
		o.NetworkBorderGroup = pointer.StringValue(IPv6.NetworkBorderGroup)
		return o
	}
	return o
}

// FindVPCCIDRBlockStatus is used to grab ec2.VpcCidrBlockStateCode from
// ec2types.Vpc.
func FindVPCCIDRBlockStatus(associationID string, vpc ec2types.Vpc) (ec2types.VpcCidrBlockStateCode, error) {
	IPv4, IPv6 := FindCIDRAssociation(associationID, vpc)

	if IPv4 != nil {
		return IPv4.CidrBlockState.State, nil
	}

	if IPv6 != nil {
		return IPv6.Ipv6CidrBlockState.State, nil
	}
	return ec2types.VpcCidrBlockStateCodeFailing, &CIDRNotFoundError{}
}

// FindCIDRAssociation will find the matching CIDRAssociation in the ec2.VPC or return nil
func FindCIDRAssociation(associationID string, vpc ec2types.Vpc) (*ec2types.VpcCidrBlockAssociation, *ec2types.VpcIpv6CidrBlockAssociation) {
	if len(vpc.CidrBlockAssociationSet) > 0 {
		for _, v := range vpc.CidrBlockAssociationSet {
			if aws.StringValue(v.AssociationId) == associationID {
				return &v, nil
			}
		}
	}
	if len(vpc.Ipv6CidrBlockAssociationSet) > 0 {
		for _, v := range vpc.Ipv6CidrBlockAssociationSet {
			if aws.StringValue(v.AssociationId) == associationID {
				return nil, &v
			}
		}
	}
	return nil, nil
}
