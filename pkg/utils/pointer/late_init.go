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
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

// LateInitialize return from if current matches the default value of T.
// Otherwise it returns current.
func LateInitialize[T comparable](current, from T) T {
	var defaultVal T
	if current == defaultVal {
		return from
	}
	return current
}

// LateInitializeValueFromPtr returns from if current matches the default value
// of T. Otherwise it returns from or the default value if from is nil as well.
func LateInitializeValueFromPtr[T comparable](current T, from *T) T {
	var defaultVal T
	if current == defaultVal {
		return ptr.Deref(from, defaultVal)
	}
	return current
}

// LateInitializeSlice returns from if current is nil and from is not empty.
// Otherwise it returns current.
func LateInitializeSlice[T any](current, from []T) []T {
	if current != nil || len(from) == 0 {
		return current
	}
	return from
}

// LateInitializeIntFrom32Ptr returns in if it's non-nil, otherwise returns from
// which is the backup for the cases in is nil.
// This function considered that nil and 0 values are same.
// However, for a *int32, nil and 0 values must be different
// because if the external AWS resource has a field with 0 value, during late
// initialization setting this value
// in CR must be allowed. Please see the LateInitializeIntFromInt32Ptr func.
func LateInitializeIntFrom32Ptr(in *int, from *int32) *int {
	if in != nil {
		return in
	}
	if from != nil && *from != 0 {
		i := int(*from)
		return &i
	}
	return nil
}

// LateInitializeIntFromInt32Ptr returns in if it's non-nil, otherwise returns from
// which is the backup for the cases in is nil.
func LateInitializeIntFromInt32Ptr(in *int, from *int32) *int {
	if in != nil {
		return in
	}

	if from != nil {
		i := int(*from)
		return &i
	}

	return nil
}

// LateInitializeTimePtr returns in if it's non-nil, otherwise returns from
// which is the backup for the cases in is nil.
func LateInitializeTimePtr(in *metav1.Time, from *time.Time) *metav1.Time {
	if in != nil {
		return in
	}
	if from != nil {
		t := metav1.NewTime(*from)
		return &t
	}
	return nil
}
