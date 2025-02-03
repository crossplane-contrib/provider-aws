package v1alpha1

import xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"

// CustomCertificate includes custom fields about certificates.
type CustomCertificate struct {
	// [HTTPS and TLS listeners] The default certificate for the listener.
	// +optional
	CertificateARN *string `json:"certificateARN,omitempty"`

	// Reference to Certificates for Certificate ARN
	// +optional
	CertificateARNRef *xpv1.Reference `json:"certificateARNRef,omitempty"`

	// Selector for references to Certificate for CertificateArn
	// +optional
	CertificateARNSelector *xpv1.Selector `json:"certificateARNSelector,omitempty"`

	// +optional
	IsDefault bool `json:"isDefault,omitempty"`
}

// CustomTargetGroupTuple includes custom fields about target groups.
// Only used with ForwardActionConfig to route to multiple target groups.
type CustomTargetGroupTuple struct { // inject refs and selectors into TargetGroupTuple
	// Provides information about how traffic will be
	// distributed between multiple target groups in a forward rule.
	TargetGroupTuple `json:",inline"`

	// Reference to TargetGroupARN used to set TargetGroupARN
	// +optional
	TargetGroupARNRef *xpv1.Reference `json:"targetGroupArnRef,omitempty"`

	// Selector for references to TargetGroup for TargetGroupARN
	// +optional
	TargetGroupARNSelector *xpv1.Selector `json:"targetGroupArnSelector,omitempty"`
}

// CustomForwardActionConfig includes custom fields about a forward action.
type CustomForwardActionConfig struct {
	// Information about the target group stickiness for a rule.
	TargetGroupStickinessConfig *TargetGroupStickinessConfig `json:"targetGroupStickinessConfig,omitempty"`

	// One or more target groups. For Network Load Balancers, you can specify a
	// single target group.
	TargetGroups []*CustomTargetGroupTuple `json:"targetGroups,omitempty"`
}

// CustomAction includes custom fields for an action.
//
// Each rule must include exactly one of the following types of actions: forward,
// fixed-response, or redirect, and it must be the last action to be performed.
type CustomAction struct {
	// Request parameters to use when integrating with Amazon Cognito to authenticate
	// users.
	AuthenticateCognitoConfig *AuthenticateCognitoActionConfig `json:"authenticateCognitoConfig,omitempty"`
	// Request parameters when using an identity provider (IdP) that is compliant
	// with OpenID Connect (OIDC) to authenticate users.
	AuthenticateOidcConfig *AuthenticateOIDCActionConfig `json:"authenticateOidcConfig,omitempty"`
	// Information about an action that returns a custom HTTP response.
	FixedResponseConfig *FixedResponseActionConfig `json:"fixedResponseConfig,omitempty"`
	// Information about a forward action.
	ForwardConfig *CustomForwardActionConfig `json:"forwardConfig,omitempty"`

	// The order for the action. This value is required for rules with multiple
	// actions. The action with the lowest value for order is performed first.
	Order *int64 `json:"order,omitempty"`
	// Information about a redirect action.
	//
	// A URI consists of the following components: protocol://hostname:port/path?query.
	// You must modify at least one of the following components to avoid a redirect
	// loop: protocol, hostname, port, or path. Any components that you do not modify
	// retain their original values.
	//
	// You can reuse URI components using the following reserved keywords:
	//
	//    * #{protocol}
	//
	//    * #{host}
	//
	//    * #{port}
	//
	//    * #{path} (the leading "/" is removed)
	//
	//    * #{query}
	//
	// For example, you can change the path to "/new/#{path}", the hostname to "example.#{host}",
	// or the query to "#{query}&value=xyz".
	RedirectConfig *RedirectActionConfig `json:"redirectConfig,omitempty"`

	// The Amazon Resource Name (ARN) of the target group. Specify only when
	// actionType is forward and you want to route to a single target group.
	// To route to one or more target groups, use ForwardConfig instead.
	// +optional
	TargetGroupARN *string `json:"targetGroupArn,omitempty"`

	// Reference to TargetGroupARN used to set TargetGroupARN
	// +optional
	TargetGroupARNRef *xpv1.Reference `json:"targetGroupArnRef,omitempty"`

	// Selector for references to TargetGroups for TargetGroupARNs
	// +optional
	TargetGroupARNSelector *xpv1.Selector `json:"targetGroupArnSelector,omitempty"`

	// The type of action.
	// +kubebuilder:validation:Required
	Type *string `json:"actionType"` // renamed json tag from "type_"
}

// CustomListenerParameters includes the custom fields of Listener.
type CustomListenerParameters struct {
	// [HTTPS and TLS listeners] The default certificate
	// for the listener. You must provide exactly one certificate.
	// Set CertificateArn to the certificate ARN but do not set IsDefault.
	// +optional
	Certificates []*CustomCertificate `json:"certificates,omitempty"`

	// The actions for the default rule.
	// +kubebuilder:validation:Required
	DefaultActions []*CustomAction `json:"defaultActions"`

	// The Amazon Resource Name (ARN) of the load balancer.
	// +optional
	LoadBalancerARN *string `json:"loadBalancerArn,omitempty"`

	// Ref to loadbalancer ARN
	// +optional
	LoadBalancerARNRef *xpv1.Reference `json:"loadBalancerArnRef,omitempty"`

	// Selector for references to LoadBalancer for LoadBalancerARN
	// +optional
	LoadBalancerARNSelector *xpv1.Selector `json:"loadBalancerArnSelector,omitempty"`
}

// CustomListenerObservation includes the custom status fields of Listener.
type CustomListenerObservation struct{}

// CustomLoadBalancerParameters includes the custom fields of LoadBalancer.
type CustomLoadBalancerParameters struct {
	// The type of load balancer. The default is application.
	Type *string `json:"loadBalancerType,omitempty"`

	// Reference to Security Groups for SecurityGroups field
	// +optional
	SecurityGroupRefs []xpv1.Reference `json:"securityGroupRefs,omitempty"`

	// Selector for references to SecurityGroups
	// +optional
	SecurityGroupSelector *xpv1.Selector `json:"securityGroupSelector,omitempty"`

	// Reference to Subnets for Subnets field
	// +optional
	SubnetRefs []xpv1.Reference `json:"subnetRefs,omitempty"`

	// Selector for references to Subnets
	// +optional
	SubnetSelector *xpv1.Selector `json:"subnetSelector,omitempty"`
}

// CustomLoadBalancerObservation includes the custom status fields of LoadBalancer.
type CustomLoadBalancerObservation struct{}

// CustomTargetGroupParameters includes the custom fields of TargetGroup.
type CustomTargetGroupParameters struct {
	// Reference to VPC for VPCID
	// +optional
	VPCIDRef *xpv1.Reference `json:"vpcIdRef,omitempty"`

	// Selector for references to VPC for VPCID
	// +optional
	VPCIDSelector *xpv1.Selector `json:"vpcIdSelector,omitempty"`
}

// CustomTargetGroupObservation includes the custom status fields of TargetGroup.
type CustomTargetGroupObservation struct{}
