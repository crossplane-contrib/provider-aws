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
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"

	"github.com/crossplane-contrib/provider-aws/apis/s3/v1beta1"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/s3/fake"
	s3testing "github.com/crossplane-contrib/provider-aws/pkg/controller/s3/testing"
	errorutils "github.com/crossplane-contrib/provider-aws/pkg/utils/errors"
)

var (
	_           SubresourceClient = &LoggingConfigurationClient{}
	bucketName                    = "test.Bucket.name"
	permission                    = "FULL_CONTROL"
	displayName                   = "name"
	email                         = "test@user.com"
	userType                      = "CanonicalUser"
	groupURI                      = "uri"
)

func generateLoggingConfig() *v1beta1.LoggingConfiguration {
	return &v1beta1.LoggingConfiguration{
		TargetBucket: &bucketName,
		TargetPrefix: prefix,
		TargetGrants: []v1beta1.TargetGrant{{
			Grantee: v1beta1.TargetGrantee{
				DisplayName:  &displayName,
				EmailAddress: &email,
				ID:           &id,
				Type:         userType,
				URI:          &groupURI,
			},
			Permission: permission,
		}},
	}
}

func generateAWSLogging() *s3types.LoggingEnabled {
	return &s3types.LoggingEnabled{
		TargetBucket: &bucketName,
		TargetGrants: []s3types.TargetGrant{{
			Grantee: &s3types.Grantee{
				DisplayName:  &displayName,
				EmailAddress: &email,
				ID:           &id,
				Type:         s3types.TypeCanonicalUser,
				URI:          &groupURI,
			},
			Permission: s3types.BucketLogsPermissionFullControl,
		}},
		TargetPrefix: &prefix,
	}
}

