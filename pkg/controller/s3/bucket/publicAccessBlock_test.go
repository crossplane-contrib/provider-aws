package bucket

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws/awserr"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/go-cmp/cmp"

	"github.com/crossplane/crossplane-runtime/pkg/test"

	"github.com/crossplane/provider-aws/apis/s3/v1beta1"
	awsclient "github.com/crossplane/provider-aws/pkg/clients"
	clients3 "github.com/crossplane/provider-aws/pkg/clients/s3"
	"github.com/crossplane/provider-aws/pkg/clients/s3/fake"
	s3Testing "github.com/crossplane/provider-aws/pkg/controller/s3/testing"
)

func TestPublicAccessBlockClient_Observe(t *testing.T) {
	type args struct {
		cl *PublicAccessBlockClient
		cr *v1beta1.Bucket
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
				cr: &v1beta1.Bucket{},
				cl: NewPublicAccessBlockClient(fake.MockBucketClient{
					MockGetPublicAccessBlockRequest: func(input *s3.GetPublicAccessBlockInput) s3.GetPublicAccessBlockRequest {
						return s3.GetPublicAccessBlockRequest{
							Request: s3Testing.CreateRequest(errBoom, &s3.GetPublicAccessBlockOutput{}),
						}
					},
				}),
			},
			want: want{
				status: NeedsUpdate,
				err:    awsclient.Wrap(errBoom, publicAccessBlockGetFailed),
			},
		},
		"NotFoundNotNeeded": {
			args: args{
				cr: &v1beta1.Bucket{
					Spec: v1beta1.BucketSpec{
						ForProvider: v1beta1.BucketParameters{},
					},
				},
				cl: NewPublicAccessBlockClient(fake.MockBucketClient{
					MockGetPublicAccessBlockRequest: func(input *s3.GetPublicAccessBlockInput) s3.GetPublicAccessBlockRequest {
						return s3.GetPublicAccessBlockRequest{
							Request: s3Testing.CreateRequest(awserr.New(clients3.PublicAccessBlockNotFoundErrCode, "error", nil), &s3.GetPublicAccessBlockOutput{}),
						}
					},
				}),
			},
			want: want{
				status: Updated,
			},
		},
		"NeedsUpdate": {
			args: args{
				cr: &v1beta1.Bucket{
					Spec: v1beta1.BucketSpec{
						ForProvider: v1beta1.BucketParameters{
							PublicAccessBlockConfiguration: &v1beta1.PublicAccessBlockConfiguration{
								BlockPublicAcls: awsclient.Bool(true),
							},
						},
					},
				},
				cl: NewPublicAccessBlockClient(fake.MockBucketClient{
					MockGetPublicAccessBlockRequest: func(input *s3.GetPublicAccessBlockInput) s3.GetPublicAccessBlockRequest {
						return s3.GetPublicAccessBlockRequest{
							Request: s3Testing.CreateRequest(nil,
								&s3.GetPublicAccessBlockOutput{PublicAccessBlockConfiguration: &s3.PublicAccessBlockConfiguration{
									BlockPublicAcls: awsclient.Bool(false),
								}}),
						}
					},
				}),
			},
			want: want{
				status: NeedsUpdate,
			},
		},
		"NeedsUpdateMissingField": {
			args: args{
				cr: &v1beta1.Bucket{
					Spec: v1beta1.BucketSpec{
						ForProvider: v1beta1.BucketParameters{
							PublicAccessBlockConfiguration: &v1beta1.PublicAccessBlockConfiguration{
								BlockPublicAcls: awsclient.Bool(true),
							},
						},
					},
				},
				cl: NewPublicAccessBlockClient(fake.MockBucketClient{
					MockGetPublicAccessBlockRequest: func(input *s3.GetPublicAccessBlockInput) s3.GetPublicAccessBlockRequest {
						return s3.GetPublicAccessBlockRequest{
							Request: s3Testing.CreateRequest(nil,
								&s3.GetPublicAccessBlockOutput{PublicAccessBlockConfiguration: &s3.PublicAccessBlockConfiguration{
									BlockPublicAcls:  awsclient.Bool(true),
									IgnorePublicAcls: awsclient.Bool(true),
								}}),
						}
					},
				}),
			},
			want: want{
				status: NeedsUpdate,
			},
		},
		"Updated": {
			args: args{
				cr: &v1beta1.Bucket{
					Spec: v1beta1.BucketSpec{
						ForProvider: v1beta1.BucketParameters{
							PublicAccessBlockConfiguration: &v1beta1.PublicAccessBlockConfiguration{
								BlockPublicAcls: awsclient.Bool(true),
							},
						},
					},
				},
				cl: NewPublicAccessBlockClient(fake.MockBucketClient{
					MockGetPublicAccessBlockRequest: func(input *s3.GetPublicAccessBlockInput) s3.GetPublicAccessBlockRequest {
						return s3.GetPublicAccessBlockRequest{
							Request: s3Testing.CreateRequest(nil,
								&s3.GetPublicAccessBlockOutput{PublicAccessBlockConfiguration: &s3.PublicAccessBlockConfiguration{
									BlockPublicAcls: awsclient.Bool(true),
								}}),
						}
					},
				}),
			},
			want: want{
				status: Updated,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			status, err := tc.args.cl.Observe(context.Background(), tc.args.cr)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.status, status); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestPublicAccessBlockClient_CreateOrUpdate(t *testing.T) {
	type args struct {
		cl *PublicAccessBlockClient
		cr *v1beta1.Bucket
	}

	type want struct {
		err error
	}

	cases := map[string]struct {
		args
		want
	}{
		"Skip": {
			args: args{
				cr: &v1beta1.Bucket{},
			},
			want: want{},
		},
		"Error": {
			args: args{
				cr: &v1beta1.Bucket{
					Spec: v1beta1.BucketSpec{
						ForProvider: v1beta1.BucketParameters{
							PublicAccessBlockConfiguration: &v1beta1.PublicAccessBlockConfiguration{},
						},
					},
				},
				cl: NewPublicAccessBlockClient(fake.MockBucketClient{
					MockPutPublicAccessBlockRequest: func(input *s3.PutPublicAccessBlockInput) s3.PutPublicAccessBlockRequest {
						return s3.PutPublicAccessBlockRequest{
							Request: s3Testing.CreateRequest(errBoom, &s3.PutPublicAccessBlockOutput{}),
						}
					},
				}),
			},
			want: want{
				err: awsclient.Wrap(errBoom, publicAccessBlockPutFailed),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			err := tc.args.cl.CreateOrUpdate(context.Background(), tc.args.cr)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestPublicAccessBlockClient_Delete(t *testing.T) {
	type args struct {
		cl *PublicAccessBlockClient
		cr *v1beta1.Bucket
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
				cr: &v1beta1.Bucket{},
				cl: NewPublicAccessBlockClient(fake.MockBucketClient{
					MockDeletePublicAccessBlockRequest: func(input *s3.DeletePublicAccessBlockInput) s3.DeletePublicAccessBlockRequest {
						return s3.DeletePublicAccessBlockRequest{
							Request: s3Testing.CreateRequest(errBoom, &s3.DeletePublicAccessBlockOutput{}),
						}
					},
				}),
			},
			want: want{
				err: awsclient.Wrap(errBoom, publicAccessBlockDeleteFailed),
			},
		},
		"GoneAlready": {
			args: args{
				cr: &v1beta1.Bucket{},
				cl: NewPublicAccessBlockClient(fake.MockBucketClient{
					MockDeletePublicAccessBlockRequest: func(input *s3.DeletePublicAccessBlockInput) s3.DeletePublicAccessBlockRequest {
						return s3.DeletePublicAccessBlockRequest{
							Request: s3Testing.CreateRequest(awserr.New(clients3.PublicAccessBlockNotFoundErrCode, "error", nil), &s3.DeletePublicAccessBlockOutput{}),
						}
					},
				}),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			err := tc.args.cl.Delete(context.Background(), tc.args.cr)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestPublicAccessBlockClient_LateInitialize(t *testing.T) {
	type args struct {
		cl *PublicAccessBlockClient
		cr *v1beta1.Bucket
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
				cr: &v1beta1.Bucket{},
				cl: NewPublicAccessBlockClient(fake.MockBucketClient{
					MockGetPublicAccessBlockRequest: func(input *s3.GetPublicAccessBlockInput) s3.GetPublicAccessBlockRequest {
						return s3.GetPublicAccessBlockRequest{
							Request: s3Testing.CreateRequest(errBoom, &s3.GetPublicAccessBlockOutput{}),
						}
					},
				}),
			},
			want: want{
				err: awsclient.Wrap(errBoom, publicAccessBlockGetFailed),
			},
		},
		"NotFoundSkip": {
			args: args{
				cr: &v1beta1.Bucket{},
				cl: NewPublicAccessBlockClient(fake.MockBucketClient{
					MockGetPublicAccessBlockRequest: func(input *s3.GetPublicAccessBlockInput) s3.GetPublicAccessBlockRequest {
						return s3.GetPublicAccessBlockRequest{
							Request: s3Testing.CreateRequest(awserr.New(clients3.PublicAccessBlockNotFoundErrCode, "error", nil), &s3.GetPublicAccessBlockOutput{}),
						}
					},
				}),
			},
		},
		"Success": {
			args: args{
				cr: &v1beta1.Bucket{},
				cl: NewPublicAccessBlockClient(fake.MockBucketClient{
					MockGetPublicAccessBlockRequest: func(input *s3.GetPublicAccessBlockInput) s3.GetPublicAccessBlockRequest {
						return s3.GetPublicAccessBlockRequest{
							Request: s3Testing.CreateRequest(nil, &s3.GetPublicAccessBlockOutput{
								PublicAccessBlockConfiguration: &s3.PublicAccessBlockConfiguration{},
							}),
						}
					},
				}),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			err := tc.args.cl.LateInitialize(context.Background(), tc.args.cr)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}
