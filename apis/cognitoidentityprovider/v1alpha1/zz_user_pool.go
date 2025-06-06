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

// Code generated by ack-generate. DO NOT EDIT.

package v1alpha1

import (
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// UserPoolParameters defines the desired state of UserPool
type UserPoolParameters struct {
	// Region is which region the UserPool will be created.
	// +kubebuilder:validation:Required
	Region string `json:"region"`
	// The available verified method a user can use to recover their password when
	// they call ForgotPassword. You can use this setting to define a preferred
	// method when a user has more than one method available. With this setting,
	// SMS doesn't qualify for a valid password recovery mechanism if the user also
	// has SMS multi-factor authentication (MFA) activated. In the absence of this
	// setting, Amazon Cognito uses the legacy behavior to determine the recovery
	// method where SMS is preferred through email.
	AccountRecoverySetting *AccountRecoverySettingType `json:"accountRecoverySetting,omitempty"`
	// The configuration for AdminCreateUser requests.
	AdminCreateUserConfig *AdminCreateUserConfigType `json:"adminCreateUserConfig,omitempty"`
	// Attributes supported as an alias for this user pool. Possible values: phone_number,
	// email, or preferred_username.
	AliasAttributes []*string `json:"aliasAttributes,omitempty"`
	// The attributes to be auto-verified. Possible values: email, phone_number.
	AutoVerifiedAttributes []*string `json:"autoVerifiedAttributes,omitempty"`
	// When active, DeletionProtection prevents accidental deletion of your user
	// pool. Before you can delete a user pool that you have protected against deletion,
	// you must deactivate this feature.
	//
	// When you try to delete a protected user pool in a DeleteUserPool API request,
	// Amazon Cognito returns an InvalidParameterException error. To delete a protected
	// user pool, send a new DeleteUserPool request after you deactivate deletion
	// protection in an UpdateUserPool API request.
	DeletionProtection *string `json:"deletionProtection,omitempty"`
	// The device-remembering configuration for a user pool. A null value indicates
	// that you have deactivated device remembering in your user pool.
	//
	// When you provide a value for any DeviceConfiguration field, you activate
	// the Amazon Cognito device-remembering feature.
	DeviceConfiguration *DeviceConfigurationType `json:"deviceConfiguration,omitempty"`
	// The email configuration of your user pool. The email configuration type sets
	// your preferred sending method, Amazon Web Services Region, and sender for
	// messages from your user pool.
	EmailConfiguration *EmailConfigurationType `json:"emailConfiguration,omitempty"`
	// This parameter is no longer used. See VerificationMessageTemplateType (https://docs.aws.amazon.com/cognito-user-identity-pools/latest/APIReference/API_VerificationMessageTemplateType.html).
	EmailVerificationMessage *string `json:"emailVerificationMessage,omitempty"`
	// This parameter is no longer used. See VerificationMessageTemplateType (https://docs.aws.amazon.com/cognito-user-identity-pools/latest/APIReference/API_VerificationMessageTemplateType.html).
	EmailVerificationSubject *string `json:"emailVerificationSubject,omitempty"`
	// The Lambda trigger configuration information for the new user pool.
	//
	// In a push model, event sources (such as Amazon S3 and custom applications)
	// need permission to invoke a function. So you must make an extra call to add
	// permission for these event sources to invoke your Lambda function.
	//
	// For more information on using the Lambda API to add permission, see AddPermission
	// (https://docs.aws.amazon.com/lambda/latest/dg/API_AddPermission.html).
	//
	// For adding permission using the CLI, see add-permission (https://docs.aws.amazon.com/cli/latest/reference/lambda/add-permission.html).
	LambdaConfig *LambdaConfigType `json:"lambdaConfig,omitempty"`
	// Specifies MFA configuration details.
	MFAConfiguration *string `json:"mfaConfiguration,omitempty"`
	// The policies associated with the new user pool.
	Policies *UserPoolPolicyType `json:"policies,omitempty"`
	// A string used to name the user pool.
	// +kubebuilder:validation:Required
	PoolName *string `json:"poolName"`
	// An array of schema attributes for the new user pool. These attributes can
	// be standard or custom attributes.
	Schema []*SchemaAttributeType `json:"schema,omitempty"`
	// A string representing the SMS authentication message.
	SmsAuthenticationMessage *string `json:"smsAuthenticationMessage,omitempty"`
	// The SMS configuration with the settings that your Amazon Cognito user pool
	// must use to send an SMS message from your Amazon Web Services account through
	// Amazon Simple Notification Service. To send SMS messages with Amazon SNS
	// in the Amazon Web Services Region that you want, the Amazon Cognito user
	// pool uses an Identity and Access Management (IAM) role in your Amazon Web
	// Services account.
	SmsConfiguration *SmsConfigurationType `json:"smsConfiguration,omitempty"`
	// This parameter is no longer used. See VerificationMessageTemplateType (https://docs.aws.amazon.com/cognito-user-identity-pools/latest/APIReference/API_VerificationMessageTemplateType.html).
	SmsVerificationMessage *string `json:"smsVerificationMessage,omitempty"`
	// The software token MFA configuration.
	SoftwareTokenMFAConfiguration *SoftwareTokenMFAConfigType `json:"softwareTokenMFAConfiguration,omitempty"`
	// The settings for updates to user attributes. These settings include the property
	// AttributesRequireVerificationBeforeUpdate, a user-pool setting that tells
	// Amazon Cognito how to handle changes to the value of your users' email address
	// and phone number attributes. For more information, see Verifying updates
	// to email addresses and phone numbers (https://docs.aws.amazon.com/cognito/latest/developerguide/user-pool-settings-email-phone-verification.html#user-pool-settings-verifications-verify-attribute-updates).
	UserAttributeUpdateSettings *UserAttributeUpdateSettingsType `json:"userAttributeUpdateSettings,omitempty"`
	// User pool add-ons. Contains settings for activation of advanced security
	// features. To log user security information but take no action, set to AUDIT.
	// To configure automatic security responses to risky traffic to your user pool,
	// set to ENFORCED.
	//
	// For more information, see Adding advanced security to a user pool (https://docs.aws.amazon.com/cognito/latest/developerguide/cognito-user-pool-settings-advanced-security.html).
	UserPoolAddOns *UserPoolAddOnsType `json:"userPoolAddOns,omitempty"`
	// The tag keys and values to assign to the user pool. A tag is a label that
	// you can use to categorize and manage user pools in different ways, such as
	// by purpose, owner, environment, or other criteria.
	UserPoolTags map[string]*string `json:"userPoolTags,omitempty"`
	// Specifies whether a user can use an email address or phone number as a username
	// when they sign up.
	UsernameAttributes []*string `json:"usernameAttributes,omitempty"`
	// Case sensitivity on the username input for the selected sign-in option. When
	// case sensitivity is set to False (case insensitive), users can sign in with
	// any combination of capital and lowercase letters. For example, username,
	// USERNAME, or UserName, or for email, email@example.com or EMaiL@eXamplE.Com.
	// For most use cases, set case sensitivity to False (case insensitive) as a
	// best practice. When usernames and email addresses are case insensitive, Amazon
	// Cognito treats any variation in case as the same user, and prevents a case
	// variation from being assigned to the same attribute for a different user.
	//
	// This configuration is immutable after you set it. For more information, see
	// UsernameConfigurationType (https://docs.aws.amazon.com/cognito-user-identity-pools/latest/APIReference/API_UsernameConfigurationType.html).
	UsernameConfiguration *UsernameConfigurationType `json:"usernameConfiguration,omitempty"`
	// The template for the verification message that the user sees when the app
	// requests permission to access the user's information.
	VerificationMessageTemplate *VerificationMessageTemplateType `json:"verificationMessageTemplate,omitempty"`
	CustomUserPoolParameters    `json:",inline"`
}

// UserPoolSpec defines the desired state of UserPool
type UserPoolSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       UserPoolParameters `json:"forProvider"`
}

