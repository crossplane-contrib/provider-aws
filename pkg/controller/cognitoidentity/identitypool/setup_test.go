package identitypool

import (
	"context"
	"testing"

	svcsdk "github.com/aws/aws-sdk-go/service/cognitoidentity"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/cognitoidentity/v1alpha1"
)

type functionModifier func(*svcapitypes.IdentityPool)

func withSpec(p svcapitypes.IdentityPoolParameters) functionModifier {
	return func(r *svcapitypes.IdentityPool) { r.Spec.ForProvider = p }
}

func withObservation(s svcapitypes.IdentityPoolObservation) functionModifier {
	return func(r *svcapitypes.IdentityPool) { r.Status.AtProvider = s }
}

func withExternalName(v string) functionModifier {
	return func(r *svcapitypes.IdentityPool) {
		meta.SetExternalName(r, v)
	}
}

func identityPool(m ...functionModifier) *svcapitypes.IdentityPool {
	cr := &svcapitypes.IdentityPool{}
	cr.Name = "test-identitypool-name"
	cr.Spec.ForProvider.IdentityPoolName = &cr.Name
	for _, f := range m {
		f(cr)
	}
	return cr
}

var (
	testString1 string = "string1"
	testString2 string = "string2"
	errBoom     error  = errors.New("boom")
	testBool1   bool   = true
	testBool2   bool   = false
)

