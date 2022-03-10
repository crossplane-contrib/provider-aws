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

// AccessPointParameters defines the desired state of AccessPoint
type AccessPointParameters struct {
	// Region is which region the AccessPoint will be created.
	// +kubebuilder:validation:Required
	Region string `json:"region"`
	// A string of up to 64 ASCII characters that Amazon EFS uses to ensure idempotent
	// creation.
	// +kubebuilder:validation:Required
	ClientToken *string `json:"clientToken"`
	// The operating system user and group applied to all file system requests made
	// using the access point.
	PosixUser *PosixUser `json:"posixUser,omitempty"`
	// Specifies the directory on the Amazon EFS file system that the access point
	// exposes as the root directory of your file system to NFS clients using the
	// access point. The clients using the access point can only access the root
	// directory and below. If the RootDirectory > Path specified does not exist,
	// EFS creates it and applies the CreationInfo settings when a client connects
	// to an access point. When specifying a RootDirectory, you need to provide
	// the Path, and the CreationInfo.
	//
	// Amazon EFS creates a root directory only if you have provided the CreationInfo:
	// OwnUid, OwnGID, and permissions for the directory. If you do not provide
	// this information, Amazon EFS does not create the root directory. If the root
	// directory does not exist, attempts to mount using the access point will fail.
	RootDirectory *RootDirectory `json:"rootDirectory,omitempty"`
	// Creates tags associated with the access point. Each tag is a key-value pair,
	// each key must be unique. For more information, see Tagging Amazon Web Services
	// resources (https://docs.aws.amazon.com/general/latest/gr/aws_tagging.html)
	// in the Amazon Web Services General Reference Guide.
	Tags                        []*Tag `json:"tags,omitempty"`
	CustomAccessPointParameters `json:",inline"`
}

// AccessPointSpec defines the desired state of AccessPoint
type AccessPointSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       AccessPointParameters `json:"forProvider"`
}

// AccessPointObservation defines the observed state of AccessPoint
type AccessPointObservation struct {
	// The unique Amazon Resource Name (ARN) associated with the access point.
	AccessPointARN *string `json:"accessPointARN,omitempty"`
	// The ID of the access point, assigned by Amazon EFS.
	AccessPointID *string `json:"accessPointID,omitempty"`
	// The ID of the EFS file system that the access point applies to.
	FileSystemID *string `json:"fileSystemID,omitempty"`
	// Identifies the lifecycle phase of the access point.
	LifeCycleState *string `json:"lifeCycleState,omitempty"`
	// The name of the access point. This is the value of the Name tag.
	Name *string `json:"name,omitempty"`
	// Identified the Amazon Web Services account that owns the access point resource.
	OwnerID *string `json:"ownerID,omitempty"`
}

// AccessPointStatus defines the observed state of AccessPoint.
type AccessPointStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          AccessPointObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// AccessPoint is the Schema for the AccessPoints API
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:subresource:status
// +kubebuilder:storageversion
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,aws}
type AccessPoint struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              AccessPointSpec   `json:"spec"`
	Status            AccessPointStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// AccessPointList contains a list of AccessPoints
type AccessPointList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AccessPoint `json:"items"`
}

// Repository type metadata.
var (
	AccessPointKind             = "AccessPoint"
	AccessPointGroupKind        = schema.GroupKind{Group: CRDGroup, Kind: AccessPointKind}.String()
	AccessPointKindAPIVersion   = AccessPointKind + "." + GroupVersion.String()
	AccessPointGroupVersionKind = GroupVersion.WithKind(AccessPointKind)
)

func init() {
	SchemeBuilder.Register(&AccessPoint{}, &AccessPointList{})
}
