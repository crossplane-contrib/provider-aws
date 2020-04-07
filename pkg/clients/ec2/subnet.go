package ec2

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/awserr"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/crossplane/provider-aws/apis/network/v1beta1"
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
}

// NewSubnetClient returns a new client using AWS credentials as JSON encoded data.
func NewSubnetClient(ctx context.Context, credentials []byte, region string, auth awsclients.AuthMethod) (SubnetClient, error) {
	cfg, err := auth(ctx, credentials, awsclients.DefaultSection, region)
	if cfg == nil {
		return nil, err
	}
	return ec2.New(*cfg), nil
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
		DefaultForAz:            aws.BoolValue(subnet.DefaultForAz),
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

	in.AssignIpv6AddressOnCreation = awsclients.LateInitializeBoolPtr(in.AssignIpv6AddressOnCreation, s.AssignIpv6AddressOnCreation)
	in.AvailabilityZone = awsclients.LateInitializeStringPtr(in.AvailabilityZone, s.AvailabilityZone)
	in.AvailabilityZoneID = awsclients.LateInitializeStringPtr(in.AvailabilityZoneID, s.AvailabilityZoneId)
	in.CIDRBlock = awsclients.LateInitializeString(in.CIDRBlock, s.CidrBlock)
	in.MapPublicIPOnLaunch = awsclients.LateInitializeBoolPtr(in.MapPublicIPOnLaunch, s.MapPublicIpOnLaunch)
	in.VPCID = awsclients.LateInitializeString(in.VPCID, s.VpcId)

	if s.Ipv6CidrBlockAssociationSet != nil {
		in.Ipv6CIDRBlock = awsclients.LateInitializeStringPtr(in.Ipv6CIDRBlock, s.Ipv6CidrBlockAssociationSet[0].Ipv6CidrBlock)
	}
}

// CreateSubnetPatch creates a *v1beta1.SubnetParameters that has only the changed
// values between the target *v1beta1.SubnetParameters and the current
// *ec2.Subnet
func CreateSubnetPatch(in *ec2.Subnet, target *v1beta1.SubnetParameters) (*v1beta1.SubnetParameters, error) {
	currentParams := &v1beta1.SubnetParameters{}
	LateInitializeSubnet(currentParams, in)

	jsonPatch, err := awsclients.CreateJSONPatch(currentParams, target)
	if err != nil {
		return nil, err
	}
	patch := &v1beta1.SubnetParameters{}
	if err := json.Unmarshal(jsonPatch, patch); err != nil {
		return nil, err
	}
	return patch, nil
}

// IsSubnetUpToDate checks whether there is a change in any of the modifiable fields.
func IsSubnetUpToDate(p v1beta1.SubnetParameters, s ec2.Subnet) (bool, error) {
	patch, err := CreateSubnetPatch(&s, &p)
	if err != nil {
		return false, err
	}
	return cmp.Equal(v1beta1.SubnetParameters{}, *patch, cmpopts.IgnoreInterfaces(struct{ resource.AttributeReferencer }{})), nil
}
