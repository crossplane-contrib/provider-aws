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

package bucketclients

import (
	"context"

	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/pkg/errors"

	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"

	"github.com/crossplane/provider-aws/apis/s3/v1beta1"
	aws "github.com/crossplane/provider-aws/pkg/clients"
	"github.com/crossplane/provider-aws/pkg/clients/s3"
)

// VersioningConfigurationClient is the client for API methods and reconciling the VersioningConfiguration
type VersioningConfigurationClient struct {
	config *v1beta1.VersioningConfiguration
	client s3.BucketClient
}

// NewVersioningConfigurationClient creates the client for Versioning Configuration
func NewVersioningConfigurationClient(bucket *v1beta1.Bucket, client s3.BucketClient) *VersioningConfigurationClient {
	return &VersioningConfigurationClient{config: bucket.Spec.Parameters.VersioningConfiguration, client: client}
}

// Observe checks if the resource exists and if it matches the local configuration
func (in *VersioningConfigurationClient) Observe(ctx context.Context, bucket *v1beta1.Bucket) (ResourceStatus, error) {
	vers, err := in.client.GetBucketVersioningRequest(&awss3.GetBucketVersioningInput{Bucket: aws.String(meta.GetExternalName(bucket))}).Send(ctx)
	if err != nil {
		return NeedsUpdate, errors.Wrap(err, "cannot get bucket encryption")
	}

	if vers.Status == "" && vers.MFADelete == "" && in.config == nil {
		return Updated, nil
	} else if vers.GetBucketVersioningOutput != nil && in.config == nil {
		return NeedsDeletion, nil
	}

	if string(vers.Status) != in.config.Status {
		return NeedsUpdate, nil
	}
	if string(vers.MFADelete) != aws.StringValue(in.config.MFADelete) {
		return NeedsUpdate, nil
	}
	return Updated, nil
}

// GeneratePutBucketVersioningInput creates the input for the PutBucketVersioning request for the S3 Client
func (in *VersioningConfigurationClient) GeneratePutBucketVersioningInput(name string) *awss3.PutBucketVersioningInput {
	return &awss3.PutBucketVersioningInput{
		Bucket: aws.String(name),
		VersioningConfiguration: &awss3.VersioningConfiguration{
			MFADelete: awss3.MFADelete(aws.StringValue(in.config.MFADelete)),
			Status:    awss3.BucketVersioningStatus(in.config.Status),
		},
	}
}

// Create sends a request to have resource created on AWS.
func (in *VersioningConfigurationClient) Create(ctx context.Context, bucket *v1beta1.Bucket) (managed.ExternalUpdate, error) {
	if in.config == nil {
		return managed.ExternalUpdate{}, nil
	}
	_, err := in.client.PutBucketVersioningRequest(in.GeneratePutBucketVersioningInput(meta.GetExternalName(bucket))).Send(ctx)
	return managed.ExternalUpdate{}, errors.Wrap(err, "cannot put bucket versioning")
}

// Delete creates the request to delete the resource on AWS or set it to the default value.
func (in *VersioningConfigurationClient) Delete(ctx context.Context, bucket *v1beta1.Bucket) error {
	input := &awss3.PutBucketVersioningInput{
		Bucket:                  aws.String(meta.GetExternalName(bucket)),
		VersioningConfiguration: &awss3.VersioningConfiguration{Status: awss3.BucketVersioningStatusSuspended},
	}
	_, err := in.client.PutBucketVersioningRequest(input).Send(ctx)
	return errors.Wrap(err, "cannot delete bucket versioning configuration")
}
