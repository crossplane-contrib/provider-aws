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

// Code generated by provider-aws codegen. DO NOT EDIT.

package v1alpha1

import (
	pointer "k8s.io/utils/pointer"
	"sort"
)

// AddTag adds a tag to this DBCluster. If it already exists, it will be overwritten.
// It returns true if the tag has been added/changed. Otherwise false.
func (mg *DBCluster) AddTag(key string, value string) bool {
	newTag := &Tag{
		Key:   &key,
		Value: &value,
	}
	updated := false
	for i, ta := range mg.Spec.ForProvider.Tags {
		if ta != nil && pointer.StringDeref(ta.Key, "") == key {
			if pointer.StringDeref(ta.Value, "") == value {
				return false
			}
			mg.Spec.ForProvider.Tags[i] = newTag
			updated = true
			break
		}
	}
	if !updated {
		mg.Spec.ForProvider.Tags = append(mg.Spec.ForProvider.Tags, newTag)
	}
	sort.Slice(mg.Spec.ForProvider.Tags, func(i int, j int) bool {
		ta := mg.Spec.ForProvider.Tags[i]
		tb := mg.Spec.ForProvider.Tags[j]
		if ta == nil {
			return true
		} else if tb == nil {
			return false
		}
		return pointer.StringDeref(ta.Key, "") < pointer.StringDeref(tb.Key, "")
	})
	return true
}

// AddTag adds a tag to this DBClusterParameterGroup. If it already exists, it will be overwritten.
// It returns true if the tag has been added/changed. Otherwise false.
func (mg *DBClusterParameterGroup) AddTag(key string, value string) bool {
	newTag := &Tag{
		Key:   &key,
		Value: &value,
	}
	updated := false
	for i, ta := range mg.Spec.ForProvider.Tags {
		if ta != nil && pointer.StringDeref(ta.Key, "") == key {
			if pointer.StringDeref(ta.Value, "") == value {
				return false
			}
			mg.Spec.ForProvider.Tags[i] = newTag
			updated = true
			break
		}
	}
	if !updated {
		mg.Spec.ForProvider.Tags = append(mg.Spec.ForProvider.Tags, newTag)
	}
	sort.Slice(mg.Spec.ForProvider.Tags, func(i int, j int) bool {
		ta := mg.Spec.ForProvider.Tags[i]
		tb := mg.Spec.ForProvider.Tags[j]
		if ta == nil {
			return true
		} else if tb == nil {
			return false
		}
		return pointer.StringDeref(ta.Key, "") < pointer.StringDeref(tb.Key, "")
	})
	return true
}

// AddTag adds a tag to this DBInstance. If it already exists, it will be overwritten.
// It returns true if the tag has been added/changed. Otherwise false.
func (mg *DBInstance) AddTag(key string, value string) bool {
	newTag := &Tag{
		Key:   &key,
		Value: &value,
	}
	updated := false
	for i, ta := range mg.Spec.ForProvider.Tags {
		if ta != nil && pointer.StringDeref(ta.Key, "") == key {
			if pointer.StringDeref(ta.Value, "") == value {
				return false
			}
			mg.Spec.ForProvider.Tags[i] = newTag
			updated = true
			break
		}
	}
	if !updated {
		mg.Spec.ForProvider.Tags = append(mg.Spec.ForProvider.Tags, newTag)
	}
	sort.Slice(mg.Spec.ForProvider.Tags, func(i int, j int) bool {
		ta := mg.Spec.ForProvider.Tags[i]
		tb := mg.Spec.ForProvider.Tags[j]
		if ta == nil {
			return true
		} else if tb == nil {
			return false
		}
		return pointer.StringDeref(ta.Key, "") < pointer.StringDeref(tb.Key, "")
	})
	return true
}

// AddTag adds a tag to this DBParameterGroup. If it already exists, it will be overwritten.
// It returns true if the tag has been added/changed. Otherwise false.
func (mg *DBParameterGroup) AddTag(key string, value string) bool {
	newTag := &Tag{
		Key:   &key,
		Value: &value,
	}
	updated := false
	for i, ta := range mg.Spec.ForProvider.Tags {
		if ta != nil && pointer.StringDeref(ta.Key, "") == key {
			if pointer.StringDeref(ta.Value, "") == value {
				return false
			}
			mg.Spec.ForProvider.Tags[i] = newTag
			updated = true
			break
		}
	}
	if !updated {
		mg.Spec.ForProvider.Tags = append(mg.Spec.ForProvider.Tags, newTag)
	}
	sort.Slice(mg.Spec.ForProvider.Tags, func(i int, j int) bool {
		ta := mg.Spec.ForProvider.Tags[i]
		tb := mg.Spec.ForProvider.Tags[j]
		if ta == nil {
			return true
		} else if tb == nil {
			return false
		}
		return pointer.StringDeref(ta.Key, "") < pointer.StringDeref(tb.Key, "")
	})
	return true
}
