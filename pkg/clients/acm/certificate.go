package acm

import (
	"context"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/acm"
	"github.com/aws/aws-sdk-go-v2/service/acm/types"
	acmtypes "github.com/aws/aws-sdk-go-v2/service/acm/types"

	"github.com/crossplane/provider-aws/apis/acm/v1alpha1"
	awsclients "github.com/crossplane/provider-aws/pkg/clients"
)

// Client defines the CertificateManager operations
type Client interface {
	// GetCertificate(*acm.GetCertificateInput) acm.GetCertificate
	DescribeCertificate(context.Context, *acm.DescribeCertificateInput, ...func(*acm.Options)) (*acm.DescribeCertificateOutput, error)
	RequestCertificate(context.Context, *acm.RequestCertificateInput, ...func(*acm.Options)) (*acm.RequestCertificateOutput, error)
	DeleteCertificate(context.Context, *acm.DeleteCertificateInput, ...func(*acm.Options)) (*acm.DeleteCertificateOutput, error)
	UpdateCertificateOptions(context.Context, *acm.UpdateCertificateOptionsInput, ...func(*acm.Options)) (*acm.UpdateCertificateOptionsOutput, error)
	ListTagsForCertificate(context.Context, *acm.ListTagsForCertificateInput, ...func(*acm.Options)) (*acm.ListTagsForCertificateOutput, error)
	AddTagsToCertificate(context.Context, *acm.AddTagsToCertificateInput, ...func(*acm.Options)) (*acm.AddTagsToCertificateOutput, error)
	RenewCertificate(context.Context, *acm.RenewCertificateInput, ...func(*acm.Options)) (*acm.RenewCertificateOutput, error)
	RemoveTagsFromCertificate(context.Context, *acm.RemoveTagsFromCertificateInput, ...func(*acm.Options)) (*acm.RemoveTagsFromCertificateOutput, error)
}

// NewClient returns a new client using AWS credentials as JSON encoded data.
func NewClient(conf aws.Config) Client {
	return acm.NewFromConfig(conf)
}

// GenerateCreateCertificateInput from CertificateSpec
func GenerateCreateCertificateInput(name string, p *v1alpha1.CertificateParameters) *acm.RequestCertificateInput {

	m := &acm.RequestCertificateInput{
		DomainName:              aws.String(p.DomainName),
		CertificateAuthorityArn: p.CertificateAuthorityARN,
	}

	if p.CertificateTransparencyLoggingPreference != nil {
		m.Options = &types.CertificateOptions{CertificateTransparencyLoggingPreference: *p.CertificateTransparencyLoggingPreference}
	}

	if p.ValidationMethod != nil {
		m.ValidationMethod = *p.ValidationMethod
	}

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
func GenerateCertificateStatus(certificate types.CertificateDetail) v1alpha1.CertificateExternalStatus {
	if certificate.Type == acmtypes.CertificateTypeAmazonIssued && len(certificate.DomainValidationOptions) > 0 {
		if certificate.DomainValidationOptions[0].ResourceRecord != nil {
			return v1alpha1.CertificateExternalStatus{
				CertificateARN:     aws.ToString(certificate.CertificateArn),
				RenewalEligibility: certificate.RenewalEligibility,
				Status:             certificate.Status,
				Type:               certificate.Type,
				ResourceRecord: &v1alpha1.ResourceRecord{
					Name:  certificate.DomainValidationOptions[0].ResourceRecord.Name,
					Value: certificate.DomainValidationOptions[0].ResourceRecord.Value,
					Type:  (*string)(&certificate.DomainValidationOptions[0].ResourceRecord.Type),
				},
			}
		}
	}

	return v1alpha1.CertificateExternalStatus{
		CertificateARN:     aws.ToString(certificate.CertificateArn),
		RenewalEligibility: certificate.RenewalEligibility,
		Status:             certificate.Status,
		Type:               certificate.Type,
	}
}

// LateInitializeCertificate fills the empty fields in *v1beta1.CertificateParameters with
// the values seen in iam.Certificate.
func LateInitializeCertificate(in *v1alpha1.CertificateParameters, certificate *types.CertificateDetail) { // nolint:gocyclo
	if certificate == nil {
		return
	}

	in.DomainName = awsclients.LateInitializeString(in.DomainName, certificate.DomainName)

	in.CertificateAuthorityARN = awsclients.LateInitializeStringPtr(in.CertificateAuthorityARN, certificate.CertificateAuthorityArn)

	if in.CertificateTransparencyLoggingPreference == nil && certificate.Options != nil {
		in.CertificateTransparencyLoggingPreference = &certificate.Options.CertificateTransparencyLoggingPreference
	}

	if in.ValidationMethod == nil && len(certificate.DomainValidationOptions) != 0 {
		in.ValidationMethod = &certificate.DomainValidationOptions[0].ValidationMethod
	}

	if len(in.SubjectAlternativeNames) == 0 && len(certificate.SubjectAlternativeNames) != 0 {
		in.SubjectAlternativeNames = make([]*string, len(certificate.SubjectAlternativeNames))
		for i := range certificate.SubjectAlternativeNames {
			in.SubjectAlternativeNames[i] = &certificate.SubjectAlternativeNames[i]
		}
	}

	if len(in.DomainValidationOptions) == 0 && len(certificate.DomainValidationOptions) != 0 {
		in.DomainValidationOptions = make([]*v1alpha1.DomainValidationOption, len(certificate.DomainValidationOptions))
		for i, val := range certificate.DomainValidationOptions {
			in.DomainValidationOptions[i] = &v1alpha1.DomainValidationOption{
				DomainName:       awsclients.StringValue(val.DomainName),
				ValidationDomain: awsclients.StringValue(val.ValidationDomain),
			}
		}
	}
}

// IsCertificateUpToDate checks whether there is a change in any of the modifiable fields.
func IsCertificateUpToDate(p v1alpha1.CertificateParameters, cd types.CertificateDetail, tags []types.Tag) bool { // nolint:gocyclo

	if *p.CertificateTransparencyLoggingPreference != cd.Options.CertificateTransparencyLoggingPreference {
		return false
	}

	if len(p.Tags) != len(tags) {
		return false
	}

	pTags := make(map[string]string, len(p.Tags))
	for _, tag := range p.Tags {
		pTags[tag.Key] = tag.Value
	}
	for _, tag := range tags {
		val, ok := pTags[*tag.Key]
		if !ok || !strings.EqualFold(val, *tag.Value) {
			return false
		}
	}

	return !aws.ToBool(p.RenewCertificate)
}

// IsErrorNotFound returns true if the error code indicates that the item was not found
func IsErrorNotFound(err error) bool {
	_, ok := err.(*acmtypes.ResourceNotFoundException)
	return ok
}
