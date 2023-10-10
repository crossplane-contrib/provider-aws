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
	"github.com/aws/aws-sdk-go-v2/aws"
	ec2type "github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

// DiffEC2Tags returns []ec2type.Tag that should be added or removed.
func DiffEC2Tags(local []ec2type.Tag, remote []ec2type.Tag) (add []ec2type.Tag, remove []ec2type.Tag) {
	var tagsToAdd = make(map[string]string, len(local))
	add = []ec2type.Tag{}
	remove = []ec2type.Tag{}
	for _, j := range local {
		tagsToAdd[aws.ToString(j.Key)] = aws.ToString(j.Value)
	}
	for _, j := range remote {
		switch val, ok := tagsToAdd[aws.ToString(j.Key)]; {
		case ok && val == aws.ToString(j.Value):
			delete(tagsToAdd, aws.ToString(j.Key))
		case !ok:
			remove = append(remove, ec2type.Tag{
				Key:   j.Key,
				Value: nil,
			})
		}
	}
	for i, j := range tagsToAdd {
		add = append(add, ec2type.Tag{
			Key:   aws.String(i),
			Value: aws.String(j),
		})
	}
	return
}
