// /*
// Copyright 2019 The Crossplane Authors.

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

//     http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
// */

package secret

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	svcsdk "github.com/aws/aws-sdk-go/service/secretsmanager"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/test"

	svcapitypes "github.com/crossplane/provider-aws/apis/secretsmanager/v1alpha1"
	awsclient "github.com/crossplane/provider-aws/pkg/clients"
	fake "github.com/crossplane/provider-aws/pkg/clients/secretsmanager/fake"
)

var (
	unexpectedItem resource.Managed
)

const (
	testExternalName    = "test-external-name"
	testSecretString    = "some-secret-value"
	testSecretName      = "test-secret"
	testSecretNamespace = "test-secret-namespace"
	testSecretKey       = "test-secret-key"
)

type mockClientModifier func(*fake.MockSecretsManagerAPI)

type args struct {
	kube client.Client
	mock mockClientModifier
	cr   resource.Managed
}

type secretModifier func(*svcapitypes.Secret)

func withExternalName(s string) secretModifier {
	return func(r *svcapitypes.Secret) { meta.SetExternalName(r, s) }
}

func withConditions(c ...xpv1.Condition) secretModifier {
	return func(r *svcapitypes.Secret) { r.Status.ConditionedStatus.Conditions = c }
}

func withStringSecretRef(value *svcapitypes.SecretReference) secretModifier {
	return func(r *svcapitypes.Secret) { r.Spec.ForProvider.StringSecretRef = value }
}

func instance(m ...secretModifier) *svcapitypes.Secret {
	cr := &svcapitypes.Secret{}
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
				kube: &test.MockClient{
					MockStatusUpdate: test.NewMockStatusUpdateFn(nil),
					MockGet: test.NewMockGetFn(nil, func(obj client.Object) error {
						s := obj.(*corev1.Secret)
						s.Data = map[string][]byte{
							testSecretKey: []byte(testSecretString),
						}
						return nil
					}),
				},
				mock: func(msma *fake.MockSecretsManagerAPI) {
					msma.EXPECT().
						DescribeSecretWithContext(
							gomock.Eq(context.Background()),
							gomock.Eq(
								&svcsdk.DescribeSecretInput{
									SecretId: awsclient.String(testExternalName),
								},
							),
						).
						Return(&svcsdk.DescribeSecretOutput{}, nil).
						Times(1)
					msma.EXPECT().
						GetSecretValueWithContext(
							gomock.Eq(context.Background()),
							gomock.Eq(
								&svcsdk.GetSecretValueInput{
									SecretId: awsclient.String(testExternalName),
								},
							),
						).
						Return(&svcsdk.GetSecretValueOutput{
							SecretString: awsclient.String(testSecretString),
						}, nil).
						Times(1)
				},
				cr: instance(
					withExternalName(testExternalName),
					withStringSecretRef(&svcapitypes.SecretReference{
						Name:      testSecretName,
						Namespace: testSecretNamespace,
						Key:       awsclient.String(testSecretKey),
					}),
				),
			},
			want: want{
				cr: instance(
					withExternalName(testExternalName),
					withConditions(xpv1.Available()),
					withStringSecretRef(&svcapitypes.SecretReference{
						Name:      testSecretName,
						Namespace: testSecretNamespace,
						Key:       awsclient.String(testSecretKey),
					}),
				),
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
		"ResourceDoesNotExist": {
			args: args{
				cr: instance(),
			},
			want: want{
				cr:     instance(),
				result: managed.ExternalObservation{},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			ctlr := gomock.NewController(t)
			client := fake.NewMockSecretsManagerAPI(ctlr)
			if tc.args.mock != nil {
				tc.args.mock(client)
			}

			opts := []option{setupExternal}
			e := newExternal(tc.args.kube, client, opts)
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
