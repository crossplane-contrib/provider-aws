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

package cluster

import (
	"context"
	"net/http"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awscache "github.com/aws/aws-sdk-go-v2/service/elasticache"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/test"

	"github.com/crossplane/provider-aws/apis/cache/v1alpha1"
	awsclient "github.com/crossplane/provider-aws/pkg/clients"
	"github.com/crossplane/provider-aws/pkg/clients/elasticache"
	"github.com/crossplane/provider-aws/pkg/clients/elasticache/fake"
)

var (
	externalName = "somecluster"
	nodeType     = "t2.small"

	errBoom = errors.New("boom")
)

type args struct {
	cache elasticache.Client
	cr    *v1alpha1.CacheCluster
}

type clusterModifier func(*v1alpha1.CacheCluster)

func withExternalName() clusterModifier {
	return func(c *v1alpha1.CacheCluster) { meta.SetExternalName(c, externalName) }
}

func withConditions(c ...xpv1.Condition) clusterModifier {
	return func(r *v1alpha1.CacheCluster) { r.Status.ConditionedStatus.Conditions = c }
}

func withSpec(p v1alpha1.CacheClusterParameters) clusterModifier {
	return func(r *v1alpha1.CacheCluster) { r.Spec.ForProvider = p }
}

func withStatus(s v1alpha1.CacheClusterObservation) clusterModifier {
	return func(r *v1alpha1.CacheCluster) { r.Status.AtProvider = s }
}

func cluster(m ...clusterModifier) *v1alpha1.CacheCluster {
	cr := &v1alpha1.CacheCluster{}
	for _, f := range m {
		f(cr)
	}
	return cr
}

var _ managed.ExternalClient = &external{}
var _ managed.ExternalConnecter = &connector{}

