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

	svcsdk "github.com/aws/aws-sdk-go/service/rds"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
)

func strPtrs(vals ...string) []*string {
	out := make([]*string, 0, len(vals))
	for _, v := range vals {
		out = append(out, pointer.ToOrNilIfZeroValue(v))
	}
	return out
}

func TestGenerateCloudWatchExportConfiguration(t *testing.T) {
	sortStrPtr := cmpopts.SortSlices(func(a, b *string) bool {
		return pointer.StringValue(a) < pointer.StringValue(b)
	})

	cases := map[string]struct {
		spec    []*string
		current []*string
		want    *svcsdk.CloudwatchLogsExportConfiguration
	}{
		"NoChange": {
			spec:    strPtrs("audit", "error"),
			current: strPtrs("error", "audit"),
			want:    nil,
		},
		"EnableOnly": {
			spec:    strPtrs("audit", "error"),
			current: strPtrs("audit"),
			want: &svcsdk.CloudwatchLogsExportConfiguration{
				EnableLogTypes:  strPtrs("error"),
				DisableLogTypes: []*string{},
			},
		},
		"DisableOnly": {
			spec:    strPtrs("audit"),
			current: strPtrs("audit", "error"),
			want: &svcsdk.CloudwatchLogsExportConfiguration{
				EnableLogTypes:  []*string{},
				DisableLogTypes: strPtrs("error"),
			},
		},
		"EnableAndDisable": {
			spec:    strPtrs("audit", "general"),
			current: strPtrs("audit", "error"),
			want: &svcsdk.CloudwatchLogsExportConfiguration{
				EnableLogTypes:  strPtrs("general"),
				DisableLogTypes: strPtrs("error"),
			},
		},
		"BothEmpty": {
			spec:    nil,
			current: nil,
			want:    nil,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := GenerateCloudWatchExportConfiguration(tc.spec, tc.current)
			if diff := cmp.Diff(tc.want, got, sortStrPtr); diff != "" {
				t.Errorf("GenerateCloudWatchExportConfiguration(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestAreSameElements(t *testing.T) {
	cases := map[string]struct {
		a1   []*string
		a2   []*string
		want bool
	}{
		"SameOrder":        {strPtrs("audit", "error"), strPtrs("audit", "error"), true},
		"DifferentOrder":   {strPtrs("audit", "error"), strPtrs("error", "audit"), true},
		"DifferentLength":  {strPtrs("audit"), strPtrs("audit", "error"), false},
		"DifferentElement": {strPtrs("audit", "error"), strPtrs("audit", "general"), false},
		"BothEmpty":        {nil, nil, true},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			if got := AreSameElements(tc.a1, tc.a2); got != tc.want {
				t.Errorf("AreSameElements(...): want %v, got %v", tc.want, got)
			}
		})
	}
}
