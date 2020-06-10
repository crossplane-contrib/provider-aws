package certificateauthority

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/acmpca"
	"github.com/google/go-cmp/cmp"

	"github.com/crossplane/provider-aws/apis/certificatemanager/v1alpha1"
	aws "github.com/crossplane/provider-aws/pkg/clients"
)

var (
	certificateAuthorityArn             = "someauthorityarn"
	customCname                         = "soemcustomname"
	revocationConfigurationEnabled      = true
	revocationConfigurationEnabledfalse = false
	renewalPermission                   = true
	s3BucketName                        = "somes3bucketname"
	commonName                          = "someCommonName"
	country                             = "someCountry"
	distinguishedNameQualifier          = "someDistinguishedNameQualifier"
	generationQualifier                 = "somegenerationQualifier"
	givenName                           = "somegivenName"
	initials                            = "someinitials"
	locality                            = "somelocality"
	organization                        = "someorganization"
	organizationalUnit                  = "someOrganizationalUnit"
	pseudonym                           = "somePseudonym"
	serialNumber                        = "someSerialNumber"
	state                               = "someState"
	surname                             = "someSurname"
	title                               = "someTitle"
	idempotencyToken                    = "someidempotencyToken"
)

func TestGenerateCreateCertificateAuthorityInput(t *testing.T) {
	cases := map[string]struct {
		in  *v1alpha1.CertificateAuthorityParameters
		out *acmpca.CreateCertificateAuthorityInput
	}{
		"Filled_Input": {
			in: &v1alpha1.CertificateAuthorityParameters{
				SigningAlgorithm:           acmpca.SigningAlgorithmSha256withecdsa,
				KeyAlgorithm:               acmpca.KeyAlgorithmRsa2048,
				CommonName:                 aws.String(commonName),
				Country:                    aws.String(country),
				DistinguishedNameQualifier: aws.String(distinguishedNameQualifier),
				GenerationQualifier:        aws.String(generationQualifier),
				GivenName:                  aws.String(givenName),
				Initials:                   aws.String(initials),
				Locality:                   aws.String(locality),
				Organization:               aws.String(organization),
				OrganizationalUnit:         aws.String(organizationalUnit),
				Pseudonym:                  aws.String(pseudonym),
				SerialNumber:               aws.String(serialNumber),
				State:                      aws.String(state),
				Surname:                    aws.String(surname),
				Title:                      aws.String(title),
				IdempotencyToken:           aws.String(idempotencyToken),
				Type:                       acmpca.CertificateAuthorityTypeRoot,
				Tags: []v1alpha1.Tag{{
					Key:   "key1",
					Value: "value1",
				}},
				CustomCname:                    aws.String(customCname),
				RevocationConfigurationEnabled: aws.Bool(revocationConfigurationEnabled),
				S3BucketName:                   aws.String(s3BucketName),
				Status:                         acmpca.CertificateAuthorityStatusActive,
			},
			out: &acmpca.CreateCertificateAuthorityInput{
				IdempotencyToken: aws.String(idempotencyToken),
				CertificateAuthorityConfiguration: &acmpca.CertificateAuthorityConfiguration{
					SigningAlgorithm: acmpca.SigningAlgorithmSha256withecdsa,
					KeyAlgorithm:     acmpca.KeyAlgorithmRsa2048,
					Subject: &acmpca.ASN1Subject{
						CommonName:                 aws.String(commonName),
						Country:                    aws.String(country),
						DistinguishedNameQualifier: aws.String(distinguishedNameQualifier),
						GenerationQualifier:        aws.String(generationQualifier),
						GivenName:                  aws.String(givenName),
						Initials:                   aws.String(initials),
						Locality:                   aws.String(locality),
						Organization:               aws.String(organization),
						OrganizationalUnit:         aws.String(organizationalUnit),
						Pseudonym:                  aws.String(pseudonym),
						SerialNumber:               aws.String(serialNumber),
						State:                      aws.String(state),
						Surname:                    aws.String(surname),
						Title:                      aws.String(title),
					},
				},
				RevocationConfiguration: &acmpca.RevocationConfiguration{
					CrlConfiguration: &acmpca.CrlConfiguration{
						CustomCname:  aws.String(customCname),
						Enabled:      aws.Bool(revocationConfigurationEnabled),
						S3BucketName: aws.String(s3BucketName),
					},
				},
				CertificateAuthorityType: acmpca.CertificateAuthorityTypeRoot,
				Tags: []acmpca.Tag{{
					Key:   aws.String("key1"),
					Value: aws.String("value1"),
				}},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			r := GenerateCreateCertificateAuthorityInput(tc.in)
			if diff := cmp.Diff(r, tc.out); diff != "" {
				t.Errorf("GenerateCreateCertificateAuthorityInput(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGenerateCertificateAuthorityConfiguration(t *testing.T) {
	cases := map[string]struct {
		in  *v1alpha1.CertificateAuthorityParameters
		out *acmpca.CertificateAuthorityConfiguration
	}{
		"Filled_Input": {
			in: &v1alpha1.CertificateAuthorityParameters{
				SigningAlgorithm:           acmpca.SigningAlgorithmSha256withecdsa,
				KeyAlgorithm:               acmpca.KeyAlgorithmRsa2048,
				CommonName:                 aws.String(commonName),
				Country:                    aws.String(country),
				DistinguishedNameQualifier: aws.String(distinguishedNameQualifier),
				GenerationQualifier:        aws.String(generationQualifier),
				GivenName:                  aws.String(givenName),
				Initials:                   aws.String(initials),
				Locality:                   aws.String(locality),
				Organization:               aws.String(organization),
				OrganizationalUnit:         aws.String(organizationalUnit),
				Pseudonym:                  aws.String(pseudonym),
				SerialNumber:               aws.String(serialNumber),
				State:                      aws.String(state),
				Surname:                    aws.String(surname),
				Title:                      aws.String(title),
			},
			out: &acmpca.CertificateAuthorityConfiguration{
				SigningAlgorithm: acmpca.SigningAlgorithmSha256withecdsa,
				KeyAlgorithm:     acmpca.KeyAlgorithmRsa2048,
				Subject: &acmpca.ASN1Subject{
					CommonName:                 aws.String(commonName),
					Country:                    aws.String(country),
					DistinguishedNameQualifier: aws.String(distinguishedNameQualifier),
					GenerationQualifier:        aws.String(generationQualifier),
					GivenName:                  aws.String(givenName),
					Initials:                   aws.String(initials),
					Locality:                   aws.String(locality),
					Organization:               aws.String(organization),
					OrganizationalUnit:         aws.String(organizationalUnit),
					Pseudonym:                  aws.String(pseudonym),
					SerialNumber:               aws.String(serialNumber),
					State:                      aws.String(state),
					Surname:                    aws.String(surname),
					Title:                      aws.String(title),
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			r := GenerateCertificateAuthorityConfiguration(tc.in)
			if diff := cmp.Diff(r, tc.out); diff != "" {
				t.Errorf("GenerateCertificateAuthorityConfiguration(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGenerateUpdateCertificateAuthorityInput(t *testing.T) {
	cases := map[string]struct {
		in  *v1alpha1.CertificateAuthority
		out *acmpca.UpdateCertificateAuthorityInput
	}{
		"Filled_Input": {
			in: &v1alpha1.CertificateAuthority{
				Spec: v1alpha1.CertificateAuthoritySpec{
					ForProvider: v1alpha1.CertificateAuthorityParameters{
						CustomCname:                    aws.String(customCname),
						RevocationConfigurationEnabled: aws.Bool(revocationConfigurationEnabled),
						S3BucketName:                   aws.String(s3BucketName),
						Status:                         acmpca.CertificateAuthorityStatusActive,
					},
				},
				Status: v1alpha1.CertificateAuthorityStatus{
					AtProvider: v1alpha1.CertificateAuthorityExternalStatus{
						CertificateAuthorityArn: certificateAuthorityArn,
					},
				},
			},
			out: &acmpca.UpdateCertificateAuthorityInput{
				CertificateAuthorityArn: aws.String(certificateAuthorityArn),
				RevocationConfiguration: &acmpca.RevocationConfiguration{
					CrlConfiguration: &acmpca.CrlConfiguration{
						CustomCname:  aws.String(customCname),
						Enabled:      aws.Bool(revocationConfigurationEnabled),
						S3BucketName: aws.String(s3BucketName),
					},
				},
				Status: acmpca.CertificateAuthorityStatusActive,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			r := GenerateUpdateCertificateAuthorityInput(tc.in)
			if diff := cmp.Diff(r, tc.out); diff != "" {
				t.Errorf("GenerateUpdateCertificateAuthorityInput(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGenerateRevocationConfiguration(t *testing.T) {
	cases := map[string]struct {
		in  *v1alpha1.CertificateAuthorityParameters
		out *acmpca.RevocationConfiguration
	}{
		"Filled_Input": {
			in: &v1alpha1.CertificateAuthorityParameters{
				CustomCname:                    aws.String(customCname),
				RevocationConfigurationEnabled: aws.Bool(revocationConfigurationEnabled),
				S3BucketName:                   aws.String(s3BucketName),
			},
			out: &acmpca.RevocationConfiguration{
				CrlConfiguration: &acmpca.CrlConfiguration{
					CustomCname:  aws.String(customCname),
					Enabled:      aws.Bool(revocationConfigurationEnabled),
					S3BucketName: aws.String(s3BucketName),
				},
			},
		},
		"PartialFilled": {
			in: &v1alpha1.CertificateAuthorityParameters{
				RevocationConfigurationEnabled: aws.Bool(revocationConfigurationEnabledfalse),
			},
			out: &acmpca.RevocationConfiguration{
				CrlConfiguration: &acmpca.CrlConfiguration{
					Enabled: aws.Bool(revocationConfigurationEnabledfalse),
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			r := GenerateRevocationConfiguration(tc.in)
			if diff := cmp.Diff(r, tc.out); diff != "" {
				t.Errorf("GenerateRevocationConfiguration(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestLateInitializeCertificateAuthority(t *testing.T) {
	type args struct {
		spec *v1alpha1.CertificateAuthorityParameters
		in   *acmpca.CertificateAuthority
	}
	cases := map[string]struct {
		args args
		want *v1alpha1.CertificateAuthorityParameters
	}{
		"AllFilledNoDiff": {
			args: args{
				spec: &v1alpha1.CertificateAuthorityParameters{
					Type:         acmpca.CertificateAuthorityTypeRoot,
					SerialNumber: aws.String(serialNumber),
				},
				in: &acmpca.CertificateAuthority{
					Type:   acmpca.CertificateAuthorityTypeRoot,
					Status: acmpca.CertificateAuthorityStatusActive,
					Serial: aws.String(serialNumber),
					RevocationConfiguration: &acmpca.RevocationConfiguration{
						CrlConfiguration: &acmpca.CrlConfiguration{
							ExpirationInDays: nil,
						},
					},
				},
			},
			want: &v1alpha1.CertificateAuthorityParameters{
				Type:             acmpca.CertificateAuthorityTypeRoot,
				Status:           acmpca.CertificateAuthorityStatusActive,
				SerialNumber:     aws.String(serialNumber),
				ExpirationInDays: nil,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			LateInitializeCertificateAuthority(tc.args.spec, tc.args.in)
			if diff := cmp.Diff(tc.args.spec, tc.want); diff != "" {
				t.Errorf("LateInitializeCertificateAuthority(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestIsCertificateAuthorityUpToDate(t *testing.T) {
	type args struct {
		p    *v1alpha1.CertificateAuthority
		cd   acmpca.CertificateAuthority
		tags []acmpca.Tag
	}

	cases := map[string]struct {
		args args
		want bool
	}{
		"SameFields": {
			args: args{
				cd: acmpca.CertificateAuthority{
					RevocationConfiguration: &acmpca.RevocationConfiguration{
						CrlConfiguration: &acmpca.CrlConfiguration{
							CustomCname:  aws.String(customCname),
							S3BucketName: aws.String(s3BucketName),
						},
					},
				},
				p: &v1alpha1.CertificateAuthority{
					Spec: v1alpha1.CertificateAuthoritySpec{
						ForProvider: v1alpha1.CertificateAuthorityParameters{
							CustomCname:                       aws.String(customCname),
							S3BucketName:                      aws.String(s3BucketName),
							CertificateRenewalPermissionAllow: true,
							Tags: []v1alpha1.Tag{{
								Key:   "key1",
								Value: "value1",
							}},
						},
					},
					Status: v1alpha1.CertificateAuthorityStatus{
						AtProvider: v1alpha1.CertificateAuthorityExternalStatus{
							CertificateAuthorityArn: certificateAuthorityArn,
							RenewalPermission:       renewalPermission,
						},
					},
				},
				tags: []acmpca.Tag{{
					Key:   aws.String("key1"),
					Value: aws.String("value1"),
				}},
			},
			want: true,
		},
		"DifferntFields": {
			args: args{
				cd: acmpca.CertificateAuthority{
					RevocationConfiguration: &acmpca.RevocationConfiguration{
						CrlConfiguration: &acmpca.CrlConfiguration{
							CustomCname:  aws.String(customCname),
							S3BucketName: aws.String(s3BucketName),
						},
					},
				},
				p: &v1alpha1.CertificateAuthority{
					Spec: v1alpha1.CertificateAuthoritySpec{
						ForProvider: v1alpha1.CertificateAuthorityParameters{
							CustomCname:                       aws.String(customCname),
							S3BucketName:                      aws.String(s3BucketName),
							CertificateRenewalPermissionAllow: true,
						},
					},
					Status: v1alpha1.CertificateAuthorityStatus{
						AtProvider: v1alpha1.CertificateAuthorityExternalStatus{
							CertificateAuthorityArn: certificateAuthorityArn,
							RenewalPermission:       renewalPermission,
						},
					},
				},
				tags: []acmpca.Tag{{
					Key:   aws.String("key1"),
					Value: aws.String("value1"),
				}},
			},
			want: false,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := IsCertificateAuthorityUpToDate(tc.args.p, tc.args.cd, tc.args.tags)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("IsCertificateAuthorityUpToDate: -want, +got:\n%s", diff)
			}
		})
	}
}
