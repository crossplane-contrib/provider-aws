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
	aws "github.com/crossplane/provider-aws/pkg/clients"
	"github.com/crossplane/provider-aws/pkg/clients/s3/fake"
	s3Testing "github.com/crossplane/provider-aws/pkg/controller/s3/testing"
)

var (
	mfadelete                   = "Enabled"
	_         SubresourceClient = &VersioningConfigurationClient{}
)

func generateVersioningConfig() *v1beta1.VersioningConfiguration {
	return &v1beta1.VersioningConfiguration{
		MFADelete: &mfadelete,
		Status:    aws.String(enabled),
	}
}

func generateAWSVersioning() *s3.VersioningConfiguration {
	return &s3.VersioningConfiguration{
		MFADelete: s3.MFADeleteEnabled,
		Status:    s3.BucketVersioningStatusEnabled,
	}
}

func TestVersioningObserve(t *testing.T) {
	type args struct {
		cl *VersioningConfigurationClient
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
				b: s3Testing.Bucket(s3Testing.WithVersioningConfig(generateVersioningConfig())),
				cl: NewVersioningConfigurationClient(fake.MockBucketClient{
					MockGetBucketVersioningRequest: func(input *s3.GetBucketVersioningInput) s3.GetBucketVersioningRequest {
						return s3.GetBucketVersioningRequest{
							Request: s3Testing.CreateRequest(errBoom, &s3.GetBucketVersioningOutput{}),
						}
					},
				}),
			},
			want: want{
				status: NeedsUpdate,
				err:    errors.Wrap(errBoom, versioningGetFailed),
			},
		},
		"UpdateNeededFull": {
			args: args{
				b: s3Testing.Bucket(s3Testing.WithVersioningConfig(generateVersioningConfig())),
				cl: NewVersioningConfigurationClient(fake.MockBucketClient{
					MockGetBucketVersioningRequest: func(input *s3.GetBucketVersioningInput) s3.GetBucketVersioningRequest {
						return s3.GetBucketVersioningRequest{
							Request: s3Testing.CreateRequest(nil, &s3.GetBucketVersioningOutput{}),
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
				b: s3Testing.Bucket(s3Testing.WithVersioningConfig(nil)),
				cl: NewVersioningConfigurationClient(fake.MockBucketClient{
					MockGetBucketVersioningRequest: func(input *s3.GetBucketVersioningInput) s3.GetBucketVersioningRequest {
						return s3.GetBucketVersioningRequest{
							Request: s3Testing.CreateRequest(nil, &s3.GetBucketVersioningOutput{
								MFADelete: s3.MFADeleteStatusEnabled,
								Status:    generateAWSVersioning().Status,
							}),
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
				b: s3Testing.Bucket(s3Testing.WithVersioningConfig(nil)),
				cl: NewVersioningConfigurationClient(fake.MockBucketClient{
					MockGetBucketVersioningRequest: func(input *s3.GetBucketVersioningInput) s3.GetBucketVersioningRequest {
						return s3.GetBucketVersioningRequest{
							Request: s3Testing.CreateRequest(nil, &s3.GetBucketVersioningOutput{}),
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
				b: s3Testing.Bucket(s3Testing.WithVersioningConfig(generateVersioningConfig())),
				cl: NewVersioningConfigurationClient(fake.MockBucketClient{
					MockGetBucketVersioningRequest: func(input *s3.GetBucketVersioningInput) s3.GetBucketVersioningRequest {
						return s3.GetBucketVersioningRequest{
							Request: s3Testing.CreateRequest(nil, &s3.GetBucketVersioningOutput{
								MFADelete: s3.MFADeleteStatusEnabled,
								Status:    generateAWSVersioning().Status,
							}),
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

func TestVersioningCreateOrUpdate(t *testing.T) {
	type args struct {
		cl *VersioningConfigurationClient
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
				b: s3Testing.Bucket(s3Testing.WithVersioningConfig(generateVersioningConfig())),
				cl: NewVersioningConfigurationClient(fake.MockBucketClient{
					MockPutBucketVersioningRequest: func(input *s3.PutBucketVersioningInput) s3.PutBucketVersioningRequest {
						return s3.PutBucketVersioningRequest{
							Request: s3Testing.CreateRequest(errBoom, &s3.PutBucketVersioningOutput{}),
						}
					},
				}),
			},
			want: want{
				err: errors.Wrap(errBoom, versioningPutFailed),
			},
		},
		"InvalidConfig": {
			args: args{
				b: s3Testing.Bucket(s3Testing.WithVersioningConfig(generateVersioningConfig())),
				cl: NewVersioningConfigurationClient(fake.MockBucketClient{
					MockPutBucketVersioningRequest: func(input *s3.PutBucketVersioningInput) s3.PutBucketVersioningRequest {
						return s3.PutBucketVersioningRequest{
							Request: s3Testing.CreateRequest(nil, &s3.PutBucketVersioningOutput{}),
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
				b: s3Testing.Bucket(s3Testing.WithVersioningConfig(generateVersioningConfig())),
				cl: NewVersioningConfigurationClient(fake.MockBucketClient{
					MockPutBucketVersioningRequest: func(input *s3.PutBucketVersioningInput) s3.PutBucketVersioningRequest {
						return s3.PutBucketVersioningRequest{
							Request: s3Testing.CreateRequest(nil, &s3.PutBucketVersioningOutput{}),
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

func TestVersioningDelete(t *testing.T) {
	type args struct {
		cl *VersioningConfigurationClient
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
				b: s3Testing.Bucket(s3Testing.WithVersioningConfig(generateVersioningConfig())),
				cl: NewVersioningConfigurationClient(fake.MockBucketClient{
					MockPutBucketVersioningRequest: func(input *s3.PutBucketVersioningInput) s3.PutBucketVersioningRequest {
						return s3.PutBucketVersioningRequest{
							Request: s3Testing.CreateRequest(errBoom, &s3.PutBucketVersioningOutput{}),
						}
					},
				}),
			},
			want: want{
				err: errors.Wrap(errBoom, versioningDeleteFailed),
			},
		},
		"SuccessfulDelete": {
			args: args{
				b: s3Testing.Bucket(s3Testing.WithVersioningConfig(generateVersioningConfig())),
				cl: NewVersioningConfigurationClient(fake.MockBucketClient{
					MockPutBucketVersioningRequest: func(input *s3.PutBucketVersioningInput) s3.PutBucketVersioningRequest {
						return s3.PutBucketVersioningRequest{
							Request: s3Testing.CreateRequest(nil, &s3.PutBucketVersioningOutput{}),
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
