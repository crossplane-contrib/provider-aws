package iam

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
)

// GroupUserMembershipClient is the external client used for GroupUserMembership Custom Resource
type GroupUserMembershipClient interface {
	AddUserToGroup(ctx context.Context, input *iam.AddUserToGroupInput, opts ...func(*iam.Options)) (*iam.AddUserToGroupOutput, error)
	RemoveUserFromGroup(ctx context.Context, input *iam.RemoveUserFromGroupInput, opts ...func(*iam.Options)) (*iam.RemoveUserFromGroupOutput, error)
	ListGroupsForUser(ctx context.Context, input *iam.ListGroupsForUserInput, opts ...func(*iam.Options)) (*iam.ListGroupsForUserOutput, error)
}

// NewGroupUserMembershipClient creates new RDS RDSClient with provided AWS Configurations/Credentials
func NewGroupUserMembershipClient(cfg aws.Config) GroupUserMembershipClient {
	return iam.NewFromConfig(cfg)
}
