/*
Copyright 2020 The Crossplane Authors.

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

package v1alpha1

import (
	"sort"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
)

// Tag defines a tag
type Tag struct {

	// Key is the name of the tag.
	Key string `json:"key"`

	// Value is the value of the tag.
	Value string `json:"value"`
}

// BuildFromECRTags returns a list of tags, off of the given ecr tags
func BuildFromECRTags(tags []ecr.Tag) []Tag {
	if len(tags) < 1 {
		return nil
	}
	res := make([]Tag, len(tags))
	for i, t := range tags {
		res[i] = Tag{aws.StringValue(t.Key), aws.StringValue(t.Value)}
	}

	return res
}

// GenerateECRTags generates a tag array with type that ECR client expects.
func GenerateECRTags(tags []Tag) []ecr.Tag {
	res := make([]ecr.Tag, len(tags))
	for i, t := range tags {
		res[i] = ecr.Tag{Key: aws.String(t.Key), Value: aws.String(t.Value)}
	}
	return res
}

// CompareTags compares arrays of v1alpha1.Tag and ecr.Tag
func CompareTags(tags []Tag, ecrTags []ecr.Tag) bool {
	if len(tags) != len(ecrTags) {
		return false
	}

	SortTags(tags, ecrTags)

	for i, t := range tags {
		if t.Key != *ecrTags[i].Key || t.Value != *ecrTags[i].Value {
			return false
		}
	}

	return true
}

// SortTags sorts array of v1alpha1.Tag and ecr.Tag on 'Key'
func SortTags(tags []Tag, ecrTags []ecr.Tag) {
	sort.Slice(tags, func(i, j int) bool {
		return tags[i].Key < tags[j].Key
	})

	sort.Slice(ecrTags, func(i, j int) bool {
		return *ecrTags[i].Key < *ecrTags[j].Key
	})
}

// TagsToMap converts []v1alpha1.Tag to map
func TagsToMap(tags []Tag) map[string]string {
	result := make(map[string]string)
	for i := range tags {
		result[tags[i].Key] = tags[i].Value
	}
	return result
}

// ECRTagsToMap converts []ecr.Tag to map
func ECRTagsToMap(ecrTags []ecr.Tag) map[string]string {
	result := make(map[string]string)
	for i := range ecrTags {
		result[aws.StringValue(ecrTags[i].Key)] = aws.StringValue(ecrTags[i].Value)
	}
	return result
}

// DiffTags returns tags that should be added or removed.
func DiffTags(spec []Tag, current []ecr.Tag) (addTags []ecr.Tag, remove []string) {
	local := TagsToMap(spec)
	remote := ECRTagsToMap(current)
	add := make(map[string]string, len(local))
	remove = []string{}
	for k, v := range local {
		add[k] = v
	}
	for k, v := range remote {
		switch val, ok := local[k]; {
		case ok && val != v:
			remove = append(remove, k)
		case !ok:
			remove = append(remove, k)
			delete(add, k)
		default:
			delete(add, k)
		}
	}
	addTags = []ecr.Tag{}
	for key, value := range add {
		value := value
		key := key
		addTags = append(addTags, ecr.Tag{Key: &key, Value: &value})
	}
	return
}
