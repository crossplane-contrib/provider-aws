package ec2

import (
	"context"
	"errors"
	"sort"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/smithy-go"

	"github.com/crossplane-contrib/provider-aws/apis/ec2/v1beta1"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
)

const (
	// AddressAddressNotFound address not found
	AddressAddressNotFound = "InvalidAddress.NotFound"
	// AddressAllocationNotFound addreess not found by allocation
	AddressAllocationNotFound = "InvalidAllocationID.NotFound"
)

// AddressClient is the external client used for ElasticIP Custom Resource
type AddressClient interface {
	AllocateAddress(ctx context.Context, input *ec2.AllocateAddressInput, opts ...func(*ec2.Options)) (*ec2.AllocateAddressOutput, error)
	DescribeAddresses(ctx context.Context, input *ec2.DescribeAddressesInput, opts ...func(*ec2.Options)) (*ec2.DescribeAddressesOutput, error)
	ReleaseAddress(ctx context.Context, input *ec2.ReleaseAddressInput, opts ...func(*ec2.Options)) (*ec2.ReleaseAddressOutput, error)
	CreateTags(ctx context.Context, input *ec2.CreateTagsInput, opts ...func(*ec2.Options)) (*ec2.CreateTagsOutput, error)
}

// IsAddressNotFoundErr returns true if the error is because the address doesn't exist
func IsAddressNotFoundErr(err error) bool {
	var awsErr smithy.APIError
	return errors.As(err, &awsErr) && (awsErr.ErrorCode() == AddressAddressNotFound || awsErr.ErrorCode() == AddressAllocationNotFound)
}

// GenerateAddressObservation is used to produce v1beta1.AddressObservation from
// ec2.Subnet
func GenerateAddressObservation(address ec2types.Address) v1beta1.AddressObservation {
	o := v1beta1.AddressObservation{
		AllocationID:            aws.ToString(address.AllocationId),
		AssociationID:           aws.ToString(address.AssociationId),
		CustomerOwnedIP:         aws.ToString(address.CustomerOwnedIp),
		CustomerOwnedIPv4Pool:   aws.ToString(address.CustomerOwnedIpv4Pool),
		InstanceID:              aws.ToString(address.InstanceId),
		NetworkBorderGroup:      aws.ToString(address.NetworkBorderGroup),
		NetworkInterfaceID:      aws.ToString(address.NetworkInterfaceId),
		NetworkInterfaceOwnerID: aws.ToString(address.NetworkInterfaceOwnerId),
		PrivateIPAddress:        aws.ToString(address.PrivateIpAddress),
		PublicIP:                aws.ToString(address.PublicIp),
		PublicIPv4Pool:          aws.ToString(address.PublicIpv4Pool),
	}
	return o
}

// LateInitializeAddress fills the empty fields in *v1beta1.AddressParameters with
// the values seen in ec2types.Address.
func LateInitializeAddress(in *v1beta1.AddressParameters, a *ec2types.Address) {
	if a == nil {
		return
	}
	in.Address = pointer.LateInitialize(in.Address, a.PublicIp)
	in.Domain = pointer.LateInitialize(in.Domain, aws.String(string(a.Domain)))
	in.CustomerOwnedIPv4Pool = pointer.LateInitialize(in.CustomerOwnedIPv4Pool, a.CustomerOwnedIpv4Pool)
	in.NetworkBorderGroup = pointer.LateInitialize(in.NetworkBorderGroup, a.NetworkBorderGroup)
	in.PublicIPv4Pool = pointer.LateInitialize(in.PublicIPv4Pool, a.PublicIpv4Pool)
	if len(in.Tags) == 0 && len(a.Tags) != 0 {
		in.Tags = BuildFromEC2Tags(a.Tags)
	}
}

// IsAddressUpToDate checks whether there is a change in any of the modifiable fields.
func IsAddressUpToDate(e v1beta1.AddressParameters, a ec2types.Address) bool {
	return CompareTags(e.Tags, a.Tags)
}

// IsStandardDomain checks whether it is set for standard domain
func IsStandardDomain(e v1beta1.AddressParameters) bool {
	return e.Domain != nil && *e.Domain == *aws.String(string(ec2types.DomainTypeStandard))
}

// GenerateEC2Tags generates a tag array with type that EC2 client expects.
func GenerateEC2Tags(tags []v1beta1.Tag) []ec2types.Tag {
	res := make([]ec2types.Tag, len(tags))
	for i, t := range tags {
		res[i] = ec2types.Tag{Key: aws.String(t.Key), Value: aws.String(t.Value)}
	}
	return res
}

// BuildFromEC2Tags returns a list of tags, off of the given ec2 tags
func BuildFromEC2Tags(tags []ec2types.Tag) []v1beta1.Tag {
	if len(tags) < 1 {
		return nil
	}
	res := make([]v1beta1.Tag, len(tags))
	for i, t := range tags {
		res[i] = v1beta1.Tag{Key: aws.ToString(t.Key), Value: aws.ToString(t.Value)}
	}

	return res
}

// CompareTags compares arrays of v1beta1.Tag and ec2types.Tag
func CompareTags(tags []v1beta1.Tag, ec2Tags []ec2types.Tag) bool {
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

// SortTags sorts array of v1beta1.Tag and ec2types.Tag on 'Key'
func SortTags(tags []v1beta1.Tag, ec2Tags []ec2types.Tag) {
	sort.Slice(tags, func(i, j int) bool {
		return tags[i].Key < tags[j].Key
	})

	sort.Slice(ec2Tags, func(i, j int) bool {
		return *ec2Tags[i].Key < *ec2Tags[j].Key
	})
}
