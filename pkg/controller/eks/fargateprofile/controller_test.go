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
	"net/http"
	"testing"

	awseks "github.com/aws/aws-sdk-go-v2/service/eks"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/test"

	"github.com/crossplane/provider-aws/apis/eks/v1alpha1"
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

func withConditions(c ...runtimev1alpha1.Condition) fargateProfileModifier {
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
					MockDescribeFargateProfileRequest: func(_ *awseks.DescribeFargateProfileInput) awseks.DescribeFargateProfileRequest {
						return awseks.DescribeFargateProfileRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awseks.DescribeFargateProfileOutput{
								FargateProfile: &awseks.FargateProfile{
									Status: awseks.FargateProfileStatusActive,
								},
							}},
						}
					},
				},
				cr: fargateProfile(),
			},
			want: want{
				cr: fargateProfile(
					withConditions(runtimev1alpha1.Available()),
					withStatus(v1alpha1.FargateProfileStatusActive)),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
				},
			},
		},
		"DeletingState": {
			args: args{
				eks: &fake.MockClient{
					MockDescribeFargateProfileRequest: func(_ *awseks.DescribeFargateProfileInput) awseks.DescribeFargateProfileRequest {
						return awseks.DescribeFargateProfileRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awseks.DescribeFargateProfileOutput{
								FargateProfile: &awseks.FargateProfile{
									Status: awseks.FargateProfileStatusDeleting,
								},
							}},
						}
					},
				},
				cr: fargateProfile(),
			},
			want: want{
				cr: fargateProfile(
					withConditions(runtimev1alpha1.Deleting()),
					withStatus(v1alpha1.FargateProfileStatusDeleting)),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
				},
			},
		},
		"FailedDescribeRequest": {
			args: args{
				eks: &fake.MockClient{
					MockDescribeFargateProfileRequest: func(_ *awseks.DescribeFargateProfileInput) awseks.DescribeFargateProfileRequest {
						return awseks.DescribeFargateProfileRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: errBoom},
						}
					},
				},
				cr: fargateProfile(),
			},
			want: want{
				cr:  fargateProfile(),
				err: errors.Wrap(errBoom, errDescribeFailed),
			},
		},
		"NotFound": {
			args: args{
				eks: &fake.MockClient{
					MockDescribeFargateProfileRequest: func(_ *awseks.DescribeFargateProfileInput) awseks.DescribeFargateProfileRequest {
						return awseks.DescribeFargateProfileRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: errors.New(awseks.ErrCodeResourceNotFoundException)},
						}
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
					MockDescribeFargateProfileRequest: func(_ *awseks.DescribeFargateProfileInput) awseks.DescribeFargateProfileRequest {
						return awseks.DescribeFargateProfileRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awseks.DescribeFargateProfileOutput{
								FargateProfile: &awseks.FargateProfile{
									Status:  awseks.FargateProfileStatusCreating,
									Subnets: subnets,
								},
							}},
						}
					},
				},
				cr: fargateProfile(),
			},
			want: want{
				cr: fargateProfile(
					withStatus(v1alpha1.FargateProfileStatusCreating),
					withConditions(runtimev1alpha1.Creating()),
					withSubnets(subnets),
				),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
				},
			},
		},
		"LateInitFailedKubeUpdate": {
			args: args{
				kube: &test.MockClient{
					MockUpdate: test.NewMockUpdateFn(errBoom),
				},
				eks: &fake.MockClient{
					MockDescribeFargateProfileRequest: func(_ *awseks.DescribeFargateProfileInput) awseks.DescribeFargateProfileRequest {
						return awseks.DescribeFargateProfileRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awseks.DescribeFargateProfileOutput{
								FargateProfile: &awseks.FargateProfile{
									Status:  awseks.FargateProfileStatusCreating,
									Subnets: subnets,
								},
							}},
						}
					},
				},
				cr: fargateProfile(),
			},
			want: want{
				cr:  fargateProfile(withSubnets(subnets)),
				err: errors.Wrap(errBoom, errKubeUpdateFailed),
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
					MockCreateFargateProfileRequest: func(input *awseks.CreateFargateProfileInput) awseks.CreateFargateProfileRequest {
						return awseks.CreateFargateProfileRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awseks.CreateFargateProfileOutput{}},
						}
					},
				},
				cr: fargateProfile(),
			},
			want: want{
				cr:     fargateProfile(withConditions(runtimev1alpha1.Creating())),
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
					withConditions(runtimev1alpha1.Creating())),
			},
		},
		"FailedRequest": {
			args: args{
				eks: &fake.MockClient{
					MockCreateFargateProfileRequest: func(input *awseks.CreateFargateProfileInput) awseks.CreateFargateProfileRequest {
						return awseks.CreateFargateProfileRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: errBoom},
						}
					},
				},
				cr: fargateProfile(),
			},
			want: want{
				cr:  fargateProfile(withConditions(runtimev1alpha1.Creating())),
				err: errors.Wrap(errBoom, errCreateFailed),
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
					MockDescribeFargateProfileRequest: func(input *awseks.DescribeFargateProfileInput) awseks.DescribeFargateProfileRequest {
						return awseks.DescribeFargateProfileRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awseks.DescribeFargateProfileOutput{
								FargateProfile: &awseks.FargateProfile{},
							}},
						}
					},
					MockTagResourceRequest: func(input *awseks.TagResourceInput) awseks.TagResourceRequest {
						return awseks.TagResourceRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awseks.TagResourceOutput{}},
						}
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
					MockDescribeFargateProfileRequest: func(input *awseks.DescribeFargateProfileInput) awseks.DescribeFargateProfileRequest {
						return awseks.DescribeFargateProfileRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awseks.DescribeFargateProfileOutput{
								FargateProfile: &awseks.FargateProfile{},
							}},
						}
					},
					MockUntagResourceRequest: func(input *awseks.UntagResourceInput) awseks.UntagResourceRequest {
						return awseks.UntagResourceRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awseks.UntagResourceOutput{}},
						}
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
					MockDescribeFargateProfileRequest: func(input *awseks.DescribeFargateProfileInput) awseks.DescribeFargateProfileRequest {
						return awseks.DescribeFargateProfileRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awseks.DescribeFargateProfileOutput{
								FargateProfile: &awseks.FargateProfile{
									Tags: map[string]string{"foo": "bar"},
								},
							}},
						}
					},
					MockUntagResourceRequest: func(input *awseks.UntagResourceInput) awseks.UntagResourceRequest {
						return awseks.UntagResourceRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: errBoom},
						}
					},
				},
				cr: fargateProfile(),
			},
			want: want{
				cr:  fargateProfile(),
				err: errors.Wrap(errBoom, errAddTagsFailed),
			},
		},
		"FailedAddTags": {
			args: args{
				eks: &fake.MockClient{
					MockDescribeFargateProfileRequest: func(input *awseks.DescribeFargateProfileInput) awseks.DescribeFargateProfileRequest {
						return awseks.DescribeFargateProfileRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awseks.DescribeFargateProfileOutput{
								FargateProfile: &awseks.FargateProfile{},
							}},
						}
					},
					MockTagResourceRequest: func(input *awseks.TagResourceInput) awseks.TagResourceRequest {
						return awseks.TagResourceRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: errBoom},
						}
					},
				},
				cr: fargateProfile(withTags(map[string]string{"foo": "bar"})),
			},
			want: want{
				cr:  fargateProfile(withTags(map[string]string{"foo": "bar"})),
				err: errors.Wrap(errBoom, errAddTagsFailed),
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
					MockDeleteFargateProfileRequest: func(input *awseks.DeleteFargateProfileInput) awseks.DeleteFargateProfileRequest {
						return awseks.DeleteFargateProfileRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awseks.DeleteFargateProfileOutput{}},
						}
					},
				},
				cr: fargateProfile(),
			},
			want: want{
				cr: fargateProfile(withConditions(runtimev1alpha1.Deleting())),
			},
		},
		"AlreadyDeleting": {
			args: args{
				cr: fargateProfile(withStatus(v1alpha1.FargateProfileStatusDeleting)),
			},
			want: want{
				cr: fargateProfile(withStatus(v1alpha1.FargateProfileStatusDeleting),
					withConditions(runtimev1alpha1.Deleting())),
			},
		},
		"AlreadyDeleted": {
			args: args{
				eks: &fake.MockClient{
					MockDeleteFargateProfileRequest: func(input *awseks.DeleteFargateProfileInput) awseks.DeleteFargateProfileRequest {
						return awseks.DeleteFargateProfileRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: errors.New(awseks.ErrCodeResourceNotFoundException)},
						}
					},
				},
				cr: fargateProfile(),
			},
			want: want{
				cr: fargateProfile(withConditions(runtimev1alpha1.Deleting())),
			},
		},
		"Failed": {
			args: args{
				eks: &fake.MockClient{
					MockDeleteFargateProfileRequest: func(input *awseks.DeleteFargateProfileInput) awseks.DeleteFargateProfileRequest {
						return awseks.DeleteFargateProfileRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: errBoom},
						}
					},
				},
				cr: fargateProfile(),
			},
			want: want{
				cr:  fargateProfile(withConditions(runtimev1alpha1.Deleting())),
				err: errors.Wrap(errBoom, errDeleteFailed),
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
