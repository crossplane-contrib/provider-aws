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

package resource

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	svcsdk "github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"

	"github.com/crossplane-contrib/provider-aws/apis/apigateway/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/apigateway/fake"
)

var (
	defaultAPI    = aws.String("1234567")
	defaultParent = aws.String("1234567")
)

type apiModifier func(*v1alpha1.Resource)

func withSpec(p v1alpha1.ResourceParameters) apiModifier {
	return func(r *v1alpha1.Resource) {
		p.RestAPIID = defaultAPI
		r.Spec.ForProvider = p
	}
}

func withDefaultParentID() apiModifier {
	return func(r *v1alpha1.Resource) { r.Spec.ForProvider.ParentResourceID = defaultParent }
}
func withExternalName(n string) apiModifier {
	return func(r *v1alpha1.Resource) { meta.SetExternalName(r, n) }
}

func Resource(m ...apiModifier) *v1alpha1.Resource {
	cr := &v1alpha1.Resource{}
	cr.Name = "test-api-name"
	for _, f := range m {
		f(cr)
	}
	return cr
}

var _ managed.TypedExternalClient[*v1alpha1.Resource] = &external{}
var _ managed.TypedExternalConnecter[*v1alpha1.Resource] = &connector{}

func TestPreCreate(t *testing.T) {
	type want struct {
		err error
		obj *svcsdk.CreateResourceInput
	}
	type args struct {
		cr  *v1alpha1.Resource
		obj *svcsdk.Resource
	}

	cases := map[string]struct {
		args
		want
	}{
		"NoParentId": {
			args: args{
				cr: Resource([]apiModifier{
					withSpec(v1alpha1.ResourceParameters{}),
				}...),
				obj: nil,
			},
			want: want{
				obj: &svcsdk.CreateResourceInput{
					RestApiId: defaultAPI,
					ParentId:  defaultParent,
				},
				err: nil,
			},
		},
		"WithParentId": {
			args: args{
				cr: Resource([]apiModifier{
					withSpec(v1alpha1.ResourceParameters{}), withDefaultParentID(),
				}...),
				obj: nil,
			},
			want: want{
				obj: &svcsdk.CreateResourceInput{
					RestApiId: defaultAPI,
					ParentId:  defaultParent,
				},
				err: nil,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			c := &custom{
				Client: &fake.MockAPIGatewayClient{
					MockGetRestAPIRootResource: func(_ context.Context, _ *string) (*string, error) {
						return defaultParent, nil
					},
				},
			}

			// Act
			in := GenerateCreateResourceInput(Resource(withSpec(v1alpha1.ResourceParameters{})))
			err := c.preCreate(context.TODO(), tc.args.cr, in)
			if err != nil {
				panic(err)
			}

			// Assert
			if diff := cmp.Diff(tc.want.obj, in); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestPreUpdate(t *testing.T) {
	type want struct {
		ops []*svcsdk.PatchOperation
		err error
	}

	type args struct {
		cr     *v1alpha1.Resource
		mockFn func(context.Context, *svcsdk.GetResourceInput, ...request.Option) (*svcsdk.Resource, error)
	}

	cases := map[string]struct {
		args
		want
	}{
		"DiffParentResource": {
			args: args{
				cr: Resource(withSpec(v1alpha1.ResourceParameters{}), withExternalName("1234"), withDefaultParentID()),
				mockFn: func(_ context.Context, in *svcsdk.GetResourceInput, opts ...request.Option) (*svcsdk.Resource, error) {
					return &svcsdk.Resource{
						Id:       in.RestApiId,
						ParentId: aws.String("1234"),
					}, nil
				},
			},
			want: want{
				ops: []*svcsdk.PatchOperation{
					{
						Op:    aws.String("replace"),
						Path:  aws.String("/parentResourceId"),
						Value: defaultParent,
					},
				},
				err: nil,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			// Act
			c := &custom{
				Client: &fake.MockAPIGatewayClient{
					MockGetResource: tc.args.mockFn,
				},
			}
			in := GenerateUpdateResourceInput(nil)

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

func TestIsUpToDate(t *testing.T) {
	type want struct {
		result bool
		err    error
	}
	type args struct {
		cr  *v1alpha1.Resource
		obj *svcsdk.Resource
	}

	cases := map[string]struct {
		args
		want
	}{
		"SameDataNoUpdate": {
			args: args{
				cr: Resource(withSpec(v1alpha1.ResourceParameters{}), withDefaultParentID(), withExternalName("1234")),
				obj: &svcsdk.Resource{
					Id:       defaultAPI,
					ParentId: defaultParent,
				},
			},
			want: want{
				result: true,
				err:    nil,
			},
		},
		"DifferentPathPartUpdate": {
			args: args{
				cr: Resource(withSpec(v1alpha1.ResourceParameters{
					PathPart: aws.String("/a"),
				}), withDefaultParentID(), withExternalName("1234")),
				obj: &svcsdk.Resource{
					Id:       defaultAPI,
					ParentId: defaultParent,
					PathPart: aws.String("/b"),
				},
			},
			want: want{
				result: true,
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
