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

package v1beta1

import xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"

// CustomSecretParameters contains the additional fields for SecretParameters.
type CustomSecretParameters struct {
	// KMSKeyIDRef is a reference to an kms/v1alpha1.Key used
	// to set the KMSKeyID field.
	// +optional
	KMSKeyIDRef *xpv1.Reference `json:"kmsKeyIDRef,omitempty"`

	// KMSKeyIDSelector selects references to kms/v1alpha1.Key
	// used to set the KMSKeyID.
	// +optional
	KMSKeyIDSelector *xpv1.Selector `json:"kmsKeyIDSelector,omitempty"`

	// StringSecretRef points to the Kubernetes Secret whose data will be sent
	// as string to AWS. If key parameter is given, only the value of that key
	// will be used. Otherwise, all data in the Secret will be marshalled into
	// JSON and sent to AWS.
	// Either StringSecretRef or BinarySecretRef must be set, but not both.
	StringSecretRef *SecretReference `json:"stringSecretRef,omitempty"`

	// BinarySecretRef points to the Kubernetes Secret whose data will be encoded
	// as binary data to AWS. If key parameter is given, only the value of that
	// key will be used. Otherwise, all data in the Secret will be marshalled
	// into JSON and sent to AWS.
	// Either StringSecretRef or BinarySecretRef must be set, but not both.
	BinarySecretRef *SecretReference `json:"binarySecretRef,omitempty"`

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

	// A JSON-formatted string constructed according to the grammar and syntax for
	// an Amazon Web Services resource-based policy. The policy in the string identifies
	// who can access or manage this secret and its versions. For information on
	// how to format a JSON parameter for the various command line tool environments,
	// see Using JSON for Parameters (http://docs.aws.amazon.com/cli/latest/userguide/cli-using-param.html#cli-using-param-json)
	// in the CLI User Guide.
	//
	// ResourcePolicy is a required field
	// +optional
	ResourcePolicy *string `json:"resourcePolicy,omitempty"`
}

// CustomSecretObservation includes the custom status fields of Secret.
type CustomSecretObservation struct{}

// A SecretReference is a reference to a secret in an arbitrary namespace.
type SecretReference struct {
	// Name of the secret.
	Name string `json:"name"`

	// Namespace of the secret.
	Namespace string `json:"namespace"`

	// Key whose value will be used. If not given, the whole map in the Secret
	// data will be used.
	Key *string `json:"key,omitempty"`

	// Type of the secret. Used to (re)create k8s secret in case of loss.
	// If not given, the controller will try to fetch the type from the referenced secret.
	Type *string `json:"type,omitempty"`
}
