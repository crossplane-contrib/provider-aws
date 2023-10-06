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
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:root=true

// HostedZone is a managed resource that represents an AWS Route53 Hosted HostedZone.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="ID",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="RRs",type="integer",JSONPath=".status.atProvider.hostedZone.resourceRecordSetCount"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,aws}
type HostedZone struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   HostedZoneSpec   `json:"spec"`
	Status HostedZoneStatus `json:"status,omitempty"`
}

// HostedZoneSpec defines the desired state of an AWS Route53 Hosted HostedZone.
type HostedZoneSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       HostedZoneParameters `json:"forProvider"`
}

// HostedZoneStatus represents the observed state of a HostedZone.
type HostedZoneStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          HostedZoneObservation `json:"atProvider,omitempty"`
}

// HostedZoneParameters define the desired state of an AWS Route53 Hosted HostedZone.
type HostedZoneParameters struct {
	// The name of the domain. Specify a fully qualified domain name, for example,
	// www.example.com. The trailing dot is optional; Amazon Route 53 assumes that
	// the domain name is fully qualified. This means that Route 53 treats www.example.com
	// (without a trailing dot) and www.example.com. (with a trailing dot) as identical.
	//
	// If you're creating a public hosted zone, this is the name you have registered
	// with your DNS registrar. If your domain name is registered with a registrar
	// other than Route 53, change the name servers for your domain to the set of
	// NameServers that CreateHostedHostedZone returns in DelegationSet.
	// +immutable
	Name string `json:"name"`

	// Config includes the Comment and PrivateZone elements. If you
	// omitted the Config and Comment elements from the request, the Config
	// and Comment elements don't appear in the response.
	// +optional
	Config *Config `json:"config,omitempty"`

	// DelegationSetId let you associate a reusable delegation set with this hosted zone.
	// It has to be the ID that Amazon Route 53 assigned to the reusable delegation set when
	// you created it. For more information about reusable delegation sets, see
	// CreateReusableDelegationSet (https://docs.aws.amazon.com/Route53/latest/APIReference/API_CreateReusableDelegationSet.html).
	// +optional
	DelegationSetID *string `json:"delegationSetId,omitempty"`

	// (Private hosted zones only) A complex type that contains information about
	// the Amazon VPC that you're associating with this hosted zone.
	//
	// You can specify only one Amazon VPC when you create a private hosted zone.
	// To associate additional Amazon VPCs with the hosted zone, use AssociateVPCWithHostedZone
	// (https://docs.aws.amazon.com/Route53/latest/APIReference/API_AssociateVPCWithHostedZone.html)
	// after you create a hosted zone.
	// +immutable
	// +optional
	VPC *VPC `json:"vpc,omitempty"`

	// Tags for this hosted zone.
	// +optional
	Tags map[string]string `json:"tags,omitempty"`
}

// Config represents the configuration of a Hosted Zone.
type Config struct {
	// Comment that you want to include about the hosted zone.
	Comment *string `json:"comment,omitempty"`

	// PrivateZone indicates whether this is a private hosted zone.
	// +immutable
	PrivateZone *bool `json:"privateZone,omitempty"`
}

// VPC is used to refer to specific VPC.
type VPC struct {
	// (Private hosted zones only) The ID of an Amazon VPC.
	// +immutable
	// +optional
	VPCID *string `json:"vpcId,omitempty"`

	// (Private hosted zones only) The region that an Amazon VPC was created in.
	// +immutable
	// +optional
	VPCRegion *string `json:"vpcRegion,omitempty"`

	// (Private hosted Hostedzones only) VPCIDRef references a VPC to retrieves its VPC Id.
	// +immutable
	// +optional
	VPCIDRef *xpv1.Reference `json:"vpcIdRef,omitempty"`

	// VPCIDSelector selects a reference to a VPC.
	// +optional
	VPCIDSelector *xpv1.Selector `json:"vpcIdSelector,omitempty"`
}

// HostedZoneObservation keeps the state for the external resource.
type HostedZoneObservation struct {
	// DelegationSet describes the name servers for this hosted zone.
	DelegationSet DelegationSet `json:"delegationSet,omitempty"`

	// HostedZone contains general information about the hosted zone.
	HostedZone HostedZoneResponse `json:"hostedZone,omitempty"`

	// A complex type that contains information about the VPCs that are associated
	// with the specified hosted zone.
	VPCs []VPCObservation `json:"vpcs,omitempty"`
}

// HostedZoneResponse stores the Hosted Zone received in the response output
type HostedZoneResponse struct {
	// CallerReference is an unique string that identifies the request and that
	// allows failed HostedZone create requests to be retried without the risk of
	// executing the operation twice.
	CallerReference string `json:"callerReference,omitempty"`

	// ID that Amazon Route 53 assigned to the hosted zone when you created
	// it.
	ID string `json:"id,omitempty"`

	// LinkedService is the service that created the hosted zone.
	LinkedService LinkedService `json:"linkedService,omitempty"`

	// The number of resource record sets in the hosted zone.
	ResourceRecordSetCount int64 `json:"resourceRecordSetCount,omitempty"`
}

// DelegationSet describes the name servers for this hosted Hostedzone.
type DelegationSet struct {
	// The value that you specified for CallerReference when you created the reusable
	// delegation set.
	CallerReference string `json:"callerReference,omitempty"`

	// The ID that Amazon Route 53 assigns to a reusable delegation set.
	ID string `json:"id,omitempty"`

	// NameServers contains a list of the authoritative name servers for a hosted Hostedzone.
	NameServers []string `json:"nameServers,omitempty"`
}

// VPCObservation is used to represent the VPC object in the HostedZone response
// object.
type VPCObservation struct {

	// VPCID is the ID of the VPC.
	VPCID string `json:"vpcId,omitempty"`

	// VPCRegion is the region where the VPC resides.
	VPCRegion string `json:"vpcRegion,omitempty"`
}

// LinkedService is the service that created the hosted zone.
type LinkedService struct {
	// Description provided by the other service.
	Description string `json:"description,omitempty"`

	// ServicePrincipal is the service that created the resource.
	ServicePrincipal string `json:"servicePrincipal,omitempty"`
}

// +kubebuilder:object:root=true

// HostedZoneList contains a list of HostedZone.
type HostedZoneList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []HostedZone `json:"items"`
}
