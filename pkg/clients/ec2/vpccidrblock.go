package ec2

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/awserr"
	"github.com/aws/aws-sdk-go-v2/service/ec2"

	"github.com/crossplane/provider-aws/apis/ec2/v1alpha1"
	awsclient "github.com/crossplane/provider-aws/pkg/clients"
)

const (
	errCIDRAssociationNotFound = "InvalidVpcCidrBlockAssociationID.NotFound"
)

// VPCCIDRBlockClient is the external client used for VPC CIDR Block Custom Resource
type VPCCIDRBlockClient interface {
	DescribeVpcsRequest(*ec2.DescribeVpcsInput) ec2.DescribeVpcsRequest
	AssociateVpcCidrBlockRequest(*ec2.AssociateVpcCidrBlockInput) ec2.AssociateVpcCidrBlockRequest
	DisassociateVpcCidrBlockRequest(*ec2.DisassociateVpcCidrBlockInput) ec2.DisassociateVpcCidrBlockRequest
}

// NewVPCCIDRBlockClient returns a new client using AWS credentials as JSON encoded data.
func NewVPCCIDRBlockClient(cfg aws.Config) VPCCIDRBlockClient {
	return ec2.New(cfg)
}

// CIDRNotFoundError will be raised when there is no Association
type CIDRNotFoundError struct{}

// Error satisfies the Error interface for CIDRNotFoundError.
func (r *CIDRNotFoundError) Error() string {
	return errCIDRAssociationNotFound
}

// IsCIDRNotFound returns true if the error code indicates that the CIDR Block Association was not found
func IsCIDRNotFound(err error) bool {
	if _, ok := err.(*CIDRNotFoundError); ok {
		return true
	}

	if awsErr, ok := err.(awserr.Error); ok {
		if awsErr.Code() == errCIDRAssociationNotFound {
			return true
		}
	}
	return false
}

// IsVpcCidrBlockUpToDate returns true if there is no update-able difference between desired
// and observed state of the resource.
func IsVpcCidrBlockUpToDate(associationID string, spec v1alpha1.VPCCIDRBlockParameters, vpc ec2.Vpc) (bool, error) {
	IPv4, IPv6 := FindCIDRAssociation(associationID, vpc)

	if IPv4 != nil {
		return *spec.CIDRBlock == *IPv4.CidrBlock, nil
	}

	if IPv6 != nil {
		return aws.StringValue(spec.IPv6CIDRBlock) == aws.StringValue(IPv6.Ipv6CidrBlock) &&
			aws.StringValue(spec.IPv6Pool) == aws.StringValue(IPv6.Ipv6Pool) &&
			aws.StringValue(spec.IPv6CIDRBlockNetworkBorderGroup) == aws.StringValue(IPv6.NetworkBorderGroup), nil
	}
	return false, &CIDRNotFoundError{}
}

// IsVpcCidrDeleting returns true if the CIDR Block is already disassociated or disassociating
func IsVpcCidrDeleting(observation v1alpha1.VPCCIDRBlockObservation) bool {
	switch {
	case observation.CIDRBlockState == nil && observation.IPv6CIDRBlockState == nil:
		return true
	case observation.CIDRBlockState != nil && (*observation.CIDRBlockState.State == string(ec2.VpcCidrBlockStateCodeDisassociating) || *observation.CIDRBlockState.State == string(ec2.VpcCidrBlockStateCodeDisassociated)):
		return true
	case observation.IPv6CIDRBlockState != nil && (*observation.IPv6CIDRBlockState.State == string(ec2.VpcCidrBlockStateCodeDisassociating) || *observation.IPv6CIDRBlockState.State == string(ec2.VpcCidrBlockStateCodeDisassociated)):
		return true
	default:
		return false
	}
}

// GenerateVpcCIDRBlockObservation is used to produce v1alpha1.VPCObservation from
// ec2.Vpc.
func GenerateVpcCIDRBlockObservation(associationID string, vpc ec2.Vpc) v1alpha1.VPCCIDRBlockObservation {
	o := v1alpha1.VPCCIDRBlockObservation{}

	IPv4, IPv6 := FindCIDRAssociation(associationID, vpc)

	if IPv4 != nil {
		o.AssociationID = IPv4.AssociationId
		o.CIDRBlockState = &v1alpha1.VPCCIDRBlockState{
			State:         awsclient.String(string(IPv4.CidrBlockState.State)),
			StatusMessage: IPv4.CidrBlockState.StatusMessage,
		}
		o.CIDRBlock = IPv4.CidrBlock
		return o
	}

	if IPv6 != nil {
		o.AssociationID = IPv6.AssociationId
		o.IPv6CIDRBlockState = &v1alpha1.VPCCIDRBlockState{
			State:         awsclient.String(string(IPv6.Ipv6CidrBlockState.State)),
			StatusMessage: IPv6.Ipv6CidrBlockState.StatusMessage,
		}
		o.IPv6CIDRBlock = IPv6.Ipv6CidrBlock
		o.IPv6Pool = IPv6.Ipv6Pool
		o.NetworkBorderGroup = IPv6.NetworkBorderGroup
		return o
	}
	return o
}

// FindVPCCIDRBlockStatus is used to grab ec2.VpcCidrBlockStateCode from
// ec2.Vpc.
func FindVPCCIDRBlockStatus(associationID string, vpc ec2.Vpc) (ec2.VpcCidrBlockStateCode, error) {
	IPv4, IPv6 := FindCIDRAssociation(associationID, vpc)

	if IPv4 != nil {
		return IPv4.CidrBlockState.State, nil
	}

	if IPv6 != nil {
		return IPv6.Ipv6CidrBlockState.State, nil
	}
	return ec2.VpcCidrBlockStateCodeFailing, &CIDRNotFoundError{}
}

// FindCIDRAssociation will find the matching CIDRAssociation in the ec2.VPC or return nil
func FindCIDRAssociation(associationID string, vpc ec2.Vpc) (*ec2.VpcCidrBlockAssociation, *ec2.VpcIpv6CidrBlockAssociation) {
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
