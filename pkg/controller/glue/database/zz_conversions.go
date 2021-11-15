/*
Copyright 2021 The Crossplane Authors.

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

// Code generated by ack-generate. DO NOT EDIT.

package database

import (
	"github.com/aws/aws-sdk-go/aws/awserr"
	svcsdk "github.com/aws/aws-sdk-go/service/glue"

	svcapitypes "github.com/crossplane/provider-aws/apis/glue/v1alpha1"
)

// NOTE(muvaf): We return pointers in case the function needs to start with an
// empty object, hence need to return a new pointer.

// GenerateGetDatabaseInput returns input for read
// operation.
func GenerateGetDatabaseInput(cr *svcapitypes.Database) *svcsdk.GetDatabaseInput {
	res := &svcsdk.GetDatabaseInput{}

	if cr.Spec.ForProvider.CatalogID != nil {
		res.SetCatalogId(*cr.Spec.ForProvider.CatalogID)
	}

	return res
}

// GenerateDatabase returns the current state in the form of *svcapitypes.Database.
func GenerateDatabase(resp *svcsdk.GetDatabaseOutput) *svcapitypes.Database {
	cr := &svcapitypes.Database{}

	return cr
}

// GenerateCreateDatabaseInput returns a create input.
func GenerateCreateDatabaseInput(cr *svcapitypes.Database) *svcsdk.CreateDatabaseInput {
	res := &svcsdk.CreateDatabaseInput{}

	if cr.Spec.ForProvider.CatalogID != nil {
		res.SetCatalogId(*cr.Spec.ForProvider.CatalogID)
	}

	return res
}

// GenerateUpdateDatabaseInput returns an update input.
func GenerateUpdateDatabaseInput(cr *svcapitypes.Database) *svcsdk.UpdateDatabaseInput {
	res := &svcsdk.UpdateDatabaseInput{}

	if cr.Spec.ForProvider.CatalogID != nil {
		res.SetCatalogId(*cr.Spec.ForProvider.CatalogID)
	}

	return res
}

// GenerateDeleteDatabaseInput returns a deletion input.
func GenerateDeleteDatabaseInput(cr *svcapitypes.Database) *svcsdk.DeleteDatabaseInput {
	res := &svcsdk.DeleteDatabaseInput{}

	if cr.Spec.ForProvider.CatalogID != nil {
		res.SetCatalogId(*cr.Spec.ForProvider.CatalogID)
	}

	return res
}

// IsNotFound returns whether the given error is of type NotFound or not.
func IsNotFound(err error) bool {
	awsErr, ok := err.(awserr.Error)
	return ok && awsErr.Code() == "EntityNotFoundException"
}
