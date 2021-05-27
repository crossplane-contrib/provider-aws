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
// +versionName=v1alpha1

package v1alpha1

import (
	"reflect"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

// Package type metadata.
const (
	Group   = "identity.aws.crossplane.io"
	Version = "v1alpha1"
)

var (
	// SchemeGroupVersion is group version used to register these objects
	SchemeGroupVersion = schema.GroupVersion{Group: Group, Version: Version}

	// SchemeBuilder is used to add go types to the GroupVersionKind scheme
	SchemeBuilder = &scheme.Builder{GroupVersion: SchemeGroupVersion}
)

// IAMUser type metadata.
var (
	IAMUserKind             = reflect.TypeOf(IAMUser{}).Name()
	IAMUserGroupKind        = schema.GroupKind{Group: Group, Kind: IAMUserKind}.String()
	IAMUserKindAPIVersion   = IAMUserKind + "." + SchemeGroupVersion.String()
	IAMUserGroupVersionKind = SchemeGroupVersion.WithKind(IAMUserKind)
)

// IAMUserPolicyAttachment type metadata.
var (
	IAMUserPolicyAttachmentKind             = reflect.TypeOf(IAMUserPolicyAttachment{}).Name()
	IAMUserPolicyAttachmentGroupKind        = schema.GroupKind{Group: Group, Kind: IAMUserPolicyAttachmentKind}.String()
	IAMUserPolicyAttachmentKindAPIVersion   = IAMUserPolicyAttachmentKind + "." + SchemeGroupVersion.String()
	IAMUserPolicyAttachmentGroupVersionKind = SchemeGroupVersion.WithKind(IAMUserPolicyAttachmentKind)
)

// IAMPolicy type metadata.
var (
	IAMPolicyKind             = reflect.TypeOf(IAMPolicy{}).Name()
	IAMPolicyGroupKind        = schema.GroupKind{Group: Group, Kind: IAMPolicyKind}.String()
	IAMPolicyKindAPIVersion   = IAMPolicyKind + "." + SchemeGroupVersion.String()
	IAMPolicyGroupVersionKind = SchemeGroupVersion.WithKind(IAMPolicyKind)
)

// IAMGroup type metadata
var (
	IAMGroupKind             = reflect.TypeOf(IAMGroup{}).Name()
	IAMGroupGroupKind        = schema.GroupKind{Group: Group, Kind: IAMGroupKind}.String()
	IAMGroupKindAPIVersion   = IAMGroupKind + "." + SchemeGroupVersion.String()
	IAMGroupGroupVersionKind = SchemeGroupVersion.WithKind(IAMGroupKind)
)

// IAMGroupUserMembership type metadata.
var (
	IAMGroupUserMembershipKind             = reflect.TypeOf(IAMGroupUserMembership{}).Name()
	IAMGroupUserMembershipGroupKind        = schema.GroupKind{Group: Group, Kind: IAMGroupUserMembershipKind}.String()
	IAMGroupUserMembershipKindAPIVersion   = IAMGroupUserMembershipKind + "." + SchemeGroupVersion.String()
	IAMGroupUserMembershipGroupVersionKind = SchemeGroupVersion.WithKind(IAMGroupUserMembershipKind)
)

// IAMGroupPolicyAttachment type metadata.
var (
	IAMGroupPolicyAttachmentKind             = reflect.TypeOf(IAMGroupPolicyAttachment{}).Name()
	IAMGroupPolicyAttachmentGroupKind        = schema.GroupKind{Group: Group, Kind: IAMGroupPolicyAttachmentKind}.String()
	IAMGroupPolicyAttachmentKindAPIVersion   = IAMGroupPolicyAttachmentKind + "." + SchemeGroupVersion.String()
	IAMGroupPolicyAttachmentGroupVersionKind = SchemeGroupVersion.WithKind(IAMGroupPolicyAttachmentKind)
)

// IAMAccessKey type metadata.
var (
	IAMAccessKeyKind             = reflect.TypeOf(IAMAccessKey{}).Name()
	IAMAccessKeyGroupKind        = schema.GroupKind{Group: Group, Kind: IAMAccessKeyKind}.String()
	IAMAccessKeyKindAPIVersion   = IAMAccessKeyKind + "." + SchemeGroupVersion.String()
	IAMAccessKeyGroupVersionKind = SchemeGroupVersion.WithKind(IAMAccessKeyKind)
)

// OpenIDConnectProvider type metadata.
var (
	OpenIDConnectProviderKind             = "OpenIDConnectProvider"
	OpenIDConnectProviderGroupKind        = schema.GroupKind{Group: Group, Kind: OpenIDConnectProviderKind}.String()
	OpenIDConnectProviderKindAPIVersion   = OpenIDConnectProviderKind + "." + SchemeGroupVersion.String()
	OpenIDConnectProviderGroupVersionKind = SchemeGroupVersion.WithKind(OpenIDConnectProviderKind)
)

func init() {
	SchemeBuilder.Register(&IAMUser{}, &IAMUserList{})
	SchemeBuilder.Register(&IAMPolicy{}, &IAMPolicyList{})
	SchemeBuilder.Register(&IAMUserPolicyAttachment{}, &IAMUserPolicyAttachmentList{})
	SchemeBuilder.Register(&IAMGroup{}, &IAMGroupList{})
	SchemeBuilder.Register(&IAMGroupUserMembership{}, &IAMGroupUserMembershipList{})
	SchemeBuilder.Register(&IAMGroupPolicyAttachment{}, &IAMGroupPolicyAttachmentList{})
	SchemeBuilder.Register(&IAMAccessKey{}, &IAMAccessKeyList{})
	SchemeBuilder.Register(&OpenIDConnectProvider{}, &OpenIDConnectProviderList{})
}
