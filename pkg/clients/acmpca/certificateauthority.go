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

package acmpca

import (
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/awserr"
	"github.com/aws/aws-sdk-go-v2/service/acmpca"

	"github.com/crossplane/provider-aws/apis/acmpca/v1alpha1"
)

// Client defines the CertificateManager operations
type Client interface {
	CreateCertificateAuthorityRequest(*acmpca.CreateCertificateAuthorityInput) acmpca.CreateCertificateAuthorityRequest
	DeleteCertificateAuthorityRequest(*acmpca.DeleteCertificateAuthorityInput) acmpca.DeleteCertificateAuthorityRequest
	UpdateCertificateAuthorityRequest(*acmpca.UpdateCertificateAuthorityInput) acmpca.UpdateCertificateAuthorityRequest
	DescribeCertificateAuthorityRequest(*acmpca.DescribeCertificateAuthorityInput) acmpca.DescribeCertificateAuthorityRequest
	ListTagsRequest(*acmpca.ListTagsInput) acmpca.ListTagsRequest
	UntagCertificateAuthorityRequest(*acmpca.UntagCertificateAuthorityInput) acmpca.UntagCertificateAuthorityRequest
	TagCertificateAuthorityRequest(*acmpca.TagCertificateAuthorityInput) acmpca.TagCertificateAuthorityRequest
}

// NewClient returns a new client using AWS credentials as JSON encoded data.
func NewClient(conf *aws.Config) (Client, error) {
	return acmpca.New(*conf), nil
}

// GenerateCreateCertificateAuthorityInput from certificateAuthorityParameters
func GenerateCreateCertificateAuthorityInput(p *v1alpha1.CertificateAuthorityParameters) *acmpca.CreateCertificateAuthorityInput {
	m := &acmpca.CreateCertificateAuthorityInput{

		CertificateAuthorityType:          p.Type,
		CertificateAuthorityConfiguration: GenerateCertificateAuthorityConfiguration(p.CertificateAuthorityConfiguration),
		RevocationConfiguration:           GenerateRevocationConfiguration(p.RevocationConfiguration),
	}

	m.Tags = make([]acmpca.Tag, len(p.Tags))
	for i, val := range p.Tags {
		m.Tags[i] = acmpca.Tag{
			Key:   aws.String(val.Key),
			Value: aws.String(val.Value),
		}
	}

	return m
}

// GenerateCertificateAuthorityConfiguration from CertificateAuthorityConfiguration
func GenerateCertificateAuthorityConfiguration(p v1alpha1.CertificateAuthorityConfiguration) *acmpca.CertificateAuthorityConfiguration { // nolint:gocyclo

	m := &acmpca.CertificateAuthorityConfiguration{
		Subject: &acmpca.ASN1Subject{
			CommonName:                 aws.String(p.Subject.CommonName),
			Country:                    aws.String(p.Subject.Country),
			DistinguishedNameQualifier: p.Subject.DistinguishedNameQualifier,
			GenerationQualifier:        p.Subject.GenerationQualifier,
			GivenName:                  p.Subject.GivenName,
			Initials:                   p.Subject.Initials,
			Locality:                   aws.String(p.Subject.Locality),
			Organization:               aws.String(p.Subject.Organization),
			OrganizationalUnit:         aws.String(p.Subject.OrganizationalUnit),
			Pseudonym:                  p.Subject.Pseudonym,
			SerialNumber:               p.Subject.SerialNumber,
			State:                      aws.String(p.Subject.State),
			Surname:                    p.Subject.Surname,
			Title:                      p.Subject.Title,
		},
		SigningAlgorithm: p.SigningAlgorithm,
		KeyAlgorithm:     p.KeyAlgorithm,
	}
	return m

}

// GenerateRevocationConfiguration from RevocationConfiguration
func GenerateRevocationConfiguration(p *v1alpha1.RevocationConfiguration) *acmpca.RevocationConfiguration {
	if p == nil {
		return nil
	}

	m := &acmpca.RevocationConfiguration{
		CrlConfiguration: &acmpca.CrlConfiguration{
			CustomCname:      p.CustomCname,
			Enabled:          aws.Bool(p.Enabled),
			ExpirationInDays: p.ExpirationInDays,
			S3BucketName:     p.S3BucketName,
		},
	}

	return m
}

