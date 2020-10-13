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
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/pkg/errors"

	"github.com/crossplane/provider-aws/apis/s3/v1beta1"
	aws "github.com/crossplane/provider-aws/pkg/clients"
	"github.com/crossplane/provider-aws/pkg/clients/s3"
)

const (
	versioningGetFailed = "cannot get Bucket versioning configuration"
	versioningPutFailed = "cannot put Bucket versioning configuration"
)

// VersioningConfigurationClient is the client for API methods and reconciling the VersioningConfiguration
type VersioningConfigurationClient struct {
	client s3.BucketClient
}

// LateInitialize is responsible for initializing the resource based on the external value
func (in *VersioningConfigurationClient) LateInitialize(ctx context.Context, bucket *v1beta1.Bucket) error {
	external, err := in.client.GetBucketVersioningRequest(&awss3.GetBucketVersioningInput{Bucket: aws.String(meta.GetExternalName(bucket))}).Send(ctx)
	if err != nil {
		return errors.Wrap(err, versioningGetFailed)
	}

	if len(external.Status) == 0 && len(external.MFADelete) == 0 {
		return nil
	}
	config := bucket.Spec.ForProvider.VersioningConfiguration
	if config == nil {
		bucket.Spec.ForProvider.VersioningConfiguration = &v1beta1.VersioningConfiguration{}
		config = bucket.Spec.ForProvider.VersioningConfiguration
	}
	config.Status = aws.LateInitializeStringPtr(config.Status, aws.String(string(external.Status)))
	config.MFADelete = aws.LateInitializeStringPtr(config.MFADelete, aws.String(string(external.MFADelete)))
	return nil
}

// NewVersioningConfigurationClient creates the client for Versioning Configuration
func NewVersioningConfigurationClient(client s3.BucketClient) *VersioningConfigurationClient {
	return &VersioningConfigurationClient{client: client}
}

// Observe checks if the resource exists and if it matches the local configuration
func (in *VersioningConfigurationClient) Observe(ctx context.Context, bucket *v1beta1.Bucket) (ResourceStatus, error) { // nolint:gocyclo
	external, err := in.client.GetBucketVersioningRequest(&awss3.GetBucketVersioningInput{Bucket: aws.String(meta.GetExternalName(bucket))}).Send(ctx)
	if err != nil {
		return NeedsUpdate, errors.Wrap(err, versioningGetFailed)
	}
	if bucket.Spec.ForProvider.VersioningConfiguration == nil {
		return Updated, nil
	}
	if string(external.Status) != aws.StringValue(bucket.Spec.ForProvider.VersioningConfiguration.Status) ||
		string(external.MFADelete) != aws.StringValue(bucket.Spec.ForProvider.VersioningConfiguration.MFADelete) {
		return NeedsUpdate, nil
	}
	return Updated, nil
}

// GeneratePutBucketVersioningInput creates the input for the PutBucketVersioning request for the S3 Client
func GeneratePutBucketVersioningInput(name string, config *v1beta1.VersioningConfiguration) *awss3.PutBucketVersioningInput {
	return &awss3.PutBucketVersioningInput{
		Bucket: aws.String(name),
		VersioningConfiguration: &awss3.VersioningConfiguration{
			MFADelete: awss3.MFADelete(aws.StringValue(config.MFADelete)),
			Status:    awss3.BucketVersioningStatus(aws.StringValue(config.Status)),
		},
	}
}

// CreateOrUpdate sends a request to have resource created on AWS.
func (in *VersioningConfigurationClient) CreateOrUpdate(ctx context.Context, bucket *v1beta1.Bucket) error {
	if bucket.Spec.ForProvider.VersioningConfiguration == nil {
		return nil
	}
	input := GeneratePutBucketVersioningInput(meta.GetExternalName(bucket), bucket.Spec.ForProvider.VersioningConfiguration)
	_, err := in.client.PutBucketVersioningRequest(input).Send(ctx)
	return errors.Wrap(err, versioningPutFailed)
}

// Delete does nothing because there is no corresponding deletion call in AWS.
func (*VersioningConfigurationClient) Delete(_ context.Context, _ *v1beta1.Bucket) error {
	return nil
}
