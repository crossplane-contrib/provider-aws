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

package dbsubnetgroup

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
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
)

const (
	testDBSubnetGroupName = "some-db-subnet-group"
	testDescription       = "some-description"
	testOtherDescription  = "some-other-description"
	testSubnetID          = "subnet-1"
	testOtherSubnetID     = "subnet-2"

	testErrDescribeDBSubnetGroupsFailed = "DescribeDBSubnetGroups failed"
	testErrCreateDBSubnetGroupFailed    = "CreateDBSubnetGroup failed"
	testErrDeleteDBSubnetGroupFailed    = "DeleteDBSubnetGroup failed"
	testErrModifyDBSubnetGroupFailed    = "ModifyDBSubnetGroup failed"
)

type args struct {
	docdb *fake.MockDocDBClient
	kube  client.Client
	cr    *svcapitypes.DBSubnetGroup
}

type docDBModifier func(*svcapitypes.DBSubnetGroup)

func subnetGroup(m ...docDBModifier) *svcapitypes.DBSubnetGroup {
	cr := &svcapitypes.DBSubnetGroup{}
	for _, f := range m {
		f(cr)
	}
	return cr
}

func withExternalName(value string) docDBModifier {
	return func(o *svcapitypes.DBSubnetGroup) {
		meta.SetExternalName(o, value)
	}
}

func withDBSubnetGroupName(value string) docDBModifier {
	return func(o *svcapitypes.DBSubnetGroup) {
		o.Status.AtProvider.DBSubnetGroupName = pointer.ToOrNilIfZeroValue(value)
	}
}

func withDescription(value string) docDBModifier {
	return func(o *svcapitypes.DBSubnetGroup) {
		o.Spec.ForProvider.DBSubnetGroupDescription = pointer.ToOrNilIfZeroValue(value)
	}
}

func withSubnetIds(values ...string) docDBModifier {
	return func(o *svcapitypes.DBSubnetGroup) {
		strArr := make([]*string, len(values))
		for i, val := range values {
			strArr[i] = pointer.ToOrNilIfZeroValue(val)
		}
		o.Spec.ForProvider.SubnetIDs = strArr
	}
}

func withSubnetIDStatus(values ...string) docDBModifier {
	return func(o *svcapitypes.DBSubnetGroup) {
		subnetArr := make([]*svcapitypes.Subnet, len(values))
		for i, val := range values {
			subnetArr[i] = &svcapitypes.Subnet{
				SubnetIdentifier: pointer.ToOrNilIfZeroValue(val),
			}
		}
		o.Status.AtProvider.Subnets = subnetArr
	}
}

func withConditions(value ...xpv1.Condition) docDBModifier {
	return func(o *svcapitypes.DBSubnetGroup) {
		o.Status.SetConditions(value...)
	}
}

