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
	"encoding/json"
	"strings"
	"testing"

	awscognitoidentityprovider "github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	awscognitoidentityprovidertypes "github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"

	"github.com/crossplane-contrib/provider-aws/apis/cognitoidentityprovider/manualv1alpha1"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/cognitoidentityprovider"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/cognitoidentityprovider/fake"
	errorutils "github.com/crossplane-contrib/provider-aws/pkg/utils/errors"
)

var (
	unexpectedItem resource.Managed
	userPoolID     = "some pool"
	Username       = "some user"
	groupName      = "some group"
	errBoom        = errors.New("boom")
)

type args struct {
	cognitoidentityprovider cognitoidentityprovider.GroupUserMembershipClient
	cr                      resource.Managed
}

type userGroupModifier func(*manualv1alpha1.GroupUserMembership)

func withExternalName(groupname string, username string, userPoolID string) userGroupModifier {
	externalAnnotation := &manualv1alpha1.ExternalAnnotation{
		UserPoolID: &userPoolID,
		Groupname:  &groupname,
		Username:   &username,
	}
	payload, err := json.Marshal(externalAnnotation)
	if err != nil {
		panic(err)
	}
	return func(r *manualv1alpha1.GroupUserMembership) { meta.SetExternalName(r, string(payload)) }
}

func withConditions(c ...xpv1.Condition) userGroupModifier {
	return func(r *manualv1alpha1.GroupUserMembership) { r.Status.ConditionedStatus.Conditions = c }
}

func withSpecGroupName(s string) userGroupModifier {
	return func(r *manualv1alpha1.GroupUserMembership) { r.Spec.ForProvider.Groupname = s }
}

func withSpecUsername(s string) userGroupModifier {
	return func(r *manualv1alpha1.GroupUserMembership) { r.Spec.ForProvider.Username = s }
}

func withSpecUserPoolID(s string) userGroupModifier {
	return func(r *manualv1alpha1.GroupUserMembership) { r.Spec.ForProvider.UserPoolID = s }
}

func withStatusGroupname(s *string) userGroupModifier {
	return func(r *manualv1alpha1.GroupUserMembership) { r.Status.AtProvider.Groupname = s }
}

