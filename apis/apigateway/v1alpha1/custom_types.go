package v1alpha1

import (
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
)

// CustomAPIKeyParameters includes the custom fields of APIKey
type CustomAPIKeyParameters struct{}

// CustomAPIKeyObservation includes the custom status fields of APIKey.
type CustomAPIKeyObservation struct{}

// CustomAuthorizerParameters includes the custom fields of Authorizer
type CustomAuthorizerParameters struct {
	// RestAPIID is the ID for the RestAPI.
	// +immutable
	// +crossplane:generate:reference:type=RestAPI
	RestAPIID *string `json:"restApiId,omitempty"`

	// RestAPIIDRef is a reference to an RestAPI used to set
	// the RestAPIID.
	// +optional
	RestAPIIDRef *xpv1.Reference `json:"restApiIdRef,omitempty"`

	// RestApiIdSelector selects references to RestAPI used
	// to set the RestAPIID.
	// +optional
	RestAPIIDSelector *xpv1.Selector `json:"restApiIdSelector,omitempty"`
}

// CustomAuthorizerObservation includes the custom status fields of Authorizer.
type CustomAuthorizerObservation struct{}

// CustomBasePathMappingParameters includes the custom fields of BasePathMapping
type CustomBasePathMappingParameters struct {
	// RestAPIID is the ID for the RestAPI.
	// +immutable
	// +crossplane:generate:reference:type=RestAPI
	RestAPIID *string `json:"restApiId,omitempty"`

	// RestAPIIDRef is a reference to an RestAPI used to set
	// the RestAPIID.
	// +optional
	RestAPIIDRef *xpv1.Reference `json:"restApiIdRef,omitempty"`

	// RestApiIdSelector selects references to RestAPI used
	// to set the RestAPIID.
	// +optional
	RestAPIIDSelector *xpv1.Selector `json:"restApiIdSelector,omitempty"`
}

// CustomBasePathMappingObservation includes the custom status fields of BasePathMapping.
type CustomBasePathMappingObservation struct{}

// CustomDeploymentParameters includes the custom fields of Deployment
type CustomDeploymentParameters struct {
	// RestAPIID is the ID for the RestAPI.
	// +immutable
	// +crossplane:generate:reference:type=RestAPI
	RestAPIID *string `json:"restApiId,omitempty"`

	// RestAPIIDRef is a reference to an RestAPI used to set
	// the RestAPIID.
	// +optional
	RestAPIIDRef *xpv1.Reference `json:"restApiIdRef,omitempty"`

	// RestApiIdSelector selects references to RestAPI used
	// to set the RestAPIID.
	// +optional
	RestAPIIDSelector *xpv1.Selector `json:"restApiIdSelector,omitempty"`
}

// CustomDeploymentObservation includes the custom status fields of Deployment.
type CustomDeploymentObservation struct{}

// CustomDocumentationPartParameters includes the custom fields of DocumentationPart
type CustomDocumentationPartParameters struct {
	// RestAPIID is the ID for the RestAPI.
	// +immutable
	// +crossplane:generate:reference:type=RestAPI
	RestAPIID *string `json:"restApiId,omitempty"`

	// RestAPIIDRef is a reference to an RestAPI used to set
	// the RestAPIID.
	// +optional
	RestAPIIDRef *xpv1.Reference `json:"restApiIdRef,omitempty"`

	// RestApiIdSelector selects references to RestAPI used
	// to set the RestAPIID.
	// +optional
	RestAPIIDSelector *xpv1.Selector `json:"restApiIdSelector,omitempty"`
}

// CustomDocumentationPartObservation includes the custom status fields of DocumentationPart.
type CustomDocumentationPartObservation struct{}

// CustomDocumentationVersionParameters includes the custom fields of DocumentationVersion
type CustomDocumentationVersionParameters struct {
	// RestAPIID is the ID for the RestAPI.
	// +immutable
	// +crossplane:generate:reference:type=RestAPI
	RestAPIID *string `json:"restApiId,omitempty"`

	// RestAPIIDRef is a reference to an RestAPI used to set
	// the RestAPIID.
	// +optional
	RestAPIIDRef *xpv1.Reference `json:"restApiIdRef,omitempty"`

	// RestApiIdSelector selects references to RestAPI used
	// to set the RestAPIID.
	// +optional
	RestAPIIDSelector *xpv1.Selector `json:"restApiIdSelector,omitempty"`
}

