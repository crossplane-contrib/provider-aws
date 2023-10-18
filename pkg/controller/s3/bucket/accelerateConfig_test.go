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
	clientss3 "github.com/crossplane-contrib/provider-aws/pkg/clients/s3"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/s3/fake"
	s3testing "github.com/crossplane-contrib/provider-aws/pkg/controller/s3/testing"
	errorutils "github.com/crossplane-contrib/provider-aws/pkg/utils/errors"
)

const (
	enabled   = "Enabled"
	suspended = "Suspended"
)

var _ SubresourceClient = &AccelerateConfigurationClient{}

func TestAccelerateObserve(t *testing.T) {
	type args struct {
		cl *AccelerateConfigurationClient
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
				b: s3testing.Bucket(s3testing.WithAccelerationConfig(&v1beta1.AccelerateConfiguration{Status: enabled})),
				cl: NewAccelerateConfigurationClient(fake.MockBucketClient{
					MockGetBucketAccelerateConfiguration: func(ctx context.Context, input *s3.GetBucketAccelerateConfigurationInput, opts []func(*s3.Options)) (*s3.GetBucketAccelerateConfigurationOutput, error) {
						return nil, errBoom
					},
				}),
			},
			want: want{
				status: NeedsUpdate,
				err:    errorutils.Wrap(errBoom, accelGetFailed),
			},
		},
		"UpdateNeeded": {
			args: args{
				b: s3testing.Bucket(s3testing.WithAccelerationConfig(&v1beta1.AccelerateConfiguration{Status: enabled})),
				cl: NewAccelerateConfigurationClient(fake.MockBucketClient{
					MockGetBucketAccelerateConfiguration: func(ctx context.Context, input *s3.GetBucketAccelerateConfigurationInput, opts []func(*s3.Options)) (*s3.GetBucketAccelerateConfigurationOutput, error) {
						return &s3.GetBucketAccelerateConfigurationOutput{Status: s3types.BucketAccelerateStatusSuspended}, nil
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
				b: s3testing.Bucket(s3testing.WithAccelerationConfig(nil)),
				cl: NewAccelerateConfigurationClient(fake.MockBucketClient{
					MockGetBucketAccelerateConfiguration: func(ctx context.Context, input *s3.GetBucketAccelerateConfigurationInput, opts []func(*s3.Options)) (*s3.GetBucketAccelerateConfigurationOutput, error) {
						return &s3.GetBucketAccelerateConfigurationOutput{}, nil
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
				b: s3testing.Bucket(s3testing.WithAccelerationConfig(&v1beta1.AccelerateConfiguration{Status: suspended})),
				cl: NewAccelerateConfigurationClient(fake.MockBucketClient{
					MockGetBucketAccelerateConfiguration: func(ctx context.Context, input *s3.GetBucketAccelerateConfigurationInput, opts []func(*s3.Options)) (*s3.GetBucketAccelerateConfigurationOutput, error) {
						return &s3.GetBucketAccelerateConfigurationOutput{Status: s3types.BucketAccelerateStatusSuspended}, nil
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

func TestAccelerateCreateOrUpdate(t *testing.T) {
	type args struct {
		cl SubresourceClient
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
				b: s3testing.Bucket(s3testing.WithAccelerationConfig(&v1beta1.AccelerateConfiguration{Status: enabled})),
				cl: NewAccelerateConfigurationClient(fake.MockBucketClient{
					MockPutBucketAccelerateConfiguration: func(ctx context.Context, input *s3.PutBucketAccelerateConfigurationInput, opts []func(*s3.Options)) (*s3.PutBucketAccelerateConfigurationOutput, error) {
						return nil, errBoom
					},
				}),
			},
			want: want{
				err: errorutils.Wrap(errBoom, accelPutFailed),
			},
		},
		"InvalidConfig": {
			args: args{
				b: s3testing.Bucket(s3testing.WithAccelerationConfig(nil)),
				cl: NewAccelerateConfigurationClient(fake.MockBucketClient{
					MockPutBucketAccelerateConfiguration: func(ctx context.Context, input *s3.PutBucketAccelerateConfigurationInput, opts []func(*s3.Options)) (*s3.PutBucketAccelerateConfigurationOutput, error) {
						return &s3.PutBucketAccelerateConfigurationOutput{}, nil
					},
				}),
			},
			want: want{
				err: nil,
			},
		},
		"SuccessfulCreate": {
			args: args{
				b: s3testing.Bucket(s3testing.WithAccelerationConfig(&v1beta1.AccelerateConfiguration{Status: enabled})),
				cl: NewAccelerateConfigurationClient(fake.MockBucketClient{
					MockPutBucketAccelerateConfiguration: func(ctx context.Context, input *s3.PutBucketAccelerateConfigurationInput, opts []func(*s3.Options)) (*s3.PutBucketAccelerateConfigurationOutput, error) {
						return &s3.PutBucketAccelerateConfigurationOutput{}, nil
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

func TestAccelLateInit(t *testing.T) {
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
				cl: NewAccelerateConfigurationClient(fake.MockBucketClient{
					MockGetBucketAccelerateConfiguration: func(ctx context.Context, input *s3.GetBucketAccelerateConfigurationInput, opts []func(*s3.Options)) (*s3.GetBucketAccelerateConfigurationOutput, error) {
						return nil, errBoom
					},
				}),
			},
			want: want{
				err: errorutils.Wrap(errBoom, accelGetFailed),
				cr:  s3testing.Bucket(),
			},
		},
		"ErrorMethodNotAllowedShortStopAndReturnNil": {
			args: args{
				b: s3testing.Bucket(),
				cl: NewAccelerateConfigurationClient(fake.MockBucketClient{
					MockGetBucketAccelerateConfiguration: func(ctx context.Context, input *s3.GetBucketAccelerateConfigurationInput, opts []func(*s3.Options)) (*s3.GetBucketAccelerateConfigurationOutput, error) {
						return nil, &smithy.GenericAPIError{Code: clientss3.MethodNotAllowed}
					},
				}),
			},
			want: want{
				err: nil,
				cr:  s3testing.Bucket(),
			},
		},
		"ErrorArgumentNotSupportedShortStopAndReturnNil": {
			args: args{
				b: s3testing.Bucket(),
				cl: NewAccelerateConfigurationClient(fake.MockBucketClient{
					MockGetBucketAccelerateConfiguration: func(ctx context.Context, input *s3.GetBucketAccelerateConfigurationInput, opts []func(*s3.Options)) (*s3.GetBucketAccelerateConfigurationOutput, error) {
						return nil, &smithy.GenericAPIError{Code: clientss3.UnsupportedArgument}
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
				cl: NewAccelerateConfigurationClient(fake.MockBucketClient{
					MockGetBucketAccelerateConfiguration: func(ctx context.Context, input *s3.GetBucketAccelerateConfigurationInput, opts []func(*s3.Options)) (*s3.GetBucketAccelerateConfigurationOutput, error) {
						return &s3.GetBucketAccelerateConfigurationOutput{Status: ""}, nil
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
				b: s3testing.Bucket(s3testing.WithAccelerationConfig(nil)),
				cl: NewAccelerateConfigurationClient(fake.MockBucketClient{
					MockGetBucketAccelerateConfiguration: func(ctx context.Context, input *s3.GetBucketAccelerateConfigurationInput, opts []func(*s3.Options)) (*s3.GetBucketAccelerateConfigurationOutput, error) {
						return &s3.GetBucketAccelerateConfigurationOutput{Status: enabled}, nil
					},
				}),
			},
			want: want{
				err: nil,
				cr:  s3testing.Bucket(s3testing.WithAccelerationConfig(&v1beta1.AccelerateConfiguration{Status: enabled})),
			},
		},
		"NoOpLateInit": {
			args: args{
				b: s3testing.Bucket(s3testing.WithAccelerationConfig(&v1beta1.AccelerateConfiguration{Status: enabled})),
				cl: NewAccelerateConfigurationClient(fake.MockBucketClient{
					MockGetBucketAccelerateConfiguration: func(ctx context.Context, input *s3.GetBucketAccelerateConfigurationInput, opts []func(*s3.Options)) (*s3.GetBucketAccelerateConfigurationOutput, error) {
						return &s3.GetBucketAccelerateConfigurationOutput{Status: suspended}, nil
					},
				}),
			},
			want: want{
				err: nil,
				cr:  s3testing.Bucket(s3testing.WithAccelerationConfig(&v1beta1.AccelerateConfiguration{Status: enabled})),
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
