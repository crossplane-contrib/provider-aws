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
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
	"github.com/aws/smithy-go/document"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/crossplane-contrib/provider-aws/apis/s3/v1beta1"
	clients3 "github.com/crossplane-contrib/provider-aws/pkg/clients/s3"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/s3/fake"
	s3testing "github.com/crossplane-contrib/provider-aws/pkg/controller/s3/testing"
	errorutils "github.com/crossplane-contrib/provider-aws/pkg/utils/errors"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
)

var (
	days        int32 = 1
	location, _       = time.LoadLocation("UTC")
	date              = metav1.Date(2020, time.September, 25, 11, 40, 0, 0, location)
	awsDate           = time.Date(2020, time.September, 25, 11, 40, 0, 0, location)
	marker            = false
	prefix            = "test-"
	id                = "test-id"
	storage           = "ONEZONE_IA"
)

var _ SubresourceClient = &LifecycleConfigurationClient{}

func generateLifecycleConfig() *v1beta1.BucketLifecycleConfiguration {
	return &v1beta1.BucketLifecycleConfiguration{
		Rules: []v1beta1.LifecycleRule{
			{
				AbortIncompleteMultipartUpload: &v1beta1.AbortIncompleteMultipartUpload{DaysAfterInitiation: 1},
				Expiration: &v1beta1.LifecycleExpiration{
					Date:                      &date,
					Days:                      days,
					ExpiredObjectDeleteMarker: marker,
				},
				Filter: &v1beta1.LifecycleRuleFilter{
					And: &v1beta1.LifecycleRuleAndOperator{
						Prefix: pointer.ToOrNilIfZeroValue(prefix),
						Tags:   tags,
					},
				},
				ID:                          pointer.ToOrNilIfZeroValue(id),
				NoncurrentVersionExpiration: &v1beta1.NoncurrentVersionExpiration{NoncurrentDays: days},
				NoncurrentVersionTransitions: []v1beta1.NoncurrentVersionTransition{{
					NoncurrentDays: days,
					StorageClass:   storage,
				}},
				Status: enabled,
				Transitions: []v1beta1.Transition{{
					Date:         &date,
					Days:         days,
					StorageClass: storage,
				}},
			},
		},
	}
}

func generateAWSLifecycle(sortTag bool) *s3types.BucketLifecycleConfiguration {
	conf := &s3types.BucketLifecycleConfiguration{
		Rules: []s3types.LifecycleRule{
			{
				AbortIncompleteMultipartUpload: &s3types.AbortIncompleteMultipartUpload{DaysAfterInitiation: 1},
				Expiration: &s3types.LifecycleExpiration{
					Date:                      &awsDate,
					Days:                      days,
					ExpiredObjectDeleteMarker: marker,
				},
				Filter: &s3types.LifecycleRuleFilterMemberAnd{
					Value: s3types.LifecycleRuleAndOperator{
						Prefix: pointer.ToOrNilIfZeroValue(prefix),
						Tags:   awsTags,
					},
				},
				ID:                          pointer.ToOrNilIfZeroValue(id),
				NoncurrentVersionExpiration: &s3types.NoncurrentVersionExpiration{NoncurrentDays: days},
				NoncurrentVersionTransitions: []s3types.NoncurrentVersionTransition{{
					NoncurrentDays: days,
					StorageClass:   s3types.TransitionStorageClassOnezoneIa,
				}},
				Status: s3types.ExpirationStatusEnabled,
				Transitions: []s3types.Transition{{
					Date:         &awsDate,
					Days:         days,
					StorageClass: s3types.TransitionStorageClassOnezoneIa,
				}},
			},
		},
	}
	if sortTag {
		sortFilterTags(conf.Rules)
	}
	return conf
}

