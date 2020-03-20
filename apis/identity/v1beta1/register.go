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
// +kubebuilder:object:generate=true
// +groupName=identity.aws.crossplane.io
// +versionName=v1beta1

package v1beta1

import (
	"reflect"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

// Package type metadata.
const (
	Group   = "identity.aws.crossplane.io"
	Version = "v1beta1"
)

var (
	// SchemeGroupVersion is group version used to register these objects
	SchemeGroupVersion = schema.GroupVersion{Group: Group, Version: Version}

	// SchemeBuilder is used to add go types to the GroupVersionKind scheme
	SchemeBuilder = &scheme.Builder{GroupVersion: SchemeGroupVersion}
)

// IAMRole type metadata.
var (
	IAMRoleKind             = reflect.TypeOf(IAMRole{}).Name()
	IAMRoleGroupKind        = schema.GroupKind{Group: Group, Kind: IAMRoleKind}.String()
	IAMRoleKindAPIVersion   = IAMRoleKind + "." + SchemeGroupVersion.String()
	IAMRoleGroupVersionKind = SchemeGroupVersion.WithKind(IAMRoleKind)
)

// IAMRolePolicyAttachment type metadata.
var (
	IAMRolePolicyAttachmentKind             = reflect.TypeOf(IAMRolePolicyAttachment{}).Name()
	IAMRolePolicyAttachmentGroupKind        = schema.GroupKind{Group: Group, Kind: IAMRolePolicyAttachmentKind}.String()
	IAMRolePolicyAttachmentKindAPIVersion   = IAMRolePolicyAttachmentKind + "." + SchemeGroupVersion.String()
	IAMRolePolicyAttachmentGroupVersionKind = SchemeGroupVersion.WithKind(IAMRolePolicyAttachmentKind)
)

func init() {
	SchemeBuilder.Register(&IAMRole{}, &IAMRoleList{})
	SchemeBuilder.Register(&IAMRolePolicyAttachment{}, &IAMRolePolicyAttachmentList{})
}
