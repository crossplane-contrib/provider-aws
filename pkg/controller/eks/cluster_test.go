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

package eks

import (
	"context"
	"net/http"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awseks "github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/test"

	"github.com/crossplane/provider-aws/apis/eks/v1beta1"
	"github.com/crossplane/provider-aws/pkg/clients/eks"
	"github.com/crossplane/provider-aws/pkg/clients/eks/fake"
)

var (
	version = "1.16"

	errBoom = errors.New("boom")
)

type args struct {
	eks  eks.Client
	kube client.Client
	cr   *v1beta1.Cluster
}

type clusterModifier func(*v1beta1.Cluster)

func withConditions(c ...runtimev1alpha1.Condition) clusterModifier {
	return func(r *v1beta1.Cluster) { r.Status.ConditionedStatus.Conditions = c }
}

func withTags(tagMaps ...map[string]string) clusterModifier {
	tags := map[string]string{}
	for _, tagMap := range tagMaps {
		for k, v := range tagMap {
			tags[k] = v
		}
	}
	return func(r *v1beta1.Cluster) { r.Spec.ForProvider.Tags = tags }
}

func withVersion(v *string) clusterModifier {
	return func(r *v1beta1.Cluster) { r.Spec.ForProvider.Version = v }
}

func withStatus(s v1beta1.ClusterStatusType) clusterModifier {
	return func(r *v1beta1.Cluster) { r.Status.AtProvider.Status = s }
}

func withConfig(c v1beta1.VpcConfigRequest) clusterModifier {
	return func(r *v1beta1.Cluster) { r.Spec.ForProvider.ResourcesVpcConfig = c }
}

func cluster(m ...clusterModifier) *v1beta1.Cluster {
	cr := &v1beta1.Cluster{}
	for _, f := range m {
		f(cr)
	}
	return cr
}

var _ managed.ExternalClient = &external{}
var _ managed.ExternalConnecter = &connector{}