func TestIsUpToDate(t *testing.T) {
	type args struct {
		cr   *svcapitypes.IdentityPool
		resp *svcsdk.IdentityPool
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
				cr:   identityPool(withSpec(svcapitypes.IdentityPoolParameters{})),
				resp: &svcsdk.IdentityPool{},
			},
			want: want{
				result: true,
				err:    nil,
			},
		},
		"UpToDateWithValues": {
			args: args{
				cr: identityPool(withSpec(svcapitypes.IdentityPoolParameters{
					AllowClassicFlow: &testBool1,
					CustomIdentityPoolParameters: svcapitypes.CustomIdentityPoolParameters{
						CognitoIdentityProviders: []*svcapitypes.Provider{
							{
								ClientID:             &testString1,
								ProviderName:         &testString1,
								ServerSideTokenCheck: &testBool1,
							},
						},
						OpenIDConnectProviderARNs: []*string{
							&testString1,
						},
						AllowUnauthenticatedIdentities: &testBool1,
					},
					DeveloperProviderName: &testString1,
					IdentityPoolTags: map[string]*string{
						testString1: &testString2,
					},
					SamlProviderARNs: []*string{
						&testString1,
					},
					SupportedLoginProviders: map[string]*string{
						testString1: &testString2,
					},
				})),
				resp: &svcsdk.IdentityPool{
					AllowClassicFlow:               &testBool1,
					AllowUnauthenticatedIdentities: &testBool1,
					CognitoIdentityProviders: []*svcsdk.Provider{
						{
							ClientId:             &testString1,
							ProviderName:         &testString1,
							ServerSideTokenCheck: &testBool1,
						},
					},
					OpenIdConnectProviderARNs: []*string{
						&testString1,
					},
					DeveloperProviderName: &testString1,
					IdentityPoolTags: map[string]*string{
						testString1: &testString2,
					},
					SamlProviderARNs: []*string{
						&testString1,
					},
					SupportedLoginProviders: map[string]*string{
						testString1: &testString2,
					},
				},
			},
			want: want{
				result: true,
				err:    nil,
			},
		},
		"ChangedAllowClassicFlow": {
			args: args{
				cr: identityPool(withSpec(svcapitypes.IdentityPoolParameters{
					AllowClassicFlow: &testBool1,
				})),
				resp: &svcsdk.IdentityPool{
					AllowClassicFlow: &testBool2,
				},
			},
			want: want{
				result: false,
				err:    nil,
			},
		},
		"ChangedAllowUnauthenticatedIdentities": {
			args: args{
				cr: identityPool(withSpec(svcapitypes.IdentityPoolParameters{
					CustomIdentityPoolParameters: svcapitypes.CustomIdentityPoolParameters{
						AllowUnauthenticatedIdentities: &testBool1,
					},
				})),
				resp: &svcsdk.IdentityPool{
					AllowUnauthenticatedIdentities: &testBool2,
				},
			},
			want: want{
				result: false,
				err:    nil,
			},
		},
		"ChangedCognitoIdentityProvidersClientID": {
			args: args{
				cr: identityPool(withSpec(svcapitypes.IdentityPoolParameters{
					CustomIdentityPoolParameters: svcapitypes.CustomIdentityPoolParameters{
						CognitoIdentityProviders: []*svcapitypes.Provider{
							{
								ClientID:             &testString1,
								ProviderName:         &testString1,
								ServerSideTokenCheck: &testBool1,
							},
						},
					},
				})),
				resp: &svcsdk.IdentityPool{
					CognitoIdentityProviders: []*svcsdk.Provider{
						{
							ClientId:             &testString2,
							ProviderName:         &testString1,
							ServerSideTokenCheck: &testBool1,
						},
					},
				},
			},
			want: want{
				result: false,
				err:    nil,
			},
		},
		"ChangedCognitoIdentityProvidersProviderName": {
			args: args{
				cr: identityPool(withSpec(svcapitypes.IdentityPoolParameters{
					CustomIdentityPoolParameters: svcapitypes.CustomIdentityPoolParameters{
						CognitoIdentityProviders: []*svcapitypes.Provider{
							{
								ClientID:             &testString1,
								ProviderName:         &testString1,
								ServerSideTokenCheck: &testBool1,
							},
						},
					},
				})),
				resp: &svcsdk.IdentityPool{
					CognitoIdentityProviders: []*svcsdk.Provider{
						{
							ClientId:             &testString1,
							ProviderName:         &testString2,
							ServerSideTokenCheck: &testBool1,
						},
					},
				},
			},
			want: want{
				result: false,
				err:    nil,
			},
		},
		"ChangedCognitoIdentityProvidersServerSideTokenCheck": {
			args: args{
				cr: identityPool(withSpec(svcapitypes.IdentityPoolParameters{
					CustomIdentityPoolParameters: svcapitypes.CustomIdentityPoolParameters{
						CognitoIdentityProviders: []*svcapitypes.Provider{
							{
								ClientID:             &testString1,
								ProviderName:         &testString1,
								ServerSideTokenCheck: &testBool1,
							},
						},
					},
				})),
				resp: &svcsdk.IdentityPool{
					CognitoIdentityProviders: []*svcsdk.Provider{
						{
							ClientId:             &testString1,
							ProviderName:         &testString1,
							ServerSideTokenCheck: &testBool2,
						},
					},
				},
			},
			want: want{
				result: false,
				err:    nil,
			},
		},
		"ChangedOpenIDConnectProviderARNs": {
			args: args{
				cr: identityPool(withSpec(svcapitypes.IdentityPoolParameters{
					CustomIdentityPoolParameters: svcapitypes.CustomIdentityPoolParameters{
						OpenIDConnectProviderARNs: []*string{
							&testString1,
						},
					},
				})),
				resp: &svcsdk.IdentityPool{
					OpenIdConnectProviderARNs: []*string{
						&testString2,
					},
				},
			},
			want: want{
				result: false,
				err:    nil,
			},
		},
		"ChangedDeveloperProviderName": {
			args: args{
				cr: identityPool(withSpec(svcapitypes.IdentityPoolParameters{
					DeveloperProviderName: &testString1,
				})),
				resp: &svcsdk.IdentityPool{
					DeveloperProviderName: &testString2,
				},
			},
			want: want{
				result: false,
				err:    nil,
			},
		},
		"ChangedIdentityPoolTags": {
			args: args{
				cr: identityPool(withSpec(svcapitypes.IdentityPoolParameters{
					IdentityPoolTags: map[string]*string{
						testString1: &testString2,
					},
				})),
				resp: &svcsdk.IdentityPool{
					IdentityPoolTags: map[string]*string{
						testString1: &testString1,
					},
				},
			},
			want: want{
				result: false,
				err:    nil,
			},
		},
		"ChangedSamlProviderARNs": {
			args: args{
				cr: identityPool(withSpec(svcapitypes.IdentityPoolParameters{
					SamlProviderARNs: []*string{
						&testString1,
					},
				})),
				resp: &svcsdk.IdentityPool{
					SamlProviderARNs: []*string{
						&testString2,
					},
				},
			},
			want: want{
				result: false,
				err:    nil,
			},
		},
		"ChangedSupportedLoginProviders": {
			args: args{
				cr: identityPool(withSpec(svcapitypes.IdentityPoolParameters{
					SupportedLoginProviders: map[string]*string{
						testString1: &testString2,
					},
				})),
				resp: &svcsdk.IdentityPool{
					SupportedLoginProviders: map[string]*string{
						testString1: &testString1,
					},
				},
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
			result, err := isUpToDate(tc.args.cr, tc.args.resp)

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

func TestPostCreate(t *testing.T) {
	type args struct {
		cr  *svcapitypes.IdentityPool
		obj *svcsdk.IdentityPool
		err error
	}

	type want struct {
		cr     *svcapitypes.IdentityPool
		result managed.ExternalCreation
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"SetExternalName": {
			args: args{
				cr: identityPool(
					withSpec(svcapitypes.IdentityPoolParameters{}),
					withObservation(svcapitypes.IdentityPoolObservation{
						IdentityPoolID: &testString1,
					}),
				),
				obj: &svcsdk.IdentityPool{
					IdentityPoolId: &testString1,
				},
				err: nil,
			},
			want: want{
				cr: identityPool(
					withSpec(svcapitypes.IdentityPoolParameters{}),
					withObservation(svcapitypes.IdentityPoolObservation{
						IdentityPoolID: &testString1,
					}),
					withExternalName(testString1),
				),
				result: managed.ExternalCreation{},
				err:    nil,
			},
		},
		"FailedCreation": {
			args: args{
				cr: identityPool(
					withSpec(svcapitypes.IdentityPoolParameters{}),
				),
				obj: nil,
				err: errBoom,
			},
			want: want{
				cr: identityPool(
					withSpec(svcapitypes.IdentityPoolParameters{}),
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
