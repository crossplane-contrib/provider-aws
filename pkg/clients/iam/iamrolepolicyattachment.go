package iam

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
)

// RolePolicyAttachmentClient is the external client used for IAMRolePolicyAttachment Custom Resource
type RolePolicyAttachmentClient interface {
	AttachRolePolicy(ctx context.Context, input *iam.AttachRolePolicyInput, opts ...func(*iam.Options)) (*iam.AttachRolePolicyOutput, error)
	ListAttachedRolePolicies(ctx context.Context, input *iam.ListAttachedRolePoliciesInput, opts ...func(*iam.Options)) (*iam.ListAttachedRolePoliciesOutput, error)
	DetachRolePolicy(ctx context.Context, input *iam.DetachRolePolicyInput, opts ...func(*iam.Options)) (*iam.DetachRolePolicyOutput, error)
}

// NewRolePolicyAttachmentClient returns a new client given an aws config
func NewRolePolicyAttachmentClient(conf aws.Config) RolePolicyAttachmentClient {
	return iam.NewFromConfig(conf)
}
