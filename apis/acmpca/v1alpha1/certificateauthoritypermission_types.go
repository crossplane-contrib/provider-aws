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

import (
	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CertificateAuthorityPermissionSpec defines the desired state of CertificateAuthorityPermission
type CertificateAuthorityPermissionSpec struct {
	runtimev1alpha1.ResourceSpec `json:",inline"`
	ForProvider                  CertificateAuthorityPermissionParameters `json:"forProvider"`
}

// An CertificateAuthorityPermissionStatus represents the observed state of an Certificate Authority Permission manager.
type CertificateAuthorityPermissionStatus struct {
	runtimev1alpha1.ResourceStatus `json:",inline"`
}

// CertificateAuthorityPermissionParameters defines the desired state of an AWS CertificateAuthority.
type CertificateAuthorityPermissionParameters struct {

	// The Amazon Resource Name (ARN) of the private certificate authority (CA)that will be used to issue the certificate.
	// +immutable
	CertificateAuthorityARN *string `json:"certificateAuthorityARN,omitempty"`

	// CertificateAuthorityARNRef references an CertificateAuthority to retrieve its Arn
	// +optional
	// +immutable
	CertificateAuthorityARNRef *runtimev1alpha1.Reference `json:"certificateAuthorityARNRef,omitempty"`

	// CertificateAuthorityARNSelector selects a reference to an CertificateAuthority to retrieve its Arn
	// +optional
	// +immutable
	CertificateAuthorityARNSelector *runtimev1alpha1.Selector `json:"certificateAuthorityARNSelector,omitempty"`

	// The actions that the specified AWS service principal can use.
	// +optional
	// +immutable
	Actions []string `json:"actions,omitempty"`

	// The AWS Service or identity
	// +optional
	// +immutable
	Principal *string `json:"principal,omitempty"`

	// Calling Account ID
	// +optional
	// +immutable
	SourceAccount *string `json:"sourceAccount,omitempty"`
}

// +kubebuilder:object:root=true

// CertificateAuthorityPermission is a managed resource that represents an AWS CertificateAuthorityPermission Manager.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,aws}
type CertificateAuthorityPermission struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CertificateAuthorityPermissionSpec   `json:"spec,omitempty"`
	Status CertificateAuthorityPermissionStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// CertificateAuthorityPermissionList contains a list of CertificateAuthorityPermission
type CertificateAuthorityPermissionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CertificateAuthorityPermission `json:"items"`
}
