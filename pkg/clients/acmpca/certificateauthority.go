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
	"context"
	"errors"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/acmpca"
	"github.com/aws/aws-sdk-go-v2/service/acmpca/types"

	"github.com/crossplane-contrib/provider-aws/apis/acmpca/v1beta1"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
)

// Client defines the CertificateManager operations
type Client interface {
	CreateCertificateAuthority(context.Context, *acmpca.CreateCertificateAuthorityInput, ...func(*acmpca.Options)) (*acmpca.CreateCertificateAuthorityOutput, error)
	DeleteCertificateAuthority(context.Context, *acmpca.DeleteCertificateAuthorityInput, ...func(*acmpca.Options)) (*acmpca.DeleteCertificateAuthorityOutput, error)
	UpdateCertificateAuthority(context.Context, *acmpca.UpdateCertificateAuthorityInput, ...func(*acmpca.Options)) (*acmpca.UpdateCertificateAuthorityOutput, error)
	DescribeCertificateAuthority(context.Context, *acmpca.DescribeCertificateAuthorityInput, ...func(*acmpca.Options)) (*acmpca.DescribeCertificateAuthorityOutput, error)
	ListTags(context.Context, *acmpca.ListTagsInput, ...func(*acmpca.Options)) (*acmpca.ListTagsOutput, error)
	UntagCertificateAuthority(context.Context, *acmpca.UntagCertificateAuthorityInput, ...func(*acmpca.Options)) (*acmpca.UntagCertificateAuthorityOutput, error)
	TagCertificateAuthority(context.Context, *acmpca.TagCertificateAuthorityInput, ...func(*acmpca.Options)) (*acmpca.TagCertificateAuthorityOutput, error)
}

// NewClient returns a new client using AWS credentials as JSON encoded data.
func NewClient(conf *aws.Config) Client {
	return acmpca.NewFromConfig(*conf)
}

// GenerateCreateCertificateAuthorityInput from certificateAuthorityParameters
func GenerateCreateCertificateAuthorityInput(p *v1beta1.CertificateAuthorityParameters) *acmpca.CreateCertificateAuthorityInput {
	m := &acmpca.CreateCertificateAuthorityInput{
		CertificateAuthorityType:          p.Type,
		CertificateAuthorityConfiguration: GenerateCertificateAuthorityConfiguration(p.CertificateAuthorityConfiguration),
		RevocationConfiguration:           GenerateRevocationConfiguration(p.RevocationConfiguration),
	}

	m.Tags = make([]types.Tag, len(p.Tags))
	for i, val := range p.Tags {
		m.Tags[i] = types.Tag{
			Key:   pointer.ToOrNilIfZeroValue(val.Key),
			Value: pointer.ToOrNilIfZeroValue(val.Value),
		}
	}

	return m
}

// GenerateCertificateAuthorityConfiguration from CertificateAuthorityConfiguration
func GenerateCertificateAuthorityConfiguration(p v1beta1.CertificateAuthorityConfiguration) *types.CertificateAuthorityConfiguration {

	m := &types.CertificateAuthorityConfiguration{
		Subject: &types.ASN1Subject{
			CommonName:                 pointer.ToOrNilIfZeroValue(p.Subject.CommonName),
			Country:                    pointer.ToOrNilIfZeroValue(p.Subject.Country),
			DistinguishedNameQualifier: p.Subject.DistinguishedNameQualifier,
			GenerationQualifier:        p.Subject.GenerationQualifier,
			GivenName:                  p.Subject.GivenName,
			Initials:                   p.Subject.Initials,
			Locality:                   pointer.ToOrNilIfZeroValue(p.Subject.Locality),
			Organization:               pointer.ToOrNilIfZeroValue(p.Subject.Organization),
			OrganizationalUnit:         pointer.ToOrNilIfZeroValue(p.Subject.OrganizationalUnit),
			Pseudonym:                  p.Subject.Pseudonym,
			SerialNumber:               p.Subject.SerialNumber,
			State:                      pointer.ToOrNilIfZeroValue(p.Subject.State),
			Surname:                    p.Subject.Surname,
			Title:                      p.Subject.Title,
		},
		SigningAlgorithm: p.SigningAlgorithm,
		KeyAlgorithm:     p.KeyAlgorithm,
	}
	return m

}

