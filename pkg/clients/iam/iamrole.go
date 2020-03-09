package iam

import (
	"net/url"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/mitchellh/copystructure"
	"github.com/pkg/errors"

	"github.com/crossplane/provider-aws/apis/identity/v1beta1"
	awsclients "github.com/crossplane/provider-aws/pkg/clients"
)

const errCheckUpToDate = "unable to determine if external resource is up to date"

// RoleClient is the external client used for IAMRole Custom Resource
type RoleClient interface {
	GetRoleRequest(*iam.GetRoleInput) iam.GetRoleRequest
	CreateRoleRequest(*iam.CreateRoleInput) iam.CreateRoleRequest
	DeleteRoleRequest(*iam.DeleteRoleInput) iam.DeleteRoleRequest
	UpdateRoleRequest(*iam.UpdateRoleInput) iam.UpdateRoleRequest
}

// NewRoleClient returns a new client using AWS credentials as JSON encoded data.
func NewRoleClient(conf *aws.Config) (RoleClient, error) {
	return iam.New(*conf), nil
}

// GenerateCreateRoleInput from IAMRoleSpec
func GenerateCreateRoleInput(name string, p *v1beta1.IAMRoleParameters) *iam.CreateRoleInput {
	m := &iam.CreateRoleInput{
		RoleName:                 aws.String(name),
		AssumeRolePolicyDocument: aws.String(p.AssumeRolePolicyDocument),
		Description:              p.Description,
		MaxSessionDuration:       p.MaxSessionDuration,
		Path:                     p.Path,
		PermissionsBoundary:      p.PermissionsBoundary,
	}

	if len(p.Tags) != 0 {
		m.Tags = make([]iam.Tag, len(p.Tags))
		for i, val := range p.Tags {
			m.Tags[i] = iam.Tag{
				Key:   &val.Key,
				Value: &val.Value,
			}
		}
	}

	return m
}

// GenerateUpdateRoleInput from IAMRoleSpec
func GenerateUpdateRoleInput(name string, p *v1beta1.IAMRoleParameters) *iam.UpdateRoleInput {
	m := &iam.UpdateRoleInput{
		RoleName:           aws.String(name),
		Description:        p.Description,
		MaxSessionDuration: p.MaxSessionDuration,
	}
	return m
}

// GenerateDeleteRoleInput from IAMRoleSpec
func GenerateDeleteRoleInput(name string) *iam.DeleteRoleInput {
	m := &iam.DeleteRoleInput{
		RoleName: aws.String(name),
	}
	return m
}

// GenerateRoleObservation is used to produce IAMRoleExternalStatus from iam.Role
func GenerateRoleObservation(role iam.Role) v1beta1.IAMRoleExternalStatus {
	return v1beta1.IAMRoleExternalStatus{
		ARN:    aws.StringValue(role.Arn),
		RoleID: aws.StringValue(role.RoleId),
	}
}

// GenerateIAMRole assigns the in IAMRoleParamters to role.
func GenerateIAMRole(in v1beta1.IAMRoleParameters, role *iam.Role) {
	s := strings.ReplaceAll(url.PathEscape(in.AssumeRolePolicyDocument), " ", "")
	role.AssumeRolePolicyDocument = &s
	role.Description = in.Description
	role.MaxSessionDuration = in.MaxSessionDuration
	role.Path = in.Path

	if len(in.Tags) != 0 {
		role.Tags = make([]iam.Tag, len(in.Tags))
		for i, val := range in.Tags {
			role.Tags[i] = iam.Tag{
				Key:   &val.Key,
				Value: &val.Value,
			}
		}
	}
}

// LateInitializeRole fills the empty fields in *v1beta1.IAMRoleParameters with
// the values seen in iam.Role.
func LateInitializeRole(in *v1beta1.IAMRoleParameters, role *iam.Role) {
	if role == nil {
		return
	}
	in.AssumeRolePolicyDocument = awsclients.LateInitializeString(in.AssumeRolePolicyDocument, role.AssumeRolePolicyDocument)
	in.Description = awsclients.LateInitializeStringPtr(in.Description, role.Description)
	in.MaxSessionDuration = awsclients.LateInitializeInt64Ptr(in.MaxSessionDuration, role.MaxSessionDuration)
	in.Path = awsclients.LateInitializeStringPtr(in.Path, role.Path)

	if in.PermissionsBoundary != nil {
		in.PermissionsBoundary = awsclients.LateInitializeStringPtr(in.PermissionsBoundary, role.PermissionsBoundary.PermissionsBoundaryArn)
	}

	for _, tag := range in.Tags {
		role.Tags = append(role.Tags, iam.Tag{Key: &tag.Key, Value: &tag.Value})
	}
}

// IsRoleUpToDate checks whether there is a change in any of the modifiable fields in role.
func IsRoleUpToDate(in *v1beta1.IAMRoleParameters, observed *iam.Role) (bool, error) {
	generated, err := copystructure.Copy(observed)
	if err != nil {
		return true, errors.Wrap(err, errCheckUpToDate)
	}
	desired, ok := generated.(*iam.Role)
	if !ok {
		return true, errors.New(errCheckUpToDate)
	}

	GenerateIAMRole(*in, desired)

	// 'AssumeRolePolicyDocument' is an escaped string in iam.Role and a normal string  in v1beta.IAMRole.
	// There is no proper way to compare them.
	return cmp.Equal(desired, observed, cmpopts.IgnoreFields(iam.Role{}, "AssumeRolePolicyDocument")), nil
}
