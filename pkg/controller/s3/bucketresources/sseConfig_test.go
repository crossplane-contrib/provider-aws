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
	clients3 "github.com/crossplane/provider-aws/pkg/clients/s3"
	"github.com/crossplane/provider-aws/pkg/clients/s3/fake"
	s3Testing "github.com/crossplane/provider-aws/pkg/controller/s3/testing"
)

var (
	sseAlgo = "AES256"
	keyID   = "test-key-id"
)

func generateSSEConfig() *v1beta1.ServerSideEncryptionConfiguration {
	return &v1beta1.ServerSideEncryptionConfiguration{
		Rules: []v1beta1.ServerSideEncryptionRule{
			{
				ApplyServerSideEncryptionByDefault: v1beta1.ServerSideEncryptionByDefault{
					KMSMasterKeyID: &keyID,
					SSEAlgorithm:   sseAlgo,
				},
			},
		},
	}
}

func generateAWSSSE() *s3.ServerSideEncryptionConfiguration {
	return &s3.ServerSideEncryptionConfiguration{
		Rules: []s3.ServerSideEncryptionRule{
			{
				ApplyServerSideEncryptionByDefault: &s3.ServerSideEncryptionByDefault{
					KMSMasterKeyID: &keyID,
					SSEAlgorithm:   s3.ServerSideEncryptionAes256,
				},
			},
		},
	}
}

