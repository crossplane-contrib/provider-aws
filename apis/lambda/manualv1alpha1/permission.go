/*
Copyright 2022 The Crossplane Authors.

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

package manualv1alpha1

import (
	"fmt"
	"hash/fnv"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

// PermissionParameters define the desired state of a Lambda Permission
type PermissionParameters struct {
	// Region is which region the Function will be created.
	// +kubebuilder:validation:Required
	Region string `json:"region"`

	// The name of the Lambda function, version, or alias. Name formats
	//
	// * Function
	// name - my-function (name-only), my-function:v1 (with alias).
	//
	// * Function ARN -
	// arn:aws:lambda:us-west-2:123456789012:function:my-function.
	//
	// * Partial ARN -
	// 123456789012:function:my-function.
	//
	// You can append a version number or alias to
	// any of the formats. The length constraint applies only to the full ARN. If you
	// specify only the function name, it is limited to 64 characters in length.
	//
	// This member is required.
	// +optional
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-aws/apis/lambda/v1beta1.Function
	// +crossplane:generate:reference:refFieldName=FunctionNameRef
	// +crossplane:generate:reference:selectorFieldName=FunctionNameSelector
	FunctionName *string `json:"functionName,omitempty"`

	// FunctionNameRef is a reference to a function used to set
	// the FunctionName.
	// +optional
	FunctionNameRef *xpv1.Reference `json:"functionNameRef,omitempty"`

	// FunctionNameSelector selects references to function used
	// to set the FunctionName.
	// +optional
	FunctionNameSelector *xpv1.Selector `json:"functionNameSelector,omitempty"`

	// The action that the principal can use on the function. For example,
	// lambda:InvokeFunction or lambda:GetFunction.
	//
	// This member is required.
	Action string `json:"action"`

	// The Amazon Web Services service or account that invokes the function. If you
	// specify a service, use SourceArn or SourceAccount to limit who can invoke the
	// function through that service.
	//
	// This member is required.
	Principal string `json:"principal"`

	// For Alexa Smart Home functions, a token that must be supplied by the invoker.
	EventSourceToken *string `json:"eventSourceToken,omitempty"`

	// The identifier for your organization in Organizations. Use this to grant
	// permissions to all the Amazon Web Services accounts under this organization.
	PrincipalOrgID *string `json:"principalOrgId,omitempty"`

	// For Amazon S3, the ID of the account that owns the resource. Use this together
	// with SourceArn to ensure that the resource is owned by the specified account. It
	// is possible for an Amazon S3 bucket to be deleted by its owner and recreated by
	// another account.
	SourceAccount *string `json:"sourceAccount,omitempty"`

	// For Amazon Web Services services, the ARN of the Amazon Web Services resource
	// that invokes the function. For example, an Amazon S3 bucket or Amazon SNS topic.
	// Note that Lambda configures the comparison using the StringLike operator.
	SourceArn *string `json:"sourceARN,omitempty"`

	// TODO: Implement revision support:

	// // Only update the policy if the revision ID matches the ID that's specified. Use
	// // this option to avoid modifying a policy that has changed since you last read it.
	// RevisionId *string `json:"revisionID,omitempty"`

	// // Specify a version or alias to add permissions to a published version of the
	// // function.
	// Qualifier *string `json:"qualifier,omitempty"`
}

// A PermissionSpec defines the desired state of a Permission.
type PermissionSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       PermissionParameters `json:"forProvider"`
}

// Hash calcuates the hash of the PermissionSpec.
func (ps *PermissionSpec) Hash() string {
	h := fnv.New64a()
	y, err := yaml.Marshal(ps)
	if err != nil {
		// I believe this should be impossible given we're marshalling a
		// known, strongly typed struct.
		return "unknown"
	}
	h.Write(y)
	return fmt.Sprintf("%x", h.Sum64())
}

// PermissionObservation keeps the state for the external resource
type PermissionObservation struct {
	RevisionID *string `json:"revisionId,omitempty"`
}

// A PermissionStatus represents the observed state of a ElasticIP.
type PermissionStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          PermissionObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A Permission is a managed resource that represents a AWS Lambda Permission.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="SID",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,aws}
type Permission struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PermissionSpec   `json:"spec"`
	Status PermissionStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// PermissionList contains a list of Permissions
type PermissionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Permission `json:"items"`
}
