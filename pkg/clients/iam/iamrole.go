package iam

import (
	"encoding/json"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/mitchellh/copystructure"
	"github.com/pkg/errors"

	"github.com/crossplane/crossplane-runtime/pkg/resource"

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
	TagRoleRequest(input *iam.TagRoleInput) iam.TagRoleRequest
	UntagRoleRequest(input *iam.UntagRoleInput) iam.UntagRoleRequest
}

// NewRoleClient returns a new client using AWS credentials as JSON encoded data.
func NewRoleClient(conf aws.Config) RoleClient {
	return iam.New(conf)
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
		for i := range p.Tags {
			m.Tags[i] = iam.Tag{
				Key:   aws.String(p.Tags[i].Key),
				Value: aws.String(p.Tags[i].Key),
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
		for i := range in.Tags {
			role.Tags[i] = iam.Tag{
				Key:   aws.String(in.Tags[i].Key),
				Value: aws.String(in.Tags[i].Value),
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
			in.Tags = append(in.Tags, v1beta1.Tag{Key: aws.StringValue(tag.Key), Value: aws.StringValue(tag.Value)})
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
func IsRoleUpToDate(in v1beta1.IAMRoleParameters, observed iam.Role) (bool, string, error) {
	generated, err := copystructure.Copy(&observed)
	if err != nil {
		return true, "", errors.Wrap(err, errCheckUpToDate)
	}
	desired, ok := generated.(*iam.Role)
	if !ok {
		return true, "", errors.New(errCheckUpToDate)
	}

	if err = GenerateIAMRole(in, desired); err != nil {
		return false, "", err
	}

	isRoleUpToDate := cmp.Equal(desired, &observed, cmpopts.IgnoreInterfaces(struct{ resource.AttributeReferencer }{}))
	observationMessage := ""

	if !isRoleUpToDate {
		observationMessage = "Found observed difference in IAM role\n"
		observationMessage += cmp.Diff(desired, &observed, cmpopts.IgnoreInterfaces(struct{ resource.AttributeReferencer }{}))

		// Add extra logging for AssumeRolePolicyDocument because cmp.Diff doesn't show the full difference
		if *desired.AssumeRolePolicyDocument != *observed.AssumeRolePolicyDocument {
			observationMessage += "\ndesired assume role policy: "
			observationMessage += *desired.AssumeRolePolicyDocument
			observationMessage += "\nobserved assume role policy: "
			observationMessage += *observed.AssumeRolePolicyDocument
		}
	}

	return isRoleUpToDate, observationMessage, nil
}

// DiffIAMTags returns the lists of tags that need to be removed and added according
// to current and desired states.
func DiffIAMTags(local []v1beta1.Tag, remote []iam.Tag) (add []iam.Tag, remove []string) {
	addMap := make(map[string]string, len(local))
	for _, t := range local {
		addMap[t.Key] = t.Value
	}
	removeMap := map[string]struct{}{}
	for _, t := range remote {
		if addMap[aws.StringValue(t.Key)] == aws.StringValue(t.Value) {
			delete(addMap, aws.StringValue(t.Key))
			continue
		}
		removeMap[aws.StringValue(t.Key)] = struct{}{}
	}
	for k, v := range addMap {
		add = append(add, iam.Tag{Key: aws.String(k), Value: aws.String(v)})
	}
	for k := range removeMap {
		remove = append(remove, k)
	}
	return add, remove
}
