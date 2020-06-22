package acmpca

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/acmpca"
)

// CAPermissionClient defines the CertificateManager operations
type CAPermissionClient interface {
	CreatePermissionRequest(*acmpca.CreatePermissionInput) acmpca.CreatePermissionRequest
	DeletePermissionRequest(*acmpca.DeletePermissionInput) acmpca.DeletePermissionRequest
	ListPermissionsRequest(*acmpca.ListPermissionsInput) acmpca.ListPermissionsRequest
}

// NewCAPermissionClient returns a new client using AWS credentials as JSON encoded data.
func NewCAPermissionClient(conf *aws.Config) (CAPermissionClient, error) {
	return acmpca.New(*conf), nil
}
