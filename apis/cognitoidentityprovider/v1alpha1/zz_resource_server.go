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

// ResourceServerParameters defines the desired state of ResourceServer
type ResourceServerParameters struct {
	// Region is which region the ResourceServer will be created.
	// +kubebuilder:validation:Required
	Region string `json:"region"`
	// A unique resource server identifier for the resource server. This could be
	// an HTTPS endpoint where the resource server is located. For example, https://my-weather-api.example.com.
	// +kubebuilder:validation:Required
	Identifier *string `json:"identifier"`
	// A friendly name for the resource server.
	// +kubebuilder:validation:Required
	Name *string `json:"name"`
	// A list of scopes. Each scope is map, where the keys are name and description.
	Scopes []*ResourceServerScopeType `json:"scopes,omitempty"`
	// The user pool ID for the user pool.
	// +kubebuilder:validation:Required
	UserPoolID                     *string `json:"userPoolID"`
	CustomResourceServerParameters `json:",inline"`
}

// ResourceServerSpec defines the desired state of ResourceServer
type ResourceServerSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       ResourceServerParameters `json:"forProvider"`
}

// ResourceServerObservation defines the observed state of ResourceServer
type ResourceServerObservation struct {
}

// ResourceServerStatus defines the observed state of ResourceServer.
type ResourceServerStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          ResourceServerObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// ResourceServer is the Schema for the ResourceServers API
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:subresource:status
// +kubebuilder:storageversion
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,aws}
type ResourceServer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              ResourceServerSpec   `json:"spec"`
	Status            ResourceServerStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ResourceServerList contains a list of ResourceServers
type ResourceServerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ResourceServer `json:"items"`
}

// Repository type metadata.
var (
	ResourceServerKind             = "ResourceServer"
	ResourceServerGroupKind        = schema.GroupKind{Group: CRDGroup, Kind: ResourceServerKind}.String()
	ResourceServerKindAPIVersion   = ResourceServerKind + "." + GroupVersion.String()
	ResourceServerGroupVersionKind = GroupVersion.WithKind(ResourceServerKind)
)

func init() {
	SchemeBuilder.Register(&ResourceServer{}, &ResourceServerList{})
}
