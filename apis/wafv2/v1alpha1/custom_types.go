package v1alpha1

// CustomWebACLObservation includes the custom status fields of WebACL
type CustomWebACLObservation struct{}

// CustomWebACLParameters includes the custom fields.
type CustomWebACLParameters struct {
	// A list of the Amazon Resource Name (ARN) of the resources to associate with the web ACL.
	// The association is implemented only for REGIONAL scope yet. The resources will be ignored if the scope is CLOUDFRONT
	AssociatedAWSResources []*AssociatedResource `json:"associatedAWSResources,omitempty"`
	// Raw JSON string for rules configuration. Mutually exclusive with rules.
	RulesJSON *string `json:"rulesJSON,omitempty"`
	// Used for regional scoped configuration. Specifies the types of resources which the provider checks while in the reconciliation loop.
	// Each type requires additional request to aws api and appropriate permissions.
	// Only enableApplicationLoadBalancer is enabled by default.
	RegionalResourceTypeAssociation *RegionalResourceTypeAssociation `json:"regionalResourceTypeAssociation,omitempty"`
}

type AssociatedResource struct {
	ResourceARN *string `json:"resourceARN,omitempty"`
}

type RegionalResourceTypeAssociation struct {
	EnableApplicationLoadBalancer *bool `json:"enableApplicationLoadBalancer,omitempty"`
	EnableApiGateway              *bool `json:"enableApiGateway,omitempty"`
	EnableAppsync                 *bool `json:"enableAppsync,omitempty"`
	EnableCognitoUserPool         *bool `json:"enableCognitoUserPool,omitempty"`
	EnableAppRunnerService        *bool `json:"enabledAppRunnerService,omitempty"`
	EnableVerifiedAccessInstance  *bool `json:"enableVerifiedAccessInstance,omitempty"`
}
