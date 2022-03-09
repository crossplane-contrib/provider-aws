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

package userpool

import (
	"context"
	"reflect"
	"time"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	svcsdk "github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/ratelimiter"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	svcapitypes "github.com/crossplane/provider-aws/apis/cognitoidentityprovider/v1alpha1"
	awsclients "github.com/crossplane/provider-aws/pkg/clients"
)

// SetupUserPool adds a controller that reconciles UserPool.
func SetupUserPool(mgr ctrl.Manager, l logging.Logger, rl workqueue.RateLimiter, poll time.Duration) error {
	name := managed.ControllerName(svcapitypes.UserPoolGroupKind)

	opts := []option{
		func(e *external) {
			e.preObserve = preObserve
			e.postObserve = postObserve
			e.preUpdate = preUpdate
			e.preDelete = preDelete
			e.postCreate = postCreate
			e.isUpToDate = isUpToDate
			e.lateInitialize = lateInitialize
		},
	}
	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(controller.Options{
			RateLimiter: ratelimiter.NewController(rl),
		}).
		For(&svcapitypes.UserPool{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(svcapitypes.UserPoolGroupVersionKind),
			managed.WithInitializers(),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), opts: opts}),
			managed.WithPollInterval(poll),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

func preObserve(_ context.Context, cr *svcapitypes.UserPool, obj *svcsdk.DescribeUserPoolInput) error {
	if meta.GetExternalName(cr) != "" {
		obj.UserPoolId = awsclients.String(meta.GetExternalName(cr))
	}
	return nil
}

func postObserve(_ context.Context, cr *svcapitypes.UserPool, obj *svcsdk.DescribeUserPoolOutput, obs managed.ExternalObservation, err error) (managed.ExternalObservation, error) {
	if err != nil {
		return managed.ExternalObservation{}, err
	}

	cr.SetConditions(xpv1.Available())

	return obs, nil
}

func preUpdate(_ context.Context, cr *svcapitypes.UserPool, obj *svcsdk.UpdateUserPoolInput) error {
	obj.UserPoolId = awsclients.String(meta.GetExternalName(cr))
	return nil
}

func preDelete(_ context.Context, cr *svcapitypes.UserPool, obj *svcsdk.DeleteUserPoolInput) (bool, error) {
	obj.UserPoolId = awsclients.String(meta.GetExternalName(cr))
	return false, nil
}

func postCreate(_ context.Context, cr *svcapitypes.UserPool, obj *svcsdk.CreateUserPoolOutput, obs managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
	if err != nil {
		return managed.ExternalCreation{}, err
	}

	meta.SetExternalName(cr, awsclients.StringValue(obj.UserPool.Id))
	return managed.ExternalCreation{ExternalNameAssigned: true}, nil
}

