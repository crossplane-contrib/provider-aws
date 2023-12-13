/*
Copyright 2023 The Crossplane Authors.

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

package rolepolicy

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsiam "github.com/aws/aws-sdk-go-v2/service/iam"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	extv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane-contrib/provider-aws/apis/iam/v1beta1"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/iam"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/iam/fake"
	errorutils "github.com/crossplane-contrib/provider-aws/pkg/utils/errors"
)

var (
	unexpectedItem resource.Managed
	arn            = "arn:aws:iam::aws:policy/aws-service-role/AccessAnalyzerServiceRolePolicy"
	documentRaw    = `{
		"Version": "2012-10-17",
		"Statement": [
		  {
			  "Sid": "VisualEditor0",
			  "Effect": "Allow",
			  "Action": "elastic-inference:Connect",
			  "Resource": "*"
		  }
		]
	  }`
	document   = makeDocument(documentRaw)
	roleName   = "my-role"
	policyName = "my-policy"

	errBoom = errors.New("boom")
)

func makeDocument(raw string) extv1.JSON {
	return extv1.JSON{
		Raw: []byte(raw),
	}
}

type args struct {
	kube client.Client
	iam  iam.RolePolicyClient
	cr   resource.Managed
}

type rolePolicyModifier func(*v1beta1.RolePolicy)

func withExternalName(s string) rolePolicyModifier {
	return func(r *v1beta1.RolePolicy) { meta.SetExternalName(r, s) }
}

func withConditions(c ...xpv1.Condition) rolePolicyModifier {
	return func(r *v1beta1.RolePolicy) { r.Status.ConditionedStatus.Conditions = c }
}

func withSpec(spec v1beta1.RolePolicyParameters) rolePolicyModifier {
	return func(r *v1beta1.RolePolicy) {
		r.Spec.ForProvider = spec
	}
}

func rolePolicy(m ...rolePolicyModifier) *v1beta1.RolePolicy {
	cr := &v1beta1.RolePolicy{}
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
		"Successful": {
			args: args{
				iam: &fake.MockRolePolicyClient{
					MockGetRolePolicy: func(ctx context.Context, input *awsiam.GetRolePolicyInput, opts []func(*awsiam.Options)) (*awsiam.GetRolePolicyOutput, error) {
						return &awsiam.GetRolePolicyOutput{
							PolicyName:     aws.String(policyName),
							RoleName:       aws.String(roleName),
							PolicyDocument: &documentRaw,
						}, nil
					},
				},
				cr: rolePolicy(withSpec(v1beta1.RolePolicyParameters{
					Document: document,
					RoleName: roleName,
				}), withExternalName(arn)),
			},
			want: want{
				cr: rolePolicy(withSpec(v1beta1.RolePolicyParameters{
					Document: document,
					RoleName: roleName,
				}), withExternalName(arn),
					withConditions(xpv1.Available())),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
				},
			},
		},
		"InvalidInput": {
			args: args{
				cr: unexpectedItem,
			},
			want: want{
				cr:  unexpectedItem,
				err: errors.New(errUnexpectedObject),
			},
		},
		"GetRolePolicyError": {
			args: args{
				iam: &fake.MockRolePolicyClient{
					MockGetRolePolicy: func(ctx context.Context, input *awsiam.GetRolePolicyInput, opts []func(*awsiam.Options)) (*awsiam.GetRolePolicyOutput, error) {
						return nil, errBoom
					},
				},
				cr: rolePolicy(withExternalName(arn)),
			},
			want: want{
				cr:  rolePolicy(withExternalName(arn)),
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
		"Successful": {
			args: args{
				kube: &test.MockClient{
					MockStatusUpdate: test.NewMockClient().MockStatusUpdate,
				},
				iam: &fake.MockRolePolicyClient{
					MockPutRolePolicy: func(ctx context.Context, input *awsiam.PutRolePolicyInput, opts []func(*awsiam.Options)) (*awsiam.PutRolePolicyOutput, error) {
						return &awsiam.PutRolePolicyOutput{}, nil
					},
				},
				cr: rolePolicy(withSpec(v1beta1.RolePolicyParameters{
					Document: document,
					RoleName: roleName,
				}),
					withExternalName(arn)),
			},
			want: want{
				cr: rolePolicy(
					withSpec(v1beta1.RolePolicyParameters{
						Document: document,
						RoleName: roleName,
					}),
					withExternalName(arn)),
				result: managed.ExternalCreation{},
			},
		},
		"InvalidInput": {
			args: args{
				cr: unexpectedItem,
			},
			want: want{
				cr:  unexpectedItem,
				err: errors.New(errUnexpectedObject),
			},
		},
		"EnsurePutRolePolicyThrowsWithCorrectException": {
			args: args{
				iam: &fake.MockRolePolicyClient{
					MockPutRolePolicy: func(ctx context.Context, input *awsiam.PutRolePolicyInput, opts []func(*awsiam.Options)) (*awsiam.PutRolePolicyOutput, error) {
						return nil, errBoom
					},
				},
				cr: rolePolicy(withSpec(v1beta1.RolePolicyParameters{
					Document: makeDocument("{}"),
				})),
			},
			want: want{
				cr: rolePolicy(withSpec(v1beta1.RolePolicyParameters{
					Document: makeDocument("{}"),
				})),
				err: errors.Wrap(errBoom, errPutRolePolicy),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: test.NewMockClient(), client: tc.iam}
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
		"Successful": {
			args: args{
				iam: &fake.MockRolePolicyClient{
					MockPutRolePolicy: func(ctx context.Context, input *awsiam.PutRolePolicyInput, opts []func(*awsiam.Options)) (*awsiam.PutRolePolicyOutput, error) {
						return &awsiam.PutRolePolicyOutput{}, nil
					},
				},
				cr: rolePolicy(withSpec(v1beta1.RolePolicyParameters{
					Document: document,
					RoleName: roleName,
				}), withExternalName(arn)),
			},
			want: want{
				cr: rolePolicy(
					withSpec(v1beta1.RolePolicyParameters{
						Document: document,
						RoleName: roleName,
					}),
					withExternalName(arn)),
				result: managed.ExternalUpdate{},
			},
		},
		"InvalidInput": {
			args: args{
				cr: unexpectedItem,
			},
			want: want{
				cr:  unexpectedItem,
				err: errors.New(errUnexpectedObject),
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
		"Successful": {
			args: args{
				iam: &fake.MockRolePolicyClient{
					MockDeleteRolePolicy: func(ctx context.Context, input *awsiam.DeleteRolePolicyInput, opts []func(*awsiam.Options)) (*awsiam.DeleteRolePolicyOutput, error) {
						return &awsiam.DeleteRolePolicyOutput{}, nil
					},
				},
				cr: rolePolicy(withExternalName(arn)),
			},
			want: want{
				cr: rolePolicy(withExternalName(arn),
					withConditions(xpv1.Deleting())),
			},
		},
		"InvalidInput": {
			args: args{
				cr: unexpectedItem,
			},
			want: want{
				cr:  unexpectedItem,
				err: errors.New(errUnexpectedObject),
			},
		},
		"DeleteRolePolicyError": {
			args: args{
				iam: &fake.MockRolePolicyClient{
					MockDeleteRolePolicy: func(ctx context.Context, input *awsiam.DeleteRolePolicyInput, opts []func(*awsiam.Options)) (*awsiam.DeleteRolePolicyOutput, error) {
						return nil, errBoom
					},
				},
				cr: rolePolicy(withExternalName(arn)),
			},
			want: want{
				cr: rolePolicy(withExternalName(arn),
					withConditions(xpv1.Deleting())),
				err: errBoom,
			},
		},
		"DeleteRolePolicyErrorNoSuchEntityException": {
			args: args{
				iam: &fake.MockRolePolicyClient{
					MockDeleteRolePolicy: func(ctx context.Context, input *awsiam.DeleteRolePolicyInput, opts []func(*awsiam.Options)) (*awsiam.DeleteRolePolicyOutput, error) {
						return nil, errors.New(iam.ErrRolePolicyNotFound)
					},
				},
				cr: rolePolicy(withExternalName(arn)),
			},
			want: want{
				cr: rolePolicy(withExternalName(arn),
					withConditions(xpv1.Deleting())),
				err: errors.New(iam.ErrRolePolicyNotFound),
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

func TestIsInlinePolicyUpToDate(t *testing.T) {
	type args struct {
		cr       extv1.JSON
		external *string
	}

	cases := map[string]struct {
		args args
		want bool
	}{
		"SameFields": {
			args: args{
				cr: makeDocument(`{
		"Version": "2012-10-17",
		"Statement": [
		  {
			"Effect": "Allow",
			"Principal": {
			  "Service": "eks.amazonaws.com"
			},
			"Action": "sts:AssumeRole"
		  }
		]
	   }`),
				external: aws.String(`{
		"Version": "2012-10-17",
		"Statement": [
		  {
			"Effect": "Allow",
			"Principal": {
			  "Service": "eks.amazonaws.com"
			},
			"Action": "sts:AssumeRole"
		  }
		]
	   }`),
			},
			want: true,
		},
		"SameFieldsEscaped": {
			args: args{
				cr: makeDocument(`{
		"Version": "2012-10-17",
		"Statement": [
		  {
			"Effect": "Allow",
			"Principal": {
			  "Service": "eks.amazonaws.com"
			},
			"Action": "sts:AssumeRole"
		  }
		]
	   }`),
				external: aws.String(`%7B%22Version%22%3A%222012-10-17%22%2C%22Statement%22%3A%5B%7B%22Effect%22%3A%22Allow%22%2C%22Principal%22%3A%7B%22Service%22%3A%22eks.amazonaws.com%22%7D%2C%22Action%22%3A%22sts%3AAssumeRole%22%7D%5D%7D`),
			},
			want: true,
		},
		"DifferentFields": {
			args: args{
				cr: makeDocument(`{
		"Version": "2012-10-17",
		"Statement": [
		  {
			"Effect": "Allow",
			"Principal": {
			  "Service": "eks.amazonaws.com"
			},
			"Action": "sts:*"
		  }
		]
	   }`),
				external: aws.String(`{
		"Version": "2012-10-17",
		"Statement": [
		  {
			"Effect": "Allow",
			"Principal": {
			  "Service": "eks.amazonaws.com"
			},
			"Action": "sts:AssumeRole"
		  }
		]
	   }`),
			},
			want: false,
		},
		"SameActionArray": {
			args: args{
				cr: makeDocument(`{
		"Version": "2012-10-17",
		"Statement": [
		  {
			"Effect": "Allow",
			"Principal": {
			  "Service": "eks.amazonaws.com"
			},
			"Action": ["sts:AssumeRole"]
		  }
		]
	   }`),
				external: aws.String(`{
		"Version": "2012-10-17",
		"Statement": [
		  {
			"Effect": "Allow",
			"Principal": {
			  "Service": "eks.amazonaws.com"
			},
			"Action": "sts:AssumeRole"
		  }
		]
	   }`),
			},
			want: true,
		},
		"DifferentActionArray": {
			args: args{
				cr: makeDocument(`{
		"Version": "2012-10-17",
		"Statement": [
		  {
			"Effect": "Allow",
			"Principal": {
			  "Service": "eks.amazonaws.com"
			},
			"Action": ["sts:AssumeRole", "sts:GetFederationToken"]
		  }
		]
	   }`),
				external: aws.String(`{
		"Version": "2012-10-17",
		"Statement": [
		  {
			"Effect": "Allow",
			"Principal": {
			  "Service": "eks.amazonaws.com"
			},
			"Action": "sts:AssumeRole"
		  }
		]
	   }`),
			},
			want: false,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got, _, _ := IsInlinePolicyUpToDate(string(tc.args.cr.Raw), tc.args.external)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}
