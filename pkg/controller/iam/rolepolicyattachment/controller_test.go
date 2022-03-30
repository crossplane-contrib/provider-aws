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

package rolepolicyattachment

import (
	"context"
	"testing"

	awsiam "github.com/aws/aws-sdk-go-v2/service/iam"
	awsiamtypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/test"

	v1beta1 "github.com/crossplane/provider-aws/apis/iam/v1beta1"
	awsclient "github.com/crossplane/provider-aws/pkg/clients"
	"github.com/crossplane/provider-aws/pkg/clients/iam"
	"github.com/crossplane/provider-aws/pkg/clients/iam/fake"
)

var (
	// an arbitrary managed resource
	unexpectedItem   resource.Managed
	roleName         = "some arbitrary name"
	specPolicyArn    = "some arbitrary arn"
	specPolicyArnTwo = "some other arbitrary arn"

	errBoom = errors.New("boom")
)

type args struct {
	iam iam.RolePolicyAttachmentClient
	cr  resource.Managed
}

type rolePolicyModifier func(*v1beta1.RolePolicyAttachment)

func withConditions(c ...xpv1.Condition) rolePolicyModifier {
	return func(r *v1beta1.RolePolicyAttachment) { r.Status.ConditionedStatus.Conditions = c }
}

func withRoleName(s *string) rolePolicyModifier {
	return func(r *v1beta1.RolePolicyAttachment) { r.Spec.ForProvider.RoleName = *s }
}

func withSpecPolicyArns(s ...string) rolePolicyModifier {
	return func(r *v1beta1.RolePolicyAttachment) { r.Spec.ForProvider.PolicyARNs = s }
}

func withStatusPolicyArns(s ...string) rolePolicyModifier {
	return func(r *v1beta1.RolePolicyAttachment) { r.Status.AtProvider.AttachedPolicyARNs = s }
}

