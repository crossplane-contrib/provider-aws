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

package sns

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/google/go-cmp/cmp"

	"github.com/crossplane/crossplane-runtime/pkg/test"

	"github.com/crossplane/provider-aws/apis/notification/v1alpha1"
	aws "github.com/crossplane/provider-aws/pkg/clients"
)

var (
	subOwner               = "owner"
	subEmailProtocol       = "email"
	subEmailEndpoint       = "xyz@abc.com"
	subArn                 = "some-arn"
	subRawMessageDelivery  = "raw-message"
	subFilterPolicy        = "filter-policy"
	subRedrivePolicy       = "redrive-policy"
	subDeliveryPolicy      = "delivery-policy"
	subWrongProtocol       = "wrong-protocol"
	subConfirmationPending = v1alpha1.ConfirmationPending
	withSubConfirmed       = v1alpha1.ConfirmationSuccessful
	subStringFalse         = "false"
	subStringTrue          = "true"
	subBoolTrue            = true
)

// Subscription Attribute Modifier
type subAttrModifier func(*map[string]string)

// subscription Observation Modifer
type subObservationModifier func(*v1alpha1.SNSSubscriptionObservation)

func subAttributes(m ...subAttrModifier) *map[string]string {
	attr := &map[string]string{}

	for _, f := range m {
		f(attr)
	}

	return attr
}

func withSubDeliveryPolicy(s *string) subAttrModifier {
	return func(attr *map[string]string) {
		(*attr)["DeliveryPolicy"] = *s
	}
}

func withSubFilterPolicy(s *string) subAttrModifier {
	return func(attr *map[string]string) {
		(*attr)["FilterPolicy"] = *s
	}
}
func withSubRawMessageDelivery(s *string) subAttrModifier {
	return func(attr *map[string]string) {
		(*attr)["RawMessageDelivery"] = *s
	}
}
func withSubRedrivePolicy(s *string) subAttrModifier {
	return func(attr *map[string]string) {
		(*attr)["RedrivePolicy"] = *s
	}
}

func withSubConfirmation(s *v1alpha1.ConfirmationStatus) subAttrModifier {
	return func(attr *map[string]string) {
		(*attr)["PendingConfirmation"] = subStringTrue
	}
}

func withSubConfirmationWasAuthenticated(s *string) subAttrModifier {
	return func(attr *map[string]string) {
		(*attr)["ConfirmationWasAuthenticated"] = *s
	}
}

func withSubOwner(s *string) subAttrModifier {
	return func(attr *map[string]string) {
		(*attr)["Owner"] = *s
	}
}

func withSubTopicARN(s *string) subAttrModifier {
	return func(attr *map[string]string) {
		(*attr)["TopicArn"] = *s
	}
}

func sub(m ...func(*sns.Subscription)) *sns.Subscription {
	o := &sns.Subscription{
		TopicArn:        aws.String(topicArn),
		Endpoint:        aws.String(subEmailEndpoint),
		Protocol:        aws.String(subEmailProtocol),
		Owner:           aws.String(subOwner),
		SubscriptionArn: aws.String(subArn),
	}

	for _, f := range m {
		f(o)
	}

	return o
}

// subscription Parameters
func subParams(m ...func(*v1alpha1.SNSSubscriptionParameters)) *v1alpha1.SNSSubscriptionParameters {
	o := &v1alpha1.SNSSubscriptionParameters{
		TopicArn:           topicArn,
		Endpoint:           &subEmailEndpoint,
		Protocol:           &subEmailProtocol,
		RedrivePolicy:      &subRedrivePolicy,
		FilterPolicy:       &subFilterPolicy,
		RawMessageDelivery: &subRawMessageDelivery,
		DeliveryPolicy:     &subDeliveryPolicy,
	}

	for _, f := range m {
		f(o)
	}

	return o
}

func subObservation(m ...func(*v1alpha1.SNSSubscriptionObservation)) *v1alpha1.SNSSubscriptionObservation {
	o := &v1alpha1.SNSSubscriptionObservation{}

	for _, f := range m {
		f(o)
	}
	return o
}

func withSubObservationOwner(s *string) subObservationModifier {
	return func(o *v1alpha1.SNSSubscriptionObservation) {
		o.Owner = s
	}
}

func withSubObservationStatus(s *v1alpha1.ConfirmationStatus) subObservationModifier {
	return func(o *v1alpha1.SNSSubscriptionObservation) {
		if *s == subConfirmationPending {
			o.Status = &subConfirmationPending
		} else {
			o.Status = &withSubConfirmed
		}
	}
}

func withSubObservationConfirmationWasAuthenticated(s bool) subObservationModifier {
	return func(o *v1alpha1.SNSSubscriptionObservation) {
		o.ConfirmationWasAuthenticated = &s
	}
}

