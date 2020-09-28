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
	"github.com/pkg/errors"

	"github.com/crossplane/provider-aws/apis/s3/v1beta1"
	aws "github.com/crossplane/provider-aws/pkg/clients"
	"github.com/crossplane/provider-aws/pkg/clients/s3"
)

var _ BucketResource = &VersioningConfigurationClient{}

// VersioningConfigurationClient is the client for API methods and reconciling the VersioningConfiguration
type VersioningConfigurationClient struct {
	config *v1beta1.VersioningConfiguration
	client s3.BucketClient
}

// LateInitialize is responsible for initializing the resource based on the external value
func (in *VersioningConfigurationClient) LateInitialize(ctx context.Context, bucket *v1beta1.Bucket) error {
	conf, err := in.client.GetBucketVersioningRequest(&awss3.GetBucketVersioningInput{Bucket: aws.String(meta.GetExternalName(bucket))}).Send(ctx)
	if err != nil {
		return errors.Wrap(err, versioningGetFailed)
	}

	if len(conf.Status) == 0 && len(conf.MFADelete) == 0 {
		return nil
	}
	if in.config == nil {
		bucket.Spec.ForProvider.VersioningConfiguration = &v1beta1.VersioningConfiguration{}
		in.config = bucket.Spec.ForProvider.VersioningConfiguration
	}
	if len(conf.Status) != 0 { // By default Status is the string ""
		in.config.Status = aws.String(string(conf.Status))
	}
	if len(conf.MFADelete) != 0 { // By default MFADelete is the string ""
		in.config.MFADelete = aws.String(string(conf.MFADelete))
	}
	return nil
}

// NewVersioningConfigurationClient creates the client for Versioning Configuration
func NewVersioningConfigurationClient(bucket *v1beta1.Bucket, client s3.BucketClient) *VersioningConfigurationClient {
	return &VersioningConfigurationClient{config: bucket.Spec.ForProvider.VersioningConfiguration, client: client}
}

// Observe checks if the resource exists and if it matches the local configuration
func (in *VersioningConfigurationClient) Observe(ctx context.Context, bucket *v1beta1.Bucket) (ResourceStatus, error) { // nolint:gocyclo
	vers, err := in.client.GetBucketVersioningRequest(&awss3.GetBucketVersioningInput{Bucket: aws.String(meta.GetExternalName(bucket))}).Send(ctx)
	if err != nil {
		return NeedsUpdate, errors.Wrap(err, versioningGetFailed)
	}

	switch {
	case len(vers.Status) == 0 && len(vers.MFADelete) == 0 && in.config == nil:
		return Updated, nil
	case (len(vers.Status) != 0 || len(vers.MFADelete) != 0) && in.config == nil:
		return NeedsDeletion, nil
	case aws.StringValue(in.config.Status) == string(vers.Status) && aws.StringValue(in.config.MFADelete) == string(vers.MFADelete):
		return Updated, nil
	default:
		return NeedsUpdate, nil
	}
}

// GeneratePutBucketVersioningInput creates the input for the PutBucketVersioning request for the S3 Client
func GeneratePutBucketVersioningInput(name string, in *VersioningConfigurationClient) *awss3.PutBucketVersioningInput {
	return &awss3.PutBucketVersioningInput{
		Bucket: aws.String(name),
		VersioningConfiguration: &awss3.VersioningConfiguration{
			MFADelete: awss3.MFADelete(aws.StringValue(in.config.MFADelete)),
			Status:    awss3.BucketVersioningStatus(aws.StringValue(in.config.Status)),
		},
	}
}

// CreateOrUpdate sends a request to have resource created on AWS.
func (in *VersioningConfigurationClient) CreateOrUpdate(ctx context.Context, bucket *v1beta1.Bucket) (managed.ExternalUpdate, error) {
	if in.config == nil {
		return managed.ExternalUpdate{}, nil
	}
	_, err := in.client.PutBucketVersioningRequest(GeneratePutBucketVersioningInput(meta.GetExternalName(bucket), in)).Send(ctx)
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
