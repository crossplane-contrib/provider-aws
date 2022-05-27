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

// NOTE(muvaf): This code ported from ACK-generated code. See details here:
// https://github.com/crossplane-contrib/provider-aws/pull/950#issue-1055573793

package alias

import (
	"github.com/aws/aws-sdk-go/aws/awserr"
	svcsdk "github.com/aws/aws-sdk-go/service/kms"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/kms/v1alpha1"
)

// NOTE(muvaf): We return pointers in case the function needs to start with an
// empty object, hence need to return a new pointer.

// GenerateListAliasesInput returns input for read
// operation.
func GenerateListAliasesInput(_ *svcapitypes.Alias) *svcsdk.ListAliasesInput {
	res := &svcsdk.ListAliasesInput{}

	return res
}

// GenerateAlias returns the current state in the form of *svcapitypes.Alias.
func GenerateAlias(resp *svcsdk.ListAliasesOutput) *svcapitypes.Alias {
	cr := &svcapitypes.Alias{}
	for _, elem := range resp.Aliases {
		if elem.TargetKeyId != nil {
			cr.Spec.ForProvider.TargetKeyID = elem.TargetKeyId
		}
	}
	return cr
}

// GenerateCreateAliasInput returns a create input.
func GenerateCreateAliasInput(cr *svcapitypes.Alias) *svcsdk.CreateAliasInput {
	res := &svcsdk.CreateAliasInput{}

	if cr.Spec.ForProvider.TargetKeyID != nil {
		res.SetTargetKeyId(*cr.Spec.ForProvider.TargetKeyID)
	}

	return res
}

// GenerateUpdateAliasInput returns an update input.
func GenerateUpdateAliasInput(cr *svcapitypes.Alias) *svcsdk.UpdateAliasInput {
	res := &svcsdk.UpdateAliasInput{}

	if cr.Spec.ForProvider.TargetKeyID != nil {
		res.SetTargetKeyId(*cr.Spec.ForProvider.TargetKeyID)
	}

	return res
}

// GenerateDeleteAliasInput returns a deletion input.
func GenerateDeleteAliasInput(_ *svcapitypes.Alias) *svcsdk.DeleteAliasInput {
	res := &svcsdk.DeleteAliasInput{}

	return res
}

// IsNotFound returns whether the given error is of type NotFound or not.
func IsNotFound(err error) bool {
	awsErr, ok := err.(awserr.Error)
	return ok && awsErr.Code() == "NotFoundException"
}
