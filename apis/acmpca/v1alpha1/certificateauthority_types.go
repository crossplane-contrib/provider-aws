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
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CertificateAuthorityParameters defines the desired state of an AWS CertificateAuthority.
type CertificateAuthorityParameters struct {
	// Region is the region you'd like your CertificateAuthority to be created in.
	Region string `json:"region"`

	// Type of the certificate authority
	// +kubebuilder:validation:Enum=ROOT;SUBORDINATE
	Type string `json:"type"`

	// RevocationConfiguration to associate with the certificateAuthority.
	// +optional
	RevocationConfiguration *RevocationConfiguration `json:"revocationConfiguration,omitempty"`

	// CertificateAuthorityConfiguration to associate with the certificateAuthority.
	CertificateAuthorityConfiguration CertificateAuthorityConfiguration `json:"certificateAuthorityConfiguration"`

	// The number of days to make a CA restorable after it has been deleted
	// +optional
	PermanentDeletionTimeInDays *int32 `json:"permanentDeletionTimeInDays,omitempty"`

	// Status of the certificate authority.
	// This value cannot be configured at creation, but can be updated to set a
	// CA to ACTIVE or DISABLED.
	// +optional
	// +kubebuilder:validation:Enum=ACTIVE;DISABLED
	Status *string `json:"status,omitempty"`

	// One or more resource tags to associate with the certificateAuthority.
	Tags []Tag `json:"tags"`
}

// Tag represents user-provided metadata that can be associated
type Tag struct {

	// The key name that can be used to look up or retrieve the associated value.
	Key string `json:"key"`

	// The value associated with this tag.
	Value string `json:"value"`
}

// RevocationConfiguration is configuration of the certificate revocation list
type RevocationConfiguration struct {

	// Boolean value that specifies certificate revocation
	Enabled bool `json:"enabled"`

	// Name of the S3 bucket that contains the CRL
	// +optional
	S3BucketName *string `json:"s3BucketName,omitempty"`

	// Alias for the CRL distribution point
	// +optional
	CustomCname *string `json:"customCname,omitempty"`

	// Number of days until a certificate expires
	// +optional
	ExpirationInDays *int32 `json:"expirationInDays,omitempty"`
}

// CertificateAuthorityConfiguration is
type CertificateAuthorityConfiguration struct {

	// Type of the public key algorithm
	// +kubebuilder:validation:Enum=RSA_2048;EC_secp384r1;EC_prime256v1;RSA_4096
	KeyAlgorithm string `json:"keyAlgorithm"`

	// Algorithm that private CA uses to sign certificate requests
	// +kubebuilder:validation:Enum=SHA512WITHECDSA;SHA256WITHECDSA;SHA384WITHECDSA;SHA512WITHRSA;SHA256WITHRSA;SHA384WITHRSA
	SigningAlgorithm string `json:"signingAlgorithm"`

	// Subject is information of Certificate Authority
	Subject Subject `json:"subject"`
}

// Subject is
type Subject struct {

	// Organization legal name
	// +immutable
	Organization string `json:"organization"`

	// Organization's subdivision or unit
	// +immutable
	OrganizationalUnit string `json:"organizationalUnit"`

	// Two-digit code that specifies the country
	// +immutable
	Country string `json:"country"`

	// State in which the subject of the certificate is located
	// +immutable
	State string `json:"state"`

	// The locality such as a city or town
	// +immutable
	Locality string `json:"locality"`

	// FQDN associated with the certificate subject
	// +immutable
	CommonName string `json:"commonName"`

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
}

// CertificateAuthorityExternalStatus keeps the state of external resource
type CertificateAuthorityExternalStatus struct {
	// String that contains the ARN of the issued certificate Authority
	CertificateAuthorityARN string `json:"certificateAuthorityARN,omitempty"`

	// Serial of the Certificate Authority
	Serial string `json:"serial,omitempty"`

	// Status is the current status of the CertificateAuthority.
	Status string `json:"status,omitempty"`
}

// CertificateAuthoritySpec defines the desired state of CertificateAuthority
type CertificateAuthoritySpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       CertificateAuthorityParameters `json:"forProvider"`
}

// An CertificateAuthorityStatus represents the observed state of an CertificateAuthority manager.
type CertificateAuthorityStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          CertificateAuthorityExternalStatus `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// CertificateAuthority is a managed resource that represents an AWS CertificateAuthority Manager.
// +kubebuilder:printcolumn:name="TYPE",type="string",JSONPath=".spec.forProvider.type"
// +kubebuilder:printcolumn:name="STATUS",type="string",JSONPath=".spec.forProvider.status"
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,aws}
// +kubebuilder:deprecatedversion:warning="Please use v1beta1 version of this resource that has identical schema."
// Deprecated: Please use v1beta1 version of this resource.
type CertificateAuthority struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CertificateAuthoritySpec   `json:"spec"`
	Status CertificateAuthorityStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// CertificateAuthorityList contains a list of CertificateAuthority
// +kubebuilder:deprecatedversion:warning="Please use v1beta1 version of this resource that has identical schema."
// Deprecated: Please use v1beta1 version of this resource.
type CertificateAuthorityList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CertificateAuthority `json:"items"`
}
