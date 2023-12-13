package iam

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
)

const (
	// ErrRolePolicyNotFound is the aws exception when the policy cannot be found on the role
	ErrRolePolicyNotFound = "NoSuchEntity"
)

// RolePolicyClient is the external client used for RolePolicy Custom Resource
type RolePolicyClient interface {
	GetRolePolicy(ctx context.Context, input *iam.GetRolePolicyInput, opts ...func(*iam.Options)) (*iam.GetRolePolicyOutput, error)
	PutRolePolicy(ctx context.Context, input *iam.PutRolePolicyInput, opts ...func(*iam.Options)) (*iam.PutRolePolicyOutput, error)
	DeleteRolePolicy(ctx context.Context, input *iam.DeleteRolePolicyInput, opts ...func(*iam.Options)) (*iam.DeleteRolePolicyOutput, error)
}

// NewRolePolicyClient returns a new client using AWS credentials as JSON encoded data.
func NewRolePolicyClient(conf aws.Config) RolePolicyClient {
	return iam.NewFromConfig(conf)
}
