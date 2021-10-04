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

package fargateprofile

import (
	"context"
	"testing"

	awseks "github.com/aws/aws-sdk-go-v2/service/eks"
	awsekstypes "github.com/aws/aws-sdk-go-v2/service/eks/types"

	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/test"

	"github.com/crossplane/provider-aws/apis/eks/v1alpha1"
	awsclient "github.com/crossplane/provider-aws/pkg/clients"
	"github.com/crossplane/provider-aws/pkg/clients/eks"
	"github.com/crossplane/provider-aws/pkg/clients/eks/fake"
)

var (
	subnets = []string{"subnet1", "subnet2"}
	errBoom = errors.New("boom")
)

type args struct {
	eks  eks.Client
	kube client.Client
	cr   *v1alpha1.FargateProfile
}

type fargateProfileModifier func(*v1alpha1.FargateProfile)

func withConditions(c ...xpv1.Condition) fargateProfileModifier {
	return func(r *v1alpha1.FargateProfile) { r.Status.ConditionedStatus.Conditions = c }
}

func withTags(tagMaps ...map[string]string) fargateProfileModifier {
	tags := map[string]string{}
	for _, tagMap := range tagMaps {
		for k, v := range tagMap {
			tags[k] = v
		}
	}
	return func(r *v1alpha1.FargateProfile) { r.Spec.ForProvider.Tags = tags }
}

func withSubnets(s []string) fargateProfileModifier {
	return func(r *v1alpha1.FargateProfile) { r.Spec.ForProvider.Subnets = s }
}

func withStatus(s v1alpha1.FargateProfileStatusType) fargateProfileModifier {
	return func(r *v1alpha1.FargateProfile) { r.Status.AtProvider.Status = s }
}

func fargateProfile(m ...fargateProfileModifier) *v1alpha1.FargateProfile {
	cr := &v1alpha1.FargateProfile{}
	for _, f := range m {
		f(cr)
	}
	return cr
}

var _ managed.ExternalClient = &external{}
var _ managed.ExternalConnecter = &connector{}

