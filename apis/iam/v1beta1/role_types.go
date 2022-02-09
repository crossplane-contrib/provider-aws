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

package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
)

// Tag represents user-provided metadata that can be associated
// with a IAM role. For more information about tagging,
// see Tagging IAM Identities (https://docs.aws.amazon.com/IAM/latest/UserGuide/id_tags.html)
// in the IAM User Guide.
type Tag struct {

	// The key name that can be used to look up or retrieve the associated value.
	// For example, Department or Cost Center are common choices.
	Key string `json:"key"`

	// The value associated with this tag. For example, tags with a key name of
	// Department could have values such as Human Resources, Accounting, and Support.
	// Tags with a key name of Cost Center might have values that consist of the
	// number associated with the different cost centers in your company. Typically,
	// many resources have tags with the same key name but with different values.
	//
	// AWS always interprets the tag Value as a single string. If you need to store
	// an array, you can store comma-separated values in the string. However, you
	// must interpret the value in your code.
	// +optional
	Value string `json:"value,omitempty"`
}

// RoleParameters define the desired state of an AWS IAM Role.
type RoleParameters struct {

	// AssumeRolePolicyDocument is the the trust relationship policy document
	// that grants an entity permission to assume the role.
	// +immutable
	AssumeRolePolicyDocument string `json:"assumeRolePolicyDocument"`

	// Description is a description of the role.
	// +optional
	Description *string `json:"description,omitempty"`

	// MaxSessionDuration is the duration (in seconds) that you want to set for the specified
	// role. The default maximum of one hour is applied. This setting can have a value from 1 hour to 12 hours.
	// Default: 3600
	// +optional
	MaxSessionDuration *int32 `json:"maxSessionDuration,omitempty"`

	// Path is the path to the role.
	// Default: /
	// +immutable
	// +optional
	Path *string `json:"path,omitempty"`

	// PermissionsBoundary is the ARN of the policy that is used to set the permissions boundary for the role.
	// +immutable
	// +optional
	PermissionsBoundary *string `json:"permissionsBoundary,omitempty"`

	// Tags. For more information about
	// tagging, see Tagging IAM Identities (https://docs.aws.amazon.com/IAM/latest/UserGuide/id_tags.html)
	// in the IAM User Guide.
	// +immutable
	// +optional
	Tags []Tag `json:"tags,omitempty"`
}

// An RoleSpec defines the desired state of an Role.
type RoleSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       RoleParameters `json:"forProvider"`
}

// RoleExternalStatus keeps the state for the external resource
type RoleExternalStatus struct {
	// ARN is the Amazon Resource Name (ARN) specifying the role. For more information
	// about ARNs and how to use them in policies, see IAM Identifiers (http://docs.aws.amazon.com/IAM/latest/UserGuide/Using_Identifiers.html)
	// in the IAM User Guide guide.
	ARN string `json:"arn"`

	// RoleID is the stable and unique string identifying the role. For more information about
	// IDs, see IAM Identifiers (http://docs.aws.amazon.com/IAM/latest/UserGuide/Using_Identifiers.html)
	// in the Using IAM guide.
	RoleID string `json:"roleID"`
}

// An RoleStatus represents the observed state of an Role.
type RoleStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          RoleExternalStatus `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// An Role is a managed resource that represents an AWS IAM Role.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,aws},shortName=iamrole
type Role struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RoleSpec   `json:"spec"`
	Status RoleStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// RoleList contains a list of Roles
type RoleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Role `json:"items"`
}
