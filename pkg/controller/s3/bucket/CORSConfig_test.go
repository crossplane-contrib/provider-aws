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

	"github.com/aws/aws-sdk-go-v2/aws/awserr"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"

	"github.com/crossplane/provider-aws/apis/s3/v1beta1"
	awsclient "github.com/crossplane/provider-aws/pkg/clients"
	clients3 "github.com/crossplane/provider-aws/pkg/clients/s3"
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
			MaxAgeSeconds:  awsclient.Int64(10),
		},
	},
	}
}

func generateAWSCORS() *s3.CORSConfiguration {
	return &s3.CORSConfiguration{CORSRules: []s3.CORSRule{
		{
			AllowedHeaders: []string{"test.header"},
			AllowedMethods: []string{"GET"},
			AllowedOrigins: []string{"test.origin"},
			ExposeHeaders:  []string{"test.expose"},
			MaxAgeSeconds:  awsclient.Int64(10),
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
					MockGetBucketCorsRequest: func(input *s3.GetBucketCorsInput) s3.GetBucketCorsRequest {
						return s3.GetBucketCorsRequest{
							Request: s3Testing.CreateRequest(errBoom, &s3.GetBucketCorsOutput{CORSRules: generateAWSCORS().CORSRules}),
						}
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
					MockGetBucketCorsRequest: func(input *s3.GetBucketCorsInput) s3.GetBucketCorsRequest {
						return s3.GetBucketCorsRequest{
							Request: s3Testing.CreateRequest(nil, &s3.GetBucketCorsOutput{CORSRules: nil}),
						}
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
					MockGetBucketCorsRequest: func(input *s3.GetBucketCorsInput) s3.GetBucketCorsRequest {
						return s3.GetBucketCorsRequest{
							Request: s3Testing.CreateRequest(nil, &s3.GetBucketCorsOutput{CORSRules: generateAWSCORS().CORSRules}),
						}
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
					MockGetBucketCorsRequest: func(input *s3.GetBucketCorsInput) s3.GetBucketCorsRequest {
						return s3.GetBucketCorsRequest{
							Request: s3Testing.CreateRequest(awserr.New(clients3.CORSNotFoundErrCode, "", nil), &s3.GetBucketCorsOutput{CORSRules: nil}),
						}
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
					MockGetBucketCorsRequest: func(input *s3.GetBucketCorsInput) s3.GetBucketCorsRequest {
						return s3.GetBucketCorsRequest{
							Request: s3Testing.CreateRequest(nil, &s3.GetBucketCorsOutput{CORSRules: nil}),
						}
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
					MockGetBucketCorsRequest: func(input *s3.GetBucketCorsInput) s3.GetBucketCorsRequest {
						return s3.GetBucketCorsRequest{
							Request: s3Testing.CreateRequest(nil, &s3.GetBucketCorsOutput{CORSRules: generateAWSCORS().CORSRules}),
						}
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
					MockPutBucketCorsRequest: func(input *s3.PutBucketCorsInput) s3.PutBucketCorsRequest {
						return s3.PutBucketCorsRequest{
							Request: s3Testing.CreateRequest(errBoom, &s3.PutBucketCorsOutput{}),
						}
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
					MockPutBucketCorsRequest: func(input *s3.PutBucketCorsInput) s3.PutBucketCorsRequest {
						return s3.PutBucketCorsRequest{
							Request: s3Testing.CreateRequest(nil, &s3.PutBucketCorsOutput{}),
						}
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
					MockPutBucketCorsRequest: func(input *s3.PutBucketCorsInput) s3.PutBucketCorsRequest {
						return s3.PutBucketCorsRequest{
							Request: s3Testing.CreateRequest(nil, &s3.PutBucketCorsOutput{}),
						}
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
					MockDeleteBucketCorsRequest: func(input *s3.DeleteBucketCorsInput) s3.DeleteBucketCorsRequest {
						return s3.DeleteBucketCorsRequest{
							Request: s3Testing.CreateRequest(errBoom, &s3.DeleteBucketCorsOutput{}),
						}
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
					MockDeleteBucketCorsRequest: func(input *s3.DeleteBucketCorsInput) s3.DeleteBucketCorsRequest {
						return s3.DeleteBucketCorsRequest{
							Request: s3Testing.CreateRequest(nil, &s3.DeleteBucketCorsOutput{}),
						}
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
					MockGetBucketCorsRequest: func(input *s3.GetBucketCorsInput) s3.GetBucketCorsRequest {
						return s3.GetBucketCorsRequest{
							Request: s3Testing.CreateRequest(errBoom, &s3.GetBucketCorsOutput{}),
						}
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
					MockGetBucketCorsRequest: func(input *s3.GetBucketCorsInput) s3.GetBucketCorsRequest {
						return s3.GetBucketCorsRequest{
							Request: s3Testing.CreateRequest(awserr.New(clients3.CORSNotFoundErrCode, "error", nil), &s3.GetBucketCorsOutput{}),
						}
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
					MockGetBucketCorsRequest: func(input *s3.GetBucketCorsInput) s3.GetBucketCorsRequest {
						return s3.GetBucketCorsRequest{
							Request: s3Testing.CreateRequest(nil, &s3.GetBucketCorsOutput{CORSRules: make([]s3.CORSRule, 0)}),
						}
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
					MockGetBucketCorsRequest: func(input *s3.GetBucketCorsInput) s3.GetBucketCorsRequest {
						return s3.GetBucketCorsRequest{
							Request: s3Testing.CreateRequest(nil, &s3.GetBucketCorsOutput{CORSRules: generateAWSCORS().CORSRules}),
						}
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
					MockGetBucketCorsRequest: func(input *s3.GetBucketCorsInput) s3.GetBucketCorsRequest {
						return s3.GetBucketCorsRequest{
							Request: s3Testing.CreateRequest(nil, &s3.GetBucketCorsOutput{CORSRules: []s3.CORSRule{
								{},
							}}),
						}
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
