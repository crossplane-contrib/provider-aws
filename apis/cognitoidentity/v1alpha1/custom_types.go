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
)

// CustomIdentityPoolParameters includes the custom fields of IdentityPool.
type CustomIdentityPoolParameters struct {

	// The Amazon Resource Names (ARN) of the OpenID Connect providers.
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-aws/apis/iam/v1beta1.OpenIDConnectProvider
	// +crossplane:generate:reference:refFieldName=OpenIDConnectProviderARNRefs
	// +crossplane:generate:reference:selectorFieldName=OpenIDConnectProviderARNSelector
	// +optional
	OpenIDConnectProviderARNs []*string `json:"openIdConnectProviderARNs,omitempty"`

	// OpenIDConnectProviderARNRefs is a list of references to OpenIDConnectProviderARNs.
	// +optional
	OpenIDConnectProviderARNRefs []xpv1.Reference `json:"openIdConnectProviderARNRefs,omitempty"`

	// OpenIDConnectProviderARNSelector selects references to OpenIDConnectProviderARNs.
	// +optional
	OpenIDConnectProviderARNSelector *xpv1.Selector `json:"openIdConnectProviderARNSelector,omitempty"`

	// An array of Amazon Cognito user pools and their client IDs.
	CognitoIdentityProviders []*Provider `json:"cognitoIdentityProviders,omitempty"`

	// TRUE if the identity pool supports unauthenticated logins.
	// +kubebuilder:validation:Required
	AllowUnauthenticatedIdentities *bool `json:"allowUnauthenticatedIdentities"`
}

// Provider contains information to Cognito UserPools and UserPoolClients
// +kubebuilder:skipversion
type Provider struct {
	// The client ID for the Amazon Cognito user pool client.
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-aws/apis/cognitoidentityprovider/v1alpha1.UserPoolClient
	// + optional
	ClientID *string `json:"clientId,omitempty"`

	// ClientIDRef is a reference to an ClientID.
	// +optional
	ClientIDRef *xpv1.Reference `json:"clientIdRef,omitempty"`

	// ClientIDSelector selects references to ClientID.
	// +optional
	ClientIDSelector *xpv1.Selector `json:"clientIdSelector,omitempty"`

	// The provider name for an Amazon Cognito user pool.
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-aws/apis/cognitoidentityprovider/v1alpha1.UserPool
	// +crossplane:generate:reference:extractor=github.com/crossplane-contrib/provider-aws/apis/cognitoidentityprovider/v1alpha1.UserPoolName()
	// +optional
	ProviderName *string `json:"providerName,omitempty"`

	// ProviderNameRef is a reference to an ProviderName.
	// +optional
	ProviderNameRef *xpv1.Reference `json:"providerNameRef,omitempty"`

	// ProviderNameSelector selects references to ProviderName.
	// +optional
	ProviderNameSelector *xpv1.Selector `json:"providerNameSelector,omitempty"`

	// Whether the server-side token validation is enabled for the identity provider’s token.
	// +optional
	ServerSideTokenCheck *bool `json:"serverSideTokenCheck,omitempty"`
}
