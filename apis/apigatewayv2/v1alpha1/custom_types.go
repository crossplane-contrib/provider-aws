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

import xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"

// CustomAPIParameters includes the custom fields.
type CustomAPIParameters struct{}

// CustomAPIMappingParameters includes the custom fields.
type CustomAPIMappingParameters struct {
	// APIID is the ID for the API.
	// +immutable
	APIID *string `json:"apiId,omitempty"`

	// APIIDRef is a reference to an API used to set
	// the APIID.
	// +optional
	APIIDRef *xpv1.Reference `json:"apiIdRef,omitempty"`

	// APIIDSelector selects references to API used
	// to set the APIID.
	// +optional
	APIIDSelector *xpv1.Selector `json:"apiIdSelector,omitempty"`

	// Stage is the name for the Stage.
	// +immutable
	Stage *string `json:"stage,omitempty"`

	// StageDRef is a reference to an Stage used to set
	// the Stage.
	// +optional
	StageRef *xpv1.Reference `json:"stageRef,omitempty"`

	// StageSelector selects references to Stage used
	// to set the Stage.
	// +optional
	StageSelector *xpv1.Selector `json:"stageSelector,omitempty"`

	// DomainName is the DomainName for the DomainName.
	// +immutable
	DomainName *string `json:"domainName,omitempty"`

	// DomainNameRef is a reference to a DomainName used to set
	// the DomainName.
	// +optional
	DomainNameRef *xpv1.Reference `json:"domainNameRef,omitempty"`

	// DomainNameSelector selects references to DomainName used
	// to set the DomainName.
	// +optional
	DomainNameSelector *xpv1.Selector `json:"domainNameSelector,omitempty"`
}

// CustomAuthorizerParameters includes the custom fields.
type CustomAuthorizerParameters struct {
	// APIID is the ID for the API.
	// +immutable
	APIID *string `json:"apiId,omitempty"`

	// APIIDRef is a reference to an API used to set
	// the APIID.
	// +optional
	APIIDRef *xpv1.Reference `json:"apiIdRef,omitempty"`

	// APIIDSelector selects references to API used
	// to set the APIID.
	// +optional
	APIIDSelector *xpv1.Selector `json:"apiIdSelector,omitempty"`
}

// CustomDeploymentParameters includes the custom fields.
type CustomDeploymentParameters struct {
	// APIID is the ID for the API.
	// +immutable
	APIID *string `json:"apiId,omitempty"`

	// APIIDRef is a reference to an API used to set
	// the APIID.
	// +optional
	APIIDRef *xpv1.Reference `json:"apiIdRef,omitempty"`

	// APIIDSelector selects references to API used
	// to set the APIID.
	// +optional
	APIIDSelector *xpv1.Selector `json:"apiIdSelector,omitempty"`

	// StageNameRef is a reference to an Stage used to set
	// the StageName.
	// +optional
	StageNameRef *xpv1.Reference `json:"stageNameRef,omitempty"`

	// StageNameSelector selects references to Stage used
	// to set the StageName.
	// +optional
	StageNameSelector *xpv1.Selector `json:"stageNameSelector,omitempty"`
}

// CustomDomainNameParameters includes the custom fields.
type CustomDomainNameParameters struct{}

// CustomIntegrationParameters includes the custom fields.
type CustomIntegrationParameters struct {
	// APIID is the ID for the API.
	// +immutable
	APIID *string `json:"apiId,omitempty"`

	// APIIDRef is a reference to an API used to set
	// the APIID.
	// +optional
	APIIDRef *xpv1.Reference `json:"apiIdRef,omitempty"`

	// APIIDSelector selects references to API used
	// to set the APIID.
	// +optional
	APIIDSelector *xpv1.Selector `json:"apiIdSelector,omitempty"`
}

// CustomIntegrationResponseParameters includes the custom fields.
type CustomIntegrationResponseParameters struct {
	// APIID is the ID for the API.
	// +immutable
	APIID *string `json:"apiId,omitempty"`

	// APIIDRef is a reference to an API used to set
	// the APIID.
	// +optional
	APIIDRef *xpv1.Reference `json:"apiIdRef,omitempty"`

	// APIIDSelector selects references to API used
	// to set the APIID.
	// +optional
	APIIDSelector *xpv1.Selector `json:"apiIdSelector,omitempty"`

	// IntegrationID is the ID for the Integration.
	// +immutable
	IntegrationID *string `json:"integrationId,omitempty"`

	// IntegrationIDRef is a reference to an Integration used to set
	// the IntegrationID.
	// +optional
	IntegrationIDRef *xpv1.Reference `json:"integrationIdRef,omitempty"`

	// IntegrationIDSelector selects references to Integration used
	// to set the IntegrationID.
	// +optional
	IntegrationIDSelector *xpv1.Selector `json:"integrationIdSelector,omitempty"`
}

