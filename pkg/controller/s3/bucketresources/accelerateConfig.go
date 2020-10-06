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

// AccelerateConfigurationClient is the client for API methods and reconciling the AccelerateConfiguration
type AccelerateConfigurationClient struct {
	client s3.BucketClient
}

// LateInitialize is responsible for initializing the resource based on the external value
func (in *AccelerateConfigurationClient) LateInitialize(ctx context.Context, bucket *v1beta1.Bucket) error {
	external, err := in.client.GetBucketAccelerateConfigurationRequest(&awss3.GetBucketAccelerateConfigurationInput{Bucket: aws.String(meta.GetExternalName(bucket))}).Send(ctx)
	if err != nil {
		return errors.Wrap(err, accelGetFailed)
	}

	// We need the second check here because by default the accelerateConfig status is not set
	// by default
	if external.GetBucketAccelerateConfigurationOutput == nil || len(external.Status) == 0 {
		return nil
	}

	config := bucket.Spec.ForProvider.AccelerateConfiguration
	if config == nil {
		bucket.Spec.ForProvider.AccelerateConfiguration = &v1beta1.AccelerateConfiguration{}
		config = bucket.Spec.ForProvider.AccelerateConfiguration
	}
	config.Status = aws.LateInitializeString(config.Status, aws.String(string(external.GetBucketAccelerateConfigurationOutput.Status)))
	return nil
}

// NewAccelerateConfigurationClient creates the client for Accelerate Configuration
func NewAccelerateConfigurationClient(client s3.BucketClient) *AccelerateConfigurationClient {
	return &AccelerateConfigurationClient{client: client}
}

// Observe checks if the resource exists and if it matches the local configuration
func (in *AccelerateConfigurationClient) Observe(ctx context.Context, bucket *v1beta1.Bucket) (ResourceStatus, error) {
	external, err := in.client.GetBucketAccelerateConfigurationRequest(&awss3.GetBucketAccelerateConfigurationInput{Bucket: aws.String(meta.GetExternalName(bucket))}).Send(ctx)
	if err != nil {
		return NeedsUpdate, errors.Wrap(err, accelGetFailed)
	}

	config := bucket.Spec.ForProvider.AccelerateConfiguration
	if external.Status == "" && config == nil {
		return Updated, nil
	} else if external.Status != "" && config == nil {
		return NeedsDeletion, nil
	}

	if string(external.Status) != config.Status {
		return NeedsUpdate, nil
	}

	return Updated, nil
}

// GenerateAccelerateConfigurationInput creates the input for the AccelerateConfiguration request for the S3 Client
func GenerateAccelerateConfigurationInput(name string, config *v1beta1.AccelerateConfiguration) *awss3.PutBucketAccelerateConfigurationInput {
	return &awss3.PutBucketAccelerateConfigurationInput{
		Bucket:                  aws.String(name),
		AccelerateConfiguration: &awss3.AccelerateConfiguration{Status: awss3.BucketAccelerateStatus(config.Status)},
	}
}

// CreateOrUpdate sends a request to have resource created on AWS
func (in *AccelerateConfigurationClient) CreateOrUpdate(ctx context.Context, bucket *v1beta1.Bucket) (managed.ExternalUpdate, error) {
	config := bucket.Spec.ForProvider.AccelerateConfiguration
	if config == nil {
		return managed.ExternalUpdate{}, nil
	}
	_, err := in.client.PutBucketAccelerateConfigurationRequest(GenerateAccelerateConfigurationInput(meta.GetExternalName(bucket), config)).Send(ctx)
	return managed.ExternalUpdate{}, errors.Wrap(err, accelPutFailed)
}

// Delete creates the request to delete the resource on AWS or set it to the default value.
func (in *AccelerateConfigurationClient) Delete(ctx context.Context, bucket *v1beta1.Bucket) error {
	_, err := in.client.PutBucketAccelerateConfigurationRequest(
		&awss3.PutBucketAccelerateConfigurationInput{
			Bucket: aws.String(meta.GetExternalName(bucket)),
			AccelerateConfiguration: &awss3.AccelerateConfiguration{
				Status: awss3.BucketAccelerateStatusSuspended,
			},
		},
	).Send(ctx)
	return errors.Wrap(err, accelDeleteFailed)
}