func userGroup(m ...userGroupModifier) *manualv1alpha1.GroupUserMembership {
	cr := &manualv1alpha1.GroupUserMembership{}
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
				cognitoidentityprovider: &fake.MockGroupUserMembershipClient{
					MockAdminListGroupsForUser: func(ctx context.Context, input *awscognitoidentityprovider.AdminListGroupsForUserInput, opts []func(*awscognitoidentityprovider.Options)) (*awscognitoidentityprovider.AdminListGroupsForUserOutput, error) {
						if strings.Compare(*input.UserPoolId, userPoolID) == 0 && strings.Compare(*input.Username, Username) == 0 {
							return &awscognitoidentityprovider.AdminListGroupsForUserOutput{
								Groups: []awscognitoidentityprovidertypes.GroupType{
									{
										GroupName:  &groupName,
										UserPoolId: &userPoolID,
									},
								},
							}, nil
						}
						return nil, &awscognitoidentityprovidertypes.ResourceNotFoundException{}
					},
				},
				cr: userGroup(withExternalName(groupName, Username, userPoolID)),
			},
			want: want{
				cr: userGroup(
					withExternalName(groupName, Username, userPoolID),
					withConditions(xpv1.Available()),
					withStatusGroupname(&groupName)),
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
				cognitoidentityprovider: &fake.MockGroupUserMembershipClient{
					MockAdminListGroupsForUser: func(ctx context.Context, input *awscognitoidentityprovider.AdminListGroupsForUserInput, opts []func(*awscognitoidentityprovider.Options)) (*awscognitoidentityprovider.AdminListGroupsForUserOutput, error) {
						return &awscognitoidentityprovider.AdminListGroupsForUserOutput{}, nil
					},
				},
				cr: userGroup(withSpecUsername(Username)),
			},
			want: want{
				cr: userGroup(withSpecUsername(Username)),
			},
		},
		"ClientError": {
			args: args{
				cognitoidentityprovider: &fake.MockGroupUserMembershipClient{
					MockAdminListGroupsForUser: func(ctx context.Context, input *awscognitoidentityprovider.AdminListGroupsForUserInput, opts []func(*awscognitoidentityprovider.Options)) (*awscognitoidentityprovider.AdminListGroupsForUserOutput, error) {
						return nil, errBoom
					},
				},
				cr: userGroup(withExternalName(groupName, Username, userPoolID)),
			},
			want: want{
				cr:  userGroup(withExternalName(groupName, Username, userPoolID)),
				err: errorutils.Wrap(errBoom, errGet),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.cognitoidentityprovider}
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
				cognitoidentityprovider: &fake.MockGroupUserMembershipClient{
					MockAdminAddUserToGroup: func(ctx context.Context, input *awscognitoidentityprovider.AdminAddUserToGroupInput, opts []func(*awscognitoidentityprovider.Options)) (*awscognitoidentityprovider.AdminAddUserToGroupOutput, error) {
						return &awscognitoidentityprovider.AdminAddUserToGroupOutput{}, nil
					},
				},
				cr: userGroup(withSpecGroupName(groupName),
					withSpecUsername(Username), withSpecUsername(Username),
					withSpecUserPoolID(userPoolID)),
			},
			want: want{
				cr: userGroup(
					withSpecGroupName(groupName),
					withSpecUsername(Username),
					withSpecUserPoolID(userPoolID),
					withExternalName(groupName, Username, userPoolID)),
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
				cognitoidentityprovider: &fake.MockGroupUserMembershipClient{
					MockAdminAddUserToGroup: func(ctx context.Context, input *awscognitoidentityprovider.AdminAddUserToGroupInput, opts []func(*awscognitoidentityprovider.Options)) (*awscognitoidentityprovider.AdminAddUserToGroupOutput, error) {
						return nil, errBoom
					},
				},
				cr: userGroup(withSpecGroupName(groupName),
					withSpecUsername(Username)),
			},
			want: want{
				cr: userGroup(withSpecGroupName(groupName),
					withSpecUsername(Username)),
				err: errorutils.Wrap(errBoom, errAdd),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.cognitoidentityprovider}
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
				cognitoidentityprovider: &fake.MockGroupUserMembershipClient{
					MockAdminRemoveUserFromGroup: func(ctx context.Context, input *awscognitoidentityprovider.AdminRemoveUserFromGroupInput, opts []func(*awscognitoidentityprovider.Options)) (*awscognitoidentityprovider.AdminRemoveUserFromGroupOutput, error) {
						return &awscognitoidentityprovider.AdminRemoveUserFromGroupOutput{}, nil
					},
				},
				cr: userGroup(withSpecGroupName(groupName),
					withSpecUsername(Username),
					withSpecUserPoolID(userPoolID)),
			},
			want: want{
				cr: userGroup(
					withSpecGroupName(groupName),
					withSpecUsername(Username),
					withSpecUserPoolID(userPoolID),
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
				cognitoidentityprovider: &fake.MockGroupUserMembershipClient{
					MockAdminRemoveUserFromGroup: func(ctx context.Context, input *awscognitoidentityprovider.AdminRemoveUserFromGroupInput, opts []func(*awscognitoidentityprovider.Options)) (*awscognitoidentityprovider.AdminRemoveUserFromGroupOutput, error) {
						return nil, errBoom
					},
				},
				cr: userGroup(withSpecGroupName(Username),
					withSpecUsername(Username)),
			},
			want: want{
				cr: userGroup(withSpecGroupName(Username),
					withSpecUsername(Username),
					withConditions(xpv1.Deleting())),
				err: errorutils.Wrap(errBoom, errRemove),
			},
		},
		"ResourceNotFoundException": {
			args: args{
				cognitoidentityprovider: &fake.MockGroupUserMembershipClient{
					MockAdminRemoveUserFromGroup: func(ctx context.Context, input *awscognitoidentityprovider.AdminRemoveUserFromGroupInput, opts []func(*awscognitoidentityprovider.Options)) (*awscognitoidentityprovider.AdminRemoveUserFromGroupOutput, error) {
						return nil, &awscognitoidentityprovidertypes.ResourceNotFoundException{}
					},
				},
				cr: userGroup(),
			},
			want: want{
				cr:  userGroup(withConditions(xpv1.Deleting())),
				err: errorutils.Wrap(&awscognitoidentityprovidertypes.ResourceNotFoundException{}, errRemove),
			},
		},
		"UserNotFoundException": {
			args: args{
				cognitoidentityprovider: &fake.MockGroupUserMembershipClient{
					MockAdminRemoveUserFromGroup: func(ctx context.Context, input *awscognitoidentityprovider.AdminRemoveUserFromGroupInput, opts []func(*awscognitoidentityprovider.Options)) (*awscognitoidentityprovider.AdminRemoveUserFromGroupOutput, error) {
						return nil, &awscognitoidentityprovidertypes.UserNotFoundException{}
					},
				},
				cr: userGroup(),
			},
			want: want{
				cr:  userGroup(withConditions(xpv1.Deleting())),
				err: errorutils.Wrap(&awscognitoidentityprovidertypes.UserNotFoundException{}, errRemove),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.cognitoidentityprovider}
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