// GenerateRevocationConfiguration from RevocationConfiguration
func GenerateRevocationConfiguration(p *v1beta1.RevocationConfiguration) *types.RevocationConfiguration {
	if p == nil {
		return nil
	}

	m := &types.RevocationConfiguration{
		CrlConfiguration: &types.CrlConfiguration{
			CustomCname:      p.CustomCname,
			Enabled:          p.Enabled,
			ExpirationInDays: p.ExpirationInDays,
			S3BucketName:     p.S3BucketName,
		},
	}

	return m
}

// LateInitializeCertificateAuthority fills the empty fields in *v1beta1.CertificateAuthorityParameters with
// the values seen in acmpca.CertificateAuthority.
func LateInitializeCertificateAuthority(in *v1beta1.CertificateAuthorityParameters, certificateAuthority *types.CertificateAuthority) { //nolint:gocyclo
	if certificateAuthority == nil {
		return
	}

	if string(in.Type) == "" && string(certificateAuthority.Type) != "" {
		in.Type = certificateAuthority.Type
	}

	// NOTE(muvaf): Only ACTIVE and DISABLED statuses can be assigned by the user
	// so these are the only variants we support in spec. The current status
	// in the status.atProvider.
	if aws.ToString(in.Status) == "" && (certificateAuthority.Status == types.CertificateAuthorityStatusActive || certificateAuthority.Status == types.CertificateAuthorityStatusDisabled) {
		in.Status = pointer.ToOrNilIfZeroValue(string(certificateAuthority.Status))
	}

	if certificateAuthority.RevocationConfiguration.CrlConfiguration.Enabled {
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
func IsCertificateAuthorityUpToDate(p *v1beta1.CertificateAuthority, cd types.CertificateAuthority, tags []types.Tag) bool { //nolint:gocyclo

	if cd.RevocationConfiguration.CrlConfiguration.Enabled {
		if !strings.EqualFold(aws.ToString(p.Spec.ForProvider.RevocationConfiguration.CustomCname), aws.ToString(cd.RevocationConfiguration.CrlConfiguration.CustomCname)) {
			return false
		}

		if !strings.EqualFold(aws.ToString(p.Spec.ForProvider.RevocationConfiguration.S3BucketName), aws.ToString(cd.RevocationConfiguration.CrlConfiguration.S3BucketName)) {
			return false
		}

		if p.Spec.ForProvider.RevocationConfiguration.Enabled != cd.RevocationConfiguration.CrlConfiguration.Enabled {
			return false
		}

		if aws.ToInt32(p.Spec.ForProvider.RevocationConfiguration.ExpirationInDays) != aws.ToInt32(cd.RevocationConfiguration.CrlConfiguration.ExpirationInDays) {
			return false
		}
	} else if p.Spec.ForProvider.RevocationConfiguration != nil {
		return false
	}

	if len(p.Spec.ForProvider.Tags) != len(tags) {
		return false
	}

	desired := aws.ToString(p.Spec.ForProvider.Status)
	if (desired == string(types.CertificateAuthorityStatusActive) || desired == string(types.CertificateAuthorityStatusDisabled)) && desired != string(cd.Status) {
		return false
	}

	pTags := make(map[string]string, len(p.Spec.ForProvider.Tags))
	for _, tag := range p.Spec.ForProvider.Tags {
		pTags[tag.Key] = tag.Value
	}
	for _, tag := range tags {
		val, ok := pTags[aws.ToString(tag.Key)]
		if !ok || !strings.EqualFold(val, aws.ToString(tag.Value)) {
			return false
		}
	}

	return true
}

// GenerateCertificateAuthorityExternalStatus is used to produce CertificateAuthorityExternalStatus from acmpca.certificateAuthorityStatus and v1alpha1.CertificateAuthority
func GenerateCertificateAuthorityExternalStatus(certificateAuthority types.CertificateAuthority) v1beta1.CertificateAuthorityExternalStatus {
	return v1beta1.CertificateAuthorityExternalStatus{
		CertificateAuthorityARN: aws.ToString(certificateAuthority.Arn),
		Serial:                  aws.ToString(certificateAuthority.Serial),
		Status:                  string(certificateAuthority.Status),
	}
}

// IsErrorNotFound returns true if the error code indicates that the item was not found
func IsErrorNotFound(err error) bool {
	var ise *types.InvalidStateException
	return errors.As(err, &ise)
}
