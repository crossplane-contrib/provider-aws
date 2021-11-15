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

package docdb

import (
	"sort"

	svcsdk "github.com/aws/aws-sdk-go/service/docdb"
	"github.com/aws/aws-sdk-go/service/docdb/docdbiface"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/pkg/errors"

	svcapitypes "github.com/crossplane/provider-aws/apis/docdb/v1alpha1"
	awsclient "github.com/crossplane/provider-aws/pkg/clients"
)

const (
	errListTagsForResource = "cannot list tags"
	errRemoveTags          = "cannot remove tags"
	errCreateTags          = "cannot create tags"
)

// AreTagsUpToDate for spec and resourceName
func AreTagsUpToDate(client docdbiface.DocDBAPI, spec []*svcapitypes.Tag, resourceName *string) (bool, error) {
	current, err := ListTagsForResource(client, resourceName)
	if err != nil {
		return false, err
	}

	add, remove := DiffTags(spec, current)

	return len(add) == 0 && len(remove) == 0, nil
}

// UpdateTagsForResource with resourceName
func UpdateTagsForResource(client docdbiface.DocDBAPI, spec []*svcapitypes.Tag, resourceName *string) error {
	current, err := ListTagsForResource(client, resourceName)
	if err != nil {
		return err
	}

	add, remove := DiffTags(spec, current)
	if len(remove) != 0 {
		if _, err := client.RemoveTagsFromResource(&svcsdk.RemoveTagsFromResourceInput{
			ResourceName: resourceName,
			TagKeys:      remove,
		}); err != nil {
			return errors.Wrap(err, errRemoveTags)
		}
	}
	if len(add) != 0 {
		if _, err := client.AddTagsToResource(&svcsdk.AddTagsToResourceInput{
			ResourceName: resourceName,
			Tags:         add,
		}); err != nil {
			return errors.Wrap(err, errCreateTags)
		}
	}

	return nil
}

// ListTagsForResource for the given resource
func ListTagsForResource(client docdbiface.DocDBAPI, resourceName *string) ([]*svcsdk.Tag, error) {
	req := &svcsdk.ListTagsForResourceInput{
		ResourceName: resourceName,
	}

	resp, err := client.ListTagsForResource(req)
	if err != nil {
		return nil, errors.Wrap(err, errListTagsForResource)
	}

	return resp.TagList, nil
}

// DiffTags between spec and current
func DiffTags(spec []*svcapitypes.Tag, current []*svcsdk.Tag) (addTags []*svcsdk.Tag, removeTags []*string) {
	currentMap := make(map[string]string, len(current))
	for _, t := range current {
		currentMap[awsclient.StringValue(t.Key)] = awsclient.StringValue(t.Value)
	}

	specMap := make(map[string]string, len(spec))
	for _, t := range spec {
		key := awsclient.StringValue(t.Key)
		val := awsclient.StringValue(t.Value)
		specMap[key] = awsclient.StringValue(t.Value)

		if currentVal, exists := currentMap[key]; exists {
			if currentVal != val {
				removeTags = append(removeTags, t.Key)
				addTags = append(addTags, &svcsdk.Tag{
					Key:   awsclient.String(key),
					Value: awsclient.String(val),
				})
			}
		} else {
			addTags = append(addTags, &svcsdk.Tag{
				Key:   awsclient.String(key),
				Value: awsclient.String(val),
			})
		}
	}

	for _, t := range current {
		key := awsclient.StringValue(t.Key)
		if _, exists := specMap[key]; !exists {
			removeTags = append(removeTags, awsclient.String(key))
		}
	}

	return addTags, removeTags
}

// AddExternalTags to spec if they don't exist
func AddExternalTags(mg resource.Managed, spec []*svcapitypes.Tag) []*svcapitypes.Tag {
	tagMap := make(map[string]struct{}, len(spec))
	for _, t := range spec {
		tagMap[awsclient.StringValue(t.Key)] = struct{}{}
	}

	tags := spec
	for _, t := range GetExternalTags(mg) {
		if _, exists := tagMap[awsclient.StringValue(t.Key)]; !exists {
			tags = append(tags, t)
		}
	}

	return tags
}

// GetExternalTags is a wrapper around resource.GetExternalTags to return a sorted array instead of a map
func GetExternalTags(mg resource.Managed) []*svcapitypes.Tag {
	externalTags := []*svcapitypes.Tag{}
	for k, v := range resource.GetExternalTags(mg) {
		externalTags = append(externalTags, &svcapitypes.Tag{Key: awsclient.String(k), Value: awsclient.String(v)})
	}

	sort.Slice(externalTags, func(i, j int) bool {
		return awsclient.StringValue(externalTags[i].Key) > awsclient.StringValue(externalTags[j].Key)
	})

	return externalTags
}
