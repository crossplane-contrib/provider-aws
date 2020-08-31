/*
Copyright 2019 The Crossplane Authors.

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

package s3bucketpolicy

import (
	"context"
	"net/http"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/awserr"
	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"

	corev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/test"

	"github.com/crossplane/provider-aws/apis/storage/v1alpha1"
	iamfake "github.com/crossplane/provider-aws/pkg/clients/iam/fake"
	"github.com/crossplane/provider-aws/pkg/clients/s3"
	"github.com/crossplane/provider-aws/pkg/clients/s3/fake"
)

var (
	// an arbitrary managed resource
	unexpectedItem resource.Managed
	bucketName     = "test.s3.crossplane.com"
	userName       = "12345667890"
	policy         = `{"Statement":[{"Action":"s3:ListBucket","Effect":"Allow","Principal":"*","Resource":"arn:aws:s3:::test.s3.crossplane.com"}],"Version":"2012-10-17"}`

	params = v1alpha1.S3BucketPolicyParameters{
		PolicyVersion: "2012-10-17",
		PolicyStatement: []v1alpha1.S3BucketPolicyStatement{
			{
				Effect: "Allow",
				Principal: &v1alpha1.S3BucketPrincipal{
					AllowAnon: true,
				},
				PolicyAction:   []string{"s3:ListBucket"},
				ResourcePath:   []string{"test.s3.crossplane.com"},
				ApplyToIAMUser: false,
			},
		},
	}
	errBoom = errors.New("boom")
)

type args struct {
	s3 s3.BucketPolicyClient
	cr resource.Managed
}

type bucketPolicyModifier func(policy *v1alpha1.S3BucketPolicy)

func withConditions(c ...corev1alpha1.Condition) bucketPolicyModifier {
	return func(r *v1alpha1.S3BucketPolicy) { r.Status.ConditionedStatus.Conditions = c }
}

func withPolicy(s *v1alpha1.S3BucketPolicyParameters) bucketPolicyModifier {
	return func(r *v1alpha1.S3BucketPolicy) { r.Spec.PolicyBody = *s }
}

func bucketPolicy(m ...bucketPolicyModifier) *v1alpha1.S3BucketPolicy {
	cr := &v1alpha1.S3BucketPolicy{
		Spec: v1alpha1.S3BucketPolicySpec{
			PolicyBody: v1alpha1.S3BucketPolicyParameters{
				UserName:        &userName,
				BucketName:      &bucketName,
				PolicyVersion:   "",
				PolicyID:        "",
				PolicyStatement: make([]v1alpha1.S3BucketPolicyStatement, 0),
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
					withConditions(corev1alpha1.Available())),
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
				err: errors.Wrap(errBoom, errGet),
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

	iamc := new(iamfake.Client)
	iamc.On("GetAccountID").Return("1234567890", nil)

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.s3, iamclient: iamc}
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
					withConditions(corev1alpha1.Creating())),
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
					withConditions(corev1alpha1.Creating())),
				err: errors.Wrap(errBoom, errAttach),
			},
		},
	}

	iamc := new(iamfake.Client)
	iamc.On("GetAccountID").Return("1234567890", nil)

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.s3, iamclient: iamc}
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

	iamc := new(iamfake.Client)
	iamc.On("GetAccountID").Return("1234567890", nil)

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.s3, iamclient: iamc}
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
					withConditions(corev1alpha1.Deleting())),
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
					withConditions(corev1alpha1.Deleting())),
				err: errors.Wrap(errBoom, errDelete),
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
				cr: bucketPolicy(withPolicy(&params), withConditions(corev1alpha1.Deleting())),
			},
		},
	}

	iamc := new(iamfake.Client)
	iamc.On("GetAccountID").Return("1234567890", nil)

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.s3, iamclient: iamc}
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
