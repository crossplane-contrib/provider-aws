/*
Copyright 2020 The Crossplane Authors.

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
	"testing"

	svcsdk "github.com/aws/aws-sdk-go/service/sns"
	"github.com/google/go-cmp/cmp"

	"github.com/crossplane/provider-aws/apis/sns/v1alpha1"
	aws "github.com/crossplane/provider-aws/pkg/clients"
)

var (
	topicName         = "some-topic"
	topicDisplayName  = "some-topic-01"
	topicDisplayName2 = "some-topic-02"
)

// Topic Attribute Modifier
type topicAttrModifier func(map[string]*string)

func topicAttributes(m ...topicAttrModifier) map[string]*string {
	attr := map[string]*string{}

	for _, f := range m {
		f(attr)
	}

	return attr
}

func withAttrDisplayName(s *string) topicAttrModifier {
	return func(attr map[string]*string) {
		attr[string(v1alpha1.TopicDisplayName)] = s
	}
}

func TestGetChangedAttributes(t *testing.T) {

	type args struct {
		p    v1alpha1.TopicParameters
		attr map[string]*string
	}

	cases := map[string]struct {
		args args
		want map[string]*string
	}{
		"NoChange": {
			args: args{
				p: v1alpha1.TopicParameters{
					Name:        &topicName,
					DisplayName: &topicDisplayName,
				},
				attr: map[string]*string{
					string(v1alpha1.TopicDisplayName): aws.String(topicDisplayName),
				},
			},
			want: topicAttributes(),
		},
		"Change": {
			args: args{
				p: v1alpha1.TopicParameters{
					Name:        &topicName,
					DisplayName: &topicDisplayName,
				},
				attr: topicAttributes(
					withAttrDisplayName(&topicDisplayName2),
				),
			},
			want: topicAttributes(
				withAttrDisplayName(&topicDisplayName),
			),
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			c := GetChangedAttributes(tc.args.p, tc.args.attr)
			if diff := cmp.Diff(tc.want, c); diff != "" {
				t.Errorf("GetChangedAttributes(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestIsSNSTopicUpToDate(t *testing.T) {
	type args struct {
		p    *v1alpha1.Topic
		attr map[string]*string
	}

	cases := map[string]struct {
		args args
		want bool
	}{
		"SameFieldsAndAllFilled": {
			args: args{
				attr: topicAttributes(
					withAttrDisplayName(&topicDisplayName),
				),
				p: &v1alpha1.Topic{
					Spec: v1alpha1.TopicSpec{
						ForProvider: v1alpha1.TopicParameters{
							DisplayName: &topicDisplayName,
						},
					},
				},
			},
			want: true,
		},
		"DifferentFields": {
			args: args{
				attr: topicAttributes(
					withAttrDisplayName(&topicDisplayName),
				),
				p: &v1alpha1.Topic{},
			},
			want: false,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got, _ := isUpToDate(true, tc.args.p, &svcsdk.GetTopicAttributesOutput{Attributes: tc.args.attr})
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("Topic : -want, +got:\n%s", diff)
			}
		})
	}
}
