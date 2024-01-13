package ecr

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"

	"github.com/crossplane-contrib/provider-aws/apis/ecr/v1beta1"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
)

var (
	// an arbitrary managed resource
	effect      = "Allow"
	statementID = pointer.ToOrNilIfZeroValue("1")
	boolCheck   = true
	testID      = "id"
	policy      = `{"Statement":[{"Action":"ecr:ListImages","Effect":"Allow","Principal":"*"}],"Version":"2012-10-17"}`
	params      = v1beta1.RepositoryPolicyParameters{
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
)

type statementModifier func(statement *v1beta1.RepositoryPolicyStatement)

func withPrincipal(s *v1beta1.RepositoryPrincipal) statementModifier {
	return func(statement *v1beta1.RepositoryPolicyStatement) {
		statement.Principal = s
	}
}

func withPolicyAction(s []string) statementModifier {
	return func(statement *v1beta1.RepositoryPolicyStatement) {
		statement.Action = s
	}
}

func withConditionBlock(m []v1beta1.Condition) statementModifier {
	return func(statement *v1beta1.RepositoryPolicyStatement) {
		statement.Condition = m
	}
}

func policyStatement(m ...statementModifier) *v1beta1.RepositoryPolicyStatement {
	cr := &v1beta1.RepositoryPolicyStatement{
		SID:    statementID,
		Effect: effect,
	}
	for _, f := range m {
		f(cr)
	}
	return cr
}

type repositoryPolicyModifier func(policy *v1beta1.RepositoryPolicy)

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

func TestSerializeRepositoryPolicyStatement(t *testing.T) {
	cases := map[string]struct {
		in  v1beta1.RepositoryPolicyStatement
		out string
		err error
	}{
		"BasicInput": {
			in:  *policyStatement(),
			out: `{"Effect":"Allow","Sid":"1"}`,
		},
		"ValidInput": {
			in: *policyStatement(
				withPrincipal(&v1beta1.RepositoryPrincipal{
					AllowAnon: &boolCheck,
				}),
				withPolicyAction([]string{"ecr:DescribeRepositories"}),
			),
			out: `{"Action":"ecr:DescribeRepositories","Effect":"Allow","Principal":"*","Sid":"1"}`,
		},
		"ComplexInput": {
			in: *policyStatement(
				withPrincipal(&v1beta1.RepositoryPrincipal{
					AWSPrincipals: []v1beta1.AWSPrincipal{
						{
							UserARN: pointer.ToOrNilIfZeroValue("arn:aws:iam::111122223333:userARN"),
						},
						{
							// Note: should be converted to full ARN when serialized
							// to avoid needless updates.
							AWSAccountID: pointer.ToOrNilIfZeroValue("111122223334"),
						},
						{
							IAMRoleARN: pointer.ToOrNilIfZeroValue("arn:aws:iam::111122223333:roleARN"),
						},
					},
				}),
				withPolicyAction([]string{"ecr:DescribeRepositories"}),
				withConditionBlock([]v1beta1.Condition{
					{
						OperatorKey: "test",
						Conditions: []v1beta1.ConditionPair{
							{
								ConditionKey:         "test",
								ConditionStringValue: pointer.ToOrNilIfZeroValue("testKey"),
							},
							{
								ConditionKey:         "test2",
								ConditionStringValue: pointer.ToOrNilIfZeroValue("testKey2"),
							},
						},
					},
				}),
			),
			out: `{"Condition":{"test":{"test":"testKey","test2":"testKey2"}},"Action":"ecr:DescribeRepositories","Effect":"Allow","Principal":{"AWS":["arn:aws:iam::111122223333:userARN","arn:aws:iam::111122223334:root","arn:aws:iam::111122223333:roleARN"]},"Sid":"1"}`,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			js, err := SerializeRepositoryPolicyStatement(tc.in)

			if diff := cmp.Diff(tc.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
				return
			}

			var out interface{}
			err = json.Unmarshal([]byte(tc.out), &out)
			if diff := cmp.Diff(tc.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
				return
			}

			if diff := cmp.Diff(js, out); diff != "" {
				t.Errorf("SerializeRepositoryPolicyStatement(...): -want, +got\n:%s", diff)
			}
		})
	}
}

func TestLateInitializePolicy(t *testing.T) {
	cases := map[string]struct {
		parameters   *v1beta1.RepositoryPolicyParameters
		policyOutput *ecr.GetRepositoryPolicyOutput
		want         *v1beta1.RepositoryPolicyParameters
	}{
		"AllOptionalFields": {
			parameters: &v1beta1.RepositoryPolicyParameters{},
			policyOutput: &ecr.GetRepositoryPolicyOutput{
				RegistryId: &testID,
			},
			want: &v1beta1.RepositoryPolicyParameters{
				RegistryID: &testID,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			LateInitializeRepositoryPolicy(tc.parameters, tc.policyOutput)
			if diff := cmp.Diff(tc.want, tc.parameters); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestFormat(t *testing.T) {
	type formatarg struct {
		cr *v1beta1.RepositoryPolicy
	}
	type want struct {
		str string
		err error
	}

	cases := map[string]struct {
		args formatarg
		want
	}{
		"ValidInput": {
			args: formatarg{
				cr: repositoryPolicy(withPolicy(&params)),
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
				cr: repositoryPolicy(withPolicy(&v1beta1.RepositoryPolicyParameters{
					RawPolicy: &policy,
				})),
			},
			want: want{
				str: policy,
			},
		},
		"NoPolicy": {
			args: formatarg{
				cr: repositoryPolicy(withPolicy(&v1beta1.RepositoryPolicyParameters{})),
			},
			want: want{
				err: errors.New(errNotSpecified),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			str, err := RawPolicyData(tc.args.cr)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.str, str); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}
