package ec2

import (
	"context"
	"errors"

	awsgo "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/smithy-go"

	"github.com/crossplane-contrib/provider-aws/apis/ec2/v1beta1"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
)

const (
	// InvalidGroupNotFound is the code that is returned by ec2 when the given VPCID is not valid
	InvalidGroupNotFound = "InvalidGroup.NotFound"

	// InvalidPermissionDuplicate is returned when you try to Authorize for a rule that already exists.
	InvalidPermissionDuplicate = "InvalidPermission.Duplicate"
)

// SecurityGroupClient is the external client used for SecurityGroup Custom Resource
type SecurityGroupClient interface {
	CreateSecurityGroup(ctx context.Context, input *ec2.CreateSecurityGroupInput, opts ...func(*ec2.Options)) (*ec2.CreateSecurityGroupOutput, error)
	DeleteSecurityGroup(ctx context.Context, input *ec2.DeleteSecurityGroupInput, opts ...func(*ec2.Options)) (*ec2.DeleteSecurityGroupOutput, error)
	DescribeSecurityGroups(ctx context.Context, input *ec2.DescribeSecurityGroupsInput, opts ...func(*ec2.Options)) (*ec2.DescribeSecurityGroupsOutput, error)
	AuthorizeSecurityGroupIngress(ctx context.Context, input *ec2.AuthorizeSecurityGroupIngressInput, opts ...func(*ec2.Options)) (*ec2.AuthorizeSecurityGroupIngressOutput, error)
	AuthorizeSecurityGroupEgress(ctx context.Context, input *ec2.AuthorizeSecurityGroupEgressInput, opts ...func(*ec2.Options)) (*ec2.AuthorizeSecurityGroupEgressOutput, error)
	RevokeSecurityGroupIngress(ctx context.Context, input *ec2.RevokeSecurityGroupIngressInput, opts ...func(*ec2.Options)) (*ec2.RevokeSecurityGroupIngressOutput, error)
	RevokeSecurityGroupEgress(ctx context.Context, input *ec2.RevokeSecurityGroupEgressInput, opts ...func(*ec2.Options)) (*ec2.RevokeSecurityGroupEgressOutput, error)
	CreateTags(ctx context.Context, input *ec2.CreateTagsInput, opts ...func(*ec2.Options)) (*ec2.CreateTagsOutput, error)
	DeleteTags(ctx context.Context, input *ec2.DeleteTagsInput, opts ...func(*ec2.Options)) (*ec2.DeleteTagsOutput, error)
}

// NewSecurityGroupClient generates client for AWS Security Group API
func NewSecurityGroupClient(cfg awsgo.Config) SecurityGroupClient {
	return ec2.NewFromConfig(cfg)
}

// IsSecurityGroupNotFoundErr returns true if the error is because the item doesn't exist
func IsSecurityGroupNotFoundErr(err error) bool {
	var awsErr smithy.APIError
	return errors.As(err, &awsErr) && awsErr.ErrorCode() == InvalidGroupNotFound
}

// IsRuleAlreadyExistsErr returns true if the error is because the rule already exists.
func IsRuleAlreadyExistsErr(err error) bool {
	var awsErr smithy.APIError
	return errors.As(err, &awsErr) && awsErr.ErrorCode() == InvalidPermissionDuplicate
}

// GenerateEC2Permissions converts object Permissions to ec2 format
func GenerateEC2Permissions(objectPerms []v1beta1.IPPermission) []ec2types.IpPermission {
	if len(objectPerms) == 0 {
		return nil
	}
	permissions := make([]ec2types.IpPermission, len(objectPerms))
	for i, p := range objectPerms {
		ipPerm := ec2types.IpPermission{
			FromPort:   p.FromPort,
			IpProtocol: pointer.ToOrNilIfZeroValue(p.IPProtocol),
			ToPort:     p.ToPort,
		}
		for _, c := range p.IPRanges {
			ipPerm.IpRanges = append(ipPerm.IpRanges, ec2types.IpRange{
				CidrIp:      pointer.ToOrNilIfZeroValue(c.CIDRIP),
				Description: c.Description,
			})
		}
		for _, c := range p.IPv6Ranges {
			ipPerm.Ipv6Ranges = append(ipPerm.Ipv6Ranges, ec2types.Ipv6Range{
				CidrIpv6:    pointer.ToOrNilIfZeroValue(c.CIDRIPv6),
				Description: c.Description,
			})
		}
		for _, c := range p.PrefixListIDs {
			ipPerm.PrefixListIds = append(ipPerm.PrefixListIds, ec2types.PrefixListId{
				Description:  c.Description,
				PrefixListId: pointer.ToOrNilIfZeroValue(c.PrefixListID),
			})
		}
		for _, c := range p.UserIDGroupPairs {
			ipPerm.UserIdGroupPairs = append(ipPerm.UserIdGroupPairs, ec2types.UserIdGroupPair{
				Description:            c.Description,
				GroupId:                c.GroupID,
				GroupName:              c.GroupName,
				UserId:                 c.UserID,
				VpcId:                  c.VPCID,
				VpcPeeringConnectionId: c.VPCPeeringConnectionID,
			})
		}
		permissions[i] = ipPerm
	}
	return permissions
}

// GenerateSGObservation is used to produce v1beta1.SecurityGroupExternalStatus from
// ec2types.SecurityGroup.
func GenerateSGObservation(sg ec2types.SecurityGroup) v1beta1.SecurityGroupObservation {
	return v1beta1.SecurityGroupObservation{
		OwnerID:         pointer.StringValue(sg.OwnerId),
		SecurityGroupID: pointer.StringValue(sg.GroupId),
	}
}

// LateInitializeSG fills the empty fields in *v1beta1.SecurityGroupParameters with
// the values seen in ec2types.SecurityGroup.
func LateInitializeSG(in *v1beta1.SecurityGroupParameters, sg *ec2types.SecurityGroup) {
	if sg == nil {
		return
	}

	in.Description = pointer.LateInitializeValueFromPtr(in.Description, sg.Description)
	in.GroupName = pointer.LateInitializeValueFromPtr(in.GroupName, sg.GroupName)
	in.VPCID = pointer.LateInitialize(in.VPCID, sg.VpcId)

	// We cannot safely late init egress/ingress rules because they are keyless arrays

	if len(in.Tags) == 0 && len(sg.Tags) != 0 {
		in.Tags = BuildFromEC2TagsV1Beta1(sg.Tags)
	}
}

// IsSGUpToDate checks if the observed security group is up to equal to the desired state
func IsSGUpToDate(sg v1beta1.SecurityGroupParameters, observed ec2types.SecurityGroup) bool {
	if !CompareTags(sg.Tags, observed.Tags) {
		return false
	}

	if !pointer.BoolValue(sg.IgnoreIngress) {
		add, remove := DiffPermissions(GenerateEC2Permissions(sg.Ingress), observed.IpPermissions)
		if len(add) > 0 || len(remove) > 0 {
			return false
		}
	}
	if !pointer.BoolValue(sg.IgnoreEgress) {
		add, remove := DiffPermissions(GenerateEC2Permissions(sg.Egress), observed.IpPermissionsEgress)
		if len(add) > 0 || len(remove) > 0 {
			return false
		}
	}
	return true
}
