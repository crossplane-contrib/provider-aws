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

package resourcepolicy

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	svcsdk "github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/cloudwatchlogs/v1alpha1"
)

const (
	testPolicyDocument = `{
		"Version": "2012-10-17",
		"Statement": [	
			{
				"Sid": "Route53LogsToCloudWatchLogs",
				"Effect": "Allow",
				"Principal": {
					"Service": "route53.amazonaws.com"
				},
				"Action": "logs:PutLogEvents",
				"Resource": "logArn"
			}
		]
	}`
	testPolicyDocument2 = `{
		"Version" : "2012-10-17",
		"Statement" : [
			{
				"Sid" : "",
				"Effect" : "Allow",
				"Principal" : {
					"Service" : "111122223333"
				},
				"Action" : "logs:PutSubscriptionFilter",
				"Resource" : "arn:aws:logs:us-east-1:123456789012:destination:testDestination"
			}
		]
	}`
	policyName      = "policyName"
	otherPolicyName = "otherPolicyName"
)

type args struct {
	describeResourcePoliciesOutput *svcsdk.DescribeResourcePoliciesOutput
	resourcePolicy                 *svcapitypes.ResourcePolicy
}

func TestIsUpToDate(t *testing.T) {
	type want struct {
		result bool
		_      string
		err    error
	}

	cases := map[string]struct {
		args args
		want want
	}{
		"SameFields": {
			args: args{
				describeResourcePoliciesOutput: &svcsdk.DescribeResourcePoliciesOutput{
					ResourcePolicies: []*svcsdk.ResourcePolicy{
						{
							PolicyName:     aws.String(policyName),
							PolicyDocument: aws.String(testPolicyDocument),
						},
						{
							PolicyName:     aws.String(otherPolicyName),
							PolicyDocument: aws.String(testPolicyDocument2),
						},
					},
				},
				resourcePolicy: &svcapitypes.ResourcePolicy{
					ObjectMeta: metav1.ObjectMeta{
						Name: policyName,
						Annotations: map[string]string{
							meta.AnnotationKeyExternalName: policyName,
						},
					},
					Spec: svcapitypes.ResourcePolicySpec{
						ForProvider: svcapitypes.ResourcePolicyParameters{
							PolicyDocument: aws.String(testPolicyDocument),
						},
					},
				},
			},
			want: want{
				result: true,
				err:    nil,
			},
		},
		"DifferentPolicyDocument": {
			args: args{
				describeResourcePoliciesOutput: &svcsdk.DescribeResourcePoliciesOutput{
					ResourcePolicies: []*svcsdk.ResourcePolicy{
						{
							PolicyName:     aws.String(policyName),
							PolicyDocument: aws.String(testPolicyDocument),
						},
					},
				},
				resourcePolicy: &svcapitypes.ResourcePolicy{
					ObjectMeta: metav1.ObjectMeta{
						Name: policyName,
						Annotations: map[string]string{
							meta.AnnotationKeyExternalName: policyName,
						},
					},
					Spec: svcapitypes.ResourcePolicySpec{
						ForProvider: svcapitypes.ResourcePolicyParameters{
							PolicyDocument: aws.String(testPolicyDocument2),
						},
					},
				},
			},
			want: want{
				result: false,
				err:    nil,
			},
		},
		"ResourcePolicyNotFound": {
			args: args{
				describeResourcePoliciesOutput: &svcsdk.DescribeResourcePoliciesOutput{
					ResourcePolicies: []*svcsdk.ResourcePolicy{
						{
							PolicyName:     aws.String(otherPolicyName),
							PolicyDocument: aws.String(testPolicyDocument2),
						},
					},
				},
				resourcePolicy: &svcapitypes.ResourcePolicy{
					ObjectMeta: metav1.ObjectMeta{
						Name: policyName,
						Annotations: map[string]string{
							meta.AnnotationKeyExternalName: policyName,
						},
					},
					Spec: svcapitypes.ResourcePolicySpec{
						ForProvider: svcapitypes.ResourcePolicyParameters{
							PolicyDocument: aws.String(testPolicyDocument),
						},
					},
				},
			},
			want: want{
				result: false,
				err:    nil,
			},
		},
		"ResourcePolicyNotFoundEmptyDescribeResourcePoliciesOutput": {
			args: args{
				describeResourcePoliciesOutput: &svcsdk.DescribeResourcePoliciesOutput{
					ResourcePolicies: []*svcsdk.ResourcePolicy{},
				},
				resourcePolicy: &svcapitypes.ResourcePolicy{
					ObjectMeta: metav1.ObjectMeta{
						Name: policyName,
						Annotations: map[string]string{
							meta.AnnotationKeyExternalName: policyName,
						},
					},
					Spec: svcapitypes.ResourcePolicySpec{
						ForProvider: svcapitypes.ResourcePolicyParameters{
							PolicyDocument: aws.String(testPolicyDocument),
						},
					},
				},
			},
			want: want{
				result: false,
				err:    nil,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {

			got, _, err := isUpToDate(context.Background(), tc.args.resourcePolicy, tc.args.describeResourcePoliciesOutput)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("isUpToDate(...): -want error, +got error: %s", diff)
			}
			if diff := cmp.Diff(tc.want.result, got); diff != "" {
				t.Errorf("isUpToDate(...): -want, +got: %s", diff)
			}
		})
	}
}

