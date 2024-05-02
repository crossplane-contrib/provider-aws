package iam

import (
	"context"
	"encoding/json"
	"net/url"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	iamtypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/aws/smithy-go/document"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"k8s.io/utils/ptr"

	"github.com/crossplane-contrib/provider-aws/apis/iam/v1beta1"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/iam/convert"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/jsonpatch"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/policy"
)

// RoleClient is the external client used for Role Custom Resource
type RoleClient interface {
	GetRole(ctx context.Context, input *iam.GetRoleInput, opts ...func(*iam.Options)) (*iam.GetRoleOutput, error)
	CreateRole(ctx context.Context, input *iam.CreateRoleInput, opts ...func(*iam.Options)) (*iam.CreateRoleOutput, error)
	DeleteRole(ctx context.Context, input *iam.DeleteRoleInput, opts ...func(*iam.Options)) (*iam.DeleteRoleOutput, error)
	UpdateRole(ctx context.Context, input *iam.UpdateRoleInput, opts ...func(*iam.Options)) (*iam.UpdateRoleOutput, error)
	PutRolePermissionsBoundary(ctx context.Context, params *iam.PutRolePermissionsBoundaryInput, optFns ...func(*iam.Options)) (*iam.PutRolePermissionsBoundaryOutput, error)
	DeleteRolePermissionsBoundary(ctx context.Context, params *iam.DeleteRolePermissionsBoundaryInput, optFns ...func(*iam.Options)) (*iam.DeleteRolePermissionsBoundaryOutput, error)
	UpdateAssumeRolePolicy(ctx context.Context, input *iam.UpdateAssumeRolePolicyInput, opts ...func(*iam.Options)) (*iam.UpdateAssumeRolePolicyOutput, error)
	TagRole(ctx context.Context, input *iam.TagRoleInput, opts ...func(*iam.Options)) (*iam.TagRoleOutput, error)
	UntagRole(ctx context.Context, input *iam.UntagRoleInput, opts ...func(*iam.Options)) (*iam.UntagRoleOutput, error)
}

// NewRoleClient returns a new client using AWS credentials as JSON encoded data.
func NewRoleClient(conf aws.Config) RoleClient {
	return iam.NewFromConfig(conf)
}

// GenerateCreateRoleInput from RoleSpec
func GenerateCreateRoleInput(name string, p *v1beta1.RoleParameters) *iam.CreateRoleInput {
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

// GenerateRoleObservation is used to produce RoleExternalStatus from iamtypes.Role
func GenerateRoleObservation(role iamtypes.Role) v1beta1.RoleExternalStatus {
	o := v1beta1.RoleExternalStatus{
		ARN:        pointer.StringValue(role.Arn),
		CreateDate: pointer.TimeToMetaTime(role.CreateDate),
		RoleID:     pointer.StringValue(role.RoleId),
	}

	if role.RoleLastUsed != nil {
		o.RoleLastUsed = &v1beta1.RoleLastUsed{
			LastUsedDate: pointer.TimeToMetaTime(role.RoleLastUsed.LastUsedDate),
			Region:       role.RoleLastUsed.Region,
		}
	}

	return o
}

// GenerateRole assigns the in RoleParamters to role.
func GenerateRole(in v1beta1.RoleParameters, role *iamtypes.Role) error {
	// iamtypes.Role has url-encoded policy document, while RoleParameters has plain.
	// Assign policy from `in` only if it is different from the one in `role`.
	if escapedPolicyDoc := role.AssumeRolePolicyDocument; escapedPolicyDoc != nil {
		policyDoc, err := url.QueryUnescape(*escapedPolicyDoc)
		if err != nil || !policy.ArePolicyDocumentsEqual(policyDoc, in.AssumeRolePolicyDocument) {
			role.AssumeRolePolicyDocument = nil
		}
	}
	if role.AssumeRolePolicyDocument == nil && in.AssumeRolePolicyDocument != "" {
		role.AssumeRolePolicyDocument = ptr.To(url.QueryEscape(in.AssumeRolePolicyDocument))
	}

	role.Description = in.Description
	role.MaxSessionDuration = in.MaxSessionDuration
	role.Path = in.Path
	if in.PermissionsBoundary != nil {
		role.PermissionsBoundary = &iamtypes.AttachedPermissionsBoundary{
			PermissionsBoundaryArn: in.PermissionsBoundary,
		}
	}

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

// LateInitializeRole fills the empty fields in *v1beta1.RoleParameters with
// the values seen in iamtypes.Role.
func LateInitializeRole(in *v1beta1.RoleParameters, role *iamtypes.Role) {
	if role == nil {
		return
	}
	in.AssumeRolePolicyDocument = pointer.LateInitializeValueFromPtr(in.AssumeRolePolicyDocument, role.AssumeRolePolicyDocument)
	in.Description = pointer.LateInitialize(in.Description, role.Description)
	in.MaxSessionDuration = pointer.LateInitialize(in.MaxSessionDuration, role.MaxSessionDuration)
	in.Path = pointer.LateInitialize(in.Path, role.Path)

	if role.PermissionsBoundary != nil {
		in.PermissionsBoundary = pointer.LateInitialize(in.PermissionsBoundary, role.PermissionsBoundary.PermissionsBoundaryArn)
	}

	if in.Tags == nil && role.Tags != nil {
		for _, tag := range role.Tags {
			in.Tags = append(in.Tags, v1beta1.Tag{Key: aws.ToString(tag.Key), Value: aws.ToString(tag.Value)})
		}
	}
}

// CreatePatch creates a *v1beta1.RoleParameters that has only the changed
// values between the target *v1beta1.RoleParameters and the current
// *iamtypes.Role
func CreatePatch(in *iamtypes.Role, target *v1beta1.RoleParameters) (*v1beta1.RoleParameters, error) {
	currentParams := &v1beta1.RoleParameters{}
	LateInitializeRole(currentParams, in)

	jsonPatch, err := jsonpatch.CreateJSONPatch(currentParams, target)
	if err != nil {
		return nil, err
	}
	patch := &v1beta1.RoleParameters{}
	if err := json.Unmarshal(jsonPatch, patch); err != nil {
		return nil, err
	}
	return patch, nil
}

// IsRoleUpToDate checks whether there is a change in any of the modifiable fields in role.
func IsRoleUpToDate(in v1beta1.RoleParameters, observed iamtypes.Role) (bool, string, error) {
	desired := (&convert.ConverterImpl{}).DeepCopyAWSRole(&observed)
	if err := GenerateRole(in, desired); err != nil {
		return false, "", err
	}

	diff := cmp.Diff(desired, &observed,
		cmpopts.IgnoreInterfaces(struct{ resource.AttributeReferencer }{}),
		cmpopts.IgnoreFields(observed, "CreateDate", "PermissionsBoundary.PermissionsBoundaryType", "RoleLastUsed"),
		cmpopts.IgnoreTypes(document.NoSerde{}), cmpopts.SortSlices(lessTag))
	if diff == "" {
		return true, diff, nil
	}

	diff = "Found observed difference in IAM role\n" + diff
	return false, diff, nil
}

func lessTag(a, b iamtypes.Tag) bool {
	if a.Key == nil {
		return b.Key != nil
	}
	if b.Key == nil {
		return false
	}
	return *a.Key <= *b.Key
}
