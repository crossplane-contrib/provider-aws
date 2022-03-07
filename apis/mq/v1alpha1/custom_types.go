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

import xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"

// CustomBrokerParameters contains the additional fields for CustomBrokerParameters
type CustomBrokerParameters struct {
	// +optional
	// +crossplane:generate:reference:type=github.com/crossplane/provider-aws/apis/ec2/v1beta1.Subnet
	// +crossplane:generate:reference:refFieldName=SubnetIDRefs
	// +crossplane:generate:reference:selectorFieldName=SubnetIDSelector
	SubnetIDs []*string `json:"subnetIDs,omitempty"`

	// SubnetIDRefs is a list of references to Subnets used to set
	// the SubnetIDs.
	// +optional
	SubnetIDRefs []xpv1.Reference `json:"subnetIDRefs,omitempty"`

	// SubnetIDsSelector selects references to Subnets used
	// to set the SubnetIDs.
	// +optional
	SubnetIDSelector *xpv1.Selector `json:"subnetIDSelector,omitempty"`

	// +crossplane:generate:reference:type=github.com/crossplane/provider-aws/apis/ec2/v1beta1.SecurityGroup
	// +crossplane:generate:reference:refFieldName=SecurityGroupIDRefs
	// +crossplane:generate:reference:selectorFieldName=SecurityGroupIDSelector
	SecurityGroups []*string `json:"securityGroups,omitempty"`

	// SecurityGroupIDRefs is a list of references to SecurityGroups used to set
	// the SecurityGroupsIDs.
	// +optional
	SecurityGroupIDRefs []xpv1.Reference `json:"securityGroupIdRefs,omitempty"`

	// SecurityGroupIDsSelector selects references to SecurityGroups used
	// to set the SecurityGroupsIDs.
	// +optional
	SecurityGroupIDSelector *xpv1.Selector `json:"securityGroupIdSelector,omitempty"`

	CustomUsers []*CustomUser `json:"users,omitempty"`
}

// CustomUser contains the fields for Users with PasswordSecretRef
type CustomUser struct {
	ConsoleAccess *bool `json:"consoleAccess,omitempty"`

	Groups []*string `json:"groups,omitempty"`

	PasswordSecretRef xpv1.SecretKeySelector `json:"passwordSecretRef,omitempty"`

	Username *string `json:"username,omitempty"`
}

// CustomUserParameters contains the additional fields for CustomUserParameters
type CustomUserParameters struct {
	// +optional
	// +crossplane:generate:reference:type=Broker
	BrokerID *string `json:"brokerID,omitempty"`

	// BrokerIDRef is a reference to a Broker used to set BrokerID.
	// +optional
	BrokerIDRef *xpv1.Reference `json:"brokerIDRef,omitempty"`

	// BrokerIDSelector selects a reference to a Broker used to set BrokerID.
	// +optional
	BrokerIDSelector *xpv1.Selector `json:"brokerIDSelector,omitempty"`

	PasswordSecretRef xpv1.SecretKeySelector `json:"passwordSecretRef,omitempty"`
}
