package ecr

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"

	"github.com/crossplane/provider-aws/apis/ecr/v1alpha1"
	aws "github.com/crossplane/provider-aws/pkg/clients"
)

var (
	// an arbitrary managed resource
	effect      = "Allow"
	statementID = aws.String("1")
	boolCheck   = true
	testID      = "id"
	policy      = `{"Statement":[{"Action":"ecr:ListImages","Effect":"Allow","Principal":"*"}],"Version":"2012-10-17"}`
	cpxPolicy   = `{"Statement":[{"Action":"ecr:ListImages","Effect":"Allow","Principal":{"AWS":["arn:aws:iam::111122223333:userARN","111122223334","arn:aws:iam::111122223333:roleARN"]}}],"Version":"2012-10-17"}`
	// Note: different sort order of principals than input above
	cpxRemPolicy = `{"Statement":[{"Action":"ecr:ListImages","Effect":"Allow","Principal":{"AWS":["111122223334","arn:aws:iam::111122223333:userARN","arn:aws:iam::111122223333:roleARN"]}}],"Version":"2012-10-17"}`
	params       = v1alpha1.RepositoryPolicyParameters{
		Policy: &v1alpha1.RepositoryPolicyBody{
			Version: "2012-10-17",
			Statements: []v1alpha1.RepositoryPolicyStatement{
				{
					Effect: "Allow",
					Principal: &v1alpha1.RepositoryPrincipal{
						AllowAnon: &boolCheck,
					},
					Action: []string{"ecr:ListImages"},
				},
			},
		},
	}
)

type statementModifier func(statement *v1alpha1.RepositoryPolicyStatement)

func withPrincipal(s *v1alpha1.RepositoryPrincipal) statementModifier {
	return func(statement *v1alpha1.RepositoryPolicyStatement) {
		statement.Principal = s
	}
}

func withPolicyAction(s []string) statementModifier {
	return func(statement *v1alpha1.RepositoryPolicyStatement) {
		statement.Action = s
	}
}

func withConditionBlock(m []v1alpha1.Condition) statementModifier {
	return func(statement *v1alpha1.RepositoryPolicyStatement) {
		statement.Condition = m
	}
}

func policyStatement(m ...statementModifier) *v1alpha1.RepositoryPolicyStatement {
	cr := &v1alpha1.RepositoryPolicyStatement{
		SID:    statementID,
		Effect: effect,
	}
	for _, f := range m {
		f(cr)
	}
	return cr
}

type repositoryPolicyModifier func(policy *v1alpha1.RepositoryPolicy)

func withPolicy(s *v1alpha1.RepositoryPolicyParameters) repositoryPolicyModifier {
	return func(r *v1alpha1.RepositoryPolicy) { r.Spec.ForProvider = *s }
}

func repositoryPolicy(m ...repositoryPolicyModifier) *v1alpha1.RepositoryPolicy {
	cr := &v1alpha1.RepositoryPolicy{
		Spec: v1alpha1.RepositoryPolicySpec{
			ForProvider: v1alpha1.RepositoryPolicyParameters{
				RepositoryName: &repositoryName,
				Policy: &v1alpha1.RepositoryPolicyBody{
					Statements: make([]v1alpha1.RepositoryPolicyStatement, 0),
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
		in  v1alpha1.RepositoryPolicyStatement
		out string
		err error
	}{
		"BasicInput": {
			in:  *policyStatement(),
			out: `{"Effect":"Allow","Sid":"1"}`,
		},
		"ValidInput": {
			in: *policyStatement(
				withPrincipal(&v1alpha1.RepositoryPrincipal{
					AllowAnon: &boolCheck,
				}),
				withPolicyAction([]string{"ecr:DescribeRepositories"}),
			),
			out: `{"Action":"ecr:DescribeRepositories","Effect":"Allow","Principal":"*","Sid":"1"}`,
		},
		"ComplexInput": {
			in: *policyStatement(
				withPrincipal(&v1alpha1.RepositoryPrincipal{
					AWSPrincipals: []v1alpha1.AWSPrincipal{
						{
							IAMUserARN: aws.String("arn:aws:iam::111122223333:userARN"),
						},
						{
							// Note: should be converted to full ARN when serialized
							// to avoid needless updates.
							AWSAccountID: aws.String("111122223334"),
						},
						{
							IAMRoleARN: aws.String("arn:aws:iam::111122223333:roleARN"),
						},
					},
				}),
				withPolicyAction([]string{"ecr:DescribeRepositories"}),
				withConditionBlock([]v1alpha1.Condition{
					{
						OperatorKey: "test",
						Conditions: []v1alpha1.ConditionPair{
							{
								ConditionKey:         "test",
								ConditionStringValue: aws.String("testKey"),
							},
							{
								ConditionKey:         "test2",
								ConditionStringValue: aws.String("testKey2"),
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
		parameters   *v1alpha1.RepositoryPolicyParameters
		policyOutput *ecr.GetRepositoryPolicyResponse
		want         *v1alpha1.RepositoryPolicyParameters
	}{
		"AllOptionalFields": {
			parameters: &v1alpha1.RepositoryPolicyParameters{},
			policyOutput: &ecr.GetRepositoryPolicyResponse{
				GetRepositoryPolicyOutput: &ecr.GetRepositoryPolicyOutput{
					RegistryId: &testID,
				},
			},
			want: &v1alpha1.RepositoryPolicyParameters{
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

func TestIsRepositoryPolicyUpToDate(t *testing.T) {
	type args struct {
		local  string
		remote string
	}

	cases := map[string]struct {
		args args
		want bool
	}{
		"SameFields": {
			args: args{
				local:  "{\"testone\": \"one\", \"testtwo\": \"two\"}",
				remote: "{\"testtwo\": \"two\", \"testone\": \"one\"}",
			},
			want: true,
		},
		"SameFieldsRealPolicy": {
			args: args{
				local:  policy,
				remote: `{"Statement":[{"Effect":"Allow","Action":"ecr:ListImages","Principal":"*"}],"Version":"2012-10-17"}`,
			},
			want: true,
		},
		"DifferentFields": {
			args: args{
				local:  "{\"testone\": \"one\", \"testtwo\": \"two\"}",
				remote: "{\"testthree\": \"three\", \"testone\": \"one\"}",
			},
			want: false,
		},
		"SameFieldsPrincipalPolicy": {
			args: args{
				local:  cpxPolicy,
				remote: cpxRemPolicy,
			},
			want: true,
		},
		"SameFieldsNumericPrincipals": {
			args: args{
				// This is to test that our slice sorting does not
				// panic with unexpected value types.
				local:  `{"Statement":[{"Effect":"Allow","Action":"ecr:ListImages","Principal":[2,1,"foo","bar"]}],"Version":"2012-10-17"}`,
				remote: `{"Statement":[{"Effect":"Allow","Action":"ecr:ListImages","Principal":[2,1,"bar","foo"]}],"Version":"2012-10-17"}`,
			},
			want: true,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := IsRepositoryPolicyUpToDate(&tc.args.local, &tc.args.remote)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestFormat(t *testing.T) {
	type formatarg struct {
		cr *v1alpha1.RepositoryPolicy
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
				cr: repositoryPolicy(withPolicy(&v1alpha1.RepositoryPolicyParameters{
					RawPolicy: &policy,
				})),
			},
			want: want{
				str: policy,
			},
		},
		"NoPolicy": {
			args: formatarg{
				cr: repositoryPolicy(withPolicy(&v1alpha1.RepositoryPolicyParameters{})),
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
