package iam

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"

	"github.com/aws/aws-sdk-go-v2/service/iam"
	iamtypes "github.com/aws/aws-sdk-go-v2/service/iam/types"

	"github.com/crossplane/provider-aws/apis/identity/v1alpha1"
	awsclients "github.com/crossplane/provider-aws/pkg/clients"
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

// LateInitializeGroupPolicy fills the empty fields in v1alpha1.GroupPolicyAttachmentParameters with
// the values seen in iamtypes.AttachedPolicy.
func LateInitializeGroupPolicy(in *v1alpha1.IAMGroupPolicyAttachmentParameters, policy *iamtypes.AttachedPolicy) {
	if policy == nil {
		return
	}
	in.PolicyARN = awsclients.LateInitializeString(in.PolicyARN, policy.PolicyArn)
}