func TestObserve(t *testing.T) {
	type want struct {
		cr     *v1beta1.Cluster
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
					MockDescribeClusterRequest: func(_ *awseks.DescribeClusterInput) awseks.DescribeClusterRequest {
						return awseks.DescribeClusterRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awseks.DescribeClusterOutput{
								Cluster: &awseks.Cluster{
									Status: awseks.ClusterStatusActive,
								},
							}},
						}
					},
				},
				cr: cluster(),
			},
			want: want{
				cr: cluster(
					withConditions(runtimev1alpha1.Available()),
					withStatus(v1beta1.ClusterStatusActive)),
				result: managed.ExternalObservation{
					ResourceExists:    true,
					ResourceUpToDate:  true,
					ConnectionDetails: eks.GetConnectionDetails(&awseks.Cluster{}, &sts.Client{}),
				},
			},
		},
		"DeletingState": {
			args: args{
				eks: &fake.MockClient{
					MockDescribeClusterRequest: func(input *awseks.DescribeClusterInput) awseks.DescribeClusterRequest {
						return awseks.DescribeClusterRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awseks.DescribeClusterOutput{
								Cluster: &awseks.Cluster{
									Status: awseks.ClusterStatusDeleting,
								},
							}},
						}
					},
				},
				cr: cluster(),
			},
			want: want{
				cr: cluster(
					withConditions(runtimev1alpha1.Deleting()),
					withStatus(v1beta1.ClusterStatusDeleting)),
				result: managed.ExternalObservation{
					ResourceExists:    true,
					ResourceUpToDate:  true,
					ConnectionDetails: eks.GetConnectionDetails(&awseks.Cluster{}, &sts.Client{}),
				},
			},
		},
		"FailedState": {
			args: args{
				eks: &fake.MockClient{
					MockDescribeClusterRequest: func(input *awseks.DescribeClusterInput) awseks.DescribeClusterRequest {
						return awseks.DescribeClusterRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awseks.DescribeClusterOutput{
								Cluster: &awseks.Cluster{
									Status: awseks.ClusterStatusFailed,
								},
							}},
						}
					},
				},
				cr: cluster(),
			},
			want: want{
				cr: cluster(
					withConditions(runtimev1alpha1.Unavailable()),
					withStatus(v1beta1.ClusterStatusFailed)),
				result: managed.ExternalObservation{
					ResourceExists:    true,
					ResourceUpToDate:  true,
					ConnectionDetails: eks.GetConnectionDetails(&awseks.Cluster{}, &sts.Client{}),
				},
			},
		},
		"FailedDescribeRequest": {
			args: args{
				eks: &fake.MockClient{
					MockDescribeClusterRequest: func(input *awseks.DescribeClusterInput) awseks.DescribeClusterRequest {
						return awseks.DescribeClusterRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: errBoom},
						}
					},
				},
				cr: cluster(),
			},
			want: want{
				cr:  cluster(),
				err: errors.Wrap(errBoom, errDescribeFailed),
			},
		},
		"NotFound": {
			args: args{
				eks: &fake.MockClient{
					MockDescribeClusterRequest: func(input *awseks.DescribeClusterInput) awseks.DescribeClusterRequest {
						return awseks.DescribeClusterRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: errors.New(awseks.ErrCodeResourceNotFoundException)},
						}
					},
				},
				cr: cluster(),
			},
			want: want{
				cr: cluster(),
			},
		},
		"LateInitSuccess": {
			args: args{
				kube: &test.MockClient{
					MockUpdate: test.NewMockUpdateFn(nil),
				},
				eks: &fake.MockClient{
					MockDescribeClusterRequest: func(input *awseks.DescribeClusterInput) awseks.DescribeClusterRequest {
						return awseks.DescribeClusterRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awseks.DescribeClusterOutput{
								Cluster: &awseks.Cluster{
									Status:  awseks.ClusterStatusCreating,
									Version: &version,
								},
							}},
						}
					},
				},
				cr: cluster(),
			},
			want: want{
				cr: cluster(
					withStatus(v1beta1.ClusterStatusCreating),
					withConditions(runtimev1alpha1.Creating()),
					withVersion(&version),
				),
				result: managed.ExternalObservation{
					ResourceExists:    true,
					ResourceUpToDate:  true,
					ConnectionDetails: eks.GetConnectionDetails(&awseks.Cluster{}, &sts.Client{}),
				},
			},
		},
		"LateInitFailedKubeUpdate": {
			args: args{
				kube: &test.MockClient{
					MockUpdate: test.NewMockUpdateFn(errBoom),
				},
				eks: &fake.MockClient{
					MockDescribeClusterRequest: func(input *awseks.DescribeClusterInput) awseks.DescribeClusterRequest {
						return awseks.DescribeClusterRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awseks.DescribeClusterOutput{
								Cluster: &awseks.Cluster{
									Status:  awseks.ClusterStatusCreating,
									Version: &version,
								},
							}},
						}
					},
				},
				cr: cluster(),
			},
			want: want{
				cr:  cluster(withVersion(&version)),
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
		cr     *v1beta1.Cluster
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
					MockCreateClusterRequest: func(input *awseks.CreateClusterInput) awseks.CreateClusterRequest {
						return awseks.CreateClusterRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awseks.CreateClusterOutput{}},
						}
					},
				},
				cr: cluster(),
			},
			want: want{
				cr:     cluster(withConditions(runtimev1alpha1.Creating())),
				result: managed.ExternalCreation{},
			},
		},
		"SuccessfulNoNeedForCreate": {
			args: args{
				cr: cluster(withStatus(v1beta1.ClusterStatusCreating)),
			},
			want: want{
				cr: cluster(
					withStatus(v1beta1.ClusterStatusCreating),
					withConditions(runtimev1alpha1.Creating())),
			},
		},
		"FailedRequest": {
			args: args{
				eks: &fake.MockClient{
					MockCreateClusterRequest: func(input *awseks.CreateClusterInput) awseks.CreateClusterRequest {
						return awseks.CreateClusterRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: errBoom},
						}
					},
				},
				cr: cluster(),
			},
			want: want{
				cr:  cluster(withConditions(runtimev1alpha1.Creating())),
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
		cr     *v1beta1.Cluster
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
					MockDescribeClusterRequest: func(input *awseks.DescribeClusterInput) awseks.DescribeClusterRequest {
						return awseks.DescribeClusterRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awseks.DescribeClusterOutput{
								Cluster: &awseks.Cluster{},
							}},
						}
					},
					MockUpdateClusterConfigRequest: func(input *awseks.UpdateClusterConfigInput) awseks.UpdateClusterConfigRequest {
						return awseks.UpdateClusterConfigRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awseks.UpdateClusterConfigOutput{}},
						}
					},
					MockTagResourceRequest: func(input *awseks.TagResourceInput) awseks.TagResourceRequest {
						return awseks.TagResourceRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awseks.TagResourceOutput{}},
						}
					},
				},
				cr: cluster(
					withTags(map[string]string{"foo": "bar"})),
			},
			want: want{
				cr: cluster(
					withTags(map[string]string{"foo": "bar"})),
			},
		},
		"SuccessfulRemoveTags": {
			args: args{
				eks: &fake.MockClient{
					MockDescribeClusterRequest: func(input *awseks.DescribeClusterInput) awseks.DescribeClusterRequest {
						return awseks.DescribeClusterRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awseks.DescribeClusterOutput{
								Cluster: &awseks.Cluster{
									Tags: map[string]string{"foo": "bar"},
								},
							}},
						}
					},
					MockUpdateClusterConfigRequest: func(input *awseks.UpdateClusterConfigInput) awseks.UpdateClusterConfigRequest {
						return awseks.UpdateClusterConfigRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awseks.UpdateClusterConfigOutput{}},
						}
					},
					MockUntagResourceRequest: func(input *awseks.UntagResourceInput) awseks.UntagResourceRequest {
						return awseks.UntagResourceRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awseks.UntagResourceOutput{}},
						}
					},
				},
				cr: cluster(),
			},
			want: want{
				cr: cluster(),
			},
		},
		"SuccessfulUpdateVersion": {
			args: args{
				eks: &fake.MockClient{
					MockUpdateClusterVersionRequest: func(input *awseks.UpdateClusterVersionInput) awseks.UpdateClusterVersionRequest {
						return awseks.UpdateClusterVersionRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awseks.UpdateClusterVersionOutput{}},
						}
					},
					MockDescribeClusterRequest: func(input *awseks.DescribeClusterInput) awseks.DescribeClusterRequest {
						return awseks.DescribeClusterRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awseks.DescribeClusterOutput{
								Cluster: &awseks.Cluster{},
							}},
						}
					},
				},
				cr: cluster(withVersion(&version)),
			},
			want: want{
				cr: cluster(withVersion(&version)),
			},
		},
		"SuccessfulUpdateCluster": {
			args: args{
				eks: &fake.MockClient{
					MockUpdateClusterConfigRequest: func(input *awseks.UpdateClusterConfigInput) awseks.UpdateClusterConfigRequest {
						return awseks.UpdateClusterConfigRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awseks.UpdateClusterConfigOutput{}},
						}
					},
					MockDescribeClusterRequest: func(input *awseks.DescribeClusterInput) awseks.DescribeClusterRequest {
						return awseks.DescribeClusterRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awseks.DescribeClusterOutput{
								Cluster: &awseks.Cluster{},
							}},
						}
					},
				},
				cr: cluster(withConfig(v1beta1.VpcConfigRequest{SubnetIDs: []string{"subnet"}})),
			},
			want: want{
				cr: cluster(withConfig(v1beta1.VpcConfigRequest{SubnetIDs: []string{"subnet"}})),
			},
		},
		"AlreadyModifying": {
			args: args{
				cr: cluster(withStatus(v1beta1.ClusterStatusUpdating)),
			},
			want: want{
				cr: cluster(withStatus(v1beta1.ClusterStatusUpdating)),
			},
		},
		"FailedDescribe": {
			args: args{
				eks: &fake.MockClient{
					MockDescribeClusterRequest: func(input *awseks.DescribeClusterInput) awseks.DescribeClusterRequest {
						return awseks.DescribeClusterRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: errBoom},
						}
					},
				},
				cr: cluster(),
			},
			want: want{
				cr:  cluster(),
				err: errors.Wrap(errBoom, errDescribeFailed),
			},
		},
		"FailedUpdateConfig": {
			args: args{
				eks: &fake.MockClient{
					MockUpdateClusterConfigRequest: func(input *awseks.UpdateClusterConfigInput) awseks.UpdateClusterConfigRequest {
						return awseks.UpdateClusterConfigRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: errBoom},
						}
					},
					MockDescribeClusterRequest: func(input *awseks.DescribeClusterInput) awseks.DescribeClusterRequest {
						return awseks.DescribeClusterRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awseks.DescribeClusterOutput{
								Cluster: &awseks.Cluster{},
							}},
						}
					},
				},
				cr: cluster(),
			},
			want: want{
				cr:  cluster(),
				err: errors.Wrap(errBoom, errUpdateConfigFailed),
			},
		},
		"FailedUpdateVersion": {
			args: args{
				eks: &fake.MockClient{
					MockUpdateClusterVersionRequest: func(input *awseks.UpdateClusterVersionInput) awseks.UpdateClusterVersionRequest {
						return awseks.UpdateClusterVersionRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: errBoom},
						}
					},
					MockDescribeClusterRequest: func(input *awseks.DescribeClusterInput) awseks.DescribeClusterRequest {
						return awseks.DescribeClusterRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awseks.DescribeClusterOutput{
								Cluster: &awseks.Cluster{},
							}},
						}
					},
				},
				cr: cluster(withVersion(&version)),
			},
			want: want{
				cr:  cluster(withVersion(&version)),
				err: errors.Wrap(errBoom, errUpdateVersionFailed),
			},
		},
		"FailedRemoveTags": {
			args: args{
				eks: &fake.MockClient{
					MockDescribeClusterRequest: func(input *awseks.DescribeClusterInput) awseks.DescribeClusterRequest {
						return awseks.DescribeClusterRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awseks.DescribeClusterOutput{
								Cluster: &awseks.Cluster{
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
				cr: cluster(),
			},
			want: want{
				cr:  cluster(),
				err: errors.Wrap(errBoom, errAddTagsFailed),
			},
		},
		"FailedAddTags": {
			args: args{
				eks: &fake.MockClient{
					MockDescribeClusterRequest: func(input *awseks.DescribeClusterInput) awseks.DescribeClusterRequest {
						return awseks.DescribeClusterRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awseks.DescribeClusterOutput{
								Cluster: &awseks.Cluster{},
							}},
						}
					},
					MockTagResourceRequest: func(input *awseks.TagResourceInput) awseks.TagResourceRequest {
						return awseks.TagResourceRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: errBoom},
						}
					},
				},
				cr: cluster(withTags(map[string]string{"foo": "bar"})),
			},
			want: want{
				cr:  cluster(withTags(map[string]string{"foo": "bar"})),
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
		cr  *v1beta1.Cluster
		err error
	}

	cases := map[string]struct {
		args
		want
	}{
		"Successful": {
			args: args{
				eks: &fake.MockClient{
					MockDeleteClusterRequest: func(input *awseks.DeleteClusterInput) awseks.DeleteClusterRequest {
						return awseks.DeleteClusterRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awseks.DeleteClusterOutput{}},
						}
					},
				},
				cr: cluster(),
			},
			want: want{
				cr: cluster(withConditions(runtimev1alpha1.Deleting())),
			},
		},
		"AlreadyDeleting": {
			args: args{
				cr: cluster(withStatus(v1beta1.ClusterStatusDeleting)),
			},
			want: want{
				cr: cluster(withStatus(v1beta1.ClusterStatusDeleting),
					withConditions(runtimev1alpha1.Deleting())),
			},
		},
		"AlreadyDeleted": {
			args: args{
				eks: &fake.MockClient{
					MockDeleteClusterRequest: func(input *awseks.DeleteClusterInput) awseks.DeleteClusterRequest {
						return awseks.DeleteClusterRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: errors.New(awseks.ErrCodeResourceNotFoundException)},
						}
					},
				},
				cr: cluster(),
			},
			want: want{
				cr: cluster(withConditions(runtimev1alpha1.Deleting())),
			},
		},
		"Failed": {
			args: args{
				eks: &fake.MockClient{
					MockDeleteClusterRequest: func(input *awseks.DeleteClusterInput) awseks.DeleteClusterRequest {
						return awseks.DeleteClusterRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: errBoom},
						}
					},
				},
				cr: cluster(),
			},
			want: want{
				cr:  cluster(withConditions(runtimev1alpha1.Deleting())),
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
		cr  *v1beta1.Cluster
		err error
	}

	cases := map[string]struct {
		args
		want
	}{
		"Successful": {
			args: args{
				cr:   cluster(withTags(map[string]string{"foo": "bar"})),
				kube: &test.MockClient{MockUpdate: test.NewMockUpdateFn(nil)},
			},
			want: want{
				cr: cluster(withTags(resource.GetExternalTags(cluster()), (map[string]string{"foo": "bar"}))),
			},
		},
		"UpdateFailed": {
			args: args{
				cr:   cluster(),
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
