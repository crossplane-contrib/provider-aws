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

package v1alpha1

import (
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/reference"
	resource "github.com/crossplane/crossplane-runtime/pkg/resource"
)

// CustomGroupParameters includes custom additional fields for GroupParameters.
type CustomGroupParameters struct {
	// The role ARN for the group.
	// +immutable
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-aws/apis/iam/v1beta1.Role
	// +crossplane:generate:reference:extractor=github.com/crossplane-contrib/provider-aws/apis/iam/v1beta1.RoleARN()
	RoleARN *string `json:"roleArn,omitempty"`

	// UserPoolIDRef is a reference to an server instance.
	// +optional
	RoleARNRef *xpv1.Reference `json:"roleArnRef,omitempty"`

	// UserPoolIDSelector selects references to an server instance.
	// +optional
	RoleARNSelector *xpv1.Selector `json:"roleArnSelector,omitempty"`

	// The user pool ID.
	// +immutable
	// +crossplane:generate:reference:type=UserPool
	UserPoolID *string `json:"userPoolId,omitempty"`

	// UserPoolIDRef is a reference to an server instance.
	// +optional
	UserPoolIDRef *xpv1.Reference `json:"userPoolIdRef,omitempty"`

	// UserPoolIDSelector selects references to an server instance.
	// +optional
	UserPoolIDSelector *xpv1.Selector `json:"userPoolIdSelector,omitempty"`
}

// CustomIdentityProviderParameters includes custom additional fields for IdentityProviderParameters.
type CustomIdentityProviderParameters struct {
	// The user pool ID.
	// +immutable
	// +crossplane:generate:reference:type=UserPool
	UserPoolID *string `json:"userPoolId,omitempty"`

	// UserPoolIDRef is a reference to an server instance.
	// +optional
	UserPoolIDRef *xpv1.Reference `json:"userPoolIdRef,omitempty"`

	// UserPoolIDSelector selects references to an server instance.
	// +optional
	UserPoolIDSelector *xpv1.Selector `json:"userPoolIdSelector,omitempty"`

	// ProviderDetailsSecretRef contins a reference to a secret containing keys according to ProviderDetails.
	// The following list describes the provider
	// detail keys for each identity provider type.
	//
	//    * For Google and Login with Amazon: client_id client_secret authorize_scopes
	//
	//    * For Facebook: client_id client_secret authorize_scopes api_version
	//
	//    * For Sign in with Apple: client_id team_id key_id private_key authorize_scopes
	//
	//    * For OIDC providers: client_id client_secret attributes_request_method
	//    oidc_issuer authorize_scopes authorize_url if not available from discovery
	//    URL specified by oidc_issuer key token_url if not available from discovery
	//    URL specified by oidc_issuer key attributes_url if not available from
	//    discovery URL specified by oidc_issuer key jwks_uri if not available from
	//    discovery URL specified by oidc_issuer key
	//
	//    * For SAML providers: MetadataFile OR MetadataURL IDPSignout optional
	// +kubebuilder:validation:Required
	ProviderDetailsSecretRef xpv1.SecretReference `json:"providerDetailsSecretRef,omitempty"`
}

// CustomUserPoolParameters includes custom additional fields for UserPoolParameters.
type CustomUserPoolParameters struct{}

// CustomUserPoolDomainParameters includes custom additional fields for UserPoolDomainParameters.
type CustomUserPoolDomainParameters struct {
	// The user pool ID.
	// +immutable
	// +crossplane:generate:reference:type=UserPool
	UserPoolID *string `json:"userPoolId,omitempty"`

	// UserPoolIDRef is a reference to an server instance.
	// +optional
	UserPoolIDRef *xpv1.Reference `json:"userPoolIdRef,omitempty"`

	// UserPoolIDSelector selects references to an server instance.
	// +optional
	UserPoolIDSelector *xpv1.Selector `json:"userPoolIdSelector,omitempty"`
}

// CustomUserPoolClientParameters includes custom additional fields for UserPoolClientParameters.
type CustomUserPoolClientParameters struct {
	// The user pool ID.
	// +immutable
	// +crossplane:generate:reference:type=UserPool
	UserPoolID *string `json:"userPoolId,omitempty"`

	// UserPoolIDRef is a reference to an server instance.
	// +optional
	UserPoolIDRef *xpv1.Reference `json:"userPoolIdRef,omitempty"`

	// UserPoolIDSelector selects references to an server instance.
	// +optional
	UserPoolIDSelector *xpv1.Selector `json:"userPoolIdSelector,omitempty"`
}

// CustomResourceServerParameters includes the custom fields of ResourceServerParameters.
type CustomResourceServerParameters struct {
	// The user pool ID.
	// +immutable
	// +crossplane:generate:reference:type=UserPool
	UserPoolID *string `json:"userPoolId,omitempty"`

	// UserPoolIDRef is a reference to an server instance.
	// +optional
	UserPoolIDRef *xpv1.Reference `json:"userPoolIdRef,omitempty"`

	// UserPoolIDSelector selects references to an server instance.
	// +optional
	UserPoolIDSelector *xpv1.Selector `json:"userPoolIdSelector,omitempty"`
}

// UserPoolName returns the status.atProvider.name of a UserPool.
func UserPoolName() reference.ExtractValueFn {
	return func(mg resource.Managed) string {
		r, ok := mg.(*UserPool)
		if !ok {
			return ""
		}

		if r.Status.AtProvider.ID == nil {
			return ""
		}

		return "cognito-idp." + r.Spec.ForProvider.Region + ".amazonaws.com/" + *r.Status.AtProvider.ID

	}
}
