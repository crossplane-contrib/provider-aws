/*
Copyright 2023 The Crossplane Authors.

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

package vpcendpoint

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/google/go-cmp/cmp"

	"github.com/crossplane-contrib/provider-aws/apis/ec2/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/ec2/fake"
)

func TestGenerateVPCEndpointTagSpecifications(t *testing.T) {
	cases := map[string]struct {
		tags map[string]string
		want []*ec2.TagSpecification
	}{
		"Empty": {
			tags: map[string]string{},
			want: []*ec2.TagSpecification{{ResourceType: aws.String(vpcEndpointTagResource), Tags: []*ec2.Tag{}}},
		},
		"SingleTag": {
			tags: map[string]string{"Name": "my-endpoint"},
			want: []*ec2.TagSpecification{{ResourceType: aws.String(vpcEndpointTagResource), Tags: []*ec2.Tag{{Key: aws.String("Name"), Value: aws.String("my-endpoint")}}}},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			if diff := cmp.Diff(tc.want, generateVPCEndpointTagSpecifications(tc.tags)); diff != "" {
				t.Errorf("generateVPCEndpointTagSpecifications() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestDiffVPCEndpointTags(t *testing.T) {
	cases := map[string]struct {
		spec       map[string]string
		current    []*ec2.Tag
		wantAdd    []*ec2.Tag
		wantRemove []*ec2.Tag
	}{
		"NoChange": {
			spec:       map[string]string{"Name": "my-endpoint"},
			current:    []*ec2.Tag{{Key: aws.String("Name"), Value: aws.String("my-endpoint")}},
			wantAdd:    nil,
			wantRemove: nil,
		},
		"AddTag": {
			spec:       map[string]string{"Name": "my-endpoint"},
			current:    nil,
			wantAdd:    []*ec2.Tag{{Key: aws.String("Name"), Value: aws.String("my-endpoint")}},
			wantRemove: nil,
		},
		"RemoveTag": {
			spec:       map[string]string{},
			current:    []*ec2.Tag{{Key: aws.String("Old"), Value: aws.String("gone")}},
			wantAdd:    nil,
			wantRemove: []*ec2.Tag{{Key: aws.String("Old"), Value: aws.String("gone")}},
		},
		"ValueChanged": {
			spec:       map[string]string{"Name": "new"},
			current:    []*ec2.Tag{{Key: aws.String("Name"), Value: aws.String("old")}},
			wantAdd:    []*ec2.Tag{{Key: aws.String("Name"), Value: aws.String("new")}},
			wantRemove: []*ec2.Tag{{Key: aws.String("Name"), Value: aws.String("old")}},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			gotAdd, gotRemove := diffVPCEndpointTags(tc.spec, tc.current)
			if diff := cmp.Diff(tc.wantAdd, gotAdd); diff != "" {
				t.Errorf("addTags mismatch (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tc.wantRemove, gotRemove); diff != "" {
				t.Errorf("removeTags mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestCreateWithTags(t *testing.T) {
	var gotInput *ec2.CreateVpcEndpointInput
	mock := &fake.MockVPCEndpointClient{
		MockCreateVpcEndpointWithContext: func(_ context.Context, input *ec2.CreateVpcEndpointInput, _ ...request.Option) (*ec2.CreateVpcEndpointOutput, error) {
			gotInput = input
			return &ec2.CreateVpcEndpointOutput{VpcEndpoint: &ec2.VpcEndpoint{VpcEndpointId: aws.String(testVPCEndpointID)}}, nil
		},
	}

	cr := vpcEndpoint(withSpec(v1alpha1.VPCEndpointParameters{
		CustomVPCEndpointParameters: v1alpha1.CustomVPCEndpointParameters{
			VPCID: aws.String(testVPCID),
			Tags:  map[string]string{"Name": "my-endpoint"},
		},
	}))

	e := newExternal(nil, mock, []option{setupExternal})
	if _, err := e.Create(context.Background(), cr); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	want := []*ec2.TagSpecification{{
		ResourceType: aws.String(vpcEndpointTagResource),
		Tags:         []*ec2.Tag{{Key: aws.String("Name"), Value: aws.String("my-endpoint")}},
	}}
	if diff := cmp.Diff(want, gotInput.TagSpecifications); diff != "" {
		t.Errorf("Create() TagSpecifications mismatch (-want +got):\n%s", diff)
	}
}

func TestUpdateTags(t *testing.T) {
	type captured struct {
		createTags *ec2.CreateTagsInput
		deleteTags *ec2.DeleteTagsInput
	}

	cases := map[string]struct {
		specTags     map[string]string
		observedTags []*ec2.Tag
		wantCreate   []*ec2.Tag
		wantDelete   []*ec2.Tag
	}{
		"AddTag": {
			specTags:     map[string]string{"Name": "my-endpoint"},
			observedTags: nil,
			wantCreate:   []*ec2.Tag{{Key: aws.String("Name"), Value: aws.String("my-endpoint")}},
		},
		"RemoveTag": {
			specTags:     map[string]string{},
			observedTags: []*ec2.Tag{{Key: aws.String("Old"), Value: aws.String("gone")}},
			wantDelete:   []*ec2.Tag{{Key: aws.String("Old"), Value: aws.String("gone")}},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			c := &captured{}
			mock := &fake.MockVPCEndpointClient{
				MockDescribeVpcEndpoints: func(*ec2.DescribeVpcEndpointsInput) (*ec2.DescribeVpcEndpointsOutput, error) {
					return &ec2.DescribeVpcEndpointsOutput{VpcEndpoints: []*ec2.VpcEndpoint{{
						VpcEndpointId: aws.String(testVPCEndpointID),
						Tags:          tc.observedTags,
					}}}, nil
				},
				MockModifyVpcEndpointWithContext: func(_ context.Context, _ *ec2.ModifyVpcEndpointInput, _ ...request.Option) (*ec2.ModifyVpcEndpointOutput, error) {
					return &ec2.ModifyVpcEndpointOutput{}, nil
				},
				MockCreateTagsWithContext: func(_ context.Context, input *ec2.CreateTagsInput, _ ...request.Option) (*ec2.CreateTagsOutput, error) {
					c.createTags = input
					return &ec2.CreateTagsOutput{}, nil
				},
				MockDeleteTagsWithContext: func(_ context.Context, input *ec2.DeleteTagsInput, _ ...request.Option) (*ec2.DeleteTagsOutput, error) {
					c.deleteTags = input
					return &ec2.DeleteTagsOutput{}, nil
				},
			}

			cr := vpcEndpoint(
				withExternalName(testVPCEndpointID),
				withSpec(v1alpha1.VPCEndpointParameters{
					CustomVPCEndpointParameters: v1alpha1.CustomVPCEndpointParameters{
						Tags: tc.specTags,
					},
				}),
			)

			e := newExternal(nil, mock, []option{setupExternal})
			if _, err := e.Update(context.Background(), cr); err != nil {
				t.Fatalf("Update() error = %v", err)
			}

			var gotCreate []*ec2.Tag
			if c.createTags != nil {
				gotCreate = c.createTags.Tags
			}
			if diff := cmp.Diff(tc.wantCreate, gotCreate); diff != "" {
				t.Errorf("CreateTags mismatch (-want +got):\n%s", diff)
			}

			var gotDelete []*ec2.Tag
			if c.deleteTags != nil {
				gotDelete = c.deleteTags.Tags
			}
			if diff := cmp.Diff(tc.wantDelete, gotDelete); diff != "" {
				t.Errorf("DeleteTags mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
