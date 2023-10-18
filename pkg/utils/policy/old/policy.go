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

// Deprecated use the policy package that contains better parser support.
package legacypolicy

import (
	"bytes"
	"encoding/json"
	"net/url"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"k8s.io/utils/ptr"
)

// CompactAndEscapeJSON removes space characters and URL-encodes the JSON string.
func CompactAndEscapeJSON(s string) (string, error) {
	buffer := new(bytes.Buffer)
	if err := json.Compact(buffer, []byte(s)); err != nil {
		return "", err
	}
	return url.QueryEscape(buffer.String()), nil
}

// IsPolicyUpToDate Marshall policies to json for a compare to get around string ordering
func IsPolicyUpToDate(local, remote *string) bool {
	var localUnmarshalled interface{}
	var remoteUnmarshalled interface{}

	var err error
	err = json.Unmarshal([]byte(ptr.Deref(local, "")), &localUnmarshalled)
	if err != nil {
		return false
	}
	err = json.Unmarshal([]byte(ptr.Deref(remote, "")), &remoteUnmarshalled)
	if err != nil {
		return false
	}

	sortSlicesOpt := cmpopts.SortSlices(func(x, y interface{}) bool {
		if a, ok := x.(string); ok {
			if b, ok := y.(string); ok {
				return a < b
			}
		}
		// Note: Unknown types in slices will not cause a panic, but
		// may not be sorted correctly. Depending on how AWS handles
		// these, it may cause constant updates - but better this than
		// panicing.
		return false
	})
	return cmp.Equal(localUnmarshalled, remoteUnmarshalled, cmpopts.EquateEmpty(), sortSlicesOpt)
}
