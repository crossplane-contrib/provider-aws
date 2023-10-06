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

package thing

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go/service/iot"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"

	"github.com/crossplane-contrib/provider-aws/apis/iot/v1alpha1"
)

var (
	testThingName     = "example-thing"
	testThingTypeName = "example-type-name"
)

var testThingOutput = iot.DescribeThingOutput{
	ThingName: &testThingName,
}

type ThingModifier func(thing *v1alpha1.Thing)

func withThingName(val string) ThingModifier {
	return func(r *v1alpha1.Thing) { r.ObjectMeta.Name = val }
}

func withThingTypeName(val *string) ThingModifier {
	return func(r *v1alpha1.Thing) { r.Spec.ForProvider.ThingTypeName = val }
}

func thing(m ...ThingModifier) *v1alpha1.Thing {
	cr := &v1alpha1.Thing{}
	for _, f := range m {
		f(cr)
	}
	return cr
}

func TestIsUpToDate(t *testing.T) {
	type args struct {
		cr   *v1alpha1.Thing
		resp iot.DescribeThingOutput
	}

	type want struct {
		isUpToDate bool
		err        error
	}

	cases := map[string]struct {
		args
		want
	}{
		"NotUpToDate": {
			args: args{
				cr:   thing(withThingName(testThingName), withThingTypeName(&testThingTypeName)),
				resp: testThingOutput,
			},
			want: want{
				isUpToDate: false,
				err:        nil,
			},
		},
		"UpToDate": {
			args: args{
				cr:   thing(withThingName(testThingName)),
				resp: testThingOutput,
			},
			want: want{
				isUpToDate: true,
				err:        nil,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			o, _, err := isUpToDate(context.Background(), tc.args.cr, &tc.args.resp)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.isUpToDate, o, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}
