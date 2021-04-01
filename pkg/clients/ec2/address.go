package ec2

import (
	"sort"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/awserr"
	"github.com/aws/aws-sdk-go-v2/service/ec2"

	"github.com/crossplane/provider-aws/apis/ec2/v1beta1"
	awsclients "github.com/crossplane/provider-aws/pkg/clients"
)

const (
	// AddressAddressNotFound address not found
	AddressAddressNotFound = "InvalidAddress.NotFound"
	// AddressAllocationNotFound addreess not found by allocation
	AddressAllocationNotFound = "InvalidAllocationID.NotFound"
)

// AddressClient is the external client used for Address Custom Resource
type AddressClient interface {
	AllocateAddressRequest(input *ec2.AllocateAddressInput) ec2.AllocateAddressRequest
	DescribeAddressesRequest(input *ec2.DescribeAddressesInput) ec2.DescribeAddressesRequest
	ReleaseAddressRequest(input *ec2.ReleaseAddressInput) ec2.ReleaseAddressRequest
	CreateTagsRequest(*ec2.CreateTagsInput) ec2.CreateTagsRequest
}

// IsAddressNotFoundErr returns true if the error is because the address doesn't exist
func IsAddressNotFoundErr(err error) bool {
	if awsErr, ok := err.(awserr.Error); ok {
		if awsErr.Code() == AddressAddressNotFound || awsErr.Code() == AddressAllocationNotFound {
			return true
		}
	}
	return false
}

// GenerateAddressObservation is used to produce v1beta1.AddressObservation from
// ec2.Subnet
func GenerateAddressObservation(address ec2.Address) v1beta1.AddressObservation {
	o := v1beta1.AddressObservation{
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

// LateInitializeAddress fills the empty fields in *v1beta1.AddressParameters with
// the values seen in ec2.Address.
func LateInitializeAddress(in *v1beta1.AddressParameters, a *ec2.Address) { // nolint:gocyclo
	if a == nil {
		return
	}
	in.Address = awsclients.LateInitializeStringPtr(in.Address, a.PublicIp)
	in.Domain = awsclients.LateInitializeStringPtr(in.Domain, aws.String(string(a.Domain)))
	in.CustomerOwnedIPv4Pool = awsclients.LateInitializeStringPtr(in.CustomerOwnedIPv4Pool, a.CustomerOwnedIpv4Pool)
	in.NetworkBorderGroup = awsclients.LateInitializeStringPtr(in.NetworkBorderGroup, a.NetworkBorderGroup)
	in.PublicIPv4Pool = awsclients.LateInitializeStringPtr(in.PublicIPv4Pool, a.PublicIpv4Pool)
	if len(in.Tags) == 0 && len(a.Tags) != 0 {
		in.Tags = BuildFromEC2Tags(a.Tags)
	}
}

// IsAddressUpToDate checks whether there is a change in any of the modifiable fields.
func IsAddressUpToDate(e v1beta1.AddressParameters, a ec2.Address) bool {
	return CompareTags(e.Tags, a.Tags)
}

// IsStandardDomain checks whether it is set for standard domain
func IsStandardDomain(e v1beta1.AddressParameters) bool {
	return e.Domain != nil && *e.Domain == *aws.String(string(ec2.DomainTypeStandard))
}

// GenerateEC2Tags generates a tag array with type that EC2 client expects.
func GenerateEC2Tags(tags []v1beta1.Tag) []ec2.Tag {
	res := make([]ec2.Tag, len(tags))
	for i, t := range tags {
		res[i] = ec2.Tag{Key: aws.String(t.Key), Value: aws.String(t.Value)}
	}
	return res
}

// BuildFromEC2Tags returns a list of tags, off of the given ec2 tags
func BuildFromEC2Tags(tags []ec2.Tag) []v1beta1.Tag {
	if len(tags) < 1 {
		return nil
	}
	res := make([]v1beta1.Tag, len(tags))
	for i, t := range tags {
		res[i] = v1beta1.Tag{Key: aws.StringValue(t.Key), Value: aws.StringValue(t.Value)}
	}

	return res
}

// CompareTags compares arrays of v1beta1.Tag and ec2.Tag
func CompareTags(tags []v1beta1.Tag, ec2Tags []ec2.Tag) bool {
	if len(tags) != len(ec2Tags) {
		return false
	}

	SortTags(tags, ec2Tags)

	for i, t := range tags {
		if t.Key != *ec2Tags[i].Key || t.Value != *ec2Tags[i].Value {
			return false
		}
	}

	return true
}

// SortTags sorts array of v1beta1.Tag and ec2.Tag on 'Key'
func SortTags(tags []v1beta1.Tag, ec2Tags []ec2.Tag) {
	sort.Slice(tags, func(i, j int) bool {
		return tags[i].Key < tags[j].Key
	})

	sort.Slice(ec2Tags, func(i, j int) bool {
		return *ec2Tags[i].Key < *ec2Tags[j].Key
	})
}
