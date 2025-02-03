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

package v1alpha1

import (
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
)

// CustomWorkspaceParameters includes custom fields about WorkspaceParameters.
type CustomWorkspaceParameters struct{}

// CustomWorkspaceObservation includes the custom status fields of Workspace.
type CustomWorkspaceObservation struct{}

// CustomRuleGroupsNamespaceParameters includes custom fields about RuleGroupsNamespaceParameters.
// workspaceID is actually required but since it's reference-able, it's not marked as required.
type CustomRuleGroupsNamespaceParameters struct {
	// workspaceID is the ID for the Workspace.
	// +immutable
	// +crossplane:generate:reference:type=Workspace
	WorkspaceID *string `json:"workspaceId,omitempty"`

	// WorkspaceIDRef is a reference to a Workspace used to set
	// the workspaceID.
	// +optional
	WorkspaceIDRef *xpv1.Reference `json:"workspaceIdRef,omitempty"`

	// WorkspaceIDSelector selects references to Workspace used
	// to set the workspaceID.
	// +optional
	WorkspaceIDSelector *xpv1.Selector `json:"workspaceIdSelector,omitempty"`
}

// CustomWorkspaceObservation includes the custom status fields of RuleGroupsNamespace.
type CustomRuleGroupsNamespaceObservation struct{}

// CustomAlertManagerDefinitionParameters includes custom fields about AlertManagerDefinitionParameters.
// workspaceID is actually required but since it's reference-able, it's not marked as required.
type CustomAlertManagerDefinitionParameters struct {
	// workspaceID is the ID for the Workspace.
	// +immutable
	// +crossplane:generate:reference:type=Workspace
	WorkspaceID *string `json:"workspaceId,omitempty"`

	// WorkspaceIDRef is a reference to a Workspace used to set
	// the workspaceID.
	// +optional
	WorkspaceIDRef *xpv1.Reference `json:"workspaceIdRef,omitempty"`

	// WorkspaceIDSelector selects references to Workspace used
	// to set the workspaceID.
	// +optional
	WorkspaceIDSelector *xpv1.Selector `json:"workspaceIdSelector,omitempty"`
}

// CustomAlertManagerDefinitionObservation includes the custom status fields of AlertManagerDefinition.
type CustomAlertManagerDefinitionObservation struct{}