func TestSSEObserve(t *testing.T) {
	type args struct {
		cl *SSEConfigurationClient
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
				b: s3Testing.Bucket(s3Testing.WithSSEConfig(generateSSEConfig())),
				cl: NewSSEConfigurationClient(
					s3Testing.Bucket(s3Testing.WithSSEConfig(generateSSEConfig())),
					fake.MockBucketClient{
						MockGetBucketEncryptionRequest: func(input *s3.GetBucketEncryptionInput) s3.GetBucketEncryptionRequest {
							return s3.GetBucketEncryptionRequest{
								Request: s3Testing.CreateRequest(errBoom, &s3.GetBucketEncryptionOutput{}),
							}
						},
					},
				),
			},
			want: want{
				status: NeedsUpdate,
				err:    errors.Wrap(errBoom, sseGetFailed),
			},
		},
		"UpdateNeeded": {
			args: args{
				b: s3Testing.Bucket(s3Testing.WithSSEConfig(generateSSEConfig())),
				cl: NewSSEConfigurationClient(
					s3Testing.Bucket(s3Testing.WithSSEConfig(generateSSEConfig())),
					fake.MockBucketClient{
						MockGetBucketEncryptionRequest: func(input *s3.GetBucketEncryptionInput) s3.GetBucketEncryptionRequest {
							return s3.GetBucketEncryptionRequest{
								Request: s3Testing.CreateRequest(nil, &s3.GetBucketEncryptionOutput{ServerSideEncryptionConfiguration: nil}),
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
				b: s3Testing.Bucket(s3Testing.WithSSEConfig(nil)),
				cl: NewSSEConfigurationClient(
					s3Testing.Bucket(s3Testing.WithSSEConfig(nil)),
					fake.MockBucketClient{
						MockGetBucketEncryptionRequest: func(input *s3.GetBucketEncryptionInput) s3.GetBucketEncryptionRequest {
							return s3.GetBucketEncryptionRequest{
								Request: s3Testing.CreateRequest(nil, &s3.GetBucketEncryptionOutput{ServerSideEncryptionConfiguration: generateAWSSSE()}),
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
				b: s3Testing.Bucket(s3Testing.WithSSEConfig(nil)),
				cl: NewSSEConfigurationClient(
					s3Testing.Bucket(s3Testing.WithSSEConfig(nil)),
					fake.MockBucketClient{
						MockGetBucketEncryptionRequest: func(input *s3.GetBucketEncryptionInput) s3.GetBucketEncryptionRequest {
							return s3.GetBucketEncryptionRequest{
								Request: s3Testing.CreateRequest(awserr.New(clients3.SSEErrCode, "", nil), &s3.GetBucketEncryptionOutput{}),
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
				b: s3Testing.Bucket(s3Testing.WithSSEConfig(nil)),
				cl: NewSSEConfigurationClient(
					s3Testing.Bucket(s3Testing.WithSSEConfig(nil)),
					fake.MockBucketClient{
						MockGetBucketEncryptionRequest: func(input *s3.GetBucketEncryptionInput) s3.GetBucketEncryptionRequest {
							return s3.GetBucketEncryptionRequest{
								Request: s3Testing.CreateRequest(nil, &s3.GetBucketEncryptionOutput{ServerSideEncryptionConfiguration: nil}),
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
				b: s3Testing.Bucket(s3Testing.WithSSEConfig(generateSSEConfig())),
				cl: NewSSEConfigurationClient(
					s3Testing.Bucket(s3Testing.WithSSEConfig(generateSSEConfig())),
					fake.MockBucketClient{
						MockGetBucketEncryptionRequest: func(input *s3.GetBucketEncryptionInput) s3.GetBucketEncryptionRequest {
							return s3.GetBucketEncryptionRequest{
								Request: s3Testing.CreateRequest(nil, &s3.GetBucketEncryptionOutput{ServerSideEncryptionConfiguration: generateAWSSSE()}),
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

func TestSSECreateOrUpdate(t *testing.T) {
	type args struct {
		cl *SSEConfigurationClient
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
				b: s3Testing.Bucket(s3Testing.WithSSEConfig(generateSSEConfig())),
				cl: NewSSEConfigurationClient(
					s3Testing.Bucket(s3Testing.WithSSEConfig(generateSSEConfig())),
					fake.MockBucketClient{
						MockPutBucketEncryptionRequest: func(input *s3.PutBucketEncryptionInput) s3.PutBucketEncryptionRequest {
							return s3.PutBucketEncryptionRequest{
								Request: s3Testing.CreateRequest(errBoom, &s3.PutBucketEncryptionOutput{}),
							}
						},
					},
				),
			},
			want: want{
				err: errors.Wrap(errBoom, ssePutFailed),
			},
		},
		"InvalidConfig": {
			args: args{
				b: s3Testing.Bucket(s3Testing.WithSSEConfig(generateSSEConfig())),
				cl: NewSSEConfigurationClient(
					s3Testing.Bucket(s3Testing.WithSSEConfig(nil)),
					fake.MockBucketClient{
						MockPutBucketEncryptionRequest: func(input *s3.PutBucketEncryptionInput) s3.PutBucketEncryptionRequest {
							return s3.PutBucketEncryptionRequest{
								Request: s3Testing.CreateRequest(nil, &s3.PutBucketEncryptionOutput{}),
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
				b: s3Testing.Bucket(s3Testing.WithSSEConfig(generateSSEConfig())),
				cl: NewSSEConfigurationClient(
					s3Testing.Bucket(s3Testing.WithSSEConfig(generateSSEConfig())),
					fake.MockBucketClient{
						MockPutBucketEncryptionRequest: func(input *s3.PutBucketEncryptionInput) s3.PutBucketEncryptionRequest {
							return s3.PutBucketEncryptionRequest{
								Request: s3Testing.CreateRequest(nil, &s3.PutBucketEncryptionOutput{}),
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
			_, err := tc.args.cl.CreateOrUpdate(context.Background(), tc.args.b)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestSSEDelete(t *testing.T) {
	type args struct {
		cl *SSEConfigurationClient
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
				b: s3Testing.Bucket(s3Testing.WithSSEConfig(generateSSEConfig())),
				cl: NewSSEConfigurationClient(
					s3Testing.Bucket(s3Testing.WithSSEConfig(generateSSEConfig())),
					fake.MockBucketClient{
						MockDeleteBucketEncryptionRequest: func(input *s3.DeleteBucketEncryptionInput) s3.DeleteBucketEncryptionRequest {
							return s3.DeleteBucketEncryptionRequest{
								Request: s3Testing.CreateRequest(errBoom, &s3.DeleteBucketEncryptionOutput{}),
							}
						},
					},
				),
			},
			want: want{
				err: errors.Wrap(errBoom, sseDeleteFailed),
			},
		},
		"SuccessfulDelete": {
			args: args{
				b: s3Testing.Bucket(s3Testing.WithSSEConfig(generateSSEConfig())),
				cl: NewSSEConfigurationClient(
					s3Testing.Bucket(s3Testing.WithSSEConfig(generateSSEConfig())),
					fake.MockBucketClient{
						MockDeleteBucketEncryptionRequest: func(input *s3.DeleteBucketEncryptionInput) s3.DeleteBucketEncryptionRequest {
							return s3.DeleteBucketEncryptionRequest{
								Request: s3Testing.CreateRequest(nil, &s3.DeleteBucketEncryptionOutput{}),
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
