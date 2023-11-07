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

package certificateauthority

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsacmpca "github.com/aws/aws-sdk-go-v2/service/acmpca"
	awsacmpcatypes "github.com/aws/aws-sdk-go-v2/service/acmpca/types"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"

	"github.com/crossplane-contrib/provider-aws/apis/acmpca/v1beta1"
	acmpca "github.com/crossplane-contrib/provider-aws/pkg/clients/acmpca"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/acmpca/fake"
	errorutils "github.com/crossplane-contrib/provider-aws/pkg/utils/errors"
)

var (
	// an arbitrary managed resource
	unexpectedItem             resource.Managed
	certificateAuthorityArn    = "someauthorityarn"
	nextToken                  = "someNextToken"
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

type certificateAuthorityModifier func(*v1beta1.CertificateAuthority)

func withConditions(c ...xpv1.Condition) certificateAuthorityModifier {
	return func(r *v1beta1.CertificateAuthority) { r.Status.ConditionedStatus.Conditions = c }
}

func withCertificateAuthorityArn() certificateAuthorityModifier {
	return func(r *v1beta1.CertificateAuthority) {
		r.Status.AtProvider.CertificateAuthorityARN = certificateAuthorityArn
		meta.SetExternalName(r, certificateAuthorityArn)
	}
}

func withCertificateAuthorityType() certificateAuthorityModifier {
	return func(r *v1beta1.CertificateAuthority) {
		r.Spec.ForProvider.Type = awsacmpcatypes.CertificateAuthorityTypeRoot
		r.Status.AtProvider.CertificateAuthorityARN = certificateAuthorityArn
		meta.SetExternalName(r, certificateAuthorityArn)
	}
}

func withCertificateAuthorityAtProviderStatus(s string) certificateAuthorityModifier {
	return func(r *v1beta1.CertificateAuthority) {
		r.Status.AtProvider.Status = s
	}
}

func withCertificateAuthorityStatus() certificateAuthorityModifier {
	status := "ACTIVE"

	return func(r *v1beta1.CertificateAuthority) {
		r.Spec.ForProvider.Status = &status
		r.Status.AtProvider.CertificateAuthorityARN = certificateAuthorityArn
		meta.SetExternalName(r, certificateAuthorityArn)
	}
}

func certificateAuthority(m ...certificateAuthorityModifier) *v1beta1.CertificateAuthority {
	cr := &v1beta1.CertificateAuthority{}
	meta.SetExternalName(cr, certificateAuthorityArn)
	for _, f := range m {
		f(cr)
	}
	return cr
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
		"ValidInput": {
			args: args{
				acmpca: &fake.MockCertificateAuthorityClient{
					MockDescribeCertificateAuthority: func(ctx context.Context, input *awsacmpca.DescribeCertificateAuthorityInput, opts []func(*awsacmpca.Options)) (*awsacmpca.DescribeCertificateAuthorityOutput, error) {
						return &awsacmpca.DescribeCertificateAuthorityOutput{
							CertificateAuthority: &awsacmpcatypes.CertificateAuthority{
								Arn:    aws.String(certificateAuthorityArn),
								Type:   awsacmpcatypes.CertificateAuthorityTypeRoot,
								Status: awsacmpcatypes.CertificateAuthorityStatusActive,
								RevocationConfiguration: &awsacmpcatypes.RevocationConfiguration{
									CrlConfiguration: &awsacmpcatypes.CrlConfiguration{
										Enabled: false,
									},
								},
								CertificateAuthorityConfiguration: &awsacmpcatypes.CertificateAuthorityConfiguration{
									SigningAlgorithm: awsacmpcatypes.SigningAlgorithmSha256withecdsa,
									KeyAlgorithm:     awsacmpcatypes.KeyAlgorithmRsa2048,
									Subject: &awsacmpcatypes.ASN1Subject{
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
										State:                      aws.String(state),
										Surname:                    aws.String(surname),
										Title:                      aws.String(title),
									},
								},
							},
						}, nil
					},
					MockListTags: func(ctx context.Context, input *awsacmpca.ListTagsInput, opts []func(*awsacmpca.Options)) (*awsacmpca.ListTagsOutput, error) {
						return &awsacmpca.ListTagsOutput{
							NextToken: aws.String(nextToken),
							Tags:      []awsacmpcatypes.Tag{{}},
						}, nil
					},
				},
				cr: certificateAuthority(),
			},
			want: want{
				cr: certificateAuthority(withCertificateAuthorityType(), withCertificateAuthorityStatus(), withCertificateAuthorityAtProviderStatus("ACTIVE"), withConditions(xpv1.Available())),
				result: managed.ExternalObservation{
					ResourceExists: true,
				},
			},
		},
		"InValidInput": {
			args: args{
				cr: unexpectedItem,
			},
			want: want{
				cr:  unexpectedItem,
				err: errors.New(errUnexpectedObject),
			},
		},
		"ClientError": {
			args: args{
				acmpca: &fake.MockCertificateAuthorityClient{
					MockDescribeCertificateAuthority: func(ctx context.Context, input *awsacmpca.DescribeCertificateAuthorityInput, opts []func(*awsacmpca.Options)) (*awsacmpca.DescribeCertificateAuthorityOutput, error) {
						return nil, errBoom
					},
				},
				cr: certificateAuthority(withCertificateAuthorityArn()),
			},
			want: want{
				cr:  certificateAuthority(withCertificateAuthorityArn()),
				err: errorutils.Wrap(errBoom, errGet),
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
		"ValidInput": {
			args: args{
				acmpca: &fake.MockCertificateAuthorityClient{
					MockCreateCertificateAuthority: func(ctx context.Context, input *awsacmpca.CreateCertificateAuthorityInput, opts []func(*awsacmpca.Options)) (*awsacmpca.CreateCertificateAuthorityOutput, error) {
						return &awsacmpca.CreateCertificateAuthorityOutput{
							CertificateAuthorityArn: aws.String(certificateAuthorityArn),
						}, nil
					},
					MockCreatePermission: func(ctx context.Context, input *awsacmpca.CreatePermissionInput, opts []func(*awsacmpca.Options)) (*awsacmpca.CreatePermissionOutput, error) {
						return &awsacmpca.CreatePermissionOutput{}, nil
					},
				},
				cr: certificateAuthority(withCertificateAuthorityArn()),
			},
			want: want{
				cr:     certificateAuthority(withCertificateAuthorityArn()),
				result: managed.ExternalCreation{},
			},
		},
		"InValidInput": {
			args: args{
				cr: unexpectedItem,
			},
			want: want{
				cr:  unexpectedItem,
				err: errors.New(errUnexpectedObject),
			},
		},
		"ClientError": {
			args: args{
				acmpca: &fake.MockCertificateAuthorityClient{
					MockCreateCertificateAuthority: func(ctx context.Context, input *awsacmpca.CreateCertificateAuthorityInput, opts []func(*awsacmpca.Options)) (*awsacmpca.CreateCertificateAuthorityOutput, error) {
						return nil, errBoom
					},
					MockCreatePermission: func(ctx context.Context, input *awsacmpca.CreatePermissionInput, opts []func(*awsacmpca.Options)) (*awsacmpca.CreatePermissionOutput, error) {
						return nil, errBoom
					},
				},
				cr: certificateAuthority(),
			},
			want: want{
				cr:  certificateAuthority(),
				err: errorutils.Wrap(errBoom, errCreate),
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
		"ValidInput": {
			args: args{
				acmpca: &fake.MockCertificateAuthorityClient{
					MockDeletePermission: func(ctx context.Context, input *awsacmpca.DeletePermissionInput, opts []func(*awsacmpca.Options)) (*awsacmpca.DeletePermissionOutput, error) {
						return &awsacmpca.DeletePermissionOutput{}, nil
					},
					MockDescribeCertificateAuthority: func(ctx context.Context, input *awsacmpca.DescribeCertificateAuthorityInput, opts []func(*awsacmpca.Options)) (*awsacmpca.DescribeCertificateAuthorityOutput, error) {
						return &awsacmpca.DescribeCertificateAuthorityOutput{
							CertificateAuthority: &awsacmpcatypes.CertificateAuthority{
								Type:   awsacmpcatypes.CertificateAuthorityTypeRoot,
								Status: awsacmpcatypes.CertificateAuthorityStatusActive,
								RevocationConfiguration: &awsacmpcatypes.RevocationConfiguration{
									CrlConfiguration: &awsacmpcatypes.CrlConfiguration{
										Enabled: false,
									},
								},
							},
						}, nil
					},
					MockUpdateCertificateAuthority: func(ctx context.Context, input *awsacmpca.UpdateCertificateAuthorityInput, opts []func(*awsacmpca.Options)) (*awsacmpca.UpdateCertificateAuthorityOutput, error) {
						return &awsacmpca.UpdateCertificateAuthorityOutput{}, nil
					},
					MockListTags: func(ctx context.Context, input *awsacmpca.ListTagsInput, opts []func(*awsacmpca.Options)) (*awsacmpca.ListTagsOutput, error) {
						return &awsacmpca.ListTagsOutput{
							NextToken: aws.String(nextToken),
							Tags:      []awsacmpcatypes.Tag{{}},
						}, nil
					},
					MockUntagCertificateAuthority: func(ctx context.Context, input *awsacmpca.UntagCertificateAuthorityInput, opts []func(*awsacmpca.Options)) (*awsacmpca.UntagCertificateAuthorityOutput, error) {
						return &awsacmpca.UntagCertificateAuthorityOutput{}, nil
					},
					MockTagCertificateAuthority: func(ctx context.Context, input *awsacmpca.TagCertificateAuthorityInput, opts []func(*awsacmpca.Options)) (*awsacmpca.TagCertificateAuthorityOutput, error) {
						return &awsacmpca.TagCertificateAuthorityOutput{}, nil
					},
					MockCreatePermission: func(ctx context.Context, input *awsacmpca.CreatePermissionInput, opts []func(*awsacmpca.Options)) (*awsacmpca.CreatePermissionOutput, error) {
						return &awsacmpca.CreatePermissionOutput{}, nil
					},
				},
				cr: certificateAuthority(withCertificateAuthorityStatus()),
			},
			want: want{
				cr: certificateAuthority(withCertificateAuthorityStatus()),
			},
		},
		"InValidInput": {
			args: args{
				cr: unexpectedItem,
			},
			want: want{
				cr:  unexpectedItem,
				err: errors.New(errUnexpectedObject),
			},
		},
		"ClientUpdateCertificateDescribeError": {
			args: args{
				acmpca: &fake.MockCertificateAuthorityClient{
					MockDeletePermission: func(ctx context.Context, input *awsacmpca.DeletePermissionInput, opts []func(*awsacmpca.Options)) (*awsacmpca.DeletePermissionOutput, error) {
						return nil, errBoom
					},
					MockDescribeCertificateAuthority: func(ctx context.Context, input *awsacmpca.DescribeCertificateAuthorityInput, opts []func(*awsacmpca.Options)) (*awsacmpca.DescribeCertificateAuthorityOutput, error) {
						return nil, errBoom
					},
					MockUpdateCertificateAuthority: func(ctx context.Context, input *awsacmpca.UpdateCertificateAuthorityInput, opts []func(*awsacmpca.Options)) (*awsacmpca.UpdateCertificateAuthorityOutput, error) {
						return nil, errBoom
					},
					MockListTags: func(ctx context.Context, input *awsacmpca.ListTagsInput, opts []func(*awsacmpca.Options)) (*awsacmpca.ListTagsOutput, error) {
						return nil, errBoom
					},
					MockUntagCertificateAuthority: func(ctx context.Context, input *awsacmpca.UntagCertificateAuthorityInput, opts []func(*awsacmpca.Options)) (*awsacmpca.UntagCertificateAuthorityOutput, error) {
						return &awsacmpca.UntagCertificateAuthorityOutput{}, nil
					},
					MockTagCertificateAuthority: func(ctx context.Context, input *awsacmpca.TagCertificateAuthorityInput, opts []func(*awsacmpca.Options)) (*awsacmpca.TagCertificateAuthorityOutput, error) {
						return &awsacmpca.TagCertificateAuthorityOutput{}, nil
					},
					MockCreatePermission: func(ctx context.Context, input *awsacmpca.CreatePermissionInput, opts []func(*awsacmpca.Options)) (*awsacmpca.CreatePermissionOutput, error) {
						return &awsacmpca.CreatePermissionOutput{}, nil
					},
				},
				cr: certificateAuthority(withCertificateAuthorityStatus()),
			},
			want: want{
				cr:  certificateAuthority(withCertificateAuthorityStatus()),
				err: errorutils.Wrap(errBoom, errCertificateAuthority),
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
		"ValidInput": {
			args: args{
				acmpca: &fake.MockCertificateAuthorityClient{
					MockDeleteCertificateAuthority: func(ctx context.Context, input *awsacmpca.DeleteCertificateAuthorityInput, opts []func(*awsacmpca.Options)) (*awsacmpca.DeleteCertificateAuthorityOutput, error) {
						return &awsacmpca.DeleteCertificateAuthorityOutput{}, nil
					},
					MockDeletePermission: func(ctx context.Context, input *awsacmpca.DeletePermissionInput, opts []func(*awsacmpca.Options)) (*awsacmpca.DeletePermissionOutput, error) {
						return &awsacmpca.DeletePermissionOutput{}, nil
					},
					MockDescribeCertificateAuthority: func(ctx context.Context, input *awsacmpca.DescribeCertificateAuthorityInput, opts []func(*awsacmpca.Options)) (*awsacmpca.DescribeCertificateAuthorityOutput, error) {
						return &awsacmpca.DescribeCertificateAuthorityOutput{
							CertificateAuthority: &awsacmpcatypes.CertificateAuthority{
								Type:   awsacmpcatypes.CertificateAuthorityTypeRoot,
								Status: awsacmpcatypes.CertificateAuthorityStatusActive,
							},
						}, nil
					},
					MockUpdateCertificateAuthority: func(ctx context.Context, input *awsacmpca.UpdateCertificateAuthorityInput, opts []func(*awsacmpca.Options)) (*awsacmpca.UpdateCertificateAuthorityOutput, error) {
						return &awsacmpca.UpdateCertificateAuthorityOutput{}, nil
					},
				},
				cr: certificateAuthority(withCertificateAuthorityArn()),
			},
			want: want{
				cr: certificateAuthority(withCertificateAuthorityArn(), withConditions(xpv1.Deleting())),
			},
		},
		"InValidInput": {
			args: args{
				cr: unexpectedItem,
			},
			want: want{
				cr:  unexpectedItem,
				err: errors.New(errUnexpectedObject),
			},
		},
		"ClientError": {
			args: args{
				acmpca: &fake.MockCertificateAuthorityClient{
					MockDeleteCertificateAuthority: func(ctx context.Context, input *awsacmpca.DeleteCertificateAuthorityInput, opts []func(*awsacmpca.Options)) (*awsacmpca.DeleteCertificateAuthorityOutput, error) {
						return nil, errBoom
					},
					MockDescribeCertificateAuthority: func(ctx context.Context, input *awsacmpca.DescribeCertificateAuthorityInput, opts []func(*awsacmpca.Options)) (*awsacmpca.DescribeCertificateAuthorityOutput, error) {
						return &awsacmpca.DescribeCertificateAuthorityOutput{
							CertificateAuthority: &awsacmpcatypes.CertificateAuthority{
								Type:   awsacmpcatypes.CertificateAuthorityTypeRoot,
								Status: awsacmpcatypes.CertificateAuthorityStatusActive,
							},
						}, nil
					},
					MockUpdateCertificateAuthority: func(ctx context.Context, input *awsacmpca.UpdateCertificateAuthorityInput, opts []func(*awsacmpca.Options)) (*awsacmpca.UpdateCertificateAuthorityOutput, error) {
						return &awsacmpca.UpdateCertificateAuthorityOutput{}, nil
					},
				},
				cr: certificateAuthority(),
			},
			want: want{
				cr:  certificateAuthority(withConditions(xpv1.Deleting())),
				err: errorutils.Wrap(errBoom, errDelete),
			},
		},
		"ResourceDoesNotExist": {
			args: args{
				acmpca: &fake.MockCertificateAuthorityClient{
					MockDeleteCertificateAuthority: func(ctx context.Context, input *awsacmpca.DeleteCertificateAuthorityInput, opts []func(*awsacmpca.Options)) (*awsacmpca.DeleteCertificateAuthorityOutput, error) {
						return nil, &awsacmpcatypes.ResourceNotFoundException{}
					},
					MockDescribeCertificateAuthority: func(ctx context.Context, input *awsacmpca.DescribeCertificateAuthorityInput, opts []func(*awsacmpca.Options)) (*awsacmpca.DescribeCertificateAuthorityOutput, error) {
						return nil, &awsacmpcatypes.ResourceNotFoundException{}
					},
					MockUpdateCertificateAuthority: func(ctx context.Context, input *awsacmpca.UpdateCertificateAuthorityInput, opts []func(*awsacmpca.Options)) (*awsacmpca.UpdateCertificateAuthorityOutput, error) {
						return nil, &awsacmpcatypes.ResourceNotFoundException{}
					},
				},
				cr: certificateAuthority(),
			},
			want: want{
				cr:  certificateAuthority(withConditions(xpv1.Deleting())),
				err: errorutils.Wrap(&awsacmpcatypes.ResourceNotFoundException{}, errDelete),
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
