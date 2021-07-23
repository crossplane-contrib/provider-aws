/*
Copyright 2021 The Crossplane Authors.
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

package ec2

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsec2 "github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/google/go-cmp/cmp"

	"github.com/crossplane/crossplane-runtime/pkg/test"

	"github.com/crossplane/provider-aws/apis/ec2/manualv1alpha1"
)

const (
	managedName = "sample-instance"
)

func TestInjectInstanceNameTagSpecification(t *testing.T) {
	type args struct {
		name     string
		tagSpecs []awsec2.TagSpecification
	}
	cases := map[string]struct {
		args args
		want []awsec2.TagSpecification
	}{
		"InstanceNameNotSupplied": {
			args: args{
				name: managedName,
				tagSpecs: []awsec2.TagSpecification{
					{
						ResourceType: "capacity-reservation",
						Tags: []awsec2.Tag{
							{
								Key:   aws.String("test"),
								Value: aws.String("test"),
							},
						},
					},
				},
			},
			want: []awsec2.TagSpecification{
				{
					ResourceType: "capacity-reservation",
					Tags: []awsec2.Tag{
						{
							Key:   aws.String("test"),
							Value: aws.String("test"),
						},
					},
				},
				{
					ResourceType: "instance",
					Tags: []awsec2.Tag{
						{
							Key:   aws.String("Name"),
							Value: aws.String(managedName),
						},
					},
				},
			},
		},
		"InstanceNameSupplied": {
			args: args{
				name: managedName,
				tagSpecs: []awsec2.TagSpecification{
					{
						ResourceType: "capacity-reservation",
						Tags: []awsec2.Tag{
							{
								Key:   aws.String("test"),
								Value: aws.String("test"),
							},
						},
					},
					{
						ResourceType: "instance",
						Tags: []awsec2.Tag{
							{
								Key:   aws.String("Name"),
								Value: aws.String("test"),
							},
						},
					},
				},
			},
			want: []awsec2.TagSpecification{
				{
					ResourceType: "capacity-reservation",
					Tags: []awsec2.Tag{
						{
							Key:   aws.String("test"),
							Value: aws.String("test"),
						},
					},
				},
				{
					ResourceType: "instance",
					Tags: []awsec2.Tag{
						{
							Key:   aws.String("Name"),
							Value: aws.String("test"),
						},
					},
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			tags := injectInstanceNameTagSpecification(tc.args.name, tc.args.tagSpecs)

			if diff := cmp.Diff(tc.want, tags, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGenerateEC2TagSpecifications(t *testing.T) {
	type args struct {
		tagSpecs []manualv1alpha1.TagSpecification
	}
	cases := map[string]struct {
		args args
		want []awsec2.TagSpecification
	}{
		"BasicTagSpecification": {
			args: args{
				tagSpecs: []manualv1alpha1.TagSpecification{
					{
						ResourceType: aws.String("capacity-reservation"),
						Tags: []manualv1alpha1.Tag{
							{
								Key:   "test",
								Value: "test",
							},
						},
					},
				},
			},
			want: []awsec2.TagSpecification{
				{
					ResourceType: "capacity-reservation",
					Tags: []awsec2.Tag{
						{
							Key:   aws.String("test"),
							Value: aws.String("test"),
						},
					},
				},
			},
		},
		"EmptyTagSpecification": {
			args: args{
				tagSpecs: []manualv1alpha1.TagSpecification{},
			},
			want: []awsec2.TagSpecification{},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			tags := generateEC2TagSpecifications(tc.args.tagSpecs)

			if diff := cmp.Diff(tc.want, tags, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestTransformTagSpecifications(t *testing.T) {
	type args struct {
		name     string
		tagSpecs []manualv1alpha1.TagSpecification
	}
	cases := map[string]struct {
		args args
		want []awsec2.TagSpecification
	}{
		"TransformTagSpecificationWithoutInstanceName": {
			args: args{
				name: managedName,
				tagSpecs: []manualv1alpha1.TagSpecification{
					{
						ResourceType: aws.String("capacity-reservation"),
						Tags: []manualv1alpha1.Tag{
							{
								Key:   "test",
								Value: "test",
							},
						},
					},
				},
			},
			want: []awsec2.TagSpecification{
				{
					ResourceType: "capacity-reservation",
					Tags: []awsec2.Tag{
						{
							Key:   aws.String("test"),
							Value: aws.String("test"),
						},
					},
				},
				{
					ResourceType: "instance",
					Tags: []awsec2.Tag{
						{
							Key:   aws.String("Name"),
							Value: aws.String(managedName),
						},
					},
				},
			},
		},
		"TransformTagSpecificationWithInstanceName": {
			args: args{
				name: managedName,
				tagSpecs: []manualv1alpha1.TagSpecification{
					{
						ResourceType: aws.String("capacity-reservation"),
						Tags: []manualv1alpha1.Tag{
							{
								Key:   "test",
								Value: "test",
							},
						},
					},
					{
						ResourceType: aws.String("instance"),
						Tags: []manualv1alpha1.Tag{
							{
								Key:   "name",
								Value: "test",
							},
						},
					},
				},
			},
			want: []awsec2.TagSpecification{
				{
					ResourceType: "capacity-reservation",
					Tags: []awsec2.Tag{
						{
							Key:   aws.String("test"),
							Value: aws.String("test"),
						},
					},
				},
				{
					ResourceType: "instance",
					Tags: []awsec2.Tag{
						{
							Key:   aws.String("name"),
							Value: aws.String("test"),
						},
					},
				},
				{
					ResourceType: "instance",
					Tags: []awsec2.Tag{
						{
							Key:   aws.String("Name"),
							Value: aws.String(managedName),
						},
					},
				},
			},
		},
		"EmptyTagSpecification": {
			args: args{
				name:     managedName,
				tagSpecs: []manualv1alpha1.TagSpecification{},
			},
			want: []awsec2.TagSpecification{
				{
					ResourceType: "instance",
					Tags: []awsec2.Tag{
						{
							Key:   aws.String("Name"),
							Value: aws.String(managedName),
						},
					},
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			tags := TransformTagSpecifications(managedName, tc.args.tagSpecs)

			if diff := cmp.Diff(tc.want, tags, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}
