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
	"github.com/aws/smithy-go/document"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/crossplane-contrib/provider-aws/apis/acmpca/v1beta1"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
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
		in  *v1beta1.CertificateAuthorityParameters
		out *acmpca.CreateCertificateAuthorityInput
	}{
		"Filled_Input": {
			in: &v1beta1.CertificateAuthorityParameters{
				Type: types.CertificateAuthorityTypeRoot,
				CertificateAuthorityConfiguration: v1beta1.CertificateAuthorityConfiguration{
					SigningAlgorithm: types.SigningAlgorithmSha256withecdsa,
					KeyAlgorithm:     types.KeyAlgorithmRsa2048,
					Subject: v1beta1.Subject{
						CommonName:                 commonName,
						Country:                    country,
						DistinguishedNameQualifier: pointer.ToOrNilIfZeroValue(distinguishedNameQualifier),
						GenerationQualifier:        pointer.ToOrNilIfZeroValue(generationQualifier),
						GivenName:                  pointer.ToOrNilIfZeroValue(givenName),
						Initials:                   pointer.ToOrNilIfZeroValue(initials),
						Locality:                   locality,
						Organization:               organization,
						OrganizationalUnit:         organizationalUnit,
						Pseudonym:                  pointer.ToOrNilIfZeroValue(pseudonym),
						SerialNumber:               pointer.ToOrNilIfZeroValue(serialNumber),
						State:                      state,
						Surname:                    pointer.ToOrNilIfZeroValue(surname),
						Title:                      pointer.ToOrNilIfZeroValue(title),
					},
				},
				Tags: []v1beta1.Tag{{
					Key:   "key1",
					Value: "value1",
				}},
			},
			out: &acmpca.CreateCertificateAuthorityInput{
				CertificateAuthorityConfiguration: &types.CertificateAuthorityConfiguration{
					SigningAlgorithm: types.SigningAlgorithmSha256withecdsa,
					KeyAlgorithm:     types.KeyAlgorithmRsa2048,
					Subject: &types.ASN1Subject{
						CommonName:                 pointer.ToOrNilIfZeroValue(commonName),
						Country:                    pointer.ToOrNilIfZeroValue(country),
						DistinguishedNameQualifier: pointer.ToOrNilIfZeroValue(distinguishedNameQualifier),
						GenerationQualifier:        pointer.ToOrNilIfZeroValue(generationQualifier),
						GivenName:                  pointer.ToOrNilIfZeroValue(givenName),
						Initials:                   pointer.ToOrNilIfZeroValue(initials),
						Locality:                   pointer.ToOrNilIfZeroValue(locality),
						Organization:               pointer.ToOrNilIfZeroValue(organization),
						OrganizationalUnit:         pointer.ToOrNilIfZeroValue(organizationalUnit),
						Pseudonym:                  pointer.ToOrNilIfZeroValue(pseudonym),
						SerialNumber:               pointer.ToOrNilIfZeroValue(serialNumber),
						State:                      pointer.ToOrNilIfZeroValue(state),
						Surname:                    pointer.ToOrNilIfZeroValue(surname),
						Title:                      pointer.ToOrNilIfZeroValue(title),
					},
				},
				CertificateAuthorityType: types.CertificateAuthorityTypeRoot,
				Tags: []types.Tag{{
					Key:   pointer.ToOrNilIfZeroValue("key1"),
					Value: pointer.ToOrNilIfZeroValue("value1"),
				}},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			r := GenerateCreateCertificateAuthorityInput(tc.in)
			if diff := cmp.Diff(r, tc.out, cmpopts.IgnoreTypes(document.NoSerde{})); diff != "" {
				t.Errorf("GenerateCreateCertificateAuthorityInput(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGenerateCertificateAuthorityConfiguration(t *testing.T) {
	cases := map[string]struct {
		in  v1beta1.CertificateAuthorityConfiguration
		out *types.CertificateAuthorityConfiguration
	}{
		"Filled_Input": {
			in: v1beta1.CertificateAuthorityConfiguration{
				SigningAlgorithm: types.SigningAlgorithmSha256withecdsa,
				KeyAlgorithm:     types.KeyAlgorithmRsa2048,
				Subject: v1beta1.Subject{
					CommonName:                 commonName,
					Country:                    country,
					DistinguishedNameQualifier: pointer.ToOrNilIfZeroValue(distinguishedNameQualifier),
					GenerationQualifier:        pointer.ToOrNilIfZeroValue(generationQualifier),
					GivenName:                  pointer.ToOrNilIfZeroValue(givenName),
					Initials:                   pointer.ToOrNilIfZeroValue(initials),
					Locality:                   locality,
					Organization:               organization,
					OrganizationalUnit:         organizationalUnit,
					Pseudonym:                  pointer.ToOrNilIfZeroValue(pseudonym),
					SerialNumber:               pointer.ToOrNilIfZeroValue(serialNumber),
					State:                      state,
					Surname:                    pointer.ToOrNilIfZeroValue(surname),
					Title:                      pointer.ToOrNilIfZeroValue(title),
				},
			},
			out: &types.CertificateAuthorityConfiguration{
				SigningAlgorithm: types.SigningAlgorithmSha256withecdsa,
				KeyAlgorithm:     types.KeyAlgorithmRsa2048,
				Subject: &types.ASN1Subject{
					CommonName:                 pointer.ToOrNilIfZeroValue(commonName),
					Country:                    pointer.ToOrNilIfZeroValue(country),
					DistinguishedNameQualifier: pointer.ToOrNilIfZeroValue(distinguishedNameQualifier),
					GenerationQualifier:        pointer.ToOrNilIfZeroValue(generationQualifier),
					GivenName:                  pointer.ToOrNilIfZeroValue(givenName),
					Initials:                   pointer.ToOrNilIfZeroValue(initials),
					Locality:                   pointer.ToOrNilIfZeroValue(locality),
					Organization:               pointer.ToOrNilIfZeroValue(organization),
					OrganizationalUnit:         pointer.ToOrNilIfZeroValue(organizationalUnit),
					Pseudonym:                  pointer.ToOrNilIfZeroValue(pseudonym),
					SerialNumber:               pointer.ToOrNilIfZeroValue(serialNumber),
					State:                      pointer.ToOrNilIfZeroValue(state),
					Surname:                    pointer.ToOrNilIfZeroValue(surname),
					Title:                      pointer.ToOrNilIfZeroValue(title),
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			r := GenerateCertificateAuthorityConfiguration(tc.in)
			if diff := cmp.Diff(r, tc.out, cmpopts.IgnoreTypes(document.NoSerde{})); diff != "" {
				t.Errorf("GenerateCertificateAuthorityConfiguration(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGenerateRevocationConfiguration(t *testing.T) {
	cases := map[string]struct {
		in  *v1beta1.RevocationConfiguration
		out *types.RevocationConfiguration
	}{
		"Filled_Input": {
			in: &v1beta1.RevocationConfiguration{
				CustomCname:  pointer.ToOrNilIfZeroValue(customCname),
				Enabled:      revocationConfigurationEnabled,
				S3BucketName: pointer.ToOrNilIfZeroValue(s3BucketName),
			},
			out: &types.RevocationConfiguration{
				CrlConfiguration: &types.CrlConfiguration{
					CustomCname:  pointer.ToOrNilIfZeroValue(customCname),
					Enabled:      revocationConfigurationEnabled,
					S3BucketName: pointer.ToOrNilIfZeroValue(s3BucketName),
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			r := GenerateRevocationConfiguration(tc.in)
			if diff := cmp.Diff(r, tc.out, cmpopts.IgnoreTypes(document.NoSerde{})); diff != "" {
				t.Errorf("GenerateRevocationConfiguration(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestLateInitializeCertificateAuthority(t *testing.T) {

	status := "ACTIVE"

	type args struct {
		spec *v1beta1.CertificateAuthorityParameters
		in   *types.CertificateAuthority
	}
	cases := map[string]struct {
		args args
		want *v1beta1.CertificateAuthorityParameters
	}{
		"AllFilledNoDiff": {
			args: args{
				spec: &v1beta1.CertificateAuthorityParameters{
					Type: types.CertificateAuthorityTypeRoot,
				},
				in: &types.CertificateAuthority{
					Type:   types.CertificateAuthorityTypeRoot,
					Status: types.CertificateAuthorityStatus(status),
					CertificateAuthorityConfiguration: &types.CertificateAuthorityConfiguration{
						Subject: &types.ASN1Subject{
							SerialNumber: pointer.ToOrNilIfZeroValue(serialNumber),
						},
					},
					RevocationConfiguration: &types.RevocationConfiguration{
						CrlConfiguration: &types.CrlConfiguration{
							Enabled: false,
						},
					},
				},
			},
			want: &v1beta1.CertificateAuthorityParameters{
				Type:   types.CertificateAuthorityTypeRoot,
				Status: &status,
				CertificateAuthorityConfiguration: v1beta1.CertificateAuthorityConfiguration{
					Subject: v1beta1.Subject{
						SerialNumber: pointer.ToOrNilIfZeroValue(serialNumber),
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
		p    *v1beta1.CertificateAuthority
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
							CustomCname:  pointer.ToOrNilIfZeroValue(customCname),
							S3BucketName: pointer.ToOrNilIfZeroValue(s3BucketName),
							Enabled:      true,
						},
					},
					Status: types.CertificateAuthorityStatus(status),
				},
				p: &v1beta1.CertificateAuthority{
					Spec: v1beta1.CertificateAuthoritySpec{
						ForProvider: v1beta1.CertificateAuthorityParameters{
							RevocationConfiguration: &v1beta1.RevocationConfiguration{
								CustomCname:  pointer.ToOrNilIfZeroValue(customCname),
								S3BucketName: pointer.ToOrNilIfZeroValue(s3BucketName),
								Enabled:      true,
							},
							Tags: []v1beta1.Tag{{
								Key:   "key1",
								Value: "value1",
							}},
							Status: &status,
						},
					},
				},
				tags: []types.Tag{{
					Key:   pointer.ToOrNilIfZeroValue("key1"),
					Value: pointer.ToOrNilIfZeroValue("value1"),
				}},
			},
			want: true,
		},
		"DifferentFields": {
			args: args{
				cd: types.CertificateAuthority{
					RevocationConfiguration: &types.RevocationConfiguration{
						CrlConfiguration: &types.CrlConfiguration{
							CustomCname:  pointer.ToOrNilIfZeroValue(customCname),
							S3BucketName: pointer.ToOrNilIfZeroValue(s3BucketName),
						},
					},
				},
				p: &v1beta1.CertificateAuthority{
					Spec: v1beta1.CertificateAuthoritySpec{
						ForProvider: v1beta1.CertificateAuthorityParameters{
							RevocationConfiguration: &v1beta1.RevocationConfiguration{
								CustomCname:  pointer.ToOrNilIfZeroValue(customCname),
								S3BucketName: pointer.ToOrNilIfZeroValue(s3BucketName),
							},
						},
					},
				},
				tags: []types.Tag{{
					Key:   pointer.ToOrNilIfZeroValue("key1"),
					Value: pointer.ToOrNilIfZeroValue("value1"),
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