// LateInitializeCertificateAuthority fills the empty fields in *v1beta1.CertificateAuthorityParameters with
// the values seen in acmpca.CertificateAuthority.
func LateInitializeCertificateAuthority(in *v1alpha1.CertificateAuthorityParameters, certificateAuthority *acmpca.CertificateAuthority) { // nolint:gocyclo
	if certificateAuthority == nil {
		return
	}

	if string(in.Type) == "" && string(certificateAuthority.Type) != "" {
		in.Type = certificateAuthority.Type
	}

	if (in.Status == nil || *in.Status == acmpca.CertificateAuthorityStatusPendingCertificate) && string(certificateAuthority.Status) != "" {
		in.Status = &certificateAuthority.Status
	}

	if aws.BoolValue(certificateAuthority.RevocationConfiguration.CrlConfiguration.Enabled) {
		if in.RevocationConfiguration.ExpirationInDays == nil && certificateAuthority.RevocationConfiguration.CrlConfiguration.ExpirationInDays != nil {
			in.RevocationConfiguration.ExpirationInDays = certificateAuthority.RevocationConfiguration.CrlConfiguration.ExpirationInDays
		}

		if in.RevocationConfiguration.CustomCname == nil && certificateAuthority.RevocationConfiguration.CrlConfiguration.CustomCname != nil {
			in.RevocationConfiguration.CustomCname = certificateAuthority.RevocationConfiguration.CrlConfiguration.CustomCname
		}
	}

	if in.CertificateAuthorityConfiguration.Subject.SerialNumber == nil && certificateAuthority.CertificateAuthorityConfiguration.Subject.SerialNumber != nil {
		in.CertificateAuthorityConfiguration.Subject.SerialNumber = certificateAuthority.CertificateAuthorityConfiguration.Subject.SerialNumber
	}
}

// IsCertificateAuthorityUpToDate checks whether there is a change in any of the modifiable fields.
func IsCertificateAuthorityUpToDate(p *v1alpha1.CertificateAuthority, cd acmpca.CertificateAuthority, tags []acmpca.Tag) bool { // nolint:gocyclo

	if aws.BoolValue(cd.RevocationConfiguration.CrlConfiguration.Enabled) {
		if !strings.EqualFold(aws.StringValue(p.Spec.ForProvider.RevocationConfiguration.CustomCname), aws.StringValue(cd.RevocationConfiguration.CrlConfiguration.CustomCname)) {
			return false
		}

		if !strings.EqualFold(aws.StringValue(p.Spec.ForProvider.RevocationConfiguration.S3BucketName), aws.StringValue(cd.RevocationConfiguration.CrlConfiguration.S3BucketName)) {
			return false
		}

		if p.Spec.ForProvider.RevocationConfiguration.Enabled != aws.BoolValue(cd.RevocationConfiguration.CrlConfiguration.Enabled) {
			return false
		}

		if aws.Int64Value(p.Spec.ForProvider.RevocationConfiguration.ExpirationInDays) != aws.Int64Value(cd.RevocationConfiguration.CrlConfiguration.ExpirationInDays) {
			return false
		}
	} else if p.Spec.ForProvider.RevocationConfiguration != nil {
		return false
	}

	if len(p.Spec.ForProvider.Tags) != len(tags) {
		return false
	}

	if *p.Spec.ForProvider.Status != cd.Status {
		return false
	}

	pTags := make(map[string]string, len(p.Spec.ForProvider.Tags))
	for _, tag := range p.Spec.ForProvider.Tags {
		pTags[tag.Key] = tag.Value
	}
	for _, tag := range tags {
		val, ok := pTags[aws.StringValue(tag.Key)]
		if !ok || !strings.EqualFold(val, aws.StringValue(tag.Value)) {
			return false
		}
	}

	return true
}

// GenerateCertificateAuthorityExternalStatus is used to produce CertificateAuthorityExternalStatus from acmpca.certificateAuthorityStatus and v1alpha1.CertificateAuthority
func GenerateCertificateAuthorityExternalStatus(certificateAuthority acmpca.CertificateAuthority) v1alpha1.CertificateAuthorityExternalStatus {
	return v1alpha1.CertificateAuthorityExternalStatus{
		CertificateAuthorityARN: aws.StringValue(certificateAuthority.Arn),
		Serial:                  aws.StringValue(certificateAuthority.Serial),
	}
}

// IsErrorNotFound returns true if the error code indicates that the item was not found
func IsErrorNotFound(err error) bool {
	if awsErr, ok := err.(awserr.Error); ok {
		if awsErr.Code() == acmpca.ErrCodeInvalidStateException {
			return true
		}
	}

	return false
}
