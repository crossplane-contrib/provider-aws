package userpoolclient

import (
	"context"
	"testing"

	svcsdk "github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/cognitoidentityprovider/v1alpha1"
)

type functionModifier func(*svcapitypes.UserPoolClient)

func withSpec(p svcapitypes.UserPoolClientParameters) functionModifier {
	return func(r *svcapitypes.UserPoolClient) { r.Spec.ForProvider = p }
}

func withObservation(s svcapitypes.UserPoolClientObservation) functionModifier {
	return func(r *svcapitypes.UserPoolClient) { r.Status.AtProvider = s }
}

func withExternalName(v string) functionModifier {
	return func(r *svcapitypes.UserPoolClient) {
		meta.SetExternalName(r, v)
	}
}

func userPoolClient(m ...functionModifier) *svcapitypes.UserPoolClient {
	cr := &svcapitypes.UserPoolClient{}
	cr.Name = "test-group-name"
	cr.Spec.ForProvider.ClientName = &cr.Name
	for _, f := range m {
		f(cr)
	}
	return cr
}

var (
	testString1       string = "string1"
	testString2       string = "string2"
	errBoom           error  = errors.New("boom")
	testNumber        int64  = 1
	testNumberChanged int64  = 2
	testBool1         bool   = true
	testBool2         bool   = false
)

