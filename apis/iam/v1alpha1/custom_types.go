package v1alpha1

import xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"

// CustomInstanceProfileParameters includes the custom fields of InstanceProfile.
type CustomInstanceProfileParameters struct {
	// Role is the ID for the Role to add to Instance Profile.
	// +immutable
	// +crossplane:generate:reference:type=github.com/crossplane/provider-aws/apis/iam/v1beta1.Role
	Role *string `json:"role,omitempty"`

	// RoleRef is a reference to an Role
	// +optional
	RoleRef *xpv1.Reference `json:"roleRef,omitempty"`

	// RoleSelector selects references to Role
	// +optional
	RoleSelector *xpv1.Selector `json:"roleSelector,omitempty"`
}
