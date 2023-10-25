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

package utils

import (
	"testing"

	svcsdk "github.com/aws/aws-sdk-go/service/efs"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/efs/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
)

func TestDiffTags(t *testing.T) {
	type args struct {
		spec    []*svcapitypes.Tag
		current []*svcsdk.Tag
	}
	type want struct {
		add    []*svcsdk.Tag
		remove []*string
	}

	sdkTag := func(k, v string) *svcsdk.Tag {
		return &svcsdk.Tag{Key: &k, Value: &v}
	}

	apiTag := func(k, v string) *svcapitypes.Tag {
		return &svcapitypes.Tag{Key: &k, Value: &v}
	}

	cases := map[string]struct {
		args args
		want want
	}{
		"DoNotAddOrRemoveAWSystemTags": {
			args: args{
				spec: []*svcapitypes.Tag{
					apiTag("foo", "bar"),
					apiTag("testAdd", "val"),
					apiTag("aws:some-system-tag", "enabled"),
				},
				current: []*svcsdk.Tag{
					sdkTag("foo", "bar"),
					sdkTag("testRemove", "val2"),
					sdkTag("aws:elasticfilesystem:default-backup", "enabled"),
				},
			},
			want: want{
				add: []*svcsdk.Tag{
					sdkTag("testAdd", "val"),
				},
				remove: []*string{
					pointer.ToOrNilIfZeroValue("testRemove"),
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			add, remove := DiffTags(tc.args.spec, tc.args.current)
			if diff := cmp.Diff(tc.want.add, add); diff != "" {
				t.Errorf("add: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.remove, remove, cmpopts.SortSlices(func(a, b string) bool { return a > b })); diff != "" {
				t.Errorf("remove: -want, +got:\n%s", diff)
			}
		})
	}
}
