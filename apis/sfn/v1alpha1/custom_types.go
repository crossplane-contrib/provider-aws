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
type CustomActivityParameters struct {
	// Settings to configure server-side encryption.
	EncryptionConfiguration *CustomEncryptionConfiguration `json:"encryptionConfiguration,omitempty"`
}

// CustomActivityObservation includes the custom status fields of Activity.
type CustomActivityObservation struct{}

// CustomStateMachineParameters includes custom additional fields for StateMachineParameters.
type CustomStateMachineParameters struct {
	// Settings to configure server-side encryption.
	EncryptionConfiguration *CustomEncryptionConfiguration `json:"encryptionConfiguration,omitempty"`

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

type CustomEncryptionConfiguration struct {
	// Maximum duration that Step Functions will reuse data keys. When the period
	// expires, Step Functions will call GenerateDataKey. Only applies to customer
	// managed keys. (when Type is CUSTOMER_MANAGED_KMS_KEY)
	// Should be between 60 seconds (1 minute) and 900 seconds (15 minutes).
	// Default is 300 seconds (5 minutes).
	KMSDataKeyReusePeriodSeconds *int64 `json:"kmsDataKeyReusePeriodSeconds,omitempty"`

	// An alias, alias ARN, key ID, or key ARN of a symmetric encryption KMS key
	// to encrypt data. To specify a KMS key in a different Amazon Web Services
	// account, you must use the key ARN or alias ARN.
	// Only needed if Type is CUSTOMER_MANAGED_KMS_KEY.
	KMSKeyID *string `json:"kmsKeyID,omitempty"`

	// Encryption type.
	// Valide values are AWS_OWNED_KEY and CUSTOMER_MANAGED_KMS_KEY
	//
	// Type is a required field
	// +kubebuilder:validation:Required
	Type string `json:"type"`
}
