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
	"testing"

	awseks "github.com/aws/aws-sdk-go-v2/service/eks"
	awsekstypes "github.com/aws/aws-sdk-go-v2/service/eks/types"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane-contrib/provider-aws/apis/eks/v1beta1"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/eks"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/eks/fake"
	errorutils "github.com/crossplane-contrib/provider-aws/pkg/utils/errors"
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

func withConditions(c ...xpv1.Condition) clusterModifier {
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

func withStatusVersion(v *string) clusterModifier {
	return func(r *v1beta1.Cluster) { r.Status.AtProvider.Version = *v }
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
					MockDescribeCluster: func(ctx context.Context, input *awseks.DescribeClusterInput, opts []func(*awseks.Options)) (*awseks.DescribeClusterOutput, error) {
						return &awseks.DescribeClusterOutput{
							Cluster: &awsekstypes.Cluster{
								Status: awsekstypes.ClusterStatusActive,
							},
						}, nil
					},
				},
				cr: cluster(),
			},
			want: want{
				cr: cluster(
					withConditions(xpv1.Available()),
					withStatus(v1beta1.ClusterStatusActive)),
				result: managed.ExternalObservation{
					ResourceExists:    true,
					ResourceUpToDate:  true,
					ConnectionDetails: eks.GetConnectionDetails(context.TODO(), &awsekstypes.Cluster{}, &fake.MockSTSClient{}),
				},
			},
		},
		"DeletingState": {
			args: args{
				eks: &fake.MockClient{
					MockDescribeCluster: func(ctx context.Context, input *awseks.DescribeClusterInput, opts []func(*awseks.Options)) (*awseks.DescribeClusterOutput, error) {
						return &awseks.DescribeClusterOutput{
							Cluster: &awsekstypes.Cluster{
								Status: awsekstypes.ClusterStatusDeleting,
							},
						}, nil
					},
				},
				cr: cluster(),
			},
			want: want{
				cr: cluster(
					withConditions(xpv1.Deleting()),
					withStatus(v1beta1.ClusterStatusDeleting)),
				result: managed.ExternalObservation{
					ResourceExists:    true,
					ResourceUpToDate:  true,
					ConnectionDetails: eks.GetConnectionDetails(context.TODO(), &awsekstypes.Cluster{}, &fake.MockSTSClient{}),
				},
			},
		},
		"FailedState": {
			args: args{
				eks: &fake.MockClient{
					MockDescribeCluster: func(ctx context.Context, input *awseks.DescribeClusterInput, opts []func(*awseks.Options)) (*awseks.DescribeClusterOutput, error) {
						return &awseks.DescribeClusterOutput{
							Cluster: &awsekstypes.Cluster{
								Status: awsekstypes.ClusterStatusFailed,
							},
						}, nil
					},
				},
				cr: cluster(),
			},
			want: want{
				cr: cluster(
					withConditions(xpv1.Unavailable()),
					withStatus(v1beta1.ClusterStatusFailed)),
				result: managed.ExternalObservation{
					ResourceExists:    true,
					ResourceUpToDate:  true,
					ConnectionDetails: eks.GetConnectionDetails(context.TODO(), &awsekstypes.Cluster{}, &fake.MockSTSClient{}),
				},
			},
		},
		"FailedDescribeRequest": {
			args: args{
				eks: &fake.MockClient{
					MockDescribeCluster: func(ctx context.Context, input *awseks.DescribeClusterInput, opts []func(*awseks.Options)) (*awseks.DescribeClusterOutput, error) {
						return nil, errBoom
					},
				},
				cr: cluster(),
			},
			want: want{
				cr:  cluster(),
				err: errorutils.Wrap(errBoom, errDescribeFailed),
			},
		},
		"NotFound": {
			args: args{
				eks: &fake.MockClient{
					MockDescribeCluster: func(ctx context.Context, input *awseks.DescribeClusterInput, opts []func(*awseks.Options)) (*awseks.DescribeClusterOutput, error) {
						return nil, &awsekstypes.ResourceNotFoundException{}
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
					MockDescribeCluster: func(ctx context.Context, input *awseks.DescribeClusterInput, opts []func(*awseks.Options)) (*awseks.DescribeClusterOutput, error) {
						return &awseks.DescribeClusterOutput{
							Cluster: &awsekstypes.Cluster{
								Status:  awsekstypes.ClusterStatusCreating,
								Version: &version,
							},
						}, nil
					},
				},
				cr: cluster(),
			},
			want: want{
				cr: cluster(
					withStatus(v1beta1.ClusterStatusCreating),
					withConditions(xpv1.Creating()),
					withVersion(&version),
					withStatusVersion(&version),
				),
				result: managed.ExternalObservation{
					ResourceExists:    true,
					ResourceUpToDate:  true,
					ConnectionDetails: eks.GetConnectionDetails(context.TODO(), &awsekstypes.Cluster{}, &fake.MockSTSClient{}),
				},
			},
		},
		"LateInitFailedKubeUpdate": {
			args: args{
				kube: &test.MockClient{
					MockUpdate: test.NewMockUpdateFn(errBoom),
				},
				eks: &fake.MockClient{
					MockDescribeCluster: func(ctx context.Context, input *awseks.DescribeClusterInput, opts []func(*awseks.Options)) (*awseks.DescribeClusterOutput, error) {
						return &awseks.DescribeClusterOutput{
							Cluster: &awsekstypes.Cluster{
								Status:  awsekstypes.ClusterStatusCreating,
								Version: &version,
							},
						}, nil
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
					MockCreateCluster: func(ctx context.Context, input *awseks.CreateClusterInput, opts []func(*awseks.Options)) (*awseks.CreateClusterOutput, error) {
						return &awseks.CreateClusterOutput{}, nil
					},
				},
				cr: cluster(),
			},
			want: want{
				cr:     cluster(withConditions(xpv1.Creating())),
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
					withConditions(xpv1.Creating())),
			},
		},
		"FailedRequest": {
			args: args{
				eks: &fake.MockClient{
					MockCreateCluster: func(ctx context.Context, input *awseks.CreateClusterInput, opts []func(*awseks.Options)) (*awseks.CreateClusterOutput, error) {
						return nil, errBoom
					},
				},
				cr: cluster(),
			},
			want: want{
				cr:  cluster(withConditions(xpv1.Creating())),
				err: errorutils.Wrap(errBoom, errCreateFailed),
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
					MockDescribeCluster: func(ctx context.Context, input *awseks.DescribeClusterInput, opts []func(*awseks.Options)) (*awseks.DescribeClusterOutput, error) {
						return &awseks.DescribeClusterOutput{
							Cluster: &awsekstypes.Cluster{},
						}, nil
					},
					MockUpdateClusterConfig: func(ctx context.Context, input *awseks.UpdateClusterConfigInput, opts []func(*awseks.Options)) (*awseks.UpdateClusterConfigOutput, error) {
						return &awseks.UpdateClusterConfigOutput{}, nil
					},
					MockTagResource: func(ctx context.Context, input *awseks.TagResourceInput, opts []func(*awseks.Options)) (*awseks.TagResourceOutput, error) {
						return &awseks.TagResourceOutput{}, nil
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
					MockDescribeCluster: func(ctx context.Context, input *awseks.DescribeClusterInput, opts []func(*awseks.Options)) (*awseks.DescribeClusterOutput, error) {
						return &awseks.DescribeClusterOutput{
							Cluster: &awsekstypes.Cluster{
								Tags: map[string]string{"foo": "bar"},
							},
						}, nil
					},
					MockUpdateClusterConfig: func(ctx context.Context, input *awseks.UpdateClusterConfigInput, opts []func(*awseks.Options)) (*awseks.UpdateClusterConfigOutput, error) {
						return &awseks.UpdateClusterConfigOutput{}, nil
					},
					MockUntagResource: func(ctx context.Context, input *awseks.UntagResourceInput, opts []func(*awseks.Options)) (*awseks.UntagResourceOutput, error) {
						return &awseks.UntagResourceOutput{}, nil
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
					MockUpdateClusterVersion: func(ctx context.Context, input *awseks.UpdateClusterVersionInput, opts []func(*awseks.Options)) (*awseks.UpdateClusterVersionOutput, error) {
						return &awseks.UpdateClusterVersionOutput{}, nil
					},
					MockDescribeCluster: func(ctx context.Context, input *awseks.DescribeClusterInput, opts []func(*awseks.Options)) (*awseks.DescribeClusterOutput, error) {
						return &awseks.DescribeClusterOutput{
							Cluster: &awsekstypes.Cluster{},
						}, nil
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
					MockUpdateClusterConfig: func(ctx context.Context, input *awseks.UpdateClusterConfigInput, opts []func(*awseks.Options)) (*awseks.UpdateClusterConfigOutput, error) {
						return &awseks.UpdateClusterConfigOutput{}, nil
					},
					MockDescribeCluster: func(ctx context.Context, input *awseks.DescribeClusterInput, opts []func(*awseks.Options)) (*awseks.DescribeClusterOutput, error) {
						return &awseks.DescribeClusterOutput{
							Cluster: &awsekstypes.Cluster{},
						}, nil
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
					MockDescribeCluster: func(ctx context.Context, input *awseks.DescribeClusterInput, opts []func(*awseks.Options)) (*awseks.DescribeClusterOutput, error) {
						return nil, errBoom
					},
				},
				cr: cluster(),
			},
			want: want{
				cr:  cluster(),
				err: errorutils.Wrap(errBoom, errDescribeFailed),
			},
		},
		"FailedUpdateConfig": {
			args: args{
				eks: &fake.MockClient{
					MockUpdateClusterConfig: func(ctx context.Context, input *awseks.UpdateClusterConfigInput, opts []func(*awseks.Options)) (*awseks.UpdateClusterConfigOutput, error) {
						return nil, errBoom
					},
					MockDescribeCluster: func(ctx context.Context, input *awseks.DescribeClusterInput, opts []func(*awseks.Options)) (*awseks.DescribeClusterOutput, error) {
						return &awseks.DescribeClusterOutput{
							Cluster: &awsekstypes.Cluster{},
						}, nil
					},
				},
				cr: cluster(),
			},
			want: want{
				cr:  cluster(),
				err: errorutils.Wrap(errBoom, errUpdateConfigFailed),
			},
		},
		"FailedUpdateVersion": {
			args: args{
				eks: &fake.MockClient{
					MockUpdateClusterVersion: func(ctx context.Context, input *awseks.UpdateClusterVersionInput, opts []func(*awseks.Options)) (*awseks.UpdateClusterVersionOutput, error) {
						return nil, errBoom
					},
					MockDescribeCluster: func(ctx context.Context, input *awseks.DescribeClusterInput, opts []func(*awseks.Options)) (*awseks.DescribeClusterOutput, error) {
						return &awseks.DescribeClusterOutput{
							Cluster: &awsekstypes.Cluster{},
						}, nil
					},
				},
				cr: cluster(withVersion(&version)),
			},
			want: want{
				cr:  cluster(withVersion(&version)),
				err: errorutils.Wrap(errBoom, errUpdateVersionFailed),
			},
		},
		"FailedRemoveTags": {
			args: args{
				eks: &fake.MockClient{
					MockDescribeCluster: func(ctx context.Context, input *awseks.DescribeClusterInput, opts []func(*awseks.Options)) (*awseks.DescribeClusterOutput, error) {
						return &awseks.DescribeClusterOutput{
							Cluster: &awsekstypes.Cluster{
								Tags: map[string]string{"foo": "bar"},
							},
						}, nil
					},
					MockUntagResource: func(ctx context.Context, input *awseks.UntagResourceInput, opts []func(*awseks.Options)) (*awseks.UntagResourceOutput, error) {
						return nil, errBoom
					},
				},
				cr: cluster(),
			},
			want: want{
				cr:  cluster(),
				err: errorutils.Wrap(errBoom, errAddTagsFailed),
			},
		},
		"FailedAddTags": {
			args: args{
				eks: &fake.MockClient{
					MockDescribeCluster: func(ctx context.Context, input *awseks.DescribeClusterInput, opts []func(*awseks.Options)) (*awseks.DescribeClusterOutput, error) {
						return &awseks.DescribeClusterOutput{
							Cluster: &awsekstypes.Cluster{},
						}, nil
					},
					MockTagResource: func(ctx context.Context, input *awseks.TagResourceInput, opts []func(*awseks.Options)) (*awseks.TagResourceOutput, error) {
						return nil, errBoom
					},
				},
				cr: cluster(withTags(map[string]string{"foo": "bar"})),
			},
			want: want{
				cr:  cluster(withTags(map[string]string{"foo": "bar"})),
				err: errorutils.Wrap(errBoom, errAddTagsFailed),
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
					MockDeleteCluster: func(ctx context.Context, input *awseks.DeleteClusterInput, opts []func(*awseks.Options)) (*awseks.DeleteClusterOutput, error) {
						return &awseks.DeleteClusterOutput{}, nil
					},
				},
				cr: cluster(),
			},
			want: want{
				cr: cluster(withConditions(xpv1.Deleting())),
			},
		},
		"AlreadyDeleting": {
			args: args{
				cr: cluster(withStatus(v1beta1.ClusterStatusDeleting)),
			},
			want: want{
				cr: cluster(withStatus(v1beta1.ClusterStatusDeleting),
					withConditions(xpv1.Deleting())),
			},
		},
		"AlreadyDeleted": {
			args: args{
				eks: &fake.MockClient{
					MockDeleteCluster: func(ctx context.Context, input *awseks.DeleteClusterInput, opts []func(*awseks.Options)) (*awseks.DeleteClusterOutput, error) {
						return nil, &awsekstypes.ResourceNotFoundException{}
					},
				},
				cr: cluster(),
			},
			want: want{
				cr: cluster(withConditions(xpv1.Deleting())),
			},
		},
		"Failed": {
			args: args{
				eks: &fake.MockClient{
					MockDeleteCluster: func(ctx context.Context, input *awseks.DeleteClusterInput, opts []func(*awseks.Options)) (*awseks.DeleteClusterOutput, error) {
						return nil, errBoom
					},
				},
				cr: cluster(),
			},
			want: want{
				cr:  cluster(withConditions(xpv1.Deleting())),
				err: errorutils.Wrap(errBoom, errDeleteFailed),
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
