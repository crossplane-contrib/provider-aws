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

package parametergroup

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/dax"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/dax/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/dax/fake"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
)

const (
	testParameterGroupName  = "test-parameter-group"
	testDescription         = "test-description"
	testOtherDescription    = "some-other-description"
	testParameterName       = "test-parameter-name"
	testParameterValue      = "test-parameter-value"
	testOtherParameterName  = "test-other-parameter-name"
	testOtherParameterValue = "test-other-parameter-value"

	testErrCreateParameterGroupFailed = "CreateParameterGroup failed"
	testErrDeleteParameterGroupFailed = "DeleteParameterGroup failed"
	testErrUpdateParameterGroupFailed = "UpdateParameterGroup failed"
)

type args struct {
	dax  *fake.MockDaxClient
	kube client.Client
	cr   *svcapitypes.ParameterGroup
}

type daxModifier func(group *svcapitypes.ParameterGroup)

func setupExternal(e *external) {
	e.preObserve = preObserve
	e.postObserve = postObserve
	e.preCreate = preCreate
	e.preUpdate = preUpdate
	e.preDelete = preDelete
	c := &custom{client: e.client, kube: e.kube}
	e.isUpToDate = c.isUpToDate
}

func instance(m ...daxModifier) *svcapitypes.ParameterGroup {
	cr := &svcapitypes.ParameterGroup{}
	for _, f := range m {
		f(cr)
	}
	return cr
}

func withExternalName(value string) daxModifier {
	return func(o *svcapitypes.ParameterGroup) {
		meta.SetExternalName(o, value)

	}
}

func withName(value string) daxModifier {
	return func(o *svcapitypes.ParameterGroup) {
		o.Name = value
	}
}

func withStatusParameterGroupName(value string) daxModifier {
	return func(o *svcapitypes.ParameterGroup) {
		o.Status.AtProvider.ParameterGroupName = pointer.ToOrNilIfZeroValue(value)
	}
}

func withDescription(value string) daxModifier {
	return func(o *svcapitypes.ParameterGroup) {
		o.Spec.ForProvider.Description = pointer.ToOrNilIfZeroValue(value)
	}
}

func withParameters(k, v string) daxModifier {
	return func(o *svcapitypes.ParameterGroup) {
		o.Spec.ForProvider.ParameterNameValues = append(o.Spec.ForProvider.ParameterNameValues, &svcapitypes.ParameterNameValue{
			ParameterName:  pointer.ToOrNilIfZeroValue(k),
			ParameterValue: pointer.ToOrNilIfZeroValue(v),
		})
	}
}

func withConditions(value ...xpv1.Condition) daxModifier {
	return func(o *svcapitypes.ParameterGroup) {
		o.Status.SetConditions(value...)
	}
}

