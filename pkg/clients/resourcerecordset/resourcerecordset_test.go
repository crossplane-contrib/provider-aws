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

package resourcerecordset

import (
	"testing"

	route53types "github.com/aws/aws-sdk-go-v2/service/route53/types"
	"github.com/google/go-cmp/cmp"

	"github.com/crossplane/provider-aws/apis/route53/v1alpha1"
)

func TestCreatePatch(t *testing.T) {

	resourceRecordSetName := "x.y.z."
	var ttl int64 = 300
	var ttl2 int64 = 200

	type args struct {
		rrSet route53types.ResourceRecordSet
		p     v1alpha1.ResourceRecordSetParameters
	}

	type want struct {
		patch *v1alpha1.ResourceRecordSetParameters
	}

	cases := map[string]struct {
		args
		want
	}{
		"SameFields": {
			args: args{
				rrSet: route53types.ResourceRecordSet{
					Name: &resourceRecordSetName,
					TTL:  &ttl,
				},
				p: v1alpha1.ResourceRecordSetParameters{
					TTL: &ttl,
				},
			},
			want: want{
				patch: &v1alpha1.ResourceRecordSetParameters{},
			},
		},
		"DifferentFields": {
			args: args{
				rrSet: route53types.ResourceRecordSet{
					Name: &resourceRecordSetName,
					TTL:  &ttl,
				},
				p: v1alpha1.ResourceRecordSetParameters{
					TTL: &ttl2,
				},
			},
			want: want{
				patch: &v1alpha1.ResourceRecordSetParameters{
					TTL: &ttl2,
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			result, _ := CreatePatch(&tc.args.rrSet, &tc.args.p)
			if diff := cmp.Diff(tc.want.patch, result); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestIsUpToDate(t *testing.T) {

	resourceRecordSetName := "x.y.z."
	var ttl int64 = 300
	var ttl2 int64 = 200

	type args struct {
		rrSet route53types.ResourceRecordSet
		p     v1alpha1.ResourceRecordSetParameters
	}

	cases := map[string]struct {
		args args
		want bool
	}{
		"SameFields": {
			args: args{
				rrSet: route53types.ResourceRecordSet{
					Name: &resourceRecordSetName,
					TTL:  &ttl,
				},
				p: v1alpha1.ResourceRecordSetParameters{
					TTL: &ttl,
				},
			},
			want: true,
		},
		"DifferentFields": {
			args: args{
				rrSet: route53types.ResourceRecordSet{
					Name: &resourceRecordSetName,
					TTL:  &ttl,
				},
				p: v1alpha1.ResourceRecordSetParameters{
					TTL: &ttl2,
				},
			},
			want: false,
		},
		"IgnoresRefs": {
			args: args{
				rrSet: route53types.ResourceRecordSet{
					Name: &resourceRecordSetName,
					TTL:  &ttl,
				},
				p: v1alpha1.ResourceRecordSetParameters{
					TTL: &ttl,
				},
			},
			want: true,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got, _ := IsUpToDate(tc.args.p, tc.args.rrSet)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}

}
