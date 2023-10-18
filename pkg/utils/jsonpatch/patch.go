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

package jsonpatch

import (
	"encoding/json"

	jsonpatch "github.com/evanphx/json-patch"
)

// TODO(muvaf): All the types that use CreateJSONPatch are known during
// development time. In order to avoid unnecessary panic checks, we can generate
// the code that creates a patch between two objects that share the same type.

// CreateJSONPatch creates a diff JSON object that can be applied to any other
// JSON object.
func CreateJSONPatch(source, destination interface{}) ([]byte, error) {
	sourceJSON, err := json.Marshal(source)
	if err != nil {
		return nil, err
	}
	destinationJSON, err := json.Marshal(destination)
	if err != nil {
		return nil, err
	}
	patchJSON, err := jsonpatch.CreateMergePatch(sourceJSON, destinationJSON)
	if err != nil {
		return nil, err
	}
	return patchJSON, nil
}
