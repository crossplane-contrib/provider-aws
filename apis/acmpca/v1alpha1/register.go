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

// NOTE: Boilerplate only. Ignore this file.

// Package v1alpha1 contains API Schema definitions for the acmpca v1alpha1 API group
// +kubebuilder:object:generate=true
// +groupName=acmpca.aws.crossplane.io
// +versionName=v1alpha1
package v1alpha1

import (
	"reflect"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

// Package type metadata.
const (
	Group   = "acmpca.aws.crossplane.io"
	Version = "v1alpha1"
)

var (
	// SchemeGroupVersion is group version used to register these objects
	SchemeGroupVersion = schema.GroupVersion{Group: Group, Version: Version}

	// SchemeBuilder is used to add go types to the GroupVersionKind scheme
	SchemeBuilder = &scheme.Builder{GroupVersion: SchemeGroupVersion}
)

// CertificateAuthority type metadata.
var (
	CertificateAuthorityKind             = reflect.TypeOf(CertificateAuthority{}).Name()
	CertificateAuthorityGroupKind        = schema.GroupKind{Group: Group, Kind: CertificateAuthorityKind}.String()
	CertificateAuthorityKindAPIVersion   = CertificateAuthorityKind + "." + SchemeGroupVersion.String()
	CertificateAuthorityGroupVersionKind = SchemeGroupVersion.WithKind(CertificateAuthorityKind)
)

// CertificateAuthorityPermission type metadata.
var (
	CertificateAuthorityPermissionKind             = reflect.TypeOf(CertificateAuthorityPermission{}).Name()
	CertificateAuthorityPermissionGroupKind        = schema.GroupKind{Group: Group, Kind: CertificateAuthorityPermissionKind}.String()
	CertificateAuthorityPermissionKindAPIVersion   = CertificateAuthorityPermissionKind + "." + SchemeGroupVersion.String()
	CertificateAuthorityPermissionGroupVersionKind = SchemeGroupVersion.WithKind(CertificateAuthorityPermissionKind)
)

func init() {
	SchemeBuilder.Register(&CertificateAuthority{}, &CertificateAuthorityList{})
	SchemeBuilder.Register(&CertificateAuthorityPermission{}, &CertificateAuthorityPermissionList{})
}
