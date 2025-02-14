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

package group

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsiam "github.com/aws/aws-sdk-go-v2/service/iam"
	awsiamtypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"

	"github.com/crossplane-contrib/provider-aws/apis/iam/v1beta1"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/iam"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/iam/fake"
	errorutils "github.com/crossplane-contrib/provider-aws/pkg/utils/errors"
)

var (
	unexpectedItem resource.Managed
	groupName      = "some group"

	errBoom = errors.New("boom")
)

const (
	groupPath = "group-path"
)

type args struct {
	iam iam.GroupClient
	cr  resource.Managed
}

type groupModifier func(*v1beta1.Group)

func withConditions(c ...xpv1.Condition) groupModifier {
	return func(r *v1beta1.Group) { r.Status.ConditionedStatus.Conditions = c }
}

func withExternalName(name string) groupModifier {
	return func(r *v1beta1.Group) { meta.SetExternalName(r, name) }
}

func withGroupPath(groupPath string) groupModifier {
	return func(r *v1beta1.Group) { r.Spec.ForProvider.Path = &groupPath }
}

func group(m ...groupModifier) *v1beta1.Group {
	cr := &v1beta1.Group{}
	for _, f := range m {
		f(cr)
	}
	return cr
}

func TestObserve(t *testing.T) {

	type want struct {
		cr     resource.Managed
		result managed.ExternalObservation
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"ValidInput": {
			args: args{
				iam: &fake.MockGroupClient{
					MockGetGroup: func(ctx context.Context, input *awsiam.GetGroupInput, opts []func(*awsiam.Options)) (*awsiam.GetGroupOutput, error) {
						return &awsiam.GetGroupOutput{
							Group: &awsiamtypes.Group{
								GroupName: aws.String(groupName),
								Path:      aws.String(groupPath),
							},
						}, nil
					},
				},
				cr: group(withExternalName(groupName),
					withGroupPath(groupPath)),
			},
			want: want{
				cr: group(withExternalName(groupName),
					withGroupPath(groupPath),
					withConditions(xpv1.Available())),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
				},
			},
		},
		"InValidInput": {
			args: args{
				cr: unexpectedItem,
			},
			want: want{
				cr:  unexpectedItem,
				err: errors.New(errUnexpectedObject),
			},
		},
		"GetGroupError": {
			args: args{
				iam: &fake.MockGroupClient{
					MockGetGroup: func(ctx context.Context, input *awsiam.GetGroupInput, opts []func(*awsiam.Options)) (*awsiam.GetGroupOutput, error) {
						return nil, errBoom
					},
				},
				cr: group(withExternalName(groupName)),
			},
			want: want{
				cr:  group(withExternalName(groupName)),
				err: errorutils.Wrap(errBoom, errGet),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.iam}
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
		cr     resource.Managed
		result managed.ExternalCreation
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"ValidInput": {
			args: args{
				iam: &fake.MockGroupClient{
					MockCreateGroup: func(ctx context.Context, input *awsiam.CreateGroupInput, opts []func(*awsiam.Options)) (*awsiam.CreateGroupOutput, error) {
						return &awsiam.CreateGroupOutput{}, nil
					},
				},
				cr: group(withExternalName(groupName)),
			},
			want: want{
				cr: group(
					withExternalName(groupName),
					withConditions(xpv1.Creating())),
			},
		},
		"InValidInput": {
			args: args{
				cr: unexpectedItem,
			},
			want: want{
				cr:  unexpectedItem,
				err: errors.New(errUnexpectedObject),
			},
		},
		"ClientError": {
			args: args{
				iam: &fake.MockGroupClient{
					MockCreateGroup: func(ctx context.Context, input *awsiam.CreateGroupInput, opts []func(*awsiam.Options)) (*awsiam.CreateGroupOutput, error) {
						return nil, errBoom
					},
				},
				cr: group(),
			},
			want: want{
				cr:  group(withConditions(xpv1.Creating())),
				err: errorutils.Wrap(errBoom, errCreate),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.iam}
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
		cr     resource.Managed
		result managed.ExternalUpdate
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"ValidInput": {
			args: args{
				iam: &fake.MockGroupClient{
					MockUpdateGroup: func(ctx context.Context, input *awsiam.UpdateGroupInput, opts []func(*awsiam.Options)) (*awsiam.UpdateGroupOutput, error) {
						return &awsiam.UpdateGroupOutput{}, nil
					},
				},
				cr: group(withExternalName(groupName)),
			},
			want: want{
				cr: group(withExternalName(groupName)),
			},
		},
		"InValidInput": {
			args: args{
				cr: unexpectedItem,
			},
			want: want{
				cr:  unexpectedItem,
				err: errors.New(errUnexpectedObject),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.iam}
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
		cr  resource.Managed
		err error
	}

	cases := map[string]struct {
		args
		want
	}{
		"ValidInput": {
			args: args{
				iam: &fake.MockGroupClient{
					MockDeleteGroup: func(ctx context.Context, input *awsiam.DeleteGroupInput, opts []func(*awsiam.Options)) (*awsiam.DeleteGroupOutput, error) {
						return &awsiam.DeleteGroupOutput{}, nil
					},
				},
				cr: group(withExternalName(groupName)),
			},
			want: want{
				cr: group(withExternalName(groupName),
					withConditions(xpv1.Deleting())),
			},
		},
		"InValidInput": {
			args: args{
				cr: unexpectedItem,
			},
			want: want{
				cr:  unexpectedItem,
				err: errors.New(errUnexpectedObject),
			},
		},
		"DeleteError": {
			args: args{
				iam: &fake.MockGroupClient{
					MockDeleteGroup: func(ctx context.Context, input *awsiam.DeleteGroupInput, opts []func(*awsiam.Options)) (*awsiam.DeleteGroupOutput, error) {
						return nil, errBoom
					},
				},
				cr: group(withExternalName(groupName)),
			},
			want: want{
				cr: group(withExternalName(groupName),
					withConditions(xpv1.Deleting())),
				err: errorutils.Wrap(errBoom, errDelete),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.iam}
			_, err := e.Delete(context.Background(), tc.args.cr)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.cr, tc.args.cr, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}
