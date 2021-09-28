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

package bucket

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"

	"github.com/crossplane/provider-aws/apis/s3/v1beta1"
	awsclient "github.com/crossplane/provider-aws/pkg/clients"
	clientss3 "github.com/crossplane/provider-aws/pkg/clients/s3"
	"github.com/crossplane/provider-aws/pkg/clients/s3/fake"
	s3Testing "github.com/crossplane/provider-aws/pkg/controller/s3/testing"
)

var _ SubresourceClient = &CORSConfigurationClient{}

func generateCORSConfig() *v1beta1.CORSConfiguration {
	return &v1beta1.CORSConfiguration{CORSRules: []v1beta1.CORSRule{
		{
			AllowedHeaders: []string{"test.header"},
			AllowedMethods: []string{"GET"},
			AllowedOrigins: []string{"test.origin"},
			ExposeHeaders:  []string{"test.expose"},
			MaxAgeSeconds:  awsclient.Int32(10),
		},
	},
	}
}

func generateAWSCORS() *s3types.CORSConfiguration {
	return &s3types.CORSConfiguration{CORSRules: []s3types.CORSRule{
		{
			AllowedHeaders: []string{"test.header"},
			AllowedMethods: []string{"GET"},
			AllowedOrigins: []string{"test.origin"},
			ExposeHeaders:  []string{"test.expose"},
			MaxAgeSeconds:  10,
		},
	},
	}
}

