package iam

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"

	"github.com/crossplane/provider-aws/apis/identity/v1beta1"
	awsclients "github.com/crossplane/provider-aws/pkg/clients"
)

// RolePolicyAttachmentClient is the external client used for IAMRolePolicyAttachment Custom Resource
type RolePolicyAttachmentClient interface {
	AttachRolePolicyRequest(*iam.AttachRolePolicyInput) iam.AttachRolePolicyRequest
	ListAttachedRolePoliciesRequest(*iam.ListAttachedRolePoliciesInput) iam.ListAttachedRolePoliciesRequest
	DetachRolePolicyRequest(*iam.DetachRolePolicyInput) iam.DetachRolePolicyRequest
}

// NewRolePolicyAttachmentClient returns a new client given an aws config
func NewRolePolicyAttachmentClient(conf aws.Config) RolePolicyAttachmentClient {
	return iam.New(conf)
}

// GenerateRolePolicyObservation is used to produce IAMRolePolicyAttachmentExternalStatus from iam.AttachedPolicy
func GenerateRolePolicyObservation(policy iam.AttachedPolicy) v1beta1.IAMRolePolicyAttachmentExternalStatus {
	return v1beta1.IAMRolePolicyAttachmentExternalStatus{
		AttachedPolicyARN: aws.StringValue(policy.PolicyArn),
	}
}

// LateInitializePolicy fills the empty fields in *v1beta1.IAMRolePolicyAttachmentParameters with
// the values seen in iam.AttachedPolicy.
func LateInitializePolicy(in *v1beta1.IAMRolePolicyAttachmentParameters, policy *iam.AttachedPolicy) {
	if policy == nil {
		return
	}
	in.PolicyARN = awsclients.LateInitializeString(in.PolicyARN, policy.PolicyArn)
}
