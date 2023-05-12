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
	"encoding/json"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"

	"github.com/crossplane-contrib/provider-aws/apis/s3/common"
	"github.com/crossplane-contrib/provider-aws/apis/s3/v1beta1"
	awsclient "github.com/crossplane-contrib/provider-aws/pkg/clients"
	s3client "github.com/crossplane-contrib/provider-aws/pkg/clients/s3"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/s3/fake"
	s3testing "github.com/crossplane-contrib/provider-aws/pkg/controller/s3/testing"
)

func makeRawPolicy(p *common.BucketPolicyBody) string {
	serialized, err := s3client.Serialize(p)
	if err != nil {
		panic(err.Error())
	}
	raw, err := json.Marshal(serialized)
	if err != nil {
		panic(err.Error())
	}
	return string(raw)
}

func TestPolicyObserve(t *testing.T) {
	var testPolicy = &common.BucketPolicyBody{
		Version: "2012-10-17",
		Statements: []common.BucketPolicyStatement{
			{
				Effect: "Allow",
				Principal: &common.BucketPrincipal{
					AllowAnon: true,
				},
				Action:   []string{"s3:ListBucket"},
				Resource: []string{"arn:aws:s3:::test.s3.crossplane.com"},
			},
		},
	}

	var testPolicyOther = &common.BucketPolicyBody{
		Version: "2012-10-17",
		Statements: []common.BucketPolicyStatement{
			{
				Effect: "Allow",
				Principal: &common.BucketPrincipal{
					AllowAnon: true,
				},
				Action:   []string{"s3:GetObject"},
				Resource: []string{"arn:aws:s3:::test.s3.crossplane.com/*"},
			},
		},
	}

	testPolicyRawShuffled := "{\"Statement\":[{\"Effect\":\"Allow\",\"Action\":\"s3:ListBucket\",\"Principal\":\"*\",\"Resource\":\"arn:aws:s3:::test.s3.crossplane.com\"}],\"Version\":\"2012-10-17\"}"
	testPolicyRaw := makeRawPolicy(testPolicy)
	testPolicyOtherRaw := makeRawPolicy(testPolicyOther)

	type args struct {
		cl *PolicyClient
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
				b: s3testing.Bucket(s3testing.WithPolicy(testPolicy)),
				cl: NewPolicyClient(fake.MockBucketClient{
					MockBucketPolicyClient: fake.MockBucketPolicyClient{
						MockGetBucketPolicy: func(ctx context.Context, input *s3.GetBucketPolicyInput, opts []func(*s3.Options)) (*s3.GetBucketPolicyOutput, error) {
							return nil, errBoom
						},
					},
				}),
			},
			want: want{
				status: NeedsUpdate,
				err:    awsclient.Wrap(errBoom, policyGetFailed),
			},
		},
		"UpdateNeeded": {
			args: args{
				b: s3testing.Bucket(s3testing.WithPolicy(testPolicy)),
				cl: NewPolicyClient(fake.MockBucketClient{
					MockBucketPolicyClient: fake.MockBucketPolicyClient{
						MockGetBucketPolicy: func(ctx context.Context, input *s3.GetBucketPolicyInput, opts []func(*s3.Options)) (*s3.GetBucketPolicyOutput, error) {
							return &s3.GetBucketPolicyOutput{
								Policy: &testPolicyOtherRaw,
							}, nil
						},
					},
				}),
			},
			want: want{
				status: NeedsUpdate,
				err:    nil,
			},
		},
		"UpdateNeededNotExist": {
			args: args{
				b: s3testing.Bucket(s3testing.WithPolicy(testPolicy)),
				cl: NewPolicyClient(fake.MockBucketClient{
					MockBucketPolicyClient: fake.MockBucketPolicyClient{
						MockGetBucketPolicy: func(ctx context.Context, input *s3.GetBucketPolicyInput, opts []func(*s3.Options)) (*s3.GetBucketPolicyOutput, error) {
							return nil, &smithy.GenericAPIError{Code: "NoSuchBucketPolicy"}
						},
					},
				}),
			},
			want: want{
				status: NeedsUpdate,
				err:    nil,
			},
		},
		"NoUpdateNotExistsAndNotSet": {
			args: args{
				b: s3testing.Bucket(),
				cl: NewPolicyClient(fake.MockBucketClient{
					MockBucketPolicyClient: fake.MockBucketPolicyClient{
						MockGetBucketPolicy: func(ctx context.Context, input *s3.GetBucketPolicyInput, opts []func(*s3.Options)) (*s3.GetBucketPolicyOutput, error) {
							return nil, &smithy.GenericAPIError{Code: "NoSuchBucketPolicy"}
						},
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
				b: s3testing.Bucket(s3testing.WithPolicy(testPolicy)),
				cl: NewPolicyClient(fake.MockBucketClient{
					MockBucketPolicyClient: fake.MockBucketPolicyClient{
						MockGetBucketPolicy: func(ctx context.Context, input *s3.GetBucketPolicyInput, opts []func(*s3.Options)) (*s3.GetBucketPolicyOutput, error) {
							return &s3.GetBucketPolicyOutput{
								Policy: &testPolicyRaw,
							}, nil
						},
					},
				}),
			},
			want: want{
				status: Updated,
				err:    nil,
			},
		},
		"DeletionNeededIfPolicyIfNull": {
			args: args{
				b: s3testing.Bucket(
					s3testing.WithPolicyUpdatePolicy(&v1beta1.BucketPolicyUpdatePolicy{
						DeletionPolicy: v1beta1.BucketPolicyDeletionPolicyIfNull,
					}),
				),
				cl: NewPolicyClient(fake.MockBucketClient{
					MockBucketPolicyClient: fake.MockBucketPolicyClient{
						MockGetBucketPolicy: func(ctx context.Context, input *s3.GetBucketPolicyInput, opts []func(*s3.Options)) (*s3.GetBucketPolicyOutput, error) {
							return &s3.GetBucketPolicyOutput{
								Policy: &testPolicyRaw,
							}, nil
						},
					},
				}),
			},
			want: want{
				status: NeedsDeletion,
				err:    nil,
			},
		},
		"NoUpdateExistsWithshuffledPolicy": {
			args: args{
				b: s3testing.Bucket(s3testing.WithPolicy(testPolicy)),
				cl: NewPolicyClient(fake.MockBucketClient{
					MockBucketPolicyClient: fake.MockBucketPolicyClient{
						MockGetBucketPolicy: func(ctx context.Context, input *s3.GetBucketPolicyInput, opts []func(*s3.Options)) (*s3.GetBucketPolicyOutput, error) {
							return &s3.GetBucketPolicyOutput{
								Policy: &testPolicyRawShuffled,
							}, nil
						},
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

func TestPolicyCreateOrUpdate(t *testing.T) {
	var testPolicy = &common.BucketPolicyBody{
		Version: "2012-10-17",
		Statements: []common.BucketPolicyStatement{
			{
				Effect: "Allow",
				Principal: &common.BucketPrincipal{
					AllowAnon: true,
				},
				Action:   []string{"s3:ListBucket"},
				Resource: []string{"arn:aws:s3:::test.s3.crossplane.com"},
			},
		},
	}

	type args struct {
		cl *PolicyClient
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
				b: s3testing.Bucket(s3testing.WithPolicy(testPolicy)),
				cl: NewPolicyClient(fake.MockBucketClient{
					MockBucketPolicyClient: fake.MockBucketPolicyClient{
						MockPutBucketPolicy: func(ctx context.Context, input *s3.PutBucketPolicyInput, opts []func(*s3.Options)) (*s3.PutBucketPolicyOutput, error) {
							return nil, errBoom
						},
					},
				}),
			},
			want: want{
				err: awsclient.Wrap(errBoom, policyPutFailed),
			},
		},
		"NoPutIfNoPolicy": {
			args: args{
				b: s3testing.Bucket(),
				cl: NewPolicyClient(fake.MockBucketClient{
					MockBucketPolicyClient: fake.MockBucketPolicyClient{
						MockPutBucketPolicy: func(ctx context.Context, input *s3.PutBucketPolicyInput, opts []func(*s3.Options)) (*s3.PutBucketPolicyOutput, error) {
							return nil, errBoom
						},
					},
				}),
			},
			want: want{
				err: nil,
			},
		},
		"SuccessfulPut": {
			args: args{
				b: s3testing.Bucket(s3testing.WithPolicy(testPolicy)),
				cl: NewPolicyClient(fake.MockBucketClient{
					MockBucketPolicyClient: fake.MockBucketPolicyClient{
						MockPutBucketPolicy: func(ctx context.Context, input *s3.PutBucketPolicyInput, opts []func(*s3.Options)) (*s3.PutBucketPolicyOutput, error) {
							return nil, nil
						},
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

func TestPolicyDelete(t *testing.T) {
	var testPolicy = &common.BucketPolicyBody{
		Version: "2012-10-17",
		Statements: []common.BucketPolicyStatement{
			{
				Effect: "Allow",
				Principal: &common.BucketPrincipal{
					AllowAnon: true,
				},
				Action:   []string{"s3:ListBucket"},
				Resource: []string{"arn:aws:s3:::test.s3.crossplane.com"},
			},
		},
	}

	type args struct {
		cl *PolicyClient
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
				b: s3testing.Bucket(s3testing.WithPolicy(testPolicy)),
				cl: NewPolicyClient(fake.MockBucketClient{
					MockBucketPolicyClient: fake.MockBucketPolicyClient{
						MockDeleteBucketPolicy: func(ctx context.Context, input *s3.DeleteBucketPolicyInput, opts []func(*s3.Options)) (*s3.DeleteBucketPolicyOutput, error) {
							return nil, errBoom
						},
					},
				}),
			},
			want: want{
				err: awsclient.Wrap(errBoom, policyDeleteFailed),
			},
		},
		"SuccessfullDelete": {
			args: args{
				b: s3testing.Bucket(s3testing.WithPolicy(testPolicy)),
				cl: NewPolicyClient(fake.MockBucketClient{
					MockBucketPolicyClient: fake.MockBucketPolicyClient{
						MockDeleteBucketPolicy: func(ctx context.Context, input *s3.DeleteBucketPolicyInput, opts []func(*s3.Options)) (*s3.DeleteBucketPolicyOutput, error) {
							return nil, nil
						},
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
