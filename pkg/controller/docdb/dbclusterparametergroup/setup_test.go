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

package dbclusterparametergroup

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/docdb"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/docdb/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/docdb/fake"
	svcutils "github.com/crossplane-contrib/provider-aws/pkg/controller/docdb/utils"
	errorutils "github.com/crossplane-contrib/provider-aws/pkg/utils/errors"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
)

const (
	testDBClusterParameterGroupName = "some-db-cluster-parameter-group"
	testDBClusterParameterGroupARN  = "some-arn"
	testDescription                 = "some-description"
	testOtherDescription            = "some-other-description"
	testFamily                      = "some-family"
	testOtherFamily                 = "some-other-family"
	testParameterName               = "some-parameter-name"
	testParameterValue              = "some-parameter-value"
	testOtherParameterName          = "some-other-parameter-name"
	testOtherParameterValue         = "some-other-parameter-value"
	testTagKey                      = "some-tag-key"
	testTagValue                    = "some-tag-value"
	testOtherTagKey                 = "some-other-tag-key"
	testOtherTagValue               = "some-other-tag-value"

	testErrDescribeDBClusterParametersFailed      = "DescribeDBClusterParameters failed"
	testErrDescribeDBClusterParameterGroupsFailed = "DescribeDBClusterParameterGroups failed"
	testErrCreateDBClusterParameterGroupFailed    = "CreateDBClusterParameterGroup failed"
	testErrDeleteDBClusterParameterGroupFailed    = "DeleteDBClusterParameterGroup failed"
	testErrModifyDBClusterParameterGroupFailed    = "ModifyDBClusterParameterGroup failed"
	testErrBoom                                   = "boom"
)

type args struct {
	docdb *fake.MockDocDBClient
	kube  client.Client
	cr    *svcapitypes.DBClusterParameterGroup
}

type docDBModifier func(*svcapitypes.DBClusterParameterGroup)

func instance(m ...docDBModifier) *svcapitypes.DBClusterParameterGroup {
	cr := &svcapitypes.DBClusterParameterGroup{}
	for _, f := range m {
		f(cr)
	}
	return cr
}

func withExternalName(value string) docDBModifier {
	return func(o *svcapitypes.DBClusterParameterGroup) {
		meta.SetExternalName(o, value)

	}
}

func withDBClusterParameterGroupName(value string) docDBModifier {
	return func(o *svcapitypes.DBClusterParameterGroup) {
		o.Status.AtProvider.DBClusterParameterGroupName = pointer.String(value)
	}
}

func withDescription(value string) docDBModifier {
	return func(o *svcapitypes.DBClusterParameterGroup) {
		o.Spec.ForProvider.Description = pointer.String(value)
	}
}

func withDBParameterGroupFamily(value string) docDBModifier {
	return func(o *svcapitypes.DBClusterParameterGroup) {
		o.Spec.ForProvider.DBParameterGroupFamily = pointer.String(value)
	}
}

func withDBClusterParameterGroupARN(value string) docDBModifier {
	return func(o *svcapitypes.DBClusterParameterGroup) {
		o.Status.AtProvider.DBClusterParameterGroupARN = pointer.String(value)
	}
}

func withConditions(value ...xpv1.Condition) docDBModifier {
	return func(o *svcapitypes.DBClusterParameterGroup) {
		o.Status.SetConditions(value...)
	}
}

func withParameters(values ...*svcapitypes.CustomParameter) docDBModifier {
	return func(o *svcapitypes.DBClusterParameterGroup) {
		if values == nil {
			o.Spec.ForProvider.Parameters = []*svcapitypes.CustomParameter{}
		} else {
			o.Spec.ForProvider.Parameters = values
		}
	}
}

func withTags(values ...*svcapitypes.Tag) docDBModifier {
	return func(o *svcapitypes.DBClusterParameterGroup) {
		if values != nil {
			o.Spec.ForProvider.Tags = values
		} else {
			o.Spec.ForProvider.Tags = []*svcapitypes.Tag{}
		}
	}
}

func mergeTags(lists ...[]*svcapitypes.Tag) []*svcapitypes.Tag {
	res := []*svcapitypes.Tag{}
	for _, list := range lists {
		res = append(res, list...)
	}
	return res
}

