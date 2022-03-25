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

package method

import (
	"github.com/aws/aws-sdk-go/aws/awserr"
	svcsdk "github.com/aws/aws-sdk-go/service/apigateway"

	svcapitypes "github.com/crossplane/provider-aws/apis/apigateway/v1alpha1"
)

// NOTE(muvaf): We return pointers in case the function needs to start with an
// empty object, hence need to return a new pointer.

// GenerateGetMethodInput returns input for read
// operation.
func GenerateGetMethodInput(cr *svcapitypes.Method) *svcsdk.GetMethodInput {
	res := &svcsdk.GetMethodInput{}

	if cr.Spec.ForProvider.HTTPMethod != nil {
		res.SetHttpMethod(*cr.Spec.ForProvider.HTTPMethod)
	}

	return res
}

// GenerateMethod returns the current state in the form of *svcapitypes.Method.
func GenerateMethod(resp *svcsdk.Method) *svcapitypes.Method {
	cr := &svcapitypes.Method{}

	if resp.ApiKeyRequired != nil {
		cr.Spec.ForProvider.APIKeyRequired = resp.ApiKeyRequired
	} else {
		cr.Spec.ForProvider.APIKeyRequired = nil
	}
	if resp.AuthorizationScopes != nil {
		f1 := []*string{}
		for _, f1iter := range resp.AuthorizationScopes {
			var f1elem string
			f1elem = *f1iter
			f1 = append(f1, &f1elem)
		}
		cr.Spec.ForProvider.AuthorizationScopes = f1
	} else {
		cr.Spec.ForProvider.AuthorizationScopes = nil
	}
	if resp.AuthorizationType != nil {
		cr.Spec.ForProvider.AuthorizationType = resp.AuthorizationType
	} else {
		cr.Spec.ForProvider.AuthorizationType = nil
	}
	if resp.AuthorizerId != nil {
		cr.Spec.ForProvider.AuthorizerID = resp.AuthorizerId
	} else {
		cr.Spec.ForProvider.AuthorizerID = nil
	}
	if resp.HttpMethod != nil {
		cr.Spec.ForProvider.HTTPMethod = resp.HttpMethod
	} else {
		cr.Spec.ForProvider.HTTPMethod = nil
	}
	if resp.OperationName != nil {
		cr.Spec.ForProvider.OperationName = resp.OperationName
	} else {
		cr.Spec.ForProvider.OperationName = nil
	}
	if resp.RequestModels != nil {
		f6 := map[string]*string{}
		for f6key, f6valiter := range resp.RequestModels {
			var f6val string
			f6val = *f6valiter
			f6[f6key] = &f6val
		}
		cr.Spec.ForProvider.RequestModels = f6
	} else {
		cr.Spec.ForProvider.RequestModels = nil
	}
	if resp.RequestParameters != nil {
		f7 := map[string]*bool{}
		for f7key, f7valiter := range resp.RequestParameters {
			var f7val bool
			f7val = *f7valiter
			f7[f7key] = &f7val
		}
		cr.Spec.ForProvider.RequestParameters = f7
	} else {
		cr.Spec.ForProvider.RequestParameters = nil
	}
	if resp.RequestValidatorId != nil {
		cr.Spec.ForProvider.RequestValidatorID = resp.RequestValidatorId
	} else {
		cr.Spec.ForProvider.RequestValidatorID = nil
	}

	return cr
}

// GeneratePutMethodInput returns a create input.
func GeneratePutMethodInput(cr *svcapitypes.Method) *svcsdk.PutMethodInput {
	res := &svcsdk.PutMethodInput{}

	if cr.Spec.ForProvider.APIKeyRequired != nil {
		res.SetApiKeyRequired(*cr.Spec.ForProvider.APIKeyRequired)
	}
	if cr.Spec.ForProvider.AuthorizationScopes != nil {
		f1 := []*string{}
		for _, f1iter := range cr.Spec.ForProvider.AuthorizationScopes {
			var f1elem string
			f1elem = *f1iter
			f1 = append(f1, &f1elem)
		}
		res.SetAuthorizationScopes(f1)
	}
	if cr.Spec.ForProvider.AuthorizationType != nil {
		res.SetAuthorizationType(*cr.Spec.ForProvider.AuthorizationType)
	}
	if cr.Spec.ForProvider.AuthorizerID != nil {
		res.SetAuthorizerId(*cr.Spec.ForProvider.AuthorizerID)
	}
	if cr.Spec.ForProvider.HTTPMethod != nil {
		res.SetHttpMethod(*cr.Spec.ForProvider.HTTPMethod)
	}
	if cr.Spec.ForProvider.OperationName != nil {
		res.SetOperationName(*cr.Spec.ForProvider.OperationName)
	}
	if cr.Spec.ForProvider.RequestModels != nil {
		f6 := map[string]*string{}
		for f6key, f6valiter := range cr.Spec.ForProvider.RequestModels {
			var f6val string
			f6val = *f6valiter
			f6[f6key] = &f6val
		}
		res.SetRequestModels(f6)
	}
	if cr.Spec.ForProvider.RequestParameters != nil {
		f7 := map[string]*bool{}
		for f7key, f7valiter := range cr.Spec.ForProvider.RequestParameters {
			var f7val bool
			f7val = *f7valiter
			f7[f7key] = &f7val
		}
		res.SetRequestParameters(f7)
	}
	if cr.Spec.ForProvider.RequestValidatorID != nil {
		res.SetRequestValidatorId(*cr.Spec.ForProvider.RequestValidatorID)
	}

	return res
}

// GenerateUpdateMethodInput returns an update input.
func GenerateUpdateMethodInput(cr *svcapitypes.Method) *svcsdk.UpdateMethodInput {
	res := &svcsdk.UpdateMethodInput{}

	if cr.Spec.ForProvider.HTTPMethod != nil {
		res.SetHttpMethod(*cr.Spec.ForProvider.HTTPMethod)
	}

	return res
}

// GenerateDeleteMethodInput returns a deletion input.
func GenerateDeleteMethodInput(cr *svcapitypes.Method) *svcsdk.DeleteMethodInput {
	res := &svcsdk.DeleteMethodInput{}

	if cr.Spec.ForProvider.HTTPMethod != nil {
		res.SetHttpMethod(*cr.Spec.ForProvider.HTTPMethod)
	}

	return res
}

// IsNotFound returns whether the given error is of type NotFound or not.
func IsNotFound(err error) bool {
	awsErr, ok := err.(awserr.Error)
	return ok && awsErr.Code() == "NotFoundException"
}
