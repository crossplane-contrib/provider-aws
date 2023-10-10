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
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Types of ResolverRuleAssociation status.
const (
	ResolverRuleAssociationStatusCreating   = "CREATING"
	ResolverRuleAssociationStatusComplete   = "COMPLETE"
	ResolverRuleAssociationStatusDeleting   = "DELETING"
	ResolverRuleAssociationStatusFailed     = "FAILED"
	ResolverRuleAssociationStatusOverridden = "OVERRIDDEN"
)

// +kubebuilder:object:root=true

// ResolverRuleAssociation is a managed resource that represents an AWS Route53 ResolverRuleAssociation.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,aws}
type ResolverRuleAssociation struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ResolverRuleAssociationSpec   `json:"spec"`
	Status ResolverRuleAssociationStatus `json:"status,omitempty"`
}

// ResolverRuleAssociationSpec defines the desired state of an AWS Route53 Hosted ResolverRuleAssociation.
type ResolverRuleAssociationSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       ResolverRuleAssociationParameters `json:"forProvider"`
}

// ResolverRuleAssociationStatus represents the observed state of a ResolverRuleAssociation.
type ResolverRuleAssociationStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          ResolverRuleAssociationObservation `json:"atProvider,omitempty"`
}

// ResolverRuleAssociationParameters define the desired state of an AWS Route53 Hosted ResolverRuleAssociation.
type ResolverRuleAssociationParameters struct {
	// Region is which region the Addon will be created.
	// +kubebuilder:validation:Required
	Region string `json:"region"`

	// The ID of the Resolver rule that you want to associate with the VPC. To list
	// the existing Resolver rules, use ListResolverRules (https://docs.aws.amazon.com/Route53/latest/APIReference/API_route53resolver_ListResolverRules.html).
	//
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-aws/apis/route53resolver/v1alpha1.ResolverRule
	// +optional
	ResolverRuleID *string `json:"resolverRuleId,omitempty"`

	// ResolverRuleIDRef is a reference to a ResolverRule used to set
	// the ResolverRuleID.
	// +immutable
	// +optional
	ResolverRuleIDRef *xpv1.Reference `json:"resolverRuleIdRef,omitempty"`

	// ResolverRuleIDSelector selects references to a ResolverRule used
	// to set the ResolverRuleID.
	// +immutable
	// +optional
	ResolverRuleIDSelector *xpv1.Selector `json:"resolverRuleIdSelector,omitempty"`

	// The ID of the VPC that you want to associate the Resolver rule with.
	//
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-aws/apis/ec2/v1beta1.VPC
	// +optional
	VPCId *string `json:"vpcId,omitempty"`

	// VPCIdRef is a reference to a VPC used to set
	// the VPCId.
	// +immutable
	// +optional
	VPCIdRef *xpv1.Reference `json:"vpcIdRef,omitempty"`

	// VPCIdSelector selects references to a VPC used
	// to set the VPCId.
	// +immutable
	// +optional
	VPCIdSelector *xpv1.Selector `json:"vpcIdSelector,omitempty"`
}

// ResolverRuleAssociationObservation keeps the state for the external resource.
type ResolverRuleAssociationObservation struct{}

// +kubebuilder:object:root=true

// ResolverRuleAssociationList contains a list of ResolverRuleAssociation.
type ResolverRuleAssociationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []ResolverRuleAssociation `json:"items"`
}
