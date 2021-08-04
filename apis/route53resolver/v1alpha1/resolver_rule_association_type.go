package v1alpha1

import (
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:root=true

// A ResolverRuleAssociation is a managed resource that represents the association between a Route53 Resolver Rule and VPC
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="STATUS",type="string",JSONPath=".status.atProvider.status"
// +kubebuilder:printcolumn:name="ID",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="VPC",type="string",JSONPath=".spec.forProvider.vpcId"
// +kubebuilder:printcolumn:name="Resolver Rule ID",type="string",JSONPath=".spec.forProvider.resolverRuleId"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,aws}
type ResolverRuleAssociation struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ResolverRuleAssociationSpec   `json:"spec"`
	Status ResolverRuleAssociationStatus `json:"status,omitempty"`
}

// ResolverRuleAssociationSpec defines the desired state of a ResolverRuleAssociation
type ResolverRuleAssociationSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       ResolverRuleAssociationParameters `json:"forProvider"`
}

// ResolverRuleAssociationParameters defines the desired state of a ResolverRuleAssociation
type ResolverRuleAssociationParameters struct {

	// Region is the region you'd like your VPC to be created in.
	Region string `json:"region"`

	// The ID of the Resolver rule that you want to associate with the VPC.
	ResolverRuleID *string `json:"resolverRuleId,omitempty"`

	// VPCIDRef references a Resolver Rule to retrieve its resolverRuleId
	ResolverRuleIDRef *xpv1.Reference `json:"resolverRuleIdRef,omitempty"`

	// ResolverRuleIDSelector selects a reference to a Resolver Rule to retrieve its resolverRuleId
	ResolverRuleIDSelector *xpv1.Selector `json:"resolverRuleIdSelector,omitempty"`

	// The ID of the VPC that you want to associate the Resolver rule with.
	VPCID *string `json:"vpcId,omitempty"`

	// VPCIDRef references a VPC to retrieve its vpcId
	VPCIDRef *xpv1.Reference `json:"vpcIdRef,omitempty"`

	// VPCIDSelector selects a reference to a VPC to retrieve its vpcId
	VPCIDSelector *xpv1.Selector `json:"vpcIdSelector,omitempty"`
}

// ResolverRuleAssociationStatus represents the observed state of a ResolverRuleAssociation
type ResolverRuleAssociationStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          ResolverRuleAssociationObservation `json:"atProvider"`
}

// ResolverRuleAssociationObservation keeps the state for the external resource
type ResolverRuleAssociationObservation struct {
	// The ID of the Resolver Rule Association
	ID string `json:"id,omitempty"`

	// The ID of the VPC that you want to associate the Resolver rule with.
	VPCID string `json:"vpcId,omitempty"`

	// The ID of the Resolver rule that you want to associate with the VPC.
	RuleID string `json:"ruleId,omitempty"`

	// The status of the Resolver rule association.
	Status string `json:"status,omitempty"`

	// A detailed description of the status of the Resolver rule association.
	StatusMessage string `json:"statusMessage,omitempty"`
}

// +kubebuilder:object:root=true

// ResolverRuleAssociationList contains a list of ResolverRuleAssociations
type ResolverRuleAssociationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ResolverRuleAssociation `json:"items"`
}
