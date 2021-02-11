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
	"net/http"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/awserr"
	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/test"

	"github.com/crossplane/provider-aws/apis/s3/v1alpha3"
	awsclient "github.com/crossplane/provider-aws/pkg/clients"
	"github.com/crossplane/provider-aws/pkg/clients/s3"
	"github.com/crossplane/provider-aws/pkg/clients/s3/fake"
)

var (
	// an arbitrary managed resource
	unexpectedItem resource.Managed
	bucketName     = "test.s3.crossplane.com"
	policy         = `{"Statement":[{"Action":"s3:ListBucket","Effect":"Allow","Principal":"*","Resource":"arn:aws:s3:::test.s3.crossplane.com"}],"Version":"2012-10-17"}`

	params = v1alpha3.BucketPolicyParameters{
		Policy: &v1alpha3.BucketPolicyBody{
			Version: "2012-10-17",
			Statements: []v1alpha3.BucketPolicyStatement{
				{
					Effect: "Allow",
					Principal: &v1alpha3.BucketPrincipal{
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
				Policy: &v1alpha3.BucketPolicyBody{
					Statements: make([]v1alpha3.BucketPolicyStatement, 0),
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
					MockGetBucketPolicyRequest: func(input *awss3.GetBucketPolicyInput) awss3.GetBucketPolicyRequest {
						return awss3.GetBucketPolicyRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awss3.GetBucketPolicyOutput{
								Policy: &policy,
							}},
						}
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
					MockGetBucketPolicyRequest: func(input *awss3.GetBucketPolicyInput) awss3.GetBucketPolicyRequest {
						return awss3.GetBucketPolicyRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Error: errBoom},
						}
					},
				},
				cr: bucketPolicy(withPolicy(&params)),
			},
			want: want{
				cr:  bucketPolicy(withPolicy(&params)),
				err: awsclient.Wrap(errBoom, errGet),
			},
		},
		"ResourceDoesNotExist": {
			args: args{
				s3: &fake.MockBucketPolicyClient{
					MockGetBucketPolicyRequest: func(input *awss3.GetBucketPolicyInput) awss3.GetBucketPolicyRequest {
						return awss3.GetBucketPolicyRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Error: awserr.New("NoSuchBucketPolicy", "", nil)},
						}
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
					MockPutBucketPolicyRequest: func(input *awss3.PutBucketPolicyInput) awss3.PutBucketPolicyRequest {
						return awss3.PutBucketPolicyRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awss3.PutBucketPolicyOutput{}},
						}
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
					MockPutBucketPolicyRequest: func(input *awss3.PutBucketPolicyInput) awss3.PutBucketPolicyRequest {
						return awss3.PutBucketPolicyRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Error: errBoom},
						}
					},
				},
				cr: bucketPolicy(withPolicy(&params)),
			},
			want: want{
				cr: bucketPolicy(
					withPolicy(&params),
					withConditions(xpv1.Creating())),
				err: awsclient.Wrap(errBoom, errAttach),
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
					MockPutBucketPolicyRequest: func(input *awss3.PutBucketPolicyInput) awss3.PutBucketPolicyRequest {
						return awss3.PutBucketPolicyRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awss3.PutBucketPolicyOutput{}},
						}
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
					MockDeleteBucketPolicyRequest: func(input *awss3.DeleteBucketPolicyInput) awss3.DeleteBucketPolicyRequest {
						return awss3.DeleteBucketPolicyRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awss3.DeleteBucketPolicyOutput{}},
						}
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
					MockDeleteBucketPolicyRequest: func(input *awss3.DeleteBucketPolicyInput) awss3.DeleteBucketPolicyRequest {
						return awss3.DeleteBucketPolicyRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Error: errBoom},
						}
					},
				},
				cr: bucketPolicy(withPolicy(&params)),
			},
			want: want{
				cr: bucketPolicy(withPolicy(&params),
					withConditions(xpv1.Deleting())),
				err: awsclient.Wrap(errBoom, errDelete),
			},
		},
		"ResourceDoesNotExist": {
			args: args{
				s3: &fake.MockBucketPolicyClient{
					MockDeleteBucketPolicyRequest: func(input *awss3.DeleteBucketPolicyInput) awss3.DeleteBucketPolicyRequest {
						return awss3.DeleteBucketPolicyRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Error: awserr.New("NoSuchBucketPolicy", "", nil)},
						}
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
			err := e.Delete(context.Background(), tc.args.cr)

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
			if diff := cmp.Diff(tc.want.str, aws.StringValue(str)); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}
