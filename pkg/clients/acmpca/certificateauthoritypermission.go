/*
Copyright 2020 The Crossplane Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package acmpca

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/acmpca"
)

// CAPermissionClient defines the CertificateManager operations
type CAPermissionClient interface {
	CreatePermission(context.Context, *acmpca.CreatePermissionInput, ...func(*acmpca.Options)) (*acmpca.CreatePermissionOutput, error)
	DeletePermission(context.Context, *acmpca.DeletePermissionInput, ...func(*acmpca.Options)) (*acmpca.DeletePermissionOutput, error)
	ListPermissions(context.Context, *acmpca.ListPermissionsInput, ...func(*acmpca.Options)) (*acmpca.ListPermissionsOutput, error)
}

// NewCAPermissionClient returns a new client using AWS credentials as JSON encoded data.
func NewCAPermissionClient(conf *aws.Config) CAPermissionClient {
	return acmpca.NewFromConfig(*conf)
}
