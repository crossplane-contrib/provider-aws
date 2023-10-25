package accesspoint

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go/service/s3control"
	"github.com/aws/aws-sdk-go/service/s3control/s3controliface"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane-contrib/provider-aws/apis/s3/common"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/s3control/fake"
	s3controlTesting "github.com/crossplane-contrib/provider-aws/pkg/controller/s3control/testing"
	errorutils "github.com/crossplane-contrib/provider-aws/pkg/utils/errors"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
)

var (
	unexpectedItem     resource.Managed
	errBoom            = errors.New("boom")
	testOutputPolicyV1 = s3control.GetAccessPointPolicyOutput{
		Policy: pointer.ToOrNilIfZeroValue(`{
						"Version": "2012-10-17",
						"Statement": [{
							"Sid": "AllowPublicRead",
							"Effect": "Allow",
							"Principal": {
								"AWS": "arn:aws:iam::1234567890:role/sso/role"
							},
							"Action": "s3:GetObject",
							"Resource": "arn:aws:s3:::my-bucket/*"
						}]
					}`),
	}
	testPolicyV1 = &common.BucketPolicyBody{
		Version: "2012-10-17",
		Statements: []common.BucketPolicyStatement{
			{
				SID:    pointer.ToOrNilIfZeroValue("AllowPublicRead"),
				Effect: "Allow",
				Principal: &common.BucketPrincipal{
					AWSPrincipals: []common.AWSPrincipal{
						{IAMRoleARN: pointer.ToOrNilIfZeroValue("arn:aws:iam::1234567890:role/sso/role")},
					},
				},
				Action:   []string{"s3:GetObject"},
				Resource: []string{"arn:aws:s3:::my-bucket/*"},
			},
		},
	}
	testPolicyV2 = &common.BucketPolicyBody{
		Version: "2012-10-17",
		Statements: []common.BucketPolicyStatement{
			{
				SID:    pointer.ToOrNilIfZeroValue("AllowPublicWrite"),
				Effect: "Allow",
				Principal: &common.BucketPrincipal{
					AWSPrincipals: []common.AWSPrincipal{
						{IAMRoleARN: pointer.ToOrNilIfZeroValue("arn:aws:iam::1234567890:role/sso/role")},
					},
				},
				Action:   []string{"s3:GetObject"},
				Resource: []string{"arn:aws:s3:::my-bucket/*"},
			},
		},
	}
)

