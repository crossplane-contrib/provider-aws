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

package v1beta1

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"

	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/test"
)

func TestSubnetIDReferencerForDBSubnetGroup(t *testing.T) {
	value := "cool"

	type args struct {
		res   resource.CanReference
		value string
	}
	type want struct {
		res resource.CanReference
		err error
	}

	cases := map[string]struct {
		reason string
		args   args
		want   want
	}{
		"AssignWrongType": {
			reason: "Assign should return an error when the supplied CanReference does not contain an *DBSubnetGroup.",
			args: args{
				res: nil,
			},
			want: want{
				err: errors.New(errResourceIsNotDBSubnetGroup),
			},
		},
		"AssignSuccessful": {
			reason: "Assign should append to Spec.SubnetIDs.",
			args: args{
				res:   &DBSubnetGroup{},
				value: value,
			},
			want: want{
				res: &DBSubnetGroup{
					Spec: DBSubnetGroupSpec{
						ForProvider: DBSubnetGroupParameters{SubnetIDs: []string{value}},
					},
				},
			},
		},
		"AssignNoOp": {
			reason: "Assign should not append existing values to Spec.SubnetIDs.",
			args: args{
				res: &DBSubnetGroup{
					Spec: DBSubnetGroupSpec{
						ForProvider: DBSubnetGroupParameters{SubnetIDs: []string{value}},
					},
				},
				value: value,
			},
			want: want{
				res: &DBSubnetGroup{
					Spec: DBSubnetGroupSpec{
						ForProvider: DBSubnetGroupParameters{SubnetIDs: []string{value}},
					},
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			r := &SubnetIDReferencerForDBSubnetGroup{}
			err := r.Assign(tc.args.res, tc.args.value)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("\nReason: %s\nAssign(...): -want error, +got error:\n%s", tc.reason, diff)
			}
			if diff := cmp.Diff(tc.want.res, tc.args.res, test.EquateErrors()); diff != "" {
				t.Errorf("\nReason: %s\nAssign(...): -want, +got:\n%s", tc.reason, diff)
			}
		})
	}
}
