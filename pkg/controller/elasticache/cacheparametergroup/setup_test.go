/*
Copyright 2021 The Crossplane Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS_IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cacheparametergroup

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	svcsdk "github.com/aws/aws-sdk-go/service/elasticache"
	"github.com/aws/aws-sdk-go/service/elasticache/elasticacheiface"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/elasticache/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
)

const (
	testCacheParameterGroupName = "some-db-subnet-group"
)

type cacheParameterGroupModifier func(*svcapitypes.CacheParameterGroup)

func cacheParameterGroup(m ...cacheParameterGroupModifier) *svcapitypes.CacheParameterGroup {
	cr := &svcapitypes.CacheParameterGroup{}
	for _, f := range m {
		f(cr)
	}
	return cr
}

func withExternalName(value string) cacheParameterGroupModifier {
	return func(o *svcapitypes.CacheParameterGroup) {
		meta.SetExternalName(o, value)
	}
}

func withCacheParameterGroupName(value string) cacheParameterGroupModifier {
	return func(o *svcapitypes.CacheParameterGroup) {
		o.Status.AtProvider.CacheParameterGroupName = pointer.ToOrNilIfZeroValue(value)
	}
}

func withParameter(k, v string) cacheParameterGroupModifier {
	return func(o *svcapitypes.CacheParameterGroup) {
		o.Spec.ForProvider.ParameterNameValues = append(o.Spec.ForProvider.ParameterNameValues, svcapitypes.ParameterNameValue{
			ParameterName:  pointer.ToOrNilIfZeroValue(k),
			ParameterValue: pointer.ToOrNilIfZeroValue(v),
		})
	}
}

// Define a mock struct to be used in your unit tests of myFunc.
type mockElastiCacheClient struct {
	elasticacheiface.ElastiCacheAPI

	DescribeCacheParametersPagesWithContextFunc func(_ aws.Context, _ *svcsdk.DescribeCacheParametersInput, cb func(*svcsdk.DescribeCacheParametersOutput, bool) bool, _ ...request.Option) error
}

func (m *mockElastiCacheClient) DescribeCacheParametersPagesWithContext(ctx aws.Context, in *svcsdk.DescribeCacheParametersInput, cb func(*svcsdk.DescribeCacheParametersOutput, bool) bool, opts ...request.Option) error {
	return m.DescribeCacheParametersPagesWithContextFunc(ctx, in, cb, opts...)
}

func TestIsUpToDate(t *testing.T) {
	type want struct {
		upToDate bool
		wantErr  error
	}

	type args struct {
		elasticache elasticacheiface.ElastiCacheAPI
		cr          *svcapitypes.CacheParameterGroup
		resp        *svcsdk.DescribeCacheParameterGroupsOutput
	}

	cases := map[string]struct {
		args args
		want want
	}{
		"upToDateSort": {
			args: args{
				elasticache: &mockElastiCacheClient{
					DescribeCacheParametersPagesWithContextFunc: func(_ aws.Context, _ *svcsdk.DescribeCacheParametersInput, cb func(*svcsdk.DescribeCacheParametersOutput, bool) bool, _ ...request.Option) error {
						cb(&svcsdk.DescribeCacheParametersOutput{
							Parameters: []*svcsdk.Parameter{
								{
									Source:         pointer.ToOrNilIfZeroValue(svcsdk.SourceTypeUser),
									ParameterName:  pointer.ToOrNilIfZeroValue("c"),
									ParameterValue: pointer.ToOrNilIfZeroValue("val3"),
								},
								{
									Source:         pointer.ToOrNilIfZeroValue(svcsdk.SourceTypeUser),
									ParameterName:  pointer.ToOrNilIfZeroValue("a"),
									ParameterValue: pointer.ToOrNilIfZeroValue("val1"),
								},
								{
									Source:         pointer.ToOrNilIfZeroValue(svcsdk.SourceTypeUser),
									ParameterName:  pointer.ToOrNilIfZeroValue("b"),
									ParameterValue: pointer.ToOrNilIfZeroValue("val2"),
								},
								{
									Source:         pointer.ToOrNilIfZeroValue(svcsdk.SourceTypeCacheParameterGroup),
									ParameterName:  pointer.ToOrNilIfZeroValue("as-default"),
									ParameterValue: pointer.ToOrNilIfZeroValue("untouched"),
								},
							},
						}, true)
						return nil
					},
				},
				cr: cacheParameterGroup(
					withCacheParameterGroupName(testCacheParameterGroupName),
					withExternalName(testCacheParameterGroupName),
					withParameter("c", "val3"),
					withParameter("b", "val2"),
					withParameter("a", "val1"),
				),
			},
			want: want{
				upToDate: true,
			},
		},
		"upToDateDiff": {
			args: args{
				elasticache: &mockElastiCacheClient{
					DescribeCacheParametersPagesWithContextFunc: func(_ aws.Context, _ *svcsdk.DescribeCacheParametersInput, cb func(*svcsdk.DescribeCacheParametersOutput, bool) bool, _ ...request.Option) error {
						cb(&svcsdk.DescribeCacheParametersOutput{
							Parameters: []*svcsdk.Parameter{
								{
									Source:         pointer.ToOrNilIfZeroValue(svcsdk.SourceTypeUser),
									ParameterName:  pointer.ToOrNilIfZeroValue("a"),
									ParameterValue: pointer.ToOrNilIfZeroValue("valx"),
								},
								{
									Source:         pointer.ToOrNilIfZeroValue(svcsdk.SourceTypeUser),
									ParameterName:  pointer.ToOrNilIfZeroValue("b"),
									ParameterValue: pointer.ToOrNilIfZeroValue("val2"),
								},
							},
						}, true)
						return nil
					},
				},
				cr: cacheParameterGroup(
					withCacheParameterGroupName(testCacheParameterGroupName),
					withExternalName(testCacheParameterGroupName),
					withParameter("a", "val1"),
					withParameter("b", "val2"),
				),
			},
			want: want{
				upToDate: false,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			opts := []option{setupExternal}
			e := newExternal(nil, tc.args.elasticache, opts)
			upToDate, _, err := e.isUpToDate(context.Background(), tc.args.cr, tc.args.resp)

			if diff := cmp.Diff(tc.want.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}

			if diff := cmp.Diff(tc.want.upToDate, upToDate); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}
