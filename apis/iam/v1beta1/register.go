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
// +groupName=iam.aws.crossplane.io
// +versionName=v1beta1

package v1beta1

import (
	"reflect"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

// Package type metadata.
const (
	CRDGroup   = "iam.aws.crossplane.io"
	CRDVersion = "v1beta1"
)

var (
	// SchemeGroupVersion is group version used to register these objects
	SchemeGroupVersion = schema.GroupVersion{Group: CRDGroup, Version: CRDVersion}

	// SchemeBuilder is used to add go types to the GroupVersionKind scheme
	SchemeBuilder = &scheme.Builder{GroupVersion: SchemeGroupVersion}
)

// Role type metadata.
var (
	RoleKind             = reflect.TypeOf(Role{}).Name()
	RoleGroupKind        = schema.GroupKind{Group: CRDGroup, Kind: RoleKind}.String()
	RoleKindAPIVersion   = RoleKind + "." + SchemeGroupVersion.String()
	RoleGroupVersionKind = SchemeGroupVersion.WithKind(RoleKind)
)

// RolePolicyAttachment type metadata.
var (
	RolePolicyAttachmentKind             = reflect.TypeOf(RolePolicyAttachment{}).Name()
	RolePolicyAttachmentGroupKind        = schema.GroupKind{Group: CRDGroup, Kind: RolePolicyAttachmentKind}.String()
	RolePolicyAttachmentKindAPIVersion   = RolePolicyAttachmentKind + "." + SchemeGroupVersion.String()
	RolePolicyAttachmentGroupVersionKind = SchemeGroupVersion.WithKind(RolePolicyAttachmentKind)
)

// User type metadata.
var (
	UserKind             = reflect.TypeOf(User{}).Name()
	UserGroupKind        = schema.GroupKind{Group: CRDGroup, Kind: UserKind}.String()
	UserKindAPIVersion   = UserKind + "." + SchemeGroupVersion.String()
	UserGroupVersionKind = SchemeGroupVersion.WithKind(UserKind)
)

// UserPolicyAttachment type metadata.
var (
	UserPolicyAttachmentKind             = reflect.TypeOf(UserPolicyAttachment{}).Name()
	UserPolicyAttachmentGroupKind        = schema.GroupKind{Group: CRDGroup, Kind: UserPolicyAttachmentKind}.String()
	UserPolicyAttachmentKindAPIVersion   = UserPolicyAttachmentKind + "." + SchemeGroupVersion.String()
	UserPolicyAttachmentGroupVersionKind = SchemeGroupVersion.WithKind(UserPolicyAttachmentKind)
)

// Policy type metadata.
var (
	PolicyKind             = reflect.TypeOf(Policy{}).Name()
	PolicyGroupKind        = schema.GroupKind{Group: CRDGroup, Kind: PolicyKind}.String()
	PolicyKindAPIVersion   = PolicyKind + "." + SchemeGroupVersion.String()
	PolicyGroupVersionKind = SchemeGroupVersion.WithKind(PolicyKind)
)

// Group type metadata
var (
	GroupKind             = reflect.TypeOf(Group{}).Name()
	GroupGroupKind        = schema.GroupKind{Group: CRDGroup, Kind: GroupKind}.String()
	GroupKindAPIVersion   = GroupKind + "." + SchemeGroupVersion.String()
	GroupGroupVersionKind = SchemeGroupVersion.WithKind(GroupKind)
)

// GroupUserMembership type metadata.
var (
	GroupUserMembershipKind             = reflect.TypeOf(GroupUserMembership{}).Name()
	GroupUserMembershipGroupKind        = schema.GroupKind{Group: CRDGroup, Kind: GroupUserMembershipKind}.String()
	GroupUserMembershipKindAPIVersion   = GroupUserMembershipKind + "." + SchemeGroupVersion.String()
	GroupUserMembershipGroupVersionKind = SchemeGroupVersion.WithKind(GroupUserMembershipKind)
)

// GroupPolicyAttachment type metadata.
var (
	GroupPolicyAttachmentKind             = reflect.TypeOf(GroupPolicyAttachment{}).Name()
	GroupPolicyAttachmentGroupKind        = schema.GroupKind{Group: CRDGroup, Kind: GroupPolicyAttachmentKind}.String()
	GroupPolicyAttachmentKindAPIVersion   = GroupPolicyAttachmentKind + "." + SchemeGroupVersion.String()
	GroupPolicyAttachmentGroupVersionKind = SchemeGroupVersion.WithKind(GroupPolicyAttachmentKind)
)

// AccessKey type metadata.
var (
	AccessKeyKind             = reflect.TypeOf(AccessKey{}).Name()
	AccessKeyGroupKind        = schema.GroupKind{Group: CRDGroup, Kind: AccessKeyKind}.String()
	AccessKeyKindAPIVersion   = AccessKeyKind + "." + SchemeGroupVersion.String()
	AccessKeyGroupVersionKind = SchemeGroupVersion.WithKind(AccessKeyKind)
)

// OpenIDConnectProvider type metadata.
var (
	OpenIDConnectProviderKind             = "OpenIDConnectProvider"
	OpenIDConnectProviderGroupKind        = schema.GroupKind{Group: CRDGroup, Kind: OpenIDConnectProviderKind}.String()
	OpenIDConnectProviderKindAPIVersion   = OpenIDConnectProviderKind + "." + SchemeGroupVersion.String()
	OpenIDConnectProviderGroupVersionKind = SchemeGroupVersion.WithKind(OpenIDConnectProviderKind)
)

// RolePolicy type metadata.
var (
	RolePolicyKind             = reflect.TypeOf(RolePolicy{}).Name()
	RolePolicyGroupKind        = schema.GroupKind{Group: CRDGroup, Kind: RolePolicyKind}.String()
	RolePolicyKindAPIVersion   = RolePolicyKind + "." + SchemeGroupVersion.String()
	RolePolicyGroupVersionKind = SchemeGroupVersion.WithKind(RolePolicyKind)
)

func init() {
	SchemeBuilder.Register(&Role{}, &RoleList{})
	SchemeBuilder.Register(&RolePolicy{}, &RolePolicyList{})
	SchemeBuilder.Register(&RolePolicyAttachment{}, &RolePolicyAttachmentList{})
	SchemeBuilder.Register(&User{}, &UserList{})
	SchemeBuilder.Register(&Policy{}, &PolicyList{})
	SchemeBuilder.Register(&UserPolicyAttachment{}, &UserPolicyAttachmentList{})
	SchemeBuilder.Register(&Group{}, &GroupList{})
	SchemeBuilder.Register(&GroupUserMembership{}, &GroupUserMembershipList{})
	SchemeBuilder.Register(&GroupPolicyAttachment{}, &GroupPolicyAttachmentList{})
	SchemeBuilder.Register(&AccessKey{}, &AccessKeyList{})
	SchemeBuilder.Register(&OpenIDConnectProvider{}, &OpenIDConnectProviderList{})
}
