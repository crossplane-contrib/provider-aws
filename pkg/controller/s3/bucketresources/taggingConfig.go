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

package bucketresources

import (
	"context"

	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"

	"github.com/crossplane/provider-aws/apis/s3/v1beta1"
	aws "github.com/crossplane/provider-aws/pkg/clients"
	"github.com/crossplane/provider-aws/pkg/clients/s3"
)

// TaggingConfigurationClient is the client for API methods and reconciling the CORSConfiguration
type TaggingConfigurationClient struct {
	client s3.BucketClient
}

// LateInitialize is responsible for initializing the resource based on the external value
func (in *TaggingConfigurationClient) LateInitialize(ctx context.Context, bucket *v1beta1.Bucket) error {
	// GetBucketTaggingRequest throws an error if nothing exists externally
	// Future work can be done to support brownfield initialization for the TaggingConfiguration
	// TODO
	return nil
}

// NewTaggingConfigurationClient creates the client for CORS Configuration
func NewTaggingConfigurationClient(client s3.BucketClient) *TaggingConfigurationClient {
	return &TaggingConfigurationClient{client: client}
}

// Observe checks if the resource exists and if it matches the local configuration
func (in *TaggingConfigurationClient) Observe(ctx context.Context, bucket *v1beta1.Bucket) (ResourceStatus, error) {
	external, err := in.client.GetBucketTaggingRequest(&awss3.GetBucketTaggingInput{Bucket: aws.String(meta.GetExternalName(bucket))}).Send(ctx)
	config := bucket.Spec.ForProvider.BucketTagging
	if err != nil {
		if s3.TaggingNotFound(err) && config == nil {
			return Updated, nil
		}
		return NeedsUpdate, errors.Wrap(err, taggingGetFailed)
	}

	switch {
	case config == nil && len(external.TagSet) == 0:
		return Updated, nil
	case config == nil && len(external.TagSet) != 0:
		return NeedsDeletion, nil
	case cmp.Equal(external.TagSet, GenerateTagging(config).TagSet):
		return Updated, nil
	default:
		return NeedsUpdate, nil
	}
}

// GenerateTagging creates the Tagging for the AWS SDK
func GenerateTagging(config *v1beta1.Tagging) *awss3.Tagging {
	if config == nil || config.TagSet == nil {
		return &awss3.Tagging{TagSet: make([]awss3.Tag, 0)}
	}
	conf := &awss3.Tagging{TagSet: make([]awss3.Tag, len(config.TagSet))}
	for i, v := range config.TagSet {
		conf.TagSet[i] = awss3.Tag{
			Key:   aws.String(v.Key),
			Value: aws.String(v.Value),
		}
	}
	return conf
}

// GeneratePutBucketTagging creates the PutBucketTaggingInput for the aws SDK
func GeneratePutBucketTagging(name string, config *v1beta1.Tagging) *awss3.PutBucketTaggingInput {
	return &awss3.PutBucketTaggingInput{
		Bucket:  aws.String(name),
		Tagging: GenerateTagging(config),
	}
}

// CreateOrUpdate sends a request to have resource created on AWS
func (in *TaggingConfigurationClient) CreateOrUpdate(ctx context.Context, bucket *v1beta1.Bucket) (managed.ExternalUpdate, error) {
	config := bucket.Spec.ForProvider.BucketTagging
	if config == nil {
		return managed.ExternalUpdate{}, nil
	}
	_, err := in.client.PutBucketTaggingRequest(GeneratePutBucketTagging(meta.GetExternalName(bucket), config)).Send(ctx)
	return managed.ExternalUpdate{}, errors.Wrap(err, taggingPutFailed)
}

// Delete creates the request to delete the resource on AWS or set it to the default value.
func (in *TaggingConfigurationClient) Delete(ctx context.Context, bucket *v1beta1.Bucket) error {
	_, err := in.client.DeleteBucketTaggingRequest(
		&awss3.DeleteBucketTaggingInput{
			Bucket: aws.String(meta.GetExternalName(bucket)),
		},
	).Send(ctx)
	return errors.Wrap(err, taggingDeleteFailed)
}