func TestObserve(t *testing.T) {
	type want struct {
		cr     *v1alpha1.CacheCluster
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
					MockDescribeCacheClustersRequest: func(input *awscache.DescribeCacheClustersInput) awscache.DescribeCacheClustersRequest {
						return awscache.DescribeCacheClustersRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awscache.DescribeCacheClustersOutput{
								CacheClusters: []awscache.CacheCluster{{
									CacheClusterStatus: aws.String(v1alpha1.StatusCreating),
								}},
							}},
						}
					},
				},
				cr: cluster(withExternalName(),
					withSpec(v1alpha1.CacheClusterParameters{
						CacheNodeType: nodeType,
						NumCacheNodes: 2,
					})),
			},
			want: want{
				cr: cluster(withConditions(xpv1.Creating()),
					withExternalName(),
					withSpec(v1alpha1.CacheClusterParameters{
						CacheNodeType: nodeType,
						NumCacheNodes: 2,
					}),
					withStatus(v1alpha1.CacheClusterObservation{
						CacheClusterStatus: v1alpha1.StatusCreating,
					})),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: false,
				},
			},
		},
		"UpToDate": {
			args: args{
				cache: &fake.MockClient{
					MockDescribeCacheClustersRequest: func(input *awscache.DescribeCacheClustersInput) awscache.DescribeCacheClustersRequest {
						return awscache.DescribeCacheClustersRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awscache.DescribeCacheClustersOutput{
								CacheClusters: []awscache.CacheCluster{{
									CacheClusterStatus: aws.String(v1alpha1.StatusAvailable),
									CacheNodeType:      aws.String(nodeType),
									NumCacheNodes:      aws.Int64(2),
									CacheClusterId:     aws.String(externalName),
								}},
							}},
						}
					},
				},
				cr: cluster(withExternalName(),
					withSpec(v1alpha1.CacheClusterParameters{
						CacheNodeType: nodeType,
						NumCacheNodes: 2,
					})),
			},
			want: want{
				cr: cluster(withConditions(xpv1.Available()),
					withExternalName(),
					withSpec(v1alpha1.CacheClusterParameters{
						CacheNodeType: nodeType,
						NumCacheNodes: 2,
					}),
					withStatus(v1alpha1.CacheClusterObservation{
						CacheClusterStatus: v1alpha1.StatusAvailable,
					})),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
				},
			},
		},
		"DescribeFail": {
			args: args{
				cache: &fake.MockClient{
					MockDescribeCacheClustersRequest: func(input *awscache.DescribeCacheClustersInput) awscache.DescribeCacheClustersRequest {
						return awscache.DescribeCacheClustersRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: errBoom},
						}
					},
				},
				cr: cluster(),
			},
			want: want{
				cr:  cluster(),
				err: awsclient.Wrap(errBoom, errDescribeCacheCluster),
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
		cr     *v1alpha1.CacheCluster
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
					MockCreateCacheClusterRequest: func(input *awscache.CreateCacheClusterInput) awscache.CreateCacheClusterRequest {
						return awscache.CreateCacheClusterRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awscache.CreateCacheClusterOutput{}},
						}
					},
				},
				cr: cluster(withExternalName(),
					withSpec(v1alpha1.CacheClusterParameters{
						CacheNodeType: nodeType,
						NumCacheNodes: 2,
					})),
			},
			want: want{
				cr: cluster(withExternalName(),
					withSpec(v1alpha1.CacheClusterParameters{
						CacheNodeType: nodeType,
						NumCacheNodes: 2,
					}), withConditions(xpv1.Creating())),
			},
		},
		"CreateFail": {
			args: args{
				cache: &fake.MockClient{
					MockCreateCacheClusterRequest: func(input *awscache.CreateCacheClusterInput) awscache.CreateCacheClusterRequest {
						return awscache.CreateCacheClusterRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Error: errBoom},
						}
					},
				},
				cr: cluster(withSpec(v1alpha1.CacheClusterParameters{
					CacheNodeType: nodeType,
					NumCacheNodes: 2,
				})),
			},
			want: want{
				cr: cluster(withSpec(v1alpha1.CacheClusterParameters{
					CacheNodeType: nodeType,
					NumCacheNodes: 2,
				}), withConditions(xpv1.Creating())),
				err: awsclient.Wrap(errBoom, errCreateCacheCluster),
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
		cr     *v1alpha1.CacheCluster
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
					MockModifyCacheClusterRequest: func(input *awscache.ModifyCacheClusterInput) awscache.ModifyCacheClusterRequest {
						return awscache.ModifyCacheClusterRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awscache.ModifyCacheClusterOutput{}},
						}
					},
				},
				cr: cluster(withExternalName(),
					withSpec(v1alpha1.CacheClusterParameters{
						CacheNodeType: nodeType,
						NumCacheNodes: 3,
					})),
			},
			want: want{
				cr: cluster(withExternalName(),
					withSpec(v1alpha1.CacheClusterParameters{
						CacheNodeType: nodeType,
						NumCacheNodes: 3,
					})),
			},
		},
		"ModifyFailed": {
			args: args{
				cache: &fake.MockClient{
					MockModifyCacheClusterRequest: func(input *awscache.ModifyCacheClusterInput) awscache.ModifyCacheClusterRequest {
						return awscache.ModifyCacheClusterRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: errBoom},
						}
					},
				},
				cr: cluster(withExternalName(),
					withSpec(v1alpha1.CacheClusterParameters{
						CacheNodeType: nodeType,
						NumCacheNodes: 3,
					}),
					withStatus(v1alpha1.CacheClusterObservation{
						CacheClusterStatus: v1alpha1.StatusAvailable,
					})),
			},
			want: want{
				cr: cluster(withExternalName(),
					withSpec(v1alpha1.CacheClusterParameters{
						CacheNodeType: nodeType,
						NumCacheNodes: 3,
					}),
					withStatus(v1alpha1.CacheClusterObservation{
						CacheClusterStatus: v1alpha1.StatusAvailable,
					})),
				err: awsclient.Wrap(errBoom, errModifyCacheCluster),
			},
		},
		"NotAvailable": {
			args: args{
				cache: &fake.MockClient{
					MockModifyCacheClusterRequest: func(input *awscache.ModifyCacheClusterInput) awscache.ModifyCacheClusterRequest {
						return awscache.ModifyCacheClusterRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: errBoom},
						}
					},
				},
				cr: cluster(withExternalName(),
					withSpec(v1alpha1.CacheClusterParameters{
						CacheNodeType: nodeType,
						NumCacheNodes: 3,
					})),
			},
			want: want{
				cr: cluster(withExternalName(),
					withSpec(v1alpha1.CacheClusterParameters{
						CacheNodeType: nodeType,
						NumCacheNodes: 3,
					})),
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
		cr  *v1alpha1.CacheCluster
		err error
	}

	cases := map[string]struct {
		args
		want
	}{
		"Successful": {
			args: args{
				cache: &fake.MockClient{
					MockDeleteCacheClusterRequest: func(input *awscache.DeleteCacheClusterInput) awscache.DeleteCacheClusterRequest {
						return awscache.DeleteCacheClusterRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awscache.DeleteCacheClusterOutput{}},
						}
					},
				},
				cr: cluster(withExternalName(), withConditions(xpv1.Deleting())),
			},
			want: want{
				cr: cluster(withExternalName(),
					withConditions(xpv1.Deleting())),
			},
		},
		"DeleteFailed": {
			args: args{
				cache: &fake.MockClient{
					MockDeleteCacheClusterRequest: func(input *awscache.DeleteCacheClusterInput) awscache.DeleteCacheClusterRequest {
						return awscache.DeleteCacheClusterRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Error: errBoom},
						}
					},
				},
				cr: cluster(withExternalName()),
			},
			want: want{
				cr: cluster(withExternalName(),
					withConditions(xpv1.Deleting())),
				err: awsclient.Wrap(errBoom, errDeleteCacheCluster),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.cache}
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
