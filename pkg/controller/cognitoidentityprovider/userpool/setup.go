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

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"

	svcsdk "github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	svcsdkapi "github.com/aws/aws-sdk-go/service/cognitoidentityprovider/cognitoidentityprovideriface"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/cognitoidentityprovider/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/apis/v1alpha1"
	awsclients "github.com/crossplane-contrib/provider-aws/pkg/clients"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
)

const (
	errMissingMFAStuff     = "no SoftwareTokenMfaConfiguration or SmsConfiguration given, unable to make MFA ON/OPTIONAL"
	errFailedGetMFARequest = "failed GetUserPoolMfaConfig request. Could not check MFAConfiguration isUptoDate-state"
	errFailedSetMFARequest = "failed SetUserPoolMfaConfig request. Could not update UserPool"
	errConflictingFields   = "fields conflicting! Please only use one of them or both with the same value"
)

// SetupUserPool adds a controller that reconciles UserPool.
func SetupUserPool(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(svcapitypes.UserPoolGroupKind)

	opts := []option{
		func(e *external) {
			h := &hooks{client: e.client}
			e.preObserve = preObserve
			e.postObserve = postObserve
			e.preUpdate = h.preUpdate
			e.preDelete = preDelete
			e.preCreate = preCreate
			e.postCreate = postCreate
			e.isUpToDate = h.isUpToDate
			e.lateInitialize = lateInitialize
		},
	}

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), v1alpha1.StoreConfigGroupVersionKind))
	}

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&svcapitypes.UserPool{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(svcapitypes.UserPoolGroupVersionKind),
			managed.WithInitializers(),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), opts: opts}),
			managed.WithPollInterval(o.PollInterval),
			managed.WithLogger(o.Logger.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
			managed.WithConnectionPublishers(cps...)))
}

type hooks struct {
	client svcsdkapi.CognitoIdentityProviderAPI
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

func (e *hooks) preUpdate(ctx context.Context, cr *svcapitypes.UserPool, obj *svcsdk.UpdateUserPoolInput) error {
	obj.UserPoolId = awsclients.String(meta.GetExternalName(cr))

	// "Cannot turn MFA functionality ON, once the user pool has been created"
	// -> concerns UpdateUserPool, not SetUserPoolMfaConfig
	// therefore, before Update request, set MFA configuration
	return e.setMfaConfiguration(ctx, cr)
}

func preDelete(_ context.Context, cr *svcapitypes.UserPool, obj *svcsdk.DeleteUserPoolInput) (bool, error) {
	obj.UserPoolId = awsclients.String(meta.GetExternalName(cr))
	return false, nil
}

func preCreate(_ context.Context, cr *svcapitypes.UserPool, obj *svcsdk.CreateUserPoolInput) error {
	// for Creation need to set MFA to OFF,
	// bc if MFA ON and no SmsConfiguration provided, AWS throws error
	// in first Update, we can use SetUserPoolMfaConfig to set all MFA stuff (e.g. Token)
	obj.MfaConfiguration = awsclients.String(svcsdk.UserPoolMfaTypeOff)

	return nil
}

func postCreate(_ context.Context, cr *svcapitypes.UserPool, obj *svcsdk.CreateUserPoolOutput, obs managed.ExternalCreation, err error) (managed.ExternalCreation, error) {
	if err != nil {
		return managed.ExternalCreation{}, err
	}
	meta.SetExternalName(cr, awsclients.StringValue(obj.UserPool.Id))

	// we cannot do a SetUserPoolMfaConfig-call here, but have to wait until first Update,
	// bc in zz_controller.go/Create all cr.specs.forProvider are set to obj.Userpool values
	// (->so no knowledge here of actual user input)
	return managed.ExternalCreation{ExternalNameAssigned: true}, nil
}

func (e *hooks) isUpToDate(cr *svcapitypes.UserPool, resp *svcsdk.DescribeUserPoolOutput) (bool, error) {
	pool := resp.UserPool
	spec := cr.Spec.ForProvider

	switch {
	case !areAccountRecoverySettingEqual(spec.AccountRecoverySetting, pool.AccountRecoverySetting),
		!areAdminCreateUserConfigEqual(spec.AdminCreateUserConfig, pool.AdminCreateUserConfig),
		!reflect.DeepEqual(spec.AutoVerifiedAttributes, pool.AutoVerifiedAttributes),
		!areDeviceConfigurationEqual(spec.DeviceConfiguration, pool.DeviceConfiguration),
		!areEmailConfigurationEqual(spec.EmailConfiguration, pool.EmailConfiguration),
		!areLambdaConfigEqual(spec.LambdaConfig, pool.LambdaConfig),
		!arePoliciesEqual(spec.Policies, pool.Policies),
		!areSchemaEqual(spec.Schema, pool.SchemaAttributes),
		awsclients.StringValue(spec.SmsAuthenticationMessage) != awsclients.StringValue(pool.SmsAuthenticationMessage),
		!areSmsConfigurationEqual(spec.SmsConfiguration, pool.SmsConfiguration),
		!areUserPoolAddOnsEqual(spec.UserPoolAddOns, pool.UserPoolAddOns),
		!areVerificationMessageTemplateEqual(spec.VerificationMessageTemplate, pool.VerificationMessageTemplate),
		!reflect.DeepEqual(spec.UserPoolTags, pool.UserPoolTags):
		return false, nil
	}

	// check the conflicting fields for isUpToDate + conflicts
	fieldsUpToDate, err := conflictingFieldsEqual(spec, pool)
	if err != nil || !fieldsUpToDate {
		return false, err
	}

	// check MFA stuff
	return e.areMFAConfigEqual(cr)
}

func areAccountRecoverySettingEqual(spec *svcapitypes.AccountRecoverySettingType, current *svcsdk.AccountRecoverySettingType) bool {
	if spec != nil {
		if current == nil {
			return false
		}
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
			!areInviteMessageTemplateEqual(spec.InviteMessageTemplate, current.InviteMessageTemplate):
			return false
		}
	}
	return true
}

