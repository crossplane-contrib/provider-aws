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

var _ BucketResource = &RequestPaymentConfigurationClient{}

// RequestPaymentConfigurationClient is the client for API methods and reconciling the PaymentConfiguration
type RequestPaymentConfigurationClient struct {
	config *v1beta1.PaymentConfiguration
	client s3.BucketClient
}

// LateInitialize is responsible for initializing the resource based on the external value
func (in *RequestPaymentConfigurationClient) LateInitialize(ctx context.Context, bucket *v1beta1.Bucket) error {
	conf, err := in.client.GetBucketRequestPaymentRequest(&awss3.GetBucketRequestPaymentInput{Bucket: aws.String(meta.GetExternalName(bucket))}).Send(ctx)
	if err != nil {
		return errors.Wrap(err, "cannot get request payment configuration")
	}

	if conf.GetBucketRequestPaymentOutput == nil {
		return nil
	}
	if in.config == nil {
		bucket.Spec.ForProvider.PayerConfiguration = &v1beta1.PaymentConfiguration{}
		in.config = bucket.Spec.ForProvider.PayerConfiguration
	}
	in.config.Payer = aws.LateInitializeString(in.config.Payer, aws.String(string(conf.Payer)))
	return nil
}

// NewRequestPaymentConfigurationClient creates the client for Payment Configuration
func NewRequestPaymentConfigurationClient(bucket *v1beta1.Bucket, client s3.BucketClient) *RequestPaymentConfigurationClient {
	return &RequestPaymentConfigurationClient{config: bucket.Spec.ForProvider.PayerConfiguration, client: client}
}

// Observe checks if the resource exists and if it matches the local configuration
func (in *RequestPaymentConfigurationClient) Observe(ctx context.Context, bucket *v1beta1.Bucket) (ResourceStatus, error) {
	conf, err := in.client.GetBucketRequestPaymentRequest(&awss3.GetBucketRequestPaymentInput{Bucket: aws.String(meta.GetExternalName(bucket))}).Send(ctx)
	if err != nil {
		return NeedsUpdate, errors.Wrap(err, "cannot get request payment configuration")
	}

	if in.config.Payer != string(conf.Payer) {
		return NeedsUpdate, nil
	}

	return Updated, nil
}

// GeneratePutBucketPaymentInput creates the input for the BucketRequestPayment request for the S3 Client
func GeneratePutBucketPaymentInput(name string, in *RequestPaymentConfigurationClient) *awss3.PutBucketRequestPaymentInput {
	bci := &awss3.PutBucketRequestPaymentInput{
		Bucket:                      aws.String(name),
		RequestPaymentConfiguration: &awss3.RequestPaymentConfiguration{Payer: awss3.Payer(in.config.Payer)},
	}

	return bci
}

// CreateOrUpdate sends a request to have resource created on AWS.
func (in *RequestPaymentConfigurationClient) CreateOrUpdate(ctx context.Context, bucket *v1beta1.Bucket) (managed.ExternalUpdate, error) {
	if in.config == nil {
		return managed.ExternalUpdate{}, nil
	}
	_, err := in.client.PutBucketRequestPaymentRequest(GeneratePutBucketPaymentInput(meta.GetExternalName(bucket), in)).Send(ctx)
	return managed.ExternalUpdate{}, errors.Wrap(err, "cannot put bucket payment")
}

// Delete creates the request to delete the resource on AWS or set it to the default value.
func (in *RequestPaymentConfigurationClient) Delete(ctx context.Context, bucket *v1beta1.Bucket) error {
	// There is no delete operation for the S3Bucket Payer
	// The value must be BucketOwner or Requester
	return nil
}
