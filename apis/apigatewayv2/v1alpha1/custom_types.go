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

// CustomAPIObservation includes the custom fields.
type CustomAPIObservation struct{}

// CustomAPIMappingParameters includes the custom fields.
type CustomAPIMappingParameters struct {
	// APIID is the ID for the API.
	// +immutable
	// +crossplane:generate:reference:type=API
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
	// +crossplane:generate:reference:type=Stage
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
	// +crossplane:generate:reference:type=DomainName
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

// CustomAPIMappingObservation includes the custom fields.
type CustomAPIMappingObservation struct{}

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

// CustomAuthorizerObservation includes the custom fields.
type CustomAuthorizerObservation struct{}

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

// CustomDeploymentObservation includes the custom fields.
type CustomDeploymentObservation struct{}

// CustomDomainNameParameters includes the custom fields.
type CustomDomainNameParameters struct{}

// CustomDomainNameObservation includes the custom fields.
type CustomDomainNameObservation struct{}

// ResponseParameters is a map of status codes and transform operations on each
// of them.
type ResponseParameters map[string]ResponseParameter

// ResponseParameter represents a single response parameter transform operation.
type ResponseParameter struct {
	// HeaderEntries is the array of header changes you'd like to make.
	// For details, see Transforming API responses in https://docs.aws.amazon.com/apigateway/latest/developerguide/http-api-parameter-mapping.html
	HeaderEntries []HeaderEntry `json:"headerEntry,omitempty"`

	// OverwriteStatusCode is the status code you'd like the response to have,
	// overwriting the one in the original response.
	// For details, see Transforming API responses in https://docs.aws.amazon.com/apigateway/latest/developerguide/http-api-parameter-mapping.html
	OverwriteStatusCode *string `json:"overwriteStatusCodeEntry,omitempty"`
}

// HeaderEntry can be used to represent a single header transform.
type HeaderEntry struct {
	// Operation is what you'd like to do with given header. Only append, overwrite
	// and remove values are supported.
	// +kubebuilder:validation:Enum=append;overwrite;remove
	Operation string `json:"operation"`

	// Name is the name of the header.
	Name string `json:"name"`

	// Value is the new value.
	Value string `json:"value"`
}

// CustomIntegrationParameters includes the custom fields.
type CustomIntegrationParameters struct {
	// NOTE(muvaf): Original type of ResponseParameters is map[string]map[string]*string,
	// but we cannot use that since kubebuilder does not support generating CRD
	// schema for map of maps.

	// Supported only for HTTP APIs. You use response parameters to transform the
	// HTTP response from a backend integration before returning the response to
	// clients. Specify a key-value map from a selection key to response parameters.
	// The selection key must be a valid HTTP status code within the range of 200-599.
	ResponseParameters ResponseParameters `json:"responseParameters,omitempty"`

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

// CustomIntegrationObservation includes the custom fields.
type CustomIntegrationObservation struct{}

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

// CustomIntegrationResponseObservation includes the custom fields.
type CustomIntegrationResponseObservation struct{}

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

// CustomModelObservation includes the custom fields.
type CustomModelObservation struct{}

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

	// Target for the route, of the form integrations/IntegrationID,
	// where IntegrationID is the identifier of an AWS API Gateway Integration
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-aws/apis/apigatewayv2/v1alpha1.Integration
	// +crossplane:generate:reference:extractor=github.com/crossplane-contrib/provider-aws/apis/apigatewayv2/v1alpha1.IntegrationID()
	// +optional
	Target *string `json:"target,omitempty"`

	// TargetRef is a reference to an Integration ID
	// +optional
	TargetRef *xpv1.Reference `json:"targetRef,omitempty"`

	// TargetSelector is a selector for an Integration ID
	// +optional
	TargetSelector *xpv1.Selector `json:"targetSelector,omitempty"`
}

// CustomRouteObservation includes the custom fields.
type CustomRouteObservation struct{}

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

// CustomRouteResponseObservation includes the custom fields.
type CustomRouteResponseObservation struct{}

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

// CustomVPCLinkObservation includes the custom fields.
type CustomVPCLinkObservation struct{}

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

// CustomStageObservation includes the custom fields.
type CustomStageObservation struct{}
