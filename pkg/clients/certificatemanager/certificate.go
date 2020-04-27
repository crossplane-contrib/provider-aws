package certificatemanager

import (
	"fmt"

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
		DomainName: aws.String(p.DomainName),
		// CertificateAuthorityArn: aws.String(p.CertificateAuthorityArn),
		ValidationMethod: acm.ValidationMethodDns,
		// IdempotencyToken:        p.IdempotencyToken,
		// Options:                 p.Options,
		// DomainValidationOptions: p.DomainValidationOptions,
		// SubjectAlternativeNames: p.SubjectAlternativeNames,
	}
	m.Tags = make([]acm.Tag, 1)
	m.Tags[0] = acm.Tag{
		Key:   aws.String("Name"),
		Value: aws.String(name),
	}

	fmt.Println(m)
	// if len(p.DomainValidationOptions) != 0 {
	// 	m.DomainValidationOptions = make([]acm.DomainValidationOption, len(p.DomainValidationOptions))
	// 	for i, val := range p.DomainValidationOptions {
	// 		m.DomainValidationOptions[i] = acm.DomainValidationOption{
	// 			DomainName:       &val.DomainName,
	// 			ValidationDomain: &val.ValidationDomain,
	// 		}
	// 	}
	// }

	// if len(p.SubjectAlternativeNames) != 0 {
	// 	for i, val := range p.SubjectAlternativeNames {
	// 		m.SubjectAlternativeNames[i] = p.SubjectAlternativeNames[i]
	// 	}
	// }

	// if len(p.Tags) != 0 {
	// 	m.Tags = make([]acm.Tag, len(p.Tags))
	// 	for i, val := range p.Tags {
	// 		m.Tags[i] = acm.Tag{
	// 			Key:   &val.Key,
	// 			Value: &val.Value,
	// 		}
	// 	}
	// }

	return m
}
