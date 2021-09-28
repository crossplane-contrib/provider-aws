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
	// SubnetIDNotFound is the code that is returned by ec2 when the given SubnetID is not valid
	SubnetIDNotFound = "InvalidSubnetID.NotFound"
)

// SubnetClient is the external client used for Subnet Custom Resource
type SubnetClient interface {
	CreateSubnet(ctx context.Context, input *ec2.CreateSubnetInput, opts ...func(*ec2.Options)) (*ec2.CreateSubnetOutput, error)
	DescribeSubnets(ctx context.Context, input *ec2.DescribeSubnetsInput, opts ...func(*ec2.Options)) (*ec2.DescribeSubnetsOutput, error)
	DeleteSubnet(ctx context.Context, input *ec2.DeleteSubnetInput, opts ...func(*ec2.Options)) (*ec2.DeleteSubnetOutput, error)
	ModifySubnetAttribute(ctx context.Context, input *ec2.ModifySubnetAttributeInput, opts ...func(*ec2.Options)) (*ec2.ModifySubnetAttributeOutput, error)
	CreateTags(ctx context.Context, input *ec2.CreateTagsInput, opts ...func(*ec2.Options)) (*ec2.CreateTagsOutput, error)
}

// NewSubnetClient returns a new client using AWS credentials as JSON encoded data.
func NewSubnetClient(cfg aws.Config) SubnetClient {
	return ec2.NewFromConfig(cfg)
}

// IsSubnetNotFoundErr returns true if the error is because the item doesn't exist
func IsSubnetNotFoundErr(err error) bool {
	if awsErr, ok := err.(smithy.APIError); ok {
		if awsErr.ErrorCode() == SubnetIDNotFound {
			return true
		}
	}

	return false
}

// GenerateSubnetObservation is used to produce v1beta1.SubnetExternalStatus from
// ec2types.Subnet
func GenerateSubnetObservation(subnet ec2types.Subnet) v1beta1.SubnetObservation {
	o := v1beta1.SubnetObservation{
		AvailableIPAddressCount: subnet.AvailableIpAddressCount,
		DefaultForAZ:            subnet.DefaultForAz,
		SubnetID:                aws.ToString(subnet.SubnetId),
		SubnetState:             string(subnet.State),
	}

	o.SubnetState = string(subnet.State)

	return o
}

// LateInitializeSubnet fills the empty fields in *v1beta1.SubnetParameters with
// the values seen in ec2types.Subnet.
func LateInitializeSubnet(in *v1beta1.SubnetParameters, s *ec2types.Subnet) { // nolint:gocyclo
	if s == nil {
		return
	}

	in.AssignIPv6AddressOnCreation = awsclients.LateInitializeBoolPtr(in.AssignIPv6AddressOnCreation, &s.AssignIpv6AddressOnCreation)
	in.AvailabilityZone = awsclients.LateInitializeStringPtr(in.AvailabilityZone, s.AvailabilityZone)
	in.AvailabilityZoneID = awsclients.LateInitializeStringPtr(in.AvailabilityZoneID, s.AvailabilityZoneId)
	in.CIDRBlock = awsclients.LateInitializeString(in.CIDRBlock, s.CidrBlock)
	in.MapPublicIPOnLaunch = awsclients.LateInitializeBoolPtr(in.MapPublicIPOnLaunch, &s.MapPublicIpOnLaunch)
	in.VPCID = awsclients.LateInitializeStringPtr(in.VPCID, s.VpcId)

	if s.Ipv6CidrBlockAssociationSet != nil {
		in.IPv6CIDRBlock = awsclients.LateInitializeStringPtr(in.IPv6CIDRBlock, s.Ipv6CidrBlockAssociationSet[0].Ipv6CidrBlock)
	}

	if len(in.Tags) == 0 && len(s.Tags) != 0 {
		in.Tags = v1beta1.BuildFromEC2Tags(s.Tags)
	}
}

// IsSubnetUpToDate checks whether there is a change in any of the modifiable fields.
func IsSubnetUpToDate(p v1beta1.SubnetParameters, s ec2types.Subnet) bool {
	if p.MapPublicIPOnLaunch != nil && (*p.MapPublicIPOnLaunch != s.MapPublicIpOnLaunch) {
		return false
	}

	if p.AssignIPv6AddressOnCreation != nil && (*p.AssignIPv6AddressOnCreation != s.AssignIpv6AddressOnCreation) {
		return false
	}

	return v1beta1.CompareTags(p.Tags, s.Tags)
}
