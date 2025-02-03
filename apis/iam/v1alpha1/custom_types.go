/*
Copyright 2023 The Crossplane Authors.

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

import xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"

// CustomInstanceProfileParameters includes the custom fields of InstanceProfile.
type CustomInstanceProfileParameters struct {
	// Role is the ID for the Role to add to Instance Profile.
	// +immutable
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-aws/apis/iam/v1beta1.Role
	Role *string `json:"role,omitempty"`

	// RoleRef is a reference to a Role
	// +optional
	RoleRef *xpv1.Reference `json:"roleRef,omitempty"`

	// RoleSelector selects references to Role
	// +optional
	RoleSelector *xpv1.Selector `json:"roleSelector,omitempty"`
}

// CustomInstanceProfileObservation includes the custom status fields of InstanceProfileObservation.
type CustomInstanceProfileObservation struct{}

// CustomServiceLinkedRoleParameters includes the custom fields of ServiceLinkedRole.
type CustomServiceLinkedRoleParameters struct{}

// CustomServiceLinkedRoleObservation includes the custom status fields of ServiceLinkedRole.
type CustomServiceLinkedRoleObservation struct{}
