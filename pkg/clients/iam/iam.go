package iam

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	iamtypes "github.com/aws/aws-sdk-go-v2/service/iam/types"

	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane/provider-aws/apis/identity/v1alpha1"
)

const (
	policyArn = "arn:aws:iam::%s:policy/%s"
)

// Client defines IAM Client operations
// mockery -case snake -name Client -output fake -outpkg fake
type Client interface {
	CreateUser(username string) (*iamtypes.AccessKey, error)
	DeleteUser(username string) error
	CreatePolicyAndAttach(username string, policyName string, policyDocument string) (string, error)
	GetPolicyVersion(policyName string) (string, error)
	UpdatePolicy(policyName string, policyDocument string) (string, error)
	DeletePolicyAndDetach(username string, policyName string) error
	GetAccountID() (string, error)
}

type iamClient struct {
	accountID *string
	iam       *iam.Client
}

// NewClient creates new AWS Client with provided AWS Configurations/Credentials
func NewClient(config aws.Config) Client {
	return &iamClient{iam: iam.NewFromConfig(config)}
}

// CreateUser - Creates an IAM User, a policy, binds user to policy and returns an access key and policy version for the user.
func (c *iamClient) CreateUser(username string) (*iamtypes.AccessKey, error) {
	err := c.createUser(username)
	if err != nil {
		return nil, fmt.Errorf("failed to create user, %s", err)
	}

	key, err := c.createAccessKey(username)
	if err != nil {
		return nil, fmt.Errorf("failed to create access key, %s", err)
	}

	return key, err
}

// CreatePolicyAndAttach - Creates the IAM policy and attaches it to the username
func (c *iamClient) CreatePolicyAndAttach(username string, policyName string, policyDocument string) (string, error) {
	currentVersion, err := c.createPolicy(username, policyDocument)
	if err != nil {
		return "", fmt.Errorf("failed to create policy, %s", err)
	}

	err = c.attachPolicyToUser(username, username)
	if err != nil {
		return "", fmt.Errorf("failed to attach policy, %s", err)
	}

	return currentVersion, nil
}

// GetPolicyVersion get the policy document for the IAM user
func (c *iamClient) GetPolicyVersion(username string) (string, error) {
	policyARN, err := c.getPolicyARN(username)
	if err != nil {
		return "", err
	}

	policyResponse, err := c.iam.GetPolicy(context.TODO(), &iam.GetPolicyInput{
		PolicyArn: aws.String(policyARN),
	})

	if err != nil {
		return "", err
	}

	return aws.ToString(policyResponse.Policy.DefaultVersionId), nil
}

// UpdatePolicy - updates the policy document for the IAM user and return current policy version
func (c *iamClient) UpdatePolicy(policyName string, policyDocument string) (string, error) {
	policyARN, err := c.getPolicyARN(policyName)
	if err != nil {
		return "", err
	}
	// Create a new policy version
	policyVersionResponse, err := c.iam.CreatePolicyVersion(context.TODO(), &iam.CreatePolicyVersionInput{PolicyArn: aws.String(policyARN), PolicyDocument: aws.String(policyDocument), SetAsDefault: true})
	if err != nil {
		return "", err
	}

	currentPolicyVersion := policyVersionResponse.PolicyVersion.VersionId
	// Delete old versions of policy - Max 5 allowed
	policyVersions, err := c.iam.ListPolicyVersions(context.TODO(), &iam.ListPolicyVersionsInput{PolicyArn: aws.String(policyARN)})
	if err != nil {
		return "", err
	}

	for _, policy := range policyVersions.Versions {
		if aws.ToString(policy.VersionId) != aws.ToString(currentPolicyVersion) {
			_, err := c.iam.DeletePolicyVersion(context.TODO(), &iam.DeletePolicyVersionInput{PolicyArn: aws.String(policyARN), VersionId: policy.VersionId})
			if err != nil {
				return "", err
			}
		}
	}

	return aws.ToString(currentPolicyVersion), nil
}

// DeletePolicyAndDetach delete the policy of PolicyName and detach it from the username provided
func (c *iamClient) DeletePolicyAndDetach(username string, policyName string) error {
	policyARN, err := c.getPolicyARN(username)
	if resource.Ignore(IsErrorNotFound, err) != nil {
		return err
	}
	if IsErrorNotFound(err) {
		return nil
	}

	_, err = c.iam.DetachUserPolicy(context.TODO(), &iam.DetachUserPolicyInput{PolicyArn: aws.String(policyARN), UserName: aws.String(username)})
	if resource.Ignore(IsErrorNotFound, err) != nil {
		return err
	}

	_, err = c.iam.DeletePolicy(context.TODO(), &iam.DeletePolicyInput{PolicyArn: aws.String(policyARN)})
	return resource.Ignore(IsErrorNotFound, err)
}

