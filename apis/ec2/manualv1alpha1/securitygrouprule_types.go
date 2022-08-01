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

package manualv1alpha1

import (
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// SecurityGroupRuleParameters define the desired state of the SecurityGroupRule
type SecurityGroupRuleParameters struct {
	// +kubebuilder:validation:Optional
	SecurityGroupRuleBlock *string `json:"SecurityGroupRuleBlock,omitempty"`

	// +kubebuilder:validation:Optional
	Description *string `json:"description,omitempty"`

	// +kubebuilder:validation:Required
	FromPort *float64 `json:"fromPort"`

	// +kubebuilder:validation:Optional
	IPv6SecurityGroupRuleBlock *string `json:"ipv6SecurityGroupRuleBlock,omitempty"`

	// +kubebuilder:validation:Optional
	PrefixListId *string `json:"prefixListId,omitempty"`

	// +kubebuilder:validation:Required
	Protocol *string `json:"protocol"`

	// Region is the region you'd like your resource to be created in.
	// +kubebuilder:validation:Required
	Region *string `json:"region"`

	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-aws/apis/ec2/v1beta1.SecurityGroup
	// +kubebuilder:validation:Optional
	SecurityGroupID *string `json:"securityGroupId,omitempty"`

	// +kubebuilder:validation:Optional
	SecurityGroupIDRef *xpv1.Reference `json:"securityGroupIdRef,omitempty"`

	// +kubebuilder:validation:Optional
	SecurityGroupIDSelector *xpv1.Selector `json:"securityGroupIdSelector,omitempty"`

	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-aws/apis/ec2/v1beta1.SecurityGroup
	// +kubebuilder:validation:Optional
	SourceSecurityGroupID *string `json:"sourceSecurityGroupId,omitempty"`

	// +kubebuilder:validation:Optional
	SourceSecurityGroupIDRef *xpv1.Reference `json:"sourceSecurityGroupIdRef,omitempty"`

	// +kubebuilder:validation:Optional
	SourceSecurityGroupIDSelector *xpv1.Selector `json:"sourceSecurityGroupIdSelector,omitempty"`

	// +kubebuilder:validation:Required
	ToPort *float64 `json:"toPort"`

	// Type of rule, ingress (inbound) or egress (outbound).
	// +kubebuilder:validation:Required
	Type *string `json:"type"`
}

// A SecurityGroupRuleSpec defines the desired state of a SecurityGroupRule.
type SecurityGroupRuleSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       SecurityGroupRuleParameters `json:"forProvider"`
}

// SecurityGroupRuleObservation keeps the state for the external resource
type SecurityGroupRuleObservation struct {
	// The association ID for the SecurityGroupRule block.
	SecurityGroupRuleID *string `json:"SecurityGroupRuleId,omitempty"`
}

// SecurityGroupRuleState represents the state of a SecurityGroupRule Block
type SecurityGroupRuleState struct {

	// The state of the SecurityGroupRule block.
	State *string `json:"state,omitempty"`

	// A message about the status of the SecurityGroupRule block, if applicable.
	StatusMessage *string `json:"statusMessage,omitempty"`
}

// A SecurityGroupRuleStatus represents the observed state of a SecurityGroupRule.
type SecurityGroupRuleStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          SecurityGroupRuleObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A SecurityGroupRule is a managed resource that represents an SecurityGroupRule
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,aws}
type SecurityGroupRule struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SecurityGroupRuleSpec   `json:"spec"`
	Status SecurityGroupRuleStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// SecurityGroupRuleList contains a list of SecurityGroupRules
type SecurityGroupRuleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SecurityGroupRule `json:"items"`
}
