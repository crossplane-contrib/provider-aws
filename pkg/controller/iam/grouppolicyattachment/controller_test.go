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

package grouppolicyattachment

import (
	"context"
	"testing"

	"github.com/crossplane/provider-aws/apis/iam/v1beta1"

	awsiam "github.com/aws/aws-sdk-go-v2/service/iam"
	awsiamtypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/test"

	awsclient "github.com/crossplane/provider-aws/pkg/clients"
	"github.com/crossplane/provider-aws/pkg/clients/iam"
	"github.com/crossplane/provider-aws/pkg/clients/iam/fake"
)

var (
	unexpectedItem resource.Managed
	policyArn      = "some arn"
	policyArnTwo   = "some other arn"
	groupName      = "some group"

	errBoom = errors.New("boom")
)

type args struct {
	iam iam.GroupPolicyAttachmentClient
	cr  resource.Managed
}

type groupPolicyModifier func(*v1beta1.GroupPolicyAttachment)

func withConditions(c ...xpv1.Condition) groupPolicyModifier {
	return func(r *v1beta1.GroupPolicyAttachment) { r.Status.ConditionedStatus.Conditions = c }
}

func withSpecGroupName(s string) groupPolicyModifier {
	return func(r *v1beta1.GroupPolicyAttachment) { r.Spec.ForProvider.GroupName = s }
}

func withSpecPolicyArns(s ...string) groupPolicyModifier {
	var arns []string
	for _, i := range s {
		arns = append(arns, i)
	}
	return func(r *v1beta1.GroupPolicyAttachment) { r.Spec.ForProvider.PolicyARNs = arns }
}

func withStatusPolicyArns(s ...string) groupPolicyModifier {
	var arns []string
	for _, i := range s {
		arns = append(arns, i)
	}
	return func(r *v1beta1.GroupPolicyAttachment) { r.Status.AtProvider.AttachedPolicyARNs = arns }
}

