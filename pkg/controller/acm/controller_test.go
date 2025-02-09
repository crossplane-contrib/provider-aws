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

package acm

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsacm "github.com/aws/aws-sdk-go-v2/service/acm"
	awsacmtype "github.com/aws/aws-sdk-go-v2/service/acm/types"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"

	"github.com/crossplane-contrib/provider-aws/apis/acm/v1beta1"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/acm"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/acm/fake"
	errorutils "github.com/crossplane-contrib/provider-aws/pkg/utils/errors"
)

var (
	// an arbitrary managed resource
	unexpectedItem resource.Managed
	domainName     = "some.site"
	certificateArn = "somearn"

	errBoom = errors.New("boom")
)

type args struct {
	acm acm.Client
	cr  resource.Managed
}

type certificateModifier func(*v1beta1.Certificate)

func withConditions(c ...xpv1.Condition) certificateModifier {
	return func(r *v1beta1.Certificate) { r.Status.ConditionedStatus.Conditions = c }
}

func withDomainName() certificateModifier {
	return func(r *v1beta1.Certificate) {
		r.Spec.ForProvider.DomainName = domainName
		meta.SetExternalName(r, certificateArn)
	}
}

func withCertificateTransparencyLoggingPreference() certificateModifier {
	certificateTransparencyLoggingPreference := string(awsacmtype.CertificateTransparencyLoggingPreferenceDisabled)

	return func(r *v1beta1.Certificate) {
		r.Spec.ForProvider.Options = &v1beta1.CertificateOptions{
			CertificateTransparencyLoggingPreference: certificateTransparencyLoggingPreference,
		}
		meta.SetExternalName(r, certificateArn)
	}
}

func withTags() certificateModifier {
	return func(r *v1beta1.Certificate) {
		r.Spec.ForProvider.Tags = append(r.Spec.ForProvider.Tags, v1beta1.Tag{
			Key:   "Name",
			Value: "somename",
		})
		meta.SetExternalName(r, certificateArn)
	}
}

func withCertificateArn() certificateModifier {
	return func(r *v1beta1.Certificate) {
		certificateTransparencyLoggingPreference := string(awsacmtype.CertificateTransparencyLoggingPreferenceDisabled)

		r.Status.AtProvider.CertificateARN = certificateArn
		r.Spec.ForProvider.Options = &v1beta1.CertificateOptions{
			CertificateTransparencyLoggingPreference: certificateTransparencyLoggingPreference,
		}
		meta.SetExternalName(r, certificateArn)
	}
}

func withStatus(status string) certificateModifier {
	return func(r *v1beta1.Certificate) {
		r.Status.AtProvider.Status = status
	}
}

