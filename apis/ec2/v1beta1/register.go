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
// +groupName=network.aws.crossplane.io
// +versionName=v1beta1

package v1beta1

import (
	"reflect"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

// Package type metadata.
const (
	Group   = "ec2.aws.crossplane.io"
	Version = "v1beta1"
)

var (
	// SchemeGroupVersion is group version used to register these objects
	SchemeGroupVersion = schema.GroupVersion{Group: Group, Version: Version}

	// SchemeBuilder is used to add go types to the GroupVersionKind scheme
	SchemeBuilder = &scheme.Builder{GroupVersion: SchemeGroupVersion}
)

// VPC type metadata.
var (
	VPCKind             = reflect.TypeOf(VPC{}).Name()
	VPCGroupKind        = schema.GroupKind{Group: Group, Kind: VPCKind}.String()
	VPCKindAPIVersion   = VPCKind + "." + SchemeGroupVersion.String()
	VPCGroupVersionKind = SchemeGroupVersion.WithKind(VPCKind)
)

// Subnet type metadata.
var (
	SubnetKind             = reflect.TypeOf(Subnet{}).Name()
	SubnetGroupKind        = schema.GroupKind{Group: Group, Kind: SubnetKind}.String()
	SubnetKindAPIVersion   = SubnetKind + "." + SchemeGroupVersion.String()
	SubnetGroupVersionKind = SchemeGroupVersion.WithKind(SubnetKind)
)

// SecurityGroup type metadata.
var (
	SecurityGroupKind             = reflect.TypeOf(SecurityGroup{}).Name()
	SecurityGroupGroupKind        = schema.GroupKind{Group: Group, Kind: SecurityGroupKind}.String()
	SecurityGroupKindAPIVersion   = SecurityGroupKind + "." + SchemeGroupVersion.String()
	SecurityGroupGroupVersionKind = SchemeGroupVersion.WithKind(SecurityGroupKind)
)

// InternetGateway type metadata.
var (
	InternetGatewayKind             = reflect.TypeOf(InternetGateway{}).Name()
	InternetGatewayGroupKind        = schema.GroupKind{Group: Group, Kind: InternetGatewayKind}.String()
	InternetGatewayKindAPIVersion   = InternetGatewayKind + "." + SchemeGroupVersion.String()
	InternetGatewayGroupVersionKind = SchemeGroupVersion.WithKind(InternetGatewayKind)
)

// RouteTable type metadata.
var (
	RouteTableKind             = reflect.TypeOf(RouteTable{}).Name()
	RouteTableGroupKind        = schema.GroupKind{Group: Group, Kind: RouteTableKind}.String()
	RouteTableKindAPIVersion   = RouteTableKind + "." + SchemeGroupVersion.String()
	RouteTableGroupVersionKind = SchemeGroupVersion.WithKind(RouteTableKind)
)

func init() {
	SchemeBuilder.Register(&VPC{}, &VPCList{})
	SchemeBuilder.Register(&Subnet{}, &SubnetList{})
	SchemeBuilder.Register(&SecurityGroup{}, &SecurityGroupList{})
	SchemeBuilder.Register(&InternetGateway{}, &InternetGatewayList{})
	SchemeBuilder.Register(&RouteTable{}, &RouteTableList{})
}