func areInviteMessageTemplateEqual(spec *svcapitypes.MessageTemplateType, current *svcsdk.MessageTemplateType) bool {
	if spec != nil {
		if current == nil {
			return false
		}
		switch {
		case awsclients.StringValue(spec.EmailMessage) != awsclients.StringValue(current.EmailMessage),
			awsclients.StringValue(spec.EmailSubject) != awsclients.StringValue(current.EmailSubject),
			awsclients.StringValue(spec.SMSMessage) != awsclients.StringValue(current.SMSMessage):
			return false
		}
	}
	return true
}

func areDeviceConfigurationEqual(spec *svcapitypes.DeviceConfigurationType, current *svcsdk.DeviceConfigurationType) bool {
	if spec != nil {
		if current == nil {
			return false
		}
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
			!areCustomEmailSenderEqual(spec.CustomEmailSender, current.CustomEmailSender),
			awsclients.StringValue(spec.CustomMessage) != awsclients.StringValue(current.CustomMessage),
			!areCustomSMSSenderEqual(spec.CustomSMSSender, current.CustomSMSSender),
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

func areCustomEmailSenderEqual(spec *svcapitypes.CustomEmailLambdaVersionConfigType, current *svcsdk.CustomEmailLambdaVersionConfigType) bool {
	if spec != nil && current != nil {
		switch {
		case awsclients.StringValue(spec.LambdaARN) != awsclients.StringValue(current.LambdaArn),
			awsclients.StringValue(spec.LambdaVersion) != awsclients.StringValue(current.LambdaVersion):
			return false
		}
	}
	return true
}

func areCustomSMSSenderEqual(spec *svcapitypes.CustomSMSLambdaVersionConfigType, current *svcsdk.CustomSMSLambdaVersionConfigType) bool {
	if spec != nil && current != nil {
		switch {
		case awsclients.StringValue(spec.LambdaARN) != awsclients.StringValue(current.LambdaArn),
			awsclients.StringValue(spec.LambdaVersion) != awsclients.StringValue(current.LambdaVersion):
			return false
		}
	}
	return true
}

func arePoliciesEqual(spec *svcapitypes.UserPoolPolicyType, current *svcsdk.UserPoolPolicyType) bool {
	if spec != nil && current != nil && spec.PasswordPolicy != nil && current.PasswordPolicy != nil {
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
	if spec != nil {
		if current == nil {
			return false
		}
		switch {
		case awsclients.StringValue(spec.ExternalID) != awsclients.StringValue(current.ExternalId),
			awsclients.StringValue(spec.SNSCallerARN) != awsclients.StringValue(current.SnsCallerArn):
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

func conflictingFieldsEqual(params svcapitypes.UserPoolParameters, pool *svcsdk.UserPoolType) (bool, error) {
	// conflicting fields, that require the user to
	// either set them both with exactly the same value
	// or to just provide one of them and leave the other on empty

	// should never be nil, bc set in lateInit, but just to be safe
	if params.VerificationMessageTemplate != nil {
		// check for conflicts and isUpTodates
		fieldUpToDate, err := conflictingFieldsHelper(params.EmailVerificationMessage, params.VerificationMessageTemplate.EmailMessage, pool.EmailVerificationMessage)
		if err != nil {
			return true, errors.Wrap(err, "EmailVerificationMessage and verificationMessageTemplate.EmailMessage")
		}
		if !fieldUpToDate {
			return false, nil
		}
		fieldUpToDate, err = conflictingFieldsHelper(params.EmailVerificationSubject, params.VerificationMessageTemplate.EmailSubject, pool.EmailVerificationSubject)
		if err != nil {
			return true, errors.Wrap(err, "EmailVerificationSubject and verificationMessageTemplate.EmailSubject")
		}
		if !fieldUpToDate {
			return false, nil
		}
		fieldUpToDate, err = conflictingFieldsHelper(params.SmsVerificationMessage, params.VerificationMessageTemplate.SmsMessage, pool.SmsVerificationMessage)
		if err != nil {
			return true, errors.Wrap(err, "SmsVerificationMessage and verificationMessageTemplate.SmsMessage")
		}
		if !fieldUpToDate {
			return false, nil
		}
	}

	return true, nil
}

// conflictingFieldsHelper checks if 2 *string fields conflict and if their value isUpToDate
func conflictingFieldsHelper(field1 *string, field2 *string, fieldAWS *string) (bool, error) {
	// both of them nil => all fine
	if field1 == nil && field2 == nil {
		return true, nil
	}
	if field1 != nil && field2 != nil {
		if awsclients.StringValue(field1) != awsclients.StringValue(field2) {
			// both of them non-nil and different => means conflict
			return true, errors.New(errConflictingFields)
		}
		// both of them non-nil, but same => check if value isUpToDate
		if awsclients.StringValue(field1) != awsclients.StringValue(fieldAWS) {

			return false, nil
		}

		return true, nil
	}
	// check which one is non-nil and if its value isUpToDate
	if field1 != nil {
		return awsclients.StringValue(field1) == awsclients.StringValue(fieldAWS), nil
	}
	return awsclients.StringValue(field2) == awsclients.StringValue(fieldAWS), nil
}

func areVerificationMessageTemplateEqual(spec *svcapitypes.VerificationMessageTemplateType, current *svcsdk.VerificationMessageTemplateType) bool {
	if spec != nil && current != nil {
		switch { // EmailMessage, EmailSubject, SmsMessage are checked for in conflictingFieldsEqual
		case awsclients.StringValue(spec.DefaultEmailOption) != awsclients.StringValue(current.DefaultEmailOption),
			awsclients.StringValue(spec.EmailMessageByLink) != awsclients.StringValue(current.EmailMessageByLink),
			awsclients.StringValue(spec.EmailSubjectByLink) != awsclients.StringValue(current.EmailSubjectByLink):
			return false
		}
	}
	return true
}

func (e *hooks) areMFAConfigEqual(cr *svcapitypes.UserPool) (bool, error) {

	out, err := e.client.GetUserPoolMfaConfig(&svcsdk.GetUserPoolMfaConfigInput{UserPoolId: awsclients.String(meta.GetExternalName(cr))})
	if err != nil {
		return true, errors.Wrap(err, errFailedGetMFARequest)
	}

	// out should not be nil, bc we set MFA to OFF in preCreate
	if awsclients.StringValue(cr.Spec.ForProvider.MFAConfiguration) != awsclients.StringValue(out.MfaConfiguration) {
		return false, nil
	}

	// only check MFAConfig stuff, if MFAConfiguration is ON/OPTIONAL,
	// bc AWS doesn't allow setting stuff in SetUserPoolMfaConfig, if MFA is OFF
	// (-> e.g. Token enabled in specs with MFA OFF)
	if awsclients.StringValue(cr.Spec.ForProvider.MFAConfiguration) != svcsdk.UserPoolMfaTypeOff {
		if cr.Spec.ForProvider.SoftwareTokenMFAConfiguration != nil && out.SoftwareTokenMfaConfiguration != nil {
			return awsclients.BoolValue(cr.Spec.ForProvider.SoftwareTokenMFAConfiguration.Enabled) == awsclients.BoolValue(out.SoftwareTokenMfaConfiguration.Enabled), nil
		}
		// no need to check SmsMfaConfiguration here,
		// bc currently it is 100% overlapping with SmsConfiguration and SmsAuthenticationMessage,
		// which are checked in other places
		// if in future there is change in this API structure (e.g. fields separation),
		// here would then be potentially the place for SmsMfaConfiguration check
	}

	return true, nil
}

func lateInitialize(cr *svcapitypes.UserPoolParameters, resp *svcsdk.DescribeUserPoolOutput) error {
	instance := resp.UserPool

	cr.MFAConfiguration = awsclients.LateInitializeStringPtr(cr.MFAConfiguration, instance.MfaConfiguration)

	if instance.AdminCreateUserConfig != nil {
		if cr.AdminCreateUserConfig == nil {
			cr.AdminCreateUserConfig = &svcapitypes.AdminCreateUserConfigType{}
		}
		cr.AdminCreateUserConfig.AllowAdminCreateUserOnly = awsclients.LateInitializeBoolPtr(cr.AdminCreateUserConfig.AllowAdminCreateUserOnly, instance.AdminCreateUserConfig.AllowAdminCreateUserOnly)
	}

	if instance.EmailConfiguration != nil {
		if cr.EmailConfiguration == nil {
			cr.EmailConfiguration = &svcapitypes.EmailConfigurationType{}
		}
		cr.EmailConfiguration.EmailSendingAccount = awsclients.LateInitializeStringPtr(cr.EmailConfiguration.EmailSendingAccount, instance.EmailConfiguration.EmailSendingAccount)
	}

	if instance.Policies != nil {
		if cr.Policies == nil {
			cr.Policies = &svcapitypes.UserPoolPolicyType{PasswordPolicy: &svcapitypes.PasswordPolicyType{}}
		}
		if instance.Policies.PasswordPolicy != nil {
			cr.Policies.PasswordPolicy.MinimumLength = awsclients.LateInitializeInt64Ptr(cr.Policies.PasswordPolicy.MinimumLength, instance.Policies.PasswordPolicy.MinimumLength)
			cr.Policies.PasswordPolicy.RequireLowercase = awsclients.LateInitializeBoolPtr(cr.Policies.PasswordPolicy.RequireLowercase, instance.Policies.PasswordPolicy.RequireLowercase)
			cr.Policies.PasswordPolicy.RequireNumbers = awsclients.LateInitializeBoolPtr(cr.Policies.PasswordPolicy.RequireNumbers, instance.Policies.PasswordPolicy.RequireNumbers)
			cr.Policies.PasswordPolicy.RequireSymbols = awsclients.LateInitializeBoolPtr(cr.Policies.PasswordPolicy.RequireSymbols, instance.Policies.PasswordPolicy.RequireSymbols)
			cr.Policies.PasswordPolicy.RequireUppercase = awsclients.LateInitializeBoolPtr(cr.Policies.PasswordPolicy.RequireUppercase, instance.Policies.PasswordPolicy.RequireUppercase)
			cr.Policies.PasswordPolicy.TemporaryPasswordValidityDays = awsclients.LateInitializeInt64Ptr(cr.Policies.PasswordPolicy.TemporaryPasswordValidityDays, instance.Policies.PasswordPolicy.TemporaryPasswordValidityDays)
		}
	}

	if instance.VerificationMessageTemplate != nil {
		if cr.VerificationMessageTemplate == nil {
			cr.VerificationMessageTemplate = &svcapitypes.VerificationMessageTemplateType{}
		}
		cr.VerificationMessageTemplate.DefaultEmailOption = awsclients.LateInitializeStringPtr(cr.VerificationMessageTemplate.DefaultEmailOption, instance.VerificationMessageTemplate.DefaultEmailOption)
	}

	// Info: to avoid redundancy+problems, do not lateInit conflicting fields
	// (e.g. VerificationMessageTemplate.SmsMessage & SmsVerificationMessage)

	return nil
}

// setMfaConfiguration sets the MFA configuration with a SetUserPoolMfaConfigWithContext-Request
func (e *hooks) setMfaConfiguration(ctx context.Context, cr *svcapitypes.UserPool) error {
	// set MFA configuration (only allowed by AWS when MFA not OFF:
	// -> "Invalid MFA configuration given, can't turn off MFA and configure an MFA together")
	if awsclients.StringValue(cr.Spec.ForProvider.MFAConfiguration) != svcsdk.UserPoolMfaTypeOff {
		mfaConfig := &svcsdk.SetUserPoolMfaConfigInput{
			MfaConfiguration: cr.Spec.ForProvider.MFAConfiguration,
			UserPoolId:       awsclients.String(meta.GetExternalName(cr)),
		}

		// even without setting it here,
		// if smsConfiguration is provided and MFA is turned ON/OPTIONAL,
		// AWS will automatically use SMS as a possible MFA method
		if cr.Spec.ForProvider.SmsConfiguration != nil {
			mfaConfig.SmsMfaConfiguration = &svcsdk.SmsMfaConfigType{
				SmsAuthenticationMessage: cr.Spec.ForProvider.SmsAuthenticationMessage,
				SmsConfiguration: &svcsdk.SmsConfigurationType{
					ExternalId:   cr.Spec.ForProvider.SmsConfiguration.ExternalID,
					SnsCallerArn: cr.Spec.ForProvider.SmsConfiguration.SNSCallerARN,
				},
			}
		}

		if cr.Spec.ForProvider.SoftwareTokenMFAConfiguration != nil {
			mfaConfig.SoftwareTokenMfaConfiguration = &svcsdk.SoftwareTokenMfaConfigType{
				Enabled: cr.Spec.ForProvider.SoftwareTokenMFAConfiguration.Enabled,
			}
		}

		// custom error here needed,
		// bc of our setting/handling of MfaConfiguration (SetUserPoolMfaConfig + UpdateUserPool)
		if mfaConfig.SmsMfaConfiguration == nil && mfaConfig.SoftwareTokenMfaConfiguration == nil {
			return errors.New(errMissingMFAStuff)
		}

		_, err := e.client.SetUserPoolMfaConfigWithContext(ctx, mfaConfig)
		if err != nil {
			return errors.Wrap(err, errFailedSetMFARequest)
		}
	}

	return nil
}
