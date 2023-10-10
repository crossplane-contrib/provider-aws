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

package secret

import (
	"context"
	_ "embed"
	"encoding/json"
	"testing"

	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/aws/aws-sdk-go/service/secretsmanager/secretsmanageriface"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane-contrib/provider-aws/apis/secretsmanager/v1beta1"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/secretsmanager/fake"
)

var (
	//go:embed testdata/issue1804_policy.json
	issue1804Policy string
	//go:embed testdata/issue1804_policy.json
	issue1804PolicyCompact string
)

type args struct {
	secretsmanager secretsmanageriface.SecretsManagerAPI
	kube           client.Client
	cr             *v1beta1.Secret
}

type secretModifier func(*v1beta1.Secret)

func withConditions(c ...xpv1.Condition) secretModifier {
	return func(r *v1beta1.Secret) { r.Status.ConditionedStatus.Conditions = c }
}

func withSpec(s v1beta1.SecretParameters) secretModifier {
	return func(r *v1beta1.Secret) { r.Spec.ForProvider = s }
}

func withExternalName(n string) secretModifier {
	return func(s *v1beta1.Secret) { meta.SetExternalName(s, n) }
}

func secret(m ...secretModifier) *v1beta1.Secret {
	cr := &v1beta1.Secret{}
	for _, f := range m {
		f(cr)
	}
	return cr
}

func TestObserve(t *testing.T) {
	testPayload := map[string]string{
		"payload": "123456",
	}
	testPayloadJSONBytes, _ := json.Marshal(testPayload) //nolint:errchkjson
	testPayloadJSON := string(testPayloadJSONBytes)

	type want struct {
		cr     *v1beta1.Secret
		result managed.ExternalObservation
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"Issue1804": {
			args: args{
				secretsmanager: &fake.MockSecretsManagerClient{
					MockDescribeSecretWithContext: func(dsi *secretsmanager.DescribeSecretInput) (*secretsmanager.DescribeSecretOutput, error) {
						return &secretsmanager.DescribeSecretOutput{}, nil
					},
					MockGetResourcePolicyWithContext: func(grpi *secretsmanager.GetResourcePolicyInput) (*secretsmanager.GetResourcePolicyOutput, error) {
						return &secretsmanager.GetResourcePolicyOutput{
							ResourcePolicy: &issue1804PolicyCompact,
						}, nil
					},
					MockGetSecretValueWithContext: func(gsvi *secretsmanager.GetSecretValueInput) (*secretsmanager.GetSecretValueOutput, error) {
						return &secretsmanager.GetSecretValueOutput{
							SecretString: &testPayloadJSON,
						}, nil
					},
				},
				kube: &test.MockClient{
					MockGet: func(ctx context.Context, key client.ObjectKey, obj client.Object) error {
						sec := obj.(*corev1.Secret)
						sec.Data = map[string][]byte{}
						for k, v := range testPayload {
							sec.Data[k] = []byte(v)
						}
						return nil
					},
				},
				cr: secret(
					withExternalName("test"),
					withSpec(v1beta1.SecretParameters{
						CustomSecretParameters: v1beta1.CustomSecretParameters{
							ResourcePolicy: &issue1804Policy,
							StringSecretRef: &v1beta1.SecretReference{
								Name:      "test-secret",
								Namespace: "test",
							},
						},
					}),
				),
			},
			want: want{
				cr: secret(
					withExternalName("test"),
					withSpec(v1beta1.SecretParameters{
						CustomSecretParameters: v1beta1.CustomSecretParameters{
							ResourcePolicy: &issue1804Policy,
							StringSecretRef: &v1beta1.SecretReference{
								Name:      "test-secret",
								Namespace: "test",
							},
						},
					}),
					withConditions(xpv1.Available()),
				),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := newExternal(tc.kube, tc.secretsmanager, []option{setupExternal})
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