func TestObserve(t *testing.T) {
	type want struct {
		cr     *svcapitypes.DBClusterParameterGroup
		result managed.ExternalObservation
		err    error
		docdb  fake.MockDocDBClientCall
	}

	cases := map[string]struct {
		args
		want
	}{
		"AvailableState_and_UpToDate_with_no_parameters": {
			args: args{
				docdb: &fake.MockDocDBClient{
					MockDescribeDBClusterParameterGroupsWithContext: func(c context.Context, ddpgi *docdb.DescribeDBClusterParameterGroupsInput, o []request.Option) (*docdb.DescribeDBClusterParameterGroupsOutput, error) {
						return &docdb.DescribeDBClusterParameterGroupsOutput{
							DBClusterParameterGroups: []*docdb.DBClusterParameterGroup{
								{
									DBClusterParameterGroupName: pointer.String(testDBClusterParameterGroupName),
								},
							},
						}, nil
					},
					MockDescribeDBClusterParametersWithContext: func(ctx context.Context, ddpi *docdb.DescribeDBClusterParametersInput, o []request.Option) (*docdb.DescribeDBClusterParametersOutput, error) {
						return &docdb.DescribeDBClusterParametersOutput{
							Parameters: []*docdb.Parameter{},
						}, nil
					},
					MockListTagsForResource: func(ltfri *docdb.ListTagsForResourceInput) (*docdb.ListTagsForResourceOutput, error) {
						return &docdb.ListTagsForResourceOutput{}, nil
					},
				},
				cr: instance(
					withExternalName(testDBClusterParameterGroupName),
				),
			},
			want: want{
				cr: instance(
					withDBClusterParameterGroupName(testDBClusterParameterGroupName),
					withExternalName(testDBClusterParameterGroupName),
					withParameters(),
					withConditions(xpv1.Available()),
				),
				result: managed.ExternalObservation{
					ResourceExists:          true,
					ResourceUpToDate:        true,
					ResourceLateInitialized: true,
				},
				docdb: fake.MockDocDBClientCall{
					DescribeDBClusterParameterGroupsWithContext: []*fake.CallDescribeDBClusterParameterGroupsWithContext{
						{
							Ctx: context.Background(),
							I: &docdb.DescribeDBClusterParameterGroupsInput{
								DBClusterParameterGroupName: pointer.String(testDBClusterParameterGroupName),
							},
						},
					},
					DescribeDBClusterParametersWithContext: []*fake.CallDescribeDBClusterParametersWithContext{
						{
							Ctx: context.Background(),
							I: &docdb.DescribeDBClusterParametersInput{
								DBClusterParameterGroupName: pointer.String(testDBClusterParameterGroupName),
							},
						},
						{
							Ctx: context.Background(),
							I: &docdb.DescribeDBClusterParametersInput{
								DBClusterParameterGroupName: pointer.String(testDBClusterParameterGroupName),
							},
						},
					},
					ListTagsForResource: []*fake.CallListTagsForResource{
						{
							I: &docdb.ListTagsForResourceInput{},
						},
					},
				},
			},
		},
		"AvailableState_and_UpToDate_with_one_parameter": {
			args: args{
				docdb: &fake.MockDocDBClient{
					MockDescribeDBClusterParameterGroupsWithContext: func(c context.Context, ddpgi *docdb.DescribeDBClusterParameterGroupsInput, o []request.Option) (*docdb.DescribeDBClusterParameterGroupsOutput, error) {
						return &docdb.DescribeDBClusterParameterGroupsOutput{
							DBClusterParameterGroups: []*docdb.DBClusterParameterGroup{
								{
									DBClusterParameterGroupName: pointer.String(testDBClusterParameterGroupName),
								},
							},
						}, nil
					},
					MockDescribeDBClusterParametersWithContext: func(ctx context.Context, ddpi *docdb.DescribeDBClusterParametersInput, o []request.Option) (*docdb.DescribeDBClusterParametersOutput, error) {
						return &docdb.DescribeDBClusterParametersOutput{
							Parameters: []*docdb.Parameter{
								{
									ParameterName:  pointer.String(testParameterName),
									ParameterValue: pointer.String(testParameterValue),
								},
							},
						}, nil
					},
					MockListTagsForResource: func(ltfri *docdb.ListTagsForResourceInput) (*docdb.ListTagsForResourceOutput, error) {
						return &docdb.ListTagsForResourceOutput{}, nil
					},
				},
				cr: instance(
					withExternalName(testDBClusterParameterGroupName),
					withParameters(
						&svcapitypes.CustomParameter{
							ParameterName:  pointer.String(testParameterName),
							ParameterValue: pointer.String(testParameterValue),
						},
					),
				),
			},
			want: want{
				cr: instance(
					withDBClusterParameterGroupName(testDBClusterParameterGroupName),
					withExternalName(testDBClusterParameterGroupName),
					withParameters(
						&svcapitypes.CustomParameter{
							ParameterName:  pointer.String(testParameterName),
							ParameterValue: pointer.String(testParameterValue),
						},
					),
					withConditions(xpv1.Available()),
				),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
				},
				docdb: fake.MockDocDBClientCall{
					DescribeDBClusterParameterGroupsWithContext: []*fake.CallDescribeDBClusterParameterGroupsWithContext{
						{
							Ctx: context.Background(),
							I: &docdb.DescribeDBClusterParameterGroupsInput{
								DBClusterParameterGroupName: pointer.String(testDBClusterParameterGroupName),
							},
						},
					},
					DescribeDBClusterParametersWithContext: []*fake.CallDescribeDBClusterParametersWithContext{
						{
							Ctx: context.Background(),
							I: &docdb.DescribeDBClusterParametersInput{
								DBClusterParameterGroupName: pointer.String(testDBClusterParameterGroupName),
							},
						},
						{
							Ctx: context.Background(),
							I: &docdb.DescribeDBClusterParametersInput{
								DBClusterParameterGroupName: pointer.String(testDBClusterParameterGroupName),
							},
						},
					},
					ListTagsForResource: []*fake.CallListTagsForResource{
						{
							I: &docdb.ListTagsForResourceInput{},
						},
					},
				},
			},
		},
		"AvailableState_and_UpToDate_with_one_spec_and_two_default_parameters": {
			args: args{
				docdb: &fake.MockDocDBClient{
					MockDescribeDBClusterParameterGroupsWithContext: func(c context.Context, ddpgi *docdb.DescribeDBClusterParameterGroupsInput, o []request.Option) (*docdb.DescribeDBClusterParameterGroupsOutput, error) {
						return &docdb.DescribeDBClusterParameterGroupsOutput{
							DBClusterParameterGroups: []*docdb.DBClusterParameterGroup{
								{
									DBClusterParameterGroupName: pointer.String(testDBClusterParameterGroupName),
								},
							},
						}, nil
					},
					MockDescribeDBClusterParametersWithContext: func(ctx context.Context, ddpi *docdb.DescribeDBClusterParametersInput, o []request.Option) (*docdb.DescribeDBClusterParametersOutput, error) {
						return &docdb.DescribeDBClusterParametersOutput{
							Parameters: []*docdb.Parameter{
								{
									ParameterName:  pointer.String(testParameterName),
									ParameterValue: pointer.String(testParameterValue),
								},
								{
									ParameterName:  pointer.String(testOtherParameterName),
									ParameterValue: pointer.String(testOtherParameterValue),
								},
							},
						}, nil
					},
					MockListTagsForResource: func(ltfri *docdb.ListTagsForResourceInput) (*docdb.ListTagsForResourceOutput, error) {
						return &docdb.ListTagsForResourceOutput{}, nil
					},
				},
				cr: instance(
					withExternalName(testDBClusterParameterGroupName),
					withParameters(
						&svcapitypes.CustomParameter{
							ParameterName:  pointer.String(testParameterName),
							ParameterValue: pointer.String(testParameterValue),
						},
					),
				),
			},
			want: want{
				cr: instance(
					withDBClusterParameterGroupName(testDBClusterParameterGroupName),
					withExternalName(testDBClusterParameterGroupName),
					withParameters(
						&svcapitypes.CustomParameter{
							ParameterName:  pointer.String(testParameterName),
							ParameterValue: pointer.String(testParameterValue),
						},
						&svcapitypes.CustomParameter{
							ParameterName:  pointer.String(testOtherParameterName),
							ParameterValue: pointer.String(testOtherParameterValue),
						},
					),
					withConditions(xpv1.Available()),
				),
				result: managed.ExternalObservation{
					ResourceExists:          true,
					ResourceUpToDate:        true,
					ResourceLateInitialized: true,
				},
				docdb: fake.MockDocDBClientCall{
					DescribeDBClusterParameterGroupsWithContext: []*fake.CallDescribeDBClusterParameterGroupsWithContext{
						{
							Ctx: context.Background(),
							I: &docdb.DescribeDBClusterParameterGroupsInput{
								DBClusterParameterGroupName: pointer.String(testDBClusterParameterGroupName),
							},
						},
					},
					DescribeDBClusterParametersWithContext: []*fake.CallDescribeDBClusterParametersWithContext{
						{
							Ctx: context.Background(),
							I: &docdb.DescribeDBClusterParametersInput{
								DBClusterParameterGroupName: pointer.String(testDBClusterParameterGroupName),
							},
						},
						{
							Ctx: context.Background(),
							I: &docdb.DescribeDBClusterParametersInput{
								DBClusterParameterGroupName: pointer.String(testDBClusterParameterGroupName),
							},
						},
					},
					ListTagsForResource: []*fake.CallListTagsForResource{
						{
							I: &docdb.ListTagsForResourceInput{},
						},
					},
				},
			},
		},
		"AvailableState_and_changed_description": {
			args: args{
				docdb: &fake.MockDocDBClient{
					MockDescribeDBClusterParameterGroupsWithContext: func(c context.Context, ddpgi *docdb.DescribeDBClusterParameterGroupsInput, o []request.Option) (*docdb.DescribeDBClusterParameterGroupsOutput, error) {
						return &docdb.DescribeDBClusterParameterGroupsOutput{
							DBClusterParameterGroups: []*docdb.DBClusterParameterGroup{
								{
									DBClusterParameterGroupName: pointer.String(testDBClusterParameterGroupName),
									Description:                 pointer.String(testDescription),
								},
							},
						}, nil
					},
					MockDescribeDBClusterParametersWithContext: func(ctx context.Context, ddpi *docdb.DescribeDBClusterParametersInput, o []request.Option) (*docdb.DescribeDBClusterParametersOutput, error) {
						return &docdb.DescribeDBClusterParametersOutput{
							Parameters: []*docdb.Parameter{},
						}, nil
					},
					MockListTagsForResource: func(ltfri *docdb.ListTagsForResourceInput) (*docdb.ListTagsForResourceOutput, error) {
						return &docdb.ListTagsForResourceOutput{}, nil
					},
				},
				cr: instance(
					withExternalName(testDBClusterParameterGroupName),
					withDescription(testOtherDescription),
				),
			},
			want: want{
				cr: instance(
					withDBClusterParameterGroupName(testDBClusterParameterGroupName),
					withExternalName(testDBClusterParameterGroupName),
					withDescription(testOtherDescription),
					withParameters(),
				),
				result: managed.ExternalObservation{
					ResourceExists:   false,
					ResourceUpToDate: false,
				},
				err: errors.Wrap(errors.New(errModifyDescription), "isUpToDate check failed"),
				docdb: fake.MockDocDBClientCall{
					DescribeDBClusterParameterGroupsWithContext: []*fake.CallDescribeDBClusterParameterGroupsWithContext{
						{
							Ctx: context.Background(),
							I: &docdb.DescribeDBClusterParameterGroupsInput{
								DBClusterParameterGroupName: pointer.String(testDBClusterParameterGroupName),
							},
						},
					},
					DescribeDBClusterParametersWithContext: []*fake.CallDescribeDBClusterParametersWithContext{
						{
							Ctx: context.Background(),
							I: &docdb.DescribeDBClusterParametersInput{
								DBClusterParameterGroupName: pointer.String(testDBClusterParameterGroupName),
							},
						},
					},
				},
			},
		},
		"AvailableState_and_changed_DBParameterGroupFamily": {
			args: args{
				docdb: &fake.MockDocDBClient{
					MockDescribeDBClusterParameterGroupsWithContext: func(c context.Context, ddpgi *docdb.DescribeDBClusterParameterGroupsInput, o []request.Option) (*docdb.DescribeDBClusterParameterGroupsOutput, error) {
						return &docdb.DescribeDBClusterParameterGroupsOutput{
							DBClusterParameterGroups: []*docdb.DBClusterParameterGroup{
								{
									DBClusterParameterGroupName: pointer.String(testDBClusterParameterGroupName),
									DBParameterGroupFamily:      pointer.String(testFamily),
								},
							},
						}, nil
					},
					MockDescribeDBClusterParametersWithContext: func(ctx context.Context, ddpi *docdb.DescribeDBClusterParametersInput, o []request.Option) (*docdb.DescribeDBClusterParametersOutput, error) {
						return &docdb.DescribeDBClusterParametersOutput{
							Parameters: []*docdb.Parameter{},
						}, nil
					},
					MockListTagsForResource: func(ltfri *docdb.ListTagsForResourceInput) (*docdb.ListTagsForResourceOutput, error) {
						return &docdb.ListTagsForResourceOutput{}, nil
					},
				},
				cr: instance(
					withExternalName(testDBClusterParameterGroupName),
					withDBParameterGroupFamily(testOtherFamily),
				),
			},
			want: want{
				cr: instance(
					withDBClusterParameterGroupName(testDBClusterParameterGroupName),
					withExternalName(testDBClusterParameterGroupName),
					withDBParameterGroupFamily(testOtherFamily),
					withParameters(),
				),
				result: managed.ExternalObservation{
					ResourceExists:   false,
					ResourceUpToDate: false,
				},
				err: errors.Wrap(errors.New(errModifyFamily), "isUpToDate check failed"),
				docdb: fake.MockDocDBClientCall{
					DescribeDBClusterParameterGroupsWithContext: []*fake.CallDescribeDBClusterParameterGroupsWithContext{
						{
							Ctx: context.Background(),
							I: &docdb.DescribeDBClusterParameterGroupsInput{
								DBClusterParameterGroupName: pointer.String(testDBClusterParameterGroupName),
							},
						},
					},
					DescribeDBClusterParametersWithContext: []*fake.CallDescribeDBClusterParametersWithContext{
						{
							Ctx: context.Background(),
							I: &docdb.DescribeDBClusterParametersInput{
								DBClusterParameterGroupName: pointer.String(testDBClusterParameterGroupName),
							},
						},
					},
				},
			},
		},
		"AvailableState_and_changed_parameter": {
			args: args{
				docdb: &fake.MockDocDBClient{
					MockDescribeDBClusterParameterGroupsWithContext: func(c context.Context, ddpgi *docdb.DescribeDBClusterParameterGroupsInput, o []request.Option) (*docdb.DescribeDBClusterParameterGroupsOutput, error) {
						return &docdb.DescribeDBClusterParameterGroupsOutput{
							DBClusterParameterGroups: []*docdb.DBClusterParameterGroup{
								{
									DBClusterParameterGroupName: pointer.String(testDBClusterParameterGroupName),
								},
							},
						}, nil
					},
					MockDescribeDBClusterParametersWithContext: func(ctx context.Context, ddpi *docdb.DescribeDBClusterParametersInput, o []request.Option) (*docdb.DescribeDBClusterParametersOutput, error) {
						return &docdb.DescribeDBClusterParametersOutput{
							Parameters: []*docdb.Parameter{
								{
									ParameterName:  pointer.String(testParameterName),
									ParameterValue: pointer.String(testParameterValue),
								},
							},
						}, nil
					},
				},
				cr: instance(
					withExternalName(testDBClusterParameterGroupName),
					withParameters(
						&svcapitypes.CustomParameter{
							ParameterName:  pointer.String(testParameterName),
							ParameterValue: pointer.String(testOtherParameterValue),
						},
					),
				),
			},
			want: want{
				cr: instance(
					withDBClusterParameterGroupName(testDBClusterParameterGroupName),
					withExternalName(testDBClusterParameterGroupName),
					withParameters(
						&svcapitypes.CustomParameter{
							ParameterName:  pointer.String(testParameterName),
							ParameterValue: pointer.String(testOtherParameterValue),
						},
					),
					withConditions(xpv1.Available()),
				),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: false,
				},
				docdb: fake.MockDocDBClientCall{
					DescribeDBClusterParameterGroupsWithContext: []*fake.CallDescribeDBClusterParameterGroupsWithContext{
						{
							Ctx: context.Background(),
							I: &docdb.DescribeDBClusterParameterGroupsInput{
								DBClusterParameterGroupName: pointer.String(testDBClusterParameterGroupName),
							},
						},
					},
					DescribeDBClusterParametersWithContext: []*fake.CallDescribeDBClusterParametersWithContext{
						{
							Ctx: context.Background(),
							I: &docdb.DescribeDBClusterParametersInput{
								DBClusterParameterGroupName: pointer.String(testDBClusterParameterGroupName),
							},
						},
						{
							Ctx: context.Background(),
							I: &docdb.DescribeDBClusterParametersInput{
								DBClusterParameterGroupName: pointer.String(testDBClusterParameterGroupName),
							},
						},
					},
				},
			},
		},
		"ErrDescribeDBClusterParameterGroupsWithContext": {
			args: args{
				docdb: &fake.MockDocDBClient{
					MockDescribeDBClusterParameterGroupsWithContext: func(c context.Context, ddpgi *docdb.DescribeDBClusterParameterGroupsInput, o []request.Option) (*docdb.DescribeDBClusterParameterGroupsOutput, error) {
						return &docdb.DescribeDBClusterParameterGroupsOutput{}, errors.New(testErrDescribeDBClusterParameterGroupsFailed)
					},
				},
				cr: instance(
					withExternalName(testDBClusterParameterGroupName),
				),
			},
			want: want{
				cr: instance(
					withExternalName(testDBClusterParameterGroupName),
				),
				result: managed.ExternalObservation{
					ResourceExists:   false,
					ResourceUpToDate: false,
				},
				err: errors.Wrap(errors.New(testErrDescribeDBClusterParameterGroupsFailed), errDescribe),
				docdb: fake.MockDocDBClientCall{
					DescribeDBClusterParameterGroupsWithContext: []*fake.CallDescribeDBClusterParameterGroupsWithContext{
						{
							Ctx: context.Background(),
							I: &docdb.DescribeDBClusterParameterGroupsInput{
								DBClusterParameterGroupName: pointer.String(testDBClusterParameterGroupName),
							},
						},
					},
				},
			},
		},
		"ErrDescribeDBClusterParameters": {
			args: args{
				docdb: &fake.MockDocDBClient{
					MockDescribeDBClusterParameterGroupsWithContext: func(c context.Context, ddpgi *docdb.DescribeDBClusterParameterGroupsInput, o []request.Option) (*docdb.DescribeDBClusterParameterGroupsOutput, error) {
						return &docdb.DescribeDBClusterParameterGroupsOutput{
							DBClusterParameterGroups: []*docdb.DBClusterParameterGroup{
								{
									DBClusterParameterGroupName: pointer.String(testDBClusterParameterGroupName),
								},
							},
						}, nil
					},
					MockDescribeDBClusterParametersWithContext: func(ctx context.Context, ddpi *docdb.DescribeDBClusterParametersInput, o []request.Option) (*docdb.DescribeDBClusterParametersOutput, error) {
						return &docdb.DescribeDBClusterParametersOutput{}, errors.New(testErrDescribeDBClusterParametersFailed)
					},
				},
				cr: instance(
					withExternalName(testDBClusterParameterGroupName),
				),
			},
			want: want{
				cr: instance(
					withExternalName(testDBClusterParameterGroupName),
				),
				result: managed.ExternalObservation{
					ResourceExists:   false,
					ResourceUpToDate: false,
				},
				err: errors.Wrap(errors.Wrap(errors.New(testErrDescribeDBClusterParametersFailed), errDescribeParameters), "late-init failed"),
				docdb: fake.MockDocDBClientCall{
					DescribeDBClusterParameterGroupsWithContext: []*fake.CallDescribeDBClusterParameterGroupsWithContext{
						{
							Ctx: context.Background(),
							I: &docdb.DescribeDBClusterParameterGroupsInput{
								DBClusterParameterGroupName: pointer.String(testDBClusterParameterGroupName),
							},
						},
					},
					DescribeDBClusterParametersWithContext: []*fake.CallDescribeDBClusterParametersWithContext{
						{
							Ctx: context.Background(),
							I: &docdb.DescribeDBClusterParametersInput{
								DBClusterParameterGroupName: pointer.String(testDBClusterParameterGroupName),
							},
						},
					},
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			opts := []option{setupExternal}
			e := newExternal(tc.args.kube, tc.args.docdb, opts)
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
			if diff := cmp.Diff(tc.want.docdb, tc.args.docdb.Called, cmpopts.IgnoreInterfaces(struct{ context.Context }{})); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestCreate(t *testing.T) {
	type want struct {
		cr     *svcapitypes.DBClusterParameterGroup
		result managed.ExternalCreation
		err    error
		docdb  fake.MockDocDBClientCall
	}

	cases := map[string]struct {
		args
		want
	}{
		"SuccessfulCreate_no_parameters": {
			args: args{
				docdb: &fake.MockDocDBClient{
					MockCreateDBClusterParameterGroupWithContext: func(c context.Context, cdpgi *docdb.CreateDBClusterParameterGroupInput, o []request.Option) (*docdb.CreateDBClusterParameterGroupOutput, error) {
						return &docdb.CreateDBClusterParameterGroupOutput{
							DBClusterParameterGroup: &docdb.DBClusterParameterGroup{
								DBClusterParameterGroupName: pointer.String(testDBClusterParameterGroupName),
							},
						}, nil
					},
				},
				cr: instance(
					withExternalName(testDBClusterParameterGroupName),
				),
			},
			want: want{
				cr: instance(
					withDBClusterParameterGroupName(testDBClusterParameterGroupName),
					withExternalName(testDBClusterParameterGroupName),
					withConditions(xpv1.Creating()),
				),
				result: managed.ExternalCreation{},
				docdb: fake.MockDocDBClientCall{
					CreateDBClusterParameterGroupWithContext: []*fake.CallCreateDBClusterParameterGroupWithContext{
						{
							Ctx: context.Background(),
							I: &docdb.CreateDBClusterParameterGroupInput{
								DBClusterParameterGroupName: pointer.String(testDBClusterParameterGroupName),
							},
						},
					},
				},
			},
		},
		"SuccessfulCreate_with_parameters": {
			args: args{
				docdb: &fake.MockDocDBClient{
					MockCreateDBClusterParameterGroupWithContext: func(c context.Context, cdpgi *docdb.CreateDBClusterParameterGroupInput, o []request.Option) (*docdb.CreateDBClusterParameterGroupOutput, error) {
						return &docdb.CreateDBClusterParameterGroupOutput{
							DBClusterParameterGroup: &docdb.DBClusterParameterGroup{
								DBClusterParameterGroupName: pointer.String(testDBClusterParameterGroupName),
							},
						}, nil
					},
				},
				cr: instance(
					withExternalName(testDBClusterParameterGroupName),
					withParameters(
						&svcapitypes.CustomParameter{
							ParameterName:  pointer.String(testParameterName),
							ParameterValue: pointer.String(testParameterValue),
						},
						&svcapitypes.CustomParameter{
							ParameterName:  pointer.String(testOtherParameterName),
							ParameterValue: pointer.String(testOtherParameterValue),
						},
					),
				),
			},
			want: want{
				cr: instance(
					withDBClusterParameterGroupName(testDBClusterParameterGroupName),
					withExternalName(testDBClusterParameterGroupName),
					withConditions(xpv1.Creating()),
					withParameters(
						&svcapitypes.CustomParameter{
							ParameterName:  pointer.String(testParameterName),
							ParameterValue: pointer.String(testParameterValue),
						},
						&svcapitypes.CustomParameter{
							ParameterName:  pointer.String(testOtherParameterName),
							ParameterValue: pointer.String(testOtherParameterValue),
						},
					),
				),
				result: managed.ExternalCreation{},
				docdb: fake.MockDocDBClientCall{
					CreateDBClusterParameterGroupWithContext: []*fake.CallCreateDBClusterParameterGroupWithContext{
						{
							Ctx: context.Background(),
							I: &docdb.CreateDBClusterParameterGroupInput{
								DBClusterParameterGroupName: pointer.String(testDBClusterParameterGroupName),
							},
						},
					},
				},
			},
		},
		"ErrorCreate": {
			args: args{
				docdb: &fake.MockDocDBClient{
					MockCreateDBClusterParameterGroupWithContext: func(c context.Context, cdpgi *docdb.CreateDBClusterParameterGroupInput, o []request.Option) (*docdb.CreateDBClusterParameterGroupOutput, error) {
						return &docdb.CreateDBClusterParameterGroupOutput{}, errors.New(testErrCreateDBClusterParameterGroupFailed)
					},
				},
				cr: instance(
					withExternalName(testDBClusterParameterGroupName),
				),
			},
			want: want{
				cr: instance(
					withExternalName(testDBClusterParameterGroupName),
					withConditions(xpv1.Creating()),
				),
				result: managed.ExternalCreation{},
				err:    errors.Wrap(errors.New(testErrCreateDBClusterParameterGroupFailed), errCreate),
				docdb: fake.MockDocDBClientCall{
					CreateDBClusterParameterGroupWithContext: []*fake.CallCreateDBClusterParameterGroupWithContext{
						{
							Ctx: context.Background(),
							I: &docdb.CreateDBClusterParameterGroupInput{
								DBClusterParameterGroupName: pointer.String(testDBClusterParameterGroupName),
							},
						},
					},
				},
			},
		},
		"ErrorCreate_Parameters": {
			args: args{
				docdb: &fake.MockDocDBClient{
					MockCreateDBClusterParameterGroupWithContext: func(c context.Context, cdpgi *docdb.CreateDBClusterParameterGroupInput, o []request.Option) (*docdb.CreateDBClusterParameterGroupOutput, error) {
						return &docdb.CreateDBClusterParameterGroupOutput{
							DBClusterParameterGroup: &docdb.DBClusterParameterGroup{
								DBClusterParameterGroupName: pointer.String(testDBClusterParameterGroupName),
							},
						}, nil
					},
				},
				cr: instance(
					withExternalName(testDBClusterParameterGroupName),
					withParameters(
						&svcapitypes.CustomParameter{
							ParameterName:  pointer.String(testParameterName),
							ParameterValue: pointer.String(testParameterValue),
						},
					),
				),
			},
			want: want{
				cr: instance(
					withDBClusterParameterGroupName(testDBClusterParameterGroupName),
					withExternalName(testDBClusterParameterGroupName),
					withConditions(xpv1.Creating()),
					withParameters(
						&svcapitypes.CustomParameter{
							ParameterName:  pointer.String(testParameterName),
							ParameterValue: pointer.String(testParameterValue),
						},
					),
				),
				result: managed.ExternalCreation{},
				docdb: fake.MockDocDBClientCall{
					CreateDBClusterParameterGroupWithContext: []*fake.CallCreateDBClusterParameterGroupWithContext{
						{
							Ctx: context.Background(),
							I: &docdb.CreateDBClusterParameterGroupInput{
								DBClusterParameterGroupName: pointer.String(testDBClusterParameterGroupName),
							},
						},
					},
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			opts := []option{setupExternal}
			e := newExternal(tc.args.kube, tc.args.docdb, opts)
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
			if diff := cmp.Diff(tc.want.docdb, tc.args.docdb.Called, cmpopts.IgnoreInterfaces(struct{ context.Context }{})); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	type want struct {
		cr    *svcapitypes.DBClusterParameterGroup
		err   error
		docdb fake.MockDocDBClientCall
	}

	cases := map[string]struct {
		args
		want
	}{
		"SuccessfulDelete": {
			args: args{
				docdb: &fake.MockDocDBClient{
					MockDeleteDBClusterParameterGroupWithContext: func(c context.Context, ddpgi *docdb.DeleteDBClusterParameterGroupInput, o []request.Option) (*docdb.DeleteDBClusterParameterGroupOutput, error) {
						return &docdb.DeleteDBClusterParameterGroupOutput{}, nil
					},
				},
				cr: instance(
					withExternalName(testDBClusterParameterGroupName),
				),
			},
			want: want{
				cr: instance(
					withExternalName(testDBClusterParameterGroupName),
					withConditions(xpv1.Deleting()),
				),
				docdb: fake.MockDocDBClientCall{
					DeleteDBClusterParameterGroupWithContext: []*fake.CallDeleteDBClusterParameterGroupWithContext{
						{
							Ctx: context.Background(),
							I: &docdb.DeleteDBClusterParameterGroupInput{
								DBClusterParameterGroupName: pointer.String(testDBClusterParameterGroupName),
							},
						},
					},
				},
			},
		},
		"ErrorDelete": {
			args: args{
				docdb: &fake.MockDocDBClient{
					MockDeleteDBClusterParameterGroupWithContext: func(c context.Context, ddpgi *docdb.DeleteDBClusterParameterGroupInput, o []request.Option) (*docdb.DeleteDBClusterParameterGroupOutput, error) {
						return &docdb.DeleteDBClusterParameterGroupOutput{}, errors.New(testErrDeleteDBClusterParameterGroupFailed)
					},
				},
				cr: instance(
					withExternalName(testDBClusterParameterGroupName),
				),
			},
			want: want{
				cr: instance(
					withExternalName(testDBClusterParameterGroupName),
					withConditions(xpv1.Deleting()),
				),
				err: errors.Wrap(errors.New(testErrDeleteDBClusterParameterGroupFailed), errDelete),
				docdb: fake.MockDocDBClientCall{
					DeleteDBClusterParameterGroupWithContext: []*fake.CallDeleteDBClusterParameterGroupWithContext{
						{
							Ctx: context.Background(),
							I: &docdb.DeleteDBClusterParameterGroupInput{
								DBClusterParameterGroupName: pointer.String(testDBClusterParameterGroupName),
							},
						},
					},
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			opts := []option{setupExternal}
			e := newExternal(tc.args.kube, tc.args.docdb, opts)
			err := e.Delete(context.Background(), tc.args.cr)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.cr, tc.args.cr, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.docdb, tc.args.docdb.Called, cmpopts.IgnoreInterfaces(struct{ context.Context }{})); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestUpdate(t *testing.T) {
	type want struct {
		cr     *svcapitypes.DBClusterParameterGroup
		err    error
		result managed.ExternalUpdate
		docdb  fake.MockDocDBClientCall
	}

	cases := map[string]struct {
		args
		want
	}{
		"SuccessfulUpdate_parameters": {
			args: args{
				docdb: &fake.MockDocDBClient{
					MockModifyDBClusterParameterGroupWithContext: func(c context.Context, mdpgi *docdb.ModifyDBClusterParameterGroupInput, o []request.Option) (*docdb.ModifyDBClusterParameterGroupOutput, error) {
						return &docdb.ModifyDBClusterParameterGroupOutput{
							DBClusterParameterGroupName: pointer.String(testDBClusterParameterGroupName),
						}, nil
					},
					MockListTagsForResource: func(ltfri *docdb.ListTagsForResourceInput) (*docdb.ListTagsForResourceOutput, error) {
						return &docdb.ListTagsForResourceOutput{
							TagList: []*docdb.Tag{},
						}, nil
					},
					MockAddTagsToResource: func(attri *docdb.AddTagsToResourceInput) (*docdb.AddTagsToResourceOutput, error) {
						return &docdb.AddTagsToResourceOutput{}, nil
					},
				},
				cr: instance(
					withDBClusterParameterGroupARN(testDBClusterParameterGroupARN),
					withExternalName(testDBClusterParameterGroupName),
					withParameters(
						&svcapitypes.CustomParameter{
							ParameterName:  pointer.String(testParameterName),
							ParameterValue: pointer.String(testParameterValue),
						},
						&svcapitypes.CustomParameter{
							ParameterName:  pointer.String(testOtherParameterName),
							ParameterValue: pointer.String(testOtherParameterValue),
						},
					),
					withTags(
						&svcapitypes.Tag{Key: pointer.String(testTagKey), Value: pointer.String(testTagValue)},
						&svcapitypes.Tag{Key: pointer.String(testOtherTagKey), Value: pointer.String(testOtherTagValue)},
					),
				),
			},
			want: want{
				cr: instance(
					withDBClusterParameterGroupARN(testDBClusterParameterGroupARN),
					withExternalName(testDBClusterParameterGroupName),
					withConditions(xpv1.Available()),
					withParameters(
						&svcapitypes.CustomParameter{
							ParameterName:  pointer.String(testParameterName),
							ParameterValue: pointer.String(testParameterValue),
						},
						&svcapitypes.CustomParameter{
							ParameterName:  pointer.String(testOtherParameterName),
							ParameterValue: pointer.String(testOtherParameterValue),
						},
					),
					withTags(
						&svcapitypes.Tag{Key: pointer.String(testTagKey), Value: pointer.String(testTagValue)},
						&svcapitypes.Tag{Key: pointer.String(testOtherTagKey), Value: pointer.String(testOtherTagValue)},
					),
				),
				docdb: fake.MockDocDBClientCall{
					ModifyDBClusterParameterGroupWithContext: []*fake.CallModifyDBClusterParameterGroupWithContext{
						{
							Ctx: context.Background(),
							I: &docdb.ModifyDBClusterParameterGroupInput{
								DBClusterParameterGroupName: pointer.String(testDBClusterParameterGroupName),
								Parameters: []*docdb.Parameter{
									{
										ParameterName:  pointer.String(testParameterName),
										ParameterValue: pointer.String(testParameterValue),
									},
									{
										ParameterName:  pointer.String(testOtherParameterName),
										ParameterValue: pointer.String(testOtherParameterValue),
									},
								},
							},
						},
					},
					ListTagsForResource: []*fake.CallListTagsForResource{
						{
							I: &docdb.ListTagsForResourceInput{
								ResourceName: pointer.String(testDBClusterParameterGroupARN),
							},
						},
					},
					AddTagsToResource: []*fake.CallAddTagsToResource{
						{
							I: &docdb.AddTagsToResourceInput{
								ResourceName: pointer.String(testDBClusterParameterGroupARN),
								Tags: []*docdb.Tag{
									{Key: pointer.String(testTagKey), Value: pointer.String(testTagValue)},
									{Key: pointer.String(testOtherTagKey), Value: pointer.String(testOtherTagValue)},
								},
							},
						},
					},
				},
			},
		},
		"ErrorUpdate": {
			args: args{
				docdb: &fake.MockDocDBClient{
					MockModifyDBClusterParameterGroupWithContext: func(c context.Context, mdpgi *docdb.ModifyDBClusterParameterGroupInput, o []request.Option) (*docdb.ModifyDBClusterParameterGroupOutput, error) {
						return &docdb.ModifyDBClusterParameterGroupOutput{}, errors.New(testErrModifyDBClusterParameterGroupFailed)
					},
				},
				cr: instance(
					withExternalName(testDBClusterParameterGroupName),
				),
			},
			want: want{
				cr: instance(
					withExternalName(testDBClusterParameterGroupName),
				),
				err: errors.Wrap(errors.New(testErrModifyDBClusterParameterGroupFailed), errUpdate),
				docdb: fake.MockDocDBClientCall{
					ModifyDBClusterParameterGroupWithContext: []*fake.CallModifyDBClusterParameterGroupWithContext{
						{
							Ctx: context.Background(),
							I: &docdb.ModifyDBClusterParameterGroupInput{
								DBClusterParameterGroupName: pointer.String(testDBClusterParameterGroupName),
								Parameters:                  []*docdb.Parameter{},
							},
						},
					},
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			opts := []option{setupExternal}
			e := newExternal(tc.args.kube, tc.args.docdb, opts)
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
			if diff := cmp.Diff(tc.want.docdb, tc.args.docdb.Called, cmpopts.IgnoreInterfaces(struct{ context.Context }{})); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestInitialize(t *testing.T) {
	type want struct {
		cr  *svcapitypes.DBClusterParameterGroup
		err error
	}

	cases := map[string]struct {
		args
		want
	}{
		"Successful": {
			args: args{
				cr: instance(withTags(
					&svcapitypes.Tag{Key: pointer.String(testTagKey), Value: pointer.String(testTagValue)},
					&svcapitypes.Tag{Key: pointer.String(testOtherTagKey), Value: pointer.String(testOtherTagValue)},
				)),
				kube: &test.MockClient{MockUpdate: test.NewMockUpdateFn(nil)},
			},
			want: want{
				cr: instance(withTags(
					mergeTags(
						[]*svcapitypes.Tag{
							{Key: pointer.String(testTagKey), Value: pointer.String(testTagValue)},
							{Key: pointer.String(testOtherTagKey), Value: pointer.String(testOtherTagValue)},
						},
						svcutils.GetExternalTags(instance()),
					)...,
				)),
			},
		},
		"UpdateFailed": {
			args: args{
				cr: instance(withTags(
					&svcapitypes.Tag{Key: pointer.String(testTagKey), Value: pointer.String(testTagValue)},
					&svcapitypes.Tag{Key: pointer.String(testOtherTagKey), Value: pointer.String(testOtherTagValue)},
				)),
				kube: &test.MockClient{MockUpdate: test.NewMockUpdateFn(errors.New(testErrBoom))},
			},
			want: want{
				cr: instance(withTags(
					mergeTags(
						[]*svcapitypes.Tag{
							{Key: pointer.String(testTagKey), Value: pointer.String(testTagValue)},
							{Key: pointer.String(testOtherTagKey), Value: pointer.String(testOtherTagValue)},
						},
						svcutils.GetExternalTags(instance()),
					)...,
				)),
				err: errorutils.Wrap(errors.New(testErrBoom), errKubeUpdateFailed),
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
			if diff := cmp.Diff(tc.want.cr, tc.args.cr, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}
