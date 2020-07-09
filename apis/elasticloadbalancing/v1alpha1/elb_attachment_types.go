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
	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ELBAttachmentParameters define the desired state of an AWS ELBAttachment.
type ELBAttachmentParameters struct {
	// Name of the Elastic Load Balancer to which the instances will attach.
	// +immutable
	// +optional
	ELBName string `json:"elbName,omitempty"`

	// ELBNameRef references an ELB to and retrieves its external-name.
	// +immutable
	// +optional
	ELBNameRef *runtimev1alpha1.Reference `json:"elbNameRef,omitempty"`

	// ELBNameSelector selects a reference to a ELB to and retrieves its external-name.
	// +immutable
	// +optional
	ELBNameSelector *runtimev1alpha1.Selector `json:"elbNameSelector,omitempty"`

	// List of identities of the instances to be attached.
	// +immutable
	InstanceID string `json:"instanceId"`
}

// An ELBAttachmentSpec defines the desired state of an ELBAttachment.
type ELBAttachmentSpec struct {
	runtimev1alpha1.ResourceSpec `json:",inline"`
	ForProvider                  ELBAttachmentParameters `json:"forProvider"`
}

// ELBAttachmentObservation keeps the state for the external resource
type ELBAttachmentObservation struct {
}

// An ELBAttachmentStatus represents the observed state of an ELBAttachmentAttachment.
type ELBAttachmentStatus struct {
	runtimev1alpha1.ResourceStatus `json:",inline"`
	AtProvider                     ELBAttachmentObservation `json:"atProvider"`
}

// +kubebuilder:object:root=true

// An ELBAttachment is a managed resource that represents attachment of an
// AWS Classic Load Balancer and an AWS EC2 instance.
// +kubebuilder:printcolumn:name="ELBNAME",type="string",JSONPath=".spec.forProvider.elbName"
// +kubebuilder:printcolumn:name="INSTANCEID",type="string",JSONPath=".spec.forProvider.instanceId"
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,aws}
type ELBAttachment struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ELBAttachmentSpec   `json:"spec"`
	Status ELBAttachmentStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ELBAttachmentList contains a list of ELBAttachmentAttachment.
type ELBAttachmentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ELBAttachment `json:"items"`
}
