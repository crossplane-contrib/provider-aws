package iam

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/iam"

	"github.com/crossplane/provider-aws/apis/identity/v1alpha1"
	awsclients "github.com/crossplane/provider-aws/pkg/clients"
)

// GroupPolicyAttachmentClient is the external client used for GroupPolicyAttachment Custom Resource
type GroupPolicyAttachmentClient interface {
	AttachGroupPolicyRequest(*iam.AttachGroupPolicyInput) iam.AttachGroupPolicyRequest
	DetachGroupPolicyRequest(*iam.DetachGroupPolicyInput) iam.DetachGroupPolicyRequest
	ListAttachedGroupPoliciesRequest(*iam.ListAttachedGroupPoliciesInput) iam.ListAttachedGroupPoliciesRequest
}

// NewGroupPolicyAttachmentClient creates new RDS RDSClient with provided AWS Configurations/Credentials
func NewGroupPolicyAttachmentClient(ctx context.Context, credentials []byte, region string, auth awsclients.AuthMethod) (GroupPolicyAttachmentClient, error) {
	cfg, err := auth(ctx, credentials, awsclients.DefaultSection, region)
	if cfg == nil {
		return nil, err
	}
	return iam.New(*cfg), err
}

// LateInitializeGroupPolicy fills the empty fields in v1alpha1.GroupPolicyAttachmentParameters with
// the values seen in iam.AttachedPolicy.
func LateInitializeGroupPolicy(in *v1alpha1.IAMGroupPolicyAttachmentParameters, policy *iam.AttachedPolicy) {
	if policy == nil {
		return
	}
	in.PolicyARN = awsclients.LateInitializeString(in.PolicyARN, policy.PolicyArn)
}
