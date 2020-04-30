package certificatemanager

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/awserr"
	"github.com/aws/aws-sdk-go-v2/service/acm"
	"github.com/aws/aws-sdk-go-v2/service/iam"

	"github.com/crossplane/provider-aws/apis/certificatemanager/v1alpha1"
	awsclients "github.com/crossplane/provider-aws/pkg/clients"
)

// Client defines the CertificateManager operations
type Client interface {
	// GetCertificateRequest(*acm.GetCertificateInput) acm.GetCertificateRequest
	DescribeCertificateRequest(*acm.DescribeCertificateInput) acm.DescribeCertificateRequest
	RequestCertificateRequest(*acm.RequestCertificateInput) acm.RequestCertificateRequest
	DeleteCertificateRequest(*acm.DeleteCertificateInput) acm.DeleteCertificateRequest
}

// NewClient returns a new client using AWS credentials as JSON encoded data.
func NewClient(conf *aws.Config) (Client, error) {
	return acm.New(*conf), nil
}

// GenerateCreateCertificateInput from CertificateSpec
func GenerateCreateCertificateInput(name string, p *v1alpha1.CertificateParameters) *acm.RequestCertificateInput {
	fmt.Println("GenerateCreateCertificateInput | Entry")
	m := &acm.RequestCertificateInput{
		DomainName:              aws.String(p.DomainName),
		CertificateAuthorityArn: p.CertificateAuthorityArn,
		IdempotencyToken:        p.IdempotencyToken,
	}

	if strings.EqualFold(p.CertificateTransparencyLoggingPreference, "DISABLED") {
		m.Options = &acm.CertificateOptions{CertificateTransparencyLoggingPreference: acm.CertificateTransparencyLoggingPreferenceDisabled}
	} else if strings.EqualFold(p.CertificateTransparencyLoggingPreference, "ENABLED") {
		m.Options = &acm.CertificateOptions{CertificateTransparencyLoggingPreference: acm.CertificateTransparencyLoggingPreferenceEnabled}
	}

	if strings.EqualFold(p.ValidationMethod, "EMAIL") {
		m.ValidationMethod = acm.ValidationMethodEmail
	} else if strings.EqualFold(p.ValidationMethod, "DNS") {
		m.ValidationMethod = acm.ValidationMethodDns
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

	p.Tags = append(p.Tags, v1alpha1.Tag{
		Key:   "Name",
		Value: name,
	})

	m.Tags = make([]acm.Tag, len(p.Tags))
	for i, val := range p.Tags {
		m.Tags[i] = acm.Tag{
			Key:   aws.String(val.Key),
			Value: aws.String(val.Value),
		}
	}
	fmt.Println("GenerateCreateCertificateInput | Exit")
	return m
}

// GenerateCertificateStatus is used to produce CertificateExternalStatus from acm.certificateStatus
func GenerateCertificateStatus(certificate acm.CertificateDetail) v1alpha1.CertificateExternalStatus {
	return v1alpha1.CertificateExternalStatus{
		CertificateArn: aws.StringValue(certificate.CertificateArn),
	}

}

// LateInitializeCertificate fills the empty fields in *v1beta1.CertificateParameters with
// the values seen in iam.Certificate.
func LateInitializeCertificate(in *v1alpha1.CertificateParameters, certificate *acm.CertificateDetail) { // nolint:gocyclo
	if certificate == nil {
		return
	}
	fmt.Println("LateInitializeCertificate | Entry")

	in.DomainName = awsclients.LateInitializeString(in.DomainName, certificate.DomainName)

	if aws.StringValue(in.CertificateAuthorityArn) == "" && certificate.CertificateAuthorityArn != nil {
		in.CertificateAuthorityArn = certificate.CertificateAuthorityArn
	}

	if in.CertificateTransparencyLoggingPreference == "" && certificate.Options != nil {
		in.CertificateTransparencyLoggingPreference = string(certificate.Options.CertificateTransparencyLoggingPreference)
	}

	if in.ValidationMethod == "" && len(certificate.DomainValidationOptions) != 0 {
		in.ValidationMethod = string(certificate.DomainValidationOptions[0].ValidationMethod)
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

	fmt.Println("LateInitializeCertificate | Exit")
}

// IsErrorNotFound returns true if the error code indicates that the item was not found
func IsErrorNotFound(err error) bool {
	if iamErr, ok := err.(awserr.Error); ok && iamErr.Code() == iam.ErrCodeNoSuchEntityException {
		return true
	}
	return false
}
