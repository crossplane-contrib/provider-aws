package iam

import (
	"context"
	"net/url"

	"github.com/crossplane-contrib/provider-aws/apis/iam/v1beta1"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	iamtypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/google/go-cmp/cmp"

	awsclients "github.com/crossplane-contrib/provider-aws/pkg/clients"
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
func IsPolicyUpToDate(in v1beta1.PolicyParameters, policy iamtypes.PolicyVersion) (bool, error) {
	// The AWS API returns Policy Document as an escaped string.
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

	compactPolicy, err := awsclients.CompactAndEscapeJSON(unescapedPolicy)
	if err != nil {
		return false, err
	}
	compactSpecPolicy, err := awsclients.CompactAndEscapeJSON(in.Document)
	if err != nil {
		return false, err
	}

	return cmp.Equal(compactPolicy, compactSpecPolicy), nil
}
