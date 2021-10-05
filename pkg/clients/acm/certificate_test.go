package acm

import (
	"testing"

	"github.com/aws/smithy-go/document"

	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/aws/aws-sdk-go-v2/service/acm"
	acmtypes "github.com/aws/aws-sdk-go-v2/service/acm/types"
	"github.com/google/go-cmp/cmp"

	"github.com/crossplane/provider-aws/apis/acm/v1alpha1"
	aws "github.com/crossplane/provider-aws/pkg/clients"
)

var (
	domainName              = "somedomain"
	certificateArn          = "somearn"
	certificateAuthorityArn = "someauthorityarn"
)

func TestGenerateCreateCertificateInput(t *testing.T) {
	certificateTransparencyLoggingPreference := acmtypes.CertificateTransparencyLoggingPreferenceDisabled
	validationMethod := acmtypes.ValidationMethodDns
	cases := map[string]struct {
		in  v1alpha1.CertificateParameters
		out acm.RequestCertificateInput
	}{
		"FilledInput": {
			in: v1alpha1.CertificateParameters{
				DomainName:                               domainName,
				CertificateAuthorityARN:                  aws.String(certificateAuthorityArn),
				CertificateTransparencyLoggingPreference: &certificateTransparencyLoggingPreference,
				ValidationMethod:                         &validationMethod,
				Tags: []v1alpha1.Tag{{
					Key:   "key1",
					Value: "value1",
				}},
			},
			out: acm.RequestCertificateInput{
				DomainName:              aws.String(domainName),
				CertificateAuthorityArn: aws.String(certificateAuthorityArn),
				Options:                 &acmtypes.CertificateOptions{CertificateTransparencyLoggingPreference: acmtypes.CertificateTransparencyLoggingPreferenceDisabled},
				ValidationMethod:        acmtypes.ValidationMethodDns,
				Tags: []acmtypes.Tag{{
					Key:   aws.String("key1"),
					Value: aws.String("value1"),
				}},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			r := GenerateCreateCertificateInput(name, &tc.in)

			if diff := cmp.Diff(r, &tc.out, cmpopts.IgnoreTypes(document.NoSerde{})); diff != "" {
				t.Errorf("GenerateCreateCertificateInput(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestLateInitializeCertificate(t *testing.T) {
	certificateTransparencyLoggingPreference := acmtypes.CertificateTransparencyLoggingPreferenceDisabled
	type args struct {
		spec *v1alpha1.CertificateParameters
		in   *acmtypes.CertificateDetail
	}
	cases := map[string]struct {
		args args
		want *v1alpha1.CertificateParameters
	}{
		"AllFilledNoDiff": {
			args: args{
				spec: &v1alpha1.CertificateParameters{
					DomainName:                               domainName,
					CertificateAuthorityARN:                  aws.String(certificateAuthorityArn),
					CertificateTransparencyLoggingPreference: &certificateTransparencyLoggingPreference,
				},
				in: &acmtypes.CertificateDetail{
					DomainName:              aws.String(domainName),
					CertificateAuthorityArn: aws.String(certificateAuthorityArn),
					Options:                 &acmtypes.CertificateOptions{CertificateTransparencyLoggingPreference: acmtypes.CertificateTransparencyLoggingPreferenceDisabled},
				},
			},
			want: &v1alpha1.CertificateParameters{
				DomainName:                               domainName,
				CertificateAuthorityARN:                  aws.String(certificateAuthorityArn),
				CertificateTransparencyLoggingPreference: &certificateTransparencyLoggingPreference,
			},
		},
		"AllFilledExternalDiff": {
			args: args{
				spec: &v1alpha1.CertificateParameters{
					DomainName:                               domainName,
					CertificateAuthorityARN:                  aws.String(certificateAuthorityArn),
					CertificateTransparencyLoggingPreference: &certificateTransparencyLoggingPreference,
				},
				in: &acmtypes.CertificateDetail{
					DomainName:              aws.String(domainName),
					CertificateAuthorityArn: aws.String(certificateAuthorityArn),
					Options:                 &acmtypes.CertificateOptions{CertificateTransparencyLoggingPreference: acmtypes.CertificateTransparencyLoggingPreferenceDisabled},
				},
			},
			want: &v1alpha1.CertificateParameters{
				DomainName:                               domainName,
				CertificateAuthorityARN:                  aws.String(certificateAuthorityArn),
				CertificateTransparencyLoggingPreference: &certificateTransparencyLoggingPreference,
			},
		},
		"PartialFilled": {
			args: args{
				spec: &v1alpha1.CertificateParameters{
					DomainName:              domainName,
					CertificateAuthorityARN: aws.String(certificateAuthorityArn),
				},
				in: &acmtypes.CertificateDetail{
					DomainName:              aws.String(domainName),
					CertificateAuthorityArn: aws.String(certificateAuthorityArn),
					Options:                 &acmtypes.CertificateOptions{CertificateTransparencyLoggingPreference: acmtypes.CertificateTransparencyLoggingPreferenceDisabled},
				},
			},
			want: &v1alpha1.CertificateParameters{
				DomainName:                               domainName,
				CertificateAuthorityARN:                  aws.String(certificateAuthorityArn),
				CertificateTransparencyLoggingPreference: &certificateTransparencyLoggingPreference,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			LateInitializeCertificate(tc.args.spec, tc.args.in)
			if diff := cmp.Diff(tc.args.spec, tc.want); diff != "" {
				t.Errorf("LateInitializeCertificate(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGenerateCertificateStatus(t *testing.T) {
	sName := "_xyz.crossplane.io."
	sType := "CNAME"
	sValue := "_xxx.zzz.acm-validations.aws."

	cases := map[string]struct {
		in  acmtypes.CertificateDetail
		out v1alpha1.CertificateExternalStatus
	}{
		"AllFilled": {
			in: acmtypes.CertificateDetail{
				CertificateArn:     aws.String(certificateArn),
				RenewalEligibility: acmtypes.RenewalEligibilityEligible,
			},
			out: v1alpha1.CertificateExternalStatus{
				CertificateARN:     certificateArn,
				RenewalEligibility: acmtypes.RenewalEligibilityEligible,
			},
		},
		"NoRoleId": {
			in: acmtypes.CertificateDetail{
				CertificateArn:     nil,
				RenewalEligibility: acmtypes.RenewalEligibilityEligible,
			},
			out: v1alpha1.CertificateExternalStatus{
				CertificateARN:     "",
				RenewalEligibility: acmtypes.RenewalEligibilityEligible,
			},
		},
		"DomainValidationOptionsResourceRecord": {
			in: acmtypes.CertificateDetail{
				CertificateArn:     aws.String(certificateArn),
				RenewalEligibility: acmtypes.RenewalEligibilityIneligible,
				Type:               acmtypes.CertificateTypeAmazonIssued,
				Status:             acmtypes.CertificateStatusPendingValidation,
				DomainValidationOptions: []acmtypes.DomainValidation{
					{
						DomainName: &sName,
						ResourceRecord: &acmtypes.ResourceRecord{
							Name:  &sName,
							Value: &sValue,
							Type:  acmtypes.RecordType(sType),
						},
					},
				},
			},
			out: v1alpha1.CertificateExternalStatus{
				CertificateARN:     certificateArn,
				RenewalEligibility: acmtypes.RenewalEligibilityIneligible,
				Type:               acmtypes.CertificateTypeAmazonIssued,
				Status:             acmtypes.CertificateStatusPendingValidation,
				ResourceRecord: &v1alpha1.ResourceRecord{
					Name:  &sName,
					Value: &sValue,
					Type:  &sType,
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			r := GenerateCertificateStatus(tc.in)
			if diff := cmp.Diff(r, tc.out); diff != "" {
				t.Errorf("GenerateCertificateStatus(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestIsCertificateUpToDate(t *testing.T) {
	certificateTransparencyLoggingPreference := acmtypes.CertificateTransparencyLoggingPreferenceDisabled
	type args struct {
		p    v1alpha1.CertificateParameters
		cd   acmtypes.CertificateDetail
		tags []acmtypes.Tag
	}

	cases := map[string]struct {
		args args
		want bool
	}{
		"SameFields": {
			args: args{
				cd: acmtypes.CertificateDetail{
					Options: &acmtypes.CertificateOptions{CertificateTransparencyLoggingPreference: acmtypes.CertificateTransparencyLoggingPreferenceDisabled},
				},
				p: v1alpha1.CertificateParameters{
					CertificateTransparencyLoggingPreference: &certificateTransparencyLoggingPreference,
					RenewCertificate:                         aws.Bool(false),
					Tags: []v1alpha1.Tag{{
						Key:   "key1",
						Value: "value1",
					}},
				},
				tags: []acmtypes.Tag{{
					Key:   aws.String("key1"),
					Value: aws.String("value1"),
				}},
			},
			want: true,
		},
		"DifferentCertificateDetail": {
			args: args{
				cd: acmtypes.CertificateDetail{
					Options: &acmtypes.CertificateOptions{CertificateTransparencyLoggingPreference: acmtypes.CertificateTransparencyLoggingPreferenceEnabled},
				},
				p: v1alpha1.CertificateParameters{
					CertificateTransparencyLoggingPreference: &certificateTransparencyLoggingPreference,
					RenewCertificate:                         aws.Bool(false),
					Tags: []v1alpha1.Tag{{
						Key:   "key1",
						Value: "value1",
					}},
				},
				tags: []acmtypes.Tag{{
					Key:   aws.String("key1"),
					Value: aws.String("value1"),
				}},
			},
			want: false,
		},
		"DifferentTags": {
			args: args{
				cd: acmtypes.CertificateDetail{
					Options: &acmtypes.CertificateOptions{CertificateTransparencyLoggingPreference: acmtypes.CertificateTransparencyLoggingPreferenceDisabled},
				},
				p: v1alpha1.CertificateParameters{
					CertificateTransparencyLoggingPreference: &certificateTransparencyLoggingPreference,
					RenewCertificate:                         aws.Bool(false),
					Tags: []v1alpha1.Tag{{
						Key:   "key1",
						Value: "value1",
					}},
				},
				tags: []acmtypes.Tag{{
					Key:   aws.String("key2"),
					Value: aws.String("value2"),
				}},
			},
			want: false,
		},
		"RenewCertificateTrueReturnsFalse": {
			args: args{
				cd: acmtypes.CertificateDetail{
					Options: &acmtypes.CertificateOptions{CertificateTransparencyLoggingPreference: acmtypes.CertificateTransparencyLoggingPreferenceDisabled},
				},
				p: v1alpha1.CertificateParameters{
					CertificateTransparencyLoggingPreference: &certificateTransparencyLoggingPreference,
					RenewCertificate:                         aws.Bool(true),
					Tags: []v1alpha1.Tag{{
						Key:   "key1",
						Value: "value1",
					}},
				},
				tags: []acmtypes.Tag{{
					Key:   aws.String("key1"),
					Value: aws.String("value1"),
				}},
			},
			want: false,
		},
		"RenewCertificateFalseReturnsTrue": {
			args: args{
				cd: acmtypes.CertificateDetail{
					Options: &acmtypes.CertificateOptions{CertificateTransparencyLoggingPreference: acmtypes.CertificateTransparencyLoggingPreferenceDisabled},
				},
				p: v1alpha1.CertificateParameters{
					CertificateTransparencyLoggingPreference: &certificateTransparencyLoggingPreference,
					RenewCertificate:                         aws.Bool(false),
					Tags: []v1alpha1.Tag{{
						Key:   "key1",
						Value: "value1",
					}},
				},
				tags: []acmtypes.Tag{{
					Key:   aws.String("key1"),
					Value: aws.String("value1"),
				}},
			},
			want: true,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := IsCertificateUpToDate(tc.args.p, tc.args.cd, tc.args.tags)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}
