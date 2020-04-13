package iam

import (
	"context"
	"sort"

	"github.com/aws/aws-sdk-go-v2/service/iam"

	"github.com/crossplane/provider-aws/apis/identity/v1alpha1"
	aws "github.com/crossplane/provider-aws/pkg/clients"
	awsclients "github.com/crossplane/provider-aws/pkg/clients"
)

// UserClient is the external client used for IAM User Custom Resource
type UserClient interface {
	GetUserRequest(*iam.GetUserInput) iam.GetUserRequest
	CreateUserRequest(*iam.CreateUserInput) iam.CreateUserRequest
	UpdateUserRequest(*iam.UpdateUserInput) iam.UpdateUserRequest
	DeleteUserRequest(*iam.DeleteUserInput) iam.DeleteUserRequest
	ListGroupsForUserRequest(*iam.ListGroupsForUserInput) iam.ListGroupsForUserRequest
	AddUserToGroupRequest(*iam.AddUserToGroupInput) iam.AddUserToGroupRequest
	RemoveUserFromGroupRequest(*iam.RemoveUserFromGroupInput) iam.RemoveUserFromGroupRequest
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
func LateInitializeUser(in *v1alpha1.UserParameters, user *iam.User) {
	if user == nil {
		return
	}

	in.Path = awsclients.LateInitializeStringPtr(in.Path, user.Path)
	if user.PermissionsBoundary != nil {
		in.PermissionsBoundary = awsclients.LateInitializeStringPtr(in.PermissionsBoundary, user.PermissionsBoundary.PermissionsBoundaryArn)
	}

	for _, tag := range user.Tags {
		in.Tags = append(in.Tags, v1alpha1.Tag{Key: *tag.Key, Value: *tag.Value})
	}
}

// CompareGroups compares list of groups specified for v1alpha1.User and
// list of actual groups for iam.User
func CompareGroups(p v1alpha1.UserParameters, groups []iam.Group) bool {
	if len(p.GroupList) != len(groups) {
		return false
	}

	sort.Slice(p.GroupList, func(i, j int) bool {
		return p.GroupList[i] < p.GroupList[j]
	})
	sort.Slice(groups, func(i, j int) bool {
		return aws.StringValue(groups[i].GroupName) < aws.StringValue(groups[j].GroupName)
	})

	for i, val := range groups {
		if p.GroupList[i] != aws.StringValue(val.GroupName) {
			return false
		}
	}

	return true
}
