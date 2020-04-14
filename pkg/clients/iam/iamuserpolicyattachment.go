package iam

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/iam"

	"github.com/crossplane/provider-aws/apis/identity/v1alpha1"
	awsclients "github.com/crossplane/provider-aws/pkg/clients"
)

// UserPolicyAttachmentClient is the external client used for UserPolicyAttachment Custom Resource
type UserPolicyAttachmentClient interface {
	AttachUserPolicyRequest(*iam.AttachUserPolicyInput) iam.AttachUserPolicyRequest
	DetachUserPolicyRequest(*iam.DetachUserPolicyInput) iam.DetachUserPolicyRequest
	ListAttachedUserPoliciesRequest(*iam.ListAttachedUserPoliciesInput) iam.ListAttachedUserPoliciesRequest
}

// NewUserPolicyAttachmentClient creates new RDS RDSClient with provided AWS Configurations/Credentials
func NewUserPolicyAttachmentClient(ctx context.Context, credentials []byte, region string, auth awsclients.AuthMethod) (UserPolicyAttachmentClient, error) {
	cfg, err := auth(ctx, credentials, awsclients.DefaultSection, region)
	if cfg == nil {
		return nil, err
	}
	return iam.New(*cfg), err
}

// LateInitializeUserPolicy fills the empty fields in v1alpha1.UserPolicyAttachmentParameters with
// the values seen in iam.AttachedPolicy.
func LateInitializeUserPolicy(in *v1alpha1.UserPolicyAttachmentParameters, policy *iam.AttachedPolicy) {
	if policy == nil {
		return
	}
	in.PolicyARN = awsclients.LateInitializeString(in.PolicyARN, policy.PolicyArn)
}
