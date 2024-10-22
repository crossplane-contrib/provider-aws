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

package ec2

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/smithy-go/document"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"k8s.io/utils/ptr"
)

func TestDiffEC2Tags(t *testing.T) {
	type args struct {
		local  []ec2types.Tag
		remote []ec2types.Tag
	}
	type want struct {
		add    []ec2types.Tag
		remove []ec2types.Tag
	}
	cases := map[string]struct {
		args
		want
	}{
		"EmptyLocalAndRemote": {
			args: args{
				local:  []ec2types.Tag{},
				remote: []ec2types.Tag{},
			},
			want: want{
				add:    []ec2types.Tag{},
				remove: []ec2types.Tag{},
			},
		},
		"TagsWithSameKeyValuesAndLength": {
			args: args{
				local: []ec2types.Tag{
					{
						Key:   aws.String("name"),
						Value: aws.String("somename"),
					},
				},
				remote: []ec2types.Tag{
					{
						Key:   aws.String("name"),
						Value: aws.String("somename"),
					},
				},
			},
			want: want{
				add:    []ec2types.Tag{},
				remove: []ec2types.Tag{},
			},
		},
		"TagsWithSameKeyValuesButDifferentOrder": {
			args: args{
				local: []ec2types.Tag{
					{
						Key:   aws.String("val"),
						Value: aws.String("key"),
					},
					{
						Key:   aws.String("name"),
						Value: aws.String("somename"),
					},
				},
				remote: []ec2types.Tag{
					{
						Key:   aws.String("name"),
						Value: aws.String("somename"),
					},
					{
						Key:   aws.String("val"),
						Value: aws.String("key"),
					},
				},
			},
			want: want{
				add:    []ec2types.Tag{},
				remove: []ec2types.Tag{},
			},
		},
		"TagsWithSameKeyDifferentValuesAndSameLength": {
			args: args{
				local: []ec2types.Tag{
					{
						Key:   aws.String("name"),
						Value: aws.String("somename"),
					},
				},
				remote: []ec2types.Tag{
					{
						Key:   aws.String("name"),
						Value: aws.String("somenames"),
					},
				},
			},
			want: want{
				add: []ec2types.Tag{
					{
						Key:   aws.String("name"),
						Value: aws.String("somename"),
					},
				},
				remove: []ec2types.Tag{},
			},
		},
		"EmptyRemoteAndMultipleInputs": {
			args: args{
				local: []ec2types.Tag{
					{
						Key:   aws.String("name"),
						Value: aws.String("somename"),
					},
					{
						Key:   aws.String("tags"),
						Value: aws.String("True"),
					},
				},
				remote: []ec2types.Tag{},
			},
			want: want{
				add: []ec2types.Tag{
					{
						Key:   aws.String("name"),
						Value: aws.String("somename"),
					},
					{
						Key:   aws.String("tags"),
						Value: aws.String("True"),
					},
				},
				remove: []ec2types.Tag{},
			},
		},
		"EmptyLocalAndMultipleRemote": {
			args: args{
				local: []ec2types.Tag{},
				remote: []ec2types.Tag{
					{
						Key:   aws.String("name"),
						Value: aws.String("somename"),
					},
					{
						Key:   aws.String("tags"),
						Value: aws.String("True"),
					},
				},
			},
			want: want{
				add: []ec2types.Tag{},
				remove: []ec2types.Tag{
					{
						Key:   aws.String("name"),
						Value: nil,
					},
					{
						Key:   aws.String("tags"),
						Value: nil,
					},
				},
			},
		},
		"LocalHaveMoreTags": {
			args: args{
				local: []ec2types.Tag{
					{
						Key:   aws.String("name"),
						Value: aws.String("somename"),
					},
					{
						Key:   aws.String("tags"),
						Value: aws.String("True"),
					},
				},
				remote: []ec2types.Tag{
					{
						Key:   aws.String("name"),
						Value: aws.String("somename"),
					},
					{
						Key:   aws.String("val"),
						Value: aws.String("key"),
					},
					{
						Key:   aws.String("val1"),
						Value: aws.String("key2"),
					},
				},
			},
			want: want{
				add: []ec2types.Tag{
					{
						Key:   aws.String("tags"),
						Value: aws.String("True"),
					},
				},
				remove: []ec2types.Tag{
					{
						Key:   aws.String("val"),
						Value: nil,
					},
					{
						Key:   aws.String("val1"),
						Value: nil,
					},
				},
			},
		},
		"RemoteHaveMoreTags": {
			args: args{
				local: []ec2types.Tag{
					{
						Key:   aws.String("name"),
						Value: aws.String("somename"),
					},
					{
						Key:   aws.String("val"),
						Value: aws.String("key"),
					},
				},
				remote: []ec2types.Tag{
					{
						Key:   aws.String("name"),
						Value: aws.String("somename"),
					},
					{
						Key:   aws.String("tags"),
						Value: aws.String("True"),
					},
					{
						Key:   aws.String("val"),
						Value: aws.String("key"),
					},
				},
			},
			want: want{
				add: []ec2types.Tag{},
				remove: []ec2types.Tag{
					{
						Key:   aws.String("tags"),
						Value: nil,
					},
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			tagCmp := cmpopts.SortSlices(func(i, j ec2types.Tag) bool {
				return ptr.Deref(i.Key, "") < ptr.Deref(j.Key, "")
			})
			add, remove := DiffEC2Tags(tc.args.local, tc.args.remote)
			if diff := cmp.Diff(tc.want.add, add, tagCmp, cmpopts.IgnoreTypes(document.NoSerde{})); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.remove, remove, tagCmp, cmpopts.IgnoreTypes(document.NoSerde{})); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}
