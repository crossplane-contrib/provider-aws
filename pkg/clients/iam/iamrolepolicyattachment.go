package iam

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"

	"github.com/crossplane/stack-aws/apis/identity/v1beta1"
)

// RolePolicyAttachmentClient is the external client used for IAMRolePolicyAttachment Custom Resource
type RolePolicyAttachmentClient interface {
	AttachRolePolicyRequest(*iam.AttachRolePolicyInput) iam.AttachRolePolicyRequest
	ListAttachedRolePoliciesRequest(*iam.ListAttachedRolePoliciesInput) iam.ListAttachedRolePoliciesRequest
	DetachRolePolicyRequest(*iam.DetachRolePolicyInput) iam.DetachRolePolicyRequest
}

// NewRolePolicyAttachmentClient returns a new client given an aws config
func NewRolePolicyAttachmentClient(conf *aws.Config) (RolePolicyAttachmentClient, error) {
	return iam.New(*conf), nil
}

// GenerateAttachRolePolicyInput from IAMRolePolicyAttachmentSpec
func GenerateAttachRolePolicyInput(p *v1beta1.IAMRolePolicyAttachmentParameters) *iam.AttachRolePolicyInput {
	m := &iam.AttachRolePolicyInput{
		PolicyArn: aws.String(p.PolicyARN),
		RoleName:  aws.String(p.RoleName),
	}
	return m
}

// GenerateDetachRolePolicyInput from IAMRolePolicyAttachmentSpec
func GenerateDetachRolePolicyInput(p *v1beta1.IAMRolePolicyAttachmentParameters) *iam.DetachRolePolicyInput {
	m := &iam.DetachRolePolicyInput{
		PolicyArn: aws.String(p.PolicyARN),
		RoleName:  aws.String(p.RoleName),
	}
	return m
}

// GenerateUpdateRolePolicyInput from IAMRolePolicyAttachmentStatus
func GenerateUpdateRolePolicyInput(p *v1beta1.IAMRolePolicyAttachmentExternalStatus) *iam.DetachRolePolicyInput {
	m := &iam.DetachRolePolicyInput{
		PolicyArn: aws.String(p.AttachedPolicyARN),
	}
	return m
}

// GenerateRolePolicyObservation is used to produce IAMRolePolicyAttachmentExternalStatus from iam.AttachedPolicy
func GenerateRolePolicyObservation(policy iam.AttachedPolicy) v1beta1.IAMRolePolicyAttachmentExternalStatus {
	return v1beta1.IAMRolePolicyAttachmentExternalStatus{
		AttachedPolicyARN: aws.StringValue(policy.PolicyArn),
	}
}
