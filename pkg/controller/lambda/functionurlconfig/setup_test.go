package functionurlconfig

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	svcsdk "github.com/aws/aws-sdk-go/service/lambda"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"

	"github.com/crossplane-contrib/provider-aws/apis/lambda/v1alpha1"
)

type args struct {
	cr  *v1alpha1.FunctionURLConfig
	obj *svcsdk.GetFunctionUrlConfigOutput
}

type functionModifier func(*v1alpha1.FunctionURLConfig)

func withSpec(p *v1alpha1.CORS) functionModifier {
	return func(r *v1alpha1.FunctionURLConfig) { r.Spec.ForProvider.CORS = p }
}

func function(m ...functionModifier) *v1alpha1.FunctionURLConfig {
	cr := &v1alpha1.FunctionURLConfig{}
	cr.Name = "test-function-config-url-name"
	for _, f := range m {
		f(cr)
	}
	return cr
}

func TestIsUpToDateCors(t *testing.T) {
	type want struct {
		result bool
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"NilSourceNoUpdate": {
			args: args{
				cr:  function(),
				obj: &svcsdk.GetFunctionUrlConfigOutput{Cors: &svcsdk.Cors{}},
			},
			want: want{
				result: true,
				err:    nil,
			},
		},
		"NilSourceNilAwsNoUpdate": {
			args: args{
				cr:  function(),
				obj: &svcsdk.GetFunctionUrlConfigOutput{},
			},
			want: want{
				result: true,
				err:    nil,
			},
		},
		"EmptySourceUpdate": {
			args: args{
				cr:  function(withSpec(&v1alpha1.CORS{})),
				obj: &svcsdk.GetFunctionUrlConfigOutput{Cors: &svcsdk.Cors{AllowCredentials: aws.Bool(true)}},
			},
			want: want{
				result: false,
				err:    nil,
			},
		},
		"NilAwsWithUpdate": {
			args: args{
				cr:  function(withSpec(&v1alpha1.CORS{})),
				obj: &svcsdk.GetFunctionUrlConfigOutput{},
			},
			want: want{
				result: false,
				err:    nil,
			},
		},
		"NeedsUpdate": {
			args: args{
				cr: function(withSpec(&v1alpha1.CORS{
					AllowHeaders: []*string{
						aws.String("X-Custom-3"),
						aws.String("X-Custom-4"),
					},
				})),
				obj: &svcsdk.GetFunctionUrlConfigOutput{
					Cors: &svcsdk.Cors{
						AllowHeaders: []*string{
							aws.String("X-Custom-1"),
							aws.String("X-Custom-2"),
						},
						AllowMethods: []*string{
							aws.String("GET"),
							aws.String("POST"),
						},
						AllowOrigins: []*string{
							aws.String("http://localhost:8081"),
							aws.String("http://localhost:8082"),
						},
						ExposeHeaders: []*string{
							aws.String("X-Custom-10"),
							aws.String("X-Custom-20"),
						},
					},
				},
			},
			want: want{
				result: false,
				err:    nil,
			},
		},
		"NoUpdateNeeded": {
			args: args{
				cr: function(withSpec(&v1alpha1.CORS{
					AllowHeaders: []*string{
						aws.String("X-Custom-1"),
						aws.String("X-Custom-2"),
					},
					AllowMethods: []*string{
						aws.String("GET"),
						aws.String("POST"),
					},
					AllowOrigins: []*string{
						aws.String("http://localhost:8081"),
						aws.String("http://localhost:8082"),
					},
					ExposeHeaders: []*string{
						aws.String("X-Custom-10"),
						aws.String("X-Custom-20"),
					},
				})),
				obj: &svcsdk.GetFunctionUrlConfigOutput{
					Cors: &svcsdk.Cors{
						AllowHeaders: []*string{
							aws.String("X-Custom-1"),
							aws.String("X-Custom-2"),
						},
						AllowMethods: []*string{
							aws.String("GET"),
							aws.String("POST"),
						},
						AllowOrigins: []*string{
							aws.String("http://localhost:8081"),
							aws.String("http://localhost:8082"),
						},
						ExposeHeaders: []*string{
							aws.String("X-Custom-10"),
							aws.String("X-Custom-20"),
						},
					},
				},
			},
			want: want{
				result: true,
				err:    nil,
			},
		},
		"NoUpdateNeededOutOfOrder": {
			args: args{
				cr: function(withSpec(&v1alpha1.CORS{
					AllowHeaders: []*string{
						aws.String("X-Custom-2"),
						aws.String("X-Custom-1"),
					},
					AllowMethods: []*string{
						aws.String("POST"),
						aws.String("GET"),
					},
					AllowOrigins: []*string{
						aws.String("http://localhost:8082"),
						aws.String("http://localhost:8081"),
					},
					ExposeHeaders: []*string{
						aws.String("X-Custom-10"),
						aws.String("X-Custom-20"),
					},
				})),
				obj: &svcsdk.GetFunctionUrlConfigOutput{
					Cors: &svcsdk.Cors{
						AllowHeaders: []*string{
							aws.String("X-Custom-1"),
							aws.String("X-Custom-2"),
						},
						AllowMethods: []*string{
							aws.String("GET"),
							aws.String("POST"),
						},
						AllowOrigins: []*string{
							aws.String("http://localhost:8081"),
							aws.String("http://localhost:8082"),
						},
						ExposeHeaders: []*string{
							aws.String("X-Custom-20"),
							aws.String("X-Custom-10"),
						},
					},
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
			result := isUpToDateCors(tc.args.cr, tc.args.obj)

			// Assert
			if diff := cmp.Diff(tc.want.result, result, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}
