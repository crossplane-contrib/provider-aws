package ec2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/awserr"
	"github.com/aws/aws-sdk-go-v2/service/ec2"

	"github.com/crossplane/provider-aws/apis/ec2/v1alpha4"
	"github.com/crossplane/provider-aws/apis/ec2/v1beta1"
	awsclients "github.com/crossplane/provider-aws/pkg/clients"
)

const (
	// ElasticIPAddressNotFound address not found
	ElasticIPAddressNotFound = "InvalidAddress.NotFound"
	// ElasticIPAllocationNotFound addreess not found by allocation
	ElasticIPAllocationNotFound = "InvalidAllocationID.NotFound"
)

// ElasticIPClient is the external client used for ElasticIP Custom Resource
type ElasticIPClient interface {
	AllocateAddressRequest(input *ec2.AllocateAddressInput) ec2.AllocateAddressRequest
	DescribeAddressesRequest(input *ec2.DescribeAddressesInput) ec2.DescribeAddressesRequest
	ReleaseAddressRequest(input *ec2.ReleaseAddressInput) ec2.ReleaseAddressRequest
	CreateTagsRequest(*ec2.CreateTagsInput) ec2.CreateTagsRequest
}

// NewElasticIPClient returns a new client using AWS credentials as JSON encoded data.
func NewElasticIPClient(ctx context.Context, credentials []byte, region string, auth awsclients.AuthMethod) (ElasticIPClient, error) {
	cfg, err := auth(ctx, credentials, awsclients.DefaultSection, region)
	if cfg == nil {
		return nil, err
	}
	return ec2.New(*cfg), nil
}

// IsAddressNotFoundErr returns true if the error is because the address doesn't exist
func IsAddressNotFoundErr(err error) bool {
	if awsErr, ok := err.(awserr.Error); ok {
		if awsErr.Code() == ElasticIPAddressNotFound || awsErr.Code() == ElasticIPAllocationNotFound {
			return true
		}
	}
	return false
}

// GenerateElasticIPObservation is used to produce v1alpha4.ElasticIPObservation from
// ec2.Subnet
func GenerateElasticIPObservation(address ec2.Address) v1alpha4.ElasticIPObservation {
	o := v1alpha4.ElasticIPObservation{
		AllocationID:            aws.StringValue(address.AllocationId),
		AssociationID:           aws.StringValue(address.AssociationId),
		CustomerOwnedIP:         aws.StringValue(address.CustomerOwnedIp),
		CustomerOwnedIPv4Pool:   aws.StringValue(address.CustomerOwnedIpv4Pool),
		InstanceID:              aws.StringValue(address.InstanceId),
		NetworkBorderGroup:      aws.StringValue(address.NetworkBorderGroup),
		NetworkInterfaceID:      aws.StringValue(address.NetworkInterfaceId),
		NetworkInterfaceOwnerID: aws.StringValue(address.NetworkInterfaceOwnerId),
		PrivateIPAddress:        aws.StringValue(address.PrivateIpAddress),
		PublicIP:                aws.StringValue(address.PublicIp),
		PublicIPv4Pool:          aws.StringValue(address.PublicIpv4Pool),
	}
	return o
}

// LateInitializeElasticIP fills the empty fields in *v1alpha4.ElasticIPParameters with
// the values seen in ec2.Address.
func LateInitializeElasticIP(in *v1alpha4.ElasticIPParameters, a *ec2.Address) { // nolint:gocyclo
	if a == nil {
		return
	}
	in.Address = awsclients.LateInitializeStringPtr(in.Address, a.PublicIp)
	in.Domain = awsclients.LateInitializeString(in.Domain, aws.String(string(a.Domain)))
	in.CustomerOwnedIPv4Pool = awsclients.LateInitializeStringPtr(in.CustomerOwnedIPv4Pool, a.CustomerOwnedIpv4Pool)
	in.NetworkBorderGroup = awsclients.LateInitializeStringPtr(in.NetworkBorderGroup, a.NetworkBorderGroup)
	in.PublicIPv4Pool = awsclients.LateInitializeStringPtr(in.PublicIPv4Pool, a.PublicIpv4Pool)
	if len(in.Tags) == 0 && len(a.Tags) != 0 {
		in.Tags = v1beta1.BuildFromEC2Tags(a.Tags)
	}
}

// IsElasticIPUpToDate checks whether there is a change in any of the modifiable fields.
func IsElasticIPUpToDate(e v1alpha4.ElasticIPParameters, a ec2.Address) bool {
	return v1beta1.CompareTags(e.Tags, a.Tags)
}