func certificate(m ...certificateModifier) *v1beta1.Certificate {
	cr := &v1beta1.Certificate{}
	meta.SetExternalName(cr, certificateArn)
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
				acm: &fake.MockCertificateClient{
					MockDescribeCertificate: func(ctx context.Context, input *awsacm.DescribeCertificateInput, opts []func(*awsacm.Options)) (*awsacm.DescribeCertificateOutput, error) {
						return &awsacm.DescribeCertificateOutput{
							Certificate: &awsacmtype.CertificateDetail{
								CertificateArn: aws.String(certificateArn),
								Options:        &awsacmtype.CertificateOptions{CertificateTransparencyLoggingPreference: awsacmtype.CertificateTransparencyLoggingPreferenceDisabled},
								Status:         awsacmtype.CertificateStatusIssued,
							},
						}, nil
					},
					MockListTagsForCertificate: func(ctx context.Context, input *awsacm.ListTagsForCertificateInput, opts []func(*awsacm.Options)) (*awsacm.ListTagsForCertificateOutput, error) {
						return &awsacm.ListTagsForCertificateOutput{
							Tags: []awsacmtype.Tag{{}},
						}, nil
					},
				},
				cr: certificate(),
			},
			want: want{
				cr: certificate(withCertificateArn(), withStatus(string(awsacmtype.CertificateStatusIssued)), withConditions(xpv1.Available())),
				result: managed.ExternalObservation{
					ResourceExists:          true,
					ResourceUpToDate:        true,
					ResourceLateInitialized: true,
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
				acm: &fake.MockCertificateClient{
					MockDescribeCertificate: func(ctx context.Context, input *awsacm.DescribeCertificateInput, opts []func(*awsacm.Options)) (*awsacm.DescribeCertificateOutput, error) {
						return nil, errBoom
					},
				},
				cr: certificate(withCertificateArn()),
			},
			want: want{
				cr:  certificate(withCertificateArn()),
				err: errorutils.Wrap(errBoom, errGet),
			},
		},
		"ResourceDoesNotExist": {
			args: args{
				acm: &fake.MockCertificateClient{
					MockDescribeCertificate: func(ctx context.Context, input *awsacm.DescribeCertificateInput, opts []func(*awsacm.Options)) (*awsacm.DescribeCertificateOutput, error) {
						return nil, &awsacmtype.ResourceNotFoundException{}
					},
				},
				cr: certificate(),
			},
			want: want{
				cr: certificate(),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{
				client: tc.acm,
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
				acm: &fake.MockCertificateClient{
					MockRequestCertificate: func(ctx context.Context, input *awsacm.RequestCertificateInput, opts []func(*awsacm.Options)) (*awsacm.RequestCertificateOutput, error) {
						return &awsacm.RequestCertificateOutput{
							CertificateArn: aws.String(certificateArn),
						}, nil
					},
				},
				cr: certificate(withDomainName()),
			},
			want: want{
				cr: certificate(
					withDomainName()),
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
				acm: &fake.MockCertificateClient{
					MockRequestCertificate: func(ctx context.Context, input *awsacm.RequestCertificateInput, opts []func(*awsacm.Options)) (*awsacm.RequestCertificateOutput, error) {
						return nil, errBoom
					},
				},
				cr: certificate(),
			},
			want: want{
				cr:  certificate(),
				err: errorutils.Wrap(errBoom, errCreate),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{
				client: tc.acm,
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
				acm: &fake.MockCertificateClient{

					MockUpdateCertificateOptions: func(ctx context.Context, input *awsacm.UpdateCertificateOptionsInput, opts []func(*awsacm.Options)) (*awsacm.UpdateCertificateOptionsOutput, error) {
						return &awsacm.UpdateCertificateOptionsOutput{}, nil
					},
					MockListTagsForCertificate: func(ctx context.Context, input *awsacm.ListTagsForCertificateInput, opts []func(*awsacm.Options)) (*awsacm.ListTagsForCertificateOutput, error) {
						return &awsacm.ListTagsForCertificateOutput{
							Tags: []awsacmtype.Tag{{}},
						}, nil
					},
					MockRemoveTagsFromCertificate: func(ctx context.Context, input *awsacm.RemoveTagsFromCertificateInput, opts []func(*awsacm.Options)) (*awsacm.RemoveTagsFromCertificateOutput, error) {
						return &awsacm.RemoveTagsFromCertificateOutput{}, nil
					},
					MockAddTagsToCertificate: func(ctx context.Context, input *awsacm.AddTagsToCertificateInput, opts []func(*awsacm.Options)) (*awsacm.AddTagsToCertificateOutput, error) {
						return &awsacm.AddTagsToCertificateOutput{}, nil
					},
				},
				cr: certificate(),
			},
			want: want{
				cr: certificate(),
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
		"ClientUpdateCertificateOptionsError": {
			args: args{
				acm: &fake.MockCertificateClient{

					MockUpdateCertificateOptions: func(ctx context.Context, input *awsacm.UpdateCertificateOptionsInput, opts []func(*awsacm.Options)) (*awsacm.UpdateCertificateOptionsOutput, error) {
						return nil, errBoom
					},
					MockListTagsForCertificate: func(ctx context.Context, input *awsacm.ListTagsForCertificateInput, opts []func(*awsacm.Options)) (*awsacm.ListTagsForCertificateOutput, error) {
						return &awsacm.ListTagsForCertificateOutput{
							Tags: []awsacmtype.Tag{{}},
						}, nil
					},
					MockRemoveTagsFromCertificate: func(ctx context.Context, input *awsacm.RemoveTagsFromCertificateInput, opts []func(*awsacm.Options)) (*awsacm.RemoveTagsFromCertificateOutput, error) {
						return &awsacm.RemoveTagsFromCertificateOutput{}, nil
					},
					MockAddTagsToCertificate: func(ctx context.Context, input *awsacm.AddTagsToCertificateInput, opts []func(*awsacm.Options)) (*awsacm.AddTagsToCertificateOutput, error) {
						return &awsacm.AddTagsToCertificateOutput{}, nil
					},
				},
				cr: certificate(withCertificateTransparencyLoggingPreference()),
			},
			want: want{
				cr:  certificate(withCertificateTransparencyLoggingPreference()),
				err: errorutils.Wrap(errBoom, errUpdate),
			},
		},
		"ClientUpdateTagsError": {
			args: args{
				acm: &fake.MockCertificateClient{

					MockUpdateCertificateOptions: func(ctx context.Context, input *awsacm.UpdateCertificateOptionsInput, opts []func(*awsacm.Options)) (*awsacm.UpdateCertificateOptionsOutput, error) {
						return &awsacm.UpdateCertificateOptionsOutput{}, nil
					},
					MockListTagsForCertificate: func(ctx context.Context, input *awsacm.ListTagsForCertificateInput, opts []func(*awsacm.Options)) (*awsacm.ListTagsForCertificateOutput, error) {
						return &awsacm.ListTagsForCertificateOutput{
							Tags: []awsacmtype.Tag{{}},
						}, nil
					},
					MockRemoveTagsFromCertificate: func(ctx context.Context, input *awsacm.RemoveTagsFromCertificateInput, opts []func(*awsacm.Options)) (*awsacm.RemoveTagsFromCertificateOutput, error) {
						return &awsacm.RemoveTagsFromCertificateOutput{}, nil
					},
					MockAddTagsToCertificate: func(ctx context.Context, input *awsacm.AddTagsToCertificateInput, opts []func(*awsacm.Options)) (*awsacm.AddTagsToCertificateOutput, error) {
						return nil, errBoom
					},
				},
				cr: certificate(withTags()),
			},
			want: want{
				cr:  certificate(withTags()),
				err: errorutils.Wrap(errBoom, errAddTagsFailed),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.acm}
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
				acm: &fake.MockCertificateClient{
					MockDeleteCertificate: func(ctx context.Context, input *awsacm.DeleteCertificateInput, opts []func(*awsacm.Options)) (*awsacm.DeleteCertificateOutput, error) {
						return &awsacm.DeleteCertificateOutput{}, nil
					},
				},
				cr: certificate(withCertificateTransparencyLoggingPreference()),
			},
			want: want{
				cr: certificate(withCertificateTransparencyLoggingPreference()),
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
				acm: &fake.MockCertificateClient{
					MockDeleteCertificate: func(ctx context.Context, input *awsacm.DeleteCertificateInput, opts []func(*awsacm.Options)) (*awsacm.DeleteCertificateOutput, error) {
						return nil, errBoom
					},
				},
				cr: certificate(),
			},
			want: want{
				cr:  certificate(),
				err: errorutils.Wrap(errBoom, errDelete),
			},
		},
		"ResourceDoesNotExist": {
			args: args{
				acm: &fake.MockCertificateClient{
					MockDeleteCertificate: func(ctx context.Context, input *awsacm.DeleteCertificateInput, opts []func(*awsacm.Options)) (*awsacm.DeleteCertificateOutput, error) {
						return nil, &awsacmtype.ResourceNotFoundException{}
					},
				},
				cr: certificate(),
			},
			want: want{
				cr: certificate(),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.acm}
			_, err := e.Delete(context.Background(), tc.args.cr)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.cr, tc.args.cr, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}
