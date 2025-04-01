/*
Copyright 2025 The Crossplane Authors.

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

// ObjectLockRule the container element for an Object Lock rule.
type ObjectLockRule struct {
	// The container element for optionally specifying the default Object Lock
	// retention settings for new objects placed in the specified bucket.
	// The DefaultRetention settings require both a mode and a period.
	// The DefaultRetention period can be either Days or Years but you must select one
	// You cannot specify Days and Years at the same time.
	// For more information, see Object Lock(https://docs.aws.amazon.com/AmazonS3/latest/userguide/object-lock.html)
	DefaultRetention *DefaultRetention `json:"defaultRetention,omitempty"`
}

// DefaultRetention the container element for the default Object Lock retention settings.
type DefaultRetention struct {
	// The number of days that you want to specify for the default retention period.
	// Must be used with Mode.
	Days *int32 `json:"days,omitempty"`

	// The default Object Lock retention mode you want to apply to new objects placed
	// in the specified bucket. Must be used with either Days or Years .
	// Valid values are "GOVERNANCE", "COMPLIANCE"
	// +kubebuilder:validation:Enum=GOVERNANCE;COMPLIANCE;
	Mode string `json:"mode"`

	// The number of years that you want to specify for the default retention period.
	// Must be used with Mode .
	Years *int32 `json:"years,omitempty"`
}
