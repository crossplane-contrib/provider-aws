package subscription

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

	"github.com/crossplane-contrib/provider-aws/apis/sns/v1beta1"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/sns"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/sns/fake"
	errorutils "github.com/crossplane-contrib/provider-aws/pkg/utils/errors"
)

var (
	// an arbitrary managed resource
	unexpectedItem resource.Managed
	subName        = "some-topic"
	errBoom        = errors.New("boom")
)

type args struct {
	sub sns.SubscriptionClient
	cr  resource.Managed
}

func makeARN(s string) string {
	return fmt.Sprintf("arn:aws:sns:ap-south-1:862356124505:%s", s)
}

// Subscription Modifier
type subModifier func(*v1beta1.Subscription)

func withConditions(c ...xpv1.Condition) subModifier {
	return func(r *v1beta1.Subscription) { r.Status.ConditionedStatus.Conditions = c }
}

func subscription(m ...subModifier) *v1beta1.Subscription {
	cr := &v1beta1.Subscription{}

	for _, f := range m {
		f(cr)
	}

	return cr
}

func withSubARN(s *string) subModifier {
	return func(t *v1beta1.Subscription) {
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
				cr: unexpectedItem,
			},
			want: want{
				cr:  unexpectedItem,
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
					MockSubscribe: func(ctx context.Context, input *awssns.SubscribeInput, opts []func(*awssns.Options)) (*awssns.SubscribeOutput, error) {
						return &awssns.SubscribeOutput{SubscriptionArn: aws.String(makeARN(subName))}, nil
					},
				},
				cr: subscription(
					withSubARN(&subName),
				),
			},
			want: want{
				cr:     subscription(withSubARN(&subName)),
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
		"ClientSubscribeError": {
			args: args{
				sub: &fake.MockSubscriptionClient{
					MockSubscribe: func(ctx context.Context, input *awssns.SubscribeInput, opts []func(*awssns.Options)) (*awssns.SubscribeOutput, error) {
						return nil, errBoom
					},
				},
				cr: subscription(
					withSubARN(&subName),
				),
			},
			want: want{
				cr: subscription(
					withSubARN(&subName)),
				err: errorutils.Wrap(errBoom, errCreate),
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
					MockGetSubscriptionAttributes: func(ctx context.Context, input *awssns.GetSubscriptionAttributesInput, opts []func(*awssns.Options)) (*awssns.GetSubscriptionAttributesOutput, error) {
						return &awssns.GetSubscriptionAttributesOutput{}, nil
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
					MockGetSubscriptionAttributes: func(ctx context.Context, input *awssns.GetSubscriptionAttributesInput, opts []func(*awssns.Options)) (*awssns.GetSubscriptionAttributesOutput, error) {
						return &awssns.GetSubscriptionAttributesOutput{}, nil
					},
					MockSetSubscriptionAttributes: func(ctx context.Context, input *awssns.SetSubscriptionAttributesInput, opts []func(*awssns.Options)) (*awssns.SetSubscriptionAttributesOutput, error) {
						return &awssns.SetSubscriptionAttributesOutput{}, nil
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
				cr: unexpectedItem,
			},
			want: want{
				cr:  unexpectedItem,
				err: errors.New(errUnexpectedObject),
			},
		},
		"ClientGetSubscriptionAttributeError": {
			args: args{
				sub: &fake.MockSubscriptionClient{
					MockGetSubscriptionAttributes: func(ctx context.Context, input *awssns.GetSubscriptionAttributesInput, opts []func(*awssns.Options)) (*awssns.GetSubscriptionAttributesOutput, error) {
						return nil, errBoom
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
				err: errorutils.Wrap(errBoom, errUpdate),
			},
		},
		"ClientSetSubscriptionAttributeError": {
			args: args{
				sub: &fake.MockSubscriptionClient{
					MockGetSubscriptionAttributes: func(ctx context.Context, input *awssns.GetSubscriptionAttributesInput, opts []func(*awssns.Options)) (*awssns.GetSubscriptionAttributesOutput, error) {
						return &awssns.GetSubscriptionAttributesOutput{
							Attributes: map[string]string{
								"DeliveryPolicy": "fake-del-policy",
							}}, nil
					},
					MockSetSubscriptionAttributes: func(ctx context.Context, input *awssns.SetSubscriptionAttributesInput, opts []func(*awssns.Options)) (*awssns.SetSubscriptionAttributesOutput, error) {
						return nil, errBoom
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
				err: errorutils.Wrap(errBoom, errUpdate),
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
					MockUnsubscribe: func(ctx context.Context, input *awssns.UnsubscribeInput, opts []func(*awssns.Options)) (*awssns.UnsubscribeOutput, error) {
						return &awssns.UnsubscribeOutput{}, nil
					},
				},
				cr: subscription(
					withSubARN(&subName),
					withConditions(xpv1.Deleting()),
				),
			},
			want: want{
				cr: subscription(
					withSubARN(&subName),
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
				sub: &fake.MockSubscriptionClient{
					MockUnsubscribe: func(ctx context.Context, input *awssns.UnsubscribeInput, opts []func(*awssns.Options)) (*awssns.UnsubscribeOutput, error) {
						return nil, errBoom
					},
				},
				cr: subscription(
					withSubARN(&subName),
				),
			},
			want: want{
				cr: subscription(
					withSubARN(&subName),
					withConditions(xpv1.Deleting()),
				),
				err: errorutils.Wrap(errBoom, errDelete),
			},
		},
		"ResourceDoesNotExist": {
			args: args{
				sub: &fake.MockSubscriptionClient{
					MockUnsubscribe: func(ctx context.Context, input *awssns.UnsubscribeInput, opts []func(*awssns.Options)) (*awssns.UnsubscribeOutput, error) {
						return &awssns.UnsubscribeOutput{}, nil
					},
				},
				cr: subscription(
					withSubARN(&subName),
				),
			},
			want: want{
				cr: subscription(
					withSubARN(&subName),
					withConditions(xpv1.Deleting()),
				),
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.sub}
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
