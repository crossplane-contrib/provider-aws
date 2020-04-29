/*
Copyright 2019 The Crossplane Authors.

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

package v1alpha3

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
)

// A ProviderSpec defines the desired state of a Provider.
type ProviderSpec struct {
	runtimev1alpha1.ProviderSpec `json:",inline"`

	// Region for managed resources created using this AWS provider.
	Region string `json:"region"`

	// UseServiceAccount indicates to use an IAM Role associated Kubernetes
	// ServiceAccount for authentication instead of a credentials Secret.
	// https://docs.aws.amazon.com/eks/latest/userguide/iam-roles-for-service-accounts.html
	//
	// If set to true, credentialsSecretRef will be ignored.
	// +optional
	UseServiceAccount *bool `json:"useServiceAccount,omitempty"`
}

// +kubebuilder:object:root=true

// A Provider configures an AWS 'provider', i.e. a connection to a particular
// AWS account using a particular AWS IAM role.
// +kubebuilder:printcolumn:name="REGION",type="string",JSONPath=".spec.region"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:printcolumn:name="SECRET-NAME",type="string",JSONPath=".spec.credentialsSecretRef.name",priority=1
// +kubebuilder:resource:scope=Cluster,categories={crossplane,provider,aws}
type Provider struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec ProviderSpec `json:"spec"`
}

// +kubebuilder:object:root=true

// ProviderList contains a list of Provider
type ProviderList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Provider `json:"items"`
}
