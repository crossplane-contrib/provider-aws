package certificatemanager

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/acm"

	"github.com/crossplane/provider-aws/apis/certificatemanager/v1alpha1"
)

// Client defines the CertificateManager operations
type Client interface {
	GetCertificateRequest(*acm.GetCertificateInput) acm.GetCertificateRequest
	RequestCertificateRequest(*acm.RequestCertificateInput) acm.RequestCertificateRequest
	DeleteCertificateRequest(*acm.DeleteCertificateInput) acm.DeleteCertificateRequest
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

	fmt.Println(m)
	return m
}
