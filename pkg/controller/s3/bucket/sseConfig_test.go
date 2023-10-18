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

	"github.com/crossplane-contrib/provider-aws/apis/s3/v1beta1"
	clients3 "github.com/crossplane-contrib/provider-aws/pkg/clients/s3"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/s3/fake"
	s3testing "github.com/crossplane-contrib/provider-aws/pkg/controller/s3/testing"
	errorutils "github.com/crossplane-contrib/provider-aws/pkg/utils/errors"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
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
					KMSMasterKeyID: pointer.ToOrNilIfZeroValue(keyID),
					SSEAlgorithm:   sseAlgo,
				},
			},
		},
	}
}

func generateSSEConfigWithBucketEncryption() *v1beta1.ServerSideEncryptionConfiguration {
	return &v1beta1.ServerSideEncryptionConfiguration{
		Rules: []v1beta1.ServerSideEncryptionRule{
			{
				BucketKeyEnabled: *pointer.ToOrNilIfZeroValue(true),
				ApplyServerSideEncryptionByDefault: v1beta1.ServerSideEncryptionByDefault{
					KMSMasterKeyID: pointer.ToOrNilIfZeroValue(keyID),
					SSEAlgorithm:   sseAlgo,
				},
			},
		},
	}
}

func generateAWSSSE() *s3types.ServerSideEncryptionConfiguration {
	return &s3types.ServerSideEncryptionConfiguration{
		Rules: []s3types.ServerSideEncryptionRule{
			{
				ApplyServerSideEncryptionByDefault: &s3types.ServerSideEncryptionByDefault{
					KMSMasterKeyID: pointer.ToOrNilIfZeroValue(keyID),
					SSEAlgorithm:   s3types.ServerSideEncryptionAes256,
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
				b: s3testing.Bucket(s3testing.WithSSEConfig(generateSSEConfig())),
				cl: NewSSEConfigurationClient(fake.MockBucketClient{
					MockGetBucketEncryption: func(ctx context.Context, input *s3.GetBucketEncryptionInput, opts []func(*s3.Options)) (*s3.GetBucketEncryptionOutput, error) {
						return nil, errBoom
					},
				}),
			},
			want: want{
				status: NeedsUpdate,
				err:    errorutils.Wrap(errBoom, sseGetFailed),
			},
		},
		"UpdateNeeded": {
			args: args{
				b: s3testing.Bucket(s3testing.WithSSEConfig(generateSSEConfig())),
				cl: NewSSEConfigurationClient(fake.MockBucketClient{
					MockGetBucketEncryption: func(ctx context.Context, input *s3.GetBucketEncryptionInput, opts []func(*s3.Options)) (*s3.GetBucketEncryptionOutput, error) {
						return &s3.GetBucketEncryptionOutput{ServerSideEncryptionConfiguration: nil}, nil
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
				b: s3testing.Bucket(s3testing.WithSSEConfig(nil)),
				cl: NewSSEConfigurationClient(fake.MockBucketClient{
					MockGetBucketEncryption: func(ctx context.Context, input *s3.GetBucketEncryptionInput, opts []func(*s3.Options)) (*s3.GetBucketEncryptionOutput, error) {
						return &s3.GetBucketEncryptionOutput{ServerSideEncryptionConfiguration: generateAWSSSE()}, nil
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
				b: s3testing.Bucket(s3testing.WithSSEConfig(nil)),
				cl: NewSSEConfigurationClient(fake.MockBucketClient{
					MockGetBucketEncryption: func(ctx context.Context, input *s3.GetBucketEncryptionInput, opts []func(*s3.Options)) (*s3.GetBucketEncryptionOutput, error) {
						return nil, &smithy.GenericAPIError{Code: clients3.SSENotFoundErrCode}
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
				b: s3testing.Bucket(s3testing.WithSSEConfig(nil)),
				cl: NewSSEConfigurationClient(fake.MockBucketClient{
					MockGetBucketEncryption: func(ctx context.Context, input *s3.GetBucketEncryptionInput, opts []func(*s3.Options)) (*s3.GetBucketEncryptionOutput, error) {
						return &s3.GetBucketEncryptionOutput{ServerSideEncryptionConfiguration: nil}, nil
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
				b: s3testing.Bucket(s3testing.WithSSEConfig(generateSSEConfig())),
				cl: NewSSEConfigurationClient(fake.MockBucketClient{
					MockGetBucketEncryption: func(ctx context.Context, input *s3.GetBucketEncryptionInput, opts []func(*s3.Options)) (*s3.GetBucketEncryptionOutput, error) {
						return &s3.GetBucketEncryptionOutput{ServerSideEncryptionConfiguration: generateAWSSSE()}, nil
					},
				}),
			},
			want: want{
				status: Updated,
				err:    nil,
			},
		},
		"NeedsUpdateEnableBucketKey": {
			args: args{
				b: s3testing.Bucket(s3testing.WithSSEConfig(generateSSEConfigWithBucketEncryption())),
				cl: NewSSEConfigurationClient(fake.MockBucketClient{
					MockGetBucketEncryption: func(ctx context.Context, input *s3.GetBucketEncryptionInput, opts []func(*s3.Options)) (*s3.GetBucketEncryptionOutput, error) {
						return &s3.GetBucketEncryptionOutput{ServerSideEncryptionConfiguration: generateAWSSSE()}, nil
					},
				}),
			},
			want: want{
				status: NeedsUpdate,
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
				b: s3testing.Bucket(s3testing.WithSSEConfig(generateSSEConfig())),
				cl: NewSSEConfigurationClient(fake.MockBucketClient{
					MockPutBucketEncryption: func(ctx context.Context, input *s3.PutBucketEncryptionInput, opts []func(*s3.Options)) (*s3.PutBucketEncryptionOutput, error) {
						return nil, errBoom
					},
				}),
			},
			want: want{
				err: errorutils.Wrap(errBoom, ssePutFailed),
			},
		},
		"InvalidConfig": {
			args: args{
				b: s3testing.Bucket(s3testing.WithSSEConfig(generateSSEConfig())),
				cl: NewSSEConfigurationClient(fake.MockBucketClient{
					MockPutBucketEncryption: func(ctx context.Context, input *s3.PutBucketEncryptionInput, opts []func(*s3.Options)) (*s3.PutBucketEncryptionOutput, error) {
						return &s3.PutBucketEncryptionOutput{}, nil
					},
				}),
			},
			want: want{
				err: nil,
			},
		},
		"SuccessfulCreate": {
			args: args{
				b: s3testing.Bucket(s3testing.WithSSEConfig(generateSSEConfig())),
				cl: NewSSEConfigurationClient(fake.MockBucketClient{
					MockPutBucketEncryption: func(ctx context.Context, input *s3.PutBucketEncryptionInput, opts []func(*s3.Options)) (*s3.PutBucketEncryptionOutput, error) {
						return &s3.PutBucketEncryptionOutput{}, nil
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
				b: s3testing.Bucket(s3testing.WithSSEConfig(generateSSEConfig())),
				cl: NewSSEConfigurationClient(fake.MockBucketClient{
					MockDeleteBucketEncryption: func(ctx context.Context, input *s3.DeleteBucketEncryptionInput, opts []func(*s3.Options)) (*s3.DeleteBucketEncryptionOutput, error) {
						return nil, errBoom
					},
				}),
			},
			want: want{
				err: errorutils.Wrap(errBoom, sseDeleteFailed),
			},
		},
		"SuccessfulDelete": {
			args: args{
				b: s3testing.Bucket(s3testing.WithSSEConfig(generateSSEConfig())),
				cl: NewSSEConfigurationClient(fake.MockBucketClient{
					MockDeleteBucketEncryption: func(ctx context.Context, input *s3.DeleteBucketEncryptionInput, opts []func(*s3.Options)) (*s3.DeleteBucketEncryptionOutput, error) {
						return &s3.DeleteBucketEncryptionOutput{}, nil
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
				b: s3testing.Bucket(),
				cl: NewSSEConfigurationClient(fake.MockBucketClient{
					MockGetBucketEncryption: func(ctx context.Context, input *s3.GetBucketEncryptionInput, opts []func(*s3.Options)) (*s3.GetBucketEncryptionOutput, error) {
						return &s3.GetBucketEncryptionOutput{}, errBoom
					},
				}),
			},
			want: want{
				err: errorutils.Wrap(errBoom, sseGetFailed),
				cr:  s3testing.Bucket(),
			},
		},
		"ErrorSSEConfigurationNotFound": {
			args: args{
				b: s3testing.Bucket(),
				cl: NewSSEConfigurationClient(fake.MockBucketClient{
					MockGetBucketEncryption: func(ctx context.Context, input *s3.GetBucketEncryptionInput, opts []func(*s3.Options)) (*s3.GetBucketEncryptionOutput, error) {
						return &s3.GetBucketEncryptionOutput{}, &smithy.GenericAPIError{Code: clients3.SSENotFoundErrCode}
					},
				}),
			},
			want: want{
				err: nil,
				cr:  s3testing.Bucket(),
			},
		},
		"NoLateInitNil": {
			args: args{
				b: s3testing.Bucket(),
				cl: NewSSEConfigurationClient(fake.MockBucketClient{
					MockGetBucketEncryption: func(ctx context.Context, input *s3.GetBucketEncryptionInput, opts []func(*s3.Options)) (*s3.GetBucketEncryptionOutput, error) {
						return &s3.GetBucketEncryptionOutput{ServerSideEncryptionConfiguration: nil}, nil
					},
				}),
			},
			want: want{
				err: nil,
				cr:  s3testing.Bucket(),
			},
		},
		"NoLateInitEmpty": {
			args: args{
				b: s3testing.Bucket(),
				cl: NewSSEConfigurationClient(fake.MockBucketClient{
					MockGetBucketEncryption: func(ctx context.Context, input *s3.GetBucketEncryptionInput, opts []func(*s3.Options)) (*s3.GetBucketEncryptionOutput, error) {
						return &s3.GetBucketEncryptionOutput{
							ServerSideEncryptionConfiguration: &s3types.ServerSideEncryptionConfiguration{
								Rules: nil,
							},
						}, nil
					},
				}),
			},
			want: want{
				err: nil,
				cr:  s3testing.Bucket(),
			},
		},
		"SuccessfulLateInit": {
			args: args{
				b: s3testing.Bucket(s3testing.WithSSEConfig(nil)),
				cl: NewSSEConfigurationClient(fake.MockBucketClient{
					MockGetBucketEncryption: func(ctx context.Context, input *s3.GetBucketEncryptionInput, opts []func(*s3.Options)) (*s3.GetBucketEncryptionOutput, error) {
						return &s3.GetBucketEncryptionOutput{ServerSideEncryptionConfiguration: generateAWSSSE()}, nil
					},
				}),
			},
			want: want{
				err: nil,
				cr:  s3testing.Bucket(s3testing.WithSSEConfig(generateSSEConfig())),
			},
		},
		"NoOpLateInit": {
			args: args{
				b: s3testing.Bucket(s3testing.WithSSEConfig(generateSSEConfig())),
				cl: NewSSEConfigurationClient(fake.MockBucketClient{
					MockGetBucketEncryption: func(ctx context.Context, input *s3.GetBucketEncryptionInput, opts []func(*s3.Options)) (*s3.GetBucketEncryptionOutput, error) {
						return &s3.GetBucketEncryptionOutput{
							ServerSideEncryptionConfiguration: &s3types.ServerSideEncryptionConfiguration{
								Rules: []s3types.ServerSideEncryptionRule{
									{},
								},
							},
						}, nil
					},
				}),
			},
			want: want{
				err: nil,
				cr:  s3testing.Bucket(s3testing.WithSSEConfig(generateSSEConfig())),
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
