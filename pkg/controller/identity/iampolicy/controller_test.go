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

package iampolicy

import (
	"context"
	"testing"

	awsiam "github.com/aws/aws-sdk-go-v2/service/iam"
	awsiamtypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/test"

	"github.com/crossplane/provider-aws/apis/identity/v1alpha1"
	awsclient "github.com/crossplane/provider-aws/pkg/clients"
	"github.com/crossplane/provider-aws/pkg/clients/iam"
	"github.com/crossplane/provider-aws/pkg/clients/iam/fake"
)

var (
	unexpectedItem resource.Managed
	arn            = "arn:aws:iam::aws:policy/aws-service-role/AccessAnalyzerServiceRolePolicy"
	name           = "policy name"
	document       = `{
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
	boolFalse = false

	errBoom = errors.New("boom")
)

type args struct {
	kube client.Client
	iam  iam.PolicyClient
	cr   resource.Managed
}

type policyModifier func(*v1alpha1.IAMPolicy)

func withExterName(s string) policyModifier {
	return func(r *v1alpha1.IAMPolicy) { meta.SetExternalName(r, s) }
}

func withConditions(c ...xpv1.Condition) policyModifier {
	return func(r *v1alpha1.IAMPolicy) { r.Status.ConditionedStatus.Conditions = c }
}

func withSpec(spec v1alpha1.IAMPolicyParameters) policyModifier {
	return func(r *v1alpha1.IAMPolicy) {
		r.Spec.ForProvider = spec
	}
}

func policy(m ...policyModifier) *v1alpha1.IAMPolicy {
	cr := &v1alpha1.IAMPolicy{}
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
				iam: &fake.MockPolicyClient{
					MockGetPolicy: func(ctx context.Context, input *awsiam.GetPolicyInput, opts []func(*awsiam.Options)) (*awsiam.GetPolicyOutput, error) {
						return &awsiam.GetPolicyOutput{
							Policy: &awsiamtypes.Policy{},
						}, nil
					},
					MockGetPolicyVersion: func(ctx context.Context, input *awsiam.GetPolicyVersionInput, opts []func(*awsiam.Options)) (*awsiam.GetPolicyVersionOutput, error) {
						return &awsiam.GetPolicyVersionOutput{
							PolicyVersion: &awsiamtypes.PolicyVersion{
								Document: &document,
							},
						}, nil
					},
				},
				cr: policy(withSpec(v1alpha1.IAMPolicyParameters{
					Document: document,
					Name:     name,
				}), withExterName(arn)),
			},
			want: want{
				cr: policy(withSpec(v1alpha1.IAMPolicyParameters{
					Document: document,
					Name:     name,
				}), withExterName(arn),
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
		"GetUPolicyError": {
			args: args{
				iam: &fake.MockPolicyClient{
					MockGetPolicy: func(ctx context.Context, input *awsiam.GetPolicyInput, opts []func(*awsiam.Options)) (*awsiam.GetPolicyOutput, error) {
						return nil, errBoom
					},
				},
				cr: policy(withExterName(arn)),
			},
			want: want{
				cr:  policy(withExterName(arn)),
				err: awsclient.Wrap(errBoom, errGet),
			},
		},
		"EmptySpecPolicy": {
			args: args{
				iam: &fake.MockPolicyClient{
					MockGetPolicy: func(ctx context.Context, input *awsiam.GetPolicyInput, opts []func(*awsiam.Options)) (*awsiam.GetPolicyOutput, error) {
						return &awsiam.GetPolicyOutput{
							Policy: &awsiamtypes.Policy{},
						}, nil
					},
					MockGetPolicyVersion: func(ctx context.Context, input *awsiam.GetPolicyVersionInput, opts []func(*awsiam.Options)) (*awsiam.GetPolicyVersionOutput, error) {
						return &awsiam.GetPolicyVersionOutput{
							PolicyVersion: &awsiamtypes.PolicyVersion{
								Document: &document,
							},
						}, nil
					},
				},
				cr: policy(withExterName(arn)),
			},
			want: want{
				cr: policy(withExterName(arn),
					withConditions(xpv1.Available())),
				result: managed.ExternalObservation{
					ResourceExists: true,
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
		"Successful": {
			args: args{
				kube: &test.MockClient{
					MockStatusUpdate: test.NewMockStatusUpdateFn(nil),
				},
				iam: &fake.MockPolicyClient{
					MockCreatePolicy: func(ctx context.Context, input *awsiam.CreatePolicyInput, opts []func(*awsiam.Options)) (*awsiam.CreatePolicyOutput, error) {
						return &awsiam.CreatePolicyOutput{
							Policy: &awsiamtypes.Policy{
								Arn: &arn,
							},
						}, nil
					},
				},
				cr: policy(withSpec(v1alpha1.IAMPolicyParameters{
					Document: document,
					Name:     name,
				})),
			},
			want: want{
				cr: policy(
					withSpec(v1alpha1.IAMPolicyParameters{
						Document: document,
						Name:     name,
					}),
					withExterName(arn)),
				result: managed.ExternalCreation{ExternalNameAssigned: true},
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
				iam: &fake.MockPolicyClient{
					MockCreatePolicy: func(ctx context.Context, input *awsiam.CreatePolicyInput, opts []func(*awsiam.Options)) (*awsiam.CreatePolicyOutput, error) {
						return nil, errBoom
					},
				},
				cr: policy(),
			},
			want: want{
				cr:  policy(),
				err: awsclient.Wrap(errBoom, errCreate),
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
				iam: &fake.MockPolicyClient{
					MockListPolicyVersions: func(ctx context.Context, input *awsiam.ListPolicyVersionsInput, opts []func(*awsiam.Options)) (*awsiam.ListPolicyVersionsOutput, error) {
						return &awsiam.ListPolicyVersionsOutput{}, nil
					},
					MockDeletePolicyVersion: func(ctx context.Context, input *awsiam.DeletePolicyVersionInput, opts []func(*awsiam.Options)) (*awsiam.DeletePolicyVersionOutput, error) {
						return &awsiam.DeletePolicyVersionOutput{}, nil
					},
					MockCreatePolicyVersion: func(ctx context.Context, input *awsiam.CreatePolicyVersionInput, opts []func(*awsiam.Options)) (*awsiam.CreatePolicyVersionOutput, error) {
						return &awsiam.CreatePolicyVersionOutput{}, nil
					},
				},
				cr: policy(withExterName(arn)),
			},
			want: want{
				cr: policy(withExterName(arn)),
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
		"ListVersionsError": {
			args: args{
				iam: &fake.MockPolicyClient{
					MockListPolicyVersions: func(ctx context.Context, input *awsiam.ListPolicyVersionsInput, opts []func(*awsiam.Options)) (*awsiam.ListPolicyVersionsOutput, error) {
						return nil, errBoom
					},
				},
				cr: policy(withExterName(arn)),
			},
			want: want{
				cr:  policy(withExterName(arn)),
				err: awsclient.Wrap(errBoom, errUpdate),
			},
		},
		"CreateVersionError": {
			args: args{
				iam: &fake.MockPolicyClient{
					MockListPolicyVersions: func(ctx context.Context, input *awsiam.ListPolicyVersionsInput, opts []func(*awsiam.Options)) (*awsiam.ListPolicyVersionsOutput, error) {
						return &awsiam.ListPolicyVersionsOutput{}, nil
					},
					MockDeletePolicyVersion: func(ctx context.Context, input *awsiam.DeletePolicyVersionInput, opts []func(*awsiam.Options)) (*awsiam.DeletePolicyVersionOutput, error) {
						return &awsiam.DeletePolicyVersionOutput{}, nil
					},
					MockCreatePolicyVersion: func(ctx context.Context, input *awsiam.CreatePolicyVersionInput, opts []func(*awsiam.Options)) (*awsiam.CreatePolicyVersionOutput, error) {
						return nil, errBoom
					},
				},
				cr: policy(withExterName(arn)),
			},
			want: want{
				cr:  policy(withExterName(arn)),
				err: awsclient.Wrap(errBoom, errUpdate),
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
				iam: &fake.MockPolicyClient{
					MockListPolicyVersions: func(ctx context.Context, input *awsiam.ListPolicyVersionsInput, opts []func(*awsiam.Options)) (*awsiam.ListPolicyVersionsOutput, error) {
						return &awsiam.ListPolicyVersionsOutput{}, nil
					},
					MockDeletePolicyVersion: func(ctx context.Context, input *awsiam.DeletePolicyVersionInput, opts []func(*awsiam.Options)) (*awsiam.DeletePolicyVersionOutput, error) {
						return &awsiam.DeletePolicyVersionOutput{}, nil
					},
					MockDeletePolicy: func(ctx context.Context, input *awsiam.DeletePolicyInput, opts []func(*awsiam.Options)) (*awsiam.DeletePolicyOutput, error) {
						return &awsiam.DeletePolicyOutput{}, nil
					},
				},
				cr: policy(withExterName(arn)),
			},
			want: want{
				cr: policy(withExterName(arn),
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
		"DeleteVersionError": {
			args: args{
				iam: &fake.MockPolicyClient{
					MockListPolicyVersions: func(ctx context.Context, input *awsiam.ListPolicyVersionsInput, opts []func(*awsiam.Options)) (*awsiam.ListPolicyVersionsOutput, error) {
						return &awsiam.ListPolicyVersionsOutput{
							Versions: []awsiamtypes.PolicyVersion{
								{
									IsDefaultVersion: boolFalse,
								},
							},
						}, nil
					},
					MockDeletePolicyVersion: func(ctx context.Context, input *awsiam.DeletePolicyVersionInput, opts []func(*awsiam.Options)) (*awsiam.DeletePolicyVersionOutput, error) {
						return nil, errBoom
					},
				},
				cr: policy(withExterName(arn)),
			},
			want: want{
				cr:  policy(withExterName(arn)),
				err: awsclient.Wrap(errBoom, errDelete),
			},
		},
		"DeletePolicyError": {
			args: args{
				iam: &fake.MockPolicyClient{
					MockListPolicyVersions: func(ctx context.Context, input *awsiam.ListPolicyVersionsInput, opts []func(*awsiam.Options)) (*awsiam.ListPolicyVersionsOutput, error) {
						return &awsiam.ListPolicyVersionsOutput{}, nil
					},
					MockDeletePolicyVersion: func(ctx context.Context, input *awsiam.DeletePolicyVersionInput, opts []func(*awsiam.Options)) (*awsiam.DeletePolicyVersionOutput, error) {
						return &awsiam.DeletePolicyVersionOutput{}, nil
					},
					MockDeletePolicy: func(ctx context.Context, input *awsiam.DeletePolicyInput, opts []func(*awsiam.Options)) (*awsiam.DeletePolicyOutput, error) {
						return nil, errBoom
					},
				},
				cr: policy(withExterName(arn)),
			},
			want: want{
				cr: policy(withExterName(arn),
					withConditions(xpv1.Deleting())),
				err: awsclient.Wrap(errBoom, errDelete),
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
