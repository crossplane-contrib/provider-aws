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
	"net/http"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsiam "github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	corev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/test"

	"github.com/crossplane/provider-aws/apis/identity/v1alpha1"
	awsv1alpha3 "github.com/crossplane/provider-aws/apis/v1alpha3"
	awsclients "github.com/crossplane/provider-aws/pkg/clients"
	"github.com/crossplane/provider-aws/pkg/clients/iam"
	"github.com/crossplane/provider-aws/pkg/clients/iam/fake"
)

var (
	unexpecedItem resource.Managed
	arn           = "some arn"
	document      = `{
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

const (
	providerName    = "aws-creds"
	secretNamespace = "crossplane-system"
	testRegion      = "us-east-1"

	connectionSecretName = "my-little-secret"
	secretKey            = "credentials"
	credData             = "confidential!"
)

type args struct {
	kube client.Client
	iam  iam.PolicyClient
	cr   resource.Managed
}

type policyModifier func(*v1alpha1.Policy)

func withConditions(c ...corev1alpha1.Condition) policyModifier {
	return func(r *v1alpha1.Policy) { r.Status.ConditionedStatus.Conditions = c }
}

func withSpec(spec v1alpha1.PolicyParameters) policyModifier {
	return func(r *v1alpha1.Policy) {
		r.Spec.ForProvider = spec
	}
}

func withStatus(status v1alpha1.PolicyObservation) policyModifier {
	return func(r *v1alpha1.Policy) {
		r.Status.AtProvider = status
	}
}

func policy(m ...policyModifier) *v1alpha1.Policy {
	cr := &v1alpha1.Policy{
		Spec: v1alpha1.PolicySpec{
			ResourceSpec: corev1alpha1.ResourceSpec{
				ProviderReference: &corev1.ObjectReference{Name: providerName},
			},
		},
	}
	for _, f := range m {
		f(cr)
	}
	return cr
}

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
		newClientFn func(ctx context.Context, credentials []byte, region string, auth awsclients.AuthMethod) (iam.PolicyClient, error)
		cr          *v1alpha1.Policy
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
				newClientFn: func(_ context.Context, credentials []byte, region string, _ awsclients.AuthMethod) (i iam.PolicyClient, e error) {
					if diff := cmp.Diff(credData, string(credentials)); diff != "" {
						t.Errorf("r: -want, +got:\n%s", diff)
					}
					if diff := cmp.Diff(testRegion, region); diff != "" {
						t.Errorf("r: -want, +got:\n%s", diff)
					}
					return nil, nil
				},
				cr: policy(),
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
				newClientFn: func(_ context.Context, credentials []byte, region string, _ awsclients.AuthMethod) (i iam.PolicyClient, e error) {
					if diff := cmp.Diff("", string(credentials)); diff != "" {
						t.Errorf("r: -want, +got:\n%s", diff)
					}
					if diff := cmp.Diff(testRegion, region); diff != "" {
						t.Errorf("r: -want, +got:\n%s", diff)
					}
					return nil, nil
				},
				cr: policy(),
			},
		},
		"ProviderGetFailed": {
			args: args{
				kube: &test.MockClient{
					MockGet: func(_ context.Context, key client.ObjectKey, obj runtime.Object) error {
						return errBoom
					},
				},
				cr: policy(),
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
				cr: policy(),
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
				cr: policy(),
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
					MockGetPolicyRequest: func(input *awsiam.GetPolicyInput) awsiam.GetPolicyRequest {
						return awsiam.GetPolicyRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Data: &awsiam.GetPolicyOutput{
								Policy: &awsiam.Policy{},
							}},
						}
					},
					MockGetPolicyVersionRequest: func(input *awsiam.GetPolicyVersionInput) awsiam.GetPolicyVersionRequest {
						return awsiam.GetPolicyVersionRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Data: &awsiam.GetPolicyVersionOutput{
								PolicyVersion: &awsiam.PolicyVersion{
									Document: &document,
								},
							}},
						}
					},
				},
				cr: policy(withSpec(v1alpha1.PolicyParameters{
					Document: document,
				}), withStatus(v1alpha1.PolicyObservation{
					Arn: arn,
				})),
			},
			want: want{
				cr: policy(withSpec(v1alpha1.PolicyParameters{
					Document: document,
				}),
					withConditions(corev1alpha1.Available())),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
				},
			},
		},
		"InValidInput": {
			args: args{
				cr: unexpecedItem,
			},
			want: want{
				cr:  unexpecedItem,
				err: errors.New(errUnexpectedObject),
			},
		},
		"GetUPolicyError": {
			args: args{
				iam: &fake.MockPolicyClient{
					MockGetPolicyRequest: func(input *awsiam.GetPolicyInput) awsiam.GetPolicyRequest {
						return awsiam.GetPolicyRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: errBoom},
						}
					},
				},
				cr: policy(withStatus(v1alpha1.PolicyObservation{
					Arn: arn,
				})),
			},
			want: want{
				cr: policy(withStatus(v1alpha1.PolicyObservation{
					Arn: arn,
				})),
				err: errors.Wrap(errBoom, errGet),
			},
		},
		"EmptySpecPolicy": {
			args: args{
				iam: &fake.MockPolicyClient{
					MockGetPolicyRequest: func(input *awsiam.GetPolicyInput) awsiam.GetPolicyRequest {
						return awsiam.GetPolicyRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Data: &awsiam.GetPolicyOutput{
								Policy: &awsiam.Policy{},
							}},
						}
					},
					MockGetPolicyVersionRequest: func(input *awsiam.GetPolicyVersionInput) awsiam.GetPolicyVersionRequest {
						return awsiam.GetPolicyVersionRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Data: &awsiam.GetPolicyVersionOutput{
								PolicyVersion: &awsiam.PolicyVersion{
									Document: &document,
								},
							}},
						}
					},
				},
				cr: policy(withStatus(v1alpha1.PolicyObservation{
					Arn: arn,
				})),
			},
			want: want{
				cr: policy(withConditions(corev1alpha1.Available())),
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
					MockCreatePolicyRequest: func(input *awsiam.CreatePolicyInput) awsiam.CreatePolicyRequest {
						return awsiam.CreatePolicyRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Data: &awsiam.CreatePolicyOutput{
								Policy: &awsiam.Policy{
									Arn: &arn,
								},
							}},
						}
					},
				},
				cr: policy(withSpec(v1alpha1.PolicyParameters{
					Document: document,
				})),
			},
			want: want{
				cr: policy(
					withSpec(v1alpha1.PolicyParameters{
						Document: document,
					}),
					withStatus(v1alpha1.PolicyObservation{
						Arn: arn,
					}),
					withConditions(corev1alpha1.Creating())),
			},
		},
		"InValidInput": {
			args: args{
				cr: unexpecedItem,
			},
			want: want{
				cr:  unexpecedItem,
				err: errors.New(errUnexpectedObject),
			},
		},
		"ClientError": {
			args: args{
				iam: &fake.MockPolicyClient{
					MockCreatePolicyRequest: func(input *awsiam.CreatePolicyInput) awsiam.CreatePolicyRequest {
						return awsiam.CreatePolicyRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: errBoom},
						}
					},
				},
				cr: policy(),
			},
			want: want{
				cr:  policy(withConditions(corev1alpha1.Creating())),
				err: errors.Wrap(errBoom, errCreate),
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
					MockListPolicyVersionsRequest: func(input *awsiam.ListPolicyVersionsInput) awsiam.ListPolicyVersionsRequest {
						return awsiam.ListPolicyVersionsRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Data: &awsiam.ListPolicyVersionsOutput{}},
						}
					},
					MockDeletePolicyVersionRequest: func(input *awsiam.DeletePolicyVersionInput) awsiam.DeletePolicyVersionRequest {
						return awsiam.DeletePolicyVersionRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Data: &awsiam.DeletePolicyVersionOutput{}},
						}
					},
					MockCreatePolicyVersionRequest: func(input *awsiam.CreatePolicyVersionInput) awsiam.CreatePolicyVersionRequest {
						return awsiam.CreatePolicyVersionRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Data: &awsiam.CreatePolicyVersionOutput{}},
						}
					},
				},
				cr: policy(withStatus(v1alpha1.PolicyObservation{
					Arn: arn,
				})),
			},
			want: want{
				cr: policy(withStatus(v1alpha1.PolicyObservation{
					Arn: arn,
				})),
			},
		},
		"InValidInput": {
			args: args{
				cr: unexpecedItem,
			},
			want: want{
				cr:  unexpecedItem,
				err: errors.New(errUnexpectedObject),
			},
		},
		"ListVersionsError": {
			args: args{
				iam: &fake.MockPolicyClient{
					MockListPolicyVersionsRequest: func(input *awsiam.ListPolicyVersionsInput) awsiam.ListPolicyVersionsRequest {
						return awsiam.ListPolicyVersionsRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: errBoom},
						}
					},
				},
				cr: policy(withStatus(v1alpha1.PolicyObservation{
					Arn: arn,
				})),
			},
			want: want{
				cr: policy(withStatus(v1alpha1.PolicyObservation{
					Arn: arn,
				})),
				err: errors.Wrap(errBoom, errUpdate),
			},
		},
		"CreateVersionError": {
			args: args{
				iam: &fake.MockPolicyClient{
					MockListPolicyVersionsRequest: func(input *awsiam.ListPolicyVersionsInput) awsiam.ListPolicyVersionsRequest {
						return awsiam.ListPolicyVersionsRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Data: &awsiam.ListPolicyVersionsOutput{}},
						}
					},
					MockDeletePolicyVersionRequest: func(input *awsiam.DeletePolicyVersionInput) awsiam.DeletePolicyVersionRequest {
						return awsiam.DeletePolicyVersionRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Data: &awsiam.DeletePolicyVersionOutput{}},
						}
					},
					MockCreatePolicyVersionRequest: func(input *awsiam.CreatePolicyVersionInput) awsiam.CreatePolicyVersionRequest {
						return awsiam.CreatePolicyVersionRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: errBoom},
						}
					},
				},
				cr: policy(withStatus(v1alpha1.PolicyObservation{
					Arn: arn,
				})),
			},
			want: want{
				cr: policy(withStatus(v1alpha1.PolicyObservation{
					Arn: arn,
				})),
				err: errors.Wrap(errBoom, errUpdate),
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
					MockListPolicyVersionsRequest: func(input *awsiam.ListPolicyVersionsInput) awsiam.ListPolicyVersionsRequest {
						return awsiam.ListPolicyVersionsRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Data: &awsiam.ListPolicyVersionsOutput{}},
						}
					},
					MockDeletePolicyVersionRequest: func(input *awsiam.DeletePolicyVersionInput) awsiam.DeletePolicyVersionRequest {
						return awsiam.DeletePolicyVersionRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Data: &awsiam.DeletePolicyVersionOutput{}},
						}
					},
					MockDeletePolicyRequest: func(input *awsiam.DeletePolicyInput) awsiam.DeletePolicyRequest {
						return awsiam.DeletePolicyRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Data: &awsiam.DeletePolicyOutput{}},
						}
					},
				},
				cr: policy(withStatus(v1alpha1.PolicyObservation{
					Arn: arn,
				})),
			},
			want: want{
				cr: policy(withStatus(v1alpha1.PolicyObservation{
					Arn: arn,
				}),
					withConditions(corev1alpha1.Deleting())),
			},
		},
		"InValidInput": {
			args: args{
				cr: unexpecedItem,
			},
			want: want{
				cr:  unexpecedItem,
				err: errors.New(errUnexpectedObject),
			},
		},
		"DeleteVersionError": {
			args: args{
				iam: &fake.MockPolicyClient{
					MockListPolicyVersionsRequest: func(input *awsiam.ListPolicyVersionsInput) awsiam.ListPolicyVersionsRequest {
						return awsiam.ListPolicyVersionsRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Data: &awsiam.ListPolicyVersionsOutput{
								Versions: []awsiam.PolicyVersion{
									{
										IsDefaultVersion: &boolFalse,
									},
								},
							}},
						}
					},
					MockDeletePolicyVersionRequest: func(input *awsiam.DeletePolicyVersionInput) awsiam.DeletePolicyVersionRequest {
						return awsiam.DeletePolicyVersionRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: errBoom},
						}
					},
				},
				cr: policy(withStatus(v1alpha1.PolicyObservation{
					Arn: arn,
				})),
			},
			want: want{
				cr: policy(withStatus(v1alpha1.PolicyObservation{
					Arn: arn,
				})),
				err: errors.Wrap(errBoom, errDelete),
			},
		},
		"DeletePolicyError": {
			args: args{
				iam: &fake.MockPolicyClient{
					MockListPolicyVersionsRequest: func(input *awsiam.ListPolicyVersionsInput) awsiam.ListPolicyVersionsRequest {
						return awsiam.ListPolicyVersionsRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Data: &awsiam.ListPolicyVersionsOutput{}},
						}
					},
					MockDeletePolicyVersionRequest: func(input *awsiam.DeletePolicyVersionInput) awsiam.DeletePolicyVersionRequest {
						return awsiam.DeletePolicyVersionRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Data: &awsiam.DeletePolicyVersionOutput{}},
						}
					},
					MockDeletePolicyRequest: func(input *awsiam.DeletePolicyInput) awsiam.DeletePolicyRequest {
						return awsiam.DeletePolicyRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Error: errBoom},
						}
					},
				},
				cr: policy(withStatus(v1alpha1.PolicyObservation{
					Arn: arn,
				})),
			},
			want: want{
				cr: policy(withStatus(v1alpha1.PolicyObservation{
					Arn: arn,
				}), withConditions(corev1alpha1.Deleting())),
				err: errors.Wrap(errBoom, errDelete),
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