// CustomDocumentationVersionObservation includes the custom status fields of DocumentationVersion.
type CustomDocumentationVersionObservation struct{}

// CustomModelParameters includes the custom fields of Model
type CustomModelParameters struct {
	// RestAPIID is the ID for the RestAPI.
	// +immutable
	// +crossplane:generate:reference:type=RestAPI
	RestAPIID *string `json:"restApiId,omitempty"`

	// RestAPIIDRef is a reference to an RestAPI used to set
	// the RestAPIID.
	// +optional
	RestAPIIDRef *xpv1.Reference `json:"restApiIdRef,omitempty"`

	// RestApiIdSelector selects references to RestAPI used
	// to set the RestAPIID.
	// +optional
	RestAPIIDSelector *xpv1.Selector `json:"restApiIdSelector,omitempty"`
}

// CustomModelObservation includes the custom status fields of Model.
type CustomModelObservation struct{}

// CustomRequestValidatorParameters includes the custom fields of RequestValidator
type CustomRequestValidatorParameters struct {
	// RestAPIID is the ID for the RestAPI.
	// +immutable
	// +crossplane:generate:reference:type=RestAPI
	RestAPIID *string `json:"restApiId,omitempty"`

	// RestAPIIDRef is a reference to an RestAPI used to set
	// the RestAPIID.
	// +optional
	RestAPIIDRef *xpv1.Reference `json:"restApiIdRef,omitempty"`

	// RestApiIdSelector selects references to RestAPI used
	// to set the RestAPIID.
	// +optional
	RestAPIIDSelector *xpv1.Selector `json:"restApiIdSelector,omitempty"`
}

// CustomRequestValidatorObservation includes the custom status fields of RequestValidator.
type CustomRequestValidatorObservation struct{}

// CustomResourceParameters includes the custom fields of Resource
type CustomResourceParameters struct {
	// RestAPIID is the ID for the RestAPI.
	// +immutable
	// +crossplane:generate:reference:type=RestAPI
	RestAPIID *string `json:"restApiId,omitempty"`

	// RestAPIIDRef is a reference to an RestAPI used to set
	// the RestAPIID.
	// +optional
	RestAPIIDRef *xpv1.Reference `json:"restApiIdRef,omitempty"`

	// RestApiIdSelector selects references to RestAPI used
	// to set the RestAPIID.
	// +optional
	RestAPIIDSelector *xpv1.Selector `json:"restApiIdSelector,omitempty"`

	// ParentID is the ID for the Parent.
	// +immutable
	// +crossplane:generate:reference:type=Resource
	ParentResourceID *string `json:"parentResourceId,omitempty"`

	// ParentIDRef is a reference to an Parent used to set
	// the ParentID.
	// +optional
	ParentResourceIDRef *xpv1.Reference `json:"parentResourceIdRef,omitempty"`

	// RestApiIdSelector selects references to Parent used
	// to set the ParentID.
	// +optional
	ParentResourceIDSelector *xpv1.Selector `json:"parentResourceIdSelector,omitempty"`
}

// CustomResourceObservation includes the custom status fields of Resource.
type CustomResourceObservation struct{}

// CustomRestAPIParameters includes the custom fields of RestAPI
type CustomRestAPIParameters struct{}

// CustomRestAPIObservation includes the custom status fields of RestAPI.
type CustomRestAPIObservation struct{}

// CustomUsagePlanKeyParameters includes the custom fields of UsagePlanKey
type CustomUsagePlanKeyParameters struct {
	// UsagePlanID is the ID for the UsagePlan.
	// +immutable
	// +crossplane:generate:reference:type=UsagePlan
	UsagePlanID *string `json:"restApiId,omitempty"`

	// UsagePlanIDRef is a reference to an UsagePlan used to set
	// the UsagePlanID.
	// +optional
	UsagePlanIDRef *xpv1.Reference `json:"restApiIdRef,omitempty"`

	// UsagePlanIdSelector selects references to UsagePlan used
	// to set the UsagePlanID.
	// +optional
	UsagePlanIDSelector *xpv1.Selector `json:"restApiIdSelector,omitempty"`
}

