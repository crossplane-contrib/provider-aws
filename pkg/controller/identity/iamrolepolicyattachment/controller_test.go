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

package iamrolepolicyattachment

import (
	"context"
	"net/http"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/awserr"
	awsiam "github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/test"

	v1beta1 "github.com/crossplane/provider-aws/apis/identity/v1beta1"
	"github.com/crossplane/provider-aws/pkg/clients/iam"
	"github.com/crossplane/provider-aws/pkg/clients/iam/fake"
)

var (
	// an arbitrary managed resource
	unexpectedItem resource.Managed
	roleName       = "some arbitrary name"
	specPolicyArn  = "some arbitrary arn"

	errBoom = errors.New("boom")
)

type args struct {
	iam iam.RolePolicyAttachmentClient
	cr  resource.Managed
}

type rolePolicyModifier func(*v1beta1.IAMRolePolicyAttachment)

func withConditions(c ...xpv1.Condition) rolePolicyModifier {
	return func(r *v1beta1.IAMRolePolicyAttachment) { r.Status.ConditionedStatus.Conditions = c }
}

func withRoleName(s *string) rolePolicyModifier {
	return func(r *v1beta1.IAMRolePolicyAttachment) { r.Spec.ForProvider.RoleName = *s }
}

func withSpecPolicyArn(s *string) rolePolicyModifier {
	return func(r *v1beta1.IAMRolePolicyAttachment) { r.Spec.ForProvider.PolicyARN = *s }
}

func withStatusPolicyArn(s *string) rolePolicyModifier {
	return func(r *v1beta1.IAMRolePolicyAttachment) { r.Status.AtProvider.AttachedPolicyARN = *s }
}

func rolePolicy(m ...rolePolicyModifier) *v1beta1.IAMRolePolicyAttachment {
	cr := &v1beta1.IAMRolePolicyAttachment{}
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
				iam: &fake.MockRolePolicyAttachmentClient{
					MockListAttachedRolePoliciesRequest: func(input *awsiam.ListAttachedRolePoliciesInput) awsiam.ListAttachedRolePoliciesRequest {
						return awsiam.ListAttachedRolePoliciesRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsiam.ListAttachedRolePoliciesOutput{
								AttachedPolicies: []awsiam.AttachedPolicy{
									{
										PolicyArn: &specPolicyArn,
									},
								},
							}},
						}
					},
				},
				cr: rolePolicy(withSpecPolicyArn(&specPolicyArn)),
			},
			want: want{
				cr: rolePolicy(withSpecPolicyArn(&specPolicyArn),
					withConditions(xpv1.Available()),
					withStatusPolicyArn(&specPolicyArn)),
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
				iam: &fake.MockRolePolicyAttachmentClient{
					MockListAttachedRolePoliciesRequest: func(input *awsiam.ListAttachedRolePoliciesInput) awsiam.ListAttachedRolePoliciesRequest {
						return awsiam.ListAttachedRolePoliciesRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: errBoom},
						}
					},
				},
				cr: rolePolicy(withRoleName(&roleName)),
			},
			want: want{
				cr:  rolePolicy(withRoleName(&roleName)),
				err: errors.Wrap(errBoom, errGet),
			},
		},
		"ResourceDoesNotExist": {
			args: args{
				iam: &fake.MockRolePolicyAttachmentClient{
					MockListAttachedRolePoliciesRequest: func(input *awsiam.ListAttachedRolePoliciesInput) awsiam.ListAttachedRolePoliciesRequest {
						return awsiam.ListAttachedRolePoliciesRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: awserr.New(awsiam.ErrCodeNoSuchEntityException, "", nil)},
						}
					},
				},
				cr: rolePolicy(),
			},
			want: want{
				cr: rolePolicy(),
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
				iam: &fake.MockRolePolicyAttachmentClient{
					MockAttachRolePolicyRequest: func(input *awsiam.AttachRolePolicyInput) awsiam.AttachRolePolicyRequest {
						return awsiam.AttachRolePolicyRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsiam.AttachRolePolicyOutput{}},
						}
					},
				},
				cr: rolePolicy(withRoleName(&roleName),
					withSpecPolicyArn(&specPolicyArn)),
			},
			want: want{
				cr: rolePolicy(
					withRoleName(&roleName),
					withSpecPolicyArn(&specPolicyArn),
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
				iam: &fake.MockRolePolicyAttachmentClient{
					MockAttachRolePolicyRequest: func(input *awsiam.AttachRolePolicyInput) awsiam.AttachRolePolicyRequest {
						return awsiam.AttachRolePolicyRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: errBoom},
						}
					},
				},
				cr: rolePolicy(withRoleName(&roleName),
					withSpecPolicyArn(&specPolicyArn)),
			},
			want: want{
				cr: rolePolicy(withRoleName(&roleName),
					withSpecPolicyArn(&specPolicyArn),
					withConditions(xpv1.Creating())),
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
				iam: &fake.MockRolePolicyAttachmentClient{
					MockAttachRolePolicyRequest: func(input *awsiam.AttachRolePolicyInput) awsiam.AttachRolePolicyRequest {
						return awsiam.AttachRolePolicyRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsiam.AttachRolePolicyOutput{}},
						}
					},
					MockDetachRolePolicyRequest: func(input *awsiam.DetachRolePolicyInput) awsiam.DetachRolePolicyRequest {
						return awsiam.DetachRolePolicyRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsiam.DetachRolePolicyOutput{}},
						}
					},
				},
				cr: rolePolicy(withRoleName(&roleName),
					withSpecPolicyArn(&specPolicyArn)),
			},
			want: want{
				cr: rolePolicy(
					withRoleName(&roleName),
					withSpecPolicyArn(&specPolicyArn)),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.iam}
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
				iam: &fake.MockRolePolicyAttachmentClient{
					MockDetachRolePolicyRequest: func(input *awsiam.DetachRolePolicyInput) awsiam.DetachRolePolicyRequest {
						return awsiam.DetachRolePolicyRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awsiam.DetachRolePolicyOutput{}},
						}
					},
				},
				cr: rolePolicy(withRoleName(&roleName),
					withSpecPolicyArn(&specPolicyArn)),
			},
			want: want{
				cr: rolePolicy(
					withRoleName(&roleName),
					withSpecPolicyArn(&specPolicyArn),
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
				iam: &fake.MockRolePolicyAttachmentClient{
					MockDetachRolePolicyRequest: func(input *awsiam.DetachRolePolicyInput) awsiam.DetachRolePolicyRequest {
						return awsiam.DetachRolePolicyRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: errBoom},
						}
					},
				},
				cr: rolePolicy(withRoleName(&roleName),
					withSpecPolicyArn(&specPolicyArn)),
			},
			want: want{
				cr: rolePolicy(withRoleName(&roleName),
					withSpecPolicyArn(&specPolicyArn),
					withConditions(xpv1.Deleting())),
				err: errors.Wrap(errBoom, errDetach),
			},
		},
		"ResourceDoesNotExist": {
			args: args{
				iam: &fake.MockRolePolicyAttachmentClient{
					MockDetachRolePolicyRequest: func(input *awsiam.DetachRolePolicyInput) awsiam.DetachRolePolicyRequest {
						return awsiam.DetachRolePolicyRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: awserr.New(awsiam.ErrCodeNoSuchEntityException, "", nil)},
						}
					},
				},
				cr: rolePolicy(),
			},
			want: want{
				cr: rolePolicy(withConditions(xpv1.Deleting())),
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
