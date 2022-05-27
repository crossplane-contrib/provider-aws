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

package routeresponse

import (
	"github.com/aws/aws-sdk-go/aws/awserr"
	svcsdk "github.com/aws/aws-sdk-go/service/apigatewayv2"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/apigatewayv2/v1alpha1"
)

// NOTE(muvaf): We return pointers in case the function needs to start with an
// empty object, hence need to return a new pointer.

// GenerateGetRouteResponseInput returns input for read
// operation.
func GenerateGetRouteResponseInput(cr *svcapitypes.RouteResponse) *svcsdk.GetRouteResponseInput {
	res := &svcsdk.GetRouteResponseInput{}

	if cr.Status.AtProvider.RouteResponseID != nil {
		res.SetRouteResponseId(*cr.Status.AtProvider.RouteResponseID)
	}

	return res
}

// GenerateRouteResponse returns the current state in the form of *svcapitypes.RouteResponse.
func GenerateRouteResponse(resp *svcsdk.GetRouteResponseOutput) *svcapitypes.RouteResponse {
	cr := &svcapitypes.RouteResponse{}

	if resp.ModelSelectionExpression != nil {
		cr.Spec.ForProvider.ModelSelectionExpression = resp.ModelSelectionExpression
	} else {
		cr.Spec.ForProvider.ModelSelectionExpression = nil
	}
	if resp.ResponseModels != nil {
		f1 := map[string]*string{}
		for f1key, f1valiter := range resp.ResponseModels {
			var f1val string
			f1val = *f1valiter
			f1[f1key] = &f1val
		}
		cr.Spec.ForProvider.ResponseModels = f1
	} else {
		cr.Spec.ForProvider.ResponseModels = nil
	}
	if resp.ResponseParameters != nil {
		f2 := map[string]*svcapitypes.ParameterConstraints{}
		for f2key, f2valiter := range resp.ResponseParameters {
			f2val := &svcapitypes.ParameterConstraints{}
			if f2valiter.Required != nil {
				f2val.Required = f2valiter.Required
			}
			f2[f2key] = f2val
		}
		cr.Spec.ForProvider.ResponseParameters = f2
	} else {
		cr.Spec.ForProvider.ResponseParameters = nil
	}
	if resp.RouteResponseId != nil {
		cr.Status.AtProvider.RouteResponseID = resp.RouteResponseId
	} else {
		cr.Status.AtProvider.RouteResponseID = nil
	}
	if resp.RouteResponseKey != nil {
		cr.Spec.ForProvider.RouteResponseKey = resp.RouteResponseKey
	} else {
		cr.Spec.ForProvider.RouteResponseKey = nil
	}

	return cr
}

// GenerateCreateRouteResponseInput returns a create input.
func GenerateCreateRouteResponseInput(cr *svcapitypes.RouteResponse) *svcsdk.CreateRouteResponseInput {
	res := &svcsdk.CreateRouteResponseInput{}

	if cr.Spec.ForProvider.ModelSelectionExpression != nil {
		res.SetModelSelectionExpression(*cr.Spec.ForProvider.ModelSelectionExpression)
	}
	if cr.Spec.ForProvider.ResponseModels != nil {
		f1 := map[string]*string{}
		for f1key, f1valiter := range cr.Spec.ForProvider.ResponseModels {
			var f1val string
			f1val = *f1valiter
			f1[f1key] = &f1val
		}
		res.SetResponseModels(f1)
	}
	if cr.Spec.ForProvider.ResponseParameters != nil {
		f2 := map[string]*svcsdk.ParameterConstraints{}
		for f2key, f2valiter := range cr.Spec.ForProvider.ResponseParameters {
			f2val := &svcsdk.ParameterConstraints{}
			if f2valiter.Required != nil {
				f2val.SetRequired(*f2valiter.Required)
			}
			f2[f2key] = f2val
		}
		res.SetResponseParameters(f2)
	}
	if cr.Spec.ForProvider.RouteResponseKey != nil {
		res.SetRouteResponseKey(*cr.Spec.ForProvider.RouteResponseKey)
	}

	return res
}

// GenerateUpdateRouteResponseInput returns an update input.
func GenerateUpdateRouteResponseInput(cr *svcapitypes.RouteResponse) *svcsdk.UpdateRouteResponseInput {
	res := &svcsdk.UpdateRouteResponseInput{}

	if cr.Spec.ForProvider.ModelSelectionExpression != nil {
		res.SetModelSelectionExpression(*cr.Spec.ForProvider.ModelSelectionExpression)
	}
	if cr.Spec.ForProvider.ResponseModels != nil {
		f2 := map[string]*string{}
		for f2key, f2valiter := range cr.Spec.ForProvider.ResponseModels {
			var f2val string
			f2val = *f2valiter
			f2[f2key] = &f2val
		}
		res.SetResponseModels(f2)
	}
	if cr.Spec.ForProvider.ResponseParameters != nil {
		f3 := map[string]*svcsdk.ParameterConstraints{}
		for f3key, f3valiter := range cr.Spec.ForProvider.ResponseParameters {
			f3val := &svcsdk.ParameterConstraints{}
			if f3valiter.Required != nil {
				f3val.SetRequired(*f3valiter.Required)
			}
			f3[f3key] = f3val
		}
		res.SetResponseParameters(f3)
	}
	if cr.Status.AtProvider.RouteResponseID != nil {
		res.SetRouteResponseId(*cr.Status.AtProvider.RouteResponseID)
	}
	if cr.Spec.ForProvider.RouteResponseKey != nil {
		res.SetRouteResponseKey(*cr.Spec.ForProvider.RouteResponseKey)
	}

	return res
}

// GenerateDeleteRouteResponseInput returns a deletion input.
func GenerateDeleteRouteResponseInput(cr *svcapitypes.RouteResponse) *svcsdk.DeleteRouteResponseInput {
	res := &svcsdk.DeleteRouteResponseInput{}

	if cr.Status.AtProvider.RouteResponseID != nil {
		res.SetRouteResponseId(*cr.Status.AtProvider.RouteResponseID)
	}

	return res
}

// IsNotFound returns whether the given error is of type NotFound or not.
func IsNotFound(err error) bool {
	awsErr, ok := err.(awserr.Error)
	return ok && awsErr.Code() == "NotFoundException"
}
