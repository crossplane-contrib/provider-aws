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

var _ BucketResource = &AccelerateConfigurationClient{}

var (
	enabled   = "Enabled"
	suspended = "Suspended"
	errBoom   = errors.New("boom")
)

// AccelerateConfigurationClient is the client for API methods and reconciling the AccelerateConfiguration
type AccelerateConfigurationClient struct {
	config *v1beta1.AccelerateConfiguration
	client s3.BucketClient
}

// LateInitialize is responsible for initializing the resource based on the external value
func (in *AccelerateConfigurationClient) LateInitialize(ctx context.Context, bucket *v1beta1.Bucket) error {
	conf, err := in.client.GetBucketAccelerateConfigurationRequest(&awss3.GetBucketAccelerateConfigurationInput{Bucket: aws.String(meta.GetExternalName(bucket))}).Send(ctx)
	if err != nil {
		return errors.Wrap(err, accelGetFailed)
	}

	// We need the second check here because by default the accelerateConfig status is not set
	// by default
	if conf.GetBucketAccelerateConfigurationOutput == nil || len(conf.Status) == 0 {
		return nil
	}

	if in.config == nil {
		bucket.Spec.ForProvider.AccelerateConfiguration = &v1beta1.AccelerateConfiguration{}
		in.config = bucket.Spec.ForProvider.AccelerateConfiguration
	}
	in.config.Status = aws.LateInitializeString(in.config.Status, aws.String(string(conf.GetBucketAccelerateConfigurationOutput.Status)))
	return nil
}

// NewAccelerateConfigurationClient creates the client for Accelerate Configuration
func NewAccelerateConfigurationClient(bucket *v1beta1.Bucket, client s3.BucketClient) *AccelerateConfigurationClient {
	return &AccelerateConfigurationClient{config: bucket.Spec.ForProvider.AccelerateConfiguration, client: client}
}

// Observe checks if the resource exists and if it matches the local configuration
func (in *AccelerateConfigurationClient) Observe(ctx context.Context, bucket *v1beta1.Bucket) (ResourceStatus, error) {
	conf, err := in.client.GetBucketAccelerateConfigurationRequest(&awss3.GetBucketAccelerateConfigurationInput{Bucket: aws.String(meta.GetExternalName(bucket))}).Send(ctx)
	if err != nil {
		return NeedsUpdate, errors.Wrap(err, accelGetFailed)
	}

	if conf.Status == "" && in.config == nil {
		return Updated, nil
	} else if conf.Status != "" && in.config == nil {
		return NeedsDeletion, nil
	}

	if string(conf.Status) != in.config.Status {
		return NeedsUpdate, nil
	}

	return Updated, nil
}

// GenerateAccelerateConfigurationInput creates the input for the AccelerateConfiguration request for the S3 Client
func GenerateAccelerateConfigurationInput(name string, in *AccelerateConfigurationClient) *awss3.PutBucketAccelerateConfigurationInput {
	return &awss3.PutBucketAccelerateConfigurationInput{
		Bucket:                  aws.String(name),
		AccelerateConfiguration: &awss3.AccelerateConfiguration{Status: awss3.BucketAccelerateStatus(in.config.Status)},
	}
}

// CreateOrUpdate sends a request to have resource created on AWS
func (in *AccelerateConfigurationClient) CreateOrUpdate(ctx context.Context, bucket *v1beta1.Bucket) (managed.ExternalUpdate, error) {
	if in.config == nil {
		return managed.ExternalUpdate{}, nil
	}
	_, err := in.client.PutBucketAccelerateConfigurationRequest(GenerateAccelerateConfigurationInput(meta.GetExternalName(bucket), in)).Send(ctx)
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
