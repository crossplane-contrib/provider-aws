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

package labels

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestDiffLabels(t *testing.T) {
	type args struct {
		local  map[string]string
		remote map[string]string
	}

	type want struct {
		addOrModify map[string]string
		remove      []string
	}

	cases := map[string]struct {
		args args
		want want
	}{
		"Add": {
			args: args{
				local:  map[string]string{"key": "val", "another": "label"},
				remote: map[string]string{},
			},
			want: want{
				addOrModify: map[string]string{
					"key":     "val",
					"another": "label",
				},
				remove: []string{},
			},
		},
		"Remove": {
			args: args{
				local: map[string]string{},

				remote: map[string]string{"key": "val", "test": "one"},
			},
			want: want{
				addOrModify: map[string]string{},
				remove:      []string{"key", "test"},
			},
		},
		"AddAndRemove": {
			args: args{
				local:  map[string]string{"key": "val", "another": "label"},
				remote: map[string]string{"key": "val", "test": "one"},
			},
			want: want{
				addOrModify: map[string]string{
					"another": "label",
				},
				remove: []string{"test"},
			},
		},
		"ModifyOnly": {
			args: args{
				local:  map[string]string{"key": "val"},
				remote: map[string]string{"key": "badval"},
			},
			want: want{
				addOrModify: map[string]string{
					"key": "val",
				},
				remove: []string{},
			},
		},
		"AddModifyRemove": {
			args: args{
				local:  map[string]string{"key": "val", "keytwo": "valtwo", "another": "tag"},
				remote: map[string]string{"key": "val", "keytwo": "badval", "test": "one"},
			},
			want: want{
				addOrModify: map[string]string{
					"keytwo":  "valtwo",
					"another": "tag",
				},
				remove: []string{"test"},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			addOrModify, remove := DiffLabels(tc.args.local, tc.args.remote)
			if diff := cmp.Diff(tc.want.addOrModify, addOrModify); diff != "" {
				t.Errorf("addOrModify: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.remove, remove, cmpopts.SortSlices(func(a, b string) bool { return a > b })); diff != "" {
				t.Errorf("remove: -want, +got:\n%s", diff)
			}
		})
	}
}
