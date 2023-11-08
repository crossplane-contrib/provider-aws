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

package user

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
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/pkg/errors"

	"github.com/crossplane-contrib/provider-aws/apis/iam/v1beta1"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/iam/fake"
	errorutils "github.com/crossplane-contrib/provider-aws/pkg/utils/errors"
)

var (
	unexpectedItem resource.Managed
	userName       = "some user"

	errBoom = errors.New("boom")

	sortTags = cmpopts.SortSlices(func(a, b v1beta1.Tag) bool {
		return a.Key > b.Key
	})

	tagComparer = cmp.Comparer(func(expected, actual awsiamtypes.Tag) bool {
		return cmp.Equal(expected.Key, actual.Key) &&
			cmp.Equal(expected.Value, actual.Value)
	})

	tagInputComparer = cmp.Comparer(func(expected, actual *awsiam.TagUserInput) bool {
		return cmp.Equal(expected.UserName, actual.UserName) &&
			cmp.Equal(expected.Tags, actual.Tags, tagComparer, sortIAMTags)
	})

	untagInputComparer = cmp.Comparer(func(expected, actual *awsiam.UntagUserInput) bool {
		return cmp.Equal(expected.UserName, actual.UserName) &&
			cmp.Equal(expected.TagKeys, actual.TagKeys, sortStrings)
	})

	sortIAMTags = cmpopts.SortSlices(func(a, b awsiamtypes.Tag) bool {
		return *a.Key > *b.Key
	})

	sortStrings = cmpopts.SortSlices(func(x, y string) bool {
		return x < y
	})
)

type args struct {
	iam *fake.MockUserClient
	cr  resource.Managed
}

type userModifier func(*v1beta1.User)

func withConditions(c ...xpv1.Condition) userModifier {
	return func(r *v1beta1.User) { r.Status.ConditionedStatus.Conditions = c }
}

func withExternalName(name string) userModifier {
	return func(r *v1beta1.User) { meta.SetExternalName(r, name) }
}

func withTags(tagMaps ...map[string]string) userModifier {
	var tagList []v1beta1.Tag
	for _, tagMap := range tagMaps {
		for k, v := range tagMap {
			tagList = append(tagList, v1beta1.Tag{Key: k, Value: v})
		}
	}
	return func(r *v1beta1.User) {
		r.Spec.ForProvider.Tags = tagList
	}
}

func withPath(path *string) userModifier {
	return func(r *v1beta1.User) {
		r.Spec.ForProvider.Path = path
	}
}

func withBoundary(boundary *string) userModifier {
	return func(r *v1beta1.User) {
		r.Spec.ForProvider.PermissionsBoundary = boundary
	}
}

