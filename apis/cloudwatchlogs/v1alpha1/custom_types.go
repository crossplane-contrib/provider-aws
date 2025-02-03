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

// CustomLogGroupParameters contains the additional fields for LogGroup.
type CustomLogGroupParameters struct {
	// The number of days to retain the log events in the specified log group.
	// If you select 0, the events in the log group are always retained and never expire.
	// +kubebuilder:validation:Enum=0;1;3;5;7;14;30;60;90;120;150;180;365;400;545;731;1827;3653
	// +optional
	RetentionInDays *int64 `json:"retentionInDays,omitempty"`

	// The Amazon Resource Name (ARN) of the CMK to use when encrypting log data.
	// For more information, see Amazon Resource Names - AWS Key Management Service
	// (AWS KMS) (https://docs.aws.amazon.com/general/latest/gr/aws-arns-and-namespaces.html#arn-syntax-kms).
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-aws/apis/kms/v1alpha1.Key
	// +crossplane:generate:reference:refFieldName=KMSKeyIDRef
	// +crossplane:generate:reference:selectorFieldName=KMSKeyIDSelector
	KMSKeyID *string `json:"kmsKeyID,omitempty"`

	// KMSKeyIDRef is a reference to a KMS Key used to set KMSKeyID.
	// +optional
	KMSKeyIDRef *xpv1.Reference `json:"kmsKeyIDRef,omitempty"`

	// KMSKeyIDSelector selects a reference to a KMS Key used to set KMSKeyID.
	// +optional
	KMSKeyIDSelector *xpv1.Selector `json:"kmsKeyIDSelector,omitempty"`
}

// CustomLogGroupObservation contains the additional status fields for LogGroup.
type CustomLogGroupObservation struct{}

// CustomResourcePolicyParameters includes the custom fields of ResourcePolicy.
type CustomResourcePolicyParameters struct{}

// CustomResourcePolicyObservation contains the additional status fields for ResourcePolicy.
type CustomResourcePolicyObservation struct{}
