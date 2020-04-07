package ec2

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-sdk-go-v2/aws/awserr"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/crossplane/provider-aws/apis/network/v1beta1"
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

// GenerateSGObservation is used to produce v1beta1.SecurityGroupExternalStatus from
// ec2.SecurityGroup.
func GenerateSGObservation(sg ec2.SecurityGroup) v1beta1.SecurityGroupObservation {
	o := v1beta1.SecurityGroupObservation{
		OwnerID:         aws.StringValue(sg.OwnerId),
		SecurityGroupID: aws.StringValue(sg.GroupId),
	}

	if len(sg.Tags) > 0 {
		o.Tags = v1beta1.BuildFromEC2Tags(sg.Tags)
	}

	return o
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
		in.Egress = v1beta1.BuildIPPermissions(sg.IpPermissionsEgress)
	}

	if len(in.Ingress) == 0 && len(sg.IpPermissions) != 0 {
		in.Ingress = v1beta1.BuildIPPermissions(sg.IpPermissions)
	}
}

// CreateSGPatch creates a *v1beta1.SecurityGroupParameters that has only the changed
// values between the target *v1beta1.SecurityGroupParameters and the current
// *ec2.SecurityGroup
func CreateSGPatch(in *ec2.SecurityGroup, target v1beta1.SecurityGroupParameters) (*v1beta1.SecurityGroupParameters, error) {
	currentParams := &v1beta1.SecurityGroupParameters{}
	LateInitializeSG(currentParams, in)

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
	patch, err := CreateSGPatch(&sg, p)
	if err != nil {
		return false, err
	}
	return cmp.Equal(&v1beta1.SecurityGroupParameters{}, patch, cmpopts.IgnoreInterfaces(struct{ resource.AttributeReferencer }{})), nil
}
