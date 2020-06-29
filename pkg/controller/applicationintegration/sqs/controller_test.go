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
package sqs

import (
	"context"
	"net/http"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awssqs "github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/test"

	"github.com/crossplane/provider-aws/apis/applicationintegration/v1alpha1"
	awsv1alpha3 "github.com/crossplane/provider-aws/apis/v1alpha3"
	awsclients "github.com/crossplane/provider-aws/pkg/clients"
	"github.com/crossplane/provider-aws/pkg/clients/sqs"
	"github.com/crossplane/provider-aws/pkg/clients/sqs/fake"
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
	attributes = map[string]string{}
	queueURL   = "someURL"
	queueName  = "some-name"

	// replaceMe = "replace-me!"
	errBoom = errors.New("boom")
)

type args struct {
	kube client.Client
	sqs  sqs.Client
	cr   *v1alpha1.Queue
}

type sqsModifier func(*v1alpha1.Queue)

func withExternalName(s string) sqsModifier {
	return func(r *v1alpha1.Queue) { meta.SetExternalName(r, s) }
}

func withConditions(c ...runtimev1alpha1.Condition) sqsModifier {
	return func(r *v1alpha1.Queue) { r.Status.ConditionedStatus.Conditions = c }
}

func withSpec(p v1alpha1.QueueParameters) sqsModifier {
	return func(r *v1alpha1.Queue) { r.Spec.ForProvider = p }
}

func withStatus(o v1alpha1.QueueObservation) sqsModifier {
	return func(r *v1alpha1.Queue) { r.Status.AtProvider = o }
}

