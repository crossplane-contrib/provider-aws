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

	"github.com/aws/aws-sdk-go-v2/aws/awserr"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/crossplane/provider-aws/apis/s3/v1beta1"
	awsclient "github.com/crossplane/provider-aws/pkg/clients"
	clients3 "github.com/crossplane/provider-aws/pkg/clients/s3"
	"github.com/crossplane/provider-aws/pkg/clients/s3/fake"
	s3Testing "github.com/crossplane/provider-aws/pkg/controller/s3/testing"
)

var (
	days        = 1
	location, _ = time.LoadLocation("UTC")
	date        = metav1.Date(2020, time.September, 25, 11, 40, 0, 0, location)
	awsDate     = time.Date(2020, time.September, 25, 11, 40, 0, 0, location)
	marker      = false
	prefix      = "test-"
	id          = "test-id"
	storage     = "ONEZONE_IA"
)

var _ SubresourceClient = &LifecycleConfigurationClient{}

func generateLifecycleConfig() *v1beta1.BucketLifecycleConfiguration {
	return &v1beta1.BucketLifecycleConfiguration{
		Rules: []v1beta1.LifecycleRule{
			{
				AbortIncompleteMultipartUpload: &v1beta1.AbortIncompleteMultipartUpload{DaysAfterInitiation: 1},
				Expiration: &v1beta1.LifecycleExpiration{
					Date:                      &date,
					Days:                      awsclient.Int64(days),
					ExpiredObjectDeleteMarker: awsclient.Bool(marker),
				},
				Filter: &v1beta1.LifecycleRuleFilter{
					And: &v1beta1.LifecycleRuleAndOperator{
						Prefix: awsclient.String(prefix),
						Tags:   tags,
					},
					Prefix: awsclient.String(prefix),
					Tag:    &tag,
				},
				ID:                          awsclient.String(id),
				NoncurrentVersionExpiration: &v1beta1.NoncurrentVersionExpiration{NoncurrentDays: awsclient.Int64(days)},
				NoncurrentVersionTransitions: []v1beta1.NoncurrentVersionTransition{{
					NoncurrentDays: awsclient.Int64(days),
					StorageClass:   storage,
				}},
				Status: enabled,
				Transitions: []v1beta1.Transition{{
					Date:         &date,
					Days:         awsclient.Int64(days),
					StorageClass: storage,
				}},
			},
		},
	}
}