func rolePolicy(m ...rolePolicyModifier) *v1beta1.RolePolicyAttachment {
	cr := &v1beta1.RolePolicyAttachment{}
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
					MockListAttachedRolePolicies: func(ctx context.Context, input *awsiam.ListAttachedRolePoliciesInput, opts []func(*awsiam.Options)) (*awsiam.ListAttachedRolePoliciesOutput, error) {
						return &awsiam.ListAttachedRolePoliciesOutput{
							AttachedPolicies: []awsiamtypes.AttachedPolicy{
								{
									PolicyArn: &specPolicyArn,
								},
							},
						}, nil
					},
				},
				cr: rolePolicy(withSpecPolicyArns(specPolicyArn)),
			},
			want: want{
				cr: rolePolicy(withSpecPolicyArns(specPolicyArn),
					withConditions(xpv1.Available()),
					withStatusPolicyArns(specPolicyArn)),
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
					MockListAttachedRolePolicies: func(ctx context.Context, input *awsiam.ListAttachedRolePoliciesInput, opts []func(*awsiam.Options)) (*awsiam.ListAttachedRolePoliciesOutput, error) {
						return nil, errBoom
					},
				},
				cr: rolePolicy(withRoleName(&roleName)),
			},
			want: want{
				cr:  rolePolicy(withRoleName(&roleName)),
				err: awsclient.Wrap(errBoom, errGet),
			},
		},
		"ResourceDoesNotExist": {
			args: args{
				iam: &fake.MockRolePolicyAttachmentClient{
					MockListAttachedRolePolicies: func(ctx context.Context, input *awsiam.ListAttachedRolePoliciesInput, opts []func(*awsiam.Options)) (*awsiam.ListAttachedRolePoliciesOutput, error) {
						return nil, &awsiamtypes.NoSuchEntityException{}
					},
				},
				cr: rolePolicy(),
			},
			want: want{
				cr: rolePolicy(),
			},
		},
		"PolicyNeedsCreate": {
			args: args{
				iam: &fake.MockRolePolicyAttachmentClient{
					MockListAttachedRolePolicies: func(ctx context.Context, input *awsiam.ListAttachedRolePoliciesInput, opts []func(*awsiam.Options)) (*awsiam.ListAttachedRolePoliciesOutput, error) {
						return &awsiam.ListAttachedRolePoliciesOutput{
							AttachedPolicies: []awsiamtypes.AttachedPolicy{
								{
									PolicyArn: &specPolicyArn,
								},
							},
						}, nil
					},
				},
				cr: rolePolicy(withRoleName(&roleName),
					withSpecPolicyArns(specPolicyArn, specPolicyArnTwo)),
			},
			want: want{
				cr: rolePolicy(
					withRoleName(&roleName),
					withSpecPolicyArns(specPolicyArn, specPolicyArnTwo),
					withConditions(xpv1.Available()),
					withStatusPolicyArns(specPolicyArn)),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: false,
				},
			},
		},
		"PolicyNeedsDelete": {
			args: args{
				iam: &fake.MockRolePolicyAttachmentClient{
					MockListAttachedRolePolicies: func(ctx context.Context, input *awsiam.ListAttachedRolePoliciesInput, opts []func(*awsiam.Options)) (*awsiam.ListAttachedRolePoliciesOutput, error) {
						return &awsiam.ListAttachedRolePoliciesOutput{
							AttachedPolicies: []awsiamtypes.AttachedPolicy{
								{
									PolicyArn: &specPolicyArn,
								},
								{
									PolicyArn: &specPolicyArnTwo,
								},
							},
						}, nil
					},
				},
				cr: rolePolicy(withRoleName(&roleName),
					withSpecPolicyArns(specPolicyArn)),
			},
			want: want{
				cr: rolePolicy(
					withRoleName(&roleName),
					withSpecPolicyArns(specPolicyArn),
					withConditions(xpv1.Available()),
					withStatusPolicyArns(specPolicyArn, specPolicyArnTwo)),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: false,
				},
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
					MockAttachRolePolicy: func(ctx context.Context, input *awsiam.AttachRolePolicyInput, opts []func(*awsiam.Options)) (*awsiam.AttachRolePolicyOutput, error) {
						return &awsiam.AttachRolePolicyOutput{}, nil
					},
				},
				cr: rolePolicy(withRoleName(&roleName),
					withSpecPolicyArns(specPolicyArn)),
			},
			want: want{
				cr: rolePolicy(
					withRoleName(&roleName),
					withSpecPolicyArns(specPolicyArn)),
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
					MockAttachRolePolicy: func(ctx context.Context, input *awsiam.AttachRolePolicyInput, opts []func(*awsiam.Options)) (*awsiam.AttachRolePolicyOutput, error) {
						return nil, errBoom
					},
				},
				cr: rolePolicy(withRoleName(&roleName),
					withSpecPolicyArns(specPolicyArn)),
			},
			want: want{
				cr: rolePolicy(withRoleName(&roleName),
					withSpecPolicyArns(specPolicyArn)),
				err: awsclient.Wrap(errBoom, errAttach),
			},
		},
		"DeprecatedInput": {
			args: args{
				iam: &fake.MockRolePolicyAttachmentClient{
					MockAttachRolePolicy: func(ctx context.Context, input *awsiam.AttachRolePolicyInput, opts []func(*awsiam.Options)) (*awsiam.AttachRolePolicyOutput, error) {
						return &awsiam.AttachRolePolicyOutput{}, nil
					},
				},
				cr: rolePolicy(withRoleName(&roleName),
					func(r *v1beta1.RolePolicyAttachment) { r.Spec.ForProvider.PolicyARN = specPolicyArn }),
			},
			want: want{
				cr: rolePolicy(
					withRoleName(&roleName),
					withSpecPolicyArns(specPolicyArn),
					func(r *v1beta1.RolePolicyAttachment) { r.Spec.ForProvider.PolicyARN = specPolicyArn },
				),
				result: managed.ExternalCreation{},
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
					MockAttachRolePolicy: func(ctx context.Context, input *awsiam.AttachRolePolicyInput, opts []func(*awsiam.Options)) (*awsiam.AttachRolePolicyOutput, error) {
						return &awsiam.AttachRolePolicyOutput{}, nil
					},
					MockListAttachedRolePolicies: func(ctx context.Context, input *awsiam.ListAttachedRolePoliciesInput, opts []func(*awsiam.Options)) (*awsiam.ListAttachedRolePoliciesOutput, error) {
						return &awsiam.ListAttachedRolePoliciesOutput{
							AttachedPolicies: []awsiamtypes.AttachedPolicy{
								{
									PolicyArn: &specPolicyArn,
								},
							},
						}, nil
					},
					MockDetachRolePolicy: func(ctx context.Context, input *awsiam.DetachRolePolicyInput, opts []func(*awsiam.Options)) (*awsiam.DetachRolePolicyOutput, error) {
						return &awsiam.DetachRolePolicyOutput{}, nil
					},
				},
				cr: rolePolicy(withRoleName(&roleName),
					withSpecPolicyArns(specPolicyArn, specPolicyArnTwo)),
			},
			want: want{
				cr: rolePolicy(
					withRoleName(&roleName),
					withSpecPolicyArns(specPolicyArn, specPolicyArnTwo)),
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
				iam: &fake.MockRolePolicyAttachmentClient{
					MockAttachRolePolicy: func(ctx context.Context, input *awsiam.AttachRolePolicyInput, opts []func(*awsiam.Options)) (*awsiam.AttachRolePolicyOutput, error) {
						return &awsiam.AttachRolePolicyOutput{}, nil
					},
					MockListAttachedRolePolicies: func(ctx context.Context, input *awsiam.ListAttachedRolePoliciesInput, opts []func(*awsiam.Options)) (*awsiam.ListAttachedRolePoliciesOutput, error) {
						return nil, errBoom
					},
					MockDetachRolePolicy: func(ctx context.Context, input *awsiam.DetachRolePolicyInput, opts []func(*awsiam.Options)) (*awsiam.DetachRolePolicyOutput, error) {
						return &awsiam.DetachRolePolicyOutput{}, nil
					},
				},
				cr: rolePolicy(withRoleName(&roleName),
					withSpecPolicyArns(specPolicyArn)),
			},
			want: want{
				cr: rolePolicy(withRoleName(&roleName),
					withSpecPolicyArns(specPolicyArn)),
				err: awsclient.Wrap(errBoom, errGet),
			},
		},
		"AttachClientError": {
			args: args{
				iam: &fake.MockRolePolicyAttachmentClient{
					MockAttachRolePolicy: func(ctx context.Context, input *awsiam.AttachRolePolicyInput, opts []func(*awsiam.Options)) (*awsiam.AttachRolePolicyOutput, error) {
						return nil, errBoom
					},
					MockListAttachedRolePolicies: func(ctx context.Context, input *awsiam.ListAttachedRolePoliciesInput, opts []func(*awsiam.Options)) (*awsiam.ListAttachedRolePoliciesOutput, error) {
						return &awsiam.ListAttachedRolePoliciesOutput{
							AttachedPolicies: []awsiamtypes.AttachedPolicy{
								{
									PolicyArn: &specPolicyArn,
								},
							},
						}, nil
					},
					MockDetachRolePolicy: func(ctx context.Context, input *awsiam.DetachRolePolicyInput, opts []func(*awsiam.Options)) (*awsiam.DetachRolePolicyOutput, error) {
						return &awsiam.DetachRolePolicyOutput{}, nil
					},
				},
				cr: rolePolicy(withRoleName(&roleName),
					withSpecPolicyArns(specPolicyArn, specPolicyArnTwo)),
			},
			want: want{
				cr: rolePolicy(withRoleName(&roleName),
					withSpecPolicyArns(specPolicyArn, specPolicyArnTwo)),
				err: awsclient.Wrap(errBoom, errAttach),
			},
		},
		"DetachClientError": {
			args: args{
				iam: &fake.MockRolePolicyAttachmentClient{
					MockAttachRolePolicy: func(ctx context.Context, input *awsiam.AttachRolePolicyInput, opts []func(*awsiam.Options)) (*awsiam.AttachRolePolicyOutput, error) {
						return &awsiam.AttachRolePolicyOutput{}, nil
					},
					MockListAttachedRolePolicies: func(ctx context.Context, input *awsiam.ListAttachedRolePoliciesInput, opts []func(*awsiam.Options)) (*awsiam.ListAttachedRolePoliciesOutput, error) {
						return &awsiam.ListAttachedRolePoliciesOutput{
							AttachedPolicies: []awsiamtypes.AttachedPolicy{
								{
									PolicyArn: &specPolicyArn,
								},
								{
									PolicyArn: &specPolicyArnTwo,
								},
							},
						}, nil
					},
					MockDetachRolePolicy: func(ctx context.Context, input *awsiam.DetachRolePolicyInput, opts []func(*awsiam.Options)) (*awsiam.DetachRolePolicyOutput, error) {
						return nil, errBoom
					},
				},
				cr: rolePolicy(withRoleName(&roleName),
					withSpecPolicyArns(specPolicyArn)),
			},
			want: want{
				cr: rolePolicy(withRoleName(&roleName),
					withSpecPolicyArns(specPolicyArn)),
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
				iam: &fake.MockRolePolicyAttachmentClient{
					MockDetachRolePolicy: func(ctx context.Context, input *awsiam.DetachRolePolicyInput, opts []func(*awsiam.Options)) (*awsiam.DetachRolePolicyOutput, error) {
						return &awsiam.DetachRolePolicyOutput{}, nil
					},
				},
				cr: rolePolicy(withRoleName(&roleName),
					withSpecPolicyArns(specPolicyArn)),
			},
			want: want{
				cr: rolePolicy(
					withRoleName(&roleName),
					withSpecPolicyArns(specPolicyArn)),
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
					MockDetachRolePolicy: func(ctx context.Context, input *awsiam.DetachRolePolicyInput, opts []func(*awsiam.Options)) (*awsiam.DetachRolePolicyOutput, error) {
						return nil, errBoom
					},
				},
				cr: rolePolicy(withRoleName(&roleName),
					withSpecPolicyArns(specPolicyArn)),
			},
			want: want{
				cr: rolePolicy(withRoleName(&roleName),
					withSpecPolicyArns(specPolicyArn)),
				err: awsclient.Wrap(errBoom, errDetach),
			},
		},
		"ResourceDoesNotExist": {
			args: args{
				iam: &fake.MockRolePolicyAttachmentClient{
					MockDetachRolePolicy: func(ctx context.Context, input *awsiam.DetachRolePolicyInput, opts []func(*awsiam.Options)) (*awsiam.DetachRolePolicyOutput, error) {
						return nil, &awsiamtypes.NoSuchEntityException{}
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
