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

package bucketpolicy

import (
	"context"
	"testing"

	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"

	"github.com/crossplane-contrib/provider-aws/apis/s3/common"
	"github.com/crossplane-contrib/provider-aws/apis/s3/v1alpha3"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/s3"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/s3/fake"
	errorutils "github.com/crossplane-contrib/provider-aws/pkg/utils/errors"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
)

var (
	// an arbitrary managed resource
	unexpectedItem resource.Managed
	bucketName     = "test.s3.crossplane.com"
	policy         = `{"Statement":[{"Action":"s3:ListBucket","Effect":"Allow","Principal":"*","Resource":"arn:aws:s3:::test.s3.crossplane.com"}],"Version":"2012-10-17"}`

	params = v1alpha3.BucketPolicyParameters{
		Policy: &common.BucketPolicyBody{
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
		},
	}
	errBoom = errors.New("boom")
)

type args struct {
	s3 s3.BucketPolicyClient
	cr resource.Managed
}

type bucketPolicyModifier func(policy *v1alpha3.BucketPolicy)

func withConditions(c ...xpv1.Condition) bucketPolicyModifier {
	return func(r *v1alpha3.BucketPolicy) { r.Status.ConditionedStatus.Conditions = c }
}

func withPolicy(s *v1alpha3.BucketPolicyParameters) bucketPolicyModifier {
	return func(r *v1alpha3.BucketPolicy) { r.Spec.Parameters = *s }
}

func bucketPolicy(m ...bucketPolicyModifier) *v1alpha3.BucketPolicy {
	cr := &v1alpha3.BucketPolicy{
		Spec: v1alpha3.BucketPolicySpec{
			Parameters: v1alpha3.BucketPolicyParameters{
				BucketName: &bucketName,
				Policy: &common.BucketPolicyBody{
					Statements: make([]common.BucketPolicyStatement, 0),
				},
			},
		},
	}
	for _, f := range m {
		f(cr)
	}
	return cr
}

