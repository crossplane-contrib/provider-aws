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
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"

	"github.com/crossplane/provider-aws/apis/s3/v1beta1"
	"github.com/crossplane/provider-aws/pkg/clients/s3/fake"
	s3Testing "github.com/crossplane/provider-aws/pkg/controller/s3/testing"
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

func generateAWSPayment() *s3.RequestPaymentConfiguration {
	return &s3.RequestPaymentConfiguration{
		Payer: s3.PayerRequester,
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
				b: s3Testing.Bucket(s3Testing.WithPayerConfig(generateRequestPaymentConfig())),
				cl: NewRequestPaymentConfigurationClient(fake.MockBucketClient{
					MockGetBucketRequestPaymentRequest: func(input *s3.GetBucketRequestPaymentInput) s3.GetBucketRequestPaymentRequest {
						return s3.GetBucketRequestPaymentRequest{
							Request: s3Testing.CreateRequest(errBoom, &s3.GetBucketRequestPaymentOutput{Payer: generateAWSPayment().Payer}),
						}
					},
				}),
			},
			want: want{
				status: NeedsUpdate,
				err:    errors.Wrap(errBoom, paymentGetFailed),
			},
		},
		"UpdateNeeded": {
			args: args{
				b: s3Testing.Bucket(s3Testing.WithPayerConfig(generateRequestPaymentConfig())),
				cl: NewRequestPaymentConfigurationClient(fake.MockBucketClient{
					MockGetBucketRequestPaymentRequest: func(input *s3.GetBucketRequestPaymentInput) s3.GetBucketRequestPaymentRequest {
						return s3.GetBucketRequestPaymentRequest{
							Request: s3Testing.CreateRequest(nil, &s3.GetBucketRequestPaymentOutput{}),
						}
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
				b: s3Testing.Bucket(s3Testing.WithPayerConfig(generateRequestPaymentConfig())),
				cl: NewRequestPaymentConfigurationClient(fake.MockBucketClient{
					MockGetBucketRequestPaymentRequest: func(input *s3.GetBucketRequestPaymentInput) s3.GetBucketRequestPaymentRequest {
						return s3.GetBucketRequestPaymentRequest{
							Request: s3Testing.CreateRequest(nil, &s3.GetBucketRequestPaymentOutput{Payer: generateAWSPayment().Payer}),
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
				b: s3Testing.Bucket(s3Testing.WithPayerConfig(generateRequestPaymentConfig())),
				cl: NewRequestPaymentConfigurationClient(fake.MockBucketClient{
					MockPutBucketRequestPaymentRequest: func(input *s3.PutBucketRequestPaymentInput) s3.PutBucketRequestPaymentRequest {
						return s3.PutBucketRequestPaymentRequest{
							Request: s3Testing.CreateRequest(errBoom, &s3.PutBucketRequestPaymentOutput{}),
						}
					},
				}),
			},
			want: want{
				err: errors.Wrap(errBoom, paymentPutFailed),
			},
		},
		"InvalidConfig": {
			args: args{
				b: s3Testing.Bucket(s3Testing.WithPayerConfig(generateRequestPaymentConfig())),
				cl: NewRequestPaymentConfigurationClient(fake.MockBucketClient{
					MockPutBucketRequestPaymentRequest: func(input *s3.PutBucketRequestPaymentInput) s3.PutBucketRequestPaymentRequest {
						return s3.PutBucketRequestPaymentRequest{
							Request: s3Testing.CreateRequest(nil, &s3.PutBucketRequestPaymentOutput{}),
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
				b: s3Testing.Bucket(s3Testing.WithPayerConfig(generateRequestPaymentConfig())),
				cl: NewRequestPaymentConfigurationClient(fake.MockBucketClient{
					MockPutBucketRequestPaymentRequest: func(input *s3.PutBucketRequestPaymentInput) s3.PutBucketRequestPaymentRequest {
						return s3.PutBucketRequestPaymentRequest{
							Request: s3Testing.CreateRequest(nil, &s3.PutBucketRequestPaymentOutput{}),
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
			_, err := tc.args.cl.CreateOrUpdate(context.Background(), tc.args.b)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}
