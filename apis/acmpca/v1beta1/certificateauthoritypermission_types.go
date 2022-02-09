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

package v1beta1

import (
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CertificateAuthorityPermissionSpec defines the desired state of CertificateAuthorityPermission
type CertificateAuthorityPermissionSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       CertificateAuthorityPermissionParameters `json:"forProvider"`
}

// An CertificateAuthorityPermissionStatus represents the observed state of an Certificate Authority Permission manager.
type CertificateAuthorityPermissionStatus struct {
	xpv1.ResourceStatus `json:",inline"`
}

// CertificateAuthorityPermissionParameters defines the desired state of an AWS CertificateAuthority.
type CertificateAuthorityPermissionParameters struct {

	// Region is the region of CertificateAuthorityPermission.
	Region string `json:"region"`

	// The Amazon Resource Name (ARN) of the private certificate authority (CA)that will be used to issue the certificate.
	// +immutable
	// +crossplane:generate:reference:type=CertificateAuthority
	CertificateAuthorityARN *string `json:"certificateAuthorityARN,omitempty"`

	// CertificateAuthorityARNRef references an CertificateAuthority to retrieve its Arn
	// +optional
	// +immutable
	CertificateAuthorityARNRef *xpv1.Reference `json:"certificateAuthorityARNRef,omitempty"`

	// CertificateAuthorityARNSelector selects a reference to an CertificateAuthority to retrieve its Arn
	// +optional
	// +immutable
	CertificateAuthorityARNSelector *xpv1.Selector `json:"certificateAuthorityARNSelector,omitempty"`

	// The actions that the specified AWS service principal can use.
	// +optional
	// +immutable
	Actions []string `json:"actions,omitempty"`

	// The AWS service or identity that receives the permission. At this
	// time, the only valid principal is acm.amazonaws.com.
	// +immutable
	// +kubebuilder:default:=acm.amazonaws.com
	Principal string `json:"principal"`

	// Calling Account ID
	// +optional
	// +immutable
	SourceAccount *string `json:"sourceAccount,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:storageversion

// CertificateAuthorityPermission is a managed resource that represents an AWS CertificateAuthorityPermission Manager.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,aws}
type CertificateAuthorityPermission struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CertificateAuthorityPermissionSpec   `json:"spec"`
	Status CertificateAuthorityPermissionStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// CertificateAuthorityPermissionList contains a list of CertificateAuthorityPermission
type CertificateAuthorityPermissionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CertificateAuthorityPermission `json:"items"`
}
