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

package bucketresources

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws/awserr"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	"golang.org/x/net/context"

	"github.com/crossplane/provider-aws/apis/s3/v1beta1"
	aws "github.com/crossplane/provider-aws/pkg/clients"
	"github.com/crossplane/provider-aws/pkg/clients/s3/fake"
)

func generateCORSConfig() *v1beta1.CORSConfiguration {
	return &v1beta1.CORSConfiguration{CORSRules: []v1beta1.CORSRule{
		{
			AllowedHeaders: []string{"test.header"},
			AllowedMethods: []string{"GET"},
			AllowedOrigins: []string{"test.origin"},
			ExposeHeaders:  []string{"test.expose"},
			MaxAgeSeconds:  aws.Int64(10),
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
			MaxAgeSeconds:  aws.Int64(10),
		},
	},
	}
}

func TestCORSExistsAndUpdated(t *testing.T) {
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
				b: bucket(withCORSConfig(generateCORSConfig())),
				cl: NewCORSConfigurationClient(
					bucket(withCORSConfig(generateCORSConfig())),
					fake.MockBucketClient{
						MockGetBucketCorsRequest: func(input *s3.GetBucketCorsInput) s3.GetBucketCorsRequest {
							return s3.GetBucketCorsRequest{
								Request: createRequest(errBoom, &s3.GetBucketCorsOutput{CORSRules: generateAWSCORS().CORSRules}),
							}
						},
					},
				),
			},
			want: want{
				status: NeedsUpdate,
				err:    errors.Wrap(errBoom, corsGetFailed),
			},
		},
		"UpdateNeeded": {
			args: args{
				b: bucket(withCORSConfig(generateCORSConfig())),
				cl: NewCORSConfigurationClient(
					bucket(withCORSConfig(generateCORSConfig())),
					fake.MockBucketClient{
						MockGetBucketCorsRequest: func(input *s3.GetBucketCorsInput) s3.GetBucketCorsRequest {
							return s3.GetBucketCorsRequest{
								Request: createRequest(nil, &s3.GetBucketCorsOutput{CORSRules: nil}),
							}
						},
					},
				),
			},
			want: want{
				status: NeedsUpdate,
				err:    nil,
			},
		},
		"NeedsDelete": {
			args: args{
				b: bucket(withCORSConfig(nil)),
				cl: NewCORSConfigurationClient(
					bucket(withCORSConfig(nil)),
					fake.MockBucketClient{
						MockGetBucketCorsRequest: func(input *s3.GetBucketCorsInput) s3.GetBucketCorsRequest {
							return s3.GetBucketCorsRequest{
								Request: createRequest(nil, &s3.GetBucketCorsOutput{CORSRules: generateAWSCORS().CORSRules}),
							}
						},
					},
				),
			},
			want: want{
				status: NeedsDeletion,
				err:    nil,
			},
		},
		"NoUpdateNotExists": {
			args: args{
				b: bucket(withCORSConfig(nil)),
				cl: NewCORSConfigurationClient(
					bucket(withCORSConfig(nil)),
					fake.MockBucketClient{
						MockGetBucketCorsRequest: func(input *s3.GetBucketCorsInput) s3.GetBucketCorsRequest {
							return s3.GetBucketCorsRequest{
								Request: createRequest(awserr.New("NoSuchCORSConfiguration", "", nil), &s3.GetBucketCorsOutput{CORSRules: nil}),
							}
						},
					},
				),
			},
			want: want{
				status: Updated,
				err:    nil,
			},
		},
		"NoUpdateNotExistsNil": {
			args: args{
				b: bucket(withCORSConfig(nil)),
				cl: NewCORSConfigurationClient(
					bucket(withCORSConfig(nil)),
					fake.MockBucketClient{
						MockGetBucketCorsRequest: func(input *s3.GetBucketCorsInput) s3.GetBucketCorsRequest {
							return s3.GetBucketCorsRequest{
								Request: createRequest(nil, &s3.GetBucketCorsOutput{CORSRules: nil}),
							}
						},
					},
				),
			},
			want: want{
				status: Updated,
				err:    nil,
			},
		},
		"NoUpdateExists": {
			args: args{
				b: bucket(withCORSConfig(generateCORSConfig())),
				cl: NewCORSConfigurationClient(
					bucket(withCORSConfig(generateCORSConfig())),
					fake.MockBucketClient{
						MockGetBucketCorsRequest: func(input *s3.GetBucketCorsInput) s3.GetBucketCorsRequest {
							return s3.GetBucketCorsRequest{
								Request: createRequest(nil, &s3.GetBucketCorsOutput{CORSRules: generateAWSCORS().CORSRules}),
							}
						},
					},
				),
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

func TestCORSCreate(t *testing.T) {
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
				b: bucket(withCORSConfig(generateCORSConfig())),
				cl: NewCORSConfigurationClient(
					bucket(withCORSConfig(generateCORSConfig())),
					fake.MockBucketClient{
						MockPutBucketCorsRequest: func(input *s3.PutBucketCorsInput) s3.PutBucketCorsRequest {
							return s3.PutBucketCorsRequest{
								Request: createRequest(errBoom, &s3.PutBucketCorsOutput{}),
							}
						},
					},
				),
			},
			want: want{
				err: errors.Wrap(errBoom, corsPutFailed),
			},
		},
		"InvalidConfig": {
			args: args{
				b: bucket(withCORSConfig(generateCORSConfig())),
				cl: NewCORSConfigurationClient(
					bucket(withCORSConfig(nil)),
					fake.MockBucketClient{
						MockPutBucketCorsRequest: func(input *s3.PutBucketCorsInput) s3.PutBucketCorsRequest {
							return s3.PutBucketCorsRequest{
								Request: createRequest(nil, &s3.PutBucketCorsOutput{}),
							}
						},
					},
				),
			},
			want: want{
				err: nil,
			},
		},
		"SuccessfulCreate": {
			args: args{
				b: bucket(withCORSConfig(generateCORSConfig())),
				cl: NewCORSConfigurationClient(
					bucket(withCORSConfig(generateCORSConfig())),
					fake.MockBucketClient{
						MockPutBucketCorsRequest: func(input *s3.PutBucketCorsInput) s3.PutBucketCorsRequest {
							return s3.PutBucketCorsRequest{
								Request: createRequest(nil, &s3.PutBucketCorsOutput{}),
							}
						},
					},
				),
			},
			want: want{
				err: nil,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := tc.args.cl.Create(context.Background(), tc.args.b)
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
				b: bucket(withCORSConfig(generateCORSConfig())),
				cl: NewCORSConfigurationClient(
					bucket(withCORSConfig(generateCORSConfig())),
					fake.MockBucketClient{
						MockDeleteBucketCorsRequest: func(input *s3.DeleteBucketCorsInput) s3.DeleteBucketCorsRequest {
							return s3.DeleteBucketCorsRequest{
								Request: createRequest(errBoom, &s3.DeleteBucketCorsOutput{}),
							}
						},
					},
				),
			},
			want: want{
				err: errors.Wrap(errBoom, corsDeleteFailed),
			},
		},
		"SuccessfulDelete": {
			args: args{
				b: bucket(withCORSConfig(generateCORSConfig())),
				cl: NewCORSConfigurationClient(
					bucket(),
					fake.MockBucketClient{
						MockDeleteBucketCorsRequest: func(input *s3.DeleteBucketCorsInput) s3.DeleteBucketCorsRequest {
							return s3.DeleteBucketCorsRequest{
								Request: createRequest(nil, &s3.DeleteBucketCorsOutput{}),
							}
						},
					},
				),
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
