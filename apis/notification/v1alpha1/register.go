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

package v1alpha1

import (
	"reflect"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

// Package type metadata.
const (
	Group   = "notification.aws.crossplane.io"
	Version = "v1alpha1"
)

var (
	// SchemeGroupVersion is group version used to register these objects
	SchemeGroupVersion = schema.GroupVersion{Group: Group, Version: Version}

	// SchemeBuilder is used to add go types to the GroupVersionKind scheme
	SchemeBuilder = &scheme.Builder{GroupVersion: SchemeGroupVersion}
)

// SNSTopic type metadata.
var (
	SNSTopicKind             = reflect.TypeOf(SNSTopic{}).Name()
	SNSTopicGroupKind        = schema.GroupKind{Group: Group, Kind: SNSTopicKind}.String()
	SNSTopicKindAPIVersion   = SNSTopicKind + "." + SchemeGroupVersion.String()
	SNSTopicGroupVersionKind = SchemeGroupVersion.WithKind(SNSTopicKind)
)

// SNSSubscription type metadata.
var (
	SNSSubscriptionKind             = reflect.TypeOf(SNSSubscription{}).Name()
	SNSSubscriptionGroupKind        = schema.GroupKind{Group: Group, Kind: SNSSubscriptionKind}.String()
	SNSSubscriptionKindAPIVersion   = SNSSubscriptionKind + "." + SchemeGroupVersion.String()
	SNSSubscriptionGroupVersionKind = SchemeGroupVersion.WithKind(SNSSubscriptionKind)
)

func init() {
	SchemeBuilder.Register(&SNSTopic{}, &SNSTopicList{})
	SchemeBuilder.Register(&SNSSubscription{}, &SNSSubscriptionList{})
}
