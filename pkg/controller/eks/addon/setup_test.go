/*
Copyright 2021 The Crossplane Authors.

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

package addon

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"

	awseks "github.com/aws/aws-sdk-go/service/eks"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/test"

	"github.com/crossplane-contrib/provider-aws/apis/eks/v1alpha1"
	awsclient "github.com/crossplane-contrib/provider-aws/pkg/clients"
	mockeksiface "github.com/crossplane-contrib/provider-aws/pkg/clients/eks/fake/eksiface"
)

var (
	testExternalName          = "test-external-name"
	testServiceAccountRoleArn = "test-role"
	testAddonName             = "test-addon"
	testAddonVersion          = "v0.0.0"
	testClusterName           = "test-cluster"
	testResolveConflict       = "test-resolve-conflict"
	testTagKey                = "test-key"
	testTagValue              = "test-value"
	testOtherTagKey           = "test-other-key"
	testOtherTagValue         = "test-other-value"
	errBoom                   = errors.New("boom")
)

type mockClientFn func(t *testing.T) *mockeksiface.MockEKSAPI

type args struct {
	eks mockClientFn
	cr  *v1alpha1.Addon
}

type AddonModifier func(*v1alpha1.Addon)

func withExternalName(val string) AddonModifier {
	return func(r *v1alpha1.Addon) { meta.SetExternalName(r, val) }
}

func withSpec(p v1alpha1.AddonParameters) AddonModifier {
	return func(r *v1alpha1.Addon) { r.Spec.ForProvider = p }
}

func withConditions(c ...xpv1.Condition) AddonModifier {
	return func(r *v1alpha1.Addon) { r.Status.ConditionedStatus.Conditions = c }
}

func withStatus(s v1alpha1.AddonObservation) AddonModifier {
	return func(r *v1alpha1.Addon) { r.Status.AtProvider = s }
}

type mockClientModifier func(me *mockeksiface.MockEKSAPI)

func mockClient(m mockClientModifier) mockClientFn {
	return func(t *testing.T) *mockeksiface.MockEKSAPI {
		ctrl := gomock.NewController(t)
		mock := mockeksiface.NewMockEKSAPI(ctrl)
		m(mock)
		return mock
	}
}

func addon(m ...AddonModifier) *v1alpha1.Addon {
	cr := &v1alpha1.Addon{}
	for _, f := range m {
		f(cr)
	}
	return cr
}

func TestObserve(t *testing.T) {
	type want struct {
		cr     *v1alpha1.Addon
		result managed.ExternalObservation
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"SuccessfulAvailable": {
			args: args{
				eks: mockClient(func(me *mockeksiface.MockEKSAPI) {
					me.EXPECT().
						DescribeAddonWithContext(
							context.Background(),
							&awseks.DescribeAddonInput{},
						).
						Return(&awseks.DescribeAddonOutput{
							Addon: &awseks.Addon{
								Status: awsclient.String(awseks.AddonStatusActive),
							},
						}, nil)
				}),
				cr: addon(
					withExternalName(testExternalName),
				),
			},
			want: want{
				cr: addon(
					withExternalName(testExternalName),
					withConditions(xpv1.Available()),
					withStatus(v1alpha1.AddonObservation{
						Status: awsclient.String(awseks.AddonStatusActive),
					}),
				),
				result: managed.ExternalObservation{
					ResourceExists:          true,
					ResourceUpToDate:        true,
					ResourceLateInitialized: false,
				},
			},
		},
		"FailedDescribeRequest": {
			args: args{
				eks: mockClient(func(me *mockeksiface.MockEKSAPI) {
					me.EXPECT().
						DescribeAddonWithContext(
							context.Background(),
							&awseks.DescribeAddonInput{},
						).
						Return(nil, errBoom)
				}),
				cr: addon(
					withExternalName(testExternalName),
				),
			},
			want: want{
				cr: addon(
					withExternalName(testExternalName),
				),
				err: awsclient.Wrap(errBoom, errDescribe),
			},
		},
		"LateInitSuccess": {
			args: args{
				eks: mockClient(func(me *mockeksiface.MockEKSAPI) {
					me.EXPECT().
						DescribeAddonWithContext(
							context.Background(),
							&awseks.DescribeAddonInput{},
						).
						Return(&awseks.DescribeAddonOutput{
							Addon: &awseks.Addon{
								ServiceAccountRoleArn: &testServiceAccountRoleArn,
								Status:                awsclient.String(awseks.AddonStatusActive),
							},
						}, nil)
				}),
				cr: addon(
					withExternalName(testExternalName),
				),
			},
			want: want{
				cr: addon(
					withExternalName(testExternalName),
					withConditions(xpv1.Available()),
					withSpec(
						v1alpha1.AddonParameters{
							ServiceAccountRoleARN: &testServiceAccountRoleArn,
						},
					),
					withStatus(v1alpha1.AddonObservation{
						Status: awsclient.String(awseks.AddonStatusActive),
					}),
				),
				result: managed.ExternalObservation{
					ResourceExists:          true,
					ResourceUpToDate:        true,
					ResourceLateInitialized: true,
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := newExternal(nil, tc.eks(t), []option{setupHooks})
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
		cr     *v1alpha1.Addon
		result managed.ExternalCreation
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"Successful": {
			args: args{
				eks: mockClient(func(me *mockeksiface.MockEKSAPI) {
					me.EXPECT().
						CreateAddonWithContext(
							context.Background(),
							&awseks.CreateAddonInput{
								AddonName:             &testAddonName,
								AddonVersion:          &testAddonVersion,
								ClusterName:           &testClusterName,
								ServiceAccountRoleArn: &testServiceAccountRoleArn,
								ResolveConflicts:      &testResolveConflict,
							},
						).
						Return(&awseks.CreateAddonOutput{
							Addon: &awseks.Addon{
								AddonArn:              &testExternalName,
								ServiceAccountRoleArn: &testServiceAccountRoleArn,
								AddonVersion:          &testAddonVersion,
								AddonName:             &testAddonName,
							},
						}, nil)
				}),
				cr: addon(
					withSpec(v1alpha1.AddonParameters{
						AddonName:             &testAddonName,
						AddonVersion:          &testAddonVersion,
						ResolveConflicts:      &testResolveConflict,
						ServiceAccountRoleARN: &testServiceAccountRoleArn,
						CustomAddonParameters: v1alpha1.CustomAddonParameters{
							ClusterName: &testClusterName,
						},
					}),
				),
			},
			want: want{
				cr: addon(
					withExternalName(testExternalName),
					withSpec(v1alpha1.AddonParameters{
						AddonName:             &testAddonName,
						AddonVersion:          &testAddonVersion,
						ResolveConflicts:      &testResolveConflict,
						ServiceAccountRoleARN: &testServiceAccountRoleArn,
						CustomAddonParameters: v1alpha1.CustomAddonParameters{
							ClusterName: &testClusterName,
						},
					}),
					withStatus(
						v1alpha1.AddonObservation{AddonARN: &testExternalName},
					),
					withConditions(xpv1.Creating()),
				),
				result: managed.ExternalCreation{},
			},
		},
		"FailedRequest": {
			args: args{
				eks: mockClient(func(me *mockeksiface.MockEKSAPI) {
					me.EXPECT().
						CreateAddonWithContext(
							context.Background(),
							&awseks.CreateAddonInput{
								AddonName:             &testAddonName,
								AddonVersion:          &testAddonVersion,
								ClusterName:           &testClusterName,
								ServiceAccountRoleArn: &testServiceAccountRoleArn,
								ResolveConflicts:      &testResolveConflict,
							},
						).
						Return(nil, errBoom)
				}),
				cr: addon(
					withSpec(v1alpha1.AddonParameters{
						AddonName:             &testAddonName,
						AddonVersion:          &testAddonVersion,
						ResolveConflicts:      &testResolveConflict,
						ServiceAccountRoleARN: &testServiceAccountRoleArn,
						CustomAddonParameters: v1alpha1.CustomAddonParameters{
							ClusterName: &testClusterName,
						},
					}),
				),
			},
			want: want{
				cr: addon(
					withSpec(v1alpha1.AddonParameters{
						AddonName:             &testAddonName,
						AddonVersion:          &testAddonVersion,
						ResolveConflicts:      &testResolveConflict,
						ServiceAccountRoleARN: &testServiceAccountRoleArn,
						CustomAddonParameters: v1alpha1.CustomAddonParameters{
							ClusterName: &testClusterName,
						},
					}),
					withConditions(xpv1.Creating()),
				),
				err: awsclient.Wrap(errBoom, errCreate),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := newExternal(nil, tc.eks(t), []option{setupHooks})
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
		cr     *v1alpha1.Addon
		result managed.ExternalUpdate
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"Successful": {
			args: args{
				eks: mockClient(func(me *mockeksiface.MockEKSAPI) {
					me.EXPECT().
						UpdateAddonWithContext(
							context.Background(),
							&awseks.UpdateAddonInput{
								AddonName:             &testAddonName,
								AddonVersion:          &testAddonVersion,
								ClusterName:           &testClusterName,
								ServiceAccountRoleArn: &testServiceAccountRoleArn,
								ResolveConflicts:      &testResolveConflict,
							},
						).
						Return(&awseks.UpdateAddonOutput{}, nil)
					me.EXPECT().
						DescribeAddonWithContext(
							context.Background(),
							&awseks.DescribeAddonInput{
								AddonName:   &testAddonName,
								ClusterName: &testClusterName,
							},
						).
						Return(&awseks.DescribeAddonOutput{
							Addon: &awseks.Addon{
								Tags: map[string]*string{
									testOtherTagKey: &testOtherTagValue,
								},
							},
						}, nil)
					me.EXPECT().
						TagResourceWithContext(
							context.Background(),
							&awseks.TagResourceInput{
								ResourceArn: &testExternalName,
								Tags: map[string]*string{
									testTagKey: &testTagValue,
								},
							}).
						Return(&awseks.TagResourceOutput{}, nil)
					me.EXPECT().
						UntagResourceWithContext(
							context.Background(),
							&awseks.UntagResourceInput{
								ResourceArn: &testExternalName,
								TagKeys:     []*string{&testOtherTagKey},
							}).
						Return(&awseks.UntagResourceOutput{}, nil)
				}),
				cr: addon(
					withExternalName(testExternalName),
					withSpec(v1alpha1.AddonParameters{
						AddonName:             &testAddonName,
						AddonVersion:          &testAddonVersion,
						ResolveConflicts:      &testResolveConflict,
						ServiceAccountRoleARN: &testServiceAccountRoleArn,
						CustomAddonParameters: v1alpha1.CustomAddonParameters{
							ClusterName: &testClusterName,
						},
						Tags: map[string]*string{
							testTagKey: &testTagValue,
						},
					}),
					withStatus(
						v1alpha1.AddonObservation{AddonARN: &testExternalName},
					),
				),
			},
			want: want{
				cr: addon(
					withExternalName(testExternalName),
					withSpec(v1alpha1.AddonParameters{
						AddonName:             &testAddonName,
						AddonVersion:          &testAddonVersion,
						ResolveConflicts:      &testResolveConflict,
						ServiceAccountRoleARN: &testServiceAccountRoleArn,
						CustomAddonParameters: v1alpha1.CustomAddonParameters{
							ClusterName: &testClusterName,
						},
						Tags: map[string]*string{
							testTagKey: &testTagValue,
						},
					}),
					withStatus(
						v1alpha1.AddonObservation{AddonARN: &testExternalName},
					),
				),
				result: managed.ExternalUpdate{},
			},
		},
		"FailedUpdateRequest": {
			args: args{
				eks: mockClient(func(me *mockeksiface.MockEKSAPI) {
					me.EXPECT().
						UpdateAddonWithContext(
							context.Background(),
							&awseks.UpdateAddonInput{
								AddonName:             &testAddonName,
								AddonVersion:          &testAddonVersion,
								ClusterName:           &testClusterName,
								ServiceAccountRoleArn: &testServiceAccountRoleArn,
								ResolveConflicts:      &testResolveConflict,
							},
						).
						Return(nil, errBoom)
				}),
				cr: addon(
					withExternalName(testExternalName),
					withSpec(v1alpha1.AddonParameters{
						AddonName:             &testAddonName,
						AddonVersion:          &testAddonVersion,
						ResolveConflicts:      &testResolveConflict,
						ServiceAccountRoleARN: &testServiceAccountRoleArn,
						CustomAddonParameters: v1alpha1.CustomAddonParameters{
							ClusterName: &testClusterName,
						},
					}),
				),
			},
			want: want{
				cr: addon(
					withExternalName(testExternalName),
					withSpec(v1alpha1.AddonParameters{
						AddonName:             &testAddonName,
						AddonVersion:          &testAddonVersion,
						ResolveConflicts:      &testResolveConflict,
						ServiceAccountRoleARN: &testServiceAccountRoleArn,
						CustomAddonParameters: v1alpha1.CustomAddonParameters{
							ClusterName: &testClusterName,
						},
					}),
				),
				err: awsclient.Wrap(errBoom, errUpdate),
			},
		},
		"FailedDescribeAddon": {
			args: args{
				eks: mockClient(func(me *mockeksiface.MockEKSAPI) {
					me.EXPECT().
						UpdateAddonWithContext(
							context.Background(),
							&awseks.UpdateAddonInput{
								AddonName:             &testAddonName,
								AddonVersion:          &testAddonVersion,
								ClusterName:           &testClusterName,
								ServiceAccountRoleArn: &testServiceAccountRoleArn,
								ResolveConflicts:      &testResolveConflict,
							},
						).
						Return(&awseks.UpdateAddonOutput{}, nil)
					me.EXPECT().
						DescribeAddonWithContext(
							context.Background(),
							&awseks.DescribeAddonInput{
								AddonName:   &testAddonName,
								ClusterName: &testClusterName,
							},
						).
						Return(nil, errBoom)
				}),
				cr: addon(
					withExternalName(testExternalName),
					withSpec(v1alpha1.AddonParameters{
						AddonName:             &testAddonName,
						AddonVersion:          &testAddonVersion,
						ResolveConflicts:      &testResolveConflict,
						ServiceAccountRoleARN: &testServiceAccountRoleArn,
						CustomAddonParameters: v1alpha1.CustomAddonParameters{
							ClusterName: &testClusterName,
						},
					}),
				),
			},
			want: want{
				cr: addon(
					withExternalName(testExternalName),
					withSpec(v1alpha1.AddonParameters{
						AddonName:             &testAddonName,
						AddonVersion:          &testAddonVersion,
						ResolveConflicts:      &testResolveConflict,
						ServiceAccountRoleARN: &testServiceAccountRoleArn,
						CustomAddonParameters: v1alpha1.CustomAddonParameters{
							ClusterName: &testClusterName,
						},
					}),
				),
				err: awsclient.Wrap(errBoom, errDescribe),
			},
		},
		"FailedTagResource": {
			args: args{
				eks: mockClient(func(me *mockeksiface.MockEKSAPI) {
					me.EXPECT().
						UpdateAddonWithContext(
							context.Background(),
							&awseks.UpdateAddonInput{
								AddonName:             &testAddonName,
								AddonVersion:          &testAddonVersion,
								ClusterName:           &testClusterName,
								ServiceAccountRoleArn: &testServiceAccountRoleArn,
								ResolveConflicts:      &testResolveConflict,
							},
						).
						Return(&awseks.UpdateAddonOutput{}, nil)
					me.EXPECT().
						DescribeAddonWithContext(
							context.Background(),
							&awseks.DescribeAddonInput{
								AddonName:   &testAddonName,
								ClusterName: &testClusterName,
							},
						).
						Return(&awseks.DescribeAddonOutput{
							Addon: &awseks.Addon{
								Tags: map[string]*string{
									testOtherTagKey: &testOtherTagValue,
								},
							},
						}, nil)
					me.EXPECT().
						TagResourceWithContext(
							context.Background(),
							&awseks.TagResourceInput{
								ResourceArn: &testExternalName,
								Tags: map[string]*string{
									testTagKey: &testTagValue,
								},
							}).
						Return(nil, errBoom)
				}),
				cr: addon(
					withExternalName(testExternalName),
					withSpec(v1alpha1.AddonParameters{
						AddonName:             &testAddonName,
						AddonVersion:          &testAddonVersion,
						ResolveConflicts:      &testResolveConflict,
						ServiceAccountRoleARN: &testServiceAccountRoleArn,
						CustomAddonParameters: v1alpha1.CustomAddonParameters{
							ClusterName: &testClusterName,
						},
						Tags: map[string]*string{
							testTagKey: &testTagValue,
						},
					}),
					withStatus(
						v1alpha1.AddonObservation{AddonARN: &testExternalName},
					),
				),
			},
			want: want{
				cr: addon(
					withExternalName(testExternalName),
					withSpec(v1alpha1.AddonParameters{
						AddonName:             &testAddonName,
						AddonVersion:          &testAddonVersion,
						ResolveConflicts:      &testResolveConflict,
						ServiceAccountRoleARN: &testServiceAccountRoleArn,
						CustomAddonParameters: v1alpha1.CustomAddonParameters{
							ClusterName: &testClusterName,
						},
						Tags: map[string]*string{
							testTagKey: &testTagValue,
						},
					}),
					withStatus(
						v1alpha1.AddonObservation{AddonARN: &testExternalName},
					),
				),
				err: awsclient.Wrap(errBoom, errTagResource),
			},
		},
		"UntagResource": {
			args: args{
				eks: mockClient(func(me *mockeksiface.MockEKSAPI) {
					me.EXPECT().
						UpdateAddonWithContext(
							context.Background(),
							&awseks.UpdateAddonInput{
								AddonName:             &testAddonName,
								AddonVersion:          &testAddonVersion,
								ClusterName:           &testClusterName,
								ServiceAccountRoleArn: &testServiceAccountRoleArn,
								ResolveConflicts:      &testResolveConflict,
							},
						).
						Return(&awseks.UpdateAddonOutput{}, nil)
					me.EXPECT().
						DescribeAddonWithContext(
							context.Background(),
							&awseks.DescribeAddonInput{
								AddonName:   &testAddonName,
								ClusterName: &testClusterName,
							},
						).
						Return(&awseks.DescribeAddonOutput{
							Addon: &awseks.Addon{
								Tags: map[string]*string{
									testOtherTagKey: &testOtherTagValue,
								},
							},
						}, nil)
					me.EXPECT().
						TagResourceWithContext(
							context.Background(),
							&awseks.TagResourceInput{
								ResourceArn: &testExternalName,
								Tags: map[string]*string{
									testTagKey: &testTagValue,
								},
							}).
						Return(&awseks.TagResourceOutput{}, nil)
					me.EXPECT().
						UntagResourceWithContext(
							context.Background(),
							&awseks.UntagResourceInput{
								ResourceArn: &testExternalName,
								TagKeys:     []*string{&testOtherTagKey},
							}).
						Return(nil, errBoom)
				}),
				cr: addon(
					withExternalName(testExternalName),
					withSpec(v1alpha1.AddonParameters{
						AddonName:             &testAddonName,
						AddonVersion:          &testAddonVersion,
						ResolveConflicts:      &testResolveConflict,
						ServiceAccountRoleARN: &testServiceAccountRoleArn,
						CustomAddonParameters: v1alpha1.CustomAddonParameters{
							ClusterName: &testClusterName,
						},
						Tags: map[string]*string{
							testTagKey: &testTagValue,
						},
					}),
					withStatus(
						v1alpha1.AddonObservation{AddonARN: &testExternalName},
					),
				),
			},
			want: want{
				cr: addon(
					withExternalName(testExternalName),
					withSpec(v1alpha1.AddonParameters{
						AddonName:             &testAddonName,
						AddonVersion:          &testAddonVersion,
						ResolveConflicts:      &testResolveConflict,
						ServiceAccountRoleARN: &testServiceAccountRoleArn,
						CustomAddonParameters: v1alpha1.CustomAddonParameters{
							ClusterName: &testClusterName,
						},
						Tags: map[string]*string{
							testTagKey: &testTagValue,
						},
					}),
					withStatus(
						v1alpha1.AddonObservation{AddonARN: &testExternalName},
					),
				),
				err: awsclient.Wrap(errBoom, errUntagResource),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := newExternal(nil, tc.eks(t), []option{setupHooks})
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
		cr  *v1alpha1.Addon
		err error
	}

	cases := map[string]struct {
		args
		want
	}{
		"Successful": {
			args: args{
				eks: mockClient(func(me *mockeksiface.MockEKSAPI) {
					me.EXPECT().
						DeleteAddonWithContext(
							context.Background(),
							&awseks.DeleteAddonInput{
								AddonName:   &testAddonName,
								ClusterName: &testClusterName,
							},
						).
						Return(&awseks.DeleteAddonOutput{}, nil)
				}),
				cr: addon(
					withSpec(v1alpha1.AddonParameters{
						AddonName: &testAddonName,
						CustomAddonParameters: v1alpha1.CustomAddonParameters{
							ClusterName: &testClusterName,
						},
					}),
				),
			},
			want: want{
				cr: addon(
					withSpec(v1alpha1.AddonParameters{
						AddonName: &testAddonName,
						CustomAddonParameters: v1alpha1.CustomAddonParameters{
							ClusterName: &testClusterName,
						},
					}),
					withConditions(xpv1.Deleting()),
				),
			},
		},
		"FailedRequest": {
			args: args{
				eks: mockClient(func(me *mockeksiface.MockEKSAPI) {
					me.EXPECT().
						DeleteAddonWithContext(
							context.Background(),
							&awseks.DeleteAddonInput{
								AddonName:   &testAddonName,
								ClusterName: &testClusterName,
							},
						).
						Return(nil, errBoom)
				}),
				cr: addon(
					withSpec(v1alpha1.AddonParameters{
						AddonName: &testAddonName,
						CustomAddonParameters: v1alpha1.CustomAddonParameters{
							ClusterName: &testClusterName,
						},
					}),
				),
			},
			want: want{
				cr: addon(
					withSpec(v1alpha1.AddonParameters{
						AddonName: &testAddonName,
						CustomAddonParameters: v1alpha1.CustomAddonParameters{
							ClusterName: &testClusterName,
						},
					}),
					withConditions(xpv1.Deleting()),
				),
				err: awsclient.Wrap(errBoom, errDelete),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := newExternal(nil, tc.eks(t), []option{setupHooks})
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
