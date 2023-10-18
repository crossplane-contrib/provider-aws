/*
Copyright 2023 The Crossplane Authors.

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

package tags

import "k8s.io/utils/ptr"

// DiffTags returns tags that should be added or removed.
func DiffTags(local, remote map[string]string) (add map[string]string, remove []string) {
	add = make(map[string]string, len(local))
	remove = []string{}
	for k, v := range local {
		add[k] = v
	}
	for k, v := range remote {
		switch val, ok := local[k]; {
		case ok && val != v:
			remove = append(remove, k)
		case !ok:
			remove = append(remove, k)
			delete(add, k)
		default:
			delete(add, k)
		}
	}
	return
}

// DiffTagsMapPtr returns which AWS Tags exist in the resource tags and which are outdated and should be removed
func DiffTagsMapPtr(spec map[string]*string, current map[string]*string) (map[string]*string, []*string) {
	addMap := make(map[string]*string, len(spec))
	removeTags := make([]*string, 0)
	for k, v := range current {
		if ptr.Deref(spec[k], "") == ptr.Deref(v, "") {
			continue
		}
		removeTags = append(removeTags, ptr.To(k))
	}
	for k, v := range spec {
		if ptr.Deref(current[k], "") == ptr.Deref(v, "") {
			continue
		}
		addMap[k] = v
	}
	return addMap, removeTags
}