func TestGenerateLifecycleConfiguration(t *testing.T) {
	type args struct {
		b *v1beta1.Bucket
	}

	type want struct {
		input []s3types.LifecycleRule
	}

	cases := map[string]struct {
		args
		want
	}{
		"SameStruct": {
			args: args{
				b: s3testing.Bucket(s3testing.WithLifecycleConfig(generateLifecycleConfig())),
			},
			want: want{
				input: generateAWSLifecycle(true).Rules,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			generated := GenerateLifecycleRules(tc.args.b.Spec.ForProvider.LifecycleConfiguration.Rules)
			if diff := cmp.Diff(generated, tc.want.input, cmpopts.IgnoreTypes(document.NoSerde{})); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestLifecycleObserve(t *testing.T) {
	type args struct {
		cl *LifecycleConfigurationClient
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
				b: s3testing.Bucket(s3testing.WithLifecycleConfig(generateLifecycleConfig())),
				cl: NewLifecycleConfigurationClient(fake.MockBucketClient{
					MockGetBucketLifecycleConfiguration: func(ctx context.Context, input *s3.GetBucketLifecycleConfigurationInput, opts []func(*s3.Options)) (*s3.GetBucketLifecycleConfigurationOutput, error) {
						return nil, errBoom
					},
				}),
			},
			want: want{
				status: NeedsUpdate,
				err:    errorutils.Wrap(errBoom, lifecycleGetFailed),
			},
		},
		"UpdateNeeded": {
			args: args{
				b: s3testing.Bucket(s3testing.WithLifecycleConfig(generateLifecycleConfig())),
				cl: NewLifecycleConfigurationClient(fake.MockBucketClient{
					MockGetBucketLifecycleConfiguration: func(ctx context.Context, input *s3.GetBucketLifecycleConfigurationInput, opts []func(*s3.Options)) (*s3.GetBucketLifecycleConfigurationOutput, error) {
						return &s3.GetBucketLifecycleConfigurationOutput{Rules: nil}, nil
					},
				}),
			},
			want: want{
				status: NeedsUpdate,
				err:    nil,
			},
		},
		"NeedsDelete": {
			args: args{
				b: s3testing.Bucket(s3testing.WithLifecycleConfig(nil)),
				cl: NewLifecycleConfigurationClient(fake.MockBucketClient{
					MockGetBucketLifecycleConfiguration: func(ctx context.Context, input *s3.GetBucketLifecycleConfigurationInput, opts []func(*s3.Options)) (*s3.GetBucketLifecycleConfigurationOutput, error) {
						return &s3.GetBucketLifecycleConfigurationOutput{Rules: generateAWSLifecycle(false).Rules}, nil
					},
				}),
			},
			want: want{
				status: NeedsDeletion,
				err:    nil,
			},
		},
		"NoUpdateNotExists": {
			args: args{
				b: s3testing.Bucket(s3testing.WithLifecycleConfig(nil)),
				cl: NewLifecycleConfigurationClient(fake.MockBucketClient{
					MockGetBucketLifecycleConfiguration: func(ctx context.Context, input *s3.GetBucketLifecycleConfigurationInput, opts []func(*s3.Options)) (*s3.GetBucketLifecycleConfigurationOutput, error) {
						return nil, &smithy.GenericAPIError{Code: clients3.LifecycleNotFoundErrCode}
					},
				}),
			},
			want: want{
				status: Updated,
				err:    nil,
			},
		},
		"NoUpdateNotExistsNil": {
			args: args{
				b: s3testing.Bucket(s3testing.WithLifecycleConfig(nil)),
				cl: NewLifecycleConfigurationClient(fake.MockBucketClient{
					MockGetBucketLifecycleConfiguration: func(ctx context.Context, input *s3.GetBucketLifecycleConfigurationInput, opts []func(*s3.Options)) (*s3.GetBucketLifecycleConfigurationOutput, error) {
						return &s3.GetBucketLifecycleConfigurationOutput{Rules: nil}, nil
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
				b: s3testing.Bucket(s3testing.WithLifecycleConfig(generateLifecycleConfig())),
				cl: NewLifecycleConfigurationClient(fake.MockBucketClient{
					MockGetBucketLifecycleConfiguration: func(ctx context.Context, input *s3.GetBucketLifecycleConfigurationInput, opts []func(*s3.Options)) (*s3.GetBucketLifecycleConfigurationOutput, error) {
						return &s3.GetBucketLifecycleConfigurationOutput{Rules: generateAWSLifecycle(false).Rules}, nil
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
			if diff := cmp.Diff(tc.want.status, status, cmpopts.IgnoreTypes(document.NoSerde{})); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestLifecycleCreateOrUpdate(t *testing.T) {
	type args struct {
		cl *LifecycleConfigurationClient
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
				b: s3testing.Bucket(s3testing.WithLifecycleConfig(generateLifecycleConfig())),
				cl: NewLifecycleConfigurationClient(fake.MockBucketClient{
					MockPutBucketLifecycleConfiguration: func(ctx context.Context, input *s3.PutBucketLifecycleConfigurationInput, opts []func(*s3.Options)) (*s3.PutBucketLifecycleConfigurationOutput, error) {
						return nil, errBoom
					},
				}),
			},
			want: want{
				err: errorutils.Wrap(errBoom, lifecyclePutFailed),
			},
		},
		"InvalidConfig": {
			args: args{
				b: s3testing.Bucket(s3testing.WithLifecycleConfig(generateLifecycleConfig())),
				cl: NewLifecycleConfigurationClient(fake.MockBucketClient{
					MockPutBucketLifecycleConfiguration: func(ctx context.Context, input *s3.PutBucketLifecycleConfigurationInput, opts []func(*s3.Options)) (*s3.PutBucketLifecycleConfigurationOutput, error) {
						return &s3.PutBucketLifecycleConfigurationOutput{}, nil
					},
				}),
			},
			want: want{
				err: nil,
			},
		},
		"SuccessfulCreate": {
			args: args{
				b: s3testing.Bucket(s3testing.WithLifecycleConfig(generateLifecycleConfig())),
				cl: NewLifecycleConfigurationClient(fake.MockBucketClient{
					MockPutBucketLifecycleConfiguration: func(ctx context.Context, input *s3.PutBucketLifecycleConfigurationInput, opts []func(*s3.Options)) (*s3.PutBucketLifecycleConfigurationOutput, error) {
						return &s3.PutBucketLifecycleConfigurationOutput{}, nil
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

func TestLifecycleDelete(t *testing.T) {
	type args struct {
		cl *LifecycleConfigurationClient
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
				b: s3testing.Bucket(s3testing.WithLifecycleConfig(generateLifecycleConfig())),
				cl: NewLifecycleConfigurationClient(fake.MockBucketClient{
					MockDeleteBucketLifecycle: func(ctx context.Context, input *s3.DeleteBucketLifecycleInput, opts []func(*s3.Options)) (*s3.DeleteBucketLifecycleOutput, error) {
						return nil, errBoom
					},
				}),
			},
			want: want{
				err: errorutils.Wrap(errBoom, lifecycleDeleteFailed),
			},
		},
		"SuccessfulDelete": {
			args: args{
				b: s3testing.Bucket(s3testing.WithLifecycleConfig(generateLifecycleConfig())),
				cl: NewLifecycleConfigurationClient(fake.MockBucketClient{
					MockDeleteBucketLifecycle: func(ctx context.Context, input *s3.DeleteBucketLifecycleInput, opts []func(*s3.Options)) (*s3.DeleteBucketLifecycleOutput, error) {
						return &s3.DeleteBucketLifecycleOutput{}, nil
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
			err := tc.args.cl.Delete(context.Background(), tc.args.b)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestLifecycleLateInit(t *testing.T) {
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
				cl: NewLifecycleConfigurationClient(fake.MockBucketClient{
					MockGetBucketLifecycleConfiguration: func(ctx context.Context, input *s3.GetBucketLifecycleConfigurationInput, opts []func(*s3.Options)) (*s3.GetBucketLifecycleConfigurationOutput, error) {
						return &s3.GetBucketLifecycleConfigurationOutput{}, errBoom
					},
				}),
			},
			want: want{
				err: errorutils.Wrap(errBoom, lifecycleGetFailed),
				cr:  s3testing.Bucket(),
			},
		},
		"ErrorLifecycleConfigurationNotFound": {
			args: args{
				b: s3testing.Bucket(),
				cl: NewLifecycleConfigurationClient(fake.MockBucketClient{
					MockGetBucketLifecycleConfiguration: func(ctx context.Context, input *s3.GetBucketLifecycleConfigurationInput, opts []func(*s3.Options)) (*s3.GetBucketLifecycleConfigurationOutput, error) {
						return &s3.GetBucketLifecycleConfigurationOutput{}, &smithy.GenericAPIError{Code: clients3.LifecycleNotFoundErrCode}
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
				cl: NewLifecycleConfigurationClient(fake.MockBucketClient{
					MockGetBucketLifecycleConfiguration: func(ctx context.Context, input *s3.GetBucketLifecycleConfigurationInput, opts []func(*s3.Options)) (*s3.GetBucketLifecycleConfigurationOutput, error) {
						return &s3.GetBucketLifecycleConfigurationOutput{Rules: make([]s3types.LifecycleRule, 0)}, nil
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
				b: s3testing.Bucket(s3testing.WithLifecycleConfig(nil)),
				cl: NewLifecycleConfigurationClient(fake.MockBucketClient{
					MockGetBucketLifecycleConfiguration: func(ctx context.Context, input *s3.GetBucketLifecycleConfigurationInput, opts []func(*s3.Options)) (*s3.GetBucketLifecycleConfigurationOutput, error) {
						return &s3.GetBucketLifecycleConfigurationOutput{Rules: generateAWSLifecycle(false).Rules}, nil
					},
				}),
			},
			want: want{
				err: nil,
				cr:  s3testing.Bucket(s3testing.WithLifecycleConfig(generateLifecycleConfig())),
			},
		},
		"NoOpLateInit": {
			args: args{
				b: s3testing.Bucket(s3testing.WithLifecycleConfig(generateLifecycleConfig())),
				cl: NewLifecycleConfigurationClient(fake.MockBucketClient{
					MockGetBucketLifecycleConfiguration: func(ctx context.Context, input *s3.GetBucketLifecycleConfigurationInput, opts []func(*s3.Options)) (*s3.GetBucketLifecycleConfigurationOutput, error) {
						return &s3.GetBucketLifecycleConfigurationOutput{Rules: []s3types.LifecycleRule{
							{},
						}}, nil
					},
				}),
			},
			want: want{
				err: nil,
				cr:  s3testing.Bucket(s3testing.WithLifecycleConfig(generateLifecycleConfig())),
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