func TestCORSObserve(t *testing.T) {
	type args struct {
		cl *CORSConfigurationClient
		b  *v1beta1.Bucket
	}

	type want struct {
		status ResourceStatus
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"Error": {
			args: args{
				b: s3Testing.Bucket(s3Testing.WithCORSConfig(generateCORSConfig())),
				cl: NewCORSConfigurationClient(fake.MockBucketClient{
					MockGetBucketCors: func(ctx context.Context, input *s3.GetBucketCorsInput, opts []func(*s3.Options)) (*s3.GetBucketCorsOutput, error) {
						return nil, errBoom
					},
				}),
			},
			want: want{
				status: NeedsUpdate,
				err:    awsclient.Wrap(errBoom, corsGetFailed),
			},
		},
		"UpdateNeeded": {
			args: args{
				b: s3Testing.Bucket(s3Testing.WithCORSConfig(generateCORSConfig())),
				cl: NewCORSConfigurationClient(fake.MockBucketClient{
					MockGetBucketCors: func(ctx context.Context, input *s3.GetBucketCorsInput, opts []func(*s3.Options)) (*s3.GetBucketCorsOutput, error) {
						return &s3.GetBucketCorsOutput{CORSRules: nil}, nil
					},
				}),
			},
			want: want{
				status: NeedsUpdate,
				err:    nil,
			},
		},
		"NeedsDelete": {
			args: args{
				b: s3Testing.Bucket(s3Testing.WithCORSConfig(nil)),
				cl: NewCORSConfigurationClient(fake.MockBucketClient{
					MockGetBucketCors: func(ctx context.Context, input *s3.GetBucketCorsInput, opts []func(*s3.Options)) (*s3.GetBucketCorsOutput, error) {
						return &s3.GetBucketCorsOutput{CORSRules: generateAWSCORS().CORSRules}, nil
					},
				}),
			},
			want: want{
				status: NeedsDeletion,
				err:    nil,
			},
		},
		"NoUpdateNotExists": {
			args: args{
				b: s3Testing.Bucket(s3Testing.WithCORSConfig(nil)),
				cl: NewCORSConfigurationClient(fake.MockBucketClient{
					MockGetBucketCors: func(ctx context.Context, input *s3.GetBucketCorsInput, opts []func(*s3.Options)) (*s3.GetBucketCorsOutput, error) {
						return nil, &smithy.GenericAPIError{Code: clientss3.CORSNotFoundErrCode}
					},
				}),
			},
			want: want{
				status: Updated,
				err:    nil,
			},
		},
		"NoUpdateNotExistsNil": {
			args: args{
				b: s3Testing.Bucket(s3Testing.WithCORSConfig(nil)),
				cl: NewCORSConfigurationClient(fake.MockBucketClient{
					MockGetBucketCors: func(ctx context.Context, input *s3.GetBucketCorsInput, opts []func(*s3.Options)) (*s3.GetBucketCorsOutput, error) {
						return &s3.GetBucketCorsOutput{CORSRules: nil}, nil
					},
				}),
			},
			want: want{
				status: Updated,
				err:    nil,
			},
		},
		"NoUpdateExists": {
			args: args{
				b: s3Testing.Bucket(s3Testing.WithCORSConfig(generateCORSConfig())),
				cl: NewCORSConfigurationClient(fake.MockBucketClient{
					MockGetBucketCors: func(ctx context.Context, input *s3.GetBucketCorsInput, opts []func(*s3.Options)) (*s3.GetBucketCorsOutput, error) {
						return &s3.GetBucketCorsOutput{CORSRules: generateAWSCORS().CORSRules}, nil
					},
				}),
			},
			want: want{
				status: Updated,
				err:    nil,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			status, err := tc.args.cl.Observe(context.Background(), tc.args.b)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.status, status); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestCORSCreateOrUpdate(t *testing.T) {
	type args struct {
		cl *CORSConfigurationClient
		b  *v1beta1.Bucket
	}

	type want struct {
		err error
	}

	cases := map[string]struct {
		args
		want
	}{
		"Error": {
			args: args{
				b: s3Testing.Bucket(s3Testing.WithCORSConfig(generateCORSConfig())),
				cl: NewCORSConfigurationClient(fake.MockBucketClient{
					MockPutBucketCors: func(ctx context.Context, input *s3.PutBucketCorsInput, opts []func(*s3.Options)) (*s3.PutBucketCorsOutput, error) {
						return nil, errBoom
					},
				}),
			},
			want: want{
				err: awsclient.Wrap(errBoom, corsPutFailed),
			},
		},
		"InvalidConfig": {
			args: args{
				b: s3Testing.Bucket(s3Testing.WithCORSConfig(generateCORSConfig())),
				cl: NewCORSConfigurationClient(fake.MockBucketClient{
					MockPutBucketCors: func(ctx context.Context, input *s3.PutBucketCorsInput, opts []func(*s3.Options)) (*s3.PutBucketCorsOutput, error) {
						return &s3.PutBucketCorsOutput{}, nil
					},
				}),
			},
			want: want{
				err: nil,
			},
		},
		"SuccessfulCreate": {
			args: args{
				b: s3Testing.Bucket(s3Testing.WithCORSConfig(generateCORSConfig())),
				cl: NewCORSConfigurationClient(fake.MockBucketClient{
					MockPutBucketCors: func(ctx context.Context, input *s3.PutBucketCorsInput, opts []func(*s3.Options)) (*s3.PutBucketCorsOutput, error) {
						return &s3.PutBucketCorsOutput{}, nil
					},
				}),
			},
			want: want{
				err: nil,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			err := tc.args.cl.CreateOrUpdate(context.Background(), tc.args.b)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestCORSDelete(t *testing.T) {
	type args struct {
		cl *CORSConfigurationClient
		b  *v1beta1.Bucket
	}

	type want struct {
		err error
	}

	cases := map[string]struct {
		args
		want
	}{
		"Error": {
			args: args{
				b: s3Testing.Bucket(s3Testing.WithCORSConfig(generateCORSConfig())),
				cl: NewCORSConfigurationClient(fake.MockBucketClient{
					MockDeleteBucketCors: func(ctx context.Context, input *s3.DeleteBucketCorsInput, opts []func(*s3.Options)) (*s3.DeleteBucketCorsOutput, error) {
						return nil, errBoom
					},
				}),
			},
			want: want{
				err: awsclient.Wrap(errBoom, corsDeleteFailed),
			},
		},
		"SuccessfulDelete": {
			args: args{
				b: s3Testing.Bucket(s3Testing.WithCORSConfig(generateCORSConfig())),
				cl: NewCORSConfigurationClient(fake.MockBucketClient{
					MockDeleteBucketCors: func(ctx context.Context, input *s3.DeleteBucketCorsInput, opts []func(*s3.Options)) (*s3.DeleteBucketCorsOutput, error) {
						return &s3.DeleteBucketCorsOutput{}, nil
					},
				}),
			},
			want: want{
				err: nil,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			err := tc.args.cl.Delete(context.Background(), tc.args.b)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestCORSLateInit(t *testing.T) {
	type args struct {
		cl SubresourceClient
		b  *v1beta1.Bucket
	}

	type want struct {
		err error
		cr  *v1beta1.Bucket
	}

	cases := map[string]struct {
		args
		want
	}{
		"Error": {
			args: args{
				b: s3Testing.Bucket(),

				cl: NewCORSConfigurationClient(fake.MockBucketClient{
					MockGetBucketCors: func(ctx context.Context, input *s3.GetBucketCorsInput, opts []func(*s3.Options)) (*s3.GetBucketCorsOutput, error) {
						return &s3.GetBucketCorsOutput{}, errBoom
					},
				}),
			},
			want: want{
				err: awsclient.Wrap(errBoom, corsGetFailed),
				cr:  s3Testing.Bucket(),
			},
		},
		"ErrorCORSErrCode": {
			args: args{
				b: s3Testing.Bucket(),

				cl: NewCORSConfigurationClient(fake.MockBucketClient{
					MockGetBucketCors: func(ctx context.Context, input *s3.GetBucketCorsInput, opts []func(*s3.Options)) (*s3.GetBucketCorsOutput, error) {
						return &s3.GetBucketCorsOutput{}, &smithy.GenericAPIError{Code: clientss3.CORSNotFoundErrCode}
					},
				}),
			},
			want: want{
				err: nil,
				cr:  s3Testing.Bucket(),
			},
		},
		"NoLateInitEmpty": {
			args: args{
				b: s3Testing.Bucket(),
				cl: NewCORSConfigurationClient(fake.MockBucketClient{
					MockGetBucketCors: func(ctx context.Context, input *s3.GetBucketCorsInput, opts []func(*s3.Options)) (*s3.GetBucketCorsOutput, error) {
						return &s3.GetBucketCorsOutput{CORSRules: make([]s3types.CORSRule, 0)}, nil
					},
				}),
			},
			want: want{
				err: nil,
				cr:  s3Testing.Bucket(),
			},
		},
		"SuccessfulLateInit": {
			args: args{
				b: s3Testing.Bucket(s3Testing.WithCORSConfig(nil)),
				cl: NewCORSConfigurationClient(fake.MockBucketClient{
					MockGetBucketCors: func(ctx context.Context, input *s3.GetBucketCorsInput, opts []func(*s3.Options)) (*s3.GetBucketCorsOutput, error) {
						return &s3.GetBucketCorsOutput{CORSRules: generateAWSCORS().CORSRules}, nil
					},
				}),
			},

			want: want{
				err: nil,
				cr:  s3Testing.Bucket(s3Testing.WithCORSConfig(generateCORSConfig())),
			},
		},
		"NoOpLateInit": {
			args: args{
				b: s3Testing.Bucket(s3Testing.WithCORSConfig(generateCORSConfig())),
				cl: NewCORSConfigurationClient(fake.MockBucketClient{
					MockGetBucketCors: func(ctx context.Context, input *s3.GetBucketCorsInput, opts []func(*s3.Options)) (*s3.GetBucketCorsOutput, error) {
						return &s3.GetBucketCorsOutput{CORSRules: []s3types.CORSRule{
							{},
						}}, nil
					},
				}),
			},

			want: want{
				err: nil,
				cr:  s3Testing.Bucket(s3Testing.WithCORSConfig(generateCORSConfig())),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			err := tc.args.cl.LateInitialize(context.Background(), tc.args.b)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.cr, tc.args.b, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}
