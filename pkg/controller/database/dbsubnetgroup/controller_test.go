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

package dbsubnetgroup

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsrds "github.com/aws/aws-sdk-go-v2/service/rds"
	awsrdstypes "github.com/aws/aws-sdk-go-v2/service/rds/types"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	v1beta1 "github.com/crossplane-contrib/provider-aws/apis/database/v1beta1"
	dbsg "github.com/crossplane-contrib/provider-aws/pkg/clients/dbsubnetgroup"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/dbsubnetgroup/fake"
	errorutils "github.com/crossplane-contrib/provider-aws/pkg/utils/errors"
)

var (
	dbSubnetGroupDescription = "arbitrary description"
	errBoom                  = errors.New("boom")
)

type args struct {
	client dbsg.Client
	kube   client.Client
	cr     *v1beta1.DBSubnetGroup
}

type dbSubnetGroupModifier func(*v1beta1.DBSubnetGroup)

func withConditions(c ...xpv1.Condition) dbSubnetGroupModifier {
	return func(sg *v1beta1.DBSubnetGroup) { sg.Status.ConditionedStatus.Conditions = c }
}

func withDBSubnetGroupStatus(s string) dbSubnetGroupModifier {
	return func(sg *v1beta1.DBSubnetGroup) { sg.Status.AtProvider.State = s }
}

func withDBSubnetGroupDescription(s string) dbSubnetGroupModifier {
	return func(sg *v1beta1.DBSubnetGroup) { sg.Spec.ForProvider.Description = s }
}

func withDBSubnetGroupTags() dbSubnetGroupModifier {
	return func(sg *v1beta1.DBSubnetGroup) {
		sg.Spec.ForProvider.Tags = []v1beta1.Tag{{Key: "arbitrary key", Value: "arbitrary value"}}
	}
}

func mockListTagsForResource(ctx context.Context, input *awsrds.ListTagsForResourceInput, opts []func(*awsrds.Options)) (*awsrds.ListTagsForResourceOutput, error) {
	return &awsrds.ListTagsForResourceOutput{TagList: []awsrdstypes.Tag{}}, nil
}

func dbSubnetGroup(m ...dbSubnetGroupModifier) *v1beta1.DBSubnetGroup {
	cr := &v1beta1.DBSubnetGroup{}
	for _, f := range m {
		f(cr)
	}
	return cr
}

var _ managed.ExternalClient = &external{}
var _ managed.ExternalConnecter = &connector{}