// UserPoolObservation defines the observed state of UserPool
type UserPoolObservation struct {
	// The Amazon Resource Name (ARN) for the user pool.
	ARN *string `json:"arn,omitempty"`
	// The date and time, in ISO 8601 (https://www.iso.org/iso-8601-date-and-time-format.html)
	// format, when the item was created.
	CreationDate *metav1.Time `json:"creationDate,omitempty"`
	// A custom domain name that you provide to Amazon Cognito. This parameter applies
	// only if you use a custom domain to host the sign-up and sign-in pages for
	// your application. An example of a custom domain name might be auth.example.com.
	//
	// For more information about adding a custom domain to your user pool, see
	// Using Your Own Domain for the Hosted UI (https://docs.aws.amazon.com/cognito/latest/developerguide/cognito-user-pools-add-custom-domain.html).
	CustomDomain *string `json:"customDomain,omitempty"`
	// The domain prefix, if the user pool has a domain associated with it.
	Domain *string `json:"domain,omitempty"`
	// Deprecated. Review error codes from API requests with EventSource:cognito-idp.amazonaws.com
	// in CloudTrail for information about problems with user pool email configuration.
	EmailConfigurationFailure *string `json:"emailConfigurationFailure,omitempty"`
	// A number estimating the size of the user pool.
	EstimatedNumberOfUsers *int64 `json:"estimatedNumberOfUsers,omitempty"`
	// The ID of the user pool.
	ID *string `json:"id,omitempty"`
	// The date and time, in ISO 8601 (https://www.iso.org/iso-8601-date-and-time-format.html)
	// format, when the item was modified.
	LastModifiedDate *metav1.Time `json:"lastModifiedDate,omitempty"`
	// The name of the user pool.
	Name *string `json:"name,omitempty"`
	// A list of the user attributes and their properties in your user pool. The
	// attribute schema contains standard attributes, custom attributes with a custom:
	// prefix, and developer attributes with a dev: prefix. For more information,
	// see User pool attributes (https://docs.aws.amazon.com/cognito/latest/developerguide/user-pool-settings-attributes.html).
	//
	// Developer-only attributes are a legacy feature of user pools, are read-only
	// to all app clients. You can create and update developer-only attributes only
	// with IAM-authenticated API operations. Use app client read/write permissions
	// instead.
	SchemaAttributes []*SchemaAttributeType `json:"schemaAttributes,omitempty"`
	// The reason why the SMS configuration can't send the messages to your users.
	//
	// This message might include comma-separated values to describe why your SMS
	// configuration can't send messages to user pool end users.
	//
	// InvalidSmsRoleAccessPolicyException
	//
	// The Identity and Access Management role that Amazon Cognito uses to send
	// SMS messages isn't properly configured. For more information, see SmsConfigurationType
	// (https://docs.aws.amazon.com/cognito-user-identity-pools/latest/APIReference/API_SmsConfigurationType.html).
	//
	// SNSSandbox
	//
	// The Amazon Web Services account is in the SNS SMS Sandbox and messages will
	// only reach verified end users. This parameter won’t get populated with
	// SNSSandbox if the user creating the user pool doesn’t have SNS permissions.
	// To learn how to move your Amazon Web Services account out of the sandbox,
	// see Moving out of the SMS sandbox (https://docs.aws.amazon.com/sns/latest/dg/sns-sms-sandbox-moving-to-production.html).
	SmsConfigurationFailure *string `json:"smsConfigurationFailure,omitempty"`
	// The status of a user pool.
	Status *string `json:"status,omitempty"`

	CustomUserPoolObservation `json:",inline"`
}

// UserPoolStatus defines the observed state of UserPool.
type UserPoolStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          UserPoolObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// UserPool is the Schema for the UserPools API
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:storageversion
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,aws}
type UserPool struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              UserPoolSpec   `json:"spec"`
	Status            UserPoolStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// UserPoolList contains a list of UserPools
type UserPoolList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []UserPool `json:"items"`
}

// Repository type metadata.
var (
	UserPoolKind             = "UserPool"
	UserPoolGroupKind        = schema.GroupKind{Group: CRDGroup, Kind: UserPoolKind}.String()
	UserPoolKindAPIVersion   = UserPoolKind + "." + GroupVersion.String()
	UserPoolGroupVersionKind = GroupVersion.WithKind(UserPoolKind)
)

func init() {
	SchemeBuilder.Register(&UserPool{}, &UserPoolList{})
}