func user(m ...userModifier) *v1beta1.User {
	cr := &v1beta1.User{}
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
				iam: &fake.MockUserClient{
					MockGetUser: func(ctx context.Context, input *awsiam.GetUserInput, opts []func(*awsiam.Options)) (*awsiam.GetUserOutput, error) {
						return &awsiam.GetUserOutput{
							User: &awsiamtypes.User{},
						}, nil
					},
				},
				cr: user(withExternalName(userName)),
			},
			want: want{
				cr: user(withExternalName(userName),
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
		"GetUserError": {
			args: args{
				iam: &fake.MockUserClient{
					MockGetUser: func(ctx context.Context, input *awsiam.GetUserInput, opts []func(*awsiam.Options)) (*awsiam.GetUserOutput, error) {
						return nil, errBoom
					},
				},
				cr: user(withExternalName(userName)),
			},
			want: want{
				cr:  user(withExternalName(userName)),
				err: errorutils.Wrap(errBoom, errGet),
			},
		},
		"DifferentTags": {
			args: args{
				iam: &fake.MockUserClient{
					MockGetUser: func(ctx context.Context, input *awsiam.GetUserInput, opts []func(*awsiam.Options)) (*awsiam.GetUserOutput, error) {
						return &awsiam.GetUserOutput{
							User: &awsiamtypes.User{
								Tags: []awsiamtypes.Tag{
									{Key: aws.String("k1"), Value: aws.String("v1")},
								},
							},
						}, nil
					},
				},
				cr: user(withExternalName(userName), withTags(map[string]string{"k1": "v2"})),
			},
			want: want{
				cr: user(withExternalName(userName),
					withConditions(xpv1.Available()),
					withTags(map[string]string{"k1": "v2"})),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: false,
				},
			},
		},
		"DifferentBoundary": {
			args: args{
				iam: &fake.MockUserClient{
					MockGetUser: func(ctx context.Context, input *awsiam.GetUserInput, opts []func(*awsiam.Options)) (*awsiam.GetUserOutput, error) {
						return &awsiam.GetUserOutput{
							User: &awsiamtypes.User{
								PermissionsBoundary: &awsiamtypes.AttachedPermissionsBoundary{
									PermissionsBoundaryArn: aws.String("old"),
								},
							},
						}, nil
					},
				},
				cr: user(withExternalName(userName), withBoundary(aws.String("new"))),
			},
			want: want{
				cr: user(withExternalName(userName),
					withConditions(xpv1.Available()),
					withBoundary(aws.String("new"))),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: false,
				},
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
				iam: &fake.MockUserClient{
					MockCreateUser: func(ctx context.Context, input *awsiam.CreateUserInput, opts []func(*awsiam.Options)) (*awsiam.CreateUserOutput, error) {
						return &awsiam.CreateUserOutput{}, nil
					},
				},
				cr: user(withExternalName(userName)),
			},
			want: want{
				cr: user(
					withExternalName(userName),
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
				iam: &fake.MockUserClient{
					MockCreateUser: func(ctx context.Context, input *awsiam.CreateUserInput, opts []func(*awsiam.Options)) (*awsiam.CreateUserOutput, error) {
						return nil, errBoom
					},
				},
				cr: user(),
			},
			want: want{
				cr:  user(withConditions(xpv1.Creating())),
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
		"InValidInput": {
			args: args{
				cr: unexpectedItem,
			},
			want: want{
				cr:  unexpectedItem,
				err: errors.New(errUnexpectedObject),
			},
		},
		"ValidInput": {
			args: args{
				iam: &fake.MockUserClient{
					MockUpdateUser: func(ctx context.Context, input *awsiam.UpdateUserInput, opts []func(*awsiam.Options)) (*awsiam.UpdateUserOutput, error) {
						return &awsiam.UpdateUserOutput{}, nil
					},
					MockGetUser: func(ctx context.Context, input *awsiam.GetUserInput, opts []func(*awsiam.Options)) (*awsiam.GetUserOutput, error) {
						return &awsiam.GetUserOutput{
							User: &awsiamtypes.User{},
						}, nil
					},
				},
				cr: user(withExternalName(userName)),
			},
			want: want{
				cr: user(withExternalName(userName)),
			},
		},
		"UpdateUserError": {
			args: args{
				iam: &fake.MockUserClient{
					MockUpdateUser: func(ctx context.Context, input *awsiam.UpdateUserInput, opts []func(*awsiam.Options)) (*awsiam.UpdateUserOutput, error) {
						return nil, errBoom
					},
					MockGetUser: func(ctx context.Context, input *awsiam.GetUserInput, opts []func(*awsiam.Options)) (*awsiam.GetUserOutput, error) {
						return &awsiam.GetUserOutput{
							User: &awsiamtypes.User{},
						}, nil
					},
				},
				cr: user(withExternalName(userName), withPath(aws.String("foo"))),
			},
			want: want{
				cr:  user(withExternalName(userName), withPath(aws.String("foo"))),
				err: errorutils.Wrap(errBoom, errUpdateUser),
			},
		},
		"GetUserError": {
			args: args{
				iam: &fake.MockUserClient{
					MockUpdateUser: func(ctx context.Context, input *awsiam.UpdateUserInput, opts []func(*awsiam.Options)) (*awsiam.UpdateUserOutput, error) {
						return &awsiam.UpdateUserOutput{}, nil
					},
					MockGetUser: func(ctx context.Context, input *awsiam.GetUserInput, opts []func(*awsiam.Options)) (*awsiam.GetUserOutput, error) {
						return nil, errBoom
					},
				},
				cr: user(withExternalName(userName)),
			},
			want: want{
				cr:  user(withExternalName(userName)),
				err: errorutils.Wrap(errBoom, errGet),
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

func TestUpdate_Tags(t *testing.T) {

	type want struct {
		cr         resource.Managed
		result     managed.ExternalUpdate
		err        error
		tagInput   *awsiam.TagUserInput
		untagInput *awsiam.UntagUserInput
	}

	cases := map[string]struct {
		args
		want
	}{
		"AddTagsError": {
			args: args{
				iam: &fake.MockUserClient{
					MockGetUser: func(ctx context.Context, input *awsiam.GetUserInput, opts []func(*awsiam.Options)) (*awsiam.GetUserOutput, error) {
						return &awsiam.GetUserOutput{
							User: &awsiamtypes.User{},
						}, nil
					},
					MockTagUser: func(ctx context.Context, params *awsiam.TagUserInput, opt []func(*awsiam.Options)) (*awsiam.TagUserOutput, error) {
						return nil, errBoom
					},
				},
				cr: user(
					withExternalName(userName),
					withTags(map[string]string{
						"key": "value",
					})),
			},
			want: want{
				cr: user(
					withExternalName(userName),
					withTags(map[string]string{
						"key": "value",
					})),
				err: errorutils.Wrap(errBoom, errTag),
			},
		},
		"AddTagsSuccess": {
			args: args{
				iam: &fake.MockUserClient{
					MockGetUser: func(ctx context.Context, input *awsiam.GetUserInput, opts []func(*awsiam.Options)) (*awsiam.GetUserOutput, error) {
						return &awsiam.GetUserOutput{
							User: &awsiamtypes.User{},
						}, nil
					},
					MockTagUser: func(ctx context.Context, params *awsiam.TagUserInput, opt []func(*awsiam.Options)) (*awsiam.TagUserOutput, error) {
						return nil, nil
					},
				},
				cr: user(
					withExternalName(userName),
					withTags(map[string]string{
						"key": "value",
					})),
			},
			want: want{
				cr: user(
					withExternalName(userName),
					withTags(map[string]string{
						"key": "value",
					})),
				tagInput: &awsiam.TagUserInput{
					UserName: aws.String(userName),
					Tags: []awsiamtypes.Tag{
						{Key: aws.String("key"), Value: aws.String("value")},
					},
				},
			},
		},
		"UpdateTagsSuccess": {
			args: args{
				iam: &fake.MockUserClient{
					MockGetUser: func(ctx context.Context, input *awsiam.GetUserInput, opts []func(*awsiam.Options)) (*awsiam.GetUserOutput, error) {
						return &awsiam.GetUserOutput{
							User: &awsiamtypes.User{
								Tags: []awsiamtypes.Tag{
									{Key: aws.String("key1"), Value: aws.String("value1")},
									{Key: aws.String("key2"), Value: aws.String("value2")},
								},
							},
						}, nil
					},
					MockTagUser: func(ctx context.Context, input *awsiam.TagUserInput, opts []func(*awsiam.Options)) (*awsiam.TagUserOutput, error) {
						return nil, nil
					},
				},
				cr: user(
					withExternalName(userName),
					withTags(map[string]string{
						"key1": "value1",
						"key2": "value1",
					})),
			},
			want: want{
				cr: user(
					withExternalName(userName),
					withTags(map[string]string{
						"key2": "value1",
						"key1": "value1",
					})),
				tagInput: &awsiam.TagUserInput{
					UserName: aws.String(userName),
					Tags: []awsiamtypes.Tag{
						{Key: aws.String("key2"), Value: aws.String("value1")},
					},
				},
			},
		},
		"RemoveTagsError": {
			args: args{
				iam: &fake.MockUserClient{
					MockGetUser: func(ctx context.Context, input *awsiam.GetUserInput, opts []func(*awsiam.Options)) (*awsiam.GetUserOutput, error) {
						return &awsiam.GetUserOutput{
							User: &awsiamtypes.User{
								Tags: []awsiamtypes.Tag{
									{Key: aws.String("key1"), Value: aws.String("value1")},
									{Key: aws.String("key2"), Value: aws.String("value2")},
								},
							},
						}, nil
					},
					MockUntagUser: func(ctx context.Context, input *awsiam.UntagUserInput, opts []func(*awsiam.Options)) (*awsiam.UntagUserOutput, error) {
						return nil, errBoom
					},
				},
				cr: user(
					withExternalName(userName),
					withTags(map[string]string{
						"key2": "value2",
					})),
			},
			want: want{
				cr: user(
					withExternalName(userName),
					withTags(map[string]string{
						"key2": "value2",
					})),
				err: errorutils.Wrap(errBoom, errUntag),
			},
		},
		"RemoveTagsSuccess": {
			args: args{
				iam: &fake.MockUserClient{
					MockGetUser: func(ctx context.Context, input *awsiam.GetUserInput, opts []func(*awsiam.Options)) (*awsiam.GetUserOutput, error) {
						return &awsiam.GetUserOutput{
							User: &awsiamtypes.User{
								Tags: []awsiamtypes.Tag{
									{Key: aws.String("key1"), Value: aws.String("value1")},
									{Key: aws.String("key2"), Value: aws.String("value2")},
								},
							},
						}, nil
					},
					MockUntagUser: func(ctx context.Context, input *awsiam.UntagUserInput, opts []func(*awsiam.Options)) (*awsiam.UntagUserOutput, error) {
						return nil, nil
					},
				},
				cr: user(
					withExternalName(userName),
					withTags(map[string]string{
						"key2": "value2",
					})),
			},
			want: want{
				cr: user(
					withExternalName(userName),
					withTags(map[string]string{
						"key2": "value2",
					})),
				untagInput: &awsiam.UntagUserInput{
					UserName: aws.String(userName),
					TagKeys:  []string{"key1"},
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			tc.iam.MockUpdateUser = func(ctx context.Context, input *awsiam.UpdateUserInput, opts []func(*awsiam.Options)) (*awsiam.UpdateUserOutput, error) {
				return nil, nil
			}

			e := &external{client: tc.iam}
			o, err := e.Update(context.Background(), tc.args.cr)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.cr, tc.args.cr, test.EquateConditions(), sortTags); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.result, o); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if tc.want.tagInput != nil {
				if diff := cmp.Diff(tc.want.tagInput, tc.iam.MockUserInput.TagUserInput, tagInputComparer); diff != "" {
					t.Errorf("r: -want, +got:\n%s", diff)
				}
			}
			if tc.want.untagInput != nil {
				if diff := cmp.Diff(tc.want.untagInput, tc.iam.MockUserInput.UntagUserInput, untagInputComparer); diff != "" {
					t.Errorf("r: -want, +got:\n%s", diff)
				}
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
				iam: &fake.MockUserClient{
					MockDeleteUser: func(ctx context.Context, input *awsiam.DeleteUserInput, opts []func(*awsiam.Options)) (*awsiam.DeleteUserOutput, error) {
						return &awsiam.DeleteUserOutput{}, nil
					},
				},
				cr: user(withExternalName(userName)),
			},
			want: want{
				cr: user(withExternalName(userName),
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
				iam: &fake.MockUserClient{
					MockDeleteUser: func(ctx context.Context, input *awsiam.DeleteUserInput, opts []func(*awsiam.Options)) (*awsiam.DeleteUserOutput, error) {
						return nil, errBoom
					},
				},
				cr: user(withExternalName(userName)),
			},
			want: want{
				cr: user(withExternalName(userName),
					withConditions(xpv1.Deleting())),
				err: errorutils.Wrap(errBoom, errDelete),
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
