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

package v1alpha1

import runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"

// CustomAPIParameters includes the custom fields.
type CustomAPIParameters struct{}

// CustomAPIMappingParameters includes the custom fields.
type CustomAPIMappingParameters struct{}

// CustomAuthorizerParameters includes the custom fields.
type CustomAuthorizerParameters struct {
	// APIID is the ID for the API.
	// +immutable
	APIID *string `json:"apiId,omitempty"`

	// APIIDRef is a reference to an API used to set
	// the APIID.
	// +optional
	APIIDRef *runtimev1alpha1.Reference `json:"apiIdRef,omitempty"`

	// APIIDSelector selects references to API used
	// to set the APIID.
	// +optional
	APIIDSelector *runtimev1alpha1.Selector `json:"apiIdSelector,omitempty"`
}

// CustomDeploymentParameters includes the custom fields.
type CustomDeploymentParameters struct{}

// CustomDomainNameParameters includes the custom fields.
type CustomDomainNameParameters struct{}

// CustomIntegrationParameters includes the custom fields.
type CustomIntegrationParameters struct{}

// CustomIntegrationResponseParameters includes the custom fields.
type CustomIntegrationResponseParameters struct{}

// CustomModelParameters includes the custom fields.
type CustomModelParameters struct{}

// CustomRouteParameters includes the custom fields.
type CustomRouteParameters struct {
	// APIID is the ID for the API.
	// +immutable
	APIID *string `json:"apiId,omitempty"`

	// APIIDRef is a reference to an API used to set
	// the APIID.
	// +optional
	APIIDRef *runtimev1alpha1.Reference `json:"apiIdRef,omitempty"`

	// APIIDSelector selects references to API used
	// to set the APIID.
	// +optional
	APIIDSelector *runtimev1alpha1.Selector `json:"apiIdSelector,omitempty"`

	// AuthorizerIDRef is a reference to an Authorizer used to set
	// the AuthorizerID.
	// +optional
	AuthorizerIDRef *runtimev1alpha1.Reference `json:"authorizerIDRef,omitempty"`

	// AuthorizerIDSelector selects references to Authorizer used
	// to set the AuthorizerID.
	// +optional
	AuthorizerIDSelector *runtimev1alpha1.Selector `json:"authorizerIDSelector,omitempty"`
}

// CustomRouteResponseParameters includes the custom fields.
type CustomRouteResponseParameters struct{}

// CustomVPCLinkParameters includes the custom fields.
type CustomVPCLinkParameters struct{}

// CustomStageParameters includes the custom fields.
type CustomStageParameters struct {
	// APIID is the ID for the API.
	// +immutable
	APIID *string `json:"apiId,omitempty"`

	// APIIDRef is a reference to an API used to set
	// the APIID.
	// +optional
	APIIDRef *runtimev1alpha1.Reference `json:"apiIdRef,omitempty"`

	// APIIDSelector selects references to API used
	// to set the APIID.
	// +optional
	APIIDSelector *runtimev1alpha1.Selector `json:"apiIdSelector,omitempty"`
}
