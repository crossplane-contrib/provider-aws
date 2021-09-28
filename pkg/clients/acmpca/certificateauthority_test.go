/*
Copyright 2020 The Crossplane Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package acmpca

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/acmpca"
	"github.com/aws/aws-sdk-go-v2/service/acmpca/types"
	"github.com/google/go-cmp/cmp"

	"github.com/crossplane/provider-aws/apis/acmpca/v1alpha1"
	aws "github.com/crossplane/provider-aws/pkg/clients"
)

var (
	customCname                    = "somecustomname"
	revocationConfigurationEnabled = true
	s3BucketName                   = "somes3bucketname"
	commonName                     = "someCommonName"
	country                        = "someCountry"
	distinguishedNameQualifier     = "someDistinguishedNameQualifier"
	generationQualifier            = "somegenerationQualifier"
	givenName                      = "somegivenName"
	initials                       = "someinitials"
	locality                       = "somelocality"
	organization                   = "someorganization"
	organizationalUnit             = "someOrganizationalUnit"
	pseudonym                      = "somePseudonym"
	serialNumber                   = "someSerialNumber"
	state                          = "someState"
	surname                        = "someSurname"
	title                          = "someTitle"
)

func TestGenerateCreateCertificateAuthorityInput(t *testing.T) {
	cases := map[string]struct {
		in  *v1alpha1.CertificateAuthorityParameters
		out *acmpca.CreateCertificateAuthorityInput
	}{
		"Filled_Input": {
			in: &v1alpha1.CertificateAuthorityParameters{
				Type: types.CertificateAuthorityTypeRoot,
				CertificateAuthorityConfiguration: v1alpha1.CertificateAuthorityConfiguration{
					SigningAlgorithm: types.SigningAlgorithmSha256withecdsa,
					KeyAlgorithm:     types.KeyAlgorithmRsa2048,
					Subject: v1alpha1.Subject{
						CommonName:                 commonName,
						Country:                    country,
						DistinguishedNameQualifier: aws.String(distinguishedNameQualifier),
						GenerationQualifier:        aws.String(generationQualifier),
						GivenName:                  aws.String(givenName),
						Initials:                   aws.String(initials),
						Locality:                   locality,
						Organization:               organization,
						OrganizationalUnit:         organizationalUnit,
						Pseudonym:                  aws.String(pseudonym),
						SerialNumber:               aws.String(serialNumber),
						State:                      state,
						Surname:                    aws.String(surname),
						Title:                      aws.String(title),
					},
				},
				Tags: []v1alpha1.Tag{{
					Key:   "key1",
					Value: "value1",
				}},
			},
			out: &acmpca.CreateCertificateAuthorityInput{
				CertificateAuthorityConfiguration: &types.CertificateAuthorityConfiguration{
					SigningAlgorithm: types.SigningAlgorithmSha256withecdsa,
					KeyAlgorithm:     types.KeyAlgorithmRsa2048,
					Subject: &types.ASN1Subject{
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
				CertificateAuthorityType: types.CertificateAuthorityTypeRoot,
				Tags: []types.Tag{{
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
		in  v1alpha1.CertificateAuthorityConfiguration
		out *types.CertificateAuthorityConfiguration
	}{
		"Filled_Input": {
			in: v1alpha1.CertificateAuthorityConfiguration{
				SigningAlgorithm: types.SigningAlgorithmSha256withecdsa,
				KeyAlgorithm:     types.KeyAlgorithmRsa2048,
				Subject: v1alpha1.Subject{
					CommonName:                 commonName,
					Country:                    country,
					DistinguishedNameQualifier: aws.String(distinguishedNameQualifier),
					GenerationQualifier:        aws.String(generationQualifier),
					GivenName:                  aws.String(givenName),
					Initials:                   aws.String(initials),
					Locality:                   locality,
					Organization:               organization,
					OrganizationalUnit:         organizationalUnit,
					Pseudonym:                  aws.String(pseudonym),
					SerialNumber:               aws.String(serialNumber),
					State:                      state,
					Surname:                    aws.String(surname),
					Title:                      aws.String(title),
				},
			},
			out: &types.CertificateAuthorityConfiguration{
				SigningAlgorithm: types.SigningAlgorithmSha256withecdsa,
				KeyAlgorithm:     types.KeyAlgorithmRsa2048,
				Subject: &types.ASN1Subject{
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

func TestGenerateRevocationConfiguration(t *testing.T) {
	cases := map[string]struct {
		in  *v1alpha1.RevocationConfiguration
		out *types.RevocationConfiguration
	}{
		"Filled_Input": {
			in: &v1alpha1.RevocationConfiguration{
				CustomCname:  aws.String(customCname),
				Enabled:      revocationConfigurationEnabled,
				S3BucketName: aws.String(s3BucketName),
			},
			out: &types.RevocationConfiguration{
				CrlConfiguration: &types.CrlConfiguration{
					CustomCname:  aws.String(customCname),
					Enabled:      revocationConfigurationEnabled,
					S3BucketName: aws.String(s3BucketName),
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

	status := "ACTIVE"

	type args struct {
		spec *v1alpha1.CertificateAuthorityParameters
		in   *types.CertificateAuthority
	}
	cases := map[string]struct {
		args args
		want *v1alpha1.CertificateAuthorityParameters
	}{
		"AllFilledNoDiff": {
			args: args{
				spec: &v1alpha1.CertificateAuthorityParameters{
					Type: types.CertificateAuthorityTypeRoot,
				},
				in: &types.CertificateAuthority{
					Type:   types.CertificateAuthorityTypeRoot,
					Status: types.CertificateAuthorityStatus(status),
					CertificateAuthorityConfiguration: &types.CertificateAuthorityConfiguration{
						Subject: &types.ASN1Subject{
							SerialNumber: aws.String(serialNumber),
						},
					},
					RevocationConfiguration: &types.RevocationConfiguration{
						CrlConfiguration: &types.CrlConfiguration{
							Enabled: false,
						},
					},
				},
			},
			want: &v1alpha1.CertificateAuthorityParameters{
				Type:   types.CertificateAuthorityTypeRoot,
				Status: &status,
				CertificateAuthorityConfiguration: v1alpha1.CertificateAuthorityConfiguration{
					Subject: v1alpha1.Subject{
						SerialNumber: aws.String(serialNumber),
					},
				},
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
	status := "ACTIVE"
	type args struct {
		p    *v1alpha1.CertificateAuthority
		cd   types.CertificateAuthority
		tags []types.Tag
	}

	cases := map[string]struct {
		args args
		want bool
	}{
		"SameFields": {
			args: args{
				cd: types.CertificateAuthority{
					RevocationConfiguration: &types.RevocationConfiguration{
						CrlConfiguration: &types.CrlConfiguration{
							CustomCname:  aws.String(customCname),
							S3BucketName: aws.String(s3BucketName),
							Enabled:      true,
						},
					},
					Status: types.CertificateAuthorityStatus(status),
				},
				p: &v1alpha1.CertificateAuthority{
					Spec: v1alpha1.CertificateAuthoritySpec{
						ForProvider: v1alpha1.CertificateAuthorityParameters{
							RevocationConfiguration: &v1alpha1.RevocationConfiguration{
								CustomCname:  aws.String(customCname),
								S3BucketName: aws.String(s3BucketName),
								Enabled:      true,
							},
							Tags: []v1alpha1.Tag{{
								Key:   "key1",
								Value: "value1",
							}},
							Status: &status,
						},
					},
				},
				tags: []types.Tag{{
					Key:   aws.String("key1"),
					Value: aws.String("value1"),
				}},
			},
			want: true,
		},
		"DifferentFields": {
			args: args{
				cd: types.CertificateAuthority{
					RevocationConfiguration: &types.RevocationConfiguration{
						CrlConfiguration: &types.CrlConfiguration{
							CustomCname:  aws.String(customCname),
							S3BucketName: aws.String(s3BucketName),
						},
					},
				},
				p: &v1alpha1.CertificateAuthority{
					Spec: v1alpha1.CertificateAuthoritySpec{
						ForProvider: v1alpha1.CertificateAuthorityParameters{
							RevocationConfiguration: &v1alpha1.RevocationConfiguration{
								CustomCname:  aws.String(customCname),
								S3BucketName: aws.String(s3BucketName),
							},
						},
					},
				},
				tags: []types.Tag{{
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