func TestFilterList(t *testing.T) {
	type want struct {
		resp *svcsdk.DescribeResourcePoliciesOutput
	}
	cases := map[string]struct {
		args args
		want want
	}{
		"ResourcePolicyFound": {
			args: args{
				describeResourcePoliciesOutput: &svcsdk.DescribeResourcePoliciesOutput{
					ResourcePolicies: []*svcsdk.ResourcePolicy{
						{
							PolicyName: aws.String(policyName),
						},
						{
							PolicyName: aws.String(otherPolicyName),
						},
					},
				},
				resourcePolicy: &svcapitypes.ResourcePolicy{
					ObjectMeta: metav1.ObjectMeta{
						Name: policyName,
						Annotations: map[string]string{
							meta.AnnotationKeyExternalName: policyName,
						},
					},
				},
			},
			want: want{
				resp: &svcsdk.DescribeResourcePoliciesOutput{
					ResourcePolicies: []*svcsdk.ResourcePolicy{
						{
							PolicyName: aws.String(policyName),
						},
					},
				},
			},
		},
		"ResourcePolicyNotFound": {
			args: args{
				describeResourcePoliciesOutput: &svcsdk.DescribeResourcePoliciesOutput{
					ResourcePolicies: []*svcsdk.ResourcePolicy{
						{
							PolicyName: aws.String(otherPolicyName),
						},
					},
				},
				resourcePolicy: &svcapitypes.ResourcePolicy{
					ObjectMeta: metav1.ObjectMeta{
						Name: policyName,
						Annotations: map[string]string{
							meta.AnnotationKeyExternalName: policyName,
						},
					},
				},
			},
			want: want{
				resp: &svcsdk.DescribeResourcePoliciesOutput{},
			},
		},
		"ResourcePolicyNotFoundEmptyDescribeResourcePoliciesOutput": {
			args: args{
				describeResourcePoliciesOutput: &svcsdk.DescribeResourcePoliciesOutput{
					ResourcePolicies: []*svcsdk.ResourcePolicy{},
				},
				resourcePolicy: &svcapitypes.ResourcePolicy{
					ObjectMeta: metav1.ObjectMeta{
						Name: policyName,
						Annotations: map[string]string{
							meta.AnnotationKeyExternalName: policyName,
						},
					},
				},
			},
			want: want{
				resp: &svcsdk.DescribeResourcePoliciesOutput{},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			resp := filterList(tc.args.resourcePolicy, tc.args.describeResourcePoliciesOutput)
			if diff := cmp.Diff(tc.want.resp, resp); diff != "" {
				t.Errorf("filterList(...): -want, +got: %s", diff)
			}
		})
	}
}
