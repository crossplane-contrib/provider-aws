package iam

import (
	"context"
	"encoding/json"
	"net/url"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	iamtypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/mitchellh/copystructure"
	"github.com/pkg/errors"

	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane/provider-aws/apis/identity/v1beta1"
	awsclients "github.com/crossplane/provider-aws/pkg/clients"
)

const (
	errCheckUpToDate      = "unable to determine if external resource is up to date"
	errPolicyJSONEscape   = "malformed AssumeRolePolicyDocument JSON"
	errPolicyJSONUnescape = "malformed AssumeRolePolicyDocument escaping"
)

// RoleClient is the external client used for IAMRole Custom Resource
type RoleClient interface {
	GetRole(ctx context.Context, input *iam.GetRoleInput, opts ...func(*iam.Options)) (*iam.GetRoleOutput, error)
	CreateRole(ctx context.Context, input *iam.CreateRoleInput, opts ...func(*iam.Options)) (*iam.CreateRoleOutput, error)
	DeleteRole(ctx context.Context, input *iam.DeleteRoleInput, opts ...func(*iam.Options)) (*iam.DeleteRoleOutput, error)
	UpdateRole(ctx context.Context, input *iam.UpdateRoleInput, opts ...func(*iam.Options)) (*iam.UpdateRoleOutput, error)
	UpdateAssumeRolePolicy(ctx context.Context, input *iam.UpdateAssumeRolePolicyInput, opts ...func(*iam.Options)) (*iam.UpdateAssumeRolePolicyOutput, error)
	TagRole(ctx context.Context, input *iam.TagRoleInput, opts ...func(*iam.Options)) (*iam.TagRoleOutput, error)
	UntagRole(ctx context.Context, input *iam.UntagRoleInput, opts ...func(*iam.Options)) (*iam.UntagRoleOutput, error)
}

// NewRoleClient returns a new client using AWS credentials as JSON encoded data.
func NewRoleClient(conf aws.Config) RoleClient {
	return iam.NewFromConfig(conf)
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
		m.Tags = make([]iamtypes.Tag, len(p.Tags))
		for i := range p.Tags {
			m.Tags[i] = iamtypes.Tag{
				Key:   &p.Tags[i].Key,
				Value: &p.Tags[i].Value,
			}
		}
	}

	return m
}

// GenerateRoleObservation is used to produce IAMRoleExternalStatus from iamtypes.Role
func GenerateRoleObservation(role iamtypes.Role) v1beta1.IAMRoleExternalStatus {
	return v1beta1.IAMRoleExternalStatus{
		ARN:    aws.ToString(role.Arn),
		RoleID: aws.ToString(role.RoleId),
	}
}

// GenerateIAMRole assigns the in IAMRoleParamters to role.
func GenerateIAMRole(in v1beta1.IAMRoleParameters, role *iamtypes.Role) error {

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
		role.Tags = make([]iamtypes.Tag, len(in.Tags))
		for i := range in.Tags {
			role.Tags[i] = iamtypes.Tag{
				Key:   &in.Tags[i].Key,
				Value: &in.Tags[i].Value,
			}
		}
	}
	return nil
}

// LateInitializeRole fills the empty fields in *v1beta1.IAMRoleParameters with
// the values seen in iamtypes.Role.
func LateInitializeRole(in *v1beta1.IAMRoleParameters, role *iamtypes.Role) {
	if role == nil {
		return
	}
	in.AssumeRolePolicyDocument = awsclients.LateInitializeString(in.AssumeRolePolicyDocument, role.AssumeRolePolicyDocument)
	in.Description = awsclients.LateInitializeStringPtr(in.Description, role.Description)
	in.MaxSessionDuration = awsclients.LateInitializeInt32Ptr(in.MaxSessionDuration, role.MaxSessionDuration)
	in.Path = awsclients.LateInitializeStringPtr(in.Path, role.Path)

	if role.PermissionsBoundary != nil {
		in.PermissionsBoundary = awsclients.LateInitializeStringPtr(in.PermissionsBoundary, role.PermissionsBoundary.PermissionsBoundaryArn)
	}

	if in.Tags == nil && role.Tags != nil {
		for _, tag := range role.Tags {
			in.Tags = append(in.Tags, v1beta1.Tag{Key: aws.ToString(tag.Key), Value: aws.ToString(tag.Value)})
		}
	}
}

// CreatePatch creates a *v1beta1.IAMRoleParameters that has only the changed
// values between the target *v1beta1.IAMRoleParameters and the current
// *iamtypes.Role
func CreatePatch(in *iamtypes.Role, target *v1beta1.IAMRoleParameters) (*v1beta1.IAMRoleParameters, error) {
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

func isAssumeRolePolicyUpToDate(a, b *string) (bool, error) {
	if a == nil || b == nil {
		return a == b, nil
	}

	jsonA, err := url.QueryUnescape(*a)
	if err != nil {
		return false, errors.Wrap(err, errPolicyJSONUnescape)
	}

	jsonB, err := url.QueryUnescape(*b)
	if err != nil {
		return false, errors.Wrap(err, errPolicyJSONUnescape)
	}

	return awsclients.IsPolicyUpToDate(&jsonA, &jsonB), nil
}

// IsRoleUpToDate checks whether there is a change in any of the modifiable fields in role.
func IsRoleUpToDate(in v1beta1.IAMRoleParameters, observed iamtypes.Role) (bool, string, error) {
	generated, err := copystructure.Copy(&observed)
	if err != nil {
		return true, "", errors.Wrap(err, errCheckUpToDate)
	}
	desired, ok := generated.(*iamtypes.Role)
	if !ok {
		return true, "", errors.New(errCheckUpToDate)
	}

	if err = GenerateIAMRole(in, desired); err != nil {
		return false, "", err
	}

	policyUpToDate, err := isAssumeRolePolicyUpToDate(desired.AssumeRolePolicyDocument, observed.AssumeRolePolicyDocument)
	if err != nil {
		return false, "", err
	}

	diff := cmp.Diff(desired, &observed, cmpopts.IgnoreInterfaces(struct{ resource.AttributeReferencer }{}), cmpopts.IgnoreFields(observed, "AssumeRolePolicyDocument"))
	if diff == "" && policyUpToDate {
		return true, diff, nil
	}

	diff = "Found observed difference in IAM role\n" + diff

	// Add extra logging for AssumeRolePolicyDocument because cmp.Diff doesn't show the full difference
	if !policyUpToDate {
		diff += "\ndesired assume role policy: "
		diff += *desired.AssumeRolePolicyDocument
		diff += "\nobserved assume role policy: "
		diff += *observed.AssumeRolePolicyDocument
	}
	return false, diff, nil
}

// DiffIAMTags returns the lists of tags that need to be removed and added according
// to current and desired states.
func DiffIAMTags(local []v1beta1.Tag, remote []iamtypes.Tag) (add []iamtypes.Tag, remove []string) {
	addMap := make(map[string]string, len(local))
	for _, t := range local {
		addMap[t.Key] = t.Value
	}
	removeMap := map[string]struct{}{}
	for _, t := range remote {
		if addMap[aws.ToString(t.Key)] == aws.ToString(t.Value) {
			delete(addMap, aws.ToString(t.Key))
			continue
		}
		removeMap[aws.ToString(t.Key)] = struct{}{}
	}
	for k, v := range addMap {
		add = append(add, iamtypes.Tag{Key: aws.String(k), Value: aws.String(v)})
	}
	for k := range removeMap {
		remove = append(remove, k)
	}
	return add, remove
}
