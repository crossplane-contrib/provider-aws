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

package iamuserpolicyattachment

import (
	"context"
	"net/http"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/awserr"
	awsiam "github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/test"

	"github.com/crossplane/provider-aws/apis/identity/v1alpha1"
	"github.com/crossplane/provider-aws/pkg/clients/iam"
	"github.com/crossplane/provider-aws/pkg/clients/iam/fake"
)

var (
	unexpectedItem resource.Managed
	policyArn      = "some arn"
	userName       = "some user"

	errBoom = errors.New("boom")
)

type args struct {
	iam iam.UserPolicyAttachmentClient
	cr  resource.Managed
}

type userPolicyModifier func(*v1alpha1.IAMUserPolicyAttachment)

func withConditions(c ...runtimev1alpha1.Condition) userPolicyModifier {
	return func(r *v1alpha1.IAMUserPolicyAttachment) { r.Status.ConditionedStatus.Conditions = c }
}

func withUserName(s string) userPolicyModifier {
	return func(r *v1alpha1.IAMUserPolicyAttachment) { r.Spec.ForProvider.UserName = s }
}

func withSpecPolicyArn(s string) userPolicyModifier {
	return func(r *v1alpha1.IAMUserPolicyAttachment) { r.Spec.ForProvider.PolicyARN = s }
}

func withStatusPolicyArn(s string) userPolicyModifier {
	return func(r *v1alpha1.IAMUserPolicyAttachment) { r.Status.AtProvider.AttachedPolicyARN = s }
}

func userPolicy(m ...userPolicyModifier) *v1alpha1.IAMUserPolicyAttachment {
	cr := &v1alpha1.IAMUserPolicyAttachment{}
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
		"VaildInput": {
			args: args{
				iam: &fake.MockUserPolicyAttachmentClient{
					MockListAttachedUserPolicies: func(input *awsiam.ListAttachedUserPoliciesInput) awsiam.ListAttachedUserPoliciesRequest {
						return awsiam.ListAttachedUserPoliciesRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsiam.ListAttachedUserPoliciesOutput{
								AttachedPolicies: []awsiam.AttachedPolicy{
									{
										PolicyArn: &policyArn,
									},
								},
							}},
						}
					},
				},
				cr: userPolicy(withUserName(userName),
					withSpecPolicyArn(policyArn)),
			},
			want: want{
				cr: userPolicy(withUserName(userName),
					withSpecPolicyArn(policyArn),
					withConditions(runtimev1alpha1.Available()),
					withStatusPolicyArn(policyArn)),
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
		"NoAttachedPolicy": {
			args: args{
				iam: &fake.MockUserPolicyAttachmentClient{
					MockListAttachedUserPolicies: func(input *awsiam.ListAttachedUserPoliciesInput) awsiam.ListAttachedUserPoliciesRequest {
						return awsiam.ListAttachedUserPoliciesRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsiam.ListAttachedUserPoliciesOutput{}},
						}
					},
				},
				cr: userPolicy(withSpecPolicyArn(policyArn)),
			},
			want: want{
				cr: userPolicy(withSpecPolicyArn(policyArn)),
			},
		},
		"ClientError": {
			args: args{
				iam: &fake.MockUserPolicyAttachmentClient{
					MockListAttachedUserPolicies: func(input *awsiam.ListAttachedUserPoliciesInput) awsiam.ListAttachedUserPoliciesRequest {
						return awsiam.ListAttachedUserPoliciesRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: errBoom},
						}
					},
				},
				cr: userPolicy(withUserName(userName)),
			},
			want: want{
				cr:  userPolicy(withUserName(userName)),
				err: errors.Wrap(errBoom, errGet),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.iam}
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
				iam: &fake.MockUserPolicyAttachmentClient{
					MockAttachUserPolicy: func(input *awsiam.AttachUserPolicyInput) awsiam.AttachUserPolicyRequest {
						return awsiam.AttachUserPolicyRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsiam.AttachUserPolicyOutput{}},
						}
					},
				},
				cr: userPolicy(withUserName(userName),
					withSpecPolicyArn(policyArn)),
			},
			want: want{
				cr: userPolicy(
					withUserName(userName),
					withSpecPolicyArn(policyArn),
					withConditions(runtimev1alpha1.Creating())),
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
				iam: &fake.MockUserPolicyAttachmentClient{
					MockAttachUserPolicy: func(input *awsiam.AttachUserPolicyInput) awsiam.AttachUserPolicyRequest {
						return awsiam.AttachUserPolicyRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: errBoom},
						}
					},
				},
				cr: userPolicy(withUserName(userName),
					withSpecPolicyArn(policyArn)),
			},
			want: want{
				cr: userPolicy(withUserName(userName),
					withSpecPolicyArn(policyArn),
					withConditions(runtimev1alpha1.Creating())),
				err: errors.Wrap(errBoom, errAttach),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.iam}
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
				iam: &fake.MockUserPolicyAttachmentClient{
					MockDetachUserPolicy: func(input *awsiam.DetachUserPolicyInput) awsiam.DetachUserPolicyRequest {
						return awsiam.DetachUserPolicyRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsiam.DetachUserPolicyOutput{}},
						}
					},
				},
				cr: userPolicy(withUserName(userName),
					withSpecPolicyArn(policyArn)),
			},
			want: want{
				cr: userPolicy(
					withUserName(userName),
					withSpecPolicyArn(policyArn),
					withConditions(runtimev1alpha1.Deleting())),
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
				iam: &fake.MockUserPolicyAttachmentClient{
					MockDetachUserPolicy: func(input *awsiam.DetachUserPolicyInput) awsiam.DetachUserPolicyRequest {
						return awsiam.DetachUserPolicyRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: errBoom},
						}
					},
				},
				cr: userPolicy(withUserName(userName),
					withSpecPolicyArn(policyArn)),
			},
			want: want{
				cr: userPolicy(withUserName(userName),
					withSpecPolicyArn(policyArn),
					withConditions(runtimev1alpha1.Deleting())),
				err: errors.Wrap(errBoom, errDetach),
			},
		},
		"ResourceDoesNotExist": {
			args: args{
				iam: &fake.MockUserPolicyAttachmentClient{
					MockDetachUserPolicy: func(input *awsiam.DetachUserPolicyInput) awsiam.DetachUserPolicyRequest {
						return awsiam.DetachUserPolicyRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: awserr.New(awsiam.ErrCodeNoSuchEntityException, "", nil)},
						}
					},
				},
				cr: userPolicy(),
			},
			want: want{
				cr: userPolicy(withConditions(runtimev1alpha1.Deleting())),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.iam}
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
