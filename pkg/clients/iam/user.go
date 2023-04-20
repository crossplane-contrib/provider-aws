package iam

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	iamtypes "github.com/aws/aws-sdk-go-v2/service/iam/types"

	"github.com/crossplane-contrib/provider-aws/apis/iam/v1beta1"
	awsclients "github.com/crossplane-contrib/provider-aws/pkg/clients"
)

// UserClient is the external client used for IAM User Custom Resource
type UserClient interface {
	GetUser(ctx context.Context, input *iam.GetUserInput, opts ...func(*iam.Options)) (*iam.GetUserOutput, error)
	CreateUser(ctx context.Context, input *iam.CreateUserInput, opts ...func(*iam.Options)) (*iam.CreateUserOutput, error)
	DeleteUser(ctx context.Context, input *iam.DeleteUserInput, opts ...func(*iam.Options)) (*iam.DeleteUserOutput, error)
	UpdateUser(ctx context.Context, input *iam.UpdateUserInput, opts ...func(*iam.Options)) (*iam.UpdateUserOutput, error)
	PutUserPermissionsBoundary(ctx context.Context, params *iam.PutUserPermissionsBoundaryInput, optFns ...func(*iam.Options)) (*iam.PutUserPermissionsBoundaryOutput, error)
	TagUser(ctx context.Context, params *iam.TagUserInput, opts ...func(*iam.Options)) (*iam.TagUserOutput, error)
	UntagUser(ctx context.Context, params *iam.UntagUserInput, opts ...func(*iam.Options)) (*iam.UntagUserOutput, error)
}

// NewUserClient returns a new client using AWS credentials as JSON encoded data.
func NewUserClient(cfg aws.Config) UserClient {
	return iam.NewFromConfig(cfg)
}

// LateInitializeUser fills the empty fields in *v1alpha1.User with
// the values seen in iam.User.
func LateInitializeUser(in *v1beta1.UserParameters, user *iamtypes.User) {
	if user == nil {
		return
	}

	in.Path = awsclients.LateInitializeStringPtr(in.Path, user.Path)
	if user.PermissionsBoundary != nil {
		in.PermissionsBoundary = awsclients.LateInitializeStringPtr(in.PermissionsBoundary, user.PermissionsBoundary.PermissionsBoundaryArn)
	}

	if in.Tags == nil && user.Tags != nil {
		for _, tag := range user.Tags {
			in.Tags = append(in.Tags, v1beta1.Tag{Key: *tag.Key, Value: *tag.Value})
		}
	}
}

// IsUserUpToDate checks whether there is a change in any of the modifiable fields in user.
func IsUserUpToDate(in v1beta1.UserParameters, observed iamtypes.User) (bool, string, error) {
	if aws.ToString(in.Path) != aws.ToString(observed.Path) {
		return false, "paths differ", nil
	}

	if observed.PermissionsBoundary == nil {
		if aws.ToString(in.PermissionsBoundary) != "" {
			return false, "permission boundary needs set", nil
		}
		return true, "", nil
	}

	return aws.ToString(in.PermissionsBoundary) == aws.ToString(observed.PermissionsBoundary.PermissionsBoundaryArn), "permission boundary", nil
}
