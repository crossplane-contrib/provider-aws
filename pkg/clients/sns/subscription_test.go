/*
Copyright 2022 The Crossplane Authors.

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
	"github.com/aws/smithy-go/document"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/crossplane-contrib/provider-aws/apis/sns/v1beta1"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
)

var (
	subOwner               = "owner"
	subEmailProtocol       = "email"
	subEmailEndpoint       = "xyz@abc.com"
	subRawMessageDelivery  = "raw-message"
	subFilterPolicy        = "filter-policy"
	subRedrivePolicy       = "redrive-policy"
	subDeliveryPolicy      = "delivery-policy"
	subConfirmationPending = v1beta1.ConfirmationPending
	withSubConfirmed       = v1beta1.ConfirmationSuccessful
	subStringFalse         = "false"
	subStringTrue          = "true"
	subBoolTrue            = true
)

// Subscription Attribute Modifier
type subAttrModifier func(*map[string]string)

// subscription Observation Modifer
type subObservationModifier func(*v1beta1.SubscriptionObservation)

func subAttributes(m ...subAttrModifier) *map[string]string {
	attr := &map[string]string{}

	for _, f := range m {
		f(attr)
	}

	return attr
}

func withSubDeliveryPolicy(s *string) subAttrModifier {
	return func(attr *map[string]string) {
		(*attr)[string(SubscriptionDeliveryPolicy)] = *s
	}
}

func withSubFilterPolicy(s *string) subAttrModifier {
	return func(attr *map[string]string) {
		(*attr)[string(SubscriptionFilterPolicy)] = *s
	}
}
func withSubRawMessageDelivery(s *string) subAttrModifier {
	return func(attr *map[string]string) {
		(*attr)[string(SubscriptionRawMessageDelivery)] = *s
	}
}
func withSubRedrivePolicy(s *string) subAttrModifier {
	return func(attr *map[string]string) {
		(*attr)[string(SubscriptionRedrivePolicy)] = *s
	}
}

func withSubConfirmation(s *v1beta1.ConfirmationStatus) subAttrModifier {
	return func(attr *map[string]string) {
		(*attr)[string(SubscriptionPendingConfirmation)] = subStringTrue
	}
}

func withSubConfirmationWasAuthenticated(s *string) subAttrModifier {
	return func(attr *map[string]string) {
		(*attr)[string(SubscriptionConfirmationWasAuthenticated)] = *s
	}
}

func withSubOwner(s *string) subAttrModifier {
	return func(attr *map[string]string) {
		(*attr)[string(SubscriptionOwner)] = *s
	}
}

// subscription Parameters
func subParams(m ...func(*v1beta1.SubscriptionParameters)) *v1beta1.SubscriptionParameters {
	o := &v1beta1.SubscriptionParameters{
		TopicARN:           topicArn,
		Endpoint:           subEmailEndpoint,
		Protocol:           subEmailProtocol,
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

func subObservation(m ...func(*v1beta1.SubscriptionObservation)) *v1beta1.SubscriptionObservation {
	o := &v1beta1.SubscriptionObservation{}

	for _, f := range m {
		f(o)
	}
	return o
}

func withSubObservationOwner(s *string) subObservationModifier {
	return func(o *v1beta1.SubscriptionObservation) {
		o.Owner = s
	}
}

func withSubObservationStatus(s *v1beta1.ConfirmationStatus) subObservationModifier {
	return func(o *v1beta1.SubscriptionObservation) {
		if *s == subConfirmationPending {
			o.Status = &subConfirmationPending
		} else {
			o.Status = &withSubConfirmed
		}
	}
}

func withSubObservationConfirmationWasAuthenticated(s bool) subObservationModifier {
	return func(o *v1beta1.SubscriptionObservation) {
		o.ConfirmationWasAuthenticated = &s
	}
}

func TestGenerateSubscribeInput(t *testing.T) {
	cases := map[string]struct {
		in  v1beta1.SubscriptionParameters
		out sns.SubscribeInput
	}{
		"FilledInput": {
			in: v1beta1.SubscriptionParameters{
				TopicARN: topicArn,
				Endpoint: subEmailEndpoint,
				Protocol: subEmailProtocol,
			},
			out: sns.SubscribeInput{
				TopicArn:              pointer.ToOrNilIfZeroValue(topicArn),
				Endpoint:              &subEmailEndpoint,
				Protocol:              &subEmailProtocol,
				ReturnSubscriptionArn: subBoolTrue,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			input := GenerateSubscribeInput(&tc.in)
			if diff := cmp.Diff(input, &tc.out, cmpopts.IgnoreTypes(document.NoSerde{})); diff != "" {
				t.Errorf("GenerateSubscribeInput(...): -want, +got\n:%s", diff)
			}
		})
	}
}

func TestLateInitializeSubscription(t *testing.T) {
	type args struct {
		spec *v1beta1.SubscriptionParameters
		attr map[string]string
	}
	cases := map[string]struct {
		args args
		want *v1beta1.SubscriptionParameters
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
				spec: subParams(func(sub *v1beta1.SubscriptionParameters) {
					sub.TopicARN = topicArn
					sub.Protocol = subEmailProtocol
					sub.Endpoint = subEmailEndpoint
				}),
				attr: map[string]string{},
			},
			want: subParams(func(sub *v1beta1.SubscriptionParameters) {
				sub.TopicARN = topicArn
				sub.Endpoint = subEmailEndpoint
				sub.Protocol = subEmailProtocol
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
		p    v1beta1.SubscriptionParameters
		attr *map[string]string
	}

	cases := map[string]struct {
		args args
		want *map[string]string
	}{
		"NoChange": {
			args: args{
				p: v1beta1.SubscriptionParameters{
					Protocol: subEmailProtocol,
					Endpoint: subEmailEndpoint,
				},
				attr: subAttributes(),
			},
			want: subAttributes(),
		},
		"Change": {
			args: args{
				p: v1beta1.SubscriptionParameters{
					Protocol: subEmailProtocol,
					Endpoint: subEmailEndpoint,
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
		out *v1beta1.SubscriptionObservation
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