func TestObserve(t *testing.T) {

	type want struct {
		cr     resource.Managed
		result managed.ExternalObservation
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"ValidInput": {
			args: args{
				s3: &fake.MockBucketPolicyClient{
					MockGetBucketPolicy: func(ctx context.Context, input *awss3.GetBucketPolicyInput, opts []func(*awss3.Options)) (*awss3.GetBucketPolicyOutput, error) {
						return &awss3.GetBucketPolicyOutput{
							Policy: &policy,
						}, nil
					},
				},
				cr: bucketPolicy(withPolicy(&params)),
			},
			want: want{
				cr: bucketPolicy(withPolicy(&params),
					withConditions(xpv1.Available())),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
				},
			},
		},
		"InValidInput": {
			args: args{
				cr: unexpectedItem,
			},
			want: want{
				cr:  unexpectedItem,
				err: errors.New(errUnexpectedObject),
			},
		},
		"ClientError": {
			args: args{
				s3: &fake.MockBucketPolicyClient{
					MockGetBucketPolicy: func(ctx context.Context, input *awss3.GetBucketPolicyInput, opts []func(*awss3.Options)) (*awss3.GetBucketPolicyOutput, error) {
						return nil, errBoom
					},
				},
				cr: bucketPolicy(withPolicy(&params)),
			},
			want: want{
				cr:  bucketPolicy(withPolicy(&params)),
				err: errorutils.Wrap(errBoom, errGet),
			},
		},
		"ResourceDoesNotExist": {
			args: args{
				s3: &fake.MockBucketPolicyClient{
					MockGetBucketPolicy: func(ctx context.Context, input *awss3.GetBucketPolicyInput, opts []func(*awss3.Options)) (*awss3.GetBucketPolicyOutput, error) {
						return nil, &smithy.GenericAPIError{Code: "NoSuchBucketPolicy"}
					},
				},
				cr: bucketPolicy(),
			},
			want: want{
				cr: bucketPolicy(),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.s3}
			o, err := e.Observe(context.Background(), tc.args.cr)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.cr, tc.args.cr, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.result, o); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestCreate(t *testing.T) {

	type want struct {
		cr     resource.Managed
		result managed.ExternalCreation
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"VaildInput": {
			args: args{
				s3: &fake.MockBucketPolicyClient{
					MockPutBucketPolicy: func(ctx context.Context, input *awss3.PutBucketPolicyInput, opts []func(*awss3.Options)) (*awss3.PutBucketPolicyOutput, error) {
						return &awss3.PutBucketPolicyOutput{}, nil
					},
				},
				cr: bucketPolicy(withPolicy(&params)),
			},
			want: want{
				cr: bucketPolicy(
					withPolicy(&params),
					withConditions(xpv1.Creating())),
			},
		},
		"InValidInput": {
			args: args{
				cr: unexpectedItem,
			},
			want: want{
				cr:  unexpectedItem,
				err: errors.New(errUnexpectedObject),
			},
		},
		"ClientError": {
			args: args{
				s3: &fake.MockBucketPolicyClient{
					MockPutBucketPolicy: func(ctx context.Context, input *awss3.PutBucketPolicyInput, opts []func(*awss3.Options)) (*awss3.PutBucketPolicyOutput, error) {
						return nil, errBoom
					},
				},
				cr: bucketPolicy(withPolicy(&params)),
			},
			want: want{
				cr: bucketPolicy(
					withPolicy(&params),
					withConditions(xpv1.Creating())),
				err: errorutils.Wrap(errBoom, errAttach),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.s3}
			o, err := e.Create(context.Background(), tc.args.cr)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.cr, tc.args.cr, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.result, o); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestUpdate(t *testing.T) {

	type want struct {
		cr     resource.Managed
		result managed.ExternalUpdate
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"VaildInput": {
			args: args{
				s3: &fake.MockBucketPolicyClient{
					MockPutBucketPolicy: func(ctx context.Context, input *awss3.PutBucketPolicyInput, opts []func(*awss3.Options)) (*awss3.PutBucketPolicyOutput, error) {
						return &awss3.PutBucketPolicyOutput{}, nil
					},
				},
				cr: bucketPolicy(withPolicy(&params)),
			},
			want: want{
				cr: bucketPolicy(withPolicy(&params)),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.s3}
			o, err := e.Update(context.Background(), tc.args.cr)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.cr, tc.args.cr, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.result, o); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestDelete(t *testing.T) {

	type want struct {
		cr  resource.Managed
		err error
	}

	cases := map[string]struct {
		args
		want
	}{
		"VaildInput": {
			args: args{
				s3: &fake.MockBucketPolicyClient{
					MockDeleteBucketPolicy: func(ctx context.Context, input *awss3.DeleteBucketPolicyInput, opts []func(*awss3.Options)) (*awss3.DeleteBucketPolicyOutput, error) {
						return &awss3.DeleteBucketPolicyOutput{}, nil
					},
				},
				cr: bucketPolicy(withPolicy(&params)),
			},
			want: want{
				cr: bucketPolicy(withPolicy(&params),
					withConditions(xpv1.Deleting())),
			},
		},
		"InValidInput": {
			args: args{
				cr: unexpectedItem,
			},
			want: want{
				cr:  unexpectedItem,
				err: errors.New(errUnexpectedObject),
			},
		},
		"ClientError": {
			args: args{
				s3: &fake.MockBucketPolicyClient{
					MockDeleteBucketPolicy: func(ctx context.Context, input *awss3.DeleteBucketPolicyInput, opts []func(*awss3.Options)) (*awss3.DeleteBucketPolicyOutput, error) {
						return nil, errBoom
					},
				},
				cr: bucketPolicy(withPolicy(&params)),
			},
			want: want{
				cr: bucketPolicy(withPolicy(&params),
					withConditions(xpv1.Deleting())),
				err: errorutils.Wrap(errBoom, errDelete),
			},
		},
		"ResourceDoesNotExist": {
			args: args{
				s3: &fake.MockBucketPolicyClient{
					MockDeleteBucketPolicy: func(ctx context.Context, input *awss3.DeleteBucketPolicyInput, opts []func(*awss3.Options)) (*awss3.DeleteBucketPolicyOutput, error) {
						return nil, &smithy.GenericAPIError{Code: "NoSuchBucketPolicy"}
					},
				},
				cr: bucketPolicy(withPolicy(&params)),
			},
			want: want{
				cr: bucketPolicy(withPolicy(&params), withConditions(xpv1.Deleting())),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.s3}
			_, err := e.Delete(context.Background(), tc.args.cr)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.cr, tc.args.cr, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestFormat(t *testing.T) {
	type formatarg struct {
		cr *v1alpha3.BucketPolicy
	}
	type want struct {
		str string
		err error
	}

	cases := map[string]struct {
		args formatarg
		want
	}{
		"VaildInput": {
			args: formatarg{
				cr: bucketPolicy(withPolicy(&params)),
			},
			want: want{
				str: policy,
			},
		},
		"InValidInput": {
			args: formatarg{
				cr: nil,
			},
			want: want{
				err: errors.New(errNotSpecified),
			},
		},
		"StringPolicy": {
			args: formatarg{
				cr: bucketPolicy(withPolicy(&v1alpha3.BucketPolicyParameters{
					RawPolicy: &policy,
				})),
			},
			want: want{
				str: policy,
			},
		},
		"NoPolicy": {
			args: formatarg{
				cr: bucketPolicy(withPolicy(&v1alpha3.BucketPolicyParameters{})),
			},
			want: want{
				err: errors.New(errNotSpecified),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{}
			str, err := e.formatBucketPolicy(tc.args.cr)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.str, pointer.StringValue(str)); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}
