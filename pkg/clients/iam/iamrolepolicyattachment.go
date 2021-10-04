package iam

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	iamtypes "github.com/aws/aws-sdk-go-v2/service/iam/types"

	"github.com/crossplane/provider-aws/apis/identity/v1beta1"
	awsclients "github.com/crossplane/provider-aws/pkg/clients"
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

// GenerateRolePolicyObservation is used to produce IAMRolePolicyAttachmentExternalStatus from iamtypes.AttachedPolicy
func GenerateRolePolicyObservation(policy iamtypes.AttachedPolicy) v1beta1.IAMRolePolicyAttachmentExternalStatus {
	return v1beta1.IAMRolePolicyAttachmentExternalStatus{
		AttachedPolicyARN: aws.ToString(policy.PolicyArn),
	}
}

// LateInitializePolicy fills the empty fields in *v1beta1.IAMRolePolicyAttachmentParameters with
// the values seen in iamtypes.AttachedPolicy.
func LateInitializePolicy(in *v1beta1.IAMRolePolicyAttachmentParameters, policy *iamtypes.AttachedPolicy) {
	if policy == nil {
		return
	}
	in.PolicyARN = awsclients.LateInitializeString(in.PolicyARN, policy.PolicyArn)
}
