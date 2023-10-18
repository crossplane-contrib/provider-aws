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

package subnetgroup

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
	testSubnetGroupName  = "test-subnet-group"
	testDescription      = "test-description"
	testOtherDescription = "some-other-description"

	testSubnetIdentifier            = "test-subnet-identifier"
	testSubnetAvailabilityZone      = "us-east-1a"
	testOtherSubnetIdentifier       = "test-other-subnet-identifier"
	testOtherSubnetAvailabilityZone = "us-east-1b"

	testErrCreateSubnetGroupFailed = "CreateSubnetGroup failed"
	testErrDeleteSubnetGroupFailed = "DeleteSubnetGroup failed"
	testErrUpdateSubnetGroupFailed = "UpdateSubnetGroup failed"
)

type args struct {
	dax  *fake.MockDaxClient
	kube client.Client
	cr   *svcapitypes.SubnetGroup
}

type daxModifier func(group *svcapitypes.SubnetGroup)

func setupExternal(e *external) {
	e.preObserve = preObserve
	e.postObserve = postObserve
	e.preCreate = preCreate
	e.preUpdate = preUpdate
	e.preDelete = preDelete
	e.isUpToDate = isUpToDate
}

func instance(m ...daxModifier) *svcapitypes.SubnetGroup {
	cr := &svcapitypes.SubnetGroup{}
	for _, f := range m {
		f(cr)
	}
	return cr
}

func withExternalName(value string) daxModifier {
	return func(o *svcapitypes.SubnetGroup) {
		meta.SetExternalName(o, value)

	}
}

func withName(value string) daxModifier {
	return func(o *svcapitypes.SubnetGroup) {
		o.Name = value
	}
}

func withDescription(value string) daxModifier {
	return func(o *svcapitypes.SubnetGroup) {
		o.Spec.ForProvider.Description = pointer.ToOrNilIfZeroValue(value)
	}
}

func withSubnetID(value string) daxModifier {
	return func(o *svcapitypes.SubnetGroup) {
		o.Spec.ForProvider.SubnetIds = append(o.Spec.ForProvider.SubnetIds, pointer.ToOrNilIfZeroValue(value))
	}
}

func withStatusSubnetGroupName(value string) daxModifier {
	return func(o *svcapitypes.SubnetGroup) {
		o.Status.AtProvider.SubnetGroupName = pointer.ToOrNilIfZeroValue(value)
	}
}

func withStatusSubnets(k, v string) daxModifier {
	return func(o *svcapitypes.SubnetGroup) {
		o.Status.AtProvider.Subnets = append(o.Status.AtProvider.Subnets, &svcapitypes.Subnet{
			SubnetIdentifier:       pointer.ToOrNilIfZeroValue(k),
			SubnetAvailabilityZone: pointer.ToOrNilIfZeroValue(v),
		})
	}
}

func withConditions(value ...xpv1.Condition) daxModifier {
	return func(o *svcapitypes.SubnetGroup) {
		o.Status.SetConditions(value...)
	}
}

