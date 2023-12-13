package iam

import (
	"context"
	"net/url"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	iamtypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"

	"github.com/crossplane-contrib/provider-aws/apis/iam/v1beta1"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
	policyutils "github.com/crossplane-contrib/provider-aws/pkg/utils/policy"
)

// PolicyClient is the external client used for Policy Custom Resource
type PolicyClient interface {
	GetPolicy(ctx context.Context, input *iam.GetPolicyInput, opts ...func(*iam.Options)) (*iam.GetPolicyOutput, error)
	CreatePolicy(ctx context.Context, input *iam.CreatePolicyInput, opts ...func(*iam.Options)) (*iam.CreatePolicyOutput, error)
	DeletePolicy(ctx context.Context, input *iam.DeletePolicyInput, opts ...func(*iam.Options)) (*iam.DeletePolicyOutput, error)
	GetPolicyVersion(ctx context.Context, input *iam.GetPolicyVersionInput, opts ...func(*iam.Options)) (*iam.GetPolicyVersionOutput, error)
	CreatePolicyVersion(ctx context.Context, input *iam.CreatePolicyVersionInput, opts ...func(*iam.Options)) (*iam.CreatePolicyVersionOutput, error)
	ListPolicyVersions(ctx context.Context, input *iam.ListPolicyVersionsInput, opts ...func(*iam.Options)) (*iam.ListPolicyVersionsOutput, error)
	DeletePolicyVersion(ctx context.Context, input *iam.DeletePolicyVersionInput, opts ...func(*iam.Options)) (*iam.DeletePolicyVersionOutput, error)
	TagPolicy(ctx context.Context, input *iam.TagPolicyInput, opts ...func(*iam.Options)) (*iam.TagPolicyOutput, error)
	UntagPolicy(ctx context.Context, input *iam.UntagPolicyInput, opts ...func(*iam.Options)) (*iam.UntagPolicyOutput, error)
}

// STSClient is the external client used for STS
type STSClient interface {
	GetCallerIdentity(ctx context.Context, input *sts.GetCallerIdentityInput, opts ...func(*sts.Options)) (*sts.GetCallerIdentityOutput, error)
}

// NewPolicyClient returns a new client using AWS credentials as JSON encoded data.
func NewPolicyClient(cfg aws.Config) PolicyClient {
	return iam.NewFromConfig(cfg)
}

// NewSTSClient creates a new STS Client.
func NewSTSClient(cfg aws.Config) STSClient {
	return sts.NewFromConfig(cfg)
}

// IsPolicyUpToDate checks whether there is a change in any of the modifiable fields in policy.
func IsPolicyUpToDate(in v1beta1.PolicyParameters, policy iamtypes.PolicyVersion) (bool, string, error) {
	externalPolicyRaw := pointer.StringValue(policy.Document)
	if externalPolicyRaw == "" || in.Document == "" {
		return false, "", nil
	}

	return IsPolicyDocumentUpToDate(in.Document, policy.Document)
}

// IsPolicyDocumentUpToDate checks whether there is a change in any of the modifiable fields in policy.
func IsPolicyDocumentUpToDate(in string, policy *string) (bool, string, error) {

	unescapedPolicy, err := url.QueryUnescape(aws.ToString(policy))
	if err != nil {
		return false, "", err
	}
	externpolicy, err := policyutils.ParsePolicyString(unescapedPolicy)
	if err != nil {
		return false, "", err
	}
	specPolicy, err := policyutils.ParsePolicyString(in)
	if err != nil {
		return false, "", err
	}

	areEqual, diff := policyutils.ArePoliciesEqal(&specPolicy, &externpolicy)
	return areEqual, diff, nil
}

// ValidatePolicyObject tries to parse the raw policy into a Policy object.
func ValidatePolicyObject(policy string) error {
	_, err := policyutils.ParsePolicyString(policy)
	return err
}
