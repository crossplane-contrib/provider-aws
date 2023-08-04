/*
Copyright 2019 The Crossplane Authors.

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

package aws

import (
	"encoding/json"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"

	jsonpatch "github.com/evanphx/json-patch"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

// String converts the supplied string for use with the AWS Go SDK.
func String(v string, o ...FieldOption) *string {
	for _, fo := range o {
		if fo == FieldRequired && v == "" {
			return aws.String(v)
		}
	}

	if v == "" {
		return nil
	}

	return aws.String(v)
}

// StringValue converts the supplied string pointer to a string, returning the
// empty string if the pointer is nil.
// TODO(muvaf): is this really meaningful? why not implement it?
func StringValue(v *string) string {
	return aws.ToString(v)
}

// StringSliceToPtr converts the supplied string array to an array of string pointers.
func StringSliceToPtr(slice []string) []*string {
	if slice == nil {
		return nil
	}

	res := make([]*string, len(slice))
	for i, s := range slice {
		res[i] = String(s)
	}
	return res
}

// StringPtrSliceToValue converts the supplied string pointer array to an array of strings.
func StringPtrSliceToValue(slice []*string) []string {
	if slice == nil {
		return nil
	}

	res := make([]string, len(slice))
	for i, s := range slice {
		res[i] = StringValue(s)
	}
	return res
}

// BoolValue calls underlying aws ToBool
func BoolValue(v *bool) bool {
	return aws.ToBool(v)
}

// Int64Value converts the supplied int64 pointer to a int64, returning
// 0 if the pointer is nil.
func Int64Value(v *int64) int64 {
	if v != nil {
		return *v
	}
	return 0
}

// Int32Value converts the supplied int32 pointer to a int32, returning
// 0 if the pointer is nil.
func Int32Value(v *int32) int32 {
	if v != nil {
		return *v
	}
	return 0
}

// TimeToMetaTime converts a standard Go time.Time to a K8s metav1.Time.
func TimeToMetaTime(t *time.Time) *metav1.Time {
	if t == nil {
		return nil
	}
	return &metav1.Time{Time: *t}
}

// LateInitializeStringPtr returns in if it's non-nil, otherwise returns from
// which is the backup for the cases in is nil.
func LateInitializeStringPtr(in *string, from *string) *string {
	if in != nil {
		return in
	}
	return from
}

// LateInitializeString returns `from` if `in` is empty and `from` is non-nil,
// in other cases it returns `in`.
func LateInitializeString(in string, from *string) string {
	if in == "" && from != nil {
		return *from
	}
	return in
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

// Int64 converts the supplied int for use with the AWS Go SDK.
func Int64(v int, o ...FieldOption) *int64 {
	for _, fo := range o {
		if fo == FieldRequired && v == 0 {
			return aws.Int64(int64(v))
		}
	}

	if v == 0 {
		return nil
	}

	return aws.Int64(int64(v))
}

// Int32 converts the supplied int for use with the AWS Go SDK.
func Int32(v int, o ...FieldOption) *int32 {
	for _, fo := range o {
		if fo == FieldRequired && v == 0 {
			return aws.Int32(int32(v))
		}
	}

	if v == 0 {
		return nil
	}

	return aws.Int32(int32(v))
}

// Int64Address returns the given *int in the form of *int64.
func Int64Address(i *int) *int64 {
	if i == nil {
		return nil
	}
	return aws.Int64(int64(*i))
}

// Int32Address returns the given *int in the form of *int32.
func Int32Address(i *int) *int32 {
	if i == nil {
		return nil
	}
	return aws.Int32(int32(*i))
}

// IntAddress converts the supplied int64 pointer to an int pointer, returning nil if
// the pointer is nil.
func IntAddress(i *int64) *int {
	if i == nil {
		return nil
	}
	r := int(*i)
	return &r
}

// IntFrom32Address converts the supplied int32 pointer to an int pointer, returning nil if
// the pointer is nil.
func IntFrom32Address(i *int32) *int {
	if i == nil {
		return nil
	}
	r := int(*i)
	return &r
}

// LateInitializeIntPtr returns in if it's non-nil, otherwise returns from
// which is the backup for the cases in is nil.
func LateInitializeIntPtr(in *int, from *int64) *int {
	if in != nil {
		return in
	}
	if from != nil {
		i := int(*from)
		return &i
	}
	return nil
}

// LateInitializeIntFrom32Ptr returns in if it's non-nil, otherwise returns from
// which is the backup for the cases in is nil.
// This function considered that nil and 0 values are same. However, for a *int32, nil and 0 values must be different
// because if the external AWS resource has a field with 0 value, during late initialization setting this value
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

// LateInitializeInt32Ptr returns in if it's non-nil, otherwise returns from
// which is the backup for the cases in is nil.
func LateInitializeInt32Ptr(in *int32, from *int32) *int32 {
	if in != nil {
		return in
	}
	return from
}

// LateInitializeInt64Ptr returns in if it's non-nil, otherwise returns from
// which is the backup for the cases in is nil.
func LateInitializeInt64Ptr(in *int64, from *int64) *int64 {
	if in != nil {
		return in
	}
	return from
}

// LateInitializeInt32 returns in if it's non-zero, otherwise returns from
// which is the backup for the cases in is zero.
func LateInitializeInt32(in int32, from int32) int32 {
	if in != 0 {
		return in
	}
	return from
}

// LateInitializeInt64 returns in if it's non-zero, otherwise returns from
// which is the backup for the cases in is zero.
func LateInitializeInt64(in int64, from int64) int64 {
	if in != 0 {
		return in
	}
	return from
}

// LateInitializeStringPtrSlice returns in if it's non-nil or from is zero
// length, otherwise it returns from.
func LateInitializeStringPtrSlice(in []*string, from []*string) []*string {
	if in != nil || len(from) == 0 {
		return in
	}

	return from
}

// LateInitializeInt64PtrSlice returns in if it's non-nil or from is zero
// length, otherwise it returns from.
func LateInitializeInt64PtrSlice(in []*int64, from []*int64) []*int64 {
	if in != nil || len(from) == 0 {
		return in
	}

	return from
}

// Bool converts the supplied bool for use with the AWS Go SDK.
func Bool(v bool, o ...FieldOption) *bool {
	for _, fo := range o {
		if fo == FieldRequired && !v {
			return aws.Bool(v)
		}
	}

	if !v {
		return nil
	}
	return aws.Bool(v)
}

// LateInitializeBoolPtr returns in if it's non-nil, otherwise returns from
// which is the backup for the cases in is nil.
func LateInitializeBoolPtr(in *bool, from *bool) *bool {
	if in != nil {
		return in
	}
	return from
}