// DeleteUser Policy and IAM User
func (c *iamClient) DeleteUser(username string) error {
	keys, err := c.iam.ListAccessKeys(context.TODO(), &iam.ListAccessKeysInput{UserName: aws.String(username)})
	if resource.Ignore(IsErrorNotFound, err) != nil {
		return err
	}
	if keys != nil {
		for _, key := range keys.AccessKeyMetadata {
			_, err = c.iam.DeleteAccessKey(context.TODO(), &iam.DeleteAccessKeyInput{AccessKeyId: key.AccessKeyId, UserName: aws.String(username)})
			if resource.Ignore(IsErrorNotFound, err) != nil {
				return err
			}
		}
	}

	_, err = c.iam.DeleteUser(context.TODO(), &iam.DeleteUserInput{UserName: aws.String(username)})
	return resource.Ignore(IsErrorNotFound, err)
}

// GetAccountID - Gets the accountID of the authenticated session.
func (c *iamClient) GetAccountID() (string, error) {
	if c.accountID == nil {
		user, err := c.iam.GetUser(context.TODO(), &iam.GetUserInput{})
		if err != nil {
			return "", err
		}

		arnData, err := arn.Parse(*user.User.Arn)
		if err != nil {
			return "", err
		}
		c.accountID = &arnData.AccountID
	}

	return aws.ToString(c.accountID), nil
}

func (c *iamClient) getPolicyARN(policyName string) (string, error) {
	accountID, err := c.GetAccountID()
	if err != nil {
		return "", err
	}
	policyARN := fmt.Sprintf(policyArn, accountID, policyName)
	return policyARN, nil
}

func (c *iamClient) createUser(username string) error {
	_, err := c.iam.CreateUser(context.TODO(), &iam.CreateUserInput{UserName: aws.String(username)})
	if err != nil && isErrorAlreadyExists(err) {
		return nil
	}
	return err
}

func (c *iamClient) createAccessKey(username string) (*iamtypes.AccessKey, error) {
	keysResponse, err := c.iam.CreateAccessKey(context.TODO(), &iam.CreateAccessKeyInput{UserName: aws.String(username)})
	if err != nil {
		return nil, err
	}

	return keysResponse.AccessKey, nil
}

func (c *iamClient) createPolicy(policyName string, policyDocument string) (string, error) {
	response, err := c.iam.CreatePolicy(context.TODO(), &iam.CreatePolicyInput{PolicyName: aws.String(policyName), PolicyDocument: aws.String(policyDocument)})
	if err != nil {
		if isErrorAlreadyExists(err) {
			return c.UpdatePolicy(policyName, policyDocument)
		}
		return "", err
	}
	return aws.ToString(response.Policy.DefaultVersionId), nil
}

func (c *iamClient) attachPolicyToUser(policyName string, username string) error {
	policyArn, err := c.getPolicyARN(policyName)
	if err != nil {
		return err
	}
	_, err = c.iam.AttachUserPolicy(context.TODO(), &iam.AttachUserPolicyInput{PolicyArn: aws.String(policyArn), UserName: aws.String(username)})
	return err
}

func isErrorAlreadyExists(err error) bool {
	var iee *iamtypes.EntityAlreadyExistsException
	return errors.As(err, &iee)
}

// IsErrorNotFound returns true if the error code indicates that the item was not found
func IsErrorNotFound(err error) bool {
	var nse *iamtypes.NoSuchEntityException
	return errors.As(err, &nse)
}

// NewErrorNotFound returns an aws error with error code indicating the item was not found.
func NewErrorNotFound() error {
	var nse *iamtypes.NoSuchEntityException
	return errors.New(*nse.Message)
}

// PolicyDocument is the structure of IAM policy document
type PolicyDocument struct {
	Version   string
	Statement []StatementEntry
}

// StatementEntry is used to define permission statements in a PolicyDocument
type StatementEntry struct {
	Sid      string
	Effect   string
	Action   []string
	Resource []string
}

// BuildIAMTags build a tag array with type that IAM client expects.
func BuildIAMTags(tags []v1alpha1.Tag) []iamtypes.Tag {
	res := make([]iamtypes.Tag, len(tags))
	for i, t := range tags {
		res[i] = iamtypes.Tag{Key: aws.String(t.Key), Value: aws.String(t.Value)}
	}
	return res
}
