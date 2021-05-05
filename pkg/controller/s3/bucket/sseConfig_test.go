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

const (
	keyID   = "test-key-id"
	sseAlgo = "AES256"
)

var (
	_ SubresourceClient = &SSEConfigurationClient{}
)

func generateSSEConfig() *v1beta1.ServerSideEncryptionConfiguration {
	return &v1beta1.ServerSideEncryptionConfiguration{
		Rules: []v1beta1.ServerSideEncryptionRule{
			{
				ApplyServerSideEncryptionByDefault: v1beta1.ServerSideEncryptionByDefault{
					KMSMasterKeyID: awsclient.String(keyID),
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
					KMSMasterKeyID: awsclient.String(keyID),
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
				cl: NewSSEConfigurationClient(fake.MockBucketClient{
					MockGetBucketEncryptionRequest: func(input *s3.GetBucketEncryptionInput) s3.GetBucketEncryptionRequest {
						return s3.GetBucketEncryptionRequest{
							Request: s3Testing.CreateRequest(errBoom, &s3.GetBucketEncryptionOutput{}),
						}
					},
				}),
			},
			want: want{
				status: NeedsUpdate,
				err:    awsclient.Wrap(errBoom, sseGetFailed),
			},
		},
		"UpdateNeeded": {
			args: args{
				b: s3Testing.Bucket(s3Testing.WithSSEConfig(generateSSEConfig())),
				cl: NewSSEConfigurationClient(fake.MockBucketClient{
					MockGetBucketEncryptionRequest: func(input *s3.GetBucketEncryptionInput) s3.GetBucketEncryptionRequest {
						return s3.GetBucketEncryptionRequest{
							Request: s3Testing.CreateRequest(nil, &s3.GetBucketEncryptionOutput{ServerSideEncryptionConfiguration: nil}),
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
				b: s3Testing.Bucket(s3Testing.WithSSEConfig(nil)),
				cl: NewSSEConfigurationClient(fake.MockBucketClient{
					MockGetBucketEncryptionRequest: func(input *s3.GetBucketEncryptionInput) s3.GetBucketEncryptionRequest {
						return s3.GetBucketEncryptionRequest{
							Request: s3Testing.CreateRequest(nil, &s3.GetBucketEncryptionOutput{ServerSideEncryptionConfiguration: generateAWSSSE()}),
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
				b: s3Testing.Bucket(s3Testing.WithSSEConfig(nil)),
				cl: NewSSEConfigurationClient(fake.MockBucketClient{
					MockGetBucketEncryptionRequest: func(input *s3.GetBucketEncryptionInput) s3.GetBucketEncryptionRequest {
						return s3.GetBucketEncryptionRequest{
							Request: s3Testing.CreateRequest(awserr.New(clients3.SSENotFoundErrCode, "", nil), &s3.GetBucketEncryptionOutput{}),
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
				b: s3Testing.Bucket(s3Testing.WithSSEConfig(nil)),
				cl: NewSSEConfigurationClient(fake.MockBucketClient{
					MockGetBucketEncryptionRequest: func(input *s3.GetBucketEncryptionInput) s3.GetBucketEncryptionRequest {
						return s3.GetBucketEncryptionRequest{
							Request: s3Testing.CreateRequest(nil, &s3.GetBucketEncryptionOutput{ServerSideEncryptionConfiguration: nil}),
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
				b: s3Testing.Bucket(s3Testing.WithSSEConfig(generateSSEConfig())),
				cl: NewSSEConfigurationClient(fake.MockBucketClient{
					MockGetBucketEncryptionRequest: func(input *s3.GetBucketEncryptionInput) s3.GetBucketEncryptionRequest {
						return s3.GetBucketEncryptionRequest{
							Request: s3Testing.CreateRequest(nil, &s3.GetBucketEncryptionOutput{ServerSideEncryptionConfiguration: generateAWSSSE()}),
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
				cl: NewSSEConfigurationClient(fake.MockBucketClient{
					MockPutBucketEncryptionRequest: func(input *s3.PutBucketEncryptionInput) s3.PutBucketEncryptionRequest {
						return s3.PutBucketEncryptionRequest{
							Request: s3Testing.CreateRequest(errBoom, &s3.PutBucketEncryptionOutput{}),
						}
					},
				}),
			},
			want: want{
				err: awsclient.Wrap(errBoom, ssePutFailed),
			},
		},
		"InvalidConfig": {
			args: args{
				b: s3Testing.Bucket(s3Testing.WithSSEConfig(generateSSEConfig())),
				cl: NewSSEConfigurationClient(fake.MockBucketClient{
					MockPutBucketEncryptionRequest: func(input *s3.PutBucketEncryptionInput) s3.PutBucketEncryptionRequest {
						return s3.PutBucketEncryptionRequest{
							Request: s3Testing.CreateRequest(nil, &s3.PutBucketEncryptionOutput{}),
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
				b: s3Testing.Bucket(s3Testing.WithSSEConfig(generateSSEConfig())),
				cl: NewSSEConfigurationClient(fake.MockBucketClient{
					MockPutBucketEncryptionRequest: func(input *s3.PutBucketEncryptionInput) s3.PutBucketEncryptionRequest {
						return s3.PutBucketEncryptionRequest{
							Request: s3Testing.CreateRequest(nil, &s3.PutBucketEncryptionOutput{}),
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
				cl: NewSSEConfigurationClient(fake.MockBucketClient{
					MockDeleteBucketEncryptionRequest: func(input *s3.DeleteBucketEncryptionInput) s3.DeleteBucketEncryptionRequest {
						return s3.DeleteBucketEncryptionRequest{
							Request: s3Testing.CreateRequest(errBoom, &s3.DeleteBucketEncryptionOutput{}),
						}
					},
				}),
			},
			want: want{
				err: awsclient.Wrap(errBoom, sseDeleteFailed),
			},
		},
		"SuccessfulDelete": {
			args: args{
				b: s3Testing.Bucket(s3Testing.WithSSEConfig(generateSSEConfig())),
				cl: NewSSEConfigurationClient(fake.MockBucketClient{
					MockDeleteBucketEncryptionRequest: func(input *s3.DeleteBucketEncryptionInput) s3.DeleteBucketEncryptionRequest {
						return s3.DeleteBucketEncryptionRequest{
							Request: s3Testing.CreateRequest(nil, &s3.DeleteBucketEncryptionOutput{}),
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

func TestSSELateInit(t *testing.T) {
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
				cl: NewSSEConfigurationClient(fake.MockBucketClient{
					MockGetBucketEncryptionRequest: func(input *s3.GetBucketEncryptionInput) s3.GetBucketEncryptionRequest {
						return s3.GetBucketEncryptionRequest{
							Request: s3Testing.CreateRequest(errBoom, &s3.GetBucketEncryptionOutput{}),
						}
					},
				}),
			},
			want: want{
				err: awsclient.Wrap(errBoom, sseGetFailed),
				cr:  s3Testing.Bucket(),
			},
		},
		"ErrorSSEConfigurationNotFound": {
			args: args{
				b: s3Testing.Bucket(),
				cl: NewSSEConfigurationClient(fake.MockBucketClient{
					MockGetBucketEncryptionRequest: func(input *s3.GetBucketEncryptionInput) s3.GetBucketEncryptionRequest {
						return s3.GetBucketEncryptionRequest{
							Request: s3Testing.CreateRequest(awserr.New(clients3.SSENotFoundErrCode, "error", nil), &s3.GetBucketEncryptionOutput{}),
						}
					},
				}),
			},
			want: want{
				err: nil,
				cr:  s3Testing.Bucket(),
			},
		},
		"NoLateInitNil": {
			args: args{
				b: s3Testing.Bucket(),
				cl: NewSSEConfigurationClient(fake.MockBucketClient{
					MockGetBucketEncryptionRequest: func(input *s3.GetBucketEncryptionInput) s3.GetBucketEncryptionRequest {
						return s3.GetBucketEncryptionRequest{
							Request: s3Testing.CreateRequest(nil, &s3.GetBucketEncryptionOutput{ServerSideEncryptionConfiguration: nil}),
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
				cl: NewSSEConfigurationClient(fake.MockBucketClient{
					MockGetBucketEncryptionRequest: func(input *s3.GetBucketEncryptionInput) s3.GetBucketEncryptionRequest {
						return s3.GetBucketEncryptionRequest{
							Request: s3Testing.CreateRequest(nil, &s3.GetBucketEncryptionOutput{
								ServerSideEncryptionConfiguration: &s3.ServerSideEncryptionConfiguration{Rules: nil},
							}),
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
				b: s3Testing.Bucket(s3Testing.WithSSEConfig(nil)),
				cl: NewSSEConfigurationClient(fake.MockBucketClient{
					MockGetBucketEncryptionRequest: func(input *s3.GetBucketEncryptionInput) s3.GetBucketEncryptionRequest {
						return s3.GetBucketEncryptionRequest{
							Request: s3Testing.CreateRequest(nil, &s3.GetBucketEncryptionOutput{ServerSideEncryptionConfiguration: generateAWSSSE()}),
						}
					},
				}),
			},
			want: want{
				err: nil,
				cr:  s3Testing.Bucket(s3Testing.WithSSEConfig(generateSSEConfig())),
			},
		},
		"NoOpLateInit": {
			args: args{
				b: s3Testing.Bucket(s3Testing.WithSSEConfig(generateSSEConfig())),
				cl: NewSSEConfigurationClient(fake.MockBucketClient{
					MockGetBucketEncryptionRequest: func(input *s3.GetBucketEncryptionInput) s3.GetBucketEncryptionRequest {
						return s3.GetBucketEncryptionRequest{
							Request: s3Testing.CreateRequest(nil, &s3.GetBucketEncryptionOutput{
								ServerSideEncryptionConfiguration: &s3.ServerSideEncryptionConfiguration{
									Rules: []s3.ServerSideEncryptionRule{
										{},
									},
								},
							}),
						}
					},
				}),
			},
			want: want{
				err: nil,
				cr:  s3Testing.Bucket(s3Testing.WithSSEConfig(generateSSEConfig())),
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
