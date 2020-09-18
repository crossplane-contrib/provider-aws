/*
Copyright 2020 The Crossplane Authors.

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

// Tagging is the container for TagSet elements.
type Tagging struct {
	// A collection for a set of tags
	// TagSet is a required field
	TagSet []Tag `json:"tagSet"`
}

// Tag is a container for a key value name pair.
type Tag struct {
	// Name of the tag.
	// Key is a required field
	Key *string `json:"key,omitempty"`

	// Value of the tag.
	// Value is a required field
	Value *string `json:"value,omitempty"`
}
