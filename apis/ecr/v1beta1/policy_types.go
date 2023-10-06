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

package v1beta1

import (
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// RepositoryPolicyParameters define the desired state of an AWS Elastic Container Repository
type RepositoryPolicyParameters struct {

	// Region is the region you'd like your RepositoryPolicy to be created in.
	Region string `json:"region"`

	// If the policy you are attempting to set on a repository policy would prevent
	// you from setting another policy in the future, you must force the SetRepositoryPolicy
	// operation. This is intended to prevent accidental repository lock outs.
	// +optional
	Force *bool `json:"force,omitempty"`

	// Policy is a well defined type which can be parsed into an JSON Repository Policy
	// either policy or rawPolicy must be specified in the policy
	// +optional
	Policy *RepositoryPolicyBody `json:"policy,omitempty"`

	// Policy stringified version of JSON repository policy
	// either policy or rawPolicy must be specified in the policy
	// +optional
	RawPolicy *string `json:"rawPolicy,omitempty"`

	// The AWS account ID associated with the registry that contains the repository.
	// If you do not specify a registry, the default registry is assumed.
	// +optional
	// +immutable
	RegistryID *string `json:"registryId,omitempty"`

	// The name of the repository to receive the policy.
	//
	// One of RepositoryName, RepositoryNameRef, or RepositoryNameSelector is required.
	// +optional
	// +immutable
	RepositoryName *string `json:"repositoryName,omitempty"`
	// contains filtered or unexported fields

	// A referencer to retrieve the name of a repository
	// One of RepositoryName, RepositoryNameRef, or RepositoryNameSelector is required.
	// +immutable
	RepositoryNameRef *xpv1.Reference `json:"repositoryNameRef,omitempty"`

	// A selector to select a referencer to retrieve the name of a repository
	// One of RepositoryName, RepositoryNameRef, or RepositoryNameSelector is required.
	// +immutable
	RepositoryNameSelector *xpv1.Selector `json:"repositoryNameSelector,omitempty"`
}

// RepositoryPolicyBody represents an ECR Repository policy in the manifest
type RepositoryPolicyBody struct {
	// Version is the current IAM policy version
	// +kubebuilder:validation:Enum="2012-10-17";"2008-10-17"
	// +kubebuilder:default:="2012-10-17"
	Version string `json:"version"`

	// ID is the policy's optional identifier
	// +immutable
	// +optional
	ID *string `json:"id,omitempty"`

	// Statements is the list of statement this policy applies
	// either jsonStatements or statements must be specified in the policy
	// +optional
	Statements []RepositoryPolicyStatement `json:"statements,omitempty"`
}

// RepositoryPolicyStatement defines an individual statement within the
// RepositoryPolicyBody
type RepositoryPolicyStatement struct {
	// Optional identifier for this statement, must be unique within the
	// policy if provided.
	// +optional
	SID *string `json:"sid,omitempty"`

	// The effect is required and specifies whether the statement results
	// in an allow or an explicit deny. Valid values for Effect are Allow and Deny.
	// +kubebuilder:validation:Enum=Allow;Deny
	Effect string `json:"effect"`

	// Used with the Repository policy to specify the principal that is allowed
	// or denied access to a resource.
	// +optional
	Principal *RepositoryPrincipal `json:"principal,omitempty"`

	// Used with the Repository policy to specify the users which are not included
	// in this policy
	// +optional
	NotPrincipal *RepositoryPrincipal `json:"notPrincipal,omitempty"`

	// Each element of the PolicyAction array describes the specific
	// action or actions that will be allowed or denied with this PolicyStatement.
	// +optional
	Action []string `json:"action,omitempty"`

	// Each element of the NotPolicyAction array will allow the property to match
	// all but the listed actions.
	// +optional
	NotAction []string `json:"notAction,omitempty"`

	// The paths on which this resource will apply
	// +optional
	Resource []string `json:"resource,omitempty"`

	// This will explicitly match all resource paths except the ones
	// specified in this array
	// +optional
	NotResource []string `json:"notResource,omitempty"`

	// Condition specifies where conditions for policy are in effect.
	// https://docs.aws.amazon.com/service-authorization/latest/reference/list_amazonelasticcontainerregistry.html#amazonelasticcontainerregistry-policy-keys
	// +optional
	Condition []Condition `json:"condition,omitempty"`
}

// RepositoryPrincipal defines the principal users affected by
// the RepositoryPolicyStatement
// Please see the AWS ECR docs for more information
// https://docs.aws.amazon.com/AmazonECR/latest/userguide/repository-policies.html
type RepositoryPrincipal struct {
	// This flag indicates if the policy should be made available
	// to all anonymous users. Principal: "*"
	// +optional
	AllowAnon *bool `json:"allowAnon,omitempty"`
	// This list contains the all of the AWS IAM users which are affected
	// by the policy statement.
	// +optional
	AWSPrincipals []AWSPrincipal `json:"awsPrincipals,omitempty"`
	// Service define the services which can have access to this bucket
	// +optional
	Service []string `json:"service,omitempty"`
	// Raw string input can be used for *
	// +optional
	Raw *string `json:"raw,omitempty"`
}

// AWSPrincipal wraps the potential values a policy
// principal can take. Only one of the values should be set.
type AWSPrincipal struct {
	// UserARN contains the ARN of an IAM user
	// +optional
	// +immutable
	UserARN *string `json:"iamUserArn,omitempty"`

	// UserARNRef contains the reference to an User
	// +optional
	UserARNRef *xpv1.Reference `json:"iamUserArnRef,omitempty"`

	// UserARNSelector queries for an User to retrieve its userName
	// +optional
	UserARNSelector *xpv1.Selector `json:"iamUserArnSelector,omitempty"`

	// AWSAccountID identifies an AWS account as the principal
	// +optional
	// +immutable
	AWSAccountID *string `json:"awsAccountId,omitempty"`

	// IAMRoleARN contains the ARN of an IAM role
	// +optional
	// +immutable
	IAMRoleARN *string `json:"iamRoleArn,omitempty"`

	// IAMRoleARNRef contains the reference to an IAMRole
	// +optional
	IAMRoleARNRef *xpv1.Reference `json:"iamRoleArnRef,omitempty"`

	// IAMRoleARNSelector queries for an IAM role to retrieve its userName
	// +optional
	IAMRoleARNSelector *xpv1.Selector `json:"iamRoleArnSelector,omitempty"`
}

// Condition represents a set of condition pairs for a Repository policy
type Condition struct {
	// OperatorKey matches the condition key and value in the policy against values in the request context
	OperatorKey string `json:"operatorKey"`

	// Conditions represents each of the key/value pairs for the operator key
	Conditions []ConditionPair `json:"conditions"`
}

// ConditionPair represents one condition inside of the set of conditions for
// a Repository policy
type ConditionPair struct {
	// ConditionKey is the key condition being applied to the parent condition
	ConditionKey string `json:"key"`

	// ConditionStringValue is the expected string value of the key from the parent condition
	// +optional
	ConditionStringValue *string `json:"stringValue,omitempty"`

	// ConditionDateValue is the expected string value of the key from the parent condition. The
	// date value must be in ISO 8601 format. The time is always midnight UTC.
	// +optional
	ConditionDateValue *metav1.Time `json:"dateValue,omitempty"`

	// ConditionNumericValue is the expected string value of the key from the parent condition
	// +optional
	ConditionNumericValue *int64 `json:"numericValue,omitempty"`

	// ConditionBooleanValue is the expected boolean value of the key from the parent condition
	// +optional
	ConditionBooleanValue *bool `json:"booleanValue,omitempty"`

	// ConditionListValue is the list value of the key from the parent condition
	// +optional
	ConditionListValue []string `json:"listValue,omitempty"`
}

// A RepositoryPolicySpec defines the desired state of a Elastic Container Repository.
type RepositoryPolicySpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       RepositoryPolicyParameters `json:"forProvider"`
}

// RepositoryPolicyObservation keeps the state for the external resource
type RepositoryPolicyObservation struct{}

// A RepositoryPolicyStatus represents the observed state of a repository policy
type RepositoryPolicyStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          RepositoryPolicyObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:storageversion

// A RepositoryPolicy is a managed resource that represents an Elastic Container Repository Policy
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="ID",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,aws}
type RepositoryPolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RepositoryPolicySpec   `json:"spec"`
	Status RepositoryPolicyStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// RepositoryPolicyList contains a list of RepositoryPolicies
type RepositoryPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RepositoryPolicy `json:"items"`
}
