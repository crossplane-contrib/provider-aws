package iam

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
)

// GroupPolicyAttachmentClient is the external client used for GroupPolicyAttachment Custom Resource
type GroupPolicyAttachmentClient interface {
	AttachGroupPolicy(ctx context.Context, input *iam.AttachGroupPolicyInput, opts ...func(*iam.Options)) (*iam.AttachGroupPolicyOutput, error)
	ListAttachedGroupPolicies(ctx context.Context, input *iam.ListAttachedGroupPoliciesInput, opts ...func(*iam.Options)) (*iam.ListAttachedGroupPoliciesOutput, error)
	DetachGroupPolicy(ctx context.Context, input *iam.DetachGroupPolicyInput, opts ...func(*iam.Options)) (*iam.DetachGroupPolicyOutput, error)
}

// NewGroupPolicyAttachmentClient creates new RDS RDSClient with provided AWS Configurations/Credentials
func NewGroupPolicyAttachmentClient(cfg aws.Config) GroupPolicyAttachmentClient {
	return iam.NewFromConfig(cfg)
}
