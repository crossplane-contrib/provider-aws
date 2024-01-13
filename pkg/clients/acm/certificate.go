package acm

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/acm"
	"github.com/aws/aws-sdk-go-v2/service/acm/types"
	acmtypes "github.com/aws/aws-sdk-go-v2/service/acm/types"

	"github.com/crossplane-contrib/provider-aws/apis/acm/v1beta1"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
)

// Client defines the CertificateManager operations
type Client interface {
	DescribeCertificate(context.Context, *acm.DescribeCertificateInput, ...func(*acm.Options)) (*acm.DescribeCertificateOutput, error)
	RequestCertificate(context.Context, *acm.RequestCertificateInput, ...func(*acm.Options)) (*acm.RequestCertificateOutput, error)
	DeleteCertificate(context.Context, *acm.DeleteCertificateInput, ...func(*acm.Options)) (*acm.DeleteCertificateOutput, error)
	UpdateCertificateOptions(context.Context, *acm.UpdateCertificateOptionsInput, ...func(*acm.Options)) (*acm.UpdateCertificateOptionsOutput, error)
	ListTagsForCertificate(context.Context, *acm.ListTagsForCertificateInput, ...func(*acm.Options)) (*acm.ListTagsForCertificateOutput, error)
	AddTagsToCertificate(context.Context, *acm.AddTagsToCertificateInput, ...func(*acm.Options)) (*acm.AddTagsToCertificateOutput, error)
	RemoveTagsFromCertificate(context.Context, *acm.RemoveTagsFromCertificateInput, ...func(*acm.Options)) (*acm.RemoveTagsFromCertificateOutput, error)
}

// NewClient returns a new client using AWS credentials as JSON encoded data.
func NewClient(conf aws.Config) Client {
	return acm.NewFromConfig(conf)
}

// GenerateCreateCertificateInput from CertificateSpec
func GenerateCreateCertificateInput(p v1beta1.CertificateParameters) *acm.RequestCertificateInput {
	m := &acm.RequestCertificateInput{
		DomainName:              aws.String(p.DomainName),
		CertificateAuthorityArn: p.CertificateAuthorityARN,
	}

	if p.Options != nil {
		m.Options = &types.CertificateOptions{
			CertificateTransparencyLoggingPreference: types.CertificateTransparencyLoggingPreference(p.Options.CertificateTransparencyLoggingPreference),
		}
	}

	m.ValidationMethod = types.ValidationMethod(p.ValidationMethod)

	if len(p.DomainValidationOptions) != 0 {
		m.DomainValidationOptions = make([]types.DomainValidationOption, len(p.DomainValidationOptions))
		for i, val := range p.DomainValidationOptions {
			m.DomainValidationOptions[i] = types.DomainValidationOption{
				DomainName:       aws.String(val.DomainName),
				ValidationDomain: aws.String(val.ValidationDomain),
			}
		}
	}

	if p.SubjectAlternativeNames != nil {
		m.SubjectAlternativeNames = make([]string, len(p.SubjectAlternativeNames))
		for i := range p.SubjectAlternativeNames {
			m.SubjectAlternativeNames[i] = *p.SubjectAlternativeNames[i]
		}
	}

	m.KeyAlgorithm = types.KeyAlgorithm(pointer.StringValue(p.KeyAlgorithm))

	m.Tags = make([]types.Tag, len(p.Tags))
	for i, val := range p.Tags {
		m.Tags[i] = types.Tag{
			Key:   aws.String(val.Key),
			Value: aws.String(val.Value),
		}
	}
	return m
}

