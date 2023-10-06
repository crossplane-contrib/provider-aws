package iam

import (
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	iamtypes "github.com/aws/aws-sdk-go-v2/service/iam/types"

	"github.com/crossplane-contrib/provider-aws/apis/iam/v1beta1"
)

// IsErrorNotFound returns true if the error code indicates that the item was not found
func IsErrorNotFound(err error) bool {
	var notFoundError *iamtypes.NoSuchEntityException
	return errors.As(err, &notFoundError)
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
func BuildIAMTags(tags []v1beta1.Tag) []iamtypes.Tag {
	res := make([]iamtypes.Tag, len(tags))
	for i, t := range tags {
		res[i] = iamtypes.Tag{Key: aws.String(t.Key), Value: aws.String(t.Value)}
	}
	return res
}