// CustomModelParameters includes the custom fields.
type CustomModelParameters struct {
	// APIID is the ID for the API.
	// +immutable
	APIID *string `json:"apiId,omitempty"`

	// APIIDRef is a reference to an API used to set
	// the APIID.
	// +optional
	APIIDRef *xpv1.Reference `json:"apiIdRef,omitempty"`

	// APIIDSelector selects references to API used
	// to set the APIID.
	// +optional
	APIIDSelector *xpv1.Selector `json:"apiIdSelector,omitempty"`
}

// CustomRouteParameters includes the custom fields.
type CustomRouteParameters struct {
	// APIID is the ID for the API.
	// +immutable
	APIID *string `json:"apiId,omitempty"`

	// APIIDRef is a reference to an API used to set
	// the APIID.
	// +optional
	APIIDRef *xpv1.Reference `json:"apiIdRef,omitempty"`

	// APIIDSelector selects references to API used
	// to set the APIID.
	// +optional
	APIIDSelector *xpv1.Selector `json:"apiIdSelector,omitempty"`

	// AuthorizerIDRef is a reference to an Authorizer used to set
	// the AuthorizerID.
	// +optional
	AuthorizerIDRef *xpv1.Reference `json:"authorizerIDRef,omitempty"`

	// AuthorizerIDSelector selects references to Authorizer used
	// to set the AuthorizerID.
	// +optional
	AuthorizerIDSelector *xpv1.Selector `json:"authorizerIDSelector,omitempty"`
}

// CustomRouteResponseParameters includes the custom fields.
type CustomRouteResponseParameters struct {
	// APIID is the ID for the API.
	// +immutable
	APIID *string `json:"apiId,omitempty"`

	// APIIDRef is a reference to an API used to set
	// the APIID.
	// +optional
	APIIDRef *xpv1.Reference `json:"apiIdRef,omitempty"`

	// APIIDSelector selects references to API used
	// to set the APIID.
	// +optional
	APIIDSelector *xpv1.Selector `json:"apiIdSelector,omitempty"`

	// RouteID is the ID for the Route.
	// +immutable
	RouteID *string `json:"routeId,omitempty"`

	// RouteIDRef is a reference to an Route used to set
	// the RouteID.
	// +optional
	RouteIDRef *xpv1.Reference `json:"routeIdRef,omitempty"`

	// RouteIDSelector selects references to Route used
	// to set the RouteID.
	// +optional
	RouteIDSelector *xpv1.Selector `json:"routeIdSelector,omitempty"`
}

// CustomVPCLinkParameters includes the custom fields.
type CustomVPCLinkParameters struct {
	// SecurityGroupIDs is the list of IDs for the SecurityGroups.
	SecurityGroupIDs []string `json:"securityGroupIds,omitempty"`

	// SecurityGroupIDRefs is a list of references to SecurityGroups used to set
	// the SecurityGroupIDs.
	// +optional
	SecurityGroupIDRefs []xpv1.Reference `json:"securityGroupIdRefs,omitempty"`

	// SecurityGroupIDsSelector selects references to SecurityGroupID used
	// to set the SecurityGroupIDs.
	// +optional
	SecurityGroupIDSelector *xpv1.Selector `json:"securityGroupIdSelector,omitempty"`

	// SecurityGroupIDs is the list of IDs for the SecurityGroups.
	SubnetIDs []string `json:"subnetIds,omitempty"`

	// SubnetIDRefs is a list of references to Subnets used to set
	// the SubnetIDs.
	// +optional
	SubnetIDRefs []xpv1.Reference `json:"subnetIdRefs,omitempty"`

	// SubnetIDsSelector selects references to Subnets used
	// to set the SubnetIDs.
	// +optional
	SubnetIDSelector *xpv1.Selector `json:"subnetIdSelector,omitempty"`
}

// CustomStageParameters includes the custom fields.
type CustomStageParameters struct {
	// APIID is the ID for the API.
	// +immutable
	APIID *string `json:"apiId,omitempty"`

	// APIIDRef is a reference to an API used to set
	// the APIID.
	// +optional
	APIIDRef *xpv1.Reference `json:"apiIdRef,omitempty"`

	// APIIDSelector selects references to API used
	// to set the APIID.
	// +optional
	APIIDSelector *xpv1.Selector `json:"apiIdSelector,omitempty"`
}
