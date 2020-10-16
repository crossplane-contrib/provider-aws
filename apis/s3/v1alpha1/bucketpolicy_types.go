/*
Copyright 2019 The Crossplane Authors.

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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
)

// BucketPolicyParameters define the desired state of an AWS BucketPolicy.
type BucketPolicyParameters struct {

	// Region is where the Bucket referenced by this BucketPolicy resides.
	Region string `json:"region"`

	// This is the current IAM policy version
	PolicyVersion string `json:"version"`

	// This is the policy's optional identifier
	PolicyID string `json:"id,omitempty"`

	// This is the list of statement this policy applies
	PolicyStatement []BucketPolicyStatement `json:"statement"`

	// BucketName presents the name of the bucket.
	// +optional
	BucketName *string `json:"bucketName,omitempty"`

	// BucketNameRef references to an S3Bucket to retrieve its bucketName
	// +optional
	BucketNameRef *runtimev1alpha1.Reference `json:"bucketNameRef,omitempty"`

	// BucketNameSelector selects a reference to an S3Bucket to retrieve its bucketName
	// +optional
	BucketNameSelector *runtimev1alpha1.Selector `json:"bucketNameSelector,omitempty"`
}

// BucketPolicyStatement defines an individual statement within the
// BucketPolicyBody
type BucketPolicyStatement struct {
	// Optional identifier for this statement, must be unique within the
	// policy if provided.
	// +optional
	StatementID *string `json:"sid,omitempty"`

	// The effect is required and specifies whether the statement results
	// in an allow or an explicit deny. Valid values for Effect are Allow and Deny.
	// +kubebuilder:validation:Enum=Allow;Deny
	Effect string `json:"effect"`

	// Used with the S3 policy to specify the principal that is allowed
	// or denied access to a resource.
	// +optional
	Principal *BucketPrincipal `json:"principal,omitempty"`

	// Used with the S3 policy to specify the users which are not included
	// in this policy
	// +optional
	NotPrincipal *BucketPrincipal `json:"notPrincipal,omitempty"`

	// Each element of the PolicyAction array describes the specific
	// action or actions that will be allowed or denied with this PolicyStatement.
	// +optional
	PolicyAction []string `json:"action,omitempty"`

	// Each element of the NotPolicyAction array will allow the property to match
	// all but the listed actions.
	// +optional
	NotPolicyAction []string `json:"notAction,omitempty"`

	// The paths on which this resource will apply
	// +optional
	ResourcePath []string `json:"resource,omitempty"`

	// This will explicitly match all resource paths except the ones
	// specified in this array
	// +optional
	NotResourcePath []string `json:"notResource,omitempty"`

	// Condition specifies where conditions for policy are in effect.
	// https://docs.aws.amazon.com/AmazonS3/latest/dev/amazon-s3-policy-keys.html
	// +optional
	ConditionBlock map[string]Condition `json:"condition,omitempty"`
}

// BucketPrincipal defines the principal users affected by
// the BucketPolicyStatement
type BucketPrincipal struct {
	// This flag indicates if the policy should be made available
	// to all anonymous users.
	AllowAnon bool `json:"allowAnon,omitempty"`

	// This list contains the all of the AWS IAM users which are affected
	// by the policy statement.
	AWSPrincipals []AWSPrincipal `json:"aws,omitempty"`

	// This string contains the identifier for any federated web identity
	// provider.
	// +optional
	Federated *string `json:"federated,omitempty"`

	// Service define the services which can have access to this bucket
	// +optional
	Service []string `json:"service,omitempty"`
}

// AWSPrincipal wraps the potential values a policy
// principal can take. Only one of the values should be set.
type AWSPrincipal struct {
	// IAMUserARN contains the ARN of an IAM user
	// +optional
	IAMUserARN *string `json:"IAMUserArn,omitempty"`

	// IAMUserARNRef contains the reference to an IAMUser
	// +optional
	IAMUserARNRef *runtimev1alpha1.Reference `json:"IAMUserArnRef,omitempty"`

	// IAMUserARNSelector queries for an IAMUser to retrieve its userName
	// +optional
	IAMUserARNSelector *runtimev1alpha1.Selector `json:"IAMUserArnSelector,omitempty"`

	// AWSAccountID identifies an AWS account as the principal
	// +optional
	AWSAccountID *string `json:"awsAccountId,omitempty"`

	// IAMRoleARN contains the ARN of an IAM role
	// +optional
	IAMRoleARN *string `json:"IAMRoleArn,omitempty"`

	// IAMRoleARNRef contains the reference to an IAMRole
	// +optional
	IAMRoleARNRef *runtimev1alpha1.Reference `json:"IAMRoleArnRef,omitempty"`

	// IAMRoleARNSelector queries for an IAM role to retrieve its userName
	// +optional
	IAMRoleARNSelector *runtimev1alpha1.Selector `json:"IAMRoleArnSelector,omitempty"`
}

// Condition represents one condition inside of the set of conditions for
// a bucket policy
type Condition struct {
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
}

// An BucketPolicySpec defines the desired state of an
// BucketPolicy.
type BucketPolicySpec struct {
	runtimev1alpha1.ResourceSpec `json:",inline"`
	PolicyBody                   BucketPolicyParameters `json:"forProvider"`
}

// An BucketPolicyStatus represents the observed state of an
// BucketPolicy.
type BucketPolicyStatus struct {
	runtimev1alpha1.ResourceStatus `json:",inline"`
}

// +kubebuilder:object:root=true

// An BucketPolicy is a managed resource that represents an AWS Bucket
// policy.
// +kubebuilder:printcolumn:name="USERNAME",type="string",JSONPath=".spec.forProvider.userName"
// +kubebuilder:printcolumn:name="BUCKETNAME",type="string",JSONPath=".spec.forProvider.bucketName"
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,aws}
type BucketPolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BucketPolicySpec   `json:"spec"`
	Status BucketPolicyStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// BucketPolicyList contains a list of BucketPolicies
type BucketPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []BucketPolicy `json:"items"`
}