// CustomUsagePlanKeyObservation includes the custom status fields of UsagePlanKey.
type CustomUsagePlanKeyObservation struct{}

// CustomUsagePlanAPIStage includes the custom fields of UsagePlan.APIStages
type CustomUsagePlanAPIStage struct {
	// RestAPIID is the ID for the RestAPI.
	// +immutable
	// +crossplane:generate:reference:type=RestAPI
	RestAPIID *string `json:"restApiId,omitempty"`

	// RestAPIIDRef is a reference to an RestAPI used to set
	// the RestAPIID.
	// +optional
	RestAPIIDRef *xpv1.Reference `json:"restApiIdRef,omitempty"`

	// RestApiIdSelector selects references to RestAPI used
	// to set the RestAPIID.
	// +optional
	RestAPIIDSelector *xpv1.Selector `json:"restApiIdSelector,omitempty"`

	Stage *string `json:"stage,omitempty"`

	Throttle map[string]*ThrottleSettings `json:"throttle,omitempty"`
}

// CustomUsagePlanParameters includes the custom fields of UsagePlan
type CustomUsagePlanParameters struct {
	// The associated API stages of the usage plan.
	APIStages []*CustomUsagePlanAPIStage `json:"apiStages,omitempty"`
}

// CustomUsagePlanObservation includes the custom status fields of UsagePlan.
type CustomUsagePlanObservation struct{}

// CustomVPCLinkParameters includes the custom fields of VPCLink
type CustomVPCLinkParameters struct{}

// CustomVPCLinkObservation includes the custom status fields of VPCLink.
type CustomVPCLinkObservation struct{}

// CustomStageCanarySettings includes the custom field Stage.CanarySettings
type CustomStageCanarySettings struct {
	// DeploymentID is the ID for the Deployment.
	// +immutable
	// +crossplane:generate:reference:type=Deployment
	DeploymentID *string `json:"deploymentId,omitempty"`

	// DeploymentIDRef is a reference to an Deployment used to set
	// the DeploymentID.
	// +optional
	DeploymentIDRef *xpv1.Reference `json:"deploymentIdRef,omitempty"`

	// DeploymentIDSelector selects references to Deployment used
	// to set the DeploymentID.
	// +optional
	DeploymentIDSelector *xpv1.Selector `json:"deploymentIdSelector,omitempty"`

	PercentTraffic *float64 `json:"percentTraffic,omitempty"`

	StageVariableOverrides map[string]*string `json:"stageVariableOverrides,omitempty"`

	UseStageCache *bool `json:"useStageCache,omitempty"`
}

// CustomStageParameters includes the custom fields of Stage
type CustomStageParameters struct {
	// RestAPIID is the ID for the RestAPI.
	// +immutable
	// +crossplane:generate:reference:type=RestAPI
	RestAPIID *string `json:"restApiId,omitempty"`

	// RestAPIIDRef is a reference to an RestAPI used to set
	// the RestAPIID.
	// +optional
	RestAPIIDRef *xpv1.Reference `json:"restApiIdRef,omitempty"`

	// RestApiIdSelector selects references to RestAPI used
	// to set the RestAPIID.
	// +optional
	RestAPIIDSelector *xpv1.Selector `json:"restApiIdSelector,omitempty"`

	// DeploymentID is the ID for the Deployment.
	// +immutable
	// +crossplane:generate:reference:type=Deployment
	DeploymentID *string `json:"deploymentId,omitempty"`

	// DeploymentIDRef is a reference to an Deployment used to set
	// the DeploymentID.
	// +optional
	DeploymentIDRef *xpv1.Reference `json:"deploymentIdRef,omitempty"`

	// DeploymentIDSelector selects references to Deployment used
	// to set the DeploymentID.
	// +optional
	DeploymentIDSelector *xpv1.Selector `json:"deploymentIdSelector,omitempty"`

	CanarySettings *CustomStageCanarySettings `json:"canarySettings,omitempty"`
}

