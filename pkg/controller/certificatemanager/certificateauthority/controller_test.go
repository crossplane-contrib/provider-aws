/*
Copyright 2019 The Crossplane Authors.

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

package certificateauthority

import (
	"context"
	"net/http"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/awserr"
	awsacmpca "github.com/aws/aws-sdk-go-v2/service/acmpca"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	corev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/test"

	v1alpha1 "github.com/crossplane/provider-aws/apis/certificatemanager/v1alpha1"
	acmpca "github.com/crossplane/provider-aws/pkg/clients/certificatemanager/certificateauthority"
	"github.com/crossplane/provider-aws/pkg/clients/certificatemanager/certificateauthority/fake"
)

const (
	providerName = "aws-creds"
	testRegion   = "us-west-2"
)

var (
	// an arbitrary managed resource
	unexpecedItem              resource.Managed
	certificateAuthorityArn    = "someauthorityarn"
	nextToken                  = "someNextToken"
	serialNumber               = "someserialNumber"
	customCname                = "soemcustomname"
	s3BucketName               = "somes3bucketname"
	commonName                 = "someCommonName"
	country                    = "someCountry"
	distinguishedNameQualifier = "someDistinguishedNameQualifier"
	generationQualifier        = "somegenerationQualifier"
	givenName                  = "somegivenName"
	initials                   = "someinitials"
	locality                   = "somelocality"
	organization               = "someorganization"
	organizationalUnit         = "someOrganizationalUnit"
	pseudonym                  = "somePseudonym"
	state                      = "someState"
	surname                    = "someSurname"
	title                      = "someTitle"

	errBoom = errors.New("boom")
)

type args struct {
	acmpca acmpca.Client
	cr     resource.Managed
}

type certificateAuthorityModifier func(*v1alpha1.CertificateAuthority)

func withConditions(c ...corev1alpha1.Condition) certificateAuthorityModifier {
	return func(r *v1alpha1.CertificateAuthority) { r.Status.ConditionedStatus.Conditions = c }
}

func withCertificateAuthorityArn() certificateAuthorityModifier {
	return func(r *v1alpha1.CertificateAuthority) {
		r.Status.AtProvider.CertificateAuthorityArn = certificateAuthorityArn
		meta.SetExternalName(r, certificateAuthorityArn)
	}
}

func withCertificateAuthorityType() certificateAuthorityModifier {
	return func(r *v1alpha1.CertificateAuthority) {
		r.Spec.ForProvider.Type = awsacmpca.CertificateAuthorityTypeRoot
		r.Status.AtProvider.CertificateAuthorityArn = certificateAuthorityArn
		meta.SetExternalName(r, certificateAuthorityArn)
	}
}

func certificateAuthority(m ...certificateAuthorityModifier) *v1alpha1.CertificateAuthority {
	cr := &v1alpha1.CertificateAuthority{
		Spec: v1alpha1.CertificateAuthoritySpec{
			ResourceSpec: corev1alpha1.ResourceSpec{
				ProviderReference: &corev1.ObjectReference{Name: providerName},
			},
		},
	}
	meta.SetExternalName(cr, certificateAuthorityArn)
	for _, f := range m {
		f(cr)
	}
	return cr
}

func TestConnect(t *testing.T) {

	type args struct {
		newClientFn func(*aws.Config) (acmpca.Client, error)
		awsConfigFn func(context.Context, client.Reader, *corev1.ObjectReference) (*aws.Config, error)
		cr          resource.Managed
	}
	type want struct {
		err error
	}

	cases := map[string]struct {
		args
		want
	}{
		"ValidInput": {
			args: args{
				newClientFn: func(config *aws.Config) (acmpca.Client, error) {
					if diff := cmp.Diff(testRegion, config.Region); diff != "" {
						t.Errorf("r: -want, +got:\n%s", diff)
					}
					return nil, nil
				},
				awsConfigFn: func(_ context.Context, _ client.Reader, p *corev1.ObjectReference) (*aws.Config, error) {
					if diff := cmp.Diff(providerName, p.Name); diff != "" {
						t.Errorf("r: -want, +got:\n%s", diff)
					}
					return &aws.Config{Region: testRegion}, nil
				},
				cr: certificateAuthority(),
			},
		},
		"InValidInput": {
			args: args{
				cr: unexpecedItem,
			},
			want: want{
				err: errors.New(errUnexpectedObject),
			},
		},
		"ProviderFailure": {
			args: args{
				newClientFn: func(config *aws.Config) (acmpca.Client, error) {
					if diff := cmp.Diff(testRegion, config.Region); diff != "" {
						t.Errorf("r: -want, +got:\n%s", diff)
					}
					return nil, errBoom
				},
				awsConfigFn: func(_ context.Context, _ client.Reader, p *corev1.ObjectReference) (*aws.Config, error) {
					if diff := cmp.Diff(providerName, p.Name); diff != "" {
						t.Errorf("r: -want, +got:\n%s", diff)
					}
					return &aws.Config{Region: testRegion}, nil
				},
				cr: certificateAuthority(),
			},
			want: want{
				err: errors.Wrap(errBoom, errClient),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			c := &connector{newClientFn: tc.newClientFn, awsConfigFn: tc.awsConfigFn}
			_, err := c.Connect(context.Background(), tc.args.cr)
			if diff := cmp.Diff(tc.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestObserve(t *testing.T) {

	type want struct {
		cr     resource.Managed
		result managed.ExternalObservation
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"VaildInput": {
			args: args{
				acmpca: &fake.MockCertificateAuthorityClient{
					MockDescribeCertificateAuthorityRequest: func(*awsacmpca.DescribeCertificateAuthorityInput) awsacmpca.DescribeCertificateAuthorityRequest {
						return awsacmpca.DescribeCertificateAuthorityRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Data: &awsacmpca.DescribeCertificateAuthorityOutput{
								CertificateAuthority: &awsacmpca.CertificateAuthority{
									Arn:  aws.String(certificateAuthorityArn),
									Type: awsacmpca.CertificateAuthorityTypeRoot,
									RevocationConfiguration: &awsacmpca.RevocationConfiguration{
										CrlConfiguration: &awsacmpca.CrlConfiguration{
											CustomCname:  aws.String(customCname),
											S3BucketName: aws.String(s3BucketName),
										},
									},
									CertificateAuthorityConfiguration: &awsacmpca.CertificateAuthorityConfiguration{
										SigningAlgorithm: awsacmpca.SigningAlgorithmSha256withecdsa,
										KeyAlgorithm:     awsacmpca.KeyAlgorithmRsa2048,
										Subject: &awsacmpca.ASN1Subject{
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
							}},
						}
					},
					MockListTagsRequest: func(input *awsacmpca.ListTagsInput) awsacmpca.ListTagsRequest {
						return awsacmpca.ListTagsRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Data: &awsacmpca.ListTagsOutput{
								NextToken: aws.String(nextToken),
								Tags:      []awsacmpca.Tag{{}},
							}},
						}
					},
				},
				cr: certificateAuthority(),
			},
			want: want{
				cr: certificateAuthority(withCertificateAuthorityType(), withConditions(corev1alpha1.Available())),
				result: managed.ExternalObservation{
					ResourceExists: true,
				},
			},
		},
		"InValidInput": {
			args: args{
				cr: unexpecedItem,
			},
			want: want{
				cr:  unexpecedItem,
				err: errors.New(errUnexpectedObject),
			},
		},
		"ClientError": {
			args: args{
				acmpca: &fake.MockCertificateAuthorityClient{
					MockDescribeCertificateAuthorityRequest: func(*awsacmpca.DescribeCertificateAuthorityInput) awsacmpca.DescribeCertificateAuthorityRequest {
						return awsacmpca.DescribeCertificateAuthorityRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: errBoom},
						}
					},
				},
				cr: certificateAuthority(withCertificateAuthorityArn()),
			},
			want: want{
				cr:  certificateAuthority(withCertificateAuthorityArn()),
				err: errors.Wrap(errBoom, errGet),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{
				client: tc.acmpca,
				kube: &test.MockClient{
					MockUpdate: test.NewMockUpdateFn(nil),
				},
			}
			o, err := e.Observe(context.Background(), tc.args.cr)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}

			if diff := cmp.Diff(tc.want.cr, tc.args.cr, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.result, o); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestCreate(t *testing.T) {

	type want struct {
		cr     resource.Managed
		result managed.ExternalCreation
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"VaildInput": {
			args: args{
				acmpca: &fake.MockCertificateAuthorityClient{
					MockCreateCertificateAuthorityRequest: func(input *awsacmpca.CreateCertificateAuthorityInput) awsacmpca.CreateCertificateAuthorityRequest {
						return awsacmpca.CreateCertificateAuthorityRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Data: &awsacmpca.CreateCertificateAuthorityOutput{
								CertificateAuthorityArn: aws.String(certificateAuthorityArn),
							}},
						}
					},
					MockCreatePermissionRequest: func(input *awsacmpca.CreatePermissionInput) awsacmpca.CreatePermissionRequest {
						return awsacmpca.CreatePermissionRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Data: &awsacmpca.CreatePermissionOutput{}},
						}
					},
				},
				cr: certificateAuthority(withCertificateAuthorityArn()),
			},
			want: want{
				cr: certificateAuthority(withCertificateAuthorityArn(), withConditions(corev1alpha1.Creating())),
			},
		},
		"InValidInput": {
			args: args{
				cr: unexpecedItem,
			},
			want: want{
				cr:  unexpecedItem,
				err: errors.New(errUnexpectedObject),
			},
		},
		"ClientError": {
			args: args{
				acmpca: &fake.MockCertificateAuthorityClient{
					MockCreateCertificateAuthorityRequest: func(input *awsacmpca.CreateCertificateAuthorityInput) awsacmpca.CreateCertificateAuthorityRequest {
						return awsacmpca.CreateCertificateAuthorityRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: errBoom},
						}
					},
					MockCreatePermissionRequest: func(input *awsacmpca.CreatePermissionInput) awsacmpca.CreatePermissionRequest {
						return awsacmpca.CreatePermissionRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: errBoom},
						}
					},
				},
				cr: certificateAuthority(),
			},
			want: want{
				cr:  certificateAuthority(withConditions(corev1alpha1.Creating())),
				err: errors.Wrap(errBoom, errCreate),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{
				client: tc.acmpca,
				kube: &test.MockClient{
					MockUpdate: test.NewMockUpdateFn(nil),
				},
			}
			o, err := e.Create(context.Background(), tc.args.cr)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.cr, tc.args.cr, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.result, o); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestUpdate(t *testing.T) {

	type want struct {
		cr     resource.Managed
		result managed.ExternalUpdate
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"VaildInput": {
			args: args{
				acmpca: &fake.MockCertificateAuthorityClient{
					MockDeletePermissionRequest: func(*awsacmpca.DeletePermissionInput) awsacmpca.DeletePermissionRequest {
						return awsacmpca.DeletePermissionRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Data: &awsacmpca.DeletePermissionOutput{}},
						}
					},
					MockDescribeCertificateAuthorityRequest: func(*awsacmpca.DescribeCertificateAuthorityInput) awsacmpca.DescribeCertificateAuthorityRequest {
						return awsacmpca.DescribeCertificateAuthorityRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Data: &awsacmpca.DescribeCertificateAuthorityOutput{
								CertificateAuthority: &awsacmpca.CertificateAuthority{
									Type:   awsacmpca.CertificateAuthorityTypeRoot,
									Status: awsacmpca.CertificateAuthorityStatusActive,
								},
							}},
						}
					},
					MockUpdateCertificateAuthorityRequest: func(*awsacmpca.UpdateCertificateAuthorityInput) awsacmpca.UpdateCertificateAuthorityRequest {
						return awsacmpca.UpdateCertificateAuthorityRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Data: &awsacmpca.UpdateCertificateAuthorityOutput{}},
						}
					},
					MockListTagsRequest: func(input *awsacmpca.ListTagsInput) awsacmpca.ListTagsRequest {
						return awsacmpca.ListTagsRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Data: &awsacmpca.ListTagsOutput{
								NextToken: aws.String(nextToken),
								Tags:      []awsacmpca.Tag{{}},
							}},
						}
					},
					MockUntagCertificateAuthorityRequest: func(input *awsacmpca.UntagCertificateAuthorityInput) awsacmpca.UntagCertificateAuthorityRequest {
						return awsacmpca.UntagCertificateAuthorityRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Data: &awsacmpca.UntagCertificateAuthorityOutput{}},
						}
					},
					MockTagCertificateAuthorityRequest: func(input *awsacmpca.TagCertificateAuthorityInput) awsacmpca.TagCertificateAuthorityRequest {
						return awsacmpca.TagCertificateAuthorityRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Data: &awsacmpca.TagCertificateAuthorityOutput{}},
						}
					},
					MockCreatePermissionRequest: func(input *awsacmpca.CreatePermissionInput) awsacmpca.CreatePermissionRequest {
						return awsacmpca.CreatePermissionRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Data: &awsacmpca.CreatePermissionOutput{}},
						}
					},
				},
				cr: certificateAuthority(),
			},
			want: want{
				cr: certificateAuthority(),
			},
		},
		"InValidInput": {
			args: args{
				cr: unexpecedItem,
			},
			want: want{
				cr:  unexpecedItem,
				err: errors.New(errUnexpectedObject),
			},
		},
		"ClientUpdateCertificateDescribeError": {
			args: args{
				acmpca: &fake.MockCertificateAuthorityClient{
					MockDeletePermissionRequest: func(*awsacmpca.DeletePermissionInput) awsacmpca.DeletePermissionRequest {
						return awsacmpca.DeletePermissionRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: errBoom},
						}
					},
					MockDescribeCertificateAuthorityRequest: func(*awsacmpca.DescribeCertificateAuthorityInput) awsacmpca.DescribeCertificateAuthorityRequest {
						return awsacmpca.DescribeCertificateAuthorityRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: errBoom},
						}
					},
					MockUpdateCertificateAuthorityRequest: func(*awsacmpca.UpdateCertificateAuthorityInput) awsacmpca.UpdateCertificateAuthorityRequest {
						return awsacmpca.UpdateCertificateAuthorityRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: errBoom},
						}
					},
					MockListTagsRequest: func(input *awsacmpca.ListTagsInput) awsacmpca.ListTagsRequest {
						return awsacmpca.ListTagsRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: errBoom},
						}
					},
					MockUntagCertificateAuthorityRequest: func(input *awsacmpca.UntagCertificateAuthorityInput) awsacmpca.UntagCertificateAuthorityRequest {
						return awsacmpca.UntagCertificateAuthorityRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Data: &awsacmpca.UntagCertificateAuthorityOutput{}},
						}
					},
					MockTagCertificateAuthorityRequest: func(input *awsacmpca.TagCertificateAuthorityInput) awsacmpca.TagCertificateAuthorityRequest {
						return awsacmpca.TagCertificateAuthorityRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Data: &awsacmpca.TagCertificateAuthorityOutput{}},
						}
					},
					MockCreatePermissionRequest: func(input *awsacmpca.CreatePermissionInput) awsacmpca.CreatePermissionRequest {
						return awsacmpca.CreatePermissionRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Data: &awsacmpca.CreatePermissionOutput{}},
						}
					},
				},
				cr: certificateAuthority(withCertificateAuthorityArn()),
			},
			want: want{
				cr:  certificateAuthority(withCertificateAuthorityArn()),
				err: errors.Wrap(errBoom, errCertificateAuthority),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.acmpca}
			o, err := e.Update(context.Background(), tc.args.cr)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.cr, tc.args.cr, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.result, o); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestDelete(t *testing.T) {

	type want struct {
		cr  resource.Managed
		err error
	}

	cases := map[string]struct {
		args
		want
	}{
		"VaildInput": {
			args: args{
				acmpca: &fake.MockCertificateAuthorityClient{
					MockDeleteCertificateAuthorityRequest: func(*awsacmpca.DeleteCertificateAuthorityInput) awsacmpca.DeleteCertificateAuthorityRequest {
						return awsacmpca.DeleteCertificateAuthorityRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Data: &awsacmpca.DeleteCertificateAuthorityOutput{}},
						}
					},
					MockDeletePermissionRequest: func(*awsacmpca.DeletePermissionInput) awsacmpca.DeletePermissionRequest {
						return awsacmpca.DeletePermissionRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Data: &awsacmpca.DeletePermissionOutput{}},
						}
					},
					MockDescribeCertificateAuthorityRequest: func(*awsacmpca.DescribeCertificateAuthorityInput) awsacmpca.DescribeCertificateAuthorityRequest {
						return awsacmpca.DescribeCertificateAuthorityRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Data: &awsacmpca.DescribeCertificateAuthorityOutput{
								CertificateAuthority: &awsacmpca.CertificateAuthority{
									Type:   awsacmpca.CertificateAuthorityTypeRoot,
									Status: awsacmpca.CertificateAuthorityStatusActive,
								},
							}},
						}
					},
					MockUpdateCertificateAuthorityRequest: func(*awsacmpca.UpdateCertificateAuthorityInput) awsacmpca.UpdateCertificateAuthorityRequest {
						return awsacmpca.UpdateCertificateAuthorityRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Data: &awsacmpca.UpdateCertificateAuthorityOutput{}},
						}
					},
				},
				cr: certificateAuthority(withCertificateAuthorityArn()),
			},
			want: want{
				cr: certificateAuthority(withCertificateAuthorityArn(), withConditions(corev1alpha1.Deleting())),
			},
		},
		"InValidInput": {
			args: args{
				cr: unexpecedItem,
			},
			want: want{
				cr:  unexpecedItem,
				err: errors.New(errUnexpectedObject),
			},
		},
		"ClientError": {
			args: args{
				acmpca: &fake.MockCertificateAuthorityClient{
					MockDeleteCertificateAuthorityRequest: func(*awsacmpca.DeleteCertificateAuthorityInput) awsacmpca.DeleteCertificateAuthorityRequest {
						return awsacmpca.DeleteCertificateAuthorityRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: errBoom},
						}
					},
					MockDescribeCertificateAuthorityRequest: func(*awsacmpca.DescribeCertificateAuthorityInput) awsacmpca.DescribeCertificateAuthorityRequest {
						return awsacmpca.DescribeCertificateAuthorityRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Data: &awsacmpca.DescribeCertificateAuthorityOutput{
								CertificateAuthority: &awsacmpca.CertificateAuthority{
									Type:   awsacmpca.CertificateAuthorityTypeRoot,
									Status: awsacmpca.CertificateAuthorityStatusActive,
								},
							}},
						}
					},
					MockUpdateCertificateAuthorityRequest: func(*awsacmpca.UpdateCertificateAuthorityInput) awsacmpca.UpdateCertificateAuthorityRequest {
						return awsacmpca.UpdateCertificateAuthorityRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Data: &awsacmpca.UpdateCertificateAuthorityOutput{}},
						}
					},
				},
				cr: certificateAuthority(),
			},
			want: want{
				cr:  certificateAuthority(withConditions(corev1alpha1.Deleting())),
				err: errors.Wrap(errBoom, errDelete),
			},
		},
		"ResourceDoesNotExist": {
			args: args{
				acmpca: &fake.MockCertificateAuthorityClient{
					MockDeleteCertificateAuthorityRequest: func(*awsacmpca.DeleteCertificateAuthorityInput) awsacmpca.DeleteCertificateAuthorityRequest {
						return awsacmpca.DeleteCertificateAuthorityRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: awserr.New(awsacmpca.ErrCodeResourceNotFoundException, "", nil)},
						}
					},
					MockDescribeCertificateAuthorityRequest: func(*awsacmpca.DescribeCertificateAuthorityInput) awsacmpca.DescribeCertificateAuthorityRequest {
						return awsacmpca.DescribeCertificateAuthorityRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: awserr.New(awsacmpca.ErrCodeResourceNotFoundException, "", nil)},
						}
					},
					MockUpdateCertificateAuthorityRequest: func(*awsacmpca.UpdateCertificateAuthorityInput) awsacmpca.UpdateCertificateAuthorityRequest {
						return awsacmpca.UpdateCertificateAuthorityRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: awserr.New(awsacmpca.ErrCodeResourceNotFoundException, "", nil)},
						}
					},
				},
				cr: certificateAuthority(),
			},
			want: want{
				cr:  certificateAuthority(withConditions(corev1alpha1.Deleting())),
				err: errors.Wrap(awserr.New(awsacmpca.ErrCodeResourceNotFoundException, "", nil), errDelete),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.acmpca}
			err := e.Delete(context.Background(), tc.args.cr)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.cr, tc.args.cr, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}
