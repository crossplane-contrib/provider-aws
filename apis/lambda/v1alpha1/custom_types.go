/*
Copyright 2020 The Crossplane Authors.

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

// CustomCodeSigningConfigParameters includes custom fields for CodeSigningConfigParameters.
type CustomCodeSigningConfigParameters struct{}

// CustomFunctionParameters includes custom fields for FunctionParameters.
type CustomFunctionParameters struct {

	// KMSKeyARNRef is a reference to a kms key used to set
	// the KMSKeyARN.
	// +optional
	KMSKeyARNRef *xpv1.Reference `json:"kmsKeyARNRef,omitempty"`

	// KMSKeyARNSelector selects references to kms key arn used
	// to set the KMSKeyARN.
	// +optional
	KMSKeyARNSelector *xpv1.Selector `json:"kmsKeyARNSelector,omitempty"`

	// RoleRef is a reference to an iam role
	// +optional
	RoleRef *xpv1.Reference `json:"roleRef,omitempty"`

	// RoleSelector selects references to iam role arn used
	// to set the lambda Role.
	// +optional
	RoleSelector *xpv1.Selector `json:"roleSelector,omitempty"`

	CustomFunctionVPCConfigParameters CustomFunctionVPCConfigParameters `json:"vpcConfig,omitempty"`
}

// CustomFunctionVPCConfigParameters includes custom fields for FunctionVPCConfigParameters.
type CustomFunctionVPCConfigParameters struct {

	// SecurityGroupIDRefs is a list of references to SecurityGroups used to set
	// the SecurityGroupIDs.
	// +optional
	SecurityGroupIDRefs []xpv1.Reference `json:"securityGroupIDRefs,omitempty"`

	// SecurityGroupIDsSelector selects references to SecurityGroupID used
	// to set the SecurityGroupIDs.
	// +optional
	SecurityGroupIDSelector *xpv1.Selector `json:"securityGroupIDSelector,omitempty"`

	// SubnetIDRefs is a list of references to Subnets used to set
	// the SubnetIDs.
	// +optional
	SubnetIDRefs []xpv1.Reference `json:"subnetIDRefs,omitempty"`

	// SubnetIDsSelector selects references to Subnets used
	// to set the SubnetIDs.
	// +optional
	SubnetIDSelector *xpv1.Selector `json:"subnetIDSelector,omitempty"`
}
