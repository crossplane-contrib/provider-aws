package ec2

import (
	"context"
	"encoding/json"

	awsgo "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/awserr"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

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
	CreateSecurityGroupRequest(input *ec2.CreateSecurityGroupInput) ec2.CreateSecurityGroupRequest
	DeleteSecurityGroupRequest(input *ec2.DeleteSecurityGroupInput) ec2.DeleteSecurityGroupRequest
	DescribeSecurityGroupsRequest(input *ec2.DescribeSecurityGroupsInput) ec2.DescribeSecurityGroupsRequest
	AuthorizeSecurityGroupIngressRequest(input *ec2.AuthorizeSecurityGroupIngressInput) ec2.AuthorizeSecurityGroupIngressRequest
	AuthorizeSecurityGroupEgressRequest(input *ec2.AuthorizeSecurityGroupEgressInput) ec2.AuthorizeSecurityGroupEgressRequest
	RevokeSecurityGroupEgressRequest(input *ec2.RevokeSecurityGroupEgressInput) ec2.RevokeSecurityGroupEgressRequest
	CreateTagsRequest(input *ec2.CreateTagsInput) ec2.CreateTagsRequest
}

// NewSecurityGroupClient generates client for AWS Security Group API
func NewSecurityGroupClient(ctx context.Context, credentials []byte, region string, auth awsclients.AuthMethod) (SecurityGroupClient, error) {
	cfg, err := auth(ctx, credentials, awsclients.DefaultSection, region)
	if cfg == nil {
		return nil, err
	}
	return ec2.New(*cfg), err
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
func GenerateEC2Permissions(objectPerms []v1beta1.IPPermission) []ec2.IpPermission {
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

// GenerateIPPermissions converts object EC2 Permissions to IPPermission format
func GenerateIPPermissions(objectPerms []ec2.IpPermission) []v1beta1.IPPermission {
	if len(objectPerms) == 0 {
		return nil
	}
	permissions := make([]v1beta1.IPPermission, len(objectPerms))
	for i, p := range objectPerms {
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

// GenerateSGObservation is used to produce v1beta1.SecurityGroupExternalStatus from
// ec2.SecurityGroup.
func GenerateSGObservation(sg ec2.SecurityGroup) v1beta1.SecurityGroupObservation {
	return v1beta1.SecurityGroupObservation{
		OwnerID:         aws.StringValue(sg.OwnerId),
		SecurityGroupID: aws.StringValue(sg.GroupId),
	}
}

// LateInitializeSG fills the empty fields in *v1beta1.SecurityGroupParameters with
// the values seen in ec2.SecurityGroup.
func LateInitializeSG(in *v1beta1.SecurityGroupParameters, sg *ec2.SecurityGroup) { // nolint:gocyclo
	if sg == nil {
		return
	}

	in.Description = awsclients.LateInitializeString(in.Description, sg.Description)
	in.GroupName = awsclients.LateInitializeString(in.GroupName, sg.GroupName)
	in.VPCID = awsclients.LateInitializeStringPtr(in.VPCID, sg.VpcId)

	if len(in.Egress) == 0 && len(sg.IpPermissionsEgress) != 0 {
		in.Egress = GenerateIPPermissions(sg.IpPermissionsEgress)
	}

	if len(in.Ingress) == 0 && len(sg.IpPermissions) != 0 {
		in.Ingress = GenerateIPPermissions(sg.IpPermissions)
	}

	if len(in.Tags) == 0 && len(sg.Tags) != 0 {
		in.Tags = v1beta1.BuildFromEC2Tags(sg.Tags)
	}
}

// CreateSGPatch creates a *v1beta1.SecurityGroupParameters that has only the changed
// values between the target *v1beta1.SecurityGroupParameters and the current
// *ec2.SecurityGroup
func CreateSGPatch(in ec2.SecurityGroup, target v1beta1.SecurityGroupParameters) (*v1beta1.SecurityGroupParameters, error) {
	currentParams := &v1beta1.SecurityGroupParameters{}

	v1beta1.SortTags(target.Tags, in.Tags)
	LateInitializeSG(currentParams, &in)

	// NOTE(muvaf): Sending -1 as FromPort or ToPort is valid but the returned
	// object does not have that value. So, in case we have sent -1, we assume
	// that the returned value is also -1 in case if it's nil.
	// See the following about usage of -1
	// https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-ec2-security-group-egress.html
	mOne := int64(-1)
	for i, spec := range target.Egress {
		if len(currentParams.Egress) <= i {
			break
		}
		if awsgo.Int64Value(spec.FromPort) == mOne {
			currentParams.Egress[i].FromPort = awsclients.LateInitializeInt64Ptr(currentParams.Egress[i].FromPort, &mOne)
		}
		if awsgo.Int64Value(spec.ToPort) == mOne {
			currentParams.Egress[i].ToPort = awsclients.LateInitializeInt64Ptr(currentParams.Egress[i].ToPort, &mOne)
		}
	}

	jsonPatch, err := awsclients.CreateJSONPatch(*currentParams, target)
	if err != nil {
		return nil, err
	}
	patch := &v1beta1.SecurityGroupParameters{}
	if err := json.Unmarshal(jsonPatch, patch); err != nil {
		return nil, err
	}

	return patch, nil
}

// IsSGUpToDate checks whether there is a change in any of the modifiable fields.
func IsSGUpToDate(p v1beta1.SecurityGroupParameters, sg ec2.SecurityGroup) (bool, error) {
	patch, err := CreateSGPatch(sg, p)
	if err != nil {
		return false, err
	}
	return cmp.Equal(&v1beta1.SecurityGroupParameters{}, patch, cmpopts.IgnoreTypes(&v1alpha1.Reference{}, &v1alpha1.Selector{})), nil
}
