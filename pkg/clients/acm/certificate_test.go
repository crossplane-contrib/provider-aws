package acm

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/acm"
	acmtypes "github.com/aws/aws-sdk-go-v2/service/acm/types"
	"github.com/aws/smithy-go/document"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/crossplane-contrib/provider-aws/apis/acm/v1beta1"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
)

var (
	domainName              = "somedomain"
	certificateArn          = "somearn"
	certificateAuthorityArn = "someauthorityarn"
)

func TestGenerateCreateCertificateInput(t *testing.T) {
	certificateTransparencyLoggingPreference := string(acmtypes.CertificateTransparencyLoggingPreferenceDisabled)
	validationMethod := string(acmtypes.ValidationMethodDns)
	cases := map[string]struct {
		in  v1beta1.CertificateParameters
		out acm.RequestCertificateInput
	}{
		"FilledInput": {
			in: v1beta1.CertificateParameters{
				DomainName:              domainName,
				CertificateAuthorityARN: pointer.ToOrNilIfZeroValue(certificateAuthorityArn),
				Options: &v1beta1.CertificateOptions{
					CertificateTransparencyLoggingPreference: certificateTransparencyLoggingPreference,
				},
				ValidationMethod: validationMethod,
				Tags: []v1beta1.Tag{{
					Key:   "key1",
					Value: "value1",
				}},
			},
			out: acm.RequestCertificateInput{
				DomainName:              pointer.ToOrNilIfZeroValue(domainName),
				CertificateAuthorityArn: pointer.ToOrNilIfZeroValue(certificateAuthorityArn),
				Options:                 &acmtypes.CertificateOptions{CertificateTransparencyLoggingPreference: acmtypes.CertificateTransparencyLoggingPreferenceDisabled},
				ValidationMethod:        acmtypes.ValidationMethodDns,
				Tags: []acmtypes.Tag{{
					Key:   pointer.ToOrNilIfZeroValue("key1"),
					Value: pointer.ToOrNilIfZeroValue("value1"),
				}},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			r := GenerateCreateCertificateInput(tc.in)

			if diff := cmp.Diff(r, &tc.out, cmpopts.IgnoreTypes(document.NoSerde{})); diff != "" {
				t.Errorf("GenerateCreateCertificateInput(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestLateInitializeCertificate(t *testing.T) {
	certificateTransparencyLoggingPreference := string(acmtypes.CertificateTransparencyLoggingPreferenceDisabled)
	type args struct {
		spec *v1beta1.CertificateParameters
		in   *acmtypes.CertificateDetail
	}
	cases := map[string]struct {
		args args
		want *v1beta1.CertificateParameters
	}{
		"AllFilledNoDiff": {
			args: args{
				spec: &v1beta1.CertificateParameters{
					DomainName:              domainName,
					CertificateAuthorityARN: pointer.ToOrNilIfZeroValue(certificateAuthorityArn),
					Options: &v1beta1.CertificateOptions{
						CertificateTransparencyLoggingPreference: certificateTransparencyLoggingPreference,
					},
				},
				in: &acmtypes.CertificateDetail{
					DomainName:              pointer.ToOrNilIfZeroValue(domainName),
					CertificateAuthorityArn: pointer.ToOrNilIfZeroValue(certificateAuthorityArn),
					Options:                 &acmtypes.CertificateOptions{CertificateTransparencyLoggingPreference: acmtypes.CertificateTransparencyLoggingPreferenceDisabled},
				},
			},
			want: &v1beta1.CertificateParameters{
				DomainName:              domainName,
				CertificateAuthorityARN: pointer.ToOrNilIfZeroValue(certificateAuthorityArn),
				Options: &v1beta1.CertificateOptions{
					CertificateTransparencyLoggingPreference: certificateTransparencyLoggingPreference,
				},
			},
		},
		"AllFilledExternalDiff": {
			args: args{
				spec: &v1beta1.CertificateParameters{
					DomainName:              domainName,
					CertificateAuthorityARN: pointer.ToOrNilIfZeroValue(certificateAuthorityArn),
					Options: &v1beta1.CertificateOptions{
						CertificateTransparencyLoggingPreference: certificateTransparencyLoggingPreference,
					},
				},
				in: &acmtypes.CertificateDetail{
					DomainName:              pointer.ToOrNilIfZeroValue(domainName),
					CertificateAuthorityArn: pointer.ToOrNilIfZeroValue(certificateAuthorityArn),
					Options:                 &acmtypes.CertificateOptions{CertificateTransparencyLoggingPreference: acmtypes.CertificateTransparencyLoggingPreferenceDisabled},
				},
			},
			want: &v1beta1.CertificateParameters{
				DomainName:              domainName,
				CertificateAuthorityARN: pointer.ToOrNilIfZeroValue(certificateAuthorityArn),
				Options: &v1beta1.CertificateOptions{
					CertificateTransparencyLoggingPreference: certificateTransparencyLoggingPreference,
				},
			},
		},
		"PartialFilled": {
			args: args{
				spec: &v1beta1.CertificateParameters{
					DomainName:              domainName,
					CertificateAuthorityARN: pointer.ToOrNilIfZeroValue(certificateAuthorityArn),
				},
				in: &acmtypes.CertificateDetail{
					DomainName:              pointer.ToOrNilIfZeroValue(domainName),
					CertificateAuthorityArn: pointer.ToOrNilIfZeroValue(certificateAuthorityArn),
					Options:                 &acmtypes.CertificateOptions{CertificateTransparencyLoggingPreference: acmtypes.CertificateTransparencyLoggingPreferenceDisabled},
				},
			},
			want: &v1beta1.CertificateParameters{
				DomainName:              domainName,
				CertificateAuthorityARN: pointer.ToOrNilIfZeroValue(certificateAuthorityArn),
				Options: &v1beta1.CertificateOptions{
					CertificateTransparencyLoggingPreference: certificateTransparencyLoggingPreference,
				},
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
		out v1beta1.CertificateExternalStatus
	}{
		"AllFilled": {
			in: acmtypes.CertificateDetail{
				CertificateArn:     pointer.ToOrNilIfZeroValue(certificateArn),
				RenewalEligibility: acmtypes.RenewalEligibilityEligible,
			},
			out: v1beta1.CertificateExternalStatus{
				CertificateARN:     certificateArn,
				RenewalEligibility: string(acmtypes.RenewalEligibilityEligible),
			},
		},
		"NoRoleId": {
			in: acmtypes.CertificateDetail{
				CertificateArn:     nil,
				RenewalEligibility: acmtypes.RenewalEligibilityEligible,
			},
			out: v1beta1.CertificateExternalStatus{
				CertificateARN:     "",
				RenewalEligibility: string(acmtypes.RenewalEligibilityEligible),
			},
		},
		"DomainValidationOptionsResourceRecord": {
			in: acmtypes.CertificateDetail{
				CertificateArn:     pointer.ToOrNilIfZeroValue(certificateArn),
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
			out: v1beta1.CertificateExternalStatus{
				CertificateARN:     certificateArn,
				RenewalEligibility: string(acmtypes.RenewalEligibilityIneligible),
				Type:               string(acmtypes.CertificateTypeAmazonIssued),
				Status:             string(acmtypes.CertificateStatusPendingValidation),
				ResourceRecord: &v1beta1.ResourceRecord{
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
	certificateTransparencyLoggingPreference := string(acmtypes.CertificateTransparencyLoggingPreferenceDisabled)
	type args struct {
		p    v1beta1.CertificateParameters
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
				p: v1beta1.CertificateParameters{
					Options: &v1beta1.CertificateOptions{
						CertificateTransparencyLoggingPreference: certificateTransparencyLoggingPreference,
					},
					Tags: []v1beta1.Tag{{
						Key:   "key1",
						Value: "value1",
					}},
				},
				tags: []acmtypes.Tag{{
					Key:   pointer.ToOrNilIfZeroValue("key1"),
					Value: pointer.ToOrNilIfZeroValue("value1"),
				}},
			},
			want: true,
		},
		"DifferentCertificateDetail": {
			args: args{
				cd: acmtypes.CertificateDetail{
					Options: &acmtypes.CertificateOptions{CertificateTransparencyLoggingPreference: acmtypes.CertificateTransparencyLoggingPreferenceEnabled},
				},
				p: v1beta1.CertificateParameters{
					Options: &v1beta1.CertificateOptions{
						CertificateTransparencyLoggingPreference: certificateTransparencyLoggingPreference,
					},
					Tags: []v1beta1.Tag{{
						Key:   "key1",
						Value: "value1",
					}},
				},
				tags: []acmtypes.Tag{{
					Key:   pointer.ToOrNilIfZeroValue("key1"),
					Value: pointer.ToOrNilIfZeroValue("value1"),
				}},
			},
			want: false,
		},
		"DifferentTags": {
			args: args{
				cd: acmtypes.CertificateDetail{
					Options: &acmtypes.CertificateOptions{CertificateTransparencyLoggingPreference: acmtypes.CertificateTransparencyLoggingPreferenceDisabled},
				},
				p: v1beta1.CertificateParameters{
					Options: &v1beta1.CertificateOptions{
						CertificateTransparencyLoggingPreference: certificateTransparencyLoggingPreference,
					},
					Tags: []v1beta1.Tag{{
						Key:   "key1",
						Value: "value1",
					}},
				},
				tags: []acmtypes.Tag{{
					Key:   pointer.ToOrNilIfZeroValue("key2"),
					Value: pointer.ToOrNilIfZeroValue("value2"),
				}},
			},
			want: false,
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
