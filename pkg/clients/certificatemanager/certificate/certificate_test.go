package certificate

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/acm"
	"github.com/google/go-cmp/cmp"

	"github.com/crossplane/provider-aws/apis/certificatemanager/v1alpha1"
	aws "github.com/crossplane/provider-aws/pkg/clients"
)

var (
	domainName                 = "somedomain"
	certificateArn             = "somearn"
	renewalEligibilityEligible = "ELIGIBLE"
	certificateAuthorityArn    = "someauthorityarn"
)

// func addCertificateOutputFields(c *acm.CertificateDetail) {
// 	c.CertificateArn = aws.String(certificateArn)
// 	c.RenewalEligibility = acm.RenewalEligibilityEligible
// }

func TestGenerateCreateCertificateInput(t *testing.T) {
	cases := map[string]struct {
		in  v1alpha1.CertificateParameters
		out acm.RequestCertificateInput
	}{
		"FilledInput": {
			in: v1alpha1.CertificateParameters{
				DomainName:                               domainName,
				CertificateAuthorityArn:                  aws.String(certificateAuthorityArn),
				CertificateTransparencyLoggingPreference: acm.CertificateTransparencyLoggingPreferenceDisabled,
				Tags: []v1alpha1.Tag{{
					Key:   "key1",
					Value: "value1",
				}},
			},
			out: acm.RequestCertificateInput{
				DomainName:              aws.String(domainName),
				CertificateAuthorityArn: aws.String(certificateAuthorityArn),
				Options:                 &acm.CertificateOptions{CertificateTransparencyLoggingPreference: acm.CertificateTransparencyLoggingPreferenceDisabled},
				Tags: []acm.Tag{{
					Key:   aws.String("key1"),
					Value: aws.String("value1"),
				}},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			r := GenerateCreateCertificateInput(name, &tc.in)

			if diff := cmp.Diff(r, &tc.out); diff != "" {
				t.Errorf("GenerateCreateCertificateInput(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestLateInitializeCertificate(t *testing.T) {
	type args struct {
		spec *v1alpha1.CertificateParameters
		in   *acm.CertificateDetail
	}
	cases := map[string]struct {
		args args
		want *v1alpha1.CertificateParameters
	}{
		"AllFilledNoDiff": {
			args: args{
				spec: &v1alpha1.CertificateParameters{
					DomainName:                               domainName,
					CertificateAuthorityArn:                  aws.String(certificateAuthorityArn),
					CertificateTransparencyLoggingPreference: acm.CertificateTransparencyLoggingPreferenceDisabled,
				},
				in: &acm.CertificateDetail{
					DomainName:              aws.String(domainName),
					CertificateAuthorityArn: aws.String(certificateAuthorityArn),
					Options:                 &acm.CertificateOptions{CertificateTransparencyLoggingPreference: acm.CertificateTransparencyLoggingPreferenceDisabled},
				},
			},
			want: &v1alpha1.CertificateParameters{
				DomainName:                               domainName,
				CertificateAuthorityArn:                  aws.String(certificateAuthorityArn),
				CertificateTransparencyLoggingPreference: acm.CertificateTransparencyLoggingPreferenceDisabled,
			},
		},
		"AllFilledExternalDiff": {
			args: args{
				spec: &v1alpha1.CertificateParameters{
					DomainName:                               domainName,
					CertificateAuthorityArn:                  aws.String(certificateAuthorityArn),
					CertificateTransparencyLoggingPreference: acm.CertificateTransparencyLoggingPreferenceDisabled,
				},
				in: &acm.CertificateDetail{
					DomainName:              aws.String(domainName),
					CertificateAuthorityArn: aws.String(certificateAuthorityArn),
					Options:                 &acm.CertificateOptions{CertificateTransparencyLoggingPreference: acm.CertificateTransparencyLoggingPreferenceDisabled},
				},
			},
			want: &v1alpha1.CertificateParameters{
				DomainName:                               domainName,
				CertificateAuthorityArn:                  aws.String(certificateAuthorityArn),
				CertificateTransparencyLoggingPreference: acm.CertificateTransparencyLoggingPreferenceDisabled,
			},
		},
		"PartialFilled": {
			args: args{
				spec: &v1alpha1.CertificateParameters{
					DomainName:              domainName,
					CertificateAuthorityArn: aws.String(certificateAuthorityArn),
				},
				in: &acm.CertificateDetail{
					DomainName:              aws.String(domainName),
					CertificateAuthorityArn: aws.String(certificateAuthorityArn),
					Options:                 &acm.CertificateOptions{CertificateTransparencyLoggingPreference: acm.CertificateTransparencyLoggingPreferenceDisabled},
				},
			},
			want: &v1alpha1.CertificateParameters{
				DomainName:                               domainName,
				CertificateAuthorityArn:                  aws.String(certificateAuthorityArn),
				CertificateTransparencyLoggingPreference: acm.CertificateTransparencyLoggingPreferenceDisabled,
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

func TestGenerateCertificateOptionRequest(t *testing.T) {
	cases := map[string]struct {
		in  *v1alpha1.CertificateParameters
		out *acm.CertificateOptions
	}{
		"DisabledOption": {
			in: &v1alpha1.CertificateParameters{
				CertificateTransparencyLoggingPreference: "DISABLED",
			},
			out: &acm.CertificateOptions{
				CertificateTransparencyLoggingPreference: acm.CertificateTransparencyLoggingPreferenceDisabled,
			},
		},
		"EnabledOption": {
			in: &v1alpha1.CertificateParameters{
				CertificateTransparencyLoggingPreference: "ENABLED",
			},
			out: &acm.CertificateOptions{
				CertificateTransparencyLoggingPreference: acm.CertificateTransparencyLoggingPreferenceEnabled,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			r := GenerateCertificateOptionRequest(tc.in)
			if diff := cmp.Diff(r, tc.out); diff != "" {
				t.Errorf("GenerateCertificateOptionRequest(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGenerateCertificateStatus(t *testing.T) {
	cases := map[string]struct {
		in  acm.CertificateDetail
		out v1alpha1.CertificateExternalStatus
	}{
		"AllFilled": {
			in: acm.CertificateDetail{
				CertificateArn:     aws.String(certificateArn),
				RenewalEligibility: acm.RenewalEligibilityEligible,
			},
			out: v1alpha1.CertificateExternalStatus{
				CertificateArn:     certificateArn,
				RenewalEligibility: renewalEligibilityEligible,
			},
		},
		"NoRoleId": {
			in: acm.CertificateDetail{
				CertificateArn:     nil,
				RenewalEligibility: acm.RenewalEligibilityEligible,
			},
			out: v1alpha1.CertificateExternalStatus{
				CertificateArn:     "",
				RenewalEligibility: renewalEligibilityEligible,
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
	type args struct {
		p    v1alpha1.CertificateParameters
		cd   acm.CertificateDetail
		tags []acm.Tag
	}

	cases := map[string]struct {
		args args
		want bool
	}{
		"SameFields": {
			args: args{
				cd: acm.CertificateDetail{
					Options: &acm.CertificateOptions{CertificateTransparencyLoggingPreference: acm.CertificateTransparencyLoggingPreferenceEnabled},
				},
				p: v1alpha1.CertificateParameters{
					CertificateTransparencyLoggingPreference: acm.CertificateTransparencyLoggingPreferenceEnabled,
					RenewCertificate:                         aws.Bool(false),
					Tags: []v1alpha1.Tag{{
						Key:   "key1",
						Value: "value1",
					}},
				},
				tags: []acm.Tag{{
					Key:   aws.String("key1"),
					Value: aws.String("value1"),
				}},
			},
			want: true,
		},
		"DifferentFields": {
			args: args{
				cd: acm.CertificateDetail{
					Options: &acm.CertificateOptions{CertificateTransparencyLoggingPreference: acm.CertificateTransparencyLoggingPreferenceEnabled},
				},
				p: v1alpha1.CertificateParameters{
					CertificateTransparencyLoggingPreference: acm.CertificateTransparencyLoggingPreferenceDisabled,
					RenewCertificate:                         aws.Bool(false),
					Tags: []v1alpha1.Tag{{
						Key:   "key1",
						Value: "value1",
					}},
				},
				tags: []acm.Tag{{
					Key:   aws.String("key1"),
					Value: aws.String("value1"),
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
