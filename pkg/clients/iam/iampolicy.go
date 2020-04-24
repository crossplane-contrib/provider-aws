package iam

import (
	"context"
	"net/url"

	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/google/go-cmp/cmp"

	"github.com/crossplane/provider-aws/apis/identity/v1alpha1"
	aws "github.com/crossplane/provider-aws/pkg/clients"
	awsclients "github.com/crossplane/provider-aws/pkg/clients"
)

// PolicyClient is the external client used for IAMPolicy Custom Resource
type PolicyClient interface {
	CreatePolicyRequest(*iam.CreatePolicyInput) iam.CreatePolicyRequest
	GetPolicyRequest(*iam.GetPolicyInput) iam.GetPolicyRequest
	DeletePolicyRequest(*iam.DeletePolicyInput) iam.DeletePolicyRequest
	GetPolicyVersionRequest(*iam.GetPolicyVersionInput) iam.GetPolicyVersionRequest
	CreatePolicyVersionRequest(*iam.CreatePolicyVersionInput) iam.CreatePolicyVersionRequest
	ListPolicyVersionsRequest(*iam.ListPolicyVersionsInput) iam.ListPolicyVersionsRequest
	DeletePolicyVersionRequest(*iam.DeletePolicyVersionInput) iam.DeletePolicyVersionRequest
}

// NewPolicyClient returns a new client using AWS credentials as JSON encoded data.
func NewPolicyClient(ctx context.Context, credentials []byte, region string, auth awsclients.AuthMethod) (PolicyClient, error) {
	cfg, err := auth(ctx, credentials, awsclients.DefaultSection, region)
	if cfg == nil {
		return nil, err
	}
	return iam.New(*cfg), nil
}

// IsPolicyUpToDate checks whether there is a change in any of the modifiable fields in policy.
func IsPolicyUpToDate(in v1alpha1.IAMPolicyParameters, policy iam.PolicyVersion) (bool, error) {
	// The AWS API reutrns Policy Document as an escaped string.
	// Due to differences in the methods to escape a string, the comparison result between
	// the spec.Document and policy.Document can sometimes be false negative (due to spaces, line feeds).
	// Escaping with a common method and then comparing is a safe way.

	if *policy.Document == "" || in.Document == "" {
		return false, nil
	}

	unescapedPolicy, err := url.QueryUnescape(aws.StringValue(policy.Document))
	if err != nil {
		return false, nil
	}

	compactIAMPolicy, err := awsclients.CompactAndEscapeJSON(unescapedPolicy)
	if err != nil {
		return false, err
	}
	compactSpecPolicy, err := aws.CompactAndEscapeJSON(in.Document)
	if err != nil {
		return false, err
	}

	return cmp.Equal(compactIAMPolicy, compactSpecPolicy), nil
}
