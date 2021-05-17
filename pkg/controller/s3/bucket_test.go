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

package s3

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/awserr"
	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	k8serrors "k8s.io/apimachinery/pkg/util/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/test"

	"github.com/crossplane/provider-aws/apis/s3/v1beta1"
	awsclient "github.com/crossplane/provider-aws/pkg/clients"
	"github.com/crossplane/provider-aws/pkg/clients/s3"
	"github.com/crossplane/provider-aws/pkg/clients/s3/fake"
	"github.com/crossplane/provider-aws/pkg/controller/s3/bucket"
	s3Testing "github.com/crossplane/provider-aws/pkg/controller/s3/testing"
)

var (
	// an arbitrary managed resource
	unexpectedItem resource.Managed
	errBoom        = errors.New("boom")
)

type args struct {
	s3   s3.BucketClient
	kube client.Client
	cr   resource.Managed
}

func TestObserve(t *testing.T) {

	type want struct {
		cr     resource.Managed
		result managed.ExternalObservation
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"InValidInput": {
			args: args{
				cr: unexpectedItem,
			},
			want: want{
				cr:  unexpectedItem,
				err: errors.New(errUnexpectedObject),
			},
		},
		"ClientError": {
			args: args{
				s3: &fake.MockBucketClient{
					MockHeadBucketRequest: func(input *awss3.HeadBucketInput) awss3.HeadBucketRequest {
						return awss3.HeadBucketRequest{
							Request: s3Testing.CreateRequest(errBoom, nil),
						}
					},
				},
				cr: s3Testing.Bucket(),
			},
			want: want{
				cr:  s3Testing.Bucket(),
				err: awsclient.Wrap(errBoom, errHead),
			},
		},
		"ResourceDoesNotExist": {
			args: args{
				s3: &fake.MockBucketClient{
					MockHeadBucketRequest: func(input *awss3.HeadBucketInput) awss3.HeadBucketRequest {
						return awss3.HeadBucketRequest{
							Request: s3Testing.CreateRequest(awserr.New(s3.BucketNotFoundErrCode, "", nil), nil),
						}
					},
				},
				cr: s3Testing.Bucket(),
			},
			want: want{
				cr: s3Testing.Bucket(),
			},
		},
		"ValidInputNoLateInitialize": {
			args: args{
				s3: s3Testing.Client(),
				cr: s3Testing.Bucket(),
			},
			want: want{
				cr: s3Testing.Bucket(
					s3Testing.WithConditions(xpv1.Available()),
					s3Testing.WithArn(fmt.Sprintf("arn:aws:s3:::%s", s3Testing.BucketName)),
				),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
					ConnectionDetails: map[string][]byte{
						xpv1.ResourceCredentialsSecretEndpointKey:  []byte(s3Testing.BucketName),
						v1beta1.ResourceCredentialsSecretRegionKey: []byte(s3Testing.Region),
					},
				},
			},
		},
		"ValidInputNoLateInitializeUpdateACLFail": {
			args: args{
				s3: s3Testing.Client(s3Testing.WithPutACL(func(input *awss3.PutBucketAclInput) awss3.PutBucketAclRequest {
					return awss3.PutBucketAclRequest{
						Request: s3Testing.CreateRequest(errBoom, &awss3.PutBucketAclOutput{}),
					}
				})),
				cr: s3Testing.Bucket(),
			},
			want: want{
				cr: s3Testing.Bucket(
					s3Testing.WithArn(fmt.Sprintf("arn:aws:s3:::%s", s3Testing.BucketName)),
				),
				err:    errBoom,
				result: managed.ExternalObservation{},
			},
		},
		"LateInitialize": {
			args: args{
				s3: s3Testing.Client(
					s3Testing.WithGetRequestPayment(func(input *awss3.GetBucketRequestPaymentInput) awss3.GetBucketRequestPaymentRequest {
						return awss3.GetBucketRequestPaymentRequest{
							Request: s3Testing.CreateRequest(nil, &awss3.GetBucketRequestPaymentOutput{Payer: awss3.PayerRequester}),
						}
					}),
				),
				cr: s3Testing.Bucket(s3Testing.WithPayerConfig(&v1beta1.PaymentConfiguration{})),
			},
			want: want{
				cr: s3Testing.Bucket(
					s3Testing.WithConditions(xpv1.Available()),
					s3Testing.WithArn(fmt.Sprintf("arn:aws:s3:::%s", s3Testing.BucketName)),
					s3Testing.WithPayerConfig(&v1beta1.PaymentConfiguration{Payer: "Requester"}),
				),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
					ConnectionDetails: map[string][]byte{
						xpv1.ResourceCredentialsSecretEndpointKey:  []byte(s3Testing.BucketName),
						v1beta1.ResourceCredentialsSecretRegionKey: []byte(s3Testing.Region),
					},
					ResourceLateInitialized: true,
				},
			},
		},
		"LateInitializeNotOccurNil": {
			// this case is the same as needing an update, we should not late init here.
			args: args{
				s3: s3Testing.Client(
					s3Testing.WithGetRequestPayment(func(input *awss3.GetBucketRequestPaymentInput) awss3.GetBucketRequestPaymentRequest {
						return awss3.GetBucketRequestPaymentRequest{
							Request: s3Testing.CreateRequest(nil, &awss3.GetBucketRequestPaymentOutput{Payer: awss3.PayerRequester}),
						}
					}),
				),
				cr: s3Testing.Bucket(),
			},
			want: want{
				cr: s3Testing.Bucket(
					s3Testing.WithArn(fmt.Sprintf("arn:aws:s3:::%s", s3Testing.BucketName)),
				),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: false,
				},
			},
		},
		"LateInitializeNotOccurExistsList": {
			// Validating that late init should not occur here because SSE already exists.
			args: args{
				s3: s3Testing.Client(
					s3Testing.WithGetSSE(func(input *awss3.GetBucketEncryptionInput) awss3.GetBucketEncryptionRequest {
						return awss3.GetBucketEncryptionRequest{
							Request: s3Testing.CreateRequest(nil, &awss3.GetBucketEncryptionOutput{
								ServerSideEncryptionConfiguration: &awss3.ServerSideEncryptionConfiguration{
									Rules: []awss3.ServerSideEncryptionRule{
										{
											ApplyServerSideEncryptionByDefault: &awss3.ServerSideEncryptionByDefault{
												KMSMasterKeyID: aws.String("1234567890"),
												SSEAlgorithm:   awss3.ServerSideEncryptionAwsKms,
											},
										},
									},
								},
							}),
						}
					}),
				),
				cr: s3Testing.Bucket(
					s3Testing.WithSSEConfig(&v1beta1.ServerSideEncryptionConfiguration{
						Rules: []v1beta1.ServerSideEncryptionRule{
							{
								ApplyServerSideEncryptionByDefault: v1beta1.ServerSideEncryptionByDefault{
									KMSMasterKeyID: aws.String("test"),
									SSEAlgorithm:   "AES256",
								},
							},
						},
					}),
				),
			},
			want: want{
				cr: s3Testing.Bucket(
					s3Testing.WithArn(fmt.Sprintf("arn:aws:s3:::%s", s3Testing.BucketName)),
					s3Testing.WithSSEConfig(&v1beta1.ServerSideEncryptionConfiguration{
						Rules: []v1beta1.ServerSideEncryptionRule{
							{
								ApplyServerSideEncryptionByDefault: v1beta1.ServerSideEncryptionByDefault{
									KMSMasterKeyID: aws.String("test"),
									SSEAlgorithm:   "AES256",
								},
							},
						},
					}),
				),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: false,
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{s3client: tc.s3, subresourceClients: bucket.NewSubresourceClients(tc.s3), kube: tc.kube}
			o, err := e.Observe(context.Background(), tc.args.cr)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.cr, tc.args.cr, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.result, o); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestCreate(t *testing.T) {

	type want struct {
		cr     resource.Managed
		result managed.ExternalCreation
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"VaildInput": {
			args: args{
				kube: &test.MockClient{
					MockUpdate: test.NewMockUpdateFn(nil),
				},
				s3: s3Testing.Client(),
				cr: s3Testing.Bucket(),
			},
			want: want{
				cr: s3Testing.Bucket(s3Testing.WithConditions(xpv1.Creating())),
			},
		},
		"InValidInput": {
			args: args{
				kube: &test.MockClient{
					MockUpdate: test.NewMockUpdateFn(nil),
				},
				cr: unexpectedItem,
			},
			want: want{
				cr:  unexpectedItem,
				err: errors.New(errUnexpectedObject),
			},
		},
		"ClientError": {
			args: args{
				kube: &test.MockClient{
					MockUpdate: test.NewMockUpdateFn(nil),
				},
				s3: s3Testing.Client(s3Testing.WithCreateBucket(func(input *awss3.CreateBucketInput) awss3.CreateBucketRequest {
					return awss3.CreateBucketRequest{
						Request: s3Testing.CreateRequest(errBoom, &awss3.CreateBucketOutput{}),
					}
				})),
				cr: s3Testing.Bucket(),
			},
			want: want{
				cr:  s3Testing.Bucket(s3Testing.WithConditions(xpv1.Creating())),
				err: awsclient.Wrap(errBoom, errCreate),
			},
		},
		"ValidInputLateInitialize": {
			args: args{
				kube: &test.MockClient{
					MockUpdate: test.NewMockUpdateFn(nil),
				},
				s3: s3Testing.Client(
					s3Testing.WithGetRequestPayment(func(input *awss3.GetBucketRequestPaymentInput) awss3.GetBucketRequestPaymentRequest {
						return awss3.GetBucketRequestPaymentRequest{
							Request: s3Testing.CreateRequest(nil, &awss3.GetBucketRequestPaymentOutput{Payer: awss3.PayerRequester}),
						}
					}),
					s3Testing.WithCreateBucket(func(input *awss3.CreateBucketInput) awss3.CreateBucketRequest {
						return awss3.CreateBucketRequest{
							Request: s3Testing.CreateRequest(nil, &awss3.CreateBucketOutput{}),
						}
					}),
				),
				cr: s3Testing.Bucket(),
			},
			want: want{
				cr: s3Testing.Bucket(
					s3Testing.WithConditions(xpv1.Creating()),
					s3Testing.WithPayerConfig(&v1beta1.PaymentConfiguration{Payer: "Requester"}),
				),
				result: managed.ExternalCreation{},
			},
		},
		"LateInitializeFailedErrors": {
			args: args{
				kube: &test.MockClient{
					MockUpdate: test.NewMockUpdateFn(nil),
				},
				s3: s3Testing.Client(
					s3Testing.WithGetRequestPayment(func(input *awss3.GetBucketRequestPaymentInput) awss3.GetBucketRequestPaymentRequest {
						return awss3.GetBucketRequestPaymentRequest{
							Request: s3Testing.CreateRequest(errBoom, &awss3.GetBucketRequestPaymentOutput{Payer: awss3.PayerRequester}),
						}
					}),
					s3Testing.WithGetSSE(func(input *awss3.GetBucketEncryptionInput) awss3.GetBucketEncryptionRequest {
						return awss3.GetBucketEncryptionRequest{
							Request: s3Testing.CreateRequest(errBoom, &awss3.GetBucketEncryptionOutput{}),
						}
					}),
					s3Testing.WithCreateBucket(func(input *awss3.CreateBucketInput) awss3.CreateBucketRequest {
						return awss3.CreateBucketRequest{
							Request: s3Testing.CreateRequest(nil, &awss3.CreateBucketOutput{}),
						}
					}),
				),
				cr: s3Testing.Bucket(),
			},
			want: want{
				cr: s3Testing.Bucket(
					s3Testing.WithConditions(xpv1.Creating()),
				),
				result: managed.ExternalCreation{},
				err:    k8serrors.NewAggregate([]error{awsclient.Wrap(errBoom, "cannot get request payment configuration"), awsclient.Wrap(errBoom, "cannot get encryption configuration")}),
			},
		},
		"ExternalDiffersFromLocalNoLateInit": {
			args: args{
				kube: &test.MockClient{
					MockUpdate: test.NewMockUpdateFn(nil),
				},
				s3: s3Testing.Client(
					s3Testing.WithGetRequestPayment(func(input *awss3.GetBucketRequestPaymentInput) awss3.GetBucketRequestPaymentRequest {
						return awss3.GetBucketRequestPaymentRequest{
							Request: s3Testing.CreateRequest(nil, &awss3.GetBucketRequestPaymentOutput{Payer: awss3.PayerRequester}),
						}
					}),
					s3Testing.WithGetSSE(func(input *awss3.GetBucketEncryptionInput) awss3.GetBucketEncryptionRequest {
						return awss3.GetBucketEncryptionRequest{
							Request: s3Testing.CreateRequest(nil, &awss3.GetBucketEncryptionOutput{
								ServerSideEncryptionConfiguration: &awss3.ServerSideEncryptionConfiguration{
									Rules: []awss3.ServerSideEncryptionRule{
										{
											ApplyServerSideEncryptionByDefault: &awss3.ServerSideEncryptionByDefault{
												KMSMasterKeyID: aws.String("1234567890"),
												SSEAlgorithm:   awss3.ServerSideEncryptionAwsKms,
											},
										},
									},
								},
							}),
						}
					}),
					s3Testing.WithCreateBucket(func(input *awss3.CreateBucketInput) awss3.CreateBucketRequest {
						return awss3.CreateBucketRequest{
							Request: s3Testing.CreateRequest(nil, &awss3.CreateBucketOutput{}),
						}
					}),
				),
				cr: s3Testing.Bucket(s3Testing.WithSSEConfig(&v1beta1.ServerSideEncryptionConfiguration{
					Rules: []v1beta1.ServerSideEncryptionRule{
						{
							ApplyServerSideEncryptionByDefault: v1beta1.ServerSideEncryptionByDefault{
								KMSMasterKeyID: aws.String("test"),
								SSEAlgorithm:   "AES256",
							},
						},
					},
				})),
			},
			want: want{
				cr: s3Testing.Bucket(
					s3Testing.WithConditions(xpv1.Creating()),
					s3Testing.WithSSEConfig(&v1beta1.ServerSideEncryptionConfiguration{
						Rules: []v1beta1.ServerSideEncryptionRule{
							{
								ApplyServerSideEncryptionByDefault: v1beta1.ServerSideEncryptionByDefault{
									KMSMasterKeyID: aws.String("test"),
									SSEAlgorithm:   "AES256",
								},
							},
						},
					}),
					s3Testing.WithPayerConfig(&v1beta1.PaymentConfiguration{Payer: "Requester"}),
				),
				result: managed.ExternalCreation{},
			},
		},
		"ValidInputLateInitializeKubeErr": {
			args: args{
				kube: &test.MockClient{
					MockUpdate: test.NewMockUpdateFn(errBoom),
				},
				s3: s3Testing.Client(
					s3Testing.WithGetRequestPayment(func(input *awss3.GetBucketRequestPaymentInput) awss3.GetBucketRequestPaymentRequest {
						return awss3.GetBucketRequestPaymentRequest{
							Request: s3Testing.CreateRequest(nil, &awss3.GetBucketRequestPaymentOutput{Payer: awss3.PayerRequester}),
						}
					}),
					s3Testing.WithCreateBucket(func(input *awss3.CreateBucketInput) awss3.CreateBucketRequest {
						return awss3.CreateBucketRequest{
							Request: s3Testing.CreateRequest(nil, &awss3.CreateBucketOutput{}),
						}
					}),
				),
				cr: s3Testing.Bucket(),
			},
			want: want{
				cr: s3Testing.Bucket(
					s3Testing.WithPayerConfig(&v1beta1.PaymentConfiguration{Payer: "Requester"}),
					s3Testing.WithConditions(xpv1.Creating()),
				),
				result: managed.ExternalCreation{},
				err:    k8serrors.NewAggregate([]error{awsclient.Wrap(errBoom, errKubeUpdateFailed)}),
			},
		},
	}

	for name, tc := range cases {
		noop := logging.NewNopLogger()
		t.Run(name, func(t *testing.T) {
			e := &external{s3client: tc.s3, kube: tc.kube, logger: noop, subresourceClients: bucket.NewSubresourceClients(tc.s3)}
			o, err := e.Create(context.Background(), tc.args.cr)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.cr, tc.args.cr, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.result, o); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestUpdate(t *testing.T) {

	type want struct {
		cr     resource.Managed
		result managed.ExternalUpdate
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"InValidInput": {
			args: args{
				cr: unexpectedItem,
			},
			want: want{
				cr:  unexpectedItem,
				err: errors.New(errUnexpectedObject),
			},
		},
		"ValidInputNoUpdate": {
			args: args{
				s3: s3Testing.Client(),
				cr: s3Testing.Bucket(),
			},
			want: want{
				cr:     s3Testing.Bucket(),
				result: managed.ExternalUpdate{},
			},
		},
		"ValidInputUpdateNeededSuccess": {
			args: args{
				s3: s3Testing.Client(
					s3Testing.WithGetRequestPayment(func(input *awss3.GetBucketRequestPaymentInput) awss3.GetBucketRequestPaymentRequest {
						return awss3.GetBucketRequestPaymentRequest{
							Request: s3Testing.CreateRequest(nil, &awss3.GetBucketRequestPaymentOutput{Payer: awss3.PayerBucketOwner}),
						}
					}),
					s3Testing.WithPutRequestPayment(func(input *awss3.PutBucketRequestPaymentInput) awss3.PutBucketRequestPaymentRequest {
						return awss3.PutBucketRequestPaymentRequest{
							Request: s3Testing.CreateRequest(nil, &awss3.PutBucketRequestPaymentOutput{}),
						}
					}),
				),
				cr: s3Testing.Bucket(s3Testing.WithPayerConfig(&v1beta1.PaymentConfiguration{Payer: "Requester"})),
			},
			want: want{
				cr: s3Testing.Bucket(
					s3Testing.WithConditions(),
					s3Testing.WithPayerConfig(&v1beta1.PaymentConfiguration{Payer: "Requester"}),
				),
				result: managed.ExternalUpdate{},
			},
		},
		"ValidInputUpdateNeededFailed": {
			args: args{
				s3: s3Testing.Client(
					s3Testing.WithGetRequestPayment(func(input *awss3.GetBucketRequestPaymentInput) awss3.GetBucketRequestPaymentRequest {
						return awss3.GetBucketRequestPaymentRequest{
							Request: s3Testing.CreateRequest(nil, &awss3.GetBucketRequestPaymentOutput{Payer: awss3.PayerBucketOwner}),
						}
					}),
					s3Testing.WithPutRequestPayment(func(input *awss3.PutBucketRequestPaymentInput) awss3.PutBucketRequestPaymentRequest {
						return awss3.PutBucketRequestPaymentRequest{
							Request: s3Testing.CreateRequest(errBoom, &awss3.PutBucketRequestPaymentOutput{}),
						}
					}),
				),
				cr: s3Testing.Bucket(s3Testing.WithPayerConfig(&v1beta1.PaymentConfiguration{Payer: "Requester"})),
			},
			want: want{
				cr: s3Testing.Bucket(
					s3Testing.WithPayerConfig(&v1beta1.PaymentConfiguration{Payer: "Requester"}),
				),
				err:    awsclient.Wrap(awsclient.Wrap(errBoom, "cannot put Bucket payment"), errCreateOrUpdate),
				result: managed.ExternalUpdate{},
			},
		},
		"ValidInputUpdateNeededObserveFailed": {
			args: args{
				s3: s3Testing.Client(
					s3Testing.WithGetRequestPayment(func(input *awss3.GetBucketRequestPaymentInput) awss3.GetBucketRequestPaymentRequest {
						return awss3.GetBucketRequestPaymentRequest{
							Request: s3Testing.CreateRequest(errBoom, &awss3.GetBucketRequestPaymentOutput{Payer: awss3.PayerBucketOwner}),
						}
					}),
					s3Testing.WithPutRequestPayment(func(input *awss3.PutBucketRequestPaymentInput) awss3.PutBucketRequestPaymentRequest {
						return awss3.PutBucketRequestPaymentRequest{
							Request: s3Testing.CreateRequest(nil, &awss3.PutBucketRequestPaymentOutput{}),
						}
					}),
				),
				cr: s3Testing.Bucket(s3Testing.WithPayerConfig(&v1beta1.PaymentConfiguration{Payer: "Requester"})),
			},
			want: want{
				cr: s3Testing.Bucket(
					s3Testing.WithConditions(xpv1.ReconcileError(awsclient.Wrap(errBoom, "cannot get request payment configuration"))),
					s3Testing.WithPayerConfig(&v1beta1.PaymentConfiguration{Payer: "Requester"}),
				),
				err:    awsclient.Wrap(errBoom, "cannot get request payment configuration"),
				result: managed.ExternalUpdate{},
			},
		},
		"ValidInputDeleteNeededSuccess": {
			args: args{
				s3: s3Testing.Client(
					s3Testing.WithDeleteSSE(func(input *awss3.DeleteBucketEncryptionInput) awss3.DeleteBucketEncryptionRequest {
						return awss3.DeleteBucketEncryptionRequest{
							Request: s3Testing.CreateRequest(nil, &awss3.DeleteBucketEncryptionOutput{}),
						}
					}),
					s3Testing.WithGetSSE(func(input *awss3.GetBucketEncryptionInput) awss3.GetBucketEncryptionRequest {
						return awss3.GetBucketEncryptionRequest{
							Request: s3Testing.CreateRequest(nil, &awss3.GetBucketEncryptionOutput{
								ServerSideEncryptionConfiguration: &awss3.ServerSideEncryptionConfiguration{
									Rules: []awss3.ServerSideEncryptionRule{
										{
											ApplyServerSideEncryptionByDefault: &awss3.ServerSideEncryptionByDefault{
												KMSMasterKeyID: aws.String("key-id"),
												SSEAlgorithm:   awss3.ServerSideEncryptionAes256,
											},
										},
									},
								},
							}),
						}
					}),
				),
				cr: s3Testing.Bucket(s3Testing.WithSSEConfig(nil)),
			},
			want: want{
				cr: s3Testing.Bucket(
					s3Testing.WithConditions(),
					s3Testing.WithSSEConfig(nil),
				),
				result: managed.ExternalUpdate{},
			},
		},
		"ValidInputDeleteNeededFailed": {
			args: args{
				s3: s3Testing.Client(
					s3Testing.WithDeleteSSE(func(input *awss3.DeleteBucketEncryptionInput) awss3.DeleteBucketEncryptionRequest {
						return awss3.DeleteBucketEncryptionRequest{
							Request: s3Testing.CreateRequest(errBoom, &awss3.DeleteBucketEncryptionOutput{}),
						}
					}),
					s3Testing.WithGetSSE(func(input *awss3.GetBucketEncryptionInput) awss3.GetBucketEncryptionRequest {
						return awss3.GetBucketEncryptionRequest{
							Request: s3Testing.CreateRequest(nil, &awss3.GetBucketEncryptionOutput{
								ServerSideEncryptionConfiguration: &awss3.ServerSideEncryptionConfiguration{
									Rules: []awss3.ServerSideEncryptionRule{
										{
											ApplyServerSideEncryptionByDefault: &awss3.ServerSideEncryptionByDefault{
												KMSMasterKeyID: aws.String("key-id"),
												SSEAlgorithm:   awss3.ServerSideEncryptionAes256,
											},
										},
									},
								},
							}),
						}
					}),
				),
				cr: s3Testing.Bucket(s3Testing.WithSSEConfig(nil)),
			},
			want: want{
				cr: s3Testing.Bucket(
					s3Testing.WithSSEConfig(nil),
				),
				err:    awsclient.Wrap(awsclient.Wrap(errBoom, "cannot delete encryption configuration"), errDelete),
				result: managed.ExternalUpdate{},
			},
		},
		"ValidInputDeleteNeededObserveFailed": {
			args: args{
				s3: s3Testing.Client(
					s3Testing.WithDeleteSSE(func(input *awss3.DeleteBucketEncryptionInput) awss3.DeleteBucketEncryptionRequest {
						return awss3.DeleteBucketEncryptionRequest{
							Request: s3Testing.CreateRequest(nil, &awss3.DeleteBucketEncryptionOutput{}),
						}
					}),
					s3Testing.WithGetSSE(func(input *awss3.GetBucketEncryptionInput) awss3.GetBucketEncryptionRequest {
						return awss3.GetBucketEncryptionRequest{
							Request: s3Testing.CreateRequest(errBoom, &awss3.GetBucketEncryptionOutput{
								ServerSideEncryptionConfiguration: &awss3.ServerSideEncryptionConfiguration{
									Rules: []awss3.ServerSideEncryptionRule{
										{
											ApplyServerSideEncryptionByDefault: &awss3.ServerSideEncryptionByDefault{
												KMSMasterKeyID: aws.String("key-id"),
												SSEAlgorithm:   awss3.ServerSideEncryptionAes256,
											},
										},
									},
								},
							}),
						}
					}),
				),
				cr: s3Testing.Bucket(s3Testing.WithSSEConfig(nil)),
			},
			want: want{
				cr: s3Testing.Bucket(
					s3Testing.WithConditions(xpv1.ReconcileError(awsclient.Wrap(errBoom, "cannot get encryption configuration"))),
					s3Testing.WithSSEConfig(nil),
				),
				err:    awsclient.Wrap(errBoom, "cannot get encryption configuration"),
				result: managed.ExternalUpdate{},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{s3client: tc.s3, subresourceClients: bucket.NewSubresourceClients(tc.s3)}
			o, err := e.Update(context.Background(), tc.args.cr)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.cr, tc.args.cr, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.result, o); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestDelete(t *testing.T) {

	type want struct {
		cr  resource.Managed
		err error
	}

	cases := map[string]struct {
		args
		want
	}{
		"VaildInput": {
			args: args{
				s3: &fake.MockBucketClient{
					MockDeleteBucketRequest: func(input *awss3.DeleteBucketInput) awss3.DeleteBucketRequest {
						return awss3.DeleteBucketRequest{
							Request: s3Testing.CreateRequest(nil, &awss3.DeleteBucketOutput{}),
						}
					},
				},
				cr: s3Testing.Bucket(),
			},
			want: want{
				cr: s3Testing.Bucket(s3Testing.WithConditions(xpv1.Deleting())),
			},
		},
		"InValidInput": {
			args: args{
				cr: unexpectedItem,
			},
			want: want{
				cr:  unexpectedItem,
				err: errors.New(errUnexpectedObject),
			},
		},
		"ClientError": {
			args: args{
				s3: &fake.MockBucketClient{
					MockDeleteBucketRequest: func(input *awss3.DeleteBucketInput) awss3.DeleteBucketRequest {
						return awss3.DeleteBucketRequest{
							Request: s3Testing.CreateRequest(errBoom, &awss3.DeleteBucketOutput{}),
						}
					},
				},
				cr: s3Testing.Bucket(),
			},
			want: want{
				cr:  s3Testing.Bucket(s3Testing.WithConditions(xpv1.Deleting())),
				err: errBoom,
			},
		},
		"ResourceDoesNotExist": {
			args: args{
				s3: &fake.MockBucketClient{
					MockDeleteBucketRequest: func(input *awss3.DeleteBucketInput) awss3.DeleteBucketRequest {
						return awss3.DeleteBucketRequest{
							Request: s3Testing.CreateRequest(awserr.New(s3.BucketNotFoundErrCode, "", nil), &awss3.DeleteBucketOutput{}),
						}
					},
				},
				cr: s3Testing.Bucket(),
			},
			want: want{
				cr: s3Testing.Bucket(s3Testing.WithConditions(xpv1.Deleting())),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{s3client: tc.s3}
			err := e.Delete(context.Background(), tc.args.cr)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.cr, tc.args.cr, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}
