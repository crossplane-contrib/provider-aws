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
package cachesubnetgroup

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awscache "github.com/aws/aws-sdk-go-v2/service/elasticache"
	awscachetypes "github.com/aws/aws-sdk-go-v2/service/elasticache/types"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"

	"github.com/crossplane-contrib/provider-aws/apis/cache/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/elasticache"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/elasticache/fake"
	errorutils "github.com/crossplane-contrib/provider-aws/pkg/utils/errors"
)

var (
	sgDescription = "some description"
	subnetID      = "some ID"

	// replaceMe = "replace-me!"
	errBoom = errors.New("boom")
)

type args struct {
	cache elasticache.Client
	cr    *v1alpha1.CacheSubnetGroup
}

type csgModifier func(*v1alpha1.CacheSubnetGroup)

func withConditions(c ...xpv1.Condition) csgModifier {
	return func(r *v1alpha1.CacheSubnetGroup) { r.Status.ConditionedStatus.Conditions = c }
}

func withSpec(p v1alpha1.CacheSubnetGroupParameters) csgModifier {
	return func(r *v1alpha1.CacheSubnetGroup) { r.Spec.ForProvider = p }
}

func csg(m ...csgModifier) *v1alpha1.CacheSubnetGroup {
	cr := &v1alpha1.CacheSubnetGroup{}
	for _, f := range m {
		f(cr)
	}
	return cr
}

var _ managed.ExternalClient = &external{}
var _ managed.ExternalConnecter = &connector{}

