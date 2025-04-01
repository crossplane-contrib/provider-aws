/*
Copyright 2025 The Crossplane Authors.

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

	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/smithy-go"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"

	"github.com/crossplane-contrib/provider-aws/apis/s3/v1beta1"
	clients3 "github.com/crossplane-contrib/provider-aws/pkg/clients/s3"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/s3/fake"
	s3Testing "github.com/crossplane-contrib/provider-aws/pkg/controller/s3/testing"
	errorutils "github.com/crossplane-contrib/provider-aws/pkg/utils/errors"
)

func TestObjectLockConfigObserve(t *testing.T) {
	type args struct {
		cl *ObjectLockConfigurationClient
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
		"ObjectLockIsNotSet": {
			args: args{
				cr: s3Testing.Bucket(),
				cl: NewObjectLockConfigurationClient(fake.MockBucketClient{
					MockGetObjectLockConfiguration: func(ctx context.Context, input *awss3.GetObjectLockConfigurationInput, opts []func(*awss3.Options)) (*awss3.GetObjectLockConfigurationOutput, error) {
						return nil, &smithy.GenericAPIError{Code: clients3.ObjectLockConfigurationErrCode}
					},
				}),
			},
			want: want{
				status: Updated,
				err:    nil,
			},
		},
		"ObjectLockExplicitlyFalse": {
			args: args{
				cr: s3Testing.Bucket(s3Testing.WithObjectLockEnabledForBucket(false)),
				cl: NewObjectLockConfigurationClient(fake.MockBucketClient{
					MockGetObjectLockConfiguration: func(ctx context.Context, input *awss3.GetObjectLockConfigurationInput, opts []func(*awss3.Options)) (*awss3.GetObjectLockConfigurationOutput, error) {
						return nil, &smithy.GenericAPIError{Code: clients3.ObjectLockConfigurationErrCode}
					},
				}),
			},
			want: want{
				status: Updated,
				err:    nil,
			},
		},
		"ObjectLockSetToFalseButWasEnabledBefore": {
			args: args{
				cr: s3Testing.Bucket(s3Testing.WithObjectLockEnabledForBucket(false)),
				cl: NewObjectLockConfigurationClient(fake.MockBucketClient{
					MockGetObjectLockConfiguration: func(ctx context.Context, input *awss3.GetObjectLockConfigurationInput, opts []func(*awss3.Options)) (*awss3.GetObjectLockConfigurationOutput, error) {
						return &awss3.GetObjectLockConfigurationOutput{ObjectLockConfiguration: &types.ObjectLockConfiguration{ObjectLockEnabled: "Enabled"}}, nil
					},
				}),
			},
			want: want{
				status: NeedsUpdate,
				err:    nil,
			},
		},
		"ObjectLockSetToTrue": {
			args: args{
				cr: s3Testing.Bucket(s3Testing.WithObjectLockEnabledForBucket(true)),
				cl: NewObjectLockConfigurationClient(fake.MockBucketClient{
					MockGetObjectLockConfiguration: func(ctx context.Context, input *awss3.GetObjectLockConfigurationInput, opts []func(*awss3.Options)) (*awss3.GetObjectLockConfigurationOutput, error) {
						return &awss3.GetObjectLockConfigurationOutput{ObjectLockConfiguration: &types.ObjectLockConfiguration{ObjectLockEnabled: "Disabled"}}, nil
					},
				}),
			},
			want: want{
				status: NeedsUpdate,
				err:    nil,
			},
		},
		"ObjectLockRuleIsSetButObjectLockNotEnabled": {
			args: args{
				cr: s3Testing.Bucket(s3Testing.WithObjectLockRule(&v1beta1.ObjectLockRule{DefaultRetention: &v1beta1.DefaultRetention{Mode: "GOVERNANCE", Days: aws.Int32(7)}})),
				cl: NewObjectLockConfigurationClient(fake.MockBucketClient{
					MockGetObjectLockConfiguration: func(ctx context.Context, input *awss3.GetObjectLockConfigurationInput, opts []func(*awss3.Options)) (*awss3.GetObjectLockConfigurationOutput, error) {
						return nil, &smithy.GenericAPIError{Code: clients3.ObjectLockConfigurationErrCode}
					},
				}),
			},
			want: want{
				status: Updated,
				err:    nil,
			},
		},
		"ObjectLockRuleIsSet": {
			args: args{
				cr: s3Testing.Bucket(
					s3Testing.WithObjectLockEnabledForBucket(true),
					s3Testing.WithObjectLockRule(&v1beta1.ObjectLockRule{DefaultRetention: &v1beta1.DefaultRetention{Mode: "GOVERNANCE", Days: aws.Int32(7)}}),
				),
				cl: NewObjectLockConfigurationClient(fake.MockBucketClient{
					MockGetObjectLockConfiguration: func(ctx context.Context, input *awss3.GetObjectLockConfigurationInput, opts []func(*awss3.Options)) (*awss3.GetObjectLockConfigurationOutput, error) {
						return nil, &smithy.GenericAPIError{Code: clients3.ObjectLockConfigurationErrCode}
					},
				}),
			},
			want: want{
				status: NeedsUpdate,
				err:    nil,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got, err := tc.args.cl.Observe(context.Background(), tc.args.cr)
			if diff := cmp.Diff(tc.want.status, got); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestObjectLockConfigCreateOrUpdate(t *testing.T) {
	type args struct {
		cl *ObjectLockConfigurationClient
		cr *v1beta1.Bucket
	}

	type want struct {
		cr  resource.Managed
		err error
	}

	cases := map[string]struct {
		args
		want
	}{
		"ObjectLockSetToFalseButWasEnabledBefore": {
			args: args{
				cr: s3Testing.Bucket(s3Testing.WithObjectLockEnabledForBucket(false)),
				cl: &ObjectLockConfigurationClient{
					cache: objectLockConfigurationCache{
						objectLockConfiguration: &types.ObjectLockConfiguration{
							ObjectLockEnabled: "Enabled",
						},
					},
				},
			},
			want: want{
				err: errorutils.Wrap(errors.New(errMsgObjectLockIsNotDisableable), errObjectLockConfigurationPutFailed),
				cr: s3Testing.Bucket(
					s3Testing.WithObjectLockEnabledForBucket(false),
					s3Testing.WithConditions(xpv1.ReconcileError(errors.New(errMsgObjectLockIsNotDisableable))),
				),
			},
		},
		"ObjectLockAPIErrorResponse": {
			args: args{
				cr: s3Testing.Bucket(s3Testing.WithObjectLockEnabledForBucket(true)),
				cl: NewObjectLockConfigurationClient(fake.MockBucketClient{
					MockPutObjectLockConfiguration: func(ctx context.Context, input *awss3.PutObjectLockConfigurationInput, opts []func(*awss3.Options)) (*awss3.PutObjectLockConfigurationOutput, error) {
						return &awss3.PutObjectLockConfigurationOutput{}, errBoom
					},
				}),
			},
			want: want{
				err: errorutils.Wrap(errBoom, errObjectLockConfigurationPutFailed),
				cr: s3Testing.Bucket(
					s3Testing.WithObjectLockEnabledForBucket(true),
					s3Testing.WithConditions(xpv1.ReconcileError(errBoom)),
				),
			},
		},
		"ObjectLockSetToTrue": {
			args: args{
				cr: s3Testing.Bucket(s3Testing.WithObjectLockEnabledForBucket(true)),
				cl: NewObjectLockConfigurationClient(fake.MockBucketClient{
					MockPutObjectLockConfiguration: func(ctx context.Context, input *awss3.PutObjectLockConfigurationInput, opts []func(*awss3.Options)) (*awss3.PutObjectLockConfigurationOutput, error) {
						return &awss3.PutObjectLockConfigurationOutput{}, nil
					},
				}),
			},
			want: want{
				err: nil,
				cr:  s3Testing.Bucket(s3Testing.WithObjectLockEnabledForBucket(true)),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			err := tc.args.cl.CreateOrUpdate(context.Background(), tc.args.cr)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.cr, tc.args.cr, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}
