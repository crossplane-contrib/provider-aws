package ec2

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/awserr"
	"github.com/aws/aws-sdk-go-v2/service/ec2"

	"github.com/crossplane/provider-aws/apis/ec2/v1beta1"
	awsclients "github.com/crossplane/provider-aws/pkg/clients"
)

const (
	// SubnetIDNotFound is the code that is returned by ec2 when the given SubnetID is not valid
	SubnetIDNotFound = "InvalidSubnetID.NotFound"
)

// SubnetClient is the external client used for Subnet Custom Resource
type SubnetClient interface {
	CreateSubnetRequest(input *ec2.CreateSubnetInput) ec2.CreateSubnetRequest
	DescribeSubnetsRequest(input *ec2.DescribeSubnetsInput) ec2.DescribeSubnetsRequest
	DeleteSubnetRequest(input *ec2.DeleteSubnetInput) ec2.DeleteSubnetRequest
	ModifySubnetAttributeRequest(input *ec2.ModifySubnetAttributeInput) ec2.ModifySubnetAttributeRequest
	CreateTagsRequest(*ec2.CreateTagsInput) ec2.CreateTagsRequest
}

// NewSubnetClient returns a new client using AWS credentials as JSON encoded data.
func NewSubnetClient(cfg aws.Config) SubnetClient {
	return ec2.New(cfg)
}

// IsSubnetNotFoundErr returns true if the error is because the item doesn't exist
func IsSubnetNotFoundErr(err error) bool {
	if awsErr, ok := err.(awserr.Error); ok {
		if awsErr.Code() == SubnetIDNotFound {
			return true
		}
	}

	return false
}

// GenerateSubnetObservation is used to produce v1beta1.SubnetExternalStatus from
// ec2.Subnet
func GenerateSubnetObservation(subnet ec2.Subnet) v1beta1.SubnetObservation {
	o := v1beta1.SubnetObservation{
		AvailableIPAddressCount: aws.Int64Value(subnet.AvailableIpAddressCount),
		DefaultForAZ:            aws.BoolValue(subnet.DefaultForAz),
		SubnetID:                aws.StringValue(subnet.SubnetId),
		SubnetState:             string(subnet.State),
	}

	v, err := subnet.State.MarshalValue()
	if err != nil {
		o.SubnetState = v
	}

	return o
}

// LateInitializeSubnet fills the empty fields in *v1beta1.SubnetParameters with
// the values seen in ec2.Subnet.
func LateInitializeSubnet(in *v1beta1.SubnetParameters, s *ec2.Subnet) { // nolint:gocyclo
	if s == nil {
		return
	}

	in.AssignIPv6AddressOnCreation = awsclients.LateInitializeBoolPtr(in.AssignIPv6AddressOnCreation, s.AssignIpv6AddressOnCreation)
	in.AvailabilityZone = awsclients.LateInitializeStringPtr(in.AvailabilityZone, s.AvailabilityZone)
	in.AvailabilityZoneID = awsclients.LateInitializeStringPtr(in.AvailabilityZoneID, s.AvailabilityZoneId)
	in.CIDRBlock = awsclients.LateInitializeString(in.CIDRBlock, s.CidrBlock)
	in.MapPublicIPOnLaunch = awsclients.LateInitializeBoolPtr(in.MapPublicIPOnLaunch, s.MapPublicIpOnLaunch)
	in.VPCID = awsclients.LateInitializeStringPtr(in.VPCID, s.VpcId)

	if s.Ipv6CidrBlockAssociationSet != nil {
		in.IPv6CIDRBlock = awsclients.LateInitializeStringPtr(in.IPv6CIDRBlock, s.Ipv6CidrBlockAssociationSet[0].Ipv6CidrBlock)
	}

	if len(in.Tags) == 0 && len(s.Tags) != 0 {
		in.Tags = v1beta1.BuildFromEC2Tags(s.Tags)
	}
}

// IsSubnetUpToDate checks whether there is a change in any of the modifiable fields.
func IsSubnetUpToDate(p v1beta1.SubnetParameters, s ec2.Subnet) bool {
	if p.MapPublicIPOnLaunch != nil && (*p.MapPublicIPOnLaunch != *s.MapPublicIpOnLaunch) {
		return false
	}

	if p.AssignIPv6AddressOnCreation != nil && (*p.AssignIPv6AddressOnCreation != *s.AssignIpv6AddressOnCreation) {
		return false
	}

	return v1beta1.CompareTags(p.Tags, s.Tags)
}
