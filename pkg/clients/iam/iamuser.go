package iam

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/iam"

	"github.com/crossplane/provider-aws/apis/identity/v1alpha1"
	awsclients "github.com/crossplane/provider-aws/pkg/clients"
)

// UserClient is the external client used for IAM User Custom Resource
type UserClient interface {
	GetUserRequest(*iam.GetUserInput) iam.GetUserRequest
	CreateUserRequest(*iam.CreateUserInput) iam.CreateUserRequest
	UpdateUserRequest(*iam.UpdateUserInput) iam.UpdateUserRequest
	DeleteUserRequest(*iam.DeleteUserInput) iam.DeleteUserRequest
}

// NewUserClient returns a new client using AWS credentials as JSON encoded data.
func NewUserClient(ctx context.Context, credentials []byte, region string, auth awsclients.AuthMethod) (UserClient, error) {
	cfg, err := auth(ctx, credentials, awsclients.DefaultSection, region)
	if cfg == nil {
		return nil, err
	}
	return iam.New(*cfg), nil
}

// LateInitializeUser fills the empty fields in *v1alpha1.User with
// the values seen in iam.User.
func LateInitializeUser(in *v1alpha1.IAMUserParameters, user *iam.User) {
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
