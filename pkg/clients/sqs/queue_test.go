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
	_ "embed"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"

	"github.com/crossplane-contrib/provider-aws/apis/sqs/v1beta1"
)

var (
	delaySeconds    int64             = 30
	kmsMasterKeyID                    = "someId"
	tagKey                            = "k"
	tagValue                          = "v"
	arn                               = "arn"
	url                               = "url"
	maxReceiveCount int64             = 5
	m               map[string]string = make(map[string]string)
)

func sqsParams(m ...func(*v1beta1.QueueParameters)) *v1beta1.QueueParameters {
	o := &v1beta1.QueueParameters{
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

	m[v1beta1.AttributeDelaySeconds] = strconv.FormatInt(delaySeconds, 10)

	for _, item := range p {
		for k, v := range item {
			m[k] = v
		}
	}

	return m
}

func TestLateInitialize(t *testing.T) {
	type args struct {
		spec *v1beta1.QueueParameters
		in   map[string]string
		tags map[string]string
	}
	cases := map[string]struct {
		args args
		want *v1beta1.QueueParameters
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
					v1beta1.AttributeKmsMasterKeyID: kmsMasterKeyID,
				}),
			},
			want: sqsParams(),
		},
		"PartialFilled": {
			args: args{
				spec: sqsParams(func(p *v1beta1.QueueParameters) {
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
			want: sqsParams(func(p *v1beta1.QueueParameters) {
				p.Tags = map[string]string{
					tagKey: tagValue,
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

var (
	//go:embed testdata/queue_policy.min.json
	testPolicyRawMin string

	//go:embed testdata/queue_policy.json
	testPolicyRaw string

	//go:embed testdata/queue_policy2.json
	testPolicy2Raw string
)

func TestIsUpToDate(t *testing.T) {
	type args struct {
		p          v1beta1.QueueParameters
		attributes map[string]string
		tags       map[string]string
	}
	type want struct {
		isUpToDate bool
		err        error
	}

	cases := map[string]struct {
		args args
		want want
	}{
		"SameFields": {
			args: args{
				p: v1beta1.QueueParameters{
					KMSMasterKeyID: &kmsMasterKeyID,
				},
				attributes: map[string]string{
					v1beta1.AttributeKmsMasterKeyID: kmsMasterKeyID,
				},
			},
			want: want{
				isUpToDate: true,
			},
		},
		"DifferentFields": {
			args: args{
				p: v1beta1.QueueParameters{
					KMSMasterKeyID: &kmsMasterKeyID,
				},
				attributes: map[string]string{},
			},
			want: want{
				isUpToDate: false,
			},
		},
		"SamePolicy": {
			args: args{
				p: v1beta1.QueueParameters{
					Policy: &testPolicyRaw,
				},
				attributes: map[string]string{
					v1beta1.AttributePolicy: testPolicyRawMin,
				},
			},
			want: want{
				isUpToDate: true,
			},
		},
		"DifferentPolicy": {
			args: args{
				p: v1beta1.QueueParameters{
					Policy: &testPolicy2Raw,
				},
				attributes: map[string]string{
					v1beta1.AttributePolicy: testPolicyRawMin,
				},
			},
			want: want{
				isUpToDate: false,
			},
		},
		"Tags": {
			args: args{
				p: v1beta1.QueueParameters{
					Tags: map[string]string{
						tagKey: tagValue,
					},
				},
				tags: map[string]string{
					tagKey: tagValue,
				},
			},
			want: want{
				isUpToDate: true,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			isUpToDate, _, err := IsUpToDate(tc.args.p, tc.args.attributes, tc.args.tags)
			if diff := cmp.Diff(tc.want.isUpToDate, isUpToDate); diff != "" {
				t.Errorf("isUpToDate: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("error: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGenerateQueueAttributes(t *testing.T) {
	cases := map[string]struct {
		in  v1beta1.QueueParameters
		out map[string]string
	}{
		"FilledInput": {
			in: *sqsParams(),
			out: map[string]string{
				v1beta1.AttributeDelaySeconds:   strconv.FormatInt(delaySeconds, 10),
				v1beta1.AttributeKmsMasterKeyID: kmsMasterKeyID,
			},
		},
		"RedrivePolicy": {
			in: *sqsParams(func(p *v1beta1.QueueParameters) {
				p.RedrivePolicy = &v1beta1.RedrivePolicy{
					DeadLetterTargetARN: &arn,
					MaxReceiveCount:     maxReceiveCount,
				}
			}),
			out: map[string]string{
				v1beta1.AttributeDelaySeconds:   strconv.FormatInt(delaySeconds, 10),
				v1beta1.AttributeRedrivePolicy:  `{"deadLetterTargetArn":"arn","maxReceiveCount":5}`,
				v1beta1.AttributeKmsMasterKeyID: kmsMasterKeyID,
			},
		},
		"EmptyInput": {
			in:  v1beta1.QueueParameters{},
			out: nil,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			r := GenerateQueueAttributes(&tc.in)
			if diff := cmp.Diff(r, tc.out); diff != "" {
				t.Errorf("GenerateQueueAttributes(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGenerateCreateAttributes(t *testing.T) {
	cases := map[string]struct {
		in  v1beta1.QueueParameters
		out map[string]string
	}{
		"EmptyInput": {
			in:  v1beta1.QueueParameters{},
			out: nil,
		},
		"FifoQueueFalseShouldNotBeSent": {
			in: v1beta1.QueueParameters{
				FIFOQueue: aws.Bool(false),
			},
			out: nil,
		},
		"FifoQueueTrue": {
			in: v1beta1.QueueParameters{
				FIFOQueue: aws.Bool(true),
			},
			out: map[string]string{
				v1beta1.AttributeFifoQueue: "true",
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			r := GenerateCreateAttributes(&tc.in)
			if diff := cmp.Diff(r, tc.out); diff != "" {
				t.Errorf("GenerateCreateAttributes(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestTagsDiff(t *testing.T) {
	type args struct {
		specTags map[string]string
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
				specTags: map[string]string{
					"k1": "v1",
					"k2": "v2",
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
				specTags: map[string]string{
					"k2": "v2",
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
				specTags: map[string]string{
					"k1": "v1",
					"k2": "v2",
					"k3": "v3",
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

func TestGetConnectionDetails(t *testing.T) {
	cases := map[string]struct {
		queue v1beta1.Queue
		want  managed.ConnectionDetails
	}{
		"ValidInstance": {
			queue: v1beta1.Queue{
				Status: v1beta1.QueueStatus{
					AtProvider: v1beta1.QueueObservation{
						URL: url,
					},
				},
			},
			want: managed.ConnectionDetails{
				xpv1.ResourceCredentialsSecretEndpointKey: []byte(url),
			},
		},
		"NilInstance": {
			queue: v1beta1.Queue{},
			want:  nil,
		}}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := GetConnectionDetails(tc.queue)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}
