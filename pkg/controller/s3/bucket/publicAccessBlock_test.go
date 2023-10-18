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
	errorutils "github.com/crossplane-contrib/provider-aws/pkg/utils/errors"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
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
					MockGetPublicAccessBlock: func(ctx context.Context, input *s3.GetPublicAccessBlockInput, opts []func(*s3.Options)) (*s3.GetPublicAccessBlockOutput, error) {
						return &s3.GetPublicAccessBlockOutput{}, errBoom
					},
				}),
			},
			want: want{
				status: NeedsUpdate,
				err:    errorutils.Wrap(errBoom, publicAccessBlockGetFailed),
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
					MockGetPublicAccessBlock: func(ctx context.Context, input *s3.GetPublicAccessBlockInput, opts []func(*s3.Options)) (*s3.GetPublicAccessBlockOutput, error) {
						return &s3.GetPublicAccessBlockOutput{}, &smithy.GenericAPIError{Code: clients3.PublicAccessBlockNotFoundErrCode}
					},
				}),
			},
			want: want{
				status: Updated,
			},
		},
		"NotFoundDisabled": {
			args: args{
				cr: &v1beta1.Bucket{
					Spec: v1beta1.BucketSpec{
						ForProvider: v1beta1.BucketParameters{
							PublicAccessBlockConfiguration: &v1beta1.PublicAccessBlockConfiguration{
								BlockPublicAcls:       pointer.ToOrNilIfZeroValue(false),
								IgnorePublicAcls:      pointer.ToOrNilIfZeroValue(false),
								BlockPublicPolicy:     pointer.ToOrNilIfZeroValue(false),
								RestrictPublicBuckets: pointer.ToOrNilIfZeroValue(false),
							},
						},
					},
				},
				cl: NewPublicAccessBlockClient(fake.MockBucketClient{
					MockGetPublicAccessBlock: func(ctx context.Context, input *s3.GetPublicAccessBlockInput, opts []func(*s3.Options)) (*s3.GetPublicAccessBlockOutput, error) {
						return &s3.GetPublicAccessBlockOutput{}, &smithy.GenericAPIError{Code: clients3.PublicAccessBlockNotFoundErrCode}
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
								BlockPublicAcls: pointer.ToOrNilIfZeroValue(true),
							},
						},
					},
				},
				cl: NewPublicAccessBlockClient(fake.MockBucketClient{
					MockGetPublicAccessBlock: func(ctx context.Context, input *s3.GetPublicAccessBlockInput, opts []func(*s3.Options)) (*s3.GetPublicAccessBlockOutput, error) {
						return &s3.GetPublicAccessBlockOutput{PublicAccessBlockConfiguration: &s3types.PublicAccessBlockConfiguration{
							BlockPublicAcls: false,
						}}, nil
					},
				}),
			},
			want: want{
				status: NeedsUpdate,
			},
		},
		"NeedsDeletion": {
			args: args{
				cr: &v1beta1.Bucket{
					Spec: v1beta1.BucketSpec{
						ForProvider: v1beta1.BucketParameters{
							PublicAccessBlockConfiguration: &v1beta1.PublicAccessBlockConfiguration{
								BlockPublicAcls:       pointer.ToOrNilIfZeroValue(false),
								IgnorePublicAcls:      pointer.ToOrNilIfZeroValue(false),
								BlockPublicPolicy:     pointer.ToOrNilIfZeroValue(false),
								RestrictPublicBuckets: pointer.ToOrNilIfZeroValue(false),
							},
						},
					},
				},
				cl: NewPublicAccessBlockClient(fake.MockBucketClient{
					MockGetPublicAccessBlock: func(ctx context.Context, input *s3.GetPublicAccessBlockInput, opts []func(*s3.Options)) (*s3.GetPublicAccessBlockOutput, error) {
						return &s3.GetPublicAccessBlockOutput{PublicAccessBlockConfiguration: &s3types.PublicAccessBlockConfiguration{
							BlockPublicAcls: true,
						}}, nil
					},
				}),
			},
			want: want{
				status: NeedsDeletion,
			},
		},
		"NeedsUpdateMissingField": {
			args: args{
				cr: &v1beta1.Bucket{
					Spec: v1beta1.BucketSpec{
						ForProvider: v1beta1.BucketParameters{
							PublicAccessBlockConfiguration: &v1beta1.PublicAccessBlockConfiguration{
								BlockPublicAcls: pointer.ToOrNilIfZeroValue(true),
							},
						},
					},
				},
				cl: NewPublicAccessBlockClient(fake.MockBucketClient{
					MockGetPublicAccessBlock: func(ctx context.Context, input *s3.GetPublicAccessBlockInput, opts []func(*s3.Options)) (*s3.GetPublicAccessBlockOutput, error) {
						return &s3.GetPublicAccessBlockOutput{PublicAccessBlockConfiguration: &s3types.PublicAccessBlockConfiguration{
							BlockPublicAcls:  true,
							IgnorePublicAcls: true,
						}}, nil
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
								BlockPublicAcls: pointer.ToOrNilIfZeroValue(true),
							},
						},
					},
				},
				cl: NewPublicAccessBlockClient(fake.MockBucketClient{
					MockGetPublicAccessBlock: func(ctx context.Context, input *s3.GetPublicAccessBlockInput, opts []func(*s3.Options)) (*s3.GetPublicAccessBlockOutput, error) {
						return &s3.GetPublicAccessBlockOutput{PublicAccessBlockConfiguration: &s3types.PublicAccessBlockConfiguration{
							BlockPublicAcls: true,
						}}, nil
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
					MockPutPublicAccessBlock: func(ctx context.Context, input *s3.PutPublicAccessBlockInput, opts []func(*s3.Options)) (*s3.PutPublicAccessBlockOutput, error) {
						return &s3.PutPublicAccessBlockOutput{}, errBoom
					},
				}),
			},
			want: want{
				err: errorutils.Wrap(errBoom, publicAccessBlockPutFailed),
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
					MockDeletePublicAccessBlock: func(ctx context.Context, input *s3.DeletePublicAccessBlockInput, opts []func(*s3.Options)) (*s3.DeletePublicAccessBlockOutput, error) {
						return &s3.DeletePublicAccessBlockOutput{}, errBoom
					},
				}),
			},
			want: want{
				err: errorutils.Wrap(errBoom, publicAccessBlockDeleteFailed),
			},
		},
		"GoneAlready": {
			args: args{
				cr: &v1beta1.Bucket{},
				cl: NewPublicAccessBlockClient(fake.MockBucketClient{
					MockDeletePublicAccessBlock: func(ctx context.Context, input *s3.DeletePublicAccessBlockInput, opts []func(*s3.Options)) (*s3.DeletePublicAccessBlockOutput, error) {
						return &s3.DeletePublicAccessBlockOutput{}, &smithy.GenericAPIError{Code: clients3.PublicAccessBlockNotFoundErrCode}
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
					MockGetPublicAccessBlock: func(ctx context.Context, input *s3.GetPublicAccessBlockInput, opts []func(*s3.Options)) (*s3.GetPublicAccessBlockOutput, error) {
						return &s3.GetPublicAccessBlockOutput{}, errBoom
					},
				}),
			},
			want: want{
				err: errorutils.Wrap(errBoom, publicAccessBlockGetFailed),
			},
		},
		"NotFoundSkip": {
			args: args{
				cr: &v1beta1.Bucket{},
				cl: NewPublicAccessBlockClient(fake.MockBucketClient{
					MockGetPublicAccessBlock: func(ctx context.Context, input *s3.GetPublicAccessBlockInput, opts []func(*s3.Options)) (*s3.GetPublicAccessBlockOutput, error) {
						return &s3.GetPublicAccessBlockOutput{}, &smithy.GenericAPIError{Code: clients3.PublicAccessBlockNotFoundErrCode}
					},
				}),
			},
		},
		"Success": {
			args: args{
				cr: &v1beta1.Bucket{},
				cl: NewPublicAccessBlockClient(fake.MockBucketClient{
					MockGetPublicAccessBlock: func(ctx context.Context, input *s3.GetPublicAccessBlockInput, opts []func(*s3.Options)) (*s3.GetPublicAccessBlockOutput, error) {
						return &s3.GetPublicAccessBlockOutput{
							PublicAccessBlockConfiguration: &s3types.PublicAccessBlockConfiguration{},
						}, nil
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
