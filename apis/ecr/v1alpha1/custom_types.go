package v1alpha1

import xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"

// CustomLifecyclePolicyParameters are custom reference parameters for the LifecyclePolicy
type CustomLifecyclePolicyParameters struct {
	// RepositoryName is the name of the Repository that the policy should attach to
	// +immutable
	// +crossplane:generate:reference:type=Repository
	RepositoryName *string `json:"repositoryName,omitempty"`

	// RepositoryNameRef is the name of the Repository that the policy should attach to
	// +optional
	RepositoryNameRef *xpv1.Reference `json:"repositoryNameRef,omitempty"`

	// RepositoryNameSelector selects a references to the Repository the policy should attach to
	// +optional
	RepositoryNameSelector *xpv1.Selector `json:"repositoryNameSelector,omitempty"`
}

// CustomLifecyclePolicyObservation includes the custom status fields of LifecyclePolicy.
type CustomLifecyclePolicyObservation struct{}
