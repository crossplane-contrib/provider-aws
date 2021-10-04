package iam

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"

	"github.com/aws/aws-sdk-go-v2/service/iam"
	iamtypes "github.com/aws/aws-sdk-go-v2/service/iam/types"

	"github.com/crossplane/provider-aws/apis/identity/v1alpha1"
	awsclients "github.com/crossplane/provider-aws/pkg/clients"
)

// UserClient is the external client used for IAM User Custom Resource
type UserClient interface {
	GetUser(ctx context.Context, input *iam.GetUserInput, opts ...func(*iam.Options)) (*iam.GetUserOutput, error)
	CreateUser(ctx context.Context, input *iam.CreateUserInput, opts ...func(*iam.Options)) (*iam.CreateUserOutput, error)
	DeleteUser(ctx context.Context, input *iam.DeleteUserInput, opts ...func(*iam.Options)) (*iam.DeleteUserOutput, error)
	UpdateUser(ctx context.Context, input *iam.UpdateUserInput, opts ...func(*iam.Options)) (*iam.UpdateUserOutput, error)
}

// NewUserClient returns a new client using AWS credentials as JSON encoded data.
func NewUserClient(cfg aws.Config) UserClient {
	return iam.NewFromConfig(cfg)
}

// LateInitializeUser fills the empty fields in *v1alpha1.User with
// the values seen in iam.User.
func LateInitializeUser(in *v1alpha1.IAMUserParameters, user *iamtypes.User) {
	if user == nil {
		return
	}

	in.Path = awsclients.LateInitializeStringPtr(in.Path, user.Path)
	if user.PermissionsBoundary != nil {
		in.PermissionsBoundary = awsclients.LateInitializeStringPtr(in.PermissionsBoundary, user.PermissionsBoundary.PermissionsBoundaryArn)
	}

	if in.Tags == nil && user.Tags != nil {
		for _, tag := range user.Tags {
			in.Tags = append(in.Tags, v1alpha1.Tag{Key: *tag.Key, Value: *tag.Value})
		}
	}
}
