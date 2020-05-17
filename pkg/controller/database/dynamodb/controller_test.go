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
package dynamodb

import (
	"context"
	"net/http"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsdynamo "github.com/aws/aws-sdk-go-v2/service/dynamodb"
	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane/provider-aws/apis/database/v1alpha1"
	awsv1alpha3 "github.com/crossplane/provider-aws/apis/v1alpha3"
	awsclients "github.com/crossplane/provider-aws/pkg/clients"
	"github.com/crossplane/provider-aws/pkg/clients/dynamodb"
	"github.com/crossplane/provider-aws/pkg/clients/dynamodb/fake"
)

const (
	providerName    = "aws-creds"
	secretNamespace = "crossplane-system"
	testRegion      = "us-east-1"

	connectionSecretName = "my-little-secret"
	secretKey            = "credentials"
	credData             = "confidential!"
)

var (
	errBoom = errors.New("boom")
)

type args struct {
	dynamo dynamodb.Client
	// kube   client.Client
	cr *v1alpha1.DynamoTable
}

type tableModifier func(*v1alpha1.DynamoTable)

func withConditions(c ...runtimev1alpha1.Condition) tableModifier {
	return func(r *v1alpha1.DynamoTable) { r.Status.ConditionedStatus.Conditions = c }
}

func withBindingPhase(p runtimev1alpha1.BindingPhase) tableModifier {
	return func(r *v1alpha1.DynamoTable) { r.Status.SetBindingPhase(p) }
}

func withStatus(s v1alpha1.DynamoTableObservation) tableModifier {
	return func(r *v1alpha1.DynamoTable) { r.Status.AtProvider = s }
}

func table(m ...tableModifier) *v1alpha1.DynamoTable {
	cr := &v1alpha1.DynamoTable{
		Spec: v1alpha1.DynamoTableSpec{
			ResourceSpec: runtimev1alpha1.ResourceSpec{
				ProviderReference: &corev1.ObjectReference{Name: providerName},
			},
		},
	}
	for _, f := range m {
		f(cr)
	}
	return cr
}

var _ managed.ExternalClient = &external{}
var _ managed.ExternalConnecter = &connector{}

