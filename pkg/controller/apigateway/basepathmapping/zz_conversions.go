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

package basepathmapping

import (
	"github.com/aws/aws-sdk-go/aws/awserr"
	svcsdk "github.com/aws/aws-sdk-go/service/apigateway"

	svcapitypes "github.com/crossplane/provider-aws/apis/apigateway/v1alpha1"
)

// NOTE(muvaf): We return pointers in case the function needs to start with an
// empty object, hence need to return a new pointer.

// GenerateGetBasePathMappingInput returns input for read
// operation.
func GenerateGetBasePathMappingInput(cr *svcapitypes.BasePathMapping) *svcsdk.GetBasePathMappingInput {
	res := &svcsdk.GetBasePathMappingInput{}

	if cr.Spec.ForProvider.BasePath != nil {
		res.SetBasePath(*cr.Spec.ForProvider.BasePath)
	}
	if cr.Spec.ForProvider.DomainName != nil {
		res.SetDomainName(*cr.Spec.ForProvider.DomainName)
	}

	return res
}

// GenerateBasePathMapping returns the current state in the form of *svcapitypes.BasePathMapping.
func GenerateBasePathMapping(resp *svcsdk.BasePathMapping) *svcapitypes.BasePathMapping {
	cr := &svcapitypes.BasePathMapping{}

	if resp.BasePath != nil {
		cr.Spec.ForProvider.BasePath = resp.BasePath
	} else {
		cr.Spec.ForProvider.BasePath = nil
	}
	if resp.RestApiId != nil {
		cr.Status.AtProvider.RestAPIID = resp.RestApiId
	} else {
		cr.Status.AtProvider.RestAPIID = nil
	}
	if resp.Stage != nil {
		cr.Spec.ForProvider.Stage = resp.Stage
	} else {
		cr.Spec.ForProvider.Stage = nil
	}

	return cr
}

// GenerateCreateBasePathMappingInput returns a create input.
func GenerateCreateBasePathMappingInput(cr *svcapitypes.BasePathMapping) *svcsdk.CreateBasePathMappingInput {
	res := &svcsdk.CreateBasePathMappingInput{}

	if cr.Spec.ForProvider.BasePath != nil {
		res.SetBasePath(*cr.Spec.ForProvider.BasePath)
	}
	if cr.Spec.ForProvider.DomainName != nil {
		res.SetDomainName(*cr.Spec.ForProvider.DomainName)
	}
	if cr.Spec.ForProvider.Stage != nil {
		res.SetStage(*cr.Spec.ForProvider.Stage)
	}

	return res
}

// GenerateUpdateBasePathMappingInput returns an update input.
func GenerateUpdateBasePathMappingInput(cr *svcapitypes.BasePathMapping) *svcsdk.UpdateBasePathMappingInput {
	res := &svcsdk.UpdateBasePathMappingInput{}

	if cr.Spec.ForProvider.BasePath != nil {
		res.SetBasePath(*cr.Spec.ForProvider.BasePath)
	}
	if cr.Spec.ForProvider.DomainName != nil {
		res.SetDomainName(*cr.Spec.ForProvider.DomainName)
	}

	return res
}

// GenerateDeleteBasePathMappingInput returns a deletion input.
func GenerateDeleteBasePathMappingInput(cr *svcapitypes.BasePathMapping) *svcsdk.DeleteBasePathMappingInput {
	res := &svcsdk.DeleteBasePathMappingInput{}

	if cr.Spec.ForProvider.BasePath != nil {
		res.SetBasePath(*cr.Spec.ForProvider.BasePath)
	}
	if cr.Spec.ForProvider.DomainName != nil {
		res.SetDomainName(*cr.Spec.ForProvider.DomainName)
	}

	return res
}

// IsNotFound returns whether the given error is of type NotFound or not.
func IsNotFound(err error) bool {
	awsErr, ok := err.(awserr.Error)
	return ok && awsErr.Code() == "NotFoundException"
}
