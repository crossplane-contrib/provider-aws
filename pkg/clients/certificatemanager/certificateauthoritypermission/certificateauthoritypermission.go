package certificateauthoritypermission

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/awserr"
	"github.com/aws/aws-sdk-go-v2/service/acmpca"
)

// Client defines the CertificateManager operations
type Client interface {
	CreatePermissionRequest(*acmpca.CreatePermissionInput) acmpca.CreatePermissionRequest
	DeletePermissionRequest(*acmpca.DeletePermissionInput) acmpca.DeletePermissionRequest
	ListPermissionsRequest(*acmpca.ListPermissionsInput) acmpca.ListPermissionsRequest
}

// NewClient returns a new client using AWS credentials as JSON encoded data.
func NewClient(conf *aws.Config) (Client, error) {
	return acmpca.New(*conf), nil
}

// IsErrorNotFound returns true if the error code indicates that the item was not found
func IsErrorNotFound(err error) bool {
	if awsErr, ok := err.(awserr.Error); ok {
		if awsErr.Code() == acmpca.ErrCodeInvalidStateException {
			return true
		}
	}

	return false
}
