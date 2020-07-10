package iam

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/iam"

	awsclients "github.com/crossplane/provider-aws/pkg/clients"
)

// GroupUserMembershipClient is the external client used for GroupUserMembership Custom Resource
type GroupUserMembershipClient interface {
	AddUserToGroupRequest(*iam.AddUserToGroupInput) iam.AddUserToGroupRequest
	RemoveUserFromGroupRequest(*iam.RemoveUserFromGroupInput) iam.RemoveUserFromGroupRequest
	ListGroupsForUserRequest(*iam.ListGroupsForUserInput) iam.ListGroupsForUserRequest
}

// NewGroupUserMembershipClient creates new RDS RDSClient with provided AWS Configurations/Credentials
func NewGroupUserMembershipClient(ctx context.Context, credentials []byte, region string, auth awsclients.AuthMethod) (GroupUserMembershipClient, error) {
	cfg, err := auth(ctx, credentials, awsclients.DefaultSection, region)
	if cfg == nil {
		return nil, err
	}
	return iam.New(*cfg), err
}
