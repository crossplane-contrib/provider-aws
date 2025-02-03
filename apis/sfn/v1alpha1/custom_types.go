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

import xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"

// CustomActivityParameters includes custom additional fields for ActivityParameters.
type CustomActivityParameters struct{}

// CustomActivityObservation includes the custom status fields of Activity.
type CustomActivityObservation struct{}

// CustomStateMachineParameters includes custom additional fields for StateMachineParameters.
type CustomStateMachineParameters struct {
	// RoleARN is the ARN for the IAMRole.
	// It has to be given directly or resolved using RoleARNRef or RoleARNSelector.
	// +immutable
	// +optional
	RoleARN *string `json:"roleArn,omitempty"`

	// RoleARNRef is a reference to an IAMRole used to set
	// the RoleARN.
	// +optional
	RoleARNRef *xpv1.Reference `json:"roleArnRef,omitempty"`

	// RoleARNSelector selects references to IAMRole used
	// to set the RoleARN.
	// +optional
	RoleARNSelector *xpv1.Selector `json:"roleArnSelector,omitempty"`

	// Determines whether a Standard or Express state machine is created.
	// You cannot update the type of a state machine once it has been created.
	// The default is STANDARD. Possible values: STANDARD, EXPRESS
	// +immutable
	// +kubebuilder:validation:Enum=STANDARD;EXPRESS
	Type StateMachineType `json:"type,omitempty"`
}

// CustomStateMachineObservation includes the custom status fields of StateMachine.
type CustomStateMachineObservation struct{}
