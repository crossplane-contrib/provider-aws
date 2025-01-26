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

package groupusermembership

import (
	"context"
	"testing"

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
	groupArn       = "some group"
	userName       = "some user"
	groupName      = "some group"
	errBoom        = errors.New("boom")
)

type args struct {
	iam iam.GroupUserMembershipClient
	cr  resource.Managed
}

type userGroupModifier func(*v1beta1.GroupUserMembership)

func withExternalName(name string) userGroupModifier {
	return func(r *v1beta1.GroupUserMembership) { meta.SetExternalName(r, name) }
}

func withConditions(c ...xpv1.Condition) userGroupModifier {
	return func(r *v1beta1.GroupUserMembership) { r.Status.ConditionedStatus.Conditions = c }
}

func withSpecGroupName(s string) userGroupModifier {
	return func(r *v1beta1.GroupUserMembership) { r.Spec.ForProvider.GroupName = s }
}

func withSpecUserName(s string) userGroupModifier {
	return func(r *v1beta1.GroupUserMembership) { r.Spec.ForProvider.UserName = s }
}

func withStatusGroupArn(s string) userGroupModifier {
	return func(r *v1beta1.GroupUserMembership) { r.Status.AtProvider.AttachedGroupARN = s }
}

func userGroup(m ...userGroupModifier) *v1beta1.GroupUserMembership {
	cr := &v1beta1.GroupUserMembership{}
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
				iam: &fake.MockGroupUserMembershipClient{
					MockListGroupsForUser: func(ctx context.Context, input *awsiam.ListGroupsForUserInput, opts []func(*awsiam.Options)) (*awsiam.ListGroupsForUserOutput, error) {
						return &awsiam.ListGroupsForUserOutput{
							Groups: []awsiamtypes.Group{
								{
									Arn:       &groupArn,
									GroupName: &groupName,
								},
							},
						}, nil
					},
				},
				cr: userGroup(withExternalName(groupName + "/" + userName)),
			},
			want: want{
				cr: userGroup(
					withExternalName(groupName+"/"+userName),
					withConditions(xpv1.Available()),
					withStatusGroupArn(groupArn)),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
				},
			},
		},
		"InvalidInput": {
			args: args{
				cr: unexpectedItem,
			},
			want: want{
				cr:  unexpectedItem,
				err: errors.New(errUnexpectedObject),
			},
		},
		"NoAttachedGroup": {
			args: args{
				iam: &fake.MockGroupUserMembershipClient{
					MockListGroupsForUser: func(ctx context.Context, input *awsiam.ListGroupsForUserInput, opts []func(*awsiam.Options)) (*awsiam.ListGroupsForUserOutput, error) {
						return &awsiam.ListGroupsForUserOutput{}, nil
					},
				},
				cr: userGroup(withSpecUserName(userName)),
			},
			want: want{
				cr: userGroup(withSpecUserName(userName)),
			},
		},
		"ClientError": {
			args: args{
				iam: &fake.MockGroupUserMembershipClient{
					MockListGroupsForUser: func(ctx context.Context, input *awsiam.ListGroupsForUserInput, opts []func(*awsiam.Options)) (*awsiam.ListGroupsForUserOutput, error) {
						return nil, errBoom
					},
				},
				cr: userGroup(withExternalName(groupName + "/" + userName)),
			},
			want: want{
				cr:  userGroup(withExternalName(groupName + "/" + userName)),
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
		"VaildInput": {
			args: args{
				iam: &fake.MockGroupUserMembershipClient{
					MockAddUserToGroup: func(ctx context.Context, input *awsiam.AddUserToGroupInput, opts []func(*awsiam.Options)) (*awsiam.AddUserToGroupOutput, error) {
						return &awsiam.AddUserToGroupOutput{}, nil
					},
				},
				cr: userGroup(withSpecGroupName(groupName),
					withSpecUserName(userName)),
			},
			want: want{
				cr: userGroup(
					withSpecGroupName(groupName),
					withSpecUserName(userName),
					withExternalName(groupName+"/"+userName)),
				result: managed.ExternalCreation{},
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
				iam: &fake.MockGroupUserMembershipClient{
					MockAddUserToGroup: func(ctx context.Context, input *awsiam.AddUserToGroupInput, opts []func(*awsiam.Options)) (*awsiam.AddUserToGroupOutput, error) {
						return nil, errBoom
					},
				},
				cr: userGroup(withSpecGroupName(groupName),
					withSpecUserName(userName)),
			},
			want: want{
				cr: userGroup(withSpecGroupName(groupName),
					withSpecUserName(userName)),
				err: errorutils.Wrap(errBoom, errAdd),
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

func TestDelete(t *testing.T) {

	type want struct {
		cr  resource.Managed
		err error
	}

	cases := map[string]struct {
		args
		want
	}{
		"VaildInput": {
			args: args{
				iam: &fake.MockGroupUserMembershipClient{
					MockRemoveUserFromGroup: func(ctx context.Context, input *awsiam.RemoveUserFromGroupInput, opts []func(*awsiam.Options)) (*awsiam.RemoveUserFromGroupOutput, error) {
						return &awsiam.RemoveUserFromGroupOutput{}, nil
					},
				},
				cr: userGroup(withSpecGroupName(groupName),
					withSpecUserName(userName)),
			},
			want: want{
				cr: userGroup(
					withSpecGroupName(groupName),
					withSpecUserName(userName),
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
		"ClientError": {
			args: args{
				iam: &fake.MockGroupUserMembershipClient{
					MockRemoveUserFromGroup: func(ctx context.Context, input *awsiam.RemoveUserFromGroupInput, opts []func(*awsiam.Options)) (*awsiam.RemoveUserFromGroupOutput, error) {
						return nil, errBoom
					},
				},
				cr: userGroup(withSpecGroupName(userName),
					withSpecUserName(userName)),
			},
			want: want{
				cr: userGroup(withSpecGroupName(userName),
					withSpecUserName(userName),
					withConditions(xpv1.Deleting())),
				err: errorutils.Wrap(errBoom, errRemove),
			},
		},
		"ResourceDoesNotExist": {
			args: args{
				iam: &fake.MockGroupUserMembershipClient{
					MockRemoveUserFromGroup: func(ctx context.Context, input *awsiam.RemoveUserFromGroupInput, opts []func(*awsiam.Options)) (*awsiam.RemoveUserFromGroupOutput, error) {
						return nil, &awsiamtypes.NoSuchEntityException{}
					},
				},
				cr: userGroup(),
			},
			want: want{
				cr:  userGroup(withConditions(xpv1.Deleting())),
				err: errorutils.Wrap(&awsiamtypes.NoSuchEntityException{}, errRemove),
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
