package iam

import (
	"github.com/aws/aws-sdk-go-v2/aws"

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
func NewGroupPolicyAttachmentClient(cfg aws.Config) GroupPolicyAttachmentClient {
	return iam.New(cfg)
}

// LateInitializeGroupPolicy fills the empty fields in v1alpha1.GroupPolicyAttachmentParameters with
// the values seen in iam.AttachedPolicy.
func LateInitializeGroupPolicy(in *v1alpha1.IAMGroupPolicyAttachmentParameters, policy *iam.AttachedPolicy) {
	if policy == nil {
		return
	}
	in.PolicyARN = awsclients.LateInitializeString(in.PolicyARN, policy.PolicyArn)
}
