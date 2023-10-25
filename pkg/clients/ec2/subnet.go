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
	DeleteTags(ctx context.Context, input *ec2.DeleteTagsInput, opts ...func(*ec2.Options)) (*ec2.DeleteTagsOutput, error)
}

// NewSubnetClient returns a new client using AWS credentials as JSON encoded data.
func NewSubnetClient(cfg aws.Config) SubnetClient {
	return ec2.NewFromConfig(cfg)
}

// IsSubnetNotFoundErr returns true if the error is because the item doesn't exist
func IsSubnetNotFoundErr(err error) bool {
	var awsErr smithy.APIError
	return errors.As(err, &awsErr) && awsErr.ErrorCode() == SubnetIDNotFound
}

// GenerateSubnetObservation is used to produce v1beta1.SubnetExternalStatus from
// ec2types.Subnet
func GenerateSubnetObservation(subnet ec2types.Subnet) v1beta1.SubnetObservation {
	o := v1beta1.SubnetObservation{
		AvailableIPAddressCount: aws.ToInt32(subnet.AvailableIpAddressCount),
		DefaultForAZ:            aws.ToBool(subnet.DefaultForAz),
		SubnetID:                aws.ToString(subnet.SubnetId),
		SubnetState:             string(subnet.State),
	}

	o.SubnetState = string(subnet.State)

	return o
}

// LateInitializeSubnet fills the empty fields in *v1beta1.SubnetParameters with
// the values seen in ec2types.Subnet.
func LateInitializeSubnet(in *v1beta1.SubnetParameters, s *ec2types.Subnet) {
	if s == nil {
		return
	}

	in.AssignIPv6AddressOnCreation = pointer.LateInitialize(in.AssignIPv6AddressOnCreation, s.AssignIpv6AddressOnCreation)
	in.AvailabilityZone = pointer.LateInitialize(in.AvailabilityZone, s.AvailabilityZone)
	in.AvailabilityZoneID = pointer.LateInitialize(in.AvailabilityZoneID, s.AvailabilityZoneId)
	in.CIDRBlock = pointer.LateInitializeValueFromPtr(in.CIDRBlock, s.CidrBlock)
	in.MapPublicIPOnLaunch = pointer.LateInitialize(in.MapPublicIPOnLaunch, s.MapPublicIpOnLaunch)
	in.VPCID = pointer.LateInitialize(in.VPCID, s.VpcId)

	if s.Ipv6CidrBlockAssociationSet != nil && len(s.Ipv6CidrBlockAssociationSet) > 0 {
		in.IPv6CIDRBlock = pointer.LateInitialize(in.IPv6CIDRBlock, s.Ipv6CidrBlockAssociationSet[0].Ipv6CidrBlock)
	}

	if len(in.Tags) == 0 && len(s.Tags) != 0 {
		in.Tags = BuildFromEC2TagsV1Beta1(s.Tags)
	}
}

// IsSubnetUpToDate checks whether there is a change in any of the modifiable fields.
func IsSubnetUpToDate(p v1beta1.SubnetParameters, s ec2types.Subnet) bool {
	if aws.ToBool(p.MapPublicIPOnLaunch) != aws.ToBool(s.MapPublicIpOnLaunch) {
		return false
	}
	if aws.ToBool(p.AssignIPv6AddressOnCreation) != aws.ToBool(s.AssignIpv6AddressOnCreation) {
		return false
	}
	return CompareTagsV1Beta1(p.Tags, s.Tags)
}
