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

package utils

import (
	svcsdk "github.com/aws/aws-sdk-go/service/transfer"
	"k8s.io/utils/ptr"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/transfer/v1alpha1"
)

// DiffTags returns tags that should be added or removed.
func DiffTags(spec []*svcapitypes.Tag, current []*svcsdk.Tag) (isUpToDate bool, add []*svcsdk.Tag, remove []*string) {
	tagsToAdd := make(map[string]string, len(spec))
	add = []*svcsdk.Tag{}
	remove = []*string{}
	for _, j := range spec {
		tagsToAdd[ptr.Deref(j.Key, "")] = ptr.Deref(j.Value, "")
	}
	for _, j := range current {
		switch val, ok := tagsToAdd[ptr.Deref(j.Key, "")]; {
		case ok && val == ptr.Deref(j.Value, ""):
			delete(tagsToAdd, ptr.Deref(j.Key, ""))
		case !ok:
			remove = append(remove, j.Key)
		}
	}
	for i, j := range tagsToAdd {
		add = append(add, &svcsdk.Tag{
			Key:   ptr.To(i),
			Value: ptr.To(j),
		})
	}
	isUpToDate = len(add) == 0 && len(remove) == 0
	return
}
