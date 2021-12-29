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

// LaunchTemplateVersionParameters defines the desired state of LaunchTemplateVersion
type LaunchTemplateVersionParameters struct {
	// Region is which region the LaunchTemplateVersion will be created.
	// +kubebuilder:validation:Required
	Region string `json:"region"`
	// Unique, case-sensitive identifier you provide to ensure the idempotency of
	// the request. For more information, see Ensuring Idempotency (https://docs.aws.amazon.com/AWSEC2/latest/APIReference/Run_Instance_Idempotency.html).
	//
	// Constraint: Maximum 128 ASCII characters.
	ClientToken *string `json:"clientToken,omitempty"`
	// Checks whether you have the required permissions for the action, without
	// actually making the request, and provides an error response. If you have
	// the required permissions, the error response is DryRunOperation. Otherwise,
	// it is UnauthorizedOperation.
	DryRun *bool `json:"dryRun,omitempty"`
	// The information for the launch template.
	// +kubebuilder:validation:Required
	LaunchTemplateData *RequestLaunchTemplateData `json:"launchTemplateData"`
	// The ID of the launch template. You must specify either the launch template
	// ID or launch template name in the request.
	LaunchTemplateID *string `json:"launchTemplateID,omitempty"`
	// The name of the launch template. You must specify either the launch template
	// ID or launch template name in the request.
	LaunchTemplateName *string `json:"launchTemplateName,omitempty"`
	// The version number of the launch template version on which to base the new
	// version. The new version inherits the same launch parameters as the source
	// version, except for parameters that you specify in LaunchTemplateData. Snapshots
	// applied to the block device mapping are ignored when creating a new version
	// unless they are explicitly included.
	SourceVersion *string `json:"sourceVersion,omitempty"`
	// A description for the version of the launch template.
	VersionDescription                    *string `json:"versionDescription,omitempty"`
	CustomLaunchTemplateVersionParameters `json:",inline"`
}

// LaunchTemplateVersionSpec defines the desired state of LaunchTemplateVersion
type LaunchTemplateVersionSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       LaunchTemplateVersionParameters `json:"forProvider"`
}

// LaunchTemplateVersionObservation defines the observed state of LaunchTemplateVersion
type LaunchTemplateVersionObservation struct {
	// Information about the launch template version.
	LaunchTemplateVersion *LaunchTemplateVersion_SDK `json:"launchTemplateVersion,omitempty"`
	// If the new version of the launch template contains parameters or parameter
	// combinations that are not valid, an error code and an error message are returned
	// for each issue that's found.
	Warning *ValidationWarning `json:"warning,omitempty"`
}

// LaunchTemplateVersionStatus defines the observed state of LaunchTemplateVersion.
type LaunchTemplateVersionStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          LaunchTemplateVersionObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// LaunchTemplateVersion is the Schema for the LaunchTemplateVersions API
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,aws}
type LaunchTemplateVersion struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              LaunchTemplateVersionSpec   `json:"spec"`
	Status            LaunchTemplateVersionStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// LaunchTemplateVersionList contains a list of LaunchTemplateVersions
type LaunchTemplateVersionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []LaunchTemplateVersion `json:"items"`
}

// Repository type metadata.
var (
	LaunchTemplateVersionKind             = "LaunchTemplateVersion"
	LaunchTemplateVersionGroupKind        = schema.GroupKind{Group: Group, Kind: LaunchTemplateVersionKind}.String()
	LaunchTemplateVersionKindAPIVersion   = LaunchTemplateVersionKind + "." + GroupVersion.String()
	LaunchTemplateVersionGroupVersionKind = GroupVersion.WithKind(LaunchTemplateVersionKind)
)

func init() {
	SchemeBuilder.Register(&LaunchTemplateVersion{}, &LaunchTemplateVersionList{})
}