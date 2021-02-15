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
	// SecretRef is used to reference the secret and the key whose value you'd
	// like to be stored in AWS.
	// The value is stored as is with no change.
	SecretRef xpv1.SecretKeySelector `json:"secretRef"`
}
