package iam

import (
	"github.com/aws/aws-sdk-go-v2/aws"

	"github.com/aws/aws-sdk-go-v2/service/iam"
)

// GroupUserMembershipClient is the external client used for GroupUserMembership Custom Resource
type GroupUserMembershipClient interface {
	AddUserToGroupRequest(*iam.AddUserToGroupInput) iam.AddUserToGroupRequest
	RemoveUserFromGroupRequest(*iam.RemoveUserFromGroupInput) iam.RemoveUserFromGroupRequest
	ListGroupsForUserRequest(*iam.ListGroupsForUserInput) iam.ListGroupsForUserRequest
}

// NewGroupUserMembershipClient creates new RDS RDSClient with provided AWS Configurations/Credentials
func NewGroupUserMembershipClient(cfg aws.Config) GroupUserMembershipClient {
	return iam.New(cfg)
}
