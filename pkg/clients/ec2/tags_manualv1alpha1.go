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
	"sort"

	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"k8s.io/utils/ptr"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/ec2/manualv1alpha1"
)

// BuildFromEC2Tags returns a list of tags, off of the given ec2 tags
func BuildFromEC2TagsManualV1alpha1(tags []types.Tag) []svcapitypes.Tag {
	if len(tags) < 1 {
		return nil
	}
	res := make([]svcapitypes.Tag, len(tags))
	for i, t := range tags {
		res[i] = svcapitypes.Tag{Key: ptr.Deref(t.Key, ""), Value: ptr.Deref(t.Value, "")}
	}

	return res
}

// GenerateEC2Tags generates a tag array with type that EC2 client expects.
func GenerateEC2TagsManualV1alpha1(tags []svcapitypes.Tag) []types.Tag {
	res := make([]types.Tag, len(tags))
	for i, t := range tags {
		res[i] = types.Tag{Key: ptr.To(t.Key), Value: ptr.To(t.Value)}
	}
	return res
}

// CompareGroupNames compares slices of group names and ec2.GroupIdentifier
func CompareGroupNames(groupNames []string, ec2Groups []types.GroupIdentifier) bool {
	if len(groupNames) != len(ec2Groups) {
		return false
	}

	sortGroupNames(groupNames, ec2Groups)

	for i, g := range groupNames {
		if g != *ec2Groups[i].GroupName {
			return false
		}
	}

	return true
}

// CompareGroupIDs compares slices of group IDs and ec2.GroupIdentifier
func CompareGroupIDs(groupIDs []string, ec2Groups []types.GroupIdentifier) bool {
	if len(groupIDs) != len(ec2Groups) {
		return false
	}

	sortGroupIDs(groupIDs, ec2Groups)

	for i, g := range groupIDs {
		if g != *ec2Groups[i].GroupId {
			return false
		}
	}

	return true
}

// sortGroupNames sorts slice of string and ec2.GroupIdentifier on 'GroupName'
func sortGroupNames(groupNames []string, ec2Groups []types.GroupIdentifier) {
	sort.Strings(groupNames)

	sort.Slice(ec2Groups, func(i, j int) bool {
		return *ec2Groups[i].GroupName < *ec2Groups[j].GroupName
	})
}

// sortGroupNames sorts slice of string and ec2.GroupIdentifier on 'GroupName'
func sortGroupIDs(groupIDs []string, ec2Groups []types.GroupIdentifier) {
	sort.Strings(groupIDs)

	sort.Slice(ec2Groups, func(i, j int) bool {
		return *ec2Groups[i].GroupId < *ec2Groups[j].GroupId
	})
}

// CompareTags compares arrays of manualv1alpha1.Tag and ec2.Tag
func CompareTagsManualV1alpha1(tags []svcapitypes.Tag, ec2Tags []types.Tag) bool {
	if len(tags) != len(ec2Tags) {
		return false
	}

	SortTagsManualV1alpha1(tags, ec2Tags)

	for i, t := range tags {
		if t.Key != *ec2Tags[i].Key || t.Value != *ec2Tags[i].Value {
			return false
		}
	}

	return true
}

// SortTags sorts array of v1beta1.Tag and ec2.Tag on 'Key'
func SortTagsManualV1alpha1(tags []svcapitypes.Tag, ec2Tags []types.Tag) {
	sort.Slice(tags, func(i, j int) bool {
		return tags[i].Key < tags[j].Key
	})

	sort.Slice(ec2Tags, func(i, j int) bool {
		return *ec2Tags[i].Key < *ec2Tags[j].Key
	})
}
