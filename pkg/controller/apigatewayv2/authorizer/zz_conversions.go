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

package authorizer

import (
	"github.com/aws/aws-sdk-go/aws/awserr"
	svcsdk "github.com/aws/aws-sdk-go/service/apigatewayv2"

	svcapitypes "github.com/crossplane/provider-aws/apis/apigatewayv2/v1alpha1"
)

// NOTE(muvaf): We return pointers in case the function needs to start with an
// empty object, hence need to return a new pointer.

// GenerateGetAuthorizerInput returns input for read
// operation.
func GenerateGetAuthorizerInput(cr *svcapitypes.Authorizer) *svcsdk.GetAuthorizerInput {
	res := &svcsdk.GetAuthorizerInput{}

	if cr.Status.AtProvider.AuthorizerID != nil {
		res.SetAuthorizerId(*cr.Status.AtProvider.AuthorizerID)
	}

	return res
}

// GenerateAuthorizer returns the current state in the form of *svcapitypes.Authorizer.
func GenerateAuthorizer(resp *svcsdk.GetAuthorizerOutput) *svcapitypes.Authorizer {
	cr := &svcapitypes.Authorizer{}

	if resp.AuthorizerCredentialsArn != nil {
		cr.Spec.ForProvider.AuthorizerCredentialsARN = resp.AuthorizerCredentialsArn
	} else {
		cr.Spec.ForProvider.AuthorizerCredentialsARN = nil
	}
	if resp.AuthorizerId != nil {
		cr.Status.AtProvider.AuthorizerID = resp.AuthorizerId
	} else {
		cr.Status.AtProvider.AuthorizerID = nil
	}
	if resp.AuthorizerPayloadFormatVersion != nil {
		cr.Spec.ForProvider.AuthorizerPayloadFormatVersion = resp.AuthorizerPayloadFormatVersion
	} else {
		cr.Spec.ForProvider.AuthorizerPayloadFormatVersion = nil
	}
	if resp.AuthorizerResultTtlInSeconds != nil {
		cr.Spec.ForProvider.AuthorizerResultTtlInSeconds = resp.AuthorizerResultTtlInSeconds
	} else {
		cr.Spec.ForProvider.AuthorizerResultTtlInSeconds = nil
	}
	if resp.AuthorizerType != nil {
		cr.Spec.ForProvider.AuthorizerType = resp.AuthorizerType
	} else {
		cr.Spec.ForProvider.AuthorizerType = nil
	}
	if resp.AuthorizerUri != nil {
		cr.Spec.ForProvider.AuthorizerURI = resp.AuthorizerUri
	} else {
		cr.Spec.ForProvider.AuthorizerURI = nil
	}
	if resp.EnableSimpleResponses != nil {
		cr.Spec.ForProvider.EnableSimpleResponses = resp.EnableSimpleResponses
	} else {
		cr.Spec.ForProvider.EnableSimpleResponses = nil
	}
	if resp.IdentitySource != nil {
		f7 := []*string{}
		for _, f7iter := range resp.IdentitySource {
			var f7elem string
			f7elem = *f7iter
			f7 = append(f7, &f7elem)
		}
		cr.Spec.ForProvider.IdentitySource = f7
	} else {
		cr.Spec.ForProvider.IdentitySource = nil
	}
	if resp.IdentityValidationExpression != nil {
		cr.Spec.ForProvider.IdentityValidationExpression = resp.IdentityValidationExpression
	} else {
		cr.Spec.ForProvider.IdentityValidationExpression = nil
	}
	if resp.JwtConfiguration != nil {
		f9 := &svcapitypes.JWTConfiguration{}
		if resp.JwtConfiguration.Audience != nil {
			f9f0 := []*string{}
			for _, f9f0iter := range resp.JwtConfiguration.Audience {
				var f9f0elem string
				f9f0elem = *f9f0iter
				f9f0 = append(f9f0, &f9f0elem)
			}
			f9.Audience = f9f0
		}
		if resp.JwtConfiguration.Issuer != nil {
			f9.Issuer = resp.JwtConfiguration.Issuer
		}
		cr.Spec.ForProvider.JWTConfiguration = f9
	} else {
		cr.Spec.ForProvider.JWTConfiguration = nil
	}
	if resp.Name != nil {
		cr.Spec.ForProvider.Name = resp.Name
	} else {
		cr.Spec.ForProvider.Name = nil
	}

	return cr
}

