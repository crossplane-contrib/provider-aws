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
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/google/go-cmp/cmp"

	"github.com/crossplane/provider-aws/apis/applicationintegration/v1alpha1"
)

var (
	delaySeconds    int64             = 30
	kmsMasterKeyID                    = "someId"
	tagKey                            = "k"
	tagValue                          = "v"
	arn                               = "arn"
	maxReceiveCount int64             = 5
	m               map[string]string = make(map[string]string)
)

func sqsParams(m ...func(*v1alpha1.QueueParameters)) *v1alpha1.QueueParameters {
	o := &v1alpha1.QueueParameters{
		DelaySeconds:   aws.Int64(delaySeconds),
		KMSMasterKeyID: aws.String(kmsMasterKeyID),
	}

	for _, f := range m {
		f(o)
	}

	return o
}

func attributes(p ...map[string]string) map[string]string {
	m := map[string]string{}

	m[v1alpha1.AttributeDelaySeconds] = strconv.FormatInt(delaySeconds, 10)

	for _, item := range p {
		for k, v := range item {
			m[k] = v
		}
	}

	return m
}

func TestLateInitialize(t *testing.T) {
	type args struct {
		spec *v1alpha1.QueueParameters
		in   map[string]string
		tags map[string]string
	}
	cases := map[string]struct {
		args args
		want *v1alpha1.QueueParameters
	}{
		"AllFilledNoDiff": {
			args: args{
				spec: sqsParams(),
				in:   attributes(),
			},
			want: sqsParams(),
		},
		"AllFilledExternalDiff": {
			args: args{
				spec: sqsParams(),
				in: attributes(map[string]string{
					v1alpha1.AttributeKmsMasterKeyID: kmsMasterKeyID,
				}),
			},
			want: sqsParams(),
		},
		"PartialFilled": {
			args: args{
				spec: sqsParams(func(p *v1alpha1.QueueParameters) {
					p.DelaySeconds = nil
				}),
				in: attributes(),
			},
			want: sqsParams(),
		},
		"PointerFields": {
			args: args{
				spec: sqsParams(),
				tags: map[string]string{
					tagKey: tagValue,
				},
			},
			want: sqsParams(func(p *v1alpha1.QueueParameters) {
				p.Tags = []v1alpha1.Tag{
					{
						Key:   tagKey,
						Value: aws.String(tagValue),
					},
				}
			}),
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			LateInitialize(tc.args.spec, tc.args.in, tc.args.tags)
			if diff := cmp.Diff(tc.want, tc.args.spec); diff != "" {
				t.Errorf("LateInitializeSpec(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestIsUpToDate(t *testing.T) {
	type args struct {
		p          v1alpha1.QueueParameters
		attributes map[string]string
		tags       map[string]string
	}

	cases := map[string]struct {
		args args
		want bool
	}{
		"SameFields": {
			args: args{
				p: v1alpha1.QueueParameters{
					KMSMasterKeyID: &kmsMasterKeyID,
				},
				attributes: map[string]string{
					v1alpha1.AttributeKmsMasterKeyID: kmsMasterKeyID,
				},
			},
			want: true,
		},
		"DifferentFields": {
			args: args{
				p: v1alpha1.QueueParameters{
					KMSMasterKeyID: &kmsMasterKeyID,
				},
				attributes: map[string]string{},
			},
			want: false,
		},
		"Tags": {
			args: args{
				p: v1alpha1.QueueParameters{
					Tags: []v1alpha1.Tag{
						{
							Key:   tagKey,
							Value: aws.String(tagValue),
						},
					},
				},
				tags: map[string]string{
					tagKey: tagValue,
				},
			},
			want: true,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := IsUpToDate(tc.args.p, tc.args.attributes, tc.args.tags)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGenerateQueueAttributes(t *testing.T) {
	cases := map[string]struct {
		in  v1alpha1.QueueParameters
		out map[string]string
	}{
		"FilledInput": {
			in: *sqsParams(),
			out: map[string]string{
				v1alpha1.AttributeDelaySeconds:   strconv.FormatInt(delaySeconds, 10),
				v1alpha1.AttributeKmsMasterKeyID: kmsMasterKeyID,
			},
		},
		"RedrivePolicy": {
			in: *sqsParams(func(p *v1alpha1.QueueParameters) {
				p.RedrivePolicy = &v1alpha1.RedrivePolicy{
					DeadLetterQueueARN: &arn,
					MaxReceiveCount:    &maxReceiveCount,
				}
			}),
			out: map[string]string{
				v1alpha1.AttributeDelaySeconds:   strconv.FormatInt(delaySeconds, 10),
				v1alpha1.AttributeRedrivePolicy:  `{"deadLetterQueueARN":"arn","maxReceiveCount":5}`,
				v1alpha1.AttributeKmsMasterKeyID: kmsMasterKeyID,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			r := GenerateQueueAttributes(&tc.in)
			if diff := cmp.Diff(r, tc.out); diff != "" {
				t.Errorf("GenerateNetworkObservation(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGenerateQueueTags(t *testing.T) {
	cases := map[string]struct {
		in  v1alpha1.QueueParameters
		out map[string]string
	}{
		"FilledInput": {
			in: *sqsParams(func(p *v1alpha1.QueueParameters) {
				p.Tags = []v1alpha1.Tag{
					{
						Key:   tagKey,
						Value: aws.String(tagValue),
					},
				}
			}),
			out: map[string]string{
				tagKey: tagValue,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			r := GenerateQueueTags(tc.in.Tags)
			if diff := cmp.Diff(r, tc.out); diff != "" {
				t.Errorf("GenerateNetworkObservation(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestTagsDiff(t *testing.T) {
	type args struct {
		specTags []v1alpha1.Tag
		sqsTags  map[string]string
	}

	type want struct {
		removed, added map[string]string
	}

	cases := map[string]struct {
		args args
		want want
	}{
		"SameTags": {
			args: args{
				specTags: []v1alpha1.Tag{
					{
						Key:   "k1",
						Value: aws.String("v1"),
					},
					{
						Key:   "k2",
						Value: aws.String("v2"),
					},
				},
				sqsTags: map[string]string{
					"k1": "v1",
					"k2": "v2",
				},
			},
			want: want{
				removed: m,
				added:   m,
			},
		},
		"RemovedTags": {
			args: args{
				specTags: []v1alpha1.Tag{
					{
						Key:   "k2",
						Value: aws.String("v2"),
					},
				},
				sqsTags: map[string]string{
					"k1": "v1",
					"k2": "v2",
				},
			},
			want: want{
				removed: map[string]string{
					"k1": "v1",
				},
				added: m,
			},
		},
		"AddedTags": {
			args: args{
				specTags: []v1alpha1.Tag{
					{
						Key:   "k1",
						Value: aws.String("v1"),
					},
					{
						Key:   "k2",
						Value: aws.String("v2"),
					},
					{
						Key:   "k3",
						Value: aws.String("v3"),
					},
				},
				sqsTags: map[string]string{
					"k1": "v1",
					"k2": "v2",
				},
			},
			want: want{
				removed: m,
				added: map[string]string{
					"k3": "v3",
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			removed, added := TagsDiff(tc.args.sqsTags, tc.args.specTags)
			if diff := cmp.Diff(tc.want.removed, removed); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.added, added); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}