func TestGetSNSSubscription(t *testing.T) {
	type args struct {
		list *sns.ListSubscriptionsByTopicResponse
		cr   *v1alpha1.SNSSubscription
	}

	type want struct {
		sub sns.Subscription
		err error
	}

	cases := map[string]struct {
		args
		want
	}{
		"SubcriptionFound": {
			args: args{
				list: &sns.ListSubscriptionsByTopicResponse{
					ListSubscriptionsByTopicOutput: &sns.ListSubscriptionsByTopicOutput{
						Subscriptions: []sns.Subscription{
							{
								TopicArn:        &topicArn,
								Endpoint:        &subEmailEndpoint,
								Owner:           &subOwner,
								SubscriptionArn: &subArn,
								Protocol:        &subEmailProtocol,
							},
						},
					},
				},
				cr: &v1alpha1.SNSSubscription{
					Spec: v1alpha1.SNSSubscriptionSpec{
						ForProvider: v1alpha1.SNSSubscriptionParameters{
							TopicArn: topicArn,
							Endpoint: &subEmailEndpoint,
							Protocol: &subEmailProtocol,
						},
					},
				},
			},
			want: want{
				sub: *sub(),
			},
		},
		"SubcriptionNotFound": {
			args: args{
				list: &sns.ListSubscriptionsByTopicResponse{
					ListSubscriptionsByTopicOutput: &sns.ListSubscriptionsByTopicOutput{
						Subscriptions: []sns.Subscription{
							{
								TopicArn:        &topicArn,
								Endpoint:        &subEmailEndpoint,
								Owner:           &subOwner,
								SubscriptionArn: &subArn,
							},
						},
					},
				},
				cr: &v1alpha1.SNSSubscription{
					Spec: v1alpha1.SNSSubscriptionSpec{
						ForProvider: v1alpha1.SNSSubscriptionParameters{
							TopicArn: topicArn,
							Endpoint: &subEmailEndpoint,
							Protocol: &subEmailProtocol,
						},
					},
				},
			},
			want: want{
				err: &SubscriptionNotFound{},
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			sub, err := GetSNSSubscription(tc.list, tc.cr)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.sub, sub); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}

		})
	}
}

func TestGenerateSubscribeInput(t *testing.T) {
	cases := map[string]struct {
		in  v1alpha1.SNSSubscriptionParameters
		out sns.SubscribeInput
	}{
		"FilledInput": {
			in: v1alpha1.SNSSubscriptionParameters{
				TopicArn: topicArn,
				Endpoint: &subEmailEndpoint,
				Protocol: &subEmailProtocol,
			},
			out: sns.SubscribeInput{
				TopicArn:              aws.String(topicArn),
				Endpoint:              &subEmailEndpoint,
				Protocol:              &subEmailProtocol,
				ReturnSubscriptionArn: &subBoolTrue,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			input := GenerateSubscribeInput(&tc.in)
			if diff := cmp.Diff(input, &tc.out); diff != "" {
				t.Errorf("GenerateSubscribeInput(...): -want, +got\n:%s", diff)
			}
		})
	}
}