// CustomStageObservation includes the custom status fields of Stage.
type CustomStageObservation struct{}

// CustomMethodParameters includes the custom fields of Method
type CustomMethodParameters struct {
	// RestAPIID is the ID for the RestAPI.
	// +immutable
	// +crossplane:generate:reference:type=RestAPI
	RestAPIID *string `json:"restApiId,omitempty"`

	// RestAPIIDRef is a reference to an RestAPI used to set
	// the RestAPIID.
	// +optional
	RestAPIIDRef *xpv1.Reference `json:"restApiIdRef,omitempty"`

	// RestApiIdSelector selects references to RestAPI used
	// to set the RestAPIID.
	// +optional
	RestAPIIDSelector *xpv1.Selector `json:"restApiIdSelector,omitempty"`

	// AuthorizerID is the ID for the Authorizer.
	// +immutable
	// +crossplane:generate:reference:type=Authorizer
	AuthorizerID *string `json:"authorizerId,omitempty"`

	// AuthorizerIDRef is a reference to an Authorizer used to set
	// the AuthorizerID.
	// +optional
	AuthorizerIDRef *xpv1.Reference `json:"authorizerIdRef,omitempty"`

	// RestApiIdSelector selects references to Authorizer used
	// to set the AuthorizerID.
	// +optional
	AuthorizerIDSelector *xpv1.Selector `json:"authorizerIdSelector,omitempty"`

	// ResourceID is the ID for the Resource.
	// +immutable
	// +crossplane:generate:reference:type=Resource
	ResourceID *string `json:"resourceId,omitempty"`

	// ResourceIDRef is a reference to an Resource used to set
	// the ResourceID.
	// +optional
	ResourceIDRef *xpv1.Reference `json:"resourceIdRef,omitempty"`

	// RestApiIdSelector selects references to Resource used
	// to set the ResourceID.
	// +optional
	ResourceIDSelector *xpv1.Selector `json:"resourceIdSelector,omitempty"`

	// RequestValidatorID is the ID for the RequestValidator.
	// +immutable
	// +crossplane:generate:reference:type=RequestValidator
	RequestValidatorID *string `json:"requestValidatorId,omitempty"`

	// RequestValidatorIDRef is a reference to an RequestValidator used to set
	// the RequestValidatorID.
	// +optional
	RequestValidatorIDRef *xpv1.Reference `json:"requestValidatorIdRef,omitempty"`

	// RequestValidatorIDSelector selects references to RequestValidator used
	// to set the RequestValidatorID.
	// +optional
	RequestValidatorIDSelector *xpv1.Selector `json:"requestValidatorIdSelector,omitempty"`
}

// CustomMethodObservation includes the custom status fields of Method.
type CustomMethodObservation struct{}

// CustomMethodResponseParameters includes the custom fields of MethodResponse
type CustomMethodResponseParameters struct {
	// RestAPIID is the ID for the RestAPI.
	// +immutable
	// +crossplane:generate:reference:type=RestAPI
	RestAPIID *string `json:"restApiId,omitempty"`

	// RestAPIIDRef is a reference to an RestAPI used to set
	// the RestAPIID.
	// +optional
	RestAPIIDRef *xpv1.Reference `json:"restApiIdRef,omitempty"`

	// RestApiIdSelector selects references to RestAPI used
	// to set the RestAPIID.
	// +optional
	RestAPIIDSelector *xpv1.Selector `json:"restApiIdSelector,omitempty"`
	// ResourceID is the ID for the Resource.
	// +immutable
	// +crossplane:generate:reference:type=Resource
	ResourceID *string `json:"resourceId,omitempty"`

	// ResourceIDRef is a reference to an Resource used to set
	// the ResourceID.
	// +optional
	ResourceIDRef *xpv1.Reference `json:"resourceIdRef,omitempty"`

	// RestApiIdSelector selects references to Resource used
	// to set the ResourceID.
	// +optional
	ResourceIDSelector *xpv1.Selector `json:"resourceIdSelector,omitempty"`
}

