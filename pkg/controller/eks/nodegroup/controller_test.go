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

package nodegroup

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

	"github.com/crossplane-contrib/provider-aws/apis/eks/manualv1alpha1"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/eks"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/eks/fake"
	errorutils "github.com/crossplane-contrib/provider-aws/pkg/utils/errors"
)

var (
	version           = "1.16"
	desiredSize int32 = 3
	force             = false

	errBoom = errors.New("boom")
)

type args struct {
	eks  eks.Client
	kube client.Client
	cr   *manualv1alpha1.NodeGroup
}

type nodeGroupModifier func(*manualv1alpha1.NodeGroup)

func withConditions(c ...xpv1.Condition) nodeGroupModifier {
	return func(r *manualv1alpha1.NodeGroup) { r.Status.ConditionedStatus.Conditions = c }
}

func withTags(tagMaps ...map[string]string) nodeGroupModifier {
	tags := map[string]string{}
	for _, tagMap := range tagMaps {
		for k, v := range tagMap {
			tags[k] = v
		}
	}
	return func(r *manualv1alpha1.NodeGroup) { r.Spec.ForProvider.Tags = tags }
}

func withVersion(v *string) nodeGroupModifier {
	return func(r *manualv1alpha1.NodeGroup) { r.Spec.ForProvider.Version = v }
}

func withStatusVersion(v *string) nodeGroupModifier {
	return func(r *manualv1alpha1.NodeGroup) { r.Status.AtProvider.Version = *v }
}

func withStatus(s manualv1alpha1.NodeGroupStatusType) nodeGroupModifier {
	return func(r *manualv1alpha1.NodeGroup) { r.Status.AtProvider.Status = s }
}

func withScalingConfig(c *manualv1alpha1.NodeGroupScalingConfig) nodeGroupModifier {
	return func(r *manualv1alpha1.NodeGroup) { r.Spec.ForProvider.ScalingConfig = c }
}

func withUpdateConfig(c *manualv1alpha1.NodeGroupUpdateConfig) nodeGroupModifier {
	return func(r *manualv1alpha1.NodeGroup) { r.Spec.ForProvider.UpdateConfig = c }
}

func withDefaultUpdateConfig() nodeGroupModifier {
	return withUpdateConfig(&manualv1alpha1.NodeGroupUpdateConfig{Force: &force})
}

func nodeGroup(m ...nodeGroupModifier) *manualv1alpha1.NodeGroup {
	cr := &manualv1alpha1.NodeGroup{}
	for _, f := range m {
		f(cr)
	}
	return cr
}

var _ managed.ExternalClient = &external{}
var _ managed.ExternalConnecter = &connector{}

