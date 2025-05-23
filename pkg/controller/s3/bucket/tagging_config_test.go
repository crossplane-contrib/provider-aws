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

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"

	"github.com/crossplane-contrib/provider-aws/apis/s3/v1beta1"
	clientss3 "github.com/crossplane-contrib/provider-aws/pkg/clients/s3"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/s3/fake"
	s3testing "github.com/crossplane-contrib/provider-aws/pkg/controller/s3/testing"
	errorutils "github.com/crossplane-contrib/provider-aws/pkg/utils/errors"
)

var (
	tag = v1beta1.Tag{
		Key:   "test",
		Value: "value",
	}
	tag1 = v1beta1.Tag{
		Key:   "xyz",
		Value: "abc",
	}
	tag2 = v1beta1.Tag{
		Key:   "abc",
		Value: "abc",
	}
	tags   = []v1beta1.Tag{tag, tag1, tag2}
	awsTag = types.Tag{
		Key:   aws.String("test"),
		Value: aws.String("value"),
	}
	awsTag1 = types.Tag{
		Key:   aws.String("xyz"),
		Value: aws.String("abc"),
	}
	awsTag2 = types.Tag{
		Key:   aws.String("abc"),
		Value: aws.String("abc"),
	}
	awsSystemTag1 = types.Tag{
		Key:   aws.String("aws:tag1"),
		Value: aws.String("1"),
	}
	awsTags                   = []types.Tag{awsTag, awsTag1, awsTag2}
	_       SubresourceClient = &TaggingConfigurationClient{}
)

func generateTaggingConfig() *v1beta1.Tagging {
	return &v1beta1.Tagging{
		TagSet: tags,
	}
}

func generateAWSTagging() *types.Tagging {
	return &types.Tagging{
		TagSet: awsTags,
	}
}