// GenerateCertificateStatus is used to produce CertificateExternalStatus from acm.certificateStatus
func GenerateCertificateStatus(certificate types.CertificateDetail) v1beta1.CertificateExternalStatus {
	if certificate.Type == acmtypes.CertificateTypeAmazonIssued && len(certificate.DomainValidationOptions) > 0 {
		if certificate.DomainValidationOptions[0].ResourceRecord != nil {
			return v1beta1.CertificateExternalStatus{
				CertificateARN:     aws.ToString(certificate.CertificateArn),
				RenewalEligibility: string(certificate.RenewalEligibility),
				Status:             string(certificate.Status),
				Type:               string(certificate.Type),
				ResourceRecord: &v1beta1.ResourceRecord{
					Name:  certificate.DomainValidationOptions[0].ResourceRecord.Name,
					Value: certificate.DomainValidationOptions[0].ResourceRecord.Value,
					Type:  (*string)(&certificate.DomainValidationOptions[0].ResourceRecord.Type),
				},
			}
		}
	}

	return v1beta1.CertificateExternalStatus{
		CertificateARN:     aws.ToString(certificate.CertificateArn),
		RenewalEligibility: string(certificate.RenewalEligibility),
		Status:             string(certificate.Status),
		Type:               string(certificate.Type),
	}
}

// LateInitializeCertificate fills the empty fields in *v1beta1.CertificateParameters with
// the values seen in iam.Certificate.
func LateInitializeCertificate(in *v1beta1.CertificateParameters, certificate *types.CertificateDetail) {
	in.CertificateAuthorityARN = pointer.LateInitialize(in.CertificateAuthorityARN, certificate.CertificateAuthorityArn)
	if in.Options == nil && certificate.Options != nil {
		in.Options = &v1beta1.CertificateOptions{
			CertificateTransparencyLoggingPreference: string(certificate.Options.CertificateTransparencyLoggingPreference),
		}
	}

	if in.SubjectAlternativeNames == nil && len(certificate.SubjectAlternativeNames) != 0 {
		in.SubjectAlternativeNames = make([]*string, len(certificate.SubjectAlternativeNames))
		for i := range certificate.SubjectAlternativeNames {
			in.SubjectAlternativeNames[i] = &certificate.SubjectAlternativeNames[i]
		}
	}

	if in.DomainValidationOptions == nil && len(certificate.DomainValidationOptions) != 0 {
		in.DomainValidationOptions = make([]*v1beta1.DomainValidationOption, len(certificate.DomainValidationOptions))
		for i, val := range certificate.DomainValidationOptions {
			in.DomainValidationOptions[i] = &v1beta1.DomainValidationOption{
				DomainName:       pointer.StringValue(val.DomainName),
				ValidationDomain: pointer.StringValue(val.ValidationDomain),
			}
		}
	}
}

// IsCertificateUpToDate checks whether there is a change in any of the modifiable fields.
func IsCertificateUpToDate(p v1beta1.CertificateParameters, cd types.CertificateDetail, tags []types.Tag) bool {
	if (p.Options != nil && cd.Options == nil) || (p.Options == nil && cd.Options != nil) {
		return false
	}
	if p.Options != nil && cd.Options != nil &&
		p.Options.CertificateTransparencyLoggingPreference != string(cd.Options.CertificateTransparencyLoggingPreference) {
		return false
	}
	add, remove := DiffTags(p.Tags, tags)
	return len(add) == 0 && len(remove) == 0
}

// IsErrorNotFound returns true if the error code indicates that the item was not found
func IsErrorNotFound(err error) bool {
	var notFoundError *acmtypes.ResourceNotFoundException
	return errors.As(err, &notFoundError)
}

// DiffTags returns tags that should be added or removed.
func DiffTags(spec []v1beta1.Tag, current []acmtypes.Tag) (addTags []acmtypes.Tag, remove []acmtypes.Tag) {
	addMap := make(map[string]string, len(spec))
	for _, t := range spec {
		addMap[t.Key] = t.Value
	}
	removeMap := map[string]string{}
	for _, t := range current {
		if addMap[aws.ToString(t.Key)] == aws.ToString(t.Value) {
			delete(addMap, aws.ToString(t.Key))
			continue
		}
		removeMap[aws.ToString(t.Key)] = aws.ToString(t.Value)
	}
	for k, v := range addMap {
		addTags = append(addTags, acmtypes.Tag{Key: aws.String(k), Value: aws.String(v)})
	}
	for k, v := range removeMap {
		remove = append(remove, acmtypes.Tag{Key: aws.String(k), Value: aws.String(v)})
	}
	return
}