func TestLoggingObserve(t *testing.T) {
	type args struct {
		cl *LoggingConfigurationClient
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
				b: s3testing.Bucket(s3testing.WithLoggingConfig(generateLoggingConfig())),
				cl: NewLoggingConfigurationClient(fake.MockBucketClient{
					MockGetBucketLogging: func(ctx context.Context, input *s3.GetBucketLoggingInput, opts []func(*s3.Options)) (*s3.GetBucketLoggingOutput, error) {
						return nil, errBoom
					},
				}),
			},
			want: want{
				status: NeedsUpdate,
				err:    errorutils.Wrap(errBoom, loggingGetFailed),
			},
		},
		"UpdateNeeded": {
			args: args{
				b: s3testing.Bucket(s3testing.WithLoggingConfig(generateLoggingConfig())),
				cl: NewLoggingConfigurationClient(fake.MockBucketClient{
					MockGetBucketLogging: func(ctx context.Context, input *s3.GetBucketLoggingInput, opts []func(*s3.Options)) (*s3.GetBucketLoggingOutput, error) {
						return &s3.GetBucketLoggingOutput{LoggingEnabled: nil}, nil
					},
				}),
			},
			want: want{
				status: NeedsUpdate,
				err:    nil,
			},
		},
		"NoUpdateNotExists": {
			args: args{
				b: s3testing.Bucket(s3testing.WithLoggingConfig(nil)),
				cl: NewLoggingConfigurationClient(fake.MockBucketClient{
					MockGetBucketLogging: func(ctx context.Context, input *s3.GetBucketLoggingInput, opts []func(*s3.Options)) (*s3.GetBucketLoggingOutput, error) {
						return &s3.GetBucketLoggingOutput{LoggingEnabled: nil}, nil
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
				b: s3testing.Bucket(s3testing.WithLoggingConfig(generateLoggingConfig())),
				cl: NewLoggingConfigurationClient(fake.MockBucketClient{
					MockGetBucketLogging: func(ctx context.Context, input *s3.GetBucketLoggingInput, opts []func(*s3.Options)) (*s3.GetBucketLoggingOutput, error) {
						return &s3.GetBucketLoggingOutput{LoggingEnabled: generateAWSLogging()}, nil
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

func TestLoggingCreateOrUpdate(t *testing.T) {
	type args struct {
		cl *LoggingConfigurationClient
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
				b: s3testing.Bucket(s3testing.WithLoggingConfig(generateLoggingConfig())),
				cl: NewLoggingConfigurationClient(fake.MockBucketClient{
					MockPutBucketLogging: func(ctx context.Context, input *s3.PutBucketLoggingInput, opts []func(*s3.Options)) (*s3.PutBucketLoggingOutput, error) {
						return nil, errBoom
					},
				}),
			},
			want: want{
				err: errorutils.Wrap(errBoom, loggingPutFailed),
			},
		},
		"InvalidConfig": {
			args: args{
				b: s3testing.Bucket(s3testing.WithLoggingConfig(generateLoggingConfig())),
				cl: NewLoggingConfigurationClient(fake.MockBucketClient{
					MockPutBucketLogging: func(ctx context.Context, input *s3.PutBucketLoggingInput, opts []func(*s3.Options)) (*s3.PutBucketLoggingOutput, error) {
						return &s3.PutBucketLoggingOutput{}, nil
					},
				}),
			},
			want: want{
				err: nil,
			},
		},
		"SuccessfulCreate": {
			args: args{
				b: s3testing.Bucket(s3testing.WithLoggingConfig(generateLoggingConfig())),
				cl: NewLoggingConfigurationClient(fake.MockBucketClient{
					MockPutBucketLogging: func(ctx context.Context, input *s3.PutBucketLoggingInput, opts []func(*s3.Options)) (*s3.PutBucketLoggingOutput, error) {
						return &s3.PutBucketLoggingOutput{}, nil
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

func TestLoggingLateInit(t *testing.T) {
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
				cl: NewLoggingConfigurationClient(fake.MockBucketClient{
					MockGetBucketLogging: func(ctx context.Context, input *s3.GetBucketLoggingInput, opts []func(*s3.Options)) (*s3.GetBucketLoggingOutput, error) {
						return &s3.GetBucketLoggingOutput{}, errBoom
					},
				}),
			},
			want: want{
				err: errorutils.Wrap(errBoom, loggingGetFailed),
				cr:  s3testing.Bucket(),
			},
		},
		"NoLateInitEmpty": {
			args: args{
				b: s3testing.Bucket(),
				cl: NewLoggingConfigurationClient(fake.MockBucketClient{
					MockGetBucketLogging: func(ctx context.Context, input *s3.GetBucketLoggingInput, opts []func(*s3.Options)) (*s3.GetBucketLoggingOutput, error) {
						return &s3.GetBucketLoggingOutput{}, nil
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
				b: s3testing.Bucket(s3testing.WithLoggingConfig(nil)),
				cl: NewLoggingConfigurationClient(fake.MockBucketClient{
					MockGetBucketLogging: func(ctx context.Context, input *s3.GetBucketLoggingInput, opts []func(*s3.Options)) (*s3.GetBucketLoggingOutput, error) {
						return &s3.GetBucketLoggingOutput{LoggingEnabled: generateAWSLogging()}, nil
					},
				}),
			},
			want: want{
				err: nil,
				cr:  s3testing.Bucket(s3testing.WithLoggingConfig(generateLoggingConfig())),
			},
		},
		"NoOpLateInit": {
			args: args{
				b: s3testing.Bucket(s3testing.WithLoggingConfig(generateLoggingConfig())),
				cl: NewLoggingConfigurationClient(fake.MockBucketClient{
					MockGetBucketLogging: func(ctx context.Context, input *s3.GetBucketLoggingInput, opts []func(*s3.Options)) (*s3.GetBucketLoggingOutput, error) {
						return &s3.GetBucketLoggingOutput{LoggingEnabled: &s3types.LoggingEnabled{}}, nil
					},
				}),
			},
			want: want{
				err: nil,
				cr:  s3testing.Bucket(s3testing.WithLoggingConfig(generateLoggingConfig())),
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