func TestConnect(t *testing.T) {
	secret := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      connectionSecretName,
			Namespace: secretNamespace,
		},
		Data: map[string][]byte{
			secretKey: []byte(credData),
		},
	}

	providerSA := func(saVal bool) awsv1alpha3.Provider {
		return awsv1alpha3.Provider{
			Spec: awsv1alpha3.ProviderSpec{
				Region:            testRegion,
				UseServiceAccount: &saVal,
				ProviderSpec: runtimev1alpha1.ProviderSpec{
					CredentialsSecretRef: &runtimev1alpha1.SecretKeySelector{
						SecretReference: runtimev1alpha1.SecretReference{
							Namespace: secretNamespace,
							Name:      connectionSecretName,
						},
						Key: secretKey,
					},
				},
			},
		}
	}
	type args struct {
		kube        client.Client
		newClientFn func(ctx context.Context, credentials []byte, region string, auth awsclients.AuthMethod) (dynamodb.Client, error)
		cr          *v1alpha1.DynamoTable
	}
	type want struct {
		err error
	}

	cases := map[string]struct {
		args
		want
	}{
		"Successful": {
			args: args{
				kube: &test.MockClient{
					MockGet: func(_ context.Context, key client.ObjectKey, obj runtime.Object) error {
						switch key {
						case client.ObjectKey{Name: providerName}:
							p := providerSA(false)
							p.DeepCopyInto(obj.(*awsv1alpha3.Provider))
							return nil
						case client.ObjectKey{Namespace: secretNamespace, Name: connectionSecretName}:
							secret.DeepCopyInto(obj.(*corev1.Secret))
							return nil
						}
						return errBoom
					},
				},
				newClientFn: func(_ context.Context, credentials []byte, region string, _ awsclients.AuthMethod) (i dynamodb.Client, e error) {
					if diff := cmp.Diff(credData, string(credentials)); diff != "" {
						t.Errorf("r: -want, +got:\n%s", diff)
					}
					if diff := cmp.Diff(testRegion, region); diff != "" {
						t.Errorf("r: -want, +got:\n%s", diff)
					}
					return nil, nil
				},
				cr: table(),
			},
		},
		"SuccessfulUseServiceAccount": {
			args: args{
				kube: &test.MockClient{
					MockGet: func(_ context.Context, key client.ObjectKey, obj runtime.Object) error {
						if key == (client.ObjectKey{Name: providerName}) {
							p := providerSA(true)
							p.DeepCopyInto(obj.(*awsv1alpha3.Provider))
							return nil
						}
						return errBoom
					},
				},
				newClientFn: func(_ context.Context, credentials []byte, region string, _ awsclients.AuthMethod) (i dynamodb.Client, e error) {
					if diff := cmp.Diff("", string(credentials)); diff != "" {
						t.Errorf("r: -want, +got:\n%s", diff)
					}
					if diff := cmp.Diff(testRegion, region); diff != "" {
						t.Errorf("r: -want, +got:\n%s", diff)
					}
					return nil, nil
				},
				cr: table(),
			},
		},
		"ProviderGetFailed": {
			args: args{
				kube: &test.MockClient{
					MockGet: func(_ context.Context, key client.ObjectKey, obj runtime.Object) error {
						return errBoom
					},
				},
				cr: table(),
			},
			want: want{
				err: errors.Wrap(errBoom, errGetProvider),
			},
		},
		"SecretGetFailed": {
			args: args{
				kube: &test.MockClient{
					MockGet: func(_ context.Context, key client.ObjectKey, obj runtime.Object) error {
						switch key {
						case client.ObjectKey{Name: providerName}:
							p := providerSA(false)
							p.DeepCopyInto(obj.(*awsv1alpha3.Provider))
							return nil
						case client.ObjectKey{Namespace: secretNamespace, Name: connectionSecretName}:
							return errBoom
						default:
							return nil
						}
					},
				},
				cr: table(),
			},
			want: want{
				err: errors.Wrap(errBoom, errGetProviderSecret),
			},
		},
		"SecretGetFailedNil": {
			args: args{
				kube: &test.MockClient{
					MockGet: func(_ context.Context, key client.ObjectKey, obj runtime.Object) error {
						switch key {
						case client.ObjectKey{Name: providerName}:
							p := providerSA(false)
							p.SetCredentialsSecretReference(nil)
							p.DeepCopyInto(obj.(*awsv1alpha3.Provider))
							return nil
						case client.ObjectKey{Namespace: secretNamespace, Name: connectionSecretName}:
							return errBoom
						default:
							return nil
						}
					},
				},
				cr: table(),
			},
			want: want{
				err: errors.New(errGetProviderSecret),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			c := &connector{kube: tc.kube, newClientFn: tc.newClientFn}
			_, err := c.Connect(context.Background(), tc.args.cr)
			if diff := cmp.Diff(tc.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestObserve(t *testing.T) {
	type want struct {
		cr     *v1alpha1.DynamoTable
		result managed.ExternalObservation
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"SuccessfulAvailable": {

			args: args{
				dynamo: &fake.MockDynamoClient{
					MockDescribe: func(input *awsdynamo.DescribeTableInput) awsdynamo.DescribeTableRequest {
						return awsdynamo.DescribeTableRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Data: &awsdynamo.DescribeTableOutput{
								Table: &awsdynamo.TableDescription{
									TableStatus: v1alpha1.DynamoTableStateAvailable,
								},
							},
							},
						}
					},
				},
				cr: table(),
			},
			want: want{
				cr: table(
					withStatus(v1alpha1.DynamoTableObservation{
						TableStatus: v1alpha1.DynamoTableStateAvailable,
					}),
					withConditions(runtimev1alpha1.Available()),
					withConditions(runtimev1alpha1.Available()),
					withBindingPhase(runtimev1alpha1.BindingPhaseUnbound)),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
				},
			},
		},
		"DeletingState": {
			args: args{
				dynamo: &fake.MockDynamoClient{
					MockDescribe: func(input *awsdynamo.DescribeTableInput) awsdynamo.DescribeTableRequest {
						return awsdynamo.DescribeTableRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Data: &awsdynamo.DescribeTableOutput{
								Table: &awsdynamo.TableDescription{
									TableStatus: v1alpha1.DynamoTableStateDeleting,
								},
							},
							},
						}
					},
				},
				cr: table(),
			},
			want: want{
				cr: table(
					withConditions(runtimev1alpha1.Deleting()),
					withStatus(v1alpha1.DynamoTableObservation{
						TableStatus: v1alpha1.DynamoTableStateDeleting,
					})),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
				},
			},
		},
		"FailedDescribeRequest": {
			args: args{
				dynamo: &fake.MockDynamoClient{
					MockDescribe: func(input *awsdynamo.DescribeTableInput) awsdynamo.DescribeTableRequest {
						return awsdynamo.DescribeTableRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: errBoom},
						}
					},
				},
				cr: table(),
			},
			want: want{
				cr:  table(),
				err: errors.Wrap(errBoom, errDescribeFailed),
			},
		},
		"NotFound": {
			args: args{
				dynamo: &fake.MockDynamoClient{
					MockDescribe: func(input *awsdynamo.DescribeTableInput) awsdynamo.DescribeTableRequest {
						return awsdynamo.DescribeTableRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: errors.New(awsdynamo.ErrCodeResourceNotFoundException)},
						}
					},
				},
				cr: table(),
			},
			want: want{
				cr: table(),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.dynamo}
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
		cr     *v1alpha1.DynamoTable
		result managed.ExternalCreation
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"Successful": {
			args: args{
				dynamo: &fake.MockDynamoClient{
					MockCreate: func(input *awsdynamo.CreateTableInput) awsdynamo.CreateTableRequest {
						return awsdynamo.CreateTableRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Data: &awsdynamo.CreateTableOutput{}},
						}
					},
				},
				cr: table(),
			},
			want: want{
				cr: table(withConditions(runtimev1alpha1.Creating())),
			},
		},
		"SuccessfulNoNeedForCreate": {
			args: args{
				cr: table(withStatus(v1alpha1.DynamoTableObservation{
					TableStatus: v1alpha1.DynamoTableStateCreating,
				})),
			},
			want: want{
				cr: table(withStatus(v1alpha1.DynamoTableObservation{
					TableStatus: v1alpha1.DynamoTableStateCreating,
				}), withConditions(runtimev1alpha1.Creating())),
			},
		},
		"FailedRequest": {
			args: args{
				dynamo: &fake.MockDynamoClient{
					MockCreate: func(input *awsdynamo.CreateTableInput) awsdynamo.CreateTableRequest {
						return awsdynamo.CreateTableRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: errBoom},
						}
					},
				},
				cr: table(),
			},
			want: want{
				cr:  table(withConditions(runtimev1alpha1.Creating())),
				err: errors.Wrap(errBoom, errCreateFailed),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.dynamo}
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
		cr     *v1alpha1.DynamoTable
		result managed.ExternalUpdate
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"Successful": {
			args: args{
				dynamo: &fake.MockDynamoClient{
					MockUpdate: func(input *awsdynamo.UpdateTableInput) awsdynamo.UpdateTableRequest {
						return awsdynamo.UpdateTableRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Data: &awsdynamo.UpdateTableOutput{}},
						}
					},
				},
				cr: table(),
			},
			want: want{
				cr: table(),
			},
		},
		"AlreadyModifying": {
			args: args{
				cr: table(withStatus(v1alpha1.DynamoTableObservation{
					TableStatus: v1alpha1.DynamoTableStateModifying,
				})),
			},
			want: want{
				cr: table(withStatus(v1alpha1.DynamoTableObservation{
					TableStatus: v1alpha1.DynamoTableStateModifying,
				})),
			},
		},
		"FailedModify": {
			args: args{
				dynamo: &fake.MockDynamoClient{
					MockUpdate: func(input *awsdynamo.UpdateTableInput) awsdynamo.UpdateTableRequest {
						return awsdynamo.UpdateTableRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: errBoom},
						}
					},
				},
				cr: table(),
			},
			want: want{
				cr:  table(),
				err: errors.Wrap(errBoom, errUpdateFailed),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.dynamo}
			u, err := e.Update(context.Background(), tc.args.cr)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.cr, tc.args.cr, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.result, u); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	type want struct {
		cr  *v1alpha1.DynamoTable
		err error
	}

	cases := map[string]struct {
		args
		want
	}{
		"Successful": {
			args: args{
				dynamo: &fake.MockDynamoClient{
					MockDelete: func(input *awsdynamo.DeleteTableInput) awsdynamo.DeleteTableRequest {
						return awsdynamo.DeleteTableRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Data: &awsdynamo.DeleteTableOutput{}},
						}
					},
				},
				cr: table(),
			},
			want: want{
				cr: table(withConditions(runtimev1alpha1.Deleting())),
			},
		},
		"AlreadyDeleting": {
			args: args{
				cr: table(withStatus(v1alpha1.DynamoTableObservation{
					TableStatus: v1alpha1.DynamoTableStateDeleting,
				})),
			},
			want: want{
				cr: table(withStatus(v1alpha1.DynamoTableObservation{
					TableStatus: v1alpha1.DynamoTableStateDeleting,
				}),
					withConditions(runtimev1alpha1.Deleting())),
			},
		},
		"AlreadyDeleted": {
			args: args{
				dynamo: &fake.MockDynamoClient{
					MockDelete: func(input *awsdynamo.DeleteTableInput) awsdynamo.DeleteTableRequest {
						return awsdynamo.DeleteTableRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: errors.New(awsdynamo.ErrCodeResourceNotFoundException)},
						}
					},
				},
				cr: table(),
			},
			want: want{
				cr: table(withConditions(runtimev1alpha1.Deleting())),
			},
		},
		"Failed": {
			args: args{
				dynamo: &fake.MockDynamoClient{
					MockDelete: func(input *awsdynamo.DeleteTableInput) awsdynamo.DeleteTableRequest {
						return awsdynamo.DeleteTableRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: errBoom},
						}
					},
				},
				cr: table(),
			},
			want: want{
				cr:  table(withConditions(runtimev1alpha1.Deleting())),
				err: errors.Wrap(errBoom, errDeleteFailed),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.dynamo}
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
