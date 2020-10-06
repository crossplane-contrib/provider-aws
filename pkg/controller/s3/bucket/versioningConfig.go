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
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/pkg/errors"

	"github.com/crossplane/provider-aws/apis/s3/v1beta1"
	aws "github.com/crossplane/provider-aws/pkg/clients"
	"github.com/crossplane/provider-aws/pkg/clients/s3"
)

const (
	versioningGetFailed    = "cannot get Bucket versioning configuration"
	versioningPutFailed    = "cannot put Bucket versioning configuration"
	versioningDeleteFailed = "cannot delete Bucket versioning configuration"
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
	if len(external.Status) != 0 { // By default Status is the string ""
		config.Status = aws.String(string(external.Status))
	}
	if len(external.MFADelete) != 0 { // By default MFADelete is the string ""
		config.MFADelete = aws.String(string(external.MFADelete))
	}
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
	config := bucket.Spec.ForProvider.VersioningConfiguration
	switch {
	case len(external.Status) == 0 && len(external.MFADelete) == 0 && config == nil:
		return Updated, nil
	case (len(external.Status) != 0 || len(external.MFADelete) != 0) && config == nil:
		return NeedsDeletion, nil
	case aws.StringValue(config.Status) == string(external.Status) && aws.StringValue(config.MFADelete) == string(external.MFADelete):
		return Updated, nil
	default:
		return NeedsUpdate, nil
	}
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
func (in *VersioningConfigurationClient) CreateOrUpdate(ctx context.Context, bucket *v1beta1.Bucket) (managed.ExternalUpdate, error) {
	config := bucket.Spec.ForProvider.VersioningConfiguration
	if config == nil {
		return managed.ExternalUpdate{}, nil
	}
	_, err := in.client.PutBucketVersioningRequest(GeneratePutBucketVersioningInput(meta.GetExternalName(bucket), config)).Send(ctx)
	return managed.ExternalUpdate{}, errors.Wrap(err, versioningPutFailed)
}

// Delete creates the request to delete the resource on AWS or set it to the default value.
func (in *VersioningConfigurationClient) Delete(ctx context.Context, bucket *v1beta1.Bucket) error {
	input := &awss3.PutBucketVersioningInput{
		Bucket: aws.String(meta.GetExternalName(bucket)),
		VersioningConfiguration: &awss3.VersioningConfiguration{
			Status:    awss3.BucketVersioningStatusSuspended,
			MFADelete: awss3.MFADeleteDisabled,
		},
	}
	_, err := in.client.PutBucketVersioningRequest(input).Send(ctx)
	return errors.Wrap(err, versioningDeleteFailed)
}
