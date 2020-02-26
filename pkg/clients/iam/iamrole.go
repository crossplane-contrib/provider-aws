package iam

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"

	"github.com/crossplane/stack-aws/apis/identity/v1beta1"
)

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
		Description:              aws.String(p.Description),
	}
	return m
}

// GenerateUpdateRoleInput from IAMRoleSpec
func GenerateUpdateRoleInput(name string, p *v1beta1.IAMRoleParameters) *iam.UpdateRoleInput {
	m := &iam.UpdateRoleInput{
		RoleName:    aws.String(name),
		Description: aws.String(p.Description),
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

// UpdateRoleExternalStatus updates the external status object, given the observation
func UpdateRoleExternalStatus(cr *v1beta1.IAMRole, observed iam.Role) {
	cr.Status.AtProvider = v1beta1.IAMRoleExternalStatus{
		ARN:    aws.StringValue(observed.Arn),
		RoleID: aws.StringValue(observed.RoleId),
	}
}
