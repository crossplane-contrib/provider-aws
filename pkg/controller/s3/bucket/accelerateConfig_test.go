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

	_ "github.com/aws/aws-sdk-go-v2/aws/awserr"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"

	"github.com/crossplane/provider-aws/apis/s3/v1beta1"
	"github.com/crossplane/provider-aws/pkg/clients/s3/fake"
	s3Testing "github.com/crossplane/provider-aws/pkg/controller/s3/testing"
)

const (
	enabled   = "Enabled"
	suspended = "Suspended"
)

var errBoom = errors.New("boom")

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
				b: s3Testing.Bucket(s3Testing.WithAccelerationConfig(&v1beta1.AccelerateConfiguration{Status: enabled})),
				cl: NewAccelerateConfigurationClient(fake.MockBucketClient{
					MockGetBucketAccelerateConfigurationRequest: func(input *s3.GetBucketAccelerateConfigurationInput) s3.GetBucketAccelerateConfigurationRequest {
						return s3.GetBucketAccelerateConfigurationRequest{
							Request: s3Testing.CreateRequest(errBoom, &s3.GetBucketAccelerateConfigurationOutput{}),
						}
					},
				}),
			},
			want: want{
				status: NeedsUpdate,
				err:    errors.Wrap(errBoom, accelGetFailed),
			},
		},
		"UpdateNeeded": {
			args: args{
				b: s3Testing.Bucket(s3Testing.WithAccelerationConfig(&v1beta1.AccelerateConfiguration{Status: enabled})),
				cl: NewAccelerateConfigurationClient(fake.MockBucketClient{
					MockGetBucketAccelerateConfigurationRequest: func(input *s3.GetBucketAccelerateConfigurationInput) s3.GetBucketAccelerateConfigurationRequest {
						return s3.GetBucketAccelerateConfigurationRequest{
							Request: s3Testing.CreateRequest(nil, &s3.GetBucketAccelerateConfigurationOutput{Status: s3.BucketAccelerateStatusSuspended}),
						}
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
				b: s3Testing.Bucket(s3Testing.WithAccelerationConfig(nil)),
				cl: NewAccelerateConfigurationClient(fake.MockBucketClient{
					MockGetBucketAccelerateConfigurationRequest: func(input *s3.GetBucketAccelerateConfigurationInput) s3.GetBucketAccelerateConfigurationRequest {
						return s3.GetBucketAccelerateConfigurationRequest{
							Request: s3Testing.CreateRequest(nil, &s3.GetBucketAccelerateConfigurationOutput{}),
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
				b: s3Testing.Bucket(s3Testing.WithAccelerationConfig(&v1beta1.AccelerateConfiguration{Status: suspended})),
				cl: NewAccelerateConfigurationClient(fake.MockBucketClient{
					MockGetBucketAccelerateConfigurationRequest: func(input *s3.GetBucketAccelerateConfigurationInput) s3.GetBucketAccelerateConfigurationRequest {
						return s3.GetBucketAccelerateConfigurationRequest{
							Request: s3Testing.CreateRequest(nil, &s3.GetBucketAccelerateConfigurationOutput{Status: s3.BucketAccelerateStatusSuspended}),
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
				b: s3Testing.Bucket(s3Testing.WithAccelerationConfig(&v1beta1.AccelerateConfiguration{Status: enabled})),
				cl: NewAccelerateConfigurationClient(fake.MockBucketClient{
					MockPutBucketAccelerateConfigurationRequest: func(input *s3.PutBucketAccelerateConfigurationInput) s3.PutBucketAccelerateConfigurationRequest {
						return s3.PutBucketAccelerateConfigurationRequest{
							Request: s3Testing.CreateRequest(errBoom, &s3.PutBucketAccelerateConfigurationOutput{}),
						}
					},
				}),
			},
			want: want{
				err: errors.Wrap(errBoom, accelPutFailed),
			},
		},
		"InvalidConfig": {
			args: args{
				b: s3Testing.Bucket(s3Testing.WithAccelerationConfig(nil)),
				cl: NewAccelerateConfigurationClient(fake.MockBucketClient{
					MockPutBucketAccelerateConfigurationRequest: func(input *s3.PutBucketAccelerateConfigurationInput) s3.PutBucketAccelerateConfigurationRequest {
						return s3.PutBucketAccelerateConfigurationRequest{
							Request: s3Testing.CreateRequest(nil, &s3.PutBucketAccelerateConfigurationOutput{}),
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
				b: s3Testing.Bucket(s3Testing.WithAccelerationConfig(&v1beta1.AccelerateConfiguration{Status: enabled})),
				cl: NewAccelerateConfigurationClient(fake.MockBucketClient{
					MockPutBucketAccelerateConfigurationRequest: func(input *s3.PutBucketAccelerateConfigurationInput) s3.PutBucketAccelerateConfigurationRequest {
						return s3.PutBucketAccelerateConfigurationRequest{
							Request: s3Testing.CreateRequest(nil, &s3.PutBucketAccelerateConfigurationOutput{}),
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