func TestLateInitializeSubscription(t *testing.T) {
	type args struct {
		spec *v1alpha1.SNSSubscriptionParameters
		attr map[string]string
	}
	cases := map[string]struct {
		args args
		want *v1alpha1.SNSSubscriptionParameters
	}{
		"AllFilledNoDiff": {
			args: args{
				spec: subParams(),
				attr: *subAttributes(
					withSubDeliveryPolicy(&subDeliveryPolicy),
					withSubRedrivePolicy(&subRedrivePolicy),
					withSubRawMessageDelivery(&subRawMessageDelivery),
					withSubFilterPolicy(&subFilterPolicy),
				),
			},
			want: subParams(),
		},
		"PartialFilled": {
			args: args{
				spec: subParams(func(sub *v1alpha1.SNSSubscriptionParameters) {
					sub.TopicArn = topicArn
					sub.Protocol = &subEmailProtocol
					sub.Endpoint = &subEmailEndpoint
				}),
				attr: map[string]string{},
			},
			want: subParams(func(sub *v1alpha1.SNSSubscriptionParameters) {
				sub.TopicArn = topicArn
				sub.Endpoint = &subEmailEndpoint
				sub.Protocol = &subEmailProtocol
			}),
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			LateInitializeSubscription(tc.args.spec, tc.args.attr)
			if diff := cmp.Diff(tc.args.spec, tc.want); diff != "" {
				t.Errorf("LateInitializeTopic(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGetChangedSubAttributes(t *testing.T) {

	type args struct {
		p    v1alpha1.SNSSubscriptionParameters
		attr *map[string]string
	}

	cases := map[string]struct {
		args args
		want *map[string]string
	}{
		"NoChange": {
			args: args{
				p: v1alpha1.SNSSubscriptionParameters{
					Protocol: &subEmailProtocol,
					Endpoint: &subEmailEndpoint,
				},
				attr: subAttributes(),
			},
			want: subAttributes(),
		},
		"Change": {
			args: args{
				p: v1alpha1.SNSSubscriptionParameters{
					Protocol: &subEmailProtocol,
					Endpoint: &subEmailEndpoint,
				},
				attr: subAttributes(),
			},
			want: subAttributes(),
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			c := GetChangedSubAttributes(tc.args.p, *tc.args.attr)
			if diff := cmp.Diff(*tc.want, c); diff != "" {
				t.Errorf("GetChangedSubAttributes(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGenerateSubscriptionObservation(t *testing.T) {
	cases := map[string]struct {
		in  *map[string]string
		out *v1alpha1.SNSSubscriptionObservation
	}{
		"All Filled": {
			in: subAttributes(
				withSubOwner(&subOwner),
				withSubConfirmation(&subConfirmationPending),
				withSubConfirmationWasAuthenticated(&subStringFalse),
			),
			out: subObservation(
				withSubObservationOwner(&subOwner),
				withSubObservationStatus(&subConfirmationPending),
				withSubObservationConfirmationWasAuthenticated(false),
			),
		},
		"Partially Filled": {
			in: subAttributes(
				withSubOwner(&subOwner),
				withSubConfirmation(&subConfirmationPending),
			),
			out: subObservation(
				withSubObservationOwner(&subOwner),
				withSubObservationStatus(&subConfirmationPending),
			),
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			observation := GenerateSubscriptionObservation(*tc.in)
			if diff := cmp.Diff(*tc.out, observation); diff != "" {
				t.Errorf("GenerateSubscriptionObservation(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestIsSNSSubscriptionUpToDate(t *testing.T) {
	type args struct {
		p    v1alpha1.SNSSubscriptionParameters
		sub  *sns.Subscription
		attr *map[string]string
	}

	cases := map[string]struct {
		args args
		want bool
	}{
		"SameFieldsAndAllFilled": {
			args: args{
				attr: subAttributes(
					withSubRawMessageDelivery(&subRawMessageDelivery),
					withSubRedrivePolicy(&subRedrivePolicy),
					withSubFilterPolicy(&subFilterPolicy),
					withSubDeliveryPolicy(&subDeliveryPolicy),
					withSubTopicARN(&topicArn),
				),
				sub: &sns.Subscription{
					Protocol: &subEmailProtocol,
					Endpoint: &subEmailEndpoint,
				},
				p: v1alpha1.SNSSubscriptionParameters{
					Protocol:           &subEmailProtocol,
					Endpoint:           &subEmailEndpoint,
					RawMessageDelivery: &subRawMessageDelivery,
					FilterPolicy:       &subFilterPolicy,
					DeliveryPolicy:     &subDeliveryPolicy,
					TopicArn:           topicArn,
					RedrivePolicy:      &subRedrivePolicy,
				},
			},
			want: true,
		},
		"DifferentFields": {
			args: args{
				attr: subAttributes(
					withSubRawMessageDelivery(&subRawMessageDelivery),
					withSubRedrivePolicy(&subRedrivePolicy),
					withSubFilterPolicy(&subFilterPolicy),
					withSubDeliveryPolicy(&subDeliveryPolicy),
					withSubTopicARN(&topicArn),
				),
				sub: &sns.Subscription{
					Protocol: &subEmailProtocol,
					Endpoint: &subEmailEndpoint,
				},
				p: v1alpha1.SNSSubscriptionParameters{
					Protocol:           &subWrongProtocol,
					Endpoint:           &subEmailEndpoint,
					RawMessageDelivery: &subRawMessageDelivery,
					FilterPolicy:       &subFilterPolicy,
					DeliveryPolicy:     &subDeliveryPolicy,
					TopicArn:           topicArn,
					RedrivePolicy:      &subRedrivePolicy,
				},
			},
			want: false,
		},
		"MissingFields": {
			args: args{
				attr: subAttributes(
					withSubRawMessageDelivery(&subRawMessageDelivery),
					withSubRedrivePolicy(&subRedrivePolicy),
					withSubFilterPolicy(&subFilterPolicy),
					withSubDeliveryPolicy(&subDeliveryPolicy),
					withSubTopicARN(&topicArn),
				),
				sub: &sns.Subscription{
					Protocol: &subEmailProtocol,
					Endpoint: &subEmailEndpoint,
				},
				p: v1alpha1.SNSSubscriptionParameters{},
			},
			want: false,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := IsSNSSubscriptionUpToDate(tc.args.p, tc.args.sub, *tc.args.attr)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("Topic : -want, +got:\n%s", diff)
			}
		})
	}
}