func TestObserve(t *testing.T) {
	type args struct {
		kube   client.Client
		client s3controliface.S3ControlAPI
		cr     resource.Managed
	}

	type want struct {
		cr     resource.Managed
		result managed.ExternalObservation
		err    error
	}

	tests := []struct {
		name string
		args args
		want want
	}{
		{name: "InValidInput",
			args: args{
				cr: unexpectedItem,
			},
			want: want{
				cr:  unexpectedItem,
				err: errors.New(errUnexpectedObject),
			},
		},
		{name: "ClientError",
			args: args{
				client: &fake.MockS3ControlClient{
					GetAccessPointWithContextErr: errBoom,
				},
				cr: s3controlTesting.AccessPoint(),
			},
			want: want{
				cr:  s3controlTesting.AccessPoint(),
				err: errorutils.Wrap(errBoom, errDescribe),
			},
		},
		{name: "ResourceDoesNotExist",
			args: args{
				client: &fake.MockS3ControlClient{
					GetAccessPointWithContextErr: s3controlTesting.NoSuchAccessPoint(),
				},
				cr: s3controlTesting.AccessPoint(),
			},
			want: want{
				cr:     s3controlTesting.AccessPoint(),
				result: managed.ExternalObservation{},
			},
		},
		{name: "ValidInput",
			args: args{
				client: &fake.MockS3ControlClient{
					GetAccessPointWithContextOutput: s3control.GetAccessPointOutput{},
				},
				cr: s3controlTesting.AccessPoint(),
			},
			want: want{
				cr: s3controlTesting.AccessPoint(
					s3controlTesting.WithConditions(xpv1.Available()),
				),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
				},
			},
		},
		{name: "ValidInputWithPolicy",
			args: args{
				client: &fake.MockS3ControlClient{
					GetAccessPointWithContextOutput: s3control.GetAccessPointOutput{},
					GetAccessPointPolicyErr:         s3controlTesting.NoSuchAccessPointPolicy(),
				},
				cr: s3controlTesting.AccessPoint(
					s3controlTesting.WithPolicy(testPolicyV1),
				),
			},
			want: want{
				cr: s3controlTesting.AccessPoint(
					s3controlTesting.WithPolicy(testPolicyV1),
					s3controlTesting.WithConditions(xpv1.Available()),
				),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: false,
				},
			},
		},
		{name: "AccessPointPolicyUpToDate",
			args: args{
				client: &fake.MockS3ControlClient{
					GetAccessPointWithContextOutput: s3control.GetAccessPointOutput{},
					GetAccessPointPolicyOutput:      testOutputPolicyV1,
				},
				cr: s3controlTesting.AccessPoint(
					s3controlTesting.WithPolicy(testPolicyV1),
				),
			},
			want: want{
				cr: s3controlTesting.AccessPoint(
					s3controlTesting.WithPolicy(testPolicyV1),
					s3controlTesting.WithConditions(xpv1.Available()),
				),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
				},
			},
		},
		{name: "AccessPointPolicyNotUpToDate",
			args: args{
				client: &fake.MockS3ControlClient{
					GetAccessPointWithContextOutput: s3control.GetAccessPointOutput{},
					GetAccessPointPolicyOutput:      testOutputPolicyV1,
				},
				cr: s3controlTesting.AccessPoint(
					s3controlTesting.WithPolicy(testPolicyV2),
				),
			},
			want: want{
				cr: s3controlTesting.AccessPoint(
					s3controlTesting.WithPolicy(testPolicyV2),
					s3controlTesting.WithConditions(xpv1.Available()),
				),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: false,
				},
			},
		},
		{name: "AccessPointPolicyNotUpToDateNeedsDeletion",
			args: args{
				client: &fake.MockS3ControlClient{
					GetAccessPointWithContextOutput: s3control.GetAccessPointOutput{},
					GetAccessPointPolicyOutput:      testOutputPolicyV1,
				},
				cr: s3controlTesting.AccessPoint(),
			},
			want: want{
				cr: s3controlTesting.AccessPoint(
					s3controlTesting.WithConditions(xpv1.Available()),
				),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: false,
				},
			},
		},
		{name: "FailedToDescribeAccessPointPolicy",
			args: args{
				client: &fake.MockS3ControlClient{
					GetAccessPointWithContextOutput: s3control.GetAccessPointOutput{},
					GetAccessPointPolicyErr:         errBoom,
				},
				cr: s3controlTesting.AccessPoint(
					s3controlTesting.WithPolicy(testPolicyV2),
				),
			},
			want: want{
				cr: s3controlTesting.AccessPoint(
					s3controlTesting.WithPolicy(testPolicyV2),
				),
				err: errors.Wrap(errorutils.Wrap(errBoom, errDescribePolicy), "isUpToDate check failed"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := newExternal(tt.args.kube, tt.args.client, createOptions())
			o, err := e.Observe(context.Background(), tt.args.cr)
			if diff := cmp.Diff(tt.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tt.want.cr, tt.args.cr, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tt.want.result, o); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestUpdate(t *testing.T) {
	type args struct {
		kube   client.Client
		client s3controliface.S3ControlAPI
		cr     resource.Managed
	}

	type want struct {
		cr     resource.Managed
		result managed.ExternalUpdate
		err    error
	}

	tests := []struct {
		name string
		args args
		want want
	}{
		{name: "InValidInput",
			args: args{
				cr: unexpectedItem,
			},
			want: want{
				cr:  unexpectedItem,
				err: errors.New(errUnexpectedObject),
			},
		},
		{name: "GetAccessPointPolicyClientError",
			args: args{
				client: &fake.MockS3ControlClient{
					GetAccessPointPolicyErr: errBoom,
				},
				cr: s3controlTesting.AccessPoint(),
			},
			want: want{
				cr: s3controlTesting.AccessPoint(
					s3controlTesting.WithConditions(xpv1.ReconcileError(errorutils.Wrap(errBoom, errDescribePolicy))),
				),
				err: errorutils.Wrap(errBoom, errDescribePolicy),
			},
		},
		{name: "DeleteAccessPointPolicyClientError",
			args: args{
				client: &fake.MockS3ControlClient{
					GetAccessPointPolicyOutput:            testOutputPolicyV1,
					DeleteAccessPointPolicyWithContextErr: errBoom,
				},
				cr: s3controlTesting.AccessPoint(),
			},
			want: want{
				cr:  s3controlTesting.AccessPoint(),
				err: errorutils.Wrap(errBoom, errDeletePolicy),
			},
		},
		{name: "DeleteAccessPointPolicyClientEmptyOutputPolicy",
			args: args{
				client: &fake.MockS3ControlClient{
					GetAccessPointPolicyOutput: s3control.GetAccessPointPolicyOutput{},
				},
				cr: s3controlTesting.AccessPoint(),
			},
			want: want{
				cr: s3controlTesting.AccessPoint(),
			},
		},
		{name: "UpdatePolicy",
			args: args{
				client: &fake.MockS3ControlClient{
					GetAccessPointPolicyOutput: testOutputPolicyV1,
				},
				cr: s3controlTesting.AccessPoint(
					s3controlTesting.WithPolicy(testPolicyV2),
				),
			},
			want: want{
				cr: s3controlTesting.AccessPoint(
					s3controlTesting.WithPolicy(testPolicyV2),
				),
			},
		},
		{name: "PutAccessPointPolicyWithContextClientError",
			args: args{
				client: &fake.MockS3ControlClient{
					GetAccessPointPolicyOutput:         testOutputPolicyV1,
					PutAccessPointPolicyWithContextErr: errBoom,
				},
				cr: s3controlTesting.AccessPoint(
					s3controlTesting.WithPolicy(testPolicyV2),
				),
			},
			want: want{
				cr: s3controlTesting.AccessPoint(
					s3controlTesting.WithPolicy(testPolicyV2),
				),
				err: errorutils.Wrap(errBoom, errPutPolicy),
			},
		},
		{name: "DontUpdatePolicy",
			args: args{
				client: &fake.MockS3ControlClient{
					GetAccessPointPolicyOutput:         testOutputPolicyV1,
					PutAccessPointPolicyWithContextErr: errBoom,
				},
				cr: s3controlTesting.AccessPoint(
					s3controlTesting.WithPolicy(testPolicyV1),
				),
			},
			want: want{
				cr: s3controlTesting.AccessPoint(
					s3controlTesting.WithPolicy(testPolicyV1),
				),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := newExternal(tt.args.kube, tt.args.client, createOptions())
			o, err := e.Update(context.Background(), tt.args.cr)
			if diff := cmp.Diff(tt.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tt.want.cr, tt.args.cr, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tt.want.result, o); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestCreate(t *testing.T) {
	type args struct {
		kube   client.Client
		client s3controliface.S3ControlAPI
		cr     resource.Managed
	}

	type want struct {
		cr     resource.Managed
		result managed.ExternalCreation
		err    error
	}

	tests := []struct {
		name string
		args args
		want want
	}{
		{name: "InValidInput",
			args: args{
				cr: unexpectedItem,
			},
			want: want{
				cr:  unexpectedItem,
				err: errors.New(errUnexpectedObject),
			},
		},
		{name: "ClientError",
			args: args{
				client: &fake.MockS3ControlClient{
					CreateAccessPointWithContextErr: errBoom,
				},
				cr: s3controlTesting.AccessPoint(),
			},
			want: want{
				cr: s3controlTesting.AccessPoint(
					s3controlTesting.WithConditions(xpv1.Creating()),
				),
				err: errorutils.Wrap(errBoom, errCreate),
			},
		},
		{name: "ValidInput",
			args: args{
				client: &fake.MockS3ControlClient{
					CreateAccessPointWithContextOutput: s3control.CreateAccessPointOutput{},
				},
				cr: s3controlTesting.AccessPoint(),
			},
			want: want{
				cr: s3controlTesting.AccessPoint(
					s3controlTesting.WithConditions(xpv1.Creating()),
				),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := newExternal(tt.args.kube, tt.args.client, createOptions())
			o, err := e.Create(context.Background(), tt.args.cr)
			if diff := cmp.Diff(tt.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tt.want.cr, tt.args.cr, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tt.want.result, o); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	type args struct {
		kube   client.Client
		client s3controliface.S3ControlAPI
		cr     resource.Managed
	}

	type want struct {
		cr  resource.Managed
		err error
	}

	tests := []struct {
		name string
		args args
		want want
	}{
		{name: "InValidInput",
			args: args{
				cr: unexpectedItem,
			},
			want: want{
				cr:  unexpectedItem,
				err: errors.New(errUnexpectedObject),
			},
		},
		{name: "ClientError",
			args: args{
				client: &fake.MockS3ControlClient{
					DeleteAccessPointWithContextErr: errBoom,
				},
				cr: s3controlTesting.AccessPoint(),
			},
			want: want{
				cr: s3controlTesting.AccessPoint(
					s3controlTesting.WithConditions(xpv1.Deleting()),
				),
				err: errorutils.Wrap(errBoom, errDelete),
			},
		},
		{name: "ValidInput",
			args: args{
				client: &fake.MockS3ControlClient{
					DeleteAccessPointWithContextOutput: s3control.DeleteAccessPointOutput{},
				},
				cr: s3controlTesting.AccessPoint(),
			},
			want: want{
				cr: s3controlTesting.AccessPoint(
					s3controlTesting.WithConditions(xpv1.Deleting()),
				),
			},
		},
		{name: "ResourceDoesNotExist",
			args: args{
				client: &fake.MockS3ControlClient{
					DeleteAccessPointWithContextErr: s3controlTesting.NoSuchAccessPoint(),
				},
				cr: s3controlTesting.AccessPoint(),
			},
			want: want{
				cr: s3controlTesting.AccessPoint(
					s3controlTesting.WithConditions(xpv1.Deleting()),
				),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := newExternal(tt.args.kube, tt.args.client, createOptions())
			err := e.Delete(context.Background(), tt.args.cr)
			if diff := cmp.Diff(tt.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tt.want.cr, tt.args.cr, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}
