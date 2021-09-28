package iam

import (
	"context"
	"net/url"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	iamtypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/google/go-cmp/cmp"

	"github.com/crossplane/provider-aws/apis/identity/v1alpha1"
	awsclients "github.com/crossplane/provider-aws/pkg/clients"
)

// PolicyClient is the external client used for IAMPolicy Custom Resource
type PolicyClient interface {
	GetPolicy(ctx context.Context, input *iam.GetPolicyInput, opts ...func(*iam.Options)) (*iam.GetPolicyOutput, error)
	CreatePolicy(ctx context.Context, input *iam.CreatePolicyInput, opts ...func(*iam.Options)) (*iam.CreatePolicyOutput, error)
	DeletePolicy(ctx context.Context, input *iam.DeletePolicyInput, opts ...func(*iam.Options)) (*iam.DeletePolicyOutput, error)
	GetPolicyVersion(ctx context.Context, input *iam.GetPolicyVersionInput, opts ...func(*iam.Options)) (*iam.GetPolicyVersionOutput, error)
	CreatePolicyVersion(ctx context.Context, input *iam.CreatePolicyVersionInput, opts ...func(*iam.Options)) (*iam.CreatePolicyVersionOutput, error)
	ListPolicyVersions(ctx context.Context, input *iam.ListPolicyVersionsInput, opts ...func(*iam.Options)) (*iam.ListPolicyVersionsOutput, error)
	DeletePolicyVersion(ctx context.Context, input *iam.DeletePolicyVersionInput, opts ...func(*iam.Options)) (*iam.DeletePolicyVersionOutput, error)
}

// NewPolicyClient returns a new client using AWS credentials as JSON encoded data.
func NewPolicyClient(cfg aws.Config) PolicyClient {
	return iam.NewFromConfig(cfg)
}

// IsPolicyUpToDate checks whether there is a change in any of the modifiable fields in policy.
func IsPolicyUpToDate(in v1alpha1.IAMPolicyParameters, policy iamtypes.PolicyVersion) (bool, error) {
	// The AWS API reutrns Policy Document as an escaped string.
	// Due to differences in the methods to escape a string, the comparison result between
	// the spec.Document and policy.Document can sometimes be false negative (due to spaces, line feeds).
	// Escaping with a common method and then comparing is a safe way.

	if *policy.Document == "" || in.Document == "" {
		return false, nil
	}

	unescapedPolicy, err := url.QueryUnescape(aws.ToString(policy.Document))
	if err != nil {
		return false, nil
	}

	compactIAMPolicy, err := awsclients.CompactAndEscapeJSON(unescapedPolicy)
	if err != nil {
		return false, err
	}
	compactSpecPolicy, err := awsclients.CompactAndEscapeJSON(in.Document)
	if err != nil {
		return false, err
	}

	return cmp.Equal(compactIAMPolicy, compactSpecPolicy), nil
}
