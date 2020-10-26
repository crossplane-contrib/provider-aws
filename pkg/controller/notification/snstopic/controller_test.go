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

package snstopic

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awssns "github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	corev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/test"

	"github.com/crossplane/provider-aws/apis/notification/v1alpha1"
	"github.com/crossplane/provider-aws/pkg/clients/sns"
	"github.com/crossplane/provider-aws/pkg/clients/sns/fake"
)

var (
	// an arbitrary managed resource
	unexpecedItem    resource.Managed
	topicName        = "some-topic"
	topicDisplayName = "some-topic-01"
	errBoom          = errors.New("boom")
	empty            = ""
)

type args struct {
	topic sns.TopicClient
	kube  client.Client
	cr    resource.Managed
}

// Topic Modifier
type topicModifier func(*v1alpha1.SNSTopic)

func makeARN(s string) string {
	return fmt.Sprintf("arn:aws:sns:ap-south-1:862356124505:%s", s)
}

func withDisplayName(s *string) topicModifier {
	return func(t *v1alpha1.SNSTopic) { t.Spec.ForProvider.DisplayName = s }
}

func withKmsMasterKeyID(s *string) topicModifier {
	return func(t *v1alpha1.SNSTopic) { t.Spec.ForProvider.KMSMasterKeyID = s }
}

func withPolicy(s *string) topicModifier {
	return func(t *v1alpha1.SNSTopic) { t.Spec.ForProvider.Policy = s }
}

func withDeliveryPolicy(s *string) topicModifier {
	return func(t *v1alpha1.SNSTopic) { t.Spec.ForProvider.DeliveryPolicy = s }
}

func withObservationOwner(s *string) topicModifier {
	return func(t *v1alpha1.SNSTopic) { t.Status.AtProvider.Owner = s }
}

func withTopicARN(s *string) topicModifier {
	return func(t *v1alpha1.SNSTopic) {
		t.Spec.ForProvider.Name = *s
		meta.SetExternalName(t, makeARN(*s))
		t.Status.AtProvider.ARN = makeARN(*s)
	}
}

func withTopicName(s *string) topicModifier {
	return func(t *v1alpha1.SNSTopic) {
		t.Spec.ForProvider.Name = *s
		meta.SetExternalName(t, *s)
	}
}

func withConditions(c ...corev1alpha1.Condition) topicModifier {
	return func(r *v1alpha1.SNSTopic) { r.Status.ConditionedStatus.Conditions = c }
}