// GenerateCreateAuthorizerInput returns a create input.
func GenerateCreateAuthorizerInput(cr *svcapitypes.Authorizer) *svcsdk.CreateAuthorizerInput {
	res := &svcsdk.CreateAuthorizerInput{}

	if cr.Spec.ForProvider.AuthorizerCredentialsARN != nil {
		res.SetAuthorizerCredentialsArn(*cr.Spec.ForProvider.AuthorizerCredentialsARN)
	}
	if cr.Spec.ForProvider.AuthorizerPayloadFormatVersion != nil {
		res.SetAuthorizerPayloadFormatVersion(*cr.Spec.ForProvider.AuthorizerPayloadFormatVersion)
	}
	if cr.Spec.ForProvider.AuthorizerResultTtlInSeconds != nil {
		res.SetAuthorizerResultTtlInSeconds(*cr.Spec.ForProvider.AuthorizerResultTtlInSeconds)
	}
	if cr.Spec.ForProvider.AuthorizerType != nil {
		res.SetAuthorizerType(*cr.Spec.ForProvider.AuthorizerType)
	}
	if cr.Spec.ForProvider.AuthorizerURI != nil {
		res.SetAuthorizerUri(*cr.Spec.ForProvider.AuthorizerURI)
	}
	if cr.Spec.ForProvider.EnableSimpleResponses != nil {
		res.SetEnableSimpleResponses(*cr.Spec.ForProvider.EnableSimpleResponses)
	}
	if cr.Spec.ForProvider.IdentitySource != nil {
		f6 := []*string{}
		for _, f6iter := range cr.Spec.ForProvider.IdentitySource {
			var f6elem string
			f6elem = *f6iter
			f6 = append(f6, &f6elem)
		}
		res.SetIdentitySource(f6)
	}
	if cr.Spec.ForProvider.IdentityValidationExpression != nil {
		res.SetIdentityValidationExpression(*cr.Spec.ForProvider.IdentityValidationExpression)
	}
	if cr.Spec.ForProvider.JWTConfiguration != nil {
		f8 := &svcsdk.JWTConfiguration{}
		if cr.Spec.ForProvider.JWTConfiguration.Audience != nil {
			f8f0 := []*string{}
			for _, f8f0iter := range cr.Spec.ForProvider.JWTConfiguration.Audience {
				var f8f0elem string
				f8f0elem = *f8f0iter
				f8f0 = append(f8f0, &f8f0elem)
			}
			f8.SetAudience(f8f0)
		}
		if cr.Spec.ForProvider.JWTConfiguration.Issuer != nil {
			f8.SetIssuer(*cr.Spec.ForProvider.JWTConfiguration.Issuer)
		}
		res.SetJwtConfiguration(f8)
	}
	if cr.Spec.ForProvider.Name != nil {
		res.SetName(*cr.Spec.ForProvider.Name)
	}

	return res
}

// GenerateUpdateAuthorizerInput returns an update input.
func GenerateUpdateAuthorizerInput(cr *svcapitypes.Authorizer) *svcsdk.UpdateAuthorizerInput {
	res := &svcsdk.UpdateAuthorizerInput{}

	if cr.Spec.ForProvider.AuthorizerCredentialsARN != nil {
		res.SetAuthorizerCredentialsArn(*cr.Spec.ForProvider.AuthorizerCredentialsARN)
	}
	if cr.Status.AtProvider.AuthorizerID != nil {
		res.SetAuthorizerId(*cr.Status.AtProvider.AuthorizerID)
	}
	if cr.Spec.ForProvider.AuthorizerPayloadFormatVersion != nil {
		res.SetAuthorizerPayloadFormatVersion(*cr.Spec.ForProvider.AuthorizerPayloadFormatVersion)
	}
	if cr.Spec.ForProvider.AuthorizerResultTtlInSeconds != nil {
		res.SetAuthorizerResultTtlInSeconds(*cr.Spec.ForProvider.AuthorizerResultTtlInSeconds)
	}
	if cr.Spec.ForProvider.AuthorizerType != nil {
		res.SetAuthorizerType(*cr.Spec.ForProvider.AuthorizerType)
	}
	if cr.Spec.ForProvider.AuthorizerURI != nil {
		res.SetAuthorizerUri(*cr.Spec.ForProvider.AuthorizerURI)
	}
	if cr.Spec.ForProvider.EnableSimpleResponses != nil {
		res.SetEnableSimpleResponses(*cr.Spec.ForProvider.EnableSimpleResponses)
	}
	if cr.Spec.ForProvider.IdentitySource != nil {
		f8 := []*string{}
		for _, f8iter := range cr.Spec.ForProvider.IdentitySource {
			var f8elem string
			f8elem = *f8iter
			f8 = append(f8, &f8elem)
		}
		res.SetIdentitySource(f8)
	}
	if cr.Spec.ForProvider.IdentityValidationExpression != nil {
		res.SetIdentityValidationExpression(*cr.Spec.ForProvider.IdentityValidationExpression)
	}
	if cr.Spec.ForProvider.JWTConfiguration != nil {
		f10 := &svcsdk.JWTConfiguration{}
		if cr.Spec.ForProvider.JWTConfiguration.Audience != nil {
			f10f0 := []*string{}
			for _, f10f0iter := range cr.Spec.ForProvider.JWTConfiguration.Audience {
				var f10f0elem string
				f10f0elem = *f10f0iter
				f10f0 = append(f10f0, &f10f0elem)
			}
			f10.SetAudience(f10f0)
		}
		if cr.Spec.ForProvider.JWTConfiguration.Issuer != nil {
			f10.SetIssuer(*cr.Spec.ForProvider.JWTConfiguration.Issuer)
		}
		res.SetJwtConfiguration(f10)
	}
	if cr.Spec.ForProvider.Name != nil {
		res.SetName(*cr.Spec.ForProvider.Name)
	}

	return res
}

// GenerateDeleteAuthorizerInput returns a deletion input.
func GenerateDeleteAuthorizerInput(cr *svcapitypes.Authorizer) *svcsdk.DeleteAuthorizerInput {
	res := &svcsdk.DeleteAuthorizerInput{}

	if cr.Status.AtProvider.AuthorizerID != nil {
		res.SetAuthorizerId(*cr.Status.AtProvider.AuthorizerID)
	}

	return res
}

// IsNotFound returns whether the given error is of type NotFound or not.
func IsNotFound(err error) bool {
	awsErr, ok := err.(awserr.Error)
	return ok && awsErr.Code() == "NotFoundException"
}
