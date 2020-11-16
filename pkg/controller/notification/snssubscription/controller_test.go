package snssubscription

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awssns "github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"

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
	unexpecedItem resource.Managed
	subName       = "some-topic"
	errBoom       = errors.New("boom")
)

type args struct {
	sub sns.SubscriptionClient
	cr  resource.Managed
}

func makeARN(s string) string {
	return fmt.Sprintf("arn:aws:sns:ap-south-1:862356124505:%s", s)
}

// Subscription Modifier
type subModifier func(*v1alpha1.SNSSubscription)

func withConditions(c ...corev1alpha1.Condition) subModifier {
	return func(r *v1alpha1.SNSSubscription) { r.Status.ConditionedStatus.Conditions = c }
}

func subscription(m ...subModifier) *v1alpha1.SNSSubscription {
	cr := &v1alpha1.SNSSubscription{}

	for _, f := range m {
		f(cr)
	}

	return cr
}

func withSubARN(s *string) subModifier {
	return func(t *v1alpha1.SNSSubscription) {
		meta.SetExternalName(t, makeARN(*s))
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
		"InValidInput": {
			args: args{
				cr: unexpecedItem,
			},
			want: want{
				cr:  unexpecedItem,
				err: errors.New(errUnexpectedObject),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.sub}
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
				sub: &fake.MockSubscriptionClient{
					MockSubscribeRequest: func(input *awssns.SubscribeInput) awssns.SubscribeRequest {
						return awssns.SubscribeRequest{
							Request: &aws.Request{
								HTTPRequest: &http.Request{},
								Data:        &awssns.SubscribeOutput{SubscriptionArn: aws.String(makeARN(subName))},
								Retryer:     aws.NoOpRetryer{},
							},
						}
					},
				},
				cr: subscription(
					withSubARN(&subName),
				),
			},
			want: want{
				cr:     subscription(withSubARN(&subName)),
				result: managed.ExternalCreation{ExternalNameAssigned: true},
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
		"ClientSubscribeError": {
			args: args{
				sub: &fake.MockSubscriptionClient{
					MockSubscribeRequest: func(input *awssns.SubscribeInput) awssns.SubscribeRequest {
						return awssns.SubscribeRequest{
							Request: &aws.Request{
								HTTPRequest: &http.Request{},
								Error:       errBoom,
								Retryer:     aws.NoOpRetryer{},
							},
						}
					},
				},
				cr: subscription(
					withSubARN(&subName),
				),
			},
			want: want{
				cr: subscription(
					withSubARN(&subName)),
				err: errors.Wrap(errBoom, errCreate),
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.sub}
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
				sub: &fake.MockSubscriptionClient{
					MockGetSubscriptionAttributesRequest: func(input *awssns.GetSubscriptionAttributesInput) awssns.GetSubscriptionAttributesRequest {
						return awssns.GetSubscriptionAttributesRequest{
							Request: &aws.Request{
								HTTPRequest: &http.Request{},
								Data:        &awssns.GetSubscriptionAttributesOutput{},
								Retryer:     aws.NoOpRetryer{},
							},
						}
					},
				},
				cr: subscription(
					withSubARN(&subName),
				),
			},
			want: want{
				cr: subscription(
					withSubARN(&subName),
				),
			},
		},
		"VaildInputWithChangedAttributes": {
			args: args{
				sub: &fake.MockSubscriptionClient{
					MockGetSubscriptionAttributesRequest: func(input *awssns.GetSubscriptionAttributesInput) awssns.GetSubscriptionAttributesRequest {
						return awssns.GetSubscriptionAttributesRequest{
							Request: &aws.Request{
								HTTPRequest: &http.Request{},
								Data:        &awssns.GetSubscriptionAttributesOutput{},
								Retryer:     aws.NoOpRetryer{},
							},
						}
					},
					MockSetSubscriptionAttributesRequest: func(input *awssns.SetSubscriptionAttributesInput) awssns.SetSubscriptionAttributesRequest {
						return awssns.SetSubscriptionAttributesRequest{
							Request: &aws.Request{
								HTTPRequest: &http.Request{},
								Data:        &awssns.SetSubscriptionAttributesOutput{},
								Retryer:     aws.NoOpRetryer{},
							},
						}
					},
				},
				cr: subscription(
					withSubARN(&subName),
				),
			},
			want: want{
				cr: subscription(
					withSubARN(&subName),
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
		"ClientGetSubscriptionAttributeError": {
			args: args{
				sub: &fake.MockSubscriptionClient{
					MockGetSubscriptionAttributesRequest: func(input *awssns.GetSubscriptionAttributesInput) awssns.GetSubscriptionAttributesRequest {
						return awssns.GetSubscriptionAttributesRequest{
							Request: &aws.Request{
								HTTPRequest: &http.Request{},
								Error:       errBoom,
								Retryer:     aws.NoOpRetryer{},
							},
						}
					},
				},
				cr: subscription(
					withSubARN(&subName),
				),
			},
			want: want{
				cr: subscription(
					withSubARN(&subName),
				),
				err: errors.Wrap(errBoom, errUpdate),
			},
		},
		"ClientSetSubscriptionAttributeError": {
			args: args{
				sub: &fake.MockSubscriptionClient{
					MockGetSubscriptionAttributesRequest: func(input *awssns.GetSubscriptionAttributesInput) awssns.GetSubscriptionAttributesRequest {
						return awssns.GetSubscriptionAttributesRequest{
							Request: &aws.Request{
								HTTPRequest: &http.Request{},
								Retryer:     aws.NoOpRetryer{},
								Data: &awssns.GetSubscriptionAttributesOutput{
									// To trigger Changed Attributes
									Attributes: map[string]string{
										"DeliveryPolicy": "fake-del-policy",
									},
								},
							},
						}
					},
					MockSetSubscriptionAttributesRequest: func(input *awssns.SetSubscriptionAttributesInput) awssns.SetSubscriptionAttributesRequest {
						return awssns.SetSubscriptionAttributesRequest{
							Request: &aws.Request{
								HTTPRequest: &http.Request{},
								Error:       errBoom,
								Retryer:     aws.NoOpRetryer{},
							},
						}
					},
				},
				cr: subscription(
					withSubARN(&subName),
				),
			},
			want: want{
				cr: subscription(
					withSubARN(&subName),
				),
				err: errors.Wrap(errBoom, errUpdate),
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.sub}
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
				sub: &fake.MockSubscriptionClient{
					MockUnsubscribeRequest: func(input *awssns.UnsubscribeInput) awssns.UnsubscribeRequest {
						return awssns.UnsubscribeRequest{
							Request: &aws.Request{
								HTTPRequest: &http.Request{},
								Data:        &awssns.UnsubscribeOutput{},
								Retryer:     aws.NoOpRetryer{},
							},
						}
					},
				},
				cr: subscription(
					withSubARN(&subName),
					withConditions(corev1alpha1.Deleting()),
				),
			},
			want: want{
				cr: subscription(
					withSubARN(&subName),
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
				sub: &fake.MockSubscriptionClient{
					MockUnsubscribeRequest: func(input *awssns.UnsubscribeInput) awssns.UnsubscribeRequest {
						return awssns.UnsubscribeRequest{
							Request: &aws.Request{
								HTTPRequest: &http.Request{},
								Error:       errBoom,
								Retryer:     aws.NoOpRetryer{},
							},
						}
					},
				},
				cr: subscription(
					withSubARN(&subName),
				),
			},
			want: want{
				cr: subscription(
					withSubARN(&subName),
					withConditions(corev1alpha1.Deleting()),
				),
				err: errors.Wrap(errBoom, errDelete),
			},
		},
		"ResourceDoesNotExist": {
			args: args{
				sub: &fake.MockSubscriptionClient{
					MockUnsubscribeRequest: func(input *awssns.UnsubscribeInput) awssns.UnsubscribeRequest {
						return awssns.UnsubscribeRequest{
							Request: &aws.Request{
								HTTPRequest: &http.Request{},
								Data:        &awssns.UnsubscribeOutput{},
								Error:       nil,
								Retryer:     aws.NoOpRetryer{},
							},
						}
					},
				},
				cr: subscription(
					withSubARN(&subName),
				),
			},
			want: want{
				cr: subscription(
					withSubARN(&subName),
					withConditions(corev1alpha1.Deleting()),
				),
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.sub}
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
