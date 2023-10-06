package iam_test

import (
	"sort"
	"testing"

	iamtypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/smithy-go/document"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/crossplane-contrib/provider-aws/apis/iam/v1beta1"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/iam"
)

func TestDiffIAMTags(t *testing.T) {
	type args struct {
		local  []v1beta1.Tag
		remote []iamtypes.Tag
	}
	type want struct {
		add    []iamtypes.Tag
		remove []string
	}
	cases := map[string]struct {
		args args
		want want
	}{
		"AllNew": {
			args: args{
				local: []v1beta1.Tag{
					{Key: "key", Value: "val"},
				},
			},
			want: want{
				add: []iamtypes.Tag{
					{Key: aws.String("key"), Value: aws.String("val")},
				},
			},
		},
		"SomeNew": {
			args: args{
				local: []v1beta1.Tag{
					{Key: "key", Value: "val"},
					{Key: "key1", Value: "val1"},
					{Key: "key2", Value: "val2"},
				},
				remote: []iamtypes.Tag{
					{Key: aws.String("key"), Value: aws.String("val")},
				},
			},
			want: want{
				add: []iamtypes.Tag{
					{Key: aws.String("key1"), Value: aws.String("val1")},
					{Key: aws.String("key2"), Value: aws.String("val2")},
				},
			},
		},
		"Update": {
			args: args{
				local: []v1beta1.Tag{
					{Key: "key", Value: "different"},
					{Key: "key1", Value: "val1"},
					{Key: "key2", Value: "val2"},
				},
				remote: []iamtypes.Tag{
					{Key: aws.String("key"), Value: aws.String("val")},
					{Key: aws.String("key1"), Value: aws.String("val1")},
					{Key: aws.String("key2"), Value: aws.String("val2")},
				},
			},
			want: want{
				add: []iamtypes.Tag{
					{Key: aws.String("key"), Value: aws.String("different")},
				},
				remove: []string{"key"},
			},
		},
		"RemoveAll": {
			args: args{
				remote: []iamtypes.Tag{
					{Key: aws.String("key"), Value: aws.String("val")},
					{Key: aws.String("key1"), Value: aws.String("val1")},
					{Key: aws.String("key2"), Value: aws.String("val2")},
				},
			},
			want: want{
				remove: []string{"key", "key1", "key2"},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			tagCmp := cmpopts.SortSlices(func(i, j iamtypes.Tag) bool {
				return aws.StringValue(i.Key) < aws.StringValue(j.Key)
			})

			crTagMap := make(map[string]string, len(tc.args.local))
			for _, v := range tc.args.local {
				crTagMap[v.Key] = v.Value
			}

			add, remove, _ := iam.DiffIAMTags(crTagMap, tc.args.remote)
			if diff := cmp.Diff(tc.want.add, add, tagCmp, cmpopts.IgnoreTypes(document.NoSerde{})); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			sort.Strings(tc.want.remove)
			sort.Strings(remove)
			if diff := cmp.Diff(tc.want.remove, remove, tagCmp, cmpopts.IgnoreTypes(document.NoSerde{})); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestDiffIAMTagsWithUpdated(t *testing.T) {
	type args struct {
		local  []v1beta1.Tag
		remote []iamtypes.Tag
	}
	type want struct {
		addOrUpdate    []iamtypes.Tag
		remove         []string
		areTagsUpdated bool
	}
	cases := map[string]struct {
		args args
		want want
	}{
		"AllNew": {
			args: args{
				local: []v1beta1.Tag{
					{Key: "key", Value: "val"},
				},
			},
			want: want{
				addOrUpdate: []iamtypes.Tag{
					{Key: aws.String("key"), Value: aws.String("val")},
				},
			},
		},
		"SomeNew": {
			args: args{
				local: []v1beta1.Tag{
					{Key: "key", Value: "val"},
					{Key: "key1", Value: "val1"},
					{Key: "key2", Value: "val2"},
				},
				remote: []iamtypes.Tag{
					{Key: aws.String("key"), Value: aws.String("val")},
				},
			},
			want: want{
				addOrUpdate: []iamtypes.Tag{
					{Key: aws.String("key1"), Value: aws.String("val1")},
					{Key: aws.String("key2"), Value: aws.String("val2")},
				},
			},
		},
		"Update": {
			args: args{
				local: []v1beta1.Tag{
					{Key: "key", Value: "different"},
					{Key: "key1", Value: "val1"},
					{Key: "key2", Value: "val2"},
				},
				remote: []iamtypes.Tag{
					{Key: aws.String("key"), Value: aws.String("val")},
					{Key: aws.String("key1"), Value: aws.String("val1")},
					{Key: aws.String("key2"), Value: aws.String("val2")},
				},
			},
			want: want{
				addOrUpdate: []iamtypes.Tag{
					{Key: aws.String("key"), Value: aws.String("different")},
				},
			},
		},
		"RemoveAll": {
			args: args{
				remote: []iamtypes.Tag{
					{Key: aws.String("key"), Value: aws.String("val")},
					{Key: aws.String("key1"), Value: aws.String("val1")},
					{Key: aws.String("key2"), Value: aws.String("val2")},
				},
			},
			want: want{
				remove: []string{"key", "key1", "key2"},
			},
		},
		"NothingToChange": {
			args: args{
				local: []v1beta1.Tag{
					{Key: "key", Value: "val"},
				},
				remote: []iamtypes.Tag{
					{Key: aws.String("key"), Value: aws.String("val")},
				},
			},
			want: want{
				areTagsUpdated: true,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			tagCmp := cmpopts.SortSlices(func(i, j iamtypes.Tag) bool {
				return aws.StringValue(i.Key) < aws.StringValue(j.Key)
			})

			add, remove, areTagsUpdated := iam.DiffIAMTagsWithUpdates(tc.args.local, tc.args.remote)
			if diff := cmp.Diff(tc.want.addOrUpdate, add, tagCmp, cmpopts.IgnoreTypes(document.NoSerde{})); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}

			sort.Strings(tc.want.remove)
			sort.Strings(remove)
			if diff := cmp.Diff(tc.want.remove, remove, tagCmp, cmpopts.IgnoreTypes(document.NoSerde{})); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}

			if tc.want.areTagsUpdated != areTagsUpdated {
				t.Errorf("r: want: %t, got:%t", tc.want.areTagsUpdated, areTagsUpdated)
			}
		})
	}
}
