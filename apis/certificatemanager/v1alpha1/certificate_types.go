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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
)

// Tag represents user-provided metadata that can be associated
// One or more resource tags to associate with the certificate.
// IAM to manage permissions, see Controlling Access Using IAM Tags (https://docs.aws.amazon.com/IAM/latest/UserGuide/access_iam-tags.html).
type Tag struct {

	// The key name that can be used to look up or retrieve the associated value.
	// For example, Department or Cost Center are common choices.
	Key string `json:"key"`

	// The value associated with this tag. For example, tags with a key name of
	// Department could have values such as Human Resources, Accounting, and Support.
	// Tags with a key name of Cost Center might have values that consist of the
	// number associated with the different cost centers in your company. Typically,
	// many resources have tags with the same key name but with different values.
	//
	// AWS always interprets the tag Value as a single string. If you need to store
	// an array, you can store comma-separated values in the string. However, you
	// must interpret the value in your code.
	// +optional
	Value string `json:"value,omitempty"`
}

// CertificateOptions to specify whether to add the certificate to a certificate transparency log
type CertificateOptions struct {

	// You can opt out of certificate transparency logging by specifying the DISABLED
	// option. Opt in by specifying ENABLED.
	CertificateTransparencyLoggingPreference string `json:"CertificateTransparencyLoggingPreference"`
	// contains filtered or unexported fields
}

// DomainValidationOption validate domain ownership.
type DomainValidationOption struct {
	// A fully qualified domain name (FQDN) in the certificate request.
	//
	// DomainName is a required field
	DomainName string `json:"DomainName"`

	// The domain name that you want ACM to use to send you validation emails. This
	// domain name is the suffix of the email addresses that you want ACM to use.
	// This must be the same as the DomainName value or a superdomain of the DomainName
	// value. For example, if you request a certificate for testing.example.com,
	// you can specify example.com for this value. In that case, ACM sends domain
	// validation emails to the following five addresses:
	//
	//    * admin@example.com
	//
	//    * administrator@example.com
	//
	//    * hostmaster@example.com
	//
	//    * postmaster@example.com
	//
	//    * webmaster@example.com
	//
	// ValidationDomain is a required field
	ValidationDomain string `json:"ValidationDomain"`
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// CertificateSpec defines the desired state of Certificate
type CertificateSpec struct {
	runtimev1alpha1.ResourceSpec `json:",inline"`
	ForProvider                  CertificateParameters `json:"forProvider"`
}

// CertificateExternalStatus keeps the state of external resource
type CertificateExternalStatus struct {
	// String that contains the ARN of the issued certificate. This must be of the
	// form:
	CertificateArn string `json:"CertificateArn"`
}

// An CertificateStatus represents the observed state of an Certificate manager.
type CertificateStatus struct {
	runtimev1alpha1.ResourceStatus `json:",inline"`
	AtProvider                     CertificateExternalStatus `json:"atProvider"`
}

// CertificateParameters defines the desired state of an AWS Certificate.
type CertificateParameters struct {

	// The Amazon Resource Name (ARN) of the private certificate authority (CA)
	// that will be used to issue the certificate. If you do not provide an ARN
	// and you are trying to request a private certificate, ACM will attempt to
	// issue a public certificate. For more information about private CAs, see the
	// AWS Certificate Manager Private Certificate Authority (PCA) (https://docs.aws.amazon.com/acm-pca/latest/userguide/PcaWelcome.html)
	// user guide. The ARN must have the following form:
	// +optional
	CertificateAuthorityArn string `json:"CertificateAuthorityArn,omitempty"`

	// Fully qualified domain name (FQDN), such as www.example.com, that you want
	// to secure with an ACM certificate. Use an asterisk (*) to create a wildcard
	// certificate that protects several sites in the same domain. For example,
	// *.example.com protects www.example.com, site.example.com, and images.example.com.
	//
	// The first domain name you enter cannot exceed 64 octets, including periods.
	// Each subsequent Subject Alternative Name (SAN), however, can be up to 253
	// octets in length.
	DomainName string `json:"DomainName"`

	// The domain name that you want ACM to use to send you emails so that you can
	// validate domain ownership.
	// +optional
	DomainValidationOptions []DomainValidationOption `json:"DomainValidationOptions,omitempty"`

	// Customer chosen string that can be used to distinguish between calls to RequestCertificate.
	// Idempotency tokens time out after one hour. Therefore, if you call RequestCertificate
	// multiple times with the same idempotency token within one hour, ACM recognizes
	// that you are requesting only one certificate and will issue only one. If
	// you change the idempotency token for each call, ACM recognizes that you are
	// requesting multiple certificates.
	// +optional
	IdempotencyToken string `json:"IdempotencyToken,omitempty"`

	// Currently, you can use this parameter to specify whether to add the certificate
	// to a certificate transparency log. Certificate transparency makes it possible
	// to detect SSL/TLS certificates that have been mistakenly or maliciously issued.
	// Certificates that have not been logged typically produce an error message
	// in a browser. For more information, see Opting Out of Certificate Transparency
	// Logging (https://docs.aws.amazon.com/acm/latest/userguide/acm-bestpractices.html#best-practices-transparency).
	// +optional
	Options CertificateOptions `json:"Options,omitempty"`

	// Additional FQDNs to be included in the Subject Alternative Name extension
	// of the ACM certificate. For example, add the name www.example.net to a certificate
	// for which the DomainName field is www.example.com if users can reach your
	// site by using either name. The maximum number of domain names that you can
	// add to an ACM certificate is 100. However, the initial quota is 10 domain
	// names. If you need more than 10 names, you must request a quota increase.
	// For more information, see Quotas (https://docs.aws.amazon.com/acm/latest/userguide/acm-limits.html).
	//
	// The maximum length of a SAN DNS name is 253 octets. The name is made up of
	// multiple labels separated by periods. No label can be longer than 63 octets.
	// Consider the following examples:
	//
	//    * (63 octets).(63 octets).(63 octets).(61 octets) is legal because the
	//    total length is 253 octets (63+1+63+1+63+1+61) and no label exceeds 63
	//    octets.
	//
	//    * (64 octets).(63 octets).(63 octets).(61 octets) is not legal because
	//    the total length exceeds 253 octets (64+1+63+1+63+1+61) and the first
	//    label exceeds 63 octets.
	//
	//    * (63 octets).(63 octets).(63 octets).(62 octets) is not legal because
	//    the total length of the DNS name (63+1+63+1+63+1+62) exceeds 253 octets.
	// +optional
	SubjectAlternativeNames []string `json:"SubjectAlternativeNames,omitempty"`

	// One or more resource tags to associate with the certificate.
	// +optional
	Tags []Tag `json:"tags,omitempty"`

	// The method you want to use if you are requesting a public certificate to
	// validate that you own or control domain. You can validate with DNS (https://docs.aws.amazon.com/acm/latest/userguide/gs-acm-validate-dns.html)
	// or validate with email (https://docs.aws.amazon.com/acm/latest/userguide/gs-acm-validate-email.html).
	// We recommend that you use DNS validation.
	// +optional
	ValidationMethod string `json:"ValidationMethod,omitempty"`
}

// +kubebuilder:object:root=true

// Certificate is a managed resource that represents an AWS Certificate Manager.
// +kubebuilder:printcolumn:name="DOMAINNAME",type="string",JSONPath=".spec.domainName"
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