func isUpToDate(cr *svcapitypes.UserPool, resp *svcsdk.DescribeUserPoolOutput) (bool, error) {
	pool := resp.UserPool

	switch {
	case !areAccountRecoverySettingEqual(cr.Spec.ForProvider.AccountRecoverySetting, pool.AccountRecoverySetting),
		!areAdminCreateUserConfigEqual(cr.Spec.ForProvider.AdminCreateUserConfig, pool.AdminCreateUserConfig),
		!reflect.DeepEqual(cr.Spec.ForProvider.AliasAttributes, pool.AliasAttributes),
		!reflect.DeepEqual(cr.Spec.ForProvider.AutoVerifiedAttributes, pool.AutoVerifiedAttributes),
		!areDeviceConfigurationEqual(cr.Spec.ForProvider.DeviceConfiguration, pool.DeviceConfiguration),
		!areEmailConfigurationEqual(cr.Spec.ForProvider.EmailConfiguration, pool.EmailConfiguration),
		awsclients.StringValue(cr.Spec.ForProvider.EmailVerificationMessage) != awsclients.StringValue(pool.EmailVerificationMessage),
		awsclients.StringValue(cr.Spec.ForProvider.EmailVerificationSubject) != awsclients.StringValue(pool.EmailVerificationSubject),
		!areLambdaConfigEqual(cr.Spec.ForProvider.LambdaConfig, pool.LambdaConfig),
		awsclients.StringValue(cr.Spec.ForProvider.MFAConfiguration) != awsclients.StringValue(pool.MfaConfiguration),
		!arePoliciesEqual(cr.Spec.ForProvider.Policies, pool.Policies),
		!areSchemaEqual(cr.Spec.ForProvider.Schema, pool.SchemaAttributes),
		awsclients.StringValue(cr.Spec.ForProvider.SmsAuthenticationMessage) != awsclients.StringValue(pool.SmsAuthenticationMessage),
		!areSmsConfigurationEqual(cr.Spec.ForProvider.SmsConfiguration, pool.SmsConfiguration),
		awsclients.StringValue(cr.Spec.ForProvider.SmsVerificationMessage) != awsclients.StringValue(pool.SmsVerificationMessage),
		!areUserPoolAddOnsEqual(cr.Spec.ForProvider.UserPoolAddOns, pool.UserPoolAddOns),
		!reflect.DeepEqual(cr.Spec.ForProvider.UsernameAttributes, pool.UsernameAttributes),
		!areUsernameConfigurationEqual(cr.Spec.ForProvider.UsernameConfiguration, pool.UsernameConfiguration),
		!areVerificationMessageTemplateEqual(cr.Spec.ForProvider.VerificationMessageTemplate, pool.VerificationMessageTemplate):
		return false, nil
	}
	return true, nil
}

func areAccountRecoverySettingEqual(spec *svcapitypes.AccountRecoverySettingType, current *svcsdk.AccountRecoverySettingType) bool {
	if spec != nil && current != nil {
		if len(spec.RecoveryMechanisms) > 0 && len(spec.RecoveryMechanisms) != len(current.RecoveryMechanisms) {
			return false
		}

		for i, s := range spec.RecoveryMechanisms {
			switch {
			case awsclients.StringValue(s.Name) != awsclients.StringValue(current.RecoveryMechanisms[i].Name),
				awsclients.Int64Value(s.Priority) != awsclients.Int64Value(current.RecoveryMechanisms[i].Priority):
				return false
			}
		}
	}

	return true
}

func areAdminCreateUserConfigEqual(spec *svcapitypes.AdminCreateUserConfigType, current *svcsdk.AdminCreateUserConfigType) bool {
	if spec != nil && current != nil {
		switch {
		case awsclients.BoolValue(spec.AllowAdminCreateUserOnly) != awsclients.BoolValue(current.AllowAdminCreateUserOnly),
			awsclients.StringValue(spec.InviteMessageTemplate.EmailMessage) != awsclients.StringValue(current.InviteMessageTemplate.EmailMessage),
			awsclients.StringValue(spec.InviteMessageTemplate.EmailSubject) != awsclients.StringValue(current.InviteMessageTemplate.EmailSubject),
			awsclients.StringValue(spec.InviteMessageTemplate.SMSMessage) != awsclients.StringValue(current.InviteMessageTemplate.SMSMessage),
			awsclients.Int64Value(spec.UnusedAccountValidityDays) != awsclients.Int64Value(current.UnusedAccountValidityDays):
			return false
		}
	}
	return true
}

func areDeviceConfigurationEqual(spec *svcapitypes.DeviceConfigurationType, current *svcsdk.DeviceConfigurationType) bool {
	if spec != nil && current != nil {
		switch {
		case awsclients.BoolValue(spec.ChallengeRequiredOnNewDevice) != awsclients.BoolValue(current.ChallengeRequiredOnNewDevice),
			awsclients.BoolValue(spec.DeviceOnlyRememberedOnUserPrompt) != awsclients.BoolValue(current.DeviceOnlyRememberedOnUserPrompt):
			return false
		}
	}
	return true
}

