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
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PolicyParameters define the desired state of an AWS IAM Policy.
type PolicyParameters struct {
	// A description of the policy.
	// +optional
	Description *string `json:"description,omitempty"`

	// The path to the policy.
	// +optional
	Path *string `json:"path,omitempty"`

	// The JSON policy document that is the content for the policy.
	Document string `json:"document"`

	// The name of the policy.
	Name string `json:"name"`

	// Tags. For more information about
	// tagging, see Tagging IAM Identities (https://docs.aws.amazon.com/IAM/latest/UserGuide/id_tags.html)
	// in the IAM User Guide.
	// +immutable
	// +optional
	Tags []Tag `json:"tags,omitempty"`
}

// A PolicySpec defines the desired state of a Policy.
type PolicySpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       PolicyParameters `json:"forProvider"`
}

// PolicyObservation keeps the state for the external resource
type PolicyObservation struct {
	// The Amazon PolicyObservation Name (ARN) of the policy
	ARN string `json:"arn,omitempty"`

	// The number of entities (users, groups, and roles) that the policy is attached
	// to.
	AttachmentCount int32 `json:"attachmentCount,omitempty"`

	// The identifier for the version of the policy that is set as the default version.
	DefaultVersionID string `json:"defaultVersionId,omitempty"`

	// Specifies whether the policy can be attached to an IAM user, group, or role.
	IsAttachable bool `json:"isAttachable,omitempty"`

	// The number of entities (users and roles) for which the policy is used to
	// set the permissions boundary.
	PermissionsBoundaryUsageCount int32 `json:"permissionsBoundaryUsageCount,omitempty"`

	// The stable and unique string identifying the policy.
	PolicyID string `json:"policyId,omitempty"`
}

// A PolicyStatus represents the observed state of a Policy.
type PolicyStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          PolicyObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A Policy is a managed resource that represents an AWS IAM Policy.
// +kubebuilder:printcolumn:name="ARN",type="string",JSONPath=".status.atProvider.arn"
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,aws}
type Policy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PolicySpec   `json:"spec"`
	Status PolicyStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// PolicyList contains a list of Policies
type PolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Policy `json:"items"`
}
