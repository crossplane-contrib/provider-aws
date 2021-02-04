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

package iamgroupusermembership

import (
	"context"
	"net/http"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsiam "github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/test"

	"github.com/crossplane/provider-aws/apis/identity/v1alpha1"
	awsclient "github.com/crossplane/provider-aws/pkg/clients"
	"github.com/crossplane/provider-aws/pkg/clients/iam"
	"github.com/crossplane/provider-aws/pkg/clients/iam/fake"
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

type userGroupModifier func(*v1alpha1.IAMGroupUserMembership)

func withConditions(c ...xpv1.Condition) userGroupModifier {
	return func(r *v1alpha1.IAMGroupUserMembership) { r.Status.ConditionedStatus.Conditions = c }
}

func withGroupName(s string) userGroupModifier {
	return func(r *v1alpha1.IAMGroupUserMembership) { r.Spec.ForProvider.GroupName = s }
}

func withSpecUserName(s string) userGroupModifier {
	return func(r *v1alpha1.IAMGroupUserMembership) { r.Spec.ForProvider.UserName = s }
}

func withStatusGroupArn(s string) userGroupModifier {
	return func(r *v1alpha1.IAMGroupUserMembership) { r.Status.AtProvider.AttachedGroupARN = s }
}

func userGroup(m ...userGroupModifier) *v1alpha1.IAMGroupUserMembership {
	cr := &v1alpha1.IAMGroupUserMembership{}
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
		"VaildInput": {
			args: args{
				iam: &fake.MockGroupUserMembershipClient{
					MockListGroupsForUser: func(input *awsiam.ListGroupsForUserInput) awsiam.ListGroupsForUserRequest {
						return awsiam.ListGroupsForUserRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsiam.ListGroupsForUserOutput{
								Groups: []awsiam.Group{
									{
										Arn:       &groupArn,
										GroupName: &groupName,
									},
								},
							}},
						}
					},
				},
				cr: userGroup(withGroupName(groupName),
					withSpecUserName(userName)),
			},
			want: want{
				cr: userGroup(withGroupName(groupName),
					withSpecUserName(userName),
					withConditions(xpv1.Available()),
					withStatusGroupArn(groupArn)),
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
		"NoAttachedGroup": {
			args: args{
				iam: &fake.MockGroupUserMembershipClient{
					MockListGroupsForUser: func(input *awsiam.ListGroupsForUserInput) awsiam.ListGroupsForUserRequest {
						return awsiam.ListGroupsForUserRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsiam.ListGroupsForUserOutput{}},
						}
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
					MockListGroupsForUser: func(input *awsiam.ListGroupsForUserInput) awsiam.ListGroupsForUserRequest {
						return awsiam.ListGroupsForUserRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Error: errBoom},
						}
					},
				},
				cr: userGroup(withGroupName(groupName)),
			},
			want: want{
				cr:  userGroup(withGroupName(groupName)),
				err: awsclient.Wrap(errBoom, errGet),
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
					MockAddUserToGroup: func(input *awsiam.AddUserToGroupInput) awsiam.AddUserToGroupRequest {
						return awsiam.AddUserToGroupRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsiam.AddUserToGroupOutput{}},
						}
					},
				},
				cr: userGroup(withGroupName(groupName),
					withSpecUserName(userName)),
			},
			want: want{
				cr: userGroup(
					withGroupName(groupName),
					withSpecUserName(userName),
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
				iam: &fake.MockGroupUserMembershipClient{
					MockAddUserToGroup: func(input *awsiam.AddUserToGroupInput) awsiam.AddUserToGroupRequest {
						return awsiam.AddUserToGroupRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Error: errBoom},
						}
					},
				},
				cr: userGroup(withGroupName(groupName),
					withSpecUserName(userName)),
			},
			want: want{
				cr: userGroup(withGroupName(groupName),
					withSpecUserName(userName),
					withConditions(xpv1.Creating())),
				err: awsclient.Wrap(errBoom, errAdd),
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
					MockRemoveUserFromGroup: func(input *awsiam.RemoveUserFromGroupInput) awsiam.RemoveUserFromGroupRequest {
						return awsiam.RemoveUserFromGroupRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsiam.RemoveUserFromGroupOutput{}},
						}
					},
				},
				cr: userGroup(withGroupName(groupName),
					withSpecUserName(userName)),
			},
			want: want{
				cr: userGroup(
					withGroupName(groupName),
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
					MockRemoveUserFromGroup: func(input *awsiam.RemoveUserFromGroupInput) awsiam.RemoveUserFromGroupRequest {
						return awsiam.RemoveUserFromGroupRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Error: errBoom},
						}
					},
				},
				cr: userGroup(withGroupName(userName),
					withSpecUserName(userName)),
			},
			want: want{
				cr: userGroup(withGroupName(userName),
					withSpecUserName(userName),
					withConditions(xpv1.Deleting())),
				err: awsclient.Wrap(errBoom, errRemove),
			},
		},
		"ResourceDoesNotExist": {
			args: args{
				iam: &fake.MockGroupUserMembershipClient{
					MockRemoveUserFromGroup: func(input *awsiam.RemoveUserFromGroupInput) awsiam.RemoveUserFromGroupRequest {
						return awsiam.RemoveUserFromGroupRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Error: errors.New(errRemove)},
						}
					},
				},
				cr: userGroup(),
			},
			want: want{
				cr:  userGroup(withConditions(xpv1.Deleting())),
				err: awsclient.Wrap(errors.New(errRemove), errRemove),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.iam}
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
