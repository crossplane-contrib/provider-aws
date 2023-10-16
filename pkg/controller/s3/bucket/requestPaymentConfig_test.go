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
	payer                   = "Requester"
	_     SubresourceClient = &RequestPaymentConfigurationClient{}
)

func generateRequestPaymentConfig() *v1beta1.PaymentConfiguration {
	return &v1beta1.PaymentConfiguration{
		Payer: payer,
	}
}

func TestRequestPaymentObserve(t *testing.T) {
	type args struct {
		cl *RequestPaymentConfigurationClient
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
				b: s3testing.Bucket(s3testing.WithPayerConfig(generateRequestPaymentConfig())),
				cl: NewRequestPaymentConfigurationClient(fake.MockBucketClient{
					MockGetBucketRequestPayment: func(ctx context.Context, input *s3.GetBucketRequestPaymentInput, opts []func(*s3.Options)) (*s3.GetBucketRequestPaymentOutput, error) {
						return nil, errBoom
					},
				}),
			},
			want: want{
				status: NeedsUpdate,
				err:    errorutils.Wrap(errBoom, paymentGetFailed),
			},
		},
		"UpdateNeeded": {
			args: args{
				b: s3testing.Bucket(s3testing.WithPayerConfig(generateRequestPaymentConfig())),
				cl: NewRequestPaymentConfigurationClient(fake.MockBucketClient{
					MockGetBucketRequestPayment: func(ctx context.Context, input *s3.GetBucketRequestPaymentInput, opts []func(*s3.Options)) (*s3.GetBucketRequestPaymentOutput, error) {
						return &s3.GetBucketRequestPaymentOutput{}, nil
					},
				}),
			},
			want: want{
				status: NeedsUpdate,
				err:    nil,
			},
		},
		"UpdateNeededPayer": {
			args: args{
				b: s3testing.Bucket(s3testing.WithPayerConfig(generateRequestPaymentConfig())),
				cl: NewRequestPaymentConfigurationClient(fake.MockBucketClient{
					MockGetBucketRequestPayment: func(ctx context.Context, input *s3.GetBucketRequestPaymentInput, opts []func(*s3.Options)) (*s3.GetBucketRequestPaymentOutput, error) {
						return &s3.GetBucketRequestPaymentOutput{Payer: s3types.PayerBucketOwner}, nil
					},
				}),
			},
			want: want{
				status: NeedsUpdate,
				err:    nil,
			},
		},
		"NoUpdateExists": {
			args: args{
				b: s3testing.Bucket(s3testing.WithPayerConfig(generateRequestPaymentConfig())),
				cl: NewRequestPaymentConfigurationClient(fake.MockBucketClient{
					MockGetBucketRequestPayment: func(ctx context.Context, input *s3.GetBucketRequestPaymentInput, opts []func(*s3.Options)) (*s3.GetBucketRequestPaymentOutput, error) {
						return &s3.GetBucketRequestPaymentOutput{Payer: s3types.PayerRequester}, nil
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

func TestRequestPaymentCreateOrUpdate(t *testing.T) {
	type args struct {
		cl *RequestPaymentConfigurationClient
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
				b: s3testing.Bucket(s3testing.WithPayerConfig(generateRequestPaymentConfig())),
				cl: NewRequestPaymentConfigurationClient(fake.MockBucketClient{
					MockPutBucketRequestPayment: func(ctx context.Context, input *s3.PutBucketRequestPaymentInput, opts []func(*s3.Options)) (*s3.PutBucketRequestPaymentOutput, error) {
						return nil, errBoom
					},
				}),
			},
			want: want{
				err: errorutils.Wrap(errBoom, paymentPutFailed),
			},
		},
		"InvalidConfig": {
			args: args{
				b: s3testing.Bucket(s3testing.WithPayerConfig(generateRequestPaymentConfig())),
				cl: NewRequestPaymentConfigurationClient(fake.MockBucketClient{
					MockPutBucketRequestPayment: func(ctx context.Context, input *s3.PutBucketRequestPaymentInput, opts []func(*s3.Options)) (*s3.PutBucketRequestPaymentOutput, error) {
						return &s3.PutBucketRequestPaymentOutput{}, nil
					},
				}),
			},
			want: want{
				err: nil,
			},
		},
		"SuccessfulCreate": {
			args: args{
				b: s3testing.Bucket(s3testing.WithPayerConfig(generateRequestPaymentConfig())),
				cl: NewRequestPaymentConfigurationClient(fake.MockBucketClient{
					MockPutBucketRequestPayment: func(ctx context.Context, input *s3.PutBucketRequestPaymentInput, opts []func(*s3.Options)) (*s3.PutBucketRequestPaymentOutput, error) {
						return &s3.PutBucketRequestPaymentOutput{}, nil
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

func TestReqPaymentLateInit(t *testing.T) {
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
				cl: NewRequestPaymentConfigurationClient(fake.MockBucketClient{
					MockGetBucketRequestPayment: func(ctx context.Context, input *s3.GetBucketRequestPaymentInput, opts []func(*s3.Options)) (*s3.GetBucketRequestPaymentOutput, error) {
						return &s3.GetBucketRequestPaymentOutput{}, errBoom
					},
				}),
			},
			want: want{
				err: errorutils.Wrap(errBoom, paymentGetFailed),
				cr:  s3testing.Bucket(),
			},
		},
		"NoLateInitEmpty": {
			args: args{
				b: s3testing.Bucket(),
				cl: NewRequestPaymentConfigurationClient(fake.MockBucketClient{
					MockGetBucketRequestPayment: func(ctx context.Context, input *s3.GetBucketRequestPaymentInput, opts []func(*s3.Options)) (*s3.GetBucketRequestPaymentOutput, error) {
						return &s3.GetBucketRequestPaymentOutput{Payer: ""}, nil
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
				b: s3testing.Bucket(s3testing.WithPayerConfig(nil)),
				cl: NewRequestPaymentConfigurationClient(fake.MockBucketClient{
					MockGetBucketRequestPayment: func(ctx context.Context, input *s3.GetBucketRequestPaymentInput, opts []func(*s3.Options)) (*s3.GetBucketRequestPaymentOutput, error) {
						return &s3.GetBucketRequestPaymentOutput{Payer: s3types.PayerRequester}, nil
					},
				}),
			},
			want: want{
				err: nil,
				cr:  s3testing.Bucket(s3testing.WithPayerConfig(generateRequestPaymentConfig())),
			},
		},
		"NoOpLateInit": {
			args: args{
				b: s3testing.Bucket(s3testing.WithPayerConfig(generateRequestPaymentConfig())),
				cl: NewRequestPaymentConfigurationClient(fake.MockBucketClient{
					MockGetBucketRequestPayment: func(ctx context.Context, input *s3.GetBucketRequestPaymentInput, opts []func(*s3.Options)) (*s3.GetBucketRequestPaymentOutput, error) {
						return &s3.GetBucketRequestPaymentOutput{Payer: s3types.PayerBucketOwner}, nil
					},
				}),
			},
			want: want{
				err: nil,
				cr:  s3testing.Bucket(s3testing.WithPayerConfig(generateRequestPaymentConfig())),
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