func TestIsUpToDate(t *testing.T) {
	type args struct {
		cr   *svcapitypes.UserPoolClient
		resp *svcsdk.DescribeUserPoolClientOutput
	}

	type want struct {
		result bool
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"UpToDate": {
			args: args{
				cr:   userPoolClient(withSpec(svcapitypes.UserPoolClientParameters{})),
				resp: &svcsdk.DescribeUserPoolClientOutput{UserPoolClient: &svcsdk.UserPoolClientType{}},
			},
			want: want{
				result: true,
				err:    nil,
			},
		},
		"ChangedAccessTokenValidity": {
			args: args{
				cr: userPoolClient(withSpec(svcapitypes.UserPoolClientParameters{
					AccessTokenValidity: &testNumberChanged,
				})),
				resp: &svcsdk.DescribeUserPoolClientOutput{UserPoolClient: &svcsdk.UserPoolClientType{
					AccessTokenValidity: &testNumber,
				}},
			},
			want: want{
				result: false,
				err:    nil,
			},
		},
		"ChangedAllowedOAuthFlows": {
			args: args{
				cr: userPoolClient(withSpec(svcapitypes.UserPoolClientParameters{
					AllowedOAuthFlows: []*string{&testString1, &testString2},
				})),
				resp: &svcsdk.DescribeUserPoolClientOutput{UserPoolClient: &svcsdk.UserPoolClientType{
					AllowedOAuthFlows: []*string{&testString2, &testString1},
				}},
			},
			want: want{
				result: false,
				err:    nil,
			},
		},
		"ChangedAllowedOAuthFlowsUserPoolClient": {
			args: args{
				cr: userPoolClient(withSpec(svcapitypes.UserPoolClientParameters{
					AllowedOAuthFlowsUserPoolClient: &testBool1,
				})),
				resp: &svcsdk.DescribeUserPoolClientOutput{UserPoolClient: &svcsdk.UserPoolClientType{
					AllowedOAuthFlowsUserPoolClient: &testBool2,
				}},
			},
			want: want{
				result: false,
				err:    nil,
			},
		},
		"ChangedAllowedOAuthScopes": {
			args: args{
				cr: userPoolClient(withSpec(svcapitypes.UserPoolClientParameters{
					AllowedOAuthScopes: []*string{&testString1, &testString2},
				})),
				resp: &svcsdk.DescribeUserPoolClientOutput{UserPoolClient: &svcsdk.UserPoolClientType{
					AllowedOAuthScopes: []*string{&testString2, &testString1},
				}},
			},
			want: want{
				result: false,
				err:    nil,
			},
		},
		"ChangedAnalyticsConfiguration": {
			args: args{
				cr: userPoolClient(withSpec(svcapitypes.UserPoolClientParameters{
					AnalyticsConfiguration: &svcapitypes.AnalyticsConfigurationType{
						ApplicationARN: &testString1,
					},
				})),
				resp: &svcsdk.DescribeUserPoolClientOutput{UserPoolClient: &svcsdk.UserPoolClientType{
					AnalyticsConfiguration: &svcsdk.AnalyticsConfigurationType{
						ApplicationArn: &testString2,
					},
				}},
			},
			want: want{
				result: false,
				err:    nil,
			},
		},
		"ChangedCallbackURLs": {
			args: args{
				cr: userPoolClient(withSpec(svcapitypes.UserPoolClientParameters{
					CallbackURLs: []*string{&testString1, &testString2},
				})),
				resp: &svcsdk.DescribeUserPoolClientOutput{UserPoolClient: &svcsdk.UserPoolClientType{
					CallbackURLs: []*string{&testString2, &testString1},
				}},
			},
			want: want{
				result: false,
				err:    nil,
			},
		},
		"ChangedDefaultRedirectURI": {
			args: args{
				cr: userPoolClient(withSpec(svcapitypes.UserPoolClientParameters{
					DefaultRedirectURI: &testString1,
				})),
				resp: &svcsdk.DescribeUserPoolClientOutput{UserPoolClient: &svcsdk.UserPoolClientType{
					DefaultRedirectURI: &testString2,
				}},
			},
			want: want{
				result: false,
				err:    nil,
			},
		},
		"ChangedExplicitAuthFlows": {
			args: args{
				cr: userPoolClient(withSpec(svcapitypes.UserPoolClientParameters{
					ExplicitAuthFlows: []*string{&testString1, &testString2},
				})),
				resp: &svcsdk.DescribeUserPoolClientOutput{UserPoolClient: &svcsdk.UserPoolClientType{
					ExplicitAuthFlows: []*string{&testString2, &testString1},
				}},
			},
			want: want{
				result: false,
				err:    nil,
			},
		},
		"ChangedIDTokenValidity": {
			args: args{
				cr: userPoolClient(withSpec(svcapitypes.UserPoolClientParameters{
					IDTokenValidity: &testNumber,
				})),
				resp: &svcsdk.DescribeUserPoolClientOutput{UserPoolClient: &svcsdk.UserPoolClientType{
					IdTokenValidity: &testNumberChanged,
				}},
			},
			want: want{
				result: false,
				err:    nil,
			},
		},
		"ChangedLogoutURLs": {
			args: args{
				cr: userPoolClient(withSpec(svcapitypes.UserPoolClientParameters{
					LogoutURLs: []*string{&testString1, &testString2},
				})),
				resp: &svcsdk.DescribeUserPoolClientOutput{UserPoolClient: &svcsdk.UserPoolClientType{
					LogoutURLs: []*string{&testString2, &testString1},
				}},
			},
			want: want{
				result: false,
				err:    nil,
			},
		},
		"ChangedPreventUserExistenceErrors": {
			args: args{
				cr: userPoolClient(withSpec(svcapitypes.UserPoolClientParameters{
					PreventUserExistenceErrors: &testString1,
				})),
				resp: &svcsdk.DescribeUserPoolClientOutput{UserPoolClient: &svcsdk.UserPoolClientType{
					PreventUserExistenceErrors: &testString2,
				}},
			},
			want: want{
				result: false,
				err:    nil,
			},
		},
		"ChangedReadAttributes": {
			args: args{
				cr: userPoolClient(withSpec(svcapitypes.UserPoolClientParameters{
					ReadAttributes: []*string{&testString1, &testString2},
				})),
				resp: &svcsdk.DescribeUserPoolClientOutput{UserPoolClient: &svcsdk.UserPoolClientType{
					ReadAttributes: []*string{&testString2, &testString1},
				}},
			},
			want: want{
				result: false,
				err:    nil,
			},
		},
		"ChangedRefreshTokenValidity": {
			args: args{
				cr: userPoolClient(withSpec(svcapitypes.UserPoolClientParameters{
					RefreshTokenValidity: &testNumber,
				})),
				resp: &svcsdk.DescribeUserPoolClientOutput{UserPoolClient: &svcsdk.UserPoolClientType{
					RefreshTokenValidity: &testNumberChanged,
				}},
			},
			want: want{
				result: false,
				err:    nil,
			},
		},
		"ChangedSupportedIdentityProviders": {
			args: args{
				cr: userPoolClient(withSpec(svcapitypes.UserPoolClientParameters{
					SupportedIdentityProviders: []*string{&testString1, &testString2},
				})),
				resp: &svcsdk.DescribeUserPoolClientOutput{UserPoolClient: &svcsdk.UserPoolClientType{
					SupportedIdentityProviders: []*string{&testString2, &testString1},
				}},
			},
			want: want{
				result: false,
				err:    nil,
			},
		},
		"ChangedTokenValidityUnits": {
			args: args{
				cr: userPoolClient(withSpec(svcapitypes.UserPoolClientParameters{
					TokenValidityUnits: &svcapitypes.TokenValidityUnitsType{
						AccessToken: &testString1,
					},
				})),
				resp: &svcsdk.DescribeUserPoolClientOutput{UserPoolClient: &svcsdk.UserPoolClientType{
					TokenValidityUnits: &svcsdk.TokenValidityUnitsType{
						AccessToken: &testString2,
					},
				}},
			},
			want: want{
				result: false,
				err:    nil,
			},
		},
		"ChangedWriteAttributes": {
			args: args{
				cr: userPoolClient(withSpec(svcapitypes.UserPoolClientParameters{
					WriteAttributes: []*string{&testString1, &testString2},
				})),
				resp: &svcsdk.DescribeUserPoolClientOutput{UserPoolClient: &svcsdk.UserPoolClientType{
					WriteAttributes: []*string{&testString2, &testString1},
				}},
			},
			want: want{
				result: false,
				err:    nil,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			// Act
			result, _, err := isUpToDate(context.Background(), tc.args.cr, tc.args.resp)

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
		cr   *svcapitypes.UserPoolClientParameters
		resp *svcsdk.DescribeUserPoolClientOutput
	}

	type want struct {
		result *svcapitypes.UserPoolClientParameters
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"NoLateInitialization": {
			args: args{
				cr: &svcapitypes.UserPoolClientParameters{
					RefreshTokenValidity: &testNumber,
				},
				resp: &svcsdk.DescribeUserPoolClientOutput{UserPoolClient: &svcsdk.UserPoolClientType{
					RefreshTokenValidity: &testNumberChanged,
				}},
			},
			want: want{
				result: &svcapitypes.UserPoolClientParameters{
					RefreshTokenValidity: &testNumber,
				},
				err: nil,
			},
		},
		"LateInitializeRefreshTokenValidity": {
			args: args{
				cr: &svcapitypes.UserPoolClientParameters{},
				resp: &svcsdk.DescribeUserPoolClientOutput{UserPoolClient: &svcsdk.UserPoolClientType{
					RefreshTokenValidity: &testNumber,
				}},
			},
			want: want{
				result: &svcapitypes.UserPoolClientParameters{
					RefreshTokenValidity: &testNumber,
				},
				err: nil,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			// Act
			err := lateInitialize(tc.args.cr, tc.args.resp)

			if diff := cmp.Diff(tc.want.result, tc.args.cr, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestPostCreate(t *testing.T) {
	type args struct {
		cr  *svcapitypes.UserPoolClient
		obj *svcsdk.CreateUserPoolClientOutput
		err error
	}

	type want struct {
		cr     *svcapitypes.UserPoolClient
		result managed.ExternalCreation
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"SetExternalNameAndConnectionDetails": {
			args: args{
				cr: userPoolClient(
					withSpec(svcapitypes.UserPoolClientParameters{
						CustomUserPoolClientParameters: svcapitypes.CustomUserPoolClientParameters{
							UserPoolID: &testString1,
						},
					}),
					withObservation(svcapitypes.UserPoolClientObservation{
						ClientID:     &testString1,
						ClientSecret: &testString2,
					}),
				),
				obj: &svcsdk.CreateUserPoolClientOutput{
					UserPoolClient: &svcsdk.UserPoolClientType{
						ClientId: &testString1,
					},
				},
				err: nil,
			},
			want: want{
				cr: userPoolClient(
					withSpec(svcapitypes.UserPoolClientParameters{
						CustomUserPoolClientParameters: svcapitypes.CustomUserPoolClientParameters{
							UserPoolID: &testString1,
						},
					}),
					withObservation(svcapitypes.UserPoolClientObservation{
						ClientID:     &testString1,
						ClientSecret: &testString2,
					}),
					withExternalName(testString1),
				),
				result: managed.ExternalCreation{
					ConnectionDetails: managed.ConnectionDetails{
						"clientID":     []byte(testString1),
						"clientSecret": []byte(testString2),
						"userPoolID":   []byte(testString1),
					},
				},
				err: nil,
			},
		},
		"FailedCreation": {
			args: args{
				cr: userPoolClient(
					withSpec(svcapitypes.UserPoolClientParameters{
						CustomUserPoolClientParameters: svcapitypes.CustomUserPoolClientParameters{
							UserPoolID: &testString1,
						},
					}),
				),
				obj: nil,
				err: errBoom,
			},
			want: want{
				cr: userPoolClient(
					withSpec(svcapitypes.UserPoolClientParameters{
						CustomUserPoolClientParameters: svcapitypes.CustomUserPoolClientParameters{
							UserPoolID: &testString1,
						},
					}),
				),
				result: managed.ExternalCreation{},
				err:    errBoom,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			// Act
			result, err := postCreate(context.Background(), tc.args.cr, tc.args.obj, managed.ExternalCreation{}, tc.args.err)

			if diff := cmp.Diff(tc.want.cr, tc.args.cr, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.result, result, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestPostUpdate(t *testing.T) {
	type args struct {
		cr  *svcapitypes.UserPoolClient
		obj *svcsdk.UpdateUserPoolClientOutput
		err error
	}

	type want struct {
		result managed.ExternalUpdate
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"SetConnectionDetails": {
			args: args{
				cr: userPoolClient(
					withSpec(svcapitypes.UserPoolClientParameters{
						CustomUserPoolClientParameters: svcapitypes.CustomUserPoolClientParameters{
							UserPoolID: &testString1,
						},
					}),
					withObservation(svcapitypes.UserPoolClientObservation{
						ClientID:     &testString1,
						ClientSecret: &testString2,
					}),
					withExternalName(testString1),
				),
				obj: &svcsdk.UpdateUserPoolClientOutput{
					UserPoolClient: &svcsdk.UserPoolClientType{
						ClientId: &testString1,
					},
				},
				err: nil,
			},
			want: want{
				result: managed.ExternalUpdate{
					ConnectionDetails: managed.ConnectionDetails{
						"clientID":     []byte(testString1),
						"clientSecret": []byte(testString2),
						"userPoolID":   []byte(testString1),
					},
				},
				err: nil,
			},
		},
		"FailedUpdate": {
			args: args{
				cr: userPoolClient(
					withSpec(svcapitypes.UserPoolClientParameters{
						CustomUserPoolClientParameters: svcapitypes.CustomUserPoolClientParameters{
							UserPoolID: &testString1,
						},
					}),
					withObservation(svcapitypes.UserPoolClientObservation{
						ClientID:     &testString1,
						ClientSecret: &testString2,
					}),
					withExternalName(testString1),
				),
				obj: nil,
				err: errBoom,
			},
			want: want{
				result: managed.ExternalUpdate{},
				err:    errBoom,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			// Act
			result, err := postUpdate(context.Background(), tc.args.cr, tc.args.obj, managed.ExternalUpdate{}, tc.args.err)

			if diff := cmp.Diff(tc.want.result, result, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}
