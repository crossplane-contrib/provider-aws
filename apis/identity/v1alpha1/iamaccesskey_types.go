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

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
)

// IAMAccessKeyParameters define the desired state of an AWS IAM Access Key.
type IAMAccessKeyParameters struct {
	// IAMUsername contains the name of the IAMUser.
	// +optional
	// +immutable
	IAMUsername string `json:"userName,omitempty"`

	// IAMUsernameRef references to an IAMUser to retrieve its userName
	// +optional
	IAMUsernameRef *xpv1.Reference `json:"userNameRef,omitempty"`

	// IAMUsernameSelector selects a reference to an IAMUser to retrieve its userName
	// +optional
	IAMUsernameSelector *xpv1.Selector `json:"userNameSelector,omitempty"`

	// The current status of this IAMAccessKey on the AWS
	// Must be either Active or Inactive.
	// +kubebuilder:validation:Enum=Active;Inactive
	Status string `json:"accessKeyStatus,omitempty"`
}

// An IAMAccessKeySpec defines the desired state of an IAM Access Key.
type IAMAccessKeySpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       IAMAccessKeyParameters `json:"forProvider"`
}

// IAMAccessKeyStatus represents the observed state of an IAM Access Key.
type IAMAccessKeyStatus struct {
	xpv1.ResourceStatus `json:",inline"`
}

// +kubebuilder:object:root=true

// An IAMAccessKey is a managed resource that represents an the Access Key for an AWS IAM User.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:printcolumn:name="STATUS",type="string",JSONPath=".spec.forProvider.accessKeyStatus"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,aws}
type IAMAccessKey struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   IAMAccessKeySpec   `json:"spec"`
	Status IAMAccessKeyStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// IAMAccessKeyList contains a list of IAM Access Keys
type IAMAccessKeyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []IAMAccessKey `json:"items"`
}
