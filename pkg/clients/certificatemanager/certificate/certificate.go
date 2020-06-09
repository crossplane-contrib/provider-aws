package certificate

import (
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/awserr"
	"github.com/aws/aws-sdk-go-v2/service/acm"

	"github.com/crossplane/provider-aws/apis/certificatemanager/v1alpha1"
	awsclients "github.com/crossplane/provider-aws/pkg/clients"
)

// Client defines the CertificateManager operations
type Client interface {
	// GetCertificateRequest(*acm.GetCertificateInput) acm.GetCertificateRequest
	DescribeCertificateRequest(*acm.DescribeCertificateInput) acm.DescribeCertificateRequest
	RequestCertificateRequest(*acm.RequestCertificateInput) acm.RequestCertificateRequest
	DeleteCertificateRequest(*acm.DeleteCertificateInput) acm.DeleteCertificateRequest
	UpdateCertificateOptionsRequest(*acm.UpdateCertificateOptionsInput) acm.UpdateCertificateOptionsRequest
	ListTagsForCertificateRequest(*acm.ListTagsForCertificateInput) acm.ListTagsForCertificateRequest
	AddTagsToCertificateRequest(*acm.AddTagsToCertificateInput) acm.AddTagsToCertificateRequest
	RenewCertificateRequest(*acm.RenewCertificateInput) acm.RenewCertificateRequest
	RemoveTagsFromCertificateRequest(*acm.RemoveTagsFromCertificateInput) acm.RemoveTagsFromCertificateRequest
}

// NewClient returns a new client using AWS credentials as JSON encoded data.
func NewClient(conf *aws.Config) (Client, error) {
	return acm.New(*conf), nil
}

// GenerateCreateCertificateInput from CertificateSpec
func GenerateCreateCertificateInput(name string, p *v1alpha1.CertificateParameters) *acm.RequestCertificateInput {
	m := &acm.RequestCertificateInput{
		DomainName:              aws.String(p.DomainName),
		CertificateAuthorityArn: p.CertificateAuthorityArn,
		IdempotencyToken:        p.IdempotencyToken,
		ValidationMethod:        p.ValidationMethod,
		Options:                 &acm.CertificateOptions{CertificateTransparencyLoggingPreference: p.CertificateTransparencyLoggingPreference},
	}

	if len(p.DomainValidationOptions) != 0 {
		m.DomainValidationOptions = make([]acm.DomainValidationOption, len(p.DomainValidationOptions))
		for i, val := range p.DomainValidationOptions {
			m.DomainValidationOptions[i] = acm.DomainValidationOption{
				DomainName:       aws.String(val.DomainName),
				ValidationDomain: aws.String(val.ValidationDomain),
			}
		}
	}

	if len(p.SubjectAlternativeNames) != 0 {
		m.SubjectAlternativeNames = make([]string, len(p.SubjectAlternativeNames))
		for i := range p.SubjectAlternativeNames {
			m.SubjectAlternativeNames[i] = p.SubjectAlternativeNames[i]
		}
	}

	m.Tags = make([]acm.Tag, len(p.Tags))
	for i, val := range p.Tags {
		m.Tags[i] = acm.Tag{
			Key:   aws.String(val.Key),
			Value: aws.String(val.Value),
		}
	}
	return m
}

// GenerateCertificateStatus is used to produce CertificateExternalStatus from acm.certificateStatus
func GenerateCertificateStatus(certificate acm.CertificateDetail) v1alpha1.CertificateExternalStatus {
	return v1alpha1.CertificateExternalStatus{
		CertificateArn:     aws.StringValue(certificate.CertificateArn),
		RenewalEligibility: string(certificate.RenewalEligibility),
		Status:             certificate.Status,
	}
}

// GenerateCertificateOptionRequest return CertificateOptions from CertificateSpec
func GenerateCertificateOptionRequest(p *v1alpha1.CertificateParameters) *acm.CertificateOptions {

	return &acm.CertificateOptions{CertificateTransparencyLoggingPreference: p.CertificateTransparencyLoggingPreference}

}

// LateInitializeCertificate fills the empty fields in *v1beta1.CertificateParameters with
// the values seen in iam.Certificate.
func LateInitializeCertificate(in *v1alpha1.CertificateParameters, certificate *acm.CertificateDetail) { // nolint:gocyclo
	if certificate == nil {
		return
	}

	in.DomainName = awsclients.LateInitializeString(in.DomainName, certificate.DomainName)

	if aws.StringValue(in.CertificateAuthorityArn) == "" && certificate.CertificateAuthorityArn != nil {
		in.CertificateAuthorityArn = certificate.CertificateAuthorityArn
	}

	if string(in.CertificateTransparencyLoggingPreference) == "" && certificate.Options != nil {
		in.CertificateTransparencyLoggingPreference = certificate.Options.CertificateTransparencyLoggingPreference
	}

	if string(in.ValidationMethod) == "" && len(certificate.DomainValidationOptions) != 0 {
		in.ValidationMethod = certificate.DomainValidationOptions[0].ValidationMethod
	}

	if len(in.SubjectAlternativeNames) == 0 && len(certificate.SubjectAlternativeNames) != 0 {
		in.SubjectAlternativeNames = make([]string, len(certificate.SubjectAlternativeNames))
		for i := range certificate.SubjectAlternativeNames {
			in.SubjectAlternativeNames[i] = certificate.SubjectAlternativeNames[i]
		}
	}

	if len(in.DomainValidationOptions) == 0 && len(certificate.DomainValidationOptions) != 0 {
		in.DomainValidationOptions = make([]*v1alpha1.DomainValidationOption, len(certificate.DomainValidationOptions))
		for i, val := range certificate.DomainValidationOptions {
			in.DomainValidationOptions[i] = &v1alpha1.DomainValidationOption{
				DomainName:       aws.StringValue(val.DomainName),
				ValidationDomain: aws.StringValue(val.ValidationDomain),
			}
		}
	}

}

// IsCertificateUpToDate checks whether there is a change in any of the modifiable fields.
func IsCertificateUpToDate(p v1alpha1.CertificateParameters, cd acm.CertificateDetail, tags []acm.Tag) bool { // nolint:gocyclo

	if !strings.EqualFold(string(p.CertificateTransparencyLoggingPreference), string(cd.Options.CertificateTransparencyLoggingPreference)) {
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
		val, ok := pTags[aws.StringValue(tag.Key)]
		if !ok || !strings.EqualFold(val, aws.StringValue(tag.Value)) {
			return false
		}
	}

	return !aws.BoolValue(p.RenewCertificate)
}

// IsErrorNotFound returns true if the error code indicates that the item was not found
func IsErrorNotFound(err error) bool {
	if acmErr, ok := err.(awserr.Error); ok && acmErr.Code() == acm.ErrCodeResourceNotFoundException {
		return true
	}
	return false
}
