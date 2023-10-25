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
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"

	"github.com/crossplane-contrib/provider-aws/apis/s3/v1beta1"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/s3/fake"
	s3testing "github.com/crossplane-contrib/provider-aws/pkg/controller/s3/testing"
	errorutils "github.com/crossplane-contrib/provider-aws/pkg/utils/errors"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
)

var (
	mfadelete                   = "Enabled"
	_         SubresourceClient = &VersioningConfigurationClient{}
)

func generateVersioningConfig() *v1beta1.VersioningConfiguration {
	return &v1beta1.VersioningConfiguration{
		MFADelete: &mfadelete,
		Status:    pointer.ToOrNilIfZeroValue(enabled),
	}
}

func generateAWSVersioning() *s3types.VersioningConfiguration {
	return &s3types.VersioningConfiguration{
		MFADelete: s3types.MFADeleteEnabled,
		Status:    s3types.BucketVersioningStatusEnabled,
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
				b: s3testing.Bucket(s3testing.WithVersioningConfig(generateVersioningConfig())),
				cl: NewVersioningConfigurationClient(fake.MockBucketClient{
					MockGetBucketVersioning: func(ctx context.Context, input *s3.GetBucketVersioningInput, opts []func(*s3.Options)) (*s3.GetBucketVersioningOutput, error) {
						return nil, errBoom
					},
				}),
			},
			want: want{
				status: NeedsUpdate,
				err:    errorutils.Wrap(errBoom, versioningGetFailed),
			},
		},
		"UpdateNeededFull": {
			args: args{
				b: s3testing.Bucket(s3testing.WithVersioningConfig(generateVersioningConfig())),
				cl: NewVersioningConfigurationClient(fake.MockBucketClient{
					MockGetBucketVersioning: func(ctx context.Context, input *s3.GetBucketVersioningInput, opts []func(*s3.Options)) (*s3.GetBucketVersioningOutput, error) {
						return &s3.GetBucketVersioningOutput{}, nil
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
				b: s3testing.Bucket(s3testing.WithVersioningConfig(nil)),
				cl: NewVersioningConfigurationClient(fake.MockBucketClient{
					MockGetBucketVersioning: func(ctx context.Context, input *s3.GetBucketVersioningInput, opts []func(*s3.Options)) (*s3.GetBucketVersioningOutput, error) {
						return &s3.GetBucketVersioningOutput{}, nil
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
				b: s3testing.Bucket(s3testing.WithVersioningConfig(generateVersioningConfig())),
				cl: NewVersioningConfigurationClient(fake.MockBucketClient{
					MockGetBucketVersioning: func(ctx context.Context, input *s3.GetBucketVersioningInput, opts []func(*s3.Options)) (*s3.GetBucketVersioningOutput, error) {
						return &s3.GetBucketVersioningOutput{
							MFADelete: s3types.MFADeleteStatusEnabled,
							Status:    generateAWSVersioning().Status,
						}, nil
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
				b: s3testing.Bucket(s3testing.WithVersioningConfig(generateVersioningConfig())),
				cl: NewVersioningConfigurationClient(fake.MockBucketClient{
					MockPutBucketVersioning: func(ctx context.Context, input *s3.PutBucketVersioningInput, opts []func(*s3.Options)) (*s3.PutBucketVersioningOutput, error) {
						return nil, errBoom
					},
				}),
			},
			want: want{
				err: errorutils.Wrap(errBoom, versioningPutFailed),
			},
		},
		"InvalidConfig": {
			args: args{
				b: s3testing.Bucket(s3testing.WithVersioningConfig(generateVersioningConfig())),
				cl: NewVersioningConfigurationClient(fake.MockBucketClient{
					MockPutBucketVersioning: func(ctx context.Context, input *s3.PutBucketVersioningInput, opts []func(*s3.Options)) (*s3.PutBucketVersioningOutput, error) {
						return &s3.PutBucketVersioningOutput{}, nil
					},
				}),
			},
			want: want{
				err: nil,
			},
		},
		"SuccessfulCreate": {
			args: args{
				b: s3testing.Bucket(s3testing.WithVersioningConfig(generateVersioningConfig())),
				cl: NewVersioningConfigurationClient(fake.MockBucketClient{
					MockPutBucketVersioning: func(ctx context.Context, input *s3.PutBucketVersioningInput, opts []func(*s3.Options)) (*s3.PutBucketVersioningOutput, error) {
						return &s3.PutBucketVersioningOutput{}, nil
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

func TestVersioningLateInit(t *testing.T) {
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
				cl: NewVersioningConfigurationClient(fake.MockBucketClient{
					MockGetBucketVersioning: func(ctx context.Context, input *s3.GetBucketVersioningInput, opts []func(*s3.Options)) (*s3.GetBucketVersioningOutput, error) {
						return &s3.GetBucketVersioningOutput{}, errBoom
					},
				}),
			},
			want: want{
				err: errorutils.Wrap(errBoom, versioningGetFailed),
				cr:  s3testing.Bucket(),
			},
		},
		"NoLateInitNil": {
			args: args{
				b: s3testing.Bucket(),
				cl: NewVersioningConfigurationClient(fake.MockBucketClient{
					MockGetBucketVersioning: func(ctx context.Context, input *s3.GetBucketVersioningInput, opts []func(*s3.Options)) (*s3.GetBucketVersioningOutput, error) {
						return &s3.GetBucketVersioningOutput{}, nil
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
				b: s3testing.Bucket(s3testing.WithVersioningConfig(nil)),
				cl: NewVersioningConfigurationClient(fake.MockBucketClient{
					MockGetBucketVersioning: func(ctx context.Context, input *s3.GetBucketVersioningInput, opts []func(*s3.Options)) (*s3.GetBucketVersioningOutput, error) {
						return &s3.GetBucketVersioningOutput{
							MFADelete: s3types.MFADeleteStatusEnabled,
							Status:    s3types.BucketVersioningStatusEnabled,
						}, nil
					},
				}),
			},
			want: want{
				err: nil,
				cr:  s3testing.Bucket(s3testing.WithVersioningConfig(generateVersioningConfig())),
			},
		},
		"NoOpLateInit": {
			args: args{
				b: s3testing.Bucket(s3testing.WithVersioningConfig(generateVersioningConfig())),
				cl: NewVersioningConfigurationClient(fake.MockBucketClient{
					MockGetBucketVersioning: func(ctx context.Context, input *s3.GetBucketVersioningInput, opts []func(*s3.Options)) (*s3.GetBucketVersioningOutput, error) {
						return &s3.GetBucketVersioningOutput{
							MFADelete: s3types.MFADeleteStatusDisabled,
							Status:    s3types.BucketVersioningStatusSuspended,
						}, nil
					},
				}),
			},
			want: want{
				err: nil,
				cr:  s3testing.Bucket(s3testing.WithVersioningConfig(generateVersioningConfig())),
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
