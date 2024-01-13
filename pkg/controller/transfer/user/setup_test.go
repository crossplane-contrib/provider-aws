/*
Copyright 2023 The Crossplane Authors.

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
	"time"

	svcsdk "github.com/aws/aws-sdk-go/service/transfer"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
	"k8s.io/utils/ptr"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/transfer/v1alpha1"
	transfermock "github.com/crossplane-contrib/provider-aws/pkg/clients/mock/transferiface"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
)

var (
	timeNow     = time.Now()
	timeNowMeta = pointer.TimeToMetaTime(&timeNow)
)

type userModifier func(cr *svcapitypes.User)

func user(mods ...userModifier) *svcapitypes.User {
	cr := &svcapitypes.User{}
	for _, m := range mods {
		m(cr)
	}
	return cr
}

func withExternalName(s string) userModifier {
	return func(cr *svcapitypes.User) { meta.SetExternalName(cr, s) }
}

func withSpec(s svcapitypes.UserParameters) userModifier {
	return func(cr *svcapitypes.User) { cr.Spec.ForProvider = s }
}

func withStatus(s svcapitypes.UserObservation) userModifier {
	return func(cr *svcapitypes.User) { cr.Status.AtProvider = s }
}

func withConditions(s ...xpv1.Condition) userModifier {
	return func(cr *svcapitypes.User) { cr.SetConditions(s...) }
}

type transferClientMockModifier func(m *transfermock.MockTransferAPI)

func TestObserve(t *testing.T) {
	type args struct {
		cr       *svcapitypes.User
		transfer transferClientMockModifier
	}

	type want struct {
		cr     *svcapitypes.User
		result managed.ExternalObservation
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"SuccessfulAvailable": {
			args: args{
				cr: user(
					withExternalName("test"),
					withSpec(svcapitypes.UserParameters{
						Region:        "us-east-1",
						HomeDirectory: ptr.To("/test"),
						HomeDirectoryMappings: []*svcapitypes.HomeDirectoryMapEntry{
							{
								Entry:  ptr.To("entry"),
								Target: ptr.To("target"),
							},
						},
						HomeDirectoryType: ptr.To("LOGICAL"),
						PosixProfile: &svcapitypes.PosixProfile{
							Gid: ptr.To(int64(1000)),
							SecondaryGids: []*int64{
								ptr.To(int64(1001)),
								ptr.To(int64(1002)),
							},
							Uid: ptr.To(int64(1005)),
						},
						Tags: []*svcapitypes.Tag{
							{
								Key:   ptr.To("key"),
								Value: ptr.To("value"),
							},
						},
						CustomUserParameters: svcapitypes.CustomUserParameters{
							ServerID: ptr.To("server"),
							Role:     ptr.To("role"),
							SSHPublicKeys: []svcapitypes.SSHPublicKeySpec{
								{
									Body: "some-public-key",
								},
							},
						},
					}),
				),
				transfer: func(m *transfermock.MockTransferAPI) {
					m.EXPECT().DescribeUserWithContext(context.Background(), &svcsdk.DescribeUserInput{
						ServerId: ptr.To("server"),
						UserName: ptr.To("test"),
					}).Return(&svcsdk.DescribeUserOutput{
						ServerId: ptr.To("server"),
						User: &svcsdk.DescribedUser{
							Arn:           ptr.To("ARN"),
							HomeDirectory: ptr.To("/test"),
							HomeDirectoryMappings: []*svcsdk.HomeDirectoryMapEntry{
								{
									Entry:  ptr.To("entry"),
									Target: ptr.To("target"),
								},
							},
							HomeDirectoryType: ptr.To("LOGICAL"),
							PosixProfile: &svcsdk.PosixProfile{
								Gid: ptr.To(int64(1000)),
								SecondaryGids: []*int64{
									ptr.To(int64(1001)),
									ptr.To(int64(1002)),
								},
								Uid: ptr.To(int64(1005)),
							},
							Role: ptr.To("role"),
							SshPublicKeys: []*svcsdk.SshPublicKey{
								{
									DateImported:     &timeNow,
									SshPublicKeyBody: ptr.To("some-public-key"),
									SshPublicKeyId:   ptr.To("key-id"),
								},
							},
							Tags: []*svcsdk.Tag{
								{
									Key:   ptr.To("key"),
									Value: ptr.To("value"),
								},
							},
							UserName: ptr.To("test"),
						},
					}, nil)
				},
			},
			want: want{
				cr: user(
					withExternalName("test"),
					withSpec(svcapitypes.UserParameters{
						Region:        "us-east-1",
						HomeDirectory: ptr.To("/test"),
						HomeDirectoryMappings: []*svcapitypes.HomeDirectoryMapEntry{
							{
								Entry:  ptr.To("entry"),
								Target: ptr.To("target"),
							},
						},
						HomeDirectoryType: ptr.To("LOGICAL"),
						PosixProfile: &svcapitypes.PosixProfile{
							Gid: ptr.To(int64(1000)),
							SecondaryGids: []*int64{
								ptr.To(int64(1001)),
								ptr.To(int64(1002)),
							},
							Uid: ptr.To(int64(1005)),
						},
						Tags: []*svcapitypes.Tag{
							{
								Key:   ptr.To("key"),
								Value: ptr.To("value"),
							},
						},
						CustomUserParameters: svcapitypes.CustomUserParameters{
							ServerID: ptr.To("server"),
							Role:     ptr.To("role"),
							SSHPublicKeys: []svcapitypes.SSHPublicKeySpec{
								{
									Body: "some-public-key",
								},
							},
						},
					}),
					withStatus(svcapitypes.UserObservation{
						ServerID: ptr.To("server"),
						ARN:      ptr.To("ARN"),
						SshPublicKeys: []*svcapitypes.SshPublicKey{
							{
								DateImported:     timeNowMeta,
								SshPublicKeyBody: ptr.To("some-public-key"),
								SshPublicKeyID:   ptr.To("key-id"),
							},
						},
						UserName: ptr.To("test"),
					}),
					withConditions(xpv1.Available()),
				),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
				},
			},
		},
		"LateInitialize": {
			args: args{
				cr: user(
					withExternalName("test"),
					withSpec(svcapitypes.UserParameters{
						Region:        "us-east-1",
						HomeDirectory: ptr.To("/test"),
						HomeDirectoryMappings: []*svcapitypes.HomeDirectoryMapEntry{
							{
								Entry:  ptr.To("entry"),
								Target: ptr.To("target"),
							},
						},
						HomeDirectoryType: ptr.To("LOGICAL"),
						PosixProfile: &svcapitypes.PosixProfile{
							Gid: ptr.To(int64(1000)),
							SecondaryGids: []*int64{
								ptr.To(int64(1001)),
								ptr.To(int64(1002)),
							},
							Uid: ptr.To(int64(1005)),
						},
						Tags: []*svcapitypes.Tag{
							{
								Key:   ptr.To("key"),
								Value: ptr.To("value"),
							},
						},
						CustomUserParameters: svcapitypes.CustomUserParameters{
							ServerID: ptr.To("server"),
							Role:     ptr.To("role"),
						},
					}),
				),
				transfer: func(m *transfermock.MockTransferAPI) {
					m.EXPECT().DescribeUserWithContext(context.Background(), &svcsdk.DescribeUserInput{
						ServerId: ptr.To("server"),
						UserName: ptr.To("test"),
					}).Return(&svcsdk.DescribeUserOutput{
						ServerId: ptr.To("server"),
						User: &svcsdk.DescribedUser{
							Arn:           ptr.To("ARN"),
							HomeDirectory: ptr.To("/test"),
							HomeDirectoryMappings: []*svcsdk.HomeDirectoryMapEntry{
								{
									Entry:  ptr.To("entry"),
									Target: ptr.To("target"),
								},
							},
							HomeDirectoryType: ptr.To("LOGICAL"),
							PosixProfile: &svcsdk.PosixProfile{
								Gid: ptr.To(int64(1000)),
								SecondaryGids: []*int64{
									ptr.To(int64(1001)),
									ptr.To(int64(1002)),
								},
								Uid: ptr.To(int64(1005)),
							},
							Role: ptr.To("role"),
							SshPublicKeys: []*svcsdk.SshPublicKey{
								{
									DateImported:     &timeNow,
									SshPublicKeyBody: ptr.To("some-public-key"),
									SshPublicKeyId:   ptr.To("key-id"),
								},
							},
							Tags: []*svcsdk.Tag{
								{
									Key:   ptr.To("key"),
									Value: ptr.To("value"),
								},
							},
							UserName: ptr.To("test"),
						},
					}, nil)
				},
			},
			want: want{
				cr: user(
					withExternalName("test"),
					withSpec(svcapitypes.UserParameters{
						Region:        "us-east-1",
						HomeDirectory: ptr.To("/test"),
						HomeDirectoryMappings: []*svcapitypes.HomeDirectoryMapEntry{
							{
								Entry:  ptr.To("entry"),
								Target: ptr.To("target"),
							},
						},
						HomeDirectoryType: ptr.To("LOGICAL"),
						PosixProfile: &svcapitypes.PosixProfile{
							Gid: ptr.To(int64(1000)),
							SecondaryGids: []*int64{
								ptr.To(int64(1001)),
								ptr.To(int64(1002)),
							},
							Uid: ptr.To(int64(1005)),
						},
						Tags: []*svcapitypes.Tag{
							{
								Key:   ptr.To("key"),
								Value: ptr.To("value"),
							},
						},
						CustomUserParameters: svcapitypes.CustomUserParameters{
							ServerID: ptr.To("server"),
							Role:     ptr.To("role"),
							SSHPublicKeys: []svcapitypes.SSHPublicKeySpec{
								{
									Body: "some-public-key",
								},
							},
						},
					}),
					withStatus(svcapitypes.UserObservation{
						ServerID: ptr.To("server"),
						ARN:      ptr.To("ARN"),
						SshPublicKeys: []*svcapitypes.SshPublicKey{
							{
								DateImported:     timeNowMeta,
								SshPublicKeyBody: ptr.To("some-public-key"),
								SshPublicKeyID:   ptr.To("key-id"),
							},
						},
						UserName: ptr.To("test"),
					}),
					withConditions(xpv1.Available()),
				),
				result: managed.ExternalObservation{
					ResourceExists:          true,
					ResourceUpToDate:        true,
					ResourceLateInitialized: true,
				},
			},
		},
		// TODO: Remove this test once the spec.forProvider.SSHPublicKeyBody is removed
		"EnsureBackwardsCompatibility": {
			args: args{
				cr: user(
					withExternalName("test"),
					withSpec(svcapitypes.UserParameters{
						Region:        "us-east-1",
						HomeDirectory: ptr.To("/test"),
						HomeDirectoryMappings: []*svcapitypes.HomeDirectoryMapEntry{
							{
								Entry:  ptr.To("entry"),
								Target: ptr.To("target"),
							},
						},
						HomeDirectoryType: ptr.To("LOGICAL"),
						PosixProfile: &svcapitypes.PosixProfile{
							Gid: ptr.To(int64(1000)),
							SecondaryGids: []*int64{
								ptr.To(int64(1001)),
								ptr.To(int64(1002)),
							},
							Uid: ptr.To(int64(1005)),
						},
						Tags: []*svcapitypes.Tag{
							{
								Key:   ptr.To("key"),
								Value: ptr.To("value"),
							},
						},
						CustomUserParameters: svcapitypes.CustomUserParameters{
							ServerID:         ptr.To("server"),
							Role:             ptr.To("role"),
							SshPublicKeyBody: ptr.To("some-public-key"),
						},
					}),
				),
				transfer: func(m *transfermock.MockTransferAPI) {
					m.EXPECT().DescribeUserWithContext(context.Background(), &svcsdk.DescribeUserInput{
						ServerId: ptr.To("server"),
						UserName: ptr.To("test"),
					}).Return(&svcsdk.DescribeUserOutput{
						ServerId: ptr.To("server"),
						User: &svcsdk.DescribedUser{
							Arn:           ptr.To("ARN"),
							HomeDirectory: ptr.To("/test"),
							HomeDirectoryMappings: []*svcsdk.HomeDirectoryMapEntry{
								{
									Entry:  ptr.To("entry"),
									Target: ptr.To("target"),
								},
							},
							HomeDirectoryType: ptr.To("LOGICAL"),
							PosixProfile: &svcsdk.PosixProfile{
								Gid: ptr.To(int64(1000)),
								SecondaryGids: []*int64{
									ptr.To(int64(1001)),
									ptr.To(int64(1002)),
								},
								Uid: ptr.To(int64(1005)),
							},
							Role: ptr.To("role"),
							SshPublicKeys: []*svcsdk.SshPublicKey{
								{
									DateImported:     &timeNow,
									SshPublicKeyBody: ptr.To("some-public-key"),
									SshPublicKeyId:   ptr.To("key-id"),
								},
							},
							Tags: []*svcsdk.Tag{
								{
									Key:   ptr.To("key"),
									Value: ptr.To("value"),
								},
							},
							UserName: ptr.To("test"),
						},
					}, nil)
				},
			},
			want: want{
				cr: user(
					withExternalName("test"),
					withSpec(svcapitypes.UserParameters{
						Region:        "us-east-1",
						HomeDirectory: ptr.To("/test"),
						HomeDirectoryMappings: []*svcapitypes.HomeDirectoryMapEntry{
							{
								Entry:  ptr.To("entry"),
								Target: ptr.To("target"),
							},
						},
						HomeDirectoryType: ptr.To("LOGICAL"),
						PosixProfile: &svcapitypes.PosixProfile{
							Gid: ptr.To(int64(1000)),
							SecondaryGids: []*int64{
								ptr.To(int64(1001)),
								ptr.To(int64(1002)),
							},
							Uid: ptr.To(int64(1005)),
						},
						Tags: []*svcapitypes.Tag{
							{
								Key:   ptr.To("key"),
								Value: ptr.To("value"),
							},
						},
						CustomUserParameters: svcapitypes.CustomUserParameters{
							ServerID:         ptr.To("server"),
							Role:             ptr.To("role"),
							SshPublicKeyBody: ptr.To("some-public-key"),
							SSHPublicKeys: []svcapitypes.SSHPublicKeySpec{
								{
									Body: "some-public-key",
								},
							},
						},
					}),
					withStatus(svcapitypes.UserObservation{
						ServerID: ptr.To("server"),
						ARN:      ptr.To("ARN"),
						SshPublicKeys: []*svcapitypes.SshPublicKey{
							{
								DateImported:     timeNowMeta,
								SshPublicKeyBody: ptr.To("some-public-key"),
								SshPublicKeyID:   ptr.To("key-id"),
							},
						},
						UserName: ptr.To("test"),
					}),
					withConditions(xpv1.Available()),
				),
				result: managed.ExternalObservation{
					ResourceExists:          true,
					ResourceUpToDate:        true,
					ResourceLateInitialized: true,
				},
			},
		},
		"DifferentKeys": {
			args: args{
				cr: user(
					withExternalName("test"),
					withSpec(svcapitypes.UserParameters{
						CustomUserParameters: svcapitypes.CustomUserParameters{
							ServerID: ptr.To("server"),
							SSHPublicKeys: []svcapitypes.SSHPublicKeySpec{
								{
									Body: "key-to-add",
								},
								{
									Body: "key-to-keep",
								},
							},
						},
					}),
				),
				transfer: func(m *transfermock.MockTransferAPI) {
					m.EXPECT().DescribeUserWithContext(context.Background(), &svcsdk.DescribeUserInput{
						ServerId: ptr.To("server"),
						UserName: ptr.To("test"),
					}).Return(&svcsdk.DescribeUserOutput{
						ServerId: ptr.To("server"),
						User: &svcsdk.DescribedUser{
							SshPublicKeys: []*svcsdk.SshPublicKey{
								{
									DateImported:     &timeNow,
									SshPublicKeyBody: ptr.To("key-to-keep"),
									SshPublicKeyId:   ptr.To("key-to-keep-id"),
								},
								{
									DateImported:     &timeNow,
									SshPublicKeyBody: ptr.To("key-to-remove"),
									SshPublicKeyId:   ptr.To("key-to-remove-id"),
								},
							},
						},
					}, nil)
				},
			},
			want: want{
				cr: user(
					withExternalName("test"),
					withSpec(svcapitypes.UserParameters{
						CustomUserParameters: svcapitypes.CustomUserParameters{
							ServerID: ptr.To("server"),
							SSHPublicKeys: []svcapitypes.SSHPublicKeySpec{
								{
									Body: "key-to-add",
								},
								{
									Body: "key-to-keep",
								},
							},
						},
					}),
					withStatus(svcapitypes.UserObservation{
						ServerID: ptr.To("server"),
						SshPublicKeys: []*svcapitypes.SshPublicKey{
							{
								DateImported:     timeNowMeta,
								SshPublicKeyBody: ptr.To("key-to-keep"),
								SshPublicKeyID:   ptr.To("key-to-keep-id"),
							},
							{
								DateImported:     timeNowMeta,
								SshPublicKeyBody: ptr.To("key-to-remove"),
								SshPublicKeyID:   ptr.To("key-to-remove-id"),
							},
						},
					}),
					withConditions(xpv1.Available()),
				),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: false,
				},
			},
		},
		"DifferentTags": {
			args: args{
				cr: user(
					withExternalName("test"),
					withSpec(svcapitypes.UserParameters{
						CustomUserParameters: svcapitypes.CustomUserParameters{
							ServerID: ptr.To("server"),
						},
						Tags: []*svcapitypes.Tag{
							{
								Key:   ptr.To("keep"),
								Value: ptr.To("bar"),
							},
							{
								Key:   ptr.To("to-remove"),
								Value: ptr.To("val"),
							},
						},
					}),
				),
				transfer: func(m *transfermock.MockTransferAPI) {
					m.EXPECT().DescribeUserWithContext(context.Background(), &svcsdk.DescribeUserInput{
						ServerId: ptr.To("server"),
						UserName: ptr.To("test"),
					}).Return(&svcsdk.DescribeUserOutput{
						ServerId: ptr.To("server"),
						User: &svcsdk.DescribedUser{
							Tags: []*svcsdk.Tag{
								{
									Key:   ptr.To("keep"),
									Value: ptr.To("bar"),
								},
								{
									Key:   ptr.To("to-add"),
									Value: ptr.To("world"),
								},
							},
						},
					}, nil)
				},
			},
			want: want{
				cr: user(
					withExternalName("test"),
					withSpec(svcapitypes.UserParameters{
						CustomUserParameters: svcapitypes.CustomUserParameters{
							ServerID: ptr.To("server"),
						},
						Tags: []*svcapitypes.Tag{
							{
								Key:   ptr.To("keep"),
								Value: ptr.To("bar"),
							},
							{
								Key:   ptr.To("to-remove"),
								Value: ptr.To("val"),
							},
						},
					}),
					withStatus(svcapitypes.UserObservation{
						ServerID: ptr.To("server"),
					}),
					withConditions(xpv1.Available()),
				),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: false,
				},
			},
		},
		"SameTagsDifferentOrder": {
			args: args{
				cr: user(
					withExternalName("test"),
					withSpec(svcapitypes.UserParameters{
						CustomUserParameters: svcapitypes.CustomUserParameters{
							ServerID: ptr.To("server"),
						},
						Tags: []*svcapitypes.Tag{
							{
								Key:   ptr.To("foo"),
								Value: ptr.To("bar"),
							},
							{
								Key:   ptr.To("hello"),
								Value: ptr.To("world"),
							},
						},
					}),
				),
				transfer: func(m *transfermock.MockTransferAPI) {
					m.EXPECT().DescribeUserWithContext(context.Background(), &svcsdk.DescribeUserInput{
						ServerId: ptr.To("server"),
						UserName: ptr.To("test"),
					}).Return(&svcsdk.DescribeUserOutput{
						ServerId: ptr.To("server"),
						User: &svcsdk.DescribedUser{
							Tags: []*svcsdk.Tag{
								{
									Key:   ptr.To("hello"),
									Value: ptr.To("world"),
								},
								{
									Key:   ptr.To("foo"),
									Value: ptr.To("bar"),
								},
							},
						},
					}, nil)
				},
			},
			want: want{
				cr: user(
					withExternalName("test"),
					withSpec(svcapitypes.UserParameters{
						CustomUserParameters: svcapitypes.CustomUserParameters{
							ServerID: ptr.To("server"),
						},
						Tags: []*svcapitypes.Tag{
							{
								Key:   ptr.To("foo"),
								Value: ptr.To("bar"),
							},
							{
								Key:   ptr.To("hello"),
								Value: ptr.To("world"),
							},
						},
					}),
					withStatus(svcapitypes.UserObservation{
						ServerID: ptr.To("server"),
					}),
					withConditions(xpv1.Available()),
				),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			transferClient := transfermock.NewMockTransferAPI(gomock.NewController(t))
			if tc.args.transfer != nil {
				tc.args.transfer(transferClient)
			}

			e := newExternal(nil, transferClient, []option{setupHooks()})
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
	type args struct {
		cr       *svcapitypes.User
		transfer transferClientMockModifier
	}

	type want struct {
		cr     *svcapitypes.User
		result managed.ExternalCreation
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"Success": {
			args: args{
				cr: user(
					withExternalName("test"),
					withSpec(svcapitypes.UserParameters{
						Region:        "us-east-1",
						HomeDirectory: ptr.To("/test"),
						HomeDirectoryMappings: []*svcapitypes.HomeDirectoryMapEntry{
							{
								Entry:  ptr.To("entry"),
								Target: ptr.To("target"),
							},
						},
						HomeDirectoryType: ptr.To("LOGICAL"),
						PosixProfile: &svcapitypes.PosixProfile{
							Gid: ptr.To(int64(1000)),
							SecondaryGids: []*int64{
								ptr.To(int64(1001)),
								ptr.To(int64(1002)),
							},
							Uid: ptr.To(int64(1005)),
						},
						Tags: []*svcapitypes.Tag{
							{
								Key:   ptr.To("key"),
								Value: ptr.To("value"),
							},
						},
						CustomUserParameters: svcapitypes.CustomUserParameters{
							ServerID: ptr.To("server"),
							Role:     ptr.To("role"),
							SSHPublicKeys: []svcapitypes.SSHPublicKeySpec{
								{
									Body: "some-public-key",
								},
							},
						},
					}),
				),
				transfer: func(m *transfermock.MockTransferAPI) {
					m.EXPECT().CreateUserWithContext(context.Background(), &svcsdk.CreateUserInput{
						ServerId:      ptr.To("server"),
						HomeDirectory: ptr.To("/test"),
						HomeDirectoryMappings: []*svcsdk.HomeDirectoryMapEntry{
							{
								Entry:  ptr.To("entry"),
								Target: ptr.To("target"),
							},
						},
						HomeDirectoryType: ptr.To("LOGICAL"),
						PosixProfile: &svcsdk.PosixProfile{
							Gid: ptr.To(int64(1000)),
							SecondaryGids: []*int64{
								ptr.To(int64(1001)),
								ptr.To(int64(1002)),
							},
							Uid: ptr.To(int64(1005)),
						},
						Tags: []*svcsdk.Tag{
							{
								Key:   ptr.To("key"),
								Value: ptr.To("value"),
							},
						},
						Role:     ptr.To("role"),
						UserName: ptr.To("test"),
					}).Return(&svcsdk.CreateUserOutput{}, nil)
				},
			},
			want: want{
				cr: user(
					withExternalName("test"),
					withSpec(svcapitypes.UserParameters{
						Region:        "us-east-1",
						HomeDirectory: ptr.To("/test"),
						HomeDirectoryMappings: []*svcapitypes.HomeDirectoryMapEntry{
							{
								Entry:  ptr.To("entry"),
								Target: ptr.To("target"),
							},
						},
						HomeDirectoryType: ptr.To("LOGICAL"),
						PosixProfile: &svcapitypes.PosixProfile{
							Gid: ptr.To(int64(1000)),
							SecondaryGids: []*int64{
								ptr.To(int64(1001)),
								ptr.To(int64(1002)),
							},
							Uid: ptr.To(int64(1005)),
						},
						Tags: []*svcapitypes.Tag{
							{
								Key:   ptr.To("key"),
								Value: ptr.To("value"),
							},
						},
						CustomUserParameters: svcapitypes.CustomUserParameters{
							ServerID: ptr.To("server"),
							Role:     ptr.To("role"),
							SSHPublicKeys: []svcapitypes.SSHPublicKeySpec{
								{
									Body: "some-public-key",
								},
							},
						},
					}),
					withConditions(xpv1.Creating()),
				),
				result: managed.ExternalCreation{},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			transferClient := transfermock.NewMockTransferAPI(gomock.NewController(t))
			if tc.args.transfer != nil {
				tc.args.transfer(transferClient)
			}

			e := newExternal(nil, transferClient, []option{setupHooks()})
			u, err := e.Create(context.Background(), tc.args.cr)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.cr, tc.args.cr, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.result, u); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestUpdate(t *testing.T) {
	type args struct {
		cr       *svcapitypes.User
		transfer transferClientMockModifier

		keyBodiesToImport []string
		keyIDsToDelete    []string

		tagsToAdd    []*svcsdk.Tag
		tagsToDelete []*string
	}

	type want struct {
		cr     *svcapitypes.User
		result managed.ExternalUpdate
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"Success": {
			args: args{
				cr: user(
					withExternalName("test"),
					withSpec(svcapitypes.UserParameters{
						Region:        "us-east-1",
						HomeDirectory: ptr.To("/test"),
						HomeDirectoryMappings: []*svcapitypes.HomeDirectoryMapEntry{
							{
								Entry:  ptr.To("entry"),
								Target: ptr.To("target"),
							},
						},
						HomeDirectoryType: ptr.To("LOGICAL"),
						PosixProfile: &svcapitypes.PosixProfile{
							Gid: ptr.To(int64(1000)),
							SecondaryGids: []*int64{
								ptr.To(int64(1001)),
								ptr.To(int64(1002)),
							},
							Uid: ptr.To(int64(1005)),
						},
						Tags: []*svcapitypes.Tag{
							{
								Key:   ptr.To("key"),
								Value: ptr.To("value"),
							},
						},
						CustomUserParameters: svcapitypes.CustomUserParameters{
							ServerID: ptr.To("server"),
							Role:     ptr.To("role"),
							SSHPublicKeys: []svcapitypes.SSHPublicKeySpec{
								{
									Body: "some-public-key",
								},
							},
						},
					}),
					withStatus(svcapitypes.UserObservation{
						ARN: ptr.To("ARN"),
					}),
				),
				keyBodiesToImport: []string{
					"key-to-add",
				},
				keyIDsToDelete: []string{
					"key-id-to-remove",
				},
				tagsToAdd: []*svcsdk.Tag{
					{
						Key:   ptr.To("foo"),
						Value: ptr.To("bar"),
					},
				},
				tagsToDelete: []*string{
					ptr.To("tag-to-remove"),
				},
				transfer: func(m *transfermock.MockTransferAPI) {
					m.EXPECT().UpdateUserWithContext(context.Background(), &svcsdk.UpdateUserInput{
						HomeDirectory: ptr.To("/test"),
						HomeDirectoryMappings: []*svcsdk.HomeDirectoryMapEntry{
							{
								Entry:  ptr.To("entry"),
								Target: ptr.To("target"),
							},
						},
						HomeDirectoryType: ptr.To("LOGICAL"),
						PosixProfile: &svcsdk.PosixProfile{
							Gid: ptr.To(int64(1000)),
							SecondaryGids: []*int64{
								ptr.To(int64(1001)),
								ptr.To(int64(1002)),
							},
							Uid: ptr.To(int64(1005)),
						},
					}).Return(&svcsdk.UpdateUserOutput{}, nil)
					m.EXPECT().ImportSshPublicKeyWithContext(context.Background(), &svcsdk.ImportSshPublicKeyInput{
						ServerId:         ptr.To("server"),
						UserName:         ptr.To("test"),
						SshPublicKeyBody: ptr.To("key-to-add"),
					}).Return(nil, nil)
					m.EXPECT().DeleteSshPublicKeyWithContext(context.Background(), &svcsdk.DeleteSshPublicKeyInput{
						ServerId:       ptr.To("server"),
						UserName:       ptr.To("test"),
						SshPublicKeyId: ptr.To("key-id-to-remove"),
					}).Return(nil, nil)
					m.EXPECT().TagResourceWithContext(context.Background(), &svcsdk.TagResourceInput{
						Arn: ptr.To("ARN"),
						Tags: []*svcsdk.Tag{
							{
								Key:   ptr.To("foo"),
								Value: ptr.To("bar"),
							},
						},
					}).Return(nil, nil)
					m.EXPECT().UntagResourceWithContext(context.Background(), &svcsdk.UntagResourceInput{
						Arn: ptr.To("ARN"),
						TagKeys: []*string{
							ptr.To("tag-to-remove"),
						},
					}).Return(nil, nil)
				},
			},
			want: want{
				cr: user(
					withExternalName("test"),
					withSpec(svcapitypes.UserParameters{
						Region:        "us-east-1",
						HomeDirectory: ptr.To("/test"),
						HomeDirectoryMappings: []*svcapitypes.HomeDirectoryMapEntry{
							{
								Entry:  ptr.To("entry"),
								Target: ptr.To("target"),
							},
						},
						HomeDirectoryType: ptr.To("LOGICAL"),
						PosixProfile: &svcapitypes.PosixProfile{
							Gid: ptr.To(int64(1000)),
							SecondaryGids: []*int64{
								ptr.To(int64(1001)),
								ptr.To(int64(1002)),
							},
							Uid: ptr.To(int64(1005)),
						},
						Tags: []*svcapitypes.Tag{
							{
								Key:   ptr.To("key"),
								Value: ptr.To("value"),
							},
						},
						CustomUserParameters: svcapitypes.CustomUserParameters{
							ServerID: ptr.To("server"),
							Role:     ptr.To("role"),
							SSHPublicKeys: []svcapitypes.SSHPublicKeySpec{
								{
									Body: "some-public-key",
								},
							},
						},
					}),
					withStatus(svcapitypes.UserObservation{
						ARN: ptr.To("ARN"),
					}),
				),
				result: managed.ExternalUpdate{},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			transferClient := transfermock.NewMockTransferAPI(gomock.NewController(t))
			if tc.args.transfer != nil {
				tc.args.transfer(transferClient)
			}

			// Need to create hooks manually here to fill the cache
			e := newExternal(nil, transferClient, []option{func(e *external) {
				h := hooks{client: e.client}
				h.cache.keyBodiesToImport = tc.args.keyBodiesToImport
				h.cache.keyIDsToDelete = tc.args.keyIDsToDelete
				h.cache.tagsToAdd = tc.args.tagsToAdd
				h.cache.tagsToDelete = tc.args.tagsToDelete
				e.postUpdate = h.postUpdate
			}})
			u, err := e.Update(context.Background(), tc.args.cr)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.cr, tc.args.cr, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.result, u); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	type args struct {
		cr       *svcapitypes.User
		transfer transferClientMockModifier
	}

	type want struct {
		cr  *svcapitypes.User
		err error
	}

	cases := map[string]struct {
		args
		want
	}{
		"Success": {
			args: args{
				cr: user(
					withExternalName("test"),
					withSpec(svcapitypes.UserParameters{
						CustomUserParameters: svcapitypes.CustomUserParameters{
							ServerID: ptr.To("server"),
						},
					}),
				),
				transfer: func(m *transfermock.MockTransferAPI) {
					m.EXPECT().DeleteUserWithContext(context.Background(), &svcsdk.DeleteUserInput{
						ServerId: ptr.To("server"),
						UserName: ptr.To("test"),
					}).Return(&svcsdk.DeleteUserOutput{}, nil)
				},
			},
			want: want{
				cr: user(
					withExternalName("test"),
					withSpec(svcapitypes.UserParameters{
						CustomUserParameters: svcapitypes.CustomUserParameters{
							ServerID: ptr.To("server"),
						},
					}),
					withConditions(xpv1.Deleting()),
				),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			transferClient := transfermock.NewMockTransferAPI(gomock.NewController(t))
			if tc.args.transfer != nil {
				tc.args.transfer(transferClient)
			}

			e := newExternal(nil, transferClient, []option{setupHooks()})
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

func TestIsSSHPublicKeysUpToDate(t *testing.T) {
	type args struct {
		cr  *svcapitypes.User
		obj *svcsdk.DescribeUserOutput
	}

	type want struct {
		isUpToDate        bool
		keyBodiesToImport []string
		keyIDsToDelete    []string
	}

	cases := map[string]struct {
		args
		want
	}{
		"AddAndRemoveKeys": {
			args: args{
				cr: user(
					withExternalName("test"),
					withSpec(svcapitypes.UserParameters{
						CustomUserParameters: svcapitypes.CustomUserParameters{
							ServerID: ptr.To("server"),
							SSHPublicKeys: []svcapitypes.SSHPublicKeySpec{
								{
									Body: "key-to-add-body",
								},
								{
									Body: "key-to-keep-body",
								},
							},
						},
					}),
				),
				obj: &svcsdk.DescribeUserOutput{
					User: &svcsdk.DescribedUser{
						SshPublicKeys: []*svcsdk.SshPublicKey{
							{
								DateImported:     &timeNow,
								SshPublicKeyBody: ptr.To("key-to-keep-body"),
								SshPublicKeyId:   ptr.To("key-to-keep"),
							},
							{
								DateImported:     &timeNow,
								SshPublicKeyBody: ptr.To("key-to-remove-body"),
								SshPublicKeyId:   ptr.To("key-to-remove"),
							},
						},
					},
				},
			},
			want: want{
				isUpToDate:        false,
				keyBodiesToImport: []string{"key-to-add-body"},
				keyIDsToDelete:    []string{"key-to-remove"},
			},
		},
		"SameKeysDifferentOrder": {
			args: args{
				cr: user(
					withExternalName("test"),
					withSpec(svcapitypes.UserParameters{
						CustomUserParameters: svcapitypes.CustomUserParameters{
							ServerID: ptr.To("server"),
							SSHPublicKeys: []svcapitypes.SSHPublicKeySpec{
								{
									Body: "key-1-body",
								},
								{
									Body: "key-2-body",
								},
							},
						},
					}),
				),
				obj: &svcsdk.DescribeUserOutput{
					User: &svcsdk.DescribedUser{
						SshPublicKeys: []*svcsdk.SshPublicKey{
							{
								DateImported:     &timeNow,
								SshPublicKeyBody: ptr.To("key-2-body"),
								SshPublicKeyId:   ptr.To("key-2"),
							},
							{
								DateImported:     &timeNow,
								SshPublicKeyBody: ptr.To("key-1-body"),
								SshPublicKeyId:   ptr.To("key-1"),
							},
						},
					},
				},
			},
			want: want{
				isUpToDate:        true,
				keyBodiesToImport: []string{},
				keyIDsToDelete:    []string{},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			isUpToDate, toAdd, toRemove := isSSHPublicKeysUpToDate(tc.args.cr, tc.args.obj)

			if diff := cmp.Diff(tc.want.isUpToDate, isUpToDate, test.EquateErrors()); diff != "" {
				t.Errorf("isUpToDate: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.keyBodiesToImport, toAdd, test.EquateErrors()); diff != "" {
				t.Errorf("keyBodiesToImport: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.keyIDsToDelete, toRemove, test.EquateErrors()); diff != "" {
				t.Errorf("keyIDsToDelete: -want, +got:\n%s", diff)
			}
		})
	}
}
