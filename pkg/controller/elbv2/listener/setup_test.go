package listener

import (
	"testing"

	svcsdk "github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/google/go-cmp/cmp"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/elbv2/v1alpha1"
)

func strPtr(s string) *string {
	return &s
}

func i64Ptr(i int64) *int64 {
	return &i
}

func boolPtr(b bool) *bool {
	return &b
}

func TestGenerateDefaultActions(t *testing.T) {
	cases := map[string]struct {
		reason string
		cr     *svcapitypes.Listener
		want   []*svcsdk.Action
	}{
		"TestEmptySpec": {
			reason: "Test generating empty actions when spec does not specify them.",
			cr:     &svcapitypes.Listener{},
			want:   []*svcsdk.Action{},
		},
		"TestForwardConfig": {
			reason: "Test generating a forwardconfig action",
			cr: &svcapitypes.Listener{
				Spec: svcapitypes.ListenerSpec{
					ForProvider: svcapitypes.ListenerParameters{
						Region: "us-east-1",
						CustomListenerParameters: svcapitypes.CustomListenerParameters{
							DefaultActions: []*svcapitypes.CustomAction{
								{
									Type: strPtr("forward"),
									ForwardConfig: &svcapitypes.CustomForwardActionConfig{
										TargetGroups: []*svcapitypes.CustomTargetGroupTuple{
											{
												TargetGroupTuple: svcapitypes.TargetGroupTuple{
													TargetGroupARN: strPtr("arn:::"),
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			want: []*svcsdk.Action{
				{
					Type: strPtr("forward"),
					ForwardConfig: &svcsdk.ForwardActionConfig{
						TargetGroups: []*svcsdk.TargetGroupTuple{
							{
								TargetGroupArn: strPtr("arn:::"),
							},
						},
					},
				},
			},
		},
		"TestFixedResponseConfig": {
			reason: "Test generating a FixedResponseConfig.",
			cr: &svcapitypes.Listener{
				Spec: svcapitypes.ListenerSpec{
					ForProvider: svcapitypes.ListenerParameters{
						Region: "us-east-1",
						CustomListenerParameters: svcapitypes.CustomListenerParameters{
							DefaultActions: []*svcapitypes.CustomAction{
								{
									Type: strPtr("fixedresponse"),
									FixedResponseConfig: &svcapitypes.FixedResponseActionConfig{
										ContentType: strPtr("text/html"),
										MessageBody: strPtr("testing"),
										StatusCode:  strPtr("200"),
									},
								},
							},
						},
					},
				},
			},
			want: []*svcsdk.Action{
				{
					Type: strPtr("fixedresponse"),
					FixedResponseConfig: &svcsdk.FixedResponseActionConfig{
						ContentType: strPtr("text/html"),
						MessageBody: strPtr("testing"),
						StatusCode:  strPtr("200"),
					},
				},
			},
		},
		"TestRedirectConfig": {
			reason: "Test generating a RedirectConfig.",
			cr: &svcapitypes.Listener{
				Spec: svcapitypes.ListenerSpec{
					ForProvider: svcapitypes.ListenerParameters{
						Region: "us-east-1",
						CustomListenerParameters: svcapitypes.CustomListenerParameters{
							DefaultActions: []*svcapitypes.CustomAction{
								{
									Type: strPtr("redirect"),
									RedirectConfig: &svcapitypes.RedirectActionConfig{
										Host:       strPtr("example.com"),
										Path:       strPtr("/test"),
										Port:       strPtr("443"),
										Protocol:   strPtr("HTTPS"),
										Query:      strPtr("testquery"),
										StatusCode: strPtr("200"),
									},
								},
							},
						},
					},
				},
			},
			want: []*svcsdk.Action{
				{
					Type: strPtr("redirect"),
					RedirectConfig: &svcsdk.RedirectActionConfig{
						Host:       strPtr("example.com"),
						Path:       strPtr("/test"),
						Port:       strPtr("443"),
						Protocol:   strPtr("HTTPS"),
						Query:      strPtr("testquery"),
						StatusCode: strPtr("200"),
					},
				},
			},
		},
		"TestAuthenticateCognitoConfig": {
			reason: "Test genrating an AuthenticateCognitoConfig",
			cr: &svcapitypes.Listener{
				Spec: svcapitypes.ListenerSpec{
					ForProvider: svcapitypes.ListenerParameters{
						Region: "us-east-1",
						CustomListenerParameters: svcapitypes.CustomListenerParameters{
							DefaultActions: []*svcapitypes.CustomAction{
								{
									AuthenticateCognitoConfig: &svcapitypes.AuthenticateCognitoActionConfig{
										AuthenticationRequestExtraParams: map[string]*string{"foo": strPtr("bar")},
										OnUnauthenticatedRequest:         strPtr("deny"),
										Scope:                            strPtr("openid"),
										SessionCookieName:                strPtr("AWSELBAuthSessionCookie"),
										SessionTimeout:                   i64Ptr(int64(604800)),
										UserPoolARN:                      strPtr("arn:::"),
										UserPoolClientID:                 strPtr("testid"),
										UserPoolDomain:                   strPtr("example.com"),
									},
								},
							},
						},
					},
				},
			},
			want: []*svcsdk.Action{
				{
					AuthenticateCognitoConfig: &svcsdk.AuthenticateCognitoActionConfig{
						AuthenticationRequestExtraParams: map[string]*string{"foo": strPtr("bar")},
						OnUnauthenticatedRequest:         strPtr("deny"),
						Scope:                            strPtr("openid"),
						SessionCookieName:                strPtr("AWSELBAuthSessionCookie"),
						SessionTimeout:                   i64Ptr(int64(604800)),
						UserPoolArn:                      strPtr("arn:::"),
						UserPoolClientId:                 strPtr("testid"),
						UserPoolDomain:                   strPtr("example.com"),
					},
				},
			},
		},
		"TestAuthenticateOidcConfig": {
			reason: "Test generating an AuthenticateCidcConfig.",
			cr: &svcapitypes.Listener{
				Spec: svcapitypes.ListenerSpec{
					ForProvider: svcapitypes.ListenerParameters{
						Region: "us-east-1",
						CustomListenerParameters: svcapitypes.CustomListenerParameters{
							DefaultActions: []*svcapitypes.CustomAction{
								{
									AuthenticateOidcConfig: &svcapitypes.AuthenticateOIDCActionConfig{
										AuthenticationRequestExtraParams: map[string]*string{"foo": strPtr("bar")},
										AuthorizationEndpoint:            strPtr("https://example.com/auth"),
										ClientID:                         strPtr("test"),
										ClientSecret:                     strPtr("supersecret"),
										Issuer:                           strPtr("https://example.com"),
										OnUnauthenticatedRequest:         strPtr("deny"),
										Scope:                            strPtr("openid"),
										SessionCookieName:                strPtr("AWSELBAuthSessionCookie"),
										SessionTimeout:                   i64Ptr(int64(604800)),
										TokenEndpoint:                    strPtr("https://example.com/token"),
										UseExistingClientSecret:          boolPtr(true),
										UserInfoEndpoint:                 strPtr("https://example.com/user"),
									},
								},
							},
						},
					},
				},
			},
			want: []*svcsdk.Action{
				{
					AuthenticateOidcConfig: &svcsdk.AuthenticateOidcActionConfig{
						AuthenticationRequestExtraParams: map[string]*string{"foo": strPtr("bar")},
						AuthorizationEndpoint:            strPtr("https://example.com/auth"),
						ClientId:                         strPtr("test"),
						ClientSecret:                     strPtr("supersecret"),
						Issuer:                           strPtr("https://example.com"),
						OnUnauthenticatedRequest:         strPtr("deny"),
						Scope:                            strPtr("openid"),
						SessionCookieName:                strPtr("AWSELBAuthSessionCookie"),
						SessionTimeout:                   i64Ptr(int64(604800)),
						TokenEndpoint:                    strPtr("https://example.com/token"),
						UseExistingClientSecret:          boolPtr(true),
						UserInfoEndpoint:                 strPtr("https://example.com/user"),
					},
				},
			},
		},
		"TestMultipleConfigs": {
			reason: "Test generating multiple configs.",
			cr: &svcapitypes.Listener{
				Spec: svcapitypes.ListenerSpec{
					ForProvider: svcapitypes.ListenerParameters{
						Region: "us-east-1",
						CustomListenerParameters: svcapitypes.CustomListenerParameters{
							DefaultActions: []*svcapitypes.CustomAction{
								{
									Type: strPtr("forward"),
									ForwardConfig: &svcapitypes.CustomForwardActionConfig{
										TargetGroups: []*svcapitypes.CustomTargetGroupTuple{
											{
												TargetGroupTuple: svcapitypes.TargetGroupTuple{
													TargetGroupARN: strPtr("arn:1::"),
												},
											},
										},
									},
								},
								{
									Type: strPtr("forward"),
									ForwardConfig: &svcapitypes.CustomForwardActionConfig{
										TargetGroups: []*svcapitypes.CustomTargetGroupTuple{
											{
												TargetGroupTuple: svcapitypes.TargetGroupTuple{
													TargetGroupARN: strPtr("arn:2::"),
												},
											},
											{
												TargetGroupTuple: svcapitypes.TargetGroupTuple{
													TargetGroupARN: strPtr("arn:3::"),
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			want: []*svcsdk.Action{
				{
					Type: strPtr("forward"),
					ForwardConfig: &svcsdk.ForwardActionConfig{
						TargetGroups: []*svcsdk.TargetGroupTuple{
							{
								TargetGroupArn: strPtr("arn:1::"),
							},
						},
					},
				},
				{
					Type: strPtr("forward"),
					ForwardConfig: &svcsdk.ForwardActionConfig{
						TargetGroups: []*svcsdk.TargetGroupTuple{
							{
								TargetGroupArn: strPtr("arn:2::"),
							},
							{
								TargetGroupArn: strPtr("arn:3::"),
							},
						},
					},
				},
			},
		},
	}

	for name, tc := range cases {

		t.Run(name, func(t *testing.T) {
			got := generateDefaultActions(tc.cr)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("%s\nExample(...): -want, +got:\n%s", tc.reason, diff)
			}
		})
	}
}
