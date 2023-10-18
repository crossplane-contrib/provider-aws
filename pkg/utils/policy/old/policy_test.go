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

package legacypolicy

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

const (
	policy = `{"Statement":[{"Action":"ecr:ListImages","Effect":"Allow","Principal":"*"}],"Version":"2012-10-17"}`

	cpxPolicy = `{"Statement":[{"Action":"ecr:ListImages","Effect":"Allow","Principal":{"AWS":["arn:aws:iam::111122223333:userARN","111122223334","arn:aws:iam::111122223333:roleARN"]}}],"Version":"2012-10-17"}`
	// Note: different sort order of principals than input above
	cpxRemPolicy = `{"Statement":[{"Action":"ecr:ListImages","Effect":"Allow","Principal":{"AWS":["111122223334","arn:aws:iam::111122223333:userARN","arn:aws:iam::111122223333:roleARN"]}}],"Version":"2012-10-17"}`
)

func TestIsPolicyUpToDate(t *testing.T) {
	type args struct {
		local  string
		remote string
	}

	cases := map[string]struct {
		args args
		want bool
	}{
		"SameFields": {
			args: args{
				local:  "{\"testone\": \"one\", \"testtwo\": \"two\"}",
				remote: "{\"testtwo\": \"two\", \"testone\": \"one\"}",
			},
			want: true,
		},
		"SameFieldsRealPolicy": {
			args: args{
				local:  policy,
				remote: `{"Statement":[{"Effect":"Allow","Action":"ecr:ListImages","Principal":"*"}],"Version":"2012-10-17"}`,
			},
			want: true,
		},
		"DifferentFields": {
			args: args{
				local:  "{\"testone\": \"one\", \"testtwo\": \"two\"}",
				remote: "{\"testthree\": \"three\", \"testone\": \"one\"}",
			},
			want: false,
		},
		"SameFieldsPrincipalPolicy": {
			args: args{
				local:  cpxPolicy,
				remote: cpxRemPolicy,
			},
			want: true,
		},
		"SameFieldsNumericPrincipals": {
			args: args{
				// This is to test that our slice sorting does not
				// panic with unexpected value types.
				local:  `{"Statement":[{"Effect":"Allow","Action":"ecr:ListImages","Principal":[2,1,"foo","bar"]}],"Version":"2012-10-17"}`,
				remote: `{"Statement":[{"Effect":"Allow","Action":"ecr:ListImages","Principal":[2,1,"bar","foo"]}],"Version":"2012-10-17"}`,
			},
			want: true,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := IsPolicyUpToDate(&tc.args.local, &tc.args.remote)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}
