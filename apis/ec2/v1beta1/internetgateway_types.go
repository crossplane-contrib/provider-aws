/*
Copyright 2019 The Crossplane Authors.

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

import (
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// AWS returns 'available` hence ec2.AttachmentStatusAttached doesn't work
// InternetGateway attachment states.
const (
	// The attachment is complete
	AttachmentStatusAvailable = "available"
	// The attachment is being created.
	AttachmentStatusAttaching = "creating"
)

// InternetGatewayParameters define the desired state of an AWS VPC Internet
// Gateway.
type InternetGatewayParameters struct {

	// TODO(muvaf): Region is a required field but in order to keep backward compatibility
	// with old Provider type and not bear the cost of bumping to v1beta2, we're
	// keeping it optional for now. Reconsider before v1beta2 or v1.

	// Region is the region you'd like your VPC to be created in.
	// +optional
	Region *string `json:"region,omitempty"`

	// VPCID is the ID of the VPC.
	// +optional
	// +crossplane:generate:reference:type=VPC
	VPCID *string `json:"vpcId,omitempty"`

	// VPCIDRef references a VPC to and retrieves its vpcId
	// +optional
	VPCIDRef *xpv1.Reference `json:"vpcIdRef,omitempty"`

	// VPCIDSelector selects a reference to a VPC to and retrieves its vpcId
	// +optional
	VPCIDSelector *xpv1.Selector `json:"vpcIdSelector,omitempty"`

	// Tags represents to current ec2 tags.
	// +optional
	Tags []Tag `json:"tags,omitempty"`
}

// An InternetGatewaySpec defines the desired state of an InternetGateway.
type InternetGatewaySpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       InternetGatewayParameters `json:"forProvider"`
}

// InternetGatewayAttachment describes the attachment of a VPC to an internet
// gateway or an egress-only internet gateway.
type InternetGatewayAttachment struct {
	// The current state of the attachment. For an internet gateway, the state
	// is available when attached to a VPC; otherwise, this value is not
	// returned.
	// +kubebuilder:validation:Enum=available;attaching;attached;detaching;detached
	AttachmentStatus string `json:"attachmentStatus"`

	// VPCID is the ID of the attached VPC.
	VPCID string `json:"vpcId"`
}

// InternetGatewayObservation keeps the state for the external resource
type InternetGatewayObservation struct {
	// Any VPCs attached to the internet gateway.
	Attachments []InternetGatewayAttachment `json:"attachments,omitempty"`

	// The ID of the internet gateway.
	InternetGatewayID string `json:"internetGatewayId"`

	// The ID of the AWS account that owns the internet gateway.
	OwnerID string `json:"ownerID"`
}

// An InternetGatewayStatus represents the observed state of an InternetGateway.
type InternetGatewayStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          InternetGatewayObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// An InternetGateway is a managed resource that represents an AWS VPC Internet
// Gateway.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="ID",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="VPC",type="string",JSONPath=".spec.forProvider.vpcId"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,aws}
// +kubebuilder:storageversion
type InternetGateway struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   InternetGatewaySpec   `json:"spec"`
	Status InternetGatewayStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// InternetGatewayList contains a list of InternetGateways
type InternetGatewayList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []InternetGateway `json:"items"`
}
