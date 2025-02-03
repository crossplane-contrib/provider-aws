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

package v1beta1

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

	// The Amazon Resource Name (ARN) of the function's execution role.
	// One of role, roleRef or roleSelector is required.
	Role *string `json:"role,omitempty"`

	// RoleRef is a reference to an iam role
	// +optional
	RoleRef *xpv1.Reference `json:"roleRef,omitempty"`

	// RoleSelector selects references to iam role arn used
	// to set the lambda Role.
	// +optional
	RoleSelector *xpv1.Selector `json:"roleSelector,omitempty"`

	// For network connectivity to AWS resources in a VPC, specify a list of security
	// groups and subnets in the VPC. When you connect a function to a VPC, it can
	// only access resources and the internet through that VPC. For more information,
	// see VPC Settings (https://docs.aws.amazon.com/lambda/latest/dg/configuration-vpc.html).
	CustomFunctionVPCConfigParameters *CustomFunctionVPCConfigParameters `json:"vpcConfig,omitempty"`

	// The code for the function.
	// +kubebuilder:validation:Required
	CustomFunctionCodeParameters CustomFunctionCodeParameters `json:"code"`
}

// CustomFunctionObservation includes the custom status fields of Function.
type CustomFunctionObservation struct{}

// CustomFunctionCodeParameters includes custom fields for FunctionCode struct.
type CustomFunctionCodeParameters struct {
	ImageURI *string `json:"imageURI,omitempty"`

	S3Key *string `json:"s3Key,omitempty"`

	S3ObjectVersion *string `json:"s3ObjectVersion,omitempty"`

	S3Bucket *string `json:"s3Bucket,omitempty"`

	// S3BucketRef is a reference to an S3 Bucket.
	// +optional
	S3BucketRef *xpv1.Reference `json:"s3BucketRef,omitempty"`

	// S3BucketSelector selects references to an S3 Bucket.
	// +optional
	S3BucketSelector *xpv1.Selector `json:"s3BucketSelector,omitempty"`
}

// CustomFunctionVPCConfigParameters includes custom fields for FunctionVPCConfigParameters.
type CustomFunctionVPCConfigParameters struct {
	SecurityGroupIDs []*string `json:"securityGroupIDs,omitempty"`

	// SecurityGroupIDRefs is a list of references to SecurityGroups used to set
	// the SecurityGroupIDs.
	// +optional
	SecurityGroupIDRefs []xpv1.Reference `json:"securityGroupIDRefs,omitempty"`

	// SecurityGroupIDsSelector selects references to SecurityGroupID used
	// to set the SecurityGroupIDs.
	// +optional
	SecurityGroupIDSelector *xpv1.Selector `json:"securityGroupIDSelector,omitempty"`

	SubnetIDs []*string `json:"subnetIDs,omitempty"`

	// SubnetIDRefs is a list of references to Subnets used to set
	// the SubnetIDs.
	// +optional
	SubnetIDRefs []xpv1.Reference `json:"subnetIDRefs,omitempty"`

	// SubnetIDsSelector selects references to Subnets used
	// to set the SubnetIDs.
	// +optional
	SubnetIDSelector *xpv1.Selector `json:"subnetIDSelector,omitempty"`
}
