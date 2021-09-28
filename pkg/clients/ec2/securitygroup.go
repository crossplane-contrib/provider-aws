package ec2

import (
	"context"
	"encoding/json"
	"sort"
	"strings"

	awsgo "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/smithy-go"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"

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
	if awsErr, ok := err.(smithy.APIError); ok {
		if awsErr.ErrorCode() == InvalidGroupNotFound {
			return true
		}
	}
	return false
}

// IsRuleAlreadyExistsErr returns true if the error is because the rule already exists.
func IsRuleAlreadyExistsErr(err error) bool {
	if awsErr, ok := err.(smithy.APIError); ok {
		if awsErr.ErrorCode() == InvalidPermissionDuplicate {
			return true
		}
	}
	return false
}

// GenerateEC2Permissions converts object Permissions to ec2 format
func GenerateEC2Permissions(objectPerms []v1beta1.IPPermission) []ec2types.IpPermission {
	if len(objectPerms) == 0 {
		return nil
	}
	permissions := make([]ec2types.IpPermission, len(objectPerms))
	for i, p := range objectPerms {
		ipPerm := ec2types.IpPermission{
			FromPort:   *p.FromPort,
			IpProtocol: aws.String(p.IPProtocol),
			ToPort:     *p.ToPort,
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

// GenerateIPPermissions converts object EC2 Permissions to IPPermission format
func GenerateIPPermissions(objectPerms []ec2types.IpPermission) []v1beta1.IPPermission {
	if len(objectPerms) == 0 {
		return nil
	}
	permissions := make([]v1beta1.IPPermission, len(objectPerms))
	for i, p := range objectPerms {
		ipPerm := v1beta1.IPPermission{
			FromPort:   awsgo.Int32(p.FromPort),
			IPProtocol: aws.StringValue(p.IpProtocol),
			ToPort:     awsgo.Int32(p.ToPort),
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
	sort.Slice(permissions, func(i, j int) bool {
		return awsgo.ToInt32(permissions[i].FromPort) < awsgo.ToInt32(permissions[j].FromPort)
	})
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
func LateInitializeSG(in *v1beta1.SecurityGroupParameters, sg *ec2types.SecurityGroup) { // nolint:gocyclo
	if sg == nil {
		return
	}

	in.Description = awsclients.LateInitializeString(in.Description, sg.Description)
	in.GroupName = awsclients.LateInitializeString(in.GroupName, sg.GroupName)
	in.VPCID = awsclients.LateInitializeStringPtr(in.VPCID, sg.VpcId)

	// If there is a mismatch in lengths, then it's a sign for desire change
	// that should be handled via add or remove.
	if len(in.Egress) == len(sg.IpPermissionsEgress) {
		in.Egress = LateInitializeIPPermissions(in.Egress, sg.IpPermissionsEgress)
	}
	if len(in.Ingress) == len(sg.IpPermissions) {
		in.Ingress = LateInitializeIPPermissions(in.Ingress, sg.IpPermissions)
	}

	if len(in.Tags) == 0 && len(sg.Tags) != 0 {
		in.Tags = v1beta1.BuildFromEC2Tags(sg.Tags)
	}
}

// LateInitializeIPPermissions returns an array of []v1beta1.IPPermission whose
// empty optional fields are filled with what's observed in []ec2types.IpPermission.
//
// Note that since there is no unique identifier for each IPPermission, its order
// is assumed to be stable and used indexes are used as identifier.
func LateInitializeIPPermissions(spec []v1beta1.IPPermission, o []ec2types.IpPermission) []v1beta1.IPPermission { // nolint:gocyclo
	if len(spec) < len(o) {
		return spec
	}
	for i := range o {
		spec[i].FromPort = awsclients.LateInitializeInt32Ptr(spec[i].FromPort, &o[i].FromPort)
		spec[i].ToPort = awsclients.LateInitializeInt32Ptr(spec[i].FromPort, &o[i].ToPort)
		spec[i].IPProtocol = awsclients.LateInitializeString(spec[i].IPProtocol, o[i].IpProtocol)

		for j := range o[i].IpRanges {
			if len(spec[i].IPRanges) == j {
				spec[i].IPRanges = append(spec[i].IPRanges, v1beta1.IPRange{})
			}
			spec[i].IPRanges[j].Description = awsclients.LateInitializeStringPtr(
				spec[i].IPRanges[j].Description,
				o[i].IpRanges[j].Description,
			)
		}
		for j := range o[i].Ipv6Ranges {
			if len(spec[i].IPv6Ranges) == j {
				spec[i].IPv6Ranges = append(spec[i].IPv6Ranges, v1beta1.IPv6Range{})
			}
			spec[i].IPv6Ranges[j].Description = awsclients.LateInitializeStringPtr(
				spec[i].IPv6Ranges[j].Description,
				o[i].Ipv6Ranges[j].Description,
			)
		}
		for j := range o[i].PrefixListIds {
			if len(spec[i].PrefixListIDs) == j {
				spec[i].PrefixListIDs = append(spec[i].PrefixListIDs, v1beta1.PrefixListID{})
			}
			spec[i].PrefixListIDs[j].Description = awsclients.LateInitializeStringPtr(
				spec[i].PrefixListIDs[j].Description,
				o[i].PrefixListIds[j].Description,
			)
		}
		for j := range o[i].UserIdGroupPairs {
			if len(spec[i].UserIDGroupPairs) == j {
				spec[i].UserIDGroupPairs = append(spec[i].UserIDGroupPairs, v1beta1.UserIDGroupPair{})
			}
			spec[i].UserIDGroupPairs[j].Description = awsclients.LateInitializeStringPtr(
				spec[i].UserIDGroupPairs[j].Description,
				o[i].UserIdGroupPairs[j].Description,
			)
			spec[i].UserIDGroupPairs[j].GroupID = awsclients.LateInitializeStringPtr(
				spec[i].UserIDGroupPairs[j].GroupID,
				o[i].UserIdGroupPairs[j].GroupId,
			)
			spec[i].UserIDGroupPairs[j].GroupName = awsclients.LateInitializeStringPtr(
				spec[i].UserIDGroupPairs[j].GroupName,
				o[i].UserIdGroupPairs[j].GroupName,
			)
			spec[i].UserIDGroupPairs[j].UserID = awsclients.LateInitializeStringPtr(
				spec[i].UserIDGroupPairs[j].UserID,
				o[i].UserIdGroupPairs[j].UserId,
			)
			spec[i].UserIDGroupPairs[j].VPCID = awsclients.LateInitializeStringPtr(
				spec[i].UserIDGroupPairs[j].VPCID,
				o[i].UserIdGroupPairs[j].VpcId,
			)
			spec[i].UserIDGroupPairs[j].VPCPeeringConnectionID = awsclients.LateInitializeStringPtr(
				spec[i].UserIDGroupPairs[j].VPCPeeringConnectionID,
				o[i].UserIdGroupPairs[j].VpcPeeringConnectionId,
			)
		}
	}
	return spec
}

// CreateSGPatch creates a *v1beta1.SecurityGroupParameters that has only the changed
// values between the target *v1beta1.SecurityGroupParameters and the current
// *ec2types.SecurityGroup
func CreateSGPatch(in ec2types.SecurityGroup, target v1beta1.SecurityGroupParameters) (*v1beta1.SecurityGroupParameters, error) { // nolint:gocyclo
	v1beta1.SortTags(target.Tags, in.Tags)
	currentParams := &v1beta1.SecurityGroupParameters{
		Description: awsclients.StringValue(in.Description),
		GroupName:   awsclients.StringValue(in.GroupName),
		VPCID:       in.VpcId,
	}
	currentParams.Tags = v1beta1.BuildFromEC2Tags(in.Tags)
	currentParams.Ingress = GenerateIPPermissions(in.IpPermissions)
	currentParams.Egress = GenerateIPPermissions(in.IpPermissionsEgress)
	// NOTE(muvaf): Sending -1 as FromPort or ToPort is valid but the returned
	// object does not have that value. So, in case we have sent -1, we assume
	// that the returned value is also -1 in case if it's nil.
	// See the following about usage of -1
	// https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-ec2-security-group-egress.html
	mOne := int32(-1)
	for i, spec := range target.Egress {
		if len(currentParams.Egress) <= i {
			break
		}
		if awsgo.ToInt32(spec.FromPort) == mOne {
			currentParams.Egress[i].FromPort = awsclients.LateInitializeInt32Ptr(currentParams.Egress[i].FromPort, &mOne)
		}
		if awsgo.ToInt32(spec.ToPort) == mOne {
			currentParams.Egress[i].ToPort = awsclients.LateInitializeInt32Ptr(currentParams.Egress[i].ToPort, &mOne)
		}
	}
	// Same happens with VPCID in egress user group id pairs. The value of that
	// field is not returned from AWS.
	for i, ingress := range currentParams.Ingress {
		for j, pair := range ingress.UserIDGroupPairs {
			if awsclients.StringValue(pair.VPCID) == "" && len(target.Ingress) > i && len(target.Ingress[i].UserIDGroupPairs) > j {
				currentParams.Ingress[i].UserIDGroupPairs[j].VPCID = target.Ingress[i].UserIDGroupPairs[j].VPCID
			}
		}
	}

	sort.Slice(target.Egress, func(i, j int) bool {
		return awsgo.ToInt32(target.Egress[i].FromPort) < awsgo.ToInt32(target.Egress[j].FromPort)
	})
	sort.Slice(target.Ingress, func(i, j int) bool {
		return awsgo.ToInt32(target.Ingress[i].FromPort) < awsgo.ToInt32(target.Ingress[j].FromPort)
	})

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
func IsSGUpToDate(p v1beta1.SecurityGroupParameters, sg ec2types.SecurityGroup) (bool, error) {
	patch, err := CreateSGPatch(sg, p)
	if err != nil {
		return false, err
	}
	return cmp.Equal(&v1beta1.SecurityGroupParameters{}, patch,
		cmpopts.IgnoreTypes(&xpv1.Reference{}, &xpv1.Selector{}),
		cmpopts.IgnoreFields(v1beta1.SecurityGroupParameters{}, "Region"),
		InsensitiveCases()), nil
}

// TODO(muvaf): We needed this for IPProtocol field; even if you send "TCP", AWS
// returns "tcp". However, this cmp.Option is probably useful for other providers,
// too. Consider making it part of crossplane-runtime.

// InsensitiveCases ignores the case sensitivity for string and *string types.
func InsensitiveCases() cmp.Option {
	return cmp.Options{
		cmp.FilterValues(func(_, _ interface{}) bool {
			return true
		}, cmp.Comparer(strings.EqualFold)),
		cmp.FilterValues(func(_, _ interface{}) bool {
			return true
		}, cmp.Comparer(func(x, y *string) bool {
			return strings.EqualFold(awsclients.StringValue(x), awsclients.StringValue(y))
		})),
	}
}
