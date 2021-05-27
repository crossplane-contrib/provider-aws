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

package iam

import (
	"testing"

	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"
)

func TestSliceDiff(t *testing.T) {
	type want struct {
		add    []string
		remove []string
	}
	type args struct {
		current []string
		desired []string
	}

	cases := map[string]struct {
		args
		want
	}{
		"AddRemove": {
			args: args{
				current: []string{"foo", "bar"},
				desired: []string{"a", "bar"},
			},
			want: want{
				add:    []string{"a"},
				remove: []string{"foo"},
			},
		},
		"NoChange": {
			args: args{
				current: []string{"foo", "bar"},
				desired: []string{"foo", "bar"},
			},
			want: want{
				add:    nil,
				remove: nil,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			add, remove := SliceDifference(tc.current, tc.desired)

			if diff := cmp.Diff(tc.want.add, add, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.remove, remove, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}