func TestObserve(t *testing.T) {
	type want struct {
		cr     *v1alpha1.CacheSubnetGroup
		result managed.ExternalObservation
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"SuccessfulAvailable": {
			args: args{
				cache: &fake.MockClient{
					MockDescribeCacheSubnetGroups: func(ctx context.Context, input *awscache.DescribeCacheSubnetGroupsInput, opts []func(*awscache.Options)) (*awscache.DescribeCacheSubnetGroupsOutput, error) {
						return &awscache.DescribeCacheSubnetGroupsOutput{
							CacheSubnetGroups: []awscachetypes.CacheSubnetGroup{{}},
						}, nil
					},
				},
				cr: csg(),
			},
			want: want{
				cr: csg(withConditions(xpv1.Available())),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
				},
			},
		},
		"UpToDate": {
			args: args{
				cache: &fake.MockClient{
					MockDescribeCacheSubnetGroups: func(ctx context.Context, input *awscache.DescribeCacheSubnetGroupsInput, opts []func(*awscache.Options)) (*awscache.DescribeCacheSubnetGroupsOutput, error) {
						return &awscache.DescribeCacheSubnetGroupsOutput{
							CacheSubnetGroups: []awscachetypes.CacheSubnetGroup{{
								CacheSubnetGroupDescription: aws.String(sgDescription),
								Subnets: []awscachetypes.Subnet{
									{
										SubnetIdentifier: aws.String(subnetID),
									},
								},
							}},
						}, nil
					},
				},
				cr: csg(withSpec(v1alpha1.CacheSubnetGroupParameters{
					Description: sgDescription,
					SubnetIDs:   []string{subnetID},
				})),
			},
			want: want{
				cr: csg(withSpec(v1alpha1.CacheSubnetGroupParameters{
					Description: sgDescription,
					SubnetIDs:   []string{subnetID},
				}), withConditions(xpv1.Available())),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
				},
			},
		},
		"DescribeFail": {
			args: args{
				cache: &fake.MockClient{
					MockDescribeCacheSubnetGroups: func(ctx context.Context, input *awscache.DescribeCacheSubnetGroupsInput, opts []func(*awscache.Options)) (*awscache.DescribeCacheSubnetGroupsOutput, error) {
						return nil, errBoom
					},
				},
				cr: csg(),
			},
			want: want{
				cr:  csg(),
				err: errorutils.Wrap(errBoom, errDescribeSubnetGroup),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.cache}
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
		cr     *v1alpha1.CacheSubnetGroup
		result managed.ExternalCreation
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"Successful": {
			args: args{
				cache: &fake.MockClient{
					MockCreateCacheSubnetGroup: func(ctx context.Context, input *awscache.CreateCacheSubnetGroupInput, opts []func(*awscache.Options)) (*awscache.CreateCacheSubnetGroupOutput, error) {
						return &awscache.CreateCacheSubnetGroupOutput{}, nil
					},
				},
				cr: csg(withSpec(v1alpha1.CacheSubnetGroupParameters{
					SubnetIDs:   []string{subnetID},
					Description: sgDescription,
				})),
			},
			want: want{
				cr: csg((withSpec(v1alpha1.CacheSubnetGroupParameters{
					SubnetIDs:   []string{subnetID},
					Description: sgDescription,
				})), withConditions(xpv1.Creating())),
			},
		},
		"CreateFail": {
			args: args{
				cache: &fake.MockClient{
					MockCreateCacheSubnetGroup: func(ctx context.Context, input *awscache.CreateCacheSubnetGroupInput, opts []func(*awscache.Options)) (*awscache.CreateCacheSubnetGroupOutput, error) {
						return nil, errBoom
					},
				},
				cr: csg(withSpec(v1alpha1.CacheSubnetGroupParameters{
					SubnetIDs:   []string{subnetID},
					Description: sgDescription,
				})),
			},
			want: want{
				cr: csg((withSpec(v1alpha1.CacheSubnetGroupParameters{
					SubnetIDs:   []string{subnetID},
					Description: sgDescription,
				})), withConditions(xpv1.Creating())),
				err: errorutils.Wrap(errBoom, errCreateSubnetGroup),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.cache}
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
		cr     *v1alpha1.CacheSubnetGroup
		result managed.ExternalUpdate
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"Successful": {
			args: args{
				cache: &fake.MockClient{
					MockModifyCacheSubnetGroup: func(ctx context.Context, input *awscache.ModifyCacheSubnetGroupInput, opts []func(*awscache.Options)) (*awscache.ModifyCacheSubnetGroupOutput, error) {
						return &awscache.ModifyCacheSubnetGroupOutput{}, nil
					},
				},
				cr: csg(withSpec(v1alpha1.CacheSubnetGroupParameters{
					SubnetIDs:   []string{subnetID},
					Description: sgDescription,
				})),
			},
			want: want{
				cr: csg((withSpec(v1alpha1.CacheSubnetGroupParameters{
					SubnetIDs:   []string{subnetID},
					Description: sgDescription,
				}))),
			},
		},
		"ModifyFailed": {
			args: args{
				cache: &fake.MockClient{
					MockModifyCacheSubnetGroup: func(ctx context.Context, input *awscache.ModifyCacheSubnetGroupInput, opts []func(*awscache.Options)) (*awscache.ModifyCacheSubnetGroupOutput, error) {
						return nil, errBoom
					},
				},
				cr: csg(withSpec(v1alpha1.CacheSubnetGroupParameters{
					SubnetIDs:   []string{subnetID},
					Description: sgDescription,
				})),
			},
			want: want{
				cr: csg((withSpec(v1alpha1.CacheSubnetGroupParameters{
					SubnetIDs:   []string{subnetID},
					Description: sgDescription,
				}))),
				err: errorutils.Wrap(errBoom, errModifySubnetGroup),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.cache}
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
		cr  *v1alpha1.CacheSubnetGroup
		err error
	}

	cases := map[string]struct {
		args
		want
	}{
		"Successful": {
			args: args{
				cache: &fake.MockClient{
					MockDeleteCacheSubnetGroup: func(ctx context.Context, input *awscache.DeleteCacheSubnetGroupInput, opts []func(*awscache.Options)) (*awscache.DeleteCacheSubnetGroupOutput, error) {
						return &awscache.DeleteCacheSubnetGroupOutput{}, nil
					},
				},
				cr: csg(withSpec(v1alpha1.CacheSubnetGroupParameters{
					SubnetIDs:   []string{subnetID},
					Description: sgDescription,
				}), withConditions(xpv1.Deleting())),
			},
			want: want{
				cr: csg((withSpec(v1alpha1.CacheSubnetGroupParameters{
					SubnetIDs:   []string{subnetID},
					Description: sgDescription,
				})), withConditions(xpv1.Deleting())),
			},
		},
		"DeleteFailed": {
			args: args{
				cache: &fake.MockClient{
					MockDeleteCacheSubnetGroup: func(ctx context.Context, input *awscache.DeleteCacheSubnetGroupInput, opts []func(*awscache.Options)) (*awscache.DeleteCacheSubnetGroupOutput, error) {
						return nil, errBoom
					},
				},
				cr: csg(withSpec(v1alpha1.CacheSubnetGroupParameters{
					SubnetIDs:   []string{subnetID},
					Description: sgDescription,
				})),
			},
			want: want{
				cr: csg((withSpec(v1alpha1.CacheSubnetGroupParameters{
					SubnetIDs:   []string{subnetID},
					Description: sgDescription,
				})), withConditions(xpv1.Deleting())),
				err: errorutils.Wrap(errBoom, errDeleteSubnetGroup),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.cache}
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