func TestObserve(t *testing.T) {
	type want struct {
		cr     *v1beta1.DBSubnetGroup
		result managed.ExternalObservation
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"SuccessfulAvailable": {
			args: args{
				client: &fake.MockDBSubnetGroupClient{
					MockDescribeDBSubnetGroups: func(ctx context.Context, input *awsrds.DescribeDBSubnetGroupsInput, opts []func(*awsrds.Options)) (*awsrds.DescribeDBSubnetGroupsOutput, error) {
						return &awsrds.DescribeDBSubnetGroupsOutput{
							DBSubnetGroups: []awsrdstypes.DBSubnetGroup{
								{
									SubnetGroupStatus: aws.String(string(v1beta1.DBSubnetGroupStateAvailable)),
								},
							},
						}, nil
					},
					MockListTagsForResource: mockListTagsForResource,
				},
				cr: dbSubnetGroup(),
			},
			want: want{
				cr: dbSubnetGroup(
					withConditions(xpv1.Available()),
					withDBSubnetGroupStatus(v1beta1.DBSubnetGroupStateAvailable),
				),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
				},
			},
		},
		"DeletingState": {
			args: args{
				client: &fake.MockDBSubnetGroupClient{
					MockDescribeDBSubnetGroups: func(ctx context.Context, input *awsrds.DescribeDBSubnetGroupsInput, opts []func(*awsrds.Options)) (*awsrds.DescribeDBSubnetGroupsOutput, error) {
						return &awsrds.DescribeDBSubnetGroupsOutput{
							DBSubnetGroups: []awsrdstypes.DBSubnetGroup{{}},
						}, nil
					},
					MockListTagsForResource: mockListTagsForResource,
				},
				cr: dbSubnetGroup(),
			},
			want: want{
				cr: dbSubnetGroup(
					withConditions(xpv1.Unavailable())),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
				},
			},
		},
		"FailedState": {
			args: args{
				client: &fake.MockDBSubnetGroupClient{
					MockDescribeDBSubnetGroups: func(ctx context.Context, input *awsrds.DescribeDBSubnetGroupsInput, opts []func(*awsrds.Options)) (*awsrds.DescribeDBSubnetGroupsOutput, error) {
						return &awsrds.DescribeDBSubnetGroupsOutput{
							DBSubnetGroups: []awsrdstypes.DBSubnetGroup{{}},
						}, nil
					},
					MockListTagsForResource: mockListTagsForResource,
				},
				cr: dbSubnetGroup(),
			},
			want: want{
				cr: dbSubnetGroup(
					withConditions(xpv1.Unavailable())),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
				},
			},
		},
		"FailedDescribeRequest": {
			args: args{
				client: &fake.MockDBSubnetGroupClient{
					MockDescribeDBSubnetGroups: func(ctx context.Context, input *awsrds.DescribeDBSubnetGroupsInput, opts []func(*awsrds.Options)) (*awsrds.DescribeDBSubnetGroupsOutput, error) {
						return nil, errBoom
					},
					MockListTagsForResource: mockListTagsForResource,
				},
				cr: dbSubnetGroup(),
			},
			want: want{
				cr:  dbSubnetGroup(),
				err: errorutils.Wrap(errBoom, errDescribe),
			},
		},
		"NotFound": {
			args: args{
				client: &fake.MockDBSubnetGroupClient{
					MockDescribeDBSubnetGroups: func(ctx context.Context, input *awsrds.DescribeDBSubnetGroupsInput, opts []func(*awsrds.Options)) (*awsrds.DescribeDBSubnetGroupsOutput, error) {
						return nil, &awsrdstypes.DBSubnetGroupNotFoundFault{}
					},
					MockListTagsForResource: mockListTagsForResource,
				},
				cr: dbSubnetGroup(),
			},
			want: want{
				cr: dbSubnetGroup(),
			},
		},
		"LateInitSuccess": {
			args: args{
				kube: &test.MockClient{
					MockUpdate: test.NewMockUpdateFn(nil),
				},
				client: &fake.MockDBSubnetGroupClient{
					MockDescribeDBSubnetGroups: func(ctx context.Context, input *awsrds.DescribeDBSubnetGroupsInput, opts []func(*awsrds.Options)) (*awsrds.DescribeDBSubnetGroupsOutput, error) {
						return &awsrds.DescribeDBSubnetGroupsOutput{
							DBSubnetGroups: []awsrdstypes.DBSubnetGroup{
								{
									DBSubnetGroupDescription: aws.String(dbSubnetGroupDescription),
								},
							},
						}, nil
					},
					MockListTagsForResource: mockListTagsForResource,
				},
				cr: dbSubnetGroup(),
			},
			want: want{
				cr: dbSubnetGroup(
					withDBSubnetGroupDescription(dbSubnetGroupDescription),
					withConditions(xpv1.Unavailable())),
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
				client: &fake.MockDBSubnetGroupClient{
					MockDescribeDBSubnetGroups: func(ctx context.Context, input *awsrds.DescribeDBSubnetGroupsInput, opts []func(*awsrds.Options)) (*awsrds.DescribeDBSubnetGroupsOutput, error) {
						return &awsrds.DescribeDBSubnetGroupsOutput{
							DBSubnetGroups: []awsrdstypes.DBSubnetGroup{
								{
									DBSubnetGroupDescription: aws.String(dbSubnetGroupDescription),
								},
							},
						}, nil
					},
					MockListTagsForResource: mockListTagsForResource,
				},
				cr: dbSubnetGroup(),
			},
			want: want{
				cr: dbSubnetGroup(
					withDBSubnetGroupDescription(dbSubnetGroupDescription),
				),
				err: errorutils.Wrap(errBoom, errLateInit),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.client}
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
		cr     *v1beta1.DBSubnetGroup
		result managed.ExternalCreation
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"Successful": {
			args: args{
				client: &fake.MockDBSubnetGroupClient{
					MockCreateDBSubnetGroup: func(ctx context.Context, input *awsrds.CreateDBSubnetGroupInput, opts []func(*awsrds.Options)) (*awsrds.CreateDBSubnetGroupOutput, error) {
						return &awsrds.CreateDBSubnetGroupOutput{}, nil
					},
				},
				cr: dbSubnetGroup(withDBSubnetGroupDescription(dbSubnetGroupDescription)),
			},
			want: want{
				cr: dbSubnetGroup(
					withDBSubnetGroupDescription(dbSubnetGroupDescription),
					withConditions(xpv1.Creating())),
			},
		},
		"FailedRequest": {
			args: args{
				client: &fake.MockDBSubnetGroupClient{
					MockCreateDBSubnetGroup: func(ctx context.Context, input *awsrds.CreateDBSubnetGroupInput, opts []func(*awsrds.Options)) (*awsrds.CreateDBSubnetGroupOutput, error) {
						return nil, errBoom
					},
				},
				cr: dbSubnetGroup(),
			},
			want: want{
				cr:  dbSubnetGroup(withConditions(xpv1.Creating())),
				err: errorutils.Wrap(errBoom, errCreate),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.client}
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
		cr     *v1beta1.DBSubnetGroup
		result managed.ExternalUpdate
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"Successful": {
			args: args{
				client: &fake.MockDBSubnetGroupClient{
					MockModifyDBSubnetGroup: func(ctx context.Context, input *awsrds.ModifyDBSubnetGroupInput, opts []func(*awsrds.Options)) (*awsrds.ModifyDBSubnetGroupOutput, error) {
						return &awsrds.ModifyDBSubnetGroupOutput{}, nil
					},
					MockDescribeDBSubnetGroups: func(ctx context.Context, input *awsrds.DescribeDBSubnetGroupsInput, opts []func(*awsrds.Options)) (*awsrds.DescribeDBSubnetGroupsOutput, error) {
						return &awsrds.DescribeDBSubnetGroupsOutput{
							DBSubnetGroups: []awsrdstypes.DBSubnetGroup{{}},
						}, nil
					},
				},
				cr: dbSubnetGroup(),
			},
			want: want{
				cr: dbSubnetGroup(),
			},
		},
		"FailedModify": {
			args: args{
				client: &fake.MockDBSubnetGroupClient{
					MockModifyDBSubnetGroup: func(ctx context.Context, input *awsrds.ModifyDBSubnetGroupInput, opts []func(*awsrds.Options)) (*awsrds.ModifyDBSubnetGroupOutput, error) {
						return nil, errBoom
					},
					MockDescribeDBSubnetGroups: func(ctx context.Context, input *awsrds.DescribeDBSubnetGroupsInput, opts []func(*awsrds.Options)) (*awsrds.DescribeDBSubnetGroupsOutput, error) {
						return &awsrds.DescribeDBSubnetGroupsOutput{
							DBSubnetGroups: []awsrdstypes.DBSubnetGroup{{}},
						}, nil
					},
				},
				cr: dbSubnetGroup(),
			},
			want: want{
				cr:  dbSubnetGroup(),
				err: errorutils.Wrap(errBoom, errUpdate),
			},
		},
		"SuccessfulWithTags": {
			args: args{
				client: &fake.MockDBSubnetGroupClient{
					MockModifyDBSubnetGroup: func(ctx context.Context, input *awsrds.ModifyDBSubnetGroupInput, opts []func(*awsrds.Options)) (*awsrds.ModifyDBSubnetGroupOutput, error) {
						return &awsrds.ModifyDBSubnetGroupOutput{}, nil
					},
					MockDescribeDBSubnetGroups: func(ctx context.Context, input *awsrds.DescribeDBSubnetGroupsInput, opts []func(*awsrds.Options)) (*awsrds.DescribeDBSubnetGroupsOutput, error) {
						return &awsrds.DescribeDBSubnetGroupsOutput{
							DBSubnetGroups: []awsrdstypes.DBSubnetGroup{{}},
						}, nil
					},
					MockAddTagsToResource: func(ctx context.Context, input *awsrds.AddTagsToResourceInput, opts []func(*awsrds.Options)) (*awsrds.AddTagsToResourceOutput, error) {
						return &awsrds.AddTagsToResourceOutput{}, nil
					},
				},
				cr: dbSubnetGroup(withDBSubnetGroupTags()),
			},
			want: want{
				cr: dbSubnetGroup(withDBSubnetGroupTags()),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.client}
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
		cr  *v1beta1.DBSubnetGroup
		err error
	}

	cases := map[string]struct {
		args
		want
	}{
		"Successful": {
			args: args{
				client: &fake.MockDBSubnetGroupClient{
					MockDeleteDBSubnetGroup: func(ctx context.Context, input *awsrds.DeleteDBSubnetGroupInput, opts []func(*awsrds.Options)) (*awsrds.DeleteDBSubnetGroupOutput, error) {
						return &awsrds.DeleteDBSubnetGroupOutput{}, nil
					},
					MockModifyDBSubnetGroup: func(ctx context.Context, input *awsrds.ModifyDBSubnetGroupInput, opts []func(*awsrds.Options)) (*awsrds.ModifyDBSubnetGroupOutput, error) {
						return &awsrds.ModifyDBSubnetGroupOutput{}, nil
					},
					MockDescribeDBSubnetGroups: func(ctx context.Context, input *awsrds.DescribeDBSubnetGroupsInput, opts []func(*awsrds.Options)) (*awsrds.DescribeDBSubnetGroupsOutput, error) {
						return &awsrds.DescribeDBSubnetGroupsOutput{
							DBSubnetGroups: []awsrdstypes.DBSubnetGroup{{}},
						}, nil
					},
				},
				cr: dbSubnetGroup(),
			},
			want: want{
				cr: dbSubnetGroup(withConditions(xpv1.Deleting())),
			},
		},
		"AlreadyDeleted": {
			args: args{
				client: &fake.MockDBSubnetGroupClient{
					MockDeleteDBSubnetGroup: func(ctx context.Context, input *awsrds.DeleteDBSubnetGroupInput, opts []func(*awsrds.Options)) (*awsrds.DeleteDBSubnetGroupOutput, error) {
						return nil, &awsrdstypes.DBSubnetGroupNotFoundFault{}
					},
				},
				cr: dbSubnetGroup(),
			},
			want: want{
				cr: dbSubnetGroup(withConditions(xpv1.Deleting())),
			},
		},
		"Failed": {
			args: args{
				client: &fake.MockDBSubnetGroupClient{
					MockDeleteDBSubnetGroup: func(ctx context.Context, input *awsrds.DeleteDBSubnetGroupInput, opts []func(*awsrds.Options)) (*awsrds.DeleteDBSubnetGroupOutput, error) {
						return nil, errBoom
					},
					MockModifyDBSubnetGroup: func(ctx context.Context, input *awsrds.ModifyDBSubnetGroupInput, opts []func(*awsrds.Options)) (*awsrds.ModifyDBSubnetGroupOutput, error) {
						return &awsrds.ModifyDBSubnetGroupOutput{}, nil
					},
					MockDescribeDBSubnetGroups: func(ctx context.Context, input *awsrds.DescribeDBSubnetGroupsInput, opts []func(*awsrds.Options)) (*awsrds.DescribeDBSubnetGroupsOutput, error) {
						return &awsrds.DescribeDBSubnetGroupsOutput{
							DBSubnetGroups: []awsrdstypes.DBSubnetGroup{{}},
						}, nil
					},
				},
				cr: dbSubnetGroup(),
			},
			want: want{
				cr:  dbSubnetGroup(withConditions(xpv1.Deleting())),
				err: errorutils.Wrap(errBoom, errDelete),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.client}
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
