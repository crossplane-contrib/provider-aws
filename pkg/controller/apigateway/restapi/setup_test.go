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

	"github.com/aws/aws-sdk-go/aws"
	svcsdk "github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"

	"github.com/crossplane-contrib/provider-aws/apis/apigateway/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/apigateway/fake"
)

type policyExample struct {
	FromSpec *string
	FromAws  *string
	Result   *string
}

var (
	polNoDetails = policyExample{
		FromSpec: aws.String(`{
			"Version": "2012-10-17",
			"Statement": [
				{
					"Effect": "Allow",
					"Principal": "*",
					"Action": "execute-api:Invoke",
					"Resource": "execute-api:/*/*/*"
				}
			]
		}`),
		FromAws: aws.String(`{\"Version\":\"2012-10-17\",\"Statement\":[{\"Action\":\"execute-api:Invoke\",\"Effect\":\"Allow\",\"Principal\":\"*\",\"Resource\":\"arn:aws:execute-api:eu-central-1:123456789012:abcdef1234/*/*/*\"}]}`),
		Result:  aws.String(`{"Statement":[{"Action":"execute-api:Invoke","Effect":"Allow","Principal":"*","Resource":"arn:aws:execute-api:eu-central-1:123456789012:abcdef1234/*/*/*"}],"Version":"2012-10-17"}`),
	}

	polWithDetails = policyExample{
		FromSpec: aws.String(`{
			"Version": "2012-10-17",
			"Statement": [
				{
					"Effect": "Allow",
					"Principal": "*",
					"Action": "execute-api:Invoke",
					"Resource": "arn:aws:execute-api:eu-central-1:123456789012:abcdef1234/*/*/*"
				}
			]
		}`),
		FromAws: polNoDetails.FromAws,
		Result:  polNoDetails.Result,
	}

	polMultipleStmts = policyExample{
		FromSpec: aws.String(`{
			"Version": "2012-10-17",
			"Statement": [
				{
					"Effect": "Deny",
					"Principal": "*",
					"Action": "execute-api:Invoke",
					"Resource": "arn:aws:execute-api:eu-central-1:123456789012:abcdef1234/*/*/*",
					"Condition": {
							"StringNotEquals": {
									"aws:sourceVpc": "vpc-1234ab123456789ab"
							}
					}
				},
				{
					"Effect": "Allow",
					"Principal": "*",
					"Action": "execute-api:Invoke",
					"Resource": "arn:aws:execute-api:eu-central-1:123456789012:abcdef1234/*/*/*"
				}
			]
		}`),
		FromAws: aws.String(`{\"Version\":\"2012-10-17\",\"Statement\":[{\"Action\":\"execute-api:Invoke\",\"Effect\":\"Deny\",\"Principal\":\"*\",\"Resource\":\"arn:aws:execute-api:eu-central-1:123456789012:abcdef1234/*/*/*\",\"Condition\":{\"StringNotEquals\":{\"aws:sourceVpc\":\"vpc-1234ab123456789ab\"}}},{\"Action\":\"execute-api:Invoke\",\"Effect\":\"Allow\",\"Principal\":\"*\",\"Resource\":\"arn:aws:execute-api:eu-central-1:123456789012:abcdef1234/*/*/*\"}]}`),
		Result:  aws.String(`{"Statement":[{"Action":"execute-api:Invoke","Condition":{"StringNotEquals":{"aws:sourceVpc":"vpc-1234ab123456789ab"}},"Effect":"Deny","Principal":"*","Resource":"arn:aws:execute-api:eu-central-1:123456789012:abcdef1234/*/*/*"},{"Action":"execute-api:Invoke","Effect":"Allow","Principal":"*","Resource":"arn:aws:execute-api:eu-central-1:123456789012:abcdef1234/*/*/*"}],"Version":"2012-10-17"}`),
	}
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

var _ managed.TypedExternalClient[*v1alpha1.RestAPI] = &external{}
var _ managed.TypedExternalConnecter[*v1alpha1.RestAPI] = &connector{}

func TestPreUpdate(t *testing.T) {
	type want struct {
		err error
		ops []*svcsdk.PatchOperation
	}

	cases := map[string]struct {
		args
		want
	}{
		"EqualsNoUpdate": {
			args: args{
				cr: restAPI([]apiModifier{
					withSpec(v1alpha1.RestAPIParameters{
						Name:   aws.String("test-a"),
						Policy: polNoDetails.FromSpec,
					}),
					withExternalName("1234567"),
				}...),
				obj: nil,
			},
			want: want{
				ops: []*svcsdk.PatchOperation{},
				err: nil,
			},
		},
		"NameUpdate": {
			args: args{
				cr: restAPI([]apiModifier{
					withSpec(v1alpha1.RestAPIParameters{
						Name:   aws.String("test-b"),
						Policy: polNoDetails.FromSpec,
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
						Name:   aws.String("test-a"),
						Policy: polMultipleStmts.FromSpec,
					}),
					withExternalName("1234567"),
				}...),
				obj: nil,
			},
			want: want{
				ops: []*svcsdk.PatchOperation{
					{
						Op:    aws.String("replace"),
						Path:  aws.String("/policy"),
						Value: polMultipleStmts.Result,
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
							Name:   aws.String("test-a"),
							Id:     apiId,
							Policy: polNoDetails.FromAws,
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

	normalized, _ := normalizePolicy(polNoDetails.FromSpec)
	cases := map[string]struct {
		args
		want
	}{
		"PolicyFixedByAws": {
			args: args{
				cr: restAPI(withSpec(v1alpha1.RestAPIParameters{
					Policy: polNoDetails.FromSpec,
				})),
				obj: &svcsdk.RestApi{
					Policy: polNoDetails.FromAws,
				},
			},
			want: want{
				result: v1alpha1.RestAPIParameters{
					Policy: polNoDetails.Result,
				},
				err: nil,
			},
		},
		"SamePolicy": {
			args: args{
				cr: restAPI(withSpec(v1alpha1.RestAPIParameters{
					Policy: polWithDetails.FromSpec,
				})),
				obj: &svcsdk.RestApi{
					Policy: polWithDetails.FromAws,
				},
			},
			want: want{
				result: v1alpha1.RestAPIParameters{
					Policy: polWithDetails.Result,
				},
				err: nil,
			},
		},
		"DiffPolicy": {
			args: args{
				cr: restAPI(withSpec(v1alpha1.RestAPIParameters{
					Policy: polNoDetails.FromSpec,
				})),
				obj: &svcsdk.RestApi{
					Policy: polMultipleStmts.FromAws,
				},
			},
			want: want{
				result: v1alpha1.RestAPIParameters{
					Policy: normalized,
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
					Policy: polNoDetails.FromSpec,
				})),
				obj: &svcsdk.RestApi{
					Policy: polNoDetails.FromAws,
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
					Policy: polMultipleStmts.FromSpec,
				})),
				obj: &svcsdk.RestApi{
					Policy: polNoDetails.FromAws,
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
					Policy: polNoDetails.FromSpec,
				})),
				obj: &svcsdk.RestApi{
					Policy: polMultipleStmts.FromAws,
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
					Name:   aws.String("test-a"),
					Policy: polNoDetails.FromSpec,
				})),
				obj: &svcsdk.RestApi{
					Name:   aws.String("test-b"),
					Policy: polNoDetails.FromAws,
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
			result, _, err := isUpToDate(context.Background(), tc.args.cr, tc.args.obj)
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
