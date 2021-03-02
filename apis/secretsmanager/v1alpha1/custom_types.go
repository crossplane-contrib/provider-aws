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

package v1alpha1

import xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"

// CustomSecretParameters contains the additional fields for SecretParameters.
type CustomSecretParameters struct {
	// PayloadType lets you choose which part of the input secret should be used.
	// SingleValue: Only value of the given key.
	// Map: The whole content of the input secret will be used after converting
	// it to JSON map.
	// +kubebuilder:validation:Enum=SingleValue;Map
	PayloadType string `json:"payloadType"`

	// SingleValueSecretRef should be filled if SingleValue is chosen as payload
	// type. It references the input secret and the key whose value you would
	// like to use.
	SingleValueSecretRef *xpv1.SecretKeySelector `json:"singleValueSecretRef,omitempty"`

	// MapSecretRef should be filled if Map is chosen as payload type. It references
	// the input secret whose data will be sent to AWS as JSON string map.
	MapSecretRef *xpv1.SecretReference `json:"mapSecretRef,omitempty"`

	// (Optional) Specifies that the secret is to be deleted without any recovery
	// window. You can't use both this parameter and the RecoveryWindowInDays parameter
	// in the same API call.
	//
	// An asynchronous background process performs the actual deletion, so there
	// can be a short delay before the operation completes. If you write code to
	// delete and then immediately recreate a secret with the same name, ensure
	// that your code includes appropriate back off and retry logic.
	//
	// Use this parameter with caution. This parameter causes the operation to skip
	// the normal waiting period before the permanent deletion that AWS would normally
	// impose with the RecoveryWindowInDays parameter. If you delete a secret with
	// the ForceDeleteWithouRecovery parameter, then you have no opportunity to
	// recover the secret. It is permanently lost.
	ForceDeleteWithoutRecovery *bool `json:"forceDeleteWithoutRecovery,omitempty"`

	// (Optional) Specifies the number of days that Secrets Manager waits before
	// it can delete the secret. You can't use both this parameter and the ForceDeleteWithoutRecovery
	// parameter in the same API call.
	//
	// This value can range from 7 to 30 days. The default value is 30.
	RecoveryWindowInDays *int64 `json:"recoveryWindowInDays,omitempty"`
}