func TestTaggingObserve(t *testing.T) {
	type args struct {
		cl *TaggingConfigurationClient
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
				b: s3testing.Bucket(s3testing.WithTaggingConfig(generateTaggingConfig())),
				cl: NewTaggingConfigurationClient(fake.MockBucketClient{
					MockGetBucketTagging: func(ctx context.Context, input *s3.GetBucketTaggingInput, opts []func(*s3.Options)) (*s3.GetBucketTaggingOutput, error) {
						return nil, errBoom
					},
				}),
			},
			want: want{
				status: NeedsUpdate,
				err:    errorutils.Wrap(errBoom, taggingGetFailed),
			},
		},
		"UpdateNeeded": {
			args: args{
				b: s3testing.Bucket(s3testing.WithTaggingConfig(generateTaggingConfig())),
				cl: NewTaggingConfigurationClient(fake.MockBucketClient{
					MockGetBucketTagging: func(ctx context.Context, input *s3.GetBucketTaggingInput, opts []func(*s3.Options)) (*s3.GetBucketTaggingOutput, error) {
						return &s3.GetBucketTaggingOutput{TagSet: nil}, nil
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
				b: s3testing.Bucket(s3testing.WithTaggingConfig(nil)),
				cl: NewTaggingConfigurationClient(fake.MockBucketClient{
					MockGetBucketTagging: func(ctx context.Context, input *s3.GetBucketTaggingInput, opts []func(*s3.Options)) (*s3.GetBucketTaggingOutput, error) {
						return &s3.GetBucketTaggingOutput{TagSet: generateAWSTagging().TagSet}, nil
					},
				}),
			},
			want: want{
				status: NeedsDeletion,
				err:    nil,
			},
		},
		"DeleteNoNeededIgnoreSystemTags": {
			args: args{
				b: s3testing.Bucket(s3testing.WithTaggingConfig(nil)),
				cl: NewTaggingConfigurationClient(fake.MockBucketClient{
					MockGetBucketTagging: func(ctx context.Context, input *s3.GetBucketTaggingInput, opts []func(*s3.Options)) (*s3.GetBucketTaggingOutput, error) {
						return &s3.GetBucketTaggingOutput{TagSet: []types.Tag{awsSystemTag1}}, nil
					},
				}),
			},
			want: want{
				status: Updated,
				err:    nil,
			},
		},
		"NoUpdateIgnoreSystemTags": {
			args: args{
				b: s3testing.Bucket(s3testing.WithTaggingConfig(generateTaggingConfig())),
				cl: NewTaggingConfigurationClient(fake.MockBucketClient{
					MockGetBucketTagging: func(ctx context.Context, input *s3.GetBucketTaggingInput, opts []func(*s3.Options)) (*s3.GetBucketTaggingOutput, error) {
						return &s3.GetBucketTaggingOutput{TagSet: []types.Tag{awsTag2, awsTag, awsTag1, awsSystemTag1}}, nil
					},
				}),
			},
			want: want{
				status: Updated,
				err:    nil,
			},
		},
		"NoUpdateNotExists": {
			args: args{
				b: s3testing.Bucket(s3testing.WithTaggingConfig(nil)),
				cl: NewTaggingConfigurationClient(fake.MockBucketClient{
					MockGetBucketTagging: func(ctx context.Context, input *s3.GetBucketTaggingInput, opts []func(*s3.Options)) (*s3.GetBucketTaggingOutput, error) {
						return nil, &smithy.GenericAPIError{Code: clientss3.TaggingNotFoundErrCode}
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
				b: s3testing.Bucket(s3testing.WithTaggingConfig(nil)),
				cl: NewTaggingConfigurationClient(fake.MockBucketClient{
					MockGetBucketTagging: func(ctx context.Context, input *s3.GetBucketTaggingInput, opts []func(*s3.Options)) (*s3.GetBucketTaggingOutput, error) {
						return &s3.GetBucketTaggingOutput{TagSet: nil}, nil
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
				b: s3testing.Bucket(s3testing.WithTaggingConfig(generateTaggingConfig())),
				cl: NewTaggingConfigurationClient(fake.MockBucketClient{
					MockGetBucketTagging: func(ctx context.Context, input *s3.GetBucketTaggingInput, opts []func(*s3.Options)) (*s3.GetBucketTaggingOutput, error) {
						return &s3.GetBucketTaggingOutput{TagSet: generateAWSTagging().TagSet}, nil
					},
				}),
			},
			want: want{
				status: Updated,
				err:    nil,
			},
		},
		"NoUpdateExistsOrder": {
			args: args{
				b: s3testing.Bucket(s3testing.WithTaggingConfig(generateTaggingConfig())),
				cl: NewTaggingConfigurationClient(fake.MockBucketClient{
					MockGetBucketTagging: func(ctx context.Context, input *s3.GetBucketTaggingInput, opts []func(*s3.Options)) (*s3.GetBucketTaggingOutput, error) {
						return &s3.GetBucketTaggingOutput{TagSet: []types.Tag{awsTag2, awsTag, awsTag1}}, nil
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

func TestTaggingCreateOrUpdate(t *testing.T) {
	type args struct {
		cl *TaggingConfigurationClient
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
				b: s3testing.Bucket(s3testing.WithTaggingConfig(generateTaggingConfig())),
				cl: NewTaggingConfigurationClient(fake.MockBucketClient{
					MockPutBucketTagging: func(ctx context.Context, input *s3.PutBucketTaggingInput, opts []func(*s3.Options)) (*s3.PutBucketTaggingOutput, error) {
						return nil, errBoom
					},
					MockGetBucketTagging: func(ctx context.Context, input *s3.GetBucketTaggingInput, opts []func(*s3.Options)) (*s3.GetBucketTaggingOutput, error) {
						return &s3.GetBucketTaggingOutput{TagSet: nil}, nil
					},
				}),
			},
			want: want{
				err: errorutils.Wrap(errBoom, taggingPutFailed),
			},
		},
		"InvalidConfig": {
			args: args{
				b: s3testing.Bucket(s3testing.WithTaggingConfig(generateTaggingConfig())),
				cl: NewTaggingConfigurationClient(fake.MockBucketClient{
					MockPutBucketTagging: func(ctx context.Context, input *s3.PutBucketTaggingInput, opts []func(*s3.Options)) (*s3.PutBucketTaggingOutput, error) {
						return &s3.PutBucketTaggingOutput{}, nil
					},
					MockGetBucketTagging: func(ctx context.Context, input *s3.GetBucketTaggingInput, opts []func(*s3.Options)) (*s3.GetBucketTaggingOutput, error) {
						return &s3.GetBucketTaggingOutput{TagSet: nil}, nil
					},
				}),
			},
			want: want{
				err: nil,
			},
		},
		"SuccessfulCreate": {
			args: args{
				b: s3testing.Bucket(s3testing.WithTaggingConfig(generateTaggingConfig())),
				cl: NewTaggingConfigurationClient(fake.MockBucketClient{
					MockPutBucketTagging: func(ctx context.Context, input *s3.PutBucketTaggingInput, opts []func(*s3.Options)) (*s3.PutBucketTaggingOutput, error) {
						return &s3.PutBucketTaggingOutput{}, nil
					},
					MockGetBucketTagging: func(ctx context.Context, input *s3.GetBucketTaggingInput, opts []func(*s3.Options)) (*s3.GetBucketTaggingOutput, error) {
						return &s3.GetBucketTaggingOutput{TagSet: nil}, nil
					},
				}),
			},
			want: want{
				err: nil,
			},
		},
		"SuccessfulCreateNoExistingTags": {
			args: args{
				b: s3testing.Bucket(s3testing.WithTaggingConfig(generateTaggingConfig())),
				cl: NewTaggingConfigurationClient(fake.MockBucketClient{
					MockPutBucketTagging: func(ctx context.Context, input *s3.PutBucketTaggingInput, opts []func(*s3.Options)) (*s3.PutBucketTaggingOutput, error) {
						return &s3.PutBucketTaggingOutput{}, nil
					},
					MockGetBucketTagging: func(ctx context.Context, input *s3.GetBucketTaggingInput, opts []func(*s3.Options)) (*s3.GetBucketTaggingOutput, error) {
						return nil, &smithy.GenericAPIError{Code: clientss3.TaggingNotFoundErrCode}
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

func TestTaggingDelete(t *testing.T) {
	type args struct {
		cl *TaggingConfigurationClient
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
				b: s3testing.Bucket(s3testing.WithTaggingConfig(generateTaggingConfig())),
				cl: NewTaggingConfigurationClient(fake.MockBucketClient{
					MockDeleteBucketTagging: func(ctx context.Context, input *s3.DeleteBucketTaggingInput, opts []func(*s3.Options)) (*s3.DeleteBucketTaggingOutput, error) {
						return nil, errBoom
					},
				}),
			},
			want: want{
				err: errorutils.Wrap(errBoom, taggingDeleteFailed),
			},
		},
		"SuccessfulDelete": {
			args: args{
				b: s3testing.Bucket(s3testing.WithTaggingConfig(generateTaggingConfig())),
				cl: NewTaggingConfigurationClient(fake.MockBucketClient{
					MockDeleteBucketTagging: func(ctx context.Context, input *s3.DeleteBucketTaggingInput, opts []func(*s3.Options)) (*s3.DeleteBucketTaggingOutput, error) {
						return &s3.DeleteBucketTaggingOutput{}, nil
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

func TestTaggingLateInit(t *testing.T) {
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
				cl: NewTaggingConfigurationClient(fake.MockBucketClient{
					MockGetBucketTagging: func(ctx context.Context, input *s3.GetBucketTaggingInput, opts []func(*s3.Options)) (*s3.GetBucketTaggingOutput, error) {
						return nil, errBoom
					},
				}),
			},
			want: want{
				err: errorutils.Wrap(errBoom, taggingGetFailed),
				cr:  s3testing.Bucket(),
			},
		},
		"ErrorTaggingNotFound": {
			args: args{
				b: s3testing.Bucket(),
				cl: NewTaggingConfigurationClient(fake.MockBucketClient{
					MockGetBucketTagging: func(ctx context.Context, input *s3.GetBucketTaggingInput, opts []func(*s3.Options)) (*s3.GetBucketTaggingOutput, error) {
						return nil, &smithy.GenericAPIError{Code: clientss3.TaggingNotFoundErrCode}
					},
				}),
			},
			want: want{
				err: nil,
				cr:  s3testing.Bucket(),
			},
		},
		"NoLateInitNil": {
			args: args{
				b: s3testing.Bucket(),
				cl: NewTaggingConfigurationClient(fake.MockBucketClient{
					MockGetBucketTagging: func(ctx context.Context, input *s3.GetBucketTaggingInput, opts []func(*s3.Options)) (*s3.GetBucketTaggingOutput, error) {
						return &s3.GetBucketTaggingOutput{TagSet: nil}, nil
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
				cl: NewTaggingConfigurationClient(fake.MockBucketClient{
					MockGetBucketTagging: func(ctx context.Context, input *s3.GetBucketTaggingInput, opts []func(*s3.Options)) (*s3.GetBucketTaggingOutput, error) {
						return &s3.GetBucketTaggingOutput{TagSet: []types.Tag{}}, nil
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
				b: s3testing.Bucket(s3testing.WithTaggingConfig(nil)),
				cl: NewTaggingConfigurationClient(fake.MockBucketClient{
					MockGetBucketTagging: func(ctx context.Context, input *s3.GetBucketTaggingInput, opts []func(*s3.Options)) (*s3.GetBucketTaggingOutput, error) {
						return &s3.GetBucketTaggingOutput{TagSet: generateAWSTagging().TagSet}, nil
					},
				}),
			},
			want: want{
				err: nil,
				cr:  s3testing.Bucket(s3testing.WithTaggingConfig(generateTaggingConfig())),
			},
		},
		"NoOpLateInit": {
			args: args{
				b: s3testing.Bucket(s3testing.WithTaggingConfig(generateTaggingConfig())),
				cl: NewTaggingConfigurationClient(fake.MockBucketClient{
					MockGetBucketTagging: func(ctx context.Context, input *s3.GetBucketTaggingInput, opts []func(*s3.Options)) (*s3.GetBucketTaggingOutput, error) {
						return &s3.GetBucketTaggingOutput{TagSet: []types.Tag{}}, nil
					},
				}),
			},
			want: want{
				err: nil,
				cr:  s3testing.Bucket(s3testing.WithTaggingConfig(generateTaggingConfig())),
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
