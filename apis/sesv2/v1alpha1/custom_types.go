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

package v1alpha1

import (
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
)

// SES states.
const (
	// The Configuration sets Sending status to Enabled
	ConfigurationSetsSendingStatusEnabled = "Enabled"
	// The Configuration sets Sending status to Disabled
	ConfigurationSetsSendingStatusDisabled = "Disabled"
	// The EmailIdentity DKIMAttributes status is pending
	DkimAttributesStatusPending = "PENDING"
	// The EmailIdentity DKIMAttributes status is successful
	DkimAttributesStatusSuccess = "SUCCESS"
	// The EmailIdentity DKIMAttributes status is failed
	DkimAttributesStatusFailed = "FAILED"
	// The EmailIdentity DKIMAttributes is temporary failed
	DkimAttributesStatusTemporaryFailure = "TEMPORARY_FAILURE"
	// The EmailIdentity DKIMAttributes is not started
	DkimAttributesStatusNotStarted = "NOT_STARTED"
)

// CustomConfigurationSetEventDestinationParameters are parameters for
type CustomConfigurationSetEventDestinationParameters struct{}

// CustomConfigurationSetEventDestinationObservation includes the custom status fields of ConfigurationSetEventDestination.
type CustomConfigurationSetEventDestinationObservation struct{}

// CustomConfigurationSetParameters are parameters for
type CustomConfigurationSetParameters struct{}

// CustomConfigurationSetObservation includes the custom status fields of ConfigurationSet.
type CustomConfigurationSetObservation struct{}

// CustomEmailIdentityParameters are parameters for
type CustomEmailIdentityParameters struct {
	// An object that contains information about the Mail-From attributes for the email identity.
	// +optional
	MailFromAttributes *MailFromAttributes `json:"mailFromAttributes,omitempty"`

	// DomainSigningPrivateKeySecretRef references the secret that contains the private key of the DKIM Authentication Token
	// Constraints: Base64 encoded format
	DomainSigningPrivateKeySecretRef *xpv1.SecretKeySelector `json:"domainSigningPrivateKeySecretRef,omitempty"`

	// The configuration set to use by default when sending from this identity.
	// Note that any configuration set defined in the email sending request takes
	// precedence.
	// +immutable
	// +crossplane:generate:reference:type=ConfigurationSet
	ConfigurationSetName *string `json:"configurationSetName,omitempty"`

	// ConfigurationSetNameRef is a reference to an API used to set
	// the ConfigurationSetName.
	// +optional
	ConfigurationSetNameRef *xpv1.Reference `json:"configurationSetNameRef,omitempty"`

	// ConfigurationSetNameSelector selects references to API used
	// to set the ConfigurationSetName.
	// +optional
	ConfigurationSetNameSelector *xpv1.Selector `json:"configurationSetNameSelector,omitempty"`
}

// CustomEmailIdentityObservation includes the custom status fields of EmailIdentity.
type CustomEmailIdentityObservation struct{}

// CustomEmailTemplateParameters are parameters for
type CustomEmailTemplateParameters struct{}

// CustomEmailTemplateObservation includes the custom status fields of EmailTemplate.
type CustomEmailTemplateObservation struct{}
