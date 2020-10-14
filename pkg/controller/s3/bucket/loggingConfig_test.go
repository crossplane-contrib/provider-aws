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
	_           SubresourceClient = &LoggingConfigurationClient{}
	bucketName                    = "test.Bucket.name"
	permission                    = "FULL_CONTROL"
	displayName                   = "name"
	email                         = "test@user.com"
	userType                      = "CanonicalUser"
	groupURI                      = "uri"
)

func generateLoggingConfig() *v1beta1.LoggingConfiguration {
	return &v1beta1.LoggingConfiguration{
		TargetBucket: &bucketName,
		TargetPrefix: prefix,
		TargetGrants: []v1beta1.TargetGrant{{
			Grantee: v1beta1.TargetGrantee{
				DisplayName:  &displayName,
				EmailAddress: &email,
				ID:           &id,
				Type:         userType,
				URI:          &groupURI,
			},
			Permission: permission,
		}},
	}
}

func generateAWSLogging() *s3.LoggingEnabled {
	return &s3.LoggingEnabled{
		TargetBucket: &bucketName,
		TargetGrants: []s3.TargetGrant{{
			Grantee: &s3.Grantee{
				DisplayName:  &displayName,
				EmailAddress: &email,
				ID:           &id,
				Type:         s3.TypeCanonicalUser,
				URI:          &groupURI,
			},
			Permission: s3.BucketLogsPermissionFullControl,
		}},
		TargetPrefix: &prefix,
	}
}

func TestLoggingObserve(t *testing.T) {
	type args struct {
		cl *LoggingConfigurationClient
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
				b: s3Testing.Bucket(s3Testing.WithLoggingConfig(generateLoggingConfig())),
				cl: NewLoggingConfigurationClient(fake.MockBucketClient{
					MockGetBucketLoggingRequest: func(input *s3.GetBucketLoggingInput) s3.GetBucketLoggingRequest {
						return s3.GetBucketLoggingRequest{
							Request: s3Testing.CreateRequest(errBoom, &s3.GetBucketLoggingOutput{LoggingEnabled: generateAWSLogging()}),
						}
					},
				}),
			},
			want: want{
				status: NeedsUpdate,
				err:    errors.Wrap(errBoom, loggingGetFailed),
			},
		},
		"UpdateNeeded": {
			args: args{
				b: s3Testing.Bucket(s3Testing.WithLoggingConfig(generateLoggingConfig())),
				cl: NewLoggingConfigurationClient(fake.MockBucketClient{
					MockGetBucketLoggingRequest: func(input *s3.GetBucketLoggingInput) s3.GetBucketLoggingRequest {
						return s3.GetBucketLoggingRequest{
							Request: s3Testing.CreateRequest(nil, &s3.GetBucketLoggingOutput{LoggingEnabled: nil}),
						}
					},
				}),
			},
			want: want{
				status: NeedsUpdate,
				err:    nil,
			},
		},
		"NoUpdateNotExists": {
			args: args{
				b: s3Testing.Bucket(s3Testing.WithLoggingConfig(nil)),
				cl: NewLoggingConfigurationClient(fake.MockBucketClient{
					MockGetBucketLoggingRequest: func(input *s3.GetBucketLoggingInput) s3.GetBucketLoggingRequest {
						return s3.GetBucketLoggingRequest{
							Request: s3Testing.CreateRequest(nil, &s3.GetBucketLoggingOutput{LoggingEnabled: nil}),
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
				b: s3Testing.Bucket(s3Testing.WithLoggingConfig(generateLoggingConfig())),
				cl: NewLoggingConfigurationClient(fake.MockBucketClient{
					MockGetBucketLoggingRequest: func(input *s3.GetBucketLoggingInput) s3.GetBucketLoggingRequest {
						return s3.GetBucketLoggingRequest{
							Request: s3Testing.CreateRequest(nil, &s3.GetBucketLoggingOutput{LoggingEnabled: generateAWSLogging()}),
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

func TestLoggingCreateOrUpdate(t *testing.T) {
	type args struct {
		cl *LoggingConfigurationClient
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
				b: s3Testing.Bucket(s3Testing.WithLoggingConfig(generateLoggingConfig())),
				cl: NewLoggingConfigurationClient(fake.MockBucketClient{
					MockPutBucketLoggingRequest: func(input *s3.PutBucketLoggingInput) s3.PutBucketLoggingRequest {
						return s3.PutBucketLoggingRequest{
							Request: s3Testing.CreateRequest(errBoom, &s3.PutBucketLoggingOutput{}),
						}
					},
				}),
			},
			want: want{
				err: errors.Wrap(errBoom, loggingPutFailed),
			},
		},
		"InvalidConfig": {
			args: args{
				b: s3Testing.Bucket(s3Testing.WithLoggingConfig(generateLoggingConfig())),
				cl: NewLoggingConfigurationClient(fake.MockBucketClient{
					MockPutBucketLoggingRequest: func(input *s3.PutBucketLoggingInput) s3.PutBucketLoggingRequest {
						return s3.PutBucketLoggingRequest{
							Request: s3Testing.CreateRequest(nil, &s3.PutBucketLoggingOutput{}),
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
				b: s3Testing.Bucket(s3Testing.WithLoggingConfig(generateLoggingConfig())),
				cl: NewLoggingConfigurationClient(fake.MockBucketClient{
					MockPutBucketLoggingRequest: func(input *s3.PutBucketLoggingInput) s3.PutBucketLoggingRequest {
						return s3.PutBucketLoggingRequest{
							Request: s3Testing.CreateRequest(nil, &s3.PutBucketLoggingOutput{}),
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
