package manualv1alpha1

import (
	"testing"

	awsec2 "github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"
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
							Key:   aws.String("name"),
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
								Key:   aws.String("name"),
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
							Key:   aws.String("name"),
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
		name     string
		tagSpecs []TagSpecification
	}
	cases := map[string]struct {
		args args
		want []awsec2.TagSpecification
	}{
		"BasicTagSpecification": {
			args: args{
				tagSpecs: []TagSpecification{
					{
						ResourceType: aws.String("capacity-reservation"),
						Tags: []Tag{
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
				tagSpecs: []TagSpecification{},
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
		tagSpecs []TagSpecification
	}
	cases := map[string]struct {
		args args
		want []awsec2.TagSpecification
	}{
		"TransformTagSpecificationWithoutInstanceName": {
			args: args{
				name: managedName,
				tagSpecs: []TagSpecification{
					{
						ResourceType: aws.String("capacity-reservation"),
						Tags: []Tag{
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
							Key:   aws.String("name"),
							Value: aws.String(managedName),
						},
					},
				},
			},
		},
		"TransformTagSpecificationWithInstanceName": {
			args: args{
				name: managedName,
				tagSpecs: []TagSpecification{
					{
						ResourceType: aws.String("capacity-reservation"),
						Tags: []Tag{
							{
								Key:   "test",
								Value: "test",
							},
						},
					},
					{
						ResourceType: aws.String("instance"),
						Tags: []Tag{
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
			},
		},
		"EmptyTagSpecification": {
			args: args{
				name:     managedName,
				tagSpecs: []TagSpecification{},
			},
			want: []awsec2.TagSpecification{
				{
					ResourceType: "instance",
					Tags: []awsec2.Tag{
						{
							Key:   aws.String("name"),
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
