/*
Copyright 2022 The Crossplane Authors.

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

// Package manualv1alpha1 is the v1alpha1 version of the lambda.aws.crossplane.io API.
// +kubebuilder:object:generate=true
// +groupName=lambda.aws.crossplane.io
// +versionName=v1alpha1
package manualv1alpha1

import (
	"reflect"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

// Package type metadata.
const (
	Group   = "lambda.aws.crossplane.io"
	Version = "v1alpha1"
)

var (
	// SchemeGroupVersion is group version used to register these objects
	SchemeGroupVersion = schema.GroupVersion{Group: Group, Version: Version}

	// SchemeBuilder is used to add go types to the GroupVersionKind scheme
	SchemeBuilder = &scheme.Builder{GroupVersion: SchemeGroupVersion}
)

// Permission type metadata.
var (
	PermissionKind             = reflect.TypeOf(Permission{}).Name()
	PermissionGroupKind        = schema.GroupKind{Group: Group, Kind: PermissionKind}.String()
	PermissionKindAPIVersion   = PermissionKind + "." + SchemeGroupVersion.String()
	PermissionGroupVersionKind = SchemeGroupVersion.WithKind(PermissionKind)
)

func init() {
	SchemeBuilder.Register(&Permission{}, &PermissionList{})
}
