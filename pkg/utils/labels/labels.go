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

package labels

// DiffLabels returns labels that should be added, modified, or removed.
func DiffLabels(local, remote map[string]string) (addOrModify map[string]string, remove []string) {
	addOrModify = make(map[string]string, len(local))
	remove = []string{}
	for k, v := range local {
		addOrModify[k] = v
	}
	for k, v := range remote {
		switch val, ok := local[k]; {
		case ok && val != v:
			// if value does not match key it will be updated by the correct
			// key-value pair being present in the returned addOrModify map
			continue
		case !ok:
			remove = append(remove, k)
			delete(addOrModify, k)
		default:
			delete(addOrModify, k)
		}
	}
	return
}
