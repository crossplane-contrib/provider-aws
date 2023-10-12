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
	_ "embed"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awssns "github.com/aws/aws-sdk-go-v2/service/sns"
	awssnstypes "github.com/aws/aws-sdk-go-v2/service/sns/types"
	"github.com/aws/smithy-go/document"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/crossplane-contrib/provider-aws/apis/sns/v1beta1"
)

var (
	empty             = ""
	falseFlag         = false
	trueFlag          = true
	topicName         = "some-topic"
	topicDisplayName  = "some-topic-01"
	topicDisplayName2 = "some-topic-02"
	topicArn          = "sometopicArn"
	topicFifo         = true
	confirmedSubs     = "1"
	pendingSubs       = "11"
	deletedSubs       = "12"
	tagKey1           = "name-1"
	tagValue1         = "value-1"
	tagKey2           = "name-2"
	tagValue2         = "value-2"

	//go:embed testdata/policy_a.json
	testPolicyA string

	//go:embed testdata/policy_a2.json
	testPolicyA2 string

	//go:embed testdata/policy_b.json
	testPolicyB string
)

// Topic Attribute Modifier
type topicAttrModifier func(map[string]string)

func topicAttributes(m ...topicAttrModifier) map[string]string {
	attr := map[string]string{}

	for _, f := range m {
		f(attr)
	}

	return attr
}

func withOwner(s *string) topicAttrModifier {
	return func(attr map[string]string) {
		attr[string(TopicOwner)] = *s
	}
}

func withARN(s *string) topicAttrModifier {
	return func(attr map[string]string) {
		attr[string(TopicARN)] = *s
	}
}

func withTopicSubs(confirmed, pending, deleted string) topicAttrModifier {
	return func(attr map[string]string) {
		attr[string(TopicSubscriptionsConfirmed)] = confirmed
		attr[string(TopicSubscriptionsPending)] = pending
		attr[string(TopicSubscriptionsDeleted)] = deleted
	}
}

func withAttrDisplayName(s *string) topicAttrModifier {
	return func(attr map[string]string) {
		attr[string(TopicDisplayName)] = *s
	}
}

func withAttrFifoTopic(b *bool) topicAttrModifier {
	return func(attr map[string]string) {
		attr[string(TopicFifoTopic)] = strconv.FormatBool(*b)
	}
}

func withAttrPolicy(s *string) topicAttrModifier {
	return func(attr map[string]string) {
		attr[string(TopicPolicy)] = *s
	}
}

// topic Observation Modifier
type topicObservationModifier func(*v1beta1.TopicObservation)

func topicObservation(m ...func(*v1beta1.TopicObservation)) *v1beta1.TopicObservation {
	o := &v1beta1.TopicObservation{}

	for _, f := range m {
		f(o)
	}

	return o
}

func withObservationOwner(s *string) topicObservationModifier {
	return func(o *v1beta1.TopicObservation) {
		o.Owner = s
	}
}

func withObservationARN(s string) topicObservationModifier {
	return func(o *v1beta1.TopicObservation) {
		o.ARN = s
	}
}

func withObservationSubs(confirmed, pending, deleted string) topicObservationModifier {
	return func(o *v1beta1.TopicObservation) {
		if s, err := strconv.ParseInt(confirmed, 10, 64); err == nil {
			n := &s
			o.ConfirmedSubscriptions = n
		}
		if s, err := strconv.ParseInt(pending, 10, 64); err == nil {
			n := &s
			o.PendingSubscriptions = n
		}
		if s, err := strconv.ParseInt(deleted, 10, 64); err == nil {
			n := &s
			o.DeletedSubscriptions = n
		}
	}
}

// topic Parameters
func topicParams(m ...func(*v1beta1.TopicParameters)) *v1beta1.TopicParameters {
	o := &v1beta1.TopicParameters{
		Name:        *aws.String(topicName),
		DisplayName: aws.String(topicDisplayName),
		Tags: []v1beta1.Tag{
			{Key: tagKey1, Value: &tagValue1},
			{Key: tagKey2, Value: &tagValue2},
		},
		KMSMasterKeyID: &empty,
		Policy:         &empty,
		DeliveryPolicy: &empty,
		FifoTopic:      &trueFlag,
	}

	for _, f := range m {
		f(o)
	}

	return o
}

// Test Cases

