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

package manualv1alpha1

// AddTag adds a tag to this FargateProfile. If it already exists, it will be overwritten.
// It returns true if the tag has been added/changed. Otherwise false.
func (mg *FargateProfile) AddTag(key string, value string) bool {
	if mg.Spec.ForProvider.Tags == nil {
		mg.Spec.ForProvider.Tags = map[string]string{key: value}
		return true
	}
	oldValue, exists := mg.Spec.ForProvider.Tags[key]
	if !exists || oldValue != value {
		mg.Spec.ForProvider.Tags[key] = value
		return true
	}
	return false
}

// AddTag adds a tag to this IdentityProviderConfig. If it already exists, it will be overwritten.
// It returns true if the tag has been added/changed. Otherwise false.
func (mg *IdentityProviderConfig) AddTag(key string, value string) bool {
	if mg.Spec.ForProvider.Tags == nil {
		mg.Spec.ForProvider.Tags = map[string]string{key: value}
		return true
	}
	oldValue, exists := mg.Spec.ForProvider.Tags[key]
	if !exists || oldValue != value {
		mg.Spec.ForProvider.Tags[key] = value
		return true
	}
	return false
}

// AddTag adds a tag to this NodeGroup. If it already exists, it will be overwritten.
// It returns true if the tag has been added/changed. Otherwise false.
func (mg *NodeGroup) AddTag(key string, value string) bool {
	if mg.Spec.ForProvider.Tags == nil {
		mg.Spec.ForProvider.Tags = map[string]string{key: value}
		return true
	}
	oldValue, exists := mg.Spec.ForProvider.Tags[key]
	if !exists || oldValue != value {
		mg.Spec.ForProvider.Tags[key] = value
		return true
	}
	return false
}
