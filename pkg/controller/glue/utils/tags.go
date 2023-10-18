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

package utils

import (
	"context"

	svcsdk "github.com/aws/aws-sdk-go/service/glue"
	"github.com/aws/aws-sdk-go/service/glue/glueiface"
	"github.com/pkg/errors"

	errorutils "github.com/crossplane-contrib/provider-aws/pkg/utils/errors"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/tags"
)

const (
	errTagResource        = "cannot tag Glue resource"
	errUntagResource      = "cannot untag Glue resource"
	errGetTagsForResource = "cannot get tags for Glue resource"
)

// GetTagsForResource for the given resource with arn
func GetTagsForResource(ctx context.Context, client glueiface.GlueAPI, arn *string) (map[string]*string, error) {
	req := &svcsdk.GetTagsInput{
		ResourceArn: arn,
	}

	resp, err := client.GetTagsWithContext(ctx, req)
	if err != nil {
		return nil, errors.Wrap(err, errGetTagsForResource)
	}

	return resp.Tags, nil
}

// UpdateTagsForResource with arn
func UpdateTagsForResource(ctx context.Context, client glueiface.GlueAPI, spec map[string]*string, arn *string) error {

	current, err := GetTagsForResource(ctx, client, arn)
	if err != nil {
		return err
	}

	add, remove := tags.DiffTagsMapPtr(spec, current)

	if len(remove) > 0 {
		_, err := client.UntagResourceWithContext(ctx, &svcsdk.UntagResourceInput{
			ResourceArn:  arn,
			TagsToRemove: remove,
		})
		if err != nil {
			return errorutils.Wrap(err, errUntagResource)
		}
	}
	// remove before add for case where we just simply update a tag
	// DiffTagsMapPtr removes "outdated" tags which includes the updated version of tags
	// other way around it still works, but we would get and need another go over this function
	if len(add) > 0 {
		_, err := client.TagResourceWithContext(ctx, &svcsdk.TagResourceInput{
			ResourceArn: arn,
			TagsToAdd:   add,
		})
		if err != nil {
			return errorutils.Wrap(err, errTagResource)
		}
	}

	return nil
}

// AreTagsUpToDate for spec and arn
func AreTagsUpToDate(client glueiface.GlueAPI, spec map[string]*string, arn *string) (bool, error) {
	current, err := client.GetTags(&svcsdk.GetTagsInput{ResourceArn: arn})
	if err != nil {
		return false, err
	}

	add, remove := tags.DiffTagsMapPtr(spec, current.Tags)

	return len(add) == 0 && len(remove) == 0, nil
}