func queue(m ...sqsModifier) *v1alpha1.Queue {
	cr := &v1alpha1.Queue{
		Spec: v1alpha1.QueueSpec{
			ResourceSpec: runtimev1alpha1.ResourceSpec{
				ProviderReference: runtimev1alpha1.Reference{Name: providerName},
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
		newClientFn func(ctx context.Context, credentials []byte, region string, auth awsclients.AuthMethod) (sqs.Client, error)
		cr          *v1alpha1.Queue
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
				newClientFn: func(_ context.Context, credentials []byte, region string, _ awsclients.AuthMethod) (i sqs.Client, e error) {
					if diff := cmp.Diff(credData, string(credentials)); diff != "" {
						t.Errorf("r: -want, +got:\n%s", diff)
					}
					if diff := cmp.Diff(testRegion, region); diff != "" {
						t.Errorf("r: -want, +got:\n%s", diff)
					}
					return nil, nil
				},
				cr: queue(),
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
				newClientFn: func(_ context.Context, credentials []byte, region string, _ awsclients.AuthMethod) (i sqs.Client, e error) {
					if diff := cmp.Diff("", string(credentials)); diff != "" {
						t.Errorf("r: -want, +got:\n%s", diff)
					}
					if diff := cmp.Diff(testRegion, region); diff != "" {
						t.Errorf("r: -want, +got:\n%s", diff)
					}
					return nil, nil
				},
				cr: queue(),
			},
		},
		"ProviderGetFailed": {
			args: args{
				kube: &test.MockClient{
					MockGet: func(_ context.Context, key client.ObjectKey, obj runtime.Object) error {
						return errBoom
					},
				},
				cr: queue(),
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
				cr: queue(),
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
				cr: queue(),
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
		cr     *v1alpha1.Queue
		result managed.ExternalObservation
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"SuccessfulAvailable": {
			args: args{
				sqs: &fake.MockSQSClient{
					MockGetQueueAttributesRequest: func(input *awssqs.GetQueueAttributesInput) awssqs.GetQueueAttributesRequest {
						return awssqs.GetQueueAttributesRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awssqs.GetQueueAttributesOutput{
								Attributes: attributes,
							}},
						}
					},
					MockListQueueTagsRequest: func(input *awssqs.ListQueueTagsInput) awssqs.ListQueueTagsRequest {
						return awssqs.ListQueueTagsRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awssqs.ListQueueTagsOutput{
								Tags: attributes,
							}},
						}
					},
					MockGetQueueURLRequest: func(input *awssqs.GetQueueUrlInput) awssqs.GetQueueUrlRequest {
						return awssqs.GetQueueUrlRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awssqs.GetQueueUrlOutput{
								QueueUrl: &queueURL,
							}},
						}
					},
				},
				cr: queue(withExternalName(queueName)),
			},
			want: want{
				cr: queue(withExternalName(queueName),
					withConditions(runtimev1alpha1.Available()),
					withStatus(v1alpha1.QueueObservation{
						URL: queueURL,
					})),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
				},
			},
		},
		"GetAttributesFail": {
			args: args{
				sqs: &fake.MockSQSClient{
					MockGetQueueURLRequest: func(input *awssqs.GetQueueUrlInput) awssqs.GetQueueUrlRequest {
						return awssqs.GetQueueUrlRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awssqs.GetQueueUrlOutput{
								QueueUrl: &queueURL,
							}},
						}
					},
					MockGetQueueAttributesRequest: func(input *awssqs.GetQueueAttributesInput) awssqs.GetQueueAttributesRequest {
						return awssqs.GetQueueAttributesRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Error: errBoom},
						}
					},
				},
				cr: queue(withExternalName(queueName)),
			},
			want: want{
				cr:  queue(withExternalName(queueName)),
				err: errors.Wrap(errBoom, errGetQueueAttributesFailed),
			},
		},
		"ListTagsFail": {
			args: args{
				sqs: &fake.MockSQSClient{
					MockGetQueueURLRequest: func(input *awssqs.GetQueueUrlInput) awssqs.GetQueueUrlRequest {
						return awssqs.GetQueueUrlRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awssqs.GetQueueUrlOutput{
								QueueUrl: &queueURL,
							}},
						}
					},
					MockGetQueueAttributesRequest: func(input *awssqs.GetQueueAttributesInput) awssqs.GetQueueAttributesRequest {
						return awssqs.GetQueueAttributesRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awssqs.GetQueueAttributesOutput{
								Attributes: attributes,
							}},
						}
					},
					MockListQueueTagsRequest: func(input *awssqs.ListQueueTagsInput) awssqs.ListQueueTagsRequest {
						return awssqs.ListQueueTagsRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Error: errBoom},
						}
					},
				},
				cr: queue(withExternalName(queueName)),
			},
			want: want{
				cr:  queue(withExternalName(queueName)),
				err: errors.Wrap(errBoom, errListQueueTagsFailed),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.sqs}
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
		cr     *v1alpha1.Queue
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
					MockUpdate: test.NewMockClient().Update,
				},
				sqs: &fake.MockSQSClient{
					MockCreateQueueRequest: func(input *awssqs.CreateQueueInput) awssqs.CreateQueueRequest {
						return awssqs.CreateQueueRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awssqs.CreateQueueOutput{
								QueueUrl: &queueURL,
							}},
						}
					},
				},
				cr: queue(withExternalName(queueURL)),
			},
			want: want{
				cr: queue(withExternalName(queueURL),
					withConditions(runtimev1alpha1.Creating())),
			},
		},
		"CreateFail": {
			args: args{
				sqs: &fake.MockSQSClient{
					MockCreateQueueRequest: func(input *awssqs.CreateQueueInput) awssqs.CreateQueueRequest {
						return awssqs.CreateQueueRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Error: errBoom},
						}
					},
				},
				cr: queue(withExternalName(queueURL),
					withSpec(v1alpha1.QueueParameters{})),
			},
			want: want{
				cr: queue(withExternalName(queueURL),
					withConditions(runtimev1alpha1.Creating())),
				err: errors.Wrap(errBoom, errCreateFailed),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{kube: tc.kube, client: tc.sqs}
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
		cr     *v1alpha1.Queue
		result managed.ExternalUpdate
		err    error
	}

	cases := map[string]struct {
		args
		want
	}{
		"Successful": {
			args: args{
				sqs: &fake.MockSQSClient{
					MockSetQueueAttributesRequest: func(input *awssqs.SetQueueAttributesInput) awssqs.SetQueueAttributesRequest {
						return awssqs.SetQueueAttributesRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awssqs.SetQueueAttributesOutput{}},
						}
					},
					MockListQueueTagsRequest: func(input *awssqs.ListQueueTagsInput) awssqs.ListQueueTagsRequest {
						return awssqs.ListQueueTagsRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awssqs.ListQueueTagsOutput{}},
						}
					},
				},
				cr: queue(withStatus(v1alpha1.QueueObservation{
					URL: queueURL,
				})),
			},
			want: want{
				cr: queue(withStatus(v1alpha1.QueueObservation{
					URL: queueURL,
				})),
			},
		},
		"TagsUpdate": {
			args: args{
				sqs: &fake.MockSQSClient{
					MockSetQueueAttributesRequest: func(input *awssqs.SetQueueAttributesInput) awssqs.SetQueueAttributesRequest {
						return awssqs.SetQueueAttributesRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awssqs.SetQueueAttributesOutput{}},
						}
					},
					MockListQueueTagsRequest: func(input *awssqs.ListQueueTagsInput) awssqs.ListQueueTagsRequest {
						return awssqs.ListQueueTagsRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awssqs.ListQueueTagsOutput{
								Tags: map[string]string{
									"k":  "v",
									"k1": "v1",
								},
							}},
						}
					},
					MockUntagQueueRequest: func(input *awssqs.UntagQueueInput) awssqs.UntagQueueRequest {
						return awssqs.UntagQueueRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awssqs.UntagQueueOutput{}},
						}
					},
					MockTagQueueRequest: func(input *awssqs.TagQueueInput) awssqs.TagQueueRequest {
						return awssqs.TagQueueRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awssqs.TagQueueOutput{}},
						}
					},
				},
				cr: queue(withSpec(v1alpha1.QueueParameters{
					Tags: []v1alpha1.Tag{
						{
							Key:   "k1",
							Value: aws.String("v1"),
						},
						{
							Key:   "k2",
							Value: aws.String("k2"),
						},
					},
				}), withStatus(v1alpha1.QueueObservation{
					URL: queueURL,
				})),
			},
			want: want{
				cr: queue(withSpec(v1alpha1.QueueParameters{
					Tags: []v1alpha1.Tag{
						{
							Key:   "k1",
							Value: aws.String("v1"),
						},
						{
							Key:   "k2",
							Value: aws.String("k2"),
						},
					},
				}), withStatus(v1alpha1.QueueObservation{
					URL: queueURL,
				})),
			},
		},
		"UpdateFailure": {
			args: args{
				sqs: &fake.MockSQSClient{
					MockSetQueueAttributesRequest: func(input *awssqs.SetQueueAttributesInput) awssqs.SetQueueAttributesRequest {
						return awssqs.SetQueueAttributesRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Error: errBoom},
						}
					},
				},
				cr: queue(withStatus(v1alpha1.QueueObservation{
					URL: queueURL,
				})),
			},
			want: want{
				cr: queue(withStatus(v1alpha1.QueueObservation{
					URL: queueURL,
				})),
				err: errors.Wrap(errBoom, errUpdateFailed),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.sqs}
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
		cr  *v1alpha1.Queue
		err error
	}

	cases := map[string]struct {
		args
		want
	}{
		"Successful": {
			args: args{
				sqs: &fake.MockSQSClient{
					MockDeleteQueueRequest: func(input *awssqs.DeleteQueueInput) awssqs.DeleteQueueRequest {
						return awssqs.DeleteQueueRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Data: &awssqs.DeleteQueueOutput{}},
						}
					},
				},
				cr: queue(withConditions(runtimev1alpha1.Deleting()),
					withStatus(v1alpha1.QueueObservation{
						URL: queueURL,
					})),
			},
			want: want{
				cr: queue(withConditions(runtimev1alpha1.Deleting()),
					withStatus(v1alpha1.QueueObservation{
						URL: queueURL,
					})),
			},
		},
		"DeleteFailure": {
			args: args{
				sqs: &fake.MockSQSClient{
					MockDeleteQueueRequest: func(input *awssqs.DeleteQueueInput) awssqs.DeleteQueueRequest {
						return awssqs.DeleteQueueRequest{
							Request: &aws.Request{HTTPRequest: &http.Request{}, Retryer: aws.NoOpRetryer{}, Error: errBoom},
						}
					},
				},
				cr: queue(withConditions(runtimev1alpha1.Deleting()),
					withStatus(v1alpha1.QueueObservation{
						URL: queueURL,
					})),
			},
			want: want{
				cr: queue(withConditions(runtimev1alpha1.Deleting()),
					withStatus(v1alpha1.QueueObservation{
						URL: queueURL,
					})),
				err: errors.Wrap(errBoom, errDeleteFailed),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.sqs}
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
