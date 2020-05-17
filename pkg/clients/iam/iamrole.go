package iam

import (
	"encoding/json"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/mitchellh/copystructure"
	"github.com/pkg/errors"

	"github.com/crossplane/provider-aws/apis/identity/v1beta1"
	awsclients "github.com/crossplane/provider-aws/pkg/clients"
)

const (
	errCheckUpToDate    = "unable to determine if external resource is up to date"
	errPolicyJSONEscape = "malformed AssumeRolePolicyDocument JSON"
)

// RoleClient is the external client used for IAMRole Custom Resource
type RoleClient interface {
	GetRoleRequest(*iam.GetRoleInput) iam.GetRoleRequest
	CreateRoleRequest(*iam.CreateRoleInput) iam.CreateRoleRequest
	DeleteRoleRequest(*iam.DeleteRoleInput) iam.DeleteRoleRequest
	UpdateRoleRequest(*iam.UpdateRoleInput) iam.UpdateRoleRequest
	UpdateAssumeRolePolicyRequest(*iam.UpdateAssumeRolePolicyInput) iam.UpdateAssumeRolePolicyRequest
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

// GenerateRoleObservation is used to produce IAMRoleExternalStatus from iam.Role
func GenerateRoleObservation(role iam.Role) v1beta1.IAMRoleExternalStatus {
	return v1beta1.IAMRoleExternalStatus{
		ARN:    aws.StringValue(role.Arn),
		RoleID: aws.StringValue(role.RoleId),
	}
}

// GenerateIAMRole assigns the in IAMRoleParamters to role.
func GenerateIAMRole(in v1beta1.IAMRoleParameters, role *iam.Role) error {

	if in.AssumeRolePolicyDocument != "" {
		s, err := awsclients.CompactAndEscapeJSON(in.AssumeRolePolicyDocument)
		if err != nil {
			return errors.Wrap(err, errPolicyJSONEscape)
		}

		role.AssumeRolePolicyDocument = &s
	}
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
	return nil
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

	if role.PermissionsBoundary != nil {
		in.PermissionsBoundary = awsclients.LateInitializeStringPtr(in.PermissionsBoundary, role.PermissionsBoundary.PermissionsBoundaryArn)
	}

	if in.Tags == nil && role.Tags != nil {
		for _, tag := range role.Tags {
			in.Tags = append(in.Tags, v1beta1.Tag{Key: *tag.Key, Value: *tag.Value})
		}
	}
}

// CreatePatch creates a *v1beta1.IAMRoleParameters that has only the changed
// values between the target *v1beta1.IAMRoleParameters and the current
// *iam.Role
func CreatePatch(in *iam.Role, target *v1beta1.IAMRoleParameters) (*v1beta1.IAMRoleParameters, error) {
	currentParams := &v1beta1.IAMRoleParameters{}
	LateInitializeRole(currentParams, in)

	jsonPatch, err := awsclients.CreateJSONPatch(currentParams, target)
	if err != nil {
		return nil, err
	}
	patch := &v1beta1.IAMRoleParameters{}
	if err := json.Unmarshal(jsonPatch, patch); err != nil {
		return nil, err
	}
	return patch, nil
}

// IsRoleUpToDate checks whether there is a change in any of the modifiable fields in role.
func IsRoleUpToDate(in v1beta1.IAMRoleParameters, observed iam.Role) (bool, error) {
	generated, err := copystructure.Copy(&observed)
	if err != nil {
		return true, errors.Wrap(err, errCheckUpToDate)
	}
	desired, ok := generated.(*iam.Role)
	if !ok {
		return true, errors.New(errCheckUpToDate)
	}

	if err = GenerateIAMRole(in, desired); err != nil {
		return false, err
	}

	return cmp.Equal(desired, &observed, cmpopts.IgnoreInterfaces(struct{ resource.AttributeReferencer }{})), nil
}
