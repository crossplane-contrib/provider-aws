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
// +kubebuilder:object:generate=true
// +groupName=network.aws.crossplane.io
// +versionName=v1alpha4

package v1alpha4

import (
	"reflect"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

// Package type metadata.
const (
	Group   = "ec2.aws.crossplane.io"
	Version = "v1alpha4"
)

var (
	// SchemeGroupVersion is group version used to register these objects
	SchemeGroupVersion = schema.GroupVersion{Group: Group, Version: Version}

	// SchemeBuilder is used to add go types to the GroupVersionKind scheme
	SchemeBuilder = &scheme.Builder{GroupVersion: SchemeGroupVersion}
)

// RouteTable type metadata.
var (
	RouteTableKind             = reflect.TypeOf(RouteTable{}).Name()
	RouteTableGroupKind        = schema.GroupKind{Group: Group, Kind: RouteTableKind}.String()
	RouteTableKindAPIVersion   = RouteTableKind + "." + SchemeGroupVersion.String()
	RouteTableGroupVersionKind = SchemeGroupVersion.WithKind(RouteTableKind)
)

// ElasticIP type metadata.
var (
	ElasticIPKind             = reflect.TypeOf(ElasticIP{}).Name()
	ElasticIPGroupKind        = schema.GroupKind{Group: Group, Kind: ElasticIPKind}.String()
	ElasticIPKindAPIVersion   = ElasticIPKind + "." + SchemeGroupVersion.String()
	ElasticIPGroupVersionKind = SchemeGroupVersion.WithKind(ElasticIPKind)
)

func init() {
	SchemeBuilder.Register(&RouteTable{}, &RouteTableList{})
	SchemeBuilder.Register(&ElasticIP{}, &ElasticIPList{})
}