func TestObserve(t *testing.T) {
	type want struct {
		cr     *manualv1alpha1.NodeGroup
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
					MockDescribeNodegroup: func(tx context.Context, input *awseks.DescribeNodegroupInput, opts []func(*awseks.Options)) (*awseks.DescribeNodegroupOutput, error) {
						return &awseks.DescribeNodegroupOutput{
							Nodegroup: &awsekstypes.Nodegroup{
								Status: awsekstypes.NodegroupStatusActive,
							},
						}, nil
					},
				},
				cr: nodeGroup(withDefaultUpdateConfig()),
			},
			want: want{
				cr: nodeGroup(
					withConditions(xpv1.Available()),
					withStatus(manualv1alpha1.NodeGroupStatusActive),
					withDefaultUpdateConfig()),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
				},
			},
		},
		"DeletingState": {
			args: args{
				eks: &fake.MockClient{
					MockDescribeNodegroup: func(tx context.Context, input *awseks.DescribeNodegroupInput, opts []func(*awseks.Options)) (*awseks.DescribeNodegroupOutput, error) {
						return &awseks.DescribeNodegroupOutput{
							Nodegroup: &awsekstypes.Nodegroup{
								Status: awsekstypes.NodegroupStatusDeleting,
							},
						}, nil
					},
				},
				cr: nodeGroup(withDefaultUpdateConfig()),
			},
			want: want{
				cr: nodeGroup(
					withConditions(xpv1.Deleting()),
					withStatus(manualv1alpha1.NodeGroupStatusDeleting),
					withDefaultUpdateConfig()),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
				},
			},
		},
		"FailedState": {
			args: args{
				eks: &fake.MockClient{
					MockDescribeNodegroup: func(tx context.Context, input *awseks.DescribeNodegroupInput, opts []func(*awseks.Options)) (*awseks.DescribeNodegroupOutput, error) {
						return &awseks.DescribeNodegroupOutput{
							Nodegroup: &awsekstypes.Nodegroup{
								Status: awsekstypes.NodegroupStatusDegraded,
							},
						}, nil
					},
				},
				cr: nodeGroup(withDefaultUpdateConfig()),
			},
			want: want{
				cr: nodeGroup(
					withConditions(xpv1.Unavailable()),
					withStatus(manualv1alpha1.NodeGroupStatusDegraded),
					withDefaultUpdateConfig()),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
				},
			},
		},
		"FailedDescribeRequest": {
			args: args{
				eks: &fake.MockClient{
					MockDescribeNodegroup: func(tx context.Context, input *awseks.DescribeNodegroupInput, opts []func(*awseks.Options)) (*awseks.DescribeNodegroupOutput, error) {
						return nil, errBoom
					},
				},
				cr: nodeGroup(),
			},
			want: want{
				cr:  nodeGroup(),
				err: errorutils.Wrap(errBoom, errDescribeFailed),
			},
		},
		"NotFound": {
			args: args{
				eks: &fake.MockClient{
					MockDescribeNodegroup: func(tx context.Context, input *awseks.DescribeNodegroupInput, opts []func(*awseks.Options)) (*awseks.DescribeNodegroupOutput, error) {
						return nil, &awsekstypes.ResourceNotFoundException{}
					},
				},
				cr: nodeGroup(),
			},
			want: want{
				cr: nodeGroup(),
			},
		},
		"LateInitSuccess": {
			args: args{
				kube: &test.MockClient{
					MockUpdate: test.NewMockUpdateFn(nil),
				},
				eks: &fake.MockClient{
					MockDescribeNodegroup: func(tx context.Context, input *awseks.DescribeNodegroupInput, opts []func(*awseks.Options)) (*awseks.DescribeNodegroupOutput, error) {
						return &awseks.DescribeNodegroupOutput{
							Nodegroup: &awsekstypes.Nodegroup{
								Status:  awsekstypes.NodegroupStatusCreating,
								Version: &version,
							},
						}, nil
					},
				},
				cr: nodeGroup(),
			},
			want: want{
				cr: nodeGroup(
					withStatus(manualv1alpha1.NodeGroupStatusCreating),
					withConditions(xpv1.Creating()),
					withVersion(&version),
					withStatusVersion(&version),
					withDefaultUpdateConfig(),
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
					MockDescribeNodegroup: func(tx context.Context, input *awseks.DescribeNodegroupInput, opts []func(*awseks.Options)) (*awseks.DescribeNodegroupOutput, error) {
						return &awseks.DescribeNodegroupOutput{
							Nodegroup: &awsekstypes.Nodegroup{
								Status:  awsekstypes.NodegroupStatusCreating,
								Version: &version,
							},
						}, nil
					},
				},
				cr: nodeGroup(),
			},
			want: want{
				cr:  nodeGroup(withVersion(&version), withUpdateConfig(&manualv1alpha1.NodeGroupUpdateConfig{Force: &force})),
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
		cr     *manualv1alpha1.NodeGroup
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
					MockCreateNodegroup: func(tx context.Context, input *awseks.CreateNodegroupInput, opts []func(*awseks.Options)) (*awseks.CreateNodegroupOutput, error) {
						return &awseks.CreateNodegroupOutput{}, nil
					},
				},
				cr: nodeGroup(),
			},
			want: want{
				cr:     nodeGroup(withConditions(xpv1.Creating())),
				result: managed.ExternalCreation{},
			},
		},
		"SuccessfulNoNeedForCreate": {
			args: args{
				cr: nodeGroup(withStatus(manualv1alpha1.NodeGroupStatusCreating)),
			},
			want: want{
				cr: nodeGroup(
					withStatus(manualv1alpha1.NodeGroupStatusCreating),
					withConditions(xpv1.Creating())),
			},
		},
		"FailedRequest": {
			args: args{
				eks: &fake.MockClient{
					MockCreateNodegroup: func(tx context.Context, input *awseks.CreateNodegroupInput, opts []func(*awseks.Options)) (*awseks.CreateNodegroupOutput, error) {
						return nil, errBoom
					},
				},
				cr: nodeGroup(),
			},
			want: want{
				cr:  nodeGroup(withConditions(xpv1.Creating())),
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
		cr     *manualv1alpha1.NodeGroup
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
					MockDescribeNodegroup: func(tx context.Context, input *awseks.DescribeNodegroupInput, opts []func(*awseks.Options)) (*awseks.DescribeNodegroupOutput, error) {
						return &awseks.DescribeNodegroupOutput{
							Nodegroup: &awsekstypes.Nodegroup{},
						}, nil
					},
					MockUpdateNodegroupConfig: func(tx context.Context, input *awseks.UpdateNodegroupConfigInput, opts []func(*awseks.Options)) (*awseks.UpdateNodegroupConfigOutput, error) {
						return &awseks.UpdateNodegroupConfigOutput{}, nil
					},
					MockTagResource: func(tx context.Context, input *awseks.TagResourceInput, opts []func(*awseks.Options)) (*awseks.TagResourceOutput, error) {
						return &awseks.TagResourceOutput{}, nil
					},
				},
				cr: nodeGroup(
					withTags(map[string]string{"foo": "bar"})),
			},
			want: want{
				cr: nodeGroup(
					withTags(map[string]string{"foo": "bar"})),
			},
		},
		"SuccessfulRemoveTags": {
			args: args{
				eks: &fake.MockClient{
					MockDescribeNodegroup: func(tx context.Context, input *awseks.DescribeNodegroupInput, opts []func(*awseks.Options)) (*awseks.DescribeNodegroupOutput, error) {
						return &awseks.DescribeNodegroupOutput{
							Nodegroup: &awsekstypes.Nodegroup{},
						}, nil
					},
					MockUpdateNodegroupConfig: func(tx context.Context, input *awseks.UpdateNodegroupConfigInput, opts []func(*awseks.Options)) (*awseks.UpdateNodegroupConfigOutput, error) {
						return &awseks.UpdateNodegroupConfigOutput{}, nil
					},
					MockUntagResource: func(tx context.Context, input *awseks.UntagResourceInput, opts []func(*awseks.Options)) (*awseks.UntagResourceOutput, error) {
						return &awseks.UntagResourceOutput{}, nil
					},
				},
				cr: nodeGroup(),
			},
			want: want{
				cr: nodeGroup(),
			},
		},
		"SuccessfulUpdateVersion": {
			args: args{
				eks: &fake.MockClient{
					MockUpdateNodegroupVersion: func(tx context.Context, input *awseks.UpdateNodegroupVersionInput, opts []func(*awseks.Options)) (*awseks.UpdateNodegroupVersionOutput, error) {
						return &awseks.UpdateNodegroupVersionOutput{}, nil
					},
					MockDescribeNodegroup: func(tx context.Context, input *awseks.DescribeNodegroupInput, opts []func(*awseks.Options)) (*awseks.DescribeNodegroupOutput, error) {
						return &awseks.DescribeNodegroupOutput{
							Nodegroup: &awsekstypes.Nodegroup{},
						}, nil
					},
				},
				cr: nodeGroup(withVersion(&version)),
			},
			want: want{
				cr: nodeGroup(withVersion(&version)),
			},
		},
		"SuccessfulUpdateNodeGroup": {
			args: args{
				eks: &fake.MockClient{
					MockUpdateNodegroupConfig: func(tx context.Context, input *awseks.UpdateNodegroupConfigInput, opts []func(*awseks.Options)) (*awseks.UpdateNodegroupConfigOutput, error) {
						return &awseks.UpdateNodegroupConfigOutput{}, nil
					},
					MockDescribeNodegroup: func(tx context.Context, input *awseks.DescribeNodegroupInput, opts []func(*awseks.Options)) (*awseks.DescribeNodegroupOutput, error) {
						return &awseks.DescribeNodegroupOutput{
							Nodegroup: &awsekstypes.Nodegroup{},
						}, nil
					},
				},
				cr: nodeGroup(withScalingConfig(&manualv1alpha1.NodeGroupScalingConfig{DesiredSize: &desiredSize})),
			},
			want: want{
				cr: nodeGroup(withScalingConfig(&manualv1alpha1.NodeGroupScalingConfig{DesiredSize: &desiredSize})),
			},
		},
		"AlreadyModifying": {
			args: args{
				cr: nodeGroup(withStatus(manualv1alpha1.NodeGroupStatusUpdating)),
			},
			want: want{
				cr: nodeGroup(withStatus(manualv1alpha1.NodeGroupStatusUpdating)),
			},
		},
		"FailedDescribe": {
			args: args{
				eks: &fake.MockClient{
					MockDescribeNodegroup: func(tx context.Context, input *awseks.DescribeNodegroupInput, opts []func(*awseks.Options)) (*awseks.DescribeNodegroupOutput, error) {
						return nil, errBoom
					},
				},
				cr: nodeGroup(),
			},
			want: want{
				cr:  nodeGroup(),
				err: errorutils.Wrap(errBoom, errDescribeFailed),
			},
		},
		"FailedUpdateConfig": {
			args: args{
				eks: &fake.MockClient{
					MockUpdateNodegroupConfig: func(tx context.Context, input *awseks.UpdateNodegroupConfigInput, opts []func(*awseks.Options)) (*awseks.UpdateNodegroupConfigOutput, error) {
						return nil, errBoom
					},
					MockDescribeNodegroup: func(tx context.Context, input *awseks.DescribeNodegroupInput, opts []func(*awseks.Options)) (*awseks.DescribeNodegroupOutput, error) {
						return &awseks.DescribeNodegroupOutput{
							Nodegroup: &awsekstypes.Nodegroup{},
						}, nil
					},
				},
				cr: nodeGroup(),
			},
			want: want{
				cr:  nodeGroup(),
				err: errorutils.Wrap(errBoom, errUpdateConfigFailed),
			},
		},
		"FailedUpdateVersion": {
			args: args{
				eks: &fake.MockClient{
					MockUpdateNodegroupVersion: func(tx context.Context, input *awseks.UpdateNodegroupVersionInput, opts []func(*awseks.Options)) (*awseks.UpdateNodegroupVersionOutput, error) {
						return nil, errBoom
					},
					MockDescribeNodegroup: func(tx context.Context, input *awseks.DescribeNodegroupInput, opts []func(*awseks.Options)) (*awseks.DescribeNodegroupOutput, error) {
						return &awseks.DescribeNodegroupOutput{
							Nodegroup: &awsekstypes.Nodegroup{},
						}, nil
					},
				},
				cr: nodeGroup(withVersion(&version)),
			},
			want: want{
				cr:  nodeGroup(withVersion(&version)),
				err: errorutils.Wrap(errBoom, errUpdateVersionFailed),
			},
		},
		"FailedRemoveTags": {
			args: args{
				eks: &fake.MockClient{
					MockDescribeNodegroup: func(tx context.Context, input *awseks.DescribeNodegroupInput, opts []func(*awseks.Options)) (*awseks.DescribeNodegroupOutput, error) {
						return &awseks.DescribeNodegroupOutput{
							Nodegroup: &awsekstypes.Nodegroup{
								Tags: map[string]string{"foo": "bar"},
							},
						}, nil
					},
					MockUntagResource: func(tx context.Context, input *awseks.UntagResourceInput, opts []func(*awseks.Options)) (*awseks.UntagResourceOutput, error) {
						return nil, errBoom
					},
				},
				cr: nodeGroup(),
			},
			want: want{
				cr:  nodeGroup(),
				err: errorutils.Wrap(errBoom, errAddTagsFailed),
			},
		},
		"FailedAddTags": {
			args: args{
				eks: &fake.MockClient{
					MockDescribeNodegroup: func(tx context.Context, input *awseks.DescribeNodegroupInput, opts []func(*awseks.Options)) (*awseks.DescribeNodegroupOutput, error) {
						return &awseks.DescribeNodegroupOutput{
							Nodegroup: &awsekstypes.Nodegroup{},
						}, nil
					},
					MockTagResource: func(tx context.Context, input *awseks.TagResourceInput, opts []func(*awseks.Options)) (*awseks.TagResourceOutput, error) {
						return nil, errBoom
					},
				},
				cr: nodeGroup(withTags(map[string]string{"foo": "bar"})),
			},
			want: want{
				cr:  nodeGroup(withTags(map[string]string{"foo": "bar"})),
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
		cr  *manualv1alpha1.NodeGroup
		err error
	}

	cases := map[string]struct {
		args
		want
	}{
		"Successful": {
			args: args{
				eks: &fake.MockClient{
					MockDeleteNodegroup: func(tx context.Context, input *awseks.DeleteNodegroupInput, opts []func(*awseks.Options)) (*awseks.DeleteNodegroupOutput, error) {
						return &awseks.DeleteNodegroupOutput{}, nil
					},
				},
				cr: nodeGroup(),
			},
			want: want{
				cr: nodeGroup(withConditions(xpv1.Deleting())),
			},
		},
		"AlreadyDeleting": {
			args: args{
				cr: nodeGroup(withStatus(manualv1alpha1.NodeGroupStatusDeleting)),
			},
			want: want{
				cr: nodeGroup(withStatus(manualv1alpha1.NodeGroupStatusDeleting),
					withConditions(xpv1.Deleting())),
			},
		},
		"AlreadyDeleted": {
			args: args{
				eks: &fake.MockClient{
					MockDeleteNodegroup: func(tx context.Context, input *awseks.DeleteNodegroupInput, opts []func(*awseks.Options)) (*awseks.DeleteNodegroupOutput, error) {
						return nil, &awsekstypes.ResourceNotFoundException{}
					},
				},
				cr: nodeGroup(),
			},
			want: want{
				cr: nodeGroup(withConditions(xpv1.Deleting())),
			},
		},
		"Failed": {
			args: args{
				eks: &fake.MockClient{
					MockDeleteNodegroup: func(tx context.Context, input *awseks.DeleteNodegroupInput, opts []func(*awseks.Options)) (*awseks.DeleteNodegroupOutput, error) {
						return nil, errBoom
					},
				},
				cr: nodeGroup(),
			},
			want: want{
				cr:  nodeGroup(withConditions(xpv1.Deleting())),
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
