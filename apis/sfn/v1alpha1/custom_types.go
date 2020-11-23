/*
Copyright 2020 The Crossplane Authors.

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

import runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"

// CustomActivityParameters includes custom additional fields for ActivityParameters.
type CustomActivityParameters struct{}

// CustomStateMachineParameters includes custom additional fields for StateMachineParameters.
type CustomStateMachineParameters struct {
	// RoleARN is the ARN for the IAMRole.
	// +immutable
	RoleARN *string `json:"roleArn,omitempty"`

	// RoleARNRef is a reference to an IAMRole used to set
	// the RoleARN.
	// +optional
	RoleARNRef *runtimev1alpha1.Reference `json:"roleArnRef,omitempty"`

	// RoleARNSelector selects references to IAMRole used
	// to set the RoleARN.
	// +optional
	RoleARNSelector *runtimev1alpha1.Selector `json:"roleArnSelector,omitempty"`
}
