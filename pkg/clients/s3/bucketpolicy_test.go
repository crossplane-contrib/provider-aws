/*
Copyright 2020 The Crossplane Authors.

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

package s3

import (
	"encoding/json"
	"testing"

	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"

	"github.com/crossplane-contrib/provider-aws/apis/s3/common"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
)

var (
	// an arbitrary managed resource
	effect      = "Allow"
	statementID = pointer.ToOrNilIfZeroValue("1")
)

type statementModifier func(statement *common.BucketPolicyStatement)

func withPrincipal(s *common.BucketPrincipal) statementModifier {
	return func(statement *common.BucketPolicyStatement) {
		statement.Principal = s
	}
}

func withPolicyAction(s []string) statementModifier {
	return func(statement *common.BucketPolicyStatement) {
		statement.Action = s
	}
}

func withResourcePath(s []string) statementModifier {
	return func(statement *common.BucketPolicyStatement) {
		statement.Resource = s
	}
}

func withConditionBlock(m []common.Condition) statementModifier {
	return func(statement *common.BucketPolicyStatement) {
		statement.Condition = m
	}
}

func policyStatement(m ...statementModifier) *common.BucketPolicyStatement {
	cr := &common.BucketPolicyStatement{
		SID:    statementID,
		Effect: effect,
	}
	for _, f := range m {
		f(cr)
	}
	return cr
}

func TestSerializeBucketPolicyStatement(t *testing.T) {
	cases := map[string]struct {
		in  common.BucketPolicyStatement
		out string
		err error
	}{
		"BasicInput": {
			in:  *policyStatement(),
			out: `{"Effect":"Allow","Sid":"1"}`,
		},
		"ValidInput": {
			in: *policyStatement(
				withPrincipal(&common.BucketPrincipal{
					AllowAnon: true,
				}),
				withPolicyAction([]string{"s3:ListBucket"}),
				withResourcePath([]string{"arn:aws:s3:::test.s3.crossplane.com"}),
			),
			out: `{"Action":"s3:ListBucket","Effect":"Allow","Principal":"*","Resource":"arn:aws:s3:::test.s3.crossplane.com","Sid":"1"}`,
		},
		"ComplexInput": {
			in: *policyStatement(
				withPrincipal(&common.BucketPrincipal{
					AWSPrincipals: []common.AWSPrincipal{
						{
							UserARN: pointer.ToOrNilIfZeroValue("arn:aws:iam::111122223333:userARN"),
						},
						{
							AWSAccountID: pointer.ToOrNilIfZeroValue("111122223333"),
						},
						{
							IAMRoleARN: pointer.ToOrNilIfZeroValue("arn:aws:iam::111122223333:roleARN"),
						},
					},
				}),
				withPolicyAction([]string{"s3:ListBucket"}),
				withResourcePath([]string{"arn:aws:s3:::test.s3.crossplane.com"}),
				withConditionBlock([]common.Condition{
					{
						OperatorKey: "test",
						Conditions: []common.ConditionPair{
							{
								ConditionKey:         "test",
								ConditionStringValue: pointer.ToOrNilIfZeroValue("testKey"),
							},
							{
								ConditionKey:         "test2",
								ConditionStringValue: pointer.ToOrNilIfZeroValue("testKey2"),
							},
						},
					},
				}),
			),
			out: `{"Condition":{"test":{"test":"testKey","test2":"testKey2"}},"Action":"s3:ListBucket","Effect":"Allow","Principal":{"AWS":["arn:aws:iam::111122223333:userARN","111122223333","arn:aws:iam::111122223333:roleARN"]},"Resource":"arn:aws:s3:::test.s3.crossplane.com","Sid":"1"}`,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			js, err := SerializeBucketPolicyStatement(tc.in)

			if diff := cmp.Diff(tc.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
				return
			}

			var out interface{}
			err = json.Unmarshal([]byte(tc.out), &out)
			if diff := cmp.Diff(tc.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
				return
			}

			if diff := cmp.Diff(js, out); diff != "" {
				t.Errorf("SerializeBucketPolicyStatement(...): -want, +got\n:%s", diff)
			}
		})
	}
}
