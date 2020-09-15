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

// S3BucketPolicyParameters define the desired state of an AWS S3BucketPolicy.
type S3BucketPolicyParameters struct {
	// This is the current IAM policy version
	PolicyVersion string `json:"version"`

	// This is the policy's optional identifier
	PolicyID string `json:"id,omitempty"`

	// This is the list of statement this policy applies
	PolicyStatement []S3BucketPolicyStatement `json:"statement"`

	// BucketName presents the name of the bucket.
	// +optional
	BucketName *string `json:"bucketName,omitempty"`

	// BucketNameRef references to an S3Bucket to retrieve its bucketName
	// +optional
	BucketNameRef *runtimev1alpha1.Reference `json:"bucketNameRef,omitempty"`

	// BucketNameSelector selects a reference to an S3Bucket to retrieve its bucketName
	// +optional
	BucketNameSelector *runtimev1alpha1.Selector `json:"bucketNameSelector,omitempty"`

	// UserName presents the name of the bucket.
	// +optional
	UserName *string `json:"userName,omitempty"`

	// UserNameRef references to an S3Bucket to retrieve its userName
	// +optional
	UserNameRef *runtimev1alpha1.Reference `json:"userNameRef,omitempty"`

	// UserNameSelector selects a reference to an S3Bucket to retrieve its userName
	// +optional
	UserNameSelector *runtimev1alpha1.Selector `json:"userNameSelector,omitempty"`
}

// Serialize is the custom marshaller for the S3BucketPolicyParameters
func (p *S3BucketPolicyParameters) Serialize() (interface{}, error) {
	m := make(map[string]interface{})
	m["Version"] = p.PolicyVersion
	if p.PolicyID != "" {
		m["Id"] = p.PolicyID
	}
	slc := make([]interface{}, len(p.PolicyStatement))
	for i, v := range p.PolicyStatement {
		msg, err := v.Serialize()
		if err != nil {
			return nil, err
		}
		slc[i] = msg
	}
	m["Statement"] = slc
	return m, nil
}

// S3BucketPolicyStatement defines an individual statement within the
// S3BucketPolicyBody
type S3BucketPolicyStatement struct {
	// Optional identifier for this statement, must be unique within the
	// policy if provided.
	StatementID string `json:"sid,omitempty"`

	// The effect is required and specifies whether the statement results
	// in an allow or an explicit deny. Valid values for Effect are Allow and Deny.
	Effect string `json:"effect"`

	// Used with the S3 policy to specify the principal that is allowed
	// or denied access to a resource.
	Principal *S3BucketPrincipal `json:"principal,omitempty"`

	// Used with the S3 policy to specify the users which are not included
	// in this policy
	NotPrincipal *S3BucketPrincipal `json:"notPrincipal,omitempty"`

	// Each element of the PolicyAction array describes the specific
	// action or actions that will be allowed or denied with this PolicyStatement.
	PolicyAction []string `json:"action,omitempty"`

	// Each element of the NotPolicyAction array will allow the property to match
	// all but the listed actions.
	NotPolicyAction []string `json:"notAction,omitempty"`

	// This flag indicates that this policy should apply to the IAMUsername
	// that was either passed in or created for this bucket, this user will
	// added to the action array
	ApplyToIAMUser bool `json:"effectIAMUser,omitempty"`

	// The paths on which this resource will apply
	ResourcePath []string `json:"resource,omitempty"`

	// This will explicitly match all resource paths except the ones
	// specified in this array
	NotResourcePath []string `json:"notResource,omitempty"`
}

func checkExistsArray(slc []string) bool {
	return len(slc) != 0
}

// Serialize is the custom marshaller for the S3BucketPolicyStatement
func (p *S3BucketPolicyStatement) Serialize() (interface{}, error) {
	m := make(map[string]interface{})
	if p.Principal != nil {
		principal, err := p.Principal.Serialize()
		if err != nil {
			return nil, err
		}
		m["Principal"] = principal
	}
	if p.NotPrincipal != nil {
		notPrincipal, err := p.NotPrincipal.Serialize()
		if err != nil {
			return nil, err
		}
		m["NotPrincipal"] = notPrincipal
	}
	if checkExistsArray(p.PolicyAction) {
		m["Action"] = tryFirst(p.PolicyAction)
	}
	if checkExistsArray(p.NotPolicyAction) {
		m["NotAction"] = tryFirst(p.NotPolicyAction)
	}
	if checkExistsArray(p.ResourcePath) {
		m["Resource"] = tryFirst(p.ResourcePath)
	}
	if checkExistsArray(p.NotResourcePath) {
		m["NotResource"] = tryFirst(p.NotResourcePath)
	}
	m["Effect"] = p.Effect
	if p.StatementID != "" {
		m["Sid"] = p.StatementID
	}
	return m, nil
}

// S3BucketPrincipal defines the principal users affected by
// the S3BucketPolicyStatement
type S3BucketPrincipal struct {
	// This flag indicates if the policy should be made available
	// to all anonymous users.
	AllowAnon bool `json:"allowAnon,omitempty"`

	// This list contains the all of the AWS IAM users which are affected
	// by the policy statement
	AWSPrincipal []string `json:"aws,omitempty"`
}

func tryFirst(slc []string) interface{} {
	if len(slc) == 1 {
		return slc[0]
	}
	return slc
}

// Serialize is the custom serializer for the S3BucketPrincipal
func (p *S3BucketPrincipal) Serialize() (interface{}, error) {
	all := "*"
	if p.AllowAnon {
		return all, nil
	}
	m := make(map[string]interface{})
	m["AWS"] = tryFirst(p.AWSPrincipal)
	return m, nil
}

// An S3BucketPolicySpec defines the desired state of an
// S3BucketPolicy.
type S3BucketPolicySpec struct {
	runtimev1alpha1.ResourceSpec `json:",inline"`
	PolicyBody                   S3BucketPolicyParameters `json:"forProvider"`
}

// An S3BucketPolicyStatus represents the observed state of an
// S3BucketPolicy.
type S3BucketPolicyStatus struct {
	runtimev1alpha1.ResourceStatus `json:",inline"`
}

// +kubebuilder:object:root=true

// An S3BucketPolicy is a managed resource that represents an AWS Bucket
// policy.
// +kubebuilder:printcolumn:name="USERNAME",type="string",JSONPath=".spec.forProvider.userName"
// +kubebuilder:printcolumn:name="BUCKETNAME",type="string",JSONPath=".spec.forProvider.bucketName"
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,aws}
type S3BucketPolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   S3BucketPolicySpec   `json:"spec"`
	Status S3BucketPolicyStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// S3BucketPolicyList contains a list of BucketPolicies
type S3BucketPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []S3BucketPolicy `json:"items"`
}
