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

package manualv1alpha1

import (
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// VPCEndpointServiceAllowedPrincipalParameters defines the desired state of VPCEndpointServiceAllowedPrincipal
type VPCEndpointServiceAllowedPrincipalParameters struct {
	// Region is which region the VPCEndpointServiceAllowedPrincipal will be created.
	// +kubebuilder:validation:Required
	Region string `json:"region"`

	// The ID of the VPC endpoint service.
	// +kubebuilder:validation:Required
	VPCEndpointServiceID string `json:"vpcEndpointServiceId"`

	// The Amazon Resource Name (ARN) of the principal.
	// Permissions are granted to the principal in this field.
	// To grant permissions to all principals, specify an asterisk (*).
	// Principal ARNs with path components aren't supported.
	// +kubebuilder:validation:Required
	PrincipalARN string `json:"principalArn"`
}

// VPCEndpointServiceAllowedPrincipalSpec defines the desired state of VPCEndpointServiceAllowedPrincipal
type VPCEndpointServiceAllowedPrincipalSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       VPCEndpointServiceAllowedPrincipalParameters `json:"forProvider"`
}

// VPCEndpointServiceAllowedPrincipalObservation defines the observed state of VPCEndpointServiceAllowedPrincipal
type VPCEndpointServiceAllowedPrincipalObservation struct {
	// The Amazon Resource Name (ARN) of the principal.
	Principal *string `json:"principal,omitempty"`

	// The type of principal.
	PrincipalType *string `json:"principalType,omitempty"`

	// The ID of the service permission.
	ServicePermissionID *string `json:"servicePermissionId,omitempty"`

	// The ID of the VPC endpoint service.
	ServiceID *string `json:"serviceId,omitempty"`
}

// VPCEndpointServiceAllowedPrincipalStatus defines the observed state of VPCEndpointServiceAllowedPrincipal.
type VPCEndpointServiceAllowedPrincipalStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          VPCEndpointServiceAllowedPrincipalObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// VPCEndpointServiceAllowedPrincipal is the Schema for the VPCEndpointServiceAllowedPrincipals API
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,aws}
type VPCEndpointServiceAllowedPrincipal struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              VPCEndpointServiceAllowedPrincipalSpec   `json:"spec"`
	Status            VPCEndpointServiceAllowedPrincipalStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// VPCEndpointServiceAllowedPrincipalList contains a list of VPCEndpointServiceAllowedPrincipals
type VPCEndpointServiceAllowedPrincipalList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VPCEndpointServiceAllowedPrincipal `json:"items"`
}
