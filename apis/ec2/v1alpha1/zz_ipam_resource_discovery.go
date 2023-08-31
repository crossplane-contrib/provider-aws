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

// IPAMResourceDiscoveryParameters defines the desired state of IPAMResourceDiscovery
type IPAMResourceDiscoveryParameters struct {
	// Region is which region the IPAMResourceDiscovery will be created.
	// +kubebuilder:validation:Required
	Region string `json:"region"`
	// A client token for the IPAM resource discovery.
	ClientToken *string `json:"clientToken,omitempty"`
	// A description for the IPAM resource discovery.
	Description *string `json:"description,omitempty"`
	// A check for whether you have the required permissions for the action without
	// actually making the request and provides an error response. If you have the
	// required permissions, the error response is DryRunOperation. Otherwise, it
	// is UnauthorizedOperation.
	DryRun *bool `json:"dryRun,omitempty"`
	// Operating Regions for the IPAM resource discovery. Operating Regions are
	// Amazon Web Services Regions where the IPAM is allowed to manage IP address
	// CIDRs. IPAM only discovers and monitors resources in the Amazon Web Services
	// Regions you select as operating Regions.
	OperatingRegions []*AddIPAMOperatingRegion `json:"operatingRegions,omitempty"`
	// Tag specifications for the IPAM resource discovery.
	TagSpecifications                     []*TagSpecification `json:"tagSpecifications,omitempty"`
	CustomIPAMResourceDiscoveryParameters `json:",inline"`
}

// IPAMResourceDiscoverySpec defines the desired state of IPAMResourceDiscovery
type IPAMResourceDiscoverySpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       IPAMResourceDiscoveryParameters `json:"forProvider"`
}

// IPAMResourceDiscoveryObservation defines the observed state of IPAMResourceDiscovery
type IPAMResourceDiscoveryObservation struct {
	// The resource discovery Amazon Resource Name (ARN).
	IPAMResourceDiscoveryARN *string `json:"ipamResourceDiscoveryARN,omitempty"`
	// The resource discovery ID.
	IPAMResourceDiscoveryID *string `json:"ipamResourceDiscoveryID,omitempty"`
	// The resource discovery Region.
	IPAMResourceDiscoveryRegion *string `json:"ipamResourceDiscoveryRegion,omitempty"`
	// Defines if the resource discovery is the default. The default resource discovery
	// is the resource discovery automatically created when you create an IPAM.
	IsDefault *bool `json:"isDefault,omitempty"`
	// The ID of the owner.
	OwnerID *string `json:"ownerID,omitempty"`
	// The lifecycle state of the resource discovery.
	//
	//    * create-in-progress - Resource discovery is being created.
	//
	//    * create-complete - Resource discovery creation is complete.
	//
	//    * create-failed - Resource discovery creation has failed.
	//
	//    * modify-in-progress - Resource discovery is being modified.
	//
	//    * modify-complete - Resource discovery modification is complete.
	//
	//    * modify-failed - Resource discovery modification has failed.
	//
	//    * delete-in-progress - Resource discovery is being deleted.
	//
	//    * delete-complete - Resource discovery deletion is complete.
	//
	//    * delete-failed - Resource discovery deletion has failed.
	//
	//    * isolate-in-progress - Amazon Web Services account that created the resource
	//    discovery has been removed and the resource discovery is being isolated.
	//
	//    * isolate-complete - Resource discovery isolation is complete.
	//
	//    * restore-in-progress - Amazon Web Services account that created the resource
	//    discovery and was isolated has been restored.
	State *string `json:"state,omitempty"`
	// A tag is a label that you assign to an Amazon Web Services resource. Each
	// tag consists of a key and an optional value. You can use tags to search and
	// filter your resources or track your Amazon Web Services costs.
	Tags []*Tag `json:"tags,omitempty"`
}

// IPAMResourceDiscoveryStatus defines the observed state of IPAMResourceDiscovery.
type IPAMResourceDiscoveryStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          IPAMResourceDiscoveryObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// IPAMResourceDiscovery is the Schema for the IPAMResourceDiscoveries API
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:storageversion
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,aws}
type IPAMResourceDiscovery struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              IPAMResourceDiscoverySpec   `json:"spec"`
	Status            IPAMResourceDiscoveryStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// IPAMResourceDiscoveryList contains a list of IPAMResourceDiscoveries
type IPAMResourceDiscoveryList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []IPAMResourceDiscovery `json:"items"`
}

// Repository type metadata.
var (
	IPAMResourceDiscoveryKind             = "IPAMResourceDiscovery"
	IPAMResourceDiscoveryGroupKind        = schema.GroupKind{Group: CRDGroup, Kind: IPAMResourceDiscoveryKind}.String()
	IPAMResourceDiscoveryKindAPIVersion   = IPAMResourceDiscoveryKind + "." + GroupVersion.String()
	IPAMResourceDiscoveryGroupVersionKind = GroupVersion.WithKind(IPAMResourceDiscoveryKind)
)

func init() {
	SchemeBuilder.Register(&IPAMResourceDiscovery{}, &IPAMResourceDiscoveryList{})
}
