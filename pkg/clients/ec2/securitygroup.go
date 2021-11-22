package ec2

import (
	"context"
	"errors"

	awsgo "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/smithy-go"

	"github.com/crossplane/provider-aws/apis/ec2/v1beta1"
	aws "github.com/crossplane/provider-aws/pkg/clients"
	awsclients "github.com/crossplane/provider-aws/pkg/clients"
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

// BuildFromEC2Permissions converts ec2 Permissions to our API's format.
func BuildFromEC2Permissions(in []ec2types.IpPermission) []v1beta1.IPPermission {
	if len(in) == 0 {
		return nil
	}
	permissions := make([]v1beta1.IPPermission, len(in))
	for i, p := range in {
		ipPerm := v1beta1.IPPermission{
			FromPort:   p.FromPort,
			IPProtocol: aws.StringValue(p.IpProtocol),
			ToPort:     p.ToPort,
		}
		for _, c := range p.IpRanges {
			ipPerm.IPRanges = append(ipPerm.IPRanges, v1beta1.IPRange{
				CIDRIP:      aws.StringValue(c.CidrIp),
				Description: c.Description,
			})
		}
		for _, c := range p.Ipv6Ranges {
			ipPerm.IPv6Ranges = append(ipPerm.IPv6Ranges, v1beta1.IPv6Range{
				CIDRIPv6:    aws.StringValue(c.CidrIpv6),
				Description: c.Description,
			})
		}
		for _, c := range p.PrefixListIds {
			ipPerm.PrefixListIDs = append(ipPerm.PrefixListIDs, v1beta1.PrefixListID{
				Description:  c.Description,
				PrefixListID: aws.StringValue(c.PrefixListId),
			})
		}
		for _, c := range p.UserIdGroupPairs {
			ipPerm.UserIDGroupPairs = append(ipPerm.UserIDGroupPairs, v1beta1.UserIDGroupPair{
				Description:            c.Description,
				GroupID:                c.GroupId,
				GroupName:              c.GroupName,
				UserID:                 c.UserId,
				VPCID:                  c.VpcId,
				VPCPeeringConnectionID: c.VpcPeeringConnectionId,
			})
		}
		permissions[i] = ipPerm
	}
	return permissions
}

// GenerateEC2Permissions converts object Permissions to ec2 format
func GenerateEC2Permissions(in []v1beta1.IPPermission) []ec2types.IpPermission {
	if len(in) == 0 {
		return nil
	}
	permissions := make([]ec2types.IpPermission, len(in))
	for i, p := range in {
		ipPerm := ec2types.IpPermission{
			FromPort:   p.FromPort,
			IpProtocol: aws.String(p.IPProtocol),
			ToPort:     p.ToPort,
		}
		for _, c := range p.IPRanges {
			ipPerm.IpRanges = append(ipPerm.IpRanges, ec2types.IpRange{
				CidrIp:      aws.String(c.CIDRIP),
				Description: c.Description,
			})
		}
		for _, c := range p.IPv6Ranges {
			ipPerm.Ipv6Ranges = append(ipPerm.Ipv6Ranges, ec2types.Ipv6Range{
				CidrIpv6:    aws.String(c.CIDRIPv6),
				Description: c.Description,
			})
		}
		for _, c := range p.PrefixListIDs {
			ipPerm.PrefixListIds = append(ipPerm.PrefixListIds, ec2types.PrefixListId{
				Description:  c.Description,
				PrefixListId: aws.String(c.PrefixListID),
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
		OwnerID:         aws.StringValue(sg.OwnerId),
		SecurityGroupID: aws.StringValue(sg.GroupId),
	}
}

// LateInitializeSG fills the empty fields in *v1beta1.SecurityGroupParameters with
// the values seen in ec2types.SecurityGroup.
func LateInitializeSG(in *v1beta1.SecurityGroupParameters, sg *ec2types.SecurityGroup) {
	if sg == nil {
		return
	}

	in.Description = awsclients.LateInitializeString(in.Description, sg.Description)
	in.GroupName = awsclients.LateInitializeString(in.GroupName, sg.GroupName)
	in.VPCID = awsclients.LateInitializeStringPtr(in.VPCID, sg.VpcId)

	// We cannot safely late init egress/ingress rules because they are
	// keyless arrays, so we follow the typical pattern of only late
	// initializing if our desired rules are zero length.
	if len(in.Ingress) == 0 && len(sg.IpPermissions) != 0 {
		in.Ingress = BuildFromEC2Permissions(sg.IpPermissions)
	}
	if len(in.Egress) == 0 && len(sg.IpPermissionsEgress) != 0 {
		in.Egress = BuildFromEC2Permissions(sg.IpPermissionsEgress)
	}

	if len(in.Tags) == 0 && len(sg.Tags) != 0 {
		in.Tags = v1beta1.BuildFromEC2Tags(sg.Tags)
	}
}

// IsSGUpToDate checks if the observed security group is up to equal to the desired state
func IsSGUpToDate(sg v1beta1.SecurityGroupParameters, observed ec2types.SecurityGroup) bool {
	if !TagsMatch(sg.Tags, observed.Tags) {
		return false
	}

	add, remove := DiffPermissions(GenerateEC2Permissions(sg.Ingress), observed.IpPermissions)
	if len(add) > 0 || len(remove) > 0 {
		return false
	}

	add, remove = DiffPermissions(GenerateEC2Permissions(sg.Egress), observed.IpPermissionsEgress)
	if len(add) > 0 || len(remove) > 0 {
		return false
	}
	return true
}
