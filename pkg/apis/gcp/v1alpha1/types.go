/*
Copyright 2018 The Crossplane Authors.

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
	corev1alpha1 "github.com/crossplaneio/crossplane/pkg/apis/core/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ProviderSpec defines the desired state of Provider
type ProviderSpec struct {
	// Important: Run "make generate" to regenerate code after modifying this file

	// GCP ServiceAccount json secret key reference
	Secret corev1.SecretKeySelector `json:"credentialsSecretRef"`

	// GCP ProjectID (name)
	ProjectID string `json:"projectID"`

	// RequiredPermissions  - list of granted GCP permissions this provider's service account is expected to have
	RequiredPermissions []string `json:"requiredPermissions,omitempty"`
}

// ProviderStatus
type ProviderStatus struct {
	corev1alpha1.ConditionedStatus
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Provider is the Schema for the instances API
// +k8s:openapi-gen=true
// +groupName=gcp
type Provider struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ProviderSpec   `json:"spec,omitempty"`
	Status ProviderStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ProviderList contains a list of Provider
type ProviderList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Provider `json:"items"`
}

// IsValid returns true if provider is valid (in ready state)
func (p *Provider) IsValid() bool {
	return p.Status.IsReady()
}
