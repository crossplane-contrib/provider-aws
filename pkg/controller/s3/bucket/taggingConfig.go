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

package bucket

import (
	"context"

	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go/document"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/crossplane-contrib/provider-aws/apis/s3/v1beta1"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/s3"
	errorutils "github.com/crossplane-contrib/provider-aws/pkg/utils/errors"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
)

const (
	taggingGetFailed    = "cannot get Bucket tagging set"
	taggingPutFailed    = "cannot put Bucket tagging set"
	taggingDeleteFailed = "cannot delete Bucket tagging set"
)

// TaggingConfigurationClient is the client for API methods and reconciling the CORSConfiguration
type TaggingConfigurationClient struct {
	client s3.BucketClient
}

// NewTaggingConfigurationClient creates the client for CORS Configuration
func NewTaggingConfigurationClient(client s3.BucketClient) *TaggingConfigurationClient {
	return &TaggingConfigurationClient{client: client}
}

// Observe checks if the resource exists and if it matches the local configuration
func (in *TaggingConfigurationClient) Observe(ctx context.Context, bucket *v1beta1.Bucket) (ResourceStatus, error) {
	external, err := in.client.GetBucketTagging(ctx, &awss3.GetBucketTaggingInput{Bucket: pointer.ToOrNilIfZeroValue(meta.GetExternalName(bucket))})
	config := bucket.Spec.ForProvider.BucketTagging
	if err != nil {
		if s3.TaggingNotFound(err) && config == nil {
			return Updated, nil
		}
		return NeedsUpdate, errorutils.Wrap(resource.Ignore(s3.TaggingNotFound, err), taggingGetFailed)
	}

	switch {
	case config == nil && len(external.TagSet) == 0:
		return Updated, nil
	case config == nil && len(external.TagSet) != 0:
		return NeedsDeletion, nil
	case cmp.Equal(s3.SortS3TagSet(external.TagSet), s3.SortS3TagSet(GenerateTagging(config).TagSet), cmpopts.IgnoreTypes(document.NoSerde{})):
		return Updated, nil
	default:
		return NeedsUpdate, nil
	}
}

// CreateOrUpdate sends a request to have resource created on AWS
func (in *TaggingConfigurationClient) CreateOrUpdate(ctx context.Context, bucket *v1beta1.Bucket) error {
	if bucket.Spec.ForProvider.BucketTagging == nil {
		return nil
	}
	input := GeneratePutBucketTagging(meta.GetExternalName(bucket), bucket.Spec.ForProvider.BucketTagging)
	_, err := in.client.PutBucketTagging(ctx, input)
	return errorutils.Wrap(err, taggingPutFailed)
}

// Delete creates the request to delete the resource on AWS or set it to the default value.
func (in *TaggingConfigurationClient) Delete(ctx context.Context, bucket *v1beta1.Bucket) error {
	_, err := in.client.DeleteBucketTagging(ctx,
		&awss3.DeleteBucketTaggingInput{
			Bucket: pointer.ToOrNilIfZeroValue(meta.GetExternalName(bucket)),
		},
	)
	return errorutils.Wrap(err, taggingDeleteFailed)
}

// LateInitialize does nothing because the resource might have been deleted by
// the user.
func (in *TaggingConfigurationClient) LateInitialize(ctx context.Context, bucket *v1beta1.Bucket) error {
	external, err := in.client.GetBucketTagging(ctx, &awss3.GetBucketTaggingInput{Bucket: pointer.ToOrNilIfZeroValue(meta.GetExternalName(bucket))})
	if err != nil {
		return errorutils.Wrap(resource.Ignore(s3.TaggingNotFound, err), taggingGetFailed)
	}

	// We need the second check here because by default the tags are not set
	if external == nil || len(external.TagSet) == 0 {
		return nil
	}

	fp := &bucket.Spec.ForProvider
	if fp.BucketTagging == nil {
		fp.BucketTagging = &v1beta1.Tagging{}
	}

	if fp.BucketTagging.TagSet == nil {
		fp.BucketTagging = GenerateLocalTagging(external.TagSet)
	}

	return nil
}

// SubresourceExists checks if the subresource this controller manages currently exists
func (in *TaggingConfigurationClient) SubresourceExists(bucket *v1beta1.Bucket) bool {
	return bucket.Spec.ForProvider.BucketTagging != nil
}

// GenerateTagging creates the awss3.Tagging for the AWS SDK
func GenerateTagging(config *v1beta1.Tagging) *types.Tagging {
	if config == nil || config.TagSet == nil {
		return &types.Tagging{TagSet: make([]types.Tag, 0)}
	}
	return &types.Tagging{TagSet: s3.CopyTags(config.TagSet)}
}

// GenerateLocalTagging creates the v1beta1.Tagging from the AWS SDK tagging
func GenerateLocalTagging(config []types.Tag) *v1beta1.Tagging {
	if config == nil {
		return nil
	}
	return &v1beta1.Tagging{TagSet: s3.CopyAWSTags(config)}
}

// GeneratePutBucketTagging creates the PutBucketTaggingInput for the aws SDK
func GeneratePutBucketTagging(name string, config *v1beta1.Tagging) *awss3.PutBucketTaggingInput {
	return &awss3.PutBucketTaggingInput{
		Bucket:  pointer.ToOrNilIfZeroValue(name),
		Tagging: GenerateTagging(config),
	}
}
