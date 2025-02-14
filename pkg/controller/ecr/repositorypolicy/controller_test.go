/*
Copyright 2021 The Crossplane Authors.

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

package repositorypolicy

import (
	"context"
	"testing"

	awsecr "github.com/aws/aws-sdk-go-v2/service/ecr"
	awsecrtypes "github.com/aws/aws-sdk-go-v2/service/ecr/types"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane-contrib/provider-aws/apis/ecr/v1beta1"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/ecr"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/ecr/fake"
	errorutils "github.com/crossplane-contrib/provider-aws/pkg/utils/errors"
)

var (
	// an arbitrary managed resource
	unexpectedItem   resource.Managed
	repositoryName   = "testRepo"
	policy           = `{"Statement":[{"Action":"ecr:ListImages","Effect":"Allow","Principal":"*"}],"Version":"2012-10-17"}`
	needUpdatePolicy = `{"Statement":[{"Action":"ecr:ListImages","Effect":"Allow","Principal":"blah"}],"Version":"2012-10-17"}`
	boolCheck        = true

	params = v1beta1.RepositoryPolicyParameters{
		Policy: &v1beta1.RepositoryPolicyBody{
			Version: "2012-10-17",
			Statements: []v1beta1.RepositoryPolicyStatement{
				{
					Effect: "Allow",
					Principal: &v1beta1.RepositoryPrincipal{
						AllowAnon: &boolCheck,
					},
					Action: []string{"ecr:ListImages"},
				},
			},
		},
	}

	errBoom = errors.New("boom")
)

type args struct {
	ecr  ecr.RepositoryPolicyClient
	kube client.Client
	cr   resource.Managed
}

type repositoryPolicyModifier func(policy *v1beta1.RepositoryPolicy)

func withConditions(c ...xpv1.Condition) repositoryPolicyModifier {
	return func(r *v1beta1.RepositoryPolicy) { r.Status.ConditionedStatus.Conditions = c }
}

func withPolicy(s *v1beta1.RepositoryPolicyParameters) repositoryPolicyModifier {
	return func(r *v1beta1.RepositoryPolicy) { r.Spec.ForProvider = *s }
}

func repositoryPolicy(m ...repositoryPolicyModifier) *v1beta1.RepositoryPolicy {
	cr := &v1beta1.RepositoryPolicy{
		Spec: v1beta1.RepositoryPolicySpec{
			ForProvider: v1beta1.RepositoryPolicyParameters{
				RepositoryName: &repositoryName,
				Policy: &v1beta1.RepositoryPolicyBody{
					Statements: make([]v1beta1.RepositoryPolicyStatement, 0),
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
				ecr: &fake.MockRepositoryPolicyClient{
					MockGet: func(ctx context.Context, input *awsecr.GetRepositoryPolicyInput, opts []func(*awsecr.Options)) (*awsecr.GetRepositoryPolicyOutput, error) {
						return &awsecr.GetRepositoryPolicyOutput{PolicyText: &policy}, nil
					},
				},
				cr: repositoryPolicy(withPolicy(&params)),
			},
			want: want{
				cr: repositoryPolicy(withPolicy(&params),
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
		"NeedUpdateInput": {
			args: args{
				ecr: &fake.MockRepositoryPolicyClient{
					MockGet: func(_ context.Context, _ *awsecr.GetRepositoryPolicyInput, _ []func(*awsecr.Options)) (*awsecr.GetRepositoryPolicyOutput, error) {
						return &awsecr.GetRepositoryPolicyOutput{
							PolicyText: &needUpdatePolicy,
						}, nil
					},
				},
				cr: repositoryPolicy(withPolicy(&params)),
			},
			want: want{
				cr: repositoryPolicy(withPolicy(&params),
					withConditions(xpv1.Available())),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: false,
				},
			},
		},
		"ClientError": {
			args: args{
				ecr: &fake.MockRepositoryPolicyClient{
					MockGet: func(ctx context.Context, input *awsecr.GetRepositoryPolicyInput, opts []func(*awsecr.Options)) (*awsecr.GetRepositoryPolicyOutput, error) {
						return &awsecr.GetRepositoryPolicyOutput{}, errBoom
					},
				},
				cr: repositoryPolicy(withPolicy(&params)),
			},
			want: want{
				cr:  repositoryPolicy(withPolicy(&params)),
				err: errorutils.Wrap(errBoom, errGet),
			},
		},
		"ResourceDoesNotExist": {
			args: args{
				ecr: &fake.MockRepositoryPolicyClient{
					MockGet: func(ctx context.Context, input *awsecr.GetRepositoryPolicyInput, opts []func(*awsecr.Options)) (*awsecr.GetRepositoryPolicyOutput, error) {
						return &awsecr.GetRepositoryPolicyOutput{}, &awsecrtypes.RepositoryPolicyNotFoundException{}
					},
				},
				cr: repositoryPolicy(),
			},
			want: want{
				cr: repositoryPolicy(),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.ecr}
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
		"ValidInput": {
			args: args{
				kube: &test.MockClient{
					MockUpdate:       test.NewMockClient().Update,
					MockStatusUpdate: test.NewMockClient().MockStatusUpdate,
				},
				ecr: &fake.MockRepositoryPolicyClient{
					MockSet: func(ctx context.Context, input *awsecr.SetRepositoryPolicyInput, opts []func(*awsecr.Options)) (*awsecr.SetRepositoryPolicyOutput, error) {
						return &awsecr.SetRepositoryPolicyOutput{}, nil
					},
				},
				cr: repositoryPolicy(withPolicy(&params)),
			},
			want: want{
				cr: repositoryPolicy(
					withPolicy(&params)),
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
				kube: &test.MockClient{
					MockStatusUpdate: test.NewMockClient().MockStatusUpdate,
				},
				ecr: &fake.MockRepositoryPolicyClient{
					MockSet: func(ctx context.Context, input *awsecr.SetRepositoryPolicyInput, opts []func(*awsecr.Options)) (*awsecr.SetRepositoryPolicyOutput, error) {
						return &awsecr.SetRepositoryPolicyOutput{}, errBoom
					},
				},
				cr: repositoryPolicy(withPolicy(&params)),
			},
			want: want{
				cr: repositoryPolicy(
					withPolicy(&params)),
				err: errorutils.Wrap(errBoom, errCreate),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.ecr, kube: tc.kube}
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
		"ValidInput": {
			args: args{
				ecr: &fake.MockRepositoryPolicyClient{
					MockSet: func(ctx context.Context, input *awsecr.SetRepositoryPolicyInput, opts []func(*awsecr.Options)) (*awsecr.SetRepositoryPolicyOutput, error) {
						return &awsecr.SetRepositoryPolicyOutput{}, nil
					},
				},
				cr: repositoryPolicy(withPolicy(&params)),
			},
			want: want{
				cr: repositoryPolicy(withPolicy(&params)),
			},
		},
		"ClientError": {
			args: args{
				ecr: &fake.MockRepositoryPolicyClient{
					MockSet: func(ctx context.Context, input *awsecr.SetRepositoryPolicyInput, opts []func(*awsecr.Options)) (*awsecr.SetRepositoryPolicyOutput, error) {
						return &awsecr.SetRepositoryPolicyOutput{}, errBoom
					},
				},
				cr: repositoryPolicy(withPolicy(&params)),
			},
			want: want{
				cr:  repositoryPolicy(withPolicy(&params)),
				err: errorutils.Wrap(errBoom, errUpdate),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.ecr}
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
		"ValidInput": {
			args: args{
				ecr: &fake.MockRepositoryPolicyClient{
					MockDelete: func(ctx context.Context, input *awsecr.DeleteRepositoryPolicyInput, opts []func(*awsecr.Options)) (*awsecr.DeleteRepositoryPolicyOutput, error) {
						return &awsecr.DeleteRepositoryPolicyOutput{}, nil
					},
				},
				cr: repositoryPolicy(withPolicy(&params)),
			},
			want: want{
				cr: repositoryPolicy(withPolicy(&params)),
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
				ecr: &fake.MockRepositoryPolicyClient{
					MockDelete: func(ctx context.Context, input *awsecr.DeleteRepositoryPolicyInput, opts []func(*awsecr.Options)) (*awsecr.DeleteRepositoryPolicyOutput, error) {
						return &awsecr.DeleteRepositoryPolicyOutput{}, errBoom
					},
				},
				cr: repositoryPolicy(withPolicy(&params)),
			},
			want: want{
				cr:  repositoryPolicy(withPolicy(&params)),
				err: errorutils.Wrap(errBoom, errDelete),
			},
		},
		"ResourceDoesNotExist": {
			args: args{
				ecr: &fake.MockRepositoryPolicyClient{
					MockDelete: func(ctx context.Context, input *awsecr.DeleteRepositoryPolicyInput, opts []func(*awsecr.Options)) (*awsecr.DeleteRepositoryPolicyOutput, error) {
						return &awsecr.DeleteRepositoryPolicyOutput{}, &awsecrtypes.RepositoryPolicyNotFoundException{}
					},
				},
				cr: repositoryPolicy(withPolicy(&params)),
			},
			want: want{
				cr: repositoryPolicy(withPolicy(&params)),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.ecr}
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
