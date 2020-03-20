package ec2

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/awserr"
	"github.com/aws/aws-sdk-go-v2/service/ec2"

	"github.com/crossplane/provider-aws/apis/network/v1alpha3"
)

const (
	// InvalidGroupNotFound is the code that is returned by ec2 when the given VPCID is not valid
	InvalidGroupNotFound = "InvalidGroup.NotFound"

	// InvalidPermissionDuplicate is returned when you try to Authorize for a rule that already exists.
	InvalidPermissionDuplicate = "InvalidPermission.Duplicate"
)

// SecurityGroupClient is the external client used for SecurityGroup Custom Resource
type SecurityGroupClient interface {
	CreateSecurityGroupRequest(input *ec2.CreateSecurityGroupInput) ec2.CreateSecurityGroupRequest
	DeleteSecurityGroupRequest(input *ec2.DeleteSecurityGroupInput) ec2.DeleteSecurityGroupRequest
	DescribeSecurityGroupsRequest(input *ec2.DescribeSecurityGroupsInput) ec2.DescribeSecurityGroupsRequest
	AuthorizeSecurityGroupIngressRequest(input *ec2.AuthorizeSecurityGroupIngressInput) ec2.AuthorizeSecurityGroupIngressRequest
	AuthorizeSecurityGroupEgressRequest(input *ec2.AuthorizeSecurityGroupEgressInput) ec2.AuthorizeSecurityGroupEgressRequest
}

// NewSecurityGroupClient returns a new client using AWS credentials as JSON encoded data.
func NewSecurityGroupClient(cfg *aws.Config) (SecurityGroupClient, error) {
	return ec2.New(*cfg), nil
}

// IsSecurityGroupNotFoundErr returns true if the error is because the item doesn't exist
func IsSecurityGroupNotFoundErr(err error) bool {
	if awsErr, ok := err.(awserr.Error); ok {
		if awsErr.Code() == InvalidGroupNotFound {
			return true
		}
	}
	return false
}

// IsRuleAlreadyExistsErr returns true if the error is because the rule already exists.
func IsRuleAlreadyExistsErr(err error) bool {
	if awsErr, ok := err.(awserr.Error); ok {
		if awsErr.Code() == InvalidPermissionDuplicate {
			return true
		}
	}
	return false
}

// GenerateEC2Permissions converts object Permissions to ec2 format
func GenerateEC2Permissions(objectPerms []v1alpha3.IPPermission) []ec2.IpPermission {
	if len(objectPerms) == 0 {
		return nil
	}
	permissions := make([]ec2.IpPermission, len(objectPerms))
	for i, p := range objectPerms {
		ipPerm := ec2.IpPermission{
			FromPort:   p.FromPort,
			IpProtocol: aws.String(p.IPProtocol),
			ToPort:     p.ToPort,
		}
		for _, c := range p.IPRanges {
			ipPerm.IpRanges = append(ipPerm.IpRanges, ec2.IpRange{
				CidrIp:      aws.String(c.CIDRIP),
				Description: c.Description,
			})
		}
		for _, c := range p.IPv6Ranges {
			ipPerm.Ipv6Ranges = append(ipPerm.Ipv6Ranges, ec2.Ipv6Range{
				CidrIpv6:    aws.String(c.CIDRIPv6),
				Description: c.Description,
			})
		}
		for _, c := range p.PrefixListIDs {
			ipPerm.PrefixListIds = append(ipPerm.PrefixListIds, ec2.PrefixListId{
				Description:  c.Description,
				PrefixListId: aws.String(c.PrefixListID),
			})
		}
		for _, c := range p.UserIDGroupPairs {
			ipPerm.UserIdGroupPairs = append(ipPerm.UserIdGroupPairs, ec2.UserIdGroupPair{
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
