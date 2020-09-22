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
	"reflect"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

// Package type metadata.
const (
	Group   = "database.aws.crossplane.io"
	Version = "v1beta1"
)

var (
	// SchemeGroupVersion is group version used to register these objects
	SchemeGroupVersion = schema.GroupVersion{Group: Group, Version: Version}

	// SchemeBuilder is used to add go types to the GroupVersionKind scheme
	SchemeBuilder = &scheme.Builder{GroupVersion: SchemeGroupVersion}
)

// RDSInstance type metadata.
var (
	RDSInstanceKind             = reflect.TypeOf(RDSInstance{}).Name()
	RDSInstanceGroupKind        = schema.GroupKind{Group: Group, Kind: RDSInstanceKind}.String()
	RDSInstanceKindAPIVersion   = RDSInstanceKind + "." + SchemeGroupVersion.String()
	RDSInstanceGroupVersionKind = SchemeGroupVersion.WithKind(RDSInstanceKind)
)

// DBSubnetGroup type metadata.
var (
	DBSubnetGroupKind             = reflect.TypeOf(DBSubnetGroup{}).Name()
	DBSubnetGroupGroupKind        = schema.GroupKind{Group: Group, Kind: DBSubnetGroupKind}.String()
	DBSubnetGroupKindAPIVersion   = DBSubnetGroupKind + "." + SchemeGroupVersion.String()
	DBSubnetGroupGroupVersionKind = SchemeGroupVersion.WithKind(DBSubnetGroupKind)
)

func init() {
	SchemeBuilder.Register(&RDSInstance{}, &RDSInstanceList{})
	SchemeBuilder.Register(&DBSubnetGroup{}, &DBSubnetGroupList{})
}