func TestObserve(t *testing.T) {
	type want struct {
		cr     *svcapitypes.ParameterGroup
		result managed.ExternalObservation
		err    error
		dax    fake.MockDaxClientCall
	}

	cases := map[string]struct {
		args
		want
	}{
		"AvailableStateAndUpToDate": {
			args: args{
				dax: &fake.MockDaxClient{
					MockDescribeParameterGroupsWithContext: func(c context.Context, dpgi *dax.DescribeParameterGroupsInput, o []request.Option) (*dax.DescribeParameterGroupsOutput, error) {
						return &dax.DescribeParameterGroupsOutput{ParameterGroups: []*dax.ParameterGroup{
							{
								ParameterGroupName: pointer.ToOrNilIfZeroValue(testParameterGroupName),
								Description:        pointer.ToOrNilIfZeroValue(testDescription),
							},
						}}, nil
					},
					MockDescribeParametersWithContext: func(ctx context.Context, input *dax.DescribeParametersInput, o []request.Option) (*dax.DescribeParametersOutput, error) {
						return &dax.DescribeParametersOutput{
							Parameters: []*dax.Parameter{{
								ParameterName:  pointer.ToOrNilIfZeroValue(testParameterName),
								ParameterValue: pointer.ToOrNilIfZeroValue(testParameterValue),
							}},
						}, nil
					},
				},
				cr: instance(
					withExternalName(testParameterGroupName),
					withDescription(testDescription),
					withParameters(testParameterName, testParameterValue),
				),
			},
			want: want{
				cr: instance(
					withExternalName(testParameterGroupName),
					withDescription(testDescription),
					withParameters(testParameterName, testParameterValue),
					withStatusParameterGroupName(testParameterGroupName),
					withConditions(xpv1.Available()),
				),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
				},
				dax: fake.MockDaxClientCall{
					DescribeParameterGroupsWithContext: []*fake.CallDescribeParameterGroupsWithContext{
						{
							Ctx: context.Background(),
							I: GenerateDescribeParameterGroupsInput(instance(
								withName(testParameterGroupName),
								withStatusParameterGroupName(testParameterGroupName))),
							Opts: nil,
						},
					},
					DescribeParametersWithContext: []*fake.CallDescribeParametersWithContext{
						{
							Ctx: context.Background(),
							I: &dax.DescribeParametersInput{
								MaxResults:         pointer.ToIntAsInt64(100),
								ParameterGroupName: pointer.ToOrNilIfZeroValue(testParameterGroupName),
							},
						},
					},
				},
			},
		},
		"AvailableStateAndOutdatedValue": {
			args: args{
				dax: &fake.MockDaxClient{
					MockDescribeParameterGroupsWithContext: func(c context.Context, dpgi *dax.DescribeParameterGroupsInput, o []request.Option) (*dax.DescribeParameterGroupsOutput, error) {
						return &dax.DescribeParameterGroupsOutput{ParameterGroups: []*dax.ParameterGroup{
							{
								ParameterGroupName: pointer.ToOrNilIfZeroValue(testParameterGroupName),
								Description:        pointer.ToOrNilIfZeroValue(testDescription),
							},
						}}, nil
					},
					MockDescribeParametersWithContext: func(ctx context.Context, input *dax.DescribeParametersInput, o []request.Option) (*dax.DescribeParametersOutput, error) {
						return &dax.DescribeParametersOutput{
							Parameters: []*dax.Parameter{{
								ParameterName:  pointer.ToOrNilIfZeroValue(testParameterName),
								ParameterValue: pointer.ToOrNilIfZeroValue(testParameterValue),
							}},
						}, nil
					},
				},
				cr: instance(
					withExternalName(testParameterGroupName),
					withDescription(testDescription),
					withParameters(testOtherParameterName, testOtherParameterValue),
				),
			},
			want: want{
				cr: instance(
					withExternalName(testParameterGroupName),
					withDescription(testDescription),
					withParameters(testOtherParameterName, testOtherParameterValue),
					withStatusParameterGroupName(testParameterGroupName),
					withConditions(xpv1.Available()),
				),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: false,
				},
				dax: fake.MockDaxClientCall{
					DescribeParameterGroupsWithContext: []*fake.CallDescribeParameterGroupsWithContext{
						{
							Ctx: context.Background(),
							I: GenerateDescribeParameterGroupsInput(instance(
								withName(testParameterGroupName),
								withStatusParameterGroupName(testParameterGroupName))),
							Opts: nil,
						},
					},
					DescribeParametersWithContext: []*fake.CallDescribeParametersWithContext{
						{
							Ctx: context.Background(),
							I: &dax.DescribeParametersInput{
								MaxResults:         pointer.ToIntAsInt64(100),
								ParameterGroupName: pointer.ToOrNilIfZeroValue(testParameterGroupName),
							},
						},
					},
				},
			},
		},
		"AvailableStateAndOutdatedDescription": {
			args: args{
				dax: &fake.MockDaxClient{
					MockDescribeParameterGroupsWithContext: func(c context.Context, dpgi *dax.DescribeParameterGroupsInput, o []request.Option) (*dax.DescribeParameterGroupsOutput, error) {
						return &dax.DescribeParameterGroupsOutput{ParameterGroups: []*dax.ParameterGroup{
							{
								ParameterGroupName: pointer.ToOrNilIfZeroValue(testParameterGroupName),
								Description:        pointer.ToOrNilIfZeroValue(testDescription),
							},
						}}, nil
					},
				},
				cr: instance(
					withExternalName(testParameterGroupName),
					withDescription(testOtherDescription),
				),
			},
			want: want{
				cr: instance(
					withExternalName(testParameterGroupName),
					withDescription(testOtherDescription),
					withStatusParameterGroupName(testParameterGroupName),
					withConditions(xpv1.Available()),
				),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: false,
				},
				dax: fake.MockDaxClientCall{
					DescribeParameterGroupsWithContext: []*fake.CallDescribeParameterGroupsWithContext{
						{
							Ctx: context.Background(),
							I: GenerateDescribeParameterGroupsInput(instance(
								withName(testParameterGroupName),
								withStatusParameterGroupName(testParameterGroupName))),
							Opts: nil,
						},
					},
				},
			},
		},
		"ErrDescribeParameterGroupsWithContext": {
			args: args{
				dax: &fake.MockDaxClient{
					MockDescribeParameterGroupsWithContext: func(c context.Context, dpgi *dax.DescribeParameterGroupsInput, o []request.Option) (*dax.DescribeParameterGroupsOutput, error) {
						return &dax.DescribeParameterGroupsOutput{}, errors.New(testErrUpdateParameterGroupFailed)
					},
				},
				cr: instance(
					withExternalName(testParameterGroupName),
				),
			},
			want: want{
				cr: instance(
					withExternalName(testParameterGroupName),
				),
				err: errors.Wrap(errors.New(testErrUpdateParameterGroupFailed), errDescribe),
				result: managed.ExternalObservation{
					ResourceExists:   false,
					ResourceUpToDate: false,
				},
				dax: fake.MockDaxClientCall{
					DescribeParameterGroupsWithContext: []*fake.CallDescribeParameterGroupsWithContext{
						{
							Ctx: context.Background(),
							I: GenerateDescribeParameterGroupsInput(instance(
								withName(testParameterGroupName),
								withStatusParameterGroupName(testParameterGroupName))),
							Opts: nil,
						},
					},
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			opts := []option{setupExternal}
			e := newExternal(tc.args.kube, tc.args.dax, opts)
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
			if diff := cmp.Diff(tc.want.dax, tc.args.dax.Called, cmpopts.IgnoreInterfaces(struct{ context.Context }{})); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestUpdate(t *testing.T) {
	type want struct {
		cr     *svcapitypes.ParameterGroup
		result managed.ExternalUpdate
		err    error
		dax    fake.MockDaxClientCall
	}

	cases := map[string]struct {
		args
		want
	}{
		"SuccessfulUpdateOneParameter": {
			args: args{
				dax: &fake.MockDaxClient{
					MockUpdateParameterGroupWithContext: func(c context.Context, upgi *dax.UpdateParameterGroupInput, o []request.Option) (*dax.UpdateParameterGroupOutput, error) {
						return &dax.UpdateParameterGroupOutput{
							ParameterGroup: &dax.ParameterGroup{
								ParameterGroupName: pointer.ToOrNilIfZeroValue(testParameterGroupName),
								Description:        pointer.ToOrNilIfZeroValue(testDescription),
							},
						}, nil
					},
				},
				cr: instance(
					withExternalName(testParameterGroupName),
					withDescription(testDescription),
					withParameters(testParameterName, testParameterValue),
					withStatusParameterGroupName(testParameterGroupName),
				),
			},
			want: want{
				cr: instance(
					withExternalName(testParameterGroupName),
					withDescription(testDescription),
					withParameters(testParameterName, testParameterValue),
					withStatusParameterGroupName(testParameterGroupName),
				),
				result: managed.ExternalUpdate{},
				dax: fake.MockDaxClientCall{
					UpdateParameterGroupsWithContext: []*fake.CallUpdateParameterGroupsWithContext{
						{
							Ctx: context.Background(),
							I: &dax.UpdateParameterGroupInput{
								ParameterGroupName: pointer.ToOrNilIfZeroValue(testParameterGroupName),
								ParameterNameValues: []*dax.ParameterNameValue{
									{
										ParameterName:  pointer.ToOrNilIfZeroValue(testParameterName),
										ParameterValue: pointer.ToOrNilIfZeroValue(testParameterValue),
									},
								},
							},
						},
					},
				},
			},
		},
		"SuccessfulUpdateSomeParameters": {
			args: args{
				dax: &fake.MockDaxClient{
					MockUpdateParameterGroupWithContext: func(c context.Context, upgi *dax.UpdateParameterGroupInput, o []request.Option) (*dax.UpdateParameterGroupOutput, error) {
						return &dax.UpdateParameterGroupOutput{
							ParameterGroup: &dax.ParameterGroup{
								Description:        pointer.ToOrNilIfZeroValue(testDescription),
								ParameterGroupName: pointer.ToOrNilIfZeroValue(testParameterGroupName),
							},
						}, nil
					},
				},
				cr: instance(
					withExternalName(testParameterGroupName),
					withDescription(testDescription),
					withParameters(testParameterName, testParameterValue),
					withParameters(testOtherParameterName, testOtherParameterValue),
					withStatusParameterGroupName(testParameterGroupName),
				),
			},
			want: want{
				cr: instance(
					withExternalName(testParameterGroupName),
					withDescription(testDescription),
					withParameters(testParameterName, testParameterValue),
					withParameters(testOtherParameterName, testOtherParameterValue),
					withStatusParameterGroupName(testParameterGroupName),
				),
				result: managed.ExternalUpdate{},
				dax: fake.MockDaxClientCall{
					UpdateParameterGroupsWithContext: []*fake.CallUpdateParameterGroupsWithContext{
						{
							Ctx: context.Background(),
							I: &dax.UpdateParameterGroupInput{
								ParameterGroupName: pointer.ToOrNilIfZeroValue(testParameterGroupName),
								ParameterNameValues: []*dax.ParameterNameValue{
									{
										ParameterName:  pointer.ToOrNilIfZeroValue(testParameterName),
										ParameterValue: pointer.ToOrNilIfZeroValue(testParameterValue),
									},
									{
										ParameterName:  pointer.ToOrNilIfZeroValue(testOtherParameterName),
										ParameterValue: pointer.ToOrNilIfZeroValue(testOtherParameterValue),
									},
								},
							},
						},
					},
				},
			},
		},
		"testErrUpdateParameterGroupFailed": {
			args: args{
				dax: &fake.MockDaxClient{
					MockUpdateParameterGroupWithContext: func(c context.Context, upgi *dax.UpdateParameterGroupInput, o []request.Option) (*dax.UpdateParameterGroupOutput, error) {
						return nil, errors.New(testErrUpdateParameterGroupFailed)
					},
				},
				cr: instance(
					withExternalName(testParameterGroupName),
					withParameters(testParameterName, testParameterValue),
					withStatusParameterGroupName(testParameterGroupName),
				),
			},
			want: want{
				cr: instance(
					withExternalName(testParameterGroupName),
					withParameters(testParameterName, testParameterValue),
					withStatusParameterGroupName(testParameterGroupName),
				),
				err:    errors.Wrap(errors.New(testErrUpdateParameterGroupFailed), errUpdate),
				result: managed.ExternalUpdate{},
				dax: fake.MockDaxClientCall{
					UpdateParameterGroupsWithContext: []*fake.CallUpdateParameterGroupsWithContext{
						{
							Ctx: context.Background(),
							I: &dax.UpdateParameterGroupInput{
								ParameterGroupName: pointer.ToOrNilIfZeroValue(testParameterGroupName),
								ParameterNameValues: []*dax.ParameterNameValue{
									{
										ParameterName:  pointer.ToOrNilIfZeroValue(testParameterName),
										ParameterValue: pointer.ToOrNilIfZeroValue(testParameterValue),
									},
								},
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
			e := newExternal(tc.args.kube, tc.args.dax, opts)
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
			if diff := cmp.Diff(tc.want.dax, tc.args.dax.Called, cmpopts.IgnoreInterfaces(struct{ context.Context }{})); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestCreate(t *testing.T) {
	type want struct {
		cr     *svcapitypes.ParameterGroup
		result managed.ExternalCreation
		err    error
		dax    fake.MockDaxClientCall
	}

	cases := map[string]struct {
		args
		want
	}{
		"SuccessfulCreateNoParameters": {
			args: args{
				dax: &fake.MockDaxClient{
					MockCreateParameterGroupWithContext: func(c context.Context, cpgi *dax.CreateParameterGroupInput, o []request.Option) (*dax.CreateParameterGroupOutput, error) {
						return &dax.CreateParameterGroupOutput{
							ParameterGroup: &dax.ParameterGroup{
								ParameterGroupName: pointer.ToOrNilIfZeroValue(testParameterGroupName),
							},
						}, nil
					},
				},
				cr: instance(
					withName(testParameterGroupName),
					withStatusParameterGroupName(testParameterGroupName),
				),
			},
			want: want{
				cr: instance(
					withName(testParameterGroupName),
					withStatusParameterGroupName(testParameterGroupName),
					withExternalName(testParameterGroupName),
					withConditions(xpv1.Creating()),
				),
				result: managed.ExternalCreation{},
				dax: fake.MockDaxClientCall{
					CreateParameterGroupWithContext: []*fake.CallCreateParameterGroupWithContext{
						{
							Ctx: context.Background(),
							I: &dax.CreateParameterGroupInput{
								ParameterGroupName: pointer.ToOrNilIfZeroValue(testParameterGroupName),
							},
						},
					},
				},
			},
		},
		"SuccessfulCreateWithParameters": {
			args: args{
				dax: &fake.MockDaxClient{
					MockCreateParameterGroupWithContext: func(c context.Context, cpgi *dax.CreateParameterGroupInput, o []request.Option) (*dax.CreateParameterGroupOutput, error) {
						return &dax.CreateParameterGroupOutput{
							ParameterGroup: &dax.ParameterGroup{
								ParameterGroupName: pointer.ToOrNilIfZeroValue(testParameterGroupName),
							},
						}, nil
					},
				},
				cr: instance(
					withName(testParameterGroupName),
					withParameters(testParameterName, testParameterValue),
					withParameters(testOtherParameterName, testOtherParameterValue),
				),
			},
			want: want{
				cr: instance(
					withName(testParameterGroupName),
					withParameters(testParameterName, testParameterValue),
					withParameters(testOtherParameterName, testOtherParameterValue),
					withStatusParameterGroupName(testParameterGroupName),
					withExternalName(testParameterGroupName),
					withConditions(xpv1.Creating()),
				),
				result: managed.ExternalCreation{},
				dax: fake.MockDaxClientCall{
					CreateParameterGroupWithContext: []*fake.CallCreateParameterGroupWithContext{
						{
							Ctx: context.Background(),
							I: &dax.CreateParameterGroupInput{
								ParameterGroupName: pointer.ToOrNilIfZeroValue(testParameterGroupName),
							},
						},
					},
				},
			},
		},
		"ErrorCreate": {
			args: args{
				dax: &fake.MockDaxClient{
					MockCreateParameterGroupWithContext: func(c context.Context, cpgi *dax.CreateParameterGroupInput, o []request.Option) (*dax.CreateParameterGroupOutput, error) {
						return &dax.CreateParameterGroupOutput{}, errors.New(testErrCreateParameterGroupFailed)
					},
				},
				cr: instance(
					withName(testParameterGroupName),
				),
			},
			want: want{
				cr: instance(
					withName(testParameterGroupName),
					withExternalName(testParameterGroupName),
					withConditions(xpv1.Creating()),
				),
				result: managed.ExternalCreation{},
				err:    errors.Wrap(errors.New(testErrCreateParameterGroupFailed), errCreate),
				dax: fake.MockDaxClientCall{
					CreateParameterGroupWithContext: []*fake.CallCreateParameterGroupWithContext{
						{
							Ctx: context.Background(),
							I: &dax.CreateParameterGroupInput{
								ParameterGroupName: pointer.ToOrNilIfZeroValue(testParameterGroupName),
							},
						},
					},
				},
			},
		},
		"ErrorCreateParameters": {
			args: args{
				dax: &fake.MockDaxClient{
					MockCreateParameterGroupWithContext: func(c context.Context, cpgi *dax.CreateParameterGroupInput, o []request.Option) (*dax.CreateParameterGroupOutput, error) {
						return &dax.CreateParameterGroupOutput{}, errors.New(testErrCreateParameterGroupFailed)
					},
				},
				cr: instance(
					withName(testParameterGroupName),
					withParameters(testParameterName, testParameterValue),
				),
			},
			want: want{
				cr: instance(
					withName(testParameterGroupName),
					withParameters(testParameterName, testParameterValue),
					withExternalName(testParameterGroupName),
					withConditions(xpv1.Creating()),
				),
				result: managed.ExternalCreation{},
				err:    errors.Wrap(errors.New(testErrCreateParameterGroupFailed), errCreate),
				dax: fake.MockDaxClientCall{
					CreateParameterGroupWithContext: []*fake.CallCreateParameterGroupWithContext{
						{
							Ctx: context.Background(),
							I: &dax.CreateParameterGroupInput{
								ParameterGroupName: pointer.ToOrNilIfZeroValue(testParameterGroupName),
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
			e := newExternal(tc.args.kube, tc.args.dax, opts)
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
			if diff := cmp.Diff(tc.want.dax, tc.args.dax.Called, cmpopts.IgnoreInterfaces(struct{ context.Context }{})); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	type want struct {
		cr  *svcapitypes.ParameterGroup
		err error
		dax fake.MockDaxClientCall
	}

	cases := map[string]struct {
		args
		want
	}{
		"SuccessfulDelete": {
			args: args{
				dax: &fake.MockDaxClient{
					MockDeleteParameterGroupWithContext: func(c context.Context, dpgi *dax.DeleteParameterGroupInput, o []request.Option) (*dax.DeleteParameterGroupOutput, error) {
						return &dax.DeleteParameterGroupOutput{}, nil
					},
				},
				cr: instance(
					withExternalName(testParameterGroupName),
				),
			},
			want: want{
				cr: instance(
					withExternalName(testParameterGroupName),
					withConditions(xpv1.Deleting()),
				),
				dax: fake.MockDaxClientCall{
					DeleteParameterGroupWithContext: []*fake.CallDeleteParameterGroupWithContext{
						{
							Ctx: context.Background(),
							I: &dax.DeleteParameterGroupInput{
								ParameterGroupName: pointer.ToOrNilIfZeroValue(testParameterGroupName),
							},
						},
					},
				},
			},
		},
		"ErrorDelete": {
			args: args{
				dax: &fake.MockDaxClient{
					MockDeleteParameterGroupWithContext: func(c context.Context, dpgi *dax.DeleteParameterGroupInput, o []request.Option) (*dax.DeleteParameterGroupOutput, error) {
						return &dax.DeleteParameterGroupOutput{}, errors.New(testErrDeleteParameterGroupFailed)
					},
				},
				cr: instance(
					withExternalName(testParameterGroupName),
				),
			},
			want: want{
				cr: instance(
					withExternalName(testParameterGroupName),
					withConditions(xpv1.Deleting()),
				),
				err: errors.Wrap(errors.New(testErrDeleteParameterGroupFailed), errDelete),
				dax: fake.MockDaxClientCall{
					DeleteParameterGroupWithContext: []*fake.CallDeleteParameterGroupWithContext{
						{
							Ctx: context.Background(),
							I: &dax.DeleteParameterGroupInput{
								ParameterGroupName: pointer.ToOrNilIfZeroValue(testParameterGroupName),
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
			e := newExternal(tc.args.kube, tc.args.dax, opts)
			err := e.Delete(context.Background(), tc.args.cr)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.cr, tc.args.cr, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.dax, tc.args.dax.Called, cmpopts.IgnoreInterfaces(struct{ context.Context }{})); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}
