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

package resourcetag

import (
	awsec2 "github.com/aws/aws-sdk-go-v2/service/ec2"

	"github.com/crossplane/provider-aws/apis/ec2/manualv1alpha1"
	awsclient "github.com/crossplane/provider-aws/pkg/clients"
)

// diffTags between spec and current
//
// This only detects tags that are either missing or already existing. It does remove already added tags to avoid conflicts
// with other controllers that might add tags to the same resource.
func diffTags(cr *manualv1alpha1.ResourceTag, current []awsec2.TagDescription) (missingTags, existingTags map[string]string) {
	currentMap := make(map[string]map[string]string)
	for _, t := range current {
		resID := awsclient.StringValue(t.ResourceId)
		key := awsclient.StringValue(t.Key)

		resTags, exists := currentMap[resID]
		if !exists {
			resTags = make(map[string]string)
			currentMap[resID] = resTags
		}

		resTags[key] = awsclient.StringValue(t.Value)
	}

	missingTags = make(map[string]string)
	existingTags = make(map[string]string)

	for _, resID := range cr.Spec.ForProvider.ResourceIDs {
		resTags, exists := currentMap[resID]
		if !exists {
			for _, t := range cr.Spec.ForProvider.Tags {
				missingTags[t.Key] = t.Value
			}
		}

		for _, t := range cr.Spec.ForProvider.Tags {
			if val, exists := resTags[t.Key]; exists && val == t.Value {
				existingTags[t.Key] = t.Value
			} else {
				missingTags[t.Key] = t.Value
			}
		}
	}

	return existingTags, missingTags
}