func TestObserve(t *testing.T) {
	type want struct {
		cr     *svcapitypes.DBSubnetGroup
		result managed.ExternalObservation
		err    error
		docdb  fake.MockDocDBClientCall
	}

	cases := map[string]struct {
		args
		want
	}{
		"AvailableState_and_UpToDate": {
			args: args{
				docdb: &fake.MockDocDBClient{
					MockDescribeDBSubnetGroupsWithContext: func(c context.Context, ddgi *docdb.DescribeDBSubnetGroupsInput, o []request.Option) (*docdb.DescribeDBSubnetGroupsOutput, error) {
						return &docdb.DescribeDBSubnetGroupsOutput{
							DBSubnetGroups: []*docdb.DBSubnetGroup{
								{
									DBSubnetGroupName: pointer.ToOrNilIfZeroValue(testDBSubnetGroupName),
								},
							},
						}, nil
					},
					MockListTagsForResource: func(ltfri *docdb.ListTagsForResourceInput) (*docdb.ListTagsForResourceOutput, error) {
						return &docdb.ListTagsForResourceOutput{}, nil
					},
				},
				cr: subnetGroup(
					withDBSubnetGroupName(testDBSubnetGroupName),
					withExternalName(testDBSubnetGroupName),
				),
			},
			want: want{
				cr: subnetGroup(
					withDBSubnetGroupName(testDBSubnetGroupName),
					withExternalName(testDBSubnetGroupName),
					withConditions(xpv1.Available()),
				),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
				},
				docdb: fake.MockDocDBClientCall{
					DescribeDBSubnetGroupsWithContext: []*fake.CallDescribeDBSubnetGroupsWithContext{
						{
							Ctx: context.Background(),
							I: &docdb.DescribeDBSubnetGroupsInput{
								DBSubnetGroupName: pointer.ToOrNilIfZeroValue(testDBSubnetGroupName),
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
					MockDescribeDBSubnetGroupsWithContext: func(c context.Context, ddgi *docdb.DescribeDBSubnetGroupsInput, o []request.Option) (*docdb.DescribeDBSubnetGroupsOutput, error) {
						return &docdb.DescribeDBSubnetGroupsOutput{
							DBSubnetGroups: []*docdb.DBSubnetGroup{
								{
									DBSubnetGroupName:        pointer.ToOrNilIfZeroValue(testDBSubnetGroupName),
									DBSubnetGroupDescription: pointer.ToOrNilIfZeroValue(testDescription),
								},
							},
						}, nil
					},
				},
				cr: subnetGroup(
					withDBSubnetGroupName(testDBSubnetGroupName),
					withExternalName(testDBSubnetGroupName),
					withDescription(testOtherDescription),
				),
			},
			want: want{
				cr: subnetGroup(
					withDBSubnetGroupName(testDBSubnetGroupName),
					withExternalName(testDBSubnetGroupName),
					withDescription(testOtherDescription),
					withConditions(xpv1.Available()),
				),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: false,
				},
				docdb: fake.MockDocDBClientCall{
					DescribeDBSubnetGroupsWithContext: []*fake.CallDescribeDBSubnetGroupsWithContext{
						{
							Ctx: context.Background(),
							I: &docdb.DescribeDBSubnetGroupsInput{
								DBSubnetGroupName: pointer.ToOrNilIfZeroValue(testDBSubnetGroupName),
							},
						},
					},
				},
			},
		},
		"AvailableState_and_equal_subnets": {
			args: args{
				docdb: &fake.MockDocDBClient{
					MockDescribeDBSubnetGroupsWithContext: func(c context.Context, ddgi *docdb.DescribeDBSubnetGroupsInput, o []request.Option) (*docdb.DescribeDBSubnetGroupsOutput, error) {
						return &docdb.DescribeDBSubnetGroupsOutput{
							DBSubnetGroups: []*docdb.DBSubnetGroup{
								{
									DBSubnetGroupName: pointer.ToOrNilIfZeroValue(testDBSubnetGroupName),
									Subnets: []*docdb.Subnet{
										{SubnetIdentifier: pointer.ToOrNilIfZeroValue(testSubnetID)},
									},
								},
							},
						}, nil
					},
					MockListTagsForResource: func(ltfri *docdb.ListTagsForResourceInput) (*docdb.ListTagsForResourceOutput, error) {
						return &docdb.ListTagsForResourceOutput{}, nil
					},
				},
				cr: subnetGroup(
					withDBSubnetGroupName(testDBSubnetGroupName),
					withExternalName(testDBSubnetGroupName),
					withSubnetIds(testSubnetID),
				),
			},
			want: want{
				cr: subnetGroup(
					withDBSubnetGroupName(testDBSubnetGroupName),
					withExternalName(testDBSubnetGroupName),
					withSubnetIds(testSubnetID),
					withSubnetIDStatus(testSubnetID),
					withConditions(xpv1.Available()),
				),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
				},
				docdb: fake.MockDocDBClientCall{
					DescribeDBSubnetGroupsWithContext: []*fake.CallDescribeDBSubnetGroupsWithContext{
						{
							Ctx: context.Background(),
							I: &docdb.DescribeDBSubnetGroupsInput{
								DBSubnetGroupName: pointer.ToOrNilIfZeroValue(testDBSubnetGroupName),
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
		"AvailableState_and_different_subnets": {
			args: args{
				docdb: &fake.MockDocDBClient{
					MockDescribeDBSubnetGroupsWithContext: func(c context.Context, ddgi *docdb.DescribeDBSubnetGroupsInput, o []request.Option) (*docdb.DescribeDBSubnetGroupsOutput, error) {
						return &docdb.DescribeDBSubnetGroupsOutput{
							DBSubnetGroups: []*docdb.DBSubnetGroup{
								{
									DBSubnetGroupName: pointer.ToOrNilIfZeroValue(testDBSubnetGroupName),
									Subnets: []*docdb.Subnet{
										{SubnetIdentifier: pointer.ToOrNilIfZeroValue(testSubnetID)},
									},
								},
							},
						}, nil
					},
				},
				cr: subnetGroup(
					withDBSubnetGroupName(testDBSubnetGroupName),
					withExternalName(testDBSubnetGroupName),
					withSubnetIds(testOtherSubnetID),
				),
			},
			want: want{
				cr: subnetGroup(
					withDBSubnetGroupName(testDBSubnetGroupName),
					withExternalName(testDBSubnetGroupName),
					withSubnetIds(testOtherSubnetID),
					withSubnetIDStatus(testSubnetID),
					withConditions(xpv1.Available()),
				),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: false,
				},
				docdb: fake.MockDocDBClientCall{
					DescribeDBSubnetGroupsWithContext: []*fake.CallDescribeDBSubnetGroupsWithContext{
						{
							Ctx: context.Background(),
							I: &docdb.DescribeDBSubnetGroupsInput{
								DBSubnetGroupName: pointer.ToOrNilIfZeroValue(testDBSubnetGroupName),
							},
						},
					},
				},
			},
		},
		"ErrDescribeDBSubnetGroups": {
			args: args{
				docdb: &fake.MockDocDBClient{
					MockDescribeDBSubnetGroupsWithContext: func(c context.Context, ddgi *docdb.DescribeDBSubnetGroupsInput, o []request.Option) (*docdb.DescribeDBSubnetGroupsOutput, error) {
						return &docdb.DescribeDBSubnetGroupsOutput{}, errors.New(testErrDescribeDBSubnetGroupsFailed)
					},
				},
				cr: subnetGroup(
					withDBSubnetGroupName(testDBSubnetGroupName),
					withExternalName(testDBSubnetGroupName),
				),
			},
			want: want{
				cr: subnetGroup(
					withDBSubnetGroupName(testDBSubnetGroupName),
					withExternalName(testDBSubnetGroupName),
				),
				result: managed.ExternalObservation{},
				err:    errors.Wrap(errors.New(testErrDescribeDBSubnetGroupsFailed), errDescribe),
				docdb: fake.MockDocDBClientCall{
					DescribeDBSubnetGroupsWithContext: []*fake.CallDescribeDBSubnetGroupsWithContext{
						{
							Ctx: context.Background(),
							I: &docdb.DescribeDBSubnetGroupsInput{
								DBSubnetGroupName: pointer.ToOrNilIfZeroValue(testDBSubnetGroupName),
							},
						},
					},
				},
			},
		},
		"EmptyDescribeDBSubnetGroupsResponse": {
			args: args{
				docdb: &fake.MockDocDBClient{
					MockDescribeDBSubnetGroupsWithContext: func(c context.Context, ddgi *docdb.DescribeDBSubnetGroupsInput, o []request.Option) (*docdb.DescribeDBSubnetGroupsOutput, error) {
						return &docdb.DescribeDBSubnetGroupsOutput{
							DBSubnetGroups: []*docdb.DBSubnetGroup{},
						}, nil
					},
				},
				cr: subnetGroup(
					withDBSubnetGroupName(testDBSubnetGroupName),
					withExternalName(testDBSubnetGroupName),
				),
			},
			want: want{
				cr: subnetGroup(
					withDBSubnetGroupName(testDBSubnetGroupName),
					withExternalName(testDBSubnetGroupName),
				),
				result: managed.ExternalObservation{
					ResourceExists:   false,
					ResourceUpToDate: false,
				},
				docdb: fake.MockDocDBClientCall{
					DescribeDBSubnetGroupsWithContext: []*fake.CallDescribeDBSubnetGroupsWithContext{
						{
							Ctx: context.Background(),
							I: &docdb.DescribeDBSubnetGroupsInput{
								DBSubnetGroupName: pointer.ToOrNilIfZeroValue(testDBSubnetGroupName),
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
		cr     *svcapitypes.DBSubnetGroup
		result managed.ExternalCreation
		err    error
		docdb  fake.MockDocDBClientCall
	}

	cases := map[string]struct {
		args
		want
	}{
		"SuccessfulCreate": {
			args: args{
				docdb: &fake.MockDocDBClient{
					MockCreateDBSubnetGroupWithContext: func(c context.Context, cdgi *docdb.CreateDBSubnetGroupInput, o []request.Option) (*docdb.CreateDBSubnetGroupOutput, error) {
						return &docdb.CreateDBSubnetGroupOutput{
							DBSubnetGroup: &docdb.DBSubnetGroup{
								DBSubnetGroupName: pointer.ToOrNilIfZeroValue(testDBSubnetGroupName),
							},
						}, nil
					},
				},
				cr: subnetGroup(
					withDBSubnetGroupName(testDBSubnetGroupName),
					withExternalName(testDBSubnetGroupName),
				),
			},
			want: want{
				cr: subnetGroup(
					withDBSubnetGroupName(testDBSubnetGroupName),
					withExternalName(testDBSubnetGroupName),
					withConditions(xpv1.Creating()),
				),
				result: managed.ExternalCreation{},
				docdb: fake.MockDocDBClientCall{
					CreateDBSubnetGroupWithContext: []*fake.CallCreateDBSubnetGroupWithContext{
						{
							Ctx: context.Background(),
							I: &docdb.CreateDBSubnetGroupInput{
								DBSubnetGroupName: pointer.ToOrNilIfZeroValue(testDBSubnetGroupName),
							},
						},
					},
				},
			},
		},
		"ErrCreate": {
			args: args{
				docdb: &fake.MockDocDBClient{
					MockCreateDBSubnetGroupWithContext: func(c context.Context, cdgi *docdb.CreateDBSubnetGroupInput, o []request.Option) (*docdb.CreateDBSubnetGroupOutput, error) {
						return &docdb.CreateDBSubnetGroupOutput{}, errors.New(testErrCreateDBSubnetGroupFailed)
					},
				},
				cr: subnetGroup(
					withDBSubnetGroupName(testDBSubnetGroupName),
					withExternalName(testDBSubnetGroupName),
				),
			},
			want: want{
				cr: subnetGroup(
					withDBSubnetGroupName(testDBSubnetGroupName),
					withExternalName(testDBSubnetGroupName),
					withConditions(xpv1.Creating()),
				),
				result: managed.ExternalCreation{},
				err:    errors.Wrap(errors.New(testErrCreateDBSubnetGroupFailed), errCreate),
				docdb: fake.MockDocDBClientCall{
					CreateDBSubnetGroupWithContext: []*fake.CallCreateDBSubnetGroupWithContext{
						{
							Ctx: context.Background(),
							I: &docdb.CreateDBSubnetGroupInput{
								DBSubnetGroupName: pointer.ToOrNilIfZeroValue(testDBSubnetGroupName),
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
		cr    *svcapitypes.DBSubnetGroup
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
					MockDeleteDBSubnetGroupWithContext: func(c context.Context, ddgi *docdb.DeleteDBSubnetGroupInput, o []request.Option) (*docdb.DeleteDBSubnetGroupOutput, error) {
						return &docdb.DeleteDBSubnetGroupOutput{}, nil
					},
				},
				cr: subnetGroup(
					withDBSubnetGroupName(testDBSubnetGroupName),
					withExternalName(testDBSubnetGroupName),
				),
			},
			want: want{
				cr: subnetGroup(
					withDBSubnetGroupName(testDBSubnetGroupName),
					withExternalName(testDBSubnetGroupName),
					withConditions(xpv1.Deleting()),
				),
				docdb: fake.MockDocDBClientCall{
					DeleteDBSubnetGroupWithContext: []*fake.CallDeleteDBSubnetGroupWithContext{
						{
							Ctx: context.Background(),
							I: &docdb.DeleteDBSubnetGroupInput{
								DBSubnetGroupName: pointer.ToOrNilIfZeroValue(testDBSubnetGroupName),
							},
						},
					},
				},
			},
		},
		"ErrDelete": {
			args: args{
				docdb: &fake.MockDocDBClient{
					MockDeleteDBSubnetGroupWithContext: func(c context.Context, ddgi *docdb.DeleteDBSubnetGroupInput, o []request.Option) (*docdb.DeleteDBSubnetGroupOutput, error) {
						return &docdb.DeleteDBSubnetGroupOutput{}, errors.New(testErrDeleteDBSubnetGroupFailed)
					},
				},
				cr: subnetGroup(
					withDBSubnetGroupName(testDBSubnetGroupName),
					withExternalName(testDBSubnetGroupName),
				),
			},
			want: want{
				cr: subnetGroup(
					withDBSubnetGroupName(testDBSubnetGroupName),
					withExternalName(testDBSubnetGroupName),
					withConditions(xpv1.Deleting()),
				),
				err: errors.Wrap(errors.New(testErrDeleteDBSubnetGroupFailed), errDelete),
				docdb: fake.MockDocDBClientCall{
					DeleteDBSubnetGroupWithContext: []*fake.CallDeleteDBSubnetGroupWithContext{
						{
							Ctx: context.Background(),
							I: &docdb.DeleteDBSubnetGroupInput{
								DBSubnetGroupName: pointer.ToOrNilIfZeroValue(testDBSubnetGroupName),
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

func TestModify(t *testing.T) {
	type want struct {
		cr     *svcapitypes.DBSubnetGroup
		result managed.ExternalUpdate
		err    error
		docdb  fake.MockDocDBClientCall
	}

	cases := map[string]struct {
		args
		want
	}{
		"SuccessfulModify_description": {
			args: args{
				docdb: &fake.MockDocDBClient{
					MockModifyDBSubnetGroupWithContext: func(c context.Context, mdgi *docdb.ModifyDBSubnetGroupInput, o []request.Option) (*docdb.ModifyDBSubnetGroupOutput, error) {
						return &docdb.ModifyDBSubnetGroupOutput{
							DBSubnetGroup: &docdb.DBSubnetGroup{
								DBSubnetGroupName: pointer.ToOrNilIfZeroValue(testDBSubnetGroupName),
							},
						}, nil
					},
					MockListTagsForResource: func(ltfri *docdb.ListTagsForResourceInput) (*docdb.ListTagsForResourceOutput, error) {
						return &docdb.ListTagsForResourceOutput{}, nil
					},
				},
				cr: subnetGroup(
					withDBSubnetGroupName(testDBSubnetGroupName),
					withExternalName(testDBSubnetGroupName),
					withDescription(testDescription),
				),
			},
			want: want{
				cr: subnetGroup(
					withDBSubnetGroupName(testDBSubnetGroupName),
					withExternalName(testDBSubnetGroupName),
					withDescription(testDescription),
					withConditions(xpv1.Available()),
				),
				result: managed.ExternalUpdate{},
				docdb: fake.MockDocDBClientCall{
					ModifyDBSubnetGroupWithContext: []*fake.CallModifyDBSubnetGroupWithContext{
						{
							Ctx: context.Background(),
							I: &docdb.ModifyDBSubnetGroupInput{
								DBSubnetGroupName:        pointer.ToOrNilIfZeroValue(testDBSubnetGroupName),
								DBSubnetGroupDescription: pointer.ToOrNilIfZeroValue(testDescription),
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
		"SuccessfulModify_update_subnet": {
			args: args{
				docdb: &fake.MockDocDBClient{
					MockModifyDBSubnetGroupWithContext: func(c context.Context, mdgi *docdb.ModifyDBSubnetGroupInput, o []request.Option) (*docdb.ModifyDBSubnetGroupOutput, error) {
						return &docdb.ModifyDBSubnetGroupOutput{
							DBSubnetGroup: &docdb.DBSubnetGroup{
								DBSubnetGroupName: pointer.ToOrNilIfZeroValue(testDBSubnetGroupName),
								Subnets: []*docdb.Subnet{
									{SubnetIdentifier: pointer.ToOrNilIfZeroValue(testSubnetID)},
								},
							},
						}, nil
					},
					MockListTagsForResource: func(ltfri *docdb.ListTagsForResourceInput) (*docdb.ListTagsForResourceOutput, error) {
						return &docdb.ListTagsForResourceOutput{}, nil
					},
				},
				cr: subnetGroup(
					withDBSubnetGroupName(testDBSubnetGroupName),
					withExternalName(testDBSubnetGroupName),
					withSubnetIds(testSubnetID),
				),
			},
			want: want{
				cr: subnetGroup(
					withDBSubnetGroupName(testDBSubnetGroupName),
					withExternalName(testDBSubnetGroupName),
					withSubnetIds(testSubnetID),
					withConditions(xpv1.Available()),
				),
				result: managed.ExternalUpdate{},
				docdb: fake.MockDocDBClientCall{
					ModifyDBSubnetGroupWithContext: []*fake.CallModifyDBSubnetGroupWithContext{
						{
							Ctx: context.Background(),
							I: &docdb.ModifyDBSubnetGroupInput{
								DBSubnetGroupName: pointer.ToOrNilIfZeroValue(testDBSubnetGroupName),
								SubnetIds: []*string{
									pointer.ToOrNilIfZeroValue(testSubnetID),
								},
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
		"ErrModify": {
			args: args{
				docdb: &fake.MockDocDBClient{
					MockModifyDBSubnetGroupWithContext: func(c context.Context, mdgi *docdb.ModifyDBSubnetGroupInput, o []request.Option) (*docdb.ModifyDBSubnetGroupOutput, error) {
						return &docdb.ModifyDBSubnetGroupOutput{}, errors.New(testErrModifyDBSubnetGroupFailed)
					},
				},
				cr: subnetGroup(
					withDBSubnetGroupName(testDBSubnetGroupName),
					withExternalName(testDBSubnetGroupName),
				),
			},
			want: want{
				cr: subnetGroup(
					withDBSubnetGroupName(testDBSubnetGroupName),
					withExternalName(testDBSubnetGroupName),
				),
				result: managed.ExternalUpdate{},
				err:    errors.Wrap(errors.New(testErrModifyDBSubnetGroupFailed), errUpdate),
				docdb: fake.MockDocDBClientCall{
					ModifyDBSubnetGroupWithContext: []*fake.CallModifyDBSubnetGroupWithContext{
						{
							Ctx: context.Background(),
							I: &docdb.ModifyDBSubnetGroupInput{
								DBSubnetGroupName: pointer.ToOrNilIfZeroValue(testDBSubnetGroupName),
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
			res, err := e.Update(context.Background(), tc.args.cr)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.cr, tc.args.cr, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.result, res, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.docdb, tc.args.docdb.Called, cmpopts.IgnoreInterfaces(struct{ context.Context }{})); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}
