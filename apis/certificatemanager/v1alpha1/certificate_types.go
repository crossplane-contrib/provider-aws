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

package v1alpha1

import (
	"github.com/aws/aws-sdk-go-v2/service/acm"
	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Tag represents user-provided metadata that can be associated
type Tag struct {

	// The key name that can be used to look up or retrieve the associated value.
	Key string `json:"key"`

	// The value associated with this tag.
	// +optional
	Value string `json:"value,omitempty"`
}

// DomainValidationOption validate domain ownership.
type DomainValidationOption struct {
	// Additinal Fully qualified domain name (FQDN),that to secure with an ACM certificate.
	// +immutable
	DomainName string `json:"domainName"`

	// Method to validate certificate
	// +immutable
	ValidationDomain string `json:"validationDomain"`
}

// CertificateSpec defines the desired state of Certificate
type CertificateSpec struct {
	runtimev1alpha1.ResourceSpec `json:",inline"`
	ForProvider                  CertificateParameters `json:"forProvider"`
}

// CertificateExternalStatus keeps the state of external resource
type CertificateExternalStatus struct {
	// String that contains the ARN of the issued certificate. This must be of the
	CertificateArn string `json:"certificateArn"`

	// Flag to check eligibility for renewal status
	RenewalEligibility string `json:"renewalEligibility"`

	// Status of the certificate
	// +kubebuilder:validation:Enum=pending_validation;issued;inactive;expired;validation_timed_out;failed
	Status acm.CertificateStatus `json:"status"`
}

// An CertificateStatus represents the observed state of an Certificate manager.
type CertificateStatus struct {
	runtimev1alpha1.ResourceStatus `json:",inline"`
	AtProvider                     CertificateExternalStatus `json:"atProvider"`
}

// CertificateParameters defines the desired state of an AWS Certificate.
type CertificateParameters struct {

	// The Amazon Resource Name (ARN) of the private certificate authority (CA)that will be used to issue the certificate.
	// +optional
	CertificateAuthorityArn *string `json:"certificateAuthorityArn,omitempty"`

	// // CertificateAuthorityArnRef references an CertificateAuthority to retrieve its Arn
	// // +optional
	// CertificateAuthorityArnRef *runtimev1alpha1.Reference `json:"certificateAuthorityArnRef,omitempty"`

	// // CertificateAuthorityArnSelector selects a reference to an CertificateAuthority to retrieve its Arn
	// // +optional
	// CertificateAuthorityArnSelector *runtimev1alpha1.Selector `json:"certificateAuthorityArnSelector,omitempty"`

	// Fully qualified domain name (FQDN),that to secure with an ACM certificate.
	// +immutable
	DomainName string `json:"domainName"`

	// The domain name that you want ACM to use to send you emails so that you can
	// validate domain ownership.
	// +optional
	// +immutable
	DomainValidationOptions []*DomainValidationOption `json:"domainValidationOptions,omitempty"`

	// Token to distinguish between calls to RequestCertificate.
	// +optional
	IdempotencyToken *string `json:"idempotencyToken,omitempty"`

	// Parameter add the certificate to a certificate transparency log.
	// +optional
	// +kubebuilder:validation:enabled;disabled
	CertificateTransparencyLoggingPreference acm.CertificateTransparencyLoggingPreference `json:"certificateTransparencyLoggingPreference,omitempty"`

	// Subject Alternative Name extension of the ACM certificate.
	// +optional
	// +immutable
	SubjectAlternativeNames []string `json:"subjectAlternativeNames,omitempty"`

	// One or more resource tags to associate with the certificate.
	// +optional
	Tags []Tag `json:"tags,omitempty"`

	// Method to validate certificate.
	// +optional
	// +kubebuilder:validation:dns;email
	ValidationMethod acm.ValidationMethod `json:"validationMethod,omitempty"`

	// Flag to renew the certificate
	// +optional
	RenewCertificate *bool `json:"renewCertificate,omitempty"`
}

// +kubebuilder:object:root=true

// Certificate is a managed resource that represents an AWS Certificate Manager.
// +kubebuilder:printcolumn:name="DOMAINNAME",type="string",JSONPath=".spec.forProvider.domainName"
// +kubebuilder:printcolumn:name="STATUS",type="string",JSONPath=".spec.atProvider.status"
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster
type Certificate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CertificateSpec   `json:"spec,omitempty"`
	Status CertificateStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// CertificateList contains a list of Certificate
type CertificateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Certificate `json:"items"`
}
