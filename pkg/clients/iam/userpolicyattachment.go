package iam

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
)

// UserPolicyAttachmentClient is the external client used for UserPolicyAttachment Custom Resource
type UserPolicyAttachmentClient interface {
	AttachUserPolicy(ctx context.Context, input *iam.AttachUserPolicyInput, opts ...func(*iam.Options)) (*iam.AttachUserPolicyOutput, error)
	ListAttachedUserPolicies(ctx context.Context, input *iam.ListAttachedUserPoliciesInput, opts ...func(*iam.Options)) (*iam.ListAttachedUserPoliciesOutput, error)
	DetachUserPolicy(ctx context.Context, input *iam.DetachUserPolicyInput, opts ...func(*iam.Options)) (*iam.DetachUserPolicyOutput, error)
}

// NewUserPolicyAttachmentClient creates new RDS RDSClient with provided AWS Configurations/Credentials
func NewUserPolicyAttachmentClient(cfg aws.Config) UserPolicyAttachmentClient {
	return iam.NewFromConfig(cfg)
}
