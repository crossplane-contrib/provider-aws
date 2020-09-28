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

package bucketresources

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws/awserr"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	"golang.org/x/net/context"

	"github.com/crossplane/provider-aws/apis/s3/v1beta1"
	clients3 "github.com/crossplane/provider-aws/pkg/clients/s3"
	"github.com/crossplane/provider-aws/pkg/clients/s3/fake"
	s3Testing "github.com/crossplane/provider-aws/pkg/controller/s3/testing"
)

func generateTaggingConfig() *v1beta1.Tagging {
	return &v1beta1.Tagging{
		TagSet: tags,
	}
}

func generateAWSTagging() *s3.Tagging {
	return &s3.Tagging{
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
				b: s3Testing.Bucket(s3Testing.WithTaggingConfig(generateTaggingConfig())),
				cl: NewTaggingConfigurationClient(
					s3Testing.Bucket(s3Testing.WithTaggingConfig(generateTaggingConfig())),
					fake.MockBucketClient{
						MockGetBucketTaggingRequest: func(input *s3.GetBucketTaggingInput) s3.GetBucketTaggingRequest {
							return s3.GetBucketTaggingRequest{
								Request: s3Testing.CreateRequest(errBoom, &s3.GetBucketTaggingOutput{}),
							}
						},
					},
				),
			},
			want: want{
				status: NeedsUpdate,
				err:    errors.Wrap(errBoom, taggingGetFailed),
			},
		},
		"UpdateNeeded": {
			args: args{
				b: s3Testing.Bucket(s3Testing.WithTaggingConfig(generateTaggingConfig())),
				cl: NewTaggingConfigurationClient(
					s3Testing.Bucket(s3Testing.WithTaggingConfig(generateTaggingConfig())),
					fake.MockBucketClient{
						MockGetBucketTaggingRequest: func(input *s3.GetBucketTaggingInput) s3.GetBucketTaggingRequest {
							return s3.GetBucketTaggingRequest{
								Request: s3Testing.CreateRequest(nil, &s3.GetBucketTaggingOutput{TagSet: nil}),
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
				b: s3Testing.Bucket(s3Testing.WithTaggingConfig(nil)),
				cl: NewTaggingConfigurationClient(
					s3Testing.Bucket(s3Testing.WithTaggingConfig(nil)),
					fake.MockBucketClient{
						MockGetBucketTaggingRequest: func(input *s3.GetBucketTaggingInput) s3.GetBucketTaggingRequest {
							return s3.GetBucketTaggingRequest{
								Request: s3Testing.CreateRequest(nil, &s3.GetBucketTaggingOutput{TagSet: generateAWSTagging().TagSet}),
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
				b: s3Testing.Bucket(s3Testing.WithTaggingConfig(nil)),
				cl: NewTaggingConfigurationClient(
					s3Testing.Bucket(s3Testing.WithTaggingConfig(nil)),
					fake.MockBucketClient{
						MockGetBucketTaggingRequest: func(input *s3.GetBucketTaggingInput) s3.GetBucketTaggingRequest {
							return s3.GetBucketTaggingRequest{
								Request: s3Testing.CreateRequest(awserr.New(clients3.TaggingErrCode, "", nil), &s3.GetBucketTaggingOutput{}),
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
		"NoUpdateNotExistsNil": {
			args: args{
				b: s3Testing.Bucket(s3Testing.WithTaggingConfig(nil)),
				cl: NewTaggingConfigurationClient(
					s3Testing.Bucket(s3Testing.WithTaggingConfig(nil)),
					fake.MockBucketClient{
						MockGetBucketTaggingRequest: func(input *s3.GetBucketTaggingInput) s3.GetBucketTaggingRequest {
							return s3.GetBucketTaggingRequest{
								Request: s3Testing.CreateRequest(nil, &s3.GetBucketTaggingOutput{TagSet: nil}),
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
				b: s3Testing.Bucket(s3Testing.WithTaggingConfig(generateTaggingConfig())),
				cl: NewTaggingConfigurationClient(
					s3Testing.Bucket(s3Testing.WithTaggingConfig(generateTaggingConfig())),
					fake.MockBucketClient{
						MockGetBucketTaggingRequest: func(input *s3.GetBucketTaggingInput) s3.GetBucketTaggingRequest {
							return s3.GetBucketTaggingRequest{
								Request: s3Testing.CreateRequest(nil, &s3.GetBucketTaggingOutput{TagSet: generateAWSTagging().TagSet}),
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
				b: s3Testing.Bucket(s3Testing.WithTaggingConfig(generateTaggingConfig())),
				cl: NewTaggingConfigurationClient(
					s3Testing.Bucket(s3Testing.WithTaggingConfig(generateTaggingConfig())),
					fake.MockBucketClient{
						MockPutBucketTaggingRequest: func(input *s3.PutBucketTaggingInput) s3.PutBucketTaggingRequest {
							return s3.PutBucketTaggingRequest{
								Request: s3Testing.CreateRequest(errBoom, &s3.PutBucketTaggingOutput{}),
							}
						},
					},
				),
			},
			want: want{
				err: errors.Wrap(errBoom, taggingPutFailed),
			},
		},
		"InvalidConfig": {
			args: args{
				b: s3Testing.Bucket(s3Testing.WithTaggingConfig(generateTaggingConfig())),
				cl: NewTaggingConfigurationClient(
					s3Testing.Bucket(s3Testing.WithTaggingConfig(nil)),
					fake.MockBucketClient{
						MockPutBucketTaggingRequest: func(input *s3.PutBucketTaggingInput) s3.PutBucketTaggingRequest {
							return s3.PutBucketTaggingRequest{
								Request: s3Testing.CreateRequest(nil, &s3.PutBucketTaggingOutput{}),
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
				b: s3Testing.Bucket(s3Testing.WithTaggingConfig(generateTaggingConfig())),
				cl: NewTaggingConfigurationClient(
					s3Testing.Bucket(s3Testing.WithTaggingConfig(generateTaggingConfig())),
					fake.MockBucketClient{
						MockPutBucketTaggingRequest: func(input *s3.PutBucketTaggingInput) s3.PutBucketTaggingRequest {
							return s3.PutBucketTaggingRequest{
								Request: s3Testing.CreateRequest(nil, &s3.PutBucketTaggingOutput{}),
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
			_, err := tc.args.cl.CreateOrUpdate(context.Background(), tc.args.b)
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
				b: s3Testing.Bucket(s3Testing.WithTaggingConfig(generateTaggingConfig())),
				cl: NewTaggingConfigurationClient(
					s3Testing.Bucket(s3Testing.WithTaggingConfig(generateTaggingConfig())),
					fake.MockBucketClient{
						MockDeleteBucketTaggingRequest: func(input *s3.DeleteBucketTaggingInput) s3.DeleteBucketTaggingRequest {
							return s3.DeleteBucketTaggingRequest{
								Request: s3Testing.CreateRequest(errBoom, &s3.DeleteBucketTaggingOutput{}),
							}
						},
					},
				),
			},
			want: want{
				err: errors.Wrap(errBoom, taggingDeleteFailed),
			},
		},
		"SuccessfulDelete": {
			args: args{
				b: s3Testing.Bucket(s3Testing.WithTaggingConfig(generateTaggingConfig())),
				cl: NewTaggingConfigurationClient(
					s3Testing.Bucket(s3Testing.WithTaggingConfig(generateTaggingConfig())),
					fake.MockBucketClient{
						MockDeleteBucketTaggingRequest: func(input *s3.DeleteBucketTaggingInput) s3.DeleteBucketTaggingRequest {
							return s3.DeleteBucketTaggingRequest{
								Request: s3Testing.CreateRequest(nil, &s3.DeleteBucketTaggingOutput{}),
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
			err := tc.args.cl.Delete(context.Background(), tc.args.b)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}
