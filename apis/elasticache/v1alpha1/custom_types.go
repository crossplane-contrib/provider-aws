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

// CustomCacheParameterGroupParameters includes the custom fields.
type CustomCacheParameterGroupParameters struct {
	// A list of parameters to associate with this DB parameter group
	// +optional
	ParameterNameValues []ParameterNameValue `json:"parameters,omitempty"`
}

// CustomUserParameters contains the addtionals fields for UserParameters
type CustomUserParameters struct {
	PasswordSecretRef []xpv1.SecretKeySelector `json:"passwordSecretRef,omitempty"`
}
