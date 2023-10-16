/*
Copyright 2022 The Crossplane Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permission and
limitations under the License.
*/

package utils

import (
	"context"

	svcsdk "github.com/aws/aws-sdk-go/service/batch"
	"github.com/aws/aws-sdk-go/service/batch/batchiface"
	"github.com/pkg/errors"

	errorutils "github.com/crossplane-contrib/provider-aws/pkg/utils/errors"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/tags"
)

const (
	errTagResource         = "cannot tag Batch resource"
	errUntagResource       = "cannot untag Batch resource"
	errListTagsForResource = "cannot list tags for Batch resource"
)

// ListTagsForResource for the given resource
func ListTagsForResource(client batchiface.BatchAPI, arn *string) (map[string]*string, error) {
	req := &svcsdk.ListTagsForResourceInput{
		ResourceArn: arn,
	}

	resp, err := client.ListTagsForResource(req)
	if err != nil {
		return nil, errors.Wrap(err, errListTagsForResource)
	}

	return resp.Tags, nil
}

// UpdateTagsForResource with resourceName
func UpdateTagsForResource(ctx context.Context, client batchiface.BatchAPI, spec map[string]*string, arn *string) error {

	current, err := ListTagsForResource(client, arn)
	if err != nil {
		return err
	}

	add, remove := tags.DiffTagsMapPtr(spec, current)

	if len(remove) > 0 {
		_, err := client.UntagResourceWithContext(ctx, &svcsdk.UntagResourceInput{
			ResourceArn: arn,
			TagKeys:     remove,
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
			Tags:        add,
		})
		if err != nil {
			return errorutils.Wrap(err, errTagResource)
		}
	}

	return nil
}
