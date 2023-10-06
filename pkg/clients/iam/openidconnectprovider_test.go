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

package iam

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"

	"github.com/crossplane-contrib/provider-aws/apis/iam/v1beta1"
)

func TestSliceDiff(t *testing.T) {
	type want struct {
		add    []string
		remove []string
	}
	type args struct {
		current []string
		desired []string
	}

	cases := map[string]struct {
		args
		want
	}{
		"AddRemove": {
			args: args{
				current: []string{"foo", "bar"},
				desired: []string{"a", "bar"},
			},
			want: want{
				add:    []string{"a"},
				remove: []string{"foo"},
			},
		},
		"NoChange": {
			args: args{
				current: []string{"foo", "bar"},
				desired: []string{"foo", "bar"},
			},
			want: want{
				add:    nil,
				remove: nil,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			add, remove := SliceDifference(tc.current, tc.desired)

			if diff := cmp.Diff(tc.want.add, add, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.remove, remove, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestIsOIDCProviderUpToDate(t *testing.T) {
	type args struct {
		input    v1beta1.OpenIDConnectProviderParameters
		observed iam.GetOpenIDConnectProviderOutput
	}

	cases := map[string]struct {
		args args
		want bool
	}{
		"DifferentClientIDList": {
			args: args{
				input:    v1beta1.OpenIDConnectProviderParameters{ClientIDList: []string{"client1", "client3"}},
				observed: iam.GetOpenIDConnectProviderOutput{ClientIDList: []string{"client2", "client3"}},
			},
			want: false,
		},
		"DifferentThumbprintList": {
			args: args{
				input:    v1beta1.OpenIDConnectProviderParameters{ThumbprintList: []string{"thumbprint1", "thumbprint3"}},
				observed: iam.GetOpenIDConnectProviderOutput{ThumbprintList: []string{"thumbprint2", "thumbprint3"}},
			},
			want: false,
		},
		"DifferentTags": {
			args: args{
				input: v1beta1.OpenIDConnectProviderParameters{Tags: []v1beta1.Tag{
					{Key: "key1", Value: "value1"},
					{Key: "key3", Value: "value3"},
				}},
				observed: iam.GetOpenIDConnectProviderOutput{Tags: []types.Tag{
					{Key: aws.String("key2"), Value: aws.String("value2")},
					{Key: aws.String("key3"), Value: aws.String("value3")},
				}},
			},
			want: false,
		},
		"UpToDate": {
			args: args{
				input: v1beta1.OpenIDConnectProviderParameters{
					ClientIDList:   []string{"client1", "client2"},
					ThumbprintList: []string{"thumbprint1", "thumbprint2"},
					Tags: []v1beta1.Tag{
						{Key: "key1", Value: "value1"},
						{Key: "key2", Value: "value2"},
					},
				},
				observed: iam.GetOpenIDConnectProviderOutput{
					ClientIDList:   []string{"client2", "client1"},
					ThumbprintList: []string{"thumbprint2", "thumbprint1"},
					Tags: []types.Tag{
						{Key: aws.String("key2"), Value: aws.String("value2")},
						{Key: aws.String("key1"), Value: aws.String("value1")},
					},
				},
			},
			want: true,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			o := IsOIDCProviderUpToDate(tc.args.input, tc.args.observed)
			if o != tc.want {
				t.Errorf("want %t, got %t", tc.want, o)
			}
		})
	}
}
