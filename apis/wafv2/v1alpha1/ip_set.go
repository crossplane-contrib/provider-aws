/*
Copyright 2026 The Crossplane Authors.

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
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// IPSetParameters defines the desired state of IPSet
type IPSetParameters struct {
	// Region is which region the IPSet will be created.
	// +kubebuilder:validation:Required
	Region string `json:"region"`
	// Contains an array of strings that specifies zero or more IP addresses or
	// blocks of IP addresses that you want WAF to inspect for in incoming requests.
	// All addresses must be specified using Classless Inter-Domain Routing (CIDR)
	// notation. WAF supports all IPv4 and IPv6 CIDR ranges except for /0.
	//
	// Example IPv4 addresses: 1.2.3.4/32, 10.0.0.0/8
	// Example IPv6 addresses: 1111:0000:0000:0000:0000:0000:0000:0111/128, 2001:db8::/32
	// +kubebuilder:validation:Required
	Addresses []*string `json:"addresses"`
	// A description of the IP set that helps with identification.
	Description *string `json:"description,omitempty"`
	// The version of the IP addresses, either IPV4 or IPV6.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=IPV4;IPV6
	IPAddressVersion *string `json:"ipAddressVersion"`
	// The name of the IP set. You cannot change the name of an IPSet after you create it.
	// +kubebuilder:validation:Required
	Name *string `json:"name"`
	// Specifies whether this is for an Amazon CloudFront distribution or for a
	// regional application.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=CLOUDFRONT;REGIONAL
	Scope *string `json:"scope"`
	// An array of key:value pairs to associate with the resource.
	Tags []*Tag `json:"tags,omitempty"`
}

// IPSetObservation defines the observed state of IPSet
type IPSetObservation struct {
	// The Amazon Resource Name (ARN) of the IP set.
	ARN *string `json:"arn,omitempty"`
	// A unique identifier for the set.
	ID *string `json:"id,omitempty"`
	// A token used for optimistic locking. WAF returns a token to your get and
	// list requests, to mark the state of the entity at the time of the request.
	// To make changes to the entity associated with the token, you provide the
	// token to operations like update and delete.
	LockToken *string `json:"lockToken,omitempty"`
}

// IPSetSpec defines the desired state of IPSet
type IPSetSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       IPSetParameters `json:"forProvider"`
}

// IPSetStatus defines the observed state of IPSet.
type IPSetStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          IPSetObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// IPSet is the Schema for the IPSets API
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,aws}
type IPSet struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              IPSetSpec   `json:"spec"`
	Status            IPSetStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// IPSetList contains a list of IPSets
type IPSetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []IPSet `json:"items"`
}

// Repository type metadata.
var (
	IPSetKind             = "IPSet"
	IPSetGroupKind        = schema.GroupKind{Group: CRDGroup, Kind: IPSetKind}.String()
	IPSetKindAPIVersion   = IPSetKind + "." + GroupVersion.String()
	IPSetGroupVersionKind = GroupVersion.WithKind(IPSetKind)
)

func init() {
	SchemeBuilder.Register(&IPSet{}, &IPSetList{})
}
