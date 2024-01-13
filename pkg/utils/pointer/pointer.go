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

package pointer

import (
	"k8s.io/utils/ptr"
)

// StringValue converts the supplied string pointer to a string, returning the
// empty string if the pointer is nil.
// TODO(muvaf): is this really meaningful? why not implement it?
func StringValue(v *string) string {
	return ptr.Deref(v, "")
}

// BoolValue calls underlying aws ToBool
func BoolValue(v *bool) bool {
	return ptr.Deref(v, false)
}

// Int64Value converts the supplied int64 pointer to a int64, returning
// 0 if the pointer is nil.
func Int64Value(v *int64) int64 {
	return ptr.Deref(v, 0)
}

// Int32Value converts the supplied int32 pointer to a int32, returning
// 0 if the pointer is nil.
func Int32Value(v *int32) int32 {
	return ptr.Deref(v, 0)
}

// Int64 converts the supplied int for use with the AWS Go SDK.
func ToIntAsInt64(v int) *int64 {
	if v == 0 {
		return nil
	}
	val64 := int64(v)
	return &val64
}

// Int32 converts the supplied int for use with the AWS Go SDK.
func ToIntAsInt32(v int) *int32 {
	if v == 0 {
		return nil
	}
	val32 := int32(v)
	return &val32
}

// Int32Address returns the given *int in the form of *int32.
func ToIntAsInt32Ptr(v *int) *int32 {
	if v == nil {
		return nil
	}
	val32 := int32(*v)
	return &val32
}

// ToInt32FromIntPtr converts an int32 pointer to an int pointer.
func ToInt32FromIntPtr(v *int32) *int {
	if v == nil {
		return nil
	}
	val := int(*v)
	return &val
}

// ToOrNilIfZeroValue returns a pointer to val if it does NOT match the default
// value of T. Otherwise it returns nil.
func ToOrNilIfZeroValue[T comparable](val T) *T {
	var defaultVal T
	if val == defaultVal {
		return nil
	}
	return &val
}
