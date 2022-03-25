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

package restapi

import (
	"context"
	"testing"

	svcsdk "github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"

	"github.com/crossplane/provider-aws/apis/apigateway/v1alpha1"
	aws "github.com/crossplane/provider-aws/pkg/clients"
	"github.com/crossplane/provider-aws/pkg/clients/apigateway/fake"
)

var (
	submittedPolSimple = `{
		"Version": "2012-10-17",
		"Statement": [
			{
				"Effect": "Allow",
				"Principal": "*",
				"Action": "execute-api:Invoke",
				"Resource": "execute-api:/*/*/*"
			}
		]
	}`
	submittedPolComplex = `{
		"Effect": "Deny",
		"Principal": "*",
		"Action": "execute-api:Invoke",
		"Resource": "execute-api:/*/*/*",
		"Condition": {
				"StringNotEquals": {
						"aws:sourceVpc": "vpc-1234ab123456789ab"
				}
		}
	},
	{
		"Version": "2012-10-17",
		"Statement": [
			{
				"Effect": "Allow",
				"Principal": "*",
				"Action": "execute-api:Invoke",
				"Resource": "execute-api:/*/*/*"
			}
		]
	}`
	expectedPolSimple  = `{\"Version\":\"2012-10-17\",\"Statement\":[{\"Effect\":\"Allow\",\"Principal\":\"*\",\"Action\":\"execute-api:Invoke\",\"Resource\":\"arn:aws:execute-api:eu-central-1:123456789012:abcdef1234\/*\/*\/*\"}]}`
	expectedPolComplex = `{\"Version\":\"2012-10-17\",\"Statement\":[{\"Effect\":\"Deny\",\"Principal\":\"*\",\"Action\":\"execute-api:Invoke\",\"Resource\":\"arn:aws:execute-api:eu-central-1:272371606098:tgcv8l3668\/*\/*\/*\",\"Condition\":{\"StringNotEquals\":{\"aws:sourceVpc\":\"vpc-0929e10bf326f116a\"}}},{\"Effect\":\"Allow\",\"Principal\":\"*\",\"Action\":\"execute-api:Invoke\",\"Resource\":\"arn:aws:execute-api:eu-central-1:272371606098:tgcv8l3668\/*\/*\/*\"}]}`
)

type args struct {
	cr  *v1alpha1.RestAPI
	obj *svcsdk.RestApi
}

type apiModifier func(*v1alpha1.RestAPI)

func withSpec(p v1alpha1.RestAPIParameters) apiModifier {
	return func(r *v1alpha1.RestAPI) { r.Spec.ForProvider = p }
}

func withExternalName(n string) apiModifier {
	return func(r *v1alpha1.RestAPI) { meta.SetExternalName(r, n) }
}

func restAPI(m ...apiModifier) *v1alpha1.RestAPI {
	cr := &v1alpha1.RestAPI{}
	cr.Name = "test-api-name"
	for _, f := range m {
		f(cr)
	}
	return cr
}

var _ managed.ExternalClient = &external{}
var _ managed.ExternalConnecter = &connector{}