func areEmailConfigurationEqual(spec *svcapitypes.EmailConfigurationType, current *svcsdk.EmailConfigurationType) bool {
	if spec != nil && current != nil {
		switch {
		case awsclients.StringValue(spec.ConfigurationSet) != awsclients.StringValue(current.ConfigurationSet),
			awsclients.StringValue(spec.EmailSendingAccount) != awsclients.StringValue(current.EmailSendingAccount),
			awsclients.StringValue(spec.From) != awsclients.StringValue(current.From),
			awsclients.StringValue(spec.ReplyToEmailAddress) != awsclients.StringValue(current.ReplyToEmailAddress),
			awsclients.StringValue(spec.SourceARN) != awsclients.StringValue(current.SourceArn):
			return false
		}
	}
	return true
}

func areLambdaConfigEqual(spec *svcapitypes.LambdaConfigType, current *svcsdk.LambdaConfigType) bool {
	if spec != nil && current != nil {
		switch {
		case awsclients.StringValue(spec.CreateAuthChallenge) != awsclients.StringValue(current.CreateAuthChallenge),
			awsclients.StringValue(spec.CustomEmailSender.LambdaARN) != awsclients.StringValue(current.CustomEmailSender.LambdaArn),
			awsclients.StringValue(spec.CustomEmailSender.LambdaVersion) != awsclients.StringValue(current.CustomEmailSender.LambdaVersion),
			awsclients.StringValue(spec.CustomMessage) != awsclients.StringValue(current.CustomMessage),
			awsclients.StringValue(spec.CustomSMSSender.LambdaARN) != awsclients.StringValue(current.CustomSMSSender.LambdaArn),
			awsclients.StringValue(spec.CustomSMSSender.LambdaVersion) != awsclients.StringValue(current.CustomSMSSender.LambdaVersion),
			awsclients.StringValue(spec.DefineAuthChallenge) != awsclients.StringValue(current.DefineAuthChallenge),
			awsclients.StringValue(spec.KMSKeyID) != awsclients.StringValue(current.KMSKeyID),
			awsclients.StringValue(spec.PostAuthentication) != awsclients.StringValue(current.PostAuthentication),
			awsclients.StringValue(spec.PostConfirmation) != awsclients.StringValue(current.PostConfirmation),
			awsclients.StringValue(spec.PreAuthentication) != awsclients.StringValue(current.PreAuthentication),
			awsclients.StringValue(spec.PreSignUp) != awsclients.StringValue(current.PreSignUp),
			awsclients.StringValue(spec.PreTokenGeneration) != awsclients.StringValue(current.PreTokenGeneration),
			awsclients.StringValue(spec.UserMigration) != awsclients.StringValue(current.UserMigration),
			awsclients.StringValue(spec.VerifyAuthChallengeResponse) != awsclients.StringValue(current.VerifyAuthChallengeResponse):
			return false
		}
	}
	return true
}

func arePoliciesEqual(spec *svcapitypes.UserPoolPolicyType, current *svcsdk.UserPoolPolicyType) bool {
	if spec != nil && current != nil {
		switch {
		case awsclients.Int64Value(spec.PasswordPolicy.MinimumLength) != awsclients.Int64Value(current.PasswordPolicy.MinimumLength),
			awsclients.BoolValue(spec.PasswordPolicy.RequireLowercase) != awsclients.BoolValue(current.PasswordPolicy.RequireLowercase),
			awsclients.BoolValue(spec.PasswordPolicy.RequireNumbers) != awsclients.BoolValue(current.PasswordPolicy.RequireNumbers),
			awsclients.BoolValue(spec.PasswordPolicy.RequireSymbols) != awsclients.BoolValue(current.PasswordPolicy.RequireSymbols),
			awsclients.BoolValue(spec.PasswordPolicy.RequireUppercase) != awsclients.BoolValue(current.PasswordPolicy.RequireUppercase),
			awsclients.Int64Value(spec.PasswordPolicy.TemporaryPasswordValidityDays) != awsclients.Int64Value(current.PasswordPolicy.TemporaryPasswordValidityDays):
			return false
		}
	}
	return true
}

