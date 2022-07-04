/*
Copyright 2022 The Crossplane Authors.

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

// AddTag adds a tag to this Key. If it already exists, it will be overwritten.
// It returns true if the tag has been added/changed. Otherwise false.
func (mg *Key) AddTag(key string, value string) bool {
	newTag := &Tag{
		TagKey:   &key,
		TagValue: &value,
	}
	for i, ta := range mg.Spec.ForProvider.Tags {
		if ta != nil && ta.TagKey != nil && *ta.TagKey == key {
			if ta.TagValue != nil && *ta.TagValue == value {
				return false
			}
			mg.Spec.ForProvider.Tags[i] = newTag
			return true
		}
	}
	mg.Spec.ForProvider.Tags = append(mg.Spec.ForProvider.Tags, newTag)
	return true
}