func generateAWSLifecycle(sortTag bool) *s3.BucketLifecycleConfiguration {
	conf := &s3.BucketLifecycleConfiguration{
		Rules: []s3.LifecycleRule{
			{
				AbortIncompleteMultipartUpload: &s3.AbortIncompleteMultipartUpload{DaysAfterInitiation: awsclient.Int64(1)},
				Expiration: &s3.LifecycleExpiration{
					Date:                      &awsDate,
					Days:                      awsclient.Int64(days),
					ExpiredObjectDeleteMarker: awsclient.Bool(marker),
				},
				Filter: &s3.LifecycleRuleFilter{
					And: &s3.LifecycleRuleAndOperator{
						Prefix: awsclient.String(prefix),
						Tags:   awsTags,
					},
					Prefix: awsclient.String(prefix),
					Tag:    &awsTag,
				},
				ID:                          awsclient.String(id),
				NoncurrentVersionExpiration: &s3.NoncurrentVersionExpiration{NoncurrentDays: awsclient.Int64(days)},
				NoncurrentVersionTransitions: []s3.NoncurrentVersionTransition{{
					NoncurrentDays: awsclient.Int64(days),
					StorageClass:   s3.TransitionStorageClassOnezoneIa,
				}},
				Status: s3.ExpirationStatusEnabled,
				Transitions: []s3.Transition{{
					Date:         &awsDate,
					Days:         awsclient.Int64(days),
					StorageClass: s3.TransitionStorageClassOnezoneIa,
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
		input []s3.LifecycleRule
	}

	cases := map[string]struct {
		args
		want
	}{
		"SameStruct": {
			args: args{
				b: s3Testing.Bucket(s3Testing.WithLifecycleConfig(generateLifecycleConfig())),
			},
			want: want{
				input: generateAWSLifecycle(true).Rules,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			generated := GenerateLifecycleRules(tc.args.b.Spec.ForProvider.LifecycleConfiguration.Rules)
			if diff := cmp.Diff(generated, tc.want.input); diff != "" {
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
				b: s3Testing.Bucket(s3Testing.WithLifecycleConfig(generateLifecycleConfig())),
				cl: NewLifecycleConfigurationClient(fake.MockBucketClient{
					MockGetBucketLifecycleConfigurationRequest: func(input *s3.GetBucketLifecycleConfigurationInput) s3.GetBucketLifecycleConfigurationRequest {
						return s3.GetBucketLifecycleConfigurationRequest{
							Request: s3Testing.CreateRequest(errBoom, &s3.GetBucketLifecycleConfigurationOutput{Rules: generateAWSLifecycle(false).Rules}),
						}
					},
				}),
			},
			want: want{
				status: NeedsUpdate,
				err:    awsclient.Wrap(errBoom, lifecycleGetFailed),
			},
		},
		"UpdateNeeded": {
			args: args{
				b: s3Testing.Bucket(s3Testing.WithLifecycleConfig(generateLifecycleConfig())),
				cl: NewLifecycleConfigurationClient(fake.MockBucketClient{
					MockGetBucketLifecycleConfigurationRequest: func(input *s3.GetBucketLifecycleConfigurationInput) s3.GetBucketLifecycleConfigurationRequest {
						return s3.GetBucketLifecycleConfigurationRequest{
							Request: s3Testing.CreateRequest(nil, &s3.GetBucketLifecycleConfigurationOutput{Rules: nil}),
						}
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
				b: s3Testing.Bucket(s3Testing.WithLifecycleConfig(nil)),
				cl: NewLifecycleConfigurationClient(fake.MockBucketClient{
					MockGetBucketLifecycleConfigurationRequest: func(input *s3.GetBucketLifecycleConfigurationInput) s3.GetBucketLifecycleConfigurationRequest {
						return s3.GetBucketLifecycleConfigurationRequest{
							Request: s3Testing.CreateRequest(nil, &s3.GetBucketLifecycleConfigurationOutput{Rules: generateAWSLifecycle(false).Rules}),
						}
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
				b: s3Testing.Bucket(s3Testing.WithLifecycleConfig(nil)),
				cl: NewLifecycleConfigurationClient(fake.MockBucketClient{
					MockGetBucketLifecycleConfigurationRequest: func(input *s3.GetBucketLifecycleConfigurationInput) s3.GetBucketLifecycleConfigurationRequest {
						return s3.GetBucketLifecycleConfigurationRequest{
							Request: s3Testing.CreateRequest(awserr.New(clients3.LifecycleNotFoundErrCode, "", nil), &s3.GetBucketLifecycleConfigurationOutput{Rules: nil}),
						}
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
				b: s3Testing.Bucket(s3Testing.WithLifecycleConfig(nil)),
				cl: NewLifecycleConfigurationClient(fake.MockBucketClient{
					MockGetBucketLifecycleConfigurationRequest: func(input *s3.GetBucketLifecycleConfigurationInput) s3.GetBucketLifecycleConfigurationRequest {
						return s3.GetBucketLifecycleConfigurationRequest{
							Request: s3Testing.CreateRequest(nil, &s3.GetBucketLifecycleConfigurationOutput{Rules: nil}),
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
				b: s3Testing.Bucket(s3Testing.WithLifecycleConfig(generateLifecycleConfig())),
				cl: NewLifecycleConfigurationClient(fake.MockBucketClient{
					MockGetBucketLifecycleConfigurationRequest: func(input *s3.GetBucketLifecycleConfigurationInput) s3.GetBucketLifecycleConfigurationRequest {
						return s3.GetBucketLifecycleConfigurationRequest{
							Request: s3Testing.CreateRequest(nil, &s3.GetBucketLifecycleConfigurationOutput{Rules: generateAWSLifecycle(false).Rules}),
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
				b: s3Testing.Bucket(s3Testing.WithLifecycleConfig(generateLifecycleConfig())),
				cl: NewLifecycleConfigurationClient(fake.MockBucketClient{
					MockPutBucketLifecycleConfigurationRequest: func(input *s3.PutBucketLifecycleConfigurationInput) s3.PutBucketLifecycleConfigurationRequest {
						return s3.PutBucketLifecycleConfigurationRequest{
							Request: s3Testing.CreateRequest(errBoom, &s3.PutBucketLifecycleConfigurationOutput{}),
						}
					},
				}),
			},
			want: want{
				err: awsclient.Wrap(errBoom, lifecyclePutFailed),
			},
		},
		"InvalidConfig": {
			args: args{
				b: s3Testing.Bucket(s3Testing.WithLifecycleConfig(generateLifecycleConfig())),
				cl: NewLifecycleConfigurationClient(fake.MockBucketClient{
					MockPutBucketLifecycleConfigurationRequest: func(input *s3.PutBucketLifecycleConfigurationInput) s3.PutBucketLifecycleConfigurationRequest {
						return s3.PutBucketLifecycleConfigurationRequest{
							Request: s3Testing.CreateRequest(nil, &s3.PutBucketLifecycleConfigurationOutput{}),
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
				b: s3Testing.Bucket(s3Testing.WithLifecycleConfig(generateLifecycleConfig())),
				cl: NewLifecycleConfigurationClient(fake.MockBucketClient{
					MockPutBucketLifecycleConfigurationRequest: func(input *s3.PutBucketLifecycleConfigurationInput) s3.PutBucketLifecycleConfigurationRequest {
						return s3.PutBucketLifecycleConfigurationRequest{
							Request: s3Testing.CreateRequest(nil, &s3.PutBucketLifecycleConfigurationOutput{}),
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
				b: s3Testing.Bucket(s3Testing.WithLifecycleConfig(generateLifecycleConfig())),
				cl: NewLifecycleConfigurationClient(fake.MockBucketClient{
					MockDeleteBucketLifecycleRequest: func(input *s3.DeleteBucketLifecycleInput) s3.DeleteBucketLifecycleRequest {
						return s3.DeleteBucketLifecycleRequest{
							Request: s3Testing.CreateRequest(errBoom, &s3.DeleteBucketLifecycleOutput{}),
						}
					},
				}),
			},
			want: want{
				err: awsclient.Wrap(errBoom, lifecycleDeleteFailed),
			},
		},
		"SuccessfulDelete": {
			args: args{
				b: s3Testing.Bucket(s3Testing.WithLifecycleConfig(generateLifecycleConfig())),
				cl: NewLifecycleConfigurationClient(fake.MockBucketClient{
					MockDeleteBucketLifecycleRequest: func(input *s3.DeleteBucketLifecycleInput) s3.DeleteBucketLifecycleRequest {
						return s3.DeleteBucketLifecycleRequest{
							Request: s3Testing.CreateRequest(nil, &s3.DeleteBucketLifecycleOutput{}),
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
				b: s3Testing.Bucket(),
				cl: NewLifecycleConfigurationClient(fake.MockBucketClient{
					MockGetBucketLifecycleConfigurationRequest: func(input *s3.GetBucketLifecycleConfigurationInput) s3.GetBucketLifecycleConfigurationRequest {
						return s3.GetBucketLifecycleConfigurationRequest{
							Request: s3Testing.CreateRequest(errBoom, &s3.GetBucketLifecycleConfigurationOutput{}),
						}
					},
				}),
			},
			want: want{
				err: awsclient.Wrap(errBoom, lifecycleGetFailed),
				cr:  s3Testing.Bucket(),
			},
		},
		"ErrorLifecycleConfigurationNotFound": {
			args: args{
				b: s3Testing.Bucket(),
				cl: NewLifecycleConfigurationClient(fake.MockBucketClient{
					MockGetBucketLifecycleConfigurationRequest: func(input *s3.GetBucketLifecycleConfigurationInput) s3.GetBucketLifecycleConfigurationRequest {
						return s3.GetBucketLifecycleConfigurationRequest{
							Request: s3Testing.CreateRequest(awserr.New(clients3.LifecycleNotFoundErrCode, "error", nil), &s3.GetBucketLifecycleConfigurationOutput{}),
						}
					},
				}),
			},
			want: want{
				err: nil,
				cr:  s3Testing.Bucket(),
			},
		},
		"NoLateInitEmpty": {
			args: args{
				b: s3Testing.Bucket(),
				cl: NewLifecycleConfigurationClient(fake.MockBucketClient{
					MockGetBucketLifecycleConfigurationRequest: func(input *s3.GetBucketLifecycleConfigurationInput) s3.GetBucketLifecycleConfigurationRequest {
						return s3.GetBucketLifecycleConfigurationRequest{
							Request: s3Testing.CreateRequest(nil, &s3.GetBucketLifecycleConfigurationOutput{
								Rules: make([]s3.LifecycleRule, 0),
							}),
						}
					},
				}),
			},
			want: want{
				err: nil,
				cr:  s3Testing.Bucket(),
			},
		},
		"SuccessfulLateInit": {
			args: args{
				b: s3Testing.Bucket(s3Testing.WithLifecycleConfig(nil)),
				cl: NewLifecycleConfigurationClient(fake.MockBucketClient{
					MockGetBucketLifecycleConfigurationRequest: func(input *s3.GetBucketLifecycleConfigurationInput) s3.GetBucketLifecycleConfigurationRequest {
						return s3.GetBucketLifecycleConfigurationRequest{
							Request: s3Testing.CreateRequest(nil, &s3.GetBucketLifecycleConfigurationOutput{Rules: generateAWSLifecycle(false).Rules}),
						}
					},
				}),
			},
			want: want{
				err: nil,
				cr:  s3Testing.Bucket(s3Testing.WithLifecycleConfig(generateLifecycleConfig())),
			},
		},
		"NoOpLateInit": {
			args: args{
				b: s3Testing.Bucket(s3Testing.WithLifecycleConfig(generateLifecycleConfig())),
				cl: NewLifecycleConfigurationClient(fake.MockBucketClient{
					MockGetBucketLifecycleConfigurationRequest: func(input *s3.GetBucketLifecycleConfigurationInput) s3.GetBucketLifecycleConfigurationRequest {
						return s3.GetBucketLifecycleConfigurationRequest{
							Request: s3Testing.CreateRequest(nil, &s3.GetBucketLifecycleConfigurationOutput{Rules: []s3.LifecycleRule{
								{},
							}}),
						}
					},
				}),
			},
			want: want{
				err: nil,
				cr:  s3Testing.Bucket(s3Testing.WithLifecycleConfig(generateLifecycleConfig())),
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