func areSchemaEqual(spec []*svcapitypes.SchemaAttributeType, current []*svcsdk.SchemaAttributeType) bool {
	if spec != nil && current != nil {
		if len(spec) > 0 && len(spec) != len(current) {
			return false
		}

		for i, s := range spec {
			switch {
			case awsclients.StringValue(s.AttributeDataType) != awsclients.StringValue(current[i].AttributeDataType),
				awsclients.BoolValue(s.DeveloperOnlyAttribute) != awsclients.BoolValue(current[i].DeveloperOnlyAttribute),
				awsclients.BoolValue(s.Mutable) != awsclients.BoolValue(current[i].Mutable),
				awsclients.StringValue(s.Name) != awsclients.StringValue(current[i].Name),
				awsclients.StringValue(s.NumberAttributeConstraints.MaxValue) != awsclients.StringValue(current[i].NumberAttributeConstraints.MaxValue),
				awsclients.StringValue(s.NumberAttributeConstraints.MinValue) != awsclients.StringValue(current[i].NumberAttributeConstraints.MinValue),
				awsclients.BoolValue(s.Required) != awsclients.BoolValue(current[i].Required),
				awsclients.StringValue(s.StringAttributeConstraints.MaxLength) != awsclients.StringValue(current[i].StringAttributeConstraints.MaxLength),
				awsclients.StringValue(s.StringAttributeConstraints.MinLength) != awsclients.StringValue(current[i].StringAttributeConstraints.MinLength):
				return false
			}
		}
	}

	return true
}

func areSmsConfigurationEqual(spec *svcapitypes.SmsConfigurationType, current *svcsdk.SmsConfigurationType) bool {
	if spec != nil && current != nil {
		switch {
		case awsclients.StringValue(spec.ExternalID) != awsclients.StringValue(current.ExternalId),
			awsclients.StringValue(spec.SnsCallerARN) != awsclients.StringValue(current.SnsCallerArn):
			return false
		}
	}
	return true
}

func areUserPoolAddOnsEqual(spec *svcapitypes.UserPoolAddOnsType, current *svcsdk.UserPoolAddOnsType) bool {
	if spec != nil && current != nil {
		return awsclients.StringValue(spec.AdvancedSecurityMode) == awsclients.StringValue(current.AdvancedSecurityMode)
	}
	return true
}

func areUsernameConfigurationEqual(spec *svcapitypes.UsernameConfigurationType, current *svcsdk.UsernameConfigurationType) bool {
	if spec != nil && current != nil {
		return awsclients.BoolValue(spec.CaseSensitive) == awsclients.BoolValue(current.CaseSensitive)
	}
	return true
}

func areVerificationMessageTemplateEqual(spec *svcapitypes.VerificationMessageTemplateType, current *svcsdk.VerificationMessageTemplateType) bool {
	if spec != nil && current != nil {
		switch {
		case awsclients.StringValue(spec.DefaultEmailOption) != awsclients.StringValue(current.DefaultEmailOption),
			awsclients.StringValue(spec.EmailMessage) != awsclients.StringValue(current.EmailMessage),
			awsclients.StringValue(spec.EmailMessageByLink) != awsclients.StringValue(current.EmailMessageByLink),
			awsclients.StringValue(spec.EmailSubject) != awsclients.StringValue(current.EmailSubject),
			awsclients.StringValue(spec.EmailSubjectByLink) != awsclients.StringValue(current.EmailSubjectByLink),
			awsclients.StringValue(spec.SmsMessage) != awsclients.StringValue(current.SmsMessage):
			return false
		}
	}
	return true
}

func lateInitialize(cr *svcapitypes.UserPoolParameters, resp *svcsdk.DescribeUserPoolOutput) error {
	instance := resp.UserPool

	cr.MFAConfiguration = awsclients.LateInitializeStringPtr(cr.MFAConfiguration, instance.MfaConfiguration)
	return nil
}