// CustomMethodResponseObservation includes the custom status fields of MethodResponse.
type CustomMethodResponseObservation struct{}

// CustomGatewayResponseParameters includes the custom fields of GatewayResponse
type CustomGatewayResponseParameters struct {
	// RestAPIID is the ID for the RestAPI.
	// +immutable
	// +crossplane:generate:reference:type=RestAPI
	RestAPIID *string `json:"restApiId,omitempty"`

	// RestAPIIDRef is a reference to an RestAPI used to set
	// the RestAPIID.
	// +optional
	RestAPIIDRef *xpv1.Reference `json:"restApiIdRef,omitempty"`

	// RestApiIdSelector selects references to RestAPI used
	// to set the RestAPIID.
	// +optional
	RestAPIIDSelector *xpv1.Selector `json:"restApiIdSelector,omitempty"`
}

// CustomGatewayResponseObservation includes the custom status fields of GatewayResponse.
type CustomGatewayResponseObservation struct{}

// CustomIntegrationResponseParameters includes the custom fields of IntegrationResponse
type CustomIntegrationResponseParameters struct {
	// RestAPIID is the ID for the RestAPI.
	// +immutable
	// +crossplane:generate:reference:type=RestAPI
	RestAPIID *string `json:"restApiId,omitempty"`

	// RestAPIIDRef is a reference to an RestAPI used to set
	// the RestAPIID.
	// +optional
	RestAPIIDRef *xpv1.Reference `json:"restApiIdRef,omitempty"`

	// RestApiIdSelector selects references to RestAPI used
	// to set the RestAPIID.
	// +optional
	RestAPIIDSelector *xpv1.Selector `json:"restApiIdSelector,omitempty"`

	// ResourceID is the ID for the Resource.
	// +immutable
	// +crossplane:generate:reference:type=Resource
	ResourceID *string `json:"resourceId,omitempty"`

	// ResourceIDRef is a reference to an Parent used to set
	// the ResourceID.
	// +optional
	ResourceIDRef *xpv1.Reference `json:"resourceIdRef,omitempty"`

	// ResourceIDSelector selects references to Parent used
	// to set the ResourceID.
	// +optional
	ResourceIDSelector *xpv1.Selector `json:"resourceIdSelector,omitempty"`
}

// CustomIntegrationResponseObservation includes the custom status fields of IntegrationResponse.
type CustomIntegrationResponseObservation struct{}

// CustomIntegrationParameters includes the custom fields of Integration
type CustomIntegrationParameters struct {
	// RestAPIID is the ID for the RestAPI.
	// +immutable
	// +crossplane:generate:reference:type=RestAPI
	RestAPIID *string `json:"restApiId,omitempty"`

	// RestAPIIDRef is a reference to an RestAPI used to set
	// the RestAPIID.
	// +optional
	RestAPIIDRef *xpv1.Reference `json:"restApiIdRef,omitempty"`

	// RestApiIdSelector selects references to RestAPI used
	// to set the RestAPIID.
	// +optional
	RestAPIIDSelector *xpv1.Selector `json:"restApiIdSelector,omitempty"`

	// ResourceID is the ID for the Resource.
	// +immutable
	// +crossplane:generate:reference:type=Resource
	ResourceID *string `json:"resourceId,omitempty"`

	// ResourceIDRef is a reference to an Parent used to set
	// the ResourceID.
	// +optional
	ResourceIDRef *xpv1.Reference `json:"resourceIdRef,omitempty"`

	// ResourceIDSelector selects references to Parent used
	// to set the ResourceID.
	// +optional
	ResourceIDSelector *xpv1.Selector `json:"resourceIdSelector,omitempty"`
}

// CustomIntegrationObservation includes the custom status fields of Integration.
type CustomIntegrationObservation struct{}

// CustomDomainNameParameters includes the custom fields of DomainName
type CustomDomainNameParameters struct{}

// CustomDomainNameObservation includes the custom status fields of DomainName.
type CustomDomainNameObservation struct{}
