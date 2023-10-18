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
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"
)

func TestCompareGroupNames(t *testing.T) {
	type args struct {
		groupNames []string
		ec2Groups  []types.GroupIdentifier
	}
	cases := map[string]struct {
		args args
		want bool
	}{
		"GroupNamesAreEqual": {
			args: args{
				groupNames: []string{"group-2", "group-1"},
				ec2Groups: []types.GroupIdentifier{
					{
						GroupId:   aws.String("group-2"),
						GroupName: aws.String("group-2"),
					},
					{
						GroupId:   aws.String("group-1"),
						GroupName: aws.String("group-1"),
					},
				},
			},
			want: true,
		},
		"GroupNamesAreEqualDifferentOrder": {
			args: args{
				groupNames: []string{"group-1", "group-2"},
				ec2Groups: []types.GroupIdentifier{
					{
						GroupId:   aws.String("group-2"),
						GroupName: aws.String("group-2"),
					},
					{
						GroupId:   aws.String("group-1"),
						GroupName: aws.String("group-1"),
					},
				},
			},
			want: true,
		},
		"GroupNamesAreNotEqual": {
			args: args{
				groupNames: []string{"group-2", "group-3"},
				ec2Groups: []types.GroupIdentifier{
					{
						GroupId:   aws.String("group-2"),
						GroupName: aws.String("group-2"),
					},
					{
						GroupId:   aws.String("group-1"),
						GroupName: aws.String("group-1"),
					},
				},
			},
			want: false,
		},
		"GroupNamesAreNotEqualDifferentLength": {
			args: args{
				groupNames: []string{"group-1", "group-2", "group-3"},
				ec2Groups: []types.GroupIdentifier{
					{
						GroupId:   aws.String("group-2"),
						GroupName: aws.String("group-2"),
					},
					{
						GroupId:   aws.String("group-1"),
						GroupName: aws.String("group-1"),
					},
				},
			},
			want: false,
		},
		"GroupNamesAreNil": {
			args: args{
				groupNames: nil,
				ec2Groups: []types.GroupIdentifier{
					{
						GroupId:   aws.String("group-2"),
						GroupName: aws.String("group-2"),
					},
					{
						GroupId:   aws.String("group-1"),
						GroupName: aws.String("group-1"),
					},
				},
			},
			want: false,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			result := CompareGroupNames(tc.args.groupNames, tc.args.ec2Groups)

			if diff := cmp.Diff(tc.want, result, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestCompareGroupIDs(t *testing.T) {
	type args struct {
		groupIDs  []string
		ec2Groups []types.GroupIdentifier
	}
	cases := map[string]struct {
		args args
		want bool
	}{
		"GroupIDsAreEqual": {
			args: args{
				groupIDs: []string{"groupid-2", "groupid-1"},
				ec2Groups: []types.GroupIdentifier{
					{
						GroupId:   aws.String("groupid-2"),
						GroupName: aws.String("group-2"),
					},
					{
						GroupId:   aws.String("groupid-1"),
						GroupName: aws.String("group-1"),
					},
				},
			},
			want: true,
		},
		"GroupIDsAreEqualDifferentOrder": {
			args: args{
				groupIDs: []string{"groupid-1", "groupid-2"},
				ec2Groups: []types.GroupIdentifier{
					{
						GroupId:   aws.String("groupid-2"),
						GroupName: aws.String("group-2"),
					},
					{
						GroupId:   aws.String("groupid-1"),
						GroupName: aws.String("group-1"),
					},
				},
			},
			want: true,
		},
		"GroupIDsAreNotEqual": {
			args: args{
				groupIDs: []string{"groupid-2", "groupid-3"},
				ec2Groups: []types.GroupIdentifier{
					{
						GroupId:   aws.String("groupid-2"),
						GroupName: aws.String("group-2"),
					},
					{
						GroupId:   aws.String("groupid-1"),
						GroupName: aws.String("group-1"),
					},
				},
			},
			want: false,
		},
		"GroupIDsAreNotEqualDifferentLength": {
			args: args{
				groupIDs: []string{"groupid-1", "groupid-2", "groupid-3"},
				ec2Groups: []types.GroupIdentifier{
					{
						GroupId:   aws.String("groupid-2"),
						GroupName: aws.String("group-2"),
					},
					{
						GroupId:   aws.String("groupid-1"),
						GroupName: aws.String("group-1"),
					},
				},
			},
			want: false,
		},
		"GroupIDsAreNil": {
			args: args{
				groupIDs: nil,
				ec2Groups: []types.GroupIdentifier{
					{
						GroupId:   aws.String("groupid-2"),
						GroupName: aws.String("group-2"),
					},
					{
						GroupId:   aws.String("groupid-1"),
						GroupName: aws.String("group-1"),
					},
				},
			},
			want: false,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			result := CompareGroupIDs(tc.args.groupIDs, tc.args.ec2Groups)

			if diff := cmp.Diff(tc.want, result, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}
