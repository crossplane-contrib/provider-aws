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
	"github.com/aws/aws-sdk-go-v2/service/acmpca"
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

// CertificateAuthoritySpec defines the desired state of CertificateAuthority
type CertificateAuthoritySpec struct {
	runtimev1alpha1.ResourceSpec `json:",inline"`
	ForProvider                  CertificateAuthorityParameters `json:"forProvider"`
}

// CertificateAuthorityExternalStatus keeps the state of external resource
type CertificateAuthorityExternalStatus struct {
	// String that contains the ARN of the issued certificate Authority
	CertificateAuthorityArn string `json:"certificateAuthorityArn"`
	RenewalPermission       bool   `json:"renewalPermission"`
}

// An CertificateAuthorityStatus represents the observed state of an CertificateAuthority manager.
type CertificateAuthorityStatus struct {
	runtimev1alpha1.ResourceStatus `json:",inline"`
	AtProvider                     CertificateAuthorityExternalStatus `json:"atProvider"`
}

// CertificateAuthorityParameters defines the desired state of an AWS CertificateAuthority.
type CertificateAuthorityParameters struct {
	// Type of the certificate authority
	// +kubebuilder:validation:Enum=root;subordinate
	Type acmpca.CertificateAuthorityType `json:"type"`

	// Status of the certificate authority
	// +optional
	// +kubebuilder:validation:Enum=creating;pending_certificate;active;deleted;disabled;expired;failed
	Status acmpca.CertificateAuthorityStatus `json:"status"`

	// Token to distinguish between calls to RequestCertificate.
	// +optional
	IdempotencyToken *string `json:"idempotencyToken,omitempty"`

	// Organization legal name
	// +immutable
	Organization *string `json:"organization"`

	// Organization's subdivision or unit
	// +immutable
	OrganizationalUnit *string `json:"organizationalUnit"`

	// Two-digit code that specifies the country
	// +immutable
	Country *string `json:"country"`

	// State in which the subject of the certificate is located
	// +immutable
	State *string `json:"state"`

	// The locality such as a city or town
	// +immutable
	Locality *string `json:"locality"`

	// FQDN associated with the certificate subject
	// +immutable
	CommonName *string `json:"commonName"`

	// Type of the public key algorithm
	// +kubebuilder:validation:Enum=rsa2048;rsa4096;ecprime256v1;ecsecp384r1
	KeyAlgorithm acmpca.KeyAlgorithm `json:"keyAlgorithm"`

	// Algorithm that private CA uses to sign certificate requests
	// +kubebuilder:validation:Enum=sha256withecdsa;sha384withecdsa;sha512withecdsa;sha256withrsa;sha384withrsa;sha512withrsa
	SigningAlgorithm acmpca.SigningAlgorithm `json:"signingAlgorithm"`

	// Boolean value that specifies certificate revocation
	RevocationConfigurationEnabled *bool `json:"revocationConfigurationEnabled"`

	// Name of the S3 bucket that contains the CRL
	S3BucketName *string `json:"s3BucketName"`

	// Alias for the CRL distribution point
	// +optional
	CustomCname *string `json:"customCname,omitempty"`

	// Number of days until a certificate expires
	ExpirationInDays *int64 `json:"expirationInDays,omitempty"`

	// The number of days to make a CA restorable after it has been deleted
	// +optional
	PermanentDeletionTimeInDays *int64 `json:"permanentDeletionTimeInDays,omitempty"`

	// The CertificateRenewalPermissionAllow decides Permissions for ACM renewals
	CertificateRenewalPermissionAllow bool `json:"certificateRenewalPermissionAllow"`

	// Disambiguating information for the certificate subject.
	// +optional
	// +immutable
	DistinguishedNameQualifier *string `json:"distinguishedNameQualifier,omitempty"`

	// Typically a qualifier appended to the name of an individual
	// +optional
	// +immutable
	GenerationQualifier *string `json:"generationQualifier,omitempty"`

	// Concatenation of first letter of the GivenName, Middle name and SurName.
	// +optional
	// +immutable
	Initials *string `json:"initials,omitempty"`

	// First name
	// +optional
	// +immutable
	GivenName *string `json:"givenName,omitempty"`

	// Shortened version of a longer GivenName
	// +optional
	// +immutable
	Pseudonym *string `json:"pseudonym,omitempty"`

	// The certificate serial number.
	// +optional
	// +immutable
	SerialNumber *string `json:"serialNumber,omitempty"`

	// Surname
	// +optional
	// +immutable
	Surname *string `json:"surname,omitempty"`

	// Title
	// +optional
	// +immutable
	Title *string `json:"title,omitempty"`

	// One or more resource tags to associate with the certificateAuthority.
	Tags []Tag `json:"tags,omitempty"`
}

// +kubebuilder:object:root=true

// CertificateAuthority is a managed resource that represents an AWS CertificateAuthority Manager.
// +kubebuilder:printcolumn:name="TYPE",type="string",JSONPath=".spec.forProvider.type"
// +kubebuilder:printcolumn:name="STATUS",type="string",JSONPath=".spec.forProvider.status"
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster
type CertificateAuthority struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CertificateAuthoritySpec   `json:"spec,omitempty"`
	Status CertificateAuthorityStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// CertificateAuthorityList contains a list of CertificateAuthority
type CertificateAuthorityList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CertificateAuthority `json:"items"`
}