func TestGenerateCreateTopicInput(t *testing.T) {
	cases := map[string]struct {
		in  v1beta1.TopicParameters
		out awssns.CreateTopicInput
	}{
		"FilledInput": {
			in: *topicParams(),
			out: awssns.CreateTopicInput{
				Name:       aws.String(topicName),
				Attributes: map[string]string{"FifoTopic": "true"},
				Tags: []awssnstypes.Tag{
					{Key: aws.String(tagKey1), Value: aws.String(tagValue1)},
					{Key: aws.String(tagKey2), Value: aws.String(tagValue2)},
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			input := GenerateCreateTopicInput(&tc.in)
			if diff := cmp.Diff(input, &tc.out, cmpopts.IgnoreTypes(document.NoSerde{})); diff != "" {
				t.Errorf("GenerateCreateTopicInput(...): -want, +got\n:%s", diff)
			}
		})
	}
}

func TestGetChangedAttributes(t *testing.T) {
	type args struct {
		p    v1beta1.TopicParameters
		attr map[string]string
	}

	type want struct {
		attrs map[string]string
		err   error
	}

	cases := map[string]struct {
		args args
		want want
	}{
		"NoChange": {
			args: args{
				p: v1beta1.TopicParameters{
					Name:        topicName,
					DisplayName: &topicDisplayName,
				},
				attr: topicAttributes(
					withAttrDisplayName(&topicDisplayName),
				),
			},
			want: want{
				attrs: topicAttributes(),
			},
		},
		"Change": {
			args: args{
				p: v1beta1.TopicParameters{
					Name:        topicName,
					DisplayName: &topicDisplayName,
				},
				attr: topicAttributes(
					withAttrDisplayName(&topicDisplayName2),
				),
			},
			want: want{
				attrs: topicAttributes(
					withAttrDisplayName(&topicDisplayName),
				),
			},
		},
		"ChangeFifo": {
			args: args{
				p: v1beta1.TopicParameters{
					Name:      topicName,
					FifoTopic: &trueFlag,
				},
				attr: topicAttributes(
					withAttrFifoTopic(&falseFlag),
				),
			},
			want: want{
				attrs: topicAttributes(
					withAttrFifoTopic(&trueFlag),
				),
			},
		},
		"SamePolicyButDifferentFormat": {
			args: args{
				p: v1beta1.TopicParameters{
					Policy: &testPolicyA,
				},
				attr: topicAttributes(
					withAttrPolicy(&testPolicyA2),
				),
			},
			want: want{
				attrs: topicAttributes(),
			},
		},
		"ChangedPolicy": {
			args: args{
				p: v1beta1.TopicParameters{
					Policy: &testPolicyA,
				},
				attr: topicAttributes(
					withAttrPolicy(&testPolicyB),
				),
			},
			want: want{
				attrs: topicAttributes(
					withAttrPolicy(&testPolicyA),
				),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			attr, err := GetChangedAttributes(tc.args.p, tc.args.attr)
			if diff := cmp.Diff(tc.want.attrs, attr); diff != "" {
				t.Errorf("GetChangedAttributes(...): -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("err: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGenerateTopicObservation(t *testing.T) {
	cases := map[string]struct {
		in  map[string]string
		out *v1beta1.TopicObservation
	}{
		"AllFilled": {
			in: topicAttributes(
				withOwner(&subOwner),
				withTopicSubs(confirmedSubs, pendingSubs, deletedSubs),
				withARN(&topicArn),
			),
			out: topicObservation(
				withObservationOwner(&subOwner),
				withObservationSubs(confirmedSubs, pendingSubs, deletedSubs),
				withObservationARN(topicArn),
			),
		},
		"NoSubscriptions": {
			in: topicAttributes(
				withOwner(&subOwner),
				withARN(&topicArn),
			),
			out: topicObservation(
				withObservationOwner(&subOwner),
				withObservationARN(topicArn),
			),
		},
		"Empty": {
			in: topicAttributes(),
			out: topicObservation(
				withObservationOwner(&empty),
				withObservationARN(empty),
			),
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			observation := GenerateTopicObservation(tc.in)
			if diff := cmp.Diff(*tc.out, observation); diff != "" {
				t.Errorf("GenerateTopicObservation(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestIsSNSTopicUpToDate(t *testing.T) {
	type args struct {
		p    v1beta1.TopicParameters
		attr map[string]string
	}
	type want struct {
		isUpToDate bool
		err        error
	}

	cases := map[string]struct {
		args args
		want want
	}{
		"SameFieldsAndAllFilled": {
			args: args{
				attr: topicAttributes(
					withAttrDisplayName(&topicDisplayName),
					withAttrFifoTopic(&topicFifo),
				),
				p: v1beta1.TopicParameters{
					DisplayName: &topicDisplayName,
					FifoTopic:   &topicFifo,
				},
			},
			want: want{
				isUpToDate: true,
			},
		},
		"DifferentFields": {
			args: args{
				attr: topicAttributes(
					withAttrDisplayName(&topicDisplayName),
				),
				p: v1beta1.TopicParameters{},
			},
			want: want{
				isUpToDate: false,
			},
		},
		"UpdateOnDifferentPolicy": {
			args: args{
				attr: topicAttributes(
					withAttrPolicy(&testPolicyA),
				),
				p: v1beta1.TopicParameters{
					Policy: &testPolicyB,
				},
			},
			want: want{
				isUpToDate: false,
			},
		},
		"NoUpdateExistsWithshuffledPolicy": {
			args: args{
				attr: topicAttributes(
					withAttrPolicy(&testPolicyA),
				),
				p: v1beta1.TopicParameters{
					Policy: &testPolicyA2,
				},
			},
			want: want{
				isUpToDate: true,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			isUpToDate, err := IsSNSTopicUpToDate(tc.args.p, tc.args.attr)
			if diff := cmp.Diff(tc.want.isUpToDate, isUpToDate); diff != "" {
				t.Errorf("isUpToDate: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("err: -want, +got:\n%s", diff)
			}
		})
	}
}
