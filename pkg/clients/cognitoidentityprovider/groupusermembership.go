package cognitoidentityprovider

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"

	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"

	"errors"

	cognitoidentityprovidertypes "github.com/aws/aws-sdk-go-v2/service/acm/types"
)

// GroupUserMembershipClient is the external client used for GroupUserMembership Custom Resource
type GroupUserMembershipClient interface {
	AdminAddUserToGroup(ctx context.Context, input *cognitoidentityprovider.AdminAddUserToGroupInput, opts ...func(*cognitoidentityprovider.Options)) (*cognitoidentityprovider.AdminAddUserToGroupOutput, error)
	AdminRemoveUserFromGroup(ctx context.Context, input *cognitoidentityprovider.AdminRemoveUserFromGroupInput, opts ...func(*cognitoidentityprovider.Options)) (*cognitoidentityprovider.AdminRemoveUserFromGroupOutput, error)
	AdminListGroupsForUser(ctx context.Context, input *cognitoidentityprovider.AdminListGroupsForUserInput, opts ...func(*cognitoidentityprovider.Options)) (*cognitoidentityprovider.AdminListGroupsForUserOutput, error)
}

// NewGroupUserMembershipClient creates new Amazton Cognito GroupUserMembershipClient with provided AWS Configurations/Credentials
func NewGroupUserMembershipClient(cfg aws.Config) GroupUserMembershipClient {
	return cognitoidentityprovider.NewFromConfig(cfg)
}

// IsErrorNotFound returns true if the error code indicates that the item was not found
func IsErrorNotFound(err error) bool {
	var notFoundError *cognitoidentityprovidertypes.ResourceNotFoundException
	return errors.As(err, &notFoundError)
}
