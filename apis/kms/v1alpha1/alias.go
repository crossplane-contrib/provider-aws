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

// NOTE(muvaf): This code ported from ACK-generated code. See details here:
// https://github.com/crossplane/provider-aws/pull/950#issue-1055573793

package v1alpha1

import (
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// AliasParameters defines the desired state of Alias
type AliasParameters struct {
	// Region is which region the Alias will be created.
	// +kubebuilder:validation:Required
	Region string `json:"region"`

	// Associates the alias with the specified customer managed CMK (https://docs.aws.amazon.com/kms/latest/developerguide/concepts.html#customer-cmk).
	// The CMK must be in the same AWS Region.
	//
	// A valid CMK ID is required. If you supply a null or empty string value, this
	// operation returns an error.
	//
	// For help finding the key ID and ARN, see Finding the Key ID and ARN (https://docs.aws.amazon.com/kms/latest/developerguide/viewing-keys.html#find-cmk-id-arn)
	// in the AWS Key Management Service Developer Guide.
	//
	// Specify the key ID or the Amazon Resource Name (ARN) of the CMK.
	//
	// For example:
	//
	//    * Key ID: 1234abcd-12ab-34cd-56ef-1234567890ab
	//
	//    * Key ARN: arn:aws:kms:us-east-2:111122223333:key/1234abcd-12ab-34cd-56ef-1234567890ab
	//
	// To get the key ID and key ARN for a CMK, use ListKeys or DescribeKey.
	// +crossplane:generate:reference:type=Key
	TargetKeyID *string `json:"targetKeyId,omitempty"`

	// TargetKeyIDRef is a reference to a KMS Key used to set TargetKeyID.
	// +optional
	TargetKeyIDRef *xpv1.Reference `json:"targetKeyIdRef,omitempty"`

	// TargetKeyIDSelector selects a reference to a KMS Key used to set TargetKeyID.
	// +optional
	TargetKeyIDSelector *xpv1.Selector `json:"targetKeyIdSelector,omitempty"`
}

// AliasSpec defines the desired state of Alias
type AliasSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       AliasParameters `json:"forProvider"`
}

// AliasObservation defines the observed state of Alias
type AliasObservation struct {
}

// AliasStatus defines the observed state of Alias.
type AliasStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          AliasObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// Alias is the Schema for the Aliases API
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,aws}
type Alias struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              AliasSpec   `json:"spec"`
	Status            AliasStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// AliasList contains a list of Aliases
type AliasList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Alias `json:"items"`
}

// Repository type metadata.
var (
	AliasKind             = "Alias"
	AliasGroupKind        = schema.GroupKind{Group: CRDGroup, Kind: AliasKind}.String()
	AliasKindAPIVersion   = AliasKind + "." + GroupVersion.String()
	AliasGroupVersionKind = GroupVersion.WithKind(AliasKind)
)

func init() {
	SchemeBuilder.Register(&Alias{}, &AliasList{})
}
