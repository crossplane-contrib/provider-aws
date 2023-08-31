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

// Code generated by ack-generate. DO NOT EDIT.

package v1alpha1

import (
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// PermissionParameters defines the desired state of Permission
type PermissionParameters struct {
	// Region is which region the Permission will be created.
	// +kubebuilder:validation:Required
	Region string `json:"region"`
	// Specifies a unique, case-sensitive identifier that you provide to ensure
	// the idempotency of the request. This lets you safely retry the request without
	// accidentally performing the same operation a second time. Passing the same
	// value to a later call to an operation requires that you also pass the same
	// value for all other parameters. We recommend that you use a UUID type of
	// value. (https://wikipedia.org/wiki/Universally_unique_identifier).
	//
	// If you don't provide this value, then Amazon Web Services generates a random
	// one for you.
	//
	// If you retry the operation with the same ClientToken, but with different
	// parameters, the retry fails with an IdempotentParameterMismatch error.
	ClientToken *string `json:"clientToken,omitempty"`
	// Specifies the name of the customer managed permission. The name must be unique
	// within the Amazon Web Services Region.
	// +kubebuilder:validation:Required
	Name *string `json:"name"`
	// A string in JSON format string that contains the following elements of a
	// resource-based policy:
	//
	//    * Effect: must be set to ALLOW.
	//
	//    * Action: specifies the actions that are allowed by this customer managed
	//    permission. The list must contain only actions that are supported by the
	//    specified resource type. For a list of all actions supported by each resource
	//    type, see Actions, resources, and condition keys for Amazon Web Services
	//    services (https://docs.aws.amazon.com/service-authorization/latest/reference/reference_policies_actions-resources-contextkeys.html)
	//    in the Identity and Access Management User Guide.
	//
	//    * Condition: (optional) specifies conditional parameters that must evaluate
	//    to true when a user attempts an action for that action to be allowed.
	//    For more information about the Condition element, see IAM policies: Condition
	//    element (https://docs.aws.amazon.com/IAM/latest/UserGuide/reference_policies_elements_condition.html)
	//    in the Identity and Access Management User Guide.
	//
	// This template can't include either the Resource or Principal elements. Those
	// are both filled in by RAM when it instantiates the resource-based policy
	// on each resource shared using this managed permission. The Resource comes
	// from the ARN of the specific resource that you are sharing. The Principal
	// comes from the list of identities added to the resource share.
	// +kubebuilder:validation:Required
	PolicyTemplate *string `json:"policyTemplate"`
	// Specifies the name of the resource type that this customer managed permission
	// applies to.
	//
	// The format is <service-code>:<resource-type> and is not case sensitive. For
	// example, to specify an Amazon EC2 Subnet, you can use the string ec2:subnet.
	// To see the list of valid values for this parameter, query the ListResourceTypes
	// operation.
	// +kubebuilder:validation:Required
	ResourceType *string `json:"resourceType"`
	// Specifies a list of one or more tag key and value pairs to attach to the
	// permission.
	Tags                       []*Tag `json:"tags,omitempty"`
	CustomPermissionParameters `json:",inline"`
}

// PermissionSpec defines the desired state of Permission
type PermissionSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       PermissionParameters `json:"forProvider"`
}

// PermissionObservation defines the observed state of Permission
type PermissionObservation struct {
	// A structure with information about this customer managed permission.
	Permission *ResourceSharePermissionSummary `json:"permission,omitempty"`
}

// PermissionStatus defines the observed state of Permission.
type PermissionStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          PermissionObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// Permission is the Schema for the Permissions API
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:storageversion
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,aws}
type Permission struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              PermissionSpec   `json:"spec"`
	Status            PermissionStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// PermissionList contains a list of Permissions
type PermissionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Permission `json:"items"`
}

// Repository type metadata.
var (
	PermissionKind             = "Permission"
	PermissionGroupKind        = schema.GroupKind{Group: CRDGroup, Kind: PermissionKind}.String()
	PermissionKindAPIVersion   = PermissionKind + "." + GroupVersion.String()
	PermissionGroupVersionKind = GroupVersion.WithKind(PermissionKind)
)

func init() {
	SchemeBuilder.Register(&Permission{}, &PermissionList{})
}