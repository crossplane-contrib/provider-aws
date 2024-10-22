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

package ec2

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/ec2/v1beta1"
)

// BuildFromEC2Tags returns a list of tags, off of the given ec2 tags
func BuildFromEC2TagsV1Beta1(tags []ec2types.Tag) []svcapitypes.Tag {
	if len(tags) < 1 {
		return nil
	}
	res := make([]svcapitypes.Tag, len(tags))
	for i, t := range tags {
		res[i] = svcapitypes.Tag{Key: aws.ToString(t.Key), Value: aws.ToString(t.Value)}
	}

	return res
}

// GenerateEC2Tags generates a tag array with type that EC2 client expects.
func GenerateEC2TagsV1Beta1(tags []svcapitypes.Tag) []ec2types.Tag {
	res := make([]ec2types.Tag, len(tags))
	for i, t := range tags {
		res[i] = ec2types.Tag{Key: aws.String(t.Key), Value: aws.String(t.Value)}
	}
	return res
}

// CompareTags compares arrays of v1beta1.Tag and ec2types.Tag
func CompareTags(spec []svcapitypes.Tag, current []ec2types.Tag) bool {
	if len(spec) != len(current) {
		return false
	}
	toAdd, toRemove := DiffEC2Tags(GenerateEC2TagsV1Beta1(spec), current)
	return len(toAdd) == 0 && len(toRemove) == 0
}
