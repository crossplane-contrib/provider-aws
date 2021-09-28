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

package v1beta1

import (
	"sort"

	"github.com/aws/aws-sdk-go-v2/aws"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

// Tag defines a tag
type Tag struct {

	// Key is the name of the tag.
	Key string `json:"key"`

	// Value is the value of the tag.
	Value string `json:"value"`
}

// BuildFromEC2Tags returns a list of tags, off of the given ec2 tags
func BuildFromEC2Tags(tags []ec2types.Tag) []Tag {
	if len(tags) < 1 {
		return nil
	}
	res := make([]Tag, len(tags))
	for i, t := range tags {
		res[i] = Tag{aws.ToString(t.Key), aws.ToString(t.Value)}
	}

	return res
}

// GenerateEC2Tags generates a tag array with type that EC2 client expects.
func GenerateEC2Tags(tags []Tag) []ec2types.Tag {
	res := make([]ec2types.Tag, len(tags))
	for i, t := range tags {
		res[i] = ec2types.Tag{Key: aws.String(t.Key), Value: aws.String(t.Value)}
	}
	return res
}

// CompareTags compares arrays of v1beta1.Tag and ec2type.Tag
func CompareTags(tags []Tag, ec2Tags []ec2types.Tag) bool {
	if len(tags) != len(ec2Tags) {
		return false
	}

	SortTags(tags, ec2Tags)

	for i, t := range tags {
		if t.Key != *ec2Tags[i].Key || t.Value != *ec2Tags[i].Value {
			return false
		}
	}

	return true
}

// SortTags sorts array of v1beta1.Tag and ec2type.Tag on 'Key'
func SortTags(tags []Tag, ec2Tags []ec2types.Tag) {
	sort.Slice(tags, func(i, j int) bool {
		return tags[i].Key < tags[j].Key
	})

	sort.Slice(ec2Tags, func(i, j int) bool {
		return *ec2Tags[i].Key < *ec2Tags[j].Key
	})
}
