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
	awss3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/crossplane/crossplane-runtime/pkg/meta"

	"github.com/crossplane/provider-aws/apis/s3/v1beta1"
	awsclient "github.com/crossplane/provider-aws/pkg/clients"
	"github.com/crossplane/provider-aws/pkg/clients/s3"
)

const (
	accelGetFailed = "cannot get Bucket accelerate configuration"
	accelPutFailed = "cannot put Bucket accelerate configuration"
)

// AccelerateConfigurationClient is the client for API methods and reconciling the AccelerateConfiguration
type AccelerateConfigurationClient struct {
	client s3.BucketClient
}

// NewAccelerateConfigurationClient creates the client for Accelerate Configuration
func NewAccelerateConfigurationClient(client s3.BucketClient) *AccelerateConfigurationClient {
	return &AccelerateConfigurationClient{client: client}
}

// Observe checks if the resource exists and if it matches the local configuration
func (in *AccelerateConfigurationClient) Observe(ctx context.Context, bucket *v1beta1.Bucket) (ResourceStatus, error) {
	external, err := in.client.GetBucketAccelerateConfiguration(ctx, &awss3.GetBucketAccelerateConfigurationInput{Bucket: awsclient.String(meta.GetExternalName(bucket))})
	if err != nil {
		// Short stop method for requests in a region without Acceleration Support
		if s3.MethodNotSupported(err) || s3.ArgumentNotSupported(err) {
			return Updated, nil
		}
		return NeedsUpdate, awsclient.Wrap(err, accelGetFailed)
	}
	if bucket.Spec.ForProvider.AccelerateConfiguration != nil &&
		bucket.Spec.ForProvider.AccelerateConfiguration.Status != string(external.Status) {
		return NeedsUpdate, nil
	}
	return Updated, nil
}

// CreateOrUpdate sends a request to have resource created on AWS
func (in *AccelerateConfigurationClient) CreateOrUpdate(ctx context.Context, bucket *v1beta1.Bucket) error {
	if bucket.Spec.ForProvider.AccelerateConfiguration == nil {
		return nil
	}
	input := GenerateAccelerateConfigurationInput(meta.GetExternalName(bucket), bucket.Spec.ForProvider.AccelerateConfiguration)
	_, err := in.client.PutBucketAccelerateConfiguration(ctx, input)
	return awsclient.Wrap(err, accelPutFailed)
}

// Delete does not do anything since AccelerateConfiguration doesn't have Delete call.
func (*AccelerateConfigurationClient) Delete(_ context.Context, _ *v1beta1.Bucket) error {
	return nil
}

// LateInitialize is responsible for initializing the resource based on the external value
func (in *AccelerateConfigurationClient) LateInitialize(ctx context.Context, bucket *v1beta1.Bucket) error {
	external, err := in.client.GetBucketAccelerateConfiguration(ctx, &awss3.GetBucketAccelerateConfigurationInput{Bucket: awsclient.String(meta.GetExternalName(bucket))})
	if err != nil {
		// Short stop method for requests without Acceleration Support
		if s3.MethodNotSupported(err) || s3.ArgumentNotSupported(err) {
			return nil
		}
		return awsclient.Wrap(err, accelGetFailed)
	}

	// We need the second check here because by default the accelerateConfig status is not set
	if external == nil || len(external.Status) == 0 {
		return nil
	}

	if bucket.Spec.ForProvider.AccelerateConfiguration == nil {
		bucket.Spec.ForProvider.AccelerateConfiguration = &v1beta1.AccelerateConfiguration{}
	}

	bucket.Spec.ForProvider.AccelerateConfiguration.Status = awsclient.LateInitializeString(
		bucket.Spec.ForProvider.AccelerateConfiguration.Status,
		awsclient.String(string(external.Status)))
	return nil
}

// SubresourceExists checks if the subresource this controller manages currently exists
func (in *AccelerateConfigurationClient) SubresourceExists(bucket *v1beta1.Bucket) bool {
	return bucket.Spec.ForProvider.AccelerateConfiguration != nil
}

// GenerateAccelerateConfigurationInput creates the input for the AccelerateConfiguration request for the S3 Client
func GenerateAccelerateConfigurationInput(name string, config *v1beta1.AccelerateConfiguration) *awss3.PutBucketAccelerateConfigurationInput {
	return &awss3.PutBucketAccelerateConfigurationInput{
		Bucket:                  awsclient.String(name),
		AccelerateConfiguration: &awss3types.AccelerateConfiguration{Status: awss3types.BucketAccelerateStatus(config.Status)},
	}
}