func TestObserve(t *testing.T) {
	type want struct {
		cr     *svcapitypes.SubnetGroup
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
					MockDescribeSubnetGroupsWithContext: func(c context.Context, dsgi *dax.DescribeSubnetGroupsInput, o []request.Option) (*dax.DescribeSubnetGroupsOutput, error) {
						return &dax.DescribeSubnetGroupsOutput{SubnetGroups: []*dax.SubnetGroup{
							{
								SubnetGroupName: pointer.ToOrNilIfZeroValue(testSubnetGroupName),
								Description:     pointer.ToOrNilIfZeroValue(testDescription),
								Subnets: []*dax.Subnet{{
									SubnetIdentifier:       pointer.ToOrNilIfZeroValue(testSubnetIdentifier),
									SubnetAvailabilityZone: pointer.ToOrNilIfZeroValue(testSubnetAvailabilityZone)}},
							},
						}}, nil
					},
				},
				cr: instance(
					withExternalName(testSubnetGroupName),
					withDescription(testDescription),
					withSubnetID(testSubnetIdentifier),
				),
			},
			want: want{
				cr: instance(
					withExternalName(testSubnetGroupName),
					withDescription(testDescription),
					withSubnetID(testSubnetIdentifier),
					withStatusSubnetGroupName(testSubnetGroupName),
					withStatusSubnets(testSubnetIdentifier, testSubnetAvailabilityZone),
					withConditions(xpv1.Available()),
				),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
				},
				dax: fake.MockDaxClientCall{
					DescribeSubnetGroupsWithContext: []*fake.CallDescribeSubnetGroupsWithContext{
						{
							Ctx: context.Background(),
							I: GenerateDescribeSubnetGroupsInput(instance(
								withName(testSubnetGroupName),
								withStatusSubnetGroupName(testSubnetGroupName))),
							Opts: nil,
						},
					},
				},
			},
		},
		"AvailableStateAndOutdatedDescription": {
			args: args{
				dax: &fake.MockDaxClient{
					MockDescribeSubnetGroupsWithContext: func(c context.Context, dsgi *dax.DescribeSubnetGroupsInput, o []request.Option) (*dax.DescribeSubnetGroupsOutput, error) {
						return &dax.DescribeSubnetGroupsOutput{SubnetGroups: []*dax.SubnetGroup{
							{
								SubnetGroupName: pointer.ToOrNilIfZeroValue(testSubnetGroupName),
								Description:     pointer.ToOrNilIfZeroValue(testDescription),
								Subnets: []*dax.Subnet{{
									SubnetIdentifier:       pointer.ToOrNilIfZeroValue(testSubnetIdentifier),
									SubnetAvailabilityZone: pointer.ToOrNilIfZeroValue(testSubnetAvailabilityZone)}},
							},
						}}, nil
					},
				},
				cr: instance(
					withExternalName(testSubnetGroupName),
					withDescription(testOtherDescription),
					withSubnetID(testSubnetIdentifier),
				),
			},
			want: want{
				cr: instance(
					withDescription(testOtherDescription),
					withSubnetID(testSubnetIdentifier),
					withStatusSubnetGroupName(testSubnetGroupName),
					withStatusSubnets(testSubnetIdentifier, testSubnetAvailabilityZone),
					withExternalName(testSubnetGroupName),
					withConditions(xpv1.Available()),
				),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: false,
				},
				dax: fake.MockDaxClientCall{
					DescribeSubnetGroupsWithContext: []*fake.CallDescribeSubnetGroupsWithContext{
						{
							Ctx: context.Background(),
							I: GenerateDescribeSubnetGroupsInput(instance(
								withName(testSubnetGroupName),
								withStatusSubnetGroupName(testSubnetGroupName))),
							Opts: nil,
						},
					},
				},
			},
		},
		"AvailableStateAndOutdatedSubnet": {
			args: args{
				dax: &fake.MockDaxClient{
					MockDescribeSubnetGroupsWithContext: func(c context.Context, dsgi *dax.DescribeSubnetGroupsInput, o []request.Option) (*dax.DescribeSubnetGroupsOutput, error) {
						return &dax.DescribeSubnetGroupsOutput{SubnetGroups: []*dax.SubnetGroup{
							{
								SubnetGroupName: pointer.ToOrNilIfZeroValue(testSubnetGroupName),
								Description:     pointer.ToOrNilIfZeroValue(testDescription),
								Subnets: []*dax.Subnet{{
									SubnetIdentifier:       pointer.ToOrNilIfZeroValue(testSubnetIdentifier),
									SubnetAvailabilityZone: pointer.ToOrNilIfZeroValue(testSubnetAvailabilityZone)}},
							},
						}}, nil
					},
				},
				cr: instance(
					withExternalName(testSubnetGroupName),
					withDescription(testDescription),
					withSubnetID(testOtherSubnetIdentifier),
				),
			},
			want: want{
				cr: instance(
					withExternalName(testSubnetGroupName),
					withDescription(testDescription),
					withSubnetID(testOtherSubnetIdentifier),
					withStatusSubnetGroupName(testSubnetGroupName),
					withStatusSubnets(testSubnetIdentifier, testSubnetAvailabilityZone),
					withConditions(xpv1.Available()),
				),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: false,
				},
				dax: fake.MockDaxClientCall{
					DescribeSubnetGroupsWithContext: []*fake.CallDescribeSubnetGroupsWithContext{
						{
							Ctx: context.Background(),
							I: GenerateDescribeSubnetGroupsInput(instance(
								withName(testSubnetGroupName),
								withStatusSubnetGroupName(testSubnetGroupName))),
							Opts: nil,
						},
					},
				},
			},
		},
		"ErrDescribeSubnetGroupsWithContext": {
			args: args{
				dax: &fake.MockDaxClient{
					MockDescribeSubnetGroupsWithContext: func(c context.Context, dsgi *dax.DescribeSubnetGroupsInput, o []request.Option) (*dax.DescribeSubnetGroupsOutput, error) {
						return &dax.DescribeSubnetGroupsOutput{}, errors.New(testErrUpdateSubnetGroupFailed)
					},
				},
				cr: instance(
					withExternalName(testSubnetGroupName),
				),
			},
			want: want{
				cr: instance(
					withExternalName(testSubnetGroupName),
				),
				err: errors.Wrap(errors.New(testErrUpdateSubnetGroupFailed), errDescribe),
				result: managed.ExternalObservation{
					ResourceExists:   false,
					ResourceUpToDate: false,
				},
				dax: fake.MockDaxClientCall{
					DescribeSubnetGroupsWithContext: []*fake.CallDescribeSubnetGroupsWithContext{
						{
							Ctx: context.Background(),
							I: GenerateDescribeSubnetGroupsInput(instance(
								withName(testSubnetGroupName),
								withStatusSubnetGroupName(testSubnetGroupName))),
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
		cr     *svcapitypes.SubnetGroup
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
					MockUpdateSubnetGroupWithContext: func(c context.Context, usgi *dax.UpdateSubnetGroupInput, o []request.Option) (*dax.UpdateSubnetGroupOutput, error) {
						return &dax.UpdateSubnetGroupOutput{
							SubnetGroup: &dax.SubnetGroup{
								SubnetGroupName: pointer.ToOrNilIfZeroValue(testSubnetGroupName),
								Subnets:         []*dax.Subnet{{SubnetIdentifier: pointer.ToOrNilIfZeroValue(testSubnetIdentifier)}},
							},
						}, nil
					},
				},
				cr: instance(
					withExternalName(testSubnetGroupName),
					withSubnetID(testSubnetIdentifier),
				),
			},
			want: want{
				cr: instance(
					withExternalName(testSubnetGroupName),
					withSubnetID(testSubnetIdentifier),
				),
				result: managed.ExternalUpdate{},
				dax: fake.MockDaxClientCall{
					UpdateSubnetGroupsWithContext: []*fake.CallUpdateSubnetGroupsWithContext{
						{
							Ctx: context.Background(),
							I: &dax.UpdateSubnetGroupInput{
								SubnetGroupName: pointer.ToOrNilIfZeroValue(testSubnetGroupName),
								SubnetIds:       []*string{pointer.ToOrNilIfZeroValue(testSubnetIdentifier)},
							},
						},
					},
				},
			},
		},
		"SuccessfulUpdateSomeParameters": {
			args: args{
				dax: &fake.MockDaxClient{
					MockUpdateSubnetGroupWithContext: func(c context.Context, usgi *dax.UpdateSubnetGroupInput, o []request.Option) (*dax.UpdateSubnetGroupOutput, error) {
						return &dax.UpdateSubnetGroupOutput{
							SubnetGroup: &dax.SubnetGroup{
								SubnetGroupName: pointer.ToOrNilIfZeroValue(testSubnetGroupName),
								Description:     pointer.ToOrNilIfZeroValue(testDescription),
								Subnets: []*dax.Subnet{
									{
										SubnetIdentifier:       pointer.ToOrNilIfZeroValue(testSubnetIdentifier),
										SubnetAvailabilityZone: pointer.ToOrNilIfZeroValue(testSubnetAvailabilityZone),
									},
									{
										SubnetIdentifier:       pointer.ToOrNilIfZeroValue(testOtherSubnetIdentifier),
										SubnetAvailabilityZone: pointer.ToOrNilIfZeroValue(testOtherSubnetAvailabilityZone),
									},
								},
							},
						}, nil
					},
				},
				cr: instance(
					withExternalName(testSubnetGroupName),
					withDescription(testDescription),
					withSubnetID(testSubnetIdentifier),
					withSubnetID(testOtherSubnetIdentifier),
				),
			},
			want: want{
				cr: instance(
					withExternalName(testSubnetGroupName),
					withDescription(testDescription),
					withSubnetID(testSubnetIdentifier),
					withSubnetID(testOtherSubnetIdentifier),
				),
				result: managed.ExternalUpdate{},
				dax: fake.MockDaxClientCall{
					UpdateSubnetGroupsWithContext: []*fake.CallUpdateSubnetGroupsWithContext{
						{
							Ctx: context.Background(),
							I: &dax.UpdateSubnetGroupInput{
								SubnetGroupName: pointer.ToOrNilIfZeroValue(testSubnetGroupName),
								Description:     pointer.ToOrNilIfZeroValue(testDescription),
								SubnetIds: []*string{
									pointer.ToOrNilIfZeroValue(testSubnetIdentifier),
									pointer.ToOrNilIfZeroValue(testOtherSubnetIdentifier),
								},
							},
						},
					},
				},
			},
		},
		"testErrUpdateSubnetGroupFailed": {
			args: args{
				dax: &fake.MockDaxClient{
					MockUpdateSubnetGroupWithContext: func(c context.Context, usgi *dax.UpdateSubnetGroupInput, o []request.Option) (*dax.UpdateSubnetGroupOutput, error) {
						return nil, errors.New(testErrUpdateSubnetGroupFailed)
					},
				},
				cr: instance(
					withExternalName(testSubnetGroupName),
					withStatusSubnetGroupName(testSubnetGroupName),
				),
			},
			want: want{
				cr: instance(
					withExternalName(testSubnetGroupName),
					withStatusSubnetGroupName(testSubnetGroupName),
				),
				err:    errors.Wrap(errors.New(testErrUpdateSubnetGroupFailed), errUpdate),
				result: managed.ExternalUpdate{},
				dax: fake.MockDaxClientCall{
					UpdateSubnetGroupsWithContext: []*fake.CallUpdateSubnetGroupsWithContext{
						{
							Ctx: context.Background(),
							I: &dax.UpdateSubnetGroupInput{
								SubnetGroupName: pointer.ToOrNilIfZeroValue(testSubnetGroupName),
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
		cr     *svcapitypes.SubnetGroup
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
					MockCreateSubnetGroupWithContext: func(c context.Context, csgi *dax.CreateSubnetGroupInput, o []request.Option) (*dax.CreateSubnetGroupOutput, error) {
						return &dax.CreateSubnetGroupOutput{
							SubnetGroup: &dax.SubnetGroup{
								SubnetGroupName: pointer.ToOrNilIfZeroValue(testSubnetGroupName),
							},
						}, nil
					},
				},
				cr: instance(
					withName(testSubnetGroupName),
					withStatusSubnetGroupName(testSubnetGroupName),
				),
			},
			want: want{
				cr: instance(
					withName(testSubnetGroupName),
					withStatusSubnetGroupName(testSubnetGroupName),
					withExternalName(testSubnetGroupName),
					withConditions(xpv1.Creating()),
				),
				result: managed.ExternalCreation{},
				dax: fake.MockDaxClientCall{
					CreateSubnetGroupWithContext: []*fake.CallCreateSubnetGroupWithContext{
						{
							Ctx: context.Background(),
							I: &dax.CreateSubnetGroupInput{
								SubnetGroupName: pointer.ToOrNilIfZeroValue(testSubnetGroupName),
							},
						},
					},
				},
			},
		},
		"SuccessfulCreateWithParameters": {
			args: args{
				dax: &fake.MockDaxClient{
					MockCreateSubnetGroupWithContext: func(c context.Context, csgi *dax.CreateSubnetGroupInput, o []request.Option) (*dax.CreateSubnetGroupOutput, error) {
						return &dax.CreateSubnetGroupOutput{
							SubnetGroup: &dax.SubnetGroup{
								SubnetGroupName: pointer.ToOrNilIfZeroValue(testSubnetGroupName),
								Subnets: []*dax.Subnet{
									{
										SubnetIdentifier:       pointer.ToOrNilIfZeroValue(testSubnetIdentifier),
										SubnetAvailabilityZone: pointer.ToOrNilIfZeroValue(testSubnetAvailabilityZone),
									},
									{
										SubnetIdentifier:       pointer.ToOrNilIfZeroValue(testOtherSubnetIdentifier),
										SubnetAvailabilityZone: pointer.ToOrNilIfZeroValue(testOtherSubnetAvailabilityZone),
									},
								},
							},
						}, nil
					},
				},
				cr: instance(
					withName(testSubnetGroupName),
					withSubnetID(testSubnetIdentifier),
					withSubnetID(testOtherSubnetIdentifier),
				),
			},
			want: want{
				cr: instance(
					withName(testSubnetGroupName),
					withSubnetID(testSubnetIdentifier),
					withSubnetID(testOtherSubnetIdentifier),
					withStatusSubnetGroupName(testSubnetGroupName),
					withStatusSubnets(testSubnetIdentifier, testSubnetAvailabilityZone),
					withStatusSubnets(testOtherSubnetIdentifier, testOtherSubnetAvailabilityZone),
					withExternalName(testSubnetGroupName),
					withConditions(xpv1.Creating()),
				),
				result: managed.ExternalCreation{},
				dax: fake.MockDaxClientCall{
					CreateSubnetGroupWithContext: []*fake.CallCreateSubnetGroupWithContext{
						{
							Ctx: context.Background(),
							I: &dax.CreateSubnetGroupInput{
								SubnetGroupName: pointer.ToOrNilIfZeroValue(testSubnetGroupName),
								SubnetIds:       []*string{pointer.ToOrNilIfZeroValue(testSubnetIdentifier), pointer.ToOrNilIfZeroValue(testOtherSubnetIdentifier)},
							},
						},
					},
				},
			},
		},
		"ErrorCreate": {
			args: args{
				dax: &fake.MockDaxClient{
					MockCreateSubnetGroupWithContext: func(c context.Context, csgi *dax.CreateSubnetGroupInput, o []request.Option) (*dax.CreateSubnetGroupOutput, error) {
						return &dax.CreateSubnetGroupOutput{}, errors.New(testErrCreateSubnetGroupFailed)
					},
				},
				cr: instance(
					withName(testSubnetGroupName),
				),
			},
			want: want{
				cr: instance(
					withName(testSubnetGroupName),
					withExternalName(testSubnetGroupName),
					withConditions(xpv1.Creating()),
				),
				result: managed.ExternalCreation{},
				err:    errors.Wrap(errors.New(testErrCreateSubnetGroupFailed), errCreate),
				dax: fake.MockDaxClientCall{
					CreateSubnetGroupWithContext: []*fake.CallCreateSubnetGroupWithContext{
						{
							Ctx: context.Background(),
							I: &dax.CreateSubnetGroupInput{
								SubnetGroupName: pointer.ToOrNilIfZeroValue(testSubnetGroupName),
							},
						},
					},
				},
			},
		},
		"ErrorCreateSubnetIds": {
			args: args{
				dax: &fake.MockDaxClient{
					MockCreateSubnetGroupWithContext: func(c context.Context, csgi *dax.CreateSubnetGroupInput, o []request.Option) (*dax.CreateSubnetGroupOutput, error) {
						return &dax.CreateSubnetGroupOutput{}, errors.New(testErrCreateSubnetGroupFailed)
					},
				},
				cr: instance(
					withName(testSubnetGroupName),
					withSubnetID(testSubnetIdentifier),
					withSubnetID(testOtherSubnetIdentifier),
				),
			},
			want: want{
				cr: instance(
					withName(testSubnetGroupName),
					withSubnetID(testSubnetIdentifier),
					withSubnetID(testOtherSubnetIdentifier),
					withExternalName(testSubnetGroupName),
					withConditions(xpv1.Creating()),
				),
				result: managed.ExternalCreation{},
				err:    errors.Wrap(errors.New(testErrCreateSubnetGroupFailed), errCreate),
				dax: fake.MockDaxClientCall{
					CreateSubnetGroupWithContext: []*fake.CallCreateSubnetGroupWithContext{
						{
							Ctx: context.Background(),
							I: &dax.CreateSubnetGroupInput{
								SubnetGroupName: pointer.ToOrNilIfZeroValue(testSubnetGroupName),
								SubnetIds:       []*string{pointer.ToOrNilIfZeroValue(testSubnetIdentifier), pointer.ToOrNilIfZeroValue(testOtherSubnetIdentifier)},
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
		cr  *svcapitypes.SubnetGroup
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
					MockDeleteSubnetGroupWithContext: func(c context.Context, dsgi *dax.DeleteSubnetGroupInput, o []request.Option) (*dax.DeleteSubnetGroupOutput, error) {
						return &dax.DeleteSubnetGroupOutput{}, nil
					},
				},
				cr: instance(
					withExternalName(testSubnetGroupName),
				),
			},
			want: want{
				cr: instance(
					withExternalName(testSubnetGroupName),
					withConditions(xpv1.Deleting()),
				),
				dax: fake.MockDaxClientCall{
					DeleteSubnetGroupWithContext: []*fake.CallDeleteSubnetGroupWithContext{
						{
							Ctx: context.Background(),
							I: &dax.DeleteSubnetGroupInput{
								SubnetGroupName: pointer.ToOrNilIfZeroValue(testSubnetGroupName),
							},
						},
					},
				},
			},
		},
		"ErrorDelete": {
			args: args{
				dax: &fake.MockDaxClient{
					MockDeleteSubnetGroupWithContext: func(c context.Context, dsgi *dax.DeleteSubnetGroupInput, o []request.Option) (*dax.DeleteSubnetGroupOutput, error) {
						return &dax.DeleteSubnetGroupOutput{}, errors.New(testErrDeleteSubnetGroupFailed)
					},
				},
				cr: instance(
					withExternalName(testSubnetGroupName),
				),
			},
			want: want{
				cr: instance(
					withExternalName(testSubnetGroupName),
					withConditions(xpv1.Deleting()),
				),
				err: errors.Wrap(errors.New(testErrDeleteSubnetGroupFailed), errDelete),
				dax: fake.MockDaxClientCall{
					DeleteSubnetGroupWithContext: []*fake.CallDeleteSubnetGroupWithContext{
						{
							Ctx: context.Background(),
							I: &dax.DeleteSubnetGroupInput{
								SubnetGroupName: pointer.ToOrNilIfZeroValue(testSubnetGroupName),
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
