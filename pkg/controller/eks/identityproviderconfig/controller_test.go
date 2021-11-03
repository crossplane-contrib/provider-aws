/*
Copyright 2021 The Crossplane Authors.

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

package identityproviderconfig

import (
	"context"
	"testing"

	awseks "github.com/aws/aws-sdk-go-v2/service/eks"
	awsekstypes "github.com/aws/aws-sdk-go-v2/service/eks/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"

	"github.com/crossplane/provider-aws/apis/eks/v1alpha1"
	awsclient "github.com/crossplane/provider-aws/pkg/clients"
	"github.com/crossplane/provider-aws/pkg/clients/eks"
	"github.com/crossplane/provider-aws/pkg/clients/eks/fake"
)

var (
	tags    = map[string]string{"tag1": "value1", "tag2": "value2"}
	errBoom = errors.New("boom")
)

type args struct {
	eks  eks.Client
	kube client.Client
	cr   *v1alpha1.IdentityProviderConfig
}

type identityProviderConfigModifier func(config *v1alpha1.IdentityProviderConfig)

func withConditions(c ...xpv1.Condition) identityProviderConfigModifier {
	return func(r *v1alpha1.IdentityProviderConfig) { r.Status.ConditionedStatus.Conditions = c }
}

func withTags(tagMaps ...map[string]string) identityProviderConfigModifier {
	tags := map[string]string{}
	for _, tagMap := range tagMaps {
		for k, v := range tagMap {
			tags[k] = v
		}
	}
	return func(r *v1alpha1.IdentityProviderConfig) { r.Spec.ForProvider.Tags = tags }
}

func withStatus(s v1alpha1.IdentityProviderConfigStatusType) identityProviderConfigModifier {
	return func(r *v1alpha1.IdentityProviderConfig) { r.Status.AtProvider.Status = s }
}

func identityProviderConfig(m ...identityProviderConfigModifier) *v1alpha1.IdentityProviderConfig {
	cr := &v1alpha1.IdentityProviderConfig{}
	for _, f := range m {
		f(cr)
	}
	return cr
}

var _ managed.ExternalClient = &external{}
var _ managed.ExternalConnecter = &connector{}

func TestObserve(t *testing.T) {
	type want struct {
		cr     *v1alpha1.IdentityProviderConfig
		result managed.ExternalObservation
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"SuccessfulAvailable": {
			args: args{
				eks: &fake.MockClient{
					MockDescribeIdentityProviderConfig: func(ctx context.Context, input *awseks.DescribeIdentityProviderConfigInput, opts []func(*awseks.Options)) (*awseks.DescribeIdentityProviderConfigOutput, error) {
						return &awseks.DescribeIdentityProviderConfigOutput{
								IdentityProviderConfig: &awsekstypes.IdentityProviderConfigResponse{
									Oidc: &awsekstypes.OidcIdentityProviderConfig{
										Status: awsekstypes.ConfigStatusActive,
									},
								},
							},
							nil
					},
				},
				cr: identityProviderConfig(),
			},
			want: want{
				cr: identityProviderConfig(
					withConditions(xpv1.Available()),
					withStatus(v1alpha1.IdentityProviderConfigStatusActive)),
				result: managed.ExternalObservation{
					ResourceExists:          true,
					ResourceUpToDate:        true,
					ResourceLateInitialized: false,
				},
			},
		},
		"DeletingState": {
			args: args{
				eks: &fake.MockClient{
					MockDescribeIdentityProviderConfig: func(ctx context.Context, input *awseks.DescribeIdentityProviderConfigInput, opts []func(*awseks.Options)) (*awseks.DescribeIdentityProviderConfigOutput, error) {
						return &awseks.DescribeIdentityProviderConfigOutput{
								IdentityProviderConfig: &awsekstypes.IdentityProviderConfigResponse{
									Oidc: &awsekstypes.OidcIdentityProviderConfig{
										Status: awsekstypes.ConfigStatusDeleting,
									},
								},
							},
							nil
					},
				},
				cr: identityProviderConfig(),
			},
			want: want{
				cr: identityProviderConfig(
					withConditions(xpv1.Deleting()),
					withStatus(v1alpha1.IdentityProviderConfigStatusDeleting)),
				result: managed.ExternalObservation{
					ResourceExists:          true,
					ResourceUpToDate:        true,
					ResourceLateInitialized: false,
				},
			},
		},
		"FailedDescribeRequest": {
			args: args{
				eks: &fake.MockClient{
					MockDescribeIdentityProviderConfig: func(ctx context.Context, input *awseks.DescribeIdentityProviderConfigInput, opts []func(*awseks.Options)) (*awseks.DescribeIdentityProviderConfigOutput, error) {
						return nil, errBoom
					},
				},
				cr: identityProviderConfig(),
			},
			want: want{
				cr:  identityProviderConfig(),
				err: awsclient.Wrap(errBoom, errDescribeFailed),
			},
		},
		"NotFound": {
			args: args{
				eks: &fake.MockClient{
					MockDescribeIdentityProviderConfig: func(ctx context.Context, input *awseks.DescribeIdentityProviderConfigInput, opts []func(*awseks.Options)) (*awseks.DescribeIdentityProviderConfigOutput, error) {
						return nil, &awsekstypes.ResourceNotFoundException{}
					},
				},
				cr: identityProviderConfig(),
			},
			want: want{
				cr: identityProviderConfig(),
			},
		},
		"LateInitSuccess": {
			args: args{
				kube: &test.MockClient{
					MockUpdate: test.NewMockUpdateFn(nil),
				},
				eks: &fake.MockClient{
					MockDescribeIdentityProviderConfig: func(ctx context.Context, input *awseks.DescribeIdentityProviderConfigInput, opts []func(*awseks.Options)) (*awseks.DescribeIdentityProviderConfigOutput, error) {
						return &awseks.DescribeIdentityProviderConfigOutput{
							IdentityProviderConfig: &awsekstypes.IdentityProviderConfigResponse{
								Oidc: &awsekstypes.OidcIdentityProviderConfig{
									Status: awsekstypes.ConfigStatusCreating,
									Tags:   tags,
								},
							},
						}, nil
					},
				},
				cr: identityProviderConfig(),
			},
			want: want{
				cr: identityProviderConfig(
					withStatus(v1alpha1.IdentityProviderConfigStatusCreating),
					withConditions(xpv1.Creating()),
					withTags(tags),
				),
				result: managed.ExternalObservation{
					ResourceExists:          true,
					ResourceUpToDate:        true,
					ResourceLateInitialized: true,
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.eks}
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
		cr     *v1alpha1.IdentityProviderConfig
		result managed.ExternalCreation
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"Successful": {
			args: args{
				eks: &fake.MockClient{
					MockAssociateIdentityProviderConfig: func(ctx context.Context, input *awseks.AssociateIdentityProviderConfigInput, opts []func(*awseks.Options)) (*awseks.AssociateIdentityProviderConfigOutput, error) {
						return &awseks.AssociateIdentityProviderConfigOutput{}, nil
					},
				},
				cr: identityProviderConfig(),
			},
			want: want{
				cr:     identityProviderConfig(withConditions(xpv1.Creating())),
				result: managed.ExternalCreation{},
			},
		},
		"SuccessfulNoNeedForCreate": {
			args: args{
				cr: identityProviderConfig(withStatus(v1alpha1.IdentityProviderConfigStatusCreating)),
			},
			want: want{
				cr: identityProviderConfig(
					withStatus(v1alpha1.IdentityProviderConfigStatusCreating),
					withConditions(xpv1.Creating())),
			},
		},
		"FailedRequest": {
			args: args{
				eks: &fake.MockClient{
					MockAssociateIdentityProviderConfig: func(ctx context.Context, input *awseks.AssociateIdentityProviderConfigInput, opts []func(*awseks.Options)) (*awseks.AssociateIdentityProviderConfigOutput, error) {
						return nil, errBoom
					},
				},
				cr: identityProviderConfig(),
			},
			want: want{
				cr:  identityProviderConfig(withConditions(xpv1.Creating())),
				err: awsclient.Wrap(errBoom, errCreateFailed),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.eks}
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
		cr     *v1alpha1.IdentityProviderConfig
		result managed.ExternalUpdate
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"SuccessfulAddTags": {
			args: args{
				eks: &fake.MockClient{
					MockDescribeIdentityProviderConfig: func(ctx context.Context, input *awseks.DescribeIdentityProviderConfigInput, opts []func(*awseks.Options)) (*awseks.DescribeIdentityProviderConfigOutput, error) {
						return &awseks.DescribeIdentityProviderConfigOutput{
							IdentityProviderConfig: &awsekstypes.IdentityProviderConfigResponse{
								Oidc: &awsekstypes.OidcIdentityProviderConfig{},
							},
						}, nil
					},
					MockTagResource: func(ctx context.Context, input *awseks.TagResourceInput, opts []func(*awseks.Options)) (*awseks.TagResourceOutput, error) {
						return &awseks.TagResourceOutput{}, nil
					},
				},
				cr: identityProviderConfig(
					withTags(map[string]string{"foo": "bar"})),
			},
			want: want{
				cr: identityProviderConfig(
					withTags(map[string]string{"foo": "bar"})),
			},
		},
		"SuccessfulRemoveTags": {
			args: args{
				eks: &fake.MockClient{
					MockDescribeIdentityProviderConfig: func(ctx context.Context, input *awseks.DescribeIdentityProviderConfigInput, opts []func(*awseks.Options)) (*awseks.DescribeIdentityProviderConfigOutput, error) {
						return &awseks.DescribeIdentityProviderConfigOutput{
							IdentityProviderConfig: &awsekstypes.IdentityProviderConfigResponse{
								Oidc: &awsekstypes.OidcIdentityProviderConfig{}},
						}, nil
					},
					MockUntagResource: func(ctx context.Context, input *awseks.UntagResourceInput, opts []func(*awseks.Options)) (*awseks.UntagResourceOutput, error) {
						return &awseks.UntagResourceOutput{}, nil
					},
				},
				cr: identityProviderConfig(),
			},
			want: want{
				cr: identityProviderConfig(),
			},
		},
		"FailedRemoveTags": {
			args: args{
				eks: &fake.MockClient{
					MockDescribeIdentityProviderConfig: func(ctx context.Context, input *awseks.DescribeIdentityProviderConfigInput, opts []func(*awseks.Options)) (*awseks.DescribeIdentityProviderConfigOutput, error) {
						return &awseks.DescribeIdentityProviderConfigOutput{
							IdentityProviderConfig: &awsekstypes.IdentityProviderConfigResponse{
								Oidc: &awsekstypes.OidcIdentityProviderConfig{
									Tags: map[string]string{"foo": "bar"},
								},
							},
						}, nil
					},
					MockUntagResource: func(ctx context.Context, input *awseks.UntagResourceInput, opts []func(*awseks.Options)) (*awseks.UntagResourceOutput, error) {
						return nil, errBoom
					},
				},
				cr: identityProviderConfig(),
			},
			want: want{
				cr:  identityProviderConfig(),
				err: awsclient.Wrap(errBoom, errAddTagsFailed),
			},
		},
		"FailedAddTags": {
			args: args{
				eks: &fake.MockClient{
					MockDescribeIdentityProviderConfig: func(ctx context.Context, input *awseks.DescribeIdentityProviderConfigInput, opts []func(*awseks.Options)) (*awseks.DescribeIdentityProviderConfigOutput, error) {
						return &awseks.DescribeIdentityProviderConfigOutput{
							IdentityProviderConfig: &awsekstypes.IdentityProviderConfigResponse{
								Oidc: &awsekstypes.OidcIdentityProviderConfig{},
							},
						}, nil
					},
					MockTagResource: func(ctx context.Context, input *awseks.TagResourceInput, opts []func(*awseks.Options)) (*awseks.TagResourceOutput, error) {
						return nil, errBoom
					},
				},
				cr: identityProviderConfig(withTags(map[string]string{"foo": "bar"})),
			},
			want: want{
				cr:  identityProviderConfig(withTags(map[string]string{"foo": "bar"})),
				err: awsclient.Wrap(errBoom, errAddTagsFailed),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.eks}
			u, err := e.Update(context.Background(), tc.args.cr)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.cr, tc.args.cr, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.result, u); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	type want struct {
		cr  *v1alpha1.IdentityProviderConfig
		err error
	}

	cases := map[string]struct {
		args
		want
	}{
		"Successful": {
			args: args{
				eks: &fake.MockClient{
					MockDisassociateIdentityProviderConfig: func(ctx context.Context, input *awseks.DisassociateIdentityProviderConfigInput, opts []func(*awseks.Options)) (*awseks.DisassociateIdentityProviderConfigOutput, error) {
						return &awseks.DisassociateIdentityProviderConfigOutput{}, nil
					},
				},
				cr: identityProviderConfig(),
			},
			want: want{
				cr: identityProviderConfig(withConditions(xpv1.Deleting())),
			},
		},
		"AlreadyDeleting": {
			args: args{
				cr: identityProviderConfig(withStatus(v1alpha1.IdentityProviderConfigStatusDeleting)),
			},
			want: want{
				cr: identityProviderConfig(withStatus(v1alpha1.IdentityProviderConfigStatusDeleting),
					withConditions(xpv1.Deleting())),
			},
		},
		"AlreadyDeleted": {
			args: args{
				eks: &fake.MockClient{
					MockDisassociateIdentityProviderConfig: func(ctx context.Context, input *awseks.DisassociateIdentityProviderConfigInput, opts []func(*awseks.Options)) (*awseks.DisassociateIdentityProviderConfigOutput, error) {
						return nil, &awsekstypes.ResourceNotFoundException{}
					},
				},
				cr: identityProviderConfig(),
			},
			want: want{
				cr: identityProviderConfig(withConditions(xpv1.Deleting())),
			},
		},
		"Failed": {
			args: args{
				eks: &fake.MockClient{
					MockDisassociateIdentityProviderConfig: func(ctx context.Context, input *awseks.DisassociateIdentityProviderConfigInput, opts []func(*awseks.Options)) (*awseks.DisassociateIdentityProviderConfigOutput, error) {
						return nil, errBoom
					},
				},
				cr: identityProviderConfig(),
			},
			want: want{
				cr:  identityProviderConfig(withConditions(xpv1.Deleting())),
				err: awsclient.Wrap(errBoom, errDeleteFailed),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.eks}
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

func TestInitialize(t *testing.T) {
	type want struct {
		cr  *v1alpha1.IdentityProviderConfig
		err error
	}

	cases := map[string]struct {
		args
		want
	}{
		"Successful": {
			args: args{
				cr:   identityProviderConfig(withTags(map[string]string{"foo": "bar"})),
				kube: &test.MockClient{MockUpdate: test.NewMockUpdateFn(nil)},
			},
			want: want{
				cr: identityProviderConfig(withTags(resource.GetExternalTags(identityProviderConfig()), map[string]string{"foo": "bar"})),
			},
		},
		"UpdateFailed": {
			args: args{
				cr:   identityProviderConfig(),
				kube: &test.MockClient{MockUpdate: test.NewMockUpdateFn(errBoom)},
			},
			want: want{
				err: errors.Wrap(errBoom, errKubeUpdateFailed),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &tagger{kube: tc.kube}
			err := e.Initialize(context.Background(), tc.args.cr)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.cr, tc.args.cr); err == nil && diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}