func TestPreUpdate(t *testing.T) {
	type want struct {
		err error
		ops []*svcsdk.PatchOperation
	}

	cases := map[string]struct {
		args
		want
	}{
		"DiffNameUpdate": {
			args: args{
				cr: restAPI([]apiModifier{
					withSpec(v1alpha1.RestAPIParameters{
						Name: aws.String("test-b"),
					}),
					withExternalName("1234567"),
				}...),
				obj: nil,
			},
			want: want{
				ops: []*svcsdk.PatchOperation{
					{
						Op:    aws.String("replace"),
						Path:  aws.String("/name"),
						Value: aws.String("test-b"),
					},
				},
				err: nil,
			},
		},
		"DiffPolUpdate": {
			args: args{
				cr: restAPI([]apiModifier{
					withSpec(v1alpha1.RestAPIParameters{
						Name: aws.String("test-b"),
					}),
					withExternalName("1234567"),
				}...),
				obj: nil,
			},
			want: want{
				ops: []*svcsdk.PatchOperation{
					{
						Op:    aws.String("replace"),
						Path:  aws.String("/name"),
						Value: aws.String("test-b"),
					},
				},
				err: nil,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			c := &custom{
				Client: &fake.MockAPIGatewayClient{
					MockGetRestAPIByID: func(_ context.Context, apiId *string) (*svcsdk.RestApi, error) {
						return &svcsdk.RestApi{
							Name: aws.String("test-a"),
							Id:   apiId,
						}, nil
					},
				},
			}

			// Act
			in := GenerateUpdateRestApiInput(nil)
			err := c.preUpdate(context.TODO(), tc.args.cr, in)
			if err != nil {
				panic(err)
			}

			// Assert
			if diff := cmp.Diff(tc.want.ops, in.PatchOperations); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestLateInitialize(t *testing.T) {
	type want struct {
		result v1alpha1.RestAPIParameters
		err    error
	}

	fixedPolResult, _ := normalizePolicy(aws.String(`{"Statement":[{"Effect":"Allow","Principal":"*","Action":"execute-api:Invoke","Resource":"arn:aws:execute-api:eu-central-1:123456789012:abcdef1234\/*\/*\/*"}],"Version":"2012-10-17"}`))
	normSubmittedPolSimple, _ := normalizePolicy(aws.String(submittedPolSimple))
	cases := map[string]struct {
		args
		want
	}{
		"PolicyFixedByAws": {
			args: args{
				cr: restAPI(withSpec(v1alpha1.RestAPIParameters{
					Policy: &submittedPolSimple,
				})),
				obj: &svcsdk.RestApi{
					Policy: unescapePolicy(aws.String(expectedPolSimple)),
				},
			},
			want: want{
				result: v1alpha1.RestAPIParameters{
					Policy: fixedPolResult,
				},
				err: nil,
			},
		},
		"DiffPolicy": {
			args: args{
				cr: restAPI(withSpec(v1alpha1.RestAPIParameters{
					Policy: &submittedPolSimple,
				})),
				obj: &svcsdk.RestApi{
					Policy: unescapePolicy(aws.String(expectedPolComplex)),
				},
			},
			want: want{
				result: v1alpha1.RestAPIParameters{
					Policy: normSubmittedPolSimple,
				},
				err: nil,
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			// Act
			err := lateInitialize(&tc.args.cr.Spec.ForProvider, tc.args.obj)
			if err != nil {
				panic(err)
			}

			// Assert
			if diff := cmp.Diff(tc.want.result, tc.args.cr.Spec.ForProvider); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestIsUpToDate(t *testing.T) {
	type want struct {
		result bool
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"PolDiffUpdate": {
			args: args{
				cr: restAPI(withSpec(v1alpha1.RestAPIParameters{
					Policy: &submittedPolSimple,
				})),
				obj: &svcsdk.RestApi{
					Policy: &expectedPolSimple,
				},
			},
			want: want{
				result: false,
				err:    nil,
			},
		},
		"PolicyAddStmtUpdate": {
			args: args{
				cr: restAPI(withSpec(v1alpha1.RestAPIParameters{
					Policy: &submittedPolSimple,
				})),
				obj: &svcsdk.RestApi{
					Policy: &expectedPolComplex,
				},
			},
			want: want{
				result: false,
				err:    nil,
			},
		},
		"PolicyRemStmtUpdate": {
			args: args{
				cr: restAPI(withSpec(v1alpha1.RestAPIParameters{
					Policy: &submittedPolComplex,
				})),
				obj: &svcsdk.RestApi{
					Policy: &expectedPolSimple,
				},
			},
			want: want{
				result: false,
				err:    nil,
			},
		},
		"DiffNameUpdate": {
			args: args{
				cr: restAPI(withSpec(v1alpha1.RestAPIParameters{
					Name: aws.String("test-a"),
				})),
				obj: &svcsdk.RestApi{
					Name: aws.String("test-b"),
				},
			},
			want: want{
				result: false,
				err:    nil,
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			// Act
			result, err := isUpToDate(tc.args.cr, tc.args.obj)
			if err != nil {
				panic(err)
			}

			// Assert
			if diff := cmp.Diff(tc.want.result, result, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}
