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

package topic

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awssns "github.com/aws/aws-sdk-go-v2/service/sns"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane-contrib/provider-aws/apis/sns/v1beta1"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/sns"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/sns/fake"
	errorutils "github.com/crossplane-contrib/provider-aws/pkg/utils/errors"
)

var (
	// an arbitrary managed resource
	unexpectedItem   resource.Managed
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
type topicModifier func(*v1beta1.Topic)

func makeARN(s string) string {
	return fmt.Sprintf("arn:aws:sns:ap-south-1:862356124505:%s", s)
}

func withDisplayName(s *string) topicModifier {
	return func(t *v1beta1.Topic) { t.Spec.ForProvider.DisplayName = s }
}

func withKmsMasterKeyID(s *string) topicModifier {
	return func(t *v1beta1.Topic) { t.Spec.ForProvider.KMSMasterKeyID = s }
}

func withPolicy(s *string) topicModifier {
	return func(t *v1beta1.Topic) { t.Spec.ForProvider.Policy = s }
}

func withDeliveryPolicy(s *string) topicModifier {
	return func(t *v1beta1.Topic) { t.Spec.ForProvider.DeliveryPolicy = s }
}

func withObservationOwner(s *string) topicModifier {
	return func(t *v1beta1.Topic) { t.Status.AtProvider.Owner = s }
}

func withTopicARN(s *string) topicModifier {
	return func(t *v1beta1.Topic) {
		t.Spec.ForProvider.Name = *s
		meta.SetExternalName(t, makeARN(*s))
		t.Status.AtProvider.ARN = makeARN(*s)
	}
}

func withTopicName(s *string) topicModifier {
	return func(t *v1beta1.Topic) {
		t.Spec.ForProvider.Name = *s
		meta.SetExternalName(t, *s)
	}
}

func withConditions(c ...xpv1.Condition) topicModifier {
	return func(r *v1beta1.Topic) { r.Status.ConditionedStatus.Conditions = c }
}

func topic(m ...topicModifier) *v1beta1.Topic {
	cr := &v1beta1.Topic{}

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
				cr: unexpectedItem,
				kube: &test.MockClient{
					MockUpdate: test.NewMockUpdateFn(nil),
				},
			},
			want: want{
				cr:  unexpectedItem,
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
					MockGetTopicAttributes: func(ctx context.Context, input *awssns.GetTopicAttributesInput, opts []func(*awssns.Options)) (*awssns.GetTopicAttributesOutput, error) {
						return nil, errBoom
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
				err: errorutils.Wrap(errBoom, errGetTopicAttr),
			},
		},
		"ClientGetTopicAttributesError": {
			args: args{
				topic: &fake.MockTopicClient{
					MockGetTopicAttributes: func(ctx context.Context, input *awssns.GetTopicAttributesInput, opts []func(*awssns.Options)) (*awssns.GetTopicAttributesOutput, error) {
						return nil, errBoom
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
				err: errorutils.Wrap(errBoom, errGetTopicAttr),
			},
		},
		"ValidInputResourceNotUpToDate": {
			args: args{
				topic: &fake.MockTopicClient{
					MockGetTopicAttributes: func(ctx context.Context, input *awssns.GetTopicAttributesInput, opts []func(*awssns.Options)) (*awssns.GetTopicAttributesOutput, error) {
						return &awssns.GetTopicAttributesOutput{
							Attributes: map[string]string{
								"TopicArn": makeARN(topicName),
							},
						}, nil
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
					withConditions(xpv1.Available()),
					withObservationOwner(&empty),
				),
				result: managed.ExternalObservation{
					ResourceExists:          true,
					ResourceUpToDate:        false,
					ResourceLateInitialized: true,
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
					MockCreateTopic: func(ctx context.Context, input *awssns.CreateTopicInput, opts []func(*awssns.Options)) (*awssns.CreateTopicOutput, error) {
						return &awssns.CreateTopicOutput{TopicArn: aws.String(topicName)}, nil
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
					withTopicName(&topicName)),
				result: managed.ExternalCreation{},
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
		"ClientCreateTopicError": {
			args: args{
				topic: &fake.MockTopicClient{
					MockCreateTopic: func(ctx context.Context, input *awssns.CreateTopicInput, opts []func(*awssns.Options)) (*awssns.CreateTopicOutput, error) {
						return nil, errBoom
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
					withTopicName(&topicName)),
				err: errorutils.Wrap(errBoom, errCreate),
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
					MockGetTopicAttributes: func(ctx context.Context, input *awssns.GetTopicAttributesInput, opts []func(*awssns.Options)) (*awssns.GetTopicAttributesOutput, error) {
						return &awssns.GetTopicAttributesOutput{}, nil
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
					MockGetTopicAttributes: func(ctx context.Context, input *awssns.GetTopicAttributesInput, opts []func(*awssns.Options)) (*awssns.GetTopicAttributesOutput, error) {
						return &awssns.GetTopicAttributesOutput{}, nil
					},
					MockSetTopicAttributes: func(ctx context.Context, input *awssns.SetTopicAttributesInput, opts []func(*awssns.Options)) (*awssns.SetTopicAttributesOutput, error) {
						return &awssns.SetTopicAttributesOutput{}, nil
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
				cr: unexpectedItem,
			},
			want: want{
				cr:  unexpectedItem,
				err: errors.New(errUnexpectedObject),
			},
		},
		"ClientGetTopicAttributeError": {
			args: args{
				topic: &fake.MockTopicClient{
					MockGetTopicAttributes: func(ctx context.Context, input *awssns.GetTopicAttributesInput, opts []func(*awssns.Options)) (*awssns.GetTopicAttributesOutput, error) {
						return nil, errBoom
					},
				},
				cr: topic(withTopicName(&topicName)),
			},
			want: want{
				cr:  topic(withTopicName(&topicName)),
				err: errorutils.Wrap(errBoom, errGetTopicAttr),
			},
		},
		"ClientSetTopicAttributeError": {
			args: args{
				topic: &fake.MockTopicClient{
					MockGetTopicAttributes: func(ctx context.Context, input *awssns.GetTopicAttributesInput, opts []func(*awssns.Options)) (*awssns.GetTopicAttributesOutput, error) {
						return &awssns.GetTopicAttributesOutput{}, nil
					},
					MockSetTopicAttributes: func(ctx context.Context, input *awssns.SetTopicAttributesInput, opts []func(*awssns.Options)) (*awssns.SetTopicAttributesOutput, error) {
						return nil, errBoom
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
				err: errorutils.Wrap(errBoom, errUpdate),
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
					MockDeleteTopic: func(ctx context.Context, input *awssns.DeleteTopicInput, opts []func(*awssns.Options)) (*awssns.DeleteTopicOutput, error) {
						return &awssns.DeleteTopicOutput{}, nil
					},
				},
				cr: topic(withTopicName(&topicName)),
			},
			want: want{
				cr: topic(
					withTopicName(&topicName),
					withConditions(xpv1.Deleting()),
				),
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
				topic: &fake.MockTopicClient{
					MockDeleteTopic: func(ctx context.Context, input *awssns.DeleteTopicInput, opts []func(*awssns.Options)) (*awssns.DeleteTopicOutput, error) {
						return nil, errBoom
					},
				},
				cr: topic(),
			},
			want: want{
				cr:  topic(withConditions(xpv1.Deleting())),
				err: errorutils.Wrap(errBoom, errDelete),
			},
		},
		"ResourceDoesNotExist": {
			args: args{
				topic: &fake.MockTopicClient{
					MockDeleteTopic: func(ctx context.Context, input *awssns.DeleteTopicInput, opts []func(*awssns.Options)) (*awssns.DeleteTopicOutput, error) {
						return &awssns.DeleteTopicOutput{}, nil
					},
				},
				cr: topic(),
			},
			want: want{
				cr: topic(withConditions(xpv1.Deleting())),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.topic}
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
