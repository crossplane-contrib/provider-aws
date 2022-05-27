package userpool

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

type functionModifier func(*svcapitypes.UserPool)

func withSpec(p svcapitypes.UserPoolParameters) functionModifier {
	return func(r *svcapitypes.UserPool) { r.Spec.ForProvider = p }
}

func withObservation(s svcapitypes.UserPoolObservation) functionModifier {
	return func(r *svcapitypes.UserPool) { r.Status.AtProvider = s }
}

func withExternalName(v string) functionModifier {
	return func(r *svcapitypes.UserPool) {
		meta.SetExternalName(r, v)
	}
}

func userPool(m ...functionModifier) *svcapitypes.UserPool {
	cr := &svcapitypes.UserPool{}
	cr.Name = "test-userpool-name"
	cr.Spec.ForProvider.PoolName = &cr.Name
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
		cr   *svcapitypes.UserPool
		resp *svcsdk.DescribeUserPoolOutput
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
				cr:   userPool(withSpec(svcapitypes.UserPoolParameters{})),
				resp: &svcsdk.DescribeUserPoolOutput{UserPool: &svcsdk.UserPoolType{}},
			},
			want: want{
				result: true,
				err:    nil,
			},
		},
		"ChangedAccountRecoverySetting": {
			args: args{
				cr: userPool(withSpec(svcapitypes.UserPoolParameters{
					AccountRecoverySetting: &svcapitypes.AccountRecoverySettingType{
						RecoveryMechanisms: []*svcapitypes.RecoveryOptionType{
							{
								Name: &testString1,
							},
						},
					},
				})),
				resp: &svcsdk.DescribeUserPoolOutput{UserPool: &svcsdk.UserPoolType{
					AccountRecoverySetting: &svcsdk.AccountRecoverySettingType{
						RecoveryMechanisms: []*svcsdk.RecoveryOptionType{
							{
								Name: &testString2,
							},
						},
					},
				}},
			},
			want: want{
				result: false,
				err:    nil,
			},
		},
		"ChangedAdminCreateUserConfig": {
			args: args{
				cr: userPool(withSpec(svcapitypes.UserPoolParameters{
					AdminCreateUserConfig: &svcapitypes.AdminCreateUserConfigType{
						InviteMessageTemplate: &svcapitypes.MessageTemplateType{
							EmailMessage: &testString1,
						},
					},
				})),
				resp: &svcsdk.DescribeUserPoolOutput{UserPool: &svcsdk.UserPoolType{
					AdminCreateUserConfig: &svcsdk.AdminCreateUserConfigType{
						InviteMessageTemplate: &svcsdk.MessageTemplateType{
							EmailMessage: &testString2,
						},
					},
				}},
			},
			want: want{
				result: false,
				err:    nil,
			},
		},
		"ChangedAliasAttributes": {
			args: args{
				cr: userPool(withSpec(svcapitypes.UserPoolParameters{
					AliasAttributes: []*string{&testString1},
				})),
				resp: &svcsdk.DescribeUserPoolOutput{UserPool: &svcsdk.UserPoolType{
					AliasAttributes: []*string{&testString2},
				}},
			},
			want: want{
				result: false,
				err:    nil,
			},
		},
		"ChangedAutoVerifiedAttributes": {
			args: args{
				cr: userPool(withSpec(svcapitypes.UserPoolParameters{
					AutoVerifiedAttributes: []*string{&testString1},
				})),
				resp: &svcsdk.DescribeUserPoolOutput{UserPool: &svcsdk.UserPoolType{
					AutoVerifiedAttributes: []*string{&testString2},
				}},
			},
			want: want{
				result: false,
				err:    nil,
			},
		},
		"ChangedDeviceConfiguration": {
			args: args{
				cr: userPool(withSpec(svcapitypes.UserPoolParameters{
					DeviceConfiguration: &svcapitypes.DeviceConfigurationType{
						ChallengeRequiredOnNewDevice: &testBool1,
					},
				})),
				resp: &svcsdk.DescribeUserPoolOutput{UserPool: &svcsdk.UserPoolType{
					DeviceConfiguration: &svcsdk.DeviceConfigurationType{
						ChallengeRequiredOnNewDevice: &testBool2,
					},
				}},
			},
			want: want{
				result: false,
				err:    nil,
			},
		},
		"ChangedEmailConfiguration": {
			args: args{
				cr: userPool(withSpec(svcapitypes.UserPoolParameters{
					EmailConfiguration: &svcapitypes.EmailConfigurationType{
						From: &testString1,
					},
				})),
				resp: &svcsdk.DescribeUserPoolOutput{UserPool: &svcsdk.UserPoolType{
					EmailConfiguration: &svcsdk.EmailConfigurationType{
						From: &testString2,
					},
				}},
			},
			want: want{
				result: false,
				err:    nil,
			},
		},
		"ChangedEmailVerificationMessage": {
			args: args{
				cr: userPool(withSpec(svcapitypes.UserPoolParameters{
					EmailVerificationMessage: &testString1,
				})),
				resp: &svcsdk.DescribeUserPoolOutput{UserPool: &svcsdk.UserPoolType{
					EmailVerificationMessage: &testString2,
				}},
			},
			want: want{
				result: false,
				err:    nil,
			},
		},
		"ChangedEmailVerificationSubject": {
			args: args{
				cr: userPool(withSpec(svcapitypes.UserPoolParameters{
					EmailVerificationSubject: &testString1,
				})),
				resp: &svcsdk.DescribeUserPoolOutput{UserPool: &svcsdk.UserPoolType{
					EmailVerificationSubject: &testString2,
				}},
			},
			want: want{
				result: false,
				err:    nil,
			},
		},
		"ChangedLambdaConfig": {
			args: args{
				cr: userPool(withSpec(svcapitypes.UserPoolParameters{
					LambdaConfig: &svcapitypes.LambdaConfigType{
						CustomEmailSender: &svcapitypes.CustomEmailLambdaVersionConfigType{
							LambdaARN: &testString1,
						},
					},
				})),
				resp: &svcsdk.DescribeUserPoolOutput{UserPool: &svcsdk.UserPoolType{
					LambdaConfig: &svcsdk.LambdaConfigType{
						CustomEmailSender: &svcsdk.CustomEmailLambdaVersionConfigType{
							LambdaArn: &testString2,
						},
					},
				}},
			},
			want: want{
				result: false,
				err:    nil,
			},
		},
		"ChangedMFAConfiguration": {
			args: args{
				cr: userPool(withSpec(svcapitypes.UserPoolParameters{
					MFAConfiguration: &testString1,
				})),
				resp: &svcsdk.DescribeUserPoolOutput{UserPool: &svcsdk.UserPoolType{
					MfaConfiguration: &testString2,
				}},
			},
			want: want{
				result: false,
				err:    nil,
			},
		},
		"ChangedPolicies": {
			args: args{
				cr: userPool(withSpec(svcapitypes.UserPoolParameters{
					Policies: &svcapitypes.UserPoolPolicyType{
						PasswordPolicy: &svcapitypes.PasswordPolicyType{
							MinimumLength: &testNumber,
						},
					},
				})),
				resp: &svcsdk.DescribeUserPoolOutput{UserPool: &svcsdk.UserPoolType{
					Policies: &svcsdk.UserPoolPolicyType{
						PasswordPolicy: &svcsdk.PasswordPolicyType{
							MinimumLength: &testNumberChanged,
						},
					},
				}},
			},
			want: want{
				result: false,
				err:    nil,
			},
		},
		"ChangedSchema": {
			args: args{
				cr: userPool(withSpec(svcapitypes.UserPoolParameters{
					Schema: []*svcapitypes.SchemaAttributeType{
						{
							NumberAttributeConstraints: &svcapitypes.NumberAttributeConstraintsType{
								MaxValue: &testString1,
							},
						},
					},
				})),
				resp: &svcsdk.DescribeUserPoolOutput{UserPool: &svcsdk.UserPoolType{
					SchemaAttributes: []*svcsdk.SchemaAttributeType{
						{
							NumberAttributeConstraints: &svcsdk.NumberAttributeConstraintsType{
								MaxValue: &testString2,
							},
						},
					},
				}},
			},
			want: want{
				result: false,
				err:    nil,
			},
		},
		"ChangedSmsAuthenticationMessage": {
			args: args{
				cr: userPool(withSpec(svcapitypes.UserPoolParameters{
					SmsAuthenticationMessage: &testString1,
				})),
				resp: &svcsdk.DescribeUserPoolOutput{UserPool: &svcsdk.UserPoolType{
					SmsAuthenticationMessage: &testString2,
				}},
			},
			want: want{
				result: false,
				err:    nil,
			},
		},
		"ChangedSmsConfiguration": {
			args: args{
				cr: userPool(withSpec(svcapitypes.UserPoolParameters{
					SmsConfiguration: &svcapitypes.SmsConfigurationType{
						ExternalID: &testString1,
					},
				})),
				resp: &svcsdk.DescribeUserPoolOutput{UserPool: &svcsdk.UserPoolType{
					SmsConfiguration: &svcsdk.SmsConfigurationType{
						ExternalId: &testString2,
					},
				}},
			},
			want: want{
				result: false,
				err:    nil,
			},
		},
		"ChangedSmsVerificationMessage": {
			args: args{
				cr: userPool(withSpec(svcapitypes.UserPoolParameters{
					SmsVerificationMessage: &testString1,
				})),
				resp: &svcsdk.DescribeUserPoolOutput{UserPool: &svcsdk.UserPoolType{
					SmsVerificationMessage: &testString2,
				}},
			},
			want: want{
				result: false,
				err:    nil,
			},
		},
		"ChangedUserPoolAddOns": {
			args: args{
				cr: userPool(withSpec(svcapitypes.UserPoolParameters{
					UserPoolAddOns: &svcapitypes.UserPoolAddOnsType{
						AdvancedSecurityMode: &testString1,
					},
				})),
				resp: &svcsdk.DescribeUserPoolOutput{UserPool: &svcsdk.UserPoolType{
					UserPoolAddOns: &svcsdk.UserPoolAddOnsType{
						AdvancedSecurityMode: &testString2,
					},
				}},
			},
			want: want{
				result: false,
				err:    nil,
			},
		},
		"ChangedUsernameAttributes": {
			args: args{
				cr: userPool(withSpec(svcapitypes.UserPoolParameters{
					UsernameAttributes: []*string{&testString1},
				})),
				resp: &svcsdk.DescribeUserPoolOutput{UserPool: &svcsdk.UserPoolType{
					UsernameAttributes: []*string{&testString2},
				}},
			},
			want: want{
				result: false,
				err:    nil,
			},
		},
		"ChangedUsernameConfiguration": {
			args: args{
				cr: userPool(withSpec(svcapitypes.UserPoolParameters{
					UsernameConfiguration: &svcapitypes.UsernameConfigurationType{
						CaseSensitive: &testBool1,
					},
				})),
				resp: &svcsdk.DescribeUserPoolOutput{UserPool: &svcsdk.UserPoolType{
					UsernameConfiguration: &svcsdk.UsernameConfigurationType{
						CaseSensitive: &testBool2,
					},
				}},
			},
			want: want{
				result: false,
				err:    nil,
			},
		},
		"ChangedVerificationMessageTemplate": {
			args: args{
				cr: userPool(withSpec(svcapitypes.UserPoolParameters{
					VerificationMessageTemplate: &svcapitypes.VerificationMessageTemplateType{
						DefaultEmailOption: &testString1,
					},
				})),
				resp: &svcsdk.DescribeUserPoolOutput{UserPool: &svcsdk.UserPoolType{
					VerificationMessageTemplate: &svcsdk.VerificationMessageTemplateType{
						DefaultEmailOption: &testString2,
					},
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

func TestLateInitialize(t *testing.T) {
	type args struct {
		cr   *svcapitypes.UserPoolParameters
		resp *svcsdk.DescribeUserPoolOutput
	}

	type want struct {
		result *svcapitypes.UserPoolParameters
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"NoLateInitialization": {
			args: args{
				cr: &svcapitypes.UserPoolParameters{
					MFAConfiguration: &testString1,
				},
				resp: &svcsdk.DescribeUserPoolOutput{
					UserPool: &svcsdk.UserPoolType{
						MfaConfiguration: &testString2,
					},
				},
			},
			want: want{
				result: &svcapitypes.UserPoolParameters{
					MFAConfiguration: &testString1,
				},
				err: nil,
			},
		},
		"LateInitializeMFAConfiguration": {
			args: args{
				cr: &svcapitypes.UserPoolParameters{},
				resp: &svcsdk.DescribeUserPoolOutput{
					UserPool: &svcsdk.UserPoolType{
						MfaConfiguration: &testString2,
					},
				},
			},
			want: want{
				result: &svcapitypes.UserPoolParameters{
					MFAConfiguration: &testString2,
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
		cr  *svcapitypes.UserPool
		obj *svcsdk.CreateUserPoolOutput
		err error
	}

	type want struct {
		cr     *svcapitypes.UserPool
		result managed.ExternalCreation
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"SetExternalNameAndConnectionDetails": {
			args: args{
				cr: userPool(
					withSpec(svcapitypes.UserPoolParameters{}),
					withObservation(svcapitypes.UserPoolObservation{
						ID: &testString1,
					}),
				),
				obj: &svcsdk.CreateUserPoolOutput{
					UserPool: &svcsdk.UserPoolType{
						Id: &testString1,
					},
				},
				err: nil,
			},
			want: want{
				cr: userPool(
					withSpec(svcapitypes.UserPoolParameters{}),
					withObservation(svcapitypes.UserPoolObservation{
						ID: &testString1,
					}),
					withExternalName(testString1),
				),
				result: managed.ExternalCreation{
					ExternalNameAssigned: true,
				},
				err: nil,
			},
		},
		"FailedCreation": {
			args: args{
				cr: userPool(
					withSpec(svcapitypes.UserPoolParameters{}),
				),
				obj: nil,
				err: errBoom,
			},
			want: want{
				cr: userPool(
					withSpec(svcapitypes.UserPoolParameters{}),
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
