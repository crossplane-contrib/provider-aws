package bucketclients

import (
	"testing"

	_ "github.com/aws/aws-sdk-go-v2/aws/awserr"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	"golang.org/x/net/context"

	"github.com/crossplane/provider-aws/apis/s3/v1beta1"
	"github.com/crossplane/provider-aws/pkg/clients/s3/fake"
)

func TestAccelerateExistsAndUpdated(t *testing.T) {
	type args struct {
		cl *AccelerateConfigurationClient
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
				cl: CreateAccelerateConfigurationClient(
					bucket(withAccelerationConfig(&v1beta1.AccelerateConfiguration{Status: enabled})),
					fake.MockBucketClient{
						MockGetBucketAccelerateConfigurationRequest: func(input *s3.GetBucketAccelerateConfigurationInput) s3.GetBucketAccelerateConfigurationRequest {
							return s3.GetBucketAccelerateConfigurationRequest{
								Request: createRequest(errBoom, &s3.GetBucketAccelerateConfigurationOutput{}),
							}
						},
					},
				),
			},
			want: want{
				status: NeedsUpdate,
				err:    errors.Wrap(errBoom, accelGetFailed),
			},
		},
		"UpdateNeeded": {
			args: args{
				cl: CreateAccelerateConfigurationClient(
					bucket(withAccelerationConfig(&v1beta1.AccelerateConfiguration{Status: "Enabled"})),
					fake.MockBucketClient{
						MockGetBucketAccelerateConfigurationRequest: func(input *s3.GetBucketAccelerateConfigurationInput) s3.GetBucketAccelerateConfigurationRequest {
							return s3.GetBucketAccelerateConfigurationRequest{
								Request: createRequest(nil, &s3.GetBucketAccelerateConfigurationOutput{Status: s3.BucketAccelerateStatusSuspended}),
							}
						},
					},
				),
			},
			want: want{
				status: NeedsUpdate,
				err:    nil,
			},
		},
		"NeedsDelete": {
			args: args{
				cl: CreateAccelerateConfigurationClient(
					bucket(withAccelerationConfig(nil)),
					fake.MockBucketClient{
						MockGetBucketAccelerateConfigurationRequest: func(input *s3.GetBucketAccelerateConfigurationInput) s3.GetBucketAccelerateConfigurationRequest {
							return s3.GetBucketAccelerateConfigurationRequest{
								Request: createRequest(nil, &s3.GetBucketAccelerateConfigurationOutput{Status: s3.BucketAccelerateStatusSuspended}),
							}
						},
					},
				),
			},
			want: want{
				status: NeedsDeletion,
				err:    nil,
			},
		},
		"NoUpdateNotExists": {
			args: args{
				cl: CreateAccelerateConfigurationClient(
					bucket(withAccelerationConfig(nil)),
					fake.MockBucketClient{
						MockGetBucketAccelerateConfigurationRequest: func(input *s3.GetBucketAccelerateConfigurationInput) s3.GetBucketAccelerateConfigurationRequest {
							return s3.GetBucketAccelerateConfigurationRequest{
								Request: createRequest(nil, &s3.GetBucketAccelerateConfigurationOutput{}),
							}
						},
					},
				),
			},
			want: want{
				status: Updated,
				err:    nil,
			},
		},
		"NoUpdateExists": {
			args: args{
				cl: CreateAccelerateConfigurationClient(
					bucket(withAccelerationConfig(&v1beta1.AccelerateConfiguration{Status: suspended})),
					fake.MockBucketClient{
						MockGetBucketAccelerateConfigurationRequest: func(input *s3.GetBucketAccelerateConfigurationInput) s3.GetBucketAccelerateConfigurationRequest {
							return s3.GetBucketAccelerateConfigurationRequest{
								Request: createRequest(nil, &s3.GetBucketAccelerateConfigurationOutput{Status: s3.BucketAccelerateStatusSuspended}),
							}
						},
					},
				),
			},
			want: want{
				status: Updated,
				err:    nil,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			status, err := tc.args.cl.ExistsAndUpdated(context.Background())
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.status, status); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestAccelerateCreate(t *testing.T) {
	type args struct {
		cl BucketResource
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
				cl: CreateAccelerateConfigurationClient(
					bucket(withAccelerationConfig(&v1beta1.AccelerateConfiguration{Status: enabled})),
					fake.MockBucketClient{
						MockPutBucketAccelerateConfigurationRequest: func(input *s3.PutBucketAccelerateConfigurationInput) s3.PutBucketAccelerateConfigurationRequest {
							return s3.PutBucketAccelerateConfigurationRequest{
								Request: createRequest(errBoom, &s3.PutBucketAccelerateConfigurationOutput{}),
							}
						},
					},
				),
			},
			want: want{
				err: errors.Wrap(errBoom, accelPutFailed),
			},
		},
		"InvalidConfig": {
			args: args{
				cl: CreateAccelerateConfigurationClient(
					bucket(withAccelerationConfig(nil)),
					fake.MockBucketClient{
						MockPutBucketAccelerateConfigurationRequest: func(input *s3.PutBucketAccelerateConfigurationInput) s3.PutBucketAccelerateConfigurationRequest {
							return s3.PutBucketAccelerateConfigurationRequest{
								Request: createRequest(nil, &s3.PutBucketAccelerateConfigurationOutput{}),
							}
						},
					},
				),
			},
			want: want{
				err: nil,
			},
		},
		"SuccessfulCreate": {
			args: args{
				cl: CreateAccelerateConfigurationClient(
					bucket(withAccelerationConfig(&v1beta1.AccelerateConfiguration{Status: enabled})),
					fake.MockBucketClient{
						MockPutBucketAccelerateConfigurationRequest: func(input *s3.PutBucketAccelerateConfigurationInput) s3.PutBucketAccelerateConfigurationRequest {
							return s3.PutBucketAccelerateConfigurationRequest{
								Request: createRequest(nil, &s3.PutBucketAccelerateConfigurationOutput{}),
							}
						},
					},
				),
			},
			want: want{
				err: nil,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := tc.args.cl.CreateResource(context.Background())
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestAccelerateDelete(t *testing.T) {
	type args struct {
		cl BucketResource
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
				cl: CreateAccelerateConfigurationClient(
					bucket(),
					fake.MockBucketClient{
						MockPutBucketAccelerateConfigurationRequest: func(input *s3.PutBucketAccelerateConfigurationInput) s3.PutBucketAccelerateConfigurationRequest {
							return s3.PutBucketAccelerateConfigurationRequest{
								Request: createRequest(errBoom, &s3.PutBucketAccelerateConfigurationOutput{}),
							}
						},
					},
				),
			},
			want: want{
				err: errors.Wrap(errBoom, accelDeleteFailed),
			},
		},
		"SuccessfulDelete": {
			args: args{
				cl: CreateAccelerateConfigurationClient(
					bucket(),
					fake.MockBucketClient{
						MockPutBucketAccelerateConfigurationRequest: func(input *s3.PutBucketAccelerateConfigurationInput) s3.PutBucketAccelerateConfigurationRequest {
							return s3.PutBucketAccelerateConfigurationRequest{
								Request: createRequest(nil, &s3.PutBucketAccelerateConfigurationOutput{}),
							}
						},
					},
				),
			},
			want: want{
				err: nil,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			err := tc.args.cl.DeleteResource(context.Background())
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}
