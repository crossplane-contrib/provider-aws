package identityprovider

import (
	"context"
	"testing"

	svcsdk "github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	v1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/cognitoidentityprovider/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/cognitoidentityprovider"
	mockclient "github.com/crossplane-contrib/provider-aws/pkg/clients/mock/cognitoidentityprovider"
)

type functionModifier func(*svcapitypes.IdentityProvider)

func withSpec(p svcapitypes.IdentityProviderParameters) functionModifier {
	return func(r *svcapitypes.IdentityProvider) { r.Spec.ForProvider = p }
}

func identityProvider(m ...functionModifier) *svcapitypes.IdentityProvider {
	cr := &svcapitypes.IdentityProvider{}
	cr.Name = "test-identityprovider-name"
	for _, f := range m {
		f(cr)
	}
	return cr
}

type mockModifier func(*mockclient.MockResolverService)

func withMockResolver(t *testing.T, mod mockModifier) *mockclient.MockResolverService {
	ctrl := gomock.NewController(t)
	mock := mockclient.NewMockResolverService(ctrl)
	mod(mock)
	return mock
}

var (
	testString1 string = "string1"
	testString2 string = "string2"
	errBoom     error  = errors.New("boom")
)

func TestIsUpToDate(t *testing.T) {
	type args struct {
		cr       *svcapitypes.IdentityProvider
		resp     *svcsdk.DescribeIdentityProviderOutput
		resolver cognitoidentityprovider.ResolverService
	}

	type want struct {
		result bool
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"UpToDateProviderDetails": {
			args: args{
				cr: identityProvider(withSpec(svcapitypes.IdentityProviderParameters{
					CustomIdentityProviderParameters: svcapitypes.CustomIdentityProviderParameters{
						ProviderDetailsSecretRef: v1.SecretReference{
							Name:      testString1,
							Namespace: testString2,
						},
					},
				})),
				resp: &svcsdk.DescribeIdentityProviderOutput{IdentityProvider: &svcsdk.IdentityProviderType{
					ProviderDetails: map[string]*string{testString1: &testString2},
				}},
				resolver: withMockResolver(t, func(mcs *mockclient.MockResolverService) {
					mcs.EXPECT().GetProviderDetails(
						context.Background(),
						nil,
						&v1.SecretReference{
							Name:      testString1,
							Namespace: testString2,
						},
					).Return(
						map[string]*string{testString1: &testString2},
						nil,
					)
				}),
			},
			want: want{
				result: true,
				err:    nil,
			},
		},
		"UpToDateProviderDetailsSecretRefContentChanged": {
			args: args{
				cr: identityProvider(withSpec(svcapitypes.IdentityProviderParameters{
					CustomIdentityProviderParameters: svcapitypes.CustomIdentityProviderParameters{
						ProviderDetailsSecretRef: v1.SecretReference{
							Name:      testString1,
							Namespace: testString2,
						},
					},
				})),
				resp: &svcsdk.DescribeIdentityProviderOutput{IdentityProvider: &svcsdk.IdentityProviderType{
					ProviderDetails: map[string]*string{testString1: &testString2},
				}},
				resolver: withMockResolver(t, func(mcs *mockclient.MockResolverService) {
					mcs.EXPECT().GetProviderDetails(
						context.Background(),
						nil,
						&v1.SecretReference{
							Name:      testString1,
							Namespace: testString2,
						},
					).Return(
						map[string]*string{testString1: &testString1},
						nil,
					)
				}),
			},
			want: want{
				result: false,
				err:    nil,
			},
		},
		"UpToDateProviderDetailsSecretRefRefChanged": {
			args: args{
				cr: identityProvider(withSpec(svcapitypes.IdentityProviderParameters{
					CustomIdentityProviderParameters: svcapitypes.CustomIdentityProviderParameters{
						ProviderDetailsSecretRef: v1.SecretReference{
							Name:      testString2,
							Namespace: testString2,
						},
					},
				})),
				resp: &svcsdk.DescribeIdentityProviderOutput{IdentityProvider: &svcsdk.IdentityProviderType{
					ProviderDetails: map[string]*string{testString1: &testString2},
				}},
				resolver: withMockResolver(t, func(mcs *mockclient.MockResolverService) {
					mcs.EXPECT().GetProviderDetails(
						context.Background(),
						nil,
						&v1.SecretReference{
							Name:      testString2,
							Namespace: testString2,
						},
					).Return(
						map[string]*string{testString1: &testString2},
						nil,
					)
				}),
			},
			want: want{
				result: true,
				err:    nil,
			},
		},
		"UpToDateProviderDetailsSecretRefError": {
			args: args{
				cr: identityProvider(withSpec(svcapitypes.IdentityProviderParameters{
					CustomIdentityProviderParameters: svcapitypes.CustomIdentityProviderParameters{
						ProviderDetailsSecretRef: v1.SecretReference{
							Name:      testString1,
							Namespace: testString2,
						},
					},
				})),
				resp: &svcsdk.DescribeIdentityProviderOutput{IdentityProvider: &svcsdk.IdentityProviderType{
					ProviderDetails: map[string]*string{testString1: &testString2},
				}},
				resolver: withMockResolver(t, func(mcs *mockclient.MockResolverService) {
					mcs.EXPECT().GetProviderDetails(
						context.Background(),
						nil,
						&v1.SecretReference{
							Name:      testString1,
							Namespace: testString2,
						},
					).Return(
						nil,
						errBoom,
					)
				}),
			},
			want: want{
				result: false,
				err:    errBoom,
			},
		},
		"ChangedAttributeMapping": {
			args: args{
				cr: identityProvider(withSpec(svcapitypes.IdentityProviderParameters{
					AttributeMapping:                 map[string]*string{testString1: &testString2},
					CustomIdentityProviderParameters: svcapitypes.CustomIdentityProviderParameters{},
				})),
				resp: &svcsdk.DescribeIdentityProviderOutput{IdentityProvider: &svcsdk.IdentityProviderType{
					AttributeMapping: map[string]*string{testString1: &testString1},
				}},
				resolver: withMockResolver(t, func(mcs *mockclient.MockResolverService) {
					mcs.EXPECT().GetProviderDetails(gomock.Any(), gomock.Any(), gomock.Any()).Times(1)
				}),
			},
			want: want{
				result: false,
				err:    nil,
			},
		},
		"ChangedIDpIdentifiers": {
			args: args{
				cr: identityProvider(withSpec(svcapitypes.IdentityProviderParameters{
					IDpIdentifiers:                   []*string{&testString2},
					CustomIdentityProviderParameters: svcapitypes.CustomIdentityProviderParameters{},
				})),
				resp: &svcsdk.DescribeIdentityProviderOutput{IdentityProvider: &svcsdk.IdentityProviderType{
					IdpIdentifiers: []*string{&testString1},
				}},
				resolver: withMockResolver(t, func(mcs *mockclient.MockResolverService) {
					mcs.EXPECT().GetProviderDetails(gomock.Any(), gomock.Any(), gomock.Any()).Times(1)
				}),
			},
			want: want{
				result: false,
				err:    nil,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &custom{resolver: tc.args.resolver, kube: nil}
			// Act
			result, _, err := e.isUpToDate(context.Background(), tc.args.cr, tc.args.resp)

			// Assert
			if diff := cmp.Diff(tc.want.result, result, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestLateInitialize(t *testing.T) {
	type args struct {
		cr      *svcapitypes.IdentityProviderParameters
		current *svcsdk.DescribeIdentityProviderOutput
	}

	type want struct {
		result *svcapitypes.IdentityProviderParameters
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"NoLateInitialization": {
			args: args{
				cr: &svcapitypes.IdentityProviderParameters{
					AttributeMapping: map[string]*string{testString1: &testString2},
				},
				current: &svcsdk.DescribeIdentityProviderOutput{IdentityProvider: &svcsdk.IdentityProviderType{
					AttributeMapping: map[string]*string{testString1: &testString1},
				}},
			},
			want: want{
				result: &svcapitypes.IdentityProviderParameters{
					AttributeMapping: map[string]*string{testString1: &testString2},
				},
				err: nil,
			},
		},
		"LateInitialize": {
			args: args{
				cr: &svcapitypes.IdentityProviderParameters{},
				current: &svcsdk.DescribeIdentityProviderOutput{IdentityProvider: &svcsdk.IdentityProviderType{
					AttributeMapping: map[string]*string{testString1: &testString2},
				}},
			},
			want: want{
				result: &svcapitypes.IdentityProviderParameters{
					AttributeMapping: map[string]*string{testString1: &testString2},
				},
				err: nil,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			// Act
			err := lateInitialize(tc.args.cr, tc.args.current)

			if diff := cmp.Diff(tc.want.result, tc.args.cr, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}