func TestObserve(t *testing.T) {
	type want struct {
		cr     *v1alpha1.FargateProfile
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
					MockDescribeFargateProfile: func(ctx context.Context, input *awseks.DescribeFargateProfileInput, opts []func(*awseks.Options)) (*awseks.DescribeFargateProfileOutput, error) {
						return &awseks.DescribeFargateProfileOutput{
							FargateProfile: &awsekstypes.FargateProfile{
								Status: awsekstypes.FargateProfileStatusActive,
							},
						}, nil
					},
				},
				cr: fargateProfile(),
			},
			want: want{
				cr: fargateProfile(
					withConditions(xpv1.Available()),
					withStatus(v1alpha1.FargateProfileStatusActive)),
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
					MockDescribeFargateProfile: func(ctx context.Context, input *awseks.DescribeFargateProfileInput, opts []func(*awseks.Options)) (*awseks.DescribeFargateProfileOutput, error) {
						return &awseks.DescribeFargateProfileOutput{
							FargateProfile: &awsekstypes.FargateProfile{
								Status: awsekstypes.FargateProfileStatusDeleting,
							},
						}, nil
					},
				},
				cr: fargateProfile(),
			},
			want: want{
				cr: fargateProfile(
					withConditions(xpv1.Deleting()),
					withStatus(v1alpha1.FargateProfileStatusDeleting)),
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
					MockDescribeFargateProfile: func(ctx context.Context, input *awseks.DescribeFargateProfileInput, opts []func(*awseks.Options)) (*awseks.DescribeFargateProfileOutput, error) {
						return nil, errBoom
					},
				},
				cr: fargateProfile(),
			},
			want: want{
				cr:  fargateProfile(),
				err: awsclient.Wrap(errBoom, errDescribeFailed),
			},
		},
		"NotFound": {
			args: args{
				eks: &fake.MockClient{
					MockDescribeFargateProfile: func(ctx context.Context, input *awseks.DescribeFargateProfileInput, opts []func(*awseks.Options)) (*awseks.DescribeFargateProfileOutput, error) {
						return nil, &awsekstypes.ResourceNotFoundException{}
					},
				},
				cr: fargateProfile(),
			},
			want: want{
				cr: fargateProfile(),
			},
		},
		"LateInitSuccess": {
			args: args{
				kube: &test.MockClient{
					MockUpdate: test.NewMockUpdateFn(nil),
				},
				eks: &fake.MockClient{
					MockDescribeFargateProfile: func(ctx context.Context, input *awseks.DescribeFargateProfileInput, opts []func(*awseks.Options)) (*awseks.DescribeFargateProfileOutput, error) {
						return &awseks.DescribeFargateProfileOutput{
							FargateProfile: &awsekstypes.FargateProfile{
								Status:  awsekstypes.FargateProfileStatusCreating,
								Subnets: subnets,
							},
						}, nil
					},
				},
				cr: fargateProfile(),
			},
			want: want{
				cr: fargateProfile(
					withStatus(v1alpha1.FargateProfileStatusCreating),
					withConditions(xpv1.Creating()),
					withSubnets(subnets),
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
		cr     *v1alpha1.FargateProfile
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
					MockCreateFargateProfile: func(ctx context.Context, input *awseks.CreateFargateProfileInput, opts []func(*awseks.Options)) (*awseks.CreateFargateProfileOutput, error) {
						return &awseks.CreateFargateProfileOutput{}, nil
					},
				},
				cr: fargateProfile(),
			},
			want: want{
				cr:     fargateProfile(withConditions(xpv1.Creating())),
				result: managed.ExternalCreation{},
			},
		},
		"SuccessfulNoNeedForCreate": {
			args: args{
				cr: fargateProfile(withStatus(v1alpha1.FargateProfileStatusCreating)),
			},
			want: want{
				cr: fargateProfile(
					withStatus(v1alpha1.FargateProfileStatusCreating),
					withConditions(xpv1.Creating())),
			},
		},
		"FailedRequest": {
			args: args{
				eks: &fake.MockClient{
					MockCreateFargateProfile: func(ctx context.Context, input *awseks.CreateFargateProfileInput, opts []func(*awseks.Options)) (*awseks.CreateFargateProfileOutput, error) {
						return nil, errBoom
					},
				},
				cr: fargateProfile(),
			},
			want: want{
				cr:  fargateProfile(withConditions(xpv1.Creating())),
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
		cr     *v1alpha1.FargateProfile
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
					MockDescribeFargateProfile: func(ctx context.Context, input *awseks.DescribeFargateProfileInput, opts []func(*awseks.Options)) (*awseks.DescribeFargateProfileOutput, error) {
						return &awseks.DescribeFargateProfileOutput{
							FargateProfile: &awsekstypes.FargateProfile{},
						}, nil
					},
					MockTagResource: func(ctx context.Context, input *awseks.TagResourceInput, opts []func(*awseks.Options)) (*awseks.TagResourceOutput, error) {
						return &awseks.TagResourceOutput{}, nil
					},
				},
				cr: fargateProfile(
					withTags(map[string]string{"foo": "bar"})),
			},
			want: want{
				cr: fargateProfile(
					withTags(map[string]string{"foo": "bar"})),
			},
		},
		"SuccessfulRemoveTags": {
			args: args{
				eks: &fake.MockClient{
					MockDescribeFargateProfile: func(ctx context.Context, input *awseks.DescribeFargateProfileInput, opts []func(*awseks.Options)) (*awseks.DescribeFargateProfileOutput, error) {
						return &awseks.DescribeFargateProfileOutput{
							FargateProfile: &awsekstypes.FargateProfile{},
						}, nil
					},
					MockUntagResource: func(ctx context.Context, input *awseks.UntagResourceInput, opts []func(*awseks.Options)) (*awseks.UntagResourceOutput, error) {
						return &awseks.UntagResourceOutput{}, nil
					},
				},
				cr: fargateProfile(),
			},
			want: want{
				cr: fargateProfile(),
			},
		},
		"FailedRemoveTags": {
			args: args{
				eks: &fake.MockClient{
					MockDescribeFargateProfile: func(ctx context.Context, input *awseks.DescribeFargateProfileInput, opts []func(*awseks.Options)) (*awseks.DescribeFargateProfileOutput, error) {
						return &awseks.DescribeFargateProfileOutput{
							FargateProfile: &awsekstypes.FargateProfile{
								Tags: map[string]string{"foo": "bar"},
							},
						}, nil
					},
					MockUntagResource: func(ctx context.Context, input *awseks.UntagResourceInput, opts []func(*awseks.Options)) (*awseks.UntagResourceOutput, error) {
						return nil, errBoom
					},
				},
				cr: fargateProfile(),
			},
			want: want{
				cr:  fargateProfile(),
				err: awsclient.Wrap(errBoom, errAddTagsFailed),
			},
		},
		"FailedAddTags": {
			args: args{
				eks: &fake.MockClient{
					MockDescribeFargateProfile: func(ctx context.Context, input *awseks.DescribeFargateProfileInput, opts []func(*awseks.Options)) (*awseks.DescribeFargateProfileOutput, error) {
						return &awseks.DescribeFargateProfileOutput{
							FargateProfile: &awsekstypes.FargateProfile{},
						}, nil
					},
					MockTagResource: func(ctx context.Context, input *awseks.TagResourceInput, opts []func(*awseks.Options)) (*awseks.TagResourceOutput, error) {
						return nil, errBoom
					},
				},
				cr: fargateProfile(withTags(map[string]string{"foo": "bar"})),
			},
			want: want{
				cr:  fargateProfile(withTags(map[string]string{"foo": "bar"})),
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
		cr  *v1alpha1.FargateProfile
		err error
	}

	cases := map[string]struct {
		args
		want
	}{
		"Successful": {
			args: args{
				eks: &fake.MockClient{
					MockDeleteFargateProfile: func(ctx context.Context, input *awseks.DeleteFargateProfileInput, opts []func(*awseks.Options)) (*awseks.DeleteFargateProfileOutput, error) {
						return &awseks.DeleteFargateProfileOutput{}, nil
					},
				},
				cr: fargateProfile(),
			},
			want: want{
				cr: fargateProfile(withConditions(xpv1.Deleting())),
			},
		},
		"AlreadyDeleting": {
			args: args{
				cr: fargateProfile(withStatus(v1alpha1.FargateProfileStatusDeleting)),
			},
			want: want{
				cr: fargateProfile(withStatus(v1alpha1.FargateProfileStatusDeleting),
					withConditions(xpv1.Deleting())),
			},
		},
		"AlreadyDeleted": {
			args: args{
				eks: &fake.MockClient{
					MockDeleteFargateProfile: func(ctx context.Context, input *awseks.DeleteFargateProfileInput, opts []func(*awseks.Options)) (*awseks.DeleteFargateProfileOutput, error) {
						return nil, &awsekstypes.ResourceNotFoundException{}
					},
				},
				cr: fargateProfile(),
			},
			want: want{
				cr: fargateProfile(withConditions(xpv1.Deleting())),
			},
		},
		"Failed": {
			args: args{
				eks: &fake.MockClient{
					MockDeleteFargateProfile: func(ctx context.Context, input *awseks.DeleteFargateProfileInput, opts []func(*awseks.Options)) (*awseks.DeleteFargateProfileOutput, error) {
						return nil, errBoom
					},
				},
				cr: fargateProfile(),
			},
			want: want{
				cr:  fargateProfile(withConditions(xpv1.Deleting())),
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
		cr  *v1alpha1.FargateProfile
		err error
	}

	cases := map[string]struct {
		args
		want
	}{
		"Successful": {
			args: args{
				cr:   fargateProfile(withTags(map[string]string{"foo": "bar"})),
				kube: &test.MockClient{MockUpdate: test.NewMockUpdateFn(nil)},
			},
			want: want{
				cr: fargateProfile(withTags(resource.GetExternalTags(fargateProfile()), (map[string]string{"foo": "bar"}))),
			},
		},
		"UpdateFailed": {
			args: args{
				cr:   fargateProfile(),
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