func groupPolicy(m ...groupPolicyModifier) *v1beta1.GroupPolicyAttachment {
	cr := &v1beta1.GroupPolicyAttachment{}
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
				iam: &fake.MockGroupPolicyAttachmentClient{
					MockListAttachedGroupPolicies: func(ctx context.Context, input *awsiam.ListAttachedGroupPoliciesInput, opts []func(*awsiam.Options)) (*awsiam.ListAttachedGroupPoliciesOutput, error) {
						return &awsiam.ListAttachedGroupPoliciesOutput{
							AttachedPolicies: []awsiamtypes.AttachedPolicy{
								{
									PolicyArn: &policyArn,
								},
							},
						}, nil
					},
				},
				cr: groupPolicy(withSpecGroupName(groupName),
					withSpecPolicyArns(policyArn)),
			},
			want: want{
				cr: groupPolicy(
					withSpecGroupName(groupName),
					withSpecPolicyArns(policyArn),
					withConditions(xpv1.Available()),
					withStatusPolicyArns(policyArn)),
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
				iam: &fake.MockGroupPolicyAttachmentClient{
					MockListAttachedGroupPolicies: func(ctx context.Context, input *awsiam.ListAttachedGroupPoliciesInput, opts []func(*awsiam.Options)) (*awsiam.ListAttachedGroupPoliciesOutput, error) {
						return &awsiam.ListAttachedGroupPoliciesOutput{}, nil
					},
				},
				cr: groupPolicy(withSpecPolicyArns(policyArn)),
			},
			want: want{
				cr: groupPolicy(withSpecPolicyArns(policyArn)),
			},
		},
		"ClientError": {
			args: args{
				iam: &fake.MockGroupPolicyAttachmentClient{
					MockListAttachedGroupPolicies: func(ctx context.Context, input *awsiam.ListAttachedGroupPoliciesInput, opts []func(*awsiam.Options)) (*awsiam.ListAttachedGroupPoliciesOutput, error) {
						return nil, errBoom
					},
				},
				cr: groupPolicy(withSpecGroupName(groupName)),
			},
			want: want{
				cr:  groupPolicy(withSpecGroupName(groupName)),
				err: awsclient.Wrap(errBoom, errGet),
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
				iam: &fake.MockGroupPolicyAttachmentClient{
					MockAttachGroupPolicy: func(ctx context.Context, input *awsiam.AttachGroupPolicyInput, opts []func(*awsiam.Options)) (*awsiam.AttachGroupPolicyOutput, error) {
						return &awsiam.AttachGroupPolicyOutput{}, nil
					},
				},
				cr: groupPolicy(withSpecGroupName(groupName),
					withSpecPolicyArns(policyArn)),
			},
			want: want{
				cr: groupPolicy(
					withSpecGroupName(groupName),
					withSpecPolicyArns(policyArn),
				),
				result: managed.ExternalCreation{},
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
				iam: &fake.MockGroupPolicyAttachmentClient{
					MockAttachGroupPolicy: func(ctx context.Context, input *awsiam.AttachGroupPolicyInput, opts []func(*awsiam.Options)) (*awsiam.AttachGroupPolicyOutput, error) {
						return nil, errBoom
					},
				},
				cr: groupPolicy(withSpecGroupName(groupName),
					withSpecPolicyArns(policyArn)),
			},
			want: want{
				cr: groupPolicy(withSpecGroupName(groupName),
					withSpecPolicyArns(policyArn)),
				err: awsclient.Wrap(errBoom, errAttach),
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
				iam: &fake.MockGroupPolicyAttachmentClient{
					MockAttachGroupPolicy: func(ctx context.Context, input *awsiam.AttachGroupPolicyInput, opts []func(*awsiam.Options)) (*awsiam.AttachGroupPolicyOutput, error) {
						return &awsiam.AttachGroupPolicyOutput{}, nil
					},
					MockListAttachedGroupPolicies: func(ctx context.Context, input *awsiam.ListAttachedGroupPoliciesInput, opts []func(*awsiam.Options)) (*awsiam.ListAttachedGroupPoliciesOutput, error) {
						return &awsiam.ListAttachedGroupPoliciesOutput{
							AttachedPolicies: []awsiamtypes.AttachedPolicy{
								{
									PolicyArn: &policyArn,
								},
							},
						}, nil
					},
					MockDetachGroupPolicy: func(ctx context.Context, input *awsiam.DetachGroupPolicyInput, opts []func(*awsiam.Options)) (*awsiam.DetachGroupPolicyOutput, error) {
						return &awsiam.DetachGroupPolicyOutput{}, nil
					},
				},
				cr: groupPolicy(withSpecGroupName(groupName),
					withSpecPolicyArns(policyArn, policyArnTwo)),
			},
			want: want{
				cr: groupPolicy(
					withSpecGroupName(groupName),
					withSpecPolicyArns(policyArn, policyArnTwo)),
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
		"ObserveClientError": {
			args: args{
				iam: &fake.MockGroupPolicyAttachmentClient{
					MockAttachGroupPolicy: func(ctx context.Context, input *awsiam.AttachGroupPolicyInput, opts []func(*awsiam.Options)) (*awsiam.AttachGroupPolicyOutput, error) {
						return &awsiam.AttachGroupPolicyOutput{}, nil
					},
					MockListAttachedGroupPolicies: func(ctx context.Context, input *awsiam.ListAttachedGroupPoliciesInput, opts []func(*awsiam.Options)) (*awsiam.ListAttachedGroupPoliciesOutput, error) {
						return nil, errBoom
					},
					MockDetachGroupPolicy: func(ctx context.Context, input *awsiam.DetachGroupPolicyInput, opts []func(*awsiam.Options)) (*awsiam.DetachGroupPolicyOutput, error) {
						return &awsiam.DetachGroupPolicyOutput{}, nil
					},
				},
				cr: groupPolicy(withSpecGroupName(groupName),
					withSpecPolicyArns(policyArn)),
			},
			want: want{
				cr: groupPolicy(withSpecGroupName(groupName),
					withSpecPolicyArns(policyArn)),
				err: awsclient.Wrap(errBoom, errGet),
			},
		},
		"AttachClientError": {
			args: args{
				iam: &fake.MockGroupPolicyAttachmentClient{
					MockAttachGroupPolicy: func(ctx context.Context, input *awsiam.AttachGroupPolicyInput, opts []func(*awsiam.Options)) (*awsiam.AttachGroupPolicyOutput, error) {
						return nil, errBoom
					},
					MockListAttachedGroupPolicies: func(ctx context.Context, input *awsiam.ListAttachedGroupPoliciesInput, opts []func(*awsiam.Options)) (*awsiam.ListAttachedGroupPoliciesOutput, error) {
						return &awsiam.ListAttachedGroupPoliciesOutput{
							AttachedPolicies: []awsiamtypes.AttachedPolicy{
								{
									PolicyArn: &policyArn,
								},
							},
						}, nil
					},
					MockDetachGroupPolicy: func(ctx context.Context, input *awsiam.DetachGroupPolicyInput, opts []func(*awsiam.Options)) (*awsiam.DetachGroupPolicyOutput, error) {
						return &awsiam.DetachGroupPolicyOutput{}, nil
					},
				},
				cr: groupPolicy(withSpecGroupName(groupName),
					withSpecPolicyArns(policyArn, policyArnTwo)),
			},
			want: want{
				cr: groupPolicy(withSpecGroupName(groupName),
					withSpecPolicyArns(policyArn, policyArnTwo)),
				err: awsclient.Wrap(errBoom, errAttach),
			},
		},
		"DetachClientError": {
			args: args{
				iam: &fake.MockGroupPolicyAttachmentClient{
					MockAttachGroupPolicy: func(ctx context.Context, input *awsiam.AttachGroupPolicyInput, opts []func(*awsiam.Options)) (*awsiam.AttachGroupPolicyOutput, error) {
						return &awsiam.AttachGroupPolicyOutput{}, nil
					},
					MockListAttachedGroupPolicies: func(ctx context.Context, input *awsiam.ListAttachedGroupPoliciesInput, opts []func(*awsiam.Options)) (*awsiam.ListAttachedGroupPoliciesOutput, error) {
						return &awsiam.ListAttachedGroupPoliciesOutput{
							AttachedPolicies: []awsiamtypes.AttachedPolicy{
								{
									PolicyArn: &policyArn,
								},
								{
									PolicyArn: &policyArnTwo,
								},
							},
						}, nil
					},
					MockDetachGroupPolicy: func(ctx context.Context, input *awsiam.DetachGroupPolicyInput, opts []func(*awsiam.Options)) (*awsiam.DetachGroupPolicyOutput, error) {
						return nil, errBoom
					},
				},
				cr: groupPolicy(withSpecGroupName(groupName),
					withSpecPolicyArns(policyArn)),
			},
			want: want{
				cr: groupPolicy(withSpecGroupName(groupName),
					withSpecPolicyArns(policyArn)),
				err: awsclient.Wrap(errBoom, errDetach),
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
				iam: &fake.MockGroupPolicyAttachmentClient{
					MockDetachGroupPolicy: func(ctx context.Context, input *awsiam.DetachGroupPolicyInput, opts []func(*awsiam.Options)) (*awsiam.DetachGroupPolicyOutput, error) {
						return &awsiam.DetachGroupPolicyOutput{}, nil
					},
				},
				cr: groupPolicy(withSpecGroupName(groupName),
					withSpecPolicyArns(policyArn)),
			},
			want: want{
				cr: groupPolicy(
					withSpecGroupName(groupName),
					withSpecPolicyArns(policyArn),
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
				iam: &fake.MockGroupPolicyAttachmentClient{
					MockDetachGroupPolicy: func(ctx context.Context, input *awsiam.DetachGroupPolicyInput, opts []func(*awsiam.Options)) (*awsiam.DetachGroupPolicyOutput, error) {
						return nil, errBoom
					},
				},
				cr: groupPolicy(withSpecGroupName(groupName),
					withSpecPolicyArns(policyArn)),
			},
			want: want{
				cr: groupPolicy(withSpecGroupName(groupName),
					withSpecPolicyArns(policyArn),
					withConditions(xpv1.Deleting())),
				err: awsclient.Wrap(errBoom, errDetach),
			},
		},
		"ResourceDoesNotExist": {
			args: args{
				iam: &fake.MockGroupPolicyAttachmentClient{
					MockDetachGroupPolicy: func(ctx context.Context, input *awsiam.DetachGroupPolicyInput, opts []func(*awsiam.Options)) (*awsiam.DetachGroupPolicyOutput, error) {
						return nil, &awsiamtypes.NoSuchEntityException{}
					},
				},
				cr: groupPolicy(),
			},
			want: want{
				cr:  groupPolicy(withConditions(xpv1.Deleting())),
				err: nil,
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
