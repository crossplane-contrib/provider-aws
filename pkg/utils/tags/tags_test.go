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

package tags

import (
	"testing"

	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"k8s.io/utils/ptr"
)

func TestDiffTags(t *testing.T) {
	type args struct {
		local  map[string]string
		remote map[string]string
	}

	type want struct {
		add    map[string]string
		remove []string
	}

	cases := map[string]struct {
		args args
		want want
	}{
		"Add": {
			args: args{
				local:  map[string]string{"key": "val", "another": "tag"},
				remote: map[string]string{},
			},
			want: want{
				add: map[string]string{
					"key":     "val",
					"another": "tag",
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
				add:    map[string]string{},
				remove: []string{"key", "test"},
			},
		},
		"AddAndRemove": {
			args: args{
				local:  map[string]string{"key": "val", "another": "tag"},
				remote: map[string]string{"key": "val", "test": "one"},
			},
			want: want{
				add: map[string]string{
					"another": "tag",
				},
				remove: []string{"test"},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			add, remove := DiffTags(tc.args.local, tc.args.remote)
			if diff := cmp.Diff(tc.want.add, add); diff != "" {
				t.Errorf("add: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.remove, remove, cmpopts.SortSlices(func(a, b string) bool { return a > b })); diff != "" {
				t.Errorf("remove: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestDiffTagsMapPtr(t *testing.T) {
	type args struct {
		cr  map[string]*string
		obj map[string]*string
	}
	type want struct {
		addTags    map[string]*string
		removeTags []*string
	}

	cases := map[string]struct {
		args
		want
	}{
		"AddNewTag": {
			args: args{
				cr: map[string]*string{
					"k1": ptr.To("exists_in_both"),
					"k2": ptr.To("only_in_cr"),
				},
				obj: map[string]*string{
					"k1": ptr.To("exists_in_both"),
				}},
			want: want{
				addTags: map[string]*string{
					"k2": ptr.To("only_in_cr"),
				},
				removeTags: []*string{},
			},
		},
		"RemoveExistingTag": {
			args: args{
				cr: map[string]*string{
					"k1": ptr.To("exists_in_both"),
				},
				obj: map[string]*string{
					"k1": ptr.To("exists_in_both"),
					"k2": ptr.To("only_in_aws"),
				}},
			want: want{
				addTags: map[string]*string{},
				removeTags: []*string{
					ptr.To("k2"),
				}},
		},
		"AddAndRemoveWhenKeyChanges": {
			args: args{
				cr: map[string]*string{
					"k1": ptr.To("exists_in_both"),
					"k2": ptr.To("same_key_different_value_1"),
				},
				obj: map[string]*string{
					"k1": ptr.To("exists_in_both"),
					"k2": ptr.To("same_key_different_value_2"),
				}},
			want: want{
				addTags: map[string]*string{
					"k2": ptr.To("same_key_different_value_1"),
				},
				removeTags: []*string{
					ptr.To("k2"),
				}},
		},
		"NoChange": {
			args: args{
				cr: map[string]*string{
					"k1": ptr.To("exists_in_both"),
				},
				obj: map[string]*string{
					"k1": ptr.To("exists_in_both"),
				}},
			want: want{
				addTags:    map[string]*string{},
				removeTags: []*string{},
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			// Act
			addTags, removeTags := DiffTagsMapPtr(tc.args.cr, tc.args.obj)

			// Assert
			if diff := cmp.Diff(tc.want.addTags, addTags, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.removeTags, removeTags, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}
