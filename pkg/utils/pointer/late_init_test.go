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
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"k8s.io/utils/ptr"
)

func expect[T any](t *testing.T, got, want T) {
	tcName := reflect.TypeOf(got).Name()
	t.Run(tcName, func(t *testing.T) {
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("-want, +got:\n%s", diff)
		}
	})
}

type TestStruct struct {
	Val int
}

func TestLateInitialize(t *testing.T) {
	// int
	expect(t, LateInitialize(nil, ptr.To(10)), ptr.To(10))
	expect(t, LateInitialize(ptr.To(0), ptr.To(10)), ptr.To(0))
	expect(t, LateInitialize(ptr.To(0), nil), ptr.To(0))

	// int32
	expect(t, LateInitialize(nil, ptr.To[int32](10)), ptr.To[int32](10))
	expect(t, LateInitialize(ptr.To[int32](0), ptr.To[int32](10)), ptr.To[int32](0))
	expect(t, LateInitialize(ptr.To[int32](0), nil), ptr.To[int32](0))

	// int64
	expect(t, LateInitialize(nil, ptr.To[int64](10)), ptr.To[int64](10))
	expect(t, LateInitialize(ptr.To[int64](0), ptr.To[int64](10)), ptr.To[int64](0))
	expect(t, LateInitialize(ptr.To[int64](0), nil), ptr.To[int64](0))

	// string
	expect(t, LateInitialize(nil, ptr.To("")), ptr.To(""))
	expect(t, LateInitialize(ptr.To(""), ptr.To("new")), ptr.To(""))
	expect(t, LateInitialize(ptr.To(""), nil), ptr.To(""))

	// bool
	expect(t, LateInitialize(nil, ptr.To(false)), ptr.To(false))
	expect(t, LateInitialize(ptr.To(false), ptr.To(true)), ptr.To(false))
	expect(t, LateInitialize(ptr.To(true), ptr.To(false)), ptr.To(true))
	expect(t, LateInitialize(ptr.To(false), nil), ptr.To(false))

	// struct
	expect(t, LateInitialize(nil, ptr.To(TestStruct{})), ptr.To(TestStruct{}))
	expect(t, LateInitialize(ptr.To(TestStruct{}), ptr.To(TestStruct{Val: 2})), ptr.To(TestStruct{}))
	expect(t, LateInitialize(ptr.To(TestStruct{}), nil), ptr.To(TestStruct{}))

	// int
	expect(t, LateInitialize(0, 10), 10)
	expect(t, LateInitialize(10, 50), 10)
	expect(t, LateInitialize(10, 0), 10)

	// int32
	expect(t, LateInitialize(int32(0), 10), 10)
	expect(t, LateInitialize(int32(10), 50), 10)
	expect(t, LateInitialize(int32(10), 0), 10)

	// int64
	expect(t, LateInitialize(int64(0), 10), 10)
	expect(t, LateInitialize(int64(10), 50), 10)
	expect(t, LateInitialize(int64(10), 0), 10)

	// string
	expect(t, LateInitialize("", "new"), "new")
	expect(t, LateInitialize("old", "new"), "old")
	expect(t, LateInitialize("old", ""), "old")

	// bool
	expect(t, LateInitialize(false, true), true)
	expect(t, LateInitialize(true, false), true)
	expect(t, LateInitialize(false, false), false)
	expect(t, LateInitialize(true, true), true)

	// struct
	expect(t, LateInitialize(TestStruct{}, TestStruct{Val: 2}), TestStruct{Val: 2})
	expect(t, LateInitialize(TestStruct{Val: 2}, TestStruct{Val: 4}), TestStruct{Val: 2})
	expect(t, LateInitialize(TestStruct{Val: 2}, TestStruct{}), TestStruct{Val: 2})
}

func TestLateInitializeValueFromPtr(t *testing.T) {
	// int
	expect(t, LateInitializeValueFromPtr(0, ptr.To(10)), 10)
	expect(t, LateInitializeValueFromPtr(10, ptr.To(50)), 10)
	expect(t, LateInitializeValueFromPtr(10, ptr.To(0)), 10)

	// int32
	expect(t, LateInitializeValueFromPtr(int32(0), ptr.To[int32](10)), 10)
	expect(t, LateInitializeValueFromPtr(int32(10), ptr.To[int32](50)), 10)
	expect(t, LateInitializeValueFromPtr(int32(10), ptr.To[int32](0)), 10)
	expect(t, LateInitializeValueFromPtr(int32(0), nil), 0)

	// int64
	expect(t, LateInitializeValueFromPtr(int64(0), ptr.To[int64](10)), 10)
	expect(t, LateInitializeValueFromPtr(int64(10), ptr.To[int64](50)), 10)
	expect(t, LateInitializeValueFromPtr(int64(10), ptr.To[int64](0)), 10)
	expect(t, LateInitializeValueFromPtr(int64(0), nil), 0)

	// string
	expect(t, LateInitializeValueFromPtr("", ptr.To("new")), "new")
	expect(t, LateInitializeValueFromPtr("old", ptr.To("new")), "old")
	expect(t, LateInitializeValueFromPtr("old", ptr.To("")), "old")
	expect(t, LateInitializeValueFromPtr("", nil), "")

	// bool
	expect(t, LateInitializeValueFromPtr(false, ptr.To(true)), true)
	expect(t, LateInitializeValueFromPtr(true, ptr.To(false)), true)
	expect(t, LateInitializeValueFromPtr(true, ptr.To(true)), true)
	expect(t, LateInitializeValueFromPtr(false, nil), false)

	// struct
	expect(t, LateInitializeValueFromPtr(TestStruct{}, &TestStruct{Val: 2}), TestStruct{Val: 2})
	expect(t, LateInitializeValueFromPtr(TestStruct{Val: 2}, &TestStruct{Val: 4}), TestStruct{Val: 2})
	expect(t, LateInitializeValueFromPtr(TestStruct{Val: 2}, &TestStruct{}), TestStruct{Val: 2})
	expect(t, LateInitializeValueFromPtr(TestStruct{Val: 2}, nil), TestStruct{Val: 2})
	expect(t, LateInitializeValueFromPtr(TestStruct{}, nil), TestStruct{})
}

func TestLateInitializeSlice(t *testing.T) {
	// values
	expect(t, LateInitializeSlice(nil, []int{1, 2, 3}), []int{1, 2, 3})
	expect(t, LateInitializeSlice([]int{}, []int{1, 2, 3}), []int{})
	expect(t, LateInitializeSlice(nil, []int{}), nil)
	expect(t, LateInitializeSlice([]int{4, 5}, []int{1, 2, 3}), []int{4, 5})
	expect(t, LateInitializeSlice([]int{4, 5}, nil), []int{4, 5})

	// pointer
	expect(t, LateInitializeSlice(nil, []*int{ptr.To(1), ptr.To(2), ptr.To(3)}), []*int{ptr.To(1), ptr.To(2), ptr.To(3)})
	expect(t, LateInitializeSlice([]*int{}, []*int{ptr.To(1), ptr.To(2), ptr.To(3)}), []*int{})
	expect(t, LateInitializeSlice(nil, []*int{}), nil)
	expect(t, LateInitializeSlice([]*int{ptr.To(4), ptr.To(5)}, []*int{ptr.To(1), ptr.To(2), ptr.To(3)}), []*int{ptr.To(4), ptr.To(5)})
	expect(t, LateInitializeSlice([]*int{ptr.To(4), ptr.To(5)}, nil), []*int{ptr.To(4), ptr.To(5)})
}