func topic(m ...topicModifier) *v1alpha1.SNSTopic {
	cr := &v1alpha1.SNSTopic{}

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
		"InValidInput": {
			args: args{
				cr: unexpecedItem,
				kube: &test.MockClient{
					MockUpdate: test.NewMockUpdateFn(nil),
				},
			},
			want: want{
				cr:  unexpecedItem,
				err: errors.New(errUnexpectedObject),
			},
		},
		"ExternalNameNotFilled": {
			args: args{
				cr: topic(),
				kube: &test.MockClient{
					MockUpdate: test.NewMockUpdateFn(nil),
				},
			},
			want: want{
				cr: topic(),
				result: managed.ExternalObservation{
					ResourceExists:   false,
					ResourceUpToDate: false,
				},
			},
		},
		"TopicNotFound": {
			args: args{
				topic: &fake.MockTopicClient{
					MockGetTopicAttributesRequest: func(input *awssns.GetTopicAttributesInput) awssns.GetTopicAttributesRequest {
						return awssns.GetTopicAttributesRequest{
							Request: &aws.Request{
								HTTPRequest: &http.Request{},
								Error:       errBoom,
								Data:        &awssns.GetTopicAttributesOutput{},
								Retryer:     aws.NoOpRetryer{},
							},
						}
					},
				},
				cr: topic(withTopicARN(&topicName)),
				kube: &test.MockClient{
					MockUpdate: test.NewMockUpdateFn(nil),
				},
			},
			want: want{
				cr: topic(withTopicARN(&topicName)),
				result: managed.ExternalObservation{
					ResourceExists:   false,
					ResourceUpToDate: false,
				},
				err: errors.Wrap(errBoom, errGetTopicAttr),
			},
		},
		"ClientGetTopicAttributesError": {
			args: args{
				topic: &fake.MockTopicClient{
					MockGetTopicAttributesRequest: func(input *awssns.GetTopicAttributesInput) awssns.GetTopicAttributesRequest {
						return awssns.GetTopicAttributesRequest{
							Request: &aws.Request{
								HTTPRequest: &http.Request{},
								Retryer:     aws.NoOpRetryer{},
								Error:       errBoom,
							},
						}
					},
				},
				cr: topic(withTopicARN(&topicName)),
				kube: &test.MockClient{
					MockUpdate: test.NewMockUpdateFn(nil),
				},
			},
			want: want{
				cr: topic(withTopicARN(&topicName)),
				result: managed.ExternalObservation{
					ResourceExists:   false,
					ResourceUpToDate: false,
				},
				err: errors.Wrap(errBoom, errGetTopicAttr),
			},
		},
		"ValidInputResourceNotUpToDate": {
			args: args{
				topic: &fake.MockTopicClient{
					MockGetTopicAttributesRequest: func(input *awssns.GetTopicAttributesInput) awssns.GetTopicAttributesRequest {
						return awssns.GetTopicAttributesRequest{
							Request: &aws.Request{
								HTTPRequest: &http.Request{},
								Retryer:     aws.NoOpRetryer{},
								Data: &awssns.GetTopicAttributesOutput{
									Attributes: map[string]string{
										"TopicArn": makeARN(topicName),
									},
								},
							},
						}
					},
				},
				kube: &test.MockClient{
					MockUpdate: test.NewMockUpdateFn(nil),
				},
				cr: topic(
					withDisplayName(&topicDisplayName),
					withTopicARN(&topicName),
				),
			},
			want: want{
				cr: topic(
					withDisplayName(&topicDisplayName),
					withTopicARN(&topicName),
					withPolicy(&empty),
					withDeliveryPolicy(&empty),
					withKmsMasterKeyID(&empty),
					withConditions(corev1alpha1.Available()),
					withObservationOwner(&empty),
				),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: false,
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.topic, kube: tc.kube}
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
				topic: &fake.MockTopicClient{
					MockCreateTopicRequest: func(input *awssns.CreateTopicInput) awssns.CreateTopicRequest {
						return awssns.CreateTopicRequest{
							Request: &aws.Request{
								HTTPRequest: &http.Request{},
								Retryer:     aws.NoOpRetryer{},
								Data:        &awssns.CreateTopicOutput{},
							},
						}
					},
				},
				kube: &test.MockClient{
					MockUpdate: test.NewMockUpdateFn(nil, func(obj runtime.Object) error {
						o := obj.(metav1.Object)
						o.SetAnnotations(map[string]string{
							meta.AnnotationKeyExternalName: topicName,
						})
						return nil
					}),
				},
				cr: topic(
					withDisplayName(&topicDisplayName),
					withTopicName(&topicName),
				),
			},
			want: want{
				cr: topic(
					withDisplayName(&topicDisplayName),
					withTopicName(&topicName),
					withConditions(corev1alpha1.Creating()),
				),
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
		"ClientCreateTopicError": {
			args: args{
				topic: &fake.MockTopicClient{
					MockCreateTopicRequest: func(input *awssns.CreateTopicInput) awssns.CreateTopicRequest {
						return awssns.CreateTopicRequest{
							Request: &aws.Request{
								HTTPRequest: &http.Request{},
								Retryer:     aws.NoOpRetryer{},
								Error:       errBoom,
							},
						}
					},
				},
				cr: topic(
					withDisplayName(&topicDisplayName),
					withTopicName(&topicName),
				),
			},
			want: want{
				cr: topic(
					withDisplayName(&topicDisplayName),
					withTopicName(&topicName),
					withConditions(corev1alpha1.Creating()),
				),
				err: errors.Wrap(errBoom, errCreate),
			},
		},
		"KubeUpdateError": {
			args: args{
				topic: &fake.MockTopicClient{
					MockCreateTopicRequest: func(input *awssns.CreateTopicInput) awssns.CreateTopicRequest {
						return awssns.CreateTopicRequest{
							Request: &aws.Request{
								HTTPRequest: &http.Request{},
								Retryer:     aws.NoOpRetryer{},
								Data:        &awssns.CreateTopicOutput{},
							},
						}
					},
				},
				kube: &test.MockClient{
					MockUpdate: test.NewMockUpdateFn(nil, func(obj runtime.Object) error {
						o := obj.(metav1.Object)
						o.SetAnnotations(map[string]string{
							meta.AnnotationKeyExternalName: topicName,
						})
						return errBoom
					}),
				},
				cr: topic(
					withDisplayName(&topicDisplayName),
					withTopicName(&topicName),
				),
			},
			want: want{
				cr: topic(
					withDisplayName(&topicDisplayName),
					withTopicName(&topicName),
					withConditions(corev1alpha1.Creating()),
				),
				err: errors.Wrap(errBoom, errKubeTopicUpdateFailed),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.topic, kube: tc.kube}
			o, err := e.Create(context.Background(), tc.args.cr)

			if diff := cmp.Diff(tc.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got\n%s", diff)
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
		"VaildInput": {
			args: args{
				topic: &fake.MockTopicClient{
					MockGetTopicAttributesRequest: func(input *awssns.GetTopicAttributesInput) awssns.GetTopicAttributesRequest {
						return awssns.GetTopicAttributesRequest{
							Request: &aws.Request{
								HTTPRequest: &http.Request{},
								Retryer:     aws.NoOpRetryer{},
								Data:        &awssns.GetTopicAttributesOutput{},
							},
						}
					},
				},
				cr: topic(withTopicName(&topicName)),
			},
			want: want{
				cr: topic(withTopicName(&topicName)),
			},
		},
		"VaildInputWithChangedAttributes": {
			args: args{
				topic: &fake.MockTopicClient{
					MockGetTopicAttributesRequest: func(input *awssns.GetTopicAttributesInput) awssns.GetTopicAttributesRequest {
						return awssns.GetTopicAttributesRequest{
							Request: &aws.Request{
								HTTPRequest: &http.Request{},
								Retryer:     aws.NoOpRetryer{},
								Data:        &awssns.GetTopicAttributesOutput{},
							},
						}
					},
					MockSetTopicAttributesRequest: func(input *awssns.SetTopicAttributesInput) awssns.SetTopicAttributesRequest {
						return awssns.SetTopicAttributesRequest{
							Request: &aws.Request{
								HTTPRequest: &http.Request{},
								Retryer:     aws.NoOpRetryer{},
								Data:        &awssns.SetTopicAttributesOutput{},
							},
						}
					},
				},
				cr: topic(
					withDisplayName(&topicDisplayName),
					withTopicName(&topicName),
				),
			},
			want: want{
				cr: topic(
					withDisplayName(&topicDisplayName),
					withTopicName(&topicName),
				),
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
		"ClientGetTopicAttributeError": {
			args: args{
				topic: &fake.MockTopicClient{
					MockGetTopicAttributesRequest: func(input *awssns.GetTopicAttributesInput) awssns.GetTopicAttributesRequest {
						return awssns.GetTopicAttributesRequest{
							Request: &aws.Request{
								HTTPRequest: &http.Request{},
								Retryer:     aws.NoOpRetryer{},
								Error:       errBoom,
							},
						}
					},
				},
				cr: topic(withTopicName(&topicName)),
			},
			want: want{
				cr:  topic(withTopicName(&topicName)),
				err: errors.Wrap(errBoom, errGetTopicAttr),
			},
		},
		"ClientSetTopicAttributeError": {
			args: args{
				topic: &fake.MockTopicClient{
					MockGetTopicAttributesRequest: func(input *awssns.GetTopicAttributesInput) awssns.GetTopicAttributesRequest {
						return awssns.GetTopicAttributesRequest{
							Request: &aws.Request{
								HTTPRequest: &http.Request{},
								Retryer:     aws.NoOpRetryer{},
								Data:        &awssns.GetTopicAttributesOutput{},
							},
						}
					},
					MockSetTopicAttributesRequest: func(input *awssns.SetTopicAttributesInput) awssns.SetTopicAttributesRequest {
						return awssns.SetTopicAttributesRequest{
							Request: &aws.Request{
								HTTPRequest: &http.Request{},
								Retryer:     aws.NoOpRetryer{},
								Error:       errBoom,
							},
						}
					},
				},
				cr: topic(
					withDisplayName(&topicDisplayName),
					withTopicName(&topicName),
				),
			},
			want: want{
				cr: topic(
					withDisplayName(&topicDisplayName),
					withTopicName(&topicName),
				),
				err: errors.Wrap(errBoom, errUpdate),
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.topic}
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
				topic: &fake.MockTopicClient{
					MockDeleteTopicRequest: func(input *awssns.DeleteTopicInput) awssns.DeleteTopicRequest {
						return awssns.DeleteTopicRequest{
							Request: &aws.Request{
								HTTPRequest: &http.Request{},
								Retryer:     aws.NoOpRetryer{},
								Data:        &awssns.DeleteTopicOutput{},
							},
						}
					},
				},
				cr: topic(withTopicName(&topicName)),
			},
			want: want{
				cr: topic(
					withTopicName(&topicName),
					withConditions(corev1alpha1.Deleting()),
				),
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
				topic: &fake.MockTopicClient{
					MockDeleteTopicRequest: func(input *awssns.DeleteTopicInput) awssns.DeleteTopicRequest {
						return awssns.DeleteTopicRequest{
							Request: &aws.Request{
								HTTPRequest: &http.Request{},
								Retryer:     aws.NoOpRetryer{},
								Error:       errBoom,
							},
						}
					},
				},
				cr: topic(),
			},
			want: want{
				cr:  topic(withConditions(corev1alpha1.Deleting())),
				err: errors.Wrap(errBoom, errDelete),
			},
		},
		"ResourceDoesNotExist": {
			args: args{
				topic: &fake.MockTopicClient{
					MockDeleteTopicRequest: func(input *awssns.DeleteTopicInput) awssns.DeleteTopicRequest {
						return awssns.DeleteTopicRequest{
							Request: &aws.Request{
								HTTPRequest: &http.Request{},
								Retryer:     aws.NoOpRetryer{},
								Data:        &awssns.DeleteTopicOutput{},
								Error:       nil,
							},
						}
					},
				},
				cr: topic(),
			},
			want: want{
				cr: topic(withConditions(corev1alpha1.Deleting())),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.topic}
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
