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

	awsiam "github.com/aws/aws-sdk-go-v2/service/iam"
	awsiamtypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"

	"github.com/crossplane-contrib/provider-aws/apis/iam/v1beta1"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/iam"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/iam/fake"
	errorutils "github.com/crossplane-contrib/provider-aws/pkg/utils/errors"
)

var (
	unexpectedItem resource.Managed
	policyArn      = "some arn"
	groupName      = "some group"

	errBoom = errors.New("boom")
)

type args struct {
	iam iam.GroupPolicyAttachmentClient
	cr  resource.Managed
}

type groupPolicyModifier func(*v1beta1.GroupPolicyAttachment)

func withExternalName(name string) groupPolicyModifier {
	return func(r *v1beta1.GroupPolicyAttachment) { meta.SetExternalName(r, name) }
}

func withConditions(c ...xpv1.Condition) groupPolicyModifier {
	return func(r *v1beta1.GroupPolicyAttachment) { r.Status.ConditionedStatus.Conditions = c }
}

func withSpecGroupName(s string) groupPolicyModifier {
	return func(r *v1beta1.GroupPolicyAttachment) { r.Spec.ForProvider.GroupName = s }
}

func withSpecPolicyArn(s string) groupPolicyModifier {
	return func(r *v1beta1.GroupPolicyAttachment) { r.Spec.ForProvider.PolicyARN = s }
}

func withStatusPolicyArn(s string) groupPolicyModifier {
	return func(r *v1beta1.GroupPolicyAttachment) { r.Status.AtProvider.AttachedPolicyARN = s }
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
					withExternalName(groupName+"/"+policyArn),
					withSpecPolicyArn(policyArn)),
			},
			want: want{
				cr: groupPolicy(
					withExternalName(groupName+"/"+policyArn),
					withSpecGroupName(groupName),
					withSpecPolicyArn(policyArn),
					withConditions(xpv1.Available()),
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
				iam: &fake.MockGroupPolicyAttachmentClient{
					MockListAttachedGroupPolicies: func(ctx context.Context, input *awsiam.ListAttachedGroupPoliciesInput, opts []func(*awsiam.Options)) (*awsiam.ListAttachedGroupPoliciesOutput, error) {
						return &awsiam.ListAttachedGroupPoliciesOutput{}, nil
					},
				},
				cr: groupPolicy(withSpecPolicyArn(policyArn)),
			},
			want: want{
				cr: groupPolicy(withSpecPolicyArn(policyArn)),
			},
		},
		"ClientError": {
			args: args{
				iam: &fake.MockGroupPolicyAttachmentClient{
					MockListAttachedGroupPolicies: func(ctx context.Context, input *awsiam.ListAttachedGroupPoliciesInput, opts []func(*awsiam.Options)) (*awsiam.ListAttachedGroupPoliciesOutput, error) {
						return nil, errBoom
					},
				},
				cr: groupPolicy(withSpecGroupName(groupName), withExternalName(groupName+"/"+policyArn)),
			},
			want: want{
				cr:  groupPolicy(withSpecGroupName(groupName), withExternalName(groupName+"/"+policyArn)),
				err: errorutils.Wrap(errBoom, errGet),
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
					withSpecPolicyArn(policyArn)),
			},
			want: want{
				cr: groupPolicy(
					withSpecGroupName(groupName),
					withSpecPolicyArn(policyArn),
					withExternalName(groupName+"/"+policyArn)),
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
					withSpecPolicyArn(policyArn)),
			},
			want: want{
				cr: groupPolicy(withSpecGroupName(groupName),
					withSpecPolicyArn(policyArn)),
				err: errorutils.Wrap(errBoom, errAttach),
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
				iam: &fake.MockGroupPolicyAttachmentClient{
					MockDetachGroupPolicy: func(ctx context.Context, input *awsiam.DetachGroupPolicyInput, opts []func(*awsiam.Options)) (*awsiam.DetachGroupPolicyOutput, error) {
						return &awsiam.DetachGroupPolicyOutput{}, nil
					},
				},
				cr: groupPolicy(withSpecGroupName(groupName),
					withSpecPolicyArn(policyArn)),
			},
			want: want{
				cr: groupPolicy(
					withSpecGroupName(groupName),
					withSpecPolicyArn(policyArn),
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
					withSpecPolicyArn(policyArn)),
			},
			want: want{
				cr: groupPolicy(withSpecGroupName(groupName),
					withSpecPolicyArn(policyArn),
					withConditions(xpv1.Deleting())),
				err: errorutils.Wrap(errBoom, errDetach),
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
