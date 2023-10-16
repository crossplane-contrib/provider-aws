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

package arn

import (
	"testing"

	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"
)

func TestParseARN(t *testing.T) {
	type args struct {
		arn string
	}
	type want struct {
		arn ARN
		err error
	}
	cases := map[string]struct {
		args args
		want want
	}{
		"WithWildcard": {
			args: args{
				arn: "arn:aws:iam::123456789012:user/Development/product_1234/*",
			},
			want: want{
				arn: ARN{
					Partition: "aws",
					Service:   "iam",
					Region:    "",
					AccountID: "123456789012",
					Resource:  "user/Development/product_1234/*",
				},
			},
		},
		"WithDots": {
			args: args{
				arn: "arn:aws:iam::123456789012:role/aws-service-role/access-analyzer.amazonaws.com/AWSServiceRoleForAccessAnalyzer",
			},
			want: want{
				arn: ARN{
					Partition: "aws",
					Service:   "iam",
					Region:    "",
					AccountID: "123456789012",
					Resource:  "role/aws-service-role/access-analyzer.amazonaws.com/AWSServiceRoleForAccessAnalyzer",
				},
			},
		},
		"WithSpaces": {
			args: args{
				arn: "arn:aws:iam::123456789012:u2f/user/JohnDoe/default (U2F security key)",
			},
			want: want{
				arn: ARN{
					Partition: "aws",
					Service:   "iam",
					Region:    "",
					AccountID: "123456789012",
					Resource:  "u2f/user/JohnDoe/default (U2F security key)",
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			arn, err := ParseARN(tc.args.arn)
			if diff := cmp.Diff(tc.want.arn, arn); diff != "" {
				t.Errorf("Wrap: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("Wrap: -want, +got:\n%s", diff)
			}
		})
	}
}
